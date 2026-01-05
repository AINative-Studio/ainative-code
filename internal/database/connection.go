package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
	"github.com/AINative-studio/ainative-code/internal/errors"
)

// ConnectionConfig holds database connection configuration
type ConnectionConfig struct {
	Path            string        // Database file path
	MaxOpenConns    int           // Maximum number of open connections
	MaxIdleConns    int           // Maximum number of idle connections
	ConnMaxLifetime time.Duration // Maximum lifetime of a connection
	ConnMaxIdleTime time.Duration // Maximum idle time of a connection
	BusyTimeout     int           // SQLite busy timeout in milliseconds
	JournalMode     string        // SQLite journal mode (WAL, DELETE, etc.)
	Synchronous     string        // SQLite synchronous mode (NORMAL, FULL, OFF)
}

// DefaultConfig returns a default database configuration
func DefaultConfig(dbPath string) *ConnectionConfig {
	return &ConnectionConfig{
		Path:            dbPath,
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 10 * time.Minute,
		BusyTimeout:     5000,      // 5 seconds
		JournalMode:     "WAL",     // Write-Ahead Logging for better concurrency
		Synchronous:     "NORMAL",  // Balance between safety and performance
	}
}

// Connect establishes a connection to the SQLite database
func Connect(config *ConnectionConfig) (*sql.DB, error) {
	if config == nil {
		config = DefaultConfig(":memory:")
	}

	// Ensure the directory exists if not using in-memory database
	if config.Path != ":memory:" {
		dir := filepath.Dir(config.Path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, errors.NewDBConnectionError(config.Path, err)
		}
	}

	// Build DSN with connection parameters
	dsn := buildDSN(config)

	// Open database connection
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, errors.NewDBConnectionError(config.Path, err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)
	db.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	// Verify the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, errors.NewDBConnectionError(config.Path, err)
	}

	// Set SQLite pragmas for optimal performance and safety
	if err := setPragmas(db, config); err != nil {
		db.Close()
		return nil, errors.NewDBConnectionError(config.Path, err)
	}

	return db, nil
}

// buildDSN constructs the SQLite DSN with connection parameters
func buildDSN(config *ConnectionConfig) string {
	dsn := config.Path + "?"
	params := []string{
		fmt.Sprintf("_busy_timeout=%d", config.BusyTimeout),
		"_txlock=immediate",
		"_foreign_keys=on",
		"_journal_mode=" + config.JournalMode,
		"_synchronous=" + config.Synchronous,
		"_fts5=on", // Enable FTS5 full-text search
	}

	for i, param := range params {
		if i > 0 {
			dsn += "&"
		}
		dsn += param
	}

	return dsn
}

// setPragmas sets SQLite pragmas for optimal configuration
func setPragmas(db *sql.DB, config *ConnectionConfig) error {
	pragmas := []string{
		"PRAGMA foreign_keys = ON",
		"PRAGMA busy_timeout = " + fmt.Sprintf("%d", config.BusyTimeout),
		"PRAGMA journal_mode = " + config.JournalMode,
		"PRAGMA synchronous = " + config.Synchronous,
		"PRAGMA cache_size = -64000", // 64MB cache
		"PRAGMA temp_store = MEMORY",
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			return fmt.Errorf("failed to set pragma: %s: %w", pragma, err)
		}
	}

	return nil
}

// Close gracefully closes the database connection
func Close(db *sql.DB) error {
	if db == nil {
		return nil
	}

	// Checkpoint WAL file if using WAL mode
	if _, err := db.Exec("PRAGMA wal_checkpoint(TRUNCATE)"); err != nil {
		// Log error but don't fail the close operation
		// In a production system, use proper logging here
	}

	return db.Close()
}

// HealthCheck verifies the database connection is healthy
func HealthCheck(db *sql.DB) error {
	if db == nil {
		return errors.NewDatabaseError(errors.ErrCodeDBConnection, "database connection is nil")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return errors.NewDBConnectionError("health check", err)
	}

	// Verify we can perform a simple query
	var result int
	if err := db.QueryRowContext(ctx, "SELECT 1").Scan(&result); err != nil {
		return errors.NewDBQueryError("health check", "test", err)
	}

	return nil
}
