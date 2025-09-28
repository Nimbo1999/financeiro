package app

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/nimbo1999/financeiro/users/internal/handler"
	"github.com/nimbo1999/financeiro/users/internal/repositories"
	"github.com/nimbo1999/financeiro/users/internal/services"
	userpb "github.com/nimbo1999/financeiro/users/pkg/grpc/users/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func (a *App) RunGRPC(port string) error {
	if port == "" {
		port = "9090"
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", port, err)
	}

	// Create gRPC server with options
	a.grpcServer = grpc.NewServer(
		grpc.MaxRecvMsgSize(4*1024*1024), // 4MB
		grpc.MaxSendMsgSize(4*1024*1024), // 4MB
	)

	// Create service dependencies
	userRepository := repositories.NewUserRepository(a.db)
	userService := services.NewUserService(userRepository)
	userGRPCHandler := handler.NewUserGRPCHandler(userService)

	// Register services
	userpb.RegisterUserServiceServer(a.grpcServer, userGRPCHandler)

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

func (a *App) Shutdown(ctx context.Context) error {
	fmt.Println("Shutting down all servers gracefully...")

	// Shutdown both HTTP and gRPC servers concurrently
	httpErrChan := make(chan error, 1)
	grpcErrChan := make(chan error, 1)

	go func() {
		httpErrChan <- a.ShutdownHTTP(ctx)
	}()

	go func() {
		grpcErrChan <- a.ShutdownGRPC(ctx)
	}()

	// Wait for both shutdowns to complete
	var httpErr, grpcErr error
	for i := 0; i < 2; i++ {
		select {
		case err := <-httpErrChan:
			httpErr = err
		case err := <-grpcErrChan:
			grpcErr = err
		}
	}

	// Close database connections
	if a.db != nil {
		sqlDB, err := a.db.DB()
		if err == nil {
			if err := sqlDB.Close(); err != nil {
				return fmt.Errorf("database close error: %w", err)
			}
		}
		fmt.Println("Database connections closed")
	}

	// Return the first error encountered
	if httpErr != nil {
		return fmt.Errorf("HTTP shutdown error: %w", httpErr)
	}
	if grpcErr != nil {
		return fmt.Errorf("gRPC shutdown error: %w", grpcErr)
	}

	fmt.Println("All servers shutdown complete")
	return nil
}