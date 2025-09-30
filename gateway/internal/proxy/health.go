package proxy

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// HealthAggregator manages health checks for all services using dependency injection
type HealthAggregator struct {
	healthChecker   HealthChecker
	serviceRegistry ServiceRegistry
}

// NewHealthAggregator creates a new HealthAggregator with injected dependencies
func NewHealthAggregator(healthChecker HealthChecker, serviceRegistry ServiceRegistry) *HealthAggregator {
	return &HealthAggregator{
		healthChecker:   healthChecker,
		serviceRegistry: serviceRegistry,
	}
}

// CheckAllServices checks the health of all configured services concurrently
func (ha *HealthAggregator) CheckAllServices() []ServiceHealth {
	services := ha.serviceRegistry.GetServices()
	results := make([]ServiceHealth, len(services))
	var wg sync.WaitGroup

	for i, svc := range services {
		wg.Add(1)
		go func(index int, service ServiceConfig) {
			defer wg.Done()
			results[index] = ha.healthChecker.CheckHealth(service.Name, service.URL)
		}(i, svc)
	}

	wg.Wait()
	return results
}

// AggregatedHealthHandler returns an HTTP handler for aggregated health checks
func (ha *HealthAggregator) AggregatedHealthHandler(w http.ResponseWriter, r *http.Request) {
	results := ha.CheckAllServices()

	// Determine overall status
	overallStatus := ha.determineOverallStatus(results)

	response := map[string]interface{}{
		"status":    overallStatus,
		"services":  results,
		"timestamp": time.Now(),
	}

	// Set appropriate HTTP status code
	if overallStatus == "healthy" {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	json.NewEncoder(w).Encode(response)
}

// determineOverallStatus determines the overall health status from individual services
func (ha *HealthAggregator) determineOverallStatus(results []ServiceHealth) string {
	for _, result := range results {
		if result.Status != "healthy" {
			return "degraded"
		}
	}
	return "healthy"
}
