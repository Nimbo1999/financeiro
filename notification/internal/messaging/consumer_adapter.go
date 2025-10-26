package messaging

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/nimbo1999/financeiro/notification/internal/services"
)

// ConsumerAdapter adapts the new EventConsumer to the old Consumer interface
type ConsumerAdapter struct {
	otpConsumer     *EventConsumer
	welcomeConsumer *EventConsumer
	logger          *slog.Logger
	otpQueue        string
	welcomeQueue    string
}

// NewConsumerAdapter creates a new consumer adapter
func NewConsumerAdapter(
	rabbitmqURL string,
	welcomeQueue string,
	otpQueue string,
	notificationSvc services.NotificationService,
) (*ConsumerAdapter, error) {
	logger := slog.Default()

	// Create OTP consumer
	otpConsumer, err := NewEventConsumer(EventConsumerConfig{
		RabbitMQURL:     rabbitmqURL,
		QueueName:       otpQueue,
		NotificationSvc: notificationSvc,
		Logger:          logger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create OTP consumer: %w", err)
	}

	// Create Welcome consumer
	welcomeConsumer, err := NewEventConsumer(EventConsumerConfig{
		RabbitMQURL:     rabbitmqURL,
		QueueName:       welcomeQueue,
		NotificationSvc: notificationSvc,
		Logger:          logger,
	})
	if err != nil {
		otpConsumer.Close()
		return nil, fmt.Errorf("failed to create welcome consumer: %w", err)
	}

	return &ConsumerAdapter{
		otpConsumer:     otpConsumer,
		welcomeConsumer: welcomeConsumer,
		logger:          logger,
		otpQueue:        otpQueue,
		welcomeQueue:    welcomeQueue,
	}, nil
}

// Start starts consuming messages from both queues
func (ca *ConsumerAdapter) Start(ctx context.Context) error {
	ca.logger.Info("starting consumer adapter")

	// Start OTP consumer in goroutine
	go func() {
		if err := ca.otpConsumer.Start(ctx, ca.otpQueue); err != nil {
			ca.logger.Error("OTP consumer error", "error", err)
		}
	}()

	// Start Welcome consumer in goroutine
	go func() {
		if err := ca.welcomeConsumer.Start(ctx, ca.welcomeQueue); err != nil {
			ca.logger.Error("welcome consumer error", "error", err)
		}
	}()

	ca.logger.Info("consumer adapter started successfully")
	return nil
}

// Stop stops both consumers
func (ca *ConsumerAdapter) Stop() error {
	ca.logger.Info("stopping consumer adapter")

	var otpErr, welcomeErr error

	if err := ca.otpConsumer.Close(); err != nil {
		ca.logger.Error("failed to close OTP consumer", "error", err)
		otpErr = err
	}

	if err := ca.welcomeConsumer.Close(); err != nil {
		ca.logger.Error("failed to close welcome consumer", "error", err)
		welcomeErr = err
	}

	if otpErr != nil {
		return otpErr
	}
	if welcomeErr != nil {
		return welcomeErr
	}

	ca.logger.Info("consumer adapter stopped successfully")
	return nil
}

// IsHealthy checks if both consumers are healthy
func (ca *ConsumerAdapter) IsHealthy() bool {
	return ca.otpConsumer.IsHealthy() && ca.welcomeConsumer.IsHealthy()
}
