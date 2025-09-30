package proxy

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// HeaderManagerTestSuite is a test suite for StandardHeaderManager
type HeaderManagerTestSuite struct {
	suite.Suite
	headerManager *StandardHeaderManager
}

// SetupTest runs before each test
func (suite *HeaderManagerTestSuite) SetupTest() {
	suite.headerManager = NewStandardHeaderManager()
}

// TestCopyRequestHeaders tests copying request headers
func (suite *HeaderManagerTestSuite) TestCopyRequestHeaders() {
	// Arrange
	srcReq := httptest.NewRequest("GET", "http://example.com", nil)
	srcReq.Header.Set("Content-Type", "application/json")
	srcReq.Header.Set("Authorization", "Bearer token")
	srcReq.Header.Set("Connection", "keep-alive") // Should be filtered
	srcReq.RemoteAddr = "192.168.1.1:1234"
	srcReq.Host = "example.com"

	dstReq := httptest.NewRequest("GET", "http://target.com", nil)

	// Act
	suite.headerManager.CopyRequestHeaders(dstReq, srcReq)

	// Assert - Regular headers should be copied
	assert.Equal(suite.T(), "application/json", dstReq.Header.Get("Content-Type"), "Content-Type should be copied")
	assert.Equal(suite.T(), "Bearer token", dstReq.Header.Get("Authorization"), "Authorization should be copied")

	// Assert - Hop-by-hop headers should be filtered
	assert.Empty(suite.T(), dstReq.Header.Get("Connection"), "Connection header should be filtered")

	// Assert - X-Forwarded headers should be added
	assert.NotEmpty(suite.T(), dstReq.Header.Get("X-Forwarded-For"), "X-Forwarded-For should be set")
	assert.Equal(suite.T(), "http", dstReq.Header.Get("X-Forwarded-Proto"), "X-Forwarded-Proto should be http")
	assert.Equal(suite.T(), "example.com", dstReq.Header.Get("X-Forwarded-Host"), "X-Forwarded-Host should be example.com")
}

// TestCopyRequestHeadersWithHTTPS tests copying request headers with HTTPS
func (suite *HeaderManagerTestSuite) TestCopyRequestHeadersWithHTTPS() {
	// Arrange
	srcReq := httptest.NewRequest("GET", "https://example.com", nil)
	srcReq.TLS = &tls.ConnectionState{} // Simulate HTTPS
	dstReq := httptest.NewRequest("GET", "http://target.com", nil)

	// Act
	suite.headerManager.CopyRequestHeaders(dstReq, srcReq)

	// Assert
	assert.Equal(suite.T(), "https", dstReq.Header.Get("X-Forwarded-Proto"), "X-Forwarded-Proto should be https")
}

// TestCopyResponseHeaders tests copying response headers
func (suite *HeaderManagerTestSuite) TestCopyResponseHeaders() {
	// Arrange
	resp := &http.Response{
		Header: make(http.Header),
	}
	resp.Header.Set("Content-Type", "application/json")
	resp.Header.Set("X-Custom-Header", "value")
	resp.Header.Set("Connection", "close") // Should be filtered

	w := httptest.NewRecorder()

	// Act
	suite.headerManager.CopyResponseHeaders(w, resp)

	// Assert - Regular headers should be copied
	assert.Equal(suite.T(), "application/json", w.Header().Get("Content-Type"), "Content-Type should be copied")
	assert.Equal(suite.T(), "value", w.Header().Get("X-Custom-Header"), "X-Custom-Header should be copied")

	// Assert - Hop-by-hop headers should be filtered
	assert.Empty(suite.T(), w.Header().Get("Connection"), "Connection header should be filtered")
}

// TestIsHopByHopHeader tests hop-by-hop header detection
func (suite *HeaderManagerTestSuite) TestIsHopByHopHeader() {
	tests := []struct {
		name     string
		header   string
		expected bool
	}{
		{"Connection is hop-by-hop", "Connection", true},
		{"Keep-Alive is hop-by-hop", "Keep-Alive", true},
		{"Content-Type is not hop-by-hop", "Content-Type", false},
		{"Authorization is not hop-by-hop", "Authorization", false},
		{"Upgrade is hop-by-hop", "Upgrade", true},
		{"X-Custom-Header is not hop-by-hop", "X-Custom-Header", false},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result := suite.headerManager.IsHopByHopHeader(tt.header)
			assert.Equal(suite.T(), tt.expected, result)
		})
	}
}

// TestHeaderManagerTestSuite runs the test suite
func TestHeaderManagerTestSuite(t *testing.T) {
	suite.Run(t, new(HeaderManagerTestSuite))
}