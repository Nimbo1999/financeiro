package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"
)

var (
	ErrCircuitOpen = errors.New("circuit breaker is open")
)

// CircuitBreakerConfig holds circuit breaker configuration
type CircuitBreakerConfig struct {
	MaxFailures  int
	Timeout      time.Duration
	ResetTimeout time.Duration
}

// DefaultCircuitBreakerConfig returns a default circuit breaker configuration
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		MaxFailures:  5,
		Timeout:      30 * time.Second,
		ResetTimeout: 60 * time.Second,
	}
}

// CircuitState represents the state of the circuit breaker
type CircuitState int

const (
	StateClosed CircuitState = iota
	StateOpen
	StateHalfOpen
)

// CircuitBreaker implements the circuit breaker pattern for resilience
type CircuitBreaker struct {
	config       CircuitBreakerConfig
	state        CircuitState
	failures     int
	lastFailTime time.Time
	mu           sync.RWMutex
}

// NewCircuitBreaker creates a new circuit breaker with the given configuration
func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker {
	return &CircuitBreaker{
		config: config,
		state:  StateClosed,
	}
}

// Handler returns the circuit breaker middleware handler
func (cb *CircuitBreaker) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if request is allowed
		if err := cb.beforeRequest(); err != nil {
			beforeRequestErr := fmt.Errorf("Service temporarily unavailable: %w", err)
			http.Error(w, beforeRequestErr.Error(), http.StatusServiceUnavailable)
			return
		}

		// Wrap response writer to capture status code
		crw := &circuitResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Execute the request
		next.ServeHTTP(crw, r)

		// Record the result
		cb.afterRequest(crw.statusCode)
	})
}

// beforeRequest checks if the circuit breaker allows the request
func (cb *CircuitBreaker) beforeRequest() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()

	switch cb.state {
	case StateClosed:
		// Circuit is closed, allow request
		return nil

	case StateOpen:
		// Check if we should transition to half-open
		if now.Sub(cb.lastFailTime) > cb.config.ResetTimeout {
			cb.state = StateHalfOpen
			cb.failures = 0
			return nil
		}
		return ErrCircuitOpen

	case StateHalfOpen:
		// Allow a single request to test the service
		return nil

	default:
		return ErrCircuitOpen
	}
}

// afterRequest records the result of the request
func (cb *CircuitBreaker) afterRequest(statusCode int) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Determine if the request was successful
	success := statusCode >= 200 && statusCode < 500

	if !success {
		cb.failures++
		cb.lastFailTime = time.Now()

		// Check if we should open the circuit
		if cb.failures >= cb.config.MaxFailures {
			cb.state = StateOpen
		}
	} else {
		// Success - reset failures
		if cb.state == StateHalfOpen {
			// Transition back to closed state
			cb.state = StateClosed
		}
		cb.failures = 0
	}
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Reset resets the circuit breaker to its initial state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.state = StateClosed
	cb.failures = 0
}

// circuitResponseWriter wraps http.ResponseWriter to capture status code
type circuitResponseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

// WriteHeader implements http.ResponseWriter
func (crw *circuitResponseWriter) WriteHeader(statusCode int) {
	if !crw.written {
		crw.statusCode = statusCode
		crw.written = true
		crw.ResponseWriter.WriteHeader(statusCode)
	}
}

// Write implements http.ResponseWriter
func (crw *circuitResponseWriter) Write(b []byte) (int, error) {
	if !crw.written {
		crw.written = true
	}
	return crw.ResponseWriter.Write(b)
}
