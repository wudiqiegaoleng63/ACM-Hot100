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

// GetEnabledLanguageByKey returns an enabled language, or nil when unavailable.
func GetEnabledLanguageByKey(db *gorm.DB, key string) (*model.LanguageConfig, error) {
	var language model.LanguageConfig
	if err := db.Where("`key` = ? AND enabled = ?", key, true).First(&language).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &language, nil
}
