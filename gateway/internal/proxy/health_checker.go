package proxy

import (
	"fmt"
	"net/http"
	"time"
)

// HTTPHealthChecker implements HealthChecker interface
type HTTPHealthChecker struct {
	httpClient *http.Client
}

// NewHTTPHealthChecker creates a new HTTPHealthChecker
func NewHTTPHealthChecker(timeout time.Duration) *HTTPHealthChecker {
	return &HTTPHealthChecker{
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// CheckHealth performs a health check on a single service
func (hc *HTTPHealthChecker) CheckHealth(name, baseURL string) ServiceHealth {
	healthURL := fmt.Sprintf("%s/health", baseURL)

	health := ServiceHealth{
		Name:      name,
		URL:       healthURL,
		Timestamp: time.Now(),
	}

	resp, err := hc.httpClient.Get(healthURL)
	if err != nil {
		health.Status = "unhealthy"
		health.Error = err.Error()
		return health
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		health.Status = "healthy"
	} else {
		health.Status = "unhealthy"
		health.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
	}
	return health
}
