package queue

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/acmhot100/server/internal/repository"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	defaultReconcileBatchSize = 100
	defaultReconcileInterval  = 30 * time.Second
)

// SubmissionReconciler repairs QUEUED MySQL submissions that were not durably
// associated with a Redis stream message after creation.
type SubmissionReconciler struct {
	db        *gorm.DB
	rdb       *redis.Client
	batchSize int
}

// NewSubmissionReconciler constructs a submission enqueue reconciler.
func NewSubmissionReconciler(db *gorm.DB, rdb *redis.Client) *SubmissionReconciler {
	return &SubmissionReconciler{db: db, rdb: rdb, batchSize: defaultReconcileBatchSize}
}

// Run periodically reconciles until the context is canceled.
func (r *SubmissionReconciler) Run(ctx context.Context) error {
	ticker := time.NewTicker(defaultReconcileInterval)
	defer ticker.Stop()
	for {
		if err := r.ReconcileOnce(ctx); err != nil {
			log.Printf("submission reconciler: %v", err)
		}
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

// ReconcileOnce scans one bounded batch and repairs each missing enqueue.
func (r *SubmissionReconciler) ReconcileOnce(ctx context.Context) error {
	submissions, err := repository.FindUnenqueuedSubmissions(r.db, r.batchSize)
	if err != nil {
		return fmt.Errorf("find unenqueued submissions: %w", err)
	}
	for _, submission := range submissions {
		messageID, err := EnqueueSubmission(ctx, r.rdb, submission.ID)
		if err != nil {
			return fmt.Errorf("enqueue submission %s: %w", submission.ID, err)
		}
		if err := repository.MarkSubmissionEnqueued(r.db, submission.ID, messageID, time.Now().UTC()); err != nil {
			return fmt.Errorf("mark submission %s enqueued: %w", submission.ID, err)
		}
	}
	return nil
}
