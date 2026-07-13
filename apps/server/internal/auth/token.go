package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

// HashToken returns the SHA-256 hex digest of a raw token.
func HashToken(rawToken string) string {
	h := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(h[:])
}

// GenerateVerifyToken generates a 32-byte random token for email verification.
// Returns the raw token (to send to user) and its SHA-256 hex hash (to store in Redis).
func GenerateVerifyToken() (rawToken string, hash string, err error) {
	return generateRandomToken()
}

// GenerateResetToken generates a 32-byte random token for password reset.
// Returns the raw token (to send to user) and its SHA-256 hex hash (to store in Redis).
func GenerateResetToken() (rawToken string, hash string, err error) {
	return generateRandomToken()
}

// generateRandomToken creates a 32-byte crypto/rand token and returns
// the hex-encoded raw token and its SHA-256 hex hash.
func generateRandomToken() (rawToken string, hash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", err
	}
	rawToken = hex.EncodeToString(b)
	hash = HashToken(rawToken)
	return
}
