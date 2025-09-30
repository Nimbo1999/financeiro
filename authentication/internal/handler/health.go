package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"gorm.io/gorm"
)

type HealthHandler struct {
	db *gorm.DB
}

func NewHealthHandler(db *gorm.DB) *HealthHandler {
	return &HealthHandler{
		db: db,
	}
}

func (h *HealthHandler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	var value int
	if err := h.db.Raw("SELECT 1").Scan(&value).Error; err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(fmt.Sprintln("Database connection error")))
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *HealthHandler) RegisterRoutes(r chi.Router) chi.Router {
	return r.Route("/health", func(router chi.Router) {
		router.Use(ContentTypeMiddleware)
		router.Use(middleware.GetHead)
		router.Get("/", h.HealthHandler)
	})
}
