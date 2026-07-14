package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func TestRateLimitRejectsRequestsBeyondFixedWindow(t *testing.T) {
	redisServer := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestID(), RateLimit(rdb, "test", 2, time.Minute))
	router.POST("/write", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	for i, want := range []int{http.StatusNoContent, http.StatusNoContent, http.StatusTooManyRequests} {
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/write", nil)
		request.RemoteAddr = "192.0.2.1:1234"
		router.ServeHTTP(response, request)
		if response.Code != want {
			t.Fatalf("request %d status = %d, want %d", i+1, response.Code, want)
		}
	}
}
