package dto

import (
	"net/http"

	"github.com/nimbo1999/financeiro/users/internal/services"
)

type CreateUserResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FullName  string `json:"full_name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	IsAdmin   bool   `json:"is_admin"`
}

type GetUserResponse struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	FullName  string  `json:"full_name"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
	DeletedAt *string `json:"deleted_at,omitempty"`
	IsAdmin   bool    `json:"is_admin"`
}

type ListUsersResponse struct {
	Data       []GetUserResponse `json:"data"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalPages int               `json:"total_pages"`
}

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"status"`
}

func ToErrorResponse(err error) ErrorResponse {
	errorsMap := map[error]ErrorResponse{
		services.ErrEmailIsRequired:    {Code: "ErrEmailIsRequired", Message: services.ErrEmailIsRequired.Error(), Status: http.StatusBadRequest},
		services.ErrInvalidUserData:    {Code: "ErrInvalidUserData", Message: services.ErrInvalidUserData.Error(), Status: http.StatusBadRequest},
		services.ErrFullNameIsRequired: {Code: "ErrFullNameIsRequired", Message: services.ErrFullNameIsRequired.Error(), Status: http.StatusBadRequest},
		services.ErrUserNotFound:       {Code: "ErrUserNotFound", Message: services.ErrUserNotFound.Error(), Status: http.StatusNotFound},
	}

	if resp, exists := errorsMap[err]; exists {
		return resp
	}
	return ErrorResponse{Code: "INTERNAL_SERVER_ERROR", Message: "Internal server error", Status: http.StatusInternalServerError}
}
