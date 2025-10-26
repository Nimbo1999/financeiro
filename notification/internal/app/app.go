package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nimbo1999/financeiro/notification/internal/config"
	"github.com/nimbo1999/financeiro/notification/internal/consumers"
	"github.com/nimbo1999/financeiro/notification/internal/handler"

	"gorm.io/gorm"
)

type App struct {
	config   *config.Config
	db       *gorm.DB
	consumer consumers.Consumer
	server   *http.Server
}

func NewApp(db *gorm.DB, consumer consumers.Consumer, cfg *config.Config) *App {
	// Initialize HTTP server (health checks)
	// Cast consumer to HealthChecker interface
	healthChecker, ok := consumer.(handler.HealthChecker)
	if !ok {
		log.Println("Warning: consumer does not implement HealthChecker interface")
		healthChecker = nil
	}

	healthHandler := handler.NewHealthHandler(db, healthChecker)
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Get("/health", healthHandler.Health)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.HTTPPort),
		Handler: router,
	}

	return &App{
		config:   cfg,
		db:       db,
		consumer: consumer,
		server:   server,
	}
}

func (a *App) RunConsumer(ctx context.Context) error {
	return a.consumer.Start(ctx)
}

func (a *App) RunHTTP() error {
	return a.server.ListenAndServe()
}

func (a *App) ShutdownApp(ctx context.Context) {
	log.Println("Shutting down gracefully...")

	// Shutdown HTTP server
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 10*time.Second)
	defer shutdownCancel()

	if err := a.server.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	// Stop consumer
	if err := a.consumer.Stop(); err != nil {
		log.Printf("Consumer shutdown error: %v", err)
	}

	// Close database
	sqlDB, err := a.db.DB()
	if err == nil {
		if err := sqlDB.Close(); err != nil {
			log.Printf("Database close error: %v", err)
		}
	}

	log.Println("Shutdown complete")
}
