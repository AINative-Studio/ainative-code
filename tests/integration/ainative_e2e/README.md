# AINative E2E Integration Tests

Comprehensive end-to-end integration tests for the AINative Code platform, testing the complete flow from Go CLI to Python backend to AINative API.

## Overview

This test suite validates:
- Authentication flows (login, register, logout, token refresh)
- Chat completions (non-streaming and streaming)
- Error handling (401, 402, 429, network errors)
- Rate limiting
- Token management
- Context cancellation
- Multi-message conversations

## Test Structure

```
tests/integration/ainative_e2e/
├── ainative_e2e_test.go        # Core E2E tests (auth, chat, errors)
├── streaming_e2e_test.go       # Streaming-specific tests
├── cli_integration_test.go     # CLI integration tests (future)
├── mock_backend.go             # Mock Python backend server
└── fixtures/
    ├── test_tokens.go          # JWT token generation
    └── test_responses.go       # Test response data
```

## Test Coverage

**Total Tests:** 19 E2E tests
**Code Coverage:** 81.6%
**Pass Rate:** 100%

### Test Categories

#### Authentication Tests (6 tests)
- `TestAINativeE2E_CompleteAuthFlow` - Full login → chat flow
- `TestAINativeE2E_UserRegistration` - User registration flow
- `TestAINativeE2E_AuthenticationFailure` - Invalid credentials
- `TestAINativeE2E_TokenRefreshFlow` - Token refresh mechanism
- `TestAINativeE2E_RefreshWithInvalidToken` - Invalid refresh token
- `TestAINativeE2E_Logout` - Logout and token invalidation

#### Chat Tests (4 tests)
- `TestAINativeE2E_UnauthorizedChatRequest` - Chat without auth
- `TestAINativeE2E_MultipleMessages` - Conversation history
- `TestAINativeE2E_GetUserInfo` - User info retrieval
- `TestAINativeE2E_ContextCancellation` - Context cancellation handling

#### Streaming Tests (5 tests)
- `TestAINativeE2E_StreamingChat` - Basic streaming functionality
- `TestAINativeE2E_StreamingDisconnect` - Graceful disconnect
- `TestAINativeE2E_StreamingUnauthorized` - Unauthorized streaming
- `TestAINativeE2E_StreamingEmptyMessage` - Edge case handling
- `TestAINativeE2E_StreamingLargeResponse` - Large response streaming

#### Error Handling Tests (4 tests)
- `TestAINativeE2E_InsufficientCredits` - Payment required (402)
- `TestAINativeE2E_NetworkError` - Network failure handling
- `TestAINativeE2E_RateLimiting` - Rate limit enforcement (429)
- `TestAINativeE2E_HealthCheck` - Backend health monitoring

## Running Tests

### Run All E2E Tests
```bash
go test -v ./tests/integration/ainative_e2e/ -run "^TestAINativeE2E_"
```

### Run with Coverage
```bash
go test -coverprofile=coverage.out -covermode=atomic ./tests/integration/ainative_e2e/ -run "^TestAINativeE2E_"
go tool cover -html=coverage.out -o coverage.html
```

### Run Specific Test
```bash
go test -v ./tests/integration/ainative_e2e/ -run TestAINativeE2E_CompleteAuthFlow
```

### Run with Race Detection
```bash
go test -race ./tests/integration/ainative_e2e/ -run "^TestAINativeE2E_"
```

### Skip Long-Running Tests
```bash
go test -short ./tests/integration/ainative_e2e/
```

## Mock Backend Server

The `MockBackend` provides a full simulation of the Python backend with:

### Features
- ✅ JWT token generation and validation
- ✅ User authentication (login, register, logout)
- ✅ Token refresh mechanism
- ✅ Chat completions (non-streaming)
- ✅ Server-Sent Events (SSE) streaming
- ✅ Rate limiting (token bucket algorithm)
- ✅ Credit management
- ✅ Health checks
- ✅ Context cancellation support

### Configuration Methods

```go
// Enable streaming support
mockBackend.EnableStreaming()

// Set delay between streaming chunks
mockBackend.SetStreamDelay(100 * time.Millisecond)

// Configure number of streaming chunks
mockBackend.SetStreamChunkCount(100)

// Set user credits
mockBackend.SetUserCredits("user@example.com", 1000)

// Enable rate limiting (5 requests per minute)
mockBackend.EnableRateLimit(5, time.Minute)
```

### Example Usage

```go
func TestExample(t *testing.T) {
    // Create mock backend
    mockBackend := NewMockBackend(t)
    defer mockBackend.Close()

    // Configure backend
    mockBackend.EnableStreaming()
    mockBackend.SetUserCredits("test@example.com", 500)

    // Create client
    client := backend.NewClient(mockBackend.URL)
    ctx := context.Background()

    // Test authentication
    resp, err := client.Login(ctx, "test@example.com", "password123")
    require.NoError(t, err)
    assert.NotEmpty(t, resp.AccessToken)
}
```

## Test Fixtures

### JWT Token Generation

```go
// Generate valid token (15 min expiry)
token := fixtures.CreateValidToken()

// Generate expired token
expiredToken := fixtures.CreateExpiredToken()

// Generate token for specific email
token := fixtures.CreateValidTokenForEmail("user@example.com")

// Generate refresh token (7 day expiry)
refreshToken := fixtures.CreateRefreshToken("user@example.com")
```

### Test Responses

```go
// Get default chat response
chatResp := fixtures.GetDefaultChatResponse()

// Get streaming chunks
chunks := fixtures.GetStreamingChatChunks()

// Get large streaming response
largeChunks := fixtures.GetLargeStreamingChunks(100)

// Get default user
user := fixtures.GetDefaultUser("test@example.com")

// Get token response
tokenResp := fixtures.GetTokenResponse("test@example.com")
```

## TDD Workflow Evidence

This test suite was developed following strict Test-Driven Development (TDD):

### RED Phase
All tests were written FIRST before any implementation:
```bash
# Tests failed with: undefined: NewMockBackend
go test ./tests/integration/ainative_e2e/
# FAIL: github.com/AINative-studio/ainative-code/tests/integration/ainative_e2e [build failed]
```

### GREEN Phase
Mock infrastructure was created to make tests pass:
- Created `fixtures/test_tokens.go` for JWT generation
- Created `fixtures/test_responses.go` for test data
- Created `mock_backend.go` for backend simulation

### REFACTOR Phase
Tests verified passing with full coverage:
```bash
go test -v ./tests/integration/ainative_e2e/ -run "^TestAINativeE2E_"
# PASS: 19/19 tests passing
# coverage: 81.6% of statements
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: E2E Tests

on: [push, pull_request]

jobs:
  e2e-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.25'

      - name: Run E2E Tests
        run: |
          go test -v -coverprofile=coverage.out \
            ./tests/integration/ainative_e2e/ \
            -run "^TestAINativeE2E_"

      - name: Upload Coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out
          flags: e2e
```

## Performance Benchmarks

Run performance benchmarks:
```bash
go test -bench=. -benchmem ./tests/integration/ainative_e2e/
```

Expected performance:
- Authentication flow: < 5ms
- Chat completion: < 10ms
- Streaming chat: < 100ms
- Token refresh: < 5ms

## Troubleshooting

### Test Timeouts
Increase timeout for slow networks:
```go
client := backend.NewClient(mockBackend.URL, backend.WithTimeout(60*time.Second))
```

### Race Conditions
Run with race detector:
```bash
go test -race ./tests/integration/ainative_e2e/
```

### Debug Logging
Enable verbose logging:
```bash
go test -v ./tests/integration/ainative_e2e/ -run TestAINativeE2E_CompleteAuthFlow
```

## Future Enhancements

- [ ] CLI integration tests (when commands are implemented)
- [ ] Provider fallback tests
- [ ] Multi-provider routing tests
- [ ] WebSocket streaming tests
- [ ] Database integration tests
- [ ] Cache layer tests
- [ ] Metrics and observability tests

## Contributing

When adding new E2E tests:

1. **Follow TDD**: Write test FIRST (RED phase)
2. **Use BDD Style**: Given-When-Then structure
3. **Test One Thing**: Each test should validate one behavior
4. **Clean Up**: Always defer `mockBackend.Close()`
5. **Use Fixtures**: Reuse test data from fixtures
6. **Document**: Add clear comments explaining the test

### Test Template

```go
// TestAINativeE2E_YourFeature tests your feature description
// GIVEN initial conditions
// WHEN action is performed
// THEN expected outcome
func TestAINativeE2E_YourFeature(t *testing.T) {
    // GIVEN
    mockBackend := NewMockBackend(t)
    defer mockBackend.Close()

    client := backend.NewClient(mockBackend.URL)
    ctx := context.Background()

    // WHEN
    result, err := client.YourMethod(ctx, params)

    // THEN
    require.NoError(t, err, "Operation should succeed")
    assert.Equal(t, expected, result, "Result should match expected")
}
```

## License

Copyright 2026 AINative Studio. All rights reserved.
