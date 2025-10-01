package services

import (
	"context"
	"time"
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
	// Will be implemented in later steps
	return nil
}

func (s *emailService) SendOTPEmail(ctx context.Context, to, code string, expiresAt time.Time) error {
	// Will be implemented in later steps
	return nil
}
