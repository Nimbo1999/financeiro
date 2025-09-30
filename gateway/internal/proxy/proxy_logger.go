package proxy

import (
	"log"
	"net/http"
)

// StandardLogger implements the Logger interface
type StandardLogger struct{}

// NewStandardLogger creates a new StandardLogger
func NewStandardLogger() *StandardLogger {
	return &StandardLogger{}
}

// LogRequest logs an incoming proxy request
func (l *StandardLogger) LogRequest(method, path, targetURL, serviceName string) {
	log.Printf("[PROXY] Incoming: %s %s -> %s (service: %s)", method, path, targetURL, serviceName)
}

// LogResponse logs a successful proxy response
func (l *StandardLogger) LogResponse(method, path, targetURL string, statusCode int, duration string) {
	log.Printf("[PROXY] Response: %s %s -> %s completed with status %d in %s",
		method, path, targetURL, statusCode, duration)
}

// LogError logs a proxy error
func (l *StandardLogger) LogError(method, path, targetURL string, duration string, err error) {
	log.Printf("[PROXY] Error: %s %s -> %s failed after %s: %v",
		method, path, targetURL, duration, err)
}

// responseWriter wraps http.ResponseWriter to capture the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// NewResponseWriter creates a new responseWriter
func NewResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK, // Default to 200 OK
	}
}

// WriteHeader captures the status code and calls the underlying WriteHeader
func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

// StatusCode returns the captured status code
func (rw *responseWriter) StatusCode() int {
	return rw.statusCode
}
