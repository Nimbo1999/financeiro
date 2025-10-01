package models

import "time"

// WelcomeEmailEvent represents the user.created event for welcome emails
type WelcomeEmailEvent struct {
	ID        string           `json:"id"`
	Type      string           `json:"type"`
	Timestamp time.Time        `json:"timestamp"`
	Data      WelcomeEmailData `json:"data"`
}

type WelcomeEmailData struct {
	UserEmail string `json:"user_email"`
	Name      string `json:"name"`
}

// OTPEmailEvent matches authentication service's AuthCodeRequestedEvent
type OTPEmailEvent struct {
	ID        string       `json:"id"`
	Type      string       `json:"type"`
	Timestamp time.Time    `json:"timestamp"`
	Source    string       `json:"source"`
	Version   string       `json:"version"`
	Data      OTPEmailData `json:"data"`
}

type OTPEmailData struct {
	UserEmail string    `json:"user_email"`
	AuthCode  string    `json:"auth_code"`
	CodeID    string    `json:"code_id"`
	ExpiresAt time.Time `json:"expires_at"`
}
