package service

import (
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"github.com/acmhot100/server/internal/auth"
	"github.com/acmhot100/server/internal/config"
	"github.com/acmhot100/server/internal/queue"
)

func TestRefreshTokenRotationAndReuse(t *testing.T) {
	redisServer := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	cfg := refreshTestConfig()
	userID := "user-123"
	oldToken, oldJTI, familyID, err := auth.GenerateRefreshToken(cfg, userID)
	if err != nil {
		t.Fatalf("generate initial refresh token: %v", err)
	}
	if err := auth.StoreRefreshSession(
		rdb,
		oldJTI,
		userID,
		familyID,
		time.Duration(cfg.JWTRefreshTTL)*time.Second,
	); err != nil {
		t.Fatalf("store initial refresh session: %v", err)
	}

	_, firstRotatedToken, err := RefreshToken(rdb, cfg, oldToken)
	if err != nil {
		t.Fatalf("first refresh: %v", err)
	}
	firstClaims, err := auth.ParseRefreshToken(cfg, firstRotatedToken)
	if err != nil {
		t.Fatalf("parse first rotated refresh token: %v", err)
	}
	if firstClaims.FamilyID != familyID {
		t.Fatalf("first rotated family ID = %q, want %q", firstClaims.FamilyID, familyID)
	}
	assertCurrentFamilyMember(t, rdb, familyID, firstClaims.ID)

	_, secondRotatedToken, err := RefreshToken(rdb, cfg, firstRotatedToken)
	if err != nil {
		t.Fatalf("second refresh: %v", err)
	}
	secondClaims, err := auth.ParseRefreshToken(cfg, secondRotatedToken)
	if err != nil {
		t.Fatalf("parse second rotated refresh token: %v", err)
	}
	if secondClaims.FamilyID != familyID {
		t.Fatalf("second rotated family ID = %q, want %q", secondClaims.FamilyID, familyID)
	}
	assertCurrentFamilyMember(t, rdb, familyID, secondClaims.ID)

	if _, _, err := RefreshToken(rdb, cfg, oldToken); !errors.Is(err, ErrTokenReuse) {
		t.Fatalf("reuse old refresh token error = %v, want ErrTokenReuse", err)
	}
	if _, _, err := auth.GetRefreshSession(rdb, secondClaims.ID); err != redis.Nil {
		t.Fatalf("latest family session remains after reuse detection: %v", err)
	}
}

func TestRefreshTokenConcurrentReuseRevokesFamily(t *testing.T) {
	redisServer := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	cfg := refreshTestConfig()
	oldToken, oldJTI, familyID, err := auth.GenerateRefreshToken(cfg, "user-123")
	if err != nil {
		t.Fatalf("generate refresh token: %v", err)
	}
	if err := auth.StoreRefreshSession(
		rdb,
		oldJTI,
		"user-123",
		familyID,
		time.Duration(cfg.JWTRefreshTTL)*time.Second,
	); err != nil {
		t.Fatalf("store refresh session: %v", err)
	}

	type result struct {
		token string
		err   error
	}
	start := make(chan struct{})
	results := make(chan result, 2)
	for range 2 {
		go func() {
			<-start
			_, token, err := RefreshToken(rdb, cfg, oldToken)
			results <- result{token: token, err: err}
		}()
	}
	close(start)

	var successfulToken string
	var reuseCount int
	for range 2 {
		result := <-results
		switch {
		case result.err == nil:
			if successfulToken != "" {
				t.Fatal("concurrent refresh token was accepted more than once")
			}
			successfulToken = result.token
		case errors.Is(result.err, ErrTokenReuse):
			reuseCount++
		default:
			t.Fatalf("unexpected concurrent refresh error: %v", result.err)
		}
	}
	if successfulToken == "" || reuseCount != 1 {
		t.Fatalf("successful token present=%t, reuse errors=%d; want one each", successfulToken != "", reuseCount)
	}

	claims, err := auth.ParseRefreshToken(cfg, successfulToken)
	if err != nil {
		t.Fatalf("parse successful refresh token: %v", err)
	}
	if _, _, err := auth.GetRefreshSession(rdb, claims.ID); err != redis.Nil {
		t.Fatalf("family remained valid after concurrent reuse: %v", err)
	}
}

func assertCurrentFamilyMember(t *testing.T, rdb *redis.Client, familyID, wantJTI string) {
	t.Helper()
	members, err := rdb.SMembers(t.Context(), queue.KeyAuthFamily(familyID)).Result()
	if err != nil {
		t.Fatalf("read refresh family members: %v", err)
	}
	if len(members) != 1 || members[0] != wantJTI {
		t.Fatalf("family members = %v, want only %q", members, wantJTI)
	}
}

func refreshTestConfig() *config.Config {
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
