package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/acmhot100/server/internal/config"
	"github.com/gin-gonic/gin"
)

func TestRequireTrustedOriginProtectsWrites(t *testing.T) {
	cfg := &config.Config{AppBaseURL: "https://acm.example"}
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestID(), RequireTrustedOrigin(cfg))
	router.POST("/write", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	for _, test := range []struct {
		name   string
		origin string
		want   int
	}{
		{name: "missing", want: http.StatusForbidden},
		{name: "foreign", origin: "https://evil.example", want: http.StatusForbidden},
		{name: "same origin", origin: cfg.AppBaseURL, want: http.StatusNoContent},
	} {
		t.Run(test.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/write", nil)
			request.Header.Set("Origin", test.origin)
			router.ServeHTTP(response, request)
			if response.Code != test.want {
				t.Fatalf("status = %d, want %d", response.Code, test.want)
			}
		})
	}
}

func TestCORSProductionAllowsOnlyConfiguredOrigin(t *testing.T) {
	cfg := &config.Config{AppEnv: "production", AppBaseURL: "https://acm.example"}
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CORSConfig(cfg))
	router.GET("/read", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/read", nil)
	request.Header.Set("Origin", "http://localhost:5173")
	router.ServeHTTP(response, request)
	if response.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatal("production must not allow localhost origins")
	}
}
