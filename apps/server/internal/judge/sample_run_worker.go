package judge

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/acmhot100/server/internal/model"
	"github.com/acmhot100/server/internal/queue"
	"github.com/acmhot100/server/internal/repository"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	defaultRunReadBlock = time.Second
	defaultMockDelay    = 750 * time.Millisecond
)

// SampleRunWorker consumes and completes asynchronous sample runs.
type SampleRunWorker struct {
	db       *gorm.DB
	rdb      *redis.Client
	consumer string
	delay    time.Duration
	sleep    func(context.Context, time.Duration) error
}

// NewSampleRunWorker constructs a mock sample-run worker.
func NewSampleRunWorker(db *gorm.DB, rdb *redis.Client, consumer string) *SampleRunWorker {
	return &SampleRunWorker{
		db:       db,
		rdb:      rdb,
		consumer: consumer,
		delay:    defaultMockDelay,
		sleep:    sleepContext,
	}
}

// EnsureGroup initializes the Redis consumer group idempotently.
func (w *SampleRunWorker) EnsureGroup(ctx context.Context) error {
	return queue.EnsureRunConsumerGroup(ctx, w.rdb)
}

// Run consumes new messages until the context is canceled.
func (w *SampleRunWorker) Run(ctx context.Context) error {
	for {
		messages, err := queue.ReadRuns(ctx, w.rdb, w.consumer, defaultRunReadBlock)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			if errors.Is(err, redis.Nil) {
				continue
			}
			return err
		}
		for _, message := range messages {
			if err := w.ProcessMessage(ctx, message); err != nil {
				return err
			}
		}
	}
}

// ProcessMessage applies idempotent state transitions and ACKs only durable results.
func (w *SampleRunWorker) ProcessMessage(ctx context.Context, message redis.XMessage) error {
	runID, ok := message.Values["run_id"].(string)
	if !ok || runID == "" || len(message.Values) != 1 {
		return fmt.Errorf("invalid sample run message %s", message.ID)
	}

	run, err := repository.GetSampleRun(w.db, runID)
	if err != nil {
		return err
	}
	if run == nil || isTerminalRunStatus(run.Status) {
		return queue.AckRun(ctx, w.rdb, message.ID)
	}

	claimed, err := repository.ClaimQueuedSampleRun(w.db, runID, time.Now().UTC())
	if err != nil {
		return err
	}
	if !claimed {
		return nil
	}
	if err := w.sleep(ctx, w.delay); err != nil {
		return err
	}
	completed, err := repository.CompleteSampleRun(w.db, runID, "", time.Now().UTC())
	if err != nil {
		return err
	}
	if !completed {
		return nil
	}
	return queue.AckRun(ctx, w.rdb, message.ID)
}

func isTerminalRunStatus(status string) bool {
	return status == model.SampleRunStatusAccepted || status == model.SampleRunStatusSystemError
}

func sleepContext(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
