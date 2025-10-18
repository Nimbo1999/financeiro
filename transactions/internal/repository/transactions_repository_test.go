package repository

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/nimbo1999/financeiro/commons"
	"github.com/nimbo1999/financeiro/transactions/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type TransactionsRepositoryTestSuite struct {
	suite.Suite
	db         *gorm.DB
	sqlDB      *sql.DB
	mock       sqlmock.Sqlmock
	repository TransactionsRepository
}

func (suite *TransactionsRepositoryTestSuite) SetupTest() {
	var err error
	suite.sqlDB, suite.mock, err = sqlmock.New()
	assert.NoError(suite.T(), err)

	dialector := postgres.New(postgres.Config{
		Conn:       suite.sqlDB,
		DriverName: "postgres",
	})

	suite.db, err = gorm.Open(dialector, &gorm.Config{})
	assert.NoError(suite.T(), err)

	suite.repository = NewTransactionsRepository(suite.db)
}

func (suite *TransactionsRepositoryTestSuite) TearDownTest() {
	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
	suite.sqlDB.Close()
}

func TestTransactionsRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(TransactionsRepositoryTestSuite))
}

// Test Create method

func (suite *TransactionsRepositoryTestSuite) TestCreate_Success() {
	// Arrange
	ctx := context.Background()
	transaction := &models.Transaction{
		ID:          "txn-123",
		Date:        time.Now(),
		Amount:      100.50,
		Description: "Test transaction",
		UserID:      "user-123",
	}

	// Create method calls CreateBatch with a slice, so we need to expect 2 sets of args (empty + actual)
	suite.mock.ExpectBegin()
	suite.mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "transactions"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	suite.mock.ExpectCommit()

	// Act
	err := suite.repository.Create(ctx, transaction)

	// Assert
	assert.NoError(suite.T(), err)
}

func (suite *TransactionsRepositoryTestSuite) TestCreate_SuccessNegativeAmmount() {
	// Arrange
	ctx := context.Background()
	transaction := &models.Transaction{
		ID:          "txn-123",
		Date:        time.Now(),
		Amount:      -100.50,
		Description: "Test transaction",
		UserID:      "user-123",
	}

	suite.mock.ExpectBegin()
	suite.mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "transactions"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, transaction.Date, transaction.Amount, transaction.Description, transaction.UserID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("bc7e3a8a-6ea3-4552-ae69-10df10b94abb"))
	suite.mock.ExpectCommit()

	// Act
	err := suite.repository.Create(ctx, transaction)

	// Assert
	assert.NoError(suite.T(), err)
}

func (suite *TransactionsRepositoryTestSuite) TestCreate_DatabaseError() {
	// Arrange
	ctx := context.Background()
	transaction := &models.Transaction{
		ID:          "txn-123",
		Date:        time.Now(),
		Amount:      100.50,
		Description: "Test transaction",
		UserID:      "user-123",
	}

	suite.mock.ExpectBegin()
	suite.mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "transactions"`)).
		WillReturnError(sql.ErrConnDone)
	suite.mock.ExpectRollback()

	// Act
	err := suite.repository.Create(ctx, transaction)

	// Assert
	assert.Error(suite.T(), err)
}

// Test CreateBatch method

func (suite *TransactionsRepositoryTestSuite) TestCreateBatch_Success() {
	// Arrange
	ctx := context.Background()
	transactions := []models.Transaction{
		{
			ID:          "txn-123",
			Date:        time.Now(),
			Amount:      100.50,
			Description: "Test transaction 1",
			UserID:      "user-123",
		},
		{
			ID:          "txn-456",
			Date:        time.Now(),
			Amount:      200.75,
			Description: "Test transaction 2",
			UserID:      "user-123",
		},
	}

	suite.mock.ExpectBegin()
	suite.mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "transactions"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2))
	suite.mock.ExpectCommit()

	// Act
	err := suite.repository.CreateBatch(ctx, transactions)

	// Assert
	assert.NoError(suite.T(), err)
}

func (suite *TransactionsRepositoryTestSuite) TestCreateBatch_DatabaseError() {
	// Arrange
	ctx := context.Background()
	transactions := []models.Transaction{
		{
			ID:          "txn-123",
			Date:        time.Now(),
			Amount:      100.50,
			Description: "Test transaction",
			UserID:      "user-123",
		},
	}

	suite.mock.ExpectBegin()
	suite.mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "transactions"`)).
		WillReturnError(sql.ErrConnDone)
	suite.mock.ExpectRollback()

	// Act
	err := suite.repository.CreateBatch(ctx, transactions)

	// Assert
	assert.Error(suite.T(), err)
}

// Test List method

func (suite *TransactionsRepositoryTestSuite) TestList_Success() {
	// Arrange
	ctx := context.Background()
	pagination := commons.NewPagination(1, 10, "created_at", commons.OrderDesc, "")

	// GORM adds WHERE deleted_at IS NULL for soft deletes
	suite.mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "transactions" WHERE "transactions"."deleted_at" IS NULL`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "date", "amount", "description", "user_id"}).
		AddRow("txn-123", time.Now(), time.Now(), nil, time.Now(), 100.50, "Transaction 1", "user-123").
		AddRow("txn-456", time.Now(), time.Now(), nil, time.Now(), 200.75, "Transaction 2", "user-123")

	suite.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "transactions" WHERE "transactions"."deleted_at" IS NULL ORDER BY created_at desc LIMIT`)).
		WillReturnRows(rows)

	// Act
	result, err := suite.repository.List(ctx, pagination)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), int64(2), result.Total)
	assert.Equal(suite.T(), 2, len(result.Data))
	assert.Equal(suite.T(), "txn-123", result.Data[0].ID)
	assert.Equal(suite.T(), "txn-456", result.Data[1].ID)
}

func (suite *TransactionsRepositoryTestSuite) TestList_WithSearchQuery() {
	// Arrange
	ctx := context.Background()
	pagination := commons.NewPagination(1, 10, "created_at", commons.OrderDesc, "groceries")

	// GORM adds WHERE deleted_at IS NULL after the custom WHERE clause
	suite.mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "transactions" WHERE description ILIKE $1 AND "transactions"."deleted_at" IS NULL`)).
		WithArgs("%groceries%").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "t_date", "amount", "t_description", "user_id"}).
		AddRow("txn-123", time.Now(), time.Now(), nil, time.Now(), 50.00, "Groceries shopping", "user-123")

	suite.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "transactions" WHERE description ILIKE $1 AND "transactions"."deleted_at" IS NULL ORDER BY created_at desc LIMIT`)).
		WillReturnRows(rows)

	// Act
	result, err := suite.repository.List(ctx, pagination)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), int64(1), result.Total)
	assert.Equal(suite.T(), 1, len(result.Data))
	assert.Contains(suite.T(), result.Data[0].Description, "Groceries")
}

func (suite *TransactionsRepositoryTestSuite) TestList_EmptyResult() {
	// Arrange
	ctx := context.Background()
	pagination := commons.NewPagination(1, 10, "created_at", commons.OrderDesc, "")

	suite.mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "transactions" WHERE "transactions"."deleted_at" IS NULL`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	// Act
	result, err := suite.repository.List(ctx, pagination)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), int64(0), result.Total)
	assert.Equal(suite.T(), 0, len(result.Data))
}

func (suite *TransactionsRepositoryTestSuite) TestList_CountError() {
	// Arrange
	ctx := context.Background()
	pagination := commons.NewPagination(1, 10, "created_at", commons.OrderDesc, "")

	suite.mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "transactions" WHERE "transactions"."deleted_at" IS NULL`)).
		WillReturnError(sql.ErrConnDone)

	// Act
	result, err := suite.repository.List(ctx, pagination)

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
}

func (suite *TransactionsRepositoryTestSuite) TestList_FindError() {
	// Arrange
	ctx := context.Background()
	pagination := commons.NewPagination(1, 10, "created_at", commons.OrderDesc, "")

	suite.mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "transactions" WHERE "transactions"."deleted_at" IS NULL`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	suite.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "transactions" WHERE "transactions"."deleted_at" IS NULL ORDER BY created_at desc LIMIT`)).
		WillReturnError(sql.ErrConnDone)

	// Act
	result, err := suite.repository.List(ctx, pagination)

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
}

func (suite *TransactionsRepositoryTestSuite) TestList_WithPagination() {
	// Arrange
	ctx := context.Background()
	pagination := commons.NewPagination(2, 5, "amount", commons.OrderAsc, "")

	suite.mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "transactions" WHERE "transactions"."deleted_at" IS NULL`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(15))

	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "date", "amount", "description", "user_id"}).
		AddRow("txn-6", time.Now(), time.Now(), nil, time.Now(), 60.00, "Transaction 6", "user-123").
		AddRow("txn-7", time.Now(), time.Now(), nil, time.Now(), 70.00, "Transaction 7", "user-123").
		AddRow("txn-8", time.Now(), time.Now(), nil, time.Now(), 80.00, "Transaction 8", "user-123").
		AddRow("txn-9", time.Now(), time.Now(), nil, time.Now(), 90.00, "Transaction 9", "user-123").
		AddRow("txn-10", time.Now(), time.Now(), nil, time.Now(), 100.00, "Transaction 10", "user-123")

	suite.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "transactions" WHERE "transactions"."deleted_at" IS NULL ORDER BY amount asc LIMIT`)).
		WillReturnRows(rows)

	// Act
	result, err := suite.repository.List(ctx, pagination)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), int64(15), result.Total)
	assert.Equal(suite.T(), 5, len(result.Data))
	assert.Equal(suite.T(), 2, result.Page)
	assert.Equal(suite.T(), 5, result.PageSize)
	assert.Equal(suite.T(), 3, result.TotalPages)
}
