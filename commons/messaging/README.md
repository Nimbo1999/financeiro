# RabbitMQ Messaging Package

A self-healing RabbitMQ messaging library with automatic reconnection, exponential backoff, and graceful error handling.

## Features

- ✅ **Automatic Reconnection**: Reconnects automatically on connection loss with exponential backoff
- ✅ **Publisher Confirms**: Optional publisher confirms for reliable message delivery
- ✅ **Consumer Auto-Recovery**: Consumers automatically resubscribe after reconnection
- ✅ **Thread-Safe**: All operations are safe for concurrent use
- ✅ **Configurable**: Flexible configuration with sensible defaults
- ✅ **Observable**: Structured logging with slog
- ✅ **Health Checks**: Built-in health check methods

## Installation

```bash
go get github.com/nimbo1999/financeiro/commons/messaging
```

## Quick Start

### Publisher Example

```go
package main

import (
    "context"
    "log/slog"
    "time"

    "github.com/nimbo1999/financeiro/commons/messaging"
)

func main() {
    // Create configuration
    config := &messaging.Config{
        URL:                 "amqp://guest:guest@localhost:5672/",
        MaxReconnectRetries: 10,
        BaseReconnectDelay:  1 * time.Second,
        MaxReconnectDelay:   30 * time.Second,
        PublisherConfirms:   true,
    }

    // Create publisher
    publisher, err := messaging.NewPublisher(config, slog.Default())
    if err != nil {
        panic(err)
    }
    defer publisher.Close()

    // Publish message
    ctx := context.Background()
    message := []byte(`{"event": "user.created", "user_id": 123}`)

    err = publisher.Publish(ctx, "events", "user.created", message)
    if err != nil {
        panic(err)
    }

    // Publish with confirmation
    err = publisher.PublishWithConfirm(ctx, "events", "user.created", message)
    if err != nil {
        panic(err)
    }
}
```

### Consumer Example

```go
package main

import (
    "context"
    "log/slog"
    "time"

    "github.com/nimbo1999/financeiro/commons/messaging"
    amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
    // Create configuration
    config := &messaging.Config{
        URL:                 "amqp://guest:guest@localhost:5672/",
        MaxReconnectRetries: 10,
        BaseReconnectDelay:  1 * time.Second,
        MaxReconnectDelay:   30 * time.Second,
        PrefetchCount:       10,
    }

    // Create consumer
    consumer, err := messaging.NewConsumer(config, slog.Default())
    if err != nil {
        panic(err)
    }
    defer consumer.Close()

    // Define message handler
    handler := func(delivery amqp.Delivery) error {
        slog.Info("received message",
            "routing_key", delivery.RoutingKey,
            "body", string(delivery.Body))

        // Process message...

        // Return nil to acknowledge, error to nack and requeue
        return nil
    }

    // Start consuming (blocks until context is cancelled)
    ctx := context.Background()
    err = consumer.Consume(ctx, "my-queue", handler)
    if err != nil {
        panic(err)
    }
}
```

## Configuration

### Config Struct

```go
type Config struct {
    // URL is the RabbitMQ connection string
    URL string

    // MaxReconnectRetries is the maximum number of reconnection attempts
    // Default: 10
    MaxReconnectRetries int

    // BaseReconnectDelay is the initial delay between reconnection attempts
    // Default: 1 second
    BaseReconnectDelay time.Duration

    // MaxReconnectDelay is the maximum delay between reconnection attempts
    // Default: 30 seconds
    MaxReconnectDelay time.Duration

    // PublisherConfirms enables publisher confirms for reliable delivery
    // Default: false
    PublisherConfirms bool

    // PrefetchCount sets the number of messages to prefetch for consumers
    // Default: 1
    PrefetchCount int
}
```

### Default Configuration

```go
config := messaging.DefaultConfig("amqp://localhost:5672")
// Returns config with all defaults applied
```

## Publisher Interface

```go
type Publisher interface {
    // Publish publishes a message
    Publish(ctx context.Context, exchange, routingKey string, body []byte) error

    // PublishWithConfirm publishes and waits for broker confirmation
    PublishWithConfirm(ctx context.Context, exchange, routingKey string, body []byte) error

    // Close closes the publisher
    Close() error

    // IsHealthy checks if the publisher is ready
    IsHealthy() bool
}
```

## Consumer Interface

```go
type Consumer interface {
    // Consume starts consuming messages
    Consume(ctx context.Context, queue string, handler MessageHandler) error

    // Close closes the consumer
    Close() error

    // IsHealthy checks if the consumer is ready
    IsHealthy() bool
}

type MessageHandler func(delivery amqp.Delivery) error
```

## Reconnection Behavior

The library implements exponential backoff for reconnection attempts:

1. **First attempt**: 1 second delay
2. **Second attempt**: 2 seconds delay
3. **Third attempt**: 4 seconds delay
4. **Fourth attempt**: 8 seconds delay
5. **Fifth attempt**: 16 seconds delay
6. **Subsequent attempts**: 30 seconds delay (capped)

After `MaxReconnectRetries` failed attempts, the library stops trying and logs an error.

## Connection Monitoring

The library automatically monitors:

- Connection close events
- Channel close events
- Connection health status

When a connection loss is detected, automatic reconnection is triggered.

## Health Checks

Both Publisher and Consumer expose health check methods:

```go
if publisher.IsHealthy() {
    // Safe to publish
}

if consumer.IsHealthy() {
    // Consumer is connected
}
```

Integrate these with your service's health check endpoint:

```go
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
    health := map[string]string{
        "rabbitmq": "unhealthy",
    }

    if publisher.IsHealthy() {
        health["rabbitmq"] = "healthy"
    }

    json.NewEncoder(w).Encode(health)
}
```

## Error Handling

### Publisher Errors

- Connection errors trigger automatic reconnection
- Publish failures return errors immediately
- PublishWithConfirm waits for broker confirmation or timeout

### Consumer Errors

- Handler errors cause messages to be nacked and requeued
- Connection errors trigger automatic reconnection and resubscription
- Panics in handlers are caught and logged

## Best Practices

1. **Use contexts with timeouts** for publish operations:
   ```go
   ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
   defer cancel()
   err := publisher.Publish(ctx, exchange, key, body)
   ```

2. **Enable publisher confirms** for critical messages:
   ```go
   config.PublisherConfirms = true
   ```

3. **Set appropriate prefetch counts** for consumers:
   ```go
   config.PrefetchCount = 10 // Process 10 messages concurrently
   ```

4. **Implement idempotent message handlers** to handle redelivery:
   ```go
   handler := func(delivery amqp.Delivery) error {
       // Check if message was already processed
       if alreadyProcessed(delivery.MessageId) {
           return nil
       }
       // Process message...
   }
   ```

5. **Use graceful shutdown**:
   ```go
   sigChan := make(chan os.Signal, 1)
   signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

   <-sigChan
   publisher.Close()
   consumer.Close()
   ```

## Testing

The package includes unit tests for configuration and initialization. For testing services that use this library:

### Mock Publisher

```go
type MockPublisher struct {
    mock.Mock
}

func (m *MockPublisher) Publish(ctx context.Context, exchange, key string, body []byte) error {
    args := m.Called(ctx, exchange, key, body)
    return args.Error(0)
}

func (m *MockPublisher) IsHealthy() bool {
    args := m.Called()
    return args.Bool(0)
}
```

### Integration Tests

For full integration tests, use testcontainers:

```go
import "github.com/testcontainers/testcontainers-go"

// Start RabbitMQ container
// Create publisher/consumer
// Test end-to-end message flow
```

## Logging

The library uses structured logging with `log/slog`. Log levels:

- **INFO**: Connection events, successful operations
- **WARN**: Reconnection attempts, delivery channel closures
- **ERROR**: Connection failures, publish errors, handler errors
- **DEBUG**: Individual message details

Configure log level in your service:

```go
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelInfo,
}))

publisher, _ := messaging.NewPublisher(config, logger)
```

## Performance Considerations

- **Connection pooling**: Each Publisher/Consumer maintains one connection. Create multiple instances if needed.
- **Channel pooling**: Channels are managed automatically and recreated on failures.
- **Message batching**: For high-throughput, consider batching messages before publishing.
- **Prefetch count**: Tune based on message processing time and concurrency needs.

## Troubleshooting

### Connection keeps reconnecting

- Check RabbitMQ server logs
- Verify network connectivity
- Check authentication credentials
- Ensure queue/exchange exists

### Messages not being consumed

- Verify queue name is correct
- Check if consumer is healthy: `consumer.IsHealthy()`
- Ensure handler doesn't panic
- Check RabbitMQ management UI for queue bindings

### Slow message processing

- Increase `PrefetchCount` for parallel processing
- Optimize message handler logic
- Consider using worker pools

## Migration from Direct AMQP Usage

Before:
```go
conn, _ := amqp.Dial(url)
ch, _ := conn.Channel()
ch.Publish(exchange, key, false, false, amqp.Publishing{Body: body})
```

After:
```go
publisher, _ := messaging.NewPublisher(config, logger)
publisher.Publish(ctx, exchange, key, body)
```

Benefits:
- Automatic reconnection
- Health checks
- Better error handling
- Structured logging
- Thread-safe operations

## License

This package is part of the Financeiro project.
