package utils

func IsEmail(s string) bool {
	if s == "" {
		return false
	}

	// Check for basic structure: local@domain
	atIndex := -1
	atCount := 0
	for i, char := range s {
		if char == '@' {
			atIndex = i
			atCount++
		}
	}

	// Email must have exactly one '@' symbol
	if atCount != 1 || atIndex == 0 || atIndex == len(s)-1 {
		return false
	}

	local := s[:atIndex]
	domain := s[atIndex+1:]

	// Validate local part (before @)
	if len(local) == 0 || len(local) > 64 {
		return false
	}

	// Local part cannot start or end with a dot
	if local[0] == '.' || local[len(local)-1] == '.' {
		return false
	}

	// Check for consecutive dots in local part
	for i := 0; i < len(local)-1; i++ {
		if local[i] == '.' && local[i+1] == '.' {
			return false
		}
	}

	// Validate characters in local part
	for _, char := range local {
		isValid := (char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '.' || char == '_' || char == '-' || char == '+'
		if !isValid {
			return false
		}
	}

	// Validate domain part (after @)
	if len(domain) == 0 || len(domain) > 255 {
		return false
	}

	// Domain must contain at least one dot
	dotIndex := -1
	for i, char := range domain {
		if char == '.' {
			dotIndex = i
			break
		}
	}

	if dotIndex == -1 || dotIndex == 0 || dotIndex == len(domain)-1 {
		return false
	}

	// Domain cannot start or end with a dot or hyphen
	if domain[0] == '.' || domain[len(domain)-1] == '.' ||
		domain[0] == '-' || domain[len(domain)-1] == '-' {
		return false
	}

	// Check for consecutive dots in domain
	for i := 0; i < len(domain)-1; i++ {
		if domain[i] == '.' && domain[i+1] == '.' {
			return false
		}
	}

	// Validate characters in domain part
	for _, char := range domain {
		isValid := (char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '.' || char == '-'
		if !isValid {
			return false
		}
	}

	// Validate TLD (top-level domain) - must be at least 2 characters
	lastDotIndex := -1
	for i := len(domain) - 1; i >= 0; i-- {
		if domain[i] == '.' {
			lastDotIndex = i
			break
		}
	}

	if lastDotIndex != -1 {
		tld := domain[lastDotIndex+1:]
		if len(tld) < 2 {
			return false
		}
		// TLD should only contain letters
		for _, char := range tld {
			if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z')) {
				return false
			}
		}
	}

	return true
}
