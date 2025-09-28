package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nimbo1999/financeiro/users/internal/handler/dto"
	"github.com/nimbo1999/financeiro/users/internal/services"
)

type HTTPHandler interface {
	RegisterRoutes(router chi.Router)
}

type UserHandler struct {
	service services.UserService
}

func NewUserHandler(service services.UserService) *UserHandler {
	return &UserHandler{
		service: service,
	}
}

func (h *UserHandler) RegisterRoutes(router chi.Router) {
	router.Use(middleware.SetHeader("Content-Type", "application/json"))
	router.Get("/{id}", h.GetUserByID)
	router.Get("/", h.ListUsers)
	router.Post("/", h.CreateUser)
}

func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	user, err := h.service.GetUserByID(id)
	if err != nil {
		errorResponse := dto.ToErrorResponse(err)
		w.WriteHeader(errorResponse.Status)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	var deletedAt *string
	if user.DeletedAt.Valid {
		str := user.DeletedAt.Time.Format("2006-01-02T15:04:05Z07:00")
		deletedAt = &str
	}

	json.NewEncoder(w).Encode(dto.GetUserResponse{
		ID:        user.ID,
		Email:     user.Email,
		FullName:  user.FullName,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		DeletedAt: deletedAt,
	})
}

func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	intPage, _ := strconv.Atoi(query.Get("page"))
	intPageSize, _ := strconv.Atoi(query.Get("page_size"))

	paginationParms := &services.PaginationParams{
		Page:     intPage,
		PageSize: intPageSize,
		OrderBy:  query.Get("order_by"),
		Sort:     query.Get("sort"),
	}

	result, err := h.service.ListUsers(paginationParms)
	if err != nil {
		errorResponse := dto.ToErrorResponse(err)
		w.WriteHeader(errorResponse.Status)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	users := make([]dto.GetUserResponse, len(result.Users))
	for i, user := range result.Users {
		users[i] = dto.GetUserResponse{
			ID:        user.ID,
			Email:     user.Email,
			FullName:  user.FullName,
			CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	json.NewEncoder(w).Encode(dto.ListUsersResponse{
		Users:      users,
		Total:      result.Total,
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalPages: result.TotalPages,
	})
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var createUser services.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&createUser); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	user, err := h.service.CreateUser(&createUser)
	if err != nil {
		errorResponse := dto.ToErrorResponse(err)
		w.WriteHeader(errorResponse.Status)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dto.CreateUserResponse{
		ID:        user.ID,
		Email:     user.Email,
		FullName:  user.FullName,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}
