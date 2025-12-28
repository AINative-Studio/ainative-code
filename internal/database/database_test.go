package database

import (
	"context"
	"testing"
)

func TestNewDB(t *testing.T) {
	sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	db := NewDB(sqlDB)
	if db == nil {
		t.Fatal("expected non-nil DB instance")
	}

	if db.Queries == nil {
		t.Error("expected Queries to be initialized")
	}

	if db.db != sqlDB {
		t.Error("expected db field to match input")
	}
}

func TestNewFromConfig(t *testing.T) {
	config := DefaultConfig(":memory:")
	db, err := NewFromConfig(config)
	if err != nil {
		t.Fatalf("failed to create DB from config: %v", err)
	}
	defer db.Close()

	if db == nil {
		t.Fatal("expected non-nil DB instance")
	}

	if err := db.Ping(); err != nil {
		t.Errorf("failed to ping database: %v", err)
	}
}

func TestInitialize(t *testing.T) {
	config := DefaultConfig(":memory:")
	db, err := Initialize(config)
	if err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}
	defer db.Close()

	// Verify migrations were run
	var count int
	err = db.db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query migrations: %v", err)
	}
	if count == 0 {
		t.Error("expected migrations to be applied")
	}

	// Verify tables exist
	tables := []string{"metadata", "sessions", "messages", "tool_executions"}
	for _, table := range tables {
		err := db.db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&count)
		if err != nil {
			t.Fatalf("failed to check for table %s: %v", table, err)
		}
		if count != 1 {
			t.Errorf("expected table %s to exist after initialization", table)
		}
	}
}

func TestInitializeContext(t *testing.T) {
	ctx := context.Background()
	config := DefaultConfig(":memory:")
	db, err := InitializeContext(ctx, config)
	if err != nil {
		t.Fatalf("failed to initialize database with context: %v", err)
	}
	defer db.Close()

	if db == nil {
		t.Fatal("expected non-nil DB instance")
	}
}

func TestDB_Close(t *testing.T) {
	config := DefaultConfig(":memory:")
	db, err := NewFromConfig(config)
	if err != nil {
		t.Fatalf("failed to create DB: %v", err)
	}

	if err := db.Close(); err != nil {
		t.Errorf("failed to close DB: %v", err)
	}

	// Verify database is closed
	if err := db.Ping(); err == nil {
		t.Error("expected error pinging closed database")
	}
}

func TestDB_Health(t *testing.T) {
	config := DefaultConfig(":memory:")
	db, err := NewFromConfig(config)
	if err != nil {
		t.Fatalf("failed to create DB: %v", err)
	}
	defer db.Close()

	if err := db.Health(); err != nil {
		t.Errorf("health check failed: %v", err)
	}
}

func TestDB_DB(t *testing.T) {
	sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	db := NewDB(sqlDB)
	if db.DB() != sqlDB {
		t.Error("DB() should return the underlying sql.DB instance")
	}
}

func TestDB_WithTx(t *testing.T) {
	config := DefaultConfig(":memory:")
	db, err := Initialize(config)
	if err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Test successful transaction
	err = db.WithTx(ctx, func(q *Queries) error {
		// Insert metadata
		return q.SetMetadata(ctx, SetMetadataParams{
			Key:   "test_key",
			Value: "test_value",
		})
	})
	if err != nil {
		t.Fatalf("transaction failed: %v", err)
	}

	// Verify data was committed
	metadata, err := db.GetMetadata(ctx, "test_key")
	if err != nil {
		t.Fatalf("failed to get metadata: %v", err)
	}
	if metadata.Value != "test_value" {
		t.Errorf("expected value 'test_value', got '%s'", metadata.Value)
	}
}

func TestDB_WithTx_Rollback(t *testing.T) {
	config := DefaultConfig(":memory:")
	db, err := Initialize(config)
	if err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Test transaction rollback on error
	err = db.WithTx(ctx, func(q *Queries) error {
		// Insert metadata
		if err := q.SetMetadata(ctx, SetMetadataParams{
			Key:   "rollback_key",
			Value: "rollback_value",
		}); err != nil {
			return err
		}

		// Return error to trigger rollback
		return &testError{msg: "intentional error"}
	})

	if err == nil {
		t.Fatal("expected transaction to fail")
	}

	// Verify data was rolled back
	_, err = db.GetMetadata(ctx, "rollback_key")
	if err == nil {
		t.Error("expected metadata to not exist after rollback")
	}
}

func TestDB_Stats(t *testing.T) {
	config := DefaultConfig(":memory:")
	db, err := NewFromConfig(config)
	if err != nil {
		t.Fatalf("failed to create DB: %v", err)
	}
	defer db.Close()

	stats := db.Stats()
	if stats.MaxOpenConnections != config.MaxOpenConns {
		t.Errorf("expected MaxOpenConnections %d, got %d", config.MaxOpenConns, stats.MaxOpenConnections)
	}
}

func TestDB_Ping(t *testing.T) {
	config := DefaultConfig(":memory:")
	db, err := NewFromConfig(config)
	if err != nil {
		t.Fatalf("failed to create DB: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Errorf("ping failed: %v", err)
	}
}

func TestDB_PingContext(t *testing.T) {
	config := DefaultConfig(":memory:")
	db, err := NewFromConfig(config)
	if err != nil {
		t.Fatalf("failed to create DB: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := db.PingContext(ctx); err != nil {
		t.Errorf("ping context failed: %v", err)
	}
}

// testError is a custom error type for testing
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
