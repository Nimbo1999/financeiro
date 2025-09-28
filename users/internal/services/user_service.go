package services

import (
	"errors"

	"github.com/nimbo1999/financeiro/users/internal/models"
	"github.com/nimbo1999/financeiro/users/internal/repositories"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidUserData    = errors.New("invalid user data")
	ErrEmailIsRequired    = errors.New("email is required")
	ErrFullNameIsRequired = errors.New("full name is required")
)

type PaginationParams = repositories.PaginationParams
type PaginatedResult = repositories.PaginatedResult

type CreateUserRequest struct {
	Email    string `json:"email"`
	FullName string `json:"full_name"`
}

type UpdateUserRequest struct {
	ID       string  `json:"id"`
	Email    *string `json:"email"`
	FullName *string `json:"full_name"`
}

type UserService interface {
	CreateUser(req *CreateUserRequest) (*models.User, error)
	GetUserByID(id string) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	UpdateUser(req *UpdateUserRequest) (*models.User, error)
	ListUsers(params *repositories.PaginationParams) (*repositories.PaginatedResult, error)
	DeleteUser(id string) error
}

type userService struct {
	userRepo repositories.UserRepository
}

func NewUserService(userRepo repositories.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

func (s *userService) CreateUser(req *CreateUserRequest) (*models.User, error) {
	if err := s.validateCreateUserRequest(req); err != nil {
		return nil, err
	}

	user := &models.User{
		Email:    req.Email,
		FullName: req.FullName,
	}

	err := s.userRepo.Create(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) validateCreateUserRequest(req *CreateUserRequest) error {
	if req == nil {
		return ErrInvalidUserData
	}
	if req.Email == "" {
		return ErrEmailIsRequired
	}
	if req.FullName == "" {
		return ErrFullNameIsRequired
	}
	return nil
}

func (s *userService) GetUserByID(id string) (*models.User, error) {
	if id == "" {
		return nil, ErrInvalidUserData
	}

	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, ErrUserNotFound
	}

	return user, nil
}

func (s *userService) GetUserByEmail(email string) (*models.User, error) {
	if email == "" {
		return nil, ErrInvalidUserData
	}

	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, ErrUserNotFound
	}

	return user, nil
}

func (s *userService) UpdateUser(req *UpdateUserRequest) (*models.User, error) {
	if req.ID == "" || req == nil {
		return nil, ErrInvalidUserData
	}

	user, err := s.userRepo.FindByID(req.ID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if req.Email != nil {
		user.Email = *req.Email
	}
	if req.FullName != nil {
		user.FullName = *req.FullName
	}

	err = s.userRepo.Update(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) ListUsers(params *repositories.PaginationParams) (*repositories.PaginatedResult, error) {
	if params != nil {
		if params.Page <= 0 {
			params.Page = 1
		}
		if params.PageSize <= 0 {
			params.PageSize = 10
		}
	}

	result, err := s.userRepo.List(params)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *userService) DeleteUser(id string) error {
	if id == "" {
		return ErrInvalidUserData
	}

	_, err := s.userRepo.FindByID(id)
	if err != nil {
		return ErrUserNotFound
	}

	return s.userRepo.Delete(id)
}
