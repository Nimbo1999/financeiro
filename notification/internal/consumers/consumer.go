package consumers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/nimbo1999/financeiro/notification/internal/models"
	"github.com/nimbo1999/financeiro/notification/internal/services"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer interface {
	Start(ctx context.Context) error
	Stop() error
}

type RabbitMQConsumer struct {
	conn            *amqp.Connection
	channel         *amqp.Channel
	welcomeQueue    string
	otpQueue        string
	notificationSvc services.NotificationService
	prefetchCount   int
}

func NewRabbitMQConsumer(
	rabbitmqURL string,
	welcomeQueue string,
	otpQueue string,
	prefetchCount int,
	notificationSvc services.NotificationService,
) (Consumer, error) {
	conn, err := amqp.DialConfig(rabbitmqURL, amqp.Config{
		Properties: amqp.Table{
			"connection_name": "notification_service",
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

	consumer := &RabbitMQConsumer{
		conn:            conn,
		channel:         channel,
		welcomeQueue:    welcomeQueue,
		otpQueue:        otpQueue,
		notificationSvc: notificationSvc,
		prefetchCount:   prefetchCount,
	}

	// Setup queue topology
	if err := consumer.setupTopology(); err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to setup topology: %w", err)
	}

	// Set QoS
	if err := channel.Qos(prefetchCount, 0, false); err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to set QoS: %w", err)
	}

	log.Println("RabbitMQ consumer initialized successfully")
	return consumer, nil
}

func (c *RabbitMQConsumer) Start(ctx context.Context) error {
	log.Println("Starting RabbitMQ consumer...")

	// Start consuming from multiple queues concurrently
	go c.consumeWelcomeEmails(ctx)
	go c.consumeOTPEmails(ctx)

	log.Println("RabbitMQ consumer started")
	return nil
}

func (c *RabbitMQConsumer) Stop() error {
	log.Println("Stopping RabbitMQ consumer...")

	if c.channel != nil {
		if err := c.channel.Close(); err != nil {
			log.Printf("Error closing channel: %v", err)
		}
	}

	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			log.Printf("Error closing connection: %v", err)
		}
	}

	log.Println("RabbitMQ consumer stopped")
	return nil
}

func (c *RabbitMQConsumer) consumeWelcomeEmails(ctx context.Context) {
	log.Printf("Starting to consume from queue: %s", c.welcomeQueue)

	msgs, err := c.channel.Consume(
		c.welcomeQueue, // queue
		"",             // consumer tag
		false,          // auto-ack
		false,          // exclusive
		false,          // no-local
		false,          // no-wait
		nil,            // args
	)
	if err != nil {
		log.Printf("Failed to register welcome email consumer: %v", err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("Welcome email consumer stopped by context")
			return
		case msg, ok := <-msgs:
			if !ok {
				log.Println("Welcome email channel closed")
				return
			}

			var event models.WelcomeEmailEvent
			if err := json.Unmarshal(msg.Body, &event); err != nil {
				log.Printf("Failed to unmarshal welcome email event: %v", err)
				msg.Nack(false, false) // Don't requeue invalid messages
				continue
			}

			log.Printf("Received welcome email event for: %s", event.Data.UserEmail)

			if err := c.notificationSvc.SendWelcomeEmail(ctx, &event); err != nil {
				log.Printf("Failed to send welcome email: %v", err)
				msg.Nack(false, true) // Requeue for retry
			} else {
				log.Printf("Welcome email sent successfully to: %s", event.Data.UserEmail)
				msg.Ack(false)
			}
		}
	}
}

func (c *RabbitMQConsumer) consumeOTPEmails(ctx context.Context) {
	log.Printf("Starting to consume from queue: %s", c.otpQueue)

	msgs, err := c.channel.Consume(
		c.otpQueue, // queue
		"",         // consumer tag
		false,      // auto-ack
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // args
	)
	if err != nil {
		log.Printf("Failed to register OTP email consumer: %v", err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("OTP email consumer stopped by context")
			return
		case msg, ok := <-msgs:
			if !ok {
				log.Println("OTP email channel closed")
				return
			}

			var event models.OTPEmailEvent
			if err := json.Unmarshal(msg.Body, &event); err != nil {
				log.Printf("Failed to unmarshal OTP email event: %v", err)
				msg.Nack(false, false) // Don't requeue invalid messages
				continue
			}

			log.Printf("Received OTP email event for: %s", event.Data.UserEmail)

			if err := c.notificationSvc.SendOTPEmail(ctx, &event); err != nil {
				log.Printf("Failed to send OTP email: %v", err)
				msg.Nack(false, true) // Requeue for retry
			} else {
				log.Printf("OTP email sent successfully to: %s", event.Data.UserEmail)
				msg.Ack(false)
			}
		}
	}
}

func (c *RabbitMQConsumer) setupTopology() error {
	// Declare exchange
	err := c.channel.ExchangeDeclare(
		"notification.exchange", // name
		"topic",                 // type
		true,                    // durable
		false,                   // auto-delete
		false,                   // internal
		false,                   // no-wait
		nil,                     // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Declare welcome queue
	_, err = c.channel.QueueDeclare(
		c.welcomeQueue, // name
		true,           // durable
		false,          // auto-delete
		false,          // exclusive
		false,          // no-wait
		amqp.Table{
			"x-dead-letter-exchange":    "notification.exchange",
			"x-dead-letter-routing-key": "notification.failed",
		},
	)
	if err != nil {
		return fmt.Errorf("failed to declare welcome queue: %w", err)
	}

	// Bind welcome queue
	err = c.channel.QueueBind(
		c.welcomeQueue,          // queue name
		"user.created",          // routing key
		"notification.exchange", // exchange
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind welcome queue: %w", err)
	}

	// Declare OTP queue
	_, err = c.channel.QueueDeclare(
		c.otpQueue, // name
		true,       // durable
		false,      // auto-delete
		false,      // exclusive
		false,      // no-wait
		amqp.Table{
			"x-dead-letter-exchange":    "notification.exchange",
			"x-dead-letter-routing-key": "notification.failed",
		},
	)
	if err != nil {
		return fmt.Errorf("failed to declare otp queue: %w", err)
	}

	// Bind OTP queue
	err = c.channel.QueueBind(
		c.otpQueue,              // queue name
		"auth.code.requested",   // routing key
		"notification.exchange", // exchange
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind otp queue: %w", err)
	}

	// Declare DLQ
	_, err = c.channel.QueueDeclare(
		"notification.dlq", // name
		true,               // durable
		false,              // auto-delete
		false,              // exclusive
		false,              // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare dlq: %w", err)
	}

	// Bind DLQ
	err = c.channel.QueueBind(
		"notification.dlq",      // queue name
		"notification.failed",   // routing key
		"notification.exchange", // exchange
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind dlq: %w", err)
	}

	log.Println("RabbitMQ topology setup completed")
	return nil
}
