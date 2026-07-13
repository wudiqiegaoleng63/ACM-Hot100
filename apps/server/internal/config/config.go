package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	// App
	AppEnv     string
	AppBaseURL string
	APIAddr    string

	// MySQL
	MySQLDSN string

	// Redis
	RedisAddr      string
	RedisPassword  string
	RedisDB        int
	RedisKeyPrefix string

	// JWT
	JWTIssuer          string
	JWTAccessAudience  string
	JWTRefreshAudience string
	JWTAccessSecret    string
	JWTRefreshSecret   string
	JWTAccessTTL       int // seconds
	JWTRefreshTTL      int // seconds

	// SMTP
	MailMode     string // "smtp" or "log"
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string
	SMTPTLSMode  string // "none", "starttls", "tls"

	// Judge
	JudgeMode     string // "local" or "remote"
	Judge0BaseURL string
}

// Load reads configuration from environment variables and .env file.
func Load() *Config {
	// Try loading .env file from multiple locations (ignore errors if not found)
	_ = godotenv.Load()                // current directory
	_ = godotenv.Load("../../.env")    // monorepo root (when running from apps/server/)
	_ = godotenv.Load("../../../.env") // monorepo root (when running from apps/server/cmd/api/)

	cfg := &Config{
		// App
		AppEnv:     getEnv("APP_ENV", "development"),
		AppBaseURL: getEnv("APP_BASE_URL", "http://localhost:3000"),
		APIAddr:    getEnv("API_ADDR", ":8080"),

		// MySQL
		MySQLDSN: getEnv("MYSQL_DSN", "root:password@tcp(127.0.0.1:3306)/acmhot100?charset=utf8mb4&parseTime=True&loc=UTC"),

		// Redis
		RedisAddr:      getEnv("REDIS_ADDR", "127.0.0.1:6379"),
		RedisPassword:  getEnv("REDIS_PASSWORD", ""),
		RedisDB:        getEnvInt("REDIS_DB", 0),
		RedisKeyPrefix: getEnv("REDIS_KEY_PREFIX", "acmhot100:"),

		// JWT
		JWTIssuer:          getEnv("JWT_ISSUER", "acmhot100"),
		JWTAccessAudience:  getEnv("JWT_ACCESS_AUDIENCE", "acmhot100-access"),
		JWTRefreshAudience: getEnv("JWT_REFRESH_AUDIENCE", "acmhot100-refresh"),
		JWTAccessSecret:    getEnv("JWT_ACCESS_SECRET", "dev-access-secret-change-me"),
		JWTRefreshSecret:   getEnv("JWT_REFRESH_SECRET", "dev-refresh-secret-change-me"),
		JWTAccessTTL:       getEnvInt("JWT_ACCESS_TTL", 900),     // 15 minutes
		JWTRefreshTTL:      getEnvInt("JWT_REFRESH_TTL", 604800), // 7 days

		// SMTP
		MailMode:     getEnv("MAIL_MODE", "smtp"),
		SMTPHost:     getEnv("SMTP_HOST", ""),
		SMTPPort:     getEnvInt("SMTP_PORT", 587),
		SMTPUsername: getEnv("SMTP_USERNAME", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:     getEnv("SMTP_FROM", ""),
		SMTPTLSMode:  getEnv("SMTP_TLS_MODE", "starttls"),

		// Judge
		JudgeMode:     getEnv("JUDGE_MODE", "remote"),
		Judge0BaseURL: getEnv("JUDGE0_BASE_URL", "http://127.0.0.1:2358"),
	}

	return cfg
}

// IsDevelopment returns true if the app is running in development mode.
func (c *Config) IsDevelopment() bool {
	return c.AppEnv == "development"
}

// SMTPAddr returns the formatted SMTP host:port address.
func (c *Config) SMTPAddr() string {
	return fmt.Sprintf("%s:%d", c.SMTPHost, c.SMTPPort)
}

// getEnv reads an environment variable or returns a default value.
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvInt reads an environment variable as an integer or returns a default value.
func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
