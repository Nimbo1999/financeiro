package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/go-chi/cors"
	"github.com/nimbo1999/financeiro/authentication/internal/services"
)

// ContentTypeMiddleware ensures the request content type is application/json
func ContentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only check content type for POST requests
		if r.Method == http.MethodPost {
			contentType := r.Header.Get("Content-Type")
			if !strings.HasPrefix(contentType, "application/json") {
				writeErrorResponse(w, http.StatusUnsupportedMediaType,
					ErrorResponse("INVALID_CONTENT_TYPE", "Content-Type must be application/json", ""))
				return
			}
		}

		// Set response content type
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// RecoveryMiddleware handles panics and returns a proper error response
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v\n%s", err, debug.Stack())

				w.Header().Set("Content-Type", "application/json")
				writeErrorResponse(w, http.StatusInternalServerError,
					ErrorResponse("INTERNAL_ERROR", "An unexpected error occurred", ""))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// LoggingMiddleware logs HTTP requests
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		log.Printf("%s %s %d %v", r.Method, r.URL.Path, wrapped.statusCode, duration)
	})
}

// ValidationMiddleware provides input validation for requests
func ValidationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add request size limit
		r.Body = http.MaxBytesReader(w, r.Body, 1024*1024) // 1MB limit

		next.ServeHTTP(w, r)
	})
}

// RateLimitMiddleware implements basic rate limiting (simplified version)
func RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This is a placeholder - in production, you'd use Redis or a proper rate limiter
		// For now, we'll let the AuthService handle rate limiting
		next.ServeHTTP(w, r)
	})
}

// SecurityHeadersMiddleware adds security headers
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		next.ServeHTTP(w, r)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Helper functions for error handling
func writeErrorResponse(w http.ResponseWriter, statusCode int, response *APIResponse) {
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode error response: %v", err)
	}
}

func writeSuccessResponse(w http.ResponseWriter, statusCode int, response *APIResponse) {
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode success response: %v", err)
		writeErrorResponse(w, http.StatusInternalServerError,
			ErrorResponse("ENCODING_ERROR", "Failed to encode response", ""))
	}
}

// Service error mapping
func mapServiceError(err error) (int, *APIResponse) {
	switch err {
	case services.ErrEmailInvalid:
		return http.StatusBadRequest, ErrorResponse("INVALID_EMAIL", "Email address is invalid", "")
	case services.ErrInvalidAuthCode:
		return http.StatusBadRequest, ErrorResponse("INVALID_CODE", "Invalid authentication code", "")
	case services.ErrCodeExpired:
		return http.StatusBadRequest, ErrorResponse("CODE_EXPIRED", "Authentication code has expired", "")
	case services.ErrCodeAlreadyUsed:
		return http.StatusBadRequest, ErrorResponse("CODE_USED", "Authentication code has already been used", "")
	case services.ErrRateLimitExceeded:
		return http.StatusTooManyRequests, ErrorResponse("RATE_LIMIT_EXCEEDED", "Too many requests", "Please wait before requesting another code")
	case services.ErrInvalidToken:
		return http.StatusUnauthorized, ErrorResponse("INVALID_TOKEN", "Invalid or expired token", "")
	case services.ErrExpiredToken:
		return http.StatusUnauthorized, ErrorResponse("TOKEN_EXPIRED", "Token has expired", "")
	case services.ErrUserNotFound:
		return http.StatusNotFound, ErrorResponse("USER_NOT_FOUND", "User not found", "")
	default:
		// For unknown errors, log them but don't expose details
		log.Printf("Unmapped service error: %v", err)
		return http.StatusInternalServerError, ErrorResponse("INTERNAL_ERROR", "An error occurred", "")
	}
}

var allowedHeaders []string = []string{
	"User-Agent",
	"Content-Type",
	"Accept",
	"Accept-Encoding",
	"Accept-Language",
	"Cache-Control",
	"Connection",
	"DNT",
	"Host",
	"Origin",
	"Pragma",
	"Referer",
	"Authorization",
	"X-User-Id",
	"X-User-Email",
}

var allowedMethods []string = []string{
	http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch, http.MethodOptions, http.MethodHead,
}

func CorsMiddleware() func(next http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   allowedMethods,
		AllowedHeaders:   allowedHeaders,
		AllowCredentials: true,
	})
}
