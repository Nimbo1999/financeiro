package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/jackc/pgx/v5"
	"github.com/nimbo1999/financeiro/authentication/internal/app"
	"github.com/nimbo1999/financeiro/authentication/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	configs := config.LoadConfigFromEnvironment()

	db, err := gorm.Open(postgres.Open(configs.PostgresConnectionString), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	application := app.New(db)

	// Initialize JWT service (shared between HTTP and gRPC servers)
	jwtService, err := application.InitializeJWTService()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize JWT service: %v", err))
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	serverErrors := make(chan error, 2)

	// Start HTTP server
	go func() {
		serverErrors <- application.RunHTTP(configs, jwtService)
	}()

	// Start gRPC server
	go func() {
		serverErrors <- application.RunGRPC(configs.GRPCPort, jwtService)
	}()

	select {
	case err := <-serverErrors:
		fmt.Printf("Server error: %v\n", err)
		os.Exit(1)

	case sig := <-shutdown:
		fmt.Printf("\nReceived signal %v, initiating graceful shutdown...\n", sig)
		ctx := context.Background()
		if err := application.Shutdown(ctx); err != nil {
			fmt.Printf("Graceful shutdown error: %v\n", err)
			os.Exit(1)
		}
	}
}
