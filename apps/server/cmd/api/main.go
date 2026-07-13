package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/acmhot100/server/internal/config"
	serverHTTP "github.com/acmhot100/server/internal/http"
	"github.com/acmhot100/server/internal/queue"
	"github.com/acmhot100/server/internal/service"
	"github.com/redis/go-redis/v9"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Parse command-line flags
	migrateFlag := flag.Bool("migrate", false, "run database migrations and exit")
	seedFlag := flag.Bool("seed", false, "run database seed and exit")
	migrateDownFlag := flag.Bool("migrate-down", false, "rollback database migrations")
	flag.Parse()

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

	// Configure connection pool
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

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

	// Handle subcommands
	if *migrateFlag {
		migrationsDir := getMigrationsDir()
		log.Printf("Running UP migrations from %s", migrationsDir)
		if err := config.RunMigrations(db, migrationsDir, config.MigrationUp); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		log.Println("Migrations completed successfully")
		return
	}

	if *migrateDownFlag {
		migrationsDir := getMigrationsDir()
		log.Printf("Running DOWN migrations from %s", migrationsDir)
		if err := config.RunMigrations(db, migrationsDir, config.MigrationDown); err != nil {
			log.Fatalf("Migration rollback failed: %v", err)
		}
		log.Println("Migration rollback completed successfully")
		return
	}

	if *seedFlag {
		seedDir := getSeedDir()
		log.Printf("Seeding database from %s", seedDir)
		if err := service.SeedAll(db, seedDir); err != nil {
			log.Fatalf("Seed failed: %v", err)
		}
		return
	}

	// Auto-run pending migrations on startup
	migrationsDir := getMigrationsDir()
	if err := config.RunMigrations(db, migrationsDir, config.MigrationUp); err != nil {
		log.Fatalf("Auto-migration failed: %v", err)
	}

	// Create and start HTTP server
	engine := serverHTTP.NewServer(cfg, db, rdb)

	srv := &http.Server{
		Addr:    cfg.APIAddr,
		Handler: engine,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting API server on %s", cfg.APIAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}

// getMigrationsDir resolves the absolute path to the migrations directory.
func getMigrationsDir() string {
	// Try relative to executable first
	exePath, err := os.Executable()
	if err == nil {
		candidate := filepath.Join(filepath.Dir(exePath), "migrations")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	// Try relative to working directory
	candidate := "migrations"
	if _, err := os.Stat(candidate); err == nil {
		abs, _ := filepath.Abs(candidate)
		return abs
	}

	// Try relative to project root (go mod directory)
	candidate = filepath.Join("..", "..", "migrations")
	if abs, err := filepath.Abs(candidate); err == nil {
		if _, err := os.Stat(abs); err == nil {
			return abs
		}
	}

	// Default fallback
	fmt.Println("Warning: migrations directory not found, using default path")
	abs, _ := filepath.Abs("migrations")
	return abs
}

// getSeedDir resolves the absolute path to the seed problems directory.
func getSeedDir() string {
	// Try relative to project root (apps/server -> ../../seed/problems)
	candidate := filepath.Join("..", "..", "seed", "problems")
	if abs, err := filepath.Abs(candidate); err == nil {
		if _, err := os.Stat(abs); err == nil {
			return abs
		}
	}

	// Try relative to working directory
	candidate = filepath.Join("seed", "problems")
	if abs, err := filepath.Abs(candidate); err == nil {
		if _, err := os.Stat(abs); err == nil {
			return abs
		}
	}

	// Default fallback
	abs, _ := filepath.Abs(filepath.Join("..", "..", "seed", "problems"))
	return abs
}
