package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/acmhot100/server/internal/auth"
	"github.com/acmhot100/server/internal/config"
)

func TestRefreshTokenRequiresCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	redisServer := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	cfg := refreshHandlerTestConfig()
	bodyToken, bodyJTI, bodyFamilyID, err := auth.GenerateRefreshToken(cfg, "user-123")
	if err != nil {
		t.Fatalf("generate body refresh token: %v", err)
	}
	if err := auth.StoreRefreshSession(
		rdb,
		bodyJTI,
		"user-123",
		bodyFamilyID,
		time.Duration(cfg.JWTRefreshTTL)*time.Second,
	); err != nil {
		t.Fatalf("store body refresh token: %v", err)
	}

	body, err := json.Marshal(map[string]string{"refresh_token": bodyToken})
	if err != nil {
		t.Fatalf("marshal request body: %v", err)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router := gin.New()
	handler := NewAuthHandler(nil, rdb, cfg)
	router.POST("/api/v1/auth/refresh", handler.RefreshToken)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusUnauthorized, response.Body.String())
	}

	var errorBody struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &errorBody); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if errorBody.Error.Code != "MISSING_TOKEN" {
		t.Fatalf("error code = %q, want MISSING_TOKEN", errorBody.Error.Code)
	}
	if _, _, err := auth.GetRefreshSession(rdb, bodyJTI); err != nil {
		t.Fatalf("body token session was consumed: %v", err)
	}
}

func refreshHandlerTestConfig() *config.Config {
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
