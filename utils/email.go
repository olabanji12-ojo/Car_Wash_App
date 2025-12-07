package utils

import (
	"crypto/tls"
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

// SendEmail sends an email using SMTP (supports both port 587 and 465)
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

	// Compose message
	msg := []byte(fmt.Sprintf(
		"To: %s\r\n"+
			"From: %s <%s>\r\n"+
			"Subject: %s\r\n"+
			"Content-Type: text/html; charset=UTF-8\r\n"+
			"\r\n"+
			"%s\r\n",
		to, config.FromName, config.FromEmail, subject, body))

	// Use different methods based on port
	if config.SMTPPort == "465" {
		// Port 465 uses SSL/TLS from the start
		return sendEmailSSL(config, to, msg)
	}

	// Port 587 uses STARTTLS
	return sendEmailSTARTTLS(config, to, msg)
}

// sendEmailSTARTTLS sends email using port 587 with STARTTLS
func sendEmailSTARTTLS(config *EmailConfig, to string, msg []byte) error {
	fmt.Printf("🔌 Connecting to SMTP server %s:%s via STARTTLS...\n", config.SMTPHost, config.SMTPPort)
	auth := smtp.PlainAuth("", config.SMTPUsername, config.SMTPPassword, config.SMTPHost)

	err := smtp.SendMail(
		config.SMTPHost+":"+config.SMTPPort,
		auth,
		config.FromEmail,
		[]string{to},
		msg,
	)

	if err != nil {
		fmt.Printf("❌ SMTP Error: %v\n", err)
		return fmt.Errorf("failed to send email: %v", err)
	}

	fmt.Println("✅ Email sent via STARTTLS")
	return nil
}

// sendEmailSSL sends email using port 465 with SSL/TLS
func sendEmailSSL(config *EmailConfig, to string, msg []byte) error {
	fmt.Printf("🔌 Connecting to SMTP server %s:%s via SSL...\n", config.SMTPHost, config.SMTPPort)

	// TLS config
	tlsConfig := &tls.Config{
		ServerName: config.SMTPHost,
	}

	// Connect to SMTP server with TLS
	conn, err := tls.Dial("tcp", config.SMTPHost+":"+config.SMTPPort, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %v", err)
	}
	defer conn.Close()
	fmt.Println("✅ SSL Connection established")

	// Create SMTP client
	client, err := smtp.NewClient(conn, config.SMTPHost)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %v", err)
	}
	defer client.Quit()

	// Authenticate
	fmt.Println("🔐 Authenticating...")
	auth := smtp.PlainAuth("", config.SMTPUsername, config.SMTPPassword, config.SMTPHost)
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP authentication failed: %v", err)
	}
	fmt.Println("✅ Authentication successful")

	// Set sender
	if err = client.Mail(config.FromEmail); err != nil {
		return fmt.Errorf("failed to set sender: %v", err)
	}

	// Set recipient
	if err = client.Rcpt(to); err != nil {
		return fmt.Errorf("failed to set recipient: %v", err)
	}

	// Send message
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %v", err)
	}

	_, err = writer.Write(msg)
	if err != nil {
		return fmt.Errorf("failed to write message: %v", err)
	}

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("failed to close writer: %v", err)
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
