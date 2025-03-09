package email

import (
	"bytes"
	"fmt"
	v "github.com/spf13/viper"
	"gopkg.in/gomail.v2"
	"html/template"
	"log"
	"path/filepath"
	"time"
)

// SendEmail - Function to send an email
func SendEmail(smtClient *SMTPClient, email, filename string) error {
	htmlBody, textBody, err := generateEmail(filename)
	if err != nil {
		return err
	}

	// Sending HTML email for other domains
	return sendHTMLEmail(smtClient, email, textBody, htmlBody)
}

// Generate email for verification
func generateEmail(filename string) (htmlBody, textBody string, err error) {
	var htmlFileName string
	if v.GetString("email_lang") == "ru" {
		htmlFileName = "ru.html"
	} else {
		htmlFileName = "en.html"
	}

	htmlBody, err = renderTemplate(htmlFileName, struct {
		FileName  string
		Timestamp string
	}{
		FileName:  filename,
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
	})

	if err != nil {
		return "", "", fmt.Errorf("error generating HTML for email confirmation: %w", err)
	}
	textBody = "Backup created successfully"
	return htmlBody, textBody, nil
}

// Render HTML template
func renderTemplate(fileName string, data interface{}) (string, error) {
	templatePath := filepath.Join("pkg/email/template", fileName)
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		log.Printf("Error loading template: %v", err)
		return "", err
	}

	var body bytes.Buffer
	err = tmpl.Execute(&body, data)
	if err != nil {
		log.Printf("Error processing template: %v", err)
		return "", err
	}

	return body.String(), nil
}

// sendHTMLEmail - Send HTML email
func sendHTMLEmail(smtpClient *SMTPClient, email, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", smtpClient.Sender)
	m.SetHeader("To", email)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer(smtpClient.Host, smtpClient.Port, smtpClient.Username, smtpClient.Password)
	err := d.DialAndSend(m)
	if err != nil {
		log.Printf("Error sending HTML email: %v", err)
		return err
	}

	log.Printf("HTML email successfully sent to %s", email)
	return nil
}
