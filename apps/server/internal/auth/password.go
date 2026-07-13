package auth

import (
	"github.com/alexedwards/argon2id"
)

// HashPassword hashes a password using Argon2id with recommended parameters.
func HashPassword(password string) (string, error) {
	hash, err := argon2id.CreateHash(password, &argon2id.Params{
		Memory:      64 * 1024,
		Iterations:  3,
		Parallelism: 2,
		SaltLength:  16,
		KeyLength:   32,
	})
	if err != nil {
		return "", err
	}
	return hash, nil
}

// CheckPassword verifies a password against an Argon2id hash.
func CheckPassword(password, hash string) (bool, error) {
	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, err
	}
	return match, nil
}
