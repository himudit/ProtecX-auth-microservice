package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha512"
	"encoding/base64"
	"errors"
	"os"

	"golang.org/x/crypto/pbkdf2"
)

const (
	SaltLength = 64
	IVLength   = 16
	TagLength  = 16
	KeyLength  = 32
	Iterations = 100000
)

// DecryptAES256GCM decrypts data encrypted by the Node.js service
func DecryptAES256GCM(cipherText string) (string, error) {
	encryptionKey := os.Getenv("ENCRYPTION_KEY")
	if encryptionKey == "" {
		return "", errors.New("ENCRYPTION_KEY not set")
	}

	data, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return "", err
	}

	if len(data) < SaltLength+IVLength+TagLength {
		return "", errors.New("ciphertext too short")
	}

	salt := data[:SaltLength]
	iv := data[SaltLength : SaltLength+IVLength]
	tag := data[SaltLength+IVLength : SaltLength+IVLength+TagLength]
	encrypted := data[SaltLength+IVLength+TagLength:]

	// Derive key
	key := pbkdf2.Key(
		[]byte(encryptionKey),
		salt,
		Iterations,
		KeyLength,
		sha512.New,
	)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCMWithNonceSize(block, 16)
	if err != nil {
		return "", err
	}

	// Node gives ciphertext + tag separately
	combined := append(encrypted, tag...)

	plain, err := gcm.Open(nil, iv, combined, nil)
	if err != nil {
		return "", err
	}

	return string(plain), nil
}
