package messaging

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ConsumerTestSuite struct {
	suite.Suite
	logger *slog.Logger
}

func (suite *ConsumerTestSuite) SetupTest() {
	suite.logger = slog.Default()
}

func (suite *ConsumerTestSuite) TestNewConsumer_NilConfig() {
	consumer, err := NewConsumer(nil, suite.logger)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), consumer)
	assert.Contains(suite.T(), err.Error(), "config cannot be nil")
}

func (suite *ConsumerTestSuite) TestNewConsumer_NilLogger() {
	config := DefaultConfig("amqp://localhost:5672")

	// This will fail because it tries to connect to RabbitMQ
	consumer, err := NewConsumer(config, nil)

	// We expect an error since RabbitMQ is not running
	// But the important thing is that nil logger doesn't cause a panic
	if err != nil {
		assert.Contains(suite.T(), err.Error(), "failed to connect")
	}
	// If it somehow succeeds (unlikely), close it
	if consumer != nil {
		consumer.Close()
	}
}

func (suite *ConsumerTestSuite) TestNewConsumer_InvalidURL() {
	config := &Config{
		URL:                 "invalid://url",
		MaxReconnectRetries: 5,
		BaseReconnectDelay:  1,
		MaxReconnectDelay:   10,
		PrefetchCount:       1,
	}

	consumer, err := NewConsumer(config, suite.logger)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), consumer)
	assert.Contains(suite.T(), err.Error(), "failed to connect")
}

// Note: More comprehensive tests for Consume, message handling, etc.
// would require integration tests with a real RabbitMQ instance.
// Those tests should be added in a separate integration test suite.

func TestConsumerTestSuite(t *testing.T) {
	suite.Run(t, new(ConsumerTestSuite))
}
