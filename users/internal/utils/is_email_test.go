package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type IsEmailTestSuite struct {
	suite.Suite
}

func (suite *IsEmailTestSuite) TestIsEmail_ValidEmails() {
	validEmails := []string{
		"user@example.com",
		"john.doe@example.com",
		"john_doe@example.com",
		"john-doe@example.com",
		"john+tag@example.com",
		"user123@example.com",
		"User@Example.COM",
		"test@sub.domain.com",
		"a@example.co.uk",
		"user.name+tag@example.co.uk",
		"test_user123@company-name.com",
		"a.b.c@domain.com",
		"user@test-domain.com",
		"1234567890@example.com",
		"user@123domain.com",
	}

	for _, email := range validEmails {
		result := IsEmail(email)
		assert.True(suite.T(), result, "Expected %s to be valid", email)
	}
}

func (suite *IsEmailTestSuite) TestIsEmail_EmptyString() {
	result := IsEmail("")
	assert.False(suite.T(), result)
}

func (suite *IsEmailTestSuite) TestIsEmail_MissingAtSymbol() {
	invalidEmails := []string{
		"userexample.com",
		"user.example.com",
		"userexamplecom",
	}

	for _, email := range invalidEmails {
		result := IsEmail(email)
		assert.False(suite.T(), result, "Expected %s to be invalid (missing @)", email)
	}
}

func (suite *IsEmailTestSuite) TestIsEmail_MultipleAtSymbols() {
	invalidEmails := []string{
		"user@@example.com",
		"user@domain@example.com",
		"@user@example.com",
	}

	for _, email := range invalidEmails {
		result := IsEmail(email)
		assert.False(suite.T(), result, "Expected %s to be invalid (multiple @)", email)
	}
}

func (suite *IsEmailTestSuite) TestIsEmail_AtSymbolAtStartOrEnd() {
	invalidEmails := []string{
		"@example.com",
		"user@",
		"@",
	}

	for _, email := range invalidEmails {
		result := IsEmail(email)
		assert.False(suite.T(), result, "Expected %s to be invalid (@ at start/end)", email)
	}
}

func (suite *IsEmailTestSuite) TestIsEmail_LocalPartTooLong() {
	// Local part longer than 64 characters
	longLocal := "thisIsAVeryLongEmailLocalPartThatExceedsTheSixtyFourCharacterLimit@example.com"
	result := IsEmail(longLocal)
	assert.False(suite.T(), result)
}

func (suite *IsEmailTestSuite) TestIsEmail_LocalPartStartsOrEndsWithDot() {
	invalidEmails := []string{
		".user@example.com",
		"user.@example.com",
		".user.@example.com",
	}

	for _, email := range invalidEmails {
		result := IsEmail(email)
		assert.False(suite.T(), result, "Expected %s to be invalid (dot at start/end of local)", email)
	}
}

func (suite *IsEmailTestSuite) TestIsEmail_ConsecutiveDotsInLocalPart() {
	invalidEmails := []string{
		"user..name@example.com",
		"user...name@example.com",
		"first..last@example.com",
	}

	for _, email := range invalidEmails {
		result := IsEmail(email)
		assert.False(suite.T(), result, "Expected %s to be invalid (consecutive dots)", email)
	}
}

func (suite *IsEmailTestSuite) TestIsEmail_InvalidCharactersInLocalPart() {
	invalidEmails := []string{
		"user#name@example.com",
		"user@name@example.com",
		"user name@example.com",
		"user!name@example.com",
		"user$name@example.com",
		"user%name@example.com",
		"user&name@example.com",
		"user*name@example.com",
		"user(name@example.com",
		"user)name@example.com",
	}

	for _, email := range invalidEmails {
		result := IsEmail(email)
		assert.False(suite.T(), result, "Expected %s to be invalid (invalid chars)", email)
	}
}

func (suite *IsEmailTestSuite) TestIsEmail_DomainTooLong() {
	// Domain longer than 255 characters
	// Create a domain that exceeds 255 characters
	longSubdomain := "verylongsubdomainname"
	longDomain := "user@"
	for len(longDomain[5:]) < 256 {
		longDomain += longSubdomain + "."
	}
	longDomain += "com"
	result := IsEmail(longDomain)
	assert.False(suite.T(), result, "Domain length: %d", len(longDomain[5:]))
}

func (suite *IsEmailTestSuite) TestIsEmail_MissingDotInDomain() {
	invalidEmails := []string{
		"user@domain",
		"user@localhost",
		"user@example",
	}

	for _, email := range invalidEmails {
		result := IsEmail(email)
		assert.False(suite.T(), result, "Expected %s to be invalid (missing dot in domain)", email)
	}
}

func (suite *IsEmailTestSuite) TestIsEmail_DotAtStartOrEndOfDomain() {
	invalidEmails := []string{
		"user@.example.com",
		"user@example.com.",
		"user@.example.com.",
	}

	for _, email := range invalidEmails {
		result := IsEmail(email)
		assert.False(suite.T(), result, "Expected %s to be invalid (dot at start/end of domain)", email)
	}
}

func (suite *IsEmailTestSuite) TestIsEmail_HyphenAtStartOrEndOfDomain() {
	invalidEmails := []string{
		"user@-example.com",
		"user@example.com-",
	}

	for _, email := range invalidEmails {
		result := IsEmail(email)
		assert.False(suite.T(), result, "Expected %s to be invalid (hyphen at start/end of domain)", email)
	}
}

func (suite *IsEmailTestSuite) TestIsEmail_ConsecutiveDotsInDomain() {
	invalidEmails := []string{
		"user@example..com",
		"user@sub..domain.com",
		"user@example...com",
	}

	for _, email := range invalidEmails {
		result := IsEmail(email)
		assert.False(suite.T(), result, "Expected %s to be invalid (consecutive dots in domain)", email)
	}
}

func (suite *IsEmailTestSuite) TestIsEmail_InvalidCharactersInDomain() {
	invalidEmails := []string{
		"user@example_domain.com",
		"user@example+domain.com",
		"user@example domain.com",
		"user@example#domain.com",
		"user@example@domain.com",
	}

	for _, email := range invalidEmails {
		result := IsEmail(email)
		assert.False(suite.T(), result, "Expected %s to be invalid (invalid chars in domain)", email)
	}
}

func (suite *IsEmailTestSuite) TestIsEmail_TLDTooShort() {
	invalidEmails := []string{
		"user@example.c",
		"user@domain.x",
	}

	for _, email := range invalidEmails {
		result := IsEmail(email)
		assert.False(suite.T(), result, "Expected %s to be invalid (TLD too short)", email)
	}
}

func (suite *IsEmailTestSuite) TestIsEmail_TLDWithNumbers() {
	invalidEmails := []string{
		"user@example.c0m",
		"user@example.co1",
		"user@example.123",
	}

	for _, email := range invalidEmails {
		result := IsEmail(email)
		assert.False(suite.T(), result, "Expected %s to be invalid (TLD with numbers)", email)
	}
}

func (suite *IsEmailTestSuite) TestIsEmail_TLDWithSpecialCharacters() {
	invalidEmails := []string{
		"user@example.c-m",
		"user@example.co_m",
	}

	for _, email := range invalidEmails {
		result := IsEmail(email)
		assert.False(suite.T(), result, "Expected %s to be invalid (TLD with special chars)", email)
	}
}

func (suite *IsEmailTestSuite) TestIsEmail_EdgeCases() {
	testCases := []struct {
		email    string
		expected bool
		reason   string
	}{
		{"a@b.co", true, "minimum valid email"},
		{"user@sub.sub.sub.example.com", true, "multiple subdomains"},
		{"123@456.com", true, "all numbers in local"},
		{"a1b2c3@example.com", true, "mixed alphanumeric"},
		{"user+tag+another@example.com", true, "multiple plus signs"},
		{"user_name_test@example.com", true, "multiple underscores"},
		{"user-name-test@example.com", true, "multiple hyphens"},
		{"@", false, "just @"},
		{"@@", false, "double @"},
		{".", false, "just dot"},
		{".@.", false, "dots and @"},
		{"user", false, "no @ or domain"},
		{"user@", false, "no domain"},
		{"@domain.com", false, "no local part"},
	}

	for _, tc := range testCases {
		result := IsEmail(tc.email)
		assert.Equal(suite.T(), tc.expected, result, "Expected %s to be %v (%s)", tc.email, tc.expected, tc.reason)
	}
}

func TestIsEmailTestSuite(t *testing.T) {
	suite.Run(t, new(IsEmailTestSuite))
}
