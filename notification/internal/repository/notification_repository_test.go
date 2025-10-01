package repository

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"notification/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type NotificationRepositoryTestSuite struct {
	suite.Suite
	db         *gorm.DB
	mock       sqlmock.Sqlmock
	repository NotificationRepository
}

func (suite *NotificationRepositoryTestSuite) SetupTest() {
	var err error
	mockDB, mock, err := sqlmock.New()
	assert.NoError(suite.T(), err)

	suite.mock = mock

	dialector := postgres.New(postgres.Config{
		Conn:       mockDB,
		DriverName: "postgres",
	})

	suite.db, err = gorm.Open(dialector, &gorm.Config{})
	assert.NoError(suite.T(), err)

	suite.repository = NewNotificationRepository(suite.db)
}

func (suite *NotificationRepositoryTestSuite) TearDownTest() {
	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

func TestNotificationRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(NotificationRepositoryTestSuite))
}

// Test Create Success
func (suite *NotificationRepositoryTestSuite) TestCreate_Success() {
	// Arrange
	eventData := json.RawMessage(`{"user_email":"test@example.com","name":"Test User"}`)
	notification := &models.Notification{
		UserID:              "660e8400-e29b-41d4-a716-446655440000",
		Email:               "test@example.com",
		NotificationSubject: "notification.welcome",
		NotificationType:    "user.created",
		Status:              "pending",
		Subject:             "Welcome!",
		RetryCount:          0,
		EventData:           eventData,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	suite.mock.ExpectBegin()
	suite.mock.ExpectQuery(`INSERT INTO "notifications"`).
		WithArgs(
			notification.UserID,
			notification.Email,
			notification.NotificationSubject,
			notification.NotificationType,
			notification.Status,
			notification.Subject,
			notification.SentAt,
			notification.FailedReason, //
			notification.RetryCount,
			notification.EventData,
			sqlmock.AnyArg(), // created_at
			sqlmock.AnyArg(), // updated_at
		).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("770e8400-e29b-41d4-a716-446655440000"))
	suite.mock.ExpectCommit()

	// Act
	err := suite.repository.Create(context.Background(), notification)

	// Assert
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "770e8400-e29b-41d4-a716-446655440000", notification.ID)
}

// Test Create DatabaseError
func (suite *NotificationRepositoryTestSuite) TestCreate_DatabaseError() {
	// Arrange
	notification := &models.Notification{
		Email:  "test@example.com",
		Status: "pending",
	}

	suite.mock.ExpectBegin()
	suite.mock.ExpectQuery(`INSERT INTO "notifications"`).
		WillReturnError(gorm.ErrInvalidData)
	suite.mock.ExpectRollback()

	// Act
	err := suite.repository.Create(context.Background(), notification)

	// Assert
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), gorm.ErrInvalidData, err)
}

// Test MarkAsSent Success
func (suite *NotificationRepositoryTestSuite) TestMarkAsSent_Success() {
	// Arrange
	id := "550e8400-e29b-41d4-a716-446655440000"

	suite.mock.ExpectBegin()
	suite.mock.ExpectExec(`UPDATE "notifications" SET`).
		WithArgs(sqlmock.AnyArg(), "sent", sqlmock.AnyArg(), id).
		WillReturnResult(sqlmock.NewResult(1, 1))
	suite.mock.ExpectCommit()

	// Act
	err := suite.repository.MarkAsSent(context.Background(), id)

	// Assert
	assert.NoError(suite.T(), err)
}

// Test MarkAsSent DatabaseError
func (suite *NotificationRepositoryTestSuite) TestMarkAsSent_DatabaseError() {
	// Arrange
	id := "550e8400-e29b-41d4-a716-446655440000"

	suite.mock.ExpectBegin()
	suite.mock.ExpectExec(`UPDATE "notifications" SET`).
		WithArgs(sqlmock.AnyArg(), "sent", sqlmock.AnyArg(), id).
		WillReturnError(gorm.ErrRecordNotFound)
	suite.mock.ExpectRollback()

	// Act
	err := suite.repository.MarkAsSent(context.Background(), id)

	// Assert
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), gorm.ErrRecordNotFound, err)
}

// Test MarkAsFailed Success
func (suite *NotificationRepositoryTestSuite) TestMarkAsFailed_Success() {
	// Arrange
	id := "550e8400-e29b-41d4-a716-446655440000"
	reason := "SMTP connection failed"

	suite.mock.ExpectBegin()
	suite.mock.ExpectExec(`UPDATE "notifications" SET`).
		WithArgs(reason, "failed", sqlmock.AnyArg(), id).
		WillReturnResult(sqlmock.NewResult(1, 1))
	suite.mock.ExpectCommit()

	// Act
	err := suite.repository.MarkAsFailed(context.Background(), id, reason)

	// Assert
	assert.NoError(suite.T(), err)
}

// Test MarkAsFailed DatabaseError
func (suite *NotificationRepositoryTestSuite) TestMarkAsFailed_DatabaseError() {
	// Arrange
	id := "550e8400-e29b-41d4-a716-446655440000"
	reason := "SMTP connection failed"

	suite.mock.ExpectBegin()
	suite.mock.ExpectExec(`UPDATE "notifications" SET`).
		WithArgs(reason, "failed", sqlmock.AnyArg(), id).
		WillReturnError(gorm.ErrInvalidDB)
	suite.mock.ExpectRollback()

	// Act
	err := suite.repository.MarkAsFailed(context.Background(), id, reason)

	// Assert
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), gorm.ErrInvalidDB, err)
}

// Test IncrementRetryCount Success
func (suite *NotificationRepositoryTestSuite) TestIncrementRetryCount_Success() {
	// Arrange
	id := "550e8400-e29b-41d4-a716-446655440000"

	suite.mock.ExpectBegin()
	suite.mock.ExpectExec(`UPDATE "notifications" SET "retry_count"=retry_count \+ 1 WHERE id`).
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(1, 1))
	suite.mock.ExpectCommit()

	// Act
	err := suite.repository.IncrementRetryCount(context.Background(), id)

	// Assert
	assert.NoError(suite.T(), err)
}

// Test IncrementRetryCount DatabaseError
func (suite *NotificationRepositoryTestSuite) TestIncrementRetryCount_DatabaseError() {
	// Arrange
	id := "550e8400-e29b-41d4-a716-446655440000"

	suite.mock.ExpectBegin()
	suite.mock.ExpectExec(`UPDATE "notifications" SET "retry_count"=retry_count \+ 1 WHERE id`).
		WithArgs(id).
		WillReturnError(gorm.ErrInvalidTransaction)
	suite.mock.ExpectRollback()

	// Act
	err := suite.repository.IncrementRetryCount(context.Background(), id)

	// Assert
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), gorm.ErrInvalidTransaction, err)
}
