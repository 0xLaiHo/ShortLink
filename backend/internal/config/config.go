package config

import "os"

// Config holds all configuration for the application
type Config struct {
	Server ServerConfig
	Redis  RedisConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port    string
	BaseURL string
	Mode    string
}

// RedisConfig holds Redis-related configuration
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:    getEnv("SERVER_PORT", "8080"),
			BaseURL: getEnv("BASE_URL", "http://localhost:8080"),
			Mode:    getEnv("GIN_MODE", "debug"),
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       0,
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
