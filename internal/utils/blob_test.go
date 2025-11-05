package utils

import (
	"backend/internal/config"
	"fmt"
	"log"
	"testing"

	"github.com/joho/godotenv"
)

func TestS3init(t *testing.T) {
	fmt.Println("Running Blob Test")
	file := "../../.env"
	if err := godotenv.Load(file); err != nil {
		log.Printf("Could not load %s: %v", file, err)
	}
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("Error Loading Config: %s\n", err)
	}
	log.Printf("R2_Endpoint: %s\n R2_Access_Key: %s\n R2_Access_Key_Id: %s\n", cfg.R2_endpoint, cfg.R2_access_key, cfg.R2_access_key_id)
	fmt.Println("Init SDK")
	InitS3SDK(cfg)
}
