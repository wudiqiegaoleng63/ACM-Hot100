package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/acmhot100/server/internal/config"
)

// AccessClaims holds the registered claims for an access token.
type AccessClaims struct {
	jwt.RegisteredClaims
}

// RefreshClaims holds the registered claims for a refresh token.
type RefreshClaims struct {
	jwt.RegisteredClaims
	FamilyID string `json:"fid,omitempty"`
}

// GenerateAccessToken creates a signed JWT access token.
func GenerateAccessToken(cfg *config.Config, userID string) (tokenString string, jti string, err error) {
	now := jwt.NewNumericDate(time.Now())
	ttl := cfg.JWTAccessTTL

	jti = uuid.New().String()

	claims := AccessClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    cfg.JWTIssuer,
			Subject:   userID,
			Audience:  jwt.ClaimStrings{cfg.JWTAccessAudience},
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(ttl) * time.Second)),
			NotBefore: now,
			IssuedAt:  now,
			ID:        jti,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err = token.SignedString([]byte(cfg.JWTAccessSecret))
	return
}

// GenerateRefreshToken creates a signed JWT refresh token with a family ID for rotation.
func GenerateRefreshToken(cfg *config.Config, userID string) (tokenString string, jti string, familyID string, err error) {
	now := jwt.NewNumericDate(time.Now())
	ttl := cfg.JWTRefreshTTL

	jti = uuid.New().String()
	familyID = uuid.New().String()

	claims := RefreshClaims{
		FamilyID: familyID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    cfg.JWTIssuer,
			Subject:   userID,
			Audience:  jwt.ClaimStrings{cfg.JWTRefreshAudience},
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(ttl) * time.Second)),
			NotBefore: now,
			IssuedAt:  now,
			ID:        jti,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err = token.SignedString([]byte(cfg.JWTRefreshSecret))
	return
}

// ParseAccessToken parses and validates an access token string.
func ParseAccessToken(cfg *config.Config, tokenString string) (*AccessClaims, error) {
	claims := &AccessClaims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		return []byte(cfg.JWTAccessSecret), nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil {
		return nil, err
	}
	return claims, nil
}

// ParseRefreshToken parses and validates a refresh token string.
func ParseRefreshToken(cfg *config.Config, tokenString string) (*RefreshClaims, error) {
	claims := &RefreshClaims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		return []byte(cfg.JWTRefreshSecret), nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil {
		return nil, err
	}
	return claims, nil
}
