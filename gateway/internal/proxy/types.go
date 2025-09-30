package proxy

import "time"

// ServiceHealth represents the health status of a service
type ServiceHealth struct {
	Name      string    `json:"name"`
	URL       string    `json:"url"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Error     string    `json:"error,omitempty"`
}

// ServiceConfig represents a service configuration
type ServiceConfig struct {
	Name string
	URL  string
}