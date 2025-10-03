package services

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"gopkg.in/gomail.v2"
)

type EmailService interface {
	SendWelcomeEmail(ctx context.Context, to, name string) error
	SendOTPEmail(ctx context.Context, to, code string, expiresAt time.Time) error
}

type emailService struct {
	smtpHost     string
	smtpPort     int
	smtpUsername string
	smtpPassword string
	smtpFrom     string
	templateSvc  TemplateService
}

func NewEmailService(smtpHost string, smtpPort int, smtpUsername, smtpPassword, smtpFrom string, templateSvc TemplateService) EmailService {
	return &emailService{
		smtpHost:     smtpHost,
		smtpPort:     smtpPort,
		smtpUsername: smtpUsername,
		smtpPassword: smtpPassword,
		smtpFrom:     smtpFrom,
		templateSvc:  templateSvc,
	}
}

func (s *emailService) SendWelcomeEmail(ctx context.Context, to, name string) error {
	body, err := s.templateSvc.RenderWelcomeEmail(name)
	if err != nil {
		return fmt.Errorf("failed to render welcome email: %w", err)
	}

	return s.sendEmail(to, "Welcome to Financeiro!", body)
}

func (s *emailService) SendOTPEmail(ctx context.Context, to, code string, expiresAt time.Time) error {
	body, err := s.templateSvc.RenderOTPEmail(code, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to render OTP email: %w", err)
	}

	return s.sendEmail(to, "Financeiro Login Code", body)
}

func (s *emailService) sendEmail(to, subject, htmlBody string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.smtpFrom)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", htmlBody)

	d := gomail.NewDialer(s.smtpHost, s.smtpPort, s.smtpUsername, s.smtpPassword)
	// Enforce TLS 1.2 or higher
	d.TLSConfig = &tls.Config{
		ServerName:         s.smtpHost,
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: false,
	}

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email to %s: %w", to, err)
	}

	return nil
}
