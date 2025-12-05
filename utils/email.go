package utils

import (
	"fmt"
	"net/smtp"
	"os"
	"strings"
)

// EmailConfig holds email configuration
type EmailConfig struct {
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
	FromName     string
}

// GetEmailConfig loads email configuration from environment variables
func GetEmailConfig() *EmailConfig {
	return &EmailConfig{
		SMTPHost:     getEnvOrDefault("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     getEnvOrDefault("SMTP_PORT", "587"),
		SMTPUsername: getEnvOrDefault("SMTP_USERNAME", ""),
		SMTPPassword: getEnvOrDefault("SMTP_PASSWORD", ""),
		FromEmail:    getEnvOrDefault("FROM_EMAIL", ""),
		FromName:     getEnvOrDefault("FROM_NAME", "CarWash App"),
	}
}

// SendEmail sends an email using SMTP
func SendEmail(to, subject, body string) error {
	config := GetEmailConfig()

	// Check if configuration is complete
	if config.SMTPUsername == "" || config.SMTPPassword == "" || config.FromEmail == "" {
		// Log email to console for development/testing
		fmt.Println("==================================================")
		fmt.Println("📧 [MOCK EMAIL] SMTP Config Missing - Logging Email")
		fmt.Printf("To: %s\n", to)
		fmt.Printf("Subject: %s\n", subject)
		fmt.Println("Body:")
		fmt.Println(body)
		fmt.Println("==================================================")
		return nil // Return success so flow continues
	}

	// Set up authentication
	auth := smtp.PlainAuth("", config.SMTPUsername, config.SMTPPassword, config.SMTPHost)

	// Compose message
	msg := []byte(fmt.Sprintf(
		"To: %s\r\n"+
			"From: %s <%s>\r\n"+
			"Subject: %s\r\n"+
			"Content-Type: text/html; charset=UTF-8\r\n"+
			"\r\n"+
			"%s\r\n",
		to, config.FromName, config.FromEmail, subject, body))

	// Send email
	err := smtp.SendMail(
		config.SMTPHost+":"+config.SMTPPort,
		auth,
		config.FromEmail,
		[]string{to},
		msg,
	)

	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}

// SendVerificationEmail sends email verification code
func SendVerificationEmail(userEmail, userName, token string) error {
	subject := "Verify Your Email - CarWash App"
	body := fmt.Sprintf(`
		<html>
		<body>
			<h2>Welcome to CarWash App!</h2>
			<p>Hi %s,</p>
			<p>Please use the following code to verify your email address:</p>
			<h1 style="color: #2563EB; letter-spacing: 5px;">%s</h1>
			<p>This code will expire in 24 hours.</p>
			<p>If you didn't create an account, please ignore this email.</p>
		</body>
		</html>
	`, userName, token)

	return SendEmail(userEmail, subject, body)
}

// SendBookingConfirmationEmail sends a booking confirmation email
func SendBookingConfirmationEmail(userEmail, userName, carwashName, bookingTime string) error {
	subject := "Booking Confirmation - CarWash App"
	body := fmt.Sprintf(`
		<html>
		<body>
			<h2>Booking Confirmed!</h2>
			<p>Hi %s,</p>
			<p>Your carwash booking has been confirmed:</p>			
			<ul>

				<li><strong>Carwash:</strong> %s</li>
				<li><strong>Time:</strong> %s</li>

			</ul>
			<p>We'll notify you when your booking is accepted by the business.</p>
			<p>Thank you for using CarWash App!</p>
		</body>
		</html>
	`, userName, carwashName, bookingTime)

	return SendEmail(userEmail, subject, body)

}

// SendOrderUpdateEmail sends order status update email
func SendOrderUpdateEmail(userEmail, userName, status, details string) error {
	subject := fmt.Sprintf("Order Update - %s", strings.Title(status))
	body := fmt.Sprintf(`
		<html>
		<body>
			<h2>Order Update</h2>
			<p>Hi %s,</p>
			<p>Your order status has been updated to: <strong>%s</strong></p>
			<p>%s</p>
			<p>Thank you for using CarWash App!</p>
		</body>
		</html>
	`, userName, strings.Title(status), details)

	return SendEmail(userEmail, subject, body)
}

// Helper function to get environment variable with default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
