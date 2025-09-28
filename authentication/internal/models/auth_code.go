package models

import (
	"time"
)

type AuthCode struct {
	ID        string     `json:"id" gorm:"primaryKey"`
	UserID    string     `json:"user_id"`
	Code      string     `json:"code"`
	ExpiresAt time.Time  `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}
