# Debugging Guide

## Overview

This guide covers debugging techniques, tools, and best practices for troubleshooting issues in AINative Code.

## Debugging with Delve

### Installing Delve

```bash
go install github.com/go-delve/delve/cmd/dlv@latest

# Verify installation
dlv version
```

### Basic Usage

#### Debug a Package

```bash
# Debug the main application
dlv debug ./cmd/ainative-code

# Debug with arguments
dlv debug ./cmd/ainative-code -- chat "Hello, world"

# Debug a specific package
dlv debug ./internal/provider/anthropic
```

#### Debug Tests

```bash
# Debug a specific test
dlv test ./internal/provider -- -test.run TestProviderChat

# Debug with verbose output
dlv test ./internal/provider -- -test.v -test.run TestProviderChat
```

### Delve Commands

```bash
# Set breakpoint
(dlv) break main.main
(dlv) break provider.go:123
(dlv) break (*Provider).Chat

# List breakpoints
(dlv) breakpoints

# Continue execution
(dlv) continue
(dlv) c

# Step into
(dlv) step
(dlv) s

# Step over
(dlv) next
(dlv) n

# Step out
(dlv) stepout

# Print variable
(dlv) print myVar
(dlv) p myVar

# Print with formatting
(dlv) print %#v myStruct

# List source code
(dlv) list
(dlv) list provider.go:100

# Show goroutines
(dlv) goroutines

# Switch goroutine
(dlv) goroutine 5

# Show stack trace
(dlv) stack
(dlv) bt

# Set condition breakpoint
(dlv) condition 1 count > 10

# Clear breakpoint
(dlv) clear 1
(dlv) clearall

# Exit
(dlv) quit
```

### Advanced Delve Usage

#### Remote Debugging

```bash
# Start headless server
dlv debug --headless --listen=:2345 --api-version=2 ./cmd/ainative-code

# Connect from another terminal or IDE
dlv connect :2345
```

#### Core Dump Analysis

```bash
# Generate core dump
GOTRACEBACK=crash ./ainative-code

# Analyze core dump
dlv core ./ainative-code core.12345
```

## IDE Debugging

### Visual Studio Code

#### Launch Configuration

Create `.vscode/launch.json`:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug Application",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/cmd/ainative-code",
      "args": ["chat"],
      "env": {
        "AINATIVE_LOG_LEVEL": "debug"
      }
    },
    {
      "name": "Debug Current Test",
      "type": "go",
      "request": "launch",
      "mode": "test",
      "program": "${fileDirname}",
      "args": ["-test.run", "^${selectedText}$"]
    },
    {
      "name": "Debug Specific Test",
      "type": "go",
      "request": "launch",
      "mode": "test",
      "program": "${workspaceFolder}/internal/provider",
      "args": ["-test.run", "TestProviderChat", "-test.v"]
    },
    {
      "name": "Attach to Process",
      "type": "go",
      "request": "attach",
      "mode": "local",
      "processId": "${command:pickProcess}"
    }
  ]
}
```

#### Debugging Steps

1. Set breakpoints by clicking in the gutter
2. Press F5 or select "Debug Application" from Run menu
3. Use debug toolbar to step through code:
   - F10: Step over
   - F11: Step into
   - Shift+F11: Step out
   - F5: Continue
4. Inspect variables in Debug panel
5. Add watches for expressions

### GoLand / IntelliJ IDEA

#### Debug Configuration

1. Run → Edit Configurations
2. Add New → Go Build
3. Configure:
   - Name: "Debug Chat"
   - Run kind: "Package"
   - Package path: `github.com/AINative-studio/ainative-code/cmd/ainative-code`
   - Program arguments: `chat`
   - Environment: `AINATIVE_LOG_LEVEL=debug`

#### Debugging Features

- Set breakpoints by clicking gutter
- Conditional breakpoints (right-click breakpoint)
- Evaluate expression (Alt+F8)
- Watch variables
- Step through code
- View goroutines

## Logging for Debugging

### Enable Debug Logging

```bash
# Via environment variable
export AINATIVE_LOG_LEVEL=debug
ainative-code chat

# Via flag
ainative-code --verbose chat

# Via config file
cat > ~/.config/ainative-code/config.yaml <<EOF
logging:
  level: debug
  format: text
EOF
```

### Using Logger in Code

```go
import "github.com/AINative-studio/ainative-code/internal/logger"

func processRequest(ctx context.Context, req Request) {
    // Debug logging
    logger.Debug("Processing request", map[string]interface{}{
        "request_id": req.ID,
        "user_id":    req.UserID,
    })

    // Info logging
    logger.Info("Request processed successfully")

    // Error logging with context
    if err := validate(req); err != nil {
        logger.Error("Validation failed", map[string]interface{}{
            "error":      err.Error(),
            "request_id": req.ID,
        })
        return
    }

    // Context-aware logging
    log := logger.WithContext(ctx)
    log.InfoWithFields("Database query", map[string]interface{}{
        "query":    "SELECT * FROM users",
        "duration": duration.Milliseconds(),
    })
}
```

### Log Levels

- **DEBUG**: Detailed information for diagnosing problems
- **INFO**: General informational messages
- **WARN**: Warning messages for potential issues
- **ERROR**: Error messages for failures
- **FATAL**: Critical errors that cause termination

### Log Output Formats

**Text Format** (development):
```
2025-01-05T10:30:45Z INFO Processing request request_id=req-123
```

**JSON Format** (production):
```json
{"level":"info","time":"2025-01-05T10:30:45Z","message":"Processing request","request_id":"req-123"}
```

## Profiling

### CPU Profiling

```bash
# Run with CPU profiling
go test -cpuprofile=cpu.prof -bench=. ./internal/provider

# Analyze profile
go tool pprof cpu.prof

# Commands in pprof
(pprof) top10        # Top 10 functions by CPU time
(pprof) list Chat    # Show annotated source for Chat function
(pprof) web          # Open graph in browser (requires graphviz)
(pprof) pdf > cpu.pdf # Generate PDF report
```

### Memory Profiling

```bash
# Run with memory profiling
go test -memprofile=mem.prof -bench=. ./internal/provider

# Analyze profile
go tool pprof mem.prof

# Commands in pprof
(pprof) top10 -alloc_space    # Top allocations
(pprof) top10 -inuse_space    # Currently in use
(pprof) list NewProvider       # Memory usage in function
```

### Live Profiling

```go
import (
    "net/http"
    _ "net/http/pprof"
)

func main() {
    // Start pprof server
    go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()

    // Rest of application
}
```

Access profiles at:
- CPU: http://localhost:6060/debug/pprof/profile?seconds=30
- Heap: http://localhost:6060/debug/pprof/heap
- Goroutines: http://localhost:6060/debug/pprof/goroutine

```bash
# Analyze live profile
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
```

### Trace Analysis

```bash
# Generate trace
go test -trace=trace.out ./internal/provider

# Analyze trace
go tool trace trace.out
```

Opens web interface showing:
- Goroutine execution timeline
- Network blocking
- Synchronization blocking
- Syscalls
- GC events

## Race Detection

### Running with Race Detector

```bash
# Run tests with race detector
go test -race ./...

# Build with race detector
go build -race ./cmd/ainative-code

# Run with race detector
./ainative-code chat
```

### Race Detection Output

```
WARNING: DATA RACE
Read at 0x00c0001234 by goroutine 7:
  main.incrementCounter()
      /path/to/file.go:10 +0x3e

Previous write at 0x00c0001234 by goroutine 8:
  main.incrementCounter()
      /path/to/file.go:10 +0x5a
```

### Fixing Race Conditions

```go
// Bad: Race condition
var counter int
func incrementCounter() {
    counter++ // Not thread-safe
}

// Good: Use mutex
var (
    counter int
    mu      sync.Mutex
)
func incrementCounter() {
    mu.Lock()
    defer mu.Unlock()
    counter++
}

// Better: Use atomic
var counter int64
func incrementCounter() {
    atomic.AddInt64(&counter, 1)
}
```

## Common Issues and Solutions

### Issue 1: Nil Pointer Dereference

**Symptom**:
```
panic: runtime error: invalid memory address or nil pointer dereference
```

**Debug**:
```bash
# Get full stack trace
GOTRACEBACK=all ./ainative-code chat
```

**Solution**:
```go
// Always check for nil
if provider == nil {
    return fmt.Errorf("provider is nil")
}

// Use safe navigation
if resp != nil && resp.Body != nil {
    defer resp.Body.Close()
}
```

### Issue 2: Goroutine Leaks

**Symptom**: Memory usage grows over time

**Debug**:
```bash
# Check goroutine count
curl http://localhost:6060/debug/pprof/goroutine?debug=1
```

**Solution**:
```go
// Always close channels
defer close(events)

// Always cancel contexts
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// Ensure goroutines can exit
go func() {
    defer close(done)
    for {
        select {
        case <-ctx.Done():
            return  // Goroutine can exit
        case item := <-items:
            process(item)
        }
    }
}()
```

### Issue 3: Context Cancellation Not Working

**Debug**:
```go
func processWithLogging(ctx context.Context) {
    for i := 0; i < 1000; i++ {
        select {
        case <-ctx.Done():
            logger.Debug("Context cancelled", map[string]interface{}{
                "iteration": i,
                "error":     ctx.Err(),
            })
            return
        default:
        }

        // Process
        logger.Debug("Processing", map[string]interface{}{"iteration": i})
        time.Sleep(100 * time.Millisecond)
    }
}
```

**Solution**:
- Check context regularly in loops
- Pass context to blocking operations
- Use `context.WithTimeout` for operations with deadlines

### Issue 4: High Memory Usage

**Debug**:
```bash
# Memory profile
go test -memprofile=mem.prof -bench=BenchmarkChat
go tool pprof -alloc_space mem.prof

# Check for leaks
go tool pprof -inuse_space mem.prof
```

**Solution**:
```go
// Limit buffer sizes
events := make(chan Event, 10)  // Buffered, not unlimited

// Close readers
defer resp.Body.Close()

// Reuse buffers
var bufPool = sync.Pool{
    New: func() interface{} {
        return new(bytes.Buffer)
    },
}

buf := bufPool.Get().(*bytes.Buffer)
defer func() {
    buf.Reset()
    bufPool.Put(buf)
}()
```

### Issue 5: Slow Performance

**Debug**:
```bash
# CPU profile
go test -cpuprofile=cpu.prof -bench=.
go tool pprof cpu.prof

# Find hotspots
(pprof) top10
(pprof) list functionName
```

**Solution**:
- Reduce allocations in hot paths
- Use sync.Pool for frequent allocations
- Cache computed values
- Use buffered channels
- Avoid unnecessary string conversions

## Network Debugging

### Using Wireshark

1. Start Wireshark
2. Filter: `tcp.port == 443 and host api.anthropic.com`
3. Inspect TLS handshake and HTTP/2 frames

### Using tcpdump

```bash
# Capture HTTP traffic
sudo tcpdump -i any -s 0 -w capture.pcap 'tcp port 443'

# Read capture
tcpdump -r capture.pcap -A
```

### HTTP Debugging Proxy

```bash
# Use mitmproxy
mitmproxy -p 8080

# Configure proxy in application
export HTTP_PROXY=http://localhost:8080
export HTTPS_PROXY=http://localhost:8080
```

### Logging HTTP Requests

```go
// Add logging to HTTP client
client := &http.Client{
    Transport: &loggingRoundTripper{
        transport: http.DefaultTransport,
    },
}

type loggingRoundTripper struct {
    transport http.RoundTripper
}

func (l *loggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
    logger.Debug("HTTP Request", map[string]interface{}{
        "method": req.Method,
        "url":    req.URL.String(),
    })

    resp, err := l.transport.RoundTrip(req)

    if err == nil {
        logger.Debug("HTTP Response", map[string]interface{}{
            "status": resp.StatusCode,
            "url":    req.URL.String(),
        })
    }

    return resp, err
}
```

## Database Debugging

### SQLite Query Logging

```go
import (
    "database/sql"
    "log"
)

// Enable query logging
db, err := sql.Open("sqlite3", "file:test.db?cache=shared&mode=rwc")
db.SetMaxOpenConns(1)  // Prevent "database locked" errors

// Log queries
type loggingStmt struct {
    *sql.Stmt
    query string
}

func (s *loggingStmt) Exec(args ...interface{}) (sql.Result, error) {
    log.Printf("SQL: %s Args: %v", s.query, args)
    return s.Stmt.Exec(args...)
}
```

### Database Inspection

```bash
# Open database
sqlite3 ~/.local/share/ainative-code/ainative.db

# Show tables
.tables

# Show schema
.schema conversations

# Query data
SELECT * FROM conversations LIMIT 10;

# Explain query plan
EXPLAIN QUERY PLAN SELECT * FROM conversations WHERE user_id = ?;
```

## Troubleshooting Checklist

### Application Won't Start

- [ ] Check Go version: `go version`
- [ ] Check dependencies: `go mod verify`
- [ ] Check configuration: `cat ~/.config/ainative-code/config.yaml`
- [ ] Check logs: `ainative-code --verbose`
- [ ] Check permissions on config directory

### Tests Failing

- [ ] Run with verbose: `go test -v ./...`
- [ ] Run specific test: `go test -v -run TestName`
- [ ] Check for race conditions: `go test -race`
- [ ] Clear test cache: `go clean -testcache`
- [ ] Check test dependencies

### Performance Issues

- [ ] Profile CPU: `go test -cpuprofile`
- [ ] Profile memory: `go test -memprofile`
- [ ] Check for goroutine leaks
- [ ] Check for excessive allocations
- [ ] Review database queries

### Memory Leaks

- [ ] Check goroutine count
- [ ] Profile memory usage
- [ ] Check for unclosed resources
- [ ] Review channel usage
- [ ] Check for circular references

## Resources

- [Delve Documentation](https://github.com/go-delve/delve/tree/master/Documentation)
- [pprof Documentation](https://github.com/google/pprof)
- [Go Diagnostics](https://go.dev/doc/diagnostics)
- [Effective Go](https://golang.org/doc/effective_go.html)

---

**Last Updated**: 2025-01-05
