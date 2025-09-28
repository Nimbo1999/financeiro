package clients

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrCircuitBreakerOpen     = errors.New("circuit breaker is open")
	ErrCircuitBreakerHalfOpen = errors.New("circuit breaker is half-open, limiting requests")
)

type CircuitBreakerState int

const (
	StateClosed CircuitBreakerState = iota
	StateOpen
	StateHalfOpen
)

type CircuitBreakerConfig struct {
	MaxFailures     int
	ResetTimeout    time.Duration
	FailureThreshold float64
	RequestTimeout  time.Duration
}

type CircuitBreaker struct {
	config       CircuitBreakerConfig
	state        CircuitBreakerState
	failures     int
	requests     int
	lastFailTime time.Time
	mutex        sync.RWMutex
}

func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker {
	if config.MaxFailures == 0 {
		config.MaxFailures = 5
	}
	if config.ResetTimeout == 0 {
		config.ResetTimeout = 60 * time.Second
	}
	if config.FailureThreshold == 0 {
		config.FailureThreshold = 0.6
	}
	if config.RequestTimeout == 0 {
		config.RequestTimeout = 5 * time.Second
	}

	return &CircuitBreaker{
		config: config,
		state:  StateClosed,
	}
}

func (cb *CircuitBreaker) Execute(ctx context.Context, fn func(ctx context.Context) error) error {
	if !cb.allowRequest() {
		return ErrCircuitBreakerOpen
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, cb.config.RequestTimeout)
	defer cancel()

	err := fn(timeoutCtx)
	cb.recordResult(err)

	return err
}

func (cb *CircuitBreaker) allowRequest() bool {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	now := time.Now()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		if now.Sub(cb.lastFailTime) > cb.config.ResetTimeout {
			cb.state = StateHalfOpen
			cb.requests = 0
			cb.failures = 0
			return true
		}
		return false
	case StateHalfOpen:
		return cb.requests < 3 // Allow limited requests in half-open state
	}

	return false
}

func (cb *CircuitBreaker) recordResult(err error) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.requests++

	if err != nil {
		cb.failures++
		cb.lastFailTime = time.Now()

		if cb.state == StateHalfOpen {
			cb.state = StateOpen
		} else if cb.failures >= cb.config.MaxFailures {
			cb.state = StateOpen
		}
	} else {
		if cb.state == StateHalfOpen {
			cb.state = StateClosed
			cb.failures = 0
			cb.requests = 0
		} else if cb.state == StateClosed {
			if cb.requests >= 10 {
				failureRate := float64(cb.failures) / float64(cb.requests)
				if failureRate < cb.config.FailureThreshold {
					cb.failures = 0
					cb.requests = 0
				}
			}
		}
	}
}

func (cb *CircuitBreaker) State() CircuitBreakerState {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state
}

func (cb *CircuitBreaker) Stats() (state CircuitBreakerState, failures int, requests int) {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state, cb.failures, cb.requests
}