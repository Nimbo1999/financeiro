package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nimbo1999/financeiro/authentication/internal/messaging"
	"gorm.io/gorm"
)

type HealthHandler struct {
	db        *gorm.DB
	publisher messaging.Publisher
}

func NewHealthHandler(db *gorm.DB, publisher messaging.Publisher) *HealthHandler {
	return &HealthHandler{
		db:        db,
		publisher: publisher,
	}
}

type HealthResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
	RabbitMQ string `json:"rabbitmq"`
}

func (h *HealthHandler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:   "healthy",
		Database: h.checkDatabase(),
		RabbitMQ: h.checkRabbitMQ(),
	}

	// If any service is unhealthy, set overall status to degraded
	if response.Database != "healthy" || response.RabbitMQ != "healthy" {
		response.Status = "degraded"
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	json.NewEncoder(w).Encode(response)
}

func (h *HealthHandler) checkDatabase() string {
	var value int
	if err := h.db.Raw("SELECT 1").Scan(&value).Error; err != nil {
		return "unhealthy"
	}
	return "healthy"
}

func (h *HealthHandler) checkRabbitMQ() string {
	if h.publisher == nil || !h.publisher.IsHealthy() {
		return "unhealthy"
	}
	return "healthy"
}

func (h *HealthHandler) RegisterRoutes(r chi.Router) chi.Router {
	return r.Route("/health", func(router chi.Router) {
		router.Use(ContentTypeMiddleware)
		router.Use(middleware.GetHead)
		router.Get("/", h.HealthHandler)
	})
}
