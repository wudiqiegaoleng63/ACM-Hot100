package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gorm.io/gorm"
)

// MigrationDirection represents the direction of a migration.
type MigrationDirection string

const (
	MigrationUp   MigrationDirection = "up"
	MigrationDown MigrationDirection = "down"
)

// RunMigrations executes SQL migration files from the migrations directory.
// It tracks applied migrations in a schema_migrations table.
func RunMigrations(db *gorm.DB, migrationsDir string, direction MigrationDirection) error {
	// Create schema_migrations table if it doesn't exist
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) NOT NULL PRIMARY KEY,
			applied_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
	`).Error; err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	// Read migration files from directory
	pattern := filepath.Join(migrationsDir, "*.sql")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	if len(files) == 0 {
		fmt.Println("No migration files found")
		return nil
	}

	// Sort files by name
	sort.Strings(files)

	if direction == MigrationDown {
		// Reverse order for down migrations
		sort.Sort(sort.Reverse(sort.StringSlice(files)))
	}

	for _, file := range files {
		filename := filepath.Base(file)

		// Determine if this file matches the requested direction
		if direction == MigrationUp && strings.HasSuffix(filename, ".down.sql") {
			continue
		}
		if direction == MigrationDown && strings.HasSuffix(filename, ".up.sql") {
			continue
		}

		// Extract version from filename (e.g., "000001" from "000001_init_schema.up.sql")
		parts := strings.SplitN(filename, "_", 2)
		if len(parts) < 1 {
			continue
		}
		version := parts[0]

		// Check if migration was already applied
		var count int64
		db.Table("schema_migrations").Where("version = ?", version).Count(&count)

		if direction == MigrationUp {
			if count > 0 {
				fmt.Printf("Migration %s already applied, skipping\n", version)
				continue
			}
		} else {
			if count == 0 {
				fmt.Printf("Migration %s not applied, skipping down\n", version)
				continue
			}
		}

		// Read and execute the migration file
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", filename, err)
		}

		if strings.TrimSpace(string(content)) == "" {
			fmt.Printf("Migration file %s is empty, skipping\n", filename)
			continue
		}

		fmt.Printf("Running migration: %s (%s)\n", filename, direction)

		// Split SQL into individual statements and execute each one
		statements := splitSQL(string(content))

		// Execute the migration in a transaction
		tx := db.Begin()
		for _, stmt := range statements {
			if err := tx.Exec(stmt).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to execute migration %s: %w", filename, err)
			}
		}

		if direction == MigrationUp {
			if err := tx.Exec("INSERT INTO schema_migrations (version) VALUES (?)", version).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to record migration %s: %w", version, err)
			}
		} else {
			if err := tx.Exec("DELETE FROM schema_migrations WHERE version = ?", version).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to remove migration record %s: %w", version, err)
			}
		}

		if err := tx.Commit().Error; err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", filename, err)
		}

		fmt.Printf("Migration %s applied successfully\n", version)
	}

	return nil
}

// splitSQL splits a SQL file content into individual statements.
// It handles multi-line statements, strips comments, and ignores empty statements.
func splitSQL(content string) []string {
	var statements []string
	var current strings.Builder

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Skip empty lines and SQL comments
		if trimmed == "" || strings.HasPrefix(trimmed, "--") {
			continue
		}
		if current.Len() > 0 {
			current.WriteString("\n")
		}
		current.WriteString(line)

		// If the line ends with a semicolon, we have a complete statement
		if strings.HasSuffix(trimmed, ";") {
			stmt := strings.TrimSpace(current.String())
			if stmt != "" {
				statements = append(statements, stmt)
			}
			current.Reset()
		}
	}

	// Handle any remaining content without trailing semicolon
	if current.Len() > 0 {
		stmt := strings.TrimSpace(current.String())
		if stmt != "" && !strings.HasPrefix(stmt, "--") {
			statements = append(statements, stmt)
		}
	}

	return statements
}
