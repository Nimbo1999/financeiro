package middleware

import (
	"net/http"

	"github.com/go-chi/cors"
)

// CORSConfig holds CORS configuration options
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

// DefaultCORSConfig returns a default CORS configuration
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodHead,
			http.MethodOptions,
		},
		AllowedHeaders: []string{
			"Accept",
			"Accept-Encoding",
			"Accept-Language",
			"User-Agent",
			"Authorization",
			"Content-Type",
			"X-Request-ID",
			"Host",
			"Origin",
		},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           3600,
	}
}

// CORSMiddleware provides CORS support
type CORSMiddleware struct {
	config CORSConfig
}

// NewCORSMiddleware creates a new CORS middleware with the given configuration
func NewCORSMiddleware(config CORSConfig) *CORSMiddleware {
	return &CORSMiddleware{
		config: config,
	}
}

// Handler returns the CORS middleware handler
func (c *CORSMiddleware) Handler(next http.Handler) http.Handler {
	return cors.New(cors.Options{
		AllowedOrigins:     c.config.AllowedOrigins,
		AllowedMethods:     c.config.AllowedMethods,
		AllowedHeaders:     c.config.AllowedHeaders,
		ExposedHeaders:     c.config.ExposedHeaders,
		AllowCredentials:   c.config.AllowCredentials,
		OptionsPassthrough: false,
		MaxAge:             c.config.MaxAge,
	}).Handler(next)
}
