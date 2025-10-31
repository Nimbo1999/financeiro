package messaging

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

// QueueConfig defines configuration for declaring queues
type QueueConfig struct {
	Name       string
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	NoWait     bool
	Args       amqp.Table
}

// QueueDeclarer provides methods for declaring queues and DLQs
type QueueDeclarer interface {
	// DeclareQueue declares a standard queue
	DeclareQueue(config QueueConfig) error

	// DeclareQueueWithDLQ declares a queue with dead letter queue setup
	DeclareQueueWithDLQ(queueName, dlqExchange, dlqRoutingKey string) error
}

type queueDeclarer struct {
	connManager *ConnectionManager
}

// NewQueueDeclarer creates a new queue declarer
func NewQueueDeclarer(connManager *ConnectionManager) QueueDeclarer {
	return &queueDeclarer{
		connManager: connManager,
	}
}

// DeclareQueue declares a standard queue
func (qd *queueDeclarer) DeclareQueue(config QueueConfig) error {
	channel, err := qd.connManager.GetChannel()
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}

	_, err = channel.QueueDeclare(
		config.Name,
		config.Durable,
		config.AutoDelete,
		config.Exclusive,
		config.NoWait,
		config.Args,
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue %s: %w", config.Name, err)
	}

	return nil
}

// DeclareQueueWithDLQ declares a queue with DLQ configuration
// Main queue will route failed messages to DLQ after rejection
func (qd *queueDeclarer) DeclareQueueWithDLQ(queueName, dlqExchange, dlqRoutingKey string) error {
	channel, err := qd.connManager.GetChannel()
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}

	// Declare DLQ exchange (topic exchange for flexibility)
	err = channel.ExchangeDeclare(
		dlqExchange, // name
		"topic",     // type
		true,        // durable
		false,       // auto-delete
		false,       // internal
		false,       // no-wait
		nil,         // args
	)
	if err != nil {
		return fmt.Errorf("failed to declare DLQ exchange: %w", err)
	}

	// Declare DLQ queue
	dlqName := queueName + ".dlq"
	_, err = channel.QueueDeclare(
		dlqName, // name
		true,    // durable
		false,   // auto-delete
		false,   // exclusive
		false,   // no-wait
		nil,     // args
	)
	if err != nil {
		return fmt.Errorf("failed to declare DLQ: %w", err)
	}

	// Bind DLQ to DLQ exchange
	err = channel.QueueBind(
		dlqName,       // queue
		dlqRoutingKey, // routing key
		dlqExchange,   // exchange
		false,         // no-wait
		nil,           // args
	)
	if err != nil {
		return fmt.Errorf("failed to bind DLQ: %w", err)
	}

	// Declare main queue with DLQ arguments
	_, err = channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // auto-delete
		false,     // exclusive
		false,     // no-wait
		amqp.Table{
			"x-dead-letter-exchange":    dlqExchange,
			"x-dead-letter-routing-key": dlqRoutingKey,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to declare main queue with DLQ: %w", err)
	}

	return nil
}
