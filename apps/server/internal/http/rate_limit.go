package http

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strconv"
	"time"

	"github.com/acmhot100/server/internal/queue"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

var fixedWindowScript = redis.NewScript(`
local current = redis.call('INCR', KEYS[1])
if current == 1 then
  redis.call('EXPIRE', KEYS[1], ARGV[1])
end
return current
`)

// RateLimit applies a Redis-backed fixed-window request limit.
func RateLimit(rdb *redis.Client, scope string, limit int64, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if rdb == nil {
			c.Next()
			return
		}

		identifier := c.ClientIP()
		key := queue.KeyRate(scope + ":" + hashRateIdentifier(identifier))
		count, err := fixedWindowScript.Run(c.Request.Context(), rdb, []string{key}, int64(window.Seconds())).Int64()
		if err != nil {
			apiError(c, http.StatusServiceUnavailable, "RATE_LIMIT_UNAVAILABLE", "request protection is temporarily unavailable")
			c.Abort()
			return
		}
		if count > limit {
			c.Header("Retry-After", formatRetryAfter(window))
			apiError(c, http.StatusTooManyRequests, "RATE_LIMITED", "too many requests, please try again later")
			c.Abort()
			return
		}
		c.Next()
	}
}

func hashRateIdentifier(identifier string) string {
	sum := sha256.Sum256([]byte(identifier))
	return hex.EncodeToString(sum[:16])
}

func formatRetryAfter(window time.Duration) string {
	seconds := int64(window.Seconds())
	if seconds < 1 {
		seconds = 1
	}
	return strconv.FormatInt(seconds, 10)
}
