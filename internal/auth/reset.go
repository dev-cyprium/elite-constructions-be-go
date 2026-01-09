package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

const (
	// ResetTokenLength is the length of the reset token in bytes
	ResetTokenLength = 32
	// ResetTokenExpiry is how long a reset token is valid
	ResetTokenExpiry = 1 * time.Hour
)

// GenerateResetToken generates a cryptographically secure reset token
func GenerateResetToken() (string, error) {
	token := make([]byte, ResetTokenLength)
	if _, err := rand.Read(token); err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	return hex.EncodeToString(token), nil
}

// HashResetToken hashes a reset token using SHA256
func HashResetToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// VerifyResetToken verifies a reset token against its hash
func VerifyResetToken(token, hash string) bool {
	computedHash := HashResetToken(token)
	return computedHash == hash
}

// IsResetTokenExpired checks if a reset token has expired
func IsResetTokenExpired(expiresAt time.Time) bool {
	return time.Now().After(expiresAt)
}
