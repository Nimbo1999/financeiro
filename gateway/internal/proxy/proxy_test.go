package proxy

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// ServiceProxyTestSuite is a test suite for ServiceProxy
type ServiceProxyTestSuite struct {
	suite.Suite
	mockRequestCreator *MockRequestCreator
	mockHeaderManager  *MockHeaderManager
	httpClient         *http.Client
	proxy              *ServiceProxy
}

// SetupTest runs before each test
func (suite *ServiceProxyTestSuite) SetupTest() {
	suite.mockRequestCreator = new(MockRequestCreator)
	suite.mockHeaderManager = new(MockHeaderManager)
	suite.httpClient = &http.Client{Timeout: 5 * time.Second}
	suite.proxy = NewServiceProxy(suite.httpClient, suite.mockRequestCreator, suite.mockHeaderManager)
}

// TearDownTest runs after each test
func (suite *ServiceProxyTestSuite) TearDownTest() {
	suite.mockRequestCreator.AssertExpectations(suite.T())
	suite.mockHeaderManager.AssertExpectations(suite.T())
}

// TestProxyRequest_Success tests successful proxy request
func (suite *ServiceProxyTestSuite) TestProxyRequest_Success() {
	// Arrange - Create test server
	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Test-Header", "test-value")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("response from target"))
	}))
	defer targetServer.Close()

	req := httptest.NewRequest("GET", "http://gateway/test", nil)
	w := httptest.NewRecorder()

	// Create proper client request without RequestURI
	proxyReq, _ := http.NewRequest("GET", targetServer.URL, nil)
	suite.mockRequestCreator.On("CreateProxyRequest", req, targetServer.URL).Return(proxyReq, nil)
	suite.mockHeaderManager.On("CopyResponseHeaders", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		respWriter := args.Get(0).(http.ResponseWriter)
		resp := args.Get(1).(*http.Response)
		for k, v := range resp.Header {
			for _, val := range v {
				respWriter.Header().Add(k, val)
			}
		}
	})

	// Act
	err := suite.proxy.ProxyRequest(w, req, targetServer.URL)

	// Assert
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	assert.Equal(suite.T(), "response from target", w.Body.String())
	assert.Equal(suite.T(), "test-value", w.Header().Get("X-Test-Header"))
}

// TestProxyRequest_CreateRequestError tests error during request creation
func (suite *ServiceProxyTestSuite) TestProxyRequest_CreateRequestError() {
	// Arrange
	req := httptest.NewRequest("GET", "http://gateway/test", nil)
	w := httptest.NewRecorder()
	expectedErr := errors.New("failed to create request")

	suite.mockRequestCreator.On("CreateProxyRequest", req, "http://target").Return(nil, expectedErr)

	// Act
	err := suite.proxy.ProxyRequest(w, req, "http://target")

	// Assert
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), expectedErr, err)
}

// TestProxyRequest_WithBody tests proxying request with body
func (suite *ServiceProxyTestSuite) TestProxyRequest_WithBody() {
	// Arrange - Create test server that echoes body
	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.WriteHeader(http.StatusCreated)
		w.Write(body)
	}))
	defer targetServer.Close()

	bodyContent := "test request body"
	req := httptest.NewRequest("POST", "http://gateway/test", strings.NewReader(bodyContent))
	w := httptest.NewRecorder()

	proxyReq, _ := http.NewRequest("POST", targetServer.URL, strings.NewReader(bodyContent))
	suite.mockRequestCreator.On("CreateProxyRequest", mock.Anything, targetServer.URL).Return(proxyReq, nil)
	suite.mockHeaderManager.On("CopyResponseHeaders", mock.Anything, mock.Anything)

	// Act
	err := suite.proxy.ProxyRequest(w, req, targetServer.URL)

	// Assert
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusCreated, w.Code)
	assert.Equal(suite.T(), bodyContent, w.Body.String())
}

// TestProxyRequest_NetworkError tests handling of network errors
func (suite *ServiceProxyTestSuite) TestProxyRequest_NetworkError() {
	// Arrange - Use very short timeout to force error
	suite.proxy.httpClient = &http.Client{Timeout: 1 * time.Nanosecond}

	req := httptest.NewRequest("GET", "http://gateway/test", nil)
	w := httptest.NewRecorder()

	proxyReq, _ := http.NewRequest("GET", "http://localhost:99999", nil)
	suite.mockRequestCreator.On("CreateProxyRequest", req, "http://localhost:99999").Return(proxyReq, nil)

	// Act
	err := suite.proxy.ProxyRequest(w, req, "http://localhost:99999")

	// Assert
	assert.Error(suite.T(), err)
}

// TestNewServiceProxy_WithNilClient tests creation with nil HTTP client
func (suite *ServiceProxyTestSuite) TestNewServiceProxy_WithNilClient() {
	// Act
	proxy := NewServiceProxy(nil, suite.mockRequestCreator, suite.mockHeaderManager)

	// Assert
	assert.NotNil(suite.T(), proxy)
	assert.NotNil(suite.T(), proxy.httpClient)
}

// TestServiceProxyTestSuite runs the test suite
func TestServiceProxyTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceProxyTestSuite))
}