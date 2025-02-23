package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"backend/internal/auth"
	"backend/internal/config"
	"backend/internal/handlers"
	"backend/internal/middleware"
	"backend/internal/models"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func generateSecureKey() ([]byte, error) {
	key := make([]byte, 32) // 256 bits
	_, err := rand.Read(key)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func setupStore() (sessions.Store, error) {
	// Generate a secure key or load from environment
	key := os.Getenv("SESSION_KEY")
	var sessionKey []byte
	
	if key == "" {
		var err error
		sessionKey, err = generateSecureKey()
		if err != nil {
			return nil, fmt.Errorf("failed to generate session key: %v", err)
		}
		// Optionally log the key for first-time setup
		log.Printf("Generated new session key: %s", base64.StdEncoding.EncodeToString(sessionKey))
	} else {
		var err error
		sessionKey, err = base64.StdEncoding.DecodeString(key)
		if err != nil {
			return nil, fmt.Errorf("invalid session key in environment: %v", err)
		}
	}

	// Create store with secure settings
	store := cookie.NewStore(sessionKey)
	store.Options(sessions.Options{
		Path:     "/",              // Cookie is valid for entire site
		MaxAge:   86400 * 7,        // 7 days
		HttpOnly: true,             // Prevent JavaScript access
		Secure:   true,             // Require HTTPS
		SameSite: http.SameSiteStrictMode,
	})

	return store, nil
}

func main() {
	// Get working directory
	wd, err := os.Getwd()
	if err != nil {
		log.Printf("Warning: Could not get working directory: %v", err)
	} else {
		log.Printf("Working directory: %s", wd)
	}

	// Try loading environment variables from multiple locations
	envFiles := []string{
		".env",
		".env.local",
		"../../.env",
		"../../.env.local",
	}

	for _, file := range envFiles {
		if err := godotenv.Load(file); err == nil {
			log.Printf("Successfully loaded environment from: %s", file)
			break
		} else {
			log.Printf("Could not load %s: %v", file, err)
		}
	}

	// Explicitly set environment variables from .env.local if not already set
	if os.Getenv("GOOGLE_CLIENT_ID") == "" {
		os.Setenv("GOOGLE_CLIENT_ID", "1056806786097-gatubd3kl6c1e027n0tbi7u3au5o27u7.apps.googleusercontent.com")
	}
	if os.Getenv("GOOGLE_CLIENT_SECRET") == "" {
		os.Setenv("GOOGLE_CLIENT_SECRET", "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")
	}

	// Print environment variables (masked)
	log.Printf("GOOGLE_CLIENT_ID present: %v", os.Getenv("GOOGLE_CLIENT_ID") != "")
	log.Printf("GOOGLE_CLIENT_SECRET present: %v", os.Getenv("GOOGLE_CLIENT_SECRET") != "")

	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Initialize DB
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.Database.Host, cfg.Database.User, cfg.Database.Password, cfg.Database.DBName, cfg.Database.Port, cfg.Database.SSLMode)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Drop existing tables to handle schema changes
	if err := db.Migrator().DropTable(&models.User{}, &models.Profile{}); err != nil {
		log.Printf("Warning: Failed to drop tables: %v", err)
	}

	// Auto migrate the schema
	err = db.AutoMigrate(
		&models.User{},
		&models.Profile{},
		&models.Event{},
		&models.Registration{},
		&models.TeamMember{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Initialize auth
	if err := auth.InitAuth(); err != nil {
		log.Printf("Warning: OAuth initialization failed: %v", err)
	}

	// Initialize router
	r := gin.Default()

	// Add CORS middleware with more permissive settings for development
	corsConfig := cors.Config{
		AllowOrigins:     []string{cfg.FrontendURL},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "Cookie"},
		ExposeHeaders:    []string{"Set-Cookie"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	r.Use(cors.New(corsConfig))

	// Setup session store with development-friendly settings
	store, err := setupStore()
	if err != nil {
		log.Fatal("Failed to setup session store:", err)
	}
	
	// Modify cookie settings for development
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
		Secure:   false, // Set to false for development
		SameSite: http.SameSiteLaxMode, // Use Lax for development
	})

	r.Use(sessions.Sessions("kthais_session", store))

	// Initialize handlers
	setupRoutes(r, db)

	// Run the server
	r.Run(":" + cfg.Server.Port)
}

func setupRoutes(r *gin.Engine, db *gorm.DB) {
	api := r.Group("/api/v1")

	// Public routes
	api.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Register all handlers
	allHandlers := []handlers.Handler{
		handlers.NewEventHandler(db),
		auth.NewAuthHandler(db),
	}

	for _, h := range allHandlers {
		h.Register(api)
	}

	// Protected routes
	protected := api.Group("/protected")
	protected.Use(middleware.AuthRequired())
	protectedHandler := handlers.NewProtectedHandler(db)
	protectedHandler.Register(protected)
}
