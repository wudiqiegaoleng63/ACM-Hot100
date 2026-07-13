package repository

import (
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"

	"github.com/acmhot100/server/internal/model"
	"gorm.io/gorm"
)

func TestMapProblemProgressPreservesAuthenticatedState(t *testing.T) {
	results := mapProblemProgress(
		[]model.Problem{{ID: "problem-1"}, {ID: "problem-2"}},
		[]model.UserProblemProgress{{ProblemID: "problem-1", State: model.ProgressSolved}},
	)
	if results[0].State != model.ProgressSolved {
		t.Fatalf("solved state = %q, want %q", results[0].State, model.ProgressSolved)
	}
	if results[1].State != model.ProgressNotStarted {
		t.Fatalf("missing state = %q, want %q", results[1].State, model.ProgressNotStarted)
	}
}

func TestGetSampleCasesFiltersHiddenCases(t *testing.T) {
	db, mock := repositoryTestDB(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `test_cases` WHERE problem_id = ? AND is_sample = ? ORDER BY order_index ASC")).
		WithArgs("problem-1", true).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "problem_id", "order_index", "input_data", "expected_output", "is_sample", "explanation_md", "created_at",
		}))

	if _, err := GetSampleCases(db, "problem-1"); err != nil {
		t.Fatalf("GetSampleCases: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestListEnabledLanguagesFiltersDisabledConfigs(t *testing.T) {
	db, mock := repositoryTestDB(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `language_configs` WHERE enabled = ? ORDER BY `key` ASC")).
		WithArgs(true).
		WillReturnRows(sqlmock.NewRows([]string{
			"key", "display_name", "judge0_language_name", "judge0_language_id", "editor_language", "source_template", "enabled",
		}))

	if _, err := ListEnabledLanguages(db); err != nil {
		t.Fatalf("ListEnabledLanguages: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func repositoryTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create SQL mock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("open GORM with SQL mock: %v", err)
	}
	return db, mock
}
