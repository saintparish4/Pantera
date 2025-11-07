package database

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/saintparish4/harmonia/config"
)

// DB is the global database connection
var DB *sql.DB

// Connect establishes a connection to PostgreSQL
func Connect(cfg *config.Config) error {
	dsn := cfg.GetDatabaseDSN()

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	// Verify connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	DB = db
	log.Println("✓ Database connection established")
	return nil
}

// Close closes the database connection
func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

// Migrate runs all pending migrations
func Migrate(migrationsPath string) error {
	log.Println("Running database migrations...")

	// Create migrations table if it doesn't exist
	if err := createMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get list of migration files
	files, err := filepath.Glob(filepath.Join(migrationsPath, "*.up.sql"))
	if err != nil {
		return fmt.Errorf("failed to read migration files: %w", err)
	}

	if len(files) == 0 {
		log.Println("No migration files found")
		return nil
	}

	// Sort files by name (which includes version number)
	sort.Strings(files)

	// Run each migration
	for _, file := range files {
		migrationName := filepath.Base(file)
		migrationName = strings.TrimSuffix(migrationName, ".up.sql")

		// Check if migration already applied
		var count int
		err := DB.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = $1", migrationName).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to check migration status: %w", err)
		}

		if count > 0 {
			log.Printf("  ⊘ Skipping %s (already applied)", migrationName)
			continue
		}

		// Read migration file
		content, err := ioutil.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		// Execute migration in a transaction
		tx, err := DB.Begin()
		if err != nil {
			return fmt.Errorf("failed to start transaction: %w", err)
		}

		if _, err := tx.Exec(string(content)); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute migration %s: %w", migrationName, err)
		}

		// Record migration
		if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", migrationName); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %s: %w", migrationName, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", migrationName, err)
		}

		log.Printf("  ✓ Applied %s", migrationName)
	}

	log.Println("✓ All migrations completed")
	return nil
}

// createMigrationsTable creates the schema_migrations table
func createMigrationsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`
	_, err := DB.Exec(query)
	return err
}

// HealthCheck verifies the database connection is healthy
func HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := DB.PingContext(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	return nil
}
