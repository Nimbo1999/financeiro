package messaging

import "context"

type Publisher interface {
	PublishEvent(ctx context.Context, event *UserCreatedEvent) error
	Close() error
}

type PublisherV2 interface {
	Publisher
	IsHealthy() bool
}
