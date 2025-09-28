package services

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nimbo1999/financeiro/authentication/internal/models"
	"github.com/nimbo1999/financeiro/authentication/internal/repository"
	"github.com/nimbo1999/financeiro/authentication/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// Mock implementations
type MockAuthCodeRepository struct {
	mock.Mock
}

func (m *MockAuthCodeRepository) Create(ctx context.Context, authCode *models.AuthCode) error {
	args := m.Called(ctx, authCode)
	return args.Error(0)
}

func (m *MockAuthCodeRepository) FindByUserID(ctx context.Context, userID string) (*models.AuthCode, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AuthCode), args.Error(1)
}

func (m *MockAuthCodeRepository) MarkAsUsed(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAuthCodeRepository) CleanExpired(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

type MockJWTService struct {
	mock.Mock
}

func (m *MockJWTService) GenerateTokenPair(ctx context.Context, userID, email string) (*TokenPair, error) {
	args := m.Called(ctx, userID, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TokenPair), args.Error(1)
}

func (m *MockJWTService) ValidateAccessToken(ctx context.Context, tokenString string) (*UserContext, error) {
	args := m.Called(ctx, tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*UserContext), args.Error(1)
}

func (m *MockJWTService) ValidateRefreshToken(ctx context.Context, tokenString string) (*UserContext, error) {
	args := m.Called(ctx, tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*UserContext), args.Error(1)
}

func (m *MockJWTService) RefreshTokens(ctx context.Context, refreshToken string) (*TokenPair, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TokenPair), args.Error(1)
}

func (m *MockJWTService) GetPublicKey() *rsa.PublicKey {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*rsa.PublicKey)
}

// Test Suite
type AuthServiceTestSuite struct {
	suite.Suite
	authService AuthService
	mockRepo    *MockAuthCodeRepository
	mockJWT     *MockJWTService
	config      *AuthConfig
}

func (suite *AuthServiceTestSuite) SetupTest() {
	suite.mockRepo = new(MockAuthCodeRepository)
	suite.mockJWT = new(MockJWTService)
	suite.config = &AuthConfig{
		CodeLength:         6,
		CodeExpiryDuration: 5 * time.Minute,
		RateLimitWindow:    1 * time.Hour,
		MaxCodesPerEmail:   3,
		MaxRequestsPerIP:   5,
		CleanupInterval:    1 * time.Hour,
	}
	suite.authService = NewAuthService(suite.mockRepo, suite.mockJWT, suite.config)
}

func (suite *AuthServiceTestSuite) TearDownTest() {
	suite.mockRepo.AssertExpectations(suite.T())
	suite.mockJWT.AssertExpectations(suite.T())
}

func TestAuthServiceTestSuite(t *testing.T) {
	suite.Run(t, new(AuthServiceTestSuite))
}

// RequestAuthCode tests
func (suite *AuthServiceTestSuite) TestRequestAuthCode_Success() {
	email := "test@example.com"

	suite.mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(authCode *models.AuthCode) bool {
		authCode.ID = uuid.NewString()
		return authCode.UserID == email &&
			len(authCode.Code) == 6 &&
			authCode.ExpiresAt.After(time.Now()) &&
			authCode.UsedAt == nil
	})).Return(nil)

	result, err := suite.authService.RequestAuthCode(context.Background(), email)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.True(suite.T(), result.Success)
	assert.NotEmpty(suite.T(), result.CodeID)
	assert.True(suite.T(), result.ExpiresAt.After(time.Now()))
}

func (suite *AuthServiceTestSuite) TestRequestAuthCode_EmptyEmail() {
	result, err := suite.authService.RequestAuthCode(context.Background(), "")

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), ErrEmailInvalid, err)
}

func (suite *AuthServiceTestSuite) TestRequestAuthCode_InvalidEmail() {
	result, err := suite.authService.RequestAuthCode(context.Background(), "invalid-email")

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), ErrEmailInvalid, err)
}

func (suite *AuthServiceTestSuite) TestRequestAuthCode_RepositoryError() {
	email := "test@example.com"

	suite.mockRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New("database error"))

	result, err := suite.authService.RequestAuthCode(context.Background(), email)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "failed to store auth code")
}

func (suite *AuthServiceTestSuite) TestRequestAuthCode_RateLimit() {
	email := "test@example.com"

	// Setup rate limiting by making multiple requests
	suite.mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Times(3)

	// Make 3 successful requests (should reach limit)
	for i := 0; i < 3; i++ {
		result, err := suite.authService.RequestAuthCode(context.Background(), email)
		assert.NoError(suite.T(), err)
		assert.NotNil(suite.T(), result)
	}

	// 4th request should be rate limited
	result, err := suite.authService.RequestAuthCode(context.Background(), email)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), ErrRateLimitExceeded, err)
}

// VerifyAuthCode tests
func (suite *AuthServiceTestSuite) TestVerifyAuthCode_Success() {
	email := "test@example.com"
	code := "123456"
	authCode := &models.AuthCode{
		ID:        "auth-id",
		UserID:    email,
		Code:      code,
		ExpiresAt: time.Now().Add(5 * time.Minute),
		UsedAt:    nil,
		CreatedAt: time.Now(),
	}

	expectedTokenPair := &TokenPair{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
	}

	suite.mockRepo.On("FindByUserID", mock.Anything, email).Return(authCode, nil)
	suite.mockRepo.On("MarkAsUsed", mock.Anything, authCode.ID).Return(nil)
	suite.mockJWT.On("GenerateTokenPair", mock.Anything, email, email).Return(expectedTokenPair, nil)

	result, err := suite.authService.VerifyAuthCode(context.Background(), email, code)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), email, result.UserID)
	assert.Equal(suite.T(), email, result.Email)
	assert.Equal(suite.T(), expectedTokenPair, result.TokenPair)
	assert.False(suite.T(), result.IsNewUser)
	assert.True(suite.T(), result.AuthenticatedAt.Before(time.Now().Add(time.Second)))
}

func (suite *AuthServiceTestSuite) TestVerifyAuthCode_EmptyEmail() {
	result, err := suite.authService.VerifyAuthCode(context.Background(), "", "123456")

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), ErrInvalidAuthCode, err)
}

func (suite *AuthServiceTestSuite) TestVerifyAuthCode_EmptyCode() {
	result, err := suite.authService.VerifyAuthCode(context.Background(), "test@example.com", "")

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), ErrInvalidAuthCode, err)
}

func (suite *AuthServiceTestSuite) TestVerifyAuthCode_InvalidEmail() {
	result, err := suite.authService.VerifyAuthCode(context.Background(), "invalid-email", "123456")

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), ErrEmailInvalid, err)
}

func (suite *AuthServiceTestSuite) TestVerifyAuthCode_CodeNotFound() {
	email := "test@example.com"
	code := "123456"

	suite.mockRepo.On("FindByUserID", mock.Anything, email).Return(nil, repository.ErrAuthCodeNotFound)

	result, err := suite.authService.VerifyAuthCode(context.Background(), email, code)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), ErrInvalidAuthCode, err)
}

func (suite *AuthServiceTestSuite) TestVerifyAuthCode_WrongCode() {
	email := "test@example.com"
	code := "123456"
	authCode := &models.AuthCode{
		ID:        "auth-id",
		UserID:    email,
		Code:      "654321", // Different code
		ExpiresAt: time.Now().Add(5 * time.Minute),
		UsedAt:    nil,
		CreatedAt: time.Now(),
	}

	suite.mockRepo.On("FindByUserID", mock.Anything, email).Return(authCode, nil)

	result, err := suite.authService.VerifyAuthCode(context.Background(), email, code)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), ErrInvalidAuthCode, err)
}

func (suite *AuthServiceTestSuite) TestVerifyAuthCode_ExpiredCode() {
	email := "test@example.com"
	code := "123456"
	authCode := &models.AuthCode{
		ID:        "auth-id",
		UserID:    email,
		Code:      code,
		ExpiresAt: time.Now().Add(-5 * time.Minute), // Expired
		UsedAt:    nil,
		CreatedAt: time.Now(),
	}

	suite.mockRepo.On("FindByUserID", mock.Anything, email).Return(authCode, nil)

	result, err := suite.authService.VerifyAuthCode(context.Background(), email, code)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), ErrCodeExpired, err)
}

func (suite *AuthServiceTestSuite) TestVerifyAuthCode_UsedCode() {
	email := "test@example.com"
	code := "123456"
	usedTime := time.Now().Add(-1 * time.Minute)
	authCode := &models.AuthCode{
		ID:        "auth-id",
		UserID:    email,
		Code:      code,
		ExpiresAt: time.Now().Add(5 * time.Minute),
		UsedAt:    &usedTime, // Already used
		CreatedAt: time.Now(),
	}

	suite.mockRepo.On("FindByUserID", mock.Anything, email).Return(authCode, nil)

	result, err := suite.authService.VerifyAuthCode(context.Background(), email, code)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), ErrCodeAlreadyUsed, err)
}

// RefreshTokens tests
func (suite *AuthServiceTestSuite) TestRefreshTokens_Success() {
	refreshToken := "refresh-token"
	expectedTokenPair := &TokenPair{
		AccessToken:  "new-access-token",
		RefreshToken: "new-refresh-token",
	}

	suite.mockJWT.On("RefreshTokens", mock.Anything, refreshToken).Return(expectedTokenPair, nil)

	result, err := suite.authService.RefreshTokens(context.Background(), refreshToken)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedTokenPair, result)
}

func (suite *AuthServiceTestSuite) TestRefreshTokens_EmptyToken() {
	result, err := suite.authService.RefreshTokens(context.Background(), "")

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), ErrInvalidToken, err)
}

func (suite *AuthServiceTestSuite) TestRefreshTokens_JWTServiceError() {
	refreshToken := "invalid-token"

	suite.mockJWT.On("RefreshTokens", mock.Anything, refreshToken).Return(nil, ErrInvalidToken)

	result, err := suite.authService.RefreshTokens(context.Background(), refreshToken)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), ErrInvalidToken, err)
}

// CleanExpiredCodes tests
func (suite *AuthServiceTestSuite) TestCleanExpiredCodes_Success() {
	suite.mockRepo.On("CleanExpired", mock.Anything).Return(nil)

	err := suite.authService.CleanExpiredCodes(context.Background())

	assert.NoError(suite.T(), err)
}

func (suite *AuthServiceTestSuite) TestCleanExpiredCodes_RepositoryError() {
	suite.mockRepo.On("CleanExpired", mock.Anything).Return(errors.New("database error"))

	err := suite.authService.CleanExpiredCodes(context.Background())

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "database error")
}

// Configuration tests
func (suite *AuthServiceTestSuite) TestNewAuthService_DefaultConfig() {
	authService := NewAuthService(suite.mockRepo, suite.mockJWT, nil)

	assert.NotNil(suite.T(), authService)
}

func (suite *AuthServiceTestSuite) TestNewAuthService_CustomConfig() {
	config := &AuthConfig{
		CodeLength:         8,
		CodeExpiryDuration: 10 * time.Minute,
		RateLimitWindow:    2 * time.Hour,
		MaxCodesPerEmail:   5,
		MaxRequestsPerIP:   10,
		CleanupInterval:    30 * time.Minute,
	}

	authService := NewAuthService(suite.mockRepo, suite.mockJWT, config)

	assert.NotNil(suite.T(), authService)
}

// Utility function tests
func TestIsValidEmail(t *testing.T) {
	testCases := []struct {
		name     string
		email    string
		expected bool
	}{
		{"Valid email", "test@example.com", true},
		{"Valid email with subdomain", "user@mail.example.com", true},
		{"Valid short email", "a@b.co", true},
		{"Empty email", "", false},
		{"Too short", "a@b", false},
		{"No @", "testexample.com", false},
		{"No domain", "test@", false},
		{"No local part", "@example.com", false},
		{"No dot in domain", "test@example", false},
		{"Dot at end", "test@example.", false},
		{"Multiple @", "test@@example.com", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := utils.IsValidEmail(tc.email)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestGenerateAuthCode(t *testing.T) {
	config := &AuthConfig{CodeLength: 6}
	service := &authService{config: config}

	// Test multiple generations
	for i := 0; i < 100; i++ {
		code, err := service.generateAuthCode()
		assert.NoError(t, err)
		assert.Len(t, code, 6)
		assert.Regexp(t, `^\d{6}$`, code)
	}
}

func TestGenerateAuthCode_DifferentLengths(t *testing.T) {
	testCases := []int{4, 6, 8, 10}

	for _, length := range testCases {
		t.Run(fmt.Sprintf("Length_%d", length), func(t *testing.T) {
			config := &AuthConfig{CodeLength: length}
			service := &authService{config: config}

			code, err := service.generateAuthCode()
			assert.NoError(t, err)
			assert.Len(t, code, length)

			// Verify it's all digits
			for _, char := range code {
				assert.True(t, char >= '0' && char <= '9')
			}
		})
	}
}

// Benchmark tests
func BenchmarkRequestAuthCode(b *testing.B) {
	mockRepo := new(MockAuthCodeRepository)
	mockJWT := new(MockJWTService)
	config := &AuthConfig{
		CodeLength:         6,
		CodeExpiryDuration: 5 * time.Minute,
		MaxCodesPerEmail:   1000, // High limit for benchmarking
	}
	service := NewAuthService(mockRepo, mockJWT, config)

	mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		email := fmt.Sprintf("test%d@example.com", i) // Different emails to avoid rate limiting
		_, _ = service.RequestAuthCode(context.Background(), email)
	}
}

func BenchmarkGenerateAuthCode(b *testing.B) {
	config := &AuthConfig{CodeLength: 6}
	service := &authService{config: config}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.generateAuthCode()
	}
}
