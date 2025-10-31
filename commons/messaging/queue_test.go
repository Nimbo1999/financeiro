package messaging

import (
	"log/slog"
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type QueueDeclarerTestSuite struct {
	suite.Suite
	config  *Config
	logger  *slog.Logger
	manager *ConnectionManager
}

func (suite *QueueDeclarerTestSuite) SetupTest() {
	suite.logger = slog.Default()
	suite.config = DefaultConfig("amqp://guest:guest@localhost:5672/")
}

func (suite *QueueDeclarerTestSuite) TearDownTest() {
	if suite.manager != nil {
		suite.manager.Close()
		suite.manager = nil
	}
}

func (suite *QueueDeclarerTestSuite) TestNewQueueDeclarer_Success() {
	manager, err := NewConnectionManager(suite.config, suite.logger)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), manager)

	declarer := NewQueueDeclarer(manager)
	assert.NotNil(suite.T(), declarer)
	assert.IsType(suite.T(), &queueDeclarer{}, declarer)
}

func (suite *QueueDeclarerTestSuite) TestDeclareQueue_NoConnection() {
	manager, err := NewConnectionManager(suite.config, suite.logger)
	assert.NoError(suite.T(), err)
	suite.manager = manager

	declarer := NewQueueDeclarer(manager)

	// Try to declare queue without connecting
	config := QueueConfig{
		Name:       "test.queue",
		Durable:    true,
		AutoDelete: false,
		Exclusive:  false,
		NoWait:     false,
		Args:       nil,
	}

	err = declarer.DeclareQueue(config)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "channel is not initialized")
}

func (suite *QueueDeclarerTestSuite) TestDeclareQueueWithDLQ_NoConnection() {
	manager, err := NewConnectionManager(suite.config, suite.logger)
	assert.NoError(suite.T(), err)
	suite.manager = manager

	declarer := NewQueueDeclarer(manager)

	// Try to declare queue with DLQ without connecting
	err = declarer.DeclareQueueWithDLQ(
		"test.queue",
		"test.dlq.exchange",
		"test.failed",
	)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "channel is not initialized")
}

func (suite *QueueDeclarerTestSuite) TestQueueConfig_AllFields() {
	// Test that QueueConfig struct accepts all expected fields
	config := QueueConfig{
		Name:       "test.queue",
		Durable:    true,
		AutoDelete: false,
		Exclusive:  false,
		NoWait:     false,
		Args: amqp.Table{
			"x-message-ttl": int32(60000),
		},
	}

	assert.Equal(suite.T(), "test.queue", config.Name)
	assert.True(suite.T(), config.Durable)
	assert.False(suite.T(), config.AutoDelete)
	assert.False(suite.T(), config.Exclusive)
	assert.False(suite.T(), config.NoWait)
	assert.NotNil(suite.T(), config.Args)
	assert.Equal(suite.T(), int32(60000), config.Args["x-message-ttl"])
}

func (suite *QueueDeclarerTestSuite) TestDeclareQueue_EmptyQueueName() {
	manager, err := NewConnectionManager(suite.config, suite.logger)
	assert.NoError(suite.T(), err)
	suite.manager = manager

	// Connect to RabbitMQ (this will fail if RabbitMQ is not running, which is expected)
	err = manager.Connect()
	if err != nil {
		suite.T().Skip("RabbitMQ not available for integration test")
		return
	}

	declarer := NewQueueDeclarer(manager)

	// Try to declare queue with empty name
	config := QueueConfig{
		Name:       "",
		Durable:    true,
		AutoDelete: false,
		Exclusive:  false,
		NoWait:     false,
		Args:       nil,
	}

	err = declarer.DeclareQueue(config)
	// RabbitMQ will generate a queue name for empty string, so this should succeed
	// if RabbitMQ is available
	if err != nil {
		// This is okay for unit tests without RabbitMQ
		assert.Contains(suite.T(), err.Error(), "failed to declare queue")
	}
}

func (suite *QueueDeclarerTestSuite) TestDeclareQueueWithDLQ_EmptyParameters() {
	manager, err := NewConnectionManager(suite.config, suite.logger)
	assert.NoError(suite.T(), err)
	suite.manager = manager

	// Connect to RabbitMQ
	err = manager.Connect()
	if err != nil {
		suite.T().Skip("RabbitMQ not available for integration test")
		return
	}

	declarer := NewQueueDeclarer(manager)

	// Try to declare queue with DLQ with empty parameters
	err = declarer.DeclareQueueWithDLQ("", "", "")
	// This should fail with RabbitMQ error about empty exchange name
	assert.Error(suite.T(), err)
}

func TestQueueDeclarerTestSuite(t *testing.T) {
	suite.Run(t, new(QueueDeclarerTestSuite))
}
