package config

import (
	"fmt"
	"os"
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
}

func LoadConfig() (*Config, error) {
	cfg := &Config{}
	
	// Add debug logging
	host := getEnv("DB_HOST", "localhost")
	fmt.Printf("Database Host: %s\n", host)
	
	// Database config
	cfg.Database.Host = host
	cfg.Database.Port = getEnv("DB_PORT", "5432")
	cfg.Database.User = getEnv("DB_USER", "postgres")
	cfg.Database.Password = getEnv("DB_PASSWORD", "password")
	cfg.Database.DBName = getEnv("DB_NAME", "kthais")
	cfg.Database.SSLMode = getEnv("DB_SSLMODE", "disable")
	cfg.Server.Port = getEnv("SERVER_PORT", "8080")
	
	// OAuth config
	cfg.OAuth.GoogleClientID = getEnv("GOOGLE_CLIENT_ID", "")
	cfg.OAuth.GoogleClientSecret = getEnv("GOOGLE_CLIENT_SECRET", "")

	// Debug OAuth configuration
	fmt.Printf("Google Client ID: %s\n", maskString(cfg.OAuth.GoogleClientID))
	fmt.Printf("Google Client Secret: %s\n", maskString(cfg.OAuth.GoogleClientSecret))

	// Log OAuth configuration status
	if cfg.OAuth.GoogleClientID == "" || cfg.OAuth.GoogleClientSecret == "" {
		fmt.Println("Warning: Google OAuth credentials not configured. OAuth functionality will be disabled.")
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
