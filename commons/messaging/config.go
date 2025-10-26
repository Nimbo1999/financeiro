package messaging

import (
	"errors"
	"time"
)

// Config holds the configuration for RabbitMQ connections
type Config struct {
	// URL is the RabbitMQ connection string (amqp://user:pass@host:port/)
	URL string

	// MaxReconnectRetries is the maximum number of reconnection attempts
	// Default: 10
	MaxReconnectRetries int

	// BaseReconnectDelay is the initial delay between reconnection attempts
	// Default: 1 second
	BaseReconnectDelay time.Duration

	// MaxReconnectDelay is the maximum delay between reconnection attempts
	// Default: 30 seconds
	MaxReconnectDelay time.Duration

	// PublisherConfirms enables publisher confirms for reliable message delivery
	// Default: false
	PublisherConfirms bool

	// PrefetchCount sets the number of messages to prefetch for consumers
	// Default: 1
	PrefetchCount int
}

// DefaultConfig returns a Config with sensible defaults
func DefaultConfig(url string) *Config {
	return &Config{
		URL:                 url,
		MaxReconnectRetries: 10,
		BaseReconnectDelay:  1 * time.Second,
		MaxReconnectDelay:   30 * time.Second,
		PublisherConfirms:   false,
		PrefetchCount:       1,
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.URL == "" {
		return errors.New("rabbitmq URL cannot be empty")
	}

	if c.MaxReconnectRetries < 0 {
		return errors.New("max reconnect retries must be >= 0")
	}

	if c.BaseReconnectDelay <= 0 {
		return errors.New("base reconnect delay must be > 0")
	}

	if c.MaxReconnectDelay <= 0 {
		return errors.New("max reconnect delay must be > 0")
	}

	if c.MaxReconnectDelay < c.BaseReconnectDelay {
		return errors.New("max reconnect delay must be >= base reconnect delay")
	}

	if c.PrefetchCount < 0 {
		return errors.New("prefetch count must be >= 0")
	}

	return nil
}

// ApplyDefaults sets default values for any unset fields
func (c *Config) ApplyDefaults() {
	if c.MaxReconnectRetries == 0 {
		c.MaxReconnectRetries = 10
	}

	if c.BaseReconnectDelay == 0 {
		c.BaseReconnectDelay = 1 * time.Second
	}

	if c.MaxReconnectDelay == 0 {
		c.MaxReconnectDelay = 30 * time.Second
	}

	if c.PrefetchCount == 0 {
		c.PrefetchCount = 1
	}
}
