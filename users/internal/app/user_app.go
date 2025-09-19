package app

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/nimbo1999/financeiro/users/internal/handler"
	"github.com/nimbo1999/financeiro/users/internal/repositories"
	"github.com/nimbo1999/financeiro/users/internal/services"
	"gorm.io/gorm"
)

type App struct {
	db     *gorm.DB
	server *http.Server
	wg     *sync.WaitGroup
}

func New(db *gorm.DB) *App {
	return &App{
		db: db,
		wg: &sync.WaitGroup{},
	}
}

func (a *App) requestTrackingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.wg.Add(1)
		defer a.wg.Done()
		next.ServeHTTP(w, r)
	})
}

func (a *App) Run(port string) error {
	if port == "" {
		port = "8080"
	}

	userRepository := repositories.NewUserRepository(a.db)
	userService := services.NewUserService(userRepository)
	userHandler := handler.NewUserHandler(userService)

	mux := chi.NewMux()
	mux.Use(a.requestTrackingMiddleware)
	mux.Route("/", userHandler.RegisterRoutes)

	a.server = &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: mux,
	}

	fmt.Println("Starting server on port:", port)
	return a.server.ListenAndServe()
}

func (a *App) Shutdown(ctx context.Context) error {
	fmt.Println("Shutting down server gracefully...")

	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if a.server != nil {
		if err := a.server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("server shutdown error: %w", err)
		}
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		a.wg.Wait()
	}()

	select {
	case <-done:
		fmt.Println("All requests completed")
	case <-shutdownCtx.Done():
		fmt.Println("Shutdown timeout reached, forcing exit")
	}

	if a.db != nil {
		sqlDB, err := a.db.DB()
		if err == nil {
			if err := sqlDB.Close(); err != nil {
				return fmt.Errorf("database close error: %w", err)
			}
		}
		fmt.Println("Database connections closed")
	}

	fmt.Println("Server shutdown complete")
	return nil
}
