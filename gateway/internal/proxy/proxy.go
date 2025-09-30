package proxy

import (
	"io"
	"net/http"
	"time"
)

// ServiceProxy implements the Proxier interface using dependency injection
type ServiceProxy struct {
	httpClient     *http.Client
	requestCreator RequestCreator
	headerManager  HeaderManager
}

// NewServiceProxy creates a new ServiceProxy with injected dependencies
func NewServiceProxy(httpClient *http.Client, requestCreator RequestCreator, headerManager HeaderManager) *ServiceProxy {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 30 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
	}

	return &ServiceProxy{
		httpClient:     httpClient,
		requestCreator: requestCreator,
		headerManager:  headerManager,
	}
}

// ProxyRequest forwards an incoming request to a target service
func (sp *ServiceProxy) ProxyRequest(w http.ResponseWriter, r *http.Request, targetURL string) error {
	// Create the proxied request using the injected request creator
	proxyReq, err := sp.requestCreator.CreateProxyRequest(r, targetURL)
	if err != nil {
		return err
	}

	// Forward the request to the target service
	resp, err := sp.httpClient.Do(proxyReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Copy response headers using the injected header manager
	sp.headerManager.CopyResponseHeaders(w, resp)

	// Set status code
	w.WriteHeader(resp.StatusCode)

	// Copy response body
	_, err = io.Copy(w, resp.Body)
	return err
}