package routes

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nimbo1999/financeiro/gateway/internal/clients"
	"github.com/nimbo1999/financeiro/gateway/internal/config"
	authmw "github.com/nimbo1999/financeiro/gateway/internal/middleware"
	"github.com/nimbo1999/financeiro/gateway/internal/proxy"
)

type Router struct {
	config           *config.Config
	mux              *chi.Mux
	proxy            *proxy.LoggingProxy
	healthAggregator *proxy.HealthAggregator
	authMiddleware   *authmw.AuthMiddleware
	authClient       *clients.AuthServiceClient
}

func NewRouter(cfg *config.Config) *Router {
	// Use factory to create proxy components following SOLID principles
	factory := proxy.NewFactory(cfg)

	// Create auth service gRPC client
	authClient, err := clients.NewAuthServiceClient(cfg.Services.AuthServiceGRPCURL)
	if err != nil {
		log.Fatalf("Failed to create auth service client: %v", err)
	}

	// Define public paths that don't require authentication
	publicPaths := []string{
		"/health",
		"/health/*",
		"/auth/*",
	}

	// Create authentication middleware
	authMiddleware := authmw.NewAuthMiddleware(authClient, publicPaths)

	r := &Router{
		config:           cfg,
		mux:              chi.NewRouter(),
		proxy:            factory.CreateLoggingProxy(),
		healthAggregator: factory.CreateHealthAggregator(),
		authMiddleware:   authMiddleware,
		authClient:       authClient,
	}
	r.setupMiddleware()
	r.setupRoutes()
	return r
}

// Close closes resources held by the router
func (r *Router) Close() error {
	if r.authClient != nil {
		return r.authClient.Close()
	}
	return nil
}

func (r *Router) setupMiddleware() {
	// Basic middleware
	r.mux.Use(middleware.RequestID)
	r.mux.Use(middleware.RealIP)
	r.mux.Use(middleware.Logger)
	r.mux.Use(middleware.Recoverer)

	// CORS middleware
	if r.config.Security.EnableCORS {
		cors := authmw.NewCORSMiddleware(authmw.DefaultCORSConfig())
		r.mux.Use(cors.Handler)
	}

	// Timeout middleware
	r.mux.Use(
		middleware.Timeout(time.Duration(r.config.Server.RequestTimeout) * time.Second),
	)

	// Circuit breaker middleware (for downstream services)
	if r.config.CircuitBreaker.Enabled {
		cbConfig := authmw.DefaultCircuitBreakerConfig()
		cbConfig.MaxFailures = r.config.CircuitBreaker.MaxFailures
		cbConfig.ResetTimeout = time.Duration(r.config.CircuitBreaker.ResetTimeout) * time.Second
		circuitBreaker := authmw.NewCircuitBreaker(cbConfig)
		r.mux.Use(circuitBreaker.Handler)
	}

	// Authentication middleware (applies to all routes, but checks for public paths)
	r.mux.Use(r.authMiddleware.Handler)
}

func (r *Router) setupRoutes() {
	// Health check endpoints
	r.mux.Route("/health", func(router chi.Router) {
		router.Use(middleware.SetHeader("Content-Type", "application/json"))
		router.Get("/", r.healthCheckHandler)
		router.Get("/services", r.healthAggregator.AggregatedHealthHandler)
	})

	// Authentication service routes (public)
	r.mux.Route("/auth", func(router chi.Router) {
		router.Use(middleware.PathRewrite("/auth", ""))
		router.HandleFunc("/*", r.proxyToAuthService)
	})

	// User service routes (protected - requires authentication)
	r.mux.Route("/users", func(router chi.Router) {
		router.Use(middleware.PathRewrite("/users", ""))
		router.HandleFunc("/*", r.proxyToUserService)
	})
}

func (r *Router) healthCheckHandler(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func (r *Router) proxyToAuthService(w http.ResponseWriter, req *http.Request) {
	if err := r.proxy.ProxyRequestWithLogging(w, req, r.config.Services.AuthServiceURL, "authentication"); err != nil {
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
	}
}

func (r *Router) proxyToUserService(w http.ResponseWriter, req *http.Request) {
	if err := r.proxy.ProxyRequestWithLogging(w, req, r.config.Services.UserServiceURL, "users"); err != nil {
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
	}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}
