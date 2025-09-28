package clients

import (
	"context"
	"time"

	userv1 "github.com/nimbo1999/financeiro/users/pkg/grpc/users/v1"
)

type ResilientUserServiceClient struct {
	client         UserServiceClient
	circuitBreaker *CircuitBreaker
}

type ResilientUserServiceConfig struct {
	UserServiceConfig    UserServiceConfig
	CircuitBreakerConfig CircuitBreakerConfig
}

func NewResilientUserServiceClient(config ResilientUserServiceConfig) (UserServiceClient, error) {
	baseClient, err := NewUserServiceClient(config.UserServiceConfig)
	if err != nil {
		return nil, err
	}

	circuitBreaker := NewCircuitBreaker(config.CircuitBreakerConfig)

	return &ResilientUserServiceClient{
		client:         baseClient,
		circuitBreaker: circuitBreaker,
	}, nil
}

func (r *ResilientUserServiceClient) GetUserByEmail(ctx context.Context, email string) (*userv1.User, bool, error) {
	var user *userv1.User
	var found bool
	var err error

	execErr := r.circuitBreaker.Execute(ctx, func(ctx context.Context) error {
		user, found, err = r.client.GetUserByEmail(ctx, email)
		return err
	})

	if execErr != nil {
		if execErr == ErrCircuitBreakerOpen {
			return nil, false, execErr
		}
		return user, found, execErr
	}

	return user, found, err
}

func (r *ResilientUserServiceClient) GetUserById(ctx context.Context, userID string) (*userv1.User, bool, error) {
	var user *userv1.User
	var found bool
	var err error

	execErr := r.circuitBreaker.Execute(ctx, func(ctx context.Context) error {
		user, found, err = r.client.GetUserById(ctx, userID)
		return err
	})

	if execErr != nil {
		if execErr == ErrCircuitBreakerOpen {
			return nil, false, execErr
		}
		return user, found, execErr
	}

	return user, found, err
}

func (r *ResilientUserServiceClient) HealthCheck(ctx context.Context) (userv1.HealthCheckResponse_Status, string, error) {
	var status userv1.HealthCheckResponse_Status
	var message string
	var err error

	execErr := r.circuitBreaker.Execute(ctx, func(ctx context.Context) error {
		status, message, err = r.client.HealthCheck(ctx)
		return err
	})

	if execErr != nil {
		if execErr == ErrCircuitBreakerOpen {
			return userv1.HealthCheckResponse_NOT_SERVING, "circuit breaker open", execErr
		}
		return status, message, execErr
	}

	return status, message, err
}

func (r *ResilientUserServiceClient) Close() error {
	return r.client.Close()
}

func DefaultResilientUserServiceConfig(address string) ResilientUserServiceConfig {
	return ResilientUserServiceConfig{
		UserServiceConfig: UserServiceConfig{
			Address:        address,
			ConnectTimeout: 10 * time.Second,
			RequestTimeout: 5 * time.Second,
			MaxRetries:     3,
		},
		CircuitBreakerConfig: CircuitBreakerConfig{
			MaxFailures:     5,
			ResetTimeout:    60 * time.Second,
			FailureThreshold: 0.6,
			RequestTimeout:  5 * time.Second,
		},
	}
}