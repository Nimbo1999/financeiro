package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"notification/internal/models"
	"notification/internal/repository"

	"github.com/google/uuid"
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
	// 1. Marshal event data
	eventData, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	// 2. Create notification record (status: pending)
	notification := &models.Notification{
		ID:                  uuid.New().String(),
		UserID:              event.ID, // Using event ID as user ID for now
		Email:               event.Data.UserEmail,
		NotificationSubject: "notification.welcome",
		NotificationType:    event.Type,
		Status:              "pending",
		Subject:             "Welcome to Financeiro!",
		RetryCount:          0,
		EventData:           eventData,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	if err := s.repository.Create(ctx, notification); err != nil {
		return fmt.Errorf("failed to create notification record: %w", err)
	}

	// 3. Send email
	err = s.emailService.SendWelcomeEmail(ctx, event.Data.UserEmail, event.Data.Name)

	// 4. Update notification status
	if err != nil {
		if updateErr := s.repository.MarkAsFailed(ctx, notification.ID, err.Error()); updateErr != nil {
			return fmt.Errorf("failed to send email and update status: %w, %w", err, updateErr)
		}
		return fmt.Errorf("failed to send welcome email: %w", err)
	}

	if err := s.repository.MarkAsSent(ctx, notification.ID); err != nil {
		return fmt.Errorf("failed to update notification as sent: %w", err)
	}

	return nil
}

func (s *notificationService) SendOTPEmail(ctx context.Context, event *models.OTPEmailEvent) error {
	// 1. Marshal event data
	eventData, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	// 2. Create notification record (status: pending)
	notification := &models.Notification{
		ID:                  uuid.New().String(),
		UserID:              event.Data.CodeID, // Using code ID as user ID for now
		Email:               event.Data.UserEmail,
		NotificationSubject: "notification.otp",
		NotificationType:    event.Type,
		Status:              "pending",
		Subject:             "Your Login Code",
		RetryCount:          0,
		EventData:           eventData,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	if err := s.repository.Create(ctx, notification); err != nil {
		return fmt.Errorf("failed to create notification record: %w", err)
	}

	// 3. Send email
	err = s.emailService.SendOTPEmail(ctx, event.Data.UserEmail, event.Data.AuthCode, event.Data.ExpiresAt)

	// 4. Update notification status
	if err != nil {
		if updateErr := s.repository.MarkAsFailed(ctx, notification.ID, err.Error()); updateErr != nil {
			return fmt.Errorf("failed to send email and update status: %w, %w", err, updateErr)
		}
		return fmt.Errorf("failed to send OTP email: %w", err)
	}

	if err := s.repository.MarkAsSent(ctx, notification.ID); err != nil {
		return fmt.Errorf("failed to update notification as sent: %w", err)
	}

	return nil
}
