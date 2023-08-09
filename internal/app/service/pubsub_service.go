package service

import (
	"context"

	"sinarlog.com/internal/entity"
)

type IPubSubService interface {
	PublishChat(ctx context.Context, topicId string, publisherId string, payload entity.Chat) error
	SubscribeChat(ctx context.Context, topicId, listenerId string, channel chan entity.Chat) error
	UnregisterClient(ctx context.Context, topicId, listenerId string) error
}
