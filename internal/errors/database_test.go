package errors

import (
	"errors"
	"strings"
	"testing"
)

func TestDatabaseError(t *testing.T) {
	t.Run("NewDBConnectionError", func(t *testing.T) {
		originalErr := errors.New("connection refused")
		err := NewDBConnectionError("postgres", originalErr)

		if err.Code() != ErrCodeDBConnection {
			t.Errorf("expected code %s, got %s", ErrCodeDBConnection, err.Code())
		}

		if err.Severity() != SeverityCritical {
			t.Errorf("expected severity %s, got %s", SeverityCritical, err.Severity())
		}

		if !err.IsRetryable() {
			t.Error("connection error should be retryable")
		}

		if err.Unwrap() != originalErr {
			t.Error("expected error to wrap original error")
		}

		userMsg := err.UserMessage()
		if !strings.Contains(userMsg, "connection") {
			t.Errorf("user message should mention connection: %s", userMsg)
		}
	})

	t.Run("NewDBQueryError", func(t *testing.T) {
		originalErr := errors.New("syntax error")
		err := NewDBQueryError("SELECT", "users", originalErr)

		if err.Code() != ErrCodeDBQuery {
			t.Errorf("expected code %s, got %s", ErrCodeDBQuery, err.Code())
		}

		if err.Table != "users" {
			t.Errorf("expected Table 'users', got '%s'", err.Table)
		}

		if err.Operation != "SELECT" {
			t.Errorf("expected Operation 'SELECT', got '%s'", err.Operation)
		}

		if err.Unwrap() != originalErr {
			t.Error("expected error to wrap original error")
		}

		errMsg := err.Error()
		if !strings.Contains(errMsg, "SELECT") || !strings.Contains(errMsg, "users") {
			t.Errorf("error message should contain operation and table: %s", errMsg)
		}
	})

	t.Run("NewDBNotFoundError", func(t *testing.T) {
		err := NewDBNotFoundError("products", "id=123")

		if err.Code() != ErrCodeDBNotFound {
			t.Errorf("expected code %s, got %s", ErrCodeDBNotFound, err.Code())
		}

		if err.Table != "products" {
			t.Errorf("expected Table 'products', got '%s'", err.Table)
		}

		if err.Severity() != SeverityLow {
			t.Errorf("expected severity %s, got %s", SeverityLow, err.Severity())
		}

		errMsg := err.Error()
		if !strings.Contains(errMsg, "id=123") {
			t.Errorf("error message should contain identifier: %s", errMsg)
		}

		userMsg := err.UserMessage()
		if !strings.Contains(userMsg, "not exist") {
			t.Errorf("user message should mention resource not exists: %s", userMsg)
		}
	})

	t.Run("NewDBDuplicateError", func(t *testing.T) {
		err := NewDBDuplicateError("users", "email", "test@example.com")

		if err.Code() != ErrCodeDBDuplicate {
			t.Errorf("expected code %s, got %s", ErrCodeDBDuplicate, err.Code())
		}

		if err.Table != "users" {
			t.Errorf("expected Table 'users', got '%s'", err.Table)
		}

		if err.Severity() != SeverityMedium {
			t.Errorf("expected severity %s, got %s", SeverityMedium, err.Severity())
		}

		errMsg := err.Error()
		if !strings.Contains(errMsg, "test@example.com") || !strings.Contains(errMsg, "email") {
			t.Errorf("error message should contain field and value: %s", errMsg)
		}

		userMsg := err.UserMessage()
		if !strings.Contains(userMsg, "already exists") {
			t.Errorf("user message should mention duplicate: %s", userMsg)
		}
	})

	t.Run("NewDBConstraintError", func(t *testing.T) {
		originalErr := errors.New("foreign key violation")
		err := NewDBConstraintError("orders", "fk_user_id", originalErr)

		if err.Code() != ErrCodeDBConstraint {
			t.Errorf("expected code %s, got %s", ErrCodeDBConstraint, err.Code())
		}

		if err.Table != "orders" {
			t.Errorf("expected Table 'orders', got '%s'", err.Table)
		}

		if err.Constraint != "fk_user_id" {
			t.Errorf("expected Constraint 'fk_user_id', got '%s'", err.Constraint)
		}

		if err.Unwrap() != originalErr {
			t.Error("expected error to wrap original error")
		}

		userMsg := err.UserMessage()
		if !strings.Contains(userMsg, "integrity constraint") {
			t.Errorf("user message should mention constraint violation: %s", userMsg)
		}
	})

	t.Run("NewDBTransactionError", func(t *testing.T) {
		originalErr := errors.New("deadlock detected")
		err := NewDBTransactionError("commit", originalErr)

		if err.Code() != ErrCodeDBTransaction {
			t.Errorf("expected code %s, got %s", ErrCodeDBTransaction, err.Code())
		}

		if err.Operation != "commit" {
			t.Errorf("expected Operation 'commit', got '%s'", err.Operation)
		}

		if err.Severity() != SeverityHigh {
			t.Errorf("expected severity %s, got %s", SeverityHigh, err.Severity())
		}

		if !err.IsRetryable() {
			t.Error("transaction error should be retryable")
		}

		if err.Unwrap() != originalErr {
			t.Error("expected error to wrap original error")
		}

		userMsg := err.UserMessage()
		if !strings.Contains(userMsg, "No changes were made") {
			t.Errorf("user message should mention rollback: %s", userMsg)
		}
	})

	t.Run("WithTable", func(t *testing.T) {
		err := NewDatabaseError(ErrCodeDBQuery, "test")
		err.WithTable("customers")

		if err.Table != "customers" {
			t.Errorf("expected Table 'customers', got '%s'", err.Table)
		}
	})

	t.Run("WithQuery", func(t *testing.T) {
		query := "SELECT * FROM users WHERE id = $1"
		err := NewDatabaseError(ErrCodeDBQuery, "test")
		err.WithQuery(query)

		if err.Query != query {
			t.Errorf("expected Query '%s', got '%s'", query, err.Query)
		}
	})

	t.Run("WithOperation", func(t *testing.T) {
		err := NewDatabaseError(ErrCodeDBQuery, "test")
		err.WithOperation("INSERT")

		if err.Operation != "INSERT" {
			t.Errorf("expected Operation 'INSERT', got '%s'", err.Operation)
		}
	})

	t.Run("WithConstraint", func(t *testing.T) {
		err := NewDatabaseError(ErrCodeDBConstraint, "test")
		err.WithConstraint("unique_email")

		if err.Constraint != "unique_email" {
			t.Errorf("expected Constraint 'unique_email', got '%s'", err.Constraint)
		}
	})

	t.Run("Method chaining", func(t *testing.T) {
		err := NewDatabaseError(ErrCodeDBQuery, "test").
			WithTable("users").
			WithQuery("SELECT * FROM users").
			WithOperation("SELECT")

		if err.Table != "users" {
			t.Error("expected Table to be set via chaining")
		}
		if err.Query != "SELECT * FROM users" {
			t.Error("expected Query to be set via chaining")
		}
		if err.Operation != "SELECT" {
			t.Error("expected Operation to be set via chaining")
		}
	})

	t.Run("Retryability", func(t *testing.T) {
		// Connection errors should be retryable
		connErr := NewDBConnectionError("postgres", errors.New("refused"))
		if !connErr.IsRetryable() {
			t.Error("connection error should be retryable")
		}

		// Transaction errors should be retryable
		txErr := NewDBTransactionError("commit", errors.New("deadlock"))
		if !txErr.IsRetryable() {
			t.Error("transaction error should be retryable")
		}

		// Constraint violations should not be retryable
		constraintErr := NewDBConstraintError("table", "constraint", errors.New("violation"))
		if constraintErr.IsRetryable() {
			t.Error("constraint error should not be retryable")
		}

		// Not found should not be retryable
		notFoundErr := NewDBNotFoundError("table", "id")
		if notFoundErr.IsRetryable() {
			t.Error("not found error should not be retryable")
		}
	})
}

func TestDatabaseErrorWrapping(t *testing.T) {
	t.Run("Wrap database error", func(t *testing.T) {
		dbErr := NewDBNotFoundError("users", "id=123")
		wrappedErr := Wrap(dbErr, ErrCodeDBQuery, "failed to fetch user")

		var baseErr *BaseError
		if !As(wrappedErr, &baseErr) {
			t.Fatal("expected BaseError")
		}

		// Check that we can still extract the original database error
		var originalDBErr *DatabaseError
		if !As(wrappedErr, &originalDBErr) {
			t.Fatal("expected to extract DatabaseError from chain")
		}

		if originalDBErr.Table != "users" {
			t.Errorf("expected Table 'users', got '%s'", originalDBErr.Table)
		}
	})
}

func BenchmarkNewDatabaseError(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewDBNotFoundError("users", "id=123")
	}
}

func ExampleNewDBNotFoundError() {
	err := NewDBNotFoundError("products", "sku=ABC123")
	println(err.Error())
	println(err.UserMessage())
	println(err.Table)
}

func ExampleNewDBConnectionError() {
	originalErr := errors.New("connection timeout")
	err := NewDBConnectionError("postgres", originalErr)
	println(err.Code())
	println(err.IsRetryable())
}
