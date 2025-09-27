package config

import "os"

type Config struct {
	Port                     string
	PostgresConnectionString string
}

func LoadConfigFromEnvironment() *Config {
	return &Config{
		Port:                     os.Getenv("PORT"),
		PostgresConnectionString: os.Getenv("POSTGRES_CONNECTION_STRING"),
	}
}
