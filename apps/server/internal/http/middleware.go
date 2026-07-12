package http

import (
	"net/http"
	"strings"

	"github.com/acmhot100/server/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestID middleware generates a unique UUID for each request and sets it
// in the context and response header.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := uuid.New().String()
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// CORSConfig returns a middleware that handles Cross-Origin Resource Sharing.
// In development mode, it allows requests from localhost:5173.
func CORSConfig(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		allowedOrigins := []string{
			"http://localhost:5173",
			"http://127.0.0.1:5173",
		}

		// In development, also allow the configured base URL
		if cfg.IsDevelopment() {
			allowedOrigins = append(allowedOrigins, cfg.AppBaseURL)
		}

		isAllowed := false
		for _, o := range allowedOrigins {
			if origin == o {
				isAllowed = true
				break
			}
		}

		if isAllowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Request-ID")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

		// Handle preflight requests
		if c.Request.Method == http.MethodOptions {
			// If origin is not allowed, still respond to preflight but without Allow-Origin
			if !isAllowed && origin != "" {
				c.Header("Access-Control-Allow-Origin", "")
			}
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		// For non-preflight requests with disallowed origin, just continue
		// The browser will block the response if origin doesn't match
		c.Next()
	}
}

// AuthRequired is a placeholder middleware for JWT authentication.
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "authorization header required",
			})
			return
		}
		// TODO: Validate JWT token
		c.Next()
	}
}
