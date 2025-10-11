package handler

import (
	"log"

	"github.com/nimbo1999/financeiro/users/internal/models"
	"github.com/nimbo1999/financeiro/users/internal/services"
	userpb "github.com/nimbo1999/financeiro/users/pkg/grpc/users/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ModelToProto converts a domain User model to protobuf User message
func ModelToProto(user *models.User) *userpb.User {
	if user == nil {
		return nil
	}

	protoUser := &userpb.User{
		Id:       user.ID,
		Email:    user.Email,
		FullName: user.FullName,
		IsAdmin:  user.IsAdmin,
	}

	// Handle timestamps with proper null checking
	if !user.CreatedAt.IsZero() {
		protoUser.CreatedAt = timestamppb.New(user.CreatedAt)
	}

	if !user.UpdatedAt.IsZero() {
		protoUser.UpdatedAt = timestamppb.New(user.UpdatedAt)
	}

	return protoUser
}

// ProtoToModel converts a protobuf User message to domain User model
func ProtoToModel(protoUser *userpb.User) *models.User {
	if protoUser == nil {
		return nil
	}

	user := &models.User{
		ID:       protoUser.Id,
		Email:    protoUser.Email,
		FullName: protoUser.FullName,
		IsAdmin:  protoUser.IsAdmin,
	}

	// Handle timestamp conversions with null checking
	if protoUser.CreatedAt != nil && protoUser.CreatedAt.IsValid() {
		user.CreatedAt = protoUser.CreatedAt.AsTime()
	}

	if protoUser.UpdatedAt != nil && protoUser.UpdatedAt.IsValid() {
		user.UpdatedAt = protoUser.UpdatedAt.AsTime()
	}

	return user
}

// MapServiceErrorToGRPC maps domain service errors to appropriate gRPC status codes
func MapServiceErrorToGRPC(err error, operation string) error {
	if err == nil {
		return nil
	}

	// Log the original error for debugging
	log.Printf("gRPC %s error: %v", operation, err)

	switch err {
	case services.ErrUserNotFound:
		return status.Error(codes.NotFound, "user not found")
	case services.ErrInvalidUserData:
		return status.Error(codes.InvalidArgument, "invalid user data")
	case services.ErrEmailIsRequired:
		return status.Error(codes.InvalidArgument, "email is required")
	case services.ErrFullNameIsRequired:
		return status.Error(codes.InvalidArgument, "full name is required")
	default:
		// Check for database-related errors
		if isDatabaseConnectionError(err) {
			log.Printf("Database connection error in %s: %v", operation, err)
			return status.Error(codes.Unavailable, "service temporarily unavailable")
		}

		if isDuplicateKeyError(err) {
			log.Printf("Duplicate key error in %s: %v", operation, err)
			return status.Error(codes.AlreadyExists, "resource already exists")
		}

		// For unknown errors, log them and return a generic internal error
		log.Printf("Unmapped service error in %s: %v", operation, err)
		return status.Error(codes.Internal, "internal server error")
	}
}

// isDatabaseConnectionError checks if the error is related to database connectivity
func isDatabaseConnectionError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	// Common database connection error patterns
	connectionErrors := []string{
		"connection refused",
		"connection reset",
		"connection timeout",
		"database is closed",
		"driver: bad connection",
		"no rows in result set", // GORM specific
	}

	for _, connErr := range connectionErrors {
		if containsSubstring(errStr, connErr) {
			return true
		}
	}
	return false
}

// isDuplicateKeyError checks if the error is related to duplicate key constraints
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	// Common duplicate key error patterns
	duplicateErrors := []string{
		"duplicate key",
		"unique constraint",
		"UNIQUE constraint failed",
		"duplicate entry",
	}

	for _, dupErr := range duplicateErrors {
		if containsSubstring(errStr, dupErr) {
			return true
		}
	}
	return false
}

// containsSubstring checks if a string contains a substring (case-insensitive)
func containsSubstring(str, substr string) bool {
	if len(substr) > len(str) {
		return false
	}

	// Simple case-insensitive substring check
	strLower := toLower(str)
	substrLower := toLower(substr)

	for i := 0; i <= len(strLower)-len(substrLower); i++ {
		if strLower[i:i+len(substrLower)] == substrLower {
			return true
		}
	}
	return false
}

// toLower converts a string to lowercase (simple implementation)
func toLower(s string) string {
	result := make([]rune, len(s))
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			result[i] = r + ('a' - 'A')
		} else {
			result[i] = r
		}
	}
	return string(result)
}

// ValidateGetUserByEmailRequest validates the GetUserByEmail request
func ValidateGetUserByEmailRequest(req *userpb.GetUserByEmailRequest) error {
	if req == nil {
		return status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	if req.Email == "" {
		return status.Error(codes.InvalidArgument, "email is required")
	}

	// Basic email format validation (simple check)
	if len(req.Email) < 3 || !containsAt(req.Email) {
		return status.Error(codes.InvalidArgument, "invalid email format")
	}

	return nil
}

// ValidateGetUserByIdRequest validates the GetUserById request
func ValidateGetUserByIdRequest(req *userpb.GetUserByIdRequest) error {
	if req == nil {
		return status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	if req.Id == "" {
		return status.Error(codes.InvalidArgument, "id is required")
	}

	return nil
}

// containsAt is a simple helper to check if string contains @ symbol
func containsAt(email string) bool {
	for _, char := range email {
		if char == '@' {
			return true
		}
	}
	return false
}
