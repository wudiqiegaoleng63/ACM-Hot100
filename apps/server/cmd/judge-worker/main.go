package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/acmhot100/server/internal/config"
	"github.com/acmhot100/server/internal/judge"
	"github.com/acmhot100/server/internal/queue"
	"github.com/redis/go-redis/v9"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Connect to MySQL
	db, err := gorm.Open(mysql.Open(cfg.MySQLDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect to MySQL: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get underlying SQL DB: %v", err)
	}
	defer sqlDB.Close()

	// Verify MySQL connection
	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("Failed to ping MySQL: %v", err)
	}
	log.Println("MySQL connected successfully")

	// Connect to Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	defer rdb.Close()

	// Verify Redis connection
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to ping Redis: %v", err)
	}
	log.Println("Redis connected successfully")

	// Set Redis key prefix for queue helpers
	queue.SetPrefix(cfg.RedisKeyPrefix)

	if cfg.JudgeMode != "mock" {
		log.Fatalf("Unsupported JUDGE_MODE %q: sample run worker currently requires mock", cfg.JudgeMode)
	}
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalf("Failed to resolve worker hostname: %v", err)
	}
	consumerName := fmt.Sprintf("sample-run-%s-%d", hostname, os.Getpid())
	worker := judge.NewSampleRunWorker(db, rdb, consumerName)
	if err := worker.EnsureGroup(ctx); err != nil {
		log.Fatalf("Failed to initialize sample run consumer group: %v", err)
	}

	workerCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	log.Printf("Judge worker started (mode: %s, consumer: %s)", cfg.JudgeMode, consumerName)
	if err := worker.Run(workerCtx); err != nil {
		log.Fatalf("Judge worker failed: %v", err)
	}
	log.Println("Judge worker shut down")
}
