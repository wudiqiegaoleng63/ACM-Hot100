package queue

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestSubmissionReconcilerEnqueuesMissingSubmissionAndStoresMetadata(t *testing.T) {
	db, mock := reconciliationTestDB(t)
	server := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	SetPrefix("reconcile-test:")
	t.Cleanup(func() { SetPrefix("") })

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `submissions` WHERE status = ? AND enqueued_at IS NULL ORDER BY created_at ASC LIMIT ?")).
		WithArgs("QUEUED", defaultReconcileBatchSize).
		WillReturnRows(sqlmock.NewRows([]string{"id", "status"}).AddRow("sub-missing", "QUEUED"))
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `submissions`").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "sub-missing", "QUEUED").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	reconciler := NewSubmissionReconciler(db, rdb)
	if err := reconciler.ReconcileOnce(context.Background()); err != nil {
		t.Fatalf("ReconcileOnce: %v", err)
	}
	entries, err := rdb.XRange(context.Background(), KeyJudgeSubmissions(), "-", "+").Result()
	if err != nil {
		t.Fatalf("XRange: %v", err)
	}
	if len(entries) != 1 || entries[0].Values["submission_id"] != "sub-missing" {
		t.Fatalf("entries = %#v, want one repaired submission", entries)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestSubmissionReconcilerDoesNotEnqueueWhenScanIsEmpty(t *testing.T) {
	db, mock := reconciliationTestDB(t)
	server := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `submissions` WHERE status = ? AND enqueued_at IS NULL ORDER BY created_at ASC LIMIT ?")).
		WithArgs("QUEUED", defaultReconcileBatchSize).
		WillReturnRows(sqlmock.NewRows([]string{"id", "status"}))

	if err := NewSubmissionReconciler(db, rdb).ReconcileOnce(context.Background()); err != nil {
		t.Fatalf("ReconcileOnce: %v", err)
	}
	if server.Exists(KeyJudgeSubmissions()) {
		t.Fatal("empty reconciliation created a stream entry")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func reconciliationTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	db, err := gorm.Open(mysql.New(mysql.Config{Conn: sqlDB, SkipInitializeWithVersion: true}), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open: %v", err)
	}
	return db, mock
}
