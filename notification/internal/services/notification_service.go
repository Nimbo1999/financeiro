package services

import (
	"context"

	"notification/internal/models"
	"notification/internal/repository"
)

type NotificationService interface {
	SendWelcomeEmail(ctx context.Context, event *models.WelcomeEmailEvent) error
	SendOTPEmail(ctx context.Context, event *models.OTPEmailEvent) error
}

type notificationService struct {
	emailService EmailService
	repository   repository.NotificationRepository
}

func NewNotificationService(emailService EmailService, repository repository.NotificationRepository) NotificationService {
	return &notificationService{
		emailService: emailService,
		repository:   repository,
	}
}

func (s *notificationService) SendWelcomeEmail(ctx context.Context, event *models.WelcomeEmailEvent) error {
	// Will be implemented in later steps
	return nil
}

func (s *notificationService) SendOTPEmail(ctx context.Context, event *models.OTPEmailEvent) error {
	// Will be implemented in later steps
	return nil
}
