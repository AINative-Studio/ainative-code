# Database Usage Guide

## Quick Start

### Initialize Database

```go
import "github.com/AINative-studio/ainative-code/internal/database"

// Create configuration
config := database.DefaultConfig("./data/ainative.db")

// Initialize database with migrations
db, err := database.Initialize(config)
if err != nil {
    log.Fatal(err)
}
defer db.Close()
```

### Using In-Memory Database (for testing)

```go
config := database.DefaultConfig(":memory:")
db, err := database.Initialize(config)
if err != nil {
    log.Fatal(err)
}
defer db.Close()
```

## Common Operations

### Session Management

```go
import (
    "context"
    "github.com/google/uuid"
)

ctx := context.Background()

// Create a session
sessionID := uuid.New().String()
err := db.CreateSession(ctx, database.CreateSessionParams{
    ID:          sessionID,
    Name:        "My Conversation",
    Status:      "active",
    Model:       strPtr("gpt-4"),
    Temperature: float64Ptr(0.7),
    MaxTokens:   int64Ptr(2000),
})

// Get a session
session, err := db.GetSession(ctx, sessionID)

// List sessions with pagination
sessions, err := db.ListSessions(ctx, database.ListSessionsParams{
    Limit:  10,
    Offset: 0,
})

// Archive a session
err = db.ArchiveSession(ctx, sessionID)

// Delete a session (soft delete)
err = db.DeleteSession(ctx, sessionID)

// Helper function for nullable strings
func strPtr(s string) *string { return &s }
func float64Ptr(f float64) *float64 { return &f }
func int64Ptr(i int64) *int64 { return &i }
```

### Message Operations

```go
// Create a message
messageID := uuid.New().String()
err := db.CreateMessage(ctx, database.CreateMessageParams{
    ID:           messageID,
    SessionID:    sessionID,
    Role:         "user",
    Content:      "Hello, AI!",
    TokensUsed:   int64Ptr(5),
    Model:        strPtr("gpt-4"),
})

// Get all messages for a session
messages, err := db.ListMessagesBySession(ctx, sessionID)

// Get messages by role
userMessages, err := db.ListMessagesByRole(ctx, database.ListMessagesByRoleParams{
    SessionID: sessionID,
    Role:      "user",
})

// Get total tokens used in session
totalTokens, err := db.GetTotalTokensUsed(ctx, sessionID)

// Search messages
results, err := db.SearchMessages(ctx, database.SearchMessagesParams{
    SessionID: sessionID,
    Content:   "%keyword%",
    Limit:     10,
    Offset:    0,
})
```

### Tool Execution Tracking

```go
// Create tool execution
toolExecID := uuid.New().String()
err := db.CreateToolExecution(ctx, database.CreateToolExecutionParams{
    ID:        toolExecID,
    MessageID: messageID,
    ToolName:  "calculator",
    Input:     `{"operation": "add", "a": 5, "b": 3}`,
    Status:    "pending",
    StartedAt: time.Now().Format(time.RFC3339),
})

// Update with result
err = db.UpdateToolExecutionOutput(ctx, database.UpdateToolExecutionOutputParams{
    Output:     strPtr(`{"result": 8}`),
    Status:     "success",
    DurationMs: int64Ptr(150),
    ID:         toolExecID,
})

// Get tool execution statistics
stats, err := db.GetToolExecutionStats(ctx, "calculator")
fmt.Printf("Total: %d, Success: %d, Failed: %d, Avg Duration: %.2fms\n",
    stats.TotalExecutions,
    stats.SuccessfulExecutions,
    stats.FailedExecutions,
    stats.AvgDurationMs,
)

// Get pending executions
pending, err := db.GetPendingToolExecutions(ctx)
```

### Metadata Operations

```go
// Set metadata
err := db.SetMetadata(ctx, database.SetMetadataParams{
    Key:   "app_version",
    Value: "1.0.0",
})

// Get metadata
metadata, err := db.GetMetadata(ctx, "app_version")
fmt.Println(metadata.Value)

// List all metadata
allMetadata, err := db.ListMetadata(ctx)
```

## Transactions

### Using Transactions

```go
// Execute operations in a transaction
err := db.WithTx(ctx, func(q *database.Queries) error {
    // Create session
    if err := q.CreateSession(ctx, sessionParams); err != nil {
        return err
    }

    // Create initial message
    if err := q.CreateMessage(ctx, messageParams); err != nil {
        return err
    }

    // All operations committed together
    return nil
})

// If any error occurs, all operations are rolled back
if err != nil {
    log.Printf("Transaction failed: %v", err)
}
```

## Migration Management

### Check Migration Status

```go
status, err := database.GetStatus(db.DB())
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Current Version: %d\n", status.CurrentVersion)
fmt.Printf("Applied Migrations: %d\n", len(status.Applied))
fmt.Printf("Pending Migrations: %d\n", len(status.Pending))

for _, migration := range status.Applied {
    fmt.Printf("  ✓ %03d_%s (applied %s)\n",
        migration.Version,
        migration.Name,
        migration.AppliedAt.Format(time.RFC3339),
    )
}
```

### Rollback Migration

```go
err := database.Rollback(db.DB())
if err != nil {
    log.Fatal(err)
}
```

## Configuration Options

### Custom Configuration

```go
config := &database.ConnectionConfig{
    Path:            "./data/ainative.db",
    MaxOpenConns:    25,
    MaxIdleConns:    10,
    ConnMaxLifetime: 2 * time.Hour,
    ConnMaxIdleTime: 30 * time.Minute,
    BusyTimeout:     10000, // 10 seconds
    JournalMode:     "WAL",
    Synchronous:     "NORMAL",
}

db, err := database.NewFromConfig(config)
```

### Health Check

```go
if err := db.Health(); err != nil {
    log.Printf("Database health check failed: %v", err)
} else {
    log.Println("Database is healthy")
}
```

### Connection Statistics

```go
stats := db.Stats()
fmt.Printf("Open Connections: %d/%d\n", stats.OpenConnections, stats.MaxOpenConnections)
fmt.Printf("Idle Connections: %d\n", stats.Idle)
fmt.Printf("In Use: %d\n", stats.InUse)
fmt.Printf("Wait Count: %d\n", stats.WaitCount)
```

## Best Practices

### 1. Always Use Context

```go
// Good
ctx := context.Background()
session, err := db.GetSession(ctx, sessionID)

// Better - with timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
session, err := db.GetSession(ctx, sessionID)
```

### 2. Use Transactions for Related Operations

```go
err := db.WithTx(ctx, func(q *database.Queries) error {
    // Multiple related operations
    if err := q.CreateSession(ctx, sessionParams); err != nil {
        return err
    }
    if err := q.CreateMessage(ctx, messageParams); err != nil {
        return err
    }
    return nil
})
```

### 3. Handle Nullable Fields Properly

```go
// Create helper functions
func strPtr(s string) *string {
    if s == "" {
        return nil
    }
    return &s
}

// Use them in queries
params := database.CreateSessionParams{
    Model: strPtr(modelName), // nil if empty, pointer otherwise
}
```

### 4. Use Pagination for Large Result Sets

```go
const pageSize = 50

for offset := 0; ; offset += pageSize {
    messages, err := db.ListMessagesBySessionPaginated(ctx,
        database.ListMessagesBySessionPaginatedParams{
            SessionID: sessionID,
            Limit:     pageSize,
            Offset:    int64(offset),
        },
    )
    if err != nil {
        return err
    }
    if len(messages) == 0 {
        break
    }

    // Process messages
    for _, msg := range messages {
        fmt.Println(msg.Content)
    }
}
```

### 5. Soft Delete for Audit Trail

```go
// Use soft delete for important data
err := db.DeleteSession(ctx, sessionID) // Sets status='deleted'

// Use hard delete only when necessary
err := db.HardDeleteSession(ctx, sessionID) // Permanently removes
```

## Error Handling

```go
import "github.com/AINative-studio/ainative-code/internal/errors"

session, err := db.GetSession(ctx, sessionID)
if err != nil {
    // Check error type
    if dbErr, ok := err.(*errors.DatabaseError); ok {
        switch dbErr.Code() {
        case errors.ErrCodeDBNotFound:
            log.Printf("Session not found: %s", sessionID)
        case errors.ErrCodeDBConnection:
            log.Printf("Database connection error: %v", err)
        default:
            log.Printf("Database error: %v", err)
        }
    }
    return err
}
```

## Testing

### Setup Test Database

```go
func setupTestDB(t *testing.T) *database.DB {
    config := database.DefaultConfig(":memory:")
    db, err := database.Initialize(config)
    if err != nil {
        t.Fatalf("failed to initialize test database: %v", err)
    }
    return db
}

func TestMyFunction(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()

    // Run tests
}
```

## Timestamps

SQLite stores timestamps as TEXT in RFC3339 format. When working with timestamps:

```go
// Storing timestamps
now := time.Now().Format(time.RFC3339)

// Parsing timestamps
timestamp, err := time.Parse(time.RFC3339, session.CreatedAt)
```

## Common Queries

### Get Latest Message in Session

```go
message, err := db.GetLatestMessage(ctx, sessionID)
```

### Count Messages by Role

```go
count, err := db.GetMessageCountByRole(ctx, database.GetMessageCountByRoleParams{
    SessionID: sessionID,
    Role:      "assistant",
})
```

### Search Sessions

```go
sessions, err := db.SearchSessions(ctx, database.SearchSessionsParams{
    Name:   "%keyword%",
    Name_2: "%keyword%", // Second parameter for ID search
    Limit:  10,
    Offset: 0,
})
```

### Get Tool Usage Statistics

```go
usageStats, err := db.GetToolUsageStats(ctx)
for _, stat := range usageStats {
    fmt.Printf("%s: %d uses, %.1f%% success rate\n",
        stat.ToolName,
        stat.TotalUses,
        float64(stat.SuccessfulUses)/float64(stat.TotalUses)*100,
    )
}
```

## Performance Tips

1. **Use indexes**: All common query patterns are already indexed
2. **Batch operations**: Use transactions for multiple related operations
3. **Connection pooling**: Configured by default for optimal performance
4. **WAL mode**: Enabled for better concurrent read performance
5. **Cache size**: Set to 64MB for faster queries

## Troubleshooting

### Database Locked Errors

```go
// Increase busy timeout
config := database.DefaultConfig("./data/ainative.db")
config.BusyTimeout = 10000 // 10 seconds
```

### Migration Failures

```go
// Check migration status
status, _ := database.GetStatus(db.DB())
fmt.Printf("Current version: %d\n", status.CurrentVersion)

// Rollback if needed
database.Rollback(db.DB())
```

### Connection Pool Exhaustion

```go
// Increase pool size
config.MaxOpenConns = 50
config.MaxIdleConns = 25
```
