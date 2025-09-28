package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/jackc/pgx/v5"
	"github.com/nimbo1999/financeiro/users/internal/app"
	"github.com/nimbo1999/financeiro/users/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	config := config.LoadConfigFromEnvironment()

	db, err := gorm.Open(postgres.Open(config.PostgresConnectionString), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	app := app.New(db)

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	serverErrors := make(chan error, 2)

	// Start HTTP server
	go func(port string) {
		serverErrors <- app.RunHTTP(port)
	}(config.HttpPort)

	// Start gRPC server
	go func(port string) {
		serverErrors <- app.RunGRPC(port)
	}(config.GrpcPort)

	select {
	case err := <-serverErrors:
		fmt.Printf("Server error: %v\n", err)
		os.Exit(1)

	case sig := <-shutdown:
		fmt.Printf("\nReceived signal %v, initiating graceful shutdown...\n", sig)

		ctx := context.Background()
		if err := app.Shutdown(ctx); err != nil {
			fmt.Printf("Graceful shutdown error: %v\n", err)
			os.Exit(1)
		}
	}
}
