package models

import (
	"backend/internal/config"
	"backend/internal/utils"
	"fmt"
	"log"
	"testing"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestBlob(t *testing.T) {
	file := "../../.env"
	if err := godotenv.Load(file); err != nil {
		log.Printf("Could not load %s: %v", file, err)
	}
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("Error Loading Config: %s\n", err)
	}
	r2, err := utils.InitS3SDK(cfg)
	if err != nil {
		t.Fatalf("Failed to init r2: %s\n", err)
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
		&BlobData{},
	)
	name := "test_file"
	ftype := ".txt"
	auid, _ := uuid.NewUUID()
	bd, err := NewBlobData(name, ftype, auid, []byte("asdfasdfasdfasdf"), db, r2)
	if err != nil {
		t.Fatalf("Error creating Blob Data: %s\n", err)
	}
	data, err := bd.GetData(r2)
	if err != nil {
		t.Fatalf("Failed to load blob data: %s\n", err)
	}
	log.Printf("Blob Data: %s\n", data)

}
