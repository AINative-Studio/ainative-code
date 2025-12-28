package database

import (
	"context"
	"database/sql"
	"testing"
)

func TestMigrate(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	if err := Migrate(db); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	// Verify migration table exists
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='schema_migrations'").Scan(&count)
	if err != nil {
		t.Fatalf("failed to check for migration table: %v", err)
	}
	if count != 1 {
		t.Errorf("expected schema_migrations table to exist")
	}

	// Verify our tables were created
	tables := []string{"metadata", "sessions", "messages", "tool_executions"}
	for _, table := range tables {
		err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&count)
		if err != nil {
			t.Fatalf("failed to check for table %s: %v", table, err)
		}
		if count != 1 {
			t.Errorf("expected table %s to exist", table)
		}
	}
}

func TestMigrateContext(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	if err := MigrateContext(ctx, db); err != nil {
		t.Fatalf("failed to run migrations with context: %v", err)
	}

	// Verify migration was applied
	var version int
	err := db.QueryRow("SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1").Scan(&version)
	if err != nil {
		t.Fatalf("failed to get migration version: %v", err)
	}
	if version != 1 {
		t.Errorf("expected version 1, got %d", version)
	}
}

func TestMigrate_Idempotent(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Run migrations twice
	if err := Migrate(db); err != nil {
		t.Fatalf("failed to run migrations first time: %v", err)
	}

	if err := Migrate(db); err != nil {
		t.Fatalf("failed to run migrations second time: %v", err)
	}

	// Verify migration was only applied once
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count)
	if err != nil {
		t.Fatalf("failed to count migrations: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 migration record, got %d", count)
	}
}

func TestMigrate_NilDB(t *testing.T) {
	err := Migrate(nil)
	if err == nil {
		t.Error("expected error for nil database, got nil")
	}
}

func TestGetStatus(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Get status before migration
	status, err := GetStatus(db)
	if err != nil {
		t.Fatalf("failed to get status: %v", err)
	}

	if status.CurrentVersion != 0 {
		t.Errorf("expected current version 0, got %d", status.CurrentVersion)
	}

	if len(status.Applied) != 0 {
		t.Errorf("expected 0 applied migrations, got %d", len(status.Applied))
	}

	if len(status.Pending) == 0 {
		t.Error("expected pending migrations, got none")
	}

	// Apply migrations
	if err := Migrate(db); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	// Get status after migration
	status, err = GetStatus(db)
	if err != nil {
		t.Fatalf("failed to get status after migration: %v", err)
	}

	if status.CurrentVersion != 1 {
		t.Errorf("expected current version 1, got %d", status.CurrentVersion)
	}

	if len(status.Applied) != 1 {
		t.Errorf("expected 1 applied migration, got %d", len(status.Applied))
	}

	if len(status.Pending) != 0 {
		t.Errorf("expected 0 pending migrations, got %d", len(status.Pending))
	}
}

func TestGetStatusContext(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	status, err := GetStatusContext(ctx, db)
	if err != nil {
		t.Fatalf("failed to get status with context: %v", err)
	}

	if status == nil {
		t.Fatal("expected non-nil status")
	}
}

func TestGetStatus_NilDB(t *testing.T) {
	_, err := GetStatus(nil)
	if err == nil {
		t.Error("expected error for nil database, got nil")
	}
}

func TestRollback(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Apply migrations
	if err := Migrate(db); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	// Verify tables exist
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='sessions'").Scan(&count)
	if err != nil {
		t.Fatalf("failed to check for sessions table: %v", err)
	}
	if count != 1 {
		t.Error("expected sessions table to exist before rollback")
	}

	// Rollback
	if err := Rollback(db); err != nil {
		t.Fatalf("failed to rollback migration: %v", err)
	}

	// Verify tables were dropped
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='sessions'").Scan(&count)
	if err != nil {
		t.Fatalf("failed to check for sessions table: %v", err)
	}
	if count != 0 {
		t.Error("expected sessions table to be dropped after rollback")
	}

	// Verify migration record was removed
	err = db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count)
	if err != nil {
		t.Fatalf("failed to count migrations: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 migration records, got %d", count)
	}
}

func TestRollback_NoMigrations(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Ensure migration table exists
	if err := Migrate(db); err != nil {
		// This will create the migration table but no migrations to rollback
	}

	// Rollback all
	if err := Rollback(db); err != nil {
		t.Fatalf("first rollback failed: %v", err)
	}

	// Try to rollback when no migrations exist
	err := Rollback(db)
	if err == nil {
		t.Error("expected error when rolling back with no migrations, got nil")
	}
}

func TestRollback_NilDB(t *testing.T) {
	err := Rollback(nil)
	if err == nil {
		t.Error("expected error for nil database, got nil")
	}
}

func TestLoadMigrations(t *testing.T) {
	migrations, err := loadMigrations()
	if err != nil {
		t.Fatalf("failed to load migrations: %v", err)
	}

	if len(migrations) == 0 {
		t.Error("expected at least one migration")
	}

	// Verify migrations are sorted by version
	for i := 1; i < len(migrations); i++ {
		if migrations[i].Version <= migrations[i-1].Version {
			t.Errorf("migrations not sorted: version %d comes after %d", migrations[i].Version, migrations[i-1].Version)
		}
	}

	// Verify first migration has required fields
	first := migrations[0]
	if first.Version != 1 {
		t.Errorf("expected first migration version 1, got %d", first.Version)
	}
	if first.Name == "" {
		t.Error("expected migration name to be set")
	}
	if first.UpSQL == "" {
		t.Error("expected migration UpSQL to be set")
	}
	if first.DownSQL == "" {
		t.Error("expected migration DownSQL to be set")
	}
}

func TestAppliedMigrations(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Ensure migration table exists
	if err := ensureMigrationTable(ctx, db); err != nil {
		t.Fatalf("failed to create migration table: %v", err)
	}

	// Should have no applied migrations initially
	applied, err := getAppliedMigrations(ctx, db)
	if err != nil {
		t.Fatalf("failed to get applied migrations: %v", err)
	}
	if len(applied) != 0 {
		t.Errorf("expected 0 applied migrations, got %d", len(applied))
	}

	// Insert a test migration record
	_, err = db.Exec(
		"INSERT INTO schema_migrations (version, name, description) VALUES (?, ?, ?)",
		1, "test_migration", "Test migration",
	)
	if err != nil {
		t.Fatalf("failed to insert test migration: %v", err)
	}

	// Should now have one applied migration
	applied, err = getAppliedMigrations(ctx, db)
	if err != nil {
		t.Fatalf("failed to get applied migrations: %v", err)
	}
	if len(applied) != 1 {
		t.Errorf("expected 1 applied migration, got %d", len(applied))
	}

	migration := applied[0]
	if migration.Version != 1 {
		t.Errorf("expected version 1, got %d", migration.Version)
	}
	if migration.Name != "test_migration" {
		t.Errorf("expected name 'test_migration', got '%s'", migration.Name)
	}
	if migration.AppliedAt.IsZero() {
		t.Error("expected AppliedAt to be set")
	}
}

func TestEnsureMigrationTable(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create migration table
	if err := ensureMigrationTable(ctx, db); err != nil {
		t.Fatalf("failed to ensure migration table: %v", err)
	}

	// Verify table exists
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='schema_migrations'").Scan(&count)
	if err != nil {
		t.Fatalf("failed to check for migration table: %v", err)
	}
	if count != 1 {
		t.Error("expected schema_migrations table to exist")
	}

	// Should be idempotent
	if err := ensureMigrationTable(ctx, db); err != nil {
		t.Errorf("failed to ensure migration table second time: %v", err)
	}
}

func TestSplitMigrationContent(t *testing.T) {
	content := `-- Migration: 001_test
-- Description: Test migration for parsing
-- Author: Test
-- Date: 2025-01-01

-- +migrate Up
CREATE TABLE test (id INTEGER PRIMARY KEY);

-- +migrate Down
DROP TABLE test;
`

	upSQL, downSQL, description := splitMigrationContent(content)

	if upSQL == "" {
		t.Error("expected upSQL to be set")
	}
	if downSQL == "" {
		t.Error("expected downSQL to be set")
	}
	if description != "Test migration for parsing" {
		t.Errorf("expected description 'Test migration for parsing', got '%s'", description)
	}

	// Verify SQL content
	if !containsMiddle(upSQL, "CREATE TABLE test") {
		t.Error("upSQL should contain CREATE TABLE statement")
	}
	if !containsMiddle(downSQL, "DROP TABLE test") {
		t.Error("downSQL should contain DROP TABLE statement")
	}
}

// setupTestDB creates an in-memory database for testing
func setupTestDB(t *testing.T) *sql.DB {
	config := DefaultConfig(":memory:")
	db, err := Connect(config)
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}
	return db
}
