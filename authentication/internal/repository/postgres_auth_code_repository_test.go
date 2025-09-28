package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/nimbo1999/financeiro/authentication/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostgresAuthCodeRepositoryTestSuite struct {
	suite.Suite
	db   *gorm.DB
	mock sqlmock.Sqlmock
	repo AuthCodeRepository
}

func (suite *PostgresAuthCodeRepositoryTestSuite) SetupTest() {
	var err error
	var sqlDB *sql.DB

	sqlDB, suite.mock, err = sqlmock.New()
	suite.Require().NoError(err)

	suite.db, err = gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{})
	suite.Require().NoError(err)

	suite.repo = NewPostgresAuthCodeRepository(suite.db)
}

func (suite *PostgresAuthCodeRepositoryTestSuite) TearDownTest() {
	suite.mock.ExpectationsWereMet()
}

func TestPostgresAuthCodeRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(PostgresAuthCodeRepositoryTestSuite))
}

// Create method tests
func (suite *PostgresAuthCodeRepositoryTestSuite) TestCreate_Success() {
	authCode := &models.AuthCode{
		ID:        "test-id",
		UserID:    "user-123",
		Code:      "123456",
		ExpiresAt: time.Now().Add(5 * time.Minute),
		CreatedAt: time.Now(),
	}

	suite.mock.ExpectBegin()
	suite.mock.ExpectExec(`INSERT INTO "auth_codes"`).
		WithArgs(authCode.ID, authCode.UserID, authCode.Code, authCode.ExpiresAt, sqlmock.AnyArg(), authCode.CreatedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))
	suite.mock.ExpectCommit()

	err := suite.repo.Create(context.Background(), authCode)

	assert.NoError(suite.T(), err)
}

func (suite *PostgresAuthCodeRepositoryTestSuite) TestCreate_NilAuthCode() {
	err := suite.repo.Create(context.Background(), nil)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "auth code cannot be nil")
}

func (suite *PostgresAuthCodeRepositoryTestSuite) TestCreate_DatabaseError() {
	authCode := &models.AuthCode{
		ID:        "test-id",
		UserID:    "user-123",
		Code:      "123456",
		ExpiresAt: time.Now().Add(5 * time.Minute),
		CreatedAt: time.Now(),
	}

	suite.mock.ExpectBegin()
	suite.mock.ExpectExec(`INSERT INTO "auth_codes"`).
		WithArgs(authCode.ID, authCode.UserID, authCode.Code, authCode.ExpiresAt, sqlmock.AnyArg(), authCode.CreatedAt).
		WillReturnError(errors.New("database connection failed"))
	suite.mock.ExpectRollback()

	err := suite.repo.Create(context.Background(), authCode)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "database connection failed")
}

// FindByUserID method tests
func (suite *PostgresAuthCodeRepositoryTestSuite) TestFindByUserID_Success() {
	userID := "user-123"
	now := time.Now()
	expectedAuthCode := models.AuthCode{
		ID:        "test-id",
		UserID:    userID,
		Code:      "123456",
		ExpiresAt: now.Add(5 * time.Minute),
		CreatedAt: now,
	}

	rows := sqlmock.NewRows([]string{"id", "user_id", "code", "expires_at", "used_at", "created_at"}).
		AddRow(expectedAuthCode.ID, expectedAuthCode.UserID, expectedAuthCode.Code,
			expectedAuthCode.ExpiresAt, nil, expectedAuthCode.CreatedAt)

	suite.mock.ExpectQuery(`SELECT \* FROM "auth_codes" WHERE user_id = \$1 ORDER BY created_at DESC,"auth_codes"\."id" LIMIT \$2`).
		WithArgs(userID, 1).
		WillReturnRows(rows)

	authCode, err := suite.repo.FindByUserID(context.Background(), userID)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), authCode)
	assert.Equal(suite.T(), expectedAuthCode.ID, authCode.ID)
	assert.Equal(suite.T(), expectedAuthCode.UserID, authCode.UserID)
	assert.Equal(suite.T(), expectedAuthCode.Code, authCode.Code)
}

func (suite *PostgresAuthCodeRepositoryTestSuite) TestFindByUserID_EmptyUserID() {
	authCode, err := suite.repo.FindByUserID(context.Background(), "")

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), authCode)
	assert.Equal(suite.T(), ErrUserIdEmpty, err)
}

func (suite *PostgresAuthCodeRepositoryTestSuite) TestFindByUserID_NotFound() {
	userID := "user-123"

	suite.mock.ExpectQuery(`SELECT \* FROM "auth_codes" WHERE user_id = \$1 ORDER BY created_at DESC,"auth_codes"\."id" LIMIT \$2`).
		WithArgs(userID, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	authCode, err := suite.repo.FindByUserID(context.Background(), userID)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), authCode)
	assert.Equal(suite.T(), ErrAuthCodeNotFound, err)
}

func (suite *PostgresAuthCodeRepositoryTestSuite) TestFindByUserID_DatabaseError() {
	userID := "user-123"

	suite.mock.ExpectQuery(`SELECT \* FROM "auth_codes" WHERE user_id = \$1 ORDER BY created_at DESC,"auth_codes"\."id" LIMIT \$2`).
		WithArgs(userID, 1).
		WillReturnError(errors.New("database connection failed"))

	authCode, err := suite.repo.FindByUserID(context.Background(), userID)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), authCode)
	assert.Contains(suite.T(), err.Error(), "database connection failed")
}

// MarkAsUsed method tests
func (suite *PostgresAuthCodeRepositoryTestSuite) TestMarkAsUsed_Success() {
	authCodeID := "test-id"

	suite.mock.ExpectBegin()
	suite.mock.ExpectExec(`UPDATE "auth_codes" SET "used_at"=\$1 WHERE id = \$2 AND used_at IS NULL`).
		WithArgs(sqlmock.AnyArg(), authCodeID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	suite.mock.ExpectCommit()

	err := suite.repo.MarkAsUsed(context.Background(), authCodeID)

	assert.NoError(suite.T(), err)
}

func (suite *PostgresAuthCodeRepositoryTestSuite) TestMarkAsUsed_EmptyID() {
	err := suite.repo.MarkAsUsed(context.Background(), "")

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrAuthCodeIdEmpty, err)
}

func (suite *PostgresAuthCodeRepositoryTestSuite) TestMarkAsUsed_NotFound() {
	authCodeID := "test-id"

	suite.mock.ExpectBegin()
	suite.mock.ExpectExec(`UPDATE "auth_codes" SET "used_at"=\$1 WHERE id = \$2 AND used_at IS NULL`).
		WithArgs(sqlmock.AnyArg(), authCodeID).
		WillReturnResult(sqlmock.NewResult(0, 0))
	suite.mock.ExpectCommit()

	err := suite.repo.MarkAsUsed(context.Background(), authCodeID)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrAuthCodeNotFound, err)
}

func (suite *PostgresAuthCodeRepositoryTestSuite) TestMarkAsUsed_DatabaseError() {
	authCodeID := "test-id"

	suite.mock.ExpectBegin()
	suite.mock.ExpectExec(`UPDATE "auth_codes" SET "used_at"=\$1 WHERE id = \$2 AND used_at IS NULL`).
		WithArgs(sqlmock.AnyArg(), authCodeID).
		WillReturnError(errors.New("database connection failed"))
	suite.mock.ExpectRollback()

	err := suite.repo.MarkAsUsed(context.Background(), authCodeID)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "database connection failed")
}

// CleanExpired method tests
func (suite *PostgresAuthCodeRepositoryTestSuite) TestCleanExpired_Success() {
	suite.mock.ExpectBegin()
	suite.mock.ExpectExec(`DELETE FROM "auth_codes" WHERE expires_at < \$1`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 3))
	suite.mock.ExpectCommit()

	err := suite.repo.CleanExpired(context.Background())

	assert.NoError(suite.T(), err)
}

func (suite *PostgresAuthCodeRepositoryTestSuite) TestCleanExpired_DatabaseError() {
	suite.mock.ExpectBegin()
	suite.mock.ExpectExec(`DELETE FROM "auth_codes" WHERE expires_at < \$1`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnError(errors.New("database connection failed"))
	suite.mock.ExpectRollback()

	err := suite.repo.CleanExpired(context.Background())

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "database connection failed")
}

// Table-driven tests for comprehensive coverage
func (suite *PostgresAuthCodeRepositoryTestSuite) TestCreate_TableDriven() {
	testCases := []struct {
		name        string
		authCode    *models.AuthCode
		setupMock   func()
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid AuthCode",
			authCode: &models.AuthCode{
				ID:        "valid-id",
				UserID:    "user-123",
				Code:      "123456",
				ExpiresAt: time.Now().Add(5 * time.Minute),
				CreatedAt: time.Now(),
			},
			setupMock: func() {
				suite.mock.ExpectBegin()
				suite.mock.ExpectExec(`INSERT INTO "auth_codes"`).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
						sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
				suite.mock.ExpectCommit()
			},
			expectError: false,
		},
		{
			name:        "Nil AuthCode",
			authCode:    nil,
			setupMock:   func() {},
			expectError: true,
			errorMsg:    "auth code cannot be nil",
		},
		{
			name: "Database Error",
			authCode: &models.AuthCode{
				ID:        "error-id",
				UserID:    "user-123",
				Code:      "123456",
				ExpiresAt: time.Now().Add(5 * time.Minute),
				CreatedAt: time.Now(),
			},
			setupMock: func() {
				suite.mock.ExpectBegin()
				suite.mock.ExpectExec(`INSERT INTO "auth_codes"`).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
						sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(errors.New("constraint violation"))
				suite.mock.ExpectRollback()
			},
			expectError: true,
			errorMsg:    "constraint violation",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			tc.setupMock()

			err := suite.repo.Create(context.Background(), tc.authCode)

			if tc.expectError {
				assert.Error(suite.T(), err)
				if tc.errorMsg != "" {
					assert.Contains(suite.T(), err.Error(), tc.errorMsg)
				}
			} else {
				assert.NoError(suite.T(), err)
			}
		})
	}
}

// Benchmark tests
func BenchmarkCreate(b *testing.B) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		b.Fatal(err)
	}
	defer sqlDB.Close()

	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		b.Fatal(err)
	}

	repo := NewPostgresAuthCodeRepository(db)
	authCode := &models.AuthCode{
		ID:        "bench-id",
		UserID:    "user-123",
		Code:      "123456",
		ExpiresAt: time.Now().Add(5 * time.Minute),
		CreatedAt: time.Now(),
	}

	for i := 0; i < b.N; i++ {
		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO "auth_codes"`).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		_ = repo.Create(context.Background(), authCode)
	}
}

// Test context cancellation
func (suite *PostgresAuthCodeRepositoryTestSuite) TestContextCancellation() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	authCode := &models.AuthCode{
		ID:        "test-id",
		UserID:    "user-123",
		Code:      "123456",
		ExpiresAt: time.Now().Add(5 * time.Minute),
		CreatedAt: time.Now(),
	}

	// Test Create with cancelled context
	err := suite.repo.Create(ctx, authCode)
	// Note: The actual behavior depends on the GORM implementation
	// This test ensures the context is properly passed through
	suite.ErrorAs(err, &context.Canceled)

	// Test FindByUserID with cancelled context
	_, err = suite.repo.FindByUserID(ctx, "user-123")
	// Should handle context cancellation gracefully
	suite.ErrorAs(err, &context.Canceled)

	// Test MarkAsUsed with cancelled context
	err = suite.repo.MarkAsUsed(ctx, "test-id")
	// Should handle context cancellation gracefully
	suite.ErrorAs(err, &context.Canceled)

	// Test CleanExpired with cancelled context
	err = suite.repo.CleanExpired(ctx)
	// Should handle context cancellation gracefully
	suite.ErrorAs(err, &context.Canceled)
}
