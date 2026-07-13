package repository

import (
	"github.com/acmhot100/server/internal/model"
	"gorm.io/gorm"
)

// ListEnabledLanguages returns public language choices in stable key order.
func ListEnabledLanguages(db *gorm.DB) ([]model.LanguageConfig, error) {
	var languages []model.LanguageConfig
	if err := db.Where("enabled = ?", true).Order("`key` ASC").Find(&languages).Error; err != nil {
		return nil, err
	}
	return languages, nil
}
