package messaging

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const userCreatedEventType = "user.created"

// UserCreatedEvent represents the event published when a user is created
type UserCreatedEvent struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	Timestamp time.Time       `json:"timestamp"`
	Data      UserCreatedData `json:"data"`
}

type UserCreatedData struct {
	UserID    string `json:"user_id"`
	UserEmail string `json:"user_email"`
	Name      string `json:"name"`
}

// ToJSON serializes the event to JSON
func (e *UserCreatedEvent) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// GetRoutingKey returns the routing key for this event
func (e *UserCreatedEvent) GetRoutingKey() string {
	return userCreatedEventType
}

// NewUserCreatedEvent creates a new user created event
func NewUserCreatedEvent(userID, email, name string) *UserCreatedEvent {
	return &UserCreatedEvent{
		ID:        uuid.NewString(),
		Type:      userCreatedEventType,
		Timestamp: time.Now().UTC(),
		Data: UserCreatedData{
			UserID:    userID,
			UserEmail: email,
			Name:      name,
		},
	}
}
