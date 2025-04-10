package email

import (
	"bytes"
	"fmt"
	"html/template"

	"backend/internal/config"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
)

// Email data struct, contains all fields used in emails
type EmailData struct {
	EmailConfig
	User     User
	Event    Event  // For event emails
	URL      string // For registration, password reset, and event emails
	ImageURL string
	Text     string // For custom text used in event survey and custom emails
}

// User struct used for tesing, should be replaced by the database implementation
type User struct {
	Name        string
	EmailAdress string
}

// Event struct used for tesing, should be replaced by the database implementation
type Event struct {
	Name     string
	Date     string
	StartsAt string
	EndsAt   string
	Location string
	URL      string
	ImageURL string
}

// Helper function to create a new EmailData struct with default values
func newEmailData() EmailData {
	return EmailData{
		EmailConfig: DefaultEmailConfig,
	}
}

// Add at package level
var emailConfig *config.Config

// Add an init function to set up the config
func InitEmailService(cfg *config.Config) {
	emailConfig = cfg
}

// Sends an email using Amazon SES.
//
// Parameters:
//   - recipient: The email adress of the recipient
//   - subject: The subject line of the email
//   - body: The HTML email body
//
// Returns:
//   - error: nil if the email was sent successfully, or an error if it failed
func sendEmail(recipient, subject, body string) error {
	if emailConfig == nil {
		return fmt.Errorf("email service not initialized")
	}

	// Create a new session using config values
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(emailConfig.SES.Region),
		Credentials: credentials.NewStaticCredentials(
			emailConfig.SES.AccessKeyID,
			emailConfig.SES.SecretAccessKey,
			"", // Token is only required for temporary security credentials retrieved via STS,
			// otherwise an empty string can be passed for this parameter.
		),
	})
	if err != nil {
		return fmt.Errorf("failed to create AWS session: %w", err)
	}

	// Create an SES session.
	svc := ses.New(sess)

	// Assemble the email.
	textBody := "Please use a HTML capable email client to view this email." // Fallback for non-HTML email clients
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
					Data:    aws.String(body),
				},
				Text: &ses.Content{
					Charset: aws.String("UTF-8"), // TODO: Replace hard-coded Charset
					Data:    aws.String(textBody),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String("UTF-8"), // TODO: Replace hard-coded Charset
				Data:    aws.String(subject),
			},
		},
		Source: aws.String(emailConfig.SES.Sender),
		ReplyToAddresses: []*string{
			aws.String(emailConfig.SES.ReplyTo),
		},
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
			return fmt.Errorf("failed to send email: %w", err)
		}

	}

	fmt.Println("Email send to address: " + recipient)
	fmt.Println(result)
	return nil
}

// Sends a registration confirmation email
//
// Parameters:
//   - user: The user struct for the recipient
//   - verificationURL: The URL for the confirmation
//
// Returns:
//   - error: nil if the email was sent successfully, or an error if it failed
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
	data.URL = verificationURL

	// Render the template into a buffer
	var htmlBody bytes.Buffer
	err = tmpl.ExecuteTemplate(&htmlBody, "base.html", data)
	if err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	// Define email parameters
	recipient := user.EmailAdress
	subject := "Complete Your KTHAIS Registration"

	return sendEmail(recipient, subject, htmlBody.String())
}

// Sends a login email
//
// Parameters:
//   - user: The user struct for the recipient
//   - loginURL: The URL for logging in
//
// Returns:
//   - error: nil if the email was sent successfully, or an error if it failed
func sendLoginEmail(user User, loginURL string) error {
	// Parse both base and password templates
	tmpl, err := template.ParseFiles(
		"templates/base.html",
		"templates/user/login.html",
	)
	if err != nil {
		return fmt.Errorf("failed to parse templates: %w", err)
	}

	// Prepare data for the email template
	data := newEmailData()
	data.User = user
	data.URL = loginURL

	// Render the template into a buffer
	var htmlBody bytes.Buffer
	err = tmpl.ExecuteTemplate(&htmlBody, "base.html", data)
	if err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	// Define email parameters
	recipient := user.EmailAdress
	subject := "Reset Your KTHAIS Password"

	return sendEmail(recipient, subject, htmlBody.String())
}

// Sends an event registration confirmation email
//
// Parameters:
//   - user: The user struct for the recipient
//   - event: The struct for the event
//
// Returns:
//   - error: nil if the email was sent successfully, or an error if it failed
func sendEventRegistrationEmail(user User, event Event) error {
	// Parse both base and password templates
	tmpl, err := template.ParseFiles(
		"templates/base.html",
		"templates/event/register.html",
	)
	if err != nil {
		return fmt.Errorf("failed to parse templates: %w", err)
	}

	// Prepare data for the email template
	data := newEmailData()
	data.User = user
	data.Event = event
	data.URL = event.URL
	data.Event = event

	// Render the template into a buffer
	var htmlBody bytes.Buffer
	err = tmpl.ExecuteTemplate(&htmlBody, "base.html", data)
	if err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	// Define email parameters
	recipient := user.EmailAdress
	subject := "Your registration to " + event.Name

	return sendEmail(recipient, subject, htmlBody.String())
}

// Sends an event reminder email
//
// Parameters:
//   - user: The user struct for the recipient
//   - event: The struct for the event
//
// Returns:
//   - error: nil if the email was sent successfully, or an error if it failed
func sendEventReminderEmail(user User, event Event) error {
	// Parse both base and password templates
	tmpl, err := template.ParseFiles(
		"templates/base.html",
		"templates/event/reminder.html",
	)
	if err != nil {
		return fmt.Errorf("failed to parse templates: %w", err)
	}

	// Prepare data for the email template
	data := newEmailData()
	data.User = user
	data.Event = event
	data.URL = event.URL
	data.Event = event

	// Render the template into a buffer
	var htmlBody bytes.Buffer
	err = tmpl.ExecuteTemplate(&htmlBody, "base.html", data)
	if err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	// Define email parameters
	recipient := user.EmailAdress
	subject := "Remember " + event.Name + "?"

	return sendEmail(recipient, subject, htmlBody.String())
}

// Sends an event cancelation email
//
// Parameters:
//   - user: The user struct for the recipient
//   - event: The struct for the event
//
// Returns:
//   - error: nil if the email was sent successfully, or an error if it failed
func sendEventCancelEmail(user User, event Event) error {
	// Parse both base and password templates
	tmpl, err := template.ParseFiles(
		"templates/base.html",
		"templates/event/cancel.html",
	)
	if err != nil {
		return fmt.Errorf("failed to parse templates: %w", err)
	}

	// Prepare data for the email template
	data := newEmailData()
	data.User = user
	data.Event = event
	data.URL = event.URL
	data.Event = event

	// Render the template into a buffer
	var htmlBody bytes.Buffer
	err = tmpl.ExecuteTemplate(&htmlBody, "base.html", data)
	if err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	// Define email parameters
	recipient := user.EmailAdress
	subject := "Remember " + event.Name + "?"

	return sendEmail(recipient, subject, htmlBody.String())
}

// Sends a custom email
//
// Parameters:
//   - user: The user struct for the recipient
//   - subject: The email subject
//   - customText: The email text
//   - customButtonText: The email button text
//   - customButtonURL: The email button URL
//   - customImageURL: The email image url (use an empty string for no image)
//
// Returns:
//   - error: nil if the email was sent successfully, or an error if it failed
func sendCustomEmail(user User, subject string, customText string, customButtonText string, customButtonURL string, customImageURL string) error {
	// Parse both base and password templates
	tmpl, err := template.ParseFiles(
		"templates/base.html",
		"templates/user/custom.html",
	)
	if err != nil {
		return fmt.Errorf("failed to parse templates: %w", err)
	}

	// Prepare data for the email template
	type CustomEmailData struct {
		EmailData
		CustomText       string
		CustomButtonText string
		CustomButtonURL  string
		CustomImageURL   string
	}
	data := CustomEmailData{
		EmailData:        newEmailData(),
		CustomText:       customText,
		CustomButtonText: customButtonText,
		CustomButtonURL:  customButtonURL,
		CustomImageURL:   customImageURL,
	}
	data.User = user

	// Render the template into a buffer
	var htmlBody bytes.Buffer
	err = tmpl.ExecuteTemplate(&htmlBody, "base.html", data)
	if err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	// Define email parameters
	recipient := user.EmailAdress

	return sendEmail(recipient, subject, htmlBody.String())
}
