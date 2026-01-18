# Backend HTTP Client

Go HTTP client package for communicating with the AINative Python backend.

## Overview

This package provides a type-safe, well-tested HTTP client for the Go CLI to interact with the Python backend API running at `http://localhost:8000`.

**Key Features:**
- Strict TDD implementation (89.9% test coverage)
- Type-safe request/response structures
- Bearer token authentication
- Comprehensive error handling
- Configurable timeouts
- Context support for cancellation

## Installation

```go
import "github.com/AINative-studio/ainative-code/internal/backend"
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/AINative-studio/ainative-code/internal/backend"
)

func main() {
    // Create client
    client := backend.NewClient("http://localhost:8000")
    ctx := context.Background()

    // Login
    resp, err := client.Login(ctx, "user@example.com", "password")
    if err != nil {
        log.Fatal(err)
    }

    // Use token for authenticated requests
    token := resp.AccessToken

    // Send chat completion
    chatReq := &backend.ChatCompletionRequest{
        Messages: []backend.Message{
            {Role: "user", Content: "Hello!"},
        },
        Model: "claude-sonnet-4-5",
    }

    chatResp, err := client.ChatCompletion(ctx, token, chatReq)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(chatResp.Choices[0].Message.Content)
}
```

## API Reference

### Client Initialization

#### `NewClient(baseURL string, opts ...ClientOption) *Client`
Creates a new HTTP client with default 30s timeout.

```go
// Default timeout (30s)
client := backend.NewClient("http://localhost:8000")

// Custom timeout
client := backend.NewClient("http://localhost:8000",
    backend.WithTimeout(60*time.Second))
```

### Authentication Methods

#### `Login(ctx context.Context, email, password string) (*TokenResponse, error)`
Authenticates user with email and password.

```go
resp, err := client.Login(ctx, "user@example.com", "password123")
if err != nil {
    return err
}
accessToken := resp.AccessToken
refreshToken := resp.RefreshToken
```

#### `Register(ctx context.Context, email, password string) (*TokenResponse, error)`
Creates a new user account.

```go
resp, err := client.Register(ctx, "newuser@example.com", "password123")
```

#### `RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error)`
Refreshes an access token.

```go
resp, err := client.RefreshToken(ctx, oldRefreshToken)
newAccessToken := resp.AccessToken
```

#### `Logout(ctx context.Context, accessToken string) error`
Logs out the current user.

```go
err := client.Logout(ctx, accessToken)
```

#### `GetMe(ctx context.Context, accessToken string) (*User, error)`
Retrieves current user information.

```go
user, err := client.GetMe(ctx, accessToken)
fmt.Printf("User: %s (%s)\n", user.Email, user.ID)
```

### Chat Completion

#### `ChatCompletion(ctx context.Context, accessToken string, req *ChatCompletionRequest) (*ChatCompletionResponse, error)`
Sends a chat completion request.

```go
req := &backend.ChatCompletionRequest{
    Messages: []backend.Message{
        {Role: "system", Content: "You are a helpful assistant."},
        {Role: "user", Content: "What is the capital of France?"},
    },
    Model:       "claude-sonnet-4-5",
    Temperature: 0.7,      // optional
    MaxTokens:   1000,     // optional
}

resp, err := client.ChatCompletion(ctx, accessToken, req)
if err != nil {
    return err
}

answer := resp.Choices[0].Message.Content
```

### Health Check

#### `HealthCheck(ctx context.Context) error`
Checks if the backend is healthy.

```go
err := client.HealthCheck(ctx)
if err != nil {
    log.Fatal("Backend is unhealthy:", err)
}
```

## Error Handling

The client provides sentinel errors for common HTTP status codes:

```go
import "errors"

resp, err := client.Login(ctx, "user@example.com", "wrongpassword")
if errors.Is(err, backend.ErrUnauthorized) {
    fmt.Println("Invalid credentials")
}

resp, err := client.ChatCompletion(ctx, token, req)
if errors.Is(err, backend.ErrPaymentRequired) {
    fmt.Println("Insufficient credits")
}
```

### Available Error Types

| Error | HTTP Status | Description |
|-------|-------------|-------------|
| `ErrUnauthorized` | 401 | Invalid or missing authentication |
| `ErrPaymentRequired` | 402 | Insufficient credits |
| `ErrBadRequest` | 400 | Invalid request format |
| `ErrNotFound` | 404 | Resource not found |
| `ErrServerError` | 500, 502, 503 | Server-side error |

## Types

### Request Types

```go
type LoginRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

type RegisterRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

type RefreshTokenRequest struct {
    RefreshToken string `json:"refresh_token"`
}

type ChatCompletionRequest struct {
    Messages    []Message `json:"messages"`
    Model       string    `json:"model"`
    Temperature float64   `json:"temperature,omitempty"`
    MaxTokens   int       `json:"max_tokens,omitempty"`
    Stream      bool      `json:"stream,omitempty"`
}

type Message struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}
```

### Response Types

```go
type TokenResponse struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    TokenType    string `json:"token_type"`
    User         User   `json:"user,omitempty"`
}

type User struct {
    ID    string `json:"id"`
    Email string `json:"email"`
}

type ChatCompletionResponse struct {
    ID      string   `json:"id"`
    Model   string   `json:"model"`
    Choices []Choice `json:"choices"`
    Usage   Usage    `json:"usage,omitempty"`
}

type Choice struct {
    Message Message `json:"message"`
    Index   int     `json:"index,omitempty"`
}

type Usage struct {
    PromptTokens     int `json:"prompt_tokens,omitempty"`
    CompletionTokens int `json:"completion_tokens,omitempty"`
    TotalTokens      int `json:"total_tokens,omitempty"`
}
```

## Testing

Run the test suite:

```bash
# Run all tests
go test -v ./internal/backend/

# Run tests with coverage
go test -cover ./internal/backend/

# Generate coverage report
go test -coverprofile=coverage.out ./internal/backend/
go tool cover -html=coverage.out
```

**Test Coverage:** 89.9% (20/20 tests passing)

## Architecture

The client uses a centralized request handler (`doRequest`) that:
1. Marshals request body to JSON
2. Creates HTTP request with context
3. Sets appropriate headers (Content-Type, Authorization)
4. Executes request with timeout
5. Handles HTTP status codes
6. Unmarshals response body

This ensures consistent error handling and reduces code duplication.

## Best Practices

### 1. Always Use Context

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

resp, err := client.Login(ctx, email, password)
```

### 2. Handle Errors Appropriately

```go
resp, err := client.ChatCompletion(ctx, token, req)
if err != nil {
    if errors.Is(err, backend.ErrUnauthorized) {
        // Refresh token or re-authenticate
    } else if errors.Is(err, backend.ErrPaymentRequired) {
        // Show payment required message
    } else {
        // Handle other errors
    }
}
```

### 3. Store Tokens Securely

```go
// Don't log or print tokens
log.Printf("User logged in: %s", resp.User.Email) // Good
log.Printf("Access token: %s", resp.AccessToken)  // BAD!

// Store tokens securely (e.g., encrypted config file)
config.SetSecure("access_token", resp.AccessToken)
```

### 4. Refresh Tokens Proactively

```go
// Check token expiration and refresh before it expires
if tokenNearExpiry(accessToken) {
    resp, err := client.RefreshToken(ctx, refreshToken)
    if err != nil {
        // Re-authenticate if refresh fails
        return reAuthenticate()
    }
    accessToken = resp.AccessToken
}
```

## Backend API Endpoints

The client communicates with these Python backend endpoints:

| Method | Endpoint | Auth Required |
|--------|----------|---------------|
| POST | `/api/v1/auth/login` | No |
| POST | `/api/v1/auth/register` | No |
| POST | `/api/v1/auth/refresh` | No |
| POST | `/api/v1/auth/logout` | Yes |
| GET | `/api/v1/auth/me` | Yes |
| POST | `/api/v1/chat/completions` | Yes |
| GET | `/health` | No |

## Examples

See `/Users/aideveloper/AINative-Code/examples/backend-client-usage.go` for comprehensive usage examples.

## License

Proprietary - AINative Platform

## See Also

- [Week 1 Python Backend](../../backend/)
- [TDD Completion Report](../../docs/ISSUE_158_COMPLETION_REPORT.md)
- [Client Usage Examples](../../examples/backend-client-usage.go)
