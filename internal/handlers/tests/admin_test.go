package handlers

import (
	"backend/internal/config"
	"backend/internal/handlers"
	"backend/internal/utils"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupRouter() (*gin.Engine, *gin.RouterGroup) {
	r := gin.Default()
	api := r.Group("/api/v1")
	return r, api
}

func getConfig() (*config.Config, error) {
	file := "../../../.env"
	loaded := false
	if _, err := os.Stat(file); err == nil {
		if err := godotenv.Load(file); err != nil {
			return nil, err
		}
		log.Println("Loaded .env file")
		loaded = true
	}
	if !loaded {
		return nil, errors.New("Env file not found")
	}
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func setupHandler(t *testing.T) (*gin.Engine, *config.Config) {
	r, api := setupRouter()
	cfg, err := getConfig()
	if err != nil {
		t.Fatalf("Error loading config: %s\n", err)
	}
	// Initialize DB
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.Database.Host, cfg.Database.User, cfg.Database.Password, cfg.Database.DBName, cfg.Database.Port, cfg.Database.SSLMode)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	adminHandler := handlers.NewAdminHandler(db, cfg)
	adminHandler.Register(api)
	return r, cfg
}

func TestIsAdmin(t *testing.T) {
	r, cfg := setupHandler(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/checkadmin/", nil)
	// Add cookies or headers as needed for authentication
	id, _ := uuid.NewUUID()
	token := utils.WriteJWT("fake.person@gmail.com", []string{"admin"}, id, cfg.JwtSigningKey, 15)
	req.AddCookie(&http.Cookie{Name: "jwt", Value: token})

	// Perform the request
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}
	//create json decoder
	_ =
	if w.Body["is_admin"] != true {
		t.Fatalf("Expected is_admin to be true, got %v", w.Body["is_admin"])
	}
}

func TestListAdmins(t *testing.T) {
	return
	r, cfg := setupHandler(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/admin/listadmins", nil)
	// Add cookies or headers as needed for authentication
	id, _ := uuid.NewUUID()
	token := utils.WriteJWT("fake.person@gmail.com", []string{"admin"}, id, cfg.JwtSigningKey, 15)
	req.AddCookie(&http.Cookie{Name: "jwt", Value: token})

	// Perform the request
	r.ServeHTTP(w, req)
}

func TestPromoteToAdmin(t *testing.T) {}
