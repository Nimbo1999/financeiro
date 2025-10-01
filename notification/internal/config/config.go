package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPPort                 string
	PostgresConnectionString string

	// RabbitMQ
	RabbitMQURL           string
	WelcomeQueueName      string
	OTPQueueName          string
	DLQName               string
	ConsumerPrefetchCount int

	// SMTP
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string
	SMTPTimeout  time.Duration

	// Templates
	TemplateDir string
}

func Load() *Config {
	return &Config{
		HTTPPort:                 getEnv("HTTP_PORT", "8080"),
		PostgresConnectionString: getEnv("POSTGRES_CONNECTION_STRING", "postgres://localhost:35432"),

		RabbitMQURL:           getEnv("RABBITMQ_URL", "amqp://localhost:5672"),
		WelcomeQueueName:      getEnv("WELCOME_QUEUE_NAME", "notification.welcome"),
		OTPQueueName:          getEnv("OTP_QUEUE_NAME", "notification.otp"),
		DLQName:               getEnv("DLQ_NAME", "notification.dlq"),
		ConsumerPrefetchCount: getEnvAsInt("CONSUMER_PREFETCH_COUNT", 10),

		SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     getEnvAsInt("SMTP_PORT", 587),
		SMTPUsername: getEnv("SMTP_USERNAME", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:     getEnv("SMTP_FROM", "noreply@localhost.com"),
		SMTPTimeout:  getEnvAsDuration("SMTP_TIMEOUT", 10*time.Second),

		TemplateDir: getEnv("TEMPLATE_DIR", "./internal/templates"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := getEnv(key, "")
	if value, err := time.ParseDuration(valueStr); err == nil {
		return value
	}
	return defaultValue
}
