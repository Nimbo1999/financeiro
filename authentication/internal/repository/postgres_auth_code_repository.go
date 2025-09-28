package repository

import (
	"context"
	"errors"
	"time"

	"github.com/nimbo1999/financeiro/authentication/internal/models"
	"gorm.io/gorm"
)

var (
	ErrAuthCodeNotFound = errors.New("auth code not found")
	ErrAuthCodeExpired  = errors.New("auth code expired")
	ErrAuthCodeUsed     = errors.New("auth code already used")
	ErrUserIdEmpty      = errors.New("user ID cannot be empty")
	ErrAuthCodeIdEmpty  = errors.New("auth code ID cannot be empty")
)

type PostgresAuthCodeRepository struct {
	db *gorm.DB
}

func NewPostgresAuthCodeRepository(db *gorm.DB) AuthCodeRepository {
	return &PostgresAuthCodeRepository{
		db: db,
	}
}

func (r *PostgresAuthCodeRepository) Create(ctx context.Context, authCode *models.AuthCode) error {
	if authCode == nil {
		return errors.New("auth code cannot be nil")
	}
	return r.db.WithContext(ctx).Create(authCode).Error
}

/*
This function just retrieves the AuthCode, but is resposibility of the service to check if
it is expired or if it has already being used.
*/
func (r *PostgresAuthCodeRepository) FindByUserID(ctx context.Context, userId string) (*models.AuthCode, error) {
	if userId == "" {
		return nil, ErrUserIdEmpty
	}

	var authCode models.AuthCode
	result := r.db.WithContext(ctx).
		Where("user_id = ?", userId).
		Order("created_at DESC").
		First(&authCode)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrAuthCodeNotFound
		}
		return nil, result.Error
	}

	// The service layer should handle these checks
	// // Check if the code is expired
	// if time.Now().After(authCode.ExpiresAt) {
	// 	return nil, ErrAuthCodeExpired
	// }

	// // Check if the code is already used
	// if authCode.UsedAt != nil {
	// 	return nil, ErrAuthCodeUsed
	// }
	return &authCode, nil
}

func (r *PostgresAuthCodeRepository) MarkAsUsed(ctx context.Context, id string) error {
	if id == "" {
		return ErrAuthCodeIdEmpty
	}

	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&models.AuthCode{}).
		Where("id = ? AND used_at IS NULL", id).
		Update("used_at", now)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrAuthCodeNotFound
	}
	return nil
}

func (r *PostgresAuthCodeRepository) CleanExpired(ctx context.Context) error {
	result := r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&models.AuthCode{})

	return result.Error
}
