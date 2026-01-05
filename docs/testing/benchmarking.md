# Performance Benchmarking Guide

## Overview

This document describes the performance benchmarking suite for AINative-Code, including how to run benchmarks, interpret results, and maintain performance standards.

## Non-Functional Requirements (NFR) Targets

The following performance targets have been established:

| Metric | Target | Category |
|--------|--------|----------|
| CLI Startup Time | < 100ms | Responsiveness |
| Memory Usage at Idle | < 100MB | Resource Usage |
| Streaming Latency (First Token) | < 50ms | Real-time Performance |
| Database Session Query | < 10ms | Data Access |
| Token Extraction | < 100ms | Code Generation |

## Benchmark Suite Structure

```
tests/benchmark/
├── cli_bench_test.go          # CLI startup benchmarks
├── memory_bench_test.go       # Memory usage benchmarks
├── streaming_bench_test.go    # Streaming latency benchmarks
├── database_bench_test.go     # Database performance
├── token_bench_test.go        # Token resolution benchmarks
├── regression_test.go         # Regression detection tests
├── helpers.go                 # Benchmark utilities
├── baseline.json              # Baseline measurements
└── reports/                   # Performance reports
    └── benchmark-output.txt
```

## Running Benchmarks

### Quick Start

Run all benchmarks:
```bash
make test-benchmark
```

Run benchmarks with verbose output:
```bash
make test-benchmark-verbose
```

### Establishing a Baseline

On first run or when establishing new performance standards:
```bash
make test-benchmark-baseline
```

This will:
1. Run all benchmarks with extended duration
2. Save results to `tests/benchmark/baseline.json`
3. Generate performance reports

### Regression Detection

Compare current performance against baseline:
```bash
make test-benchmark-compare
```

This will fail CI if any benchmark regresses by more than 10%.

## Individual Benchmark Categories

### 1. CLI Startup Benchmarks

**File**: `cli_bench_test.go`

**What it measures**:
- Binary initialization time
- Command creation overhead
- Configuration loading time
- Help text generation

**Key benchmarks**:
```bash
# Run only CLI benchmarks
go test -bench=BenchmarkCLI -benchmem ./tests/benchmark/
```

**Expected results**:
- `BenchmarkCLIStartup`: Should complete in < 100ms
- `BenchmarkCLIBinaryLoad`: Full initialization < 100ms

### 2. Memory Usage Benchmarks

**File**: `memory_bench_test.go`

**What it measures**:
- Memory allocation at startup
- Idle state memory consumption
- Memory growth over time
- Garbage collection impact

**Key benchmarks**:
```bash
# Run only memory benchmarks
go test -bench=BenchmarkMemory -benchmem ./tests/benchmark/
```

**Expected results**:
- `BenchmarkMemoryAtStartup`: < 100MB allocated
- `BenchmarkMemoryAtIdle`: < 100MB sustained
- `BenchmarkMemoryLeakDetection`: < 10MB growth per 100 iterations

### 3. Streaming Latency Benchmarks

**File**: `streaming_bench_test.go`

**What it measures**:
- Time to first token
- Overall streaming response time
- Channel operation overhead
- Concurrent stream handling

**Key benchmarks**:
```bash
# Run only streaming benchmarks
go test -bench=BenchmarkStreaming -benchmem ./tests/benchmark/
```

**Expected results**:
- `BenchmarkStreamingTimeToFirstToken`: < 50ms
- Channel overhead: < 1μs per operation

### 4. Database Performance Benchmarks

**File**: `database_bench_test.go`

**What it measures**:
- Connection establishment time
- Session query performance
- Message insertion/retrieval
- Transaction overhead
- Concurrent access patterns

**Key benchmarks**:
```bash
# Run only database benchmarks
go test -bench=BenchmarkDatabase -benchmem ./tests/benchmark/
```

**Expected results**:
- `BenchmarkSessionQueries/GetSession`: < 1ms
- `BenchmarkSessionQueries/ListSessions`: < 10ms
- `BenchmarkMessageInsertion/SingleMessage`: < 100μs

### 5. Token Resolution Benchmarks

**File**: `token_bench_test.go`

**What it measures**:
- Design token extraction speed
- Parser performance (CSS/SCSS/LESS)
- Token validation overhead
- Code generation performance

**Key benchmarks**:
```bash
# Run only token benchmarks
go test -bench=BenchmarkToken -benchmem ./tests/benchmark/
```

**Expected results**:
- `BenchmarkTokenExtraction`: < 100ms for typical file
- `BenchmarkTokenParsing/CSS`: < 10μs per token
- `BenchmarkCodeGeneration`: < 50ms

## Understanding Benchmark Output

### Standard Go Benchmark Format

```
BenchmarkCLIStartup-8           100      52.3 ms/op      15.2 MB/alloc    1250 allocs/op
```

Reading this output:
- `BenchmarkCLIStartup-8`: Benchmark name with parallelism (8 CPUs)
- `100`: Number of iterations run
- `52.3 ms/op`: Average time per operation
- `15.2 MB/alloc`: Memory allocated per operation
- `1250 allocs/op`: Number of allocations per operation

### Custom Metrics

Our benchmarks also report custom metrics:

```
BenchmarkCLIStartup-8    100    52.3 ms/op    50.0 ms/startup    15.2 MB/alloc
```

Additional metrics:
- `ms/startup`: Actual startup time measurement
- `MB/alloc`: Memory allocation in megabytes
- `μs/get`: Microseconds for get operations

## Baseline Management

### Baseline File Format

The `baseline.json` file contains:

```json
{
  "timestamp": "2025-01-04T12:00:00Z",
  "results": {
    "BenchmarkCLIStartup": {
      "name": "BenchmarkCLIStartup",
      "ns_per_op": 50000000,
      "ms_per_op": 50.0,
      "passed_target": true,
      "target_value": 100.0
    }
  }
}
```

### When to Update Baseline

Update the baseline when:
1. Legitimate performance improvements are made
2. Hardware or Go version changes significantly
3. Major refactoring is complete
4. NFR targets are officially adjusted

**Never** update the baseline to hide regressions.

## CI Integration

### GitHub Actions Workflow

The benchmark suite integrates with CI/CD:

```yaml
- name: Run Performance Benchmarks
  run: make test-benchmark-compare

- name: Upload Benchmark Results
  uses: actions/upload-artifact@v3
  with:
    name: benchmark-results
    path: tests/benchmark/reports/
```

### Regression Detection

Regressions are detected when:
- Current performance is >10% slower than baseline
- Any benchmark exceeds its NFR target
- Memory usage grows beyond acceptable limits

CI will fail if regressions are detected.

## Performance Optimization Tips

### If CLI Startup is Slow

1. Profile import costs: `go tool pprof -top cpu.prof`
2. Reduce global initialization
3. Lazy-load heavy dependencies
4. Use `init()` functions sparingly

### If Memory Usage is High

1. Check for memory leaks with `BenchmarkMemoryLeakDetection`
2. Review object pooling opportunities
3. Reduce retained references
4. Use `runtime.GC()` strategically

### If Streaming is Slow

1. Optimize channel buffer sizes
2. Reduce allocation in hot paths
3. Profile goroutine overhead
4. Consider worker pool patterns

### If Database Queries are Slow

1. Add appropriate indexes
2. Use prepared statements
3. Batch operations when possible
4. Enable WAL mode for SQLite
5. Tune connection pool settings

## Advanced Usage

### Running Specific Benchmarks

```bash
# Run only startup benchmarks
go test -bench=BenchmarkCLIStartup -benchmem ./tests/benchmark/

# Run with longer benchmark time
go test -bench=. -benchtime=10s ./tests/benchmark/

# Run with CPU profiling
go test -bench=. -cpuprofile=cpu.prof ./tests/benchmark/

# Run with memory profiling
go test -bench=. -memprofile=mem.prof ./tests/benchmark/
```

### Analyzing Profiles

```bash
# Analyze CPU profile
go tool pprof cpu.prof

# Analyze memory profile
go tool pprof mem.prof

# Generate flamegraph
go tool pprof -http=:8080 cpu.prof
```

### Benchmark Helpers

The `helpers.go` file provides utilities:

```go
// Measure memory
m := benchmark.MeasureMemory()
fmt.Printf("Allocated: %.2f MB\n", benchmark.BytesToMB(m.Alloc))

// Record result with target
result := benchmark.RecordResult(b, "MyBenchmark", 100.0)

// Compare with baseline
regressions, _ := benchmark.CompareWithBaseline(results, baseline, 10.0)
```

## Troubleshooting

### Benchmarks Are Unstable

- Ensure machine is idle during benchmarks
- Run benchmarks multiple times
- Increase `-benchtime` for more stable results
- Disable CPU frequency scaling

### Memory Benchmarks Vary

- Force GC before measurements: `runtime.GC()`
- Add sleep to let GC settle: `time.Sleep(100*time.Millisecond)`
- Run with `GODEBUG=gctrace=1` to see GC activity

### CI Benchmarks Differ from Local

- CI machines may have different specs
- Baseline may need separate CI version
- Consider using relative comparisons only

## Best Practices

1. **Run benchmarks on consistent hardware**
2. **Establish baseline on representative system**
3. **Update baseline only for legitimate improvements**
4. **Monitor trends over time, not just absolute values**
5. **Profile before optimizing**
6. **Test optimizations with benchmarks**
7. **Document performance-critical code paths**
8. **Review benchmark results in PRs**

## Reporting Issues

If you encounter benchmark failures:

1. Check if it's a real regression or environmental
2. Run locally to reproduce
3. Profile to identify bottleneck
4. Create issue with:
   - Benchmark name and output
   - Expected vs actual performance
   - System information
   - Reproduction steps

## References

- [Go Benchmark Documentation](https://pkg.go.dev/testing#hdr-Benchmarks)
- [Writing Benchmarks in Go](https://dave.cheney.net/2013/06/30/how-to-write-benchmarks-in-go)
- [Go Performance Tips](https://github.com/dgryski/go-perfbook)
- [Profiling Go Programs](https://go.dev/blog/pprof)

## Changelog

- **2025-01-04**: Initial benchmark suite implementation (TASK-083)
  - CLI startup benchmarks
  - Memory usage benchmarks
  - Streaming latency benchmarks
  - Database performance benchmarks
  - Token resolution benchmarks
  - Baseline and regression detection
  - CI integration
