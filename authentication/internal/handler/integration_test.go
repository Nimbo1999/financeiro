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
	"github.com/nimbo1999/financeiro/authentication/pkg/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Integration test that shows the full authentication flow
func TestFullAuthenticationFlow(t *testing.T) {
	// Setup JWT service with real keys
	privateKeyPEM, publicKeyPEM, err := crypto.GenerateRSAKeyPair(2048)
	assert.NoError(t, err)

	privateKey, err := crypto.ParseRSAPrivateKey(privateKeyPEM)
	assert.NoError(t, err)

	publicKey, err := crypto.ParseRSAPublicKey(publicKeyPEM)
	assert.NoError(t, err)

	jwtConfig := &services.JWTConfig{
		PrivateKey:           privateKey,
		PublicKey:            publicKey,
		AccessTokenDuration:  1 * time.Hour,
		RefreshTokenDuration: 7 * 24 * time.Hour,
		Issuer:               "test-service",
	}

	jwtService := services.NewJWTService(jwtConfig)

	// Setup mock auth service for integration test
	mockAuthService := new(MockAuthService)

	// Setup handler
	handler := NewAuthHandler(mockAuthService)

	// Test email
	email := "integration@test.com"

	// Step 1: Request auth code
	t.Run("Request Auth Code", func(t *testing.T) {
		expectedResult := &services.AuthCodeResult{
			CodeID:    "test-code-id",
			ExpiresAt: time.Now().Add(5 * time.Minute),
			Success:   true,
		}

		mockAuthService.On("RequestAuthCode", mock.Anything, email).Return(expectedResult, nil)

		reqBody := RequestCodeRequest{Email: email}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/request-code", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		handler.RequestCodeHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)
	})

	// Step 2: Verify auth code
	var tokens *services.TokenPair
	t.Run("Verify Auth Code", func(t *testing.T) {
		code := "123456"

		// Generate real tokens for this test
		realTokens, err := jwtService.GenerateTokenPair(context.Background(), email, email)
		assert.NoError(t, err)

		expectedResult := &services.AuthResult{
			UserID:          email,
			Email:           email,
			TokenPair:       realTokens,
			IsNewUser:       false,
			AuthenticatedAt: time.Now(),
		}

		mockAuthService.On("VerifyAuthCode", mock.Anything, email, code).Return(expectedResult, nil)

		reqBody := VerifyCodeRequest{Email: email, Code: code}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/verify-code", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		handler.VerifyCodeHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response APIResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)

		// Extract tokens from response
		data := response.Data.(map[string]interface{})
		tokensData := data["tokens"].(map[string]interface{})
		tokens = &services.TokenPair{
			AccessToken:  tokensData["access_token"].(string),
			RefreshToken: tokensData["refresh_token"].(string),
		}

		assert.NotEmpty(t, tokens.AccessToken)
		assert.NotEmpty(t, tokens.RefreshToken)
	})

	// Step 3: Refresh tokens
	t.Run("Refresh Tokens", func(t *testing.T) {
		// Generate new tokens for refresh response
		newTokens, err := jwtService.GenerateTokenPair(context.Background(), email, email)
		assert.NoError(t, err)

		mockAuthService.On("RefreshTokens", mock.Anything, tokens.RefreshToken).Return(newTokens, nil)

		reqBody := RefreshTokenRequest{RefreshToken: tokens.RefreshToken}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		handler.RefreshTokenHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response APIResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)

		// Verify tokens are returned
		data := response.Data.(map[string]interface{})
		tokensData := data["tokens"].(map[string]interface{})
		responseTokens := &services.TokenPair{
			AccessToken:  tokensData["access_token"].(string),
			RefreshToken: tokensData["refresh_token"].(string),
		}

		assert.NotEmpty(t, responseTokens.AccessToken)
		assert.NotEmpty(t, responseTokens.RefreshToken)
	})

	// Verify all mock expectations
	mockAuthService.AssertExpectations(t)
}

// Test middleware chain
func TestMiddlewareChain(t *testing.T) {
	mockService := new(MockAuthService)
	handler := NewAuthHandler(mockService)

	t.Run("Content Type Validation", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/request-code", bytes.NewBuffer([]byte(`{"email":"test@example.com"}`)))
		req.Header.Set("Content-Type", "text/plain") // Wrong content type

		w := httptest.NewRecorder()

		// Apply content type middleware
		middleware := ContentTypeMiddleware(http.HandlerFunc(handler.RequestCodeHandler))
		middleware.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnsupportedMediaType, w.Code)
	})

	t.Run("Request Size Limit", func(t *testing.T) {
		// Create a large request body (over 1MB)
		largeBody := make([]byte, 1024*1024+1)
		for i := range largeBody {
			largeBody[i] = 'a'
		}

		req := httptest.NewRequest(http.MethodPost, "/request-code", bytes.NewBuffer(largeBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()

		// Apply validation middleware
		middleware := ValidationMiddleware(http.HandlerFunc(handler.RequestCodeHandler))
		middleware.ServeHTTP(w, req)

		// Should reject large requests
		assert.NotEqual(t, http.StatusOK, w.Code)
	})

	t.Run("Recovery Middleware", func(t *testing.T) {
		panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("test panic")
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		// Apply recovery middleware
		middleware := RecoveryMiddleware(panicHandler)
		middleware.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.False(t, response.Success)
		assert.Equal(t, "INTERNAL_ERROR", response.Error.Code)
	})
}
