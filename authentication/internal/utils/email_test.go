package utils

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type EmailTestSuite struct {
	suite.Suite
}

func TestEmailTestSuite(t *testing.T) {
	suite.Run(t, new(EmailTestSuite))
}

func (suite *EmailTestSuite) Test_validEmailTest() {
	validEmails := []string{
		"test@example.com",
		"user.name@domain.co",
		"first.last@sub.domain.com",
	}

	for _, email := range validEmails {
		suite.Assert().True(IsValidEmail(email), email)
	}
}

func (suite *EmailTestSuite) Test_invalidEmailTest() {
	invalidEmails := []string{
		"plainaddress",
		"@no-local-part.com",
		"Outlook Contact <outlook@example.com>",
	}

	for _, email := range invalidEmails {
		suite.Assert().False(IsValidEmail(email), email)
	}
}

func (suite *EmailTestSuite) Test_should_have_at_least_5_characters() {
	suite.Assert().False(IsValidEmail("a@b"))
	suite.Assert().True(IsValidEmail("a@b.c"))
}

func (suite *EmailTestSuite) Test_should_have_the_at_symbol() {
	suite.Assert().False(IsValidEmail("plainaddress"))
	suite.Assert().False(IsValidEmail("missingatsymbol.com"))
	suite.Assert().False(IsValidEmail("two@@atsymbols.com"))
	suite.Assert().True(IsValidEmail("valid@example.com"))
}
