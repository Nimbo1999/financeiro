package repository

import (
	"context"
	"fmt"

	"github.com/nimbo1999/financeiro/commons"
	"github.com/nimbo1999/financeiro/transactions/internal/models"
	"gorm.io/gorm"
)

type TransactionsRepository interface {
	Create(context.Context, *models.Transaction) error
	CreateBatch(context.Context, []models.Transaction) error
	List(context.Context, commons.Pagination) (*commons.PaginatedResult[models.Transaction], error)
}

type repository struct {
	db *gorm.DB
}

func (r *repository) Create(ctx context.Context, entity *models.Transaction) error {
	entities := []models.Transaction{}
	entities = append(entities, *entity)
	return r.CreateBatch(ctx, entities)
}

func (r *repository) CreateBatch(ctx context.Context, entities []models.Transaction) error {
	return r.db.WithContext(ctx).Omit("ID").Create(entities).Error
}

func (r *repository) List(ctx context.Context, pagination commons.Pagination) (*commons.PaginatedResult[models.Transaction], error) {
	var entities []models.Transaction
	var total int64

	query := r.db.Model(&models.Transaction{})
	if pagination.GetQuery() != "" {
		query = query.Where("description ILIKE ?", fmt.Sprintf("%%%s%%", pagination.GetQuery()))
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	if total == 0 {
		return commons.NewPaginatedResult([]models.Transaction{}, total, pagination), nil
	}

	err := query.Order(pagination.Order()).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Find(&entities).Error
	if err != nil {
		return nil, err
	}
	return commons.NewPaginatedResult(entities, total, pagination), nil
}

func NewTransactionsRepository(db *gorm.DB) TransactionsRepository {
	return &repository{db: db}
}
