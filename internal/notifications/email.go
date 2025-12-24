package notifications

import (
	"encoding/base64"
	"fmt"
	"net/smtp"
	"strings"
)

// EmailService handles email notifications
type EmailService struct {
	server          string
	port            int
	username        string
	passwordEncoded string
	fromAddress     string
}

// NewEmailService creates a new email notification service
func NewEmailService(server string, port int, username, passwordEncoded, fromAddress string) *EmailService {
	return &EmailService{
		server:          server,
		port:            port,
		username:        username,
		passwordEncoded: passwordEncoded,
		fromAddress:     fromAddress,
	}
}

// Send sends an email notification
func (e *EmailService) Send(toAddress, subject, body string) error {
	// Decode the password from base64
	password, err := DecodePassword(e.passwordEncoded)
	if err != nil {
		return fmt.Errorf("failed to decode password: %w", err)
	}

	// Build email message
	message := e.buildMessage(toAddress, subject, body)

	// Connect and send
	addr := fmt.Sprintf("%s:%d", e.server, e.port)
	auth := smtp.PlainAuth("", e.username, password, e.server)

	err = smtp.SendMail(addr, auth, e.fromAddress, []string{toAddress}, []byte(message))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// buildMessage constructs the email message with headers
func (e *EmailService) buildMessage(to, subject, body string) string {
	headers := make(map[string]string)
	headers["From"] = e.fromAddress
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	var message strings.Builder
	for k, v := range headers {
		message.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	message.WriteString("\r\n")
	message.WriteString(body)

	return message.String()
}

// EncodePassword encodes a password using base64 for storage
func EncodePassword(password string) string {
	return base64.StdEncoding.EncodeToString([]byte(password))
}

// DecodePassword decodes a base64-encoded password
func DecodePassword(encoded string) (string, error) {
	if encoded == "" {
		return "", nil
	}
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

// TestConnection tests the email configuration
func (e *EmailService) TestConnection(toAddress, password string) error {
	subject := "nonraidUI Test Notification"
	body := `
<html>
<body>
	<h2>Test Notification from nonraidUI</h2>
	<p>This is a test email to verify your email notification settings are configured correctly.</p>
	<p>If you received this email, your SMTP settings are working properly!</p>
	<hr>
	<p><small>Sent from nonraidUI</small></p>
</body>
</html>
`

	// Build message
	message := e.buildMessage(toAddress, subject, body)

	// Connect and send with plaintext password
	addr := fmt.Sprintf("%s:%d", e.server, e.port)
	auth := smtp.PlainAuth("", e.username, password, e.server)

	err := smtp.SendMail(addr, auth, e.fromAddress, []string{toAddress}, []byte(message))
	if err != nil {
		return fmt.Errorf("failed to send test email: %w", err)
	}

	return nil
}
