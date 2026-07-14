package repository

import (
	"time"

	"github.com/acmhot100/server/internal/model"
	"gorm.io/gorm"
)

// CreateSubmission creates a QUEUED submission and updates progress in a single transaction.
// The submission source_code is an immutable snapshot.
func CreateSubmission(db *gorm.DB, submission *model.Submission) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(submission).Error; err != nil {
			return err
		}

		// Upsert progress: set ATTEMPTED and increment attempt_count
		now := time.Now().UTC()
		result := tx.Model(&model.UserProblemProgress{}).
			Where("user_id = ? AND problem_id = ?", submission.UserID, submission.ProblemID).
			Updates(map[string]interface{}{
				"state":            model.ProgressAttempted,
				"attempt_count":    gorm.Expr("attempt_count + 1"),
				"last_submitted_at": now,
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			progress := &model.UserProblemProgress{
				UserID:          submission.UserID,
				ProblemID:       submission.ProblemID,
				State:           model.ProgressAttempted,
				AttemptCount:    1,
				LastSubmittedAt: &now,
			}
			if err := tx.Create(progress).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// ListSubmissions returns paginated submissions for a user with optional filters.
func ListSubmissions(db *gorm.DB, userID string, problemSlug, status, languageKey string, page, pageSize int) ([]model.Submission, int, error) {
	query := db.Model(&model.Submission{}).Where("user_id = ?", userID)

	if problemSlug != "" {
		var problem model.Problem
		if err := db.Where("slug = ?", problemSlug).Select("id").First(&problem).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// No matching problem means no submissions
				return nil, 0, nil
			}
			return nil, 0, err
		}
		query = query.Where("problem_id = ?", problem.ID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if languageKey != "" {
		query = query.Where("language_key = ?", languageKey)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var submissions []model.Submission
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&submissions).Error; err != nil {
		return nil, 0, err
	}
	return submissions, int(total), nil
}

// GetSubmissionForUser returns a submission only when it belongs to the requesting user.
func GetSubmissionForUser(db *gorm.DB, submissionID, userID string) (*model.Submission, error) {
	var submission model.Submission
	if err := db.Preload("CaseResults", func(db *gorm.DB) *gorm.DB {
		return db.Order("case_index ASC")
	}).Where("id = ? AND user_id = ?", submissionID, userID).First(&submission).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &submission, nil
}

// MarkSubmissionEnqueued stores queue metadata after XADD succeeds.
func MarkSubmissionEnqueued(db *gorm.DB, submissionID, messageID string, enqueuedAt time.Time) error {
	return db.Model(&model.Submission{}).
		Where("id = ? AND status = ?", submissionID, model.SubmissionStatusQueued).
		Updates(map[string]interface{}{
			"stream_message_id": messageID,
			"enqueued_at":       enqueuedAt,
		}).Error
}

// FindUnenqueuedSubmissions returns QUEUED submissions missing enqueued_at for reconciliation.
func FindUnenqueuedSubmissions(db *gorm.DB, limit int) ([]model.Submission, error) {
	var submissions []model.Submission
	if err := db.Where("status = ? AND enqueued_at IS NULL", model.SubmissionStatusQueued).
		Order("created_at ASC").
		Limit(limit).
		Find(&submissions).Error; err != nil {
		return nil, err
	}
	return submissions, nil
}
