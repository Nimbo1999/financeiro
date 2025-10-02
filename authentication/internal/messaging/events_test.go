package messaging

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type EventsTestSuite struct {
	suite.Suite
}

func TestEventsTestSuite(t *testing.T) {
	suite.Run(t, new(EventsTestSuite))
}

func (suite *EventsTestSuite) TestNewAuthCodeRequestedEvent() {
	userID := "user-123"
	userEmail := "test@example.com"
	authCode := "123456"
	codeID := "code-123"
	expiresAt := time.Now().Add(5 * time.Minute)

	event := NewAuthCodeRequestedEvent(userID, userEmail, authCode, codeID, expiresAt)

	assert.NotNil(suite.T(), event)
	assert.Equal(suite.T(), EventTypeAuthCodeRequested, event.Type)
	assert.Equal(suite.T(), "authentication-service", event.Source)
	assert.Equal(suite.T(), "1.0", event.Version)
	assert.NotEmpty(suite.T(), event.ID)
	assert.WithinDuration(suite.T(), time.Now(), event.Timestamp, time.Second)

	assert.Equal(suite.T(), userID, event.Data.UserID)
	assert.Equal(suite.T(), userEmail, event.Data.UserEmail)
	assert.Equal(suite.T(), authCode, event.Data.AuthCode)
	assert.Equal(suite.T(), codeID, event.Data.CodeID)
	assert.Equal(suite.T(), expiresAt, event.Data.ExpiresAt)
}

func (suite *EventsTestSuite) TestNewAuthCodeVerifiedEvent() {
	userEmail := "test@example.com"
	userID := "user-123"
	codeID := "code-123"
	isNewUser := true

	event := NewAuthCodeVerifiedEvent(userEmail, userID, codeID, isNewUser)

	assert.NotNil(suite.T(), event)
	assert.Equal(suite.T(), EventTypeAuthCodeVerified, event.Type)
	assert.Equal(suite.T(), "authentication-service", event.Source)
	assert.Equal(suite.T(), "1.0", event.Version)
	assert.NotEmpty(suite.T(), event.ID)

	assert.Equal(suite.T(), userEmail, event.Data.UserEmail)
	assert.Equal(suite.T(), userID, event.Data.UserID)
	assert.Equal(suite.T(), codeID, event.Data.CodeID)
	assert.Equal(suite.T(), isNewUser, event.Data.IsNewUser)
	assert.Equal(suite.T(), "passwordless_code", event.Data.AuthenticationMethod)
	assert.WithinDuration(suite.T(), time.Now(), event.Data.VerifiedAt, time.Second)
}

func (suite *EventsTestSuite) TestNewAuthCodeExpiredEvent() {
	userEmail := "test@example.com"
	codeID := "code-123"

	event := NewAuthCodeExpiredEvent(userEmail, codeID)

	assert.NotNil(suite.T(), event)
	assert.Equal(suite.T(), EventTypeAuthCodeExpired, event.Type)
	assert.Equal(suite.T(), userEmail, event.Data.UserEmail)
	assert.Equal(suite.T(), codeID, event.Data.CodeID)
	assert.WithinDuration(suite.T(), time.Now(), event.Data.ExpiredAt, time.Second)
}

func (suite *EventsTestSuite) TestNewAuthCodeFailedEvent() {
	userEmail := "test@example.com"
	codeID := "code-123"
	reason := "invalid_code"
	attemptsCount := 3

	event := NewAuthCodeFailedEvent(userEmail, codeID, reason, attemptsCount)

	assert.NotNil(suite.T(), event)
	assert.Equal(suite.T(), EventTypeAuthCodeFailed, event.Type)
	assert.Equal(suite.T(), userEmail, event.Data.UserEmail)
	assert.Equal(suite.T(), codeID, event.Data.CodeID)
	assert.Equal(suite.T(), reason, event.Data.Reason)
	assert.Equal(suite.T(), attemptsCount, event.Data.AttemptsCount)
	assert.WithinDuration(suite.T(), time.Now(), event.Data.FailedAt, time.Second)
}

func (suite *EventsTestSuite) TestAuthCodeRequestedEvent_EventInterface() {
	event := NewAuthCodeRequestedEvent("user-123", "test@example.com", "123456", "code-123", time.Now())

	assert.Equal(suite.T(), EventTypeAuthCodeRequested, event.GetType())
	assert.NotEmpty(suite.T(), event.GetID())
	assert.WithinDuration(suite.T(), time.Now(), event.GetTimestamp(), time.Second)
	assert.Equal(suite.T(), string(EventTypeAuthCodeRequested), event.GetRoutingKey())

	jsonData, err := event.ToJSON()
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), jsonData)

	// Verify JSON can be unmarshaled
	var unmarshaled AuthCodeRequestedEvent
	err = json.Unmarshal(jsonData, &unmarshaled)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), event.ID, unmarshaled.ID)
	assert.Equal(suite.T(), event.Type, unmarshaled.Type)
}

func (suite *EventsTestSuite) TestAuthCodeVerifiedEvent_EventInterface() {
	event := NewAuthCodeVerifiedEvent("test@example.com", "user-123", "code-123", false)

	assert.Equal(suite.T(), EventTypeAuthCodeVerified, event.GetType())
	assert.NotEmpty(suite.T(), event.GetID())
	assert.WithinDuration(suite.T(), time.Now(), event.GetTimestamp(), time.Second)
	assert.Equal(suite.T(), string(EventTypeAuthCodeVerified), event.GetRoutingKey())

	jsonData, err := event.ToJSON()
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), jsonData)
}

func (suite *EventsTestSuite) TestAuthCodeExpiredEvent_EventInterface() {
	event := NewAuthCodeExpiredEvent("test@example.com", "code-123")

	assert.Equal(suite.T(), EventTypeAuthCodeExpired, event.GetType())
	assert.Equal(suite.T(), string(EventTypeAuthCodeExpired), event.GetRoutingKey())

	jsonData, err := event.ToJSON()
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), jsonData)
}

func (suite *EventsTestSuite) TestAuthCodeFailedEvent_EventInterface() {
	event := NewAuthCodeFailedEvent("test@example.com", "code-123", "invalid_code", 3)

	assert.Equal(suite.T(), EventTypeAuthCodeFailed, event.GetType())
	assert.Equal(suite.T(), string(EventTypeAuthCodeFailed), event.GetRoutingKey())

	jsonData, err := event.ToJSON()
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), jsonData)
}

func TestEventTypes(t *testing.T) {
	testCases := []struct {
		name      string
		eventType EventType
		expected  string
	}{
		{"Auth Code Requested", EventTypeAuthCodeRequested, "auth.code.requested"},
		{"Auth Code Verified", EventTypeAuthCodeVerified, "auth.code.verified"},
		{"Auth Code Expired", EventTypeAuthCodeExpired, "auth.code.expired"},
		{"Auth Code Failed", EventTypeAuthCodeFailed, "auth.code.failed"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, string(tc.eventType))
		})
	}
}

func TestGenerateEventID(t *testing.T) {
	// Test that event IDs are unique
	id1 := generateEventID()
	time.Sleep(1 * time.Millisecond) // Ensure different timestamp
	id2 := generateEventID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
	assert.Len(t, id1, 21) // Format: YYYYMMDDHHMMSS.SSSSSS
}

func TestEventSerialization(t *testing.T) {
	testCases := []struct {
		name  string
		event Event
	}{
		{
			"AuthCodeRequestedEvent",
			NewAuthCodeRequestedEvent("user-123", "test@example.com", "123456", "code-123", time.Now()),
		},
		{
			"AuthCodeVerifiedEvent",
			NewAuthCodeVerifiedEvent("test@example.com", "user-123", "code-123", true),
		},
		{
			"AuthCodeExpiredEvent",
			NewAuthCodeExpiredEvent("test@example.com", "code-123"),
		},
		{
			"AuthCodeFailedEvent",
			NewAuthCodeFailedEvent("test@example.com", "code-123", "invalid_code", 3),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jsonData, err := tc.event.ToJSON()
			assert.NoError(t, err)
			assert.NotEmpty(t, jsonData)

			// Verify it's valid JSON
			var result map[string]interface{}
			err = json.Unmarshal(jsonData, &result)
			assert.NoError(t, err)

			// Check common fields
			assert.Equal(t, tc.event.GetID(), result["id"])
			assert.Equal(t, string(tc.event.GetType()), result["type"])
			assert.Equal(t, "authentication-service", result["source"])
			assert.Equal(t, "1.0", result["version"])
		})
	}
}
