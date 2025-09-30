package proxy

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// LoggingProxyTestSuite is a test suite for LoggingProxy
type LoggingProxyTestSuite struct {
	suite.Suite
	mockProxier *MockProxier
	mockLogger  *MockLogger
	loggingProxy *LoggingProxy
}

// SetupTest runs before each test
func (suite *LoggingProxyTestSuite) SetupTest() {
	suite.mockProxier = new(MockProxier)
	suite.mockLogger = new(MockLogger)
	suite.loggingProxy = NewLoggingProxy(suite.mockProxier, suite.mockLogger)
}

// TearDownTest runs after each test
func (suite *LoggingProxyTestSuite) TearDownTest() {
	suite.mockProxier.AssertExpectations(suite.T())
	suite.mockLogger.AssertExpectations(suite.T())
}

// TestProxyRequest_Success tests successful proxy request with logging
func (suite *LoggingProxyTestSuite) TestProxyRequest_Success() {
	// Arrange
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	suite.mockLogger.On("LogRequest", "GET", "/test", "http://target", "").Return()
	suite.mockProxier.On("ProxyRequest", mock.Anything, req, "http://target").Run(func(args mock.Arguments) {
		rw := args.Get(0).(http.ResponseWriter)
		rw.WriteHeader(http.StatusOK)
	}).Return(nil)
	suite.mockLogger.On("LogResponse", "GET", "/test", "http://target", http.StatusOK, mock.AnythingOfType("string")).Return()

	// Act
	err := suite.loggingProxy.ProxyRequest(w, req, "http://target")

	// Assert
	assert.NoError(suite.T(), err)
}

// TestProxyRequest_Error tests proxy request with error
func (suite *LoggingProxyTestSuite) TestProxyRequest_Error() {
	// Arrange
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	expectedErr := errors.New("proxy failed")

	suite.mockLogger.On("LogRequest", "GET", "/test", "http://target", "").Return()
	suite.mockProxier.On("ProxyRequest", mock.Anything, req, "http://target").Return(expectedErr)
	suite.mockLogger.On("LogError", "GET", "/test", "http://target", mock.AnythingOfType("string"), expectedErr).Return()

	// Act
	err := suite.loggingProxy.ProxyRequest(w, req, "http://target")

	// Assert
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), expectedErr, err)
}

// TestProxyRequestWithLogging_Success tests proxy with service name logging
func (suite *LoggingProxyTestSuite) TestProxyRequestWithLogging_Success() {
	// Arrange
	req := httptest.NewRequest("POST", "/api/users", nil)
	w := httptest.NewRecorder()

	suite.mockLogger.On("LogRequest", "POST", "/api/users", "http://user-service", "users").Return()
	suite.mockProxier.On("ProxyRequest", mock.Anything, req, "http://user-service").Run(func(args mock.Arguments) {
		rw := args.Get(0).(http.ResponseWriter)
		rw.WriteHeader(http.StatusCreated)
	}).Return(nil)
	suite.mockLogger.On("LogResponse", "POST", "/api/users", "http://user-service", http.StatusCreated, mock.AnythingOfType("string")).Return()

	// Act
	err := suite.loggingProxy.ProxyRequestWithLogging(w, req, "http://user-service", "users")

	// Assert
	assert.NoError(suite.T(), err)
}

// TestProxyRequestWithLogging_Error tests proxy with error and service name
func (suite *LoggingProxyTestSuite) TestProxyRequestWithLogging_Error() {
	// Arrange
	req := httptest.NewRequest("GET", "/api/users", nil)
	w := httptest.NewRecorder()
	expectedErr := errors.New("service unavailable")

	suite.mockLogger.On("LogRequest", "GET", "/api/users", "http://user-service", "users").Return()
	suite.mockProxier.On("ProxyRequest", mock.Anything, req, "http://user-service").Return(expectedErr)
	suite.mockLogger.On("LogError", "GET", "/api/users", "http://user-service", mock.AnythingOfType("string"), expectedErr).Return()

	// Act
	err := suite.loggingProxy.ProxyRequestWithLogging(w, req, "http://user-service", "users")

	// Assert
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), expectedErr, err)
}

// TestLoggingProxyTestSuite runs the test suite
func TestLoggingProxyTestSuite(t *testing.T) {
	suite.Run(t, new(LoggingProxyTestSuite))
}

// ResponseWriterTestSuite is a test suite for responseWriter
type ResponseWriterTestSuite struct {
	suite.Suite
}

// TestResponseWriter_StatusCode tests status code capture
func (suite *ResponseWriterTestSuite) TestResponseWriter_StatusCode() {
	// Arrange
	w := httptest.NewRecorder()
	rw := NewResponseWriter(w)

	// Act
	rw.WriteHeader(http.StatusCreated)

	// Assert
	assert.Equal(suite.T(), http.StatusCreated, rw.StatusCode())
	assert.Equal(suite.T(), http.StatusCreated, w.Code)
}

// TestResponseWriter_DefaultStatusCode tests default status code
func (suite *ResponseWriterTestSuite) TestResponseWriter_DefaultStatusCode() {
	// Arrange
	w := httptest.NewRecorder()
	rw := NewResponseWriter(w)

	// Assert - Default should be 200
	assert.Equal(suite.T(), http.StatusOK, rw.StatusCode())
}

// TestResponseWriter_Write tests writing response
func (suite *ResponseWriterTestSuite) TestResponseWriter_Write() {
	// Arrange
	w := httptest.NewRecorder()
	rw := NewResponseWriter(w)

	// Act
	rw.WriteHeader(http.StatusAccepted)
	rw.Write([]byte("test body"))

	// Assert
	assert.Equal(suite.T(), http.StatusAccepted, rw.StatusCode())
	assert.Equal(suite.T(), "test body", w.Body.String())
}

// TestResponseWriterTestSuite runs the test suite
func TestResponseWriterTestSuite(t *testing.T) {
	suite.Run(t, new(ResponseWriterTestSuite))
}
