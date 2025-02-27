package email

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSendEmail(t *testing.T) {
	sender := "jack@gugolz.se"
	recipient := "jack.gugolz@gmail.com"
	subject := "Test Email"

	htmlBody, err := os.ReadFile("templates/user/register.html")
	if err != nil {
		t.Fatalf("Failed to read email template: %v", err)
	}

	body := "Please use an html email client to view this email."

	err = sendEmail(sender, recipient, subject, string(htmlBody), body)

	assert.NotEmpty(t, string(htmlBody), "HTML body should not be empty")
	assert.Nil(t, err, "SendEmail should not return an error")
}

func TestSendBaseEmail(t *testing.T) {
	mockUser := User{
		Name:        "Jack Gugolz",
		EmailAdress: "jack.gugolz@gmail.com",
	}

	err := sendBaseEmail(mockUser)
	assert.Nil(t, err, "SendBaseEmail should not return an error")
}

func TestSendRegistrationEmail(t *testing.T) {
	mockUser := User{
		Name:        "Jack Gugolz",
		EmailAdress: "jack.gugolz@gmail.com",
	}
	verificationURL := "https://kthais.com/verify?token=abc123"

	err := SendRegistrationEmail(mockUser, verificationURL)
	assert.Nil(t, err, "SendRegistrationEmail should not return an error")
}
