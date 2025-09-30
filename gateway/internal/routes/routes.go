package routes

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nimbo1999/financeiro/gateway/internal/config"
	"github.com/nimbo1999/financeiro/gateway/internal/proxy"
)

type Router struct {
	config           *config.Config
	mux              *chi.Mux
	proxy            *proxy.LoggingProxy
	healthAggregator *proxy.HealthAggregator
}

func NewRouter(cfg *config.Config) *Router {
	// Use factory to create proxy components following SOLID principles
	factory := proxy.NewFactory(cfg)

	r := &Router{
		config:           cfg,
		mux:              chi.NewRouter(),
		proxy:            factory.CreateLoggingProxy(),
		healthAggregator: factory.CreateHealthAggregator(),
	}
	r.setupMiddleware()
	r.setupRoutes()
	return r
}

func (r *Router) setupMiddleware() {
	// Basic middleware
	r.mux.Use(middleware.RequestID)
	r.mux.Use(middleware.RealIP)
	r.mux.Use(middleware.Logger)
	r.mux.Use(middleware.Recoverer)
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

	// User service routes (will be protected in Step 3.3)
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
