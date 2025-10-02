package messaging

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Publisher interface for publishing events to RabbitMQ
type Publisher interface {
	PublishEvent(ctx context.Context, event Event) error
	PublishWithRetry(ctx context.Context, event Event, maxRetries int) error
	Close() error
	IsHealthy() bool
}

// PublisherConfig contains configuration for the publisher
type PublisherConfig struct {
	ExchangeName   string
	Mandatory      bool
	Immediate      bool
	RetryDelay     time.Duration
	MaxRetries     int
	ConfirmMode    bool
	PublishTimeout time.Duration
}

type publisher struct {
	connection RabbitMQConnection
	config     PublisherConfig
	channel    *amqp.Channel
	mutex      sync.RWMutex
	confirms   chan amqp.Confirmation
	isOpen     bool
}

// NewPublisher creates a new RabbitMQ publisher
func NewPublisher(connection RabbitMQConnection, queueManager QueueManager, config PublisherConfig) (Publisher, error) {
	if config.RetryDelay == 0 {
		config.RetryDelay = 1 * time.Second
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.PublishTimeout == 0 {
		config.PublishTimeout = 30 * time.Second
	}
	if config.ExchangeName == "" {
		config.ExchangeName = "notification.exchange"
	}

	p := &publisher{
		connection: connection,
		config:     config,
	}

	err := p.setupChannel()
	if err != nil {
		return nil, fmt.Errorf("failed to setup publisher channel: %w", err)
	}

	if queueManager != nil {
		err = queueManager.SetupTopology()
		if err != nil {
			p.Close()
			return nil, fmt.Errorf("failed to setup RabbitMQ topology: %w", err)
		}
	}
	log.Println("RabbitMQ publisher created and ready")

	return p, nil
}

func (p *publisher) setupChannel() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	channel, err := p.connection.GetChannel()
	if err != nil {
		return err
	}

	// Enable publisher confirms if configured
	if p.config.ConfirmMode {
		err = channel.Confirm(false)
		if err != nil {
			channel.Close()
			return fmt.Errorf("failed to enable publisher confirms: %w", err)
		}

		p.confirms = make(chan amqp.Confirmation, 1)
		channel.NotifyPublish(p.confirms)
	}

	p.channel = channel
	p.isOpen = true

	log.Println("Publisher channel setup completed")
	return nil
}

func (p *publisher) PublishEvent(ctx context.Context, event Event) error {
	return p.PublishWithRetry(ctx, event, p.config.MaxRetries)
}

func (p *publisher) PublishWithRetry(ctx context.Context, event Event, maxRetries int) error {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		err := p.publishSingle(ctx, event)
		if err == nil {
			return nil
		}

		lastErr = err
		log.Printf("Publish attempt %d failed: %v", attempt+1, err)

		// Don't wait after the last attempt
		if attempt < maxRetries {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(p.config.RetryDelay):
				// Continue to next attempt
			}

			// Try to re-establish channel if connection failed
			if !p.IsHealthy() {
				p.setupChannel()
			}
		}
	}

	return fmt.Errorf("failed to publish after %d attempts, last error: %w", maxRetries+1, lastErr)
}

func (p *publisher) publishSingle(ctx context.Context, event Event) error {
	if !p.IsHealthy() {
		return fmt.Errorf("publisher is not healthy")
	}

	// Serialize event to JSON
	body, err := event.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize event: %w", err)
	}

	// Create publication context with timeout
	pubCtx, cancel := context.WithTimeout(ctx, p.config.PublishTimeout)
	defer cancel()

	// Prepare message
	msg := amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent, // Make message persistent
		Timestamp:    event.GetTimestamp(),
		MessageId:    event.GetID(),
		Type:         string(event.GetType()),
		Body:         body,
		Headers: amqp.Table{
			"event_type":   string(event.GetType()),
			"event_id":     event.GetID(),
			"published_at": time.Now().UTC().Format(time.RFC3339),
			"publisher":    "authentication-service",
		},
	}

	p.mutex.RLock()
	channel := p.channel
	p.mutex.RUnlock()

	if channel == nil {
		return fmt.Errorf("channel is not available")
	}

	// Publish message
	err = channel.PublishWithContext(
		pubCtx,
		p.config.ExchangeName,
		event.GetRoutingKey(),
		p.config.Mandatory,
		p.config.Immediate,
		msg,
	)

	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	// Wait for confirmation if enabled
	if p.config.ConfirmMode {
		select {
		case confirm := <-p.confirms:
			if !confirm.Ack {
				return fmt.Errorf("message was not acknowledged by broker")
			}
		case <-pubCtx.Done():
			return fmt.Errorf("timeout waiting for publish confirmation")
		}
	}

	log.Printf("Published event: %s (ID: %s, RoutingKey: %s)",
		event.GetType(), event.GetID(), event.GetRoutingKey())

	return nil
}

func (p *publisher) Close() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.channel != nil && !p.channel.IsClosed() {
		err := p.channel.Close()
		if err != nil {
			log.Printf("Error closing publisher channel: %v", err)
		}
	}

	if p.confirms != nil {
		close(p.confirms)
	}

	p.isOpen = false
	log.Println("Publisher closed")
	return nil
}

func (p *publisher) IsHealthy() bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	return p.isOpen &&
		p.connection.IsConnected() &&
		p.channel != nil &&
		!p.channel.IsClosed()
}

// DefaultPublisherConfig returns a default publisher configuration
func DefaultPublisherConfig() PublisherConfig {
	return PublisherConfig{
		ExchangeName:   "notification.exchange",
		Mandatory:      false,
		Immediate:      false,
		RetryDelay:     1 * time.Second,
		MaxRetries:     3,
		ConfirmMode:    true,
		PublishTimeout: 30 * time.Second,
	}
}
