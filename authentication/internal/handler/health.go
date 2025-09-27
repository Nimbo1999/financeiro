package handler

import (
	"fmt"
	"net/http"

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

func (h *HealthHandler) HealthRoute(w http.ResponseWriter, r *http.Request) {
	var value int
	if err := h.db.Raw("SELECT 1").Scan(&value).Error; err != nil {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(fmt.Sprintln("Database connection error")))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintln("Database and service are healthy")))
}
