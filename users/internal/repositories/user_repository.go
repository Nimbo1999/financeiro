package repositories

import (
	"github.com/nimbo1999/financeiro/users/internal/models"
	"gorm.io/gorm"
)

type PaginationParams struct {
	Page     int
	PageSize int
	OrderBy  string
	Sort     string
}

type PaginatedResult struct {
	Users      []models.User
	Total      int64
	Page       int
	PageSize   int
	TotalPages int
}

type UserRepository interface {
	Create(user *models.User) error
	FindByID(id string) (*models.User, error)
	Update(user *models.User) error
	Delete(id string) error
	List(params *PaginationParams) (*PaginatedResult, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) Create(user *models.User) error {
	return r.db.Omit("ID").Create(user).Error
}

func (r *userRepository) FindByID(id string) (*models.User, error) {
	var user models.User
	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) List(params *PaginationParams) (*PaginatedResult, error) {
	if params == nil {
		params = &PaginationParams{
			Page:     1,
			PageSize: 10,
			OrderBy:  "updated_at",
			Sort:     "DESC",
		}
	}

	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 10
	}
	if params.OrderBy == "" {
		params.OrderBy = "updated_at"
	}
	if params.Sort == "" {
		params.Sort = "DESC"
	}

	var users []models.User
	var total int64

	err := r.db.Model(&models.User{}).Count(&total).Error
	if err != nil {
		return nil, err
	}

	offset := (params.Page - 1) * params.PageSize
	orderClause := params.OrderBy + " " + params.Sort
	err = r.db.Order(orderClause).
		Offset(offset).
		Limit(params.PageSize).
		Find(&users).Error
	if err != nil {
		return nil, err
	}

	totalPages := int((total + int64(params.PageSize) - 1) / int64(params.PageSize))

	return &PaginatedResult{
		Users:      users,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}, nil
}

func (r *userRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&models.User{}).Error
}
