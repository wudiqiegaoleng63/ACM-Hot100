package judge

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestSampleRunWorkerProcessesQueuedRunAndAcknowledges(t *testing.T) {
	db, mock := judgeTestDB(t)
	redisServer := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `sample_runs` WHERE id = ? ORDER BY `sample_runs`.`id` LIMIT ?")).
		WithArgs("run-1", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "status"}).AddRow("run-1", "QUEUED"))
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `sample_runs`").
		WithArgs(sqlmock.AnyArg(), "RUNNING", sqlmock.AnyArg(), "run-1", "QUEUED").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `sample_runs`").
		WithArgs("", sqlmock.AnyArg(), "", "AC", sqlmock.AnyArg(), "run-1", "RUNNING").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	worker := NewSampleRunWorker(db, rdb, "worker-test")
	worker.delay = 500 * time.Millisecond
	var slept time.Duration
	worker.sleep = func(_ context.Context, delay time.Duration) error {
		slept = delay
		return nil
	}

	if err := worker.EnsureGroup(ctx); err != nil {
		t.Fatalf("EnsureGroup: %v", err)
	}
	messageID, err := rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: "judge:runs",
		Values: map[string]interface{}{"run_id": "run-1"},
	}).Result()
	if err != nil {
		t.Fatalf("XAdd: %v", err)
	}
	streams, err := rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group: "judge-workers", Consumer: "worker-test", Streams: []string{"judge:runs", ">"}, Count: 1,
	}).Result()
	if err != nil {
		t.Fatalf("XReadGroup: %v", err)
	}
	if err := worker.ProcessMessage(ctx, streams[0].Messages[0]); err != nil {
		t.Fatalf("ProcessMessage: %v", err)
	}
	if slept < 500*time.Millisecond || slept > time.Second {
		t.Fatalf("mock delay = %v, want 500-1000ms", slept)
	}
	pending, err := rdb.XPending(ctx, "judge:runs", "judge-workers").Result()
	if err != nil {
		t.Fatalf("XPending: %v", err)
	}
	if pending.Count != 0 {
		t.Fatalf("pending = %d, want 0 after ACK for %s", pending.Count, messageID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func judgeTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
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
