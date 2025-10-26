package messaging

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// @todo: remove this legacy health monitor once the new one is fully integrated
// @todo: remove this legacy health monitor once the new one is fully integrated
// @todo: remove this legacy health monitor once the new one is fully integrated
// @todo: remove this legacy health monitor once the new one is fully integrated
// @todo: remove this legacy health monitor once the new one is fully integrated
// @todo: remove this legacy health monitor once the new one is fully integrated
// @todo: remove this legacy health monitor once the new one is fully integrated
// @todo: remove this legacy health monitor once the new one is fully integrated
// @todo: remove this legacy health monitor once the new one is fully integrated
// @todo: remove this legacy health monitor once the new one is fully integrated
// @todo: remove this legacy health monitor once the new one is fully integrated
// @todo: remove this legacy health monitor once the new one is fully integrated
// @todo: remove this legacy health monitor once the new one is fully integrated
// @todo: remove this legacy health monitor once the new one is fully integrated
// @todo: remove this legacy health monitor once the new one is fully integrated

// HealthStatus represents the health status of the messaging system
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusDegraded  HealthStatus = "degraded"
)

// HealthCheck represents a health check result
type HealthCheck struct {
	Component string                 `json:"component"`
	Status    HealthStatus           `json:"status"`
	Message   string                 `json:"message"`
	Timestamp time.Time              `json:"timestamp"`
	Duration  time.Duration          `json:"duration_ms"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// HealthMonitor monitors the health of messaging components
type HealthMonitor interface {
	Start(ctx context.Context) error
	Stop() error
	GetStatus() HealthCheck
	GetDetailedStatus() map[string]HealthCheck
	IsHealthy() bool
	AddHealthChecker(name string, checker HealthChecker)
}

// HealthChecker interface for individual component health checks
type HealthChecker interface {
	CheckHealth(ctx context.Context) HealthCheck
}

type healthMonitor struct {
	connection    RabbitMQConnection
	publisher     Publisher
	checkers      map[string]HealthChecker
	status        map[string]HealthCheck
	checkInterval time.Duration
	mutex         sync.RWMutex
	stopCh        chan struct{}
	isRunning     bool
}

// NewHealthMonitor creates a new health monitor
func NewHealthMonitor(connection RabbitMQConnection, publisher Publisher) HealthMonitor {
	hm := &healthMonitor{
		connection:    connection,
		publisher:     publisher,
		checkers:      make(map[string]HealthChecker),
		status:        make(map[string]HealthCheck),
		checkInterval: 30 * time.Second,
		stopCh:        make(chan struct{}),
	}

	// Add default health checkers
	hm.AddHealthChecker("connection", &connectionHealthChecker{connection: connection})
	if publisher != nil {
		hm.AddHealthChecker("publisher", &publisherHealthChecker{publisher: publisher})
	}

	return hm
}

func (h *healthMonitor) Start(ctx context.Context) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if h.isRunning {
		return nil
	}

	h.isRunning = true
	go h.runHealthChecks(ctx)

	log.Println("Health monitor started")
	return nil
}

func (h *healthMonitor) Stop() error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if !h.isRunning {
		return nil
	}

	close(h.stopCh)
	h.isRunning = false

	log.Println("Health monitor stopped")
	return nil
}

func (h *healthMonitor) runHealthChecks(ctx context.Context) {
	ticker := time.NewTicker(h.checkInterval)
	defer ticker.Stop()

	// Run initial check
	h.performHealthChecks(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-h.stopCh:
			return
		case <-ticker.C:
			h.performHealthChecks(ctx)
		}
	}
}

func (h *healthMonitor) performHealthChecks(ctx context.Context) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	for name, checker := range h.checkers {
		checkCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		result := checker.CheckHealth(checkCtx)
		cancel()

		h.status[name] = result

		if result.Status != HealthStatusHealthy {
			log.Printf("Health check failed for %s: %s", name, result.Message)
		}
	}
}

func (h *healthMonitor) GetStatus() HealthCheck {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	overall := HealthCheck{
		Component: "messaging",
		Status:    HealthStatusHealthy,
		Message:   "All components healthy",
		Timestamp: time.Now().UTC(),
	}

	unhealthyCount := 0
	totalCount := len(h.status)

	for _, status := range h.status {
		if status.Status == HealthStatusUnhealthy {
			unhealthyCount++
		}
	}

	if unhealthyCount > 0 {
		if unhealthyCount == totalCount {
			overall.Status = HealthStatusUnhealthy
			overall.Message = "All components unhealthy"
		} else {
			overall.Status = HealthStatusDegraded
			overall.Message = fmt.Sprintf("%d of %d components unhealthy", unhealthyCount, totalCount)
		}
	}

	overall.Metadata = map[string]interface{}{
		"total_components":     totalCount,
		"unhealthy_components": unhealthyCount,
		"last_check":           time.Now().UTC().Format(time.RFC3339),
	}

	return overall
}

func (h *healthMonitor) GetDetailedStatus() map[string]HealthCheck {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	result := make(map[string]HealthCheck)
	for name, status := range h.status {
		result[name] = status
	}

	return result
}

func (h *healthMonitor) IsHealthy() bool {
	return h.GetStatus().Status == HealthStatusHealthy
}

func (h *healthMonitor) AddHealthChecker(name string, checker HealthChecker) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.checkers[name] = checker
}

// Connection health checker
type connectionHealthChecker struct {
	connection RabbitMQConnection
}

func (c *connectionHealthChecker) CheckHealth(ctx context.Context) HealthCheck {
	start := time.Now()

	check := HealthCheck{
		Component: "rabbitmq_connection",
		Timestamp: start,
	}

	if c.connection == nil {
		check.Status = HealthStatusUnhealthy
		check.Message = "Connection is nil"
		check.Duration = time.Since(start)
		return check
	}

	if !c.connection.IsConnected() {
		check.Status = HealthStatusUnhealthy
		check.Message = "Not connected to RabbitMQ"
		check.Duration = time.Since(start)
		return check
	}

	if !c.connection.IsHealthy() {
		check.Status = HealthStatusDegraded
		check.Message = "Connection exists but health check failed"
		check.Duration = time.Since(start)
		return check
	}

	check.Status = HealthStatusHealthy
	check.Message = "Connected and healthy"
	check.Duration = time.Since(start)
	return check
}

// Publisher health checker
type publisherHealthChecker struct {
	publisher Publisher
}

func (p *publisherHealthChecker) CheckHealth(ctx context.Context) HealthCheck {
	start := time.Now()

	check := HealthCheck{
		Component: "rabbitmq_publisher",
		Timestamp: start,
	}

	if p.publisher == nil {
		check.Status = HealthStatusUnhealthy
		check.Message = "Publisher is nil"
		check.Duration = time.Since(start)
		return check
	}

	if !p.publisher.IsHealthy() {
		check.Status = HealthStatusUnhealthy
		check.Message = "Publisher is not healthy"
		check.Duration = time.Since(start)
		return check
	}

	check.Status = HealthStatusHealthy
	check.Message = "Publisher is healthy"
	check.Duration = time.Since(start)
	return check
}
