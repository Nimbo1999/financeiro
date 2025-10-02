package messaging

import (
	"encoding/json"
	"time"
)

// EventType represents the type of authentication event
type EventType string

const (
	EventTypeAuthCodeRequested EventType = "auth.code.requested"
	EventTypeAuthCodeVerified  EventType = "auth.code.verified"
	EventTypeAuthCodeExpired   EventType = "auth.code.expired"
	EventTypeAuthCodeFailed    EventType = "auth.code.failed"
)

// BaseEvent contains common fields for all events
type BaseEvent struct {
	ID        string    `json:"id"`
	Type      EventType `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	Source    string    `json:"source"`
	Version   string    `json:"version"`
}

// AuthCodeRequestedEvent is published when an authentication code is requested
type AuthCodeRequestedEvent struct {
	BaseEvent
	Data AuthCodeRequestedData `json:"data"`
}

type AuthCodeRequestedData struct {
	UserID    string    `json:"user_id"`
	UserEmail string    `json:"user_email"`
	AuthCode  string    `json:"auth_code"`
	CodeID    string    `json:"code_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

// AuthCodeVerifiedEvent is published when an authentication code is successfully verified
type AuthCodeVerifiedEvent struct {
	BaseEvent
	Data AuthCodeVerifiedData `json:"data"`
}

type AuthCodeVerifiedData struct {
	UserEmail            string    `json:"user_email"`
	UserID               string    `json:"user_id"`
	CodeID               string    `json:"code_id"`
	VerifiedAt           time.Time `json:"verified_at"`
	IsNewUser            bool      `json:"is_new_user"`
	AuthenticationMethod string    `json:"authentication_method"`
}

// AuthCodeExpiredEvent is published when an authentication code expires
type AuthCodeExpiredEvent struct {
	BaseEvent
	Data AuthCodeExpiredData `json:"data"`
}

type AuthCodeExpiredData struct {
	UserEmail string    `json:"user_email"`
	CodeID    string    `json:"code_id"`
	ExpiredAt time.Time `json:"expired_at"`
}

// AuthCodeFailedEvent is published when authentication code verification fails
type AuthCodeFailedEvent struct {
	BaseEvent
	Data AuthCodeFailedData `json:"data"`
}

type AuthCodeFailedData struct {
	UserEmail     string    `json:"user_email"`
	CodeID        string    `json:"code_id,omitempty"`
	FailedAt      time.Time `json:"failed_at"`
	Reason        string    `json:"reason"`
	AttemptsCount int       `json:"attempts_count,omitempty"`
}

// Event interface for all authentication events
type Event interface {
	GetType() EventType
	GetID() string
	GetTimestamp() time.Time
	ToJSON() ([]byte, error)
	GetRoutingKey() string
}

// NewAuthCodeRequestedEvent creates a new auth code requested event
func NewAuthCodeRequestedEvent(userID, userEmail, authCode, codeID string, expiresAt time.Time) *AuthCodeRequestedEvent {
	return &AuthCodeRequestedEvent{
		BaseEvent: BaseEvent{
			ID:        generateEventID(),
			Type:      EventTypeAuthCodeRequested,
			Timestamp: time.Now().UTC(),
			Source:    "authentication-service",
			Version:   "1.0",
		},
		Data: AuthCodeRequestedData{
			UserID:    userID,
			UserEmail: userEmail,
			AuthCode:  authCode,
			CodeID:    codeID,
			ExpiresAt: expiresAt,
		},
	}
}

// NewAuthCodeVerifiedEvent creates a new auth code verified event
func NewAuthCodeVerifiedEvent(userEmail, userID, codeID string, isNewUser bool) *AuthCodeVerifiedEvent {
	return &AuthCodeVerifiedEvent{
		BaseEvent: BaseEvent{
			ID:        generateEventID(),
			Type:      EventTypeAuthCodeVerified,
			Timestamp: time.Now().UTC(),
			Source:    "authentication-service",
			Version:   "1.0",
		},
		Data: AuthCodeVerifiedData{
			UserEmail:            userEmail,
			UserID:               userID,
			CodeID:               codeID,
			VerifiedAt:           time.Now().UTC(),
			IsNewUser:            isNewUser,
			AuthenticationMethod: "passwordless_code",
		},
	}
}

// NewAuthCodeExpiredEvent creates a new auth code expired event
func NewAuthCodeExpiredEvent(userEmail, codeID string) *AuthCodeExpiredEvent {
	return &AuthCodeExpiredEvent{
		BaseEvent: BaseEvent{
			ID:        generateEventID(),
			Type:      EventTypeAuthCodeExpired,
			Timestamp: time.Now().UTC(),
			Source:    "authentication-service",
			Version:   "1.0",
		},
		Data: AuthCodeExpiredData{
			UserEmail: userEmail,
			CodeID:    codeID,
			ExpiredAt: time.Now().UTC(),
		},
	}
}

// NewAuthCodeFailedEvent creates a new auth code failed event
func NewAuthCodeFailedEvent(userEmail, codeID, reason string, attemptsCount int) *AuthCodeFailedEvent {
	return &AuthCodeFailedEvent{
		BaseEvent: BaseEvent{
			ID:        generateEventID(),
			Type:      EventTypeAuthCodeFailed,
			Timestamp: time.Now().UTC(),
			Source:    "authentication-service",
			Version:   "1.0",
		},
		Data: AuthCodeFailedData{
			UserEmail:     userEmail,
			CodeID:        codeID,
			FailedAt:      time.Now().UTC(),
			Reason:        reason,
			AttemptsCount: attemptsCount,
		},
	}
}

// Implement Event interface for AuthCodeRequestedEvent
func (e *AuthCodeRequestedEvent) GetType() EventType      { return e.Type }
func (e *AuthCodeRequestedEvent) GetID() string           { return e.ID }
func (e *AuthCodeRequestedEvent) GetTimestamp() time.Time { return e.Timestamp }
func (e *AuthCodeRequestedEvent) GetRoutingKey() string   { return string(e.Type) }

func (e *AuthCodeRequestedEvent) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// Implement Event interface for AuthCodeVerifiedEvent
func (e *AuthCodeVerifiedEvent) GetType() EventType      { return e.Type }
func (e *AuthCodeVerifiedEvent) GetID() string           { return e.ID }
func (e *AuthCodeVerifiedEvent) GetTimestamp() time.Time { return e.Timestamp }
func (e *AuthCodeVerifiedEvent) GetRoutingKey() string   { return string(e.Type) }

func (e *AuthCodeVerifiedEvent) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// Implement Event interface for AuthCodeExpiredEvent
func (e *AuthCodeExpiredEvent) GetType() EventType      { return e.Type }
func (e *AuthCodeExpiredEvent) GetID() string           { return e.ID }
func (e *AuthCodeExpiredEvent) GetTimestamp() time.Time { return e.Timestamp }
func (e *AuthCodeExpiredEvent) GetRoutingKey() string   { return string(e.Type) }

func (e *AuthCodeExpiredEvent) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// Implement Event interface for AuthCodeFailedEvent
func (e *AuthCodeFailedEvent) GetType() EventType      { return e.Type }
func (e *AuthCodeFailedEvent) GetID() string           { return e.ID }
func (e *AuthCodeFailedEvent) GetTimestamp() time.Time { return e.Timestamp }
func (e *AuthCodeFailedEvent) GetRoutingKey() string   { return string(e.Type) }

func (e *AuthCodeFailedEvent) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// generateEventID generates a unique event ID
func generateEventID() string {
	// Using a simple timestamp-based ID for now
	// In production, you might want to use UUID or a more sophisticated ID generator
	return time.Now().Format("20060102150405.000000")
}
