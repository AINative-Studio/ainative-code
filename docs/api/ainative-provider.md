# AINative Provider API Reference

## Overview

This document provides comprehensive API reference for AINative's cloud authentication and hosted inference commands.

## Table of Contents

- [Authentication Commands](#authentication-commands)
- [Chat Commands](#chat-commands)
- [Configuration Commands](#configuration-commands)
- [Go Package API](#go-package-api)
- [HTTP API Reference](#http-api-reference)

---

## Authentication Commands

### `auth login-backend`

Authenticate with AINative platform using email and password.

**Syntax:**
```bash
ainative-code auth login-backend --email EMAIL --password PASSWORD
```

**Flags:**

| Flag | Short | Type | Required | Description |
|------|-------|------|----------|-------------|
| `--email` | `-e` | string | Yes | User email address |
| `--password` | `-p` | string | Yes | User password |

**Examples:**

```bash
# Basic login
ainative-code auth login-backend \
  --email user@example.com \
  --password mypassword

# Short form
ainative-code auth login-backend -e user@example.com -p mypassword
```

**Response:**
```
Successfully logged in as user@example.com
```

**Exit Codes:**
- `0` - Success
- `1` - Authentication failed
- `2` - Network error

**Stored Data:**
- Access token (in config file)
- Refresh token (in config file)
- User email
- User ID

**Related Commands:**
- `auth logout-backend` - Logout
- `auth refresh-backend` - Refresh token
- `auth whoami` - Check status

---

### `auth logout-backend`

Clear stored authentication credentials.

**Syntax:**
```bash
ainative-code auth logout-backend
```

**Flags:** None

**Examples:**

```bash
# Logout from AINative
ainative-code auth logout-backend
```

**Response:**
```
Successfully logged out
```

**What it does:**
1. Calls backend logout endpoint (if token is valid)
2. Clears access token from config
3. Clears refresh token from config
4. Removes user email and ID

**Exit Codes:**
- `0` - Success
- `1` - Error

**Related Commands:**
- `auth login-backend` - Login again

---

### `auth refresh-backend`

Refresh the access token using the stored refresh token.

**Syntax:**
```bash
ainative-code auth refresh-backend
```

**Flags:** None

**Examples:**

```bash
# Refresh access token
ainative-code auth refresh-backend
```

**Response:**
```
Token refreshed successfully
```

**When to use:**
- Access token has expired
- Before critical operations requiring authentication
- Testing token refresh flow

**Requirements:**
- Valid refresh token in config
- Backend must be accessible

**Exit Codes:**
- `0` - Success
- `1` - Refresh failed (no refresh token or token invalid)

**Related Commands:**
- `auth login-backend` - Login if refresh fails

---

### `auth whoami`

Display current authentication status and user information.

**Syntax:**
```bash
ainative-code auth whoami
```

**Flags:** None

**Examples:**

```bash
# Check authentication status
ainative-code auth whoami
```

**Response (authenticated):**
```
Authenticated User:
  Email: user@example.com
  Token Type: Bearer
  Expires In: 15m 30s
```

**Response (not authenticated):**
```
Not authenticated

Run 'ainative-code auth login-backend' to authenticate
```

**Exit Codes:**
- `0` - Success (shows status regardless)

**Related Commands:**
- `auth login-backend` - Login
- `auth token status` - Detailed token info

---

### `auth token status`

Show detailed token expiration information.

**Syntax:**
```bash
ainative-code auth token status
```

**Flags:** None

**Examples:**

```bash
# View token status
ainative-code auth token status
```

**Response:**
```
Token Status:
─────────────────────────────────────────
Access Token:  eyJhbGciOiJIUzI1...
Refresh Token: eyJhbGciOiJIUzI1...
Token Type:    Bearer
Expires At:    Mon, 17 Jan 2026 15:30:00 PST
Time Until Expiry: 15m 30s

Status: ✓ VALID
```

**Status Indicators:**
- `✓ VALID` - Token is valid and not expiring soon
- `⚠️ EXPIRING SOON` - Token expires in < 5 minutes
- `❌ EXPIRED` - Token has expired

**Exit Codes:**
- `0` - Success

---

### `auth token refresh`

Manually refresh access token (alias for `auth refresh-backend`).

**Syntax:**
```bash
ainative-code auth token refresh
```

See [`auth refresh-backend`](#auth-refresh-backend) for details.

---

## Chat Commands

### `chat-ainative`

Send chat messages using AINative backend with intelligent provider selection.

**Syntax:**
```bash
ainative-code chat-ainative [OPTIONS] [MESSAGE]
```

**Flags:**

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--message` | `-m` | string | - | Message to send (required) |
| `--model` | - | string | auto | Model to use |
| `--provider` | - | string | auto | Provider to use |
| `--auto-provider` | - | bool | false | Enable automatic provider selection |
| `--stream` | - | bool | false | Enable streaming responses |
| `--verbose` | `-v` | bool | false | Show detailed information |

**Examples:**

**Basic chat:**
```bash
ainative-code chat-ainative --message "Hello, world!"
```

**Short form:**
```bash
ainative-code chat-ainative -m "Explain quantum computing"
```

**Positional argument:**
```bash
ainative-code chat-ainative "Write a Python function"
```

**Specific model:**
```bash
ainative-code chat-ainative \
  --message "Explain REST APIs" \
  --model claude-sonnet-4-5
```

**Auto provider selection:**
```bash
ainative-code chat-ainative \
  --message "Hello" \
  --auto-provider
```

**Manual provider:**
```bash
ainative-code chat-ainative \
  --message "Hello" \
  --provider anthropic
```

**Streaming:**
```bash
ainative-code chat-ainative \
  --message "Count from 1 to 10" \
  --stream
```

**Verbose mode:**
```bash
ainative-code chat-ainative \
  --message "Hello" \
  --verbose
```

**Verbose output includes:**
- Selected provider
- Model used
- Request parameters
- Response time
- Token usage
- Credits consumed (future)

**Exit Codes:**
- `0` - Success
- `1` - Request failed
- `2` - Authentication required

**Requirements:**
- Valid access token (use `auth login-backend`)
- Backend must be running
- Internet connectivity

**Related Commands:**
- `auth login-backend` - Authenticate first
- `auth whoami` - Check authentication

---

## Configuration Commands

### `config get`

Get configuration value.

**Syntax:**
```bash
ainative-code config get KEY
```

**Examples:**

```bash
# Get backend URL
ainative-code config get backend_url

# Get preferred provider
ainative-code config get ainative.preferred_provider
```

---

### `config set`

Set configuration value.

**Syntax:**
```bash
ainative-code config set KEY VALUE
```

**Examples:**

```bash
# Set backend URL
ainative-code config set backend_url "http://localhost:8000"

# Set preferred provider
ainative-code config set ainative.preferred_provider anthropic
```

---

## Go Package API

### Backend Client

**Package:** `github.com/AINative-studio/ainative-code/internal/backend`

#### NewClient

Create a new backend client.

```go
func NewClient(baseURL string) *Client
```

**Parameters:**
- `baseURL` - Backend API base URL (e.g., "http://localhost:8000")

**Returns:**
- `*Client` - Backend client instance

**Example:**
```go
import "github.com/AINative-studio/ainative-code/internal/backend"

client := backend.NewClient("http://localhost:8000")
```

---

#### Login

Authenticate with email and password.

```go
func (c *Client) Login(ctx context.Context, email, password string) (*LoginResponse, error)
```

**Parameters:**
- `ctx` - Context for cancellation
- `email` - User email address
- `password` - User password

**Returns:**
- `*LoginResponse` - Login response with tokens
- `error` - Error if login fails

**LoginResponse Structure:**
```go
type LoginResponse struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    TokenType    string `json:"token_type"`
    ExpiresIn    int    `json:"expires_in"`
    User         User   `json:"user"`
}

type User struct {
    ID      string `json:"id"`
    Email   string `json:"email"`
    Credits int    `json:"credits"`
}
```

**Example:**
```go
ctx := context.Background()
resp, err := client.Login(ctx, "user@example.com", "password")
if err != nil {
    log.Fatalf("Login failed: %v", err)
}

fmt.Printf("Logged in as %s\n", resp.User.Email)
fmt.Printf("Access Token: %s\n", resp.AccessToken)
```

---

#### ChatCompletion

Send chat completion request.

```go
func (c *Client) ChatCompletion(ctx context.Context, token string, req *ChatCompletionRequest) (*ChatCompletionResponse, error)
```

**Parameters:**
- `ctx` - Context for cancellation
- `token` - Access token
- `req` - Chat completion request

**Returns:**
- `*ChatCompletionResponse` - Chat response
- `error` - Error if request fails

**ChatCompletionRequest Structure:**
```go
type ChatCompletionRequest struct {
    Messages     []Message `json:"messages"`
    Model        string    `json:"model,omitempty"`
    Provider     string    `json:"provider,omitempty"`
    Stream       bool      `json:"stream,omitempty"`
    MaxTokens    int       `json:"max_tokens,omitempty"`
    Temperature  float64   `json:"temperature,omitempty"`
}

type Message struct {
    Role    string `json:"role"`    // "user", "assistant", "system"
    Content string `json:"content"`
}
```

**ChatCompletionResponse Structure:**
```go
type ChatCompletionResponse struct {
    ID       string   `json:"id"`
    Model    string   `json:"model"`
    Provider string   `json:"provider"`
    Choices  []Choice `json:"choices"`
    Usage    Usage    `json:"usage"`
}

type Choice struct {
    Index   int     `json:"index"`
    Message Message `json:"message"`
}

type Usage struct {
    PromptTokens     int `json:"prompt_tokens"`
    CompletionTokens int `json:"completion_tokens"`
    TotalTokens      int `json:"total_tokens"`
}
```

**Example:**
```go
req := &backend.ChatCompletionRequest{
    Messages: []backend.Message{
        {Role: "user", Content: "Hello, world!"},
    },
    Model: "claude-sonnet-4-5",
}

resp, err := client.ChatCompletion(ctx, accessToken, req)
if err != nil {
    log.Fatalf("Chat failed: %v", err)
}

fmt.Printf("Response: %s\n", resp.Choices[0].Message.Content)
```

---

#### RefreshToken

Refresh access token.

```go
func (c *Client) RefreshToken(ctx context.Context, refreshToken string) (*RefreshResponse, error)
```

**Parameters:**
- `ctx` - Context for cancellation
- `refreshToken` - Refresh token

**Returns:**
- `*RefreshResponse` - New tokens
- `error` - Error if refresh fails

**RefreshResponse Structure:**
```go
type RefreshResponse struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    TokenType    string `json:"token_type"`
    ExpiresIn    int    `json:"expires_in"`
}
```

**Example:**
```go
resp, err := client.RefreshToken(ctx, refreshToken)
if err != nil {
    log.Fatalf("Refresh failed: %v", err)
}

fmt.Printf("New access token: %s\n", resp.AccessToken)
```

---

### Provider Selector

**Package:** `github.com/AINative-studio/ainative-code/internal/provider`

#### NewSelector

Create a new provider selector.

```go
func NewSelector(opts ...SelectorOption) *Selector
```

**Options:**
- `WithProviders(...string)` - Set available providers
- `WithUserPreference(string)` - Set user preferred provider
- `WithFallback(bool)` - Enable/disable fallback

**Example:**
```go
import "github.com/AINative-studio/ainative-code/internal/provider"

selector := provider.NewSelector(
    provider.WithProviders("anthropic", "openai", "google"),
    provider.WithUserPreference("anthropic"),
    provider.WithFallback(true),
)
```

---

#### Select

Select best provider for request.

```go
func (s *Selector) Select(ctx context.Context, user *User) (string, error)
```

**Parameters:**
- `ctx` - Context for cancellation
- `user` - User information (credits, preferences)

**Returns:**
- `string` - Selected provider name
- `error` - Error if no provider available

**Example:**
```go
user := &provider.User{
    Email:   "test@example.com",
    Credits: 1000,
}

providerName, err := selector.Select(ctx, user)
if err != nil {
    log.Fatalf("Provider selection failed: %v", err)
}

fmt.Printf("Selected provider: %s\n", providerName)
```

---

## HTTP API Reference

### Authentication Endpoints

#### POST /v1/auth/login

Login with email and password.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "yourpassword"
}
```

**Response (200 OK):**
```json
{
  "access_token": "eyJhbGc...",
  "refresh_token": "eyJhbGc...",
  "token_type": "Bearer",
  "expires_in": 900,
  "user": {
    "id": "user123",
    "email": "user@example.com",
    "credits": 1000
  }
}
```

**Errors:**
- `400 Bad Request` - Invalid request body
- `401 Unauthorized` - Invalid credentials

---

#### POST /v1/auth/refresh

Refresh access token.

**Request:**
```json
{
  "refresh_token": "eyJhbGc..."
}
```

**Response (200 OK):**
```json
{
  "access_token": "eyJhbGc...",
  "refresh_token": "eyJhbGc...",
  "token_type": "Bearer",
  "expires_in": 900
}
```

**Errors:**
- `400 Bad Request` - Missing refresh token
- `401 Unauthorized` - Invalid refresh token

---

#### POST /v1/auth/logout

Logout and invalidate tokens.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response (200 OK):**
```json
{
  "message": "Successfully logged out"
}
```

---

### Chat Endpoints

#### POST /v1/chat/completions

Send chat completion request.

**Headers:**
```
Authorization: Bearer <access_token>
Content-Type: application/json
```

**Request:**
```json
{
  "messages": [
    {"role": "user", "content": "Hello, world!"}
  ],
  "model": "claude-sonnet-4-5",
  "provider": "anthropic",
  "stream": false,
  "max_tokens": 4096,
  "temperature": 0.7
}
```

**Response (200 OK):**
```json
{
  "id": "chatcmpl-123",
  "model": "claude-sonnet-4-5",
  "provider": "anthropic",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "Hello! How can I help you today?"
      }
    }
  ],
  "usage": {
    "prompt_tokens": 10,
    "completion_tokens": 15,
    "total_tokens": 25
  }
}
```

**Errors:**
- `401 Unauthorized` - Invalid or expired token
- `402 Payment Required` - Insufficient credits
- `400 Bad Request` - Invalid request
- `503 Service Unavailable` - Provider unavailable

---

## Error Codes

| HTTP Code | Error Type | Description |
|-----------|------------|-------------|
| 400 | Bad Request | Invalid request parameters |
| 401 | Unauthorized | Authentication required or invalid token |
| 402 | Payment Required | Insufficient credits |
| 404 | Not Found | Resource not found |
| 429 | Too Many Requests | Rate limit exceeded |
| 500 | Internal Server Error | Server error |
| 503 | Service Unavailable | Provider or service unavailable |

---

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `AINATIVE_BACKEND_URL` | Backend API URL | `http://localhost:8000` |
| `AINATIVE_ACCESS_TOKEN` | Access token | - |
| `AINATIVE_REFRESH_TOKEN` | Refresh token | - |
| `AINATIVE_PREFERRED_PROVIDER` | Preferred provider | `anthropic` |
| `AINATIVE_TIMEOUT` | Request timeout (seconds) | `120` |

---

## Next Steps

- [Getting Started Guide](../guides/ainative-getting-started.md) - Setup instructions
- [Authentication Guide](../guides/authentication.md) - Authentication details
- [Hosted Inference Guide](../guides/hosted-inference.md) - Chat features
- [Troubleshooting Guide](../guides/troubleshooting.md) - Common issues
