package app

import (
	"context"
	"crypto/rsa"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nimbo1999/financeiro/authentication/internal/clients"
	"github.com/nimbo1999/financeiro/authentication/internal/config"
	"github.com/nimbo1999/financeiro/authentication/internal/handler"
	"github.com/nimbo1999/financeiro/authentication/internal/messaging"
	"github.com/nimbo1999/financeiro/authentication/internal/repository"
	"github.com/nimbo1999/financeiro/authentication/internal/services"
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

func (a *App) Run(config *config.Config) error {
	if config.HTTPPort == "" {
		config.HTTPPort = "8080"
	}

	rsaKeys, err := readRSAKeys()
	if err != nil {
		return fmt.Errorf("failed to read RSA keys: %w", err)
	}

	jwtConfig := &services.JWTConfig{
		PrivateKey: rsaKeys.PrivateKey,
		PublicKey:  rsaKeys.PublicKey,
	}

	authCodeRepository := repository.NewPostgresAuthCodeRepository(a.db)
	jwtService := services.NewJWTService(jwtConfig)
	userServiceClient, err := clients.NewUserServiceClient(clients.UserServiceConfig{
		Address: config.UserGRPCAddress,
	})

	if err != nil {
		return fmt.Errorf("failed to create user service client: %w", err)
	}

	rabbitMqConnection := messaging.NewRabbitMQConnection(messaging.RabbitMQConfig{
		URL: config.RabbitMQURL,
	})

	if err = rabbitMqConnection.Connect(); err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	publisher, err := messaging.NewPublisher(rabbitMqConnection, messaging.NewQueueManager(rabbitMqConnection), messaging.DefaultPublisherConfig())
	if err != nil {
		return fmt.Errorf("failed to create RabbitMQ publisher: %w", err)
	}

	authCodeService := services.NewAuthService(authCodeRepository, jwtService, userServiceClient, publisher, nil)
	healthHandler := handler.NewHealthHandler(a.db)
	authHandler := handler.NewAuthHandler(authCodeService)

	mux := chi.NewMux()
	// Global middleware
	mux.Use(a.requestTrackingMiddleware)
	mux.Use(middleware.GetHead)
	mux.Use(handler.CorsMiddleware())
	mux.Use(handler.RecoveryMiddleware)
	mux.Use(handler.LoggingMiddleware)
	mux.Use(handler.SecurityHeadersMiddleware)
	handler.RegisterRoutes(mux, healthHandler, authHandler)

	a.server = &http.Server{
		Addr:    fmt.Sprintf(":%s", config.HTTPPort),
		Handler: mux,
	}

	fmt.Println("Starting server on port:", config.HTTPPort)
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

type RSAKeys struct {
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
}

// @TODO: for debug we are reading keys from files, but in production we should use a secure vault like AWS KMS or HashiCorp Vault
func readRSAKeys() (*RSAKeys, error) {
	privateKeyFile, err := os.ReadFile("private_key.pem")
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}
	publicKeyFile, err := os.ReadFile("public_key.pem")
	if err != nil {
		return nil, fmt.Errorf("failed to read public key file: %w", err)
	}
	privateKey, err := services.LoadRSAPrivateKey(privateKeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load RSA private key: %w", err)
	}
	publicKey, err := services.LoadRSAPublicKey(publicKeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load RSA public key: %w", err)
	}
	return &RSAKeys{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
	}, nil
}
