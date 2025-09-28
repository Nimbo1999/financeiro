package handler

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/nimbo1999/financeiro/authentication/internal/services"
)

type AuthHandler struct {
	authService services.AuthService
}

func NewAuthHandler(authService services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// RequestCodeHandler handles POST /auth/request-code
func (h *AuthHandler) RequestCodeHandler(w http.ResponseWriter, r *http.Request) {
	var req RequestCodeRequest

	// Validate request
	if !ValidateRequest(w, r, &req) {
		return
	}

	// Call auth service
	result, err := h.authService.RequestAuthCode(r.Context(), req.Email)
	if err != nil {
		log.Printf("RequestAuthCode error for email %s: %v", req.Email, err)
		statusCode, errorResponse := mapServiceError(err)
		writeErrorResponse(w, statusCode, errorResponse)
		return
	}

	// Build response
	response := &RequestCodeResponse{
		CodeID:    result.CodeID,
		ExpiresAt: result.ExpiresAt,
		Message:   "Authentication code sent successfully",
	}

	writeSuccessResponse(w, http.StatusOK,
		SuccessResponse(response, "Authentication code generated successfully"))
}

// VerifyCodeHandler handles POST /auth/verify-code
func (h *AuthHandler) VerifyCodeHandler(w http.ResponseWriter, r *http.Request) {
	var req VerifyCodeRequest

	// Validate request
	if !ValidateRequest(w, r, &req) {
		return
	}

	// Call auth service
	result, err := h.authService.VerifyAuthCode(r.Context(), req.Email, req.Code)
	if err != nil {
		log.Printf("VerifyAuthCode error for email %s: %v", req.Email, err)
		statusCode, errorResponse := mapServiceError(err)
		writeErrorResponse(w, statusCode, errorResponse)
		return
	}

	// Build response
	response := &VerifyCodeResponse{
		UserID:          result.UserID,
		Email:           result.Email,
		Tokens:          result.TokenPair,
		IsNewUser:       result.IsNewUser,
		AuthenticatedAt: result.AuthenticatedAt,
	}

	writeSuccessResponse(w, http.StatusOK,
		SuccessResponse(response, "Authentication successful"))
}

// RefreshTokenHandler handles POST /auth/refresh
func (h *AuthHandler) RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	var req RefreshTokenRequest

	// Validate request
	if !ValidateRequest(w, r, &req) {
		return
	}

	// Call auth service
	tokenPair, err := h.authService.RefreshTokens(r.Context(), req.RefreshToken)
	if err != nil {
		log.Printf("RefreshTokens error: %v", err)
		statusCode, errorResponse := mapServiceError(err)
		writeErrorResponse(w, statusCode, errorResponse)
		return
	}

	// Build response
	response := &RefreshTokenResponse{
		Tokens: tokenPair,
	}

	writeSuccessResponse(w, http.StatusOK,
		SuccessResponse(response, "Tokens refreshed successfully"))
}

func (h *AuthHandler) RegisterRoutes(r chi.Router) chi.Router {
	return r.Route("/auth", func(r chi.Router) {
		// Apply middleware to all auth routes
		r.Use(ContentTypeMiddleware)
		r.Use(ValidationMiddleware)
		r.Use(RateLimitMiddleware)

		// Public authentication endpoints
		r.Post("/request-code", h.RequestCodeHandler)
		r.Post("/verify-code", h.VerifyCodeHandler)
		r.Post("/refresh", h.RefreshTokenHandler)
	})
}
