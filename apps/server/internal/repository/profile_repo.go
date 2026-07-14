package repository

import (
	"github.com/acmhot100/server/internal/model"
	"gorm.io/gorm"
)

// ProfileProgressSummary aggregates the authenticated user's progress over published problems.
type ProfileProgressSummary struct {
	TotalProblems int `json:"total_problems"`
	Solved        int `json:"solved"`
	Attempted     int `json:"attempted"`
	NotStarted    int `json:"not_started"`
}

// StageProgress aggregates progress for one training stage.
type StageProgress struct {
	Stage      string `json:"stage"`
	Total      int    `json:"total"`
	Solved     int    `json:"solved"`
	Attempted  int    `json:"attempted"`
	NotStarted int    `json:"not_started"`
}

// GetProfileProgressSummary returns three-state progress counts for published problems.
func GetProfileProgressSummary(db *gorm.DB, userID string) (ProfileProgressSummary, error) {
	var summary ProfileProgressSummary
	err := profileProgressQuery(db, userID).
		Select(`COUNT(*) AS total_problems,
			COALESCE(SUM(CASE WHEN upp.state = ? THEN 1 ELSE 0 END), 0) AS solved,
			COALESCE(SUM(CASE WHEN upp.state = ? THEN 1 ELSE 0 END), 0) AS attempted,
			COALESCE(SUM(CASE WHEN upp.state IS NULL OR upp.state = ? THEN 1 ELSE 0 END), 0) AS not_started`,
			model.ProgressSolved, model.ProgressAttempted, model.ProgressNotStarted).
		Scan(&summary).Error
	return summary, err
}

// GetProfileProgressByStage returns progress grouped by stage in problem order.
func GetProfileProgressByStage(db *gorm.DB, userID string) ([]StageProgress, error) {
	var stages []StageProgress
	err := profileProgressQuery(db, userID).
		Select(`problems.stage AS stage,
			COUNT(*) AS total,
			COALESCE(SUM(CASE WHEN upp.state = ? THEN 1 ELSE 0 END), 0) AS solved,
			COALESCE(SUM(CASE WHEN upp.state = ? THEN 1 ELSE 0 END), 0) AS attempted,
			COALESCE(SUM(CASE WHEN upp.state IS NULL OR upp.state = ? THEN 1 ELSE 0 END), 0) AS not_started`,
			model.ProgressSolved, model.ProgressAttempted, model.ProgressNotStarted).
		Group("problems.stage").
		Order("MIN(problems.order_index) ASC").
		Scan(&stages).Error
	return stages, err
}

func profileProgressQuery(db *gorm.DB, userID string) *gorm.DB {
	return db.Table("problems").
		Joins("LEFT JOIN user_problem_progress AS upp ON upp.problem_id = problems.id AND upp.user_id = ?", userID).
		Where("problems.is_published = ?", true)
}
