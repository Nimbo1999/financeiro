package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nimbo1999/financeiro/commons"
	"github.com/nimbo1999/financeiro/transactions/internal/service"
	"gorm.io/gorm"
)

type HTTPHandler interface {
	RegisterRoutes(router chi.Router)
}

type transactionHandler struct {
	service service.TransactionService
}

func (h *transactionHandler) RegisterRoutes(router chi.Router) {
	router.Use(middleware.SetHeader("Content-Type", "application/json"))
	router.Post("/", h.CreateTransaction)
	router.Get("/", h.ListTransactions)
}

func (h *transactionHandler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	var request CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		statusCode, err := mapServiceError(err)
		writeErrorResponse(w, statusCode, err)
		return
	}

	transactionDate, err := time.Parse(time.RFC3339, request.Date)
	if err != nil {
		statusCode, apiErr := mapServiceError(err)
		writeErrorResponse(w, statusCode, apiErr)
		return
	}

	if err := h.service.Create(r.Context(), request.UserID, request.Description, request.Amount, transactionDate); err != nil {
		statusCode, apiErr := mapServiceError(err)
		writeErrorResponse(w, statusCode, apiErr)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"message":"Transaction created successfully"}`))
}

func (h *transactionHandler) ListTransactions(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	intPage, _ := strconv.Atoi(query.Get("page"))
	intPageSize, _ := strconv.Atoi(query.Get("page_size"))
	sortBy := query.Get("sort_by")
	order := query.Get("order")
	searchQuery := query.Get("query")

	pagination := commons.NewPagination(intPage, intPageSize, order, commons.Order(sortBy), searchQuery)
	result, err := h.service.List(r.Context(), pagination)
	if err != nil {
		statusCode, apiErr := mapServiceError(err)
		writeErrorResponse(w, statusCode, apiErr)
		return
	}

	var data []TransactionVO
	for _, tx := range result.Data {
		data = append(data, TransactionVOFromModel(tx))
	}

	response := commons.NewPaginatedResult(data, result.Total, pagination)
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

func TransactionHandler(service service.TransactionService) HTTPHandler {
	return &transactionHandler{
		service: service,
	}
}

type healthHandler struct {
	db *gorm.DB
}

func HealthHandler(db *gorm.DB) HTTPHandler {
	return &healthHandler{
		db: db,
	}
}

func (h *healthHandler) RegisterRoutes(router chi.Router) {
	router.Use(middleware.SetHeader("Content-Type", "application/json"))
	router.Get("/", h.HealthCheck)
}

func (h *healthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if err := h.db.Raw("SELECT 1").Scan(new(int)).Error; err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "unhealthy"})
		return
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}
