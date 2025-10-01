package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type CORSMiddlewareTestSuite struct {
	suite.Suite
	middleware *CORSMiddleware
}

func TestCORSMiddlewareTestSuite(t *testing.T) {
	suite.Run(t, new(CORSMiddlewareTestSuite))
}

func (s *CORSMiddlewareTestSuite) SetupTest() {
	config := CORSConfig{
		AllowedOrigins:   []string{"http://localhost:3000", "https://example.com"},
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodOptions},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           3600,
	}
	s.middleware = NewCORSMiddleware(config)
}

func (s *CORSMiddlewareTestSuite) TestCORSMiddleware_AllowedOrigin() {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rec := httptest.NewRecorder()

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Act
	handler := s.middleware.Handler(nextHandler)
	handler.ServeHTTP(rec, req)

	// Assert
	assert.Contains(s.T(), rec.Header().Values("Access-Control-Allow-Origin"), "http://localhost:3000")
	assert.Equal(s.T(), strings.ToLower(rec.Header().Get("Access-Control-Expose-Headers")), "x-request-id")
	assert.Equal(s.T(), "true", rec.Header().Get("Access-Control-Allow-Credentials"))
	assert.Equal(s.T(), http.StatusOK, rec.Code)
}

func (s *CORSMiddlewareTestSuite) TestCORSMiddleware_DisallowedOrigin() {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://malicious.com")
	rec := httptest.NewRecorder()

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Act
	handler := s.middleware.Handler(nextHandler)
	handler.ServeHTTP(rec, req)

	// Assert
	assert.Empty(s.T(), rec.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(s.T(), http.StatusOK, rec.Code)
}

func (s *CORSMiddlewareTestSuite) TestCORSMiddleware_PreflightRequest() {
	// Arrange
	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	rec := httptest.NewRecorder()

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("preflight success"))
	})

	// Act
	handler := s.middleware.Handler(nextHandler)
	handler.ServeHTTP(rec, req)

	// Assert
	assert.Equal(s.T(), "https://example.com", rec.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(s.T(), http.StatusOK, rec.Code)
}

func (s *CORSMiddlewareTestSuite) TestCORSMiddleware_WildcardOrigin() {
	// Arrange
	config := CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{http.MethodGet},
	}
	middleware := NewCORSMiddleware(config)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://any-domain.com")
	rec := httptest.NewRecorder()

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Act
	handler := middleware.Handler(nextHandler)
	handler.ServeHTTP(rec, req)

	// Assert
	assert.Equal(s.T(), "*", rec.Header().Get("Access-Control-Allow-Origin"))
}

func (s *CORSMiddlewareTestSuite) TestDefaultCORSConfig() {
	// Act
	config := DefaultCORSConfig()

	// Assert
	assert.Contains(s.T(), config.AllowedOrigins, "*")
	assert.Contains(s.T(), config.AllowedMethods, http.MethodGet)
	assert.Contains(s.T(), config.AllowedMethods, http.MethodPost)
	assert.Contains(s.T(), config.AllowedHeaders, "Authorization", "Content-Type")
	assert.True(s.T(), config.AllowCredentials)
	assert.Equal(s.T(), 3600, config.MaxAge)
}
