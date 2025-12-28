# Testing Guide

This guide covers testing practices, strategies, and examples for the AINative Code project.

## Table of Contents

- [Testing Philosophy](#testing-philosophy)
- [Test Types](#test-types)
- [Running Tests](#running-tests)
- [Writing Unit Tests](#writing-unit-tests)
- [Writing Integration Tests](#writing-integration-tests)
- [Test Coverage](#test-coverage)
- [Benchmarking](#benchmarking)
- [Mocking and Testing Patterns](#mocking-and-testing-patterns)
- [Best Practices](#best-practices)

## Testing Philosophy

AINative Code follows these testing principles:

1. **Test-Driven Development (TDD)**: Write tests before implementation when possible
2. **High Coverage**: Maintain 80%+ code coverage
3. **Fast Execution**: Unit tests should run in milliseconds
4. **Isolation**: Tests should be independent and not rely on external state
5. **Clear Assertions**: Test one thing per test function
6. **Readable**: Tests serve as documentation

## Test Types

### 1. Unit Tests

Test individual functions and methods in isolation.

**Location**: `*_test.go` files alongside source code

**Example**: `internal/logger/logger_test.go`

### 2. Integration Tests

Test interaction between multiple components.

**Location**: `tests/` directory or tagged with `//go:build integration`

**Example**: Database integration tests

### 3. Benchmark Tests

Measure performance of critical code paths.

**Location**: `*_bench_test.go` files

**Example**: `internal/logger/logger_bench_test.go`

### 4. Example Tests

Executable examples that appear in documentation.

**Location**: `*_example_test.go` files

**Example**: `internal/logger/example_test.go`

## Running Tests

### Basic Test Commands

```bash
# Run all tests
make test

# Run tests with verbose output
go test -v ./...

# Run tests in a specific package
go test -v ./internal/logger

# Run a specific test function
go test -v -run TestLogLevels ./internal/logger

# Run tests matching a pattern
go test -v -run "Test.*Context" ./...
```

### Test with Coverage

```bash
# Run tests with coverage
make test-coverage

# View coverage in browser
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Check coverage for specific package
go test -cover ./internal/logger

# Get detailed coverage report
go test -coverprofile=coverage.out -covermode=atomic ./...
go tool cover -func=coverage.out
```

### Integration Tests

```bash
# Run integration tests only
make test-integration

# Or with build tags
go test -v -tags=integration ./tests/...

# Skip integration tests
go test -v -short ./...  # Integration tests use t.Skip() in short mode
```

### Benchmark Tests

```bash
# Run all benchmarks
make test-benchmark

# Run specific benchmark
go test -bench=BenchmarkLogger ./internal/logger

# Run benchmarks with memory stats
go test -bench=. -benchmem ./...

# Compare benchmark results
go test -bench=. -benchmem ./... > old.txt
# Make changes...
go test -bench=. -benchmem ./... > new.txt
benchcmp old.txt new.txt  # Requires golang.org/x/tools/cmd/benchcmp
```

### Race Detection

```bash
# Run tests with race detector
go test -race ./...

# Race detection in CI
make ci  # Includes race detection
```

## Writing Unit Tests

### Basic Test Structure

```go
package mypackage

import "testing"

func TestFunctionName(t *testing.T) {
    // Arrange: Set up test data
    input := "test input"
    expected := "expected output"

    // Act: Execute the function
    result := MyFunction(input)

    // Assert: Verify the result
    if result != expected {
        t.Errorf("MyFunction(%q) = %q; want %q", input, result, expected)
    }
}
```

### Table-Driven Tests

**Recommended pattern for testing multiple scenarios:**

```go
func TestValidation(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {
            name:    "valid input",
            input:   "valid@example.com",
            wantErr: false,
        },
        {
            name:    "invalid email",
            input:   "invalid",
            wantErr: true,
        },
        {
            name:    "empty string",
            input:   "",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateEmail(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateEmail(%q) error = %v, wantErr %v",
                    tt.input, err, tt.wantErr)
            }
        })
    }
}
```

### Testing with Context

```go
func TestWithContext(t *testing.T) {
    ctx := context.Background()
    ctx = logger.WithRequestID(ctx, "req-123")

    log := logger.WithContext(ctx)
    log.Info("test message")

    // Verify context values
    requestID, ok := logger.GetRequestID(ctx)
    if !ok || requestID != "req-123" {
        t.Errorf("Expected request_id 'req-123', got %q (ok=%v)", requestID, ok)
    }
}
```

### Testing Error Cases

```go
func TestErrorHandling(t *testing.T) {
    // Test expected error
    err := FunctionThatShouldFail()
    if err == nil {
        t.Error("Expected error, got nil")
    }

    // Test specific error type
    var targetErr *MyCustomError
    if !errors.As(err, &targetErr) {
        t.Errorf("Expected MyCustomError, got %T", err)
    }

    // Test error message
    expectedMsg := "specific error message"
    if err.Error() != expectedMsg {
        t.Errorf("Error message = %q; want %q", err.Error(), expectedMsg)
    }
}
```

### Testing with Temporary Files

```go
func TestFileOperations(t *testing.T) {
    // Create temporary directory
    tmpDir := t.TempDir() // Automatically cleaned up after test

    // Create temporary file
    tmpFile := filepath.Join(tmpDir, "test.log")

    // Test file operations
    err := WriteToFile(tmpFile, "content")
    if err != nil {
        t.Fatalf("Failed to write file: %v", err)
    }

    // Verify file content
    content, err := os.ReadFile(tmpFile)
    if err != nil {
        t.Fatalf("Failed to read file: %v", err)
    }

    if string(content) != "content" {
        t.Errorf("File content = %q; want %q", string(content), "content")
    }
}
```

### Real Example from Project

From `internal/logger/logger_test.go`:

```go
func TestStructuredLogging(t *testing.T) {
    tmpFile := filepath.Join(t.TempDir(), "test.log")

    config := &Config{
        Level:  InfoLevel,
        Format: JSONFormat,
        Output: tmpFile,
    }

    logger, err := New(config)
    if err != nil {
        t.Fatalf("Failed to create logger: %v", err)
    }

    // Test structured logging
    fields := map[string]interface{}{
        "user_id":    "user123",
        "session_id": "session456",
        "count":      42,
        "active":     true,
    }

    logger.InfoWithFields("structured message", fields)

    // Read and parse log file
    content, err := os.ReadFile(tmpFile)
    if err != nil {
        t.Fatalf("Failed to read log file: %v", err)
    }

    var logEntry map[string]interface{}
    if err := json.Unmarshal(content, &logEntry); err != nil {
        t.Fatalf("Failed to parse JSON log: %v", err)
    }

    // Verify all fields
    if logEntry["message"] != "structured message" {
        t.Errorf("Expected message 'structured message', got '%v'", logEntry["message"])
    }

    if logEntry["user_id"] != "user123" {
        t.Errorf("Expected user_id 'user123', got '%v'", logEntry["user_id"])
    }
}
```

## Writing Integration Tests

### Integration Test Structure

```go
//go:build integration
// +build integration

package tests

import (
    "testing"
    "database/sql"
)

func TestDatabaseIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    // Set up test database
    db, cleanup := setupTestDB(t)
    defer cleanup()

    // Run integration test
    queries := database.New(db)
    user, err := queries.CreateUser(context.Background(), "test@example.com")
    if err != nil {
        t.Fatalf("Failed to create user: %v", err)
    }

    // Verify
    if user.Email != "test@example.com" {
        t.Errorf("User email = %q; want %q", user.Email, "test@example.com")
    }
}

func setupTestDB(t *testing.T) (*sql.DB, func()) {
    // Create test database
    db, err := sql.Open("sqlite3", ":memory:")
    if err != nil {
        t.Fatalf("Failed to open test database: %v", err)
    }

    // Run migrations
    if err := runMigrations(db); err != nil {
        t.Fatalf("Failed to run migrations: %v", err)
    }

    // Return cleanup function
    cleanup := func() {
        db.Close()
    }

    return db, cleanup
}
```

### API Integration Tests

```go
func TestAPIIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping API integration test")
    }

    // Set up test server
    server := httptest.NewServer(handler)
    defer server.Close()

    // Create client
    client := NewClient(server.URL)

    // Test API call
    resp, err := client.GetUser(context.Background(), "user123")
    if err != nil {
        t.Fatalf("API call failed: %v", err)
    }

    // Verify response
    if resp.ID != "user123" {
        t.Errorf("User ID = %q; want %q", resp.ID, "user123")
    }
}
```

## Test Coverage

### Coverage Requirements

AINative Code maintains **80% minimum code coverage**.

### Check Coverage

```bash
# Run tests with coverage
make test-coverage-check

# This will fail if coverage is below 80%
```

### Coverage Report

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View in terminal
go tool cover -func=coverage.out

# View in browser
go tool cover -html=coverage.out

# Get total coverage percentage
go tool cover -func=coverage.out | grep total
```

### Improving Coverage

1. **Identify uncovered code**:
   ```bash
   go tool cover -html=coverage.out
   # Red = uncovered, Green = covered
   ```

2. **Add tests for uncovered code**
3. **Verify improvement**:
   ```bash
   go test -cover ./internal/mypackage
   ```

### Coverage Exclusions

Some code may be excluded from coverage:
- Generated code (marked with `// Code generated`)
- Main function
- Panic/fatal error handlers
- Platform-specific code not testable on current OS

## Benchmarking

### Writing Benchmarks

```go
func BenchmarkMyFunction(b *testing.B) {
    // Setup (not measured)
    input := "test data"

    // Reset timer after setup
    b.ResetTimer()

    // Run function b.N times
    for i := 0; i < b.N; i++ {
        MyFunction(input)
    }
}
```

### Benchmark with Memory Stats

```go
func BenchmarkWithAllocs(b *testing.B) {
    // Report allocations
    b.ReportAllocs()

    for i := 0; i < b.N; i++ {
        MyFunction()
    }
}
```

### Real Benchmark Example

From `internal/logger/logger_bench_test.go`:

```go
func BenchmarkSimpleLog(b *testing.B) {
    logger, _ := New(&Config{
        Level:  InfoLevel,
        Format: JSONFormat,
        Output: "/dev/null",
    })

    b.ResetTimer()
    b.ReportAllocs()

    for i := 0; i < b.N; i++ {
        logger.Info("benchmark message")
    }
}
```

### Running Benchmarks

```bash
# Run all benchmarks
go test -bench=. ./...

# Run specific benchmark
go test -bench=BenchmarkSimpleLog ./internal/logger

# With memory allocations
go test -bench=. -benchmem ./...

# Run for specific duration
go test -bench=. -benchtime=10s ./...

# Run with different input sizes
go test -bench=. -benchmem -count=5 ./...
```

### Benchmark Results Interpretation

```
BenchmarkSimpleLog-8    500000    2000 ns/op    0 B/op    0 allocs/op
```

- `BenchmarkSimpleLog-8`: Benchmark name with GOMAXPROCS
- `500000`: Number of iterations
- `2000 ns/op`: Nanoseconds per operation
- `0 B/op`: Bytes allocated per operation
- `0 allocs/op`: Number of allocations per operation

## Mocking and Testing Patterns

### Interface-Based Testing

```go
// Define interface
type UserStore interface {
    GetUser(ctx context.Context, id string) (*User, error)
    CreateUser(ctx context.Context, email string) (*User, error)
}

// Mock implementation
type MockUserStore struct {
    GetUserFunc    func(ctx context.Context, id string) (*User, error)
    CreateUserFunc func(ctx context.Context, email string) (*User, error)
}

func (m *MockUserStore) GetUser(ctx context.Context, id string) (*User, error) {
    if m.GetUserFunc != nil {
        return m.GetUserFunc(ctx, id)
    }
    return nil, errors.New("not implemented")
}

// Use in test
func TestWithMock(t *testing.T) {
    mock := &MockUserStore{
        GetUserFunc: func(ctx context.Context, id string) (*User, error) {
            return &User{ID: id, Email: "test@example.com"}, nil
        },
    }

    user, err := mock.GetUser(context.Background(), "user123")
    if err != nil {
        t.Fatalf("Unexpected error: %v", err)
    }

    if user.ID != "user123" {
        t.Errorf("User ID = %q; want %q", user.ID, "user123")
    }
}
```

### Test Fixtures

```go
// Create test fixtures
func newTestUser(t *testing.T) *User {
    t.Helper()
    return &User{
        ID:    "test-user",
        Email: "test@example.com",
    }
}

func TestWithFixture(t *testing.T) {
    user := newTestUser(t)
    // Use user in test
}
```

### Table-Driven Tests with Subtests

```go
func TestValidation(t *testing.T) {
    tests := map[string]struct {
        input   string
        wantErr bool
    }{
        "valid email":   {"test@example.com", false},
        "invalid email": {"invalid", true},
        "empty string":  {"", true},
    }

    for name, tc := range tests {
        t.Run(name, func(t *testing.T) {
            err := Validate(tc.input)
            if (err != nil) != tc.wantErr {
                t.Errorf("Validate(%q) error = %v, wantErr %v",
                    tc.input, err, tc.wantErr)
            }
        })
    }
}
```

## Best Practices

### 1. Test Naming

```go
// Good: Descriptive test names
func TestValidateEmail_WithValidEmail_ReturnsNoError(t *testing.T) {}
func TestValidateEmail_WithInvalidEmail_ReturnsError(t *testing.T) {}

// Better: Use subtests for clarity
func TestValidateEmail(t *testing.T) {
    t.Run("valid email returns no error", func(t *testing.T) {})
    t.Run("invalid email returns error", func(t *testing.T) {})
}
```

### 2. Use t.Helper()

```go
func assertNoError(t *testing.T, err error) {
    t.Helper() // Mark as helper function
    if err != nil {
        t.Fatalf("Unexpected error: %v", err)
    }
}

func TestSomething(t *testing.T) {
    err := DoSomething()
    assertNoError(t, err) // Failure points to this line, not inside assertNoError
}
```

### 3. Clean Up Resources

```go
func TestWithCleanup(t *testing.T) {
    // Setup
    resource := setupResource()

    // Register cleanup (runs even if test fails)
    t.Cleanup(func() {
        resource.Close()
    })

    // Or use defer
    defer resource.Close()

    // Test code
}
```

### 4. Use t.TempDir()

```go
func TestFileOperations(t *testing.T) {
    // Automatically cleaned up
    tmpDir := t.TempDir()

    // Use tmpDir for test files
}
```

### 5. Avoid Testing Implementation Details

```go
// Bad: Testing internal implementation
func TestInternalState(t *testing.T) {
    obj := NewObject()
    if obj.internalCounter != 0 {
        t.Error("Internal counter should be 0")
    }
}

// Good: Testing behavior
func TestBehavior(t *testing.T) {
    obj := NewObject()
    result := obj.DoSomething()
    if result != expected {
        t.Errorf("DoSomething() = %v; want %v", result, expected)
    }
}
```

### 6. Test Edge Cases

```go
func TestEdgeCases(t *testing.T) {
    tests := []struct {
        name  string
        input int
    }{
        {"zero", 0},
        {"negative", -1},
        {"max int", math.MaxInt64},
        {"min int", math.MinInt64},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test edge case
        })
    }
}
```

### 7. Parallel Tests

```go
func TestParallel(t *testing.T) {
    t.Parallel() // Run in parallel with other parallel tests

    // Test code
}

func TestWithSubtests(t *testing.T) {
    tests := []struct {
        name string
        // ...
    }{
        // test cases
    }

    for _, tt := range tests {
        tt := tt // Capture range variable
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel() // Run subtests in parallel
            // Test code
        })
    }
}
```

## Quick Reference

### Common Test Commands

```bash
# Basic testing
go test ./...                    # Run all tests
go test -v ./...                 # Verbose output
go test -run TestName ./...      # Run specific test
go test -short ./...             # Skip long-running tests

# Coverage
go test -cover ./...             # Show coverage
go test -coverprofile=c.out ./...  # Generate coverage file
go tool cover -html=c.out        # View coverage in browser

# Benchmarks
go test -bench=. ./...           # Run benchmarks
go test -bench=. -benchmem ./... # With memory stats

# Race detection
go test -race ./...              # Detect race conditions

# Integration
go test -tags=integration ./...  # Run integration tests

# Makefile shortcuts
make test                        # Run all tests
make test-coverage               # Tests with coverage report
make test-integration            # Integration tests only
make test-benchmark              # Run benchmarks
make ci                          # Run all CI checks
```

---

**Next**: [Debugging Guide](debugging.md) | [Code Style Guidelines](code-style.md) | [Build Instructions](build.md)
