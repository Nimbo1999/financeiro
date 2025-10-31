package messaging

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	// MaxRetries is the maximum number of retry attempts before sending to DLQ
	MaxRetries = 5
)

// GetRetryCount extracts the retry count from message headers
// RabbitMQ tracks retries in the x-death header array
func GetRetryCount(delivery amqp.Delivery) int64 {
	if delivery.Headers == nil {
		return 0
	}

	// x-death is an array of death records
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
