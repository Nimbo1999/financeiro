package service

import (
	"context"
	"errors"
	"time"

	"github.com/nimbo1999/financeiro/commons"
	"github.com/nimbo1999/financeiro/transactions/internal/models"
	"github.com/nimbo1999/financeiro/transactions/internal/repository"
)

var ErrInvalidTransaction = errors.New("invalid transaction")

type TransactionService interface {
	Create(ctx context.Context, userID, description string, amount float64, date time.Time) error
	List(ctx context.Context, pagination commons.Pagination) (*commons.PaginatedResult[models.Transaction], error)
}

type service struct {
	repo repository.TransactionsRepository
}

func (s *service) Create(ctx context.Context, userID, description string, amount float64, date time.Time) error {
	transaction := models.NewTransaction(userID, description, amount, date)
	if !transaction.IsValid() {
		return ErrInvalidTransaction
	}
	return s.repo.Create(ctx, transaction)
}

func (s *service) List(ctx context.Context, pagination commons.Pagination) (*commons.PaginatedResult[models.Transaction], error) {
	return s.repo.List(ctx, pagination)
}

func NewTransactionService(repo repository.TransactionsRepository) TransactionService {
	return &service{repo: repo}
}
