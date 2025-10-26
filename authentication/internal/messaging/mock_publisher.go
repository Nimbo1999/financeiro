package messaging

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockPublisher is a mock implementation of the Publisher interface for testing
type MockPublisher struct {
	mock.Mock
}

func (m *MockPublisher) PublishEvent(ctx context.Context, event Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockPublisher) PublishWithRetry(ctx context.Context, event Event, maxRetries int) error {
	args := m.Called(ctx, event, maxRetries)
	return args.Error(0)
}

func (m *MockPublisher) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockPublisher) IsHealthy() bool {
	args := m.Called()
	return args.Bool(0)
}
