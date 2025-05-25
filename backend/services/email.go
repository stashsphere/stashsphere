package services

import (
	"fmt"
	"strings"

	"github.com/emersion/go-sasl"
	"github.com/emersion/go-smtp"
	"github.com/stashsphere/backend/config"
)

type EmailService interface {
	Deliver(identifier string, subject string, body string) error
}

type SMTPEmailService struct {
	config config.StashSphereMailConfig
}

func NewEmailService(config config.StashSphereMailConfig) EmailService {
	if config.Backend == "stdout" {
		return StdoutEmailService{}
	} else {
		return SMTPEmailService{config}
	}
}

func (h SMTPEmailService) Deliver(identifier string, subject string, body string) error {
	config := h.config
	auth := sasl.NewPlainClient("", config.User, config.Password)
	to := []string{identifier}
	msg := strings.NewReader(
		fmt.Sprintf(
			"From: %s\r\n"+
				"To: %s\r\n"+
				"Subject: %s\r\n"+
				"\r\n"+
				"%s\r\n", config.FromAddr, identifier, subject, body),
	)
	err := smtp.SendMail(
		fmt.Sprintf("%s:%d", config.Host, config.Port),
		auth,
		config.FromAddr,
		to,
		msg)
	return err
}

type StdoutEmailService struct {
}

func (h StdoutEmailService) Deliver(identifier string, subject string, body string) error {
	fmt.Printf("To: %s\n", identifier)
	fmt.Printf("Subject: %s\n\n", subject)
	fmt.Printf("Body: \n%s\n", body)
	return nil
}

type TestEmail struct {
	To      string
	Subject string
	Body    string
}

type TestEmailService struct {
	mails []TestEmail
}

func (h TestEmailService) Deliver(identifier string, subject string, body string) error {
	newMail := TestEmail{
		To:      identifier,
		Subject: subject,
		Body:    body,
	}
	h.mails = append(h.mails, newMail)
	return nil
}

func (h TestEmailService) Clear() {
	h.mails = []TestEmail{}
}
