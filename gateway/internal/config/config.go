package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Server   ServerConfig
	Services ServiceConfig
}

type ServerConfig struct {
	Port         int
	ReadTimeout  int
	WriteTimeout int
}

type ServiceConfig struct {
	AuthServiceURL string
	UserServiceURL string
	// Add more services as needed
}

func LoadFromEnv() (*Config, error) {
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

	authServiceURL := getEnv("AUTH_SERVICE_URL", "http://localhost:8081")
	userServiceURL := getEnv("USER_SERVICE_URL", "http://localhost:8082")

	return &Config{
		Server: ServerConfig{
			Port:         port,
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
		},
		Services: ServiceConfig{
			AuthServiceURL: authServiceURL,
			UserServiceURL: userServiceURL,
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
