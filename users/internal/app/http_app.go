package app

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nimbo1999/financeiro/users/internal/handler"
	"github.com/nimbo1999/financeiro/users/internal/messaging"
	"github.com/nimbo1999/financeiro/users/internal/repositories"
	"github.com/nimbo1999/financeiro/users/internal/services"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

type App struct {
	db         *gorm.DB
	httpServer *http.Server
	grpcServer *grpc.Server
	publisher  messaging.Publisher
	wg         *sync.WaitGroup
}

func New(db *gorm.DB, publisher messaging.Publisher) *App {
	return &App{
		db:        db,
		publisher: publisher,
		wg:        &sync.WaitGroup{},
	}
}

func (a *App) requestTrackingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.wg.Add(1)
		defer a.wg.Done()
		next.ServeHTTP(w, r)
	})
}

func (a *App) RunHTTP(port string) error {
	if port == "" {
		port = "8080"
	}

	userRepository := repositories.NewUserRepository(a.db)
	userService := services.NewUserService(userRepository, a.publisher)
	userHandler := handler.NewUserHandler(userService)

	mux := chi.NewMux()
	mux.Use(middleware.Logger)
	mux.Use(a.requestTrackingMiddleware)
	mux.Use(cors)
	mux.Route("/", userHandler.RegisterRoutes)

	a.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: mux,
	}

	fmt.Println("Starting HTTP server on port:", port)
	return a.httpServer.ListenAndServe()
}

func (a *App) ShutdownHTTP(ctx context.Context) error {
	fmt.Println("Shutting down HTTP server gracefully...")

	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if a.httpServer != nil {
		if err := a.httpServer.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("HTTP server shutdown error: %w", err)
		}
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		a.wg.Wait()
	}()

	select {
	case <-done:
		fmt.Println("All HTTP requests completed")
	case <-shutdownCtx.Done():
		fmt.Println("HTTP shutdown timeout reached, forcing exit")
	}

	fmt.Println("HTTP server shutdown complete")
	return nil
}
