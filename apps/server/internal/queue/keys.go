package queue

import "fmt"

// ─── Redis key prefix ───────────────────────────────────────────────────────

// prefix is the Redis key prefix (e.g., "acmhot100:")
var prefix string

// SetPrefix sets the global Redis key prefix used by all key helpers.
func SetPrefix(p string) {
	prefix = p
}

// K builds a fully-qualified Redis key with the configured prefix.
func K(key string) string {
	return prefix + key
}

// ─── Auth keys ───────────────────────────────────────────────────────────────

// KeyAuthVerify returns the key for email verification tokens.
// Usage: KeyAuthVerify(token) -> "{prefix}auth:verify:{token}"
func KeyAuthVerify(token string) string {
	return K(fmt.Sprintf("auth:verify:%s", token))
}

// KeyAuthVerifyUser returns the key tracking which user a verify token belongs to.
// Usage: KeyAuthVerifyUser(userID) -> "{prefix}auth:verify:user:{userID}"
func KeyAuthVerifyUser(userID string) string {
	return K(fmt.Sprintf("auth:verify:user:%s", userID))
}

// KeyAuthReset returns the key for password reset tokens.
// Usage: KeyAuthReset(token) -> "{prefix}auth:reset:{token}"
func KeyAuthReset(token string) string {
	return K(fmt.Sprintf("auth:reset:%s", token))
}

// KeyAuthRefresh returns the key for refresh token validation.
// Usage: KeyAuthRefresh(tokenID) -> "{prefix}auth:refresh:{tokenID}"
func KeyAuthRefresh(tokenID string) string {
	return K(fmt.Sprintf("auth:refresh:%s", tokenID))
}

// KeyAuthFamily returns the key for token family tracking (refresh token rotation).
// Usage: KeyAuthFamily(familyID) -> "{prefix}auth:family:{familyID}"
func KeyAuthFamily(familyID string) string {
	return K(fmt.Sprintf("auth:family:%s", familyID))
}

// KeyAuthDeny returns the key for denied/blacklisted tokens.
// Usage: KeyAuthDeny(tokenID) -> "{prefix}auth:deny:{tokenID}"
func KeyAuthDeny(tokenID string) string {
	return K(fmt.Sprintf("auth:deny:%s", tokenID))
}

// KeyAuthUserFamilies returns the key tracking all token families for a user.
// Usage: KeyAuthUserFamilies(userID) -> "{prefix}auth:user_families:{userID}"
func KeyAuthUserFamilies(userID string) string {
	return K(fmt.Sprintf("auth:user_families:%s", userID))
}

// ─── Rate limiting keys ─────────────────────────────────────────────────────

// KeyRate returns the key for rate limiting.
// Usage: KeyRate(identifier) -> "{prefix}rate:{identifier}"
func KeyRate(identifier string) string {
	return K(fmt.Sprintf("rate:%s", identifier))
}

// ─── Judge queue keys ───────────────────────────────────────────────────────

const (
	// StreamName is the Redis stream name for judge submissions.
	StreamName = "judge:submissions"

	// ConsumerGroup is the Redis consumer group name for judge workers.
	ConsumerGroup = "judge-workers"
)

// KeyJudgeSubmissions returns the key for the judge submissions stream.
func KeyJudgeSubmissions() string {
	return K(StreamName)
}

// KeyJudgeLock returns the key for judge submission locks.
// Usage: KeyJudgeLock(submissionID) -> "{prefix}judge:lock:{submissionID}"
func KeyJudgeLock(submissionID string) string {
	return K(fmt.Sprintf("judge:lock:%s", submissionID))
}

// ─── Queue operations ───────────────────────────────────────────────────────

// RedisClient is an interface abstracting the Redis operations needed by the queue.
type RedisClient interface {
	XAdd(stream string, values map[string]interface{}) error
	XReadGroup(group string, consumer string, count int64, blockMs int64) ([]StreamMessage, error)
}

// StreamMessage represents a message read from a Redis stream.
type StreamMessage struct {
	ID     string
	Fields map[string]interface{}
}

// EnqueueSubmission adds a submission event to the judge queue stream.
// This is a placeholder implementation.
func EnqueueSubmission(rdb RedisClient, submissionID string, payload map[string]interface{}) error {
	return rdb.XAdd(StreamName, payload)
}

// ReadSubmissions reads submission events from the judge queue stream
// using a consumer group. This is a placeholder implementation.
func ReadSubmissions(rdb RedisClient, consumerGroup string, consumerName string, count int64, blockMs int64) ([]StreamMessage, error) {
	return rdb.XReadGroup(consumerGroup, consumerName, count, blockMs)
}
