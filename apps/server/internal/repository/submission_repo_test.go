package repository

import (
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/acmhot100/server/internal/model"
)

func TestCreateSubmissionSetsQueuedAndWritesProgress(t *testing.T) {
	db, mock := repositoryTestDB(t)

	// Expect transaction: INSERT submission, then upsert progress
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `submissions`")).
		WillReturnResult(sqlmock.NewResult(0, 1))
	// Progress upsert: UPDATE first (RowsAffected=0 means new record), then INSERT
	mock.ExpectExec("UPDATE `user_problem_progress`").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `user_problem_progress`")).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	submission := &model.Submission{
		ID:        "sub-1",
		UserID:    "user-1",
		ProblemID: "problem-1",
		Status:    model.SubmissionStatusQueued,
	}
	if err := CreateSubmission(db, submission); err != nil {
		t.Fatalf("CreateSubmission: %v", err)
	}
	if submission.Status != model.SubmissionStatusQueued {
		t.Fatalf("status = %s, want QUEUED", submission.Status)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestCreateSubmissionPreservesSolvedProgress(t *testing.T) {
	db, mock := repositoryTestDB(t)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `submissions`")).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE `user_problem_progress` SET .*`state`=CASE WHEN state = \\? THEN state ELSE \\? END").
		WithArgs(sqlmock.AnyArg(), model.ProgressSolved, model.ProgressAttempted, "user-1", "problem-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	submission := &model.Submission{ID: "sub-solved", UserID: "user-1", ProblemID: "problem-1", Status: model.SubmissionStatusQueued}
	if err := CreateSubmission(db, submission); err != nil {
		t.Fatalf("CreateSubmission: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestCreateSubmissionIncrementsExistingProgress(t *testing.T) {
	db, mock := repositoryTestDB(t)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `submissions`")).
		WillReturnResult(sqlmock.NewResult(0, 1))
	// Progress already exists: UPDATE returns RowsAffected=1
	mock.ExpectExec("UPDATE `user_problem_progress`").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	submission := &model.Submission{
		ID:        "sub-2",
		UserID:    "user-1",
		ProblemID: "problem-1",
		Status:    model.SubmissionStatusQueued,
	}
	if err := CreateSubmission(db, submission); err != nil {
		t.Fatalf("CreateSubmission: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestWriteSubmissionResultPersistsSanitizedErrorMessage(t *testing.T) {
	db, mock := repositoryTestDB(t)
	judgedAt := time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `submissions` SET .*`error_message`").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := WriteSubmissionResult(db, "sub-3", model.SubmissionStatusRuntimeError, 1, 3, 20, 2048, "", "Runtime Error at [path]", judgedAt); err != nil {
		t.Fatalf("WriteSubmissionResult: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestGetSubmissionForUserRejectsOtherUsersSubmission(t *testing.T) {
	db, mock := repositoryTestDB(t)
	// Query with user_id = "user-2" but submission belongs to "user-1"
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `submissions` WHERE id = ? AND user_id = ? ORDER BY `submissions`.`id` LIMIT ?")).
		WithArgs("sub-1", "user-2", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "status"}))

	submission, err := GetSubmissionForUser(db, "sub-1", "user-2")
	if err != nil {
		t.Fatalf("GetSubmissionForUser: %v", err)
	}
	if submission != nil {
		t.Fatal("should not return another user's submission (IDOR)")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}
