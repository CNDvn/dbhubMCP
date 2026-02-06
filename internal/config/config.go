package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all configuration for the MCP server
type Config struct {
	// Database configuration
	DBType          string // "mysql" or "postgres"
	DBHost          string
	DBPort          int
	DBName          string
	DBUser          string
	DBPassword      string
	DBMaxConns      int
	DBMaxIdleConns  int
	DBConnTimeout   time.Duration

	// Query execution limits
	QueryTimeout    time.Duration
	MaxRows         int

	// Server configuration
	LogLevel        string

	// Transport configuration
	TransportType   string   // "stdio" or "http"
	HTTPAddr        string   // ":8080"
	HTTPCORSOrigins []string // ["*"]
	HTTPAPIKey      string   // Optional
}

// LoadFromEnv loads configuration from environment variables
func LoadFromEnv() (*Config, error) {
	cfg := &Config{
		DBType:         getEnv("DB_TYPE", "mysql"),
		DBHost:         getEnv("DB_HOST", "localhost"),
		DBPort:         getEnvInt("DB_PORT", 3309),
		DBName:         getEnv("DB_NAME", "test"),
		DBUser:         getEnv("DB_USER", "root"),
		DBPassword:     getEnv("DB_PASSWORD", "123456"),
		DBMaxConns:     getEnvInt("DB_MAX_CONNS", 10),
		DBMaxIdleConns: getEnvInt("DB_MAX_IDLE_CONNS", 5),
		DBConnTimeout:  time.Duration(getEnvInt("DB_CONN_TIMEOUT_SEC", 10)) * time.Second,
		QueryTimeout:   time.Duration(getEnvInt("QUERY_TIMEOUT_SEC", 30)) * time.Second,
		MaxRows:        getEnvInt("MAX_ROWS", 1000),
		LogLevel:       getEnv("LOG_LEVEL", "info"),

		// Transport configuration
		TransportType:   getEnv("TRANSPORT_TYPE", "stdio"),
		HTTPAddr:        getEnv("HTTP_ADDR", ":8080"),
		HTTPCORSOrigins: getEnvSlice("HTTP_CORS_ORIGINS", []string{"*"}),
		HTTPAPIKey:      getEnv("HTTP_API_KEY", ""),
	}

	// Validate required fields
	if cfg.DBName == "" {
		return nil, fmt.Errorf("DB_NAME is required")
	}
	if cfg.DBUser == "" {
		return nil, fmt.Errorf("DB_USER is required")
	}
	if cfg.DBType != "mysql" && cfg.DBType != "postgres" {
		return nil, fmt.Errorf("DB_TYPE must be 'mysql' or 'postgres', got: %s", cfg.DBType)
	}
	if cfg.TransportType != "stdio" && cfg.TransportType != "http" {
		return nil, fmt.Errorf("TRANSPORT_TYPE must be 'stdio' or 'http', got: %s", cfg.TransportType)
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvSlice(key string, defaultValue []string) []string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return strings.Split(value, ",")
}
