package repository

import (
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestGetSampleCaseForProblemFiltersHiddenAndForeignCases(t *testing.T) {
	db, mock := repositoryTestDB(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `test_cases` WHERE id = ? AND problem_id = ? AND is_sample = ? ORDER BY `test_cases`.`id` LIMIT ?")).
		WithArgs("sample-1", "problem-1", true, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "problem_id", "is_sample", "input_data"}).
			AddRow("sample-1", "problem-1", true, "1 2\n"))

	testCase, err := GetSampleCaseForProblem(db, "problem-1", "sample-1")
	if err != nil {
		t.Fatalf("GetSampleCaseForProblem: %v", err)
	}
	if testCase == nil || testCase.ID != "sample-1" {
		t.Fatalf("test case = %#v, want sample-1", testCase)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestSampleRunStateTransitionsRequireExpectedCurrentStatus(t *testing.T) {
	db, mock := repositoryTestDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `sample_runs`").
		WithArgs(sqlmock.AnyArg(), "RUNNING", sqlmock.AnyArg(), "run-1", "QUEUED").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	claimed, err := ClaimQueuedSampleRun(db, "run-1", time.Now())
	if err != nil {
		t.Fatalf("ClaimQueuedSampleRun: %v", err)
	}
	if claimed {
		t.Fatal("claimed terminal or already-running sample run")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}
