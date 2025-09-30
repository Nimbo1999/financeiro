package proxy

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// HealthAggregatorTestSuite is a test suite for HealthAggregator
type HealthAggregatorTestSuite struct {
	suite.Suite
	mockHealthChecker   *MockHealthChecker
	mockServiceRegistry *MockServiceRegistry
	aggregator          *HealthAggregator
}

// SetupTest runs before each test
func (suite *HealthAggregatorTestSuite) SetupTest() {
	suite.mockHealthChecker = new(MockHealthChecker)
	suite.mockServiceRegistry = new(MockServiceRegistry)
	suite.aggregator = NewHealthAggregator(suite.mockHealthChecker, suite.mockServiceRegistry)
}

// TearDownTest runs after each test
func (suite *HealthAggregatorTestSuite) TearDownTest() {
	suite.mockHealthChecker.AssertExpectations(suite.T())
	suite.mockServiceRegistry.AssertExpectations(suite.T())
}

// TestCheckAllServices_AllHealthy tests checking all healthy services
func (suite *HealthAggregatorTestSuite) TestCheckAllServices_AllHealthy() {
	// Arrange
	services := []ServiceConfig{
		{Name: "service1", URL: "http://service1"},
		{Name: "service2", URL: "http://service2"},
	}
	suite.mockServiceRegistry.On("GetServices").Return(services)

	suite.mockHealthChecker.On("CheckHealth", "service1", "http://service1").
		Return(ServiceHealth{Name: "service1", URL: "http://service1", Status: "healthy", Timestamp: time.Now()})
	suite.mockHealthChecker.On("CheckHealth", "service2", "http://service2").
		Return(ServiceHealth{Name: "service2", URL: "http://service2", Status: "healthy", Timestamp: time.Now()})

	// Act
	results := suite.aggregator.CheckAllServices()

	// Assert
	assert.Len(suite.T(), results, 2)
	assert.Equal(suite.T(), "healthy", results[0].Status)
	assert.Equal(suite.T(), "healthy", results[1].Status)
}

// TestAggregatedHealthHandler_Healthy tests the HTTP handler with healthy services
func (suite *HealthAggregatorTestSuite) TestAggregatedHealthHandler_Healthy() {
	// Arrange
	services := []ServiceConfig{{Name: "service1", URL: "http://service1"}}
	suite.mockServiceRegistry.On("GetServices").Return(services)
	suite.mockHealthChecker.On("CheckHealth", "service1", "http://service1").
		Return(ServiceHealth{Name: "service1", URL: "http://service1", Status: "healthy", Timestamp: time.Now()})

	req := httptest.NewRequest("GET", "/health/services", nil)
	w := httptest.NewRecorder()

	// Act
	suite.aggregator.AggregatedHealthHandler(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "healthy", response["status"])
}

// TestHealthAggregatorTestSuite runs the test suite
func TestHealthAggregatorTestSuite(t *testing.T) {
	suite.Run(t, new(HealthAggregatorTestSuite))
}

// HTTPHealthCheckerTestSuite is a test suite for HTTPHealthChecker
type HTTPHealthCheckerTestSuite struct {
	suite.Suite
	checker *HTTPHealthChecker
}

// SetupTest runs before each test
func (suite *HTTPHealthCheckerTestSuite) SetupTest() {
	suite.checker = NewHTTPHealthChecker(5 * time.Second)
}

// TestCheckHealth_Healthy tests checking a healthy service
func (suite *HTTPHealthCheckerTestSuite) TestCheckHealth_Healthy() {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(suite.T(), "/health", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	// Act
	result := suite.checker.CheckHealth("test-service", server.URL)

	// Assert
	assert.Equal(suite.T(), "healthy", result.Status)
	assert.Equal(suite.T(), "test-service", result.Name)
	assert.Equal(suite.T(), server.URL, result.URL)
	assert.Empty(suite.T(), result.Error)
}

// TestHTTPHealthCheckerTestSuite runs the test suite
func TestHTTPHealthCheckerTestSuite(t *testing.T) {
	suite.Run(t, new(HTTPHealthCheckerTestSuite))
}
