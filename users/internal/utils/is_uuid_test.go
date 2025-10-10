package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type IsUUIDTestSuite struct {
	suite.Suite
}

func (suite *IsUUIDTestSuite) TestIsUUID_ValidUUIDs() {
	validUUIDs := []string{
		// UUID v1
		"c9bf9e58-0aa4-11ef-b939-0242ac120002",
		"d4f5e6a7-0aa4-11ef-b939-0242ac120002",

		// UUID v4 (most common)
		"550e8400-e29b-41d4-a716-446655440000",
		"6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		"f47ac10b-58cc-4372-a567-0e02b2c3d479",
		"123e4567-e89b-12d3-a456-426614174000",
		"a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",

		// UUID v5
		"74738ff5-5367-5958-9aee-98fffdcd1876",

		// Edge cases with all zeros
		"00000000-0000-0000-0000-000000000000",

		// Edge cases with all f's
		"ffffffff-ffff-ffff-ffff-ffffffffffff",

		// Uppercase
		"550E8400-E29B-41D4-A716-446655440000",
		"F47AC10B-58CC-4372-A567-0E02B2C3D479",

		// Mixed case
		"550e8400-E29b-41D4-a716-446655440000",
	}

	for _, uuid := range validUUIDs {
		result := IsUUID(uuid)
		assert.True(suite.T(), result, "Expected %s to be a valid UUID", uuid)
	}
}

func (suite *IsUUIDTestSuite) TestIsUUID_EmptyString() {
	result := IsUUID("")
	assert.False(suite.T(), result)
}

func (suite *IsUUIDTestSuite) TestIsUUID_InvalidLength() {
	invalidUUIDs := []string{
		"550e8400-e29b-41d4-a716",                    // Too short
		"550e8400-e29b-41d4-a716-446655440000-extra", // Too long
		"550e8400",                                    // Way too short
		"550e8400-e29b-41d4-a716-446655440000000",    // Extra character
	}

	for _, uuid := range invalidUUIDs {
		result := IsUUID(uuid)
		assert.False(suite.T(), result, "Expected %s to be invalid (wrong length)", uuid)
	}
}

func (suite *IsUUIDTestSuite) TestIsUUID_MissingHyphens() {
	// Note: google/uuid library accepts UUIDs without hyphens (32 hex chars)
	// So we test with strings that have missing hyphens in the standard format
	invalidUUIDs := []string{
		"550e8400-e29b41d4-a716-446655440000",        // Missing one hyphen
		"550e8400-e29b-41d4a716-446655440000",        // Missing one hyphen
		"550e8400-e29b-41d4-a716446655440000",        // Missing one hyphen
	}

	for _, uuid := range invalidUUIDs {
		result := IsUUID(uuid)
		assert.False(suite.T(), result, "Expected %s to be invalid (missing hyphens)", uuid)
	}
}

func (suite *IsUUIDTestSuite) TestIsUUID_WrongHyphenPositions() {
	invalidUUIDs := []string{
		"550e840-0e29b-41d4-a716-446655440000",  // Wrong hyphen position
		"550e8400-e29-b41d4-a716-446655440000",  // Wrong hyphen position
		"550e8400-e29b-41d-4a716-446655440000",  // Wrong hyphen position
		"550e8400-e29b-41d4-a71-6446655440000",  // Wrong hyphen position
	}

	for _, uuid := range invalidUUIDs {
		result := IsUUID(uuid)
		assert.False(suite.T(), result, "Expected %s to be invalid (wrong hyphen positions)", uuid)
	}
}

func (suite *IsUUIDTestSuite) TestIsUUID_InvalidCharacters() {
	invalidUUIDs := []string{
		"550e8400-e29b-41d4-a716-44665544000g",  // 'g' is not hex
		"550e8400-e29b-41d4-a716-44665544000z",  // 'z' is not hex
		"550e8400-e29b-41d4-a716-44665544000!",  // Special character
		"550e8400-e29b-41d4-a716-44665544000 ",  // Space
		"550e8400-e29b-41d4-a716-44665544000@",  // @ symbol
		"550e8400-e29b-41d4-a716-44665544000#",  // # symbol
	}

	for _, uuid := range invalidUUIDs {
		result := IsUUID(uuid)
		assert.False(suite.T(), result, "Expected %s to be invalid (invalid characters)", uuid)
	}
}

func (suite *IsUUIDTestSuite) TestIsUUID_TooManyHyphens() {
	invalidUUIDs := []string{
		"550e-8400-e29b-41d4-a716-446655440000",  // 5 hyphens instead of 4
		"550e8400--e29b-41d4-a716-446655440000",  // Double hyphen
		"550e8400-e29b--41d4-a716-446655440000",  // Double hyphen
	}

	for _, uuid := range invalidUUIDs {
		result := IsUUID(uuid)
		assert.False(suite.T(), result, "Expected %s to be invalid (too many hyphens)", uuid)
	}
}

func (suite *IsUUIDTestSuite) TestIsUUID_NotUUIDFormat() {
	invalidStrings := []string{
		"not-a-uuid",
		"this-is-not-a-valid-uuid-string",
		"550e8400-e29b-41d4-a716",                // Incomplete UUID
		"random-string",
		"email@example.com",
		"123456",
	}

	for _, str := range invalidStrings {
		result := IsUUID(str)
		assert.False(suite.T(), result, "Expected %s to be invalid (not UUID format)", str)
	}
}

func (suite *IsUUIDTestSuite) TestIsUUID_WithWhitespace() {
	invalidUUIDs := []string{
		"550e8400 -e29b-41d4-a716-446655440000",  // Space in middle
		"550e8400-e29b-41d4-a716-446655440000\n", // Newline
		"550e8400-e29b-41d4-a716-446655440000\t", // Tab
	}

	for _, uuid := range invalidUUIDs {
		result := IsUUID(uuid)
		assert.False(suite.T(), result, "Expected %s to be invalid (contains whitespace)", uuid)
	}
}

func (suite *IsUUIDTestSuite) TestIsUUID_EdgeCases() {
	testCases := []struct {
		input    string
		expected bool
		reason   string
	}{
		{"00000000-0000-0000-0000-000000000000", true, "nil UUID (all zeros)"},
		{"FFFFFFFF-FFFF-FFFF-FFFF-FFFFFFFFFFFF", true, "all F's uppercase"},
		{"ffffffff-ffff-ffff-ffff-ffffffffffff", true, "all f's lowercase"},
		// Note: google/uuid library accepts these formats
		{"{550e8400-e29b-41d4-a716-446655440000}", true, "wrapped in braces (accepted by google/uuid)"},
		{"urn:uuid:550e8400-e29b-41d4-a716-446655440000", true, "URN format (accepted by google/uuid)"},
		{"550e8400e29b41d4a716446655440000", true, "no hyphens (accepted by google/uuid)"},
		{"", false, "empty string"},
		{"nil", false, "literal 'nil'"},
		{"null", false, "literal 'null'"},
	}

	for _, tc := range testCases {
		result := IsUUID(tc.input)
		assert.Equal(suite.T(), tc.expected, result, "Expected %s to be %v (%s)", tc.input, tc.expected, tc.reason)
	}
}

func (suite *IsUUIDTestSuite) TestIsUUID_AllVersions() {
	// Test different UUID versions (all should be valid)
	uuidVersions := []string{
		// Version 1 (time-based)
		"c9bf9e58-0aa4-11ef-b939-0242ac120002",

		// Version 2 (DCE Security)
		"000003e8-0aa5-21ef-b200-325096b39f47",

		// Version 3 (MD5 hash)
		"a3bb189e-8bf9-3888-9912-ace4e6543002",

		// Version 4 (random)
		"f47ac10b-58cc-4372-a567-0e02b2c3d479",

		// Version 5 (SHA-1 hash)
		"74738ff5-5367-5958-9aee-98fffdcd1876",
	}

	for _, uuid := range uuidVersions {
		result := IsUUID(uuid)
		assert.True(suite.T(), result, "Expected %s to be a valid UUID", uuid)
	}
}

func TestIsUUIDTestSuite(t *testing.T) {
	suite.Run(t, new(IsUUIDTestSuite))
}
