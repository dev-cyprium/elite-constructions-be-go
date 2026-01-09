package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	// Argon2id parameters (recommended values)
	memory      = 64 * 1024 // 64 MB
	iterations  = 3
	parallelism = 2         // uint8
	saltLength  = 16
	keyLength   = 32
)

// HashPassword hashes a password using Argon2id
func HashPassword(password string) (string, error) {
	// Generate a random salt
	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Hash the password
	hash := argon2.IDKey([]byte(password), salt, iterations, memory, uint8(parallelism), keyLength)

	// Encode salt and hash to base64
	saltBase64 := base64.RawStdEncoding.EncodeToString(salt)
	hashBase64 := base64.RawStdEncoding.EncodeToString(hash)

	// Return formatted string: $argon2id$v=19$m=65536,t=3,p=2$salt$hash
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, memory, iterations, parallelism, saltBase64, hashBase64), nil
}

// VerifyPassword verifies a password against an Argon2id hash
func VerifyPassword(password, encodedHash string) (bool, error) {
	// Parse the encoded hash
	// Format: $argon2id$v=19$m=65536,t=3,p=2$salt$hash
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 || parts[0] != "" || parts[1] != "argon2id" {
		return false, fmt.Errorf("invalid hash format: expected $argon2id$v=X$m=Y,t=Z,p=W$salt$hash")
	}

	var version int
	var memory, iterations, parallelism uint32

	// Parse version: v=19
	_, err := fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil {
		return false, fmt.Errorf("invalid version format: %w", err)
	}

	// Parse parameters: m=65536,t=3,p=2
	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism)
	if err != nil {
		return false, fmt.Errorf("invalid parameters format: %w", err)
	}

	saltBase64 := parts[4]
	hashBase64 := parts[5]

	if version != argon2.Version {
		return false, fmt.Errorf("incompatible version: %d", version)
	}

	// Decode salt and hash
	salt, err := base64.RawStdEncoding.DecodeString(saltBase64)
	if err != nil {
		return false, fmt.Errorf("failed to decode salt: %w", err)
	}

	hash, err := base64.RawStdEncoding.DecodeString(hashBase64)
	if err != nil {
		return false, fmt.Errorf("failed to decode hash: %w", err)
	}

	// Compute the hash of the provided password
	// Use keyLength constant to ensure consistency with hash generation
	computedHash := argon2.IDKey([]byte(password), salt, iterations, memory, uint8(parallelism), keyLength)

	// Compare hashes in constant time
	return subtle.ConstantTimeCompare(hash, computedHash) == 1, nil
}
