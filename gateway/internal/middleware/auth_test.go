package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// MockTokenValidator is a mock implementation of TokenValidator
type MockTokenValidator struct {
	mock.Mock
}

func (m *MockTokenValidator) ValidateToken(ctx context.Context, token string) (*UserContext, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*UserContext), args.Error(1)
}

// AuthMiddlewareTestSuite is a test suite for AuthMiddleware
type AuthMiddlewareTestSuite struct {
	suite.Suite
	mockValidator *MockTokenValidator
	middleware    *AuthMiddleware
}

func (suite *AuthMiddlewareTestSuite) SetupTest() {
	suite.mockValidator = new(MockTokenValidator)
	publicPaths := []string{"/health", "/public/*", "/auth/*"}
	suite.middleware = NewAuthMiddleware(suite.mockValidator, publicPaths)
}

func (suite *AuthMiddlewareTestSuite) TearDownTest() {
	suite.mockValidator.AssertExpectations(suite.T())
}

// TestPublicPath_NoAuthRequired tests that public paths bypass authentication
func (suite *AuthMiddlewareTestSuite) TestPublicPath_NoAuthRequired() {
	// Arrange
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	nextCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	// Act
	handler := suite.middleware.Handler(nextHandler)
	handler.ServeHTTP(w, req)

	// Assert
	assert.True(suite.T(), nextCalled, "Next handler should be called for public paths")
	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

// TestPublicPathWithWildcard tests wildcard public paths
func (suite *AuthMiddlewareTestSuite) TestPublicPathWithWildcard() {
	// Arrange
	req := httptest.NewRequest("GET", "/public/some/nested/path", nil)
	w := httptest.NewRecorder()

	nextCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	// Act
	handler := suite.middleware.Handler(nextHandler)
	handler.ServeHTTP(w, req)

	// Assert
	assert.True(suite.T(), nextCalled)
	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

// TestProtectedPath_MissingAuthHeader tests protected path without auth header
func (suite *AuthMiddlewareTestSuite) TestProtectedPath_MissingAuthHeader() {
	// Arrange
	req := httptest.NewRequest("GET", "/protected/resource", nil)
	w := httptest.NewRecorder()

	nextCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	// Act
	handler := suite.middleware.Handler(nextHandler)
	handler.ServeHTTP(w, req)

	// Assert
	assert.False(suite.T(), nextCalled, "Next handler should not be called without auth")
	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
	assert.Contains(suite.T(), w.Body.String(), "missing authorization header")
}

// TestProtectedPath_InvalidAuthHeader tests protected path with invalid auth header format
func (suite *AuthMiddlewareTestSuite) TestProtectedPath_InvalidAuthHeader() {
	// Arrange
	req := httptest.NewRequest("GET", "/protected/resource", nil)
	req.Header.Set("Authorization", "InvalidFormat")
	w := httptest.NewRecorder()

	nextCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	// Act
	handler := suite.middleware.Handler(nextHandler)
	handler.ServeHTTP(w, req)

	// Assert
	assert.False(suite.T(), nextCalled)
	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
	assert.Contains(suite.T(), w.Body.String(), "invalid authorization header")
}

// TestProtectedPath_ValidToken tests protected path with valid token
func (suite *AuthMiddlewareTestSuite) TestProtectedPath_ValidToken() {
	// Arrange
	req := httptest.NewRequest("GET", "/protected/resource", nil)
	req.Header.Set("Authorization", "Bearer valid-token-123")
	w := httptest.NewRecorder()

	expectedUserContext := &UserContext{
		UserID: "user-123",
		Email:  "user@example.com",
	}

	suite.mockValidator.On("ValidateToken", mock.Anything, "valid-token-123").
		Return(expectedUserContext, nil)

	nextCalled := false
	var capturedHeaders http.Header
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		capturedHeaders = r.Header.Clone()
		w.WriteHeader(http.StatusOK)
	})

	// Act
	handler := suite.middleware.Handler(nextHandler)
	handler.ServeHTTP(w, req)

	// Assert
	assert.True(suite.T(), nextCalled, "Next handler should be called with valid token")
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	assert.Equal(suite.T(), "user-123", capturedHeaders.Get("X-User-ID"))
	assert.Equal(suite.T(), "user@example.com", capturedHeaders.Get("X-User-Email"))
}

// TestProtectedPath_InvalidToken tests protected path with invalid token
func (suite *AuthMiddlewareTestSuite) TestProtectedPath_InvalidToken() {
	// Arrange
	req := httptest.NewRequest("GET", "/protected/resource", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()

	suite.mockValidator.On("ValidateToken", mock.Anything, "invalid-token").
		Return(nil, errors.New("token validation failed"))

	nextCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	// Act
	handler := suite.middleware.Handler(nextHandler)
	handler.ServeHTTP(w, req)

	// Assert
	assert.False(suite.T(), nextCalled)
	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
	assert.Contains(suite.T(), w.Body.String(), "Invalid token")
}

// TestExtractToken tests token extraction functionality
func (suite *AuthMiddlewareTestSuite) TestExtractToken_Success() {
	// Arrange
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer my-token-value")

	// Act
	token, err := suite.middleware.extractToken(req)

	// Assert
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "my-token-value", token)
}

// TestExtractToken_CaseInsensitive tests that Bearer is case-insensitive
func (suite *AuthMiddlewareTestSuite) TestExtractToken_CaseInsensitive() {
	// Arrange
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "bearer my-token-value")

	// Act
	token, err := suite.middleware.extractToken(req)

	// Assert
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "my-token-value", token)
}

// TestAuthMiddlewareTestSuite runs the test suite
func TestAuthMiddlewareTestSuite(t *testing.T) {
	suite.Run(t, new(AuthMiddlewareTestSuite))
}