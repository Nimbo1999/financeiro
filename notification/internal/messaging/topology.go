package messaging

import (
	"fmt"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
)

// TopologyManager handles RabbitMQ topology setup
type TopologyManager struct {
	url    string
	logger *slog.Logger
}

// NewTopologyManager creates a new topology manager
func NewTopologyManager(url string, logger *slog.Logger) *TopologyManager {
	if logger == nil {
		logger = slog.Default()
	}

	return &TopologyManager{
		url:    url,
		logger: logger,
	}
}

// SetupTopology declares all exchanges, queues, and bindings
func (tm *TopologyManager) SetupTopology() error {
	tm.logger.Info("setting up RabbitMQ topology")

	// Connect to RabbitMQ
	conn, err := amqp.Dial(tm.url)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	defer conn.Close()

	// Create channel
	channel, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to create channel: %w", err)
	}
	defer channel.Close()

	// Declare exchange
	if err := tm.declareExchange(channel); err != nil {
		return err
	}

	// Declare queues
	if err := tm.declareQueues(channel); err != nil {
		return err
	}

	// Bind queues
	if err := tm.bindQueues(channel); err != nil {
		return err
	}

	tm.logger.Info("RabbitMQ topology setup completed successfully")
	return nil
}

func (tm *TopologyManager) declareExchange(channel *amqp.Channel) error {
	err := channel.ExchangeDeclare(
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

	tm.logger.Debug("exchange declared", "name", "notification.exchange")
	return nil
}

func (tm *TopologyManager) declareQueues(channel *amqp.Channel) error {
	queues := []struct {
		name string
		args amqp.Table
	}{
		{
			name: "notification.welcome",
			args: amqp.Table{
				"x-dead-letter-exchange":    "notification.exchange",
				"x-dead-letter-routing-key": "notification.failed",
			},
		},
		{
			name: "notification.otp",
			args: amqp.Table{
				"x-dead-letter-exchange":    "notification.exchange",
				"x-dead-letter-routing-key": "notification.failed",
			},
		},
		{
			name: "notification.dlq",
			args: nil,
		},
	}

	for _, q := range queues {
		_, err := channel.QueueDeclare(
			q.name, // name
			true,   // durable
			false,  // auto-delete
			false,  // exclusive
			false,  // no-wait
			q.args, // arguments
		)
		if err != nil {
			return fmt.Errorf("failed to declare queue %s: %w", q.name, err)
		}

		tm.logger.Debug("queue declared", "name", q.name)
	}

	return nil
}

func (tm *TopologyManager) bindQueues(channel *amqp.Channel) error {
	bindings := []struct {
		queue      string
		routingKey string
	}{
		{
			queue:      "notification.welcome",
			routingKey: "user.created",
		},
		{
			queue:      "notification.otp",
			routingKey: "auth.code.requested",
		},
		{
			queue:      "notification.dlq",
			routingKey: "notification.failed",
		},
	}

	for _, b := range bindings {
		err := channel.QueueBind(
			b.queue,                 // queue name
			b.routingKey,            // routing key
			"notification.exchange", // exchange
			false,                   // no-wait
			nil,                     // arguments
		)
		if err != nil {
			return fmt.Errorf("failed to bind queue %s with routing key %s: %w",
				b.queue, b.routingKey, err)
		}

		tm.logger.Debug("queue bound",
			"queue", b.queue,
			"routing_key", b.routingKey)
	}

	return nil
}
