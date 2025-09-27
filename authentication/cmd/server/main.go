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

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- application.Run(configs.Port)
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
