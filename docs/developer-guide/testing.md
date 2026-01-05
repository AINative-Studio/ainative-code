# Testing Guide

## Overview

Testing is a critical part of the AINative Code development process. We maintain a **minimum 80% code coverage** requirement and follow best practices for unit, integration, and end-to-end testing.

## Test Organization

### Directory Structure

```
ainative-code/
├── internal/
│   ├── provider/
│   │   ├── provider.go
│   │   ├── provider_test.go      # Unit tests
│   │   ├── anthropic/
│   │   │   ├── provider.go
│   │   │   └── provider_test.go  # Unit tests
│   │   └── mock/
│   │       └── provider.go       # Mock for testing
│   └── tools/
│       ├── interface.go
│       ├── interface_test.go
│       └── validator_test.go
├── tests/
│   ├── integration/              # Integration tests
│   │   ├── auth_test.go
│   │   ├── design_test.go
│   │   └── strapi_test.go
│   ├── security/                 # Security tests
│   │   ├── auth_security_test.go
│   │   └── sql_injection_test.go
│   └── benchmark/                # Benchmark tests
│       ├── streaming_bench_test.go
│       └── token_bench_test.go
```

### Test Types

1. **Unit Tests**: Test individual functions and methods in isolation
2. **Integration Tests**: Test component interactions
3. **Security Tests**: Validate security controls
4. **Benchmark Tests**: Measure performance
5. **End-to-End Tests**: Test complete user workflows

## Running Tests

### All Tests

```bash
# Run all tests
make test

# Or with go directly
go test -v -race ./...

# Run with coverage
make test-coverage

# Run with detailed coverage
go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
go tool cover -html=coverage.out
```

### Specific Test Types

```bash
# Unit tests only (default)
go test ./internal/...

# Integration tests
make test-integration
# Or: go test -v -timeout=10m -tags=integration ./tests/integration/...

# Security tests
go test -v ./tests/security/...

# Benchmark tests
make test-benchmark
# Or: go test -bench=. -benchmem ./...
```

### Running Specific Tests

```bash
# Run specific package
go test -v ./internal/provider/...

# Run specific test function
go test -v -run TestProviderChat ./internal/provider

# Run tests matching pattern
go test -v -run "TestProvider.*" ./...

# Run with race detector
go test -race -run TestConcurrent ./...
```

### Test Flags

```bash
# Verbose output
go test -v ./...

# Race detection
go test -race ./...

# Coverage
go test -cover ./...
go test -coverprofile=coverage.out ./...

# Timeout
go test -timeout 30s ./...

# Short mode (skip long tests)
go test -short ./...

# Parallel execution
go test -parallel 4 ./...

# Run tests sequentially
go test -p 1 ./...
```

## Writing Unit Tests

### Basic Test Structure

```go
package provider

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestProviderChat(t *testing.T) {
    // Arrange
    provider := NewMockProvider()
    ctx := context.Background()
    messages := []Message{
        {Role: "user", Content: "Hello"},
    }

    // Act
    response, err := provider.Chat(ctx, messages)

    // Assert
    require.NoError(t, err)
    assert.NotEmpty(t, response.Content)
    assert.Greater(t, response.Usage.TotalTokens, 0)
}
```

### Table-Driven Tests

```go
func TestMessageValidation(t *testing.T) {
    tests := []struct {
        name    string
        message Message
        want    bool
        wantErr bool
    }{
        {
            name:    "valid user message",
            message: Message{Role: "user", Content: "Hello"},
            want:    true,
            wantErr: false,
        },
        {
            name:    "empty content",
            message: Message{Role: "user", Content: ""},
            want:    false,
            wantErr: true,
        },
        {
            name:    "invalid role",
            message: Message{Role: "invalid", Content: "Hello"},
            want:    false,
            wantErr: true,
        },
        {
            name:    "system message",
            message: Message{Role: "system", Content: "You are helpful"},
            want:    true,
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ValidateMessage(tt.message)

            if tt.wantErr {
                assert.Error(t, err)
                return
            }

            require.NoError(t, err)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

### Testing with Subtests

```go
func TestConfigLoader(t *testing.T) {
    t.Run("LoadFromFile", func(t *testing.T) {
        config, err := LoadConfig("testdata/config.yaml")
        require.NoError(t, err)
        assert.NotNil(t, config)
    })

    t.Run("LoadFromEnv", func(t *testing.T) {
        t.Setenv("AINATIVE_PROVIDER", "anthropic")
        config, err := LoadConfig("")
        require.NoError(t, err)
        assert.Equal(t, "anthropic", config.Provider)
    })

    t.Run("InvalidPath", func(t *testing.T) {
        _, err := LoadConfig("nonexistent.yaml")
        assert.Error(t, err)
    })
}
```

### Testing with Context

```go
func TestProviderChatWithTimeout(t *testing.T) {
    provider := NewAnthropicProvider(testAPIKey)

    // Test with timeout context
    ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
    defer cancel()

    messages := []Message{{Role: "user", Content: "Hello"}}

    // This should timeout
    _, err := provider.Chat(ctx, messages)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "context deadline exceeded")
}

func TestProviderChatWithCancellation(t *testing.T) {
    provider := NewAnthropicProvider(testAPIKey)
    ctx, cancel := context.WithCancel(context.Background())

    // Cancel immediately
    cancel()

    messages := []Message{{Role: "user", Content: "Hello"}}
    _, err := provider.Chat(ctx, messages)

    assert.Error(t, err)
    assert.Contains(t, err.Error(), "context canceled")
}
```

## Using Mocks

### Creating Mocks

```go
// internal/provider/mock/provider.go
package mock

import (
    "context"
    "github.com/AINative-studio/ainative-code/internal/provider"
)

type Provider struct {
    ChatFunc   func(ctx context.Context, messages []provider.Message, opts ...provider.ChatOption) (provider.Response, error)
    StreamFunc func(ctx context.Context, messages []provider.Message, opts ...provider.StreamOption) (<-chan provider.Event, error)
}

func (m *Provider) Chat(ctx context.Context, messages []provider.Message, opts ...provider.ChatOption) (provider.Response, error) {
    if m.ChatFunc != nil {
        return m.ChatFunc(ctx, messages, opts...)
    }
    return provider.Response{}, nil
}

func (m *Provider) Stream(ctx context.Context, messages []provider.Message, opts ...provider.StreamOption) (<-chan provider.Event, error) {
    if m.StreamFunc != nil {
        return m.StreamFunc(ctx, messages, opts...)
    }

    events := make(chan provider.Event)
    close(events)
    return events, nil
}

func (m *Provider) Name() string { return "mock" }
func (m *Provider) Models() []string { return []string{"mock-model"} }
func (m *Provider) Close() error { return nil }
```

### Using Mocks in Tests

```go
func TestChatHandlerWithMock(t *testing.T) {
    // Create mock provider
    mockProvider := &mock.Provider{
        ChatFunc: func(ctx context.Context, messages []provider.Message, opts ...provider.ChatOption) (provider.Response, error) {
            return provider.Response{
                Content: "Mocked response",
                Usage: provider.Usage{
                    PromptTokens:     10,
                    CompletionTokens: 20,
                    TotalTokens:      30,
                },
            }, nil
        },
    }

    // Use mock in test
    handler := NewChatHandler(mockProvider)
    response, err := handler.HandleChat(context.Background(), "Hello")

    require.NoError(t, err)
    assert.Equal(t, "Mocked response", response)
}
```

### Testify Mock Framework

```go
import (
    "github.com/stretchr/testify/mock"
)

type MockProvider struct {
    mock.Mock
}

func (m *MockProvider) Chat(ctx context.Context, messages []provider.Message, opts ...provider.ChatOption) (provider.Response, error) {
    args := m.Called(ctx, messages, opts)
    return args.Get(0).(provider.Response), args.Error(1)
}

func TestWithTestifyMock(t *testing.T) {
    mockProvider := new(MockProvider)

    // Set expectations
    mockProvider.On("Chat", mock.Anything, mock.Anything, mock.Anything).
        Return(provider.Response{Content: "Hello"}, nil)

    // Use mock
    response, err := mockProvider.Chat(context.Background(), nil)

    require.NoError(t, err)
    assert.Equal(t, "Hello", response.Content)

    // Verify expectations were met
    mockProvider.AssertExpectations(t)
}
```

## Integration Tests

### Build Tags

Integration tests use build tags to separate them from unit tests:

```go
//go:build integration
// +build integration

package integration

import "testing"

func TestZeroDBIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    // Integration test code
}
```

### Running Integration Tests

```bash
# Run integration tests
go test -v -tags=integration ./tests/integration/...

# Or with Make
make test-integration

# Skip in short mode
go test -short -tags=integration ./tests/integration/...
```

### Example Integration Test

```go
//go:build integration

package integration

import (
    "context"
    "testing"
    "time"

    "github.com/AINative-studio/ainative-code/internal/client/zerodb"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestZeroDBQuery(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    // Setup
    client := zerodb.NewClient(zerodb.Config{
        Endpoint: getTestEndpoint(),
        APIKey:   getTestAPIKey(),
    })
    defer client.Close()

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // Execute
    result, err := client.Query(ctx, "SELECT 1 as value")

    // Assert
    require.NoError(t, err)
    assert.NotNil(t, result)
}

func getTestEndpoint() string {
    // Read from environment or use default
    endpoint := os.Getenv("ZERODB_TEST_ENDPOINT")
    if endpoint == "" {
        endpoint = "http://localhost:8080"
    }
    return endpoint
}
```

## Test Fixtures

### Using testdata Directory

```
internal/config/
├── loader.go
├── loader_test.go
└── testdata/
    ├── config.yaml
    ├── invalid-config.yaml
    └── empty-config.yaml
```

```go
func TestLoadConfig(t *testing.T) {
    tests := []struct {
        name    string
        file    string
        wantErr bool
    }{
        {"valid config", "testdata/config.yaml", false},
        {"invalid config", "testdata/invalid-config.yaml", true},
        {"empty config", "testdata/empty-config.yaml", false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := LoadConfig(tt.file)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### Temporary Files and Directories

```go
func TestDatabaseOperations(t *testing.T) {
    // Create temporary directory
    tmpDir := t.TempDir() // Automatically cleaned up after test

    dbPath := filepath.Join(tmpDir, "test.db")
    db, err := OpenDatabase(dbPath)
    require.NoError(t, err)
    defer db.Close()

    // Test database operations
    err = db.CreateTable("users")
    assert.NoError(t, err)
}

func TestConfigFile(t *testing.T) {
    // Create temporary file
    tmpFile, err := os.CreateTemp("", "config-*.yaml")
    require.NoError(t, err)
    defer os.Remove(tmpFile.Name())

    // Write test data
    _, err = tmpFile.WriteString("provider: anthropic\n")
    require.NoError(t, err)
    tmpFile.Close()

    // Test with temp file
    config, err := LoadConfig(tmpFile.Name())
    require.NoError(t, err)
    assert.Equal(t, "anthropic", config.Provider)
}
```

## Test Coverage

### Coverage Requirements

- **Minimum Coverage**: 80%
- **Critical Packages**: 90%+ (auth, provider, config)
- **UI Packages**: 60%+ (TUI testing is complex)

### Generating Coverage Reports

```bash
# Generate coverage profile
go test -coverprofile=coverage.out ./...

# View coverage in terminal
go tool cover -func=coverage.out

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html

# Open in browser
open coverage.html
```

### Coverage by Package

```bash
# Show coverage for each package
go test -cover ./...

# Detailed coverage
go test -coverprofile=coverage.out -covermode=atomic ./...
go tool cover -func=coverage.out | grep -E "^total|internal/"
```

### Coverage Enforcement

```bash
# Check coverage meets threshold
make test-coverage-check

# Or manually
go test -coverprofile=coverage.out ./...
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
if (( $(echo "$COVERAGE < 80.0" | bc -l) )); then
    echo "Coverage $COVERAGE% is below threshold 80%"
    exit 1
fi
```

### Excluding from Coverage

Use build tags to exclude files from coverage:

```go
//go:build !test

package main

// This file is excluded from coverage
```

## Benchmarking

### Writing Benchmarks

```go
func BenchmarkProviderChat(b *testing.B) {
    provider := NewMockProvider()
    ctx := context.Background()
    messages := []Message{{Role: "user", Content: "Hello"}}

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = provider.Chat(ctx, messages)
    }
}

func BenchmarkConfigLoad(b *testing.B) {
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = LoadConfig("testdata/config.yaml")
    }
}
```

### Benchmarking with Memory

```go
func BenchmarkLoggerMemory(b *testing.B) {
    logger := NewLogger(Config{Level: INFO})

    b.ReportAllocs()
    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        logger.Info("Benchmark message")
    }
}
```

### Running Benchmarks

```bash
# Run all benchmarks
go test -bench=. ./...

# Run specific benchmark
go test -bench=BenchmarkProviderChat ./internal/provider

# With memory statistics
go test -bench=. -benchmem ./...

# Save results for comparison
go test -bench=. -benchmem ./... > bench-new.txt

# Compare with baseline
go install golang.org/x/perf/cmd/benchstat@latest
benchstat bench-old.txt bench-new.txt
```

### Benchmark Output

```
BenchmarkProviderChat-8         1000000    1234 ns/op    256 B/op    4 allocs/op
BenchmarkLoggerMemory-8         5000000     345 ns/op      0 B/op    0 allocs/op
```

- `BenchmarkProviderChat-8`: Benchmark name with GOMAXPROCS
- `1000000`: Number of iterations
- `1234 ns/op`: Nanoseconds per operation
- `256 B/op`: Bytes allocated per operation
- `4 allocs/op`: Allocations per operation

## Testing Best Practices

### 1. Use Table-Driven Tests

**Good**:
```go
func TestValidation(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    bool
        wantErr bool
    }{
        {"valid", "test@example.com", true, false},
        {"invalid", "not-an-email", false, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Validate(tt.input)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                require.NoError(t, err)
                assert.Equal(t, tt.want, got)
            }
        })
    }
}
```

### 2. Use testify for Assertions

```go
// require: Fails immediately
require.NoError(t, err)
require.NotNil(t, result)

// assert: Continues test
assert.Equal(t, expected, actual)
assert.Greater(t, actual, 0)
assert.Contains(t, list, item)
```

### 3. Test Edge Cases

```go
func TestDivide(t *testing.T) {
    tests := []struct {
        name    string
        a, b    int
        want    int
        wantErr bool
    }{
        {"normal", 10, 2, 5, false},
        {"divide by zero", 10, 0, 0, true},
        {"negative", -10, 2, -5, false},
        {"zero dividend", 0, 5, 0, false},
    }
    // ...
}
```

### 4. Use t.Helper() for Test Helpers

```go
func assertValid(t *testing.T, input string) {
    t.Helper() // Marks this as a helper function

    if !IsValid(input) {
        t.Errorf("expected %q to be valid", input)
    }
}

func TestValidation(t *testing.T) {
    assertValid(t, "test@example.com")
}
```

### 5. Clean Up Resources

```go
func TestDatabase(t *testing.T) {
    db, err := OpenDB()
    require.NoError(t, err)
    defer db.Close() // Always close resources

    t.Cleanup(func() {
        // Additional cleanup
        os.Remove(db.Path())
    })

    // Test code...
}
```

### 6. Test Parallel When Possible

```go
func TestConcurrent(t *testing.T) {
    t.Parallel() // Run in parallel with other tests

    // Test code...
}
```

### 7. Use Descriptive Test Names

```go
// Good
func TestProvider_Chat_ReturnsError_WhenAPIKeyInvalid(t *testing.T)
func TestConfig_Load_ParsesYAML_Successfully(t *testing.T)

// Bad
func TestProvider(t *testing.T)
func TestConfig(t *testing.T)
```

## Continuous Integration

### GitHub Actions

Tests run automatically on:
- Push to `main` or `develop`
- Pull requests
- Manual workflow dispatch

### CI Test Matrix

```yaml
strategy:
  matrix:
    os: [ubuntu-latest, macos-latest, windows-latest]
    go-version: [1.21, 1.22]
```

### Coverage Reporting

Coverage reports are automatically uploaded to Codecov:

```bash
# In CI
go test -race -coverprofile=coverage.out -covermode=atomic ./...
bash <(curl -s https://codecov.io/bash)
```

## Troubleshooting Tests

### Tests Fail Randomly

Likely race condition. Run with race detector:

```bash
go test -race ./...
```

### Tests Timeout

Increase timeout:

```bash
go test -timeout 5m ./...
```

### Database Locked Errors

Run tests sequentially:

```bash
go test -p 1 ./...
```

### Coverage Not Generated

Ensure all packages are tested:

```bash
go test -coverprofile=coverage.out ./...
```

## Resources

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Testify Documentation](https://github.com/stretchr/testify)
- [Table Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests)
- [Advanced Testing Patterns](https://go.dev/blog/subtests)

---

**Last Updated**: 2025-01-05
