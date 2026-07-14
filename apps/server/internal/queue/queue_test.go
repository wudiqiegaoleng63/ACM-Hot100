package queue

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestRunStreamConsumerGroupRoundTrip(t *testing.T) {
	server := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	SetPrefix("queue-test:")
	t.Cleanup(func() { SetPrefix("") })
	ctx := context.Background()

	if err := EnsureRunConsumerGroup(ctx, rdb); err != nil {
		t.Fatalf("EnsureRunConsumerGroup first call: %v", err)
	}
	if err := EnsureRunConsumerGroup(ctx, rdb); err != nil {
		t.Fatalf("EnsureRunConsumerGroup second call: %v", err)
	}
	messageID, err := EnqueueRun(ctx, rdb, "run-1")
	if err != nil {
		t.Fatalf("EnqueueRun: %v", err)
	}
	entries, err := rdb.XRange(ctx, KeyJudgeRuns(), "-", "+").Result()
	if err != nil {
		t.Fatalf("XRange: %v", err)
	}
	if len(entries) != 1 || len(entries[0].Values) != 1 || entries[0].Values["run_id"] != "run-1" {
		t.Fatalf("stream entries = %#v, want only run_id", entries)
	}

	messages, err := ReadRuns(ctx, rdb, "consumer-1", 10*time.Millisecond)
	if err != nil {
		t.Fatalf("ReadRuns: %v", err)
	}
	if len(messages) != 1 || messages[0].ID != messageID || messages[0].Values["run_id"] != "run-1" {
		t.Fatalf("messages = %#v, want run-1", messages)
	}
	if err := AckRun(ctx, rdb, messageID); err != nil {
		t.Fatalf("AckRun: %v", err)
	}
	pending, err := rdb.XPending(ctx, KeyJudgeRuns(), ConsumerGroup).Result()
	if err != nil {
		t.Fatalf("XPending: %v", err)
	}
	if pending.Count != 0 {
		t.Fatalf("pending count = %d, want 0", pending.Count)
	}
}

func TestSubmissionStreamConsumerGroupRoundTrip(t *testing.T) {
	server := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	SetPrefix("queue-test:")
	t.Cleanup(func() { SetPrefix("") })
	ctx := context.Background()

	if err := EnsureSubmissionConsumerGroup(ctx, rdb); err != nil {
		t.Fatalf("EnsureSubmissionConsumerGroup first call: %v", err)
	}
	if err := EnsureSubmissionConsumerGroup(ctx, rdb); err != nil {
		t.Fatalf("EnsureSubmissionConsumerGroup second call: %v", err)
	}
	messageID, err := EnqueueSubmission(ctx, rdb, "sub-1")
	if err != nil {
		t.Fatalf("EnqueueSubmission: %v", err)
	}
	entries, err := rdb.XRange(ctx, KeyJudgeSubmissions(), "-", "+").Result()
	if err != nil {
		t.Fatalf("XRange: %v", err)
	}
	if len(entries) != 1 || len(entries[0].Values) != 1 || entries[0].Values["submission_id"] != "sub-1" {
		t.Fatalf("stream entries = %#v, want only submission_id", entries)
	}

	messages, err := ReadSubmissions(ctx, rdb, "consumer-1", 10*time.Millisecond)
	if err != nil {
		t.Fatalf("ReadSubmissions: %v", err)
	}
	if len(messages) != 1 || messages[0].ID != messageID || messages[0].Values["submission_id"] != "sub-1" {
		t.Fatalf("messages = %#v, want sub-1", messages)
	}
	if err := AckSubmission(ctx, rdb, messageID); err != nil {
		t.Fatalf("AckSubmission: %v", err)
	}
	pending, err := rdb.XPending(ctx, KeyJudgeSubmissions(), ConsumerGroup).Result()
	if err != nil {
		t.Fatalf("XPending: %v", err)
	}
	if pending.Count != 0 {
		t.Fatalf("pending count = %d, want 0", pending.Count)
	}
}
