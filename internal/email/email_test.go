package email

import (
	"log"
	"os"
	"testing"

	"backend/internal/config"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

var (
	mockUser = User{
		Name:        "Jack Gugolz",
		EmailAdress: "jack.gugolz@gmail.com",
	}

	mockEvent = Event{
		Name:     "Test Event",
		Date:     "30/2",
		StartsAt: "06:00",
		EndsAt:   "22:00",
		Location: "Plattan",
		URL:      "http://kthais.com",
		ImageURL: "https://kthais.com/files/__sized__/event/picture/Asort_Ventures_-_Website_Poster-crop-c0-5__0-5-1500x1000-70.jpg",
	}
)

func init() {
	// Load .env file if it exists
	if err := godotenv.Load("../../.env"); err != nil {
		log.Printf("No .env file found: %v", err)
	}
}

func TestMain(m *testing.M) {
	// Load config before running tests
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Printf("CONFIG ERROR: %+v", err)
		os.Exit(1)
	}
	log.Println("Loaded config:", cfg)

	// Initialize email service with config
	InitEmailService(cfg)
	log.Println("Initialized email service")

	// Run tests
	os.Exit(m.Run())
}

func TestSendRegistrationEmail(t *testing.T) {
	verificationURL := "https://kthais.com"

	err := SendRegistrationEmail(mockUser, verificationURL)
	assert.Nil(t, err, "SendRegistrationEmail should not return an error")
}

func TestSendLoginEmail(t *testing.T) {
	passwordResetURL := "http://kthais.com"

	err := sendLoginEmail(mockUser, passwordResetURL)
	assert.Nil(t, err, "sendLoginEmail should not return an error")
}

func TestSendEventRegistrationEmail(t *testing.T) {
	err := sendEventRegistrationEmail(mockUser, mockEvent)
	assert.Nil(t, err, "sendEventRegistrationEmail should not return an error")
}

func TestSendEventReminderEmail(t *testing.T) {
	err := sendEventReminderEmail(mockUser, mockEvent)
	assert.Nil(t, err, "sendEventReminderEmail should not return an error")
}

func TestSendEventCancelEmail(t *testing.T) {
	err := sendEventCancelEmail(mockUser, mockEvent)
	assert.Nil(t, err, "sendEventCancelEmail should not return an error")
}

func TestSendCustomEmail(t *testing.T) {
	err := sendCustomEmail(mockUser, "Custom email", "Custom email text :)", "Button text", "https://kthais.com", "")
	assert.Nil(t, err, "sendCustomEmail should not return an error")
}

func TestSendCustomEmailWithImage(t *testing.T) {
	err := sendCustomEmail(mockUser, "Custom email with image", "Custom email text :)", "Button text", "https://kthais.com", "https://kthais.com/files/__sized__/event/picture/Asort_Ventures_-_Website_Poster-crop-c0-5__0-5-1500x1000-70.jpg")
	assert.Nil(t, err, "sendCustomEmail should not return an error")
}
