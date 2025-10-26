package messaging

import (
	"context"
	"fmt"
	"log/slog"

	commonsmsg "github.com/nimbo1999/financeiro/commons/messaging"
)

// EventPublisher wraps the commons Publisher to work with user events
type EventPublisher struct {
	publisher    commonsmsg.Publisher
	exchangeName string
	logger       *slog.Logger
}

// EventPublisherConfig contains configuration for the event publisher
type EventPublisherConfig struct {
	RabbitMQURL  string
	ExchangeName string
	Logger       *slog.Logger
}

// NewEventPublisher creates a new event publisher using commons messaging
func NewEventPublisher(config EventPublisherConfig) (PublisherV2, error) {
	if config.ExchangeName == "" {
		config.ExchangeName = "notification.exchange"
	}

	if config.Logger == nil {
		config.Logger = slog.Default()
	}

	// Create commons messaging config with self-healing capabilities
	messagingConfig := commonsmsg.DefaultConfig(config.RabbitMQURL)
	messagingConfig.PublisherConfirms = true // Enable publisher confirms for reliability
	messagingConfig.ConnectionName = "users-service-publisher"

	// Create the underlying commons publisher
	publisher, err := commonsmsg.NewPublisher(messagingConfig, config.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create commons publisher: %w", err)
	}

	ep := &EventPublisher{
		publisher:    publisher,
		exchangeName: config.ExchangeName,
		logger:       config.Logger,
	}

	config.Logger.Info("event publisher created successfully",
		"exchange", config.ExchangeName,
		"self_healing", true)

	return ep, nil
}

// PublishEvent publishes a user created event
func (ep *EventPublisher) PublishEvent(ctx context.Context, event *UserCreatedEvent) error {
	// Serialize event to JSON
	body, err := event.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize event: %w", err)
	}

	// Get routing key from event
	routingKey := event.GetRoutingKey()

	// Publish using commons publisher with confirms
	err = ep.publisher.PublishWithConfirm(ctx, ep.exchangeName, routingKey, body)
	if err != nil {
		ep.logger.Error("failed to publish user created event",
			"user_id", event.Data.UserID,
			"user_email", event.Data.UserEmail,
			"routing_key", routingKey,
			"error", err)
		return fmt.Errorf("failed to publish event: %w", err)
	}

	ep.logger.Info("user created event published successfully",
		"user_id", event.Data.UserID,
		"user_email", event.Data.UserEmail,
		"routing_key", routingKey)

	return nil
}

// Close closes the publisher and its underlying connection
func (ep *EventPublisher) Close() error {
	ep.logger.Info("closing event publisher")
	return ep.publisher.Close()
}

// IsHealthy checks if the publisher is healthy
func (ep *EventPublisher) IsHealthy() bool {
	return ep.publisher.IsHealthy()
}
