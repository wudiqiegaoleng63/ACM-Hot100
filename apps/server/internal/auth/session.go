package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/acmhot100/server/internal/queue"
)

// StoreRefreshSession stores a refresh token session in Redis.
// The key is the JTI, and the value is "userID:familyID".
func StoreRefreshSession(rdb *redis.Client, jti, userID, familyID string, ttl time.Duration) error {
	ctx := context.Background()
	val := fmt.Sprintf("%s:%s", userID, familyID)

	// Store the session keyed by JTI
	if err := rdb.Set(ctx, queue.KeyAuthRefresh(jti), val, ttl).Err(); err != nil {
		return err
	}

	// Add JTI to the family set
	if err := rdb.SAdd(ctx, queue.KeyAuthFamily(familyID), jti).Err(); err != nil {
		return err
	}

	// Set TTL on the family set to match the refresh TTL
	rdb.Expire(ctx, queue.KeyAuthFamily(familyID), ttl)

	// Track user -> families mapping for RevokeAllUserFamilies
	if err := rdb.SAdd(ctx, queue.KeyAuthUserFamilies(userID), familyID).Err(); err != nil {
		return err
	}

	return nil
}

// GetRefreshSession retrieves the userID and familyID for a refresh token JTI.
func GetRefreshSession(rdb *redis.Client, jti string) (userID, familyID string, err error) {
	ctx := context.Background()
	val, err := rdb.Get(ctx, queue.KeyAuthRefresh(jti)).Result()
	if err != nil {
		return "", "", err
	}

	// Parse "userID:familyID"
	parts := splitTwo(val, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid session data")
	}
	return parts[0], parts[1], nil
}

// DeleteRefreshSession removes a single refresh token session.
func DeleteRefreshSession(rdb *redis.Client, jti string) error {
	ctx := context.Background()
	return rdb.Del(ctx, queue.KeyAuthRefresh(jti)).Err()
}

// RevokeTokenFamily revokes all refresh tokens in a family by deleting
// the family set and all individual session keys.
func RevokeTokenFamily(rdb *redis.Client, familyID string) error {
	ctx := context.Background()
	familyKey := queue.KeyAuthFamily(familyID)

	// Get all JTIs in the family
	jtis, err := rdb.SMembers(ctx, familyKey).Result()
	if err != nil {
		return err
	}

	// Delete all individual session keys and the family set
	if len(jtis) > 0 {
		keys := make([]string, len(jtis))
		for i, jti := range jtis {
			keys[i] = queue.KeyAuthRefresh(jti)
		}
		rdb.Del(ctx, keys...)
	}

	rdb.Del(ctx, familyKey)
	return nil
}

// RevokeAllUserFamilies revokes all refresh token families for a user.
func RevokeAllUserFamilies(rdb *redis.Client, userID string) error {
	ctx := context.Background()
	userFamiliesKey := queue.KeyAuthUserFamilies(userID)

	// Get all family IDs for this user
	familyIDs, err := rdb.SMembers(ctx, userFamiliesKey).Result()
	if err != nil {
		return err
	}

	// Revoke each family
	for _, familyID := range familyIDs {
		if err := RevokeTokenFamily(rdb, familyID); err != nil {
			return err
		}
	}

	// Clean up the user families set
	rdb.Del(ctx, userFamiliesKey)
	return nil
}

// DetectTokenReuse checks if a refresh token JTI has already been consumed.
// If the JTI is not in the session store but the family still exists,
// this indicates token reuse (theft), and the entire family should be revoked.
func DetectTokenReuse(rdb *redis.Client, jti, familyID string) (bool, error) {
	ctx := context.Background()

	// Check if the JTI session exists
	_, err := rdb.Get(ctx, queue.KeyAuthRefresh(jti)).Result()
	if err == redis.Nil {
		// JTI not found - check if the family still exists
		exists, err := rdb.Exists(ctx, queue.KeyAuthFamily(familyID)).Result()
		if err != nil {
			return false, err
		}
		if exists > 0 {
			// Family exists but JTI doesn't = reuse detected
			return true, nil
		}
		// Family also gone - token is simply expired
		return false, nil
	}
	if err != nil {
		return false, err
	}

	// JTI found - not reused
	return false, nil
}

// StoreDeniedAccessJTI adds an access token JTI to the deny list (for logout).
func StoreDeniedAccessJTI(rdb *redis.Client, jti string, ttl time.Duration) error {
	ctx := context.Background()
	return rdb.Set(ctx, queue.KeyAuthDeny(jti), "1", ttl).Err()
}

// IsAccessDenied checks if an access token JTI is in the deny list.
func IsAccessDenied(rdb *redis.Client, jti string) (bool, error) {
	ctx := context.Background()
	val, err := rdb.Exists(ctx, queue.KeyAuthDeny(jti)).Result()
	if err != nil {
		return false, err
	}
	return val > 0, nil
}

// splitTwo splits a string by the first occurrence of sep.
func splitTwo(s, sep string) []string {
	for i := 0; i < len(s)-len(sep)+1; i++ {
		if s[i:i+len(sep)] == sep {
			return []string{s[:i], s[i+len(sep):]}
		}
	}
	return []string{s}
}
