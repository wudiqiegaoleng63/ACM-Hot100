package repository

import (
	"time"

	"github.com/acmhot100/server/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CreateSubmission creates a QUEUED submission and updates progress in a single transaction.
// The submission source_code is an immutable snapshot.
func CreateSubmission(db *gorm.DB, submission *model.Submission) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(submission).Error; err != nil {
			return err
		}

		// Upsert progress without downgrading an already solved problem.
		now := time.Now().UTC()
		result := tx.Model(&model.UserProblemProgress{}).
			Where("user_id = ? AND problem_id = ?", submission.UserID, submission.ProblemID).
			Updates(map[string]interface{}{
				"state":             gorm.Expr("CASE WHEN state = ? THEN state ELSE ? END", model.ProgressSolved, model.ProgressAttempted),
				"attempt_count":     gorm.Expr("attempt_count + 1"),
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

// ClaimQueuedSubmission atomically moves a submission from QUEUED to the given status.
func ClaimQueuedSubmission(db *gorm.DB, submissionID, nextStatus string, claimedAt time.Time) (bool, error) {
	result := db.Model(&model.Submission{}).
		Where("id = ? AND status = ?", submissionID, model.SubmissionStatusQueued).
		Updates(map[string]interface{}{
			"status":     nextStatus,
			"claimed_at": claimedAt,
		})
	return result.RowsAffected == 1, result.Error
}

// PrepareSubmissionRetry atomically returns a crashed in-flight submission to QUEUED
// and records one recovery attempt.
func PrepareSubmissionRetry(db *gorm.DB, submissionID string, maxRetries int) (bool, error) {
	result := db.Model(&model.Submission{}).
		Where("id = ? AND status IN ? AND retry_count < ?", submissionID, []string{
			model.SubmissionStatusCompiling,
			model.SubmissionStatusRunning,
		}, maxRetries).
		Updates(map[string]interface{}{
			"status":      model.SubmissionStatusQueued,
			"claimed_at":  nil,
			"retry_count": gorm.Expr("retry_count + 1"),
		})
	return result.RowsAffected == 1, result.Error
}

// GetSubmissionByID returns a submission by ID for worker-side state inspection.
func GetSubmissionByID(db *gorm.DB, submissionID string) (*model.Submission, error) {
	var submission model.Submission
	if err := db.Where("id = ?", submissionID).First(&submission).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &submission, nil
}

// WriteSubmissionResult updates the submission row with the final judge result fields.
func WriteSubmissionResult(db *gorm.DB, submissionID, status string, passedCases, totalCases, totalTimeMs, peakMemoryKb int, compilerOutput, errorMessage string, judgedAt time.Time) (bool, error) {
	result := db.Model(&model.Submission{}).
		Where("id = ? AND status IN ?", submissionID, []string{model.SubmissionStatusCompiling, model.SubmissionStatusRunning}).
		Updates(map[string]interface{}{
			"status":          status,
			"passed_cases":    passedCases,
			"total_cases":     totalCases,
			"time_ms":         totalTimeMs,
			"memory_kb":       peakMemoryKb,
			"compiler_output": compilerOutput,
			"error_message":   errorMessage,
			"judged_at":       judgedAt,
		})
	return result.RowsAffected == 1, result.Error
}

// CaseResultInput is a plain data struct for writing case results without importing judge.
type CaseResultInput struct {
	CaseIndex    int
	Status       string
	TimeMs       *int
	MemoryKb     *int
	ActualOutput string
	IsSample     bool
}

// WriteCaseResults creates case result rows for a submission.
func WriteCaseResults(db *gorm.DB, submissionID string, caseResults []CaseResultInput) error {
	for _, cr := range caseResults {
		caseResult := &model.SubmissionCaseResult{
			ID:           uuid.New().String(),
			SubmissionID: submissionID,
			CaseIndex:    cr.CaseIndex,
			Status:       cr.Status,
			TimeMs:       cr.TimeMs,
			MemoryKb:     cr.MemoryKb,
			ActualOutput: cr.ActualOutput,
			IsSample:     cr.IsSample,
		}
		if err := db.Create(caseResult).Error; err != nil {
			return err
		}
	}
	return nil
}

// SetProgressSolved sets the user's progress to SOLVED (never downgrades).
func SetProgressSolved(db *gorm.DB, userID, problemID string, firstACAt time.Time) error {
	result := db.Model(&model.UserProblemProgress{}).
		Where("user_id = ? AND problem_id = ? AND state != ?", userID, problemID, model.ProgressSolved).
		Updates(map[string]interface{}{
			"state":       model.ProgressSolved,
			"first_ac_at": firstACAt,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		progress, err := GetProgress(db, userID, problemID)
		if err != nil {
			return err
		}
		if progress != nil && progress.State == model.ProgressSolved {
			return nil
		}
		progress = &model.UserProblemProgress{
			UserID:    userID,
			ProblemID: problemID,
			State:     model.ProgressSolved,
			FirstACAt: &firstACAt,
		}
		return db.Create(progress).Error
	}
	return nil
}

// SetProgressAttemptedIfNotStarted sets progress to ATTEMPTED only if currently NOT_STARTED.
func SetProgressAttemptedIfNotStarted(db *gorm.DB, userID, problemID string) {
	db.Model(&model.UserProblemProgress{}).
		Where("user_id = ? AND problem_id = ? AND state = ?", userID, problemID, model.ProgressNotStarted).
		Update("state", model.ProgressAttempted)
}

// MarkSubmissionSystemError sets a submission to SYSTEM_ERROR state.
func MarkSubmissionSystemError(db *gorm.DB, submissionID, message string, judgedAt time.Time) error {
	return db.Model(&model.Submission{}).
		Where("id = ? AND status NOT IN ?", submissionID, []string{
			model.SubmissionStatusAccepted, model.SubmissionStatusWrongAnswer,
			model.SubmissionStatusTimeLimit, model.SubmissionStatusMemoryLimit,
			model.SubmissionStatusRuntimeError, model.SubmissionStatusCompileError,
			model.SubmissionStatusSystemError,
		}).
		Updates(map[string]interface{}{
			"status":        model.SubmissionStatusSystemError,
			"error_message": message,
			"judged_at":     judgedAt,
		}).Error
}

// IncrementSubmissionRetryCount atomically increments the retry count.
func IncrementSubmissionRetryCount(db *gorm.DB, submissionID string) error {
	return db.Model(&model.Submission{}).
		Where("id = ?", submissionID).
		Update("retry_count", gorm.Expr("retry_count + 1")).Error
}
