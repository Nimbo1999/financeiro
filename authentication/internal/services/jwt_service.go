package services

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrExpiredToken     = errors.New("token expired")
	ErrInvalidTokenType = errors.New("invalid token type")
)

type TokenType string

const (
	AccessTokenType  TokenType = "access"
	RefreshTokenType TokenType = "refresh"
)

type Claims struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	TokenType TokenType `json:"token_type"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type UserContext struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
}

type JWTService interface {
	// GenerateTokenPair creates both access and refresh tokens for a user
	GenerateTokenPair(ctx context.Context, userID, email string) (*TokenPair, error)

	// ValidateAccessToken validates an access token and returns user context
	ValidateAccessToken(ctx context.Context, tokenString string) (*UserContext, error)

	// ValidateRefreshToken validates a refresh token and returns user context
	ValidateRefreshToken(ctx context.Context, tokenString string) (*UserContext, error)

	// RefreshTokens validates a refresh token and generates a new token pair
	RefreshTokens(ctx context.Context, refreshToken string) (*TokenPair, error)

	// GetPublicKey returns the public key for token validation
	GetPublicKey() *rsa.PublicKey
}

type JWTConfig struct {
	PrivateKey           *rsa.PrivateKey
	PublicKey            *rsa.PublicKey
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
	Issuer               string
}

type jwtService struct {
	config *JWTConfig
}

func NewJWTService(config *JWTConfig) JWTService {
	if config.AccessTokenDuration == 0 {
		config.AccessTokenDuration = 1 * time.Hour
	}
	if config.RefreshTokenDuration == 0 {
		config.RefreshTokenDuration = 7 * 24 * time.Hour
	}
	if config.Issuer == "" {
		config.Issuer = "authentication-service"
	}

	return &jwtService{
		config: config,
	}
}

func (s *jwtService) GenerateTokenPair(ctx context.Context, userID, email string) (*TokenPair, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}
	if email == "" {
		return nil, errors.New("email cannot be empty")
	}

	now := time.Now()
	jti := uuid.New().String()

	accessTokenString, err := s.generateToken(userID, email, jti, AccessTokenType, now)
	if err != nil {
		return nil, err
	}

	refreshTokenString, err := s.generateToken(userID, email, jti, RefreshTokenType, now)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
	}, nil
}

func (s *jwtService) generateToken(userID, email, jti string, tokenType TokenType, now time.Time) (string, error) {
	tokenExpireDuration := s.config.AccessTokenDuration
	if tokenType == RefreshTokenType {
		tokenExpireDuration = s.config.RefreshTokenDuration
	}

	// Generate token
	claims := &Claims{
		UserID:    userID,
		Email:     email,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        fmt.Sprintf("%s-%s", jti, tokenType),
			Issuer:    s.config.Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(tokenExpireDuration)),
			NotBefore: jwt.NewNumericDate(now),
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(s.config.PrivateKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (s *jwtService) ValidateAccessToken(ctx context.Context, tokenString string) (*UserContext, error) {
	claims, err := s.validateToken(tokenString, AccessTokenType)
	if err != nil {
		return nil, err
	}

	return &UserContext{
		UserID: claims.UserID,
		Email:  claims.Email,
	}, nil
}

func (s *jwtService) ValidateRefreshToken(ctx context.Context, tokenString string) (*UserContext, error) {
	claims, err := s.validateToken(tokenString, RefreshTokenType)
	if err != nil {
		return nil, err
	}

	return &UserContext{
		UserID: claims.UserID,
		Email:  claims.Email,
	}, nil
}

func (s *jwtService) RefreshTokens(ctx context.Context, refreshToken string) (*TokenPair, error) {
	userContext, err := s.ValidateRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, err
	}

	return s.GenerateTokenPair(ctx, userContext.UserID, userContext.Email)
}

func (s *jwtService) GetPublicKey() *rsa.PublicKey {
	return s.config.PublicKey
}

func (s *jwtService) validateToken(tokenString string, expectedType TokenType) (*Claims, error) {
	if tokenString == "" {
		return nil, ErrInvalidToken
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		// Verify the signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.config.PublicKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	// Verify token type
	if claims.TokenType != expectedType {
		return nil, ErrInvalidTokenType
	}

	// Additional validation
	if claims.UserID == "" || claims.Email == "" {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func LoadRSAPrivateKey(key []byte) (*rsa.PrivateKey, error) {
	return jwt.ParseRSAPrivateKeyFromPEM(key)
}

func LoadRSAPublicKey(key []byte) (*rsa.PublicKey, error) {
	return jwt.ParseRSAPublicKeyFromPEM(key)
}
