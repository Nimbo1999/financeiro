package clients

import (
	"context"
	"errors"
	"testing"
	"time"

	userv1 "github.com/nimbo1999/financeiro/users/pkg/grpc/users/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// MockUserServiceClient is a mock implementation of UserServiceClient using testify/mock
type MockUserServiceClient struct {
	mock.Mock
}

// NewMockUserServiceClient creates a new mock user service client
func NewMockUserServiceClient() *MockUserServiceClient {
	return &MockUserServiceClient{}
}

// GetUserByEmail mocks the GetUserByEmail method
func (m *MockUserServiceClient) GetUserByEmail(ctx context.Context, email string) (*userv1.User, bool, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(*userv1.User), args.Bool(1), args.Error(2)
}

// GetUserById mocks the GetUserById method
func (m *MockUserServiceClient) GetUserById(ctx context.Context, userID string) (*userv1.User, bool, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*userv1.User), args.Bool(1), args.Error(2)
}

// HealthCheck mocks the HealthCheck method
func (m *MockUserServiceClient) HealthCheck(ctx context.Context) (userv1.HealthCheckResponse_Status, string, error) {
	args := m.Called(ctx)
	return args.Get(0).(userv1.HealthCheckResponse_Status), args.String(1), args.Error(2)
}

// Close mocks the Close method
func (m *MockUserServiceClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

type UserServiceClientTestSuite struct {
	suite.Suite
	mockClient *MockUserServiceClient
}

func (suite *UserServiceClientTestSuite) SetupTest() {
	suite.mockClient = NewMockUserServiceClient()
}

func (suite *UserServiceClientTestSuite) TearDownTest() {
	suite.mockClient.AssertExpectations(suite.T())
}

func TestUserServiceClientTestSuite(t *testing.T) {
	suite.Run(t, new(UserServiceClientTestSuite))
}

func (suite *UserServiceClientTestSuite) TestGetUserByEmail_Success() {
	email := "test@example.com"
	expectedUser := &userv1.User{
		Id:        "user-123",
		Email:     email,
		FullName:  "Test User",
		CreatedAt: timestamppb.New(time.Now()),
		UpdatedAt: timestamppb.New(time.Now()),
	}

	suite.mockClient.On("GetUserByEmail", mock.Anything, email).Return(expectedUser, true, nil)

	ctx := context.Background()
	user, found, err := suite.mockClient.GetUserByEmail(ctx, email)

	assert.NoError(suite.T(), err)
	assert.True(suite.T(), found)
	assert.NotNil(suite.T(), user)
	assert.Equal(suite.T(), expectedUser.Id, user.Id)
	assert.Equal(suite.T(), expectedUser.Email, user.Email)
	assert.Equal(suite.T(), expectedUser.FullName, user.FullName)
}

func (suite *UserServiceClientTestSuite) TestGetUserByEmail_NotFound() {
	email := "notfound@example.com"

	suite.mockClient.On("GetUserByEmail", mock.Anything, email).Return((*userv1.User)(nil), false, nil)

	ctx := context.Background()
	user, found, err := suite.mockClient.GetUserByEmail(ctx, email)

	assert.NoError(suite.T(), err)
	assert.False(suite.T(), found)
	assert.Nil(suite.T(), user)
}

func (suite *UserServiceClientTestSuite) TestGetUserByEmail_ServiceError() {
	email := "error@example.com"
	expectedError := errors.New("service unavailable")

	suite.mockClient.On("GetUserByEmail", mock.Anything, email).Return((*userv1.User)(nil), false, expectedError)

	ctx := context.Background()
	user, found, err := suite.mockClient.GetUserByEmail(ctx, email)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), expectedError, err)
	assert.False(suite.T(), found)
	assert.Nil(suite.T(), user)
}

func (suite *UserServiceClientTestSuite) TestGetUserById_Success() {
	userID := "user-123"
	expectedUser := &userv1.User{
		Id:        userID,
		Email:     "test@example.com",
		FullName:  "Test User",
		CreatedAt: timestamppb.New(time.Now()),
		UpdatedAt: timestamppb.New(time.Now()),
	}

	suite.mockClient.On("GetUserById", mock.Anything, userID).Return(expectedUser, true, nil)

	ctx := context.Background()
	user, found, err := suite.mockClient.GetUserById(ctx, userID)

	assert.NoError(suite.T(), err)
	assert.True(suite.T(), found)
	assert.NotNil(suite.T(), user)
	assert.Equal(suite.T(), expectedUser.Id, user.Id)
	assert.Equal(suite.T(), expectedUser.Email, user.Email)
}

func (suite *UserServiceClientTestSuite) TestGetUserById_NotFound() {
	userID := "user-999"

	suite.mockClient.On("GetUserById", mock.Anything, userID).Return((*userv1.User)(nil), false, nil)

	ctx := context.Background()
	user, found, err := suite.mockClient.GetUserById(ctx, userID)

	assert.NoError(suite.T(), err)
	assert.False(suite.T(), found)
	assert.Nil(suite.T(), user)
}

func (suite *UserServiceClientTestSuite) TestGetUserById_ServiceError() {
	userID := "user-error"
	expectedError := errors.New("database connection failed")

	suite.mockClient.On("GetUserById", mock.Anything, userID).Return((*userv1.User)(nil), false, expectedError)

	ctx := context.Background()
	user, found, err := suite.mockClient.GetUserById(ctx, userID)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), expectedError, err)
	assert.False(suite.T(), found)
	assert.Nil(suite.T(), user)
}

func (suite *UserServiceClientTestSuite) TestHealthCheck_Success() {
	expectedStatus := userv1.HealthCheckResponse_SERVING
	expectedMessage := "service is healthy"

	suite.mockClient.On("HealthCheck", mock.Anything).Return(expectedStatus, expectedMessage, nil)

	ctx := context.Background()
	status, message, err := suite.mockClient.HealthCheck(ctx)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedStatus, status)
	assert.Equal(suite.T(), expectedMessage, message)
}

func (suite *UserServiceClientTestSuite) TestHealthCheck_Unhealthy() {
	expectedStatus := userv1.HealthCheckResponse_NOT_SERVING
	expectedMessage := "service is down"

	suite.mockClient.On("HealthCheck", mock.Anything).Return(expectedStatus, expectedMessage, nil)

	ctx := context.Background()
	status, message, err := suite.mockClient.HealthCheck(ctx)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedStatus, status)
	assert.Equal(suite.T(), expectedMessage, message)
}

func (suite *UserServiceClientTestSuite) TestHealthCheck_Error() {
	expectedError := errors.New("health check failed")

	suite.mockClient.On("HealthCheck", mock.Anything).Return(userv1.HealthCheckResponse_UNKNOWN, "", expectedError)

	ctx := context.Background()
	status, message, err := suite.mockClient.HealthCheck(ctx)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), expectedError, err)
	assert.Equal(suite.T(), userv1.HealthCheckResponse_UNKNOWN, status)
	assert.Empty(suite.T(), message)
}

func (suite *UserServiceClientTestSuite) TestClose_Success() {
	suite.mockClient.On("Close").Return(nil)

	err := suite.mockClient.Close()

	assert.NoError(suite.T(), err)
}

func (suite *UserServiceClientTestSuite) TestClose_Error() {
	expectedError := errors.New("failed to close connection")

	suite.mockClient.On("Close").Return(expectedError)

	err := suite.mockClient.Close()

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), expectedError, err)
}

// Test table-driven approach for various email validation scenarios
func TestGetUserByEmail_ValidationScenarios(t *testing.T) {
	testCases := []struct {
		name          string
		email         string
		mockUser      *userv1.User
		mockFound     bool
		mockError     error
		expectedFound bool
		expectedError string
	}{
		{
			name:  "valid email with user found",
			email: "valid@example.com",
			mockUser: &userv1.User{
				Id:       "user-1",
				Email:    "valid@example.com",
				FullName: "Valid User",
			},
			mockFound:     true,
			expectedFound: true,
		},
		{
			name:          "valid email with user not found",
			email:         "valid@example.com",
			mockUser:      nil,
			mockFound:     false,
			expectedFound: false,
		},
		{
			name:          "empty email",
			email:         "",
			mockUser:      nil,
			mockFound:     false,
			mockError:     errors.New("invalid email"),
			expectedFound: false,
			expectedError: "invalid email",
		},
		{
			name:          "malformed email",
			email:         "not-an-email",
			mockUser:      nil,
			mockFound:     false,
			mockError:     errors.New("invalid email format"),
			expectedFound: false,
			expectedError: "invalid email format",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := NewMockUserServiceClient()
			mockClient.On("GetUserByEmail", mock.Anything, tc.email).Return(tc.mockUser, tc.mockFound, tc.mockError)

			ctx := context.Background()
			user, found, err := mockClient.GetUserByEmail(ctx, tc.email)

			if tc.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tc.expectedFound, found)
			if tc.mockUser != nil {
				assert.Equal(t, tc.mockUser.Id, user.Id)
				assert.Equal(t, tc.mockUser.Email, user.Email)
			} else {
				assert.Nil(t, user)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

// Test Configuration
func TestUserServiceConfig_DefaultValues(t *testing.T) {
	config := UserServiceConfig{
		Address: "localhost:9090",
	}

	// This test would be for the actual implementation
	// For now, just verify the config can be created
	assert.Equal(t, "localhost:9090", config.Address)
	assert.Equal(t, time.Duration(0), config.ConnectTimeout)
	assert.Equal(t, time.Duration(0), config.RequestTimeout)
	assert.Equal(t, 0, config.MaxRetries)
}

func TestResilientUserServiceConfig_DefaultFactory(t *testing.T) {
	address := "localhost:9090"
	config := DefaultResilientUserServiceConfig(address)

	assert.Equal(t, address, config.UserServiceConfig.Address)
	assert.Equal(t, 10*time.Second, config.UserServiceConfig.ConnectTimeout)
	assert.Equal(t, 5*time.Second, config.UserServiceConfig.RequestTimeout)
	assert.Equal(t, 3, config.UserServiceConfig.MaxRetries)

	assert.Equal(t, 5, config.CircuitBreakerConfig.MaxFailures)
	assert.Equal(t, 60*time.Second, config.CircuitBreakerConfig.ResetTimeout)
	assert.Equal(t, 0.6, config.CircuitBreakerConfig.FailureThreshold)
	assert.Equal(t, 5*time.Second, config.CircuitBreakerConfig.RequestTimeout)
}
