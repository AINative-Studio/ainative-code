# API Integration Tests for AINative Code Platform

Comprehensive integration tests for AINative platform external API integrations.

## Overview

This test suite validates integration with 5 external platform APIs:
- **OAuth Authentication** (PKCE flow)
- **ZeroDB** (Vector operations, NoSQL tables, Agent memory)
- **Design Token Extraction** (CSS/SCSS parsing)
- **Strapi CMS** (Blog post CRUD operations)
- **RLHF Feedback** (Rating and correction submission)

## Test Execution

**Total execution time**: ~2.8 seconds
**Total test count**: 60+ test cases
**All tests passing**: ✅

```bash
# Run all API integration tests
go test -v -timeout 3m ./tests/integration/...

# Results:
# PASS - 2.824s
```

## Test Coverage

### 1. OAuth Authentication (`auth_test.go`)

#### PKCE Parameter Generation
- ✅ Generate valid PKCE parameters (verifier, challenge, state)
- ✅ Generate unique parameters on each call
- ✅ Validate code verifier format (43-128 characters)
- ✅ Reject invalid code verifiers (too short, too long, invalid characters)

#### Authorization URL Construction
- ✅ Construct valid authorization URL with all required parameters
- ✅ Include code challenge, challenge method, and state

#### Code Exchange
- ✅ Successfully exchange authorization code for tokens
- ✅ Reject invalid authorization codes
- ✅ Reject mismatched code verifiers (PKCE validation)

#### Token Refresh
- ✅ Successfully refresh access token with valid refresh token
- ✅ Reject invalid refresh tokens

#### Error Handling
- ✅ Handle 401/403 unauthorized errors
- ✅ Handle 429 rate limit errors with Retry-After header
- ✅ Handle 500 server errors
- ✅ Handle timeout scenarios

#### State Validation
- ✅ Validate CSRF state parameter matches
- ✅ Reject mismatched state (CSRF protection)

### 2. ZeroDB Operations (`zerodb_test.go`)

#### Vector Operations
- ✅ Upsert vector embeddings with metadata
- ✅ Search vectors by similarity
- ✅ List all vectors
- ✅ Reject empty vectors

#### NoSQL Table Operations
- ✅ Create new tables
- ✅ Insert records into tables
- ✅ Query records from tables
- ✅ Update records in tables
- ✅ Reject operations on non-existent tables (404 error)

#### Agent Memory Operations
- ✅ Store agent memory with embeddings
- ✅ Search agent memory by similarity and session

#### Error Handling
- ✅ Handle 401 unauthorized errors
- ✅ Handle 429 rate limit errors
- ✅ Validate API request format (reject malformed JSON)

### 3. Design Token Extraction (`design_test.go`)

#### CSS/SCSS Parsing
- ✅ Parse CSS files and extract color tokens
- ✅ Parse SCSS files and extract typography tokens
- ✅ Extract specific token types (colors only, typography only, etc.)
- ✅ Reject empty content

#### Token Validation
- ✅ Validate correct token structure
- ✅ Reject invalid token structure (missing required fields)

#### Export Formats
- ✅ Export tokens as JSON
- ✅ Export tokens as CSS (`:root` variables)
- ✅ Export tokens as SCSS (`$variable` format)
- ✅ Reject unsupported export formats

#### Error Handling
- ✅ Handle 401 unauthorized errors
- ✅ Handle 429 rate limit errors
- ✅ Handle invalid JSON requests

### 4. Strapi CMS Operations (`strapi_test.go`)

#### Blog Post CRUD
- ✅ Create new blog posts
- ✅ List all blog posts with pagination
- ✅ Return empty list when no posts exist
- ✅ Update existing posts
- ✅ Return 404 for non-existent posts
- ✅ Delete existing posts
- ✅ Return 404 when deleting non-existent posts

#### Publish/Unpublish Workflow
- ✅ Publish draft posts
- ✅ Unpublish published posts

#### Error Handling
- ✅ Handle 401 unauthorized errors
- ✅ Handle 429 rate limit errors
- ✅ Require authorization header
- ✅ Reject invalid JSON requests

### 5. RLHF Feedback (`rlhf_test.go`)

#### Feedback Rating Submission
- ✅ Submit feedback with rating (1-5), comment, and tags
- ✅ Submit feedback with minimal required data
- ✅ Validate rating range (1-5)
- ✅ Require message_id field

#### Correction Submission
- ✅ Submit corrections for incorrect responses
- ✅ Validate correction is not empty

#### Feedback Structure Validation
- ✅ Validate complete feedback structure
- ✅ Handle multiple feedback submissions
- ✅ Track different feedback types (rating vs correction)

#### Feedback Listing
- ✅ List all submitted feedback

#### Error Handling
- ✅ Handle 401 unauthorized errors
- ✅ Handle 429 rate limit errors
- ✅ Handle invalid JSON requests
- ✅ Require API key header

## Mock Infrastructure

All tests use `httptest` mock servers with realistic API behavior:

### Mock Servers (`tests/integration/mocks/`)

| Server | File | Features |
|--------|------|----------|
| **AuthServer** | `auth_server.go` | OAuth 2.0 endpoints, JWT generation with RSA signing, PKCE validation |
| **ZeroDBServer** | `zerodb_server.go` | Vector storage, NoSQL tables, Agent memory, Thread-safe operations |
| **DesignServer** | `design_server.go` | CSS/SCSS parsing, Token extraction, Multiple export formats |
| **StrapiServer** | `strapi_server.go` | RESTful blog API, Publish/unpublish workflow, Strapi-compatible responses |
| **RLHFServer** | `rlhf_server.go` | Feedback submission, Correction tracking, Type classification |

### Mock Server Features

All mock servers support:
- ✅ Configurable error scenarios (401, 429, 500, timeout)
- ✅ Rate limiting with Retry-After headers
- ✅ Authentication validation
- ✅ Proper HTTP status codes
- ✅ Thread-safe concurrent access
- ✅ Automatic cleanup with defer

## Running Tests

### Run All API Integration Tests
```bash
go test -v ./tests/integration/...
```

### Run Specific Test Suite
```bash
# OAuth tests only
go test -v -run TestOAuth ./tests/integration/...

# ZeroDB tests only
go test -v -run TestZeroDB ./tests/integration/...

# Design token tests only
go test -v -run TestDesign ./tests/integration/...

# Strapi tests only
go test -v -run TestStrapi ./tests/integration/...

# RLHF tests only
go test -v -run TestRLHF ./tests/integration/...
```

### Run with Timeout
```bash
go test -v -timeout 3m ./tests/integration/...
```

### Run Specific Test Case
```bash
go test -v -run TestOAuthLoginFlow_PKCE/should_generate_valid_PKCE_parameters ./tests/integration/...
```

## Test Structure

All tests follow the **Given-When-Then** pattern:

```go
t.Run("should perform some action", func(t *testing.T) {
    // Given: Setup test conditions
    server := mocks.NewMockServer()
    defer server.Close()

    testData := map[string]interface{}{
        "key": "value",
    }

    // When: Execute the action
    body, _ := json.Marshal(testData)
    req, _ := http.NewRequest("POST", server.GetURL()+"/api/endpoint", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-API-Key", "test-key")

    client := &http.Client{Timeout: 5 * time.Second}
    resp, err := client.Do(req)

    // Then: Verify expectations
    require.NoError(t, err)
    defer resp.Body.Close()

    assert.Equal(t, http.StatusOK, resp.StatusCode)
})
```

## Performance Metrics

- **Total Execution Time**: 2.824 seconds
- **Average per test**: ~0.047 seconds
- **Fastest test**: < 0.001 seconds (PKCE validation)
- **Slowest test**: ~2 seconds (timeout handling)
- **All tests**: ✅ Pass

## Error Scenario Coverage

Each test suite includes comprehensive error testing:

| Error Type | Status Code | Tests |
|------------|-------------|-------|
| **Unauthorized** | 401/403 | Missing/invalid credentials |
| **Rate Limit** | 429 | Too many requests, includes Retry-After header |
| **Bad Request** | 400 | Invalid JSON, missing required fields |
| **Not Found** | 404 | Resource does not exist |
| **Server Error** | 500 | Internal server errors |
| **Timeout** | - | Request exceeds timeout duration |

## Acceptance Criteria

- [x] All 5 test scenarios implemented (OAuth, ZeroDB, Design, Strapi, RLHF)
- [x] Mock servers for all external APIs
- [x] Error scenarios tested (401, 429, 500, timeout)
- [x] Tests run in < 3 minutes (actual: ~2.8 seconds)
- [x] Proper cleanup after tests (all use defer)

## Best Practices

### 1. Test Isolation
Each test creates its own mock server instance:
```go
server := mocks.NewAuthServer()
defer server.Close()
```

### 2. Resource Cleanup
All resources properly closed with `defer`:
```go
resp, err := client.Do(req)
require.NoError(t, err)
defer resp.Body.Close()  // Always close response bodies
```

### 3. Descriptive Test Names
Test names clearly describe the scenario:
```go
t.Run("should_reject_invalid_authorization_code", func(t *testing.T) {
    // Test code
})
```

### 4. Comprehensive Coverage
Each test suite covers:
- ✅ Happy path (successful operations)
- ✅ Edge cases (empty data, minimal data)
- ✅ Error conditions (auth failures, rate limits, validation errors)

### 5. Fast Execution
Mock servers eliminate network latency:
- No real API calls
- No network delays
- Deterministic results

## CI/CD Integration

These tests are designed for continuous integration:
- ✅ No external dependencies
- ✅ Deterministic results (no flaky tests)
- ✅ Fast execution (< 3 seconds)
- ✅ Clear error messages
- ✅ Comprehensive coverage

## Mock Server Configuration

### Simulating Error Scenarios

```go
// Simulate authentication failure
server.ShouldFailAuth = true

// Simulate rate limiting
server.ShouldRateLimit = true

// Simulate timeout
server.SimulateTimeout(10 * time.Second)

// Simulate server error
server.SimulateServerError()
```

### Adding Test Data

```go
// Auth: Add valid authorization code
authServer.AddValidCode("test_code", "test_verifier")

// Auth: Add valid refresh token
authServer.AddRefreshToken("valid_refresh_token")

// Strapi: Add blog post
postID := strapiServer.AddPost(map[string]interface{}{
    "title": "Test Post",
    "content": "Test Content",
})

// Design: Add design token
designServer.AddToken(mocks.DesignToken{
    Name:  "primary-color",
    Value: "#007bff",
    Type:  "color",
})
```

## Future Enhancements

Potential areas for expansion:
- [ ] WebSocket integration tests (if applicable)
- [ ] Performance/load testing
- [ ] Contract testing with real API schemas
- [ ] Retry logic and exponential backoff testing
- [ ] Circuit breaker pattern testing
- [ ] Token expiration and automatic refresh handling
- [ ] Batch operation testing
- [ ] Concurrent request handling tests

## Debugging Tests

### Run with verbose output
```bash
go test -v ./tests/integration -run TestOAuth
```

### Run single test
```bash
go test -v ./tests/integration -run TestOAuthLoginFlow_PKCE/should_generate_valid_PKCE_parameters
```

### Check specific mock server behavior
```go
// Verify endpoint was called
assert.True(t, server.TokenCalled)
assert.True(t, server.RefreshCalled)
```

## References

- [OAuth 2.0 PKCE RFC 7636](https://datatracker.ietf.org/doc/html/rfc7636)
- [Go httptest Package](https://pkg.go.dev/net/http/httptest)
- [Testify Assertions](https://github.com/stretchr/testify)
- [JWT RS256 Signing](https://jwt.io/introduction)
