package service

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/acmhot100/server/internal/auth"
	"github.com/acmhot100/server/internal/config"
	"github.com/acmhot100/server/internal/model"
	"github.com/acmhot100/server/internal/queue"
	"github.com/acmhot100/server/internal/repository"
)

var (
	ErrInvalidInput       = errors.New("invalid input")
	ErrEmailTaken         = errors.New("email already taken")
	ErrUsernameTaken      = errors.New("username already taken")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailNotVerified   = errors.New("email not verified")
	ErrTokenExpired       = errors.New("token expired or not found")
	ErrTokenReuse         = errors.New("token reuse detected")
	ErrAlreadyVerified    = errors.New("email already verified")
	ErrRateLimited        = errors.New("please wait before requesting another email")
	ErrUserNotFound       = errors.New("user not found")
)

// Register creates a new user account and sends a verification email.
func Register(db *gorm.DB, rdb *redis.Client, cfg *config.Config, email, username, password string) error {
	// Validate input
	email = strings.TrimSpace(email)
	username = strings.TrimSpace(username)

	if err := validateEmail(email); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}
	if err := validateUsername(username); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}
	if err := validatePassword(password); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	// Check uniqueness
	exists, err := repository.EmailExists(db, email)
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}
	if exists {
		return ErrEmailTaken
	}

	exists, err = repository.UsernameExists(db, username)
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}
	if exists {
		return ErrUsernameTaken
	}

	// Hash password
	hash, err := auth.HashPassword(password)
	if err != nil {
		return fmt.Errorf("password hashing failed: %w", err)
	}

	// Create user with PENDING status
	user := &model.User{
		ID:           uuid.New().String(),
		Email:        email,
		Username:     username,
		PasswordHash: hash,
		Status:       model.UserStatusPending,
	}

	if err := repository.CreateUser(db, user); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	// Generate verify token
	rawToken, tokenHash, err := auth.GenerateVerifyToken()
	if err != nil {
		return fmt.Errorf("failed to generate verify token: %w", err)
	}

	// Store token hash in Redis (TTL 30 minutes)
	ctx := context.Background()
	verifyKey := queue.KeyAuthVerify(tokenHash)
	verifyUserKey := queue.KeyAuthVerifyUser(user.ID)

	pipe := rdb.Pipeline()
	pipe.Set(ctx, verifyKey, user.ID, 30*time.Minute)
	pipe.Set(ctx, verifyUserKey, tokenHash, 30*time.Minute)
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to store verify token: %w", err)
	}

	// Send verification email (non-blocking, log error but don't fail registration)
	if sendErr := auth.SendVerificationEmail(cfg, email, rawToken, cfg.AppBaseURL); sendErr != nil {
		// Log but don't fail - user can resend verification
		fmt.Printf("Warning: failed to send verification email: %v\n", sendErr)
	}

	return nil
}

// VerifyEmail verifies a user's email using the raw token from the email link.
func VerifyEmail(db *gorm.DB, rdb *redis.Client, rawToken string) error {
	tokenHash := auth.HashToken(rawToken)
	ctx := context.Background()
	verifyKey := queue.KeyAuthVerify(tokenHash)

	// Look up the token in Redis
	userID, err := rdb.Get(ctx, verifyKey).Result()
	if err == redis.Nil {
		return ErrTokenExpired
	}
	if err != nil {
		return fmt.Errorf("redis error: %w", err)
	}

	// Atomic: delete the token (single use) and the user tracking key
	pipe := rdb.Pipeline()
	pipe.Del(ctx, verifyKey)
	pipe.Del(ctx, queue.KeyAuthVerifyUser(userID))
	pipe.Exec(ctx)

	// Update user status to ACTIVE
	user, err := repository.GetUserByID(db, userID)
	if err != nil {
		return ErrUserNotFound
	}

	if user.Status == model.UserStatusActive {
		return ErrAlreadyVerified
	}

	now := time.Now()
	user.EmailVerifiedAt = &now
	user.Status = model.UserStatusActive

	if err := repository.UpdateUser(db, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// ResendVerification sends a new verification email to the user.
func ResendVerification(db *gorm.DB, rdb *redis.Client, cfg *config.Config, email string) error {
	user, err := repository.GetUserByEmail(db, email)
	if err != nil {
		// Don't reveal whether the email exists
		return nil
	}

	if user.Status == model.UserStatusActive {
		return ErrAlreadyVerified
	}

	ctx := context.Background()
	verifyUserKey := queue.KeyAuthVerifyUser(user.ID)

	// Rate limit: 60 second cooldown
	oldHash, err := rdb.Get(ctx, verifyUserKey).Result()
	if err == nil {
		// Old token exists - check TTL for rate limiting
		ttl, err := rdb.TTL(ctx, verifyUserKey).Result()
		if err == nil && ttl > 0 {
			remaining := 30*time.Minute - ttl
			if remaining < 60*time.Second {
				return ErrRateLimited
			}
		}
		// Delete old token
		rdb.Del(ctx, queue.KeyAuthVerify(oldHash))
		rdb.Del(ctx, verifyUserKey)
	}

	// Generate new token
	rawToken, tokenHash, err := auth.GenerateVerifyToken()
	if err != nil {
		return fmt.Errorf("failed to generate verify token: %w", err)
	}

	// Store new token in Redis
	pipe := rdb.Pipeline()
	pipe.Set(ctx, queue.KeyAuthVerify(tokenHash), user.ID, 30*time.Minute)
	pipe.Set(ctx, verifyUserKey, tokenHash, 30*time.Minute)
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to store verify token: %w", err)
	}

	// Send email
	if sendErr := auth.SendVerificationEmail(cfg, email, rawToken, cfg.AppBaseURL); sendErr != nil {
		fmt.Printf("Warning: failed to send verification email: %v\n", sendErr)
	}

	return nil
}

// Login authenticates a user and returns access and refresh tokens as cookies.
func Login(db *gorm.DB, rdb *redis.Client, cfg *config.Config, email, password string) (accessToken, refreshToken string, err error) {
	user, err := repository.GetUserByEmail(db, email)
	if err != nil {
		return "", "", ErrInvalidCredentials
	}

	// Check if email is verified
	if user.Status != model.UserStatusActive {
		return "", "", ErrInvalidCredentials
	}

	// Check password
	match, err := auth.CheckPassword(password, user.PasswordHash)
	if err != nil || !match {
		return "", "", ErrInvalidCredentials
	}

	// Generate token pair
	accessTok, _, err := auth.GenerateAccessToken(cfg, user.ID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshTok, refreshJTI, familyID, err := auth.GenerateRefreshToken(cfg, user.ID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Store refresh session in Redis
	refreshTTL := time.Duration(cfg.JWTRefreshTTL) * time.Second
	if err := auth.StoreRefreshSession(rdb, refreshJTI, user.ID, familyID, refreshTTL); err != nil {
		return "", "", fmt.Errorf("failed to store refresh session: %w", err)
	}

	return accessTok, refreshTok, nil
}

// RefreshToken rotates a refresh token and returns a new token pair.
func RefreshToken(rdb *redis.Client, cfg *config.Config, oldRefreshToken string) (newAccessToken, newRefreshToken string, err error) {
	// Parse the old refresh token
	claims, err := auth.ParseRefreshToken(cfg, oldRefreshToken)
	if err != nil {
		return "", "", fmt.Errorf("invalid refresh token: %w", err)
	}

	jti := claims.ID
	userID := claims.Subject
	familyID := claims.FamilyID

	// Check if the JTI is in the deny list
	denied, err := rdb.Exists(context.Background(), queue.KeyAuthDeny(jti)).Result()
	if err != nil {
		return "", "", fmt.Errorf("redis error: %w", err)
	}
	if denied > 0 {
		return "", "", ErrTokenExpired
	}

	// Generate the replacement pair before atomically consuming the old session.
	newAccessTok, _, err := auth.GenerateAccessToken(cfg, userID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshTok, newRefreshJTI, _, err := auth.GenerateRefreshTokenInFamily(cfg, userID, familyID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	refreshTTL := time.Duration(cfg.JWTRefreshTTL) * time.Second
	rotationResult, err := auth.RotateRefreshSession(
		rdb,
		jti,
		newRefreshJTI,
		userID,
		familyID,
		refreshTTL,
	)
	if err != nil {
		return "", "", fmt.Errorf("failed to rotate refresh session: %w", err)
	}

	switch rotationResult {
	case auth.RefreshRotationSucceeded:
		return newAccessTok, newRefreshTok, nil
	case auth.RefreshRotationReuse:
		if err := auth.RevokeTokenFamily(rdb, familyID); err != nil {
			return "", "", fmt.Errorf("failed to revoke reused token family: %w", err)
		}
		return "", "", ErrTokenReuse
	case auth.RefreshRotationExpired, auth.RefreshRotationMismatch:
		return "", "", ErrTokenExpired
	default:
		return "", "", fmt.Errorf("unknown refresh rotation result: %d", rotationResult)
	}
}

// Logout invalidates the current access and refresh tokens.
func Logout(rdb *redis.Client, accessTokenJTI string, accessTTL time.Duration, refreshTokenJTI string) error {
	ctx := context.Background()

	// Delete refresh session
	if refreshTokenJTI != "" {
		rdb.Del(ctx, queue.KeyAuthRefresh(refreshTokenJTI))
	}

	// Add access JTI to deny list with remaining TTL
	if accessTokenJTI != "" && accessTTL > 0 {
		if err := auth.StoreDeniedAccessJTI(rdb, accessTokenJTI, accessTTL); err != nil {
			return fmt.Errorf("failed to deny access token: %w", err)
		}
	}

	return nil
}

// LogoutAll revokes all refresh token families for a user.
func LogoutAll(rdb *redis.Client, userID string) error {
	return auth.RevokeAllUserFamilies(rdb, userID)
}

// ForgotPassword initiates a password reset. Always returns nil to prevent email enumeration.
func ForgotPassword(db *gorm.DB, rdb *redis.Client, cfg *config.Config, email string) error {
	user, err := repository.GetUserByEmail(db, email)
	if err != nil {
		// User not found - silently return nil
		return nil
	}

	// Only send if user is verified
	if user.Status != model.UserStatusActive {
		return nil
	}

	// Generate reset token
	rawToken, tokenHash, err := auth.GenerateResetToken()
	if err != nil {
		return nil // Don't leak errors
	}

	// Store token hash in Redis (TTL 20 minutes)
	ctx := context.Background()
	resetKey := queue.KeyAuthReset(tokenHash)
	if err := rdb.Set(ctx, resetKey, user.ID, 20*time.Minute).Err(); err != nil {
		return nil // Don't leak errors
	}

	// Send email (non-blocking)
	if sendErr := auth.SendResetPasswordEmail(cfg, email, rawToken, cfg.AppBaseURL); sendErr != nil {
		fmt.Printf("Warning: failed to send reset email: %v\n", sendErr)
	}

	return nil
}

// ResetPassword resets a user's password using the raw token from the reset email.
func ResetPassword(db *gorm.DB, rdb *redis.Client, rawToken string, newPassword string) error {
	if err := validatePassword(newPassword); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	tokenHash := auth.HashToken(rawToken)
	ctx := context.Background()
	resetKey := queue.KeyAuthReset(tokenHash)

	// Look up the token in Redis
	userID, err := rdb.Get(ctx, resetKey).Result()
	if err == redis.Nil {
		return ErrTokenExpired
	}
	if err != nil {
		return fmt.Errorf("redis error: %w", err)
	}

	// Atomic: delete the token (single use)
	rdb.Del(ctx, resetKey)

	// Update password
	user, err := repository.GetUserByID(db, userID)
	if err != nil {
		return ErrUserNotFound
	}

	hash, err := auth.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("password hashing failed: %w", err)
	}

	user.PasswordHash = hash
	if err := repository.UpdateUser(db, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	// Revoke ALL refresh token families for this user
	auth.RevokeAllUserFamilies(rdb, userID)

	return nil
}

// GetCurrentUser retrieves the current user by ID.
func GetCurrentUser(db *gorm.DB, userID string) (*model.User, error) {
	user, err := repository.GetUserByID(db, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// ─── Validation helpers ─────────────────────────────────────────────────────

func validateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email is required")
	}
	_, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("invalid email format")
	}
	return nil
}

func validateUsername(username string) error {
	if len(username) < 3 || len(username) > 32 {
		return fmt.Errorf("username must be 3-32 characters")
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	return nil
}
