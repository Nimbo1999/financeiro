package proxy

import "net/http"

// StandardHeaderManager implements HeaderManager interface
type StandardHeaderManager struct {
	hopByHopHeaders map[string]bool
}

// NewStandardHeaderManager creates a new StandardHeaderManager
func NewStandardHeaderManager() *StandardHeaderManager {
	return &StandardHeaderManager{
		hopByHopHeaders: map[string]bool{
			"Connection":          true,
			"Keep-Alive":          true,
			"Proxy-Authenticate":  true,
			"Proxy-Authorization": true,
			"Te":                  true,
			"Trailers":            true,
			"Transfer-Encoding":   true,
			"Upgrade":             true,
		},
	}
}

// CopyRequestHeaders copies headers from source to destination request
func (hm *StandardHeaderManager) CopyRequestHeaders(dst, src *http.Request) {
	// Copy all headers except hop-by-hop headers
	for key, values := range src.Header {
		if hm.IsHopByHopHeader(key) {
			continue
		}
		for _, value := range values {
			dst.Header.Add(key, value)
		}
	}

	// Set X-Forwarded headers
	if clientIP := src.RemoteAddr; clientIP != "" {
		dst.Header.Set("X-Forwarded-For", clientIP)
	}
	if src.TLS != nil {
		dst.Header.Set("X-Forwarded-Proto", "https")
	} else {
		dst.Header.Set("X-Forwarded-Proto", "http")
	}
	dst.Header.Set("X-Forwarded-Host", src.Host)
}

// CopyResponseHeaders copies headers from response to response writer
func (hm *StandardHeaderManager) CopyResponseHeaders(w http.ResponseWriter, resp *http.Response) {
	for key, values := range resp.Header {
		if hm.IsHopByHopHeader(key) {
			continue
		}
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
}

// IsHopByHopHeader checks if a header should not be forwarded
func (hm *StandardHeaderManager) IsHopByHopHeader(header string) bool {
	return hm.hopByHopHeaders[header]
}