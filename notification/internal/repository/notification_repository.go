package repository

import (
	"context"

	"notification/internal/models"

	"gorm.io/gorm"
)

type NotificationRepository interface {
	Create(ctx context.Context, notification *models.Notification) error
	MarkAsSent(ctx context.Context, id string) error
	MarkAsFailed(ctx context.Context, id string, reason string) error
	IncrementRetryCount(ctx context.Context, id string) error
}

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(ctx context.Context, notification *models.Notification) error {
	return r.db.WithContext(ctx).Create(notification).Error
}

func (r *notificationRepository) MarkAsSent(ctx context.Context, id string) error {
	// Will be implemented in Step 2
	return nil
}

func (r *notificationRepository) MarkAsFailed(ctx context.Context, id string, reason string) error {
	// Will be implemented in Step 2
	return nil
}

func (r *notificationRepository) IncrementRetryCount(ctx context.Context, id string) error {
	// Will be implemented in Step 2
	return nil
}
