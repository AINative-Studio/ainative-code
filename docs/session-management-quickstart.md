# Session Management Quick Start Guide

## Overview

The AINative-Code session management system provides comprehensive conversation persistence, full-text search, and multi-format export capabilities.

## Quick Setup

### 1. Database Initialization

```go
import (
    "github.com/AINative-studio/ainative-code/internal/database"
    "github.com/AINative-studio/ainative-code/internal/session"
)

// Initialize database with default config
config := database.DefaultConfig("~/.ainative/ainative.db")
db, err := database.Initialize(config)
if err != nil {
    log.Fatal(err)
}
defer db.Close()

// Create session manager
mgr := session.NewSQLiteManager(db)
```

### 2. Create a Session

```go
import "github.com/google/uuid"

sess := &session.Session{
    ID:          uuid.New().String(),
    Name:        "My Project Discussion",
    Status:      session.StatusActive,
    Model:       strPtr("claude-3-5-sonnet-20241022"),
    Temperature: float64Ptr(0.7),
    MaxTokens:   int64Ptr(4096),
    Settings: map[string]any{
        "theme":     "dark",
        "auto_save": true,
    },
}

err = mgr.CreateSession(context.Background(), sess)
```

### 3. Add Messages

```go
// User message
userMsg := &session.Message{
    ID:        uuid.New().String(),
    SessionID: sess.ID,
    Role:      session.RoleUser,
    Content:   "How do I implement authentication in Go?",
    TokensUsed: int64Ptr(25),
}
err = mgr.AddMessage(context.Background(), userMsg)

// Assistant response
assistantMsg := &session.Message{
    ID:           uuid.New().String(),
    SessionID:    sess.ID,
    Role:         session.RoleAssistant,
    Content:      "Here's how to implement authentication...",
    ParentID:     &userMsg.ID,
    TokensUsed:   int64Ptr(150),
    Model:        strPtr("claude-3-5-sonnet-20241022"),
    FinishReason: strPtr("end_turn"),
}
err = mgr.AddMessage(context.Background(), assistantMsg)
```

### 4. Retrieve Session with Messages

```go
// Get session
session, err := mgr.GetSession(context.Background(), sess.ID)

// Get all messages
messages, err := mgr.GetMessages(context.Background(), sess.ID)

// Get paginated messages
messages, err := mgr.GetMessagesPaginated(context.Background(), sess.ID, 10, 0)

// Get session with summary
summary, err := mgr.GetSessionSummary(context.Background(), sess.ID)
fmt.Printf("Session: %s (%d messages, %d tokens)\n",
    summary.Name, summary.MessageCount, summary.TotalTokens)
```

### 5. Search Messages

```go
// Basic search
opts := &session.SearchOptions{
    Query:  "authentication",
    Limit:  50,
    Offset: 0,
}
results, err := mgr.SearchAllMessages(context.Background(), opts)

// Search with date range
dateFrom := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
dateTo := time.Date(2026, 1, 5, 23, 59, 59, 0, time.UTC)
opts.DateFrom = &dateFrom
opts.DateTo = &dateTo
results, err = mgr.SearchAllMessages(context.Background(), opts)

// Search with provider filter
opts.Provider = "claude"
results, err = mgr.SearchAllMessages(context.Background(), opts)

// Process results
for _, result := range results.Results {
    fmt.Printf("Session: %s\n", result.SessionName)
    fmt.Printf("Score: %.2f\n", result.RelevanceScore)
    fmt.Printf("Snippet: %s\n\n", result.Snippet)
}
```

### 6. Export Session

```go
// Export to JSON
file, _ := os.Create("session-export.json")
defer file.Close()

exporter := session.NewExporter(&session.ExporterOptions{
    IncludeMetadata: true,
    PrettyPrint:     true,
})

err = exporter.ExportToJSON(file, session, messages)

// Export to Markdown
mdFile, _ := os.Create("session-export.md")
defer mdFile.Close()
err = exporter.ExportToMarkdown(mdFile, session, messages)

// Export to HTML
htmlFile, _ := os.Create("session-export.html")
defer htmlFile.Close()
err = exporter.ExportToHTML(htmlFile, session, messages)
```

### 7. List Sessions

```go
// List all active sessions
sessions, err := mgr.ListSessions(context.Background())

// List with filters
sessions, err = mgr.ListSessions(
    context.Background(),
    session.WithStatus(session.StatusActive),
    session.WithLimit(10),
    session.WithOffset(0),
)

// List archived sessions
sessions, err = mgr.ListSessions(
    context.Background(),
    session.WithStatus(session.StatusArchived),
)
```

### 8. Update Session

```go
// Modify session
session.Name = "Updated Project Name"
session.Settings["theme"] = "light"

err = mgr.UpdateSession(context.Background(), session)
```

### 9. Delete Session

```go
// Soft delete (status = 'deleted')
err = mgr.DeleteSession(context.Background(), sess.ID)

// Archive session (status = 'archived')
err = mgr.ArchiveSession(context.Background(), sess.ID)

// Hard delete (permanent removal)
err = mgr.HardDeleteSession(context.Background(), sess.ID)
```

## CLI Usage

### Session Commands

```bash
# Create session
ainative-code session create "My Project"

# List sessions
ainative-code session list
ainative-code session list --all --limit 50

# Show session details
ainative-code session show <session-id>

# Delete session
ainative-code session delete <session-id>

# View messages
ainative-code session messages <session-id>
```

### Export Commands

```bash
# Export to JSON
ainative-code session export <session-id>

# Export to Markdown
ainative-code session export <session-id> --format markdown -o conversation.md

# Export to HTML
ainative-code session export <session-id> --format html -o report.html

# Custom template
ainative-code session export <session-id> --template custom.tmpl -o output.txt
```

### Search Commands

```bash
# Basic search
ainative-code session search "authentication"

# Limit results
ainative-code session search "golang" --limit 10

# Date range
ainative-code session search "error" \
  --date-from "2026-01-01" \
  --date-to "2026-01-05"

# Provider filter
ainative-code session search "explain" --provider claude

# JSON output
ainative-code session search "database" --json
```

## Helper Functions

```go
// String pointer helper
func strPtr(s string) *string {
    return &s
}

// Float64 pointer helper
func float64Ptr(f float64) *float64 {
    return &f
}

// Int64 pointer helper
func int64Ptr(i int64) *int64 {
    return &i
}
```

## Error Handling

```go
import "errors"

// Check for specific errors
if errors.Is(err, session.ErrSessionNotFound) {
    log.Println("Session not found")
}

if errors.Is(err, session.ErrInvalidSessionID) {
    log.Println("Invalid session ID")
}

// Get error context
var sessErr *session.SessionError
if errors.As(err, &sessErr) {
    log.Printf("Operation: %s, Context: %s, Error: %v",
        sessErr.Op, sessErr.Context, sessErr.Err)
}
```

## Transaction Example

```go
// The manager handles transactions internally
// For custom transactions, use the database directly
err := db.WithTx(ctx, func(q *database.Queries) error {
    // Create session
    if err := q.CreateSession(ctx, params); err != nil {
        return err
    }

    // Add messages
    for _, msg := range messages {
        if err := q.CreateMessage(ctx, msgParams); err != nil {
            return err
        }
    }

    return nil
})
```

## Environment Variables

```bash
# Set custom database path
export AINATIVE_DB_PATH=~/my-custom-path/sessions.db

# FTS5 support (if needed)
export CGO_ENABLED=1
```

## Best Practices

1. **Always use context with timeout:**
   ```go
   ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
   defer cancel()
   ```

2. **Close database connections:**
   ```go
   defer db.Close()
   ```

3. **Use pagination for large result sets:**
   ```go
   messages, err := mgr.GetMessagesPaginated(ctx, sessionID, 100, offset)
   ```

4. **Validate input before operations:**
   ```go
   if sessionID == "" {
       return errors.New("session ID cannot be empty")
   }
   ```

5. **Handle errors appropriately:**
   ```go
   if err != nil {
       log.Printf("Failed to create session: %v", err)
       return err
   }
   ```

## Common Patterns

### Resume Session Flow

```go
// 1. List recent sessions
sessions, _ := mgr.ListSessions(ctx, session.WithLimit(10))

// 2. Select session to resume
sessionID := sessions[0].ID

// 3. Load messages
messages, _ := mgr.GetMessages(ctx, sessionID)

// 4. Display conversation history
for _, msg := range messages {
    fmt.Printf("[%s]: %s\n", msg.Role, msg.Content)
}

// 5. Add new message and continue
newMsg := &session.Message{
    ID:        uuid.New().String(),
    SessionID: sessionID,
    Role:      session.RoleUser,
    Content:   "Continue from where we left off...",
}
mgr.AddMessage(ctx, newMsg)

// 6. Touch session to update timestamp
mgr.TouchSession(ctx, sessionID)
```

### Conversation Thread Navigation

```go
// Get full conversation thread from a message
messageID := "some-message-id"
thread, err := mgr.GetConversationThread(ctx, messageID)

// Thread is returned in chronological order
for _, msg := range thread {
    indent := ""
    if msg.ParentID != nil {
        indent = "  "
    }
    fmt.Printf("%s[%s]: %s\n", indent, msg.Role, msg.Content)
}
```

### Search and Export Workflow

```go
// 1. Search for relevant conversations
results, _ := mgr.SearchAllMessages(ctx, &session.SearchOptions{
    Query:  "authentication implementation",
    Limit:  5,
})

// 2. Select most relevant session
sessionID := results.Results[0].Message.SessionID

// 3. Get full session
session, _ := mgr.GetSession(ctx, sessionID)
messages, _ := mgr.GetMessages(ctx, sessionID)

// 4. Export for documentation
exporter := session.NewExporter(nil)
file, _ := os.Create("auth-discussion.md")
defer file.Close()
exporter.ExportToMarkdown(file, session, messages)
```

## Troubleshooting

### FTS5 Not Available

If you get "no such module: fts5" error:

```bash
# Option 1: Build with CGO
CGO_ENABLED=1 go build

# Option 2: Use system SQLite
brew install sqlite3  # macOS
apt-get install libsqlite3-dev  # Ubuntu

# Option 3: Use Docker with FTS5
docker run -v $(pwd):/app golang:1.21 bash -c \
  "apt-get update && apt-get install -y libsqlite3-dev && go test ./..."
```

### Database Locked Error

If you get "database is locked" error:

```go
// Increase busy timeout
config := database.DefaultConfig(dbPath)
config.BusyTimeout = 10000  // 10 seconds

// Use WAL mode for better concurrency
config.JournalMode = "WAL"
```

### Migration Errors

If migrations fail:

```go
// Check migration status
status, err := database.GetStatus(db.DB())
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Current version: %d\n", status.CurrentVersion)
fmt.Printf("Pending migrations: %d\n", len(status.Pending))

// Rollback last migration if needed
err = database.Rollback(db.DB())
```

## Additional Resources

- **Full Documentation:** `/docs/session-management-completion-report.md`
- **Architecture:** See "Architecture Diagram" section in completion report
- **Test Examples:** `/internal/session/*_test.go`
- **CLI Source:** `/internal/cmd/session.go`
- **Database Schema:** `/internal/database/schema/schema.sql`

## Support

For issues or questions:
1. Check the completion report for detailed information
2. Review test files for usage examples
3. Examine CLI commands for integration patterns
4. Consult error definitions in `internal/session/errors.go`
