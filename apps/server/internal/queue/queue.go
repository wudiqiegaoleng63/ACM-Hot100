package queue

import (
	"context"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const runReadCount = 1
const submissionReadCount = 1

// EnsureRunConsumerGroup creates the shared judge worker group from the start of the stream.
func EnsureRunConsumerGroup(ctx context.Context, rdb *redis.Client) error {
	err := rdb.XGroupCreateMkStream(ctx, KeyJudgeRuns(), ConsumerGroup, "0").Err()
	if err != nil && strings.Contains(err.Error(), "BUSYGROUP") {
		return nil
	}
	return err
}

// EnsureSubmissionConsumerGroup creates the shared judge worker group for submissions.
func EnsureSubmissionConsumerGroup(ctx context.Context, rdb *redis.Client) error {
	err := rdb.XGroupCreateMkStream(ctx, KeyJudgeSubmissions(), ConsumerGroup, "0").Err()
	if err != nil && strings.Contains(err.Error(), "BUSYGROUP") {
		return nil
	}
	return err
}

// EnqueueRun adds only the immutable run identifier to the sample-run stream.
func EnqueueRun(ctx context.Context, rdb *redis.Client, runID string) (string, error) {
	return rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: KeyJudgeRuns(),
		Values: map[string]interface{}{"run_id": runID},
	}).Result()
}

// EnqueueSubmission adds only the submission identifier to the submissions stream.
func EnqueueSubmission(ctx context.Context, rdb *redis.Client, submissionID string) (string, error) {
	return rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: KeyJudgeSubmissions(),
		Values: map[string]interface{}{"submission_id": submissionID},
	}).Result()
}

// ReadRuns blocks for the next new run assigned to this consumer.
func ReadRuns(ctx context.Context, rdb *redis.Client, consumer string, block time.Duration) ([]redis.XMessage, error) {
	streams, err := rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    ConsumerGroup,
		Consumer: consumer,
		Streams:  []string{KeyJudgeRuns(), ">"},
		Count:    runReadCount,
		Block:    block,
	}).Result()
	if err != nil {
		return nil, err
	}
	if len(streams) == 0 {
		return nil, nil
	}
	return streams[0].Messages, nil
}

// ReadSubmissions blocks for the next new submission assigned to this consumer.
func ReadSubmissions(ctx context.Context, rdb *redis.Client, consumer string, block time.Duration) ([]redis.XMessage, error) {
	streams, err := rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    ConsumerGroup,
		Consumer: consumer,
		Streams:  []string{KeyJudgeSubmissions(), ">"},
		Count:    submissionReadCount,
		Block:    block,
	}).Result()
	if err != nil {
		return nil, err
	}
	if len(streams) == 0 {
		return nil, nil
	}
	return streams[0].Messages, nil
}

// AckRun acknowledges a durably processed run message.
func AckRun(ctx context.Context, rdb *redis.Client, messageID string) error {
	return rdb.XAck(ctx, KeyJudgeRuns(), ConsumerGroup, messageID).Err()
}

// AckSubmission acknowledges a durably processed submission message.
func AckSubmission(ctx context.Context, rdb *redis.Client, messageID string) error {
	return rdb.XAck(ctx, KeyJudgeSubmissions(), ConsumerGroup, messageID).Err()
}
