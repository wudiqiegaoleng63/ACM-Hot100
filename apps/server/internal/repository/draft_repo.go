package repository

import (
	"github.com/acmhot100/server/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GetDraft returns the user's draft for a problem+language, or nil if not found.
func GetDraft(db *gorm.DB, userID, problemID, languageKey string) (*model.Draft, error) {
	var draft model.Draft
	if err := db.Where("user_id = ? AND problem_id = ? AND language_key = ?", userID, problemID, languageKey).
		First(&draft).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &draft, nil
}

// UpsertDraft creates or updates a draft using ON DUPLICATE KEY UPDATE.
func UpsertDraft(db *gorm.DB, userID, problemID, languageKey, sourceCode string) error {
	draft := model.Draft{
		UserID:      userID,
		ProblemID:   problemID,
		LanguageKey: languageKey,
		SourceCode:  sourceCode,
	}

	return db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "user_id"},
			{Name: "problem_id"},
			{Name: "language_key"},
		},
		DoUpdates: clause.AssignmentColumns([]string{"source_code"}),
	}).Create(&draft).Error
}
