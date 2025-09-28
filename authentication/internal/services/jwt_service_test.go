package services

import (
	"context"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/nimbo1999/financeiro/authentication/pkg/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type JWTServiceTestSuite struct {
	suite.Suite
	jwtService JWTService
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	config     *JWTConfig
}

func (suite *JWTServiceTestSuite) SetupTest() {
	// Generate test RSA key pair
	privateKeyPEM, publicKeyPEM, err := crypto.GenerateRSAKeyPair(2048)
	suite.Require().NoError(err)

	suite.privateKey, err = crypto.ParseRSAPrivateKey(privateKeyPEM)
	suite.Require().NoError(err)

	suite.publicKey, err = crypto.ParseRSAPublicKey(publicKeyPEM)
	suite.Require().NoError(err)

	suite.config = &JWTConfig{
		PrivateKey:           suite.privateKey,
		PublicKey:            suite.publicKey,
		AccessTokenDuration:  1 * time.Hour,
		RefreshTokenDuration: 7 * 24 * time.Hour,
		Issuer:               "test-service",
	}

	suite.jwtService = NewJWTService(suite.config)
}

func TestJWTServiceTestSuite(t *testing.T) {
	suite.Run(t, new(JWTServiceTestSuite))
}

// GenerateTokenPair tests
func (suite *JWTServiceTestSuite) TestGenerateTokenPair_Success() {
	userID := "user-123"
	email := "test@example.com"

	tokenPair, err := suite.jwtService.GenerateTokenPair(context.Background(), userID, email)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), tokenPair)
	assert.NotEmpty(suite.T(), tokenPair.AccessToken)
	assert.NotEmpty(suite.T(), tokenPair.RefreshToken)
	assert.NotEqual(suite.T(), tokenPair.AccessToken, tokenPair.RefreshToken)
}

func (suite *JWTServiceTestSuite) TestGenerateTokenPair_EmptyUserID() {
	email := "test@example.com"

	tokenPair, err := suite.jwtService.GenerateTokenPair(context.Background(), "", email)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), tokenPair)
	assert.Contains(suite.T(), err.Error(), "user ID cannot be empty")
}

func (suite *JWTServiceTestSuite) TestGenerateTokenPair_EmptyEmail() {
	userID := "user-123"

	tokenPair, err := suite.jwtService.GenerateTokenPair(context.Background(), userID, "")

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), tokenPair)
	assert.Contains(suite.T(), err.Error(), "email cannot be empty")
}

// ValidateAccessToken tests
func (suite *JWTServiceTestSuite) TestValidateAccessToken_Success() {
	userID := "user-123"
	email := "test@example.com"

	tokenPair, err := suite.jwtService.GenerateTokenPair(context.Background(), userID, email)
	suite.Require().NoError(err)

	userContext, err := suite.jwtService.ValidateAccessToken(context.Background(), tokenPair.AccessToken)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), userContext)
	assert.Equal(suite.T(), userID, userContext.UserID)
	assert.Equal(suite.T(), email, userContext.Email)
}

func (suite *JWTServiceTestSuite) TestValidateAccessToken_InvalidToken() {
	userContext, err := suite.jwtService.ValidateAccessToken(context.Background(), "invalid-token")

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), userContext)
	assert.Equal(suite.T(), ErrInvalidToken, err)
}

func (suite *JWTServiceTestSuite) TestValidateAccessToken_EmptyToken() {
	userContext, err := suite.jwtService.ValidateAccessToken(context.Background(), "")

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), userContext)
	assert.Equal(suite.T(), ErrInvalidToken, err)
}

func (suite *JWTServiceTestSuite) TestValidateAccessToken_WrongTokenType() {
	userID := "user-123"
	email := "test@example.com"

	tokenPair, err := suite.jwtService.GenerateTokenPair(context.Background(), userID, email)
	suite.Require().NoError(err)

	// Try to validate refresh token as access token
	userContext, err := suite.jwtService.ValidateAccessToken(context.Background(), tokenPair.RefreshToken)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), userContext)
	assert.Equal(suite.T(), ErrInvalidTokenType, err)
}

// ValidateRefreshToken tests
func (suite *JWTServiceTestSuite) TestValidateRefreshToken_Success() {
	userID := "user-123"
	email := "test@example.com"

	tokenPair, err := suite.jwtService.GenerateTokenPair(context.Background(), userID, email)
	suite.Require().NoError(err)

	userContext, err := suite.jwtService.ValidateRefreshToken(context.Background(), tokenPair.RefreshToken)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), userContext)
	assert.Equal(suite.T(), userID, userContext.UserID)
	assert.Equal(suite.T(), email, userContext.Email)
}

func (suite *JWTServiceTestSuite) TestValidateRefreshToken_WrongTokenType() {
	userID := "user-123"
	email := "test@example.com"

	tokenPair, err := suite.jwtService.GenerateTokenPair(context.Background(), userID, email)
	suite.Require().NoError(err)

	// Try to validate access token as refresh token
	userContext, err := suite.jwtService.ValidateRefreshToken(context.Background(), tokenPair.AccessToken)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), userContext)
	assert.Equal(suite.T(), ErrInvalidTokenType, err)
}

// RefreshTokens tests
func (suite *JWTServiceTestSuite) TestRefreshTokens_Success() {
	userID := "user-123"
	email := "test@example.com"

	originalTokenPair, err := suite.jwtService.GenerateTokenPair(context.Background(), userID, email)
	suite.Require().NoError(err)

	newTokenPair, err := suite.jwtService.RefreshTokens(context.Background(), originalTokenPair.RefreshToken)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), newTokenPair)
	assert.NotEmpty(suite.T(), newTokenPair.AccessToken)
	assert.NotEmpty(suite.T(), newTokenPair.RefreshToken)
	assert.NotEqual(suite.T(), originalTokenPair.AccessToken, newTokenPair.AccessToken)
	assert.NotEqual(suite.T(), originalTokenPair.RefreshToken, newTokenPair.RefreshToken)

	// Verify new tokens are valid
	userContext, err := suite.jwtService.ValidateAccessToken(context.Background(), newTokenPair.AccessToken)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), userID, userContext.UserID)
	assert.Equal(suite.T(), email, userContext.Email)
}

func (suite *JWTServiceTestSuite) TestRefreshTokens_InvalidRefreshToken() {
	newTokenPair, err := suite.jwtService.RefreshTokens(context.Background(), "invalid-token")

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), newTokenPair)
}

// GetPublicKey tests
func (suite *JWTServiceTestSuite) TestGetPublicKey() {
	publicKey := suite.jwtService.GetPublicKey()

	assert.NotNil(suite.T(), publicKey)
	assert.Equal(suite.T(), suite.publicKey, publicKey)
}

// Token expiration tests
func (suite *JWTServiceTestSuite) TestExpiredToken() {
	// Create a service with very short token duration
	shortConfig := &JWTConfig{
		PrivateKey:           suite.privateKey,
		PublicKey:            suite.publicKey,
		AccessTokenDuration:  1 * time.Millisecond,
		RefreshTokenDuration: 1 * time.Millisecond,
		Issuer:               "test-service",
	}
	shortJWTService := NewJWTService(shortConfig)

	userID := "user-123"
	email := "test@example.com"

	tokenPair, err := shortJWTService.GenerateTokenPair(context.Background(), userID, email)
	suite.Require().NoError(err)

	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)

	// Try to validate expired access token
	userContext, err := shortJWTService.ValidateAccessToken(context.Background(), tokenPair.AccessToken)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), userContext)
	assert.Equal(suite.T(), ErrExpiredToken, err)

	// Try to validate expired refresh token
	userContext, err = shortJWTService.ValidateRefreshToken(context.Background(), tokenPair.RefreshToken)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), userContext)
	assert.Equal(suite.T(), ErrExpiredToken, err)
}

// Configuration defaults tests
func (suite *JWTServiceTestSuite) TestNewJWTService_DefaultValues() {
	config := &JWTConfig{
		PrivateKey: suite.privateKey,
		PublicKey:  suite.publicKey,
	}

	jwtService := NewJWTService(config)

	assert.NotNil(suite.T(), jwtService)
	assert.Equal(suite.T(), 1*time.Hour, config.AccessTokenDuration)
	assert.Equal(suite.T(), 7*24*time.Hour, config.RefreshTokenDuration)
	assert.Equal(suite.T(), "authentication-service", config.Issuer)
}

// Token claims validation tests
func (suite *JWTServiceTestSuite) TestTokenClaims() {
	userID := "user-123"
	email := "test@example.com"

	tokenPair, err := suite.jwtService.GenerateTokenPair(context.Background(), userID, email)
	suite.Require().NoError(err)

	// Parse access token claims
	accessToken, err := jwt.ParseWithClaims(tokenPair.AccessToken, &Claims{}, func(token *jwt.Token) (any, error) {
		return suite.publicKey, nil
	})
	suite.Require().NoError(err)

	accessClaims := accessToken.Claims.(*Claims)
	suite.Equal(userID, accessClaims.UserID)
	suite.Equal(email, accessClaims.Email)
	suite.Equal(AccessTokenType, accessClaims.TokenType)
	suite.Equal("test-service", accessClaims.Issuer)
	suite.NotEmpty(accessClaims.ID)
	suite.True(accessClaims.ExpiresAt.After(time.Now()))

	// Parse refresh token claims
	refreshToken, err := jwt.ParseWithClaims(tokenPair.RefreshToken, &Claims{}, func(token *jwt.Token) (any, error) {
		return suite.publicKey, nil
	})
	suite.Require().NoError(err)

	refreshClaims := refreshToken.Claims.(*Claims)
	suite.Equal(userID, refreshClaims.UserID)
	suite.Equal(email, refreshClaims.Email)
	suite.Equal(RefreshTokenType, refreshClaims.TokenType)
	suite.Equal("test-service", refreshClaims.Issuer)
	suite.NotEmpty(refreshClaims.ID)
	suite.True(refreshClaims.ExpiresAt.After(accessClaims.ExpiresAt.Time))
}

// Table-driven tests
func (suite *JWTServiceTestSuite) TestValidateToken_TableDriven() {
	testCases := []struct {
		name        string
		tokenString string
		tokenType   TokenType
		expectError bool
		expectedErr error
	}{
		{
			name:        "Empty token",
			tokenString: "",
			tokenType:   AccessTokenType,
			expectError: true,
			expectedErr: ErrInvalidToken,
		},
		{
			name:        "Malformed token",
			tokenString: "malformed.token.here",
			tokenType:   AccessTokenType,
			expectError: true,
			expectedErr: ErrInvalidToken,
		},
		{
			name:        "Token with invalid signature",
			tokenString: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.invalid",
			tokenType:   AccessTokenType,
			expectError: true,
			expectedErr: ErrInvalidToken,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			var err error
			if tc.tokenType == AccessTokenType {
				_, err = suite.jwtService.ValidateAccessToken(context.Background(), tc.tokenString)
			} else {
				_, err = suite.jwtService.ValidateRefreshToken(context.Background(), tc.tokenString)
			}

			if tc.expectError {
				assert.Error(suite.T(), err)
				if tc.expectedErr != nil {
					assert.Equal(suite.T(), tc.expectedErr, err)
				}
			} else {
				assert.NoError(suite.T(), err)
			}
		})
	}
}

// Benchmark tests
func BenchmarkGenerateTokenPair(b *testing.B) {
	privateKeyPEM, publicKeyPEM, err := crypto.GenerateRSAKeyPair(2048)
	if err != nil {
		b.Fatal(err)
	}

	privateKey, err := crypto.ParseRSAPrivateKey(privateKeyPEM)
	if err != nil {
		b.Fatal(err)
	}

	publicKey, err := crypto.ParseRSAPublicKey(publicKeyPEM)
	if err != nil {
		b.Fatal(err)
	}

	config := &JWTConfig{
		PrivateKey:           privateKey,
		PublicKey:            publicKey,
		AccessTokenDuration:  1 * time.Hour,
		RefreshTokenDuration: 7 * 24 * time.Hour,
		Issuer:               "bench-service",
	}

	jwtService := NewJWTService(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = jwtService.GenerateTokenPair(context.Background(), "user-123", "test@example.com")
	}
}

func BenchmarkValidateAccessToken(b *testing.B) {
	privateKeyPEM, publicKeyPEM, err := crypto.GenerateRSAKeyPair(2048)
	if err != nil {
		b.Fatal(err)
	}

	privateKey, err := crypto.ParseRSAPrivateKey(privateKeyPEM)
	if err != nil {
		b.Fatal(err)
	}

	publicKey, err := crypto.ParseRSAPublicKey(publicKeyPEM)
	if err != nil {
		b.Fatal(err)
	}

	config := &JWTConfig{
		PrivateKey:           privateKey,
		PublicKey:            publicKey,
		AccessTokenDuration:  1 * time.Hour,
		RefreshTokenDuration: 7 * 24 * time.Hour,
		Issuer:               "bench-service",
	}

	jwtService := NewJWTService(config)

	// Generate a token for benchmarking
	tokenPair, err := jwtService.GenerateTokenPair(context.Background(), "user-123", "test@example.com")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = jwtService.ValidateAccessToken(context.Background(), tokenPair.AccessToken)
	}
}
