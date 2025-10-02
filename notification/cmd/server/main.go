package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"notification/internal/app"
	"notification/internal/config"
	"notification/internal/consumers"
	"notification/internal/repository"
	"notification/internal/services"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// 1. Load config
	cfg := config.LoadFromEnvironment()

	// 2. Connect to database
	db, err := connectDB(cfg.PostgresConnectionString)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 3. Initialize repositories
	notificationRepo := repository.NewNotificationRepository(db)

	// 4. Initialize services
	templateSvc := services.NewTemplateService(cfg.TemplateDir)
	emailSvc := services.NewEmailService(
		cfg.SMTPHost,
		cfg.SMTPPort,
		cfg.SMTPUsername,
		cfg.SMTPPassword,
		cfg.SMTPFrom,
		templateSvc,
	)
	notificationSvc := services.NewNotificationService(emailSvc, notificationRepo)

	// 5. Initialize RabbitMQ consumer
	consumer, err := consumers.NewRabbitMQConsumer(
		cfg.RabbitMQURL,
		cfg.WelcomeQueueName,
		cfg.OTPQueueName,
		cfg.ConsumerPrefetchCount,
		notificationSvc,
	)
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}

	// 6. Initialize and run application
	application := app.NewApp(db, consumer, cfg)

	ctx := context.Background()

	// Start HTTP server
	go func() {
		if err := application.RunHTTP(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Start consumer
	go func() {
		if err := application.RunConsumer(ctx); err != nil {
			log.Fatalf("Consumer error: %v", err)
		}
	}()

	log.Printf("HTTP server listening on port %s\n", cfg.HTTPPort)
	log.Println("Consumer has been started")

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	application.ShutdownApp(ctx)
}

func connectDB(connString string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(connString), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	log.Println("Database connection established")
	return db, nil
}
