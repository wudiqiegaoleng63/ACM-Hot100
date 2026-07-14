package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/acmhot100/server/internal/config"
	"github.com/acmhot100/server/internal/judge"
	"github.com/acmhot100/server/internal/queue"
	"github.com/redis/go-redis/v9"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// Load configuration
	cfg := config.Load()

	if err := cfg.ValidateProduction(); err != nil {
		log.Fatalf("Unsafe production configuration: %v", err)
	}

	// Connect to MySQL
	db, err := gorm.Open(mysql.Open(cfg.MySQLDSN), &gorm.Config{
		Logger: config.GORMLogger(cfg),
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

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalf("Failed to resolve worker hostname: %v", err)
	}
	consumerName := fmt.Sprintf("worker-%s-%d", hostname, os.Getpid())

	// Create judge adapter based on mode
	var adapter judge.Adapter
	switch cfg.JudgeMode {
	case "mock":
		adapter = judge.NewFakeAdapter(judge.FakeACResult(1))
		log.Println("Judge adapter: mock (always AC)")
	case "judge0":
		judge0Adapter := judge.NewJudge0Adapter(db, judge.Judge0AdapterConfig{
			BaseURL:        cfg.Judge0BaseURL,
			ConnectTimeout: 5 * time.Second,
			TotalTimeout:   60 * time.Second,
		})
		adapter = judge0Adapter
		log.Printf("Judge adapter: Judge0 (%s)", cfg.Judge0BaseURL)
	default:
		log.Fatalf("Unsupported JUDGE_MODE %q: only 'mock' and 'judge0' are supported", cfg.JudgeMode)
	}

	// Initialize sample run worker
	sampleRunWorker := judge.NewSampleRunWorker(db, rdb, consumerName+"-runs")
	if err := sampleRunWorker.EnsureGroup(ctx); err != nil {
		log.Fatalf("Failed to initialize sample run consumer group: %v", err)
	}

	// Initialize submission worker
	submissionWorker := judge.NewSubmissionWorker(db, rdb, consumerName+"-subs", adapter)
	if err := submissionWorker.EnsureGroup(ctx); err != nil {
		log.Fatalf("Failed to initialize submission consumer group: %v", err)
	}

	// Set up graceful shutdown
	workerCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Printf("Judge worker started (mode: %s, consumer: %s)", cfg.JudgeMode, consumerName)

	// Run both workers concurrently
	errCh := make(chan error, 2)

	go func() {
		if err := sampleRunWorker.Run(workerCtx); err != nil {
			errCh <- fmt.Errorf("sample run worker: %w", err)
		}
		errCh <- nil
	}()

	go func() {
		if err := submissionWorker.Run(workerCtx); err != nil {
			errCh <- fmt.Errorf("submission worker: %w", err)
		}
		errCh <- nil
	}()

	// Wait for context cancellation or worker errors
	select {
	case <-workerCtx.Done():
		log.Println("Shutdown signal received, waiting for workers to finish...")
	case err := <-errCh:
		if err != nil {
			log.Fatalf("Worker error: %v", err)
		}
	}

	// Wait for both workers to complete
	for i := 0; i < 2; i++ {
		if err := <-errCh; err != nil {
			log.Printf("Worker error during shutdown: %v", err)
		}
	}

	log.Println("Judge worker shut down")
}
