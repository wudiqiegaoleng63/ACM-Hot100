package repository

import (
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestGetProfileProgressSummaryScopesToUserAndPublishedProblems(t *testing.T) {
	db, mock := repositoryTestDB(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) AS total_problems,")).
		WithArgs("SOLVED", "ATTEMPTED", "NOT_STARTED", "user-1", true).
		WillReturnRows(sqlmock.NewRows([]string{"total_problems", "solved", "attempted", "not_started"}).
			AddRow(5, 2, 1, 2))

	summary, err := GetProfileProgressSummary(db, "user-1")
	if err != nil {
		t.Fatalf("GetProfileProgressSummary: %v", err)
	}
	if summary.TotalProblems != 5 || summary.Solved != 2 || summary.Attempted != 1 || summary.NotStarted != 2 {
		t.Fatalf("summary = %#v", summary)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestGetProfileProgressByStageKeepsStageOrder(t *testing.T) {
	db, mock := repositoryTestDB(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT problems.stage AS stage,")).
		WithArgs("SOLVED", "ATTEMPTED", "NOT_STARTED", "user-1", true).
		WillReturnRows(sqlmock.NewRows([]string{"stage", "total", "solved", "attempted", "not_started"}).
			AddRow("数组与哈希", 2, 1, 1, 0).
			AddRow("树与图", 1, 0, 0, 1))

	stages, err := GetProfileProgressByStage(db, "user-1")
	if err != nil {
		t.Fatalf("GetProfileProgressByStage: %v", err)
	}
	if len(stages) != 2 || stages[0].Stage != "数组与哈希" || stages[1].NotStarted != 1 {
		t.Fatalf("stages = %#v", stages)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}
