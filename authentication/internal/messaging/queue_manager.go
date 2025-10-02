package messaging

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type QueueManager interface {
	DeclareQueues() error
	DeclareExchanges() error
	BindQueues() error
	SetupTopology() error
}

type QueueConfig struct {
	Name       string
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	NoWait     bool
	Args       amqp.Table
}

type ExchangeConfig struct {
	Name       string
	Type       string
	Durable    bool
	AutoDelete bool
	Internal   bool
	NoWait     bool
	Args       amqp.Table
}

type BindingConfig struct {
	QueueName    string
	ExchangeName string
	RoutingKey   string
	NoWait       bool
	Args         amqp.Table
}

type queueManager struct {
	connection RabbitMQConnection
	queues     []QueueConfig
	exchanges  []ExchangeConfig
	bindings   []BindingConfig
}

func NewQueueManager(connection RabbitMQConnection) QueueManager {
	return &queueManager{
		connection: connection,
		queues:     getDefaultQueues(),
		exchanges:  getDefaultExchanges(),
		bindings:   getDefaultBindings(),
	}
}

func (q *queueManager) SetupTopology() error {
	if err := q.DeclareExchanges(); err != nil {
		return fmt.Errorf("failed to declare exchanges: %w", err)
	}

	if err := q.DeclareQueues(); err != nil {
		return fmt.Errorf("failed to declare queues: %w", err)
	}

	if err := q.BindQueues(); err != nil {
		return fmt.Errorf("failed to bind queues: %w", err)
	}
	return nil
}

func (q *queueManager) DeclareExchanges() error {
	channel, err := q.connection.GetChannel()
	if err != nil {
		return err
	}
	defer channel.Close()

	for _, exchange := range q.exchanges {
		err := channel.ExchangeDeclare(
			exchange.Name,
			exchange.Type,
			exchange.Durable,
			exchange.AutoDelete,
			exchange.Internal,
			exchange.NoWait,
			exchange.Args,
		)
		if err != nil {
			return fmt.Errorf("failed to declare exchange %s: %w", exchange.Name, err)
		}
	}

	return nil
}

func (q *queueManager) DeclareQueues() error {
	channel, err := q.connection.GetChannel()
	if err != nil {
		return err
	}
	defer channel.Close()

	for _, queue := range q.queues {
		_, err := channel.QueueDeclare(
			queue.Name,
			queue.Durable,
			queue.AutoDelete,
			queue.Exclusive,
			queue.NoWait,
			queue.Args,
		)
		if err != nil {
			return fmt.Errorf("failed to declare queue %s: %w", queue.Name, err)
		}
	}

	return nil
}

func (q *queueManager) BindQueues() error {
	channel, err := q.connection.GetChannel()
	if err != nil {
		return err
	}
	defer channel.Close()

	for _, binding := range q.bindings {
		err := channel.QueueBind(
			binding.QueueName,
			binding.RoutingKey,
			binding.ExchangeName,
			binding.NoWait,
			binding.Args,
		)
		if err != nil {
			return fmt.Errorf("failed to bind queue %s to exchange %s: %w",
				binding.QueueName, binding.ExchangeName, err)
		}
	}

	return nil
}

// Default configuration for authentication service
func getDefaultExchanges() []ExchangeConfig {
	return []ExchangeConfig{
		{
			Name:       "notification.exchange",
			Type:       "topic",
			Durable:    true,
			AutoDelete: false,
			Internal:   false,
			NoWait:     false,
			Args:       nil,
		},
	}
}

func getDefaultQueues() []QueueConfig {
	return []QueueConfig{
		{
			Name:       "notification.otp",
			Durable:    true,
			AutoDelete: false,
			Exclusive:  false,
			NoWait:     false,
			Args: amqp.Table{
				"x-dead-letter-exchange":    "notification.exchange",
				"x-dead-letter-routing-key": "notification.failed",
			},
		},
		{
			Name:       "notification.dlq",
			Durable:    true,
			AutoDelete: false,
			Exclusive:  false,
			NoWait:     false,
			Args:       nil,
		},
	}
}

func getDefaultBindings() []BindingConfig {
	return []BindingConfig{
		{
			QueueName:    "notification.otp",
			ExchangeName: "notification.exchange",
			RoutingKey:   "auth.code.requested",
			NoWait:       false,
			Args:         nil,
		},
		{
			QueueName:    "notification.dlq",
			ExchangeName: "notification.exchange",
			RoutingKey:   "notification.failed",
			NoWait:       false,
			Args:         nil,
		},
	}
}
