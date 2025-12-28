# Debugging Guide

This guide provides debugging strategies, tools, and techniques for the AINative Code project.

## Table of Contents

- [Debugging Tools](#debugging-tools)
- [Debug Build](#debug-build)
- [Using Delve Debugger](#using-delve-debugger)
- [Logging for Debugging](#logging-for-debugging)
- [Common Issues](#common-issues)
- [Performance Debugging](#performance-debugging)
- [Memory Debugging](#memory-debugging)
- [Debugging Techniques](#debugging-techniques)

## Debugging Tools

### 1. Delve (dlv) - Go Debugger

Install Delve:
```bash
# Install latest version
go install github.com/go-delve/delve/cmd/dlv@latest

# Verify installation
dlv version
```

### 2. VSCode Go Extension

Recommended for visual debugging:
- Install "Go" extension by Google
- Built-in Delve integration
- Breakpoints, watches, call stack visualization

### 3. GoLand/IntelliJ IDEA

Professional IDE with advanced debugging:
- Visual debugger with breakpoints
- Variable inspection
- Step debugging
- Conditional breakpoints

### 4. Built-in Go Tools

```bash
# Print detailed execution trace
go build -x ./cmd/ainative-code

# Generate CPU profile
go test -cpuprofile=cpu.prof ./...

# Generate memory profile
go test -memprofile=mem.prof ./...

# Race detector
go test -race ./...
go build -race ./cmd/ainative-code
```

## Debug Build

### Build with Debug Symbols

```bash
# Build with debug info (no optimizations)
go build -gcflags="all=-N -l" -o build/ainative-code-debug ./cmd/ainative-code

# Flags explained:
# -N: Disable optimizations
# -l: Disable inlining
```

### Build for Debugging

```bash
# Debug build with additional info
go build \
  -gcflags="all=-N -l" \
  -ldflags="-X main.version=debug" \
  -o build/ainative-code-debug \
  ./cmd/ainative-code
```

### Run Debug Build

```bash
# Run with debug output
./build/ainative-code-debug --verbose

# Run with environment variables
DEBUG=1 LOG_LEVEL=debug ./build/ainative-code-debug
```

## Using Delve Debugger

### Basic Delve Commands

#### Start Debugging

```bash
# Debug main package
dlv debug ./cmd/ainative-code

# Debug with arguments
dlv debug ./cmd/ainative-code -- --config config.yaml

# Debug a test
dlv test ./internal/logger

# Debug specific test function
dlv test ./internal/logger -- -test.run TestLogLevels
```

#### Attach to Running Process

```bash
# Find process ID
ps aux | grep ainative-code

# Attach to process
dlv attach <PID>
```

### Delve Interactive Commands

```bash
# Set breakpoint
(dlv) break main.main
(dlv) break internal/logger/logger.go:42
(dlv) break Logger.Info

# List breakpoints
(dlv) breakpoints

# Clear breakpoint
(dlv) clear 1

# Continue execution
(dlv) continue
(dlv) c

# Step into function
(dlv) step
(dlv) s

# Step over function
(dlv) next
(dlv) n

# Step out of function
(dlv) stepout

# Print variable
(dlv) print myVar
(dlv) p myVar

# Print all local variables
(dlv) locals

# Print function arguments
(dlv) args

# View goroutines
(dlv) goroutines

# Switch goroutine
(dlv) goroutine 1

# View call stack
(dlv) stack
(dlv) bt

# Evaluate expression
(dlv) print len(mySlice)
(dlv) print myMap["key"]

# Set variable value
(dlv) set myVar = 42

# Exit debugger
(dlv) exit
(dlv) quit
```

### Conditional Breakpoints

```bash
# Break when condition is true
(dlv) break logger.go:42
(dlv) condition 1 level == "error"

# Break after N hits
(dlv) break logger.go:42
(dlv) condition 1 %10  # Every 10th hit
```

### VSCode Debug Configuration

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
      "args": ["--config", "configs/config.yaml"],
      "env": {
        "LOG_LEVEL": "debug"
      },
      "showLog": true
    },
    {
      "name": "Debug Test",
      "type": "go",
      "request": "launch",
      "mode": "test",
      "program": "${workspaceFolder}/internal/logger",
      "args": ["-test.v", "-test.run", "TestLogLevels"],
      "showLog": true
    },
    {
      "name": "Attach to Process",
      "type": "go",
      "request": "attach",
      "mode": "local",
      "processId": 0
    }
  ]
}
```

## Logging for Debugging

### Enable Debug Logging

```bash
# Set log level via environment
export LOG_LEVEL=debug
./build/ainative-code

# Set log level via config
ainative-code --log-level debug

# Enable verbose output
ainative-code --verbose
```

### Structured Debug Logging

```go
import "github.com/AINative-studio/ainative-code/internal/logger"

func debugFunction() {
    // Create debug logger
    log := logger.New(&logger.Config{
        Level:  logger.DebugLevel,
        Format: logger.TextFormat,
        Output: "stdout",
    })

    // Debug with context
    log.DebugWithFields("Processing request", map[string]interface{}{
        "request_id": "req-123",
        "method":     "GET",
        "path":       "/api/users",
    })

    // Debug variable state
    log.Debugf("Variable value: %+v", myStruct)
}
```

### Temporary Debug Output

```go
// Add temporary debug prints (remove before commit)
fmt.Printf("DEBUG: myVar = %+v\n", myVar)
fmt.Printf("DEBUG: Entering function %s\n", "FunctionName")

// Better: Use debug logger
logger.Debug("Entering function", map[string]interface{}{
    "function": "FunctionName",
    "args":     args,
})
```

## Common Issues

### 1. Application Crashes

**Debug with panic recovery**:

```go
defer func() {
    if r := recover(); r != nil {
        logger.Error("Panic recovered", map[string]interface{}{
            "panic": r,
            "stack": string(debug.Stack()),
        })
    }
}()
```

**Get stack trace**:

```bash
# Run with GOTRACEBACK
GOTRACEBACK=all ./build/ainative-code

# Or GOTRACEBACK=crash for core dump
GOTRACEBACK=crash ./build/ainative-code
```

### 2. Database Issues

**Enable SQL logging**:

```go
// Add logging to database queries
db.SetLogger(logger.New(&logger.Config{
    Level: logger.DebugLevel,
}))

// Log all queries
logger.DebugWithFields("Executing query", map[string]interface{}{
    "query": query,
    "args":  args,
})
```

**Debug with SQLite CLI**:

```bash
# Open database
sqlite3 ~/.local/share/ainative-code/ainative.db

# List tables
.tables

# Show schema
.schema users

# Run query
SELECT * FROM users;

# Enable column headers
.headers on
.mode column
SELECT * FROM users;
```

### 3. Configuration Issues

**Debug configuration loading**:

```go
// Print loaded config
logger.DebugWithFields("Configuration loaded", map[string]interface{}{
    "config": fmt.Sprintf("%+v", config),
})

// Verify specific values
logger.Debugf("API Key present: %v", config.APIKey != "")
logger.Debugf("Database path: %s", config.DatabasePath)
```

**Check configuration file**:

```bash
# Validate YAML syntax
yamllint configs/config.yaml

# Print parsed config
cat configs/config.yaml | yq .

# Check file permissions
ls -la configs/config.yaml
```

### 4. Race Conditions

**Detect races**:

```bash
# Build with race detector
go build -race -o build/ainative-code-race ./cmd/ainative-code

# Run with race detection
./build/ainative-code-race

# Test with race detection
go test -race ./...
```

**Fix race conditions**:

```go
// Use mutexes
type SafeCounter struct {
    mu    sync.Mutex
    count int
}

func (c *SafeCounter) Increment() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.count++
}

// Or use channels
func safeOperation() {
    done := make(chan bool)
    go func() {
        // Do work
        done <- true
    }()
    <-done
}
```

### 5. Memory Leaks

**Detect memory leaks**:

```bash
# Run with memory profiling
go test -memprofile=mem.prof ./...

# Analyze memory profile
go tool pprof mem.prof

# In pprof interactive mode:
(pprof) top10        # Show top 10 memory allocations
(pprof) list FunctionName  # Show memory allocations in function
(pprof) web          # Visualize as graph (requires Graphviz)
```

## Performance Debugging

### CPU Profiling

```bash
# Generate CPU profile from test
go test -cpuprofile=cpu.prof -bench=. ./...

# Analyze CPU profile
go tool pprof cpu.prof

# Interactive commands:
(pprof) top10       # Top 10 CPU consumers
(pprof) list main.main  # Show CPU usage in function
(pprof) web         # Visualize as graph
(pprof) pdf > cpu.pdf   # Export as PDF
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
        logger.Info("Starting pprof server on :6060")
        http.ListenAndServe("localhost:6060", nil)
    }()

    // Your application code
}
```

Access profiles:
```bash
# CPU profile (30 seconds)
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Heap profile
go tool pprof http://localhost:6060/debug/pprof/heap

# Goroutine profile
go tool pprof http://localhost:6060/debug/pprof/goroutine

# View in browser
open http://localhost:6060/debug/pprof/
```

### Execution Tracing

```bash
# Generate trace
go test -trace=trace.out ./...

# Analyze trace
go tool trace trace.out

# Opens browser with interactive trace viewer
```

### Benchmarking for Performance

```bash
# Run benchmarks
go test -bench=. -benchmem ./internal/logger

# Compare benchmarks
go test -bench=. -benchmem ./... > old.txt
# Make changes...
go test -bench=. -benchmem ./... > new.txt

# Compare results
benchstat old.txt new.txt
```

## Memory Debugging

### Find Memory Leaks

```go
// Add memory stats logging
import "runtime"

func logMemStats() {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)

    logger.InfoWithFields("Memory stats", map[string]interface{}{
        "alloc_mb":      m.Alloc / 1024 / 1024,
        "total_alloc_mb": m.TotalAlloc / 1024 / 1024,
        "sys_mb":        m.Sys / 1024 / 1024,
        "num_gc":        m.NumGC,
        "goroutines":    runtime.NumGoroutine(),
    })
}

// Call periodically
ticker := time.NewTicker(10 * time.Second)
go func() {
    for range ticker.C {
        logMemStats()
    }
}()
```

### Heap Dump Analysis

```bash
# Force garbage collection and dump heap
curl http://localhost:6060/debug/pprof/heap > heap.prof

# Analyze heap dump
go tool pprof heap.prof

# Find allocations
(pprof) top10
(pprof) list FunctionName
```

### Check for Goroutine Leaks

```bash
# Get goroutine dump
curl http://localhost:6060/debug/pprof/goroutine > goroutine.prof

# Analyze
go tool pprof goroutine.prof

# Or view as text
curl http://localhost:6060/debug/pprof/goroutine?debug=2
```

## Debugging Techniques

### 1. Printf Debugging

```go
// Strategic print statements
func debugFlow() {
    fmt.Println("1. Starting function")

    result, err := operation()
    fmt.Printf("2. Operation result: %+v, err: %v\n", result, err)

    if err != nil {
        fmt.Println("3. Error path")
        return
    }

    fmt.Println("4. Success path")
}
```

### 2. Binary Search Debugging

Comment out half the code to find the problematic section:

```go
func problematicFunction() {
    // Step 1
    // step1()  // Comment out

    // Step 2
    step2()

    // Step 3
    // step3()  // Comment out

    // Narrow down which step causes issue
}
```

### 3. Rubber Duck Debugging

Explain the code to someone (or something) line by line. Often reveals the issue.

### 4. Git Bisect

Find which commit introduced a bug:

```bash
# Start bisect
git bisect start

# Mark current commit as bad
git bisect bad

# Mark last known good commit
git bisect good v1.0.0

# Git will check out commits for you to test
# After each test:
git bisect good  # or git bisect bad

# When found, reset
git bisect reset
```

### 5. Minimal Reproduction

Create minimal test case:

```go
func TestMinimalRepro(t *testing.T) {
    // Simplest code that reproduces the bug
    result := BuggyFunction("input")

    if result != expected {
        t.Errorf("Bug reproduced: got %v, want %v", result, expected)
    }
}
```

## IDE-Specific Debugging

### VSCode

**Set breakpoints**:
- Click left of line number
- F9 on current line

**Debug actions**:
- F5: Start debugging
- F10: Step over
- F11: Step into
- Shift+F11: Step out
- F5: Continue

**Debug console**:
- Evaluate expressions
- Call functions
- Inspect variables

### GoLand

**Breakpoints**:
- Click gutter or Ctrl+F8
- Right-click breakpoint for conditions

**Debug toolbar**:
- Step Over (F8)
- Step Into (F7)
- Step Out (Shift+F8)
- Run to Cursor (Alt+F9)

**Evaluate Expression**: Alt+F8

## Quick Reference

### Debug Commands

```bash
# Build for debugging
go build -gcflags="all=-N -l" -o debug ./cmd/ainative-code

# Run with debugger
dlv debug ./cmd/ainative-code

# Debug test
dlv test ./internal/logger

# Attach to process
dlv attach <PID>

# Enable debug logging
LOG_LEVEL=debug ./build/ainative-code

# Race detection
go test -race ./...

# CPU profiling
go test -cpuprofile=cpu.prof ./...
go tool pprof cpu.prof

# Memory profiling
go test -memprofile=mem.prof ./...
go tool pprof mem.prof

# Execution trace
go test -trace=trace.out ./...
go tool trace trace.out
```

### Common dlv Commands

```bash
break <location>     # Set breakpoint
continue            # Continue execution
next                # Step over
step                # Step into
stepout             # Step out
print <var>         # Print variable
locals              # Show local variables
stack               # Show call stack
goroutines          # List goroutines
clear <n>           # Clear breakpoint
quit                # Exit debugger
```

### Logging Levels

```bash
# Set via environment
export LOG_LEVEL=debug

# Levels (lowest to highest):
# - debug: Detailed debugging info
# - info: General informational messages
# - warn: Warning messages
# - error: Error messages
# - fatal: Fatal errors (exits)
```

## Debugging Checklist

When debugging an issue:

- [ ] Can you reproduce the issue consistently?
- [ ] What changed recently? (code, config, environment)
- [ ] Check logs for errors or warnings
- [ ] Verify configuration is correct
- [ ] Test with minimal example
- [ ] Use debugger to inspect state
- [ ] Check for race conditions (go test -race)
- [ ] Profile for performance issues
- [ ] Review recent commits (git log)
- [ ] Search for similar issues (GitHub, Stack Overflow)
- [ ] Ask for help with clear reproduction steps

## Additional Resources

- [Delve Documentation](https://github.com/go-delve/delve/tree/master/Documentation)
- [Go pprof Documentation](https://pkg.go.dev/runtime/pprof)
- [Effective Go - Debugging](https://go.dev/doc/effective_go)
- [Dave Cheney's Blog](https://dave.cheney.net/) - Excellent Go debugging articles

---

**Next**: [Code Style Guidelines](code-style.md) | [Testing Guide](testing.md) | [Build Instructions](build.md)
