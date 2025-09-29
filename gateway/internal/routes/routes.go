package routes

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nimbo1999/financeiro/gateway/internal/config"
)

type Router struct {
	config *config.Config
	mux    *chi.Mux
}

func NewRouter(cfg *config.Config) *Router {
	r := &Router{
		config: cfg,
		mux:    chi.NewRouter(),
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
	// Health check endpoint
	r.mux.Route("/health", func(router chi.Router) {
		router.Use(middleware.SetHeader("Content-Type", "application/json"))
		router.Use(middleware.GetHead)
		router.Get("/", r.healthCheckHandler)
	})

	// Service routes will be added in Step 3.2
	// Protected routes will be added in Step 3.3
}

func (r *Router) healthCheckHandler(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}
