package config

import (
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	App               AppConfig
	HTTP              ServerConfig
	GRPC              ServerConfig
	MySQL             MySQLConfig
	Log               LogConfig
	InternalEndpoints InternalEndpointsConfig
}

type AppConfig struct {
	ServiceName string
	APIKey      string
}

type ServerConfig struct {
	Host string
	Port string
}

type MySQLConfig struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type LogConfig struct {
	Level string
}

type InternalEndpointsConfig struct {
	AuthGRPCAddr string
}

// Load reads configuration from environment variables (and .env when present).
func Load() (*Config, error) {
	_ = godotenv.Load()

	mysqlDSN := os.Getenv("MYSQL_DSN")
	if mysqlDSN == "" {
		return nil, errors.New("MYSQL_DSN environment variable is required")
	}

	return &Config{
		App: AppConfig{
			ServiceName: getEnv("APP_SERVICE_NAME", "profile-service"),
			APIKey:      getEnv("APP_API_KEY", ""),
		},
		HTTP: ServerConfig{
			Host: getEnv("HTTP_HOST", "0.0.0.0"),
			Port: getEnv("HTTP_PORT", "8080"),
		},
		GRPC: ServerConfig{
			Host: getEnv("GRPC_HOST", "0.0.0.0"),
			Port: getEnv("GRPC_PORT", "9090"),
		},
		MySQL: MySQLConfig{
			DSN:             mysqlDSN,
			MaxOpenConns:    getIntEnv("MYSQL_MAX_OPEN_CONNS", 10),
			MaxIdleConns:    getIntEnv("MYSQL_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getDurationEnv("MYSQL_CONN_MAX_LIFETIME_MINUTES", 30*time.Minute),
		},
		Log: LogConfig{
			Level: getEnv("LOG_LEVEL", "info"),
		},
		InternalEndpoints: InternalEndpointsConfig{
			AuthGRPCAddr: getEnv("AUTH_SERVICE_GRPC_ADDR", "localhost:9090"),
		},
	}, nil
}

// getEnv returns the env value or the default if empty.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getIntEnv returns the int env value or the default if empty/invalid.
func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if n, err := strconv.Atoi(value); err == nil {
			return n
		}
	}
	return defaultValue
}

// getDurationEnv returns a minutes-based duration from env or the default.
func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if minutes, err := strconv.Atoi(value); err == nil {
			return time.Duration(minutes) * time.Minute
		}
	}
	return defaultValue
}
