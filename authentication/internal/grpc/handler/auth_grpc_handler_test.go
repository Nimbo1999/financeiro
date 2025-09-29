package handler

import (
	"context"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/nimbo1999/financeiro/authentication/internal/services"
	authpb "github.com/nimbo1999/financeiro/authentication/pkg/grpc/auth/v1"
	"github.com/nimbo1999/financeiro/authentication/pkg/crypto"
)

func TestAuthGRPCHandler_ValidateToken_Success(t *testing.T) {
	// Setup: Generate RSA keys and JWT service
	privateKeyPEM, publicKeyPEM, err := crypto.GenerateRSAKeyPair(2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA keys: %v", err)
	}

	privateKey, err := crypto.ParseRSAPrivateKey(privateKeyPEM)
	if err != nil {
		t.Fatalf("Failed to parse private key: %v", err)
	}

	publicKey, err := crypto.ParseRSAPublicKey(publicKeyPEM)
	if err != nil {
		t.Fatalf("Failed to parse public key: %v", err)
	}

	jwtService := services.NewJWTService(&services.JWTConfig{
		PrivateKey:           privateKey,
		PublicKey:            publicKey,
		AccessTokenDuration:  1 * time.Hour,
		RefreshTokenDuration: 7 * 24 * time.Hour,
		Issuer:               "test-service",
	})

	// Create handler
	handler := NewAuthGRPCHandler(jwtService)

	// Generate a valid token
	ctx := context.Background()
	tokenPair, err := jwtService.GenerateTokenPair(ctx, "user-123", "user@example.com")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Test: Validate the token
	req := &authpb.ValidateTokenRequest{
		Token: tokenPair.AccessToken,
	}

	resp, err := handler.ValidateToken(ctx, req)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	// Assertions
	if !resp.Valid {
		t.Errorf("Expected token to be valid, got invalid. Error: %s", resp.ErrorMessage)
	}

	if resp.UserContext == nil {
		t.Fatal("Expected user context, got nil")
	}

	if resp.UserContext.UserId != "user-123" {
		t.Errorf("Expected user ID 'user-123', got '%s'", resp.UserContext.UserId)
	}

	if resp.UserContext.Email != "user@example.com" {
		t.Errorf("Expected email 'user@example.com', got '%s'", resp.UserContext.Email)
	}
}

func TestAuthGRPCHandler_ValidateToken_InvalidToken(t *testing.T) {
	// Setup: Generate RSA keys and JWT service
	privateKeyPEM, publicKeyPEM, err := crypto.GenerateRSAKeyPair(2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA keys: %v", err)
	}

	privateKey, err := crypto.ParseRSAPrivateKey(privateKeyPEM)
	if err != nil {
		t.Fatalf("Failed to parse private key: %v", err)
	}

	publicKey, err := crypto.ParseRSAPublicKey(publicKeyPEM)
	if err != nil {
		t.Fatalf("Failed to parse public key: %v", err)
	}

	jwtService := services.NewJWTService(&services.JWTConfig{
		PrivateKey:           privateKey,
		PublicKey:            publicKey,
		AccessTokenDuration:  1 * time.Hour,
		RefreshTokenDuration: 7 * 24 * time.Hour,
		Issuer:               "test-service",
	})

	// Create handler
	handler := NewAuthGRPCHandler(jwtService)

	// Test: Validate an invalid token
	ctx := context.Background()
	req := &authpb.ValidateTokenRequest{
		Token: "invalid-token",
	}

	resp, err := handler.ValidateToken(ctx, req)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	// Assertions
	if resp.Valid {
		t.Error("Expected token to be invalid, got valid")
	}

	if resp.ErrorMessage == "" {
		t.Error("Expected error message for invalid token")
	}
}

func TestAuthGRPCHandler_ValidateToken_EmptyToken(t *testing.T) {
	// Setup: Generate RSA keys and JWT service
	privateKeyPEM, publicKeyPEM, err := crypto.GenerateRSAKeyPair(2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA keys: %v", err)
	}

	privateKey, err := crypto.ParseRSAPrivateKey(privateKeyPEM)
	if err != nil {
		t.Fatalf("Failed to parse private key: %v", err)
	}

	publicKey, err := crypto.ParseRSAPublicKey(publicKeyPEM)
	if err != nil {
		t.Fatalf("Failed to parse public key: %v", err)
	}

	jwtService := services.NewJWTService(&services.JWTConfig{
		PrivateKey:           privateKey,
		PublicKey:            publicKey,
		AccessTokenDuration:  1 * time.Hour,
		RefreshTokenDuration: 7 * 24 * time.Hour,
		Issuer:               "test-service",
	})

	// Create handler
	handler := NewAuthGRPCHandler(jwtService)

	// Test: Validate empty token
	ctx := context.Background()
	req := &authpb.ValidateTokenRequest{
		Token: "",
	}

	resp, err := handler.ValidateToken(ctx, req)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	// Assertions
	if resp.Valid {
		t.Error("Expected token to be invalid, got valid")
	}
}

func TestAuthGRPCHandler_HealthCheck(t *testing.T) {
	// Setup: Generate RSA keys and JWT service
	privateKeyPEM, publicKeyPEM, err := crypto.GenerateRSAKeyPair(2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA keys: %v", err)
	}

	privateKey, err := crypto.ParseRSAPrivateKey(privateKeyPEM)
	if err != nil {
		t.Fatalf("Failed to parse private key: %v", err)
	}

	publicKey, err := crypto.ParseRSAPublicKey(publicKeyPEM)
	if err != nil {
		t.Fatalf("Failed to parse public key: %v", err)
	}

	jwtService := services.NewJWTService(&services.JWTConfig{
		PrivateKey:           privateKey,
		PublicKey:            publicKey,
		AccessTokenDuration:  1 * time.Hour,
		RefreshTokenDuration: 7 * 24 * time.Hour,
		Issuer:               "test-service",
	})

	// Create handler
	handler := NewAuthGRPCHandler(jwtService)

	// Test: Health check
	ctx := context.Background()
	req := &authpb.HealthCheckRequest{}

	resp, err := handler.HealthCheck(ctx, req)
	if err != nil {
		t.Fatalf("HealthCheck failed: %v", err)
	}

	// Assertions
	if resp.Status != authpb.HealthCheckResponse_SERVING {
		t.Errorf("Expected status SERVING, got %v", resp.Status)
	}

	if resp.Message == "" {
		t.Error("Expected health check message")
	}
}

func TestAuthGRPCHandler_HealthCheck_NoJWTService(t *testing.T) {
	// Create handler with nil JWT service
	handler := &AuthGRPCHandler{
		jwtService: nil,
	}

	// Test: Health check with nil service
	ctx := context.Background()
	req := &authpb.HealthCheckRequest{}

	resp, err := handler.HealthCheck(ctx, req)
	if err != nil {
		t.Fatalf("HealthCheck failed: %v", err)
	}

	// Assertions
	if resp.Status != authpb.HealthCheckResponse_NOT_SERVING {
		t.Errorf("Expected status NOT_SERVING, got %v", resp.Status)
	}
}

func TestAuthGRPCHandler_ValidateToken_WrongTokenType(t *testing.T) {
	// Setup: Generate RSA keys and JWT service
	privateKeyPEM, publicKeyPEM, err := crypto.GenerateRSAKeyPair(2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA keys: %v", err)
	}

	privateKey, err := crypto.ParseRSAPrivateKey(privateKeyPEM)
	if err != nil {
		t.Fatalf("Failed to parse private key: %v", err)
	}

	publicKey, err := crypto.ParseRSAPublicKey(publicKeyPEM)
	if err != nil {
		t.Fatalf("Failed to parse public key: %v", err)
	}

	jwtService := services.NewJWTService(&services.JWTConfig{
		PrivateKey:           privateKey,
		PublicKey:            publicKey,
		AccessTokenDuration:  1 * time.Hour,
		RefreshTokenDuration: 7 * 24 * time.Hour,
		Issuer:               "test-service",
	})

	// Create handler
	handler := NewAuthGRPCHandler(jwtService)

	// Generate a refresh token (not access token)
	ctx := context.Background()
	tokenPair, err := jwtService.GenerateTokenPair(ctx, "user-123", "user@example.com")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Test: Try to validate refresh token as access token
	req := &authpb.ValidateTokenRequest{
		Token: tokenPair.RefreshToken,
	}

	resp, err := handler.ValidateToken(ctx, req)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	// Assertions
	if resp.Valid {
		t.Error("Expected refresh token to be invalid for access token validation")
	}

	if resp.ErrorMessage != "invalid token type" {
		t.Errorf("Expected error message 'invalid token type', got '%s'", resp.ErrorMessage)
	}
}

// Mock JWT service for testing
type mockJWTService struct {
	publicKey *rsa.PublicKey
}

func (m *mockJWTService) GenerateTokenPair(ctx context.Context, userID, email string) (*services.TokenPair, error) {
	return nil, nil
}

func (m *mockJWTService) ValidateAccessToken(ctx context.Context, tokenString string) (*services.UserContext, error) {
	return nil, services.ErrInvalidToken
}

func (m *mockJWTService) ValidateRefreshToken(ctx context.Context, tokenString string) (*services.UserContext, error) {
	return nil, services.ErrInvalidToken
}

func (m *mockJWTService) RefreshTokens(ctx context.Context, refreshToken string) (*services.TokenPair, error) {
	return nil, nil
}

func (m *mockJWTService) GetPublicKey() *rsa.PublicKey {
	return m.publicKey
}