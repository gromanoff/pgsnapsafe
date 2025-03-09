package email

import (
	"crypto/tls"
	"fmt"
	"log"

	"gopkg.in/gomail.v2"
)

type SMTPClient struct {
	Host     string
	Port     int
	Username string
	Password string
	Sender   string
}

// NewSMTPClient - Creating a new SMTP client
func NewSMTPClient(host string, port int, username, password, sender string) *SMTPClient {
	return &SMTPClient{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		Sender:   sender,
	}
}

func (s *SMTPClient) SendEmail(to, subject, body string) error {
	// Creating a new message
	m := gomail.NewMessage()
	m.SetHeader("From", s.Username)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	// Setting up TLS
	d := gomail.NewDialer(s.Host, s.Port, s.Username, s.Password)
	d.TLSConfig = &tls.Config{
		ServerName:         s.Host, // Server name verification
		InsecureSkipVerify: false,  // Disabling insecure mode
	}

	// Attempting to send the email
	err := d.DialAndSend(m)
	if err != nil {
		log.Printf("Error sending email to %s: %v", to, err)
		return fmt.Errorf("failed to send email to %s: %w", to, err)
	}

	log.Printf("Email successfully sent to %s", to)
	return nil
}
