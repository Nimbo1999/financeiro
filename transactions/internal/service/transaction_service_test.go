package service

import (
	"context"
	"testing"
	"time"

	"github.com/nimbo1999/financeiro/commons"
	"github.com/nimbo1999/financeiro/transactions/internal/models"
	"github.com/nimbo1999/financeiro/transactions/internal/repository"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type TransactionsRepositoryMock struct {
	mock.Mock
}

func (m *TransactionsRepositoryMock) Create(ctx context.Context, transaction *models.Transaction) error {
	args := m.Called(ctx, transaction)
	return args.Error(0)
}

func (m *TransactionsRepositoryMock) List(ctx context.Context, pagination commons.Pagination) (*commons.PaginatedResult[models.Transaction], error) {
	args := m.Called(ctx, pagination)
	return args.Get(0).(*commons.PaginatedResult[models.Transaction]), args.Error(1)
}

func (m *TransactionsRepositoryMock) CreateBatch(ctx context.Context, entities []models.Transaction) error {
	args := m.Called(ctx, entities)
	return args.Error(0)
}

type TransactionsServiceTestSuite struct {
	suite.Suite
	repository repository.TransactionsRepository
}

func (s *TransactionsServiceTestSuite) SetupTest() {
	s.repository = new(TransactionsRepositoryMock)
}

func (s *TransactionsServiceTestSuite) TestCreate_Success() {
	// Arrange
	ctx := context.Background()
	service := NewTransactionService(s.repository)
	now := time.Now()
	userID := "user123"
	description := "Test Transaction"
	amount := 100.0

	// Set expectations
	s.repository.(*TransactionsRepositoryMock).On("Create", mock.Anything, mock.Anything).Return(nil)

	// Act
	err := service.Create(ctx, userID, description, amount, now)

	// Assert
	s.NoError(err)
}

func (s *TransactionsServiceTestSuite) TestCreate_ErrorOnInvalidTransaction() {
	// Arrange
	ctx := context.Background()
	service := NewTransactionService(s.repository)
	now := time.Now()
	userID := ""
	description := "Test Transaction"
	amount := 100.0

	// Set expectations
	s.repository.(*TransactionsRepositoryMock).On("Create", mock.Anything, mock.Anything).Return(nil)

	// Act
	err := service.Create(ctx, userID, description, amount, now)

	// Assert
	s.Error(err)
	s.Equal(ErrInvalidTransaction, err)
}

func (s *TransactionsServiceTestSuite) TestList_Success() {
	// Arrange
	ctx := context.Background()
	service := NewTransactionService(s.repository)
	data := []models.Transaction{
		*models.NewTransaction("user1", "Desc1", 50.0, time.Now()),
		*models.NewTransaction("user2", "Desc2", 75.0, time.Now()),
	}
	pagination := commons.NewPagination(1, 10, "updated_at", commons.OrderDesc, "")
	paginatedResult := commons.NewPaginatedResult(data, int64(len(data)), pagination)

	// Set expectations
	s.repository.(*TransactionsRepositoryMock).On("List", mock.Anything, mock.Anything).
		Return(paginatedResult, nil)

	// Act
	result, err := service.List(ctx, pagination)

	// Assert
	s.NoError(err)
	if s.NotEmpty(result) {
		s.Equal(1, result.Page)
		s.Equal(10, result.PageSize)
		s.Equal(int64(2), result.Total)
		s.Equal(1, result.TotalPages)
		s.Equal(data, result.Data)
	}
}

func Test_TransactionsServiceSuite(t *testing.T) {
	suite.Run(t, new(TransactionsServiceTestSuite))
}
