// Package helpers provides test helper utilities for integration tests.
package helpers

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/AINative-studio/ainative-code/internal/database"
	"github.com/stretchr/testify/require"
)

// SetupTestDB creates a temporary SQLite database for testing.
// Returns the database instance and a cleanup function.
func SetupTestDB(t *testing.T) (*database.DB, func()) {
	t.Helper()

	// Create temp directory for test database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Connect to database
	config := database.DefaultConfig(dbPath)
	conn, err := database.Connect(config)
	require.NoError(t, err, "Failed to connect to database")

	// Create DB wrapper
	db := database.NewDB(conn)

	// Run migrations
	ctx := context.Background()
	err = database.MigrateContext(ctx, conn)
	require.NoError(t, err, "Failed to run migrations")

	cleanup := func() {
		// Close database connection
		conn.Close()
		os.RemoveAll(tmpDir)
	}

	return db, cleanup
}

// SetupInMemoryDB creates an in-memory SQLite database for testing.
// This is faster than file-based databases but doesn't persist between runs.
func SetupInMemoryDB(t *testing.T) (*database.DB, func()) {
	t.Helper()

	// Connect to in-memory database
	config := database.DefaultConfig(":memory:")
	conn, err := database.Connect(config)
	require.NoError(t, err, "Failed to connect to in-memory database")

	// Create DB wrapper
	db := database.NewDB(conn)

	// Run migrations
	ctx := context.Background()
	err = database.MigrateContext(ctx, conn)
	require.NoError(t, err, "Failed to run migrations")

	cleanup := func() {
		// Close database connection
		conn.Close()
	}

	return db, cleanup
}
