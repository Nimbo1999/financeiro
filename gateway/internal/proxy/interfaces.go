package proxy

import "net/http"

// Proxier defines the interface for proxying HTTP requests
type Proxier interface {
	ProxyRequest(w http.ResponseWriter, r *http.Request, targetURL string) error
}

// HeaderManager defines the interface for managing HTTP headers
type HeaderManager interface {
	CopyRequestHeaders(dst, src *http.Request)
	CopyResponseHeaders(w http.ResponseWriter, resp *http.Response)
	IsHopByHopHeader(header string) bool
}

// RequestCreator defines the interface for creating proxy requests
type RequestCreator interface {
	CreateProxyRequest(r *http.Request, targetURL string) (*http.Request, error)
}

// Logger defines the interface for logging proxy operations
type Logger interface {
	LogRequest(method, path, targetURL, serviceName string)
	LogResponse(method, path, targetURL string, statusCode int, duration string)
	LogError(method, path, targetURL string, duration string, err error)
}

// HealthChecker defines the interface for checking service health
type HealthChecker interface {
	CheckHealth(name, url string) ServiceHealth
}

// ServiceRegistry defines the interface for managing service configurations
type ServiceRegistry interface {
	GetServices() []ServiceConfig
}