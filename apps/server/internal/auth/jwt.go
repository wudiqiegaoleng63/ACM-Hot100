package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/acmhot100/server/internal/config"
)

// AccessClaims holds the registered claims for an access token.
type AccessClaims struct {
	jwt.RegisteredClaims
}

// Validate enforces required access-token claims beyond the standard parser checks.
func (c AccessClaims) Validate() error {
	if c.ID == "" || c.NotBefore == nil {
		return errors.New("missing required access token claims")
	}
	return nil
}

// RefreshClaims holds the registered claims for a refresh token.
type RefreshClaims struct {
	jwt.RegisteredClaims
	FamilyID string `json:"fid,omitempty"`
}

// Validate enforces required refresh-token claims beyond the standard parser checks.
func (c RefreshClaims) Validate() error {
	if c.ID == "" || c.NotBefore == nil || c.FamilyID == "" {
		return errors.New("missing required refresh token claims")
	}
	return nil
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

// GenerateRefreshToken creates a signed JWT refresh token with a new family ID.
func GenerateRefreshToken(cfg *config.Config, userID string) (tokenString string, jti string, familyID string, err error) {
	return GenerateRefreshTokenInFamily(cfg, userID, uuid.New().String())
}

// GenerateRefreshTokenInFamily creates a signed JWT refresh token in an existing family.
func GenerateRefreshTokenInFamily(cfg *config.Config, userID, familyID string) (tokenString string, jti string, returnedFamilyID string, err error) {
	now := jwt.NewNumericDate(time.Now())
	ttl := cfg.JWTRefreshTTL

	jti = uuid.New().String()

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
	returnedFamilyID = familyID
	return
}

// ParseAccessToken parses and validates an access token string.
func ParseAccessToken(cfg *config.Config, tokenString string) (*AccessClaims, error) {
	claims := &AccessClaims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		return []byte(cfg.JWTAccessSecret), nil
	},
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithIssuer(cfg.JWTIssuer),
		jwt.WithAudience(cfg.JWTAccessAudience),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
	)
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
	},
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithIssuer(cfg.JWTIssuer),
		jwt.WithAudience(cfg.JWTRefreshAudience),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
	)
	if err != nil {
		return nil, err
	}
	return claims, nil
}
