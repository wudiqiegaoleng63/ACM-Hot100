package config

import (
	"log"
	"os"
	"time"

	"gorm.io/gorm/logger"
)

// GORMLogger avoids writing query parameters such as source code or test data to logs.
func GORMLogger(cfg *Config) logger.Interface {
	level := logger.Warn
	if cfg.IsDevelopment() {
		level = logger.Info
	}
	return logger.New(log.New(os.Stdout, "", log.LstdFlags), logger.Config{
		SlowThreshold:             time.Second,
		LogLevel:                  level,
		IgnoreRecordNotFoundError: true,
		ParameterizedQueries:      true,
		Colorful:                  cfg.IsDevelopment(),
	})
}
