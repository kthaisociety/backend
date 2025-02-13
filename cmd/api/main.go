package main

import (
	"fmt"
	"log"

	"backend/internal/config"
	"backend/internal/handlers"
	"backend/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
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

	// Initialize router
	r := gin.Default()

	// Initialize handlers
	setupRoutes(r, db)

	// Run the server
	r.Run(":" + cfg.Server.Port)
}

func setupRoutes(r *gin.Engine, db *gorm.DB) {
	api := r.Group("/api/v1")

	// Health check
	api.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// Register all handlers
	handlers := []handlers.Handler{
		handlers.NewEventHandler(db),
		// Add other handlers here as they are implemented
		// handlers.NewUserHandler(db),
		// handlers.NewProfileHandler(db),
		// handlers.NewRegistrationHandler(db),
		// handlers.NewTeamMemberHandler(db),
	}

	for _, h := range handlers {
		h.Register(api)
	}
}
