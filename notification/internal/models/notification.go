package models

import (
	"encoding/json"
	"time"
)

type Notification struct {
	ID                  string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID              string
	Email               string
	NotificationSubject string // 'notification.welcome', 'notification.otp'
	NotificationType    string // 'user.created', 'auth.code.requested'
	Status              string // 'pending', 'sent', 'failed'
	Subject             string
	SentAt              *time.Time
	FailedReason        string
	RetryCount          int
	EventData           json.RawMessage
	CreatedAt           time.Time
	UpdatedAt           time.Time
}
