package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"

	"github.com/redis/go-redis/v9"
)

// HashToken returns the SHA-256 hex digest of a raw token.
func HashToken(rawToken string) string {
	h := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(h[:])
}

// ConsumeOneTimeToken atomically reads and deletes a token value.
func ConsumeOneTimeToken(rdb *redis.Client, key string) (string, error) {
	return rdb.GetDel(context.Background(), key).Result()
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
