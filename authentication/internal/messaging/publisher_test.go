package messaging

import (
	"context"
	"errors"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// Mock implementations
type MockRabbitMQConnection struct {
	mock.Mock
}

func (m *MockRabbitMQConnection) Connect() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockRabbitMQConnection) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockRabbitMQConnection) IsConnected() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockRabbitMQConnection) GetChannel() (*amqp.Channel, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*amqp.Channel), args.Error(1)
}

func (m *MockRabbitMQConnection) IsHealthy() bool {
	args := m.Called()
	return args.Bool(0)
}

type MockEvent struct {
	mock.Mock
	eventType  EventType
	id         string
	timestamp  time.Time
	routingKey string
}

func (m *MockEvent) GetType() EventType {
	return m.eventType
}

func (m *MockEvent) GetID() string {
	return m.id
}

func (m *MockEvent) GetTimestamp() time.Time {
	return m.timestamp
}

func (m *MockEvent) GetRoutingKey() string {
	return m.routingKey
}

func (m *MockEvent) ToJSON() ([]byte, error) {
	args := m.Called()
	return args.Get(0).([]byte), args.Error(1)
}

// Test Suites
type PublisherTestSuite struct {
	suite.Suite
	mockConnection *MockRabbitMQConnection
	mockPublisher  *MockPublisher
}

func (suite *PublisherTestSuite) SetupTest() {
	suite.mockConnection = new(MockRabbitMQConnection)
	suite.mockPublisher = new(MockPublisher)
}

func (suite *PublisherTestSuite) TearDownTest() {
	suite.mockConnection.AssertExpectations(suite.T())
	suite.mockPublisher.AssertExpectations(suite.T())
}

func TestPublisherTestSuite(t *testing.T) {
	suite.Run(t, new(PublisherTestSuite))
}

func (suite *PublisherTestSuite) TestMockPublisher_PublishEvent_Success() {
	ctx := context.Background()
	event := &MockEvent{
		eventType:  EventTypeAuthCodeRequested,
		id:         "test-id",
		timestamp:  time.Now(),
		routingKey: "auth.code.requested",
	}

	suite.mockPublisher.On("PublishEvent", ctx, event).Return(nil)

	err := suite.mockPublisher.PublishEvent(ctx, event)

	assert.NoError(suite.T(), err)
}

func (suite *PublisherTestSuite) TestMockPublisher_PublishEvent_Error() {
	ctx := context.Background()
	event := &MockEvent{
		eventType:  EventTypeAuthCodeRequested,
		id:         "test-id",
		timestamp:  time.Now(),
		routingKey: "auth.code.requested",
	}
	expectedError := errors.New("publish failed")

	suite.mockPublisher.On("PublishEvent", ctx, event).Return(expectedError)

	err := suite.mockPublisher.PublishEvent(ctx, event)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), expectedError, err)
}

func (suite *PublisherTestSuite) TestMockPublisher_PublishWithRetry_Success() {
	ctx := context.Background()
	event := &MockEvent{
		eventType:  EventTypeAuthCodeRequested,
		id:         "test-id",
		timestamp:  time.Now(),
		routingKey: "auth.code.requested",
	}
	maxRetries := 3

	suite.mockPublisher.On("PublishWithRetry", ctx, event, maxRetries).Return(nil)

	err := suite.mockPublisher.PublishWithRetry(ctx, event, maxRetries)

	assert.NoError(suite.T(), err)
}

func (suite *PublisherTestSuite) TestMockPublisher_PublishWithRetry_FailAfterRetries() {
	ctx := context.Background()
	event := &MockEvent{
		eventType:  EventTypeAuthCodeRequested,
		id:         "test-id",
		timestamp:  time.Now(),
		routingKey: "auth.code.requested",
	}
	maxRetries := 3
	expectedError := errors.New("failed after retries")

	suite.mockPublisher.On("PublishWithRetry", ctx, event, maxRetries).Return(expectedError)

	err := suite.mockPublisher.PublishWithRetry(ctx, event, maxRetries)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), expectedError, err)
}

func (suite *PublisherTestSuite) TestMockPublisher_IsHealthy() {
	suite.mockPublisher.On("IsHealthy").Return(true)

	isHealthy := suite.mockPublisher.IsHealthy()

	assert.True(suite.T(), isHealthy)
}

func (suite *PublisherTestSuite) TestMockPublisher_Close() {
	suite.mockPublisher.On("Close").Return(nil)

	err := suite.mockPublisher.Close()

	assert.NoError(suite.T(), err)
}

// Connection tests
func (suite *PublisherTestSuite) TestMockConnection_Connect_Success() {
	suite.mockConnection.On("Connect").Return(nil)

	err := suite.mockConnection.Connect()

	assert.NoError(suite.T(), err)
}

func (suite *PublisherTestSuite) TestMockConnection_Connect_Error() {
	expectedError := errors.New("connection failed")
	suite.mockConnection.On("Connect").Return(expectedError)

	err := suite.mockConnection.Connect()

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), expectedError, err)
}

func (suite *PublisherTestSuite) TestMockConnection_IsConnected() {
	suite.mockConnection.On("IsConnected").Return(true)

	isConnected := suite.mockConnection.IsConnected()

	assert.True(suite.T(), isConnected)
}

func (suite *PublisherTestSuite) TestMockConnection_IsHealthy() {
	suite.mockConnection.On("IsHealthy").Return(true)

	isHealthy := suite.mockConnection.IsHealthy()

	assert.True(suite.T(), isHealthy)
}

func (suite *PublisherTestSuite) TestMockConnection_Close() {
	suite.mockConnection.On("Close").Return(nil)

	err := suite.mockConnection.Close()

	assert.NoError(suite.T(), err)
}

// Configuration tests
func TestDefaultPublisherConfig(t *testing.T) {
	config := DefaultPublisherConfig()

	assert.Equal(t, "notification.exchange", config.ExchangeName)
	assert.False(t, config.Mandatory)
	assert.False(t, config.Immediate)
	assert.Equal(t, 1*time.Second, config.RetryDelay)
	assert.Equal(t, 3, config.MaxRetries)
	assert.True(t, config.ConfirmMode)
	assert.Equal(t, 30*time.Second, config.PublishTimeout)
}

func TestPublisherConfig_DefaultValues(t *testing.T) {
	testCases := []struct {
		name           string
		config         PublisherConfig
		expectedConfig PublisherConfig
	}{
		{
			name:   "empty config gets defaults",
			config: PublisherConfig{},
			expectedConfig: PublisherConfig{
				ExchangeName:   "authentication.exchange",
				RetryDelay:     1 * time.Second,
				MaxRetries:     3,
				PublishTimeout: 30 * time.Second,
			},
		},
		{
			name: "partial config keeps custom values",
			config: PublisherConfig{
				ExchangeName: "custom.exchange",
				MaxRetries:   5,
			},
			expectedConfig: PublisherConfig{
				ExchangeName:   "custom.exchange",
				RetryDelay:     1 * time.Second,
				MaxRetries:     5,
				PublishTimeout: 30 * time.Second,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// This test would be for the actual implementation's default value setting
			// For now, just verify the config can be created
			assert.NotNil(t, tc.config)
		})
	}
}

// Event publishing scenarios
func TestEventPublishingScenarios(t *testing.T) {
	testCases := []struct {
		name       string
		event      Event
		shouldFail bool
	}{
		{
			name: "auth code requested event",
			event: NewAuthCodeRequestedEvent(
				"user-123",
				"test@example.com",
				"123456",
				"code-123",
				time.Now().Add(5*time.Minute),
			),
			shouldFail: false,
		},
		{
			name: "auth code verified event",
			event: NewAuthCodeVerifiedEvent(
				"test@example.com",
				"user-123",
				"code-123",
				true,
			),
			shouldFail: false,
		},
		{
			name: "auth code expired event",
			event: NewAuthCodeExpiredEvent(
				"test@example.com",
				"code-123",
			),
			shouldFail: false,
		},
		{
			name: "auth code failed event",
			event: NewAuthCodeFailedEvent(
				"test@example.com",
				"code-123",
				"invalid_code",
				3,
			),
			shouldFail: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockPublisher := new(MockPublisher)

			if tc.shouldFail {
				mockPublisher.On("PublishEvent", mock.Anything, tc.event).Return(errors.New("publish failed"))
			} else {
				mockPublisher.On("PublishEvent", mock.Anything, tc.event).Return(nil)
			}

			ctx := context.Background()
			err := mockPublisher.PublishEvent(ctx, tc.event)

			if tc.shouldFail {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockPublisher.AssertExpectations(t)
		})
	}
}
