# Integration Tests

This document describes the integration testing strategy for AINative-Code.

## Overview

Integration tests verify critical user workflows across multiple components of the system. Unlike unit tests that test individual functions in isolation, integration tests validate that different parts of the application work together correctly.

## Test Scenarios

### 1. Session Persistence and Resume

**Location**: `tests/integration/session_test.go`

**Purpose**: Verify that chat sessions are correctly persisted to the database and can be resumed later.

**Test Cases**:
- Session creation and retrieval
- Session persistence across operations
- Session update and modification
- Session listing with filters
- Soft and hard deletion
- Message CRUD operations
- Session export/import
- Token usage tracking
- Search functionality
- Conversation thread retrieval
- Session archiving and restoration
- Concurrent operations

**Key Features Tested**:
- SQLite database integration
- Transaction management
- Data integrity
- Concurrent access
- Session state management

### 2. ZeroDB Operations

**Location**: `tests/integration/zerodb_test.go`

**Purpose**: Verify ZeroDB NoSQL, vector, and quantum operations through the API client.

**Test Cases**:
- Table creation and listing
- Document insert, query, update, delete
- Memory store and retrieval
- Memory listing and clearing
- Vector collection operations
- Vector insert and similarity search
- Vector deletion
- Quantum entanglement
- Quantum state measurement
- Quantum compression/decompression
- Quantum-enhanced search
- Error handling and validation

**Key Features Tested**:
- HTTP API client integration
- Mock server responses
- Request/response parsing
- Error handling
- API contract validation

### 3. Design Token Extraction

**Location**: `tests/integration/design_test.go`

**Purpose**: Verify design token extraction and multi-format code generation.

**Test Cases**:
- Design token validation
- CSS generation
- SCSS generation
- JSON generation
- Tailwind configuration generation
- TypeScript type generation
- Token formatting
- Multi-format generation
- Error handling for invalid tokens

**Key Features Tested**:
- Token parsing and validation
- Code generation pipelines
- Format-specific output
- Template rendering
- Validation rules

### 4. OAuth Login Flow

**Status**: To be implemented

**Purpose**: Verify OAuth 2.0 PKCE authentication flow.

**Planned Test Cases**:
- Authorization request generation
- Code exchange for tokens
- Token storage
- Token refresh
- Token validation
- Error handling

### 5. Chat Session with LLM Provider

**Status**: To be implemented

**Purpose**: Verify chat interaction with Anthropic Claude API.

**Planned Test Cases**:
- Message sending and streaming
- Tool use integration
- Context management
- Error handling
- Rate limiting

### 6. Tool Execution

**Status**: To be implemented

**Purpose**: Verify bash command and file operation tools.

**Planned Test Cases**:
- Bash command execution
- File read/write operations
- Security validation
- Error handling

### 7. Strapi Content Management

**Status**: To be implemented

**Purpose**: Verify Strapi CMS integration for blog management.

**Planned Test Cases**:
- Blog post creation
- Content listing
- Update operations
- Deletion
- Error handling

### 8. RLHF Feedback Submission

**Status**: To be implemented

**Purpose**: Verify feedback collection and submission to ZeroDB.

**Planned Test Cases**:
- Feedback capture
- Rating submission
- Metadata collection
- Error handling

## Test Infrastructure

### Directory Structure

```
tests/integration/
├── docker-compose.yml      # Docker services for testing
├── fixtures/               # Test data fixtures
│   ├── sessions.go
│   └── design.go
├── helpers/                # Test helper utilities
│   ├── database.go
│   └── mock_server.go
├── session_test.go         # Session integration tests
├── zerodb_test.go          # ZeroDB integration tests
├── design_test.go          # Design token tests
└── suite_test.go           # Test suite configuration
```

### Test Helpers

#### Database Helpers

```go
// SetupTestDB creates a temporary SQLite database
db, cleanup := helpers.SetupTestDB(t)
defer cleanup()

// SetupInMemoryDB creates an in-memory database (faster)
db, cleanup := helpers.SetupInMemoryDB(t)
defer cleanup()
```

#### Mock Server Helpers

```go
// MockZeroDBServer creates a mock ZeroDB API server
server, cleanup := helpers.MockZeroDBServer(t)
defer cleanup()

// MockAuthServer creates a mock OAuth server
server, cleanup := helpers.MockAuthServer(t)
defer cleanup()
```

### Test Fixtures

```go
// NewTestSession creates a test session with defaults
session := fixtures.NewTestSession()

// NewTestMessage creates a test message
msg := fixtures.NewTestMessage(sessionID, role, content)

// TestDesignTokens returns sample design tokens
tokens := fixtures.TestDesignTokens()
```

## Running Integration Tests

### Run All Integration Tests

```bash
make test-integration
```

### Run Integration Tests with Coverage

```bash
make test-integration-coverage
```

### Run Specific Test Suite

```bash
go test -v -tags=integration ./tests/integration -run TestSessionIntegrationTestSuite
```

### Run Specific Test Case

```bash
go test -v -tags=integration ./tests/integration -run TestSessionIntegrationTestSuite/TestSessionCreationAndRetrieval
```

## Docker Test Environment

Integration tests can optionally use Docker for external service mocks:

```bash
# Start mock services
cd tests/integration
docker-compose up -d

# Run tests
make test-integration

# Stop mock services
docker-compose down
```

**Note**: Currently, tests use in-process mock servers and don't require Docker for execution. Docker support is provided for future expansion.

## Test Guidelines

### BDD-Style Test Naming

Tests follow the Given-When-Then pattern:

```go
func (s *TestSuite) TestFeatureName() {
    // Given: Initial state or preconditions
    ctx := context.Background()
    testData := fixtures.NewTestData()

    // When: Action being tested
    result, err := s.systemUnderTest.DoSomething(ctx, testData)

    // Then: Expected outcomes
    s.Require().NoError(err)
    s.Equal(expectedValue, result)
}
```

### Test Isolation

- Each test should be independent
- Use `SetupTest()` and `TearDownTest()` for per-test setup/cleanup
- Don't rely on test execution order
- Clean up resources in teardown

### Assertions

- Use `Require()` for critical assertions that should stop the test
- Use `Assert()` for assertions that allow the test to continue
- Provide descriptive error messages

### Performance

- Individual tests should complete in < 30 seconds
- Total suite runtime should be < 5 minutes
- Use in-memory databases where possible
- Run independent tests in parallel when safe

## Coverage Requirements

- Integration test coverage: >= 80% for tested code paths
- Critical workflows must have comprehensive coverage
- Error paths must be tested
- Edge cases should be covered

## Continuous Integration

Integration tests are run as part of the CI pipeline:

```yaml
# .github/workflows/test.yml
- name: Run Integration Tests
  run: make test-integration-coverage
```

## Troubleshooting

### Tests Timing Out

- Check for deadlocks in concurrent operations
- Ensure cleanup functions are called
- Increase timeout if needed: `go test -timeout=15m`

### Database Errors

- Ensure migrations are applied in test setup
- Check that database is properly cleaned between tests
- Verify file permissions for file-based SQLite

### Mock Server Issues

- Verify mock server is started before tests
- Check handler registration
- Review request/response logging

### Coverage Gaps

- Run with coverage: `make test-integration-coverage`
- Review coverage report: `go tool cover -html=integration-coverage.out`
- Add tests for uncovered critical paths

## Best Practices

1. **Test Real Workflows**: Integration tests should simulate actual user workflows
2. **Use Real Database**: Use actual SQLite database (in-memory or temporary file)
3. **Mock External APIs**: Use mock HTTP servers for external dependencies
4. **Clean State**: Each test should start with a clean state
5. **Meaningful Assertions**: Verify actual behavior, not just absence of errors
6. **Error Cases**: Test error conditions and edge cases
7. **Documentation**: Add comments explaining complex test scenarios
8. **Fixtures**: Use shared fixtures for common test data
9. **Helpers**: Extract common setup into helper functions
10. **Parallel Execution**: Enable parallel execution when tests are independent

## Future Enhancements

- [ ] Add OAuth flow integration tests
- [ ] Add LLM provider integration tests
- [ ] Add tool execution tests
- [ ] Add Strapi CMS tests
- [ ] Add RLHF feedback tests
- [ ] Add performance benchmarks
- [ ] Add stress tests for concurrent operations
- [ ] Add integration with external test databases
- [ ] Add API contract testing
- [ ] Add cross-platform testing

## References

- [Go Testing Package](https://pkg.go.dev/testing)
- [Testify Suite](https://pkg.go.dev/github.com/stretchr/testify/suite)
- [Integration Testing Best Practices](https://martinfowler.com/bliki/IntegrationTest.html)
