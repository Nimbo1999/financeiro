package proxy

import (
	"net/http"
	"time"
)

// LoggingProxy wraps a Proxier with logging capabilities (Decorator pattern)
type LoggingProxy struct {
	proxier Proxier
	logger  Logger
}

// NewLoggingProxy creates a new LoggingProxy with injected dependencies
func NewLoggingProxy(proxier Proxier, logger Logger) *LoggingProxy {
	return &LoggingProxy{
		proxier: proxier,
		logger:  logger,
	}
}

// ProxyRequest forwards a request with detailed logging
func (lp *LoggingProxy) ProxyRequest(w http.ResponseWriter, r *http.Request, targetURL string) error {
	start := time.Now()

	// Log incoming request
	lp.logger.LogRequest(r.Method, r.URL.Path, targetURL, "")

	// Wrap the response writer to capture status code
	wrappedWriter := NewResponseWriter(w)

	// Proxy the request using the underlying proxier
	err := lp.proxier.ProxyRequest(wrappedWriter, r, targetURL)

	// Calculate duration
	duration := time.Since(start).String()

	// Log response or error
	if err != nil {
		lp.logger.LogError(r.Method, r.URL.Path, targetURL, duration, err)
		return err
	}

	lp.logger.LogResponse(r.Method, r.URL.Path, targetURL, wrappedWriter.StatusCode(), duration)
	return nil
}

// ProxyRequestWithLogging forwards a request with service-specific logging
func (lp *LoggingProxy) ProxyRequestWithLogging(w http.ResponseWriter, r *http.Request, targetURL, serviceName string) error {
	start := time.Now()

	// Log incoming request with service name
	lp.logger.LogRequest(r.Method, r.URL.Path, targetURL, serviceName)

	// Wrap the response writer to capture status code
	wrappedWriter := NewResponseWriter(w)

	// Proxy the request
	err := lp.proxier.ProxyRequest(wrappedWriter, r, targetURL)

	// Calculate duration
	duration := time.Since(start).String()

	// Log response or error
	if err != nil {
		lp.logger.LogError(r.Method, r.URL.Path, targetURL, duration, err)
		return err
	}

	lp.logger.LogResponse(r.Method, r.URL.Path, targetURL, wrappedWriter.StatusCode(), duration)
	return nil
}