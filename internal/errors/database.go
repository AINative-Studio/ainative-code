package errors

import "fmt"

// DatabaseError represents database-related errors
type DatabaseError struct {
	*BaseError
	Table      string
	Query      string
	Operation  string
	Constraint string
}

// NewDatabaseError creates a new database error
func NewDatabaseError(code ErrorCode, message string) *DatabaseError {
	baseErr := newError(code, message, SeverityHigh, false)
	return &DatabaseError{
		BaseError: baseErr,
	}
}

// NewDBConnectionError creates an error for database connection failures
func NewDBConnectionError(dbName string, cause error) *DatabaseError {
	msg := fmt.Sprintf("Failed to connect to database '%s'", dbName)
	userMsg := "Database connection error: Unable to connect to the database. Please check your connection settings."

	baseErr := newError(ErrCodeDBConnection, msg, SeverityCritical, true)
	baseErr.cause = cause
	baseErr.userMsg = userMsg

	return &DatabaseError{
		BaseError: baseErr,
	}
}

// NewDBQueryError creates an error for query execution failures
func NewDBQueryError(operation, table string, cause error) *DatabaseError {
	msg := fmt.Sprintf("Database query failed: %s on table '%s'", operation, table)
	userMsg := "Database error: The requested operation could not be completed. Please try again."

	baseErr := newError(ErrCodeDBQuery, msg, SeverityMedium, false)
	baseErr.cause = cause
	baseErr.userMsg = userMsg

	return &DatabaseError{
		BaseError: baseErr,
		Table:     table,
		Operation: operation,
	}
}

// NewDBNotFoundError creates an error for record not found
func NewDBNotFoundError(table, identifier string) *DatabaseError {
	msg := fmt.Sprintf("Record not found in table '%s': %s", table, identifier)
	userMsg := "Not found: The requested resource does not exist."

	err := NewDatabaseError(ErrCodeDBNotFound, msg)
	err.userMsg = userMsg
	err.Table = table
	err.severity = SeverityLow
	return err
}

// NewDBDuplicateError creates an error for duplicate key violations
func NewDBDuplicateError(table, field, value string) *DatabaseError {
	msg := fmt.Sprintf("Duplicate entry '%s' for field '%s' in table '%s'", value, field, table)
	userMsg := fmt.Sprintf("Duplicate entry: A record with this %s already exists.", field)

	err := NewDatabaseError(ErrCodeDBDuplicate, msg)
	err.userMsg = userMsg
	err.Table = table
	err.severity = SeverityMedium
	return err
}

// NewDBConstraintError creates an error for constraint violations
func NewDBConstraintError(table, constraint string, cause error) *DatabaseError {
	msg := fmt.Sprintf("Constraint violation in table '%s': %s", table, constraint)
	userMsg := "Database error: The operation violates a data integrity constraint. Please check your input."

	baseErr := newError(ErrCodeDBConstraint, msg, SeverityMedium, false)
	baseErr.cause = cause
	baseErr.userMsg = userMsg

	return &DatabaseError{
		BaseError:  baseErr,
		Table:      table,
		Constraint: constraint,
	}
}

// NewDBTransactionError creates an error for transaction failures
func NewDBTransactionError(operation string, cause error) *DatabaseError {
	msg := fmt.Sprintf("Transaction failed during %s", operation)
	userMsg := "Database transaction error: The operation could not be completed. No changes were made."

	baseErr := newError(ErrCodeDBTransaction, msg, SeverityHigh, true)
	baseErr.cause = cause
	baseErr.userMsg = userMsg

	return &DatabaseError{
		BaseError: baseErr,
		Operation: operation,
	}
}

// WithTable sets the database table name
func (e *DatabaseError) WithTable(table string) *DatabaseError {
	e.Table = table
	return e
}

// WithQuery sets the query that failed
func (e *DatabaseError) WithQuery(query string) *DatabaseError {
	e.Query = query
	return e
}

// WithOperation sets the database operation
func (e *DatabaseError) WithOperation(operation string) *DatabaseError {
	e.Operation = operation
	return e
}

// WithConstraint sets the constraint that was violated
func (e *DatabaseError) WithConstraint(constraint string) *DatabaseError {
	e.Constraint = constraint
	return e
}
