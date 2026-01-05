package benchmark

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// BenchmarkResult represents a single benchmark result
type BenchmarkResult struct {
	Name          string        `json:"name"`
	Iterations    int           `json:"iterations"`
	NsPerOp       int64         `json:"ns_per_op"`
	MsPerOp       float64       `json:"ms_per_op"`
	BytesPerOp    int64         `json:"bytes_per_op"`
	AllocsPerOp   int64         `json:"allocs_per_op"`
	MemAllocMB    float64       `json:"mem_alloc_mb"`
	Timestamp     time.Time     `json:"timestamp"`
	GoVersion     string        `json:"go_version"`
	OS            string        `json:"os"`
	Arch          string        `json:"arch"`
	PassedTarget  bool          `json:"passed_target"`
	TargetValue   float64       `json:"target_value,omitempty"`
	ActualValue   float64       `json:"actual_value,omitempty"`
	TargetUnit    string        `json:"target_unit,omitempty"`
}

// BenchmarkReport contains all benchmark results
type BenchmarkReport struct {
	Timestamp   time.Time         `json:"timestamp"`
	Results     []BenchmarkResult `json:"results"`
	Summary     BenchmarkSummary  `json:"summary"`
}

// BenchmarkSummary provides high-level statistics
type BenchmarkSummary struct {
	TotalBenchmarks int     `json:"total_benchmarks"`
	PassedTargets   int     `json:"passed_targets"`
	FailedTargets   int     `json:"failed_targets"`
	AverageNsPerOp  float64 `json:"average_ns_per_op"`
	TotalAllocsMB   float64 `json:"total_allocs_mb"`
}

// Baseline represents baseline measurements for regression detection
type Baseline struct {
	Timestamp time.Time                  `json:"timestamp"`
	Results   map[string]BenchmarkResult `json:"results"`
}

// TestHelper provides common test utilities
type TestHelper struct {
	TempDir string
	Cleanup func()
}

// NewTestHelper creates a new test helper with cleanup
func NewTestHelper(t *testing.T) *TestHelper {
	tempDir, err := os.MkdirTemp("", "benchmark-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	return &TestHelper{
		TempDir: tempDir,
		Cleanup: func() {
			os.RemoveAll(tempDir)
		},
	}
}

// MeasureMemory captures current memory statistics
func MeasureMemory() runtime.MemStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m
}

// BytesToMB converts bytes to megabytes
func BytesToMB(bytes uint64) float64 {
	return float64(bytes) / 1024 / 1024
}

// NsToMs converts nanoseconds to milliseconds
func NsToMs(ns int64) float64 {
	return float64(ns) / 1_000_000
}

// RecordResult records a benchmark result with target validation
func RecordResult(b *testing.B, name string, targetMs float64) BenchmarkResult {
	actualMs := float64(b.Elapsed().Nanoseconds()) / float64(b.N) / 1_000_000

	result := BenchmarkResult{
		Name:         name,
		Iterations:   b.N,
		NsPerOp:      b.Elapsed().Nanoseconds() / int64(b.N),
		MsPerOp:      actualMs,
		BytesPerOp:   0, // Will be set if b.SetBytes is called
		AllocsPerOp:  0, // Will be populated from benchmem
		Timestamp:    time.Now(),
		GoVersion:    runtime.Version(),
		OS:           runtime.GOOS,
		Arch:         runtime.GOARCH,
		PassedTarget: false,
		TargetValue:  targetMs,
		ActualValue:  actualMs,
		TargetUnit:   "ms",
	}

	if targetMs > 0 {
		result.PassedTarget = result.MsPerOp <= targetMs
	}

	return result
}

// SaveBaseline saves benchmark results as baseline
func SaveBaseline(results []BenchmarkResult, path string) error {
	baseline := Baseline{
		Timestamp: time.Now(),
		Results:   make(map[string]BenchmarkResult),
	}

	for _, result := range results {
		baseline.Results[result.Name] = result
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create baseline directory: %w", err)
	}

	data, err := json.MarshalIndent(baseline, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal baseline: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write baseline: %w", err)
	}

	return nil
}

// LoadBaseline loads baseline measurements from file
func LoadBaseline(path string) (*Baseline, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No baseline exists yet
		}
		return nil, fmt.Errorf("failed to read baseline: %w", err)
	}

	var baseline Baseline
	if err := json.Unmarshal(data, &baseline); err != nil {
		return nil, fmt.Errorf("failed to unmarshal baseline: %w", err)
	}

	return &baseline, nil
}

// CompareWithBaseline compares current results with baseline
func CompareWithBaseline(current []BenchmarkResult, baseline *Baseline, threshold float64) ([]string, error) {
	if baseline == nil {
		return nil, nil // No baseline to compare against
	}

	var regressions []string

	for _, result := range current {
		baselineResult, ok := baseline.Results[result.Name]
		if !ok {
			continue // New benchmark, skip comparison
		}

		// Calculate percentage change
		percentChange := ((float64(result.NsPerOp) - float64(baselineResult.NsPerOp)) / float64(baselineResult.NsPerOp)) * 100

		if percentChange > threshold {
			regressions = append(regressions, fmt.Sprintf(
				"%s: %.2f%% slower (baseline: %.2fms, current: %.2fms)",
				result.Name,
				percentChange,
				baselineResult.MsPerOp,
				result.MsPerOp,
			))
		}
	}

	return regressions, nil
}

// GenerateReport creates a comprehensive benchmark report
func GenerateReport(results []BenchmarkResult, baseline *Baseline) *BenchmarkReport {
	report := &BenchmarkReport{
		Timestamp: time.Now(),
		Results:   results,
		Summary: BenchmarkSummary{
			TotalBenchmarks: len(results),
		},
	}

	var totalNs int64
	var totalAllocsMB float64

	for _, result := range results {
		totalNs += result.NsPerOp
		totalAllocsMB += result.MemAllocMB

		if result.PassedTarget {
			report.Summary.PassedTargets++
		} else {
			report.Summary.FailedTargets++
		}
	}

	if len(results) > 0 {
		report.Summary.AverageNsPerOp = float64(totalNs) / float64(len(results))
	}
	report.Summary.TotalAllocsMB = totalAllocsMB

	return report
}

// SaveReport saves the benchmark report to a file
func SaveReport(report *BenchmarkReport, path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create report directory: %w", err)
	}

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write report: %w", err)
	}

	return nil
}

// GetTestContext returns a context with reasonable timeout for benchmarks
func GetTestContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 30*time.Second)
}

// MinDuration ensures a minimum duration for timing measurements
func MinDuration(d time.Duration, min time.Duration) time.Duration {
	if d < min {
		return min
	}
	return d
}
