package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nimbo1999/financeiro/users/internal/handler/dto"
	"github.com/nimbo1999/financeiro/users/internal/models"
	"github.com/nimbo1999/financeiro/users/internal/services"
	"github.com/nimbo1999/financeiro/users/internal/utils"
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
	router.Get("/health", h.HealthCheck)
	router.Patch("/{email}/change-admin-state", h.ChangeAdminState)
	router.Get("/{email}", h.GetUserByIdOrEmail)
	router.Get("/", h.ListUsers)
	router.Post("/", h.CreateUser)
}

/*
@todo: refactor this method to separate GetByID and GetByEmail
it needs to identify if the param is an UUID or an email and threat it accordingly
*/
func (h *UserHandler) GetUserByIdOrEmail(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "email")
	var (
		err  error
		user *models.User
	)

	if utils.IsUUID(id) {
		user, err = h.service.GetUserByID(id)
	} else if utils.IsEmail(id) {
		user, err = h.service.GetUserByEmail(id)
	} else {
		err = services.ErrInvalidUserData
	}

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
		IsAdmin:   user.IsAdmin,
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
		FullName: query.Get("full_name"),
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
			IsAdmin:   user.IsAdmin,
		}
	}

	json.NewEncoder(w).Encode(dto.ListUsersResponse{
		Data:       users,
		Total:      result.Total,
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalPages: result.TotalPages,
	})
}

func (h *UserHandler) ChangeAdminState(w http.ResponseWriter, r *http.Request) {
	email := chi.URLParam(r, "email")
	if err := h.service.ChangeAdminState(email); err != nil {
		errorResponse := dto.ToErrorResponse(err)
		w.WriteHeader(errorResponse.Status)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}
	w.WriteHeader(http.StatusNoContent)
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
		IsAdmin:   user.IsAdmin,
	})
}

func (h *UserHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
