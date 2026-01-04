package helpers

import (
	"context"
	"database/sql"
	"testing"

	"github.com/AINative-studio/ainative-code/internal/database"
	_ "github.com/mattn/go-sqlite3"
)

// SetupTestDB creates an in-memory SQLite database for testing
func SetupTestDB(t *testing.T) *database.DB {
	t.Helper()

	// Create in-memory database
	sqlDB, err := sql.Open("sqlite3", ":memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to open in-memory database: %v", err)
	}

	// Run migrations first
	ctx := context.Background()
	if err := database.MigrateContext(ctx, sqlDB); err != nil {
		sqlDB.Close()
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Wrap in our database type
	db := database.NewDB(sqlDB)

	// Register cleanup
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("Failed to close test database: %v", err)
		}
	})

	return db
}

// SetupTestDBWithData creates a test database and populates it with test data
func SetupTestDBWithData(t *testing.T) *database.DB {
	t.Helper()

	db := SetupTestDB(t)

	// TODO: Load test data from fixtures
	// This could be extended to load sessions.json and messages.json

	return db
}

// CleanupDB clears all data from the database tables
func CleanupDB(t *testing.T, db *database.DB) {
	t.Helper()

	ctx := context.Background()

	// Delete in reverse order of foreign key dependencies
	tables := []string{
		"messages",
		"sessions",
		"tool_executions",
	}

	for _, table := range tables {
		query := "DELETE FROM " + table
		if _, err := db.DB().ExecContext(ctx, query); err != nil {
			t.Logf("Warning: Failed to clean table %s: %v", table, err)
		}
	}
}

// AssertTableEmpty asserts that a table has no rows
func AssertTableEmpty(t *testing.T, db *database.DB, tableName string) {
	t.Helper()

	ctx := context.Background()
	query := "SELECT COUNT(*) FROM " + tableName

	var count int64
	err := db.DB().QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count rows in %s: %v", tableName, err)
	}

	if count != 0 {
		t.Errorf("Expected table %s to be empty, but found %d rows", tableName, count)
	}
}

// AssertTableRowCount asserts that a table has the expected number of rows
func AssertTableRowCount(t *testing.T, db *database.DB, tableName string, expected int64) {
	t.Helper()

	ctx := context.Background()
	query := "SELECT COUNT(*) FROM " + tableName

	var count int64
	err := db.DB().QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count rows in %s: %v", tableName, err)
	}

	if count != expected {
		t.Errorf("Expected table %s to have %d rows, but found %d", tableName, expected, count)
	}
}
