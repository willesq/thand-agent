package common

import (
	"crypto/rand"
	"encoding/base64"
)

// GenerateSecureRandomString generates a cryptographically secure random string of the specified length
func GenerateSecureRandomString(length int) (string, error) {
	// Calculate the number of bytes needed (base64 encoding expands by ~4/3)
	byteLength := (length*3 + 3) / 4

	bytes := make([]byte, byteLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// Encode to base64 and trim to the requested length
	encoded := base64.URLEncoding.EncodeToString(bytes)
	return encoded[:length], nil
}
