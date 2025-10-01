package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type CircuitBreakerTestSuite struct {
	suite.Suite
	circuitBreaker *CircuitBreaker
}

func TestCircuitBreakerTestSuite(t *testing.T) {
	suite.Run(t, new(CircuitBreakerTestSuite))
}

func (s *CircuitBreakerTestSuite) SetupTest() {
	config := CircuitBreakerConfig{
		MaxFailures:  3,
		Timeout:      30 * time.Second,
		ResetTimeout: 100 * time.Millisecond, // Short timeout for testing
	}
	s.circuitBreaker = NewCircuitBreaker(config)
}

func (s *CircuitBreakerTestSuite) TestCircuitBreaker_InitialStateClosed() {
	// Assert
	assert.Equal(s.T(), StateClosed, s.circuitBreaker.GetState())
}

func (s *CircuitBreakerTestSuite) TestCircuitBreaker_AllowsSuccessfulRequests() {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Act
	handler := s.circuitBreaker.Handler(nextHandler)
	handler.ServeHTTP(rec, req)

	// Assert
	assert.Equal(s.T(), http.StatusOK, rec.Code)
	assert.Equal(s.T(), StateClosed, s.circuitBreaker.GetState())
}

func (s *CircuitBreakerTestSuite) TestCircuitBreaker_OpensAfterMaxFailures() {
	// Arrange
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	handler := s.circuitBreaker.Handler(nextHandler)

	// Act - Make failing requests
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}

	// Assert
	assert.Equal(s.T(), StateOpen, s.circuitBreaker.GetState())
}

func (s *CircuitBreakerTestSuite) TestCircuitBreaker_RejectsRequestsWhenOpen() {
	// Arrange
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	handler := s.circuitBreaker.Handler(nextHandler)

	// Act - Make failing requests to open circuit
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}

	// Try another request when circuit is open
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// Assert
	assert.Equal(s.T(), http.StatusServiceUnavailable, rec.Code)
	assert.Contains(s.T(), rec.Body.String(), "Service temporarily unavailable")
}

func (s *CircuitBreakerTestSuite) TestCircuitBreaker_TransitionsToHalfOpenAfterTimeout() {
	// Arrange
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	handler := s.circuitBreaker.Handler(nextHandler)

	// Act - Open the circuit
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}

	assert.Equal(s.T(), StateOpen, s.circuitBreaker.GetState())

	// Wait for reset timeout
	time.Sleep(150 * time.Millisecond)

	// Make a request - should transition to half-open
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	// Change handler to return success
	successHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler = s.circuitBreaker.Handler(successHandler)
	handler.ServeHTTP(rec, req)

	// Assert
	assert.Equal(s.T(), http.StatusOK, rec.Code)
	assert.Equal(s.T(), StateClosed, s.circuitBreaker.GetState())
}

func (s *CircuitBreakerTestSuite) TestCircuitBreaker_ResetsOnSuccessInHalfOpen() {
	// Arrange
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	handler := s.circuitBreaker.Handler(nextHandler)

	// Act - Open the circuit
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}

	// Wait for reset timeout
	time.Sleep(150 * time.Millisecond)

	// Success request in half-open state
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	successHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler = s.circuitBreaker.Handler(successHandler)
	handler.ServeHTTP(rec, req)

	// Assert
	assert.Equal(s.T(), StateClosed, s.circuitBreaker.GetState())
}

func (s *CircuitBreakerTestSuite) TestCircuitBreaker_ClientErrorsDoNotTrigger() {
	// Arrange
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest) // 4xx errors should not trigger circuit
	})

	handler := s.circuitBreaker.Handler(nextHandler)

	// Act - Make requests with client errors
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}

	// Assert
	assert.Equal(s.T(), StateClosed, s.circuitBreaker.GetState())
}

func (s *CircuitBreakerTestSuite) TestCircuitBreaker_Reset() {
	// Arrange
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	handler := s.circuitBreaker.Handler(nextHandler)

	// Act - Open the circuit
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}

	assert.Equal(s.T(), StateOpen, s.circuitBreaker.GetState())

	// Reset the circuit
	s.circuitBreaker.Reset()

	// Assert
	assert.Equal(s.T(), StateClosed, s.circuitBreaker.GetState())
}

func (s *CircuitBreakerTestSuite) TestDefaultCircuitBreakerConfig() {
	// Act
	config := DefaultCircuitBreakerConfig()

	// Assert
	assert.Equal(s.T(), 5, config.MaxFailures)
	assert.Equal(s.T(), 30*time.Second, config.Timeout)
	assert.Equal(s.T(), 60*time.Second, config.ResetTimeout)
}

func (s *CircuitBreakerTestSuite) TestCircuitResponseWriter_WriteHeader() {
	// Arrange
	rec := httptest.NewRecorder()
	crw := &circuitResponseWriter{ResponseWriter: rec, statusCode: http.StatusOK}

	// Act
	crw.WriteHeader(http.StatusCreated)

	// Assert
	assert.Equal(s.T(), http.StatusCreated, crw.statusCode)
	assert.True(s.T(), crw.written)
	assert.Equal(s.T(), http.StatusCreated, rec.Code)
}

func (s *CircuitBreakerTestSuite) TestCircuitResponseWriter_Write() {
	// Arrange
	rec := httptest.NewRecorder()
	crw := &circuitResponseWriter{ResponseWriter: rec, statusCode: http.StatusOK}

	// Act
	n, err := crw.Write([]byte("test"))

	// Assert
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 4, n)
	assert.True(s.T(), crw.written)
	assert.Equal(s.T(), "test", rec.Body.String())
}
