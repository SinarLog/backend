package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"cloud.google.com/go/pubsub"
	"google.golang.org/api/iterator"
	"sinarlog.com/internal/entity"
)

var (
	once           sync.Once
	psService      *pubsubService
	roomPrefix     string = "room-"
	listenerPrefix string = "listener-"
)

type pubsubService struct {
	ps          *pubsub.Client
	collections map[string]*collection
	mu          sync.RWMutex
}

type collection struct {
	topic   *pubsub.Topic
	subs    []*pubsub.Subscription
	clients []string
	mu      sync.Mutex
}

func init() {
	t := time.NewTicker(10 * time.Second)

	go func() {
		for range t.C {
			for key := range psService.collections {
				if len(psService.collections[key].clients) == 0 {
					psService.collections[key].topic.Stop()
					delete(psService.collections, key)
				}
			}
		}
	}()
}

func NewPubSubService(ps *pubsub.Client) *pubsubService {
	if psService == nil {
		once.Do(func() {
			psService = new(pubsubService)

			psService.collections = make(map[string]*collection)
			psService.ps = ps
		})
	}

	return psService
}

// PublshChat publishes a chat into the pubsub service.
// This  will then be consumed by listeners, if there are.
func (service *pubsubService) PublishChat(ctx context.Context, topicID string, publisherID string, payload entity.Chat) error {
	topic, err := service.findOrCreateTopic(ctx, roomPrefix+topicID)
	if err != nil {
		return err
	}

	json, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("json marshal error: %s", err)
	}
	r := topic.Publish(ctx, &pubsub.Message{
		ID:   topicID,
		Data: json,
	})

	_, err = r.Get(ctx)
	if err != nil {
		return fmt.Errorf("pubsub: result.Get: %w", err)
	}
	return nil
}

// SubscribeChat lets a user to be subscribed to
// a topic (in this case is a room) and will receive
// incoming messages in that room.
func (service *pubsubService) SubscribeChat(ctx context.Context, topicID, listenerID string, channel chan entity.Chat) error {
	topicID = roomPrefix + topicID
	subID := topicID + "-" + listenerPrefix + listenerID
	sub, err := service.findOrCreateSubscription(ctx, topicID, subID)
	if err != nil {
		return err
	}

	if err := sub.Receive(ctx, func(ctx context.Context, m *pubsub.Message) {
		var chat entity.Chat
		if err := json.Unmarshal(m.Data, &chat); err != nil {
			log.Fatalf("unable to unmarshal incoming pubsub message: %s\n", err)
		}

		channel <- chat

		m.Ack()
	}); err != nil {
		return err
	}

	return nil
}

// UnregisterClient will unregisters client from the current
// hash list of clients when the client is no longer subscribing
// to the topic. This should be called when client is about to
// be disconnected from the chat.
func (service *pubsubService) UnregisterClient(ctx context.Context, topicID, listenerID string) error {
	collection, exist := service.collections[roomPrefix+topicID]
	if !exist {
		return fmt.Errorf("the topic is not found in hash")
	}

	collection.mu.Lock()
	defer collection.mu.Unlock()

	for i, v := range collection.clients {
		if v == listenerPrefix+listenerID {
			collection.clients = append(collection.clients[:i], collection.clients[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("unable to find client in hash")
}

// findOrCreateTopic is a helper function used to find
// or create a new topic. It firstly finds the topic id
// by the room id in the collections hash map. If it is
// not found, it then finds the topic in the google cloud
// itself. If still not found, we then create a new topic.
func (service *pubsubService) findOrCreateTopic(ctx context.Context, topicID string) (*pubsub.Topic, error) {
	service.mu.Lock()
	defer service.mu.Unlock()

	// Finds in current collections
	if col, found := service.collections[topicID]; found {
		return col.topic, nil
	}

	// Find topic in google cloud
	it := service.ps.Topics(ctx)
	for {
		t, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		if t.ID() == topicID {
			// Make new collection for the topic
			service.collections[t.ID()] = &collection{
				topic: t,
			}
			return t, nil
		}
	}

	// Create a new topic
	topic, err := service.ps.CreateTopic(ctx, topicID)
	if err != nil {
		return nil, err
	}

	// Make new collection for the topic
	service.collections[topic.ID()] = &collection{
		topic: topic,
	}
	return topic, nil
}

// findOrCreateSubscription is used to find a subscription
// based on the topic or creates if none is found. It is
// done by first looking through the hash map. If not found
// it looks through the subs of the given topic in the google
// cloud. If still not found it will create a new sub to the
// topic.
func (service *pubsubService) findOrCreateSubscription(ctx context.Context, topicID, listenerID string) (*pubsub.Subscription, error) {
	topic, err := service.findOrCreateTopic(ctx, topicID)
	if err != nil {
		return nil, err
	}

	service.collections[topic.ID()].mu.Lock()
	defer service.collections[topic.ID()].mu.Unlock()

	// Finds in current collections
	for _, sub := range service.collections[topic.ID()].subs {
		if sub.ID() == listenerID {
			return sub, nil
		}
	}

	// Finds in google cloud
	it := topic.Subscriptions(ctx)
	for {
		sub, err := it.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return nil, err
		}

		if sub.ID() == listenerID {
			// Store in collection
			service.addSubsToTopic(ctx, topic.ID(), sub)
			service.addClientToSubs(ctx, topic.ID(), listenerID)

			return sub, nil
		}
	}

	// Create new sub
	sub, err := service.ps.CreateSubscription(ctx, listenerID, pubsub.SubscriptionConfig{
		Topic: topic,
	})
	if err != nil {
		return nil, err
	}

	// Store in collection
	service.addSubsToTopic(ctx, topic.ID(), sub)
	service.addClientToSubs(ctx, topic.ID(), listenerID)

	return sub, nil
}

// addClientToSubs adds a listener (in this case is a client)
// to the sub of the given topic.
func (service *pubsubService) addClientToSubs(ctx context.Context, topicID, listenerID string) {
	if col, found := service.collections[topicID]; found {
		for _, v := range col.clients {
			if v == listenerID {
				return
			}
		}
		col.clients = append(col.clients, listenerID)
	}
}

// addSubsToTopic adds pubsub subscription instance to the
// given topic.
func (service *pubsubService) addSubsToTopic(ctx context.Context, topicID string, sub *pubsub.Subscription) {
	if col, found := service.collections[topicID]; found {
		for _, v := range col.subs {
			if v == sub {
				return
			}
		}
		col.subs = append(col.subs, sub)
	}
}
