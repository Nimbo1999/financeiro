package messaging

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ConfigTestSuite struct {
	suite.Suite
}

func (suite *ConfigTestSuite) TestDefaultConfig_Success() {
	url := "amqp://localhost:5672"
	config := DefaultConfig(url)

	assert.Equal(suite.T(), url, config.URL)
	assert.Equal(suite.T(), 10, config.MaxReconnectRetries)
	assert.Equal(suite.T(), 1*time.Second, config.BaseReconnectDelay)
	assert.Equal(suite.T(), 30*time.Second, config.MaxReconnectDelay)
	assert.Equal(suite.T(), false, config.PublisherConfirms)
	assert.Equal(suite.T(), 1, config.PrefetchCount)
}

func (suite *ConfigTestSuite) TestValidate_Success() {
	config := &Config{
		URL:                 "amqp://localhost:5672",
		MaxReconnectRetries: 5,
		BaseReconnectDelay:  1 * time.Second,
		MaxReconnectDelay:   10 * time.Second,
		PublisherConfirms:   true,
		PrefetchCount:       10,
	}

	err := config.Validate()
	assert.NoError(suite.T(), err)
}

func (suite *ConfigTestSuite) TestValidate_EmptyURL() {
	config := &Config{
		URL:                 "",
		MaxReconnectRetries: 5,
		BaseReconnectDelay:  1 * time.Second,
		MaxReconnectDelay:   10 * time.Second,
	}

	err := config.Validate()
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "URL cannot be empty")
}

func (suite *ConfigTestSuite) TestValidate_NegativeMaxRetries() {
	config := &Config{
		URL:                 "amqp://localhost:5672",
		MaxReconnectRetries: -1,
		BaseReconnectDelay:  1 * time.Second,
		MaxReconnectDelay:   10 * time.Second,
	}

	err := config.Validate()
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "max reconnect retries must be >= 0")
}

func (suite *ConfigTestSuite) TestValidate_InvalidBaseDelay() {
	config := &Config{
		URL:                 "amqp://localhost:5672",
		MaxReconnectRetries: 5,
		BaseReconnectDelay:  0,
		MaxReconnectDelay:   10 * time.Second,
	}

	err := config.Validate()
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "base reconnect delay must be > 0")
}

func (suite *ConfigTestSuite) TestValidate_InvalidMaxDelay() {
	config := &Config{
		URL:                 "amqp://localhost:5672",
		MaxReconnectRetries: 5,
		BaseReconnectDelay:  1 * time.Second,
		MaxReconnectDelay:   0,
	}

	err := config.Validate()
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "max reconnect delay must be > 0")
}

func (suite *ConfigTestSuite) TestValidate_MaxDelayLessThanBaseDelay() {
	config := &Config{
		URL:                 "amqp://localhost:5672",
		MaxReconnectRetries: 5,
		BaseReconnectDelay:  10 * time.Second,
		MaxReconnectDelay:   1 * time.Second,
	}

	err := config.Validate()
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "max reconnect delay must be >= base reconnect delay")
}

func (suite *ConfigTestSuite) TestValidate_NegativePrefetchCount() {
	config := &Config{
		URL:                 "amqp://localhost:5672",
		MaxReconnectRetries: 5,
		BaseReconnectDelay:  1 * time.Second,
		MaxReconnectDelay:   10 * time.Second,
		PrefetchCount:       -1,
	}

	err := config.Validate()
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "prefetch count must be >= 0")
}

func (suite *ConfigTestSuite) TestApplyDefaults_AllFieldsZero() {
	config := &Config{
		URL: "amqp://localhost:5672",
	}

	config.ApplyDefaults()

	assert.Equal(suite.T(), 10, config.MaxReconnectRetries)
	assert.Equal(suite.T(), 1*time.Second, config.BaseReconnectDelay)
	assert.Equal(suite.T(), 30*time.Second, config.MaxReconnectDelay)
	assert.Equal(suite.T(), 1, config.PrefetchCount)
}

func (suite *ConfigTestSuite) TestApplyDefaults_SomeFieldsSet() {
	config := &Config{
		URL:                 "amqp://localhost:5672",
		MaxReconnectRetries: 5,
		BaseReconnectDelay:  2 * time.Second,
	}

	config.ApplyDefaults()

	// Should keep custom values
	assert.Equal(suite.T(), 5, config.MaxReconnectRetries)
	assert.Equal(suite.T(), 2*time.Second, config.BaseReconnectDelay)

	// Should apply defaults for unset values
	assert.Equal(suite.T(), 30*time.Second, config.MaxReconnectDelay)
	assert.Equal(suite.T(), 1, config.PrefetchCount)
}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}
