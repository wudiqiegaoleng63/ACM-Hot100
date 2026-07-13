package repository

import (
	"time"

	"github.com/acmhot100/server/internal/model"
	"gorm.io/gorm"
)

// CreateSampleRun persists an immutable sample-run snapshot.
func CreateSampleRun(db *gorm.DB, run *model.SampleRun) error {
	return db.Create(run).Error
}

// GetSampleRunForUser returns a run only when it belongs to the requesting user.
func GetSampleRunForUser(db *gorm.DB, runID, userID string) (*model.SampleRun, error) {
	var run model.SampleRun
	if err := db.Where("id = ? AND user_id = ?", runID, userID).First(&run).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &run, nil
}

// GetSampleRun returns a run by ID for worker-side state inspection.
func GetSampleRun(db *gorm.DB, runID string) (*model.SampleRun, error) {
	var run model.SampleRun
	if err := db.Where("id = ?", runID).First(&run).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &run, nil
}

// ClaimQueuedSampleRun atomically moves a run from QUEUED to RUNNING.
func ClaimQueuedSampleRun(db *gorm.DB, runID string, startedAt time.Time) (bool, error) {
	result := db.Model(&model.SampleRun{}).
		Where("id = ? AND status = ?", runID, model.SampleRunStatusQueued).
		Updates(map[string]interface{}{
			"status":     model.SampleRunStatusRunning,
			"started_at": startedAt,
		})
	return result.RowsAffected == 1, result.Error
}

// CompleteSampleRun atomically records a terminal mock result from RUNNING.
func CompleteSampleRun(db *gorm.DB, runID, output string, finishedAt time.Time) (bool, error) {
	result := db.Model(&model.SampleRun{}).
		Where("id = ? AND status = ?", runID, model.SampleRunStatusRunning).
		Updates(map[string]interface{}{
			"status":        model.SampleRunStatusAccepted,
			"output_data":   output,
			"error_message": "",
			"finished_at":   finishedAt,
		})
	return result.RowsAffected == 1, result.Error
}

// MarkSampleRunEnqueued stores queue metadata after XADD succeeds.
func MarkSampleRunEnqueued(db *gorm.DB, runID, messageID string, enqueuedAt time.Time) error {
	return db.Model(&model.SampleRun{}).
		Where("id = ? AND status = ?", runID, model.SampleRunStatusQueued).
		Updates(map[string]interface{}{
			"stream_message_id": messageID,
			"enqueued_at":       enqueuedAt,
		}).Error
}

// MarkSampleRunSystemError prevents an enqueue failure from leaving an eternal QUEUED row.
func MarkSampleRunSystemError(db *gorm.DB, runID, message string, finishedAt time.Time) error {
	return db.Model(&model.SampleRun{}).
		Where("id = ? AND status = ?", runID, model.SampleRunStatusQueued).
		Updates(map[string]interface{}{
			"status":        model.SampleRunStatusSystemError,
			"error_message": message,
			"finished_at":   finishedAt,
		}).Error
}
