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

		allowedOrigins := []string{cfg.AppBaseURL}
		if cfg.IsDevelopment() {
			allowedOrigins = append(allowedOrigins,
				"http://localhost:5173",
				"http://127.0.0.1:5173",
			)
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

		c.Next()
	}
}

// RequireTrustedOrigin rejects cross-site state-changing requests authenticated by cookies.
func RequireTrustedOrigin(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodGet || c.Request.Method == http.MethodHead || c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}
		origin := c.GetHeader("Origin")
		if origin == "" || origin != cfg.AppBaseURL {
			apiError(c, http.StatusForbidden, "INVALID_ORIGIN", "request origin is not allowed")
			c.Abort()
			return
		}
		c.Next()
	}
}

// SecurityHeaders applies browser hardening headers to API responses.
func SecurityHeaders(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		if cfg.AppEnv == "production" {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
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
