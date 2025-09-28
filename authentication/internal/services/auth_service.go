package services

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/nimbo1999/financeiro/authentication/internal/models"
	"github.com/nimbo1999/financeiro/authentication/internal/repository"
	"github.com/nimbo1999/financeiro/authentication/internal/utils"
)

var (
	ErrEmailInvalid         = errors.New("email is invalid")
	ErrCodeExpired          = errors.New("auth code has expired")
	ErrCodeAlreadyUsed      = errors.New("auth code has already been used")
	ErrRateLimitExceeded    = errors.New("rate limit exceeded")
	ErrCodeGenerationFailed = errors.New("failed to generate auth code")
	ErrUserNotFound         = errors.New("user not found")
	ErrInvalidAuthCode      = errors.New("invalid auth code")
)

type AuthService interface {
	// RequestAuthCode generates and stores a new auth code for the given email
	RequestAuthCode(ctx context.Context, email string) (*AuthCodeResult, error)

	// VerifyAuthCode validates an auth code and returns JWT tokens if valid
	VerifyAuthCode(ctx context.Context, email, code string) (*AuthResult, error)

	// RefreshTokens validates a refresh token and generates new token pair
	RefreshTokens(ctx context.Context, refreshToken string) (*TokenPair, error)

	// CleanExpiredCodes removes all expired auth codes from storage
	CleanExpiredCodes(ctx context.Context) error
}

type AuthCodeResult struct {
	CodeID    string    `json:"code_id"`
	ExpiresAt time.Time `json:"expires_at"`
	Success   bool      `json:"success"`
}

type AuthResult struct {
	UserID          string     `json:"user_id"`
	Email           string     `json:"email"`
	TokenPair       *TokenPair `json:"tokens"`
	IsNewUser       bool       `json:"is_new_user"`
	AuthenticatedAt time.Time  `json:"authenticated_at"`
}

type AuthConfig struct {
	CodeLength         int           // Length of auth code (default: 6)
	CodeExpiryDuration time.Duration // How long codes are valid (default: 5 minutes)
	RateLimitWindow    time.Duration // Rate limiting window (default: 1 hour)
	MaxCodesPerEmail   int           // Max codes per email in window (default: 3)
	MaxRequestsPerIP   int           // Max requests per IP in window (default: 5)
	CleanupInterval    time.Duration // How often to clean expired codes (default: 1 hour)
}

type authService struct {
	authRepo   repository.AuthCodeRepository
	jwtService JWTService
	config     *AuthConfig
	/* @todo: refactor the rateLimiter to store the information in redis */
	rateLimiter map[string]*rateLimitEntry // Simple in-memory rate limiter
}

type rateLimitEntry struct {
	count     int
	resetTime time.Time
}

func NewAuthService(authRepo repository.AuthCodeRepository, jwtService JWTService, config *AuthConfig) AuthService {
	if config == nil {
		config = &AuthConfig{}
	}

	// Set defaults
	if config.CodeLength == 0 {
		config.CodeLength = 6
	}
	if config.CodeExpiryDuration == 0 {
		config.CodeExpiryDuration = 5 * time.Minute
	}
	if config.RateLimitWindow == 0 {
		config.RateLimitWindow = 1 * time.Hour
	}
	if config.MaxCodesPerEmail == 0 {
		config.MaxCodesPerEmail = 3
	}
	if config.MaxRequestsPerIP == 0 {
		config.MaxRequestsPerIP = 5
	}
	if config.CleanupInterval == 0 {
		config.CleanupInterval = 1 * time.Hour
	}

	return &authService{
		authRepo:    authRepo,
		jwtService:  jwtService,
		config:      config,
		rateLimiter: make(map[string]*rateLimitEntry),
	}
}

func (s *authService) RequestAuthCode(ctx context.Context, email string) (*AuthCodeResult, error) {
	if email == "" {
		return nil, ErrEmailInvalid
	}

	// Normalize email
	email = strings.ToLower(strings.TrimSpace(email))

	// Basic email validation
	if !utils.IsValidEmail(email) {
		return nil, ErrEmailInvalid
	}

	// Check rate limiting for email
	if err := s.checkRateLimit(email); err != nil {
		return nil, err
	}

	// Generate auth code
	code, err := s.generateAuthCode()
	if err != nil {
		return nil, ErrCodeGenerationFailed
	}

	now := time.Now()

	// Create auth code record
	authCode := &models.AuthCode{
		UserID:    email, // Using email as user ID for now
		Code:      code,
		ExpiresAt: now.Add(s.config.CodeExpiryDuration),
		CreatedAt: now,
	}

	// Store in repository
	if err := s.authRepo.Create(ctx, authCode); err != nil {
		return nil, fmt.Errorf("failed to store auth code: %w", err)
	}

	// Update rate limiter
	s.updateRateLimit(email)

	return &AuthCodeResult{
		CodeID:    authCode.ID,
		ExpiresAt: authCode.ExpiresAt,
		Success:   true,
	}, nil
}

func (s *authService) VerifyAuthCode(ctx context.Context, email, code string) (*AuthResult, error) {
	if email == "" || code == "" {
		return nil, ErrInvalidAuthCode
	}

	// Normalize email
	email = strings.ToLower(strings.TrimSpace(email))

	// Basic email validation
	if !utils.IsValidEmail(email) {
		return nil, ErrEmailInvalid
	}

	// @todo: retrieve the user from user service by email to use the user ID instead of the email. For now, we will keep using the email as user ID.

	// Find the most recent auth code for this email
	authCode, err := s.authRepo.FindByUserID(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrAuthCodeNotFound) {
			return nil, ErrInvalidAuthCode
		}
		return nil, fmt.Errorf("failed to find auth code: %w", err)
	}

	// Validate the code matches
	if authCode.Code != code {
		return nil, ErrInvalidAuthCode
	}

	// Check if code is expired
	if time.Now().After(authCode.ExpiresAt) {
		return nil, ErrCodeExpired
	}

	// Check if code is already used
	if authCode.UsedAt != nil {
		return nil, ErrCodeAlreadyUsed
	}

	// Mark code as used
	if err := s.authRepo.MarkAsUsed(ctx, authCode.ID); err != nil {
		return nil, fmt.Errorf("failed to mark code as used: %w", err)
	}

	// @todo: again, for now, we are using email as user ID. In the future, integrate with user service to get actual user ID and check the user status before anything.
	// Generate JWT tokens
	tokenPair, err := s.jwtService.GenerateTokenPair(ctx, email, email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return &AuthResult{
		UserID:          email,
		Email:           email,
		TokenPair:       tokenPair,
		IsNewUser:       false, // @todo: For now, treat all as existing users
		AuthenticatedAt: time.Now(),
	}, nil
}

func (s *authService) RefreshTokens(ctx context.Context, refreshToken string) (*TokenPair, error) {
	if refreshToken == "" {
		return nil, ErrInvalidToken
	}

	// Use JWT service to refresh tokens
	tokenPair, err := s.jwtService.RefreshTokens(ctx, refreshToken)
	if err != nil {
		return nil, err
	}

	return tokenPair, nil
}

func (s *authService) CleanExpiredCodes(ctx context.Context) error {
	return s.authRepo.CleanExpired(ctx)
}

// generateAuthCode creates a cryptographically secure random 6-digit code
func (s *authService) generateAuthCode() (string, error) {
	// Calculate the maximum value for the given code length
	max := int64(1)
	for i := 0; i < s.config.CodeLength; i++ {
		max *= 10
	}
	max -= 1 // e.g., for 6 digits: 999999

	// Generate random number
	n, err := rand.Int(rand.Reader, big.NewInt(max+1))
	if err != nil {
		return "", err
	}

	// Format with leading zeros
	format := fmt.Sprintf("%%0%dd", s.config.CodeLength)
	return fmt.Sprintf(format, n.Int64()), nil
}

// checkRateLimit verifies if the email has exceeded rate limits
func (s *authService) checkRateLimit(email string) error {
	now := time.Now()
	entry, exists := s.rateLimiter[email]

	if !exists {
		// First request for this email
		return nil
	}

	// Check if rate limit window has expired
	if now.After(entry.resetTime) {
		// Reset the counter
		delete(s.rateLimiter, email)
		return nil
	}

	// Check if limit exceeded
	if entry.count >= s.config.MaxCodesPerEmail {
		return ErrRateLimitExceeded
	}

	return nil
}

// updateRateLimit updates the rate limiter for an email
func (s *authService) updateRateLimit(email string) {
	now := time.Now()
	entry, exists := s.rateLimiter[email]

	if !exists || now.After(entry.resetTime) {
		// Create new entry or reset expired entry
		s.rateLimiter[email] = &rateLimitEntry{
			count:     1,
			resetTime: now.Add(s.config.RateLimitWindow),
		}
	} else {
		// Increment existing entry
		entry.count++
	}
}
