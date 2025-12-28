package database

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	dbPath := "/tmp/test.db"
	config := DefaultConfig(dbPath)

	if config.Path != dbPath {
		t.Errorf("expected path %s, got %s", dbPath, config.Path)
	}

	if config.MaxOpenConns != 10 {
		t.Errorf("expected MaxOpenConns 10, got %d", config.MaxOpenConns)
	}

	if config.MaxIdleConns != 5 {
		t.Errorf("expected MaxIdleConns 5, got %d", config.MaxIdleConns)
	}

	if config.BusyTimeout != 5000 {
		t.Errorf("expected BusyTimeout 5000, got %d", config.BusyTimeout)
	}

	if config.JournalMode != "WAL" {
		t.Errorf("expected JournalMode WAL, got %s", config.JournalMode)
	}

	if config.Synchronous != "NORMAL" {
		t.Errorf("expected Synchronous NORMAL, got %s", config.Synchronous)
	}
}

func TestConnect_InMemory(t *testing.T) {
	config := DefaultConfig(":memory:")
	db, err := Connect(config)
	if err != nil {
		t.Fatalf("failed to connect to in-memory database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Errorf("failed to ping database: %v", err)
	}
}

func TestConnect_FileDatabase(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	config := DefaultConfig(dbPath)
	db, err := Connect(config)
	if err != nil {
		t.Fatalf("failed to connect to file database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Errorf("failed to ping database: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Errorf("database file was not created at %s", dbPath)
	}
}

func TestConnect_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "subdir", "test.db")

	config := DefaultConfig(dbPath)
	db, err := Connect(config)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer db.Close()

	// Verify directory was created
	dir := filepath.Dir(dbPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Errorf("directory was not created at %s", dir)
	}
}

func TestConnect_NilConfig(t *testing.T) {
	db, err := Connect(nil)
	if err != nil {
		t.Fatalf("failed to connect with nil config: %v", err)
	}
	defer db.Close()

	// Should default to in-memory
	if err := db.Ping(); err != nil {
		t.Errorf("failed to ping database: %v", err)
	}
}

func TestConnect_ConnectionPool(t *testing.T) {
	config := DefaultConfig(":memory:")
	config.MaxOpenConns = 15
	config.MaxIdleConns = 7
	config.ConnMaxLifetime = 2 * time.Hour
	config.ConnMaxIdleTime = 15 * time.Minute

	db, err := Connect(config)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer db.Close()

	stats := db.Stats()
	if stats.MaxOpenConnections != 15 {
		t.Errorf("expected MaxOpenConnections 15, got %d", stats.MaxOpenConnections)
	}
}

func TestBuildDSN(t *testing.T) {
	config := &ConnectionConfig{
		Path:        "/tmp/test.db",
		BusyTimeout: 5000,
		JournalMode: "WAL",
		Synchronous: "NORMAL",
	}

	dsn := buildDSN(config)

	expectedParams := []string{
		"_busy_timeout=5000",
		"_txlock=immediate",
		"_foreign_keys=on",
		"_journal_mode=WAL",
		"_synchronous=NORMAL",
	}

	for _, param := range expectedParams {
		if !containsParam(dsn, param) {
			t.Errorf("DSN missing expected parameter: %s", param)
		}
	}
}

func TestHealthCheck(t *testing.T) {
	config := DefaultConfig(":memory:")
	db, err := Connect(config)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer db.Close()

	if err := HealthCheck(db); err != nil {
		t.Errorf("health check failed: %v", err)
	}
}

func TestHealthCheck_NilDB(t *testing.T) {
	err := HealthCheck(nil)
	if err == nil {
		t.Error("expected error for nil database, got nil")
	}
}

func TestHealthCheck_ClosedDB(t *testing.T) {
	config := DefaultConfig(":memory:")
	db, err := Connect(config)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	db.Close()

	err = HealthCheck(db)
	if err == nil {
		t.Error("expected error for closed database, got nil")
	}
}

func TestClose(t *testing.T) {
	config := DefaultConfig(":memory:")
	db, err := Connect(config)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}

	if err := Close(db); err != nil {
		t.Errorf("failed to close database: %v", err)
	}

	// Verify database is closed
	if err := db.Ping(); err == nil {
		t.Error("expected error pinging closed database, got nil")
	}
}

func TestClose_NilDB(t *testing.T) {
	if err := Close(nil); err != nil {
		t.Errorf("expected nil error for closing nil database, got: %v", err)
	}
}

// Helper function to check if DSN contains a parameter
func containsParam(dsn, param string) bool {
	return contains(dsn, param)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
