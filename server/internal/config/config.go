package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration values for the server.
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	CORS     CORSConfig
	RateLimit RateLimitConfig
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Port    string
	GinMode string
}

// DatabaseConfig holds PostgreSQL connection configuration.
type DatabaseConfig struct {
	URL string
}

// JWTConfig holds JWT authentication configuration.
type JWTConfig struct {
	AccessSecret  string
	RefreshSecret string
	AccessExpiry  time.Duration
	RefreshExpiry time.Duration
}

// CORSConfig holds CORS policy configuration.
type CORSConfig struct {
	AllowedOrigins []string
}

// RateLimitConfig holds rate limiting configuration.
type RateLimitConfig struct {
	RPS   float64
	Burst int
}

// Load reads environment variables (with optional .env file) and returns
// a fully-populated Config struct. It returns an error if any required
// variable is missing or has an invalid value.
func Load() (*Config, error) {
	// Attempt to load .env; ignore error if the file does not exist because
	// environment variables may be provided by the host (Docker, systemd, etc.).
	_ = godotenv.Load()

	port := getEnvOrDefault("SERVER_PORT", "8080")
	ginMode := getEnvOrDefault("GIN_MODE", "debug")

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("config: DATABASE_URL is required")
	}

	accessSecret := os.Getenv("JWT_ACCESS_SECRET")
	if accessSecret == "" {
		return nil, fmt.Errorf("config: JWT_ACCESS_SECRET is required")
	}

	refreshSecret := os.Getenv("JWT_REFRESH_SECRET")
	if refreshSecret == "" {
		return nil, fmt.Errorf("config: JWT_REFRESH_SECRET is required")
	}

	accessExpiry, err := time.ParseDuration(getEnvOrDefault("JWT_ACCESS_EXPIRY", "15m"))
	if err != nil {
		return nil, fmt.Errorf("config: invalid JWT_ACCESS_EXPIRY: %w", err)
	}

	refreshExpiry, err := time.ParseDuration(getEnvOrDefault("JWT_REFRESH_EXPIRY", "168h"))
	if err != nil {
		return nil, fmt.Errorf("config: invalid JWT_REFRESH_EXPIRY: %w", err)
	}

	origins := strings.Split(getEnvOrDefault("CORS_ALLOWED_ORIGINS", "http://localhost:3000"), ",")
	for i := range origins {
		origins[i] = strings.TrimSpace(origins[i])
	}

	rps, err := strconv.ParseFloat(getEnvOrDefault("RATE_LIMIT_RPS", "10"), 64)
	if err != nil {
		return nil, fmt.Errorf("config: invalid RATE_LIMIT_RPS: %w", err)
	}

	burst, err := strconv.Atoi(getEnvOrDefault("RATE_LIMIT_BURST", "20"))
	if err != nil {
		return nil, fmt.Errorf("config: invalid RATE_LIMIT_BURST: %w", err)
	}

	return &Config{
		Server: ServerConfig{
			Port:    port,
			GinMode: ginMode,
		},
		Database: DatabaseConfig{
			URL: dbURL,
		},
		JWT: JWTConfig{
			AccessSecret:  accessSecret,
			RefreshSecret: refreshSecret,
			AccessExpiry:  accessExpiry,
			RefreshExpiry: refreshExpiry,
		},
		CORS: CORSConfig{
			AllowedOrigins: origins,
		},
		RateLimit: RateLimitConfig{
			RPS:   rps,
			Burst: burst,
		},
	}, nil
}

// getEnvOrDefault returns the value of the environment variable named by key,
// or the provided default value if the variable is not set.
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
