package messaging

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ConnectionManagerTestSuite struct {
	suite.Suite
	logger *slog.Logger
}

func (suite *ConnectionManagerTestSuite) SetupTest() {
	suite.logger = slog.Default()
}

func (suite *ConnectionManagerTestSuite) TestNewConnectionManager_Success() {
	config := DefaultConfig("amqp://localhost:5672")

	cm, err := NewConnectionManager(config, suite.logger)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), cm)
	assert.Equal(suite.T(), config, cm.config)
	assert.NotNil(suite.T(), cm.closeChan)
	assert.NotNil(suite.T(), cm.doneChan)
}

func (suite *ConnectionManagerTestSuite) TestNewConnectionManager_NilConfig() {
	cm, err := NewConnectionManager(nil, suite.logger)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), cm)
	assert.Contains(suite.T(), err.Error(), "config cannot be nil")
}

func (suite *ConnectionManagerTestSuite) TestNewConnectionManager_InvalidConfig() {
	config := &Config{
		URL:                 "", // Empty URL
		MaxReconnectRetries: 5,
		BaseReconnectDelay:  1,
		MaxReconnectDelay:   10,
	}

	cm, err := NewConnectionManager(config, suite.logger)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), cm)
	assert.Contains(suite.T(), err.Error(), "invalid config")
}

func (suite *ConnectionManagerTestSuite) TestNewConnectionManager_NilLogger() {
	config := DefaultConfig("amqp://localhost:5672")

	cm, err := NewConnectionManager(config, nil)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), cm)
	assert.NotNil(suite.T(), cm.logger) // Should use default logger
}

func (suite *ConnectionManagerTestSuite) TestIsHealthy_NotConnected() {
	config := DefaultConfig("amqp://localhost:5672")
	cm, err := NewConnectionManager(config, suite.logger)
	assert.NoError(suite.T(), err)

	healthy := cm.IsHealthy()

	assert.False(suite.T(), healthy)
}

func (suite *ConnectionManagerTestSuite) TestClose_NotConnected() {
	config := DefaultConfig("amqp://localhost:5672")
	cm, err := NewConnectionManager(config, suite.logger)
	assert.NoError(suite.T(), err)

	err = cm.Close()

	assert.NoError(suite.T(), err)
	assert.True(suite.T(), cm.closed.Load())
}

func (suite *ConnectionManagerTestSuite) TestClose_MultipleTimes() {
	config := DefaultConfig("amqp://localhost:5672")
	cm, err := NewConnectionManager(config, suite.logger)
	assert.NoError(suite.T(), err)

	err1 := cm.Close()
	err2 := cm.Close()

	assert.NoError(suite.T(), err1)
	assert.NoError(suite.T(), err2)
	assert.True(suite.T(), cm.closed.Load())
}

func (suite *ConnectionManagerTestSuite) TestGetChannel_NotConnected() {
	config := DefaultConfig("amqp://localhost:5672")
	cm, err := NewConnectionManager(config, suite.logger)
	assert.NoError(suite.T(), err)

	channel, err := cm.GetChannel()

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), channel)
	assert.Contains(suite.T(), err.Error(), "channel is not initialized")
}

func (suite *ConnectionManagerTestSuite) TestGetConnection_NotConnected() {
	config := DefaultConfig("amqp://localhost:5672")
	cm, err := NewConnectionManager(config, suite.logger)
	assert.NoError(suite.T(), err)

	conn, err := cm.GetConnection()

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), conn)
	assert.Contains(suite.T(), err.Error(), "connection is closed")
}

func (suite *ConnectionManagerTestSuite) TestMaskPassword() {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Standard AMQP URL",
			input:    "amqp://user:password@localhost:5672/",
			expected: "amqp://***",
		},
		{
			name:     "Short URL",
			input:    "amqp://",
			expected: "***",
		},
		{
			name:     "Empty URL",
			input:    "",
			expected: "***",
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			result := maskPassword(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestConnectionManagerTestSuite(t *testing.T) {
	suite.Run(t, new(ConnectionManagerTestSuite))
}
