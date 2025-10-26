package messaging

import (
	"context"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Publisher interface for publishing events

type publisher struct {
	conn         *amqp.Connection
	channel      *amqp.Channel
	exchangeName string
}

// NewPublisher creates a new RabbitMQ publisher
func NewPublisher(rabbitmqURL, exchangeName string) (Publisher, error) {
	conn, err := amqp.DialConfig(rabbitmqURL, amqp.Config{
		Properties: amqp.Table{
			"connection_name": "user_service",
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare exchange
	err = channel.ExchangeDeclare(
		exchangeName, // name
		"topic",      // type
		true,         // durable
		false,        // auto-delete
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	log.Printf("RabbitMQ publisher initialized (exchange: %s)", exchangeName)

	return &publisher{
		conn:         conn,
		channel:      channel,
		exchangeName: exchangeName,
	}, nil
}

func (p *publisher) PublishEvent(ctx context.Context, event *UserCreatedEvent) error {
	body, err := event.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize event: %w", err)
	}

	err = p.channel.PublishWithContext(
		ctx,
		p.exchangeName,
		event.GetRoutingKey(),
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	log.Printf("Published user.created event for user: %s", event.Data.UserEmail)
	return nil
}

func (p *publisher) Close() error {
	if p.channel != nil {
		if err := p.channel.Close(); err != nil {
			log.Printf("Error closing channel: %v", err)
		}
	}

	if p.conn != nil {
		if err := p.conn.Close(); err != nil {
			log.Printf("Error closing connection: %v", err)
		}
	}

	log.Println("Publisher closed")
	return nil
}
