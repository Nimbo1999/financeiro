package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/jackc/pgx/v5"
	"github.com/nimbo1999/financeiro/migrator"
	"github.com/nimbo1999/financeiro/users/internal/app"
	"github.com/nimbo1999/financeiro/users/internal/config"
	"github.com/nimbo1999/financeiro/users/internal/messaging"
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
	publisher, err := messaging.NewPublisher(cfg.RabbitMQURL, "notification.exchange")
	if err != nil {
		log.Printf("Warning: Failed to initialize RabbitMQ publisher: %v", err)
		log.Println("User service will run without event publishing")
	}

	app := app.New(db, publisher)

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	serverErrors := make(chan error, 2)

	// Start HTTP server
	go func() {
		serverErrors <- app.RunHTTP(cfg.HttpPort)
	}()

	// Start gRPC server
	go func() {
		serverErrors <- app.RunGRPC(cfg.GrpcPort)
	}()

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
