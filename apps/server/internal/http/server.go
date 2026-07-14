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

	// Health check (exposes judge_mode so frontend can show Mock Judge badge)
	r.GET("/api/v1/health", healthCheck(db, rdb, cfg))

	// API v1 route group
	v1 := r.Group("/api/v1")
	{
		// Auth routes
		authHandler := NewAuthHandler(db, rdb, cfg)
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/register", authHandler.Register)
			authGroup.POST("/verify-email", authHandler.VerifyEmail)
			authGroup.POST("/resend-verification", authHandler.ResendVerification)
			authGroup.POST("/login", authHandler.Login)
			authGroup.POST("/refresh", authHandler.RefreshToken)
			authGroup.POST("/logout", authHandler.Logout)
			authGroup.POST("/logout-all", RequireAuth(cfg, rdb), authHandler.LogoutAll)
			authGroup.POST("/forgot-password", authHandler.ForgotPassword)
			authGroup.POST("/reset-password", authHandler.ResetPassword)
			authGroup.GET("/me", RequireAuth(cfg, rdb), authHandler.GetCurrentUser)
		}

		// Problems routes
		problems := v1.Group("/problems")
		problems.Use(OptionalAuth(cfg, rdb))
		{
			problems.GET("", listProblems(db))
			problems.GET("/:slug", getProblem(db))
			problems.GET("/:slug/navigation", getProblemNavigation(db))
			problems.POST("/:slug/run", RequireAuth(cfg, rdb), createSampleRun(db, rdb))
			problems.POST("/:slug/submissions", RequireAuth(cfg, rdb), createSubmission(db, rdb))

			// Draft routes (require auth)
			drafts := problems.Group("/:slug/drafts")
			drafts.Use(RequireAuth(cfg, rdb))
			{
				drafts.PUT("/:language_key", saveDraft(db))
				drafts.GET("/:language_key", getDraft(db))
			}
		}

		// Sample run routes
		v1.GET("/runs/:id", RequireAuth(cfg, rdb), getSampleRun(db))

		// Public content metadata
		v1.GET("/tags", listTags(db))
		v1.GET("/languages", listLanguages(db))

		// Submissions routes
		submissions := v1.Group("/submissions")
		submissions.Use(RequireAuth(cfg, rdb))
		{
			submissions.GET("", listSubmissions(db))
			submissions.GET("/:id", getSubmission(db))
		}

		// Profile routes
		profile := v1.Group("/profile")
		profile.Use(RequireAuth(cfg, rdb))
		{
			profile.GET("/summary", getProfileSummary(db))
			profile.GET("/progress-by-stage", getProfileProgressByStage(db))
		}
	}

	return r
}

// healthCheck returns a handler that checks MySQL and Redis connectivity.
func healthCheck(db *gorm.DB, rdb *redis.Client, cfg *config.Config) gin.HandlerFunc {
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
			"status":     "ok",
			"services":   services,
			"judge_mode": cfg.JudgeMode,
		})
	}
}
