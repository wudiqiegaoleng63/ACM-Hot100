package service

import (
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/acmhot100/server/internal/auth"
	"github.com/acmhot100/server/internal/model"
	"github.com/acmhot100/server/internal/queue"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestVerifyEmailConsumesTokenAndActivatesUser(t *testing.T) {
	db, mock := loginTestDB(t)
	server := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	queue.SetPrefix("auth-flow-test:")
	t.Cleanup(func() { queue.SetPrefix("") })

	rawToken := "verification-token"
	userID := "user-verify"
	if err := rdb.Set(t.Context(), queue.KeyAuthVerify(auth.HashToken(rawToken)), userID, time.Minute).Err(); err != nil {
		t.Fatalf("store verification token: %v", err)
	}
	if err := rdb.Set(t.Context(), queue.KeyAuthVerifyUser(userID), auth.HashToken(rawToken), time.Minute).Err(); err != nil {
		t.Fatalf("store user token tracking: %v", err)
	}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE id = ? ORDER BY `users`.`id` LIMIT ?")).
		WithArgs(userID, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "username", "password_hash", "status"}).
			AddRow(userID, "verify@example.local", "verify-user", "hash", model.UserStatusPending))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `users` SET")).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := VerifyEmail(db, rdb, rawToken); err != nil {
		t.Fatalf("VerifyEmail: %v", err)
	}
	if err := VerifyEmail(db, rdb, rawToken); !errors.Is(err, ErrTokenExpired) {
		t.Fatalf("second VerifyEmail error = %v, want ErrTokenExpired", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestLogoutRevokesRefreshAndDeniesAccess(t *testing.T) {
	server := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	queue.SetPrefix("auth-flow-test:")
	t.Cleanup(func() { queue.SetPrefix("") })

	if err := auth.StoreRefreshSession(rdb, "refresh-jti", "user-1", "family-1", time.Hour); err != nil {
		t.Fatalf("StoreRefreshSession: %v", err)
	}
	if err := Logout(rdb, "access-jti", time.Minute, "refresh-jti"); err != nil {
		t.Fatalf("Logout: %v", err)
	}
	if _, _, err := auth.GetRefreshSession(rdb, "refresh-jti"); !errors.Is(err, redis.Nil) {
		t.Fatalf("refresh session still valid: %v", err)
	}
	denied, err := auth.IsAccessDenied(rdb, "access-jti")
	if err != nil || !denied {
		t.Fatalf("access denied = %t, %v; want true", denied, err)
	}
}

func TestForgotPasswordDoesNotRevealMissingEmail(t *testing.T) {
	db, mock := loginTestDB(t)
	server := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE email = ? ORDER BY `users`.`id` LIMIT ?")).
		WithArgs("missing@example.local", 1).
		WillReturnError(errors.New("record not found"))

	if err := ForgotPassword(db, rdb, refreshTestConfig(), "missing@example.local"); err != nil {
		t.Fatalf("ForgotPassword leaked missing email through error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}
