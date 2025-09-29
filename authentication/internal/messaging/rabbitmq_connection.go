package messaging

import (
	"fmt"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQConnection interface {
	Connect() error
	Close() error
	IsConnected() bool
	GetChannel() (*amqp.Channel, error)
	IsHealthy() bool
}

type RabbitMQConfig struct {
	URL             string
	ReconnectDelay  time.Duration
	MaxRetries      int
	HeartbeatDelay  time.Duration
	ConnectionName  string
}

type rabbitMQConnection struct {
	config     RabbitMQConfig
	connection *amqp.Connection
	mutex      sync.RWMutex
	isConnected bool
	retryCount int
}

func NewRabbitMQConnection(config RabbitMQConfig) RabbitMQConnection {
	if config.ReconnectDelay == 0 {
		config.ReconnectDelay = 5 * time.Second
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 10
	}
	if config.HeartbeatDelay == 0 {
		config.HeartbeatDelay = 60 * time.Second
	}
	if config.ConnectionName == "" {
		config.ConnectionName = "authentication-service"
	}

	return &rabbitMQConnection{
		config: config,
	}
}

func (r *rabbitMQConnection) Connect() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.isConnected && r.connection != nil && !r.connection.IsClosed() {
		return nil
	}

	// Configure connection properties
	config := amqp.Config{
		Heartbeat: r.config.HeartbeatDelay,
		Properties: amqp.Table{
			"connection_name": r.config.ConnectionName,
		},
	}

	conn, err := amqp.DialConfig(r.config.URL, config)
	if err != nil {
		r.retryCount++
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	r.connection = conn
	r.isConnected = true
	r.retryCount = 0

	// Setup connection close handler
	go r.handleConnectionClose()

	log.Printf("Connected to RabbitMQ at %s", r.maskURL(r.config.URL))
	return nil
}

func (r *rabbitMQConnection) Close() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.connection != nil && !r.connection.IsClosed() {
		err := r.connection.Close()
		if err != nil {
			return fmt.Errorf("failed to close RabbitMQ connection: %w", err)
		}
	}

	r.isConnected = false
	r.connection = nil
	log.Println("RabbitMQ connection closed")
	return nil
}

func (r *rabbitMQConnection) IsConnected() bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return r.isConnected && r.connection != nil && !r.connection.IsClosed()
}

func (r *rabbitMQConnection) GetChannel() (*amqp.Channel, error) {
	if !r.IsConnected() {
		return nil, fmt.Errorf("not connected to RabbitMQ")
	}

	r.mutex.RLock()
	conn := r.connection
	r.mutex.RUnlock()

	channel, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to create RabbitMQ channel: %w", err)
	}

	return channel, nil
}

func (r *rabbitMQConnection) IsHealthy() bool {
	if !r.IsConnected() {
		return false
	}

	// Try to create and close a channel to test connection health
	channel, err := r.GetChannel()
	if err != nil {
		return false
	}
	defer channel.Close()

	return true
}

func (r *rabbitMQConnection) handleConnectionClose() {
	r.mutex.RLock()
	conn := r.connection
	r.mutex.RUnlock()

	if conn == nil {
		return
	}

	// Wait for connection close
	closeErr := <-conn.NotifyClose(make(chan *amqp.Error))
	if closeErr != nil {
		log.Printf("RabbitMQ connection closed: %v", closeErr)
	}

	r.mutex.Lock()
	r.isConnected = false
	r.mutex.Unlock()

	// Attempt to reconnect
	r.reconnect()
}

func (r *rabbitMQConnection) reconnect() {
	for r.retryCount < r.config.MaxRetries {
		log.Printf("Attempting to reconnect to RabbitMQ (attempt %d/%d)", r.retryCount+1, r.config.MaxRetries)

		err := r.Connect()
		if err == nil {
			log.Println("Successfully reconnected to RabbitMQ")
			return
		}

		log.Printf("Reconnection failed: %v. Retrying in %v", err, r.config.ReconnectDelay)
		time.Sleep(r.config.ReconnectDelay)
	}

	log.Printf("Failed to reconnect to RabbitMQ after %d attempts", r.config.MaxRetries)
}

func (r *rabbitMQConnection) maskURL(url string) string {
	// Simple URL masking to hide credentials in logs
	// This is a basic implementation - you might want to use a more robust URL parser
	if len(url) > 15 {
		return url[:7] + "***" + url[len(url)-5:]
	}
	return "***"
}