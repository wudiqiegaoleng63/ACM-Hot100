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

// RefreshRotationResult describes the outcome of an atomic refresh-token rotation.
type RefreshRotationResult int64

const (
	RefreshRotationSucceeded RefreshRotationResult = 1
	RefreshRotationReuse     RefreshRotationResult = 2
	RefreshRotationExpired   RefreshRotationResult = 3
	RefreshRotationMismatch  RefreshRotationResult = 4
)

var rotateRefreshSessionScript = redis.NewScript(`
local current = redis.call('GET', KEYS[1])
if not current then
  if redis.call('EXISTS', KEYS[2]) == 1 then
    return 2
  end
  return 3
end

if current ~= ARGV[2] then
  return 4
end

redis.call('DEL', KEYS[1])
redis.call('SREM', KEYS[2], ARGV[1])
redis.call('SET', KEYS[3], ARGV[3], 'PX', ARGV[5])
redis.call('SADD', KEYS[2], ARGV[4])
redis.call('PEXPIRE', KEYS[2], ARGV[5])
return 1
`)

// RotateRefreshSession atomically consumes an old refresh session and stores its replacement.
func RotateRefreshSession(rdb *redis.Client, oldJTI, newJTI, userID, familyID string, ttl time.Duration) (RefreshRotationResult, error) {
	ctx := context.Background()
	expectedValue := fmt.Sprintf("%s:%s", userID, familyID)
	result, err := rotateRefreshSessionScript.Run(
		ctx,
		rdb,
		[]string{
			queue.KeyAuthRefresh(oldJTI),
			queue.KeyAuthFamily(familyID),
			queue.KeyAuthRefresh(newJTI),
		},
		oldJTI,
		expectedValue,
		expectedValue,
		newJTI,
		ttl.Milliseconds(),
	).Int64()
	if err != nil {
		return 0, err
	}
	return RefreshRotationResult(result), nil
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

var revokeTokenFamilyScript = redis.NewScript(`
local members = redis.call('SMEMBERS', KEYS[1])
for _, member in ipairs(members) do
  redis.call('DEL', ARGV[1] .. member)
end
redis.call('DEL', KEYS[1])
return #members
`)

// RevokeTokenFamily atomically deletes a family and all of its refresh sessions.
func RevokeTokenFamily(rdb *redis.Client, familyID string) error {
	ctx := context.Background()
	refreshKeyPrefix := queue.KeyAuthRefresh("")
	return revokeTokenFamilyScript.Run(
		ctx,
		rdb,
		[]string{queue.KeyAuthFamily(familyID)},
		refreshKeyPrefix,
	).Err()
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
