package repository

import (
	"context"

	"github.com/nimbo1999/financeiro/authentication/internal/models"
)

type AuthCodeRepository interface {
	// Create stores a new auth code in the repository
	Create(ctx context.Context, authCode *models.AuthCode) error

	// FindByUserId retrieves the most recent auth code by its user ID
	FindByUserID(ctx context.Context, userId string) (*models.AuthCode, error)

	// MarkAsUsed marks an auth code as used by setting the UsedAt timestamp
	MarkAsUsed(ctx context.Context, id string) error

	// CleanExpired removes all expired auth codes from the repository
	CleanExpired(ctx context.Context) error
}
