package messaging

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type PublisherTestSuite struct {
	suite.Suite
	logger *slog.Logger
}

func (suite *PublisherTestSuite) SetupTest() {
	suite.logger = slog.Default()
}

func (suite *PublisherTestSuite) TestNewPublisher_NilConfig() {
	publisher, err := NewPublisher(nil, suite.logger)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), publisher)
	assert.Contains(suite.T(), err.Error(), "config cannot be nil")
}

func (suite *PublisherTestSuite) TestNewPublisher_NilLogger() {
	config := DefaultConfig("amqp://localhost:5672")

	// This will fail because it tries to connect to RabbitMQ
	publisher, err := NewPublisher(config, nil)

	// We expect an error since RabbitMQ is not running
	// But the important thing is that nil logger doesn't cause a panic
	if err != nil {
		assert.Contains(suite.T(), err.Error(), "failed to connect")
	}
	// If it somehow succeeds (unlikely), close it
	if publisher != nil {
		publisher.Close()
	}
}

func (suite *PublisherTestSuite) TestNewPublisher_InvalidURL() {
	config := &Config{
		URL:                 "invalid://url",
		MaxReconnectRetries: 5,
		BaseReconnectDelay:  1,
		MaxReconnectDelay:   10,
		PrefetchCount:       1,
	}

	publisher, err := NewPublisher(config, suite.logger)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), publisher)
	assert.Contains(suite.T(), err.Error(), "failed to connect")
}

// Note: More comprehensive tests for Publish, PublishWithConfirm, etc.
// would require integration tests with a real RabbitMQ instance.
// Those tests should be added in a separate integration test suite.

func TestPublisherTestSuite(t *testing.T) {
	suite.Run(t, new(PublisherTestSuite))
}
