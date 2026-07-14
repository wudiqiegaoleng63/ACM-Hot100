package judge

import (
	"context"
	"regexp"
	"sync"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/acmhot100/server/internal/model"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestSubmissionWorkerProcessesQueuedSubmissionAndAcknowledges(t *testing.T) {
	db, mock := judgeTestDB(t)
	redisServer := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	ctx := context.Background()

	// Mock adapter returns AC with 3 cases
	adapter := NewFakeAdapter(FakeACResult(3))

	// Expect: check submission status (QUEUED), claim (QUEUED→COMPILING),
	// then CompleteSubmission (COMPILING/RUNNING→AC + case results + progress)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `submissions` WHERE id = ? ORDER BY `submissions`.`id` LIMIT ?")).
		WithArgs("sub-1", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "status"}).AddRow("sub-1", "QUEUED"))
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `submissions`").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "sub-1", "QUEUED").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	// CompleteSubmission transaction
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `submissions`").
		WillReturnResult(sqlmock.NewResult(0, 1))
	// 3 case result inserts
	for i := 0; i < 3; i++ {
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `submission_case_results`")).
			WillReturnResult(sqlmock.NewResult(0, 1))
	}
	// Read submission for progress update
	mock.ExpectQuery("SELECT .+ FROM `submissions` WHERE id = ?").
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "problem_id"}).AddRow("sub-1", "user-1", "problem-1"))
	// Progress update (SOLVED)
	mock.ExpectExec("UPDATE `user_problem_progress`").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	worker := NewSubmissionWorker(db, rdb, "worker-test", adapter)
	if err := worker.EnsureGroup(ctx); err != nil {
		t.Fatalf("EnsureGroup: %v", err)
	}

	// Enqueue a submission
	messageID, err := rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: "judge:submissions",
		Values: map[string]interface{}{"submission_id": "sub-1"},
	}).Result()
	if err != nil {
		t.Fatalf("XAdd: %v", err)
	}

	// Read the message
	streams, err := rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group: "judge-workers", Consumer: "worker-test", Streams: []string{"judge:submissions", ">"}, Count: 1,
	}).Result()
	if err != nil {
		t.Fatalf("XReadGroup: %v", err)
	}

	if err := worker.ProcessMessage(ctx, streams[0].Messages[0]); err != nil {
		t.Fatalf("ProcessMessage: %v", err)
	}

	// Verify ACK
	pending, err := rdb.XPending(ctx, "judge:submissions", "judge-workers").Result()
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

func TestSubmissionWorkerDuplicateMessageDoesNotJudgeTerminalSubmission(t *testing.T) {
	db, mock := judgeTestDB(t)
	redisServer := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	ctx := context.Background()
	adapter := &countingAdapter{result: FakeACResult(1)}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `submissions` WHERE id = ? ORDER BY `submissions`.`id` LIMIT ?")).
		WithArgs("sub-terminal", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "status"}).AddRow("sub-terminal", model.SubmissionStatusAccepted))

	worker := NewSubmissionWorker(db, rdb, "worker-test", adapter)
	if err := worker.EnsureGroup(ctx); err != nil {
		t.Fatalf("EnsureGroup: %v", err)
	}
	messageID, err := rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: "judge:submissions", Values: map[string]interface{}{"submission_id": "sub-terminal"},
	}).Result()
	if err != nil {
		t.Fatalf("XAdd: %v", err)
	}
	streams, err := rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group: "judge-workers", Consumer: "worker-test", Streams: []string{"judge:submissions", ">"}, Count: 1,
	}).Result()
	if err != nil {
		t.Fatalf("XReadGroup: %v", err)
	}
	if err := worker.ProcessMessage(ctx, streams[0].Messages[0]); err != nil {
		t.Fatalf("ProcessMessage: %v", err)
	}
	if adapter.calls != 0 {
		t.Fatalf("adapter calls = %d, want 0 for terminal duplicate", adapter.calls)
	}
	pending, err := rdb.XPending(ctx, "judge:submissions", "judge-workers").Result()
	if err != nil || pending.Count != 0 {
		t.Fatalf("pending = %#v, %v after ACK for %s", pending, err, messageID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

type countingAdapter struct {
	mu     sync.Mutex
	calls  int
	result *JudgeResult
}

func (a *countingAdapter) Judge(context.Context, string) (*JudgeResult, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.calls++
	return a.result, nil
}

func TestSubmissionWorkerMarksRetryExhaustionBeforeAcknowledging(t *testing.T) {
	db, mock := judgeTestDB(t)
	redisServer := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	ctx := context.Background()
	adapter := &countingAdapter{result: FakeACResult(1)}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `submissions` WHERE id = ? ORDER BY `submissions`.`id` LIMIT ?")).
		WithArgs("sub-exhausted", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "status", "retry_count"}).
			AddRow("sub-exhausted", model.SubmissionStatusRunning, maxRetryCount))
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `submissions`").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	worker := NewSubmissionWorker(db, rdb, "recovery-worker", adapter)
	if err := worker.EnsureGroup(ctx); err != nil {
		t.Fatalf("EnsureGroup: %v", err)
	}
	_, err := rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: "judge:submissions", Values: map[string]interface{}{"submission_id": "sub-exhausted"},
	}).Result()
	if err != nil {
		t.Fatalf("XAdd: %v", err)
	}
	streams, err := rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group: "judge-workers", Consumer: "dead-worker", Streams: []string{"judge:submissions", ">"}, Count: 1,
	}).Result()
	if err != nil {
		t.Fatalf("XReadGroup: %v", err)
	}
	if err := worker.ProcessMessage(ctx, streams[0].Messages[0]); err != nil {
		t.Fatalf("ProcessMessage: %v", err)
	}
	if adapter.calls != 0 {
		t.Fatalf("adapter calls = %d, want 0 after retry exhaustion", adapter.calls)
	}
	pending, err := rdb.XPending(ctx, "judge:submissions", "judge-workers").Result()
	if err != nil || pending.Count != 0 {
		t.Fatalf("pending = %#v, %v; want durable SYSTEM_ERROR before ACK", pending, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestSubmissionWorkerSkipsTerminalSubmission(t *testing.T) {
	db, mock := judgeTestDB(t)
	redisServer := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	ctx := context.Background()

	adapter := NewFakeAdapter(FakeACResult(1))

	// Submission is already AC (terminal)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `submissions` WHERE id = ? ORDER BY `submissions`.`id` LIMIT ?")).
		WithArgs("sub-2", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "status"}).AddRow("sub-2", model.SubmissionStatusAccepted))

	worker := NewSubmissionWorker(db, rdb, "worker-test", adapter)
	if err := worker.EnsureGroup(ctx); err != nil {
		t.Fatalf("EnsureGroup: %v", err)
	}

	messageID, err := rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: "judge:submissions",
		Values: map[string]interface{}{"submission_id": "sub-2"},
	}).Result()
	if err != nil {
		t.Fatalf("XAdd: %v", err)
	}

	streams, err := rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group: "judge-workers", Consumer: "worker-test", Streams: []string{"judge:submissions", ">"}, Count: 1,
	}).Result()
	if err != nil {
		t.Fatalf("XReadGroup: %v", err)
	}

	if err := worker.ProcessMessage(ctx, streams[0].Messages[0]); err != nil {
		t.Fatalf("ProcessMessage: %v", err)
	}

	// Should ACK even for terminal submissions
	pending, err := rdb.XPending(ctx, "judge:submissions", "judge-workers").Result()
	if err != nil {
		t.Fatalf("XPending: %v", err)
	}
	if pending.Count != 0 {
		t.Fatalf("pending = %d, want 0 for terminal submission %s", pending.Count, messageID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}
