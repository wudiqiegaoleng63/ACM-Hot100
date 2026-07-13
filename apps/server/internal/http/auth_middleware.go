package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/acmhot100/server/internal/auth"
	"github.com/acmhot100/server/internal/config"
	"github.com/redis/go-redis/v9"
)

// RequireAuth returns middleware that validates the access JWT from cookie
// and checks the deny list in Redis. Sets "userID" in the Gin context.
func RequireAuth(cfg *config.Config, rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken, err := c.Cookie("access_token")
		if err != nil || accessToken == "" {
			errorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
			c.Abort()
			return
		}

		claims, err := auth.ParseAccessToken(cfg, accessToken)
		if err != nil {
			errorResponse(c, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid or expired access token")
			c.Abort()
			return
		}

		// Check deny list in Redis
		denied, err := auth.IsAccessDenied(rdb, claims.ID)
		if err != nil {
			// Redis error - allow the request through (fail open)
			// but log the error
		} else if denied {
			errorResponse(c, http.StatusUnauthorized, "TOKEN_REVOKED", "Token has been revoked")
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("userID", claims.Subject)
		c.Set("accessJTI", claims.ID)

		// Store remaining TTL for potential logout
		if claims.ExpiresAt != nil {
			remaining := time.Until(claims.ExpiresAt.Time)
			if remaining > 0 {
				c.Set("accessTTL", remaining)
			}
		}

		c.Next()
	}
}

// OptionalAuth returns middleware that validates the access JWT if present
// but does not reject the request if it's missing or invalid.
func OptionalAuth(cfg *config.Config, rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken, err := c.Cookie("access_token")
		if err != nil || accessToken == "" {
			c.Next()
			return
		}

		claims, err := auth.ParseAccessToken(cfg, accessToken)
		if err != nil {
			c.Next()
			return
		}

		// Check deny list
		denied, err := auth.IsAccessDenied(rdb, claims.ID)
		if err == nil && denied {
			c.Next()
			return
		}

		// Set user info in context
		c.Set("userID", claims.Subject)
		c.Set("accessJTI", claims.ID)

		if claims.ExpiresAt != nil {
			remaining := time.Until(claims.ExpiresAt.Time)
			if remaining > 0 {
				c.Set("accessTTL", remaining)
			}
		}

		c.Next()
	}
}
