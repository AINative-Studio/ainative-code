package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/AINative-studio/ainative-code/internal/errors"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Migration represents a database migration
type Migration struct {
	Version     int
	Name        string
	UpSQL       string
	DownSQL     string
	AppliedAt   time.Time
	Description string
}

// MigrationStatus represents the status of migrations
type MigrationStatus struct {
	CurrentVersion int
	Applied        []Migration
	Pending        []Migration
}

const (
	migrationTableSQL = `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		description TEXT
	);
	`
)

// Migrate runs all pending migrations
func Migrate(db *sql.DB) error {
	return MigrateContext(context.Background(), db)
}

// MigrateContext runs all pending migrations with context
func MigrateContext(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return errors.NewDatabaseError(errors.ErrCodeDBConnection, "database connection is nil")
	}

	// Ensure migration table exists
	if err := ensureMigrationTable(ctx, db); err != nil {
		return errors.NewDBQueryError("create migration table", "schema_migrations", err)
	}

	// Load migrations from embedded filesystem
	migrations, err := loadMigrations()
	if err != nil {
		return errors.NewDatabaseError(errors.ErrCodeDBQuery, fmt.Sprintf("failed to load migrations: %v", err))
	}

	// Get applied migrations
	applied, err := getAppliedMigrations(ctx, db)
	if err != nil {
		return errors.NewDBQueryError("get applied migrations", "schema_migrations", err)
	}

	appliedVersions := make(map[int]bool)
	for _, m := range applied {
		appliedVersions[m.Version] = true
	}

	// Apply pending migrations
	for _, migration := range migrations {
		if appliedVersions[migration.Version] {
			continue
		}

		if err := applyMigration(ctx, db, migration); err != nil {
			return errors.NewDBTransactionError(
				fmt.Sprintf("migration %03d_%s", migration.Version, migration.Name),
				err,
			)
		}
	}

	return nil
}

// Rollback rolls back the last applied migration
func Rollback(db *sql.DB) error {
	return RollbackContext(context.Background(), db)
}

// RollbackContext rolls back the last applied migration with context
func RollbackContext(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return errors.NewDatabaseError(errors.ErrCodeDBConnection, "database connection is nil")
	}

	// Get applied migrations
	applied, err := getAppliedMigrations(ctx, db)
	if err != nil {
		return errors.NewDBQueryError("get applied migrations", "schema_migrations", err)
	}

	if len(applied) == 0 {
		return errors.NewDatabaseError(errors.ErrCodeDBQuery, "no migrations to rollback")
	}

	// Get the last applied migration
	lastMigration := applied[len(applied)-1]

	// Load migrations to get down SQL
	migrations, err := loadMigrations()
	if err != nil {
		return errors.NewDatabaseError(errors.ErrCodeDBQuery, fmt.Sprintf("failed to load migrations: %v", err))
	}

	var migrationToRollback *Migration
	for _, m := range migrations {
		if m.Version == lastMigration.Version {
			migrationToRollback = &m
			break
		}
	}

	if migrationToRollback == nil {
		return errors.NewDatabaseError(
			errors.ErrCodeDBQuery,
			fmt.Sprintf("migration file not found for version %d", lastMigration.Version),
		)
	}

	// Rollback the migration
	if err := rollbackMigration(ctx, db, *migrationToRollback); err != nil {
		return errors.NewDBTransactionError(
			fmt.Sprintf("rollback migration %03d_%s", migrationToRollback.Version, migrationToRollback.Name),
			err,
		)
	}

	return nil
}

// GetStatus returns the current migration status
func GetStatus(db *sql.DB) (*MigrationStatus, error) {
	return GetStatusContext(context.Background(), db)
}

// GetStatusContext returns the current migration status with context
func GetStatusContext(ctx context.Context, db *sql.DB) (*MigrationStatus, error) {
	if db == nil {
		return nil, errors.NewDatabaseError(errors.ErrCodeDBConnection, "database connection is nil")
	}

	// Ensure migration table exists
	if err := ensureMigrationTable(ctx, db); err != nil {
		return nil, errors.NewDBQueryError("create migration table", "schema_migrations", err)
	}

	// Load all migrations
	allMigrations, err := loadMigrations()
	if err != nil {
		return nil, errors.NewDatabaseError(errors.ErrCodeDBQuery, fmt.Sprintf("failed to load migrations: %v", err))
	}

	// Get applied migrations
	applied, err := getAppliedMigrations(ctx, db)
	if err != nil {
		return nil, errors.NewDBQueryError("get applied migrations", "schema_migrations", err)
	}

	appliedVersions := make(map[int]bool)
	currentVersion := 0
	for _, m := range applied {
		appliedVersions[m.Version] = true
		if m.Version > currentVersion {
			currentVersion = m.Version
		}
	}

	// Determine pending migrations
	var pending []Migration
	for _, m := range allMigrations {
		if !appliedVersions[m.Version] {
			pending = append(pending, m)
		}
	}

	return &MigrationStatus{
		CurrentVersion: currentVersion,
		Applied:        applied,
		Pending:        pending,
	}, nil
}

// ensureMigrationTable creates the migration tracking table if it doesn't exist
func ensureMigrationTable(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, migrationTableSQL)
	return err
}

// loadMigrations loads all migration files from the embedded filesystem
func loadMigrations() ([]Migration, error) {
	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var migrations []Migration
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		migration, err := parseMigrationFile(entry.Name())
		if err != nil {
			return nil, fmt.Errorf("failed to parse migration %s: %w", entry.Name(), err)
		}

		migrations = append(migrations, migration)
	}

	// Sort migrations by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// parseMigrationFile parses a migration file and extracts up/down SQL
func parseMigrationFile(filename string) (Migration, error) {
	// Parse version and name from filename (e.g., "001_initial_schema.sql")
	var version int
	var name string
	_, err := fmt.Sscanf(filename, "%d_%s", &version, &name)
	if err != nil {
		return Migration{}, fmt.Errorf("invalid migration filename format: %s", filename)
	}

	// Remove .sql extension from name
	name = strings.TrimSuffix(name, ".sql")

	// Read file content
	content, err := fs.ReadFile(migrationsFS, filepath.Join("migrations", filename))
	if err != nil {
		return Migration{}, fmt.Errorf("failed to read migration file: %w", err)
	}

	// Split content into up and down migrations
	upSQL, downSQL, description := splitMigrationContent(string(content))

	return Migration{
		Version:     version,
		Name:        name,
		UpSQL:       upSQL,
		DownSQL:     downSQL,
		Description: description,
	}, nil
}

// splitMigrationContent splits migration content into up and down SQL
func splitMigrationContent(content string) (upSQL, downSQL, description string) {
	lines := strings.Split(content, "\n")
	var currentSection string
	var upLines, downLines []string
	var descLine string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Extract description from comment
		if strings.HasPrefix(trimmed, "-- Description:") {
			descLine = strings.TrimPrefix(trimmed, "-- Description:")
			descLine = strings.TrimSpace(descLine)
			continue
		}

		// Detect section markers
		if strings.Contains(trimmed, "+migrate Up") {
			currentSection = "up"
			continue
		}
		if strings.Contains(trimmed, "+migrate Down") {
			currentSection = "down"
			continue
		}

		// Skip migration metadata comments
		if strings.HasPrefix(trimmed, "-- Migration:") ||
			strings.HasPrefix(trimmed, "-- Author:") ||
			strings.HasPrefix(trimmed, "-- Date:") {
			continue
		}

		// Add line to appropriate section
		switch currentSection {
		case "up":
			upLines = append(upLines, line)
		case "down":
			downLines = append(downLines, line)
		}
	}

	upSQL = strings.TrimSpace(strings.Join(upLines, "\n"))
	downSQL = strings.TrimSpace(strings.Join(downLines, "\n"))
	description = descLine

	return
}

// getAppliedMigrations retrieves all applied migrations from the database
func getAppliedMigrations(ctx context.Context, db *sql.DB) ([]Migration, error) {
	query := `
		SELECT version, name, applied_at, COALESCE(description, '')
		FROM schema_migrations
		ORDER BY version ASC
	`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var migrations []Migration
	for rows.Next() {
		var m Migration
		if err := rows.Scan(&m.Version, &m.Name, &m.AppliedAt, &m.Description); err != nil {
			return nil, err
		}
		migrations = append(migrations, m)
	}

	return migrations, rows.Err()
}

// applyMigration applies a single migration within a transaction
func applyMigration(ctx context.Context, db *sql.DB, migration Migration) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Execute migration SQL
	if _, err := tx.ExecContext(ctx, migration.UpSQL); err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Record migration
	recordSQL := `
		INSERT INTO schema_migrations (version, name, description)
		VALUES (?, ?, ?)
	`
	if _, err := tx.ExecContext(ctx, recordSQL, migration.Version, migration.Name, migration.Description); err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	return tx.Commit()
}

// rollbackMigration rolls back a single migration within a transaction
func rollbackMigration(ctx context.Context, db *sql.DB, migration Migration) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Execute rollback SQL
	if _, err := tx.ExecContext(ctx, migration.DownSQL); err != nil {
		return fmt.Errorf("failed to execute rollback SQL: %w", err)
	}

	// Remove migration record
	deleteSQL := `DELETE FROM schema_migrations WHERE version = ?`
	if _, err := tx.ExecContext(ctx, deleteSQL, migration.Version); err != nil {
		return fmt.Errorf("failed to remove migration record: %w", err)
	}

	return tx.Commit()
}
