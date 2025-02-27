package email

import (
	"bytes"
	"fmt"
	"html/template"

	//go get -u github.com/aws/aws-sdk-go
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
)

type EmailData struct {
	EmailConfig
	User             User
	VerificationURL  string // For registration emails
	PasswordResetURL string // For password reset emails
}

type User struct {
	Name        string
	EmailAdress string
}

// Helper function to create a new EmailData struct with default values
func newEmailData() EmailData {
	return EmailData{
		EmailConfig: DefaultEmailConfig,
	}
}

func sendEmail(sender, recipient, subject, htmlBody string, body string) error {
	// Create a new session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("eu-north-1")}, // TODO: Is this needed? Should be in the SES config
	)
	if err != nil {
		return fmt.Errorf("failed to create AWS session: %w", err)
	}

	// Create an SES session.
	svc := ses.New(sess)

	// Assemble the email.
	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			CcAddresses: []*string{},
			ToAddresses: []*string{
				aws.String(recipient),
			},
		},
		Message: &ses.Message{ // TODO: Use the body function parameter
			Body: &ses.Body{
				Html: &ses.Content{ // How should we handle the Html body?
					Charset: aws.String("UTF-8"), // TODO: Replace hard-coded Charset (if needed)
					Data:    aws.String(htmlBody),
				},
				Text: &ses.Content{
					Charset: aws.String("UTF-8"), // TODO: Replace hard-coded Charset
					Data:    aws.String(body),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String("UTF-8"), // TODO: Replace hard-coded Charset
				Data:    aws.String(subject),
			},
		},
		Source: aws.String(sender),
		// Uncomment to use a configuration set
		//ConfigurationSetName: aws.String(ConfigurationSet),
	}

	// Attempt to send the email.
	result, err := svc.SendEmail(input)

	// Display error messages if they occur.
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ses.ErrCodeMessageRejected:
				return fmt.Errorf("message rejected: %s", aerr.Error())
			case ses.ErrCodeMailFromDomainNotVerifiedException:
				return fmt.Errorf("mail from domain not verified: %s", aerr.Error())
			case ses.ErrCodeConfigurationSetDoesNotExistException:
				return fmt.Errorf("configuration set does not exist: %s", aerr.Error())
			default:
				return fmt.Errorf("SES error: %s", aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			return fmt.Errorf("failed to send email: %w", err)
		}

	}

	fmt.Println("Email Sent to address: " + recipient)
	fmt.Println(result)
	return nil
}

// This functions will need to be edited when we have decided on how
// users and associated attributes are stored.
func sendBaseEmail(user User) error {
	// Parse and render the email template
	tmpl, err := template.ParseFiles("templates/base.html")

	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Create data with default config
	data := newEmailData()
	data.User = user

	// Render the template into a buffer
	var htmlBody bytes.Buffer
	err = tmpl.ExecuteTemplate(&htmlBody, "base.html", data)
	if err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	// Define email parameters
	sender := "jack@gugolz.se"
	recipient := user.EmailAdress
	subject := "Test email" // TODO: Change to actual subject
	plainTextBody := "Please use a HTML capable email client to view this email."

	return sendEmail(sender, recipient, subject, htmlBody.String(), plainTextBody)
}

// SendRegistrationEmail sends a registration confirmation email to the user
func SendRegistrationEmail(user User, verificationURL string) error {
	// Parse both base and registration templates
	tmpl, err := template.ParseFiles(
		"templates/base.html",
		"templates/user/register.html",
	)
	if err != nil {
		return fmt.Errorf("failed to parse templates: %w", err)
	}

	// Prepare data for the email template
	data := newEmailData()
	data.User = user
	data.VerificationURL = verificationURL

	// Render the template into a buffer
	var htmlBody bytes.Buffer
	err = tmpl.ExecuteTemplate(&htmlBody, "base.html", data)
	if err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	// Define email parameters
	sender := "jack@gugolz.se"
	recipient := user.EmailAdress
	subject := "Complete Your KTHAIS Registration"
	plainTextBody := "Please use a HTML capable email client to view this email. To complete your registration, please visit: " + verificationURL

	return sendEmail(sender, recipient, subject, htmlBody.String(), plainTextBody)
}
