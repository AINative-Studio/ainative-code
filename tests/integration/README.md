# Integration Tests

This directory contains integration tests for AINative Code. These tests verify that different components work together correctly in realistic scenarios.

## Overview

Integration tests validate:
- **Chat Sessions**: Provider initialization, chat requests, streaming responses, and error handling
- **Tool Execution**: Bash commands, file operations, output validation, timeout/cancellation
- **Session Persistence**: Database operations, message storage, export/import, search functionality

## Directory Structure

```
tests/
├── integration/
│   ├── README.md              # This file
│   ├── chat_test.go          # Chat session integration tests
│   ├── tools_test.go         # Tool execution integration tests
│   ├── session_test.go       # Session persistence integration tests
│   └── helpers.go            # Shared test helpers
├── fixtures/
│   ├── config.yaml           # Test configuration
│   ├── messages.json         # Sample messages
│   └── sessions.json         # Sample sessions
└── helpers/
    ├── mock_provider.go      # Mock LLM provider implementation
    ├── mock_server.go        # Mock HTTP server for API testing
    └── test_db.go            # Test database utilities
```

## Running Tests

### Run All Integration Tests

```bash
go test -v ./tests/integration/...
```

### Run Specific Test File

```bash
# Chat tests only
go test -v ./tests/integration -run TestChatSession

# Tool tests only
go test -v ./tests/integration -run TestToolExecution

# Session persistence tests only
go test -v ./tests/integration -run TestSessionPersistence
```

### Run Individual Test

```bash
# Run a specific test
go test -v ./tests/integration -run TestChatSession_Initialize

# Run tests matching a pattern
go test -v ./tests/integration -run "Chat.*Error"
```

### Run with Coverage

```bash
# Generate coverage report
go test -v -cover ./tests/integration/...

# Generate detailed coverage report
go test -v -coverprofile=coverage.out ./tests/integration/...
go tool cover -html=coverage.out -o coverage.html
```

### Run with Race Detector

```bash
go test -v -race ./tests/integration/...
```

## Test Scenarios

### 1. Chat Session Tests (`chat_test.go`)

#### Provider Initialization
- ✓ Valid configuration with all parameters
- ✓ Configuration with defaults
- ✓ Error handling for invalid configuration

#### Chat Requests
- ✓ Single message chat request
- ✓ Multi-turn conversation
- ✓ Empty message handling
- ✓ Request parameter validation

#### Streaming Responses
- ✓ Default streaming events
- ✓ Custom event sequences
- ✓ Event type validation
- ✓ Content accumulation

#### Error Handling
- ✓ Authentication errors (invalid API key)
- ✓ Rate limit errors
- ✓ Timeout errors
- ✓ Context cancellation
- ✓ Network errors

#### Concurrency
- ✓ Multiple concurrent chat requests
- ✓ Thread safety verification
- ✓ Request isolation

**Example:**
```go
func TestChatSession_SendRequest(t *testing.T) {
    provider := helpers.NewMockProvider("mock")
    defer provider.Close()

    req := &providers.ChatRequest{
        Messages: []providers.Message{
            {Role: providers.RoleUser, Content: "Hello"},
        },
        Model: "test-model-v1",
    }

    resp, err := provider.Chat(context.Background(), req)
    require.NoError(t, err)
    assert.NotNil(t, resp)
}
```

### 2. Tool Execution Tests (`tools_test.go`)

#### Bash Command Execution
- ✓ Simple commands (echo, pwd, date)
- ✓ Commands with arguments
- ✓ Commands with environment variables
- ✓ Custom working directory
- ✓ Command timeout handling
- ✓ Context cancellation

#### File Operations
- ✓ Write file to disk
- ✓ Read file from disk
- ✓ Create nested directories
- ✓ Read non-existent file (error case)
- ✓ Security: prevent writing outside allowed directory

#### Output Validation
- ✓ Capture stdout separately
- ✓ Capture stderr separately
- ✓ Capture combined output
- ✓ Handle non-zero exit codes
- ✓ Output formatting

#### Security & Permissions
- ✓ Command whitelist enforcement
- ✓ Reject unauthorized commands
- ✓ File path validation
- ✓ Directory traversal prevention

#### Metadata Collection
- ✓ Exit code tracking
- ✓ Execution duration
- ✓ Output byte counts
- ✓ Environment variables
- ✓ Working directory

**Example:**
```go
func TestToolExecution_BashCommand(t *testing.T) {
    tool := builtin.NewExecCommandTool([]string{"echo"}, "")

    input := map[string]interface{}{
        "command": "echo",
        "args":    []interface{}{"Hello", "World"},
    }

    result, err := tool.Execute(context.Background(), input)
    require.NoError(t, err)
    assert.True(t, result.Success)
    assert.Contains(t, result.Output, "Hello World")
}
```

### 3. Session Persistence Tests (`session_test.go`)

#### Session Management
- ✓ Create new session
- ✓ Create session with settings (model, temperature, etc.)
- ✓ Get session by ID
- ✓ Update session
- ✓ Archive session (soft delete)
- ✓ Delete session (soft delete)
- ✓ Hard delete session
- ✓ List sessions with filters
- ✓ Prevent duplicate session IDs

#### Message Operations
- ✓ Add user message
- ✓ Add assistant message with metadata
- ✓ Add message with parent reference (threading)
- ✓ Get all messages for session
- ✓ Get paginated messages
- ✓ Get message count
- ✓ Update message
- ✓ Delete message

#### Session Resume
- ✓ Save conversation to database
- ✓ Resume session from database
- ✓ Load conversation history
- ✓ Maintain message order
- ✓ Preserve message threading

#### Export/Import
- ✓ Export session as JSON
- ✓ Export session as Markdown
- ✓ Export session as plain text
- ✓ Import session from JSON
- ✓ Preserve all metadata during export/import

#### Search
- ✓ Search messages by content
- ✓ Search sessions by name
- ✓ Full-text search
- ✓ Filter search results

#### Statistics
- ✓ Get message count per session
- ✓ Calculate total tokens used
- ✓ Get session summary with stats

**Example:**
```go
func TestSessionPersistence_CreateSession(t *testing.T) {
    db := helpers.SetupTestDB(t)
    mgr := session.NewSQLiteManager(db)

    sess := &session.Session{
        ID:        uuid.New().String(),
        Name:      "Test Session",
        Status:    session.StatusActive,
    }

    err := mgr.CreateSession(context.Background(), sess)
    require.NoError(t, err)

    retrieved, err := mgr.GetSession(context.Background(), sess.ID)
    require.NoError(t, err)
    assert.Equal(t, sess.Name, retrieved.Name)
}
```

## Test Infrastructure

### Mock Provider

The `MockProvider` simulates an LLM provider for testing without making real API calls:

```go
provider := helpers.NewMockProvider("mock")
defer provider.Close()

// Configure custom response
provider.SetChatResponse(&providers.Response{
    Content: "Custom response",
    Model:   "test-model",
}, nil)

// Configure streaming events
provider.SetStreamEvents([]providers.Event{
    {Type: providers.EventTextDelta, Data: "Hello"},
}, nil)
```

### Mock Server

The `MockServer` creates a test HTTP server for API endpoint testing:

```go
server := helpers.NewMockServer(t)
// Server automatically closes via t.Cleanup()

// Configure endpoint
server.SetResponse("POST", "/api/chat", http.StatusOK, map[string]interface{}{
    "message": "response",
})

// Verify endpoint was called
helpers.AssertEndpointCalled(t, server, "POST", "/api/chat")
```

### Test Database

The `SetupTestDB` creates an in-memory SQLite database:

```go
db := helpers.SetupTestDB(t)
// Database automatically closes and cleans up via t.Cleanup()

// Verify table is empty
helpers.AssertTableEmpty(t, db, "sessions")

// Check row count
helpers.AssertTableRowCount(t, db, "messages", 5)

// Clean specific tables
helpers.CleanupDB(t, db)
```

## Best Practices

### Test Isolation

Each test should be completely independent:

```go
func TestExample(t *testing.T) {
    // Create fresh database for this test
    db := helpers.SetupTestDB(t)

    // Test runs in isolation
    // ...

    // Cleanup happens automatically via t.Cleanup()
}
```

### Use Subtests

Group related tests using `t.Run`:

```go
func TestChatSession_Scenarios(t *testing.T) {
    t.Run("successful request", func(t *testing.T) {
        // Test code
    })

    t.Run("authentication error", func(t *testing.T) {
        // Test code
    })
}
```

### Table-Driven Tests

Use table-driven tests for multiple similar scenarios:

```go
tests := []struct {
    name    string
    input   string
    wantErr bool
}{
    {"valid input", "test", false},
    {"empty input", "", true},
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        // Test with tt.input
    })
}
```

### Context Management

Always use context with timeout for operations that could hang:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

result, err := operation(ctx)
```

### Cleanup

Use `t.Cleanup()` for resource cleanup:

```go
func TestExample(t *testing.T) {
    resource := setupResource()

    t.Cleanup(func() {
        resource.Close()
    })

    // Test code
}
```

## Performance Expectations

Integration tests should complete within reasonable time:

- **Individual test**: < 1 second
- **Test file**: < 10 seconds
- **Full suite**: < 2 minutes

Tests exceeding these limits should be investigated for:
- Unnecessary waits or sleeps
- Inefficient database operations
- Missing timeouts
- Resource leaks

## Debugging Tests

### Verbose Output

```bash
# Run with verbose output
go test -v ./tests/integration -run TestName

# Show test output even for passing tests
go test -v -test.v ./tests/integration
```

### Debug Individual Test

```bash
# Run single test with verbose output
go test -v ./tests/integration -run TestChatSession_Initialize/valid_configuration

# Add additional logging
AINATIVE_LOG_LEVEL=debug go test -v ./tests/integration -run TestName
```

### Using Delve Debugger

```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug specific test
dlv test ./tests/integration -- -test.run TestChatSession_Initialize
```

## Continuous Integration

These tests run automatically on:
- Every pull request
- Every commit to main branch
- Nightly builds

CI pipeline requirements:
- All tests must pass
- Coverage >= 80%
- No race conditions
- Execution time < 2 minutes

## Troubleshooting

### Tests Hang

- Check for missing context timeouts
- Verify channels are properly closed
- Look for goroutine leaks

### Flaky Tests

- Ensure proper test isolation
- Remove timing dependencies
- Use synchronization primitives instead of sleeps

### Database Errors

- Check migration scripts are up to date
- Verify foreign key constraints
- Ensure proper cleanup between tests

### Mock Provider Issues

- Verify mock is configured before use
- Check that responses match expected types
- Ensure streaming events are properly closed

## Contributing

When adding new integration tests:

1. Follow existing test patterns and naming conventions
2. Add test documentation to this README
3. Ensure tests are isolated and deterministic
4. Add fixtures for complex test data
5. Update the test count in acceptance criteria
6. Run full test suite before submitting PR

## Test Fixtures

Test fixtures in `tests/fixtures/` provide sample data:

- `config.yaml`: Test configuration with mock provider settings
- `messages.json`: Sample conversation messages
- `sessions.json`: Sample session data with various states

Load fixtures in tests:

```go
// Load config
config := helpers.LoadTestConfig(t)

// Load sample messages
messages := helpers.LoadTestMessages(t)

// Load sample sessions
sessions := helpers.LoadTestSessions(t)
```

## References

- [Go Testing Package](https://pkg.go.dev/testing)
- [Testify Documentation](https://github.com/stretchr/testify)
- [Table-Driven Tests in Go](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)
- [Advanced Go Testing Tutorial](https://quii.gitbook.io/learn-go-with-tests/)
