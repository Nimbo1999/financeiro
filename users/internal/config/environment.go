package config

import "os"

type Config struct {
	HttpPort                 string
	GrpcPort                 string
	PostgresConnectionString string
	RabbitMQURL              string
}

func LoadConfigFromEnvironment() *Config {
	return &Config{
		HttpPort:                 os.Getenv("HTTP_PORT"),
		GrpcPort:                 os.Getenv("GRPC_PORT"),
		PostgresConnectionString: os.Getenv("POSTGRES_CONNECTION_STRING"),
		RabbitMQURL:              os.Getenv("RABBITMQ_URL"),
	}
}
