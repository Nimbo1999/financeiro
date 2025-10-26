package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	commonsmsg "github.com/nimbo1999/financeiro/commons/messaging"
	"github.com/nimbo1999/financeiro/notification/internal/services"
	amqp "github.com/rabbitmq/amqp091-go"
)

// EventConsumer wraps the commons Consumer to handle authentication events
type EventConsumer struct {
	consumer        commonsmsg.Consumer
	notificationSvc services.NotificationService
	logger          *slog.Logger
	ctx             context.Context
	cancelFunc      context.CancelFunc
}

// EventConsumerConfig contains configuration for the event consumer
type EventConsumerConfig struct {
	RabbitMQURL     string
	QueueName       string
	NotificationSvc services.NotificationService
	Logger          *slog.Logger
}

// NewEventConsumer creates a new event consumer using commons messaging
func NewEventConsumer(config EventConsumerConfig) (*EventConsumer, error) {
	if config.Logger == nil {
		config.Logger = slog.Default()
	}

	if config.QueueName == "" {
		config.QueueName = "notification.otp"
	}

	// Create commons messaging config with self-healing capabilities
	messagingConfig := commonsmsg.DefaultConfig(config.RabbitMQURL)
	messagingConfig.PrefetchCount = 10 // Process up to 10 messages concurrently
	messagingConfig.ConnectionName = "notification-consumer"

	// Create the underlying commons consumer
	consumer, err := commonsmsg.NewConsumer(messagingConfig, config.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create commons consumer: %w", err)
	}

	ec := &EventConsumer{
		consumer:        consumer,
		notificationSvc: config.NotificationSvc,
		logger:          config.Logger,
	}

	config.Logger.Info("event consumer created successfully",
		"queue", config.QueueName,
		"self_healing", true,
		"prefetch_count", messagingConfig.PrefetchCount)

	return ec, nil
}

// Start starts consuming messages from the queue
func (ec *EventConsumer) Start(ctx context.Context, queueName string) error {
	ec.ctx, ec.cancelFunc = context.WithCancel(ctx)

	ec.logger.Info("starting event consumer", "queue", queueName)

	// Define the message handler
	handler := func(delivery amqp.Delivery) error {
		return ec.handleMessage(delivery)
	}

	// Start consuming (blocks until context is cancelled)
	return ec.consumer.Consume(ec.ctx, queueName, handler)
}

// handleMessage processes a single message
func (ec *EventConsumer) handleMessage(delivery amqp.Delivery) error {
	ec.logger.Debug("received message",
		"routing_key", delivery.RoutingKey,
		"message_id", delivery.MessageId,
		"body_size", len(delivery.Body))

	// Check routing key to determine event type
	switch delivery.RoutingKey {
	case "auth.code.requested":
		return ec.handleOTPEmailEvent(delivery.Body)
	case "user.created":
		return ec.handleWelcomeEmailEvent(delivery.Body)
	default:
		ec.logger.Warn("unknown routing key",
			"routing_key", delivery.RoutingKey)
		// Return nil to acknowledge unknown messages (don't requeue)
		return nil
	}
}

// handleOTPEmailEvent processes OTP email events
func (ec *EventConsumer) handleOTPEmailEvent(body []byte) error {
	var event OTPEmailEvent
	if err := json.Unmarshal(body, &event); err != nil {
		ec.logger.Error("failed to unmarshal OTP email event", "error", err)
		// Don't requeue invalid JSON
		return nil
	}

	ec.logger.Info("processing OTP email event",
		"user_email", event.Data.UserEmail,
		"event_id", event.ID)

	// Convert to models.OTPEmailEvent
	modelEvent := event.ToModel()

	if err := ec.notificationSvc.SendOTPEmail(ec.ctx, modelEvent); err != nil {
		ec.logger.Error("failed to send OTP email",
			"user_email", event.Data.UserEmail,
			"error", err)
		// Return error to nack and requeue
		return fmt.Errorf("failed to send OTP email: %w", err)
	}

	ec.logger.Info("OTP email sent successfully",
		"user_email", event.Data.UserEmail)

	return nil
}

// handleWelcomeEmailEvent processes welcome email events
func (ec *EventConsumer) handleWelcomeEmailEvent(body []byte) error {
	var event WelcomeEmailEvent
	if err := json.Unmarshal(body, &event); err != nil {
		ec.logger.Error("failed to unmarshal welcome email event", "error", err)
		// Don't requeue invalid JSON
		return nil
	}

	ec.logger.Info("processing welcome email event",
		"user_email", event.Data.UserEmail,
		"event_id", event.ID)

	// Convert to models.WelcomeEmailEvent
	modelEvent := event.ToModel()

	if err := ec.notificationSvc.SendWelcomeEmail(ec.ctx, modelEvent); err != nil {
		ec.logger.Error("failed to send welcome email",
			"user_email", event.Data.UserEmail,
			"error", err)
		// Return error to nack and requeue
		return fmt.Errorf("failed to send welcome email: %w", err)
	}

	ec.logger.Info("welcome email sent successfully",
		"user_email", event.Data.UserEmail)

	return nil
}

// Close closes the consumer
func (ec *EventConsumer) Close() error {
	ec.logger.Info("closing event consumer")

	if ec.cancelFunc != nil {
		ec.cancelFunc()
	}

	return ec.consumer.Close()
}

// IsHealthy checks if the consumer is healthy
func (ec *EventConsumer) IsHealthy() bool {
	return ec.consumer.IsHealthy()
}
