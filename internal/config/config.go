package config

import (
	"fmt"
	"log"
	"os"
	"strings"
)

type Config struct {
	Database struct {
		Host     string
		Port     string
		User     string
		Password string
		DBName   string
		SSLMode  string
	}
	Server struct {
		Port string
	}
	OAuth struct {
		GoogleClientID     string
		GoogleClientSecret string
	}
	AllowedOrigins []string
	BackendURL     string
	Redis          struct {
		Host     string
		Port     string
		Password string
	}
	SessionKey      string
	DevelopmentMode bool
}

func LoadConfig() (*Config, error) {
	cfg := &Config{}

	// Database config
	cfg.Database.Host = getEnv("DB_HOST", "localhost")
	cfg.Database.Port = getEnv("DB_PORT", "5432")
	cfg.Database.User = getEnv("DB_USER", "postgres")
	cfg.Database.Password = getEnv("DB_PASSWORD", "password")
	cfg.Database.DBName = getEnv("DB_NAME", "kthais")
	cfg.Database.SSLMode = getEnv("DB_SSLMODE", "disable")
	cfg.Server.Port = getEnv("SERVER_PORT", "8080")

	// Redis config
	cfg.Redis.Host = getEnv("REDIS_HOST", "localhost")
	cfg.Redis.Port = getEnv("REDIS_PORT", "6379")
	cfg.Redis.Password = getEnv("REDIS_PASSWORD", "")

	// Load allowed origins from environment variable
	// Format: comma-separated list of origins
	allowedOriginsStr := getEnv("ALLOWED_ORIGINS", "http://localhost:3000")
	cfg.AllowedOrigins = strings.Split(allowedOriginsStr, ",")

	// Trim spaces from each origin
	for i, origin := range cfg.AllowedOrigins {
		cfg.AllowedOrigins[i] = strings.TrimSpace(origin)
	}

	cfg.BackendURL = getEnv("BACKEND_URL", "http://localhost:8080")

	// OAuth config
	cfg.OAuth.GoogleClientID = getEnv("GOOGLE_CLIENT_ID", "")
	cfg.OAuth.GoogleClientSecret = getEnv("GOOGLE_CLIENT_SECRET", "")

	cfg.SessionKey = getEnv("SESSION_KEY", "")
	cfg.DevelopmentMode = getEnv("DEVELOPMENT", "true") == "true"

	// Debug OAuth configuration
	fmt.Printf("Google Client ID: %s\n", maskString(cfg.OAuth.GoogleClientID))
	fmt.Printf("Google Client Secret: %s\n", maskString(cfg.OAuth.GoogleClientSecret))

	// Log OAuth configuration status
	if cfg.OAuth.GoogleClientID == "" || cfg.OAuth.GoogleClientSecret == "" {
		log.Fatalf("Warning: Google OAuth credentials not configured. OAuth functionality will be disabled.")
		os.Exit(1)
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// maskString returns a masked version of the string for secure logging
func maskString(s string) string {
	if len(s) <= 8 {
		return "****"
	}
	return s[:4] + "..." + s[len(s)-4:]
}
