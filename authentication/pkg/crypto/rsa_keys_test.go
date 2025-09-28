package crypto

import (
	"crypto/rsa"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type RSAKeysTestSuite struct {
	suite.Suite
}

func TestRSAKeysTestSuite(t *testing.T) {
	suite.Run(t, new(RSAKeysTestSuite))
}

func (suite *RSAKeysTestSuite) TestGenerateRSAKeyPair_Success() {
	privateKeyPEM, publicKeyPEM, err := GenerateRSAKeyPair(2048)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), privateKeyPEM)
	assert.NotNil(suite.T(), publicKeyPEM)
	assert.Contains(suite.T(), string(privateKeyPEM), "-----BEGIN RSA PRIVATE KEY-----")
	assert.Contains(suite.T(), string(privateKeyPEM), "-----END RSA PRIVATE KEY-----")
	assert.Contains(suite.T(), string(publicKeyPEM), "-----BEGIN PUBLIC KEY-----")
	assert.Contains(suite.T(), string(publicKeyPEM), "-----END PUBLIC KEY-----")
}

func (suite *RSAKeysTestSuite) TestGenerateRSAKeyPair_InvalidKeySize() {
	privateKeyPEM, publicKeyPEM, err := GenerateRSAKeyPair(1024)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), privateKeyPEM)
	assert.Nil(suite.T(), publicKeyPEM)
	assert.Contains(suite.T(), err.Error(), "key size must be at least 2048 bits")
}

func (suite *RSAKeysTestSuite) TestParseRSAPrivateKey_Success() {
	// Generate a key pair first
	privateKeyPEM, _, err := GenerateRSAKeyPair(2048)
	suite.Require().NoError(err)

	// Parse the private key
	privateKey, err := ParseRSAPrivateKey(privateKeyPEM)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), privateKey)
	assert.IsType(suite.T(), &rsa.PrivateKey{}, privateKey)
	assert.Equal(suite.T(), 2048, privateKey.Size()*8) // Size() returns bytes, so multiply by 8 for bits
}

func (suite *RSAKeysTestSuite) TestParseRSAPrivateKey_InvalidPEM() {
	invalidPEM := []byte("invalid-pem-data")

	privateKey, err := ParseRSAPrivateKey(invalidPEM)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), privateKey)
	assert.Contains(suite.T(), err.Error(), "failed to decode PEM block")
}

func (suite *RSAKeysTestSuite) TestParseRSAPrivateKey_WrongPEMType() {
	wrongTypePEM := []byte(`-----BEGIN CERTIFICATE-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA
-----END CERTIFICATE-----`)

	privateKey, err := ParseRSAPrivateKey(wrongTypePEM)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), privateKey)
	assert.Contains(suite.T(), err.Error(), "invalid PEM block type")
}

func (suite *RSAKeysTestSuite) TestParseRSAPublicKey_Success() {
	// Generate a key pair first
	_, publicKeyPEM, err := GenerateRSAKeyPair(2048)
	suite.Require().NoError(err)

	// Parse the public key
	publicKey, err := ParseRSAPublicKey(publicKeyPEM)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), publicKey)
	assert.IsType(suite.T(), &rsa.PublicKey{}, publicKey)
	assert.Equal(suite.T(), 2048, publicKey.Size()*8) // Size() returns bytes, so multiply by 8 for bits
}

func (suite *RSAKeysTestSuite) TestParseRSAPublicKey_InvalidPEM() {
	invalidPEM := []byte("invalid-pem-data")

	publicKey, err := ParseRSAPublicKey(invalidPEM)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), publicKey)
	assert.Contains(suite.T(), err.Error(), "failed to decode PEM block")
}

func (suite *RSAKeysTestSuite) TestParseRSAPublicKey_WrongPEMType() {
	wrongTypePEM := []byte(`-----BEGIN CERTIFICATE-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA
-----END CERTIFICATE-----`)

	publicKey, err := ParseRSAPublicKey(wrongTypePEM)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), publicKey)
	assert.Contains(suite.T(), err.Error(), "invalid PEM block type")
}

func (suite *RSAKeysTestSuite) TestKeyPairCompatibility() {
	// Generate a key pair
	privateKeyPEM, publicKeyPEM, err := GenerateRSAKeyPair(2048)
	suite.Require().NoError(err)

	// Parse both keys
	privateKey, err := ParseRSAPrivateKey(privateKeyPEM)
	suite.Require().NoError(err)

	publicKey, err := ParseRSAPublicKey(publicKeyPEM)
	suite.Require().NoError(err)

	// Verify they are compatible by comparing the public key from private key
	assert.Equal(suite.T(), publicKey.N, privateKey.PublicKey.N)
	assert.Equal(suite.T(), publicKey.E, privateKey.PublicKey.E)
}

func (suite *RSAKeysTestSuite) TestGenerateMultipleKeyPairs() {
	// Generate multiple key pairs and ensure they are different
	privateKeyPEM1, publicKeyPEM1, err := GenerateRSAKeyPair(2048)
	suite.Require().NoError(err)

	privateKeyPEM2, publicKeyPEM2, err := GenerateRSAKeyPair(2048)
	suite.Require().NoError(err)

	// Keys should be different
	assert.NotEqual(suite.T(), string(privateKeyPEM1), string(privateKeyPEM2))
	assert.NotEqual(suite.T(), string(publicKeyPEM1), string(publicKeyPEM2))

	// But both should be valid
	privateKey1, err := ParseRSAPrivateKey(privateKeyPEM1)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), privateKey1)

	privateKey2, err := ParseRSAPrivateKey(privateKeyPEM2)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), privateKey2)

	publicKey1, err := ParseRSAPublicKey(publicKeyPEM1)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), publicKey1)

	publicKey2, err := ParseRSAPublicKey(publicKeyPEM2)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), publicKey2)
}

// Table-driven tests
func (suite *RSAKeysTestSuite) TestGenerateRSAKeyPair_TableDriven() {
	testCases := []struct {
		name        string
		bits        int
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid 2048 bits",
			bits:        2048,
			expectError: false,
		},
		{
			name:        "Valid 4096 bits",
			bits:        4096,
			expectError: false,
		},
		{
			name:        "Invalid 1024 bits",
			bits:        1024,
			expectError: true,
			errorMsg:    "key size must be at least 2048 bits",
		},
		{
			name:        "Invalid 512 bits",
			bits:        512,
			expectError: true,
			errorMsg:    "key size must be at least 2048 bits",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			privateKeyPEM, publicKeyPEM, err := GenerateRSAKeyPair(tc.bits)

			if tc.expectError {
				assert.Error(suite.T(), err)
				assert.Nil(suite.T(), privateKeyPEM)
				assert.Nil(suite.T(), publicKeyPEM)
				if tc.errorMsg != "" {
					assert.Contains(suite.T(), err.Error(), tc.errorMsg)
				}
			} else {
				assert.NoError(suite.T(), err)
				assert.NotNil(suite.T(), privateKeyPEM)
				assert.NotNil(suite.T(), publicKeyPEM)

				// Verify the keys can be parsed
				privateKey, err := ParseRSAPrivateKey(privateKeyPEM)
				assert.NoError(suite.T(), err)
				assert.Equal(suite.T(), tc.bits, privateKey.Size()*8)

				publicKey, err := ParseRSAPublicKey(publicKeyPEM)
				assert.NoError(suite.T(), err)
				assert.Equal(suite.T(), tc.bits, publicKey.Size()*8)
			}
		})
	}
}

// Benchmark tests
func BenchmarkGenerateRSAKeyPair2048(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _, _ = GenerateRSAKeyPair(2048)
	}
}

func BenchmarkGenerateRSAKeyPair4096(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _, _ = GenerateRSAKeyPair(4096)
	}
}

func BenchmarkParseRSAPrivateKey(b *testing.B) {
	// Generate a key pair for benchmarking
	privateKeyPEM, _, err := GenerateRSAKeyPair(2048)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseRSAPrivateKey(privateKeyPEM)
	}
}

func BenchmarkParseRSAPublicKey(b *testing.B) {
	// Generate a key pair for benchmarking
	_, publicKeyPEM, err := GenerateRSAKeyPair(2048)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseRSAPublicKey(publicKeyPEM)
	}
}