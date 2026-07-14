package judge

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/acmhot100/server/internal/model"
	"github.com/acmhot100/server/internal/queue"
	"github.com/acmhot100/server/internal/repository"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	defaultSubmissionReadBlock = 5 * time.Second
	submissionLockTTL          = 300 * time.Second // 5 minutes
	maxRetryCount              = 2
)

// SubmissionWorker consumes and judges formal submissions.
type SubmissionWorker struct {
	db       *gorm.DB
	rdb      *redis.Client
	adapter  Adapter
	consumer string
}

// NewSubmissionWorker constructs a submission worker with the given judge adapter.
func NewSubmissionWorker(db *gorm.DB, rdb *redis.Client, consumer string, adapter Adapter) *SubmissionWorker {
	return &SubmissionWorker{
		db:       db,
		rdb:      rdb,
		adapter:  adapter,
		consumer: consumer,
	}
}

// EnsureGroup initializes the Redis consumer group idempotently.
func (w *SubmissionWorker) EnsureGroup(ctx context.Context) error {
	return queue.EnsureSubmissionConsumerGroup(ctx, w.rdb)
}

// Run consumes new messages until the context is canceled.
func (w *SubmissionWorker) Run(ctx context.Context) error {
	for {
		messages, err := queue.ReadSubmissions(ctx, w.rdb, w.consumer, defaultSubmissionReadBlock)
		if err != nil && !errors.Is(err, redis.Nil) {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			return err
		}
		for _, message := range messages {
			if err := w.ProcessMessage(ctx, message); err != nil {
				log.Printf("submission worker: process message %s: %v", message.ID, err)
				// Don't return — continue processing other messages
			}
		}
		if err := w.reclaimPending(ctx); err != nil {
			log.Printf("submission worker: reclaim pending: %v", err)
		}
	}
}

// reclaimPending re-delivers submission messages stuck with a dead consumer past 5 minutes.
// Each recovered submission's retry_count is tracked; after maxRetryCount it is marked SYSTEM_ERROR.
func (w *SubmissionWorker) reclaimPending(ctx context.Context) error {
	messages, err := queue.ClaimPendingSubmissions(ctx, w.rdb, w.consumer)
	if err != nil {
		return err
	}
	for _, message := range messages {
		if err := w.ProcessMessage(ctx, message); err != nil {
			log.Printf("submission worker: reclaim message %s: %v", message.ID, err)
		}
	}
	return nil
}

// ProcessMessage applies idempotent state transitions and ACKs only durable results.
func (w *SubmissionWorker) ProcessMessage(ctx context.Context, message redis.XMessage) error {
	submissionID, ok := message.Values["submission_id"].(string)
	if !ok || submissionID == "" || len(message.Values) != 1 {
		return fmt.Errorf("invalid submission message %s", message.ID)
	}

	// Check if already terminal
	submission, err := repository.GetSubmissionByID(w.db, submissionID)
	if err != nil {
		return err
	}
	if submission == nil || IsTerminalSubmissionStatus(submission.Status) {
		return queue.AckSubmission(ctx, w.rdb, message.ID)
	}

	// Enforce retry limit to avoid infinite re-delivery of a stuck submission.
	if submission.RetryCount >= maxRetryCount && submission.Status == model.SubmissionStatusQueued {
		if err := repository.MarkSubmissionSystemError(w.db, submissionID, "submission exceeded retry limit", time.Now().UTC()); err != nil {
			return err
		}
		return queue.AckSubmission(ctx, w.rdb, message.ID)
	}

	// Acquire distributed lock
	lockKey := queue.KeyJudgeLock(submissionID)
	locked, err := w.rdb.SetNX(ctx, lockKey, w.consumer, submissionLockTTL).Result()
	if err != nil {
		return fmt.Errorf("acquire lock: %w", err)
	}
	if !locked {
		// Another worker has this submission
		return nil
	}
	defer func() {
		_, _ = w.rdb.Del(ctx, lockKey).Result()
	}()

	// A reclaimed message can point at a submission left in an in-flight state by
	// a crashed worker. Return it to QUEUED and count this recovery before claiming it.
	if submission.Status == model.SubmissionStatusCompiling || submission.Status == model.SubmissionStatusRunning {
		if submission.RetryCount >= maxRetryCount {
			if err := repository.MarkSubmissionSystemError(w.db, submissionID, "submission exceeded retry limit", time.Now().UTC()); err != nil {
				return err
			}
			return queue.AckSubmission(ctx, w.rdb, message.ID)
		}
		prepared, err := repository.PrepareSubmissionRetry(w.db, submissionID, maxRetryCount)
		if err != nil {
			return err
		}
		if !prepared {
			return nil
		}
	}

	// Claim: QUEUED → COMPILING
	now := time.Now().UTC()
	claimed, err := repository.ClaimQueuedSubmission(w.db, submissionID, model.SubmissionStatusCompiling, now)
	if err != nil {
		return err
	}
	if !claimed {
		return nil
	}

	// Execute judge
	result, err := w.adapter.Judge(ctx, submissionID)
	if err != nil {
		// Judge execution failed — ACK only after the terminal state is durable.
		if markErr := repository.MarkSubmissionSystemError(w.db, submissionID, fmt.Sprintf("judge execution failed: %v", err), time.Now().UTC()); markErr != nil {
			return markErr
		}
		return queue.AckSubmission(ctx, w.rdb, message.ID)
	}

	// Write result to database in a transaction
	if err := w.completeSubmission(ctx, submissionID, result); err != nil {
		log.Printf("submission worker: complete submission %s: %v", submissionID, err)
		// Don't ACK — will be retried
		return err
	}

	// ACK only after durable result
	return queue.AckSubmission(ctx, w.rdb, message.ID)
}

// completeSubmission writes the judge result, case results, and progress in a single transaction.
func (w *SubmissionWorker) completeSubmission(_ context.Context, submissionID string, result *JudgeResult) error {
	judgedAt := time.Now().UTC()
	return w.db.Transaction(func(tx *gorm.DB) error {
		updated, err := repository.WriteSubmissionResult(tx, submissionID, result.Status,
			result.PassedCases, result.TotalCases, result.TotalTimeMs, result.PeakMemoryKb,
			result.CompilerOutput, result.ErrorMessage, judgedAt)
		if err != nil {
			return err
		}
		if !updated {
			return nil
		}

		caseInputs := make([]repository.CaseResultInput, len(result.CaseResults))
		for i, cr := range result.CaseResults {
			caseInputs[i] = repository.CaseResultInput{
				CaseIndex:    i,
				Status:       cr.Status,
				TimeMs:       &cr.TimeMs,
				MemoryKb:     &cr.MemoryKb,
				ActualOutput: cr.ActualOutput,
				IsSample:     false,
			}
		}
		if err := repository.WriteCaseResults(tx, submissionID, caseInputs); err != nil {
			return err
		}

		var submission model.Submission
		if err := tx.Where("id = ?", submissionID).Select("user_id", "problem_id").First(&submission).Error; err != nil {
			return err
		}

		if result.Status == model.SubmissionStatusAccepted {
			return repository.SetProgressSolved(tx, submission.UserID, submission.ProblemID, judgedAt)
		}
		repository.SetProgressAttemptedIfNotStarted(tx, submission.UserID, submission.ProblemID)
		return nil
	})
}
