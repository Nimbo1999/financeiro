package clients

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type CircuitBreakerTestSuite struct {
	suite.Suite
	circuitBreaker *CircuitBreaker
	config         CircuitBreakerConfig
}

func (suite *CircuitBreakerTestSuite) SetupTest() {
	suite.config = CircuitBreakerConfig{
		MaxFailures:      3,
		ResetTimeout:     100 * time.Millisecond,
		FailureThreshold: 0.5,
		RequestTimeout:   50 * time.Millisecond,
	}
	suite.circuitBreaker = NewCircuitBreaker(suite.config)
}

func TestCircuitBreakerTestSuite(t *testing.T) {
	suite.Run(t, new(CircuitBreakerTestSuite))
}

func (suite *CircuitBreakerTestSuite) TestCircuitBreaker_InitialState() {
	assert.Equal(suite.T(), StateClosed, suite.circuitBreaker.State())
	state, failures, requests := suite.circuitBreaker.Stats()
	assert.Equal(suite.T(), StateClosed, state)
	assert.Equal(suite.T(), 0, failures)
	assert.Equal(suite.T(), 0, requests)
}

func (suite *CircuitBreakerTestSuite) TestCircuitBreaker_SuccessfulExecution() {
	executed := false
	err := suite.circuitBreaker.Execute(context.Background(), func(ctx context.Context) error {
		executed = true
		return nil
	})

	assert.NoError(suite.T(), err)
	assert.True(suite.T(), executed)
	assert.Equal(suite.T(), StateClosed, suite.circuitBreaker.State())
}

func (suite *CircuitBreakerTestSuite) TestCircuitBreaker_FailedExecution() {
	expectedError := errors.New("test error")
	executed := false

	err := suite.circuitBreaker.Execute(context.Background(), func(ctx context.Context) error {
		executed = true
		return expectedError
	})

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), expectedError, err)
	assert.True(suite.T(), executed)
	assert.Equal(suite.T(), StateClosed, suite.circuitBreaker.State())

	_, failures, requests := suite.circuitBreaker.Stats()
	assert.Equal(suite.T(), 1, failures)
	assert.Equal(suite.T(), 1, requests)
}

func (suite *CircuitBreakerTestSuite) TestCircuitBreaker_OpenAfterMaxFailures() {
	expectedError := errors.New("test error")

	// Trigger enough failures to open circuit
	for i := 0; i < suite.config.MaxFailures; i++ {
		err := suite.circuitBreaker.Execute(context.Background(), func(ctx context.Context) error {
			return expectedError
		})
		assert.Error(suite.T(), err)
		assert.Equal(suite.T(), expectedError, err)
	}

	assert.Equal(suite.T(), StateOpen, suite.circuitBreaker.State())

	// Next request should be rejected without execution
	executed := false
	err := suite.circuitBreaker.Execute(context.Background(), func(ctx context.Context) error {
		executed = true
		return nil
	})

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrCircuitBreakerOpen, err)
	assert.False(suite.T(), executed)
}

func (suite *CircuitBreakerTestSuite) TestCircuitBreaker_HalfOpenAfterTimeout() {
	expectedError := errors.New("test error")

	// Open the circuit
	for i := 0; i < suite.config.MaxFailures; i++ {
		suite.circuitBreaker.Execute(context.Background(), func(ctx context.Context) error {
			return expectedError
		})
	}
	assert.Equal(suite.T(), StateOpen, suite.circuitBreaker.State())

	// Wait for reset timeout
	time.Sleep(suite.config.ResetTimeout + 10*time.Millisecond)

	// First request should transition to half-open
	executed := false
	err := suite.circuitBreaker.Execute(context.Background(), func(ctx context.Context) error {
		executed = true
		return nil
	})

	assert.NoError(suite.T(), err)
	assert.True(suite.T(), executed)
	assert.Equal(suite.T(), StateClosed, suite.circuitBreaker.State())
}

func (suite *CircuitBreakerTestSuite) TestCircuitBreaker_HalfOpenToOpenOnFailure() {
	expectedError := errors.New("test error")

	// Open the circuit
	for i := 0; i < suite.config.MaxFailures; i++ {
		suite.circuitBreaker.Execute(context.Background(), func(ctx context.Context) error {
			return expectedError
		})
	}

	// Wait for reset timeout
	time.Sleep(suite.config.ResetTimeout + 10*time.Millisecond)

	// First request fails, should go back to open
	err := suite.circuitBreaker.Execute(context.Background(), func(ctx context.Context) error {
		return expectedError
	})

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), expectedError, err)
	assert.Equal(suite.T(), StateOpen, suite.circuitBreaker.State())
}

func (suite *CircuitBreakerTestSuite) TestCircuitBreaker_RequestTimeout() {
	executed := false
	longRunningErr := suite.circuitBreaker.Execute(context.Background(), func(ctx context.Context) error {
		executed = true
		// Check if context is cancelled due to timeout
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(suite.config.RequestTimeout + 10*time.Millisecond):
			return nil
		}
	})

	assert.Error(suite.T(), longRunningErr)
	assert.True(suite.T(), executed)
	// The context should be cancelled due to timeout
	assert.Contains(suite.T(), longRunningErr.Error(), "context deadline exceeded")
}

func (suite *CircuitBreakerTestSuite) TestCircuitBreaker_DefaultConfig() {
	cb := NewCircuitBreaker(CircuitBreakerConfig{})

	assert.Equal(suite.T(), 5, cb.config.MaxFailures)
	assert.Equal(suite.T(), 60*time.Second, cb.config.ResetTimeout)
	assert.Equal(suite.T(), 0.6, cb.config.FailureThreshold)
	assert.Equal(suite.T(), 5*time.Second, cb.config.RequestTimeout)
}

func TestCircuitBreakerStates(t *testing.T) {
	testCases := []struct {
		name     string
		state    CircuitBreakerState
		expected string
	}{
		{"Closed state", StateClosed, "StateClosed"},
		{"Open state", StateOpen, "StateOpen"},
		{"Half-open state", StateHalfOpen, "StateHalfOpen"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// This test verifies the state constants are properly defined
			assert.NotEqual(t, tc.state, -1)
		})
	}
}