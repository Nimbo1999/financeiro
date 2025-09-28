package handler

import (
	"time"

	"github.com/nimbo1999/financeiro/authentication/internal/services"
)

// Request DTOs
type RequestCodeRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type VerifyCodeRequest struct {
	Email string `json:"email" validate:"required,email"`
	Code  string `json:"code" validate:"required,len=6,numeric"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// Response DTOs
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

type RequestCodeResponse struct {
	CodeID    string    `json:"code_id"`
	ExpiresAt time.Time `json:"expires_at"`
	Message   string    `json:"message"`
}

type VerifyCodeResponse struct {
	UserID          string                `json:"user_id"`
	Email           string                `json:"email"`
	Tokens          *services.TokenPair   `json:"tokens"`
	IsNewUser       bool                  `json:"is_new_user"`
	AuthenticatedAt time.Time             `json:"authenticated_at"`
}

type RefreshTokenResponse struct {
	Tokens *services.TokenPair `json:"tokens"`
}

// Internal DTOs for validation
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Helper functions to create standard responses
func SuccessResponse(data interface{}, message string) *APIResponse {
	return &APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
}

func ErrorResponse(code, message, details string) *APIResponse {
	return &APIResponse{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
}

func ValidationErrorResponse(errors []ValidationError) *APIResponse {
	return &APIResponse{
		Success: false,
		Error: &ErrorInfo{
			Code:    "VALIDATION_ERROR",
			Message: "Invalid input data",
			Details: "Please check the request format and try again",
		},
		Data: map[string]interface{}{
			"validation_errors": errors,
		},
	}
}