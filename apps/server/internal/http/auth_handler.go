package http

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/acmhot100/server/internal/auth"
	"github.com/acmhot100/server/internal/config"
	"github.com/acmhot100/server/internal/service"
)

// AuthHandler holds dependencies for auth HTTP handlers.
type AuthHandler struct {
	db  *gorm.DB
	rdb *redis.Client
	cfg *config.Config
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(db *gorm.DB, rdb *redis.Client, cfg *config.Config) *AuthHandler {
	return &AuthHandler{db: db, rdb: rdb, cfg: cfg}
}

// ─── Request/Response types ─────────────────────────────────────────────────

type registerRequest struct {
	Email    string `json:"email" binding:"required"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type verifyEmailRequest struct {
	Token string `json:"token" binding:"required"`
}

type resendVerificationRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type loginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type forgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type resetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// errorResponse builds the standard error response format.
func errorResponse(c *gin.Context, statusCode int, code string, message string) {
	requestID, _ := c.Get("request_id")
	c.JSON(statusCode, gin.H{
		"error": gin.H{
			"code":    code,
			"message": message,
		},
		"request_id": requestID,
	})
}

// setAuthCookies sets HttpOnly SameSite=Lax cookies for access and refresh tokens.
func setAuthCookies(c *gin.Context, accessToken, refreshToken string) {
	secure := c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https"
	accessTTL := 15 * time.Minute
	refreshTTL := 7 * 24 * time.Hour

	accessCookie := buildCookieHeader("access_token", accessToken, accessTTL, secure)
	refreshCookie := buildCookieHeader("refresh_token", refreshToken, refreshTTL, secure)
	c.Writer.Header()["Set-Cookie"] = []string{accessCookie, refreshCookie}
}

// buildCookieHeader builds a Set-Cookie header string with HttpOnly and SameSite=Lax.
func buildCookieHeader(name, value string, ttl time.Duration, secure bool) string {
	sameSite := "Lax"
	secureFlag := ""
	if secure {
		secureFlag = "; Secure"
	}
	return fmt.Sprintf("%s=%s; Path=/; Max-Age=%d; HttpOnly; SameSite=%s%s",
		name, value, int(ttl.Seconds()), sameSite, secureFlag)
}

// clearAuthCookies removes the auth cookies.
func clearAuthCookies(c *gin.Context) {
	secure := c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https"
	secureFlag := ""
	if secure {
		secureFlag = "; Secure"
	}
	accessCookie := fmt.Sprintf("access_token=; Path=/; Max-Age=0; HttpOnly; SameSite=Lax%s", secureFlag)
	refreshCookie := fmt.Sprintf("refresh_token=; Path=/; Max-Age=0; HttpOnly; SameSite=Lax%s", secureFlag)
	c.Writer.Header()["Set-Cookie"] = []string{accessCookie, refreshCookie}
}

// ─── Handlers ───────────────────────────────────────────────────────────────

// Register handles POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body")
		return
	}

	err := service.Register(h.db, h.rdb, h.cfg, req.Email, req.Username, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidInput):
			errorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		case errors.Is(err, service.ErrEmailTaken):
			errorResponse(c, http.StatusConflict, "EMAIL_ALREADY_EXISTS", "Email is already registered")
		case errors.Is(err, service.ErrUsernameTaken):
			errorResponse(c, http.StatusConflict, "USERNAME_ALREADY_EXISTS", "Username is already taken")
		default:
			errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to register")
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Registration successful. Please check your email to verify your account.",
	})
}

// VerifyEmail handles POST /api/v1/auth/verify-email
func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var req verifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body")
		return
	}

	err := service.VerifyEmail(h.db, h.rdb, req.Token)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTokenExpired):
			errorResponse(c, http.StatusBadRequest, "TOKEN_EXPIRED", "Verification token has expired or already been used")
		case errors.Is(err, service.ErrAlreadyVerified):
			errorResponse(c, http.StatusOK, "ALREADY_VERIFIED", "Email is already verified")
		default:
			errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to verify email")
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Email verified successfully",
	})
}

// ResendVerification handles POST /api/v1/auth/resend-verification
func (h *AuthHandler) ResendVerification(c *gin.Context) {
	var req resendVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body")
		return
	}

	err := service.ResendVerification(h.db, h.rdb, h.cfg, req.Email)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrAlreadyVerified):
			errorResponse(c, http.StatusBadRequest, "ALREADY_VERIFIED", "Email is already verified")
		case errors.Is(err, service.ErrRateLimited):
			errorResponse(c, http.StatusTooManyRequests, "RATE_LIMITED", "Please wait before requesting another verification email")
		default:
			// Don't reveal if email exists
			c.JSON(http.StatusOK, gin.H{
				"message": "If the email exists and is not verified, a new verification email has been sent",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "If the email exists and is not verified, a new verification email has been sent",
	})
}

// Login handles POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body")
		return
	}

	accessToken, refreshToken, err := service.Login(h.db, h.rdb, h.cfg, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			errorResponse(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid email or password")
		} else if errors.Is(err, service.ErrEmailNotVerified) {
			errorResponse(c, http.StatusForbidden, "EMAIL_NOT_VERIFIED", "Email address is not verified")
		} else {
			errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Login failed")
		}
		return
	}

	setAuthCookies(c, accessToken, refreshToken)

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
	})
}

// RefreshToken handles POST /api/v1/auth/refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	refreshTokenStr, err := c.Cookie("refresh_token")
	if err != nil || refreshTokenStr == "" {
		errorResponse(c, http.StatusUnauthorized, "MISSING_TOKEN", "Refresh token required")
		return
	}

	newAccessToken, newRefreshToken, err := service.RefreshToken(h.rdb, h.cfg, refreshTokenStr)
	if err != nil {
		if errors.Is(err, service.ErrTokenReuse) {
			clearAuthCookies(c)
			errorResponse(c, http.StatusUnauthorized, "TOKEN_REUSE", "Token reuse detected. All sessions have been revoked.")
		} else if errors.Is(err, service.ErrTokenExpired) {
			errorResponse(c, http.StatusUnauthorized, "TOKEN_EXPIRED", "Refresh token has expired")
		} else {
			errorResponse(c, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid refresh token")
		}
		return
	}

	setAuthCookies(c, newAccessToken, newRefreshToken)

	c.JSON(http.StatusOK, gin.H{
		"message": "Tokens refreshed successfully",
	})
}

// Logout handles POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// Extract access token JTI from cookie
	accessJTI := ""
	var accessTTL time.Duration
	if accessTokenStr, err := c.Cookie("access_token"); err == nil {
		if claims, err := auth.ParseAccessToken(h.cfg, accessTokenStr); err == nil {
			accessJTI = claims.ID
			if claims.ExpiresAt != nil {
				remaining := time.Until(claims.ExpiresAt.Time)
				if remaining > 0 {
					accessTTL = remaining
				}
			}
		}
	}

	// Extract refresh token JTI from cookie
	refreshJTI := ""
	if refreshTokenStr, err := c.Cookie("refresh_token"); err == nil {
		if claims, err := auth.ParseRefreshToken(h.cfg, refreshTokenStr); err == nil {
			refreshJTI = claims.ID
		}
	}

	_ = service.Logout(h.rdb, accessJTI, accessTTL, refreshJTI)

	clearAuthCookies(c)

	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}

// LogoutAll handles POST /api/v1/auth/logout-all
func (h *AuthHandler) LogoutAll(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		errorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	if err := service.LogoutAll(h.rdb, userID.(string)); err != nil {
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to logout all sessions")
		return
	}

	clearAuthCookies(c)

	c.JSON(http.StatusOK, gin.H{
		"message": "All sessions have been revoked",
	})
}

// ForgotPassword handles POST /api/v1/auth/forgot-password
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req forgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body")
		return
	}

	// Always returns nil - no email enumeration
	_ = service.ForgotPassword(h.db, h.rdb, h.cfg, req.Email)

	c.JSON(http.StatusOK, gin.H{
		"message": "If the email exists and is verified, a password reset link has been sent",
	})
}

// ResetPassword handles POST /api/v1/auth/reset-password
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req resetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body")
		return
	}

	err := service.ResetPassword(h.db, h.rdb, req.Token, req.NewPassword)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidInput):
			errorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		case errors.Is(err, service.ErrTokenExpired):
			errorResponse(c, http.StatusBadRequest, "TOKEN_EXPIRED", "Reset token has expired or already been used")
		default:
			errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to reset password")
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password reset successfully",
	})
}

// GetCurrentUser handles GET /api/v1/auth/me
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		errorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	user, err := service.GetCurrentUser(h.db, userID.(string))
	if err != nil {
		errorResponse(c, http.StatusNotFound, "USER_NOT_FOUND", "User not found")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}
