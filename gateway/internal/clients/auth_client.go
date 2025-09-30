package clients

import (
	"context"
	"fmt"
	"time"

	authv1 "github.com/nimbo1999/financeiro/authentication/pkg/grpc/auth/v1"
	"github.com/nimbo1999/financeiro/gateway/internal/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// AuthServiceClient wraps the gRPC client for authentication service
type AuthServiceClient struct {
	client authv1.AuthServiceClient
	conn   *grpc.ClientConn
}

// NewAuthServiceClient creates a new authentication service gRPC client
func NewAuthServiceClient(serverAddr string) (*AuthServiceClient, error) {
	// Create gRPC connection with connection pooling
	conn, err := grpc.NewClient(
		serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(10*1024*1024), // 10MB
			grpc.MaxCallSendMsgSize(10*1024*1024), // 10MB
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC client: %w", err)
	}

	client := authv1.NewAuthServiceClient(conn)

	return &AuthServiceClient{
		client: client,
		conn:   conn,
	}, nil
}

// ValidateToken validates a JWT token by calling the auth service
func (c *AuthServiceClient) ValidateToken(ctx context.Context, token string) (*middleware.UserContext, error) {
	// Set timeout for the gRPC call
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Call the gRPC service
	resp, err := c.client.ValidateToken(ctx, &authv1.ValidateTokenRequest{
		Token: token,
	})
	if err != nil {
		return nil, fmt.Errorf("gRPC call failed: %w", err)
	}

	// Check if token is valid
	if !resp.Valid {
		return nil, fmt.Errorf("invalid token: %s", resp.ErrorMessage)
	}

	// Check if user context is present
	if resp.UserContext == nil {
		return nil, fmt.Errorf("missing user context in response")
	}

	// Return user context
	return &middleware.UserContext{
		UserID: resp.UserContext.UserId,
		Email:  resp.UserContext.Email,
	}, nil
}

// Close closes the gRPC connection
func (c *AuthServiceClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// HealthCheck checks if the auth service is healthy
func (c *AuthServiceClient) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	resp, err := c.client.HealthCheck(ctx, &authv1.HealthCheckRequest{})
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	if resp.Status != authv1.HealthCheckResponse_SERVING {
		return fmt.Errorf("service not serving: %s", resp.Message)
	}

	return nil
}