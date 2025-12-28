package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/AINative-studio/ainative-code/internal/errors"
)

// DB wraps the SQLC Queries interface with additional functionality
type DB struct {
	*Queries
	db *sql.DB
}

// NewDB creates a new database instance with the given connection
func NewDB(sqlDB *sql.DB) *DB {
	return &DB{
		Queries: New(sqlDB),
		db:      sqlDB,
	}
}

// NewFromConfig creates a new database instance from configuration
func NewFromConfig(config *ConnectionConfig) (*DB, error) {
	sqlDB, err := Connect(config)
	if err != nil {
		return nil, err
	}

	return &DB{
		Queries: New(sqlDB),
		db:      sqlDB,
	}, nil
}

// Initialize sets up the database connection and runs migrations
func Initialize(config *ConnectionConfig) (*DB, error) {
	return InitializeContext(context.Background(), config)
}

// InitializeContext sets up the database connection and runs migrations with context
func InitializeContext(ctx context.Context, config *ConnectionConfig) (*DB, error) {
	// Connect to database
	db, err := Connect(config)
	if err != nil {
		return nil, err
	}

	// Run migrations
	if err := MigrateContext(ctx, db); err != nil {
		db.Close()
		return nil, err
	}

	return &DB{
		Queries: New(db),
		db:      db,
	}, nil
}

// Close closes the database connection
func (d *DB) Close() error {
	return Close(d.db)
}

// Health performs a health check on the database
func (d *DB) Health() error {
	return HealthCheck(d.db)
}

// DB returns the underlying *sql.DB instance
func (d *DB) DB() *sql.DB {
	return d.db
}

// WithTx executes a function within a database transaction
func (d *DB) WithTx(ctx context.Context, fn func(*Queries) error) error {
	return d.WithTxOptions(ctx, nil, fn)
}

// WithTxOptions executes a function within a database transaction with options
func (d *DB) WithTxOptions(ctx context.Context, opts *sql.TxOptions, fn func(*Queries) error) error {
	tx, err := d.db.BeginTx(ctx, opts)
	if err != nil {
		return errors.NewDBTransactionError("begin", err)
	}

	// Create a new Queries instance for the transaction using the generated WithTx method
	qtx := d.Queries.WithTx(tx)

	// Execute the function
	if err := fn(qtx); err != nil {
		// Attempt to rollback
		if rbErr := tx.Rollback(); rbErr != nil {
			return errors.NewDBTransactionError(
				"rollback after error",
				fmt.Errorf("transaction error: %w, rollback error: %v", err, rbErr),
			)
		}
		return err
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return errors.NewDBTransactionError("commit", err)
	}

	return nil
}

// ExecInTx is a helper that executes SQL in a transaction
func (d *DB) ExecInTx(ctx context.Context, query string, args ...interface{}) error {
	return d.WithTx(ctx, func(q *Queries) error {
		_, err := d.db.ExecContext(ctx, query, args...)
		return err
	})
}

// Stats returns database statistics
func (d *DB) Stats() sql.DBStats {
	return d.db.Stats()
}

// Ping verifies the database connection is alive
func (d *DB) Ping() error {
	return d.db.Ping()
}

// PingContext verifies the database connection is alive with context
func (d *DB) PingContext(ctx context.Context) error {
	return d.db.PingContext(ctx)
}
