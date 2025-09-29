package clients

import (
	"context"
	"fmt"
	"time"

	userv1 "github.com/nimbo1999/financeiro/users/pkg/grpc/users/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

type UserServiceClient interface {
	GetUserByEmail(ctx context.Context, email string) (*userv1.User, bool, error)
	GetUserById(ctx context.Context, userID string) (*userv1.User, bool, error)
	HealthCheck(ctx context.Context) (userv1.HealthCheckResponse_Status, string, error)
	Close() error
}

type userServiceClient struct {
	conn   *grpc.ClientConn
	client userv1.UserServiceClient
}

type UserServiceConfig struct {
	Address        string
	ConnectTimeout time.Duration
	RequestTimeout time.Duration
	MaxRetries     int
}

func NewUserServiceClient(config UserServiceConfig) (UserServiceClient, error) {
	if config.ConnectTimeout == 0 {
		config.ConnectTimeout = 10 * time.Second
	}
	if config.RequestTimeout == 0 {
		config.RequestTimeout = 5 * time.Second
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}

	conn, err := grpc.NewClient(config.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                30 * time.Second,
			Timeout:             3 * time.Second,
			PermitWithoutStream: true,
		}),
		grpc.WithDefaultCallOptions(
			grpc.WaitForReady(true),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create user service client: %w", err)
	}

	// Verify connection within timeout
	ctx, cancel := context.WithTimeout(context.Background(), config.ConnectTimeout)
	defer cancel()

	// Try to establish connection by attempting to connect
	conn.Connect()

	// Wait for connection to be established or timeout
	for {
		state := conn.GetState()
		if state == connectivity.Ready {
			break
		}
		if state == connectivity.TransientFailure {
			conn.Close()
			return nil, fmt.Errorf("failed to connect to user service at %s: connection failed", config.Address)
		}

		if !conn.WaitForStateChange(ctx, state) {
			conn.Close()
			return nil, fmt.Errorf("failed to connect to user service at %s: timeout after %v", config.Address, config.ConnectTimeout)
		}
	}

	client := userv1.NewUserServiceClient(conn)

	return &userServiceClient{
		conn:   conn,
		client: client,
	}, nil
}

func (c *userServiceClient) GetUserByEmail(ctx context.Context, email string) (*userv1.User, bool, error) {
	if c.conn.GetState() != connectivity.Ready && c.conn.GetState() != connectivity.Idle {
		return nil, false, fmt.Errorf("user service connection not ready, state: %v", c.conn.GetState())
	}

	req := &userv1.GetUserByEmailRequest{
		Email: email,
	}

	resp, err := c.client.GetUserByEmail(ctx, req)
	if err != nil {
		return nil, false, fmt.Errorf("failed to get user by email: %w", err)
	}

	return resp.User, resp.Found, nil
}

func (c *userServiceClient) GetUserById(ctx context.Context, userID string) (*userv1.User, bool, error) {
	if c.conn.GetState() != connectivity.Ready && c.conn.GetState() != connectivity.Idle {
		return nil, false, fmt.Errorf("user service connection not ready, state: %v", c.conn.GetState())
	}

	req := &userv1.GetUserByIdRequest{
		Id: userID,
	}

	resp, err := c.client.GetUserById(ctx, req)
	if err != nil {
		return nil, false, fmt.Errorf("failed to get user by id: %w", err)
	}

	return resp.User, resp.Found, nil
}

func (c *userServiceClient) HealthCheck(ctx context.Context) (userv1.HealthCheckResponse_Status, string, error) {
	req := &userv1.HealthCheckRequest{}

	resp, err := c.client.HealthCheck(ctx, req)
	if err != nil {
		return userv1.HealthCheckResponse_UNKNOWN, "", fmt.Errorf("failed to perform health check: %w", err)
	}

	return resp.Status, resp.Message, nil
}

func (c *userServiceClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
