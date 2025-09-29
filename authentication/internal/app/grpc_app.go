package app

import (
	"context"
	"fmt"
	"net"
	"time"

	grpchandler "github.com/nimbo1999/financeiro/authentication/internal/grpc/handler"
	"github.com/nimbo1999/financeiro/authentication/internal/grpc/interceptors"
	"github.com/nimbo1999/financeiro/authentication/internal/services"
	authpb "github.com/nimbo1999/financeiro/authentication/pkg/grpc/auth/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

func (a *App) RunGRPC(port string, jwtService services.JWTService) error {
	if port == "" {
		port = "9090"
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", port, err)
	}

	// Create gRPC server with options and interceptors
	a.grpcServer = grpc.NewServer(
		grpc.MaxRecvMsgSize(4*1024*1024), // 4MB
		grpc.MaxSendMsgSize(4*1024*1024), // 4MB
		grpc.ChainUnaryInterceptor(
			interceptors.RecoveryUnaryInterceptor(),
			interceptors.LoggingUnaryInterceptor(),
		),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             30 * time.Second, // Minimum time between client pings
			PermitWithoutStream: true,             // Allow pings without active streams
		}),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     15 * time.Minute, // Close connection after idle time
			MaxConnectionAge:      30 * time.Minute, // Close connection after this time
			MaxConnectionAgeGrace: 5 * time.Second,  // Grace period for closing
			Time:                  15 * time.Second, // Server ping interval
			Timeout:               1 * time.Second,  // Ping timeout
		}),
	)

	// Create gRPC handler
	authGRPCHandler := grpchandler.NewAuthGRPCHandler(jwtService)

	// Register services
	authpb.RegisterAuthServiceServer(a.grpcServer, authGRPCHandler)

	// Enable reflection for development (optional)
	reflection.Register(a.grpcServer)

	fmt.Println("Starting gRPC server on port:", port)
	return a.grpcServer.Serve(listener)
}

func (a *App) ShutdownGRPC(ctx context.Context) error {
	fmt.Println("Shutting down gRPC server gracefully...")

	if a.grpcServer != nil {
		// Channel to signal when graceful shutdown is complete
		done := make(chan struct{})

		go func() {
			defer close(done)
			a.grpcServer.GracefulStop()
		}()

		// Wait for graceful shutdown or timeout
		shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		select {
		case <-done:
			fmt.Println("gRPC server shutdown complete")
		case <-shutdownCtx.Done():
			fmt.Println("gRPC shutdown timeout reached, forcing stop")
			a.grpcServer.Stop()
		}
	}

	return nil
}
