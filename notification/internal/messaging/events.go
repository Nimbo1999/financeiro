package messaging

import (
	"time"

	"github.com/nimbo1999/financeiro/notification/internal/models"
)

// OTPEmailEvent represents the authentication code requested event
type OTPEmailEvent struct {
	ID        string       `json:"id"`
	Type      string       `json:"type"`
	Timestamp time.Time    `json:"timestamp"`
	Source    string       `json:"source"`
	Version   string       `json:"version"`
	Data      OTPEmailData `json:"data"`
}

type OTPEmailData struct {
	UserID    string    `json:"user_id"`
	UserEmail string    `json:"user_email"`
	AuthCode  string    `json:"auth_code"`
	CodeID    string    `json:"code_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

// ToModel converts messaging event to model event
func (e *OTPEmailEvent) ToModel() *models.OTPEmailEvent {
	return &models.OTPEmailEvent{
		ID:        e.ID,
		Type:      e.Type,
		Timestamp: e.Timestamp,
		Source:    e.Source,
		Version:   e.Version,
		Data: models.OTPEmailData{
			UserID:    e.Data.UserID,
			UserEmail: e.Data.UserEmail,
			AuthCode:  e.Data.AuthCode,
			CodeID:    e.Data.CodeID,
			ExpiresAt: e.Data.ExpiresAt,
		},
	}
}

// WelcomeEmailEvent represents a user creation event
type WelcomeEmailEvent struct {
	ID        string           `json:"id"`
	Type      string           `json:"type"`
	Timestamp time.Time        `json:"timestamp"`
	Source    string           `json:"source"`
	Version   string           `json:"version"`
	Data      WelcomeEmailData `json:"data"`
}

type WelcomeEmailData struct {
	UserID    string    `json:"user_id"`
	UserEmail string    `json:"user_email"`
	Name      string    `json:"user_name,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// ToModel converts messaging event to model event
func (e *WelcomeEmailEvent) ToModel() *models.WelcomeEmailEvent {
	return &models.WelcomeEmailEvent{
		ID:        e.ID,
		Type:      e.Type,
		Timestamp: e.Timestamp,
		Data: models.WelcomeEmailData{
			UserID:    e.Data.UserID,
			UserEmail: e.Data.UserEmail,
			Name:      e.Data.Name,
		},
	}
}
