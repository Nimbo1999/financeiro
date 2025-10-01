package consumers

import (
	"context"

	"notification/internal/services"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer interface {
	Start(ctx context.Context) error
	Stop() error
}

type RabbitMQConsumer struct {
	conn              *amqp.Connection
	channel           *amqp.Channel
	welcomeQueue      string
	otpQueue          string
	notificationSvc   services.NotificationService
	prefetchCount     int
}

func NewRabbitMQConsumer(
	rabbitmqURL string,
	welcomeQueue string,
	otpQueue string,
	prefetchCount int,
	notificationSvc services.NotificationService,
) (Consumer, error) {
	// Will be implemented in later steps
	return &RabbitMQConsumer{
		welcomeQueue:    welcomeQueue,
		otpQueue:        otpQueue,
		notificationSvc: notificationSvc,
		prefetchCount:   prefetchCount,
	}, nil
}

func (c *RabbitMQConsumer) Start(ctx context.Context) error {
	// Will be implemented in later steps
	return nil
}

func (c *RabbitMQConsumer) Stop() error {
	// Will be implemented in later steps
	return nil
}

func (c *RabbitMQConsumer) consumeWelcomeEmails(ctx context.Context) {
	// Will be implemented in later steps
}

func (c *RabbitMQConsumer) consumeOTPEmails(ctx context.Context) {
	// Will be implemented in later steps
}

func (c *RabbitMQConsumer) setupTopology() error {
	// Will be implemented in later steps
	return nil
}
