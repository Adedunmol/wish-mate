package email

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
)

// go to google app passwords and create an app and use the details given

type Email struct {
	ToAddr   string `json:"to_addr"`
	Subject  string `json:"subject"`
	Template string `json:"template"`
	Vars     any    `json:"vars"`
}

func SendHTMLEmail(to []string, subject, htmlBody string) error {
	from := os.Getenv("FROM_EMAIL")
	password := os.Getenv("FROM_EMAIL_PASSWORD")
	smtpAddr := os.Getenv("SMTP_ADDR")
	smtpPort := os.Getenv("SMTP_PORT")
	adminEmail := os.Getenv("ADMIN_EMAIL")

	auth := smtp.PlainAuth("", from, password, smtpAddr)

	// Build full headers
	headers := make(map[string]string)
	headers["From"] = adminEmail
	headers["To"] = strings.Join(to, ", ")
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=\"UTF-8\""

	// Construct email message
	var msg strings.Builder
	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n")   // Blank line between headers and body
	msg.WriteString(htmlBody) // HTML content

	// Send email
	return smtp.SendMail(
		smtpAddr+":"+smtpPort,
		auth,
		from,
		to,
		[]byte(msg.String()),
	)
}

func parseTemplate(data Email) (bytes.Buffer, error) {

	tmplDir := os.Getenv("TEMPLATES_DIR")
	if tmplDir == "" {
		tmplDir = "./templates" // fallback
	}

	templatePath := filepath.Join(tmplDir, data.Template+".html")

	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return bytes.Buffer{}, fmt.Errorf("error parsing template: %v", err)
	}

	var rendered bytes.Buffer
	if err := tmpl.Execute(&rendered, data.Vars); err != nil {
		return bytes.Buffer{}, fmt.Errorf("error executing template: %v", err)
	}

	return rendered, nil
}

func (e Email) SendTemplateEmail() error {

	to := strings.Split(e.ToAddr, ",")

	rendered, err := parseTemplate(e)
	if err != nil {
		return err
	}

	return SendHTMLEmail(to, e.Subject, rendered.String())
}
