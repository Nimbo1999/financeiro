package utils

import "strings"

// IsValidEmail performs basic email validation
func IsValidEmail(email string) bool {
	if len(email) < 5 { // Minimum: a@b.c
		return false
	}

	// Check for spaces
	if strings.Count(strings.TrimSpace(email), " ") > 0 {
		return false
	}

	// Check for exactly one @ symbol
	atCount := strings.Count(email, "@")
	if atCount != 1 {
		return false
	}

	atIndex := strings.Index(email, "@")
	if atIndex <= 0 || atIndex == len(email)-1 {
		return false
	}

	dotIndex := strings.LastIndex(email, ".")
	if dotIndex <= atIndex+1 || dotIndex == len(email)-1 {
		return false
	}

	return true
}
