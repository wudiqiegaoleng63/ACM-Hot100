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
		if err != nil && !errors.Is(err, redis.Nil) {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			return err
		}
		for _, message := range messages {
			if err := w.ProcessMessage(ctx, message); err != nil {
				return err
			}
		}
		if err := w.reclaimPending(ctx); err != nil {
			return err
		}
	}
}

// reclaimPending re-delivers sample-run messages stuck with a dead consumer past 5 minutes.
func (w *SampleRunWorker) reclaimPending(ctx context.Context) error {
	messages, err := queue.ClaimPendingRuns(ctx, w.rdb, w.consumer)
	if err != nil {
		return err
	}
	for _, message := range messages {
		if err := w.ProcessMessage(ctx, message); err != nil {
			return err
		}
	}
	return nil
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

	if run.Status == model.SampleRunStatusQueued {
		claimed, err := repository.ClaimQueuedSampleRun(w.db, runID, time.Now().UTC())
		if err != nil {
			return err
		}
		if !claimed {
			return nil
		}
	}
	// A reclaimed RUNNING row represents work interrupted after the durable claim;
	// mock execution is deterministic, so it is safe to finish it again.
	if err := w.sleep(ctx, w.delay); err != nil {
		return err
	}
	completed, err := repository.CompleteSampleRun(w.db, runID, "", time.Now().UTC())
	if err != nil {
		return err
	}
	if !completed {
		latest, getErr := repository.GetSampleRun(w.db, runID)
		if getErr != nil {
			return getErr
		}
		if latest == nil || isTerminalRunStatus(latest.Status) {
			return queue.AckRun(ctx, w.rdb, message.ID)
		}
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
