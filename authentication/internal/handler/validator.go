package handler

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"github.com/nimbo1999/financeiro/authentication/internal/utils"
)

var (
	// Regular expressions for validation
	numericRegex = regexp.MustCompile(`^\d+$`)
)

// ValidateAndDecodeRequest validates and decodes JSON request
func ValidateAndDecodeRequest(r *http.Request, dst interface{}) []ValidationError {
	var errors []ValidationError

	// Decode JSON
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		errors = append(errors, ValidationError{
			Field:   "body",
			Message: "Invalid JSON format",
		})
		return errors
	}

	// Validate based on type
	switch v := dst.(type) {
	case *RequestCodeRequest:
		errors = append(errors, validateRequestCodeRequest(v)...)
	case *VerifyCodeRequest:
		errors = append(errors, validateVerifyCodeRequest(v)...)
	case *RefreshTokenRequest:
		errors = append(errors, validateRefreshTokenRequest(v)...)
	}

	return errors
}

func validateRequestCodeRequest(req *RequestCodeRequest) []ValidationError {
	var errors []ValidationError

	// Validate email
	if req.Email == "" {
		errors = append(errors, ValidationError{
			Field:   "email",
			Message: "Email is required",
		})
	} else {
		// Normalize email
		req.Email = strings.ToLower(strings.TrimSpace(req.Email))

		if !utils.IsValidEmail(req.Email) {
			errors = append(errors, ValidationError{
				Field:   "email",
				Message: "Email format is invalid",
			})
		}
	}

	return errors
}

func validateVerifyCodeRequest(req *VerifyCodeRequest) []ValidationError {
	var errors []ValidationError

	// Validate email
	if req.Email == "" {
		errors = append(errors, ValidationError{
			Field:   "email",
			Message: "Email is required",
		})
	} else {
		// Normalize email
		req.Email = strings.ToLower(strings.TrimSpace(req.Email))

		if !utils.IsValidEmail(req.Email) {
			errors = append(errors, ValidationError{
				Field:   "email",
				Message: "Email format is invalid",
			})
		}
	}

	// Validate code
	if req.Code == "" {
		errors = append(errors, ValidationError{
			Field:   "code",
			Message: "Authentication code is required",
		})
	} else {
		req.Code = strings.TrimSpace(req.Code)

		if len(req.Code) != 6 {
			errors = append(errors, ValidationError{
				Field:   "code",
				Message: "Authentication code must be exactly 6 digits",
			})
		} else if !numericRegex.MatchString(req.Code) {
			errors = append(errors, ValidationError{
				Field:   "code",
				Message: "Authentication code must contain only numbers",
			})
		}
	}

	return errors
}

func validateRefreshTokenRequest(req *RefreshTokenRequest) []ValidationError {
	var errors []ValidationError

	// Validate refresh token
	if req.RefreshToken == "" {
		errors = append(errors, ValidationError{
			Field:   "refresh_token",
			Message: "Refresh token is required",
		})
	} else {
		req.RefreshToken = strings.TrimSpace(req.RefreshToken)

		if len(req.RefreshToken) < 10 {
			errors = append(errors, ValidationError{
				Field:   "refresh_token",
				Message: "Refresh token format is invalid",
			})
		}
	}

	return errors
}

// ValidateRequest is a helper function that validates a request and writes error response if invalid
func ValidateRequest(w http.ResponseWriter, r *http.Request, dst interface{}) bool {
	errors := ValidateAndDecodeRequest(r, dst)
	if len(errors) > 0 {
		writeErrorResponse(w, http.StatusBadRequest, ValidationErrorResponse(errors))
		return false
	}
	return true
}