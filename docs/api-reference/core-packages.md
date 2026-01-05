# Core Packages API Reference

This document covers the core packages of AINative Code: client, config, session, and database.

## Table of Contents

- [Client Package](#client-package)
- [Session Package](#session-package)
- [Database Package](#database-package)

## Client Package

**Import Path**: `github.com/AINative-studio/ainative-code/internal/client`

The client package provides an HTTP client for interacting with AINative platform APIs.

### Types

#### Client

```go
type Client struct {
    // Contains filtered or unexported fields
}
```

The Client represents an HTTP client for AINative platform API interactions with automatic JWT authentication, retry logic, and error handling.

### Functions

#### New

```go
func New(opts ...Option) *Client
```

Creates a new API client with the specified options.

**Parameters**:
- `opts` - Functional options for configuring the client

**Returns**:
- `*Client` - Configured client instance

**Example**:

```go
package main

import (
    "time"
    "github.com/AINative-studio/ainative-code/internal/client"
    "github.com/AINative-studio/ainative-code/internal/auth"
)

func main() {
    // Create auth client
    authClient, _ := auth.NewClient(auth.DefaultClientOptions())

    // Create API client with options
    apiClient := client.New(
        client.WithBaseURL("https://api.ainative.studio"),
        client.WithTimeout(30 * time.Second),
        client.WithMaxRetries(3),
        client.WithAuthClient(authClient),
    )
}
```

### Methods

#### Get

```go
func (c *Client) Get(ctx context.Context, path string, opts ...RequestOption) ([]byte, error)
```

Performs a GET request to the specified path.

**Parameters**:
- `ctx` - Context for cancellation and timeout
- `path` - API endpoint path (e.g., "/api/v1/users")
- `opts` - Optional request configuration

**Returns**:
- `[]byte` - Response body
- `error` - Error if request fails

**Example**:

```go
ctx := context.Background()
response, err := client.Get(ctx, "/api/v1/health")
if err != nil {
    log.Fatalf("GET request failed: %v", err)
}
fmt.Printf("Response: %s\n", response)
```

#### Post

```go
func (c *Client) Post(ctx context.Context, path string, body interface{}, opts ...RequestOption) ([]byte, error)
```

Performs a POST request with JSON body.

**Parameters**:
- `ctx` - Context for cancellation and timeout
- `path` - API endpoint path
- `body` - Request body (will be JSON-marshaled)
- `opts` - Optional request configuration

**Returns**:
- `[]byte` - Response body
- `error` - Error if request fails

**Example**:

```go
type CreateUserRequest struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

ctx := context.Background()
req := CreateUserRequest{
    Name:  "John Doe",
    Email: "john@example.com",
}

response, err := client.Post(ctx, "/api/v1/users", req)
if err != nil {
    log.Fatalf("POST request failed: %v", err)
}
```

#### Put

```go
func (c *Client) Put(ctx context.Context, path string, body interface{}, opts ...RequestOption) ([]byte, error)
```

Performs a PUT request with JSON body.

#### Patch

```go
func (c *Client) Patch(ctx context.Context, path string, body interface{}, opts ...RequestOption) ([]byte, error)
```

Performs a PATCH request with JSON body.

#### Delete

```go
func (c *Client) Delete(ctx context.Context, path string, opts ...RequestOption) ([]byte, error)
```

Performs a DELETE request.

### Client Options

#### WithBaseURL

```go
func WithBaseURL(baseURL string) Option
```

Sets the base URL for API requests.

#### WithTimeout

```go
func WithTimeout(timeout time.Duration) Option
```

Sets the HTTP request timeout (default: 30 seconds).

#### WithMaxRetries

```go
func WithMaxRetries(maxRetries int) Option
```

Sets the maximum number of retry attempts for failed requests (default: 3).

#### WithAuthClient

```go
func WithAuthClient(authClient auth.Client) Option
```

Sets the authentication client for JWT token management.

#### WithHTTPClient

```go
func WithHTTPClient(httpClient *http.Client) Option
```

Sets a custom HTTP client.

### Request Options

#### WithHeader

```go
func WithHeader(key, value string) RequestOption
```

Adds a custom header to the request.

**Example**:

```go
response, err := client.Get(ctx, "/api/v1/data",
    client.WithHeader("X-Custom-Header", "value"),
)
```

#### WithHeaders

```go
func WithHeaders(headers map[string]string) RequestOption
```

Adds multiple custom headers.

#### WithQueryParam

```go
func WithQueryParam(key, value string) RequestOption
```

Adds a query parameter to the request.

**Example**:

```go
response, err := client.Get(ctx, "/api/v1/users",
    client.WithQueryParam("page", "1"),
    client.WithQueryParam("limit", "10"),
)
```

#### WithQueryParams

```go
func WithQueryParams(params map[string]string) RequestOption
```

Adds multiple query parameters.

#### WithSkipAuth

```go
func WithSkipAuth() RequestOption
```

Skips JWT token injection for this request (useful for public endpoints).

**Example**:

```go
response, err := client.Get(ctx, "/api/v1/public/info",
    client.WithSkipAuth(),
)
```

#### WithDisableRetry

```go
func WithDisableRetry() RequestOption
```

Disables retry logic for this request.

### Error Handling

The client automatically retries requests on the following HTTP status codes:
- 429 (Too Many Requests)
- 500 (Internal Server Error)
- 502 (Bad Gateway)
- 503 (Service Unavailable)
- 504 (Gateway Timeout)

Retry logic uses exponential backoff: 1s, 2s, 4s, 8s...

**Example**:

```go
response, err := client.Get(ctx, "/api/v1/data")
if err != nil {
    // Check if it's a 401 error
    if strings.Contains(err.Error(), "HTTP 401") {
        // Re-authenticate
        tokens, _ := authClient.Authenticate(ctx)
    }
    return err
}
```

### Best Practices

1. **Always use context with timeout**:
```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
```

2. **Handle authentication errors**:
```go
response, err := client.Get(ctx, "/api/v1/data")
if err != nil {
    if strings.Contains(err.Error(), "authentication failed") {
        // Re-authenticate user
    }
}
```

3. **Use request options for customization**:
```go
response, err := client.Post(ctx, "/api/v1/data", body,
    client.WithHeader("X-Request-ID", requestID),
    client.WithQueryParam("async", "true"),
)
```

---

## Session Package

**Import Path**: `github.com/AINative-studio/ainative-code/internal/session`

The session package provides types and utilities for managing conversation sessions and messages.

### Types

#### Session

```go
type Session struct {
    ID          string         `json:"id"`
    Name        string         `json:"name"`
    CreatedAt   time.Time      `json:"created_at"`
    UpdatedAt   time.Time      `json:"updated_at"`
    Status      SessionStatus  `json:"status"`
    Model       *string        `json:"model,omitempty"`
    Temperature *float64       `json:"temperature,omitempty"`
    MaxTokens   *int64         `json:"max_tokens,omitempty"`
    Settings    map[string]any `json:"settings,omitempty"`
}
```

Represents a conversation session.

**Example**:

```go
session := session.Session{
    ID:        "sess_123",
    Name:      "Code Review Session",
    Status:    session.StatusActive,
    CreatedAt: time.Now(),
    UpdatedAt: time.Now(),
    Settings: map[string]any{
        "theme": "dark",
        "auto_save": true,
    },
}
```

#### SessionStatus

```go
type SessionStatus string

const (
    StatusActive   SessionStatus = "active"
    StatusArchived SessionStatus = "archived"
    StatusDeleted  SessionStatus = "deleted"
)
```

**Methods**:

```go
func (s SessionStatus) IsValid() bool
```

Checks if a session status is valid.

#### Message

```go
type Message struct {
    ID           string         `json:"id"`
    SessionID    string         `json:"session_id"`
    Role         MessageRole    `json:"role"`
    Content      string         `json:"content"`
    Timestamp    time.Time      `json:"timestamp"`
    ParentID     *string        `json:"parent_id,omitempty"`
    TokensUsed   *int64         `json:"tokens_used,omitempty"`
    Model        *string        `json:"model,omitempty"`
    FinishReason *string        `json:"finish_reason,omitempty"`
    Metadata     map[string]any `json:"metadata,omitempty"`
}
```

Represents a conversation message.

**Example**:

```go
message := session.Message{
    ID:        "msg_456",
    SessionID: "sess_123",
    Role:      session.RoleUser,
    Content:   "How do I implement authentication?",
    Timestamp: time.Now(),
}
```

#### MessageRole

```go
type MessageRole string

const (
    RoleUser      MessageRole = "user"
    RoleAssistant MessageRole = "assistant"
    RoleSystem    MessageRole = "system"
    RoleTool      MessageRole = "tool"
)
```

**Methods**:

```go
func (r MessageRole) IsValid() bool
```

#### SessionSummary

```go
type SessionSummary struct {
    Session
    MessageCount int64 `json:"message_count"`
    TotalTokens  int64 `json:"total_tokens,omitempty"`
}
```

Session with summary information.

#### SessionExport

```go
type SessionExport struct {
    Session  Session   `json:"session"`
    Messages []Message `json:"messages"`
}
```

Represents an exported session with all messages.

#### ExportFormat

```go
type ExportFormat string

const (
    ExportFormatJSON     ExportFormat = "json"
    ExportFormatMarkdown ExportFormat = "markdown"
    ExportFormatText     ExportFormat = "text"
)
```

### Functions

#### MarshalSettings

```go
func MarshalSettings(settings map[string]any) (string, error)
```

Marshals settings map to JSON string for database storage.

**Example**:

```go
settings := map[string]any{
    "theme": "dark",
    "auto_save": true,
}

jsonStr, err := session.MarshalSettings(settings)
if err != nil {
    log.Fatalf("Failed to marshal settings: %v", err)
}
```

#### UnmarshalSettings

```go
func UnmarshalSettings(data string) (map[string]any, error)
```

Unmarshals JSON string to settings map.

#### MarshalMetadata

```go
func MarshalMetadata(metadata map[string]any) (string, error)
```

Marshals metadata map to JSON string.

#### UnmarshalMetadata

```go
func UnmarshalMetadata(data string) (map[string]any, error)
```

Unmarshals JSON string to metadata map.

### Usage Examples

#### Creating a Session

```go
import (
    "time"
    "github.com/AINative-studio/ainative-code/internal/session"
)

// Create new session
sess := session.Session{
    ID:        generateID(),
    Name:      "My Session",
    Status:    session.StatusActive,
    CreatedAt: time.Now(),
    UpdatedAt: time.Now(),
}

// Set optional parameters
model := "claude-3-sonnet"
temp := 0.7
maxTok := int64(4096)

sess.Model = &model
sess.Temperature = &temp
sess.MaxTokens = &maxTok
```

#### Adding Messages

```go
// Create user message
userMsg := session.Message{
    ID:        generateID(),
    SessionID: sess.ID,
    Role:      session.RoleUser,
    Content:   "What is the capital of France?",
    Timestamp: time.Now(),
}

// Create assistant message
assistantMsg := session.Message{
    ID:        generateID(),
    SessionID: sess.ID,
    Role:      session.RoleAssistant,
    Content:   "The capital of France is Paris.",
    Timestamp: time.Now(),
}
```

#### Exporting a Session

```go
// Create export
export := session.SessionExport{
    Session: sess,
    Messages: []session.Message{userMsg, assistantMsg},
}

// Convert to JSON
jsonData, err := json.Marshal(export)
if err != nil {
    log.Fatalf("Failed to export session: %v", err)
}

// Save to file
os.WriteFile("session_export.json", jsonData, 0644)
```

---

## Database Package

**Import Path**: `github.com/AINative-studio/ainative-code/internal/database`

The database package provides a wrapper around SQLC-generated queries with transaction support and connection management.

### Types

#### DB

```go
type DB struct {
    *Queries
    // Contains filtered or unexported fields
}
```

DB wraps the SQLC Queries interface with additional functionality.

#### ConnectionConfig

```go
type ConnectionConfig struct {
    Driver          string
    DSN             string
    MaxOpenConns    int
    MaxIdleConns    int
    ConnMaxLifetime time.Duration
}
```

Configuration for database connection.

### Functions

#### NewDB

```go
func NewDB(sqlDB *sql.DB) *DB
```

Creates a new database instance with the given connection.

#### NewFromConfig

```go
func NewFromConfig(config *ConnectionConfig) (*DB, error)
```

Creates a new database instance from configuration.

**Example**:

```go
config := &database.ConnectionConfig{
    Driver:          "sqlite3",
    DSN:             "ainative.db",
    MaxOpenConns:    10,
    MaxIdleConns:    5,
    ConnMaxLifetime: 1 * time.Hour,
}

db, err := database.NewFromConfig(config)
if err != nil {
    log.Fatalf("Failed to create database: %v", err)
}
defer db.Close()
```

#### Initialize

```go
func Initialize(config *ConnectionConfig) (*DB, error)
```

Sets up the database connection and runs migrations.

**Example**:

```go
db, err := database.Initialize(&database.ConnectionConfig{
    Driver: "sqlite3",
    DSN:    "ainative.db",
})
if err != nil {
    log.Fatalf("Failed to initialize database: %v", err)
}
defer db.Close()
```

#### InitializeContext

```go
func InitializeContext(ctx context.Context, config *ConnectionConfig) (*DB, error)
```

Sets up the database connection and runs migrations with context.

### Methods

#### Close

```go
func (d *DB) Close() error
```

Closes the database connection.

#### Health

```go
func (d *DB) Health() error
```

Performs a health check on the database.

**Example**:

```go
if err := db.Health(); err != nil {
    log.Printf("Database health check failed: %v", err)
}
```

#### DB

```go
func (d *DB) DB() *sql.DB
```

Returns the underlying *sql.DB instance.

#### WithTx

```go
func (d *DB) WithTx(ctx context.Context, fn func(*Queries) error) error
```

Executes a function within a database transaction.

**Example**:

```go
ctx := context.Background()

err := db.WithTx(ctx, func(q *database.Queries) error {
    // Create session
    sess, err := q.CreateSession(ctx, database.CreateSessionParams{
        ID:   "sess_123",
        Name: "My Session",
    })
    if err != nil {
        return err
    }

    // Create message
    _, err = q.CreateMessage(ctx, database.CreateMessageParams{
        ID:        "msg_456",
        SessionID: sess.ID,
        Role:      "user",
        Content:   "Hello",
    })
    return err
})

if err != nil {
    log.Fatalf("Transaction failed: %v", err)
}
```

#### WithTxOptions

```go
func (d *DB) WithTxOptions(ctx context.Context, opts *sql.TxOptions, fn func(*Queries) error) error
```

Executes a function within a database transaction with custom options.

**Example**:

```go
opts := &sql.TxOptions{
    Isolation: sql.LevelSerializable,
    ReadOnly:  false,
}

err := db.WithTxOptions(ctx, opts, func(q *database.Queries) error {
    // Perform database operations
    return nil
})
```

#### ExecInTx

```go
func (d *DB) ExecInTx(ctx context.Context, query string, args ...interface{}) error
```

Executes SQL in a transaction.

#### Stats

```go
func (d *DB) Stats() sql.DBStats
```

Returns database statistics.

**Example**:

```go
stats := db.Stats()
fmt.Printf("Open connections: %d\n", stats.OpenConnections)
fmt.Printf("In use: %d\n", stats.InUse)
fmt.Printf("Idle: %d\n", stats.Idle)
```

#### Ping

```go
func (d *DB) Ping() error
```

Verifies the database connection is alive.

#### PingContext

```go
func (d *DB) PingContext(ctx context.Context) error
```

Verifies the database connection with context.

### Usage Examples

#### Basic Database Operations

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/AINative-studio/ainative-code/internal/database"
)

func main() {
    ctx := context.Background()

    // Initialize database
    db, err := database.Initialize(&database.ConnectionConfig{
        Driver: "sqlite3",
        DSN:    "ainative.db",
    })
    if err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }
    defer db.Close()

    // Create a session
    sess, err := db.CreateSession(ctx, database.CreateSessionParams{
        ID:     "sess_" + generateID(),
        Name:   "Code Review",
        Status: "active",
    })
    if err != nil {
        log.Fatalf("Failed to create session: %v", err)
    }

    log.Printf("Created session: %s", sess.ID)
}
```

#### Using Transactions

```go
// Perform multiple operations atomically
err := db.WithTx(ctx, func(q *database.Queries) error {
    // Create session
    sess, err := q.CreateSession(ctx, database.CreateSessionParams{
        ID:     "sess_123",
        Name:   "Transaction Test",
        Status: "active",
    })
    if err != nil {
        return err
    }

    // Add messages
    for i := 0; i < 5; i++ {
        _, err := q.CreateMessage(ctx, database.CreateMessageParams{
            ID:        fmt.Sprintf("msg_%d", i),
            SessionID: sess.ID,
            Role:      "user",
            Content:   fmt.Sprintf("Message %d", i),
        })
        if err != nil {
            return err // Transaction will rollback
        }
    }

    return nil // Transaction will commit
})

if err != nil {
    log.Printf("Transaction failed and rolled back: %v", err)
}
```

#### Monitoring Database Health

```go
import "time"

// Periodic health check
ticker := time.NewTicker(30 * time.Second)
defer ticker.Stop()

go func() {
    for range ticker.C {
        if err := db.Health(); err != nil {
            log.Printf("Database health check failed: %v", err)
            // Take remedial action
        }

        stats := db.Stats()
        log.Printf("DB Stats - Open: %d, InUse: %d, Idle: %d",
            stats.OpenConnections, stats.InUse, stats.Idle)
    }
}()
```

### Best Practices

1. **Always close database connections**:
```go
db, err := database.Initialize(config)
if err != nil {
    log.Fatal(err)
}
defer db.Close()
```

2. **Use transactions for multiple related operations**:
```go
err := db.WithTx(ctx, func(q *database.Queries) error {
    // Multiple operations that should be atomic
    return nil
})
```

3. **Set appropriate connection pool settings**:
```go
config := &database.ConnectionConfig{
    Driver:          "sqlite3",
    DSN:             "ainative.db",
    MaxOpenConns:    25,  // Limit concurrent connections
    MaxIdleConns:    5,   // Keep some connections idle
    ConnMaxLifetime: 1 * time.Hour,
}
```

4. **Use context for all database operations**:
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

session, err := db.GetSession(ctx, sessionID)
```

5. **Monitor database statistics**:
```go
stats := db.Stats()
if stats.WaitCount > 100 {
    log.Println("High wait count, consider increasing MaxOpenConns")
}
```
