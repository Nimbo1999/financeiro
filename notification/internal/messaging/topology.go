package messaging

import (
	"fmt"
	"log/slog"

	commonsmessaging "github.com/nimbo1999/financeiro/commons/messaging"
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

// SetupTopology declares all exchanges, queues, and bindings using commons QueueDeclarer
func (tm *TopologyManager) SetupTopology() error {
	tm.logger.Info("setting up RabbitMQ topology")

	// Create connection manager from commons
	config := commonsmessaging.DefaultConfig(tm.url)
	connManager, err := commonsmessaging.NewConnectionManager(config, tm.logger)
	if err != nil {
		return fmt.Errorf("failed to create connection manager: %w", err)
	}

	if err := connManager.Connect(); err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	defer connManager.Close()

	// Get channel for exchange declaration
	channel, err := connManager.GetChannel()
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}

	// Declare main exchange
	if err := tm.declareExchange(channel); err != nil {
		return err
	}

	// Use QueueDeclarer from commons for DLQ setup
	queueDeclarer := commonsmessaging.NewQueueDeclarer(connManager)

	// Declare queues with DLQ using commons helper
	if err := tm.declareQueuesWithDLQ(queueDeclarer); err != nil {
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

func (tm *TopologyManager) declareQueuesWithDLQ(queueDeclarer commonsmessaging.QueueDeclarer) error {
	// Define queues that need DLQ support
	queuesWithDLQ := []struct {
		name          string
		dlqExchange   string
		dlqRoutingKey string
	}{
		{
			name:          "notification.welcome",
			dlqExchange:   "notification.dlq.exchange",
			dlqRoutingKey: "notification.welcome.failed",
		},
		{
			name:          "notification.otp",
			dlqExchange:   "notification.dlq.exchange",
			dlqRoutingKey: "notification.otp.failed",
		},
	}

	// Declare each queue with DLQ using commons helper
	for _, q := range queuesWithDLQ {
		if err := queueDeclarer.DeclareQueueWithDLQ(q.name, q.dlqExchange, q.dlqRoutingKey); err != nil {
			return fmt.Errorf("failed to declare queue %s with DLQ: %w", q.name, err)
		}
		tm.logger.Debug("queue with DLQ declared", "name", q.name, "dlq", q.name+".dlq")
	}

	return nil
}

func (tm *TopologyManager) bindQueues(channel *amqp.Channel) error {
	bindings := []struct {
		queue      string
		routingKey string
		exchange   string
	}{
		{
			queue:      "notification.welcome",
			routingKey: "user.created",
			exchange:   "notification.exchange",
		},
		{
			queue:      "notification.otp",
			routingKey: "auth.code.requested",
			exchange:   "notification.exchange",
		},
		// Bind DLQ queues to their routing keys
		{
			queue:      "notification.welcome.dlq",
			routingKey: "notification.welcome.failed",
			exchange:   "notification.dlq.exchange",
		},
		{
			queue:      "notification.otp.dlq",
			routingKey: "notification.otp.failed",
			exchange:   "notification.dlq.exchange",
		},
	}

	for _, b := range bindings {
		err := channel.QueueBind(
			b.queue,      // queue name
			b.routingKey, // routing key
			b.exchange,   // exchange
			false,        // no-wait
			nil,          // arguments
		)
		if err != nil {
			return fmt.Errorf("failed to bind queue %s with routing key %s: %w",
				b.queue, b.routingKey, err)
		}

		tm.logger.Debug("queue bound",
			"queue", b.queue,
			"routing_key", b.routingKey,
			"exchange", b.exchange)
	}

	return nil
}
