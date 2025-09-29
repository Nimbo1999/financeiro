package handler

import (
	"context"
	"log"

	"github.com/nimbo1999/financeiro/authentication/internal/services"
	authpb "github.com/nimbo1999/financeiro/authentication/pkg/grpc/auth/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthGRPCHandler struct {
	authpb.UnimplementedAuthServiceServer
	jwtService services.JWTService
}

func NewAuthGRPCHandler(jwtService services.JWTService) *AuthGRPCHandler {
	return &AuthGRPCHandler{
		jwtService: jwtService,
	}
}

func (h *AuthGRPCHandler) ValidateToken(ctx context.Context, req *authpb.ValidateTokenRequest) (*authpb.ValidateTokenResponse, error) {
	// Validate request
	if err := ValidateTokenRequest(req); err != nil {
		return &authpb.ValidateTokenResponse{
			Valid:        false,
			ErrorMessage: err.Error(),
		}, nil
	}

	// Validate the access token using JWT service
	userContext, err := h.jwtService.ValidateAccessToken(ctx, req.Token)
	if err != nil {
		log.Printf("Token validation failed: %v", err)

		// Return validation result with error message
		return &authpb.ValidateTokenResponse{
			Valid:        false,
			ErrorMessage: mapJWTErrorToMessage(err),
		}, nil
	}

	// Token is valid, return user context
	return &authpb.ValidateTokenResponse{
		Valid: true,
		UserContext: &authpb.UserContext{
			UserId: userContext.UserID,
			Email:  userContext.Email,
		},
	}, nil
}

func (h *AuthGRPCHandler) HealthCheck(ctx context.Context, req *authpb.HealthCheckRequest) (*authpb.HealthCheckResponse, error) {
	// Perform basic health check
	// Since this service is stateless and only validates tokens,
	// we can check if the JWT service is properly configured
	if h.jwtService == nil || h.jwtService.GetPublicKey() == nil {
		log.Printf("Health check failed: JWT service not properly configured")
		return &authpb.HealthCheckResponse{
			Status:  authpb.HealthCheckResponse_NOT_SERVING,
			Message: "Service unavailable - JWT configuration issues",
		}, nil
	}

	return &authpb.HealthCheckResponse{
		Status:  authpb.HealthCheckResponse_SERVING,
		Message: "Authentication service is healthy",
	}, nil
}

// ValidateTokenRequest validates the ValidateToken request
func ValidateTokenRequest(req *authpb.ValidateTokenRequest) error {
	if req == nil {
		return status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	if req.Token == "" {
		return status.Error(codes.InvalidArgument, "token is required")
	}

	return nil
}

// mapJWTErrorToMessage maps JWT service errors to user-friendly messages
func mapJWTErrorToMessage(err error) string {
	switch err {
	case services.ErrInvalidToken:
		return "invalid token"
	case services.ErrExpiredToken:
		return "token expired"
	case services.ErrInvalidTokenType:
		return "invalid token type"
	default:
		return "token validation failed"
	}
}