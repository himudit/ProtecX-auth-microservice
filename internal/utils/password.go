package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// HashPassword hashes a plain text password using Argon2id.
// Returns the encoded hash including the salt.
func HashPassword(password string) (string, error) {
	// Generate a random 16-byte salt
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}

	// Argon2id parameters (you can tweak these)
	timeCost := uint32(1)
	memory := uint32(64 * 1024) // 64 MB
	threads := uint8(4)
	keyLen := uint32(32)

	hash := argon2.IDKey([]byte(password), salt, timeCost, memory, threads, keyLen)

	// Encode salt + hash for storage
	saltStr := base64.RawStdEncoding.EncodeToString(salt)
	hashStr := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf("%s.%s", saltStr, hashStr), nil
}

func VerifyPassword(storedHash, password string) (bool, error) {
	// Split stored hash into salt and hash
	parts := strings.SplitN(storedHash, ".", 2)
	if len(parts) != 2 {
		return false, fmt.Errorf("invalid stored hash format")
	}

	saltStr := parts[0]
	hashStr := parts[1]

	salt, err := base64.RawStdEncoding.DecodeString(saltStr)
	if err != nil {
		return false, err
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(hashStr)
	if err != nil {
		return false, err
	}

	// Must match parameters used in HashPassword
	timeCost := uint32(1)
	memory := uint32(64 * 1024)
	threads := uint8(4)
	keyLen := uint32(32)

	computedHash := argon2.IDKey([]byte(password), salt, timeCost, memory, threads, keyLen)

	if len(computedHash) != len(expectedHash) {
		return false, nil
	}

	// Constant-time comparison
	diff := byte(0)
	for i := 0; i < len(expectedHash); i++ {
		diff |= computedHash[i] ^ expectedHash[i]
	}

	return diff == 0, nil
}
