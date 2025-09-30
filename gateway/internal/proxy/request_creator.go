package proxy

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// StandardRequestCreator implements RequestCreator interface
type StandardRequestCreator struct {
	headerManager HeaderManager
	timeout       time.Duration
}

// NewStandardRequestCreator creates a new StandardRequestCreator
func NewStandardRequestCreator(headerManager HeaderManager, timeout time.Duration) *StandardRequestCreator {
	return &StandardRequestCreator{
		headerManager: headerManager,
		timeout:       timeout,
	}
}

// CreateProxyRequest creates a new HTTP request for proxying
func (rc *StandardRequestCreator) CreateProxyRequest(r *http.Request, targetURL string) (*http.Request, error) {
	// Parse target URL
	target, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}

	// Create new context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), rc.timeout)
	_ = cancel // Will be called when the original request context is done

	// Build target URL with original path and query
	fullTargetURL := *target
	fullTargetURL.Path = r.URL.Path
	fullTargetURL.RawQuery = r.URL.RawQuery

	// Create new request
	proxyReq, err := http.NewRequestWithContext(ctx, r.Method, fullTargetURL.String(), r.Body)
	if err != nil {
		return nil, err
	}

	// Copy headers from original request
	rc.headerManager.CopyRequestHeaders(proxyReq, r)

	return proxyReq, nil
}