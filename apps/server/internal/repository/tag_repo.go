package repository

import (
	"github.com/acmhot100/server/internal/model"
	"gorm.io/gorm"
)

// ListTags returns all tags ordered by name.
func ListTags(db *gorm.DB) ([]model.Tag, error) {
	var tags []model.Tag
	if err := db.Order("name ASC").Find(&tags).Error; err != nil {
		return nil, err
	}
	return tags, nil
}
