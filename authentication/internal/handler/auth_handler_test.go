package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/nimbo1999/financeiro/authentication/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// Mock AuthService
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) RequestAuthCode(ctx context.Context, email string) (*services.AuthCodeResult, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.AuthCodeResult), args.Error(1)
}

func (m *MockAuthService) VerifyAuthCode(ctx context.Context, email, code string) (*services.AuthResult, error) {
	args := m.Called(ctx, email, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.AuthResult), args.Error(1)
}

func (m *MockAuthService) RefreshTokens(ctx context.Context, refreshToken string) (*services.TokenPair, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.TokenPair), args.Error(1)
}

func (m *MockAuthService) CleanExpiredCodes(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Test Suite
type AuthHandlerTestSuite struct {
	suite.Suite
	handler     *AuthHandler
	mockService *MockAuthService
}

func (suite *AuthHandlerTestSuite) SetupTest() {
	suite.mockService = new(MockAuthService)
	suite.handler = NewAuthHandler(suite.mockService)
}

func (suite *AuthHandlerTestSuite) TearDownTest() {
	suite.mockService.AssertExpectations(suite.T())
}

func TestAuthHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(AuthHandlerTestSuite))
}

// RequestCodeHandler tests
func (suite *AuthHandlerTestSuite) TestRequestCodeHandler_Success() {
	email := "test@example.com"
	expectedResult := &services.AuthCodeResult{
		CodeID:    "code-123",
		ExpiresAt: time.Now().Add(5 * time.Minute),
		Success:   true,
	}

	suite.mockService.On("RequestAuthCode", mock.Anything, email).Return(expectedResult, nil)

	// Create request
	reqBody := RequestCodeRequest{Email: email}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/request-code", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	suite.handler.RequestCodeHandler(w, req)

	// Assert response
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
	assert.NotNil(suite.T(), response.Data)
}

func (suite *AuthHandlerTestSuite) TestRequestCodeHandler_InvalidEmail() {
	// Create request with invalid email
	reqBody := RequestCodeRequest{Email: "invalid-email"}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/request-code", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	suite.handler.RequestCodeHandler(w, req)

	// Assert response
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response.Success)
	assert.NotNil(suite.T(), response.Error)
}

func (suite *AuthHandlerTestSuite) TestRequestCodeHandler_EmptyEmail() {
	// Create request with empty email
	reqBody := RequestCodeRequest{Email: ""}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/request-code", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	suite.handler.RequestCodeHandler(w, req)

	// Assert response
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response.Success)
}

func (suite *AuthHandlerTestSuite) TestRequestCodeHandler_ServiceError() {
	email := "test@example.com"

	suite.mockService.On("RequestAuthCode", mock.Anything, email).Return(nil, services.ErrRateLimitExceeded)

	// Create request
	reqBody := RequestCodeRequest{Email: email}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/request-code", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	suite.handler.RequestCodeHandler(w, req)

	// Assert response
	assert.Equal(suite.T(), http.StatusTooManyRequests, w.Code)

	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response.Success)
	assert.Equal(suite.T(), "RATE_LIMIT_EXCEEDED", response.Error.Code)
}

// VerifyCodeHandler tests
func (suite *AuthHandlerTestSuite) TestVerifyCodeHandler_Success() {
	email := "test@example.com"
	code := "123456"
	expectedResult := &services.AuthResult{
		UserID: email,
		Email:  email,
		TokenPair: &services.TokenPair{
			AccessToken:  "access-token",
			RefreshToken: "refresh-token",
		},
		IsNewUser:       false,
		AuthenticatedAt: time.Now(),
	}

	suite.mockService.On("VerifyAuthCode", mock.Anything, email, code).Return(expectedResult, nil)

	// Create request
	reqBody := VerifyCodeRequest{Email: email, Code: code}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/verify-code", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	suite.handler.VerifyCodeHandler(w, req)

	// Assert response
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
	assert.NotNil(suite.T(), response.Data)
}

func (suite *AuthHandlerTestSuite) TestVerifyCodeHandler_InvalidCode() {
	email := "test@example.com"
	code := "12345" // Only 5 digits

	// Create request
	reqBody := VerifyCodeRequest{Email: email, Code: code}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/verify-code", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	suite.handler.VerifyCodeHandler(w, req)

	// Assert response
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response.Success)
}

func (suite *AuthHandlerTestSuite) TestVerifyCodeHandler_ExpiredCode() {
	email := "test@example.com"
	code := "123456"

	suite.mockService.On("VerifyAuthCode", mock.Anything, email, code).Return(nil, services.ErrCodeExpired)

	// Create request
	reqBody := VerifyCodeRequest{Email: email, Code: code}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/verify-code", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	suite.handler.VerifyCodeHandler(w, req)

	// Assert response
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response.Success)
	assert.Equal(suite.T(), "CODE_EXPIRED", response.Error.Code)
}

// RefreshTokenHandler tests
func (suite *AuthHandlerTestSuite) TestRefreshTokenHandler_Success() {
	refreshToken := "valid-refresh-token"
	expectedTokenPair := &services.TokenPair{
		AccessToken:  "new-access-token",
		RefreshToken: "new-refresh-token",
	}

	suite.mockService.On("RefreshTokens", mock.Anything, refreshToken).Return(expectedTokenPair, nil)

	// Create request
	reqBody := RefreshTokenRequest{RefreshToken: refreshToken}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	suite.handler.RefreshTokenHandler(w, req)

	// Assert response
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
	assert.NotNil(suite.T(), response.Data)
}

func (suite *AuthHandlerTestSuite) TestRefreshTokenHandler_EmptyToken() {
	// Create request with empty refresh token
	reqBody := RefreshTokenRequest{RefreshToken: ""}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	suite.handler.RefreshTokenHandler(w, req)

	// Assert response
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response.Success)
}

func (suite *AuthHandlerTestSuite) TestRefreshTokenHandler_InvalidToken() {
	refreshToken := "invalid-refresh-token"

	suite.mockService.On("RefreshTokens", mock.Anything, refreshToken).Return(nil, services.ErrInvalidToken)

	// Create request
	reqBody := RefreshTokenRequest{RefreshToken: refreshToken}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	suite.handler.RefreshTokenHandler(w, req)

	// Assert response
	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)

	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response.Success)
	assert.Equal(suite.T(), "INVALID_TOKEN", response.Error.Code)
}

// Test invalid JSON
func (suite *AuthHandlerTestSuite) TestInvalidJSON() {
	req := httptest.NewRequest(http.MethodPost, "/request-code", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.handler.RequestCodeHandler(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response.Success)
}
