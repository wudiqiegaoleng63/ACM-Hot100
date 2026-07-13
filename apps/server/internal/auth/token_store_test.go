package auth

import (
	"errors"
	"sync"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestConsumeOneTimeTokenIsAtomic(t *testing.T) {
	redisServer := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	const tokenKey = "auth:test:token"
	if err := rdb.Set(t.Context(), tokenKey, "user-123", 0).Err(); err != nil {
		t.Fatalf("store token: %v", err)
	}

	const consumers = 16
	start := make(chan struct{})
	results := make(chan error, consumers)
	var wg sync.WaitGroup
	for range consumers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			value, err := ConsumeOneTimeToken(rdb, tokenKey)
			if err == nil && value != "user-123" {
				results <- errors.New("consumer received unexpected user ID")
				return
			}
			results <- err
		}()
	}
	close(start)
	wg.Wait()
	close(results)

	var succeeded, missing int
	for err := range results {
		switch {
		case err == nil:
			succeeded++
		case errors.Is(err, redis.Nil):
			missing++
		default:
			t.Fatalf("unexpected consume error: %v", err)
		}
	}

	if succeeded != 1 || missing != consumers-1 {
		t.Fatalf("successful consumers = %d, missing = %d; want 1 and %d", succeeded, missing, consumers-1)
	}
}
