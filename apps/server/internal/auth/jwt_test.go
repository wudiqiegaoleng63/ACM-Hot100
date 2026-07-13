package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/acmhot100/server/internal/config"
)

func TestGenerateAndParseJWT(t *testing.T) {
	t.Parallel()

	cfg := testJWTConfig()
	userID := "user-123"

	tests := []struct {
		name       string
		generate   func(*config.Config, string) (string, string, string, error)
		parse      func(*config.Config, string) (jwt.Claims, error)
		audience   string
		secretName string
		ttl        time.Duration
	}{
		{
			name: "access token",
			generate: func(cfg *config.Config, userID string) (string, string, string, error) {
				token, jti, err := GenerateAccessToken(cfg, userID)
				return token, jti, "", err
			},
			parse: func(cfg *config.Config, token string) (jwt.Claims, error) {
				return ParseAccessToken(cfg, token)
			},
			audience:   cfg.JWTAccessAudience,
			secretName: "access",
			ttl:        time.Duration(cfg.JWTAccessTTL) * time.Second,
		},
		{
			name: "refresh token",
			generate: func(cfg *config.Config, userID string) (string, string, string, error) {
				return GenerateRefreshToken(cfg, userID)
			},
			parse: func(cfg *config.Config, token string) (jwt.Claims, error) {
				return ParseRefreshToken(cfg, token)
			},
			audience:   cfg.JWTRefreshAudience,
			secretName: "refresh",
			ttl:        time.Duration(cfg.JWTRefreshTTL) * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			before := time.Now()
			tokenString, generatedJTI, familyID, err := tt.generate(cfg, userID)
			if err != nil {
				t.Fatalf("generate %s token: %v", tt.secretName, err)
			}
			claims, err := tt.parse(cfg, tokenString)
			if err != nil {
				t.Fatalf("parse generated %s token: %v", tt.secretName, err)
			}

			issuer, err := claims.GetIssuer()
			if err != nil || issuer != cfg.JWTIssuer {
				t.Errorf("issuer = %q, %v; want %q", issuer, err, cfg.JWTIssuer)
			}
			subject, err := claims.GetSubject()
			if err != nil || subject != userID {
				t.Errorf("subject = %q, %v; want %q", subject, err, userID)
			}
			audience, err := claims.GetAudience()
			if err != nil || len(audience) != 1 || audience[0] != tt.audience {
				t.Errorf("audience = %v, %v; want [%q]", audience, err, tt.audience)
			}
			expiresAt, err := claims.GetExpirationTime()
			if err != nil || expiresAt == nil {
				t.Fatalf("expiration = %v, %v; want a timestamp", expiresAt, err)
			}
			if expiresAt.Time.Before(before.Add(tt.ttl-time.Second)) || expiresAt.Time.After(time.Now().Add(tt.ttl+time.Second)) {
				t.Errorf("expiration %v is not approximately %v from generation", expiresAt.Time, tt.ttl)
			}

			switch typedClaims := claims.(type) {
			case *AccessClaims:
				if typedClaims.ID != generatedJTI || typedClaims.ID == "" {
					t.Errorf("access JTI = %q, want generated JTI %q", typedClaims.ID, generatedJTI)
				}
			case *RefreshClaims:
				if typedClaims.ID != generatedJTI || typedClaims.ID == "" {
					t.Errorf("refresh JTI = %q, want generated JTI %q", typedClaims.ID, generatedJTI)
				}
				if typedClaims.FamilyID != familyID || typedClaims.FamilyID == "" {
					t.Errorf("family ID = %q, want generated family ID %q", typedClaims.FamilyID, familyID)
				}
			default:
				t.Fatalf("unexpected claims type %T", claims)
			}
		})
	}
}

func TestParseAccessTokenRejectsInvalidClaims(t *testing.T) {
	t.Parallel()

	cfg := testJWTConfig()
	now := time.Now()

	tests := []struct {
		name          string
		claims        AccessClaims
		signingMethod jwt.SigningMethod
	}{
		{
			name: "expired",
			claims: AccessClaims{RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(now.Add(-time.Minute)),
			}},
			signingMethod: jwt.SigningMethodHS256,
		},
		{
			name: "not valid yet",
			claims: AccessClaims{RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
				NotBefore: jwt.NewNumericDate(now.Add(time.Minute)),
			}},
			signingMethod: jwt.SigningMethodHS256,
		},
		{
			name: "disallowed algorithm",
			claims: AccessClaims{RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
			}},
			signingMethod: jwt.SigningMethodHS384,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			token := jwt.NewWithClaims(tt.signingMethod, tt.claims)
			tokenString, err := token.SignedString([]byte(cfg.JWTAccessSecret))
			if err != nil {
				t.Fatalf("sign test token: %v", err)
			}
			if _, err := ParseAccessToken(cfg, tokenString); err == nil {
				t.Fatal("ParseAccessToken accepted an invalid token")
			}
		})
	}
}

func TestParseJWTRejectsInvalidRequiredClaims(t *testing.T) {
	t.Parallel()

	cfg := testJWTConfig()
	now := time.Now()
	validRegisteredClaims := func(audience string) jwt.RegisteredClaims {
		return jwt.RegisteredClaims{
			Issuer:    cfg.JWTIssuer,
			Subject:   "user-123",
			Audience:  jwt.ClaimStrings{audience},
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
			NotBefore: jwt.NewNumericDate(now.Add(-time.Minute)),
			IssuedAt:  jwt.NewNumericDate(now.Add(-time.Minute)),
			ID:        "jti-123",
		}
	}

	accessTests := []struct {
		name   string
		mutate func(*AccessClaims)
	}{
		{name: "wrong issuer", mutate: func(claims *AccessClaims) { claims.Issuer = "other-issuer" }},
		{name: "wrong audience", mutate: func(claims *AccessClaims) { claims.Audience = jwt.ClaimStrings{"other-audience"} }},
		{name: "missing expiration", mutate: func(claims *AccessClaims) { claims.ExpiresAt = nil }},
		{name: "missing not before", mutate: func(claims *AccessClaims) { claims.NotBefore = nil }},
		{name: "missing JTI", mutate: func(claims *AccessClaims) { claims.ID = "" }},
	}

	for _, tt := range accessTests {
		t.Run("access "+tt.name, func(t *testing.T) {
			t.Parallel()
			claims := AccessClaims{RegisteredClaims: validRegisteredClaims(cfg.JWTAccessAudience)}
			tt.mutate(&claims)
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
			tokenString, err := token.SignedString([]byte(cfg.JWTAccessSecret))
			if err != nil {
				t.Fatalf("sign access token: %v", err)
			}
			if _, err := ParseAccessToken(cfg, tokenString); err == nil {
				t.Fatal("ParseAccessToken accepted invalid required claims")
			}
		})
	}

	refreshTests := []struct {
		name   string
		mutate func(*RefreshClaims)
	}{
		{name: "wrong issuer", mutate: func(claims *RefreshClaims) { claims.Issuer = "other-issuer" }},
		{name: "wrong audience", mutate: func(claims *RefreshClaims) { claims.Audience = jwt.ClaimStrings{"other-audience"} }},
		{name: "missing expiration", mutate: func(claims *RefreshClaims) { claims.ExpiresAt = nil }},
		{name: "missing not before", mutate: func(claims *RefreshClaims) { claims.NotBefore = nil }},
		{name: "missing JTI", mutate: func(claims *RefreshClaims) { claims.ID = "" }},
		{name: "missing family ID", mutate: func(claims *RefreshClaims) { claims.FamilyID = "" }},
	}

	for _, tt := range refreshTests {
		t.Run("refresh "+tt.name, func(t *testing.T) {
			t.Parallel()
			claims := RefreshClaims{
				RegisteredClaims: validRegisteredClaims(cfg.JWTRefreshAudience),
				FamilyID:         "family-123",
			}
			tt.mutate(&claims)
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
			tokenString, err := token.SignedString([]byte(cfg.JWTRefreshSecret))
			if err != nil {
				t.Fatalf("sign refresh token: %v", err)
			}
			if _, err := ParseRefreshToken(cfg, tokenString); err == nil {
				t.Fatal("ParseRefreshToken accepted invalid required claims")
			}
		})
	}
}

func testJWTConfig() *config.Config {
	return &config.Config{
		JWTIssuer:          "test-issuer",
		JWTAccessAudience:  "test-access",
		JWTRefreshAudience: "test-refresh",
		JWTAccessSecret:    "test-access-secret-at-least-32-bytes",
		JWTRefreshSecret:   "test-refresh-secret-at-least-32-bytes",
		JWTAccessTTL:       900,
		JWTRefreshTTL:      604800,
	}
}
