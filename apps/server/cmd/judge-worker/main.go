package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/acmhot100/server/internal/config"
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

	log.Printf("Judge worker started (mode: %s)", cfg.JudgeMode)

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Placeholder: main worker loop
	// In production, this would consume from the Redis stream and process submissions
	for {
		select {
		case <-quit:
			log.Println("Judge worker shutting down...")
			return
		default:
			// TODO: Poll Redis stream for new submissions
			// For now, just wait for shutdown signal
			<-quit
			log.Println("Judge worker shutting down...")
			return
		}
	}
}
