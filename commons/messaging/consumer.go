package messaging

import (
	"context"
	"fmt"
	"log/slog"
	"maps"
	"sync"
	"sync/atomic"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// MessageHandler is a function that processes a message
// Return nil to acknowledge the message, or an error to nack and requeue
type MessageHandler func(delivery amqp.Delivery) error

// Consumer defines the interface for consuming messages from RabbitMQ
type Consumer interface {
	// Consume starts consuming messages from the specified queue
	Consume(ctx context.Context, queue string, handler MessageHandler) error

	// Close closes the consumer and its underlying connection
	Close() error

	// IsHealthy checks if the consumer is healthy and ready to consume
	IsHealthy() bool
}

// consumer implements the Consumer interface with auto-recovery
type consumer struct {
	connManager *ConnectionManager
	config      *Config
	logger      *slog.Logger

	consuming atomic.Bool
	closed    atomic.Bool

	mu sync.Mutex
}

// NewConsumer creates a new consumer with the given configuration
func NewConsumer(config *Config, logger *slog.Logger) (Consumer, error) {
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

	c := &consumer{
		connManager: connManager,
		config:      config,
		logger:      logger,
	}

	return c, nil
}

// Consume starts consuming messages from the specified queue
func (c *consumer) Consume(ctx context.Context, queue string, handler MessageHandler) error {
	if c.closed.Load() {
		return fmt.Errorf("consumer is closed")
	}

	if !c.consuming.CompareAndSwap(false, true) {
		return fmt.Errorf("already consuming")
	}
	defer c.consuming.Store(false)

	c.logger.Info("starting consumer", "queue", queue)

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("consumer context cancelled", "queue", queue)
			return ctx.Err()
		default:
			if err := c.consumeMessages(ctx, queue, handler); err != nil {
				if c.closed.Load() {
					return nil
				}

				c.logger.Error("consumer error, will retry",
					"queue", queue,
					"error", err)

				// Wait before retrying
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(c.config.BaseReconnectDelay):
				}
			}
		}
	}
}

// consumeMessages handles the actual message consumption
func (c *consumer) consumeMessages(ctx context.Context, queue string, handler MessageHandler) error {
	channel, err := c.connManager.GetChannel()
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}

	// Set QoS (prefetch count)
	if err := channel.Qos(c.config.PrefetchCount, 0, false); err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	// Start consuming
	deliveries, err := channel.Consume(
		queue,
		"",    // consumer tag (auto-generated)
		false, // auto-ack (we'll manually ack)
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	c.logger.Info("consumer started successfully", "queue", queue)

	// Process messages
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case delivery, ok := <-deliveries:
			if !ok {
				c.logger.Warn("delivery channel closed, will resubscribe", "queue", queue)
				return fmt.Errorf("delivery channel closed")
			}

			c.handleMessage(delivery, handler)
		}
	}
}

// handleMessage processes a single message
func (c *consumer) handleMessage(delivery amqp.Delivery, handler MessageHandler) {
	defer func() {
		if r := recover(); r != nil {
			retryCount := GetRetryCount(delivery)
			c.logger.Error("panic in message handler",
				"panic", r,
				"delivery_tag", delivery.DeliveryTag,
				"retry_count", retryCount)

			// Check if we should retry or move to DLQ
			if ShouldRetry(delivery) {
				c.logger.Warn("requeuing message after panic",
					"delivery_tag", delivery.DeliveryTag,
					"retry_count", retryCount,
					"max_retries", MaxRetries)
				c.requeueWithRetry(delivery)
			} else {
				c.logger.Error("max retries exceeded, moving to DLQ",
					"delivery_tag", delivery.DeliveryTag,
					"retry_count", retryCount,
					"max_retries", MaxRetries)
				delivery.Reject(false) // Send to DLQ (no requeue)
			}
		}
	}()

	c.logger.Debug("processing message",
		"delivery_tag", delivery.DeliveryTag,
		"routing_key", delivery.RoutingKey,
		"body_size", len(delivery.Body),
		"retry_count", GetRetryCount(delivery))

	start := time.Now()
	err := handler(delivery)
	duration := time.Since(start)

	if err != nil {
		retryCount := GetRetryCount(delivery)
		c.logger.Error("message handler error",
			"error", err,
			"delivery_tag", delivery.DeliveryTag,
			"routing_key", delivery.RoutingKey,
			"duration_ms", duration.Milliseconds(),
			"retry_count", retryCount)

		// Determine retry strategy based on retry count
		if ShouldRetry(delivery) {
			c.logger.Warn("requeuing message for retry",
				"delivery_tag", delivery.DeliveryTag,
				"retry_count", retryCount,
				"max_retries", MaxRetries)

			// Republish with incremented retry count, then ack
			c.requeueWithRetry(delivery)
		} else {
			c.logger.Error("max retries exceeded, moving to DLQ",
				"delivery_tag", delivery.DeliveryTag,
				"retry_count", retryCount,
				"max_retries", MaxRetries,
				"error", err)

			// Reject without requeue (sends to DLQ)
			if rejectErr := delivery.Reject(false); rejectErr != nil {
				c.logger.Error("failed to reject message",
					"error", rejectErr,
					"delivery_tag", delivery.DeliveryTag)
			}
		}
		return
	}

	// Acknowledge the message
	if ackErr := delivery.Ack(false); ackErr != nil {
		c.logger.Error("failed to ack message",
			"error", ackErr,
			"delivery_tag", delivery.DeliveryTag)
		return
	}

	c.logger.Debug("message processed successfully",
		"delivery_tag", delivery.DeliveryTag,
		"routing_key", delivery.RoutingKey,
		"duration_ms", duration.Milliseconds())
}

// requeueWithRetry republishes the message with an incremented retry count
func (c *consumer) requeueWithRetry(delivery amqp.Delivery) {
	// Get current retry count and increment
	currentRetry := GetRetryCount(delivery)
	newRetryCount := currentRetry + 1

	// Get channel for republishing
	channel, err := c.connManager.GetChannel()
	if err != nil {
		c.logger.Error("failed to get channel for requeue",
			"error", err,
			"delivery_tag", delivery.DeliveryTag)
		// Fall back to nack
		delivery.Nack(false, true)
		return
	}

	// Copy headers and add/update retry count
	headers := make(amqp.Table)
	if delivery.Headers != nil {
		maps.Copy(headers, delivery.Headers)
	}
	headers[RetryCountHeader] = int32(newRetryCount)

	// Republish message with updated retry count
	err = channel.Publish(
		delivery.Exchange,   // exchange
		delivery.RoutingKey, // routing key
		false,               // mandatory
		false,               // immediate
		amqp.Publishing{
			Headers:         headers,
			ContentType:     delivery.ContentType,
			ContentEncoding: delivery.ContentEncoding,
			DeliveryMode:    delivery.DeliveryMode,
			Priority:        delivery.Priority,
			CorrelationId:   delivery.CorrelationId,
			ReplyTo:         delivery.ReplyTo,
			Expiration:      delivery.Expiration,
			MessageId:       delivery.MessageId,
			Timestamp:       delivery.Timestamp,
			Type:            delivery.Type,
			UserId:          delivery.UserId,
			AppId:           delivery.AppId,
			Body:            delivery.Body,
		},
	)

	if err != nil {
		c.logger.Error("failed to republish message",
			"error", err,
			"delivery_tag", delivery.DeliveryTag,
			"retry_count", newRetryCount)
		// Fall back to nack on republish failure
		delivery.Nack(false, true)
		return
	}

	// Ack the original message since we successfully republished
	if ackErr := delivery.Ack(false); ackErr != nil {
		c.logger.Error("failed to ack message after republish",
			"error", ackErr,
			"delivery_tag", delivery.DeliveryTag)
	}

	c.logger.Debug("message republished with retry count",
		"delivery_tag", delivery.DeliveryTag,
		"retry_count", newRetryCount)
}

// Close closes the consumer and its underlying connection
func (c *consumer) Close() error {
	if !c.closed.CompareAndSwap(false, true) {
		return nil // Already closed
	}

	c.logger.Info("closing consumer")
	return c.connManager.Close()
}

// IsHealthy checks if the consumer is healthy and ready to consume
func (c *consumer) IsHealthy() bool {
	return c.connManager.IsHealthy()
}

// ConsumeWithOptions provides advanced consumption options
type ConsumeOptions struct {
	// Queue name to consume from
	Queue string

	// ConsumerTag is an optional consumer identifier
	ConsumerTag string

	// AutoAck enables automatic acknowledgement
	AutoAck bool

	// Exclusive makes this consumer the only one for the queue
	Exclusive bool

	// Args are optional arguments
	Args amqp.Table
}

// ConsumeWithOptions starts consuming with custom options
func (c *consumer) ConsumeWithOptions(ctx context.Context, opts ConsumeOptions, handler MessageHandler) error {
	if c.closed.Load() {
		return fmt.Errorf("consumer is closed")
	}

	if !c.consuming.CompareAndSwap(false, true) {
		return fmt.Errorf("already consuming")
	}
	defer c.consuming.Store(false)

	c.logger.Info("starting consumer with options",
		"queue", opts.Queue,
		"consumer_tag", opts.ConsumerTag,
		"auto_ack", opts.AutoAck)

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("consumer context cancelled", "queue", opts.Queue)
			return ctx.Err()
		default:
			if err := c.consumeMessagesWithOptions(ctx, opts, handler); err != nil {
				if c.closed.Load() {
					return nil
				}

				c.logger.Error("consumer error, will retry",
					"queue", opts.Queue,
					"error", err)

				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(c.config.BaseReconnectDelay):
				}
			}
		}
	}
}

// consumeMessagesWithOptions handles message consumption with custom options
func (c *consumer) consumeMessagesWithOptions(ctx context.Context, opts ConsumeOptions, handler MessageHandler) error {
	channel, err := c.connManager.GetChannel()
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}

	if err := channel.Qos(c.config.PrefetchCount, 0, false); err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	deliveries, err := channel.Consume(
		opts.Queue,
		opts.ConsumerTag,
		opts.AutoAck,
		opts.Exclusive,
		false, // no-local
		false, // no-wait
		opts.Args,
	)
	if err != nil {
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	c.logger.Info("consumer started successfully", "queue", opts.Queue)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case delivery, ok := <-deliveries:
			if !ok {
				c.logger.Warn("delivery channel closed, will resubscribe", "queue", opts.Queue)
				return fmt.Errorf("delivery channel closed")
			}

			if opts.AutoAck {
				// Just handle the message without acking
				go func(d amqp.Delivery) {
					defer func() {
						if r := recover(); r != nil {
							c.logger.Error("panic in message handler", "panic", r)
						}
					}()
					if err := handler(d); err != nil {
						c.logger.Error("message handler error", "error", err)
					}
				}(delivery)
			} else {
				c.handleMessage(delivery, handler)
			}
		}
	}
}
