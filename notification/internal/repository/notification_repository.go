package repository

import (
	"context"
	"time"

	"github.com/nimbo1999/financeiro/notification/internal/models"

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
	return r.db.WithContext(ctx).Omit("ID").Create(notification).Error
}

func (r *notificationRepository) MarkAsSent(ctx context.Context, id string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     "sent",
			"sent_at":    now,
			"updated_at": now,
		}).Error
}

func (r *notificationRepository) MarkAsFailed(ctx context.Context, id string, reason string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":        "failed",
			"failed_reason": reason,
			"updated_at":    now,
		}).Error
}

func (r *notificationRepository) IncrementRetryCount(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("id = ?", id).
		UpdateColumn("retry_count", gorm.Expr("retry_count + 1")).
		Error
}
