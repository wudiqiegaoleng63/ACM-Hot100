package repository

import (
	"github.com/acmhot100/server/internal/model"
	"gorm.io/gorm"
)

// GetProgress returns a user's progress for a specific problem, or nil if not found.
func GetProgress(db *gorm.DB, userID, problemID string) (*model.UserProblemProgress, error) {
	var progress model.UserProblemProgress
	if err := db.Where("user_id = ? AND problem_id = ?", userID, problemID).
		First(&progress).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &progress, nil
}

// GetProgressByUser returns all progress records for a user.
func GetProgressByUser(db *gorm.DB, userID string) ([]model.UserProblemProgress, error) {
	var progress []model.UserProblemProgress
	if err := db.Where("user_id = ?", userID).Find(&progress).Error; err != nil {
		return nil, err
	}
	return progress, nil
}
