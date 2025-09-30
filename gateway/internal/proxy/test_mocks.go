package proxy

import (
	"net/http"

	"github.com/stretchr/testify/mock"
)

// Test mocks - these are only used in tests and defined in the same package to avoid import cycles

// MockProxier is a mock implementation of the Proxier interface
type MockProxier struct {
	mock.Mock
}

func (m *MockProxier) ProxyRequest(w http.ResponseWriter, r *http.Request, targetURL string) error {
	args := m.Called(w, r, targetURL)
	return args.Error(0)
}

// MockHeaderManager is a mock implementation of the HeaderManager interface
type MockHeaderManager struct {
	mock.Mock
}

func (m *MockHeaderManager) CopyRequestHeaders(dst, src *http.Request) {
	m.Called(dst, src)
}

func (m *MockHeaderManager) CopyResponseHeaders(w http.ResponseWriter, resp *http.Response) {
	m.Called(w, resp)
}

func (m *MockHeaderManager) IsHopByHopHeader(header string) bool {
	args := m.Called(header)
	return args.Bool(0)
}

// MockRequestCreator is a mock implementation of the RequestCreator interface
type MockRequestCreator struct {
	mock.Mock
}

func (m *MockRequestCreator) CreateProxyRequest(r *http.Request, targetURL string) (*http.Request, error) {
	args := m.Called(r, targetURL)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*http.Request), args.Error(1)
}

// MockLogger is a mock implementation of the Logger interface
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) LogRequest(method, path, targetURL, serviceName string) {
	m.Called(method, path, targetURL, serviceName)
}

func (m *MockLogger) LogResponse(method, path, targetURL string, statusCode int, duration string) {
	m.Called(method, path, targetURL, statusCode, duration)
}

func (m *MockLogger) LogError(method, path, targetURL string, duration string, err error) {
	m.Called(method, path, targetURL, duration, err)
}

// MockHealthChecker is a mock implementation of the HealthChecker interface
type MockHealthChecker struct {
	mock.Mock
}

func (m *MockHealthChecker) CheckHealth(name, url string) ServiceHealth {
	args := m.Called(name, url)
	return args.Get(0).(ServiceHealth)
}

// MockServiceRegistry is a mock implementation of the ServiceRegistry interface
type MockServiceRegistry struct {
	mock.Mock
}

func (m *MockServiceRegistry) GetServices() []ServiceConfig {
	args := m.Called()
	return args.Get(0).([]ServiceConfig)
}