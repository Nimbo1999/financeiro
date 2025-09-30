package middleware

import (
	"context"
	"net/http"
	"strings"
)

// AuthMiddleware handles authentication for protected routes
type AuthMiddleware struct {
	tokenValidator TokenValidator
	publicPaths    map[string]bool
}

// TokenValidator defines the interface for validating JWT tokens
type TokenValidator interface {
	ValidateToken(ctx context.Context, token string) (*UserContext, error)
}

// UserContext contains user information extracted from the token
type UserContext struct {
	UserID string
	Email  string
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(validator TokenValidator, publicPaths []string) *AuthMiddleware {
	pathMap := make(map[string]bool)
	for _, path := range publicPaths {
		pathMap[path] = true
	}

	return &AuthMiddleware{
		tokenValidator: validator,
		publicPaths:    pathMap,
	}
}

// Handler returns an HTTP middleware that validates JWT tokens
func (am *AuthMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the path is public
		if am.isPublicPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Extract token from Authorization header
		token, err := am.extractToken(r)
		if err != nil {
			http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
			return
		}

		// Validate token
		userContext, err := am.tokenValidator.ValidateToken(r.Context(), token)
		if err != nil {
			http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
			return
		}

		// Inject user context headers
		r.Header.Set("X-User-ID", userContext.UserID)
		r.Header.Set("X-User-Email", userContext.Email)

		// Call next handler
		next.ServeHTTP(w, r)
	})
}

// extractToken extracts the JWT token from the Authorization header
func (am *AuthMiddleware) extractToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", ErrMissingAuthHeader
	}

	// Check if it's a Bearer token
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", ErrInvalidAuthHeader
	}

	return parts[1], nil
}

// isPublicPath checks if a path should bypass authentication
func (am *AuthMiddleware) isPublicPath(path string) bool {
	// Check exact match
	if am.publicPaths[path] {
		return true
	}

	// Check prefix match for public patterns
	for publicPath := range am.publicPaths {
		if strings.HasSuffix(publicPath, "/*") {
			prefix := strings.TrimSuffix(publicPath, "/*")
			if strings.HasPrefix(path, prefix) {
				return true
			}
		}
	}

	return false
}

// Error definitions
var (
	ErrMissingAuthHeader = &AuthError{Message: "missing authorization header"}
	ErrInvalidAuthHeader = &AuthError{Message: "invalid authorization header format"}
)

// AuthError represents an authentication error
type AuthError struct {
	Message string
}

func (e *AuthError) Error() string {
	return e.Message
}