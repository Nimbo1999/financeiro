package config

import "os"

type Config struct {
	HTTPPort                 string
	GRPCPort                 string
	PostgresConnectionString string
	UserGRPCAddress          string
	RabbitMQURL              string
}

func LoadConfigFromEnvironment() *Config {
	return &Config{
		HTTPPort:                 os.Getenv("HTTP_PORT"),
		GRPCPort:                 os.Getenv("GRPC_PORT"),
		PostgresConnectionString: os.Getenv("POSTGRES_CONNECTION_STRING"),
		UserGRPCAddress:          os.Getenv("USER_GRPC_ADDRESS"),
		RabbitMQURL:              os.Getenv("RABBITMQ_URL"),
	}
}
