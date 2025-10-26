package messaging

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// ConnectionManager manages RabbitMQ connections with automatic reconnection
type ConnectionManager struct {
	config *Config

	conn    *amqp.Connection
	channel *amqp.Channel
	mu      sync.RWMutex

	reconnecting atomic.Bool
	connected    atomic.Bool
	closed       atomic.Bool

	closeChan chan struct{}
	doneChan  chan struct{}

	logger *slog.Logger
}

// NewConnectionManager creates a new connection manager
func NewConnectionManager(config *Config, logger *slog.Logger) (*ConnectionManager, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	config.ApplyDefaults()
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	if logger == nil {
		logger = slog.Default()
	}

	cm := &ConnectionManager{
		config:    config,
		closeChan: make(chan struct{}),
		doneChan:  make(chan struct{}),
		logger:    logger,
	}

	return cm, nil
}

// Connect establishes a connection to RabbitMQ
func (cm *ConnectionManager) Connect() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.closed.Load() {
		return fmt.Errorf("connection manager is closed")
	}

	if cm.conn != nil && !cm.conn.IsClosed() {
		return nil // Already connected
	}

	return cm.connect()
}

// connect performs the actual connection (must be called with lock held)
func (cm *ConnectionManager) connect() error {
	cm.logger.Info("connecting to RabbitMQ", "url", maskPassword(cm.config.URL))

	conn, err := amqp.DialConfig(cm.config.URL, amqp.Config{
		Properties: amqp.Table{
			"connection_name": cm.config.ConnectionName,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to create channel: %w", err)
	}

	cm.conn = conn
	cm.channel = channel
	cm.connected.Store(true)

	cm.logger.Info("successfully connected to RabbitMQ")

	// Start monitoring connection and channel close events
	go cm.monitorConnection()

	return nil
}

// monitorConnection watches for connection and channel close events
func (cm *ConnectionManager) monitorConnection() {
	connCloseChan := make(chan *amqp.Error, 1)
	channelCloseChan := make(chan *amqp.Error, 1)

	cm.mu.RLock()
	if cm.conn != nil {
		cm.conn.NotifyClose(connCloseChan)
	}
	if cm.channel != nil {
		cm.channel.NotifyClose(channelCloseChan)
	}
	cm.mu.RUnlock()

	select {
	case err := <-connCloseChan:
		if err != nil {
			cm.logger.Error("RabbitMQ connection closed", "error", err)
			cm.handleConnectionLoss()
		}
	case err := <-channelCloseChan:
		if err != nil {
			cm.logger.Error("RabbitMQ channel closed", "error", err)
			cm.handleConnectionLoss()
		}
	case <-cm.closeChan:
		return
	}
}

// handleConnectionLoss handles connection loss and triggers reconnection
func (cm *ConnectionManager) handleConnectionLoss() {
	if cm.closed.Load() {
		return
	}

	cm.connected.Store(false)

	// Only one goroutine should attempt reconnection
	if !cm.reconnecting.CompareAndSwap(false, true) {
		return
	}
	defer cm.reconnecting.Store(false)

	cm.logger.Warn("connection lost, initiating reconnection")

	if err := cm.reconnectWithBackoff(); err != nil {
		cm.logger.Error("failed to reconnect after all retries", "error", err)
	}
}

// reconnectWithBackoff attempts to reconnect with exponential backoff
func (cm *ConnectionManager) reconnectWithBackoff() error {
	delay := cm.config.BaseReconnectDelay

	for attempt := 1; attempt <= cm.config.MaxReconnectRetries; attempt++ {
		if cm.closed.Load() {
			return fmt.Errorf("connection manager is closed")
		}

		cm.logger.Info("attempting to reconnect to RabbitMQ",
			"attempt", attempt,
			"max_retries", cm.config.MaxReconnectRetries,
			"delay", delay)

		cm.mu.Lock()
		err := cm.connect()
		cm.mu.Unlock()

		if err == nil {
			cm.logger.Info("successfully reconnected to RabbitMQ")
			return nil
		}

		cm.logger.Error("reconnection attempt failed", "error", err, "attempt", attempt)

		// Wait before next retry
		time.Sleep(delay)

		// Exponential backoff: 1s, 2s, 4s, 8s, 16s, 30s (capped)
		delay *= 2
		if delay > cm.config.MaxReconnectDelay {
			delay = cm.config.MaxReconnectDelay
		}
	}

	return fmt.Errorf("failed to reconnect after %d attempts", cm.config.MaxReconnectRetries)
}

// GetChannel returns the current channel
func (cm *ConnectionManager) GetChannel() (*amqp.Channel, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if cm.channel == nil {
		return nil, fmt.Errorf("channel is not initialized")
	}

	if cm.conn == nil || cm.conn.IsClosed() {
		return nil, fmt.Errorf("connection is closed")
	}

	return cm.channel, nil
}

// GetConnection returns the current connection
func (cm *ConnectionManager) GetConnection() (*amqp.Connection, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if cm.conn == nil || cm.conn.IsClosed() {
		return nil, fmt.Errorf("connection is closed")
	}

	return cm.conn, nil
}

// IsHealthy checks if the connection is healthy
func (cm *ConnectionManager) IsHealthy() bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if cm.conn == nil || cm.conn.IsClosed() {
		return false
	}

	if cm.channel == nil {
		return false
	}

	return cm.connected.Load()
}

// Close closes the connection and channel
func (cm *ConnectionManager) Close() error {
	if !cm.closed.CompareAndSwap(false, true) {
		return nil // Already closed
	}

	close(cm.closeChan)

	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.logger.Info("closing RabbitMQ connection")

	var err error
	if cm.channel != nil {
		if channelErr := cm.channel.Close(); channelErr != nil {
			err = fmt.Errorf("failed to close channel: %w", channelErr)
		}
		cm.channel = nil
	}

	if cm.conn != nil && !cm.conn.IsClosed() {
		if connErr := cm.conn.Close(); connErr != nil {
			if err != nil {
				err = fmt.Errorf("%w; failed to close connection: %v", err, connErr)
			} else {
				err = fmt.Errorf("failed to close connection: %w", connErr)
			}
		}
		cm.conn = nil
	}

	cm.connected.Store(false)

	return err
}

// WaitForConnection blocks until the connection is established or timeout occurs
func (cm *ConnectionManager) WaitForConnection(ctx context.Context) error {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for connection: %w", ctx.Err())
		case <-ticker.C:
			if cm.IsHealthy() {
				return nil
			}
		}
	}
}

// maskPassword masks the password in the RabbitMQ URL for logging
func maskPassword(url string) string {
	// Simple masking for security - just show the protocol
	if len(url) > 10 {
		return url[:7] + "***"
	}
	return "***"
}
