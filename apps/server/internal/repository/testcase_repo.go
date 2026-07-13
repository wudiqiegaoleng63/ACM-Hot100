package repository

import (
	"github.com/acmhot100/server/internal/model"
	"gorm.io/gorm"
)

// GetSampleCases returns only sample test cases (is_sample = true) for a problem,
// ordered by order_index.
func GetSampleCases(db *gorm.DB, problemID string) ([]model.TestCase, error) {
	var cases []model.TestCase
	if err := db.Where("problem_id = ? AND is_sample = ?", problemID, true).
		Order("order_index ASC").
		Find(&cases).Error; err != nil {
		return nil, err
	}
	return cases, nil
}

// GetSampleCaseForProblem returns a public sample belonging to the requested problem.
func GetSampleCaseForProblem(db *gorm.DB, problemID, sampleCaseID string) (*model.TestCase, error) {
	var testCase model.TestCase
	if err := db.Where("id = ? AND problem_id = ? AND is_sample = ?", sampleCaseID, problemID, true).
		First(&testCase).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &testCase, nil
}

// GetAllCases returns all test cases for a problem, including hidden ones.
// This should only be used by the judge worker.
func GetAllCases(db *gorm.DB, problemID string) ([]model.TestCase, error) {
	var cases []model.TestCase
	if err := db.Where("problem_id = ?", problemID).
		Order("order_index ASC").
		Find(&cases).Error; err != nil {
		return nil, err
	}
	return cases, nil
}
