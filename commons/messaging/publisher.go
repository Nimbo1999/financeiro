package messaging

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Publisher defines the interface for publishing messages to RabbitMQ
type Publisher interface {
	// Publish publishes a message to the specified exchange and routing key
	Publish(ctx context.Context, exchange, routingKey string, body []byte) error

	// PublishWithConfirm publishes a message and waits for broker confirmation
	PublishWithConfirm(ctx context.Context, exchange, routingKey string, body []byte) error

	// Close closes the publisher and its underlying connection
	Close() error

	// IsHealthy checks if the publisher is healthy and ready to publish
	IsHealthy() bool
}

// publisher implements the Publisher interface with auto-recovery
type publisher struct {
	connManager *ConnectionManager
	config      *Config
	logger      *slog.Logger
}

// NewPublisher creates a new publisher with the given configuration
func NewPublisher(config *Config, logger *slog.Logger) (Publisher, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if logger == nil {
		logger = slog.Default()
	}

	connManager, err := NewConnectionManager(config, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection manager: %w", err)
	}

	if err := connManager.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	p := &publisher{
		connManager: connManager,
		config:      config,
		logger:      logger,
	}

	return p, nil
}

// Publish publishes a message to the specified exchange and routing key
func (p *publisher) Publish(ctx context.Context, exchange, routingKey string, body []byte) error {
	channel, err := p.connManager.GetChannel()
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}

	publishing := amqp.Publishing{
		ContentType:  "application/json",
		Body:         body,
		DeliveryMode: amqp.Persistent, // Make messages persistent
		Timestamp:    time.Now(),
	}

	err = channel.PublishWithContext(
		ctx,
		exchange,
		routingKey,
		false, // mandatory
		false, // immediate
		publishing,
	)

	if err != nil {
		p.logger.Error("failed to publish message",
			"exchange", exchange,
			"routing_key", routingKey,
			"error", err)
		return fmt.Errorf("failed to publish message: %w", err)
	}

	p.logger.Debug("message published successfully",
		"exchange", exchange,
		"routing_key", routingKey,
		"body_size", len(body))

	return nil
}

// PublishWithConfirm publishes a message and waits for broker confirmation
func (p *publisher) PublishWithConfirm(ctx context.Context, exchange, routingKey string, body []byte) error {
	channel, err := p.connManager.GetChannel()
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}

	// Enable publisher confirms if not already enabled
	if err := channel.Confirm(false); err != nil {
		return fmt.Errorf("failed to enable publisher confirms: %w", err)
	}

	// Create confirmation channel
	confirms := channel.NotifyPublish(make(chan amqp.Confirmation, 1))

	publishing := amqp.Publishing{
		ContentType:  "application/json",
		Body:         body,
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now(),
	}

	err = channel.PublishWithContext(
		ctx,
		exchange,
		routingKey,
		false, // mandatory
		false, // immediate
		publishing,
	)

	if err != nil {
		p.logger.Error("failed to publish message with confirm",
			"exchange", exchange,
			"routing_key", routingKey,
			"error", err)
		return fmt.Errorf("failed to publish message: %w", err)
	}

	// Wait for confirmation
	select {
	case confirmation := <-confirms:
		if !confirmation.Ack {
			p.logger.Error("message not confirmed by broker",
				"exchange", exchange,
				"routing_key", routingKey)
			return fmt.Errorf("message not confirmed by broker")
		}
		p.logger.Debug("message published and confirmed successfully",
			"exchange", exchange,
			"routing_key", routingKey,
			"body_size", len(body))
		return nil
	case <-ctx.Done():
		return fmt.Errorf("timeout waiting for confirmation: %w", ctx.Err())
	}
}

// Close closes the publisher and its underlying connection
func (p *publisher) Close() error {
	p.logger.Info("closing publisher")
	return p.connManager.Close()
}

// IsHealthy checks if the publisher is healthy and ready to publish
func (p *publisher) IsHealthy() bool {
	return p.connManager.IsHealthy()
}

// PublishWithRetry publishes a message with retry logic
func (p *publisher) PublishWithRetry(ctx context.Context, exchange, routingKey string, body []byte, maxRetries int) error {
	delay := p.config.BaseReconnectDelay

	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := p.Publish(ctx, exchange, routingKey, body)
		if err == nil {
			return nil
		}

		p.logger.Warn("publish attempt failed",
			"attempt", attempt,
			"max_retries", maxRetries,
			"error", err)

		if attempt == maxRetries {
			return fmt.Errorf("failed to publish after %d attempts: %w", maxRetries, err)
		}

		// Wait before retry with exponential backoff
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled during retry: %w", ctx.Err())
		case <-time.After(delay):
		}

		delay *= 2
		if delay > p.config.MaxReconnectDelay {
			delay = p.config.MaxReconnectDelay
		}
	}

	return fmt.Errorf("failed to publish after %d attempts", maxRetries)
}
