package config

import "os"

type Config struct {
	HttpPort                 string
	PostgresConnectionString string
}

func LoadConfigFromEnvironment() *Config {
	return &Config{
		HttpPort:                 os.Getenv("HTTP_PORT"),
		PostgresConnectionString: os.Getenv("POSTGRES_CONNECTION_STRING"),
	}
}
