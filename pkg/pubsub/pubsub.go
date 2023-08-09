package pubsub

import (
	"context"
	"log"
	"sync"

	"cloud.google.com/go/pubsub"
	"google.golang.org/api/option"
)

var (
	once       sync.Once
	psInstance *PubSub
)

var (
	_defaultEnvironment string = "DEVELOPMENT"
	_defaultProjectId   string = "LOCAL"
)

type PubSub struct {
	keyPath   string
	env       string
	projectId string
	Client    *pubsub.Client
}

func GetPubSubClient(ctx context.Context, options ...Option) *PubSub {
	if psInstance == nil {
		once.Do(func() {
			psInstance = &PubSub{
				keyPath:   "",
				env:       _defaultEnvironment,
				projectId: _defaultProjectId,
			}

			for _, opt := range options {
				opt(psInstance)
			}

			psInstance.connect(ctx)
		})
	}

	return psInstance
}

func (psInstance *PubSub) connect(ctx context.Context) {
	switch psInstance.env {
	case "PRODUCTION":
		client, err := pubsub.NewClient(ctx, psInstance.projectId)
		if err != nil {
			log.Fatalf("unable to create pubsub client: %s\n", err)
		}

		psInstance.Client = client
	default:
		if psInstance.keyPath == "" {
			log.Fatalln("DEVELOPMENT mode for pubsub requires a service account path...")
		}

		opt := option.WithCredentialsFile(psInstance.keyPath)
		client, err := pubsub.NewClient(ctx, psInstance.projectId, opt)
		if err != nil {
			log.Fatalf("unable to create pubsub client: %s\n", err)
		}

		psInstance.Client = client
	}
}
