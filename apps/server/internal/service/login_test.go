package service

import (
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/acmhot100/server/internal/auth"
	"github.com/acmhot100/server/internal/model"
)

func TestLoginPendingUserOnlyRevealedAfterCorrectPassword(t *testing.T) {
	passwordHash, err := auth.HashPassword("CorrectPass123!")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	tests := []struct {
		name     string
		password string
		wantErr  error
	}{
		{name: "wrong password", password: "WrongPass123!", wantErr: ErrInvalidCredentials},
		{name: "correct password", password: "CorrectPass123!", wantErr: ErrEmailNotVerified},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := loginTestDB(t)
			rows := sqlmock.NewRows([]string{
				"id", "email", "username", "password_hash", "email_verified_at", "status", "created_at", "updated_at",
			}).AddRow(
				"user-123", "pending@example.local", "pending-user", passwordHash, nil,
				model.UserStatusPending, time.Now(), time.Now(),
			)
			mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE email = ? ORDER BY `users`.`id` LIMIT ?")).
				WithArgs("pending@example.local", 1).
				WillReturnRows(rows)

			redisServer := miniredis.RunT(t)
			rdb := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
			t.Cleanup(func() { _ = rdb.Close() })

			_, _, err := Login(db, rdb, refreshTestConfig(), "pending@example.local", tt.password)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Login() error = %v, want %v", err, tt.wantErr)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("unmet SQL expectations: %v", err)
			}
		})
	}
}

func loginTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
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
