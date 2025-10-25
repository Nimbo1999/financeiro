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
	"github.com/nimbo1999/financeiro/authentication/internal/clients"
	"github.com/nimbo1999/financeiro/authentication/internal/config"
	"github.com/nimbo1999/financeiro/authentication/internal/handler"
	"github.com/nimbo1999/financeiro/authentication/internal/messaging"
	"github.com/nimbo1999/financeiro/authentication/internal/repository"
	"github.com/nimbo1999/financeiro/authentication/internal/services"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

type App struct {
	db         *gorm.DB
	server     *http.Server
	grpcServer *grpc.Server
	wg         *sync.WaitGroup
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

func (a *App) RunHTTP(config *config.Config, jwtService services.JWTService) error {
	if config.HTTPPort == "" {
		config.HTTPPort = "8080"
	}

	authCodeRepository := repository.NewPostgresAuthCodeRepository(a.db)
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
	healthHandler := handler.NewHealthHandler(a.db, rabbitMqConnection)
	authHandler := handler.NewAuthHandler(authCodeService)

	mux := chi.NewMux()
	// Global middleware
	mux.Use(a.requestTrackingMiddleware)
	mux.Use(handler.CorsMiddleware())
	mux.Use(handler.RecoveryMiddleware)
	mux.Use(handler.LoggingMiddleware)
	mux.Use(handler.SecurityHeadersMiddleware)
	handler.RegisterRoutes(mux, healthHandler, authHandler)

	a.server = &http.Server{
		Addr:    fmt.Sprintf(":%s", config.HTTPPort),
		Handler: mux,
	}

	fmt.Println("Starting HTTP server on port:", config.HTTPPort)
	return a.server.ListenAndServe()
}

func (a *App) InitializeJWTService() (services.JWTService, error) {
	// @TODO: for debug we are reading keys from files, but in production we should use a secure vault like AWS KMS or HashiCorp Vault
	rsaKeys, err := readRSAKeys()
	if err != nil {
		return nil, fmt.Errorf("failed to read RSA keys: %w", err)
	}

	jwtConfig := &services.JWTConfig{
		PrivateKey: rsaKeys.PrivateKey,
		PublicKey:  rsaKeys.PublicKey,
	}

	return services.NewJWTService(jwtConfig), nil
}

func (a *App) ShutdownHTTP(ctx context.Context) error {
	fmt.Println("Shutting down HTTP server gracefully...")

	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if a.server != nil {
		if err := a.server.Shutdown(shutdownCtx); err != nil {
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

func (a *App) Shutdown(ctx context.Context) error {
	fmt.Println("Shutting down all servers gracefully...")

	// Shutdown both HTTP and gRPC servers concurrently
	httpErrChan := make(chan error, 1)
	grpcErrChan := make(chan error, 1)

	go func() {
		httpErrChan <- a.ShutdownHTTP(ctx)
	}()

	go func() {
		grpcErrChan <- a.ShutdownGRPC(ctx)
	}()

	// Wait for both shutdowns to complete
	var httpErr, grpcErr error
	for i := 0; i < 2; i++ {
		select {
		case err := <-httpErrChan:
			httpErr = err
		case err := <-grpcErrChan:
			grpcErr = err
		}
	}

	// Close database connections
	if a.db != nil {
		sqlDB, err := a.db.DB()
		if err == nil {
			if err := sqlDB.Close(); err != nil {
				return fmt.Errorf("database close error: %w", err)
			}
		}
		fmt.Println("Database connections closed")
	}

	// Return the first error encountered
	if httpErr != nil {
		return fmt.Errorf("HTTP shutdown error: %w", httpErr)
	}
	if grpcErr != nil {
		return fmt.Errorf("gRPC shutdown error: %w", grpcErr)
	}

	fmt.Println("All servers shutdown complete")
	return nil
}

type RSAKeys struct {
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
}

// @TODO: for debug we are reading keys from files, but in production we should use a secure vault like AWS KMS or HashiCorp Vault
func readRSAKeys() (*RSAKeys, error) {
	privateKeyFile, err := os.ReadFile("/etc/jwt-keys/private_key.pem")
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}
	publicKeyFile, err := os.ReadFile("/etc/jwt-keys/public_key.pem")
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
