package main

import (
	"fmt"
	"log"
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
	if err := db.Migrator().DropTable(&models.Profile{}, &models.User{}); err != nil {
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

	// Setup session middleware
	store := cookie.NewStore([]byte("secret"))  // In production, use a proper secret key
	r.Use(sessions.Sessions("kthais_session", store))

	// Add CORS middleware
	config := cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"}, // env variable
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	r.Use(cors.New(config))

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
