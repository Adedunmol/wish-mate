package email

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
	"os"
	"strings"
)

// go to google app passwords and create an app and use the details given

func SendHTMLEmail(to []string, subject, htmlBody string) error {

	auth := smtp.PlainAuth(
		"",
		os.Getenv("FROM_EMAIL"),
		os.Getenv("FROM_EMAIL_PASSWORD"),
		os.Getenv("FROM_EMAIL_SMTP"),
	)

	headers := "MIME-version: 1.0;\nContent-Type: text/html; charset: UTF-8"

	message := "Subject: " + subject + "\n" + headers + "\n\n" + htmlBody

	return smtp.SendMail(
		os.Getenv("SMTP_ADDR"),
		auth,
		os.Getenv("FROM_EMAIL"),
		to,
		[]byte(message),
	)
}

type Email struct {
	ToAddr   string            `json:"to_addr"`
	Subject  string            `json:"subject"`
	Template string            `json:"template"`
	Vars     map[string]string `json:"vars"`
}

func parseTemplate(data Email) (bytes.Buffer, error) {

	tmpl, err := template.ParseFiles("../templates/" + data.Template + ".html")
	if err != nil {
		return bytes.Buffer{}, fmt.Errorf("error parsing template: %v", err)
	}

	var rendered bytes.Buffer
	if err := tmpl.Execute(&rendered, data); err != nil {
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
