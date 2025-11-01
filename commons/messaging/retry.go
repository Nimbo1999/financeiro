package messaging

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	// MaxRetries is the maximum number of retry attempts before sending to DLQ
	MaxRetries = 5

	// RetryCountHeader is the custom header used to track retry attempts
	RetryCountHeader = "x-retry-count"
)

// GetRetryCount extracts the retry count from message headers
// We use a custom header since x-death is only populated when messages are dead-lettered
func GetRetryCount(delivery amqp.Delivery) int64 {
	if delivery.Headers == nil {
		return 0
	}

	// Try to get custom retry count header first
	if count, ok := delivery.Headers[RetryCountHeader].(int32); ok {
		return int64(count)
	}

	// Fall back to x-death if present (for messages that came from DLQ)
	xDeath, ok := delivery.Headers["x-death"].([]interface{})
	if !ok || len(xDeath) == 0 {
		return 0
	}

	// Each death record is a table, get the first one
	deathRecord, ok := xDeath[0].(amqp.Table)
	if !ok {
		return 0
	}

	// Extract count from the death record
	count, ok := deathRecord["count"].(int64)
	if !ok {
		return 0
	}

	return count
}

// ShouldRetry determines if a message should be retried
func ShouldRetry(delivery amqp.Delivery) bool {
	retryCount := GetRetryCount(delivery)
	return retryCount < MaxRetries
}

// ShouldMoveToDLQ determines if a message should be moved to DLQ
func ShouldMoveToDLQ(delivery amqp.Delivery) bool {
	return !ShouldRetry(delivery)
}
