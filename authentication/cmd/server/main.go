package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/jackc/pgx/v5"
	"github.com/nimbo1999/financeiro/authentication/internal/app"
	"github.com/nimbo1999/financeiro/authentication/internal/config"
	"github.com/nimbo1999/financeiro/authentication/internal/messaging"
	"github.com/nimbo1999/financeiro/migrator"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB
var cfg *config.Config

func init() {
	cfg = config.LoadConfigFromEnvironment()

	var err error
	db, err = gorm.Open(postgres.Open(cfg.PostgresConnectionString), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	sqlDB, err := db.DB()
	if err != nil {
		panic("failed to get database instance")
	}

	if err = migrator.Migrate(sqlDB); err != nil {
		if err == migrator.ErrNoChange {
			return
		}
		panic(fmt.Sprintf("failed to run migrations: %v", err))
	}
	log.Println("Database migrated successfully!")
}

func main() {
	// Create event publisher with self-healing capabilities
	publisher, err := messaging.NewEventPublisher(messaging.EventPublisherConfig{
		RabbitMQURL:  cfg.RabbitMQURL,
		ExchangeName: "notification.exchange",
		Logger:       nil, // Will use default logger
	})
	if err != nil {
		panic(fmt.Sprintf("failed to create event publisher: %w", err))
	}

	application := app.New(db, publisher)

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
		serverErrors <- application.RunHTTP(cfg, jwtService)
	}()

	// Start gRPC server
	go func() {
		serverErrors <- application.RunGRPC(cfg.GRPCPort, jwtService)
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
