package http

import (
	"net/http"

	"github.com/acmhot100/server/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Server creates and configures a new Gin engine with all routes and middleware.
func NewServer(cfg *config.Config, db *gorm.DB, rdb *redis.Client) *gin.Engine {
	if cfg.IsDevelopment() {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// Global middleware
	r.Use(gin.Recovery())
	r.Use(RequestID())
	r.Use(CORSConfig(cfg))

	// Health check
	r.GET("/api/v1/health", healthCheck(db, rdb))

	// API v1 route group
	v1 := r.Group("/api/v1")
	{
		// Auth routes (placeholder)
		auth := v1.Group("/auth")
		_ = auth

		// Problems routes (placeholder)
		problems := v1.Group("/problems")
		_ = problems

		// Submissions routes (placeholder)
		submissions := v1.Group("/submissions")
		_ = submissions

		// User routes (placeholder)
		users := v1.Group("/users")
		_ = users
	}

	return r
}

// healthCheck returns a handler that checks MySQL and Redis connectivity.
func healthCheck(db *gorm.DB, rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		services := make(map[string]string)

		// Check MySQL
		sqlDB, err := db.DB()
		if err != nil {
			services["mysql"] = "error"
		} else if err := sqlDB.Ping(); err != nil {
			services["mysql"] = "error"
		} else {
			services["mysql"] = "ok"
		}

		// Check Redis
		if err := rdb.Ping(c.Request.Context()).Err(); err != nil {
			services["redis"] = "error"
		} else {
			services["redis"] = "ok"
		}

		status := http.StatusOK
		for _, s := range services {
			if s != "ok" {
				status = http.StatusServiceUnavailable
				break
			}
		}

		c.JSON(status, gin.H{
			"status":   "ok",
			"services": services,
		})
	}
}
