package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Server         ServerConfig
	Services       ServiceConfig
	Security       SecurityConfig
	CircuitBreaker CircuitBreakerConfig
}

type ServerConfig struct {
	Port           int
	ReadTimeout    int
	WriteTimeout   int
	RequestTimeout int
}

type SecurityConfig struct {
	EnableCORS     bool
	AllowedOrigins []string
}

type CircuitBreakerConfig struct {
	Enabled      bool
	MaxFailures  int
	ResetTimeout int
}

type ServiceConfig struct {
	AuthServiceURL         string
	AuthServiceGRPCURL     string
	UserServiceURL         string
	NotificationServiceURL string
	TransactionsServiceURL string
	// Add more services as needed
}

func Load() (*Config, error) {
	port, err := getEnvAsInt("GATEWAY_PORT", 8080)
	if err != nil {
		return nil, fmt.Errorf("invalid GATEWAY_PORT: %w", err)
	}

	readTimeout, err := getEnvAsInt("GATEWAY_READ_TIMEOUT", 15)
	if err != nil {
		return nil, fmt.Errorf("invalid GATEWAY_READ_TIMEOUT: %w", err)
	}

	writeTimeout, err := getEnvAsInt("GATEWAY_WRITE_TIMEOUT", 15)
	if err != nil {
		return nil, fmt.Errorf("invalid GATEWAY_WRITE_TIMEOUT: %w", err)
	}

	requestTimeout, err := getEnvAsInt("GATEWAY_REQUEST_TIMEOUT", 30)
	if err != nil {
		return nil, fmt.Errorf("invalid GATEWAY_REQUEST_TIMEOUT: %w", err)
	}

	authServiceURL := getEnv("AUTH_SERVICE_URL", "http://localhost:8080")
	authServiceGRPCURL := getEnv("AUTH_SERVICE_GRPC_URL", "localhost:9090")
	userServiceURL := getEnv("USER_SERVICE_URL", "http://localhost:8080")
	notificationServiceURL := getEnv("NOTIFICATION_SERVICE_URL", "http://localhost:8080")
	transactionsServiceURL := getEnv("TRANSACTIONS_SERVICE_URL", "http://localhost:8080")

	// Security configuration
	enableCORS, err := getEnvAsBool("GATEWAY_ENABLE_CORS", true)
	if err != nil {
		return nil, fmt.Errorf("invalid GATEWAY_ENABLE_CORS: %w", err)
	}

	allowedOrigins := getEnvAsSlice("GATEWAY_ALLOWED_ORIGINS", []string{"*"})

	// Circuit breaker configuration
	enableCircuitBreaker, err := getEnvAsBool("GATEWAY_ENABLE_CIRCUIT_BREAKER", true)
	if err != nil {
		return nil, fmt.Errorf("invalid GATEWAY_ENABLE_CIRCUIT_BREAKER: %w", err)
	}

	circuitBreakerMaxFailures, err := getEnvAsInt("GATEWAY_CIRCUIT_BREAKER_MAX_FAILURES", 5)
	if err != nil {
		return nil, fmt.Errorf("invalid GATEWAY_CIRCUIT_BREAKER_MAX_FAILURES: %w", err)
	}

	circuitBreakerResetTimeout, err := getEnvAsInt("GATEWAY_CIRCUIT_BREAKER_RESET_TIMEOUT", 60)
	if err != nil {
		return nil, fmt.Errorf("invalid GATEWAY_CIRCUIT_BREAKER_RESET_TIMEOUT: %w", err)
	}

	return &Config{
		Server: ServerConfig{
			Port:           port,
			ReadTimeout:    readTimeout,
			WriteTimeout:   writeTimeout,
			RequestTimeout: requestTimeout,
		},
		Services: ServiceConfig{
			AuthServiceURL:         authServiceURL,
			AuthServiceGRPCURL:     authServiceGRPCURL,
			UserServiceURL:         userServiceURL,
			NotificationServiceURL: notificationServiceURL,
			TransactionsServiceURL: transactionsServiceURL,
		},
		Security: SecurityConfig{
			EnableCORS:     enableCORS,
			AllowedOrigins: allowedOrigins,
		},
		CircuitBreaker: CircuitBreakerConfig{
			Enabled:      enableCircuitBreaker,
			MaxFailures:  circuitBreakerMaxFailures,
			ResetTimeout: circuitBreakerResetTimeout,
		},
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) (int, error) {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue, nil
	}
	return strconv.Atoi(valueStr)
}

func getEnvAsBool(key string, defaultValue bool) (bool, error) {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue, nil
	}
	return strconv.ParseBool(valueStr)
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}
	// Split by comma
	var result []string
	for _, v := range splitAndTrim(valueStr, ",") {
		if v != "" {
			result = append(result, v)
		}
	}
	if len(result) == 0 {
		return defaultValue
	}
	return result
}

func splitAndTrim(s, sep string) []string {
	var result []string
	for _, v := range strings.Split(s, sep) {
		trimmed := strings.TrimSpace(v)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
