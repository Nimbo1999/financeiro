package proxy

import (
	"net/http"
	"time"

	"github.com/nimbo1999/financeiro/gateway/internal/config"
)

// Factory creates and wires proxy components following SOLID principles
type Factory struct {
	config *config.Config
}

// NewFactory creates a new Factory
func NewFactory(cfg *config.Config) *Factory {
	return &Factory{config: cfg}
}

// CreateLoggingProxy creates a fully configured LoggingProxy
func (f *Factory) CreateLoggingProxy() *LoggingProxy {
	timeout := 30 * time.Second
	// Create dependencies
	headerManager := NewStandardHeaderManager()
	requestCreator := NewStandardRequestCreator(headerManager, timeout)

	httpClient := &http.Client{
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// Create base proxy
	serviceProxy := NewServiceProxy(httpClient, requestCreator, headerManager)

	// Create logger
	logger := NewStandardLogger()

	// Wrap proxy with logging
	return NewLoggingProxy(serviceProxy, logger)
}

// CreateHealthAggregator creates a fully configured HealthAggregator
func (f *Factory) CreateHealthAggregator() *HealthAggregator {
	// Create health checker
	healthChecker := NewHTTPHealthChecker(5 * time.Second)

	// Create service registry with configured services
	services := []ServiceConfig{
		{
			Name: "authentication",
			URL:  f.config.Services.AuthServiceURL,
		},
		{
			Name: "users",
			URL:  f.config.Services.UserServiceURL,
		},
		{
			Name: "notifications",
			URL:  f.config.Services.NotificationServiceURL,
		},
	}
	serviceRegistry := NewInMemoryServiceRegistry(services)

	// Create health aggregator
	return NewHealthAggregator(healthChecker, serviceRegistry)
}
