package handler

import (
	"encoding/json"
	"net/http"

	"gorm.io/gorm"
)

type HealthResponse struct {
	Status   string `json:"status"`
	Service  string `json:"service"`
	Database string `json:"database"`
	RabbitMQ string `json:"rabbitmq"`
}

type HealthChecker interface {
	IsHealthy() bool
}

type HealthHandler struct {
	db       *gorm.DB
	consumer HealthChecker
}

func NewHealthHandler(db *gorm.DB, consumer HealthChecker) *HealthHandler {
	return &HealthHandler{
		db:       db,
		consumer: consumer,
	}
}

func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := HealthResponse{
		Status:   "healthy",
		Service:  "notification",
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
	if h.db == nil {
		return "unhealthy"
	}

	sqlDB, err := h.db.DB()
	if err != nil {
		return "unhealthy"
	}

	if err := sqlDB.Ping(); err != nil {
		return "unhealthy"
	}

	return "healthy"
}

func (h *HealthHandler) checkRabbitMQ() string {
	if h.consumer == nil || !h.consumer.IsHealthy() {
		return "unhealthy"
	}
	return "healthy"
}
