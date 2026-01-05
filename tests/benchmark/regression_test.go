package benchmark

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// TestRegressionDetection verifies regression detection works correctly
func TestRegressionDetection(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	// Create baseline
	baseline := &Baseline{
		Results: map[string]BenchmarkResult{
			"BenchmarkFast": {
				Name:    "BenchmarkFast",
				NsPerOp: 1000000, // 1ms
				MsPerOp: 1.0,
			},
			"BenchmarkSlow": {
				Name:    "BenchmarkSlow",
				NsPerOp: 10000000, // 10ms
				MsPerOp: 10.0,
			},
		},
	}

	// Save baseline
	baselinePath := filepath.Join(helper.TempDir, "baseline.json")
	if err := SaveBaseline([]BenchmarkResult{
		baseline.Results["BenchmarkFast"],
		baseline.Results["BenchmarkSlow"],
	}, baselinePath); err != nil {
		t.Fatalf("Failed to save baseline: %v", err)
	}

	// Load baseline
	loaded, err := LoadBaseline(baselinePath)
	if err != nil {
		t.Fatalf("Failed to load baseline: %v", err)
	}

	if loaded == nil {
		t.Fatal("Loaded baseline is nil")
	}

	// Test: No regression (within threshold)
	currentResults := []BenchmarkResult{
		{
			Name:    "BenchmarkFast",
			NsPerOp: 1050000, // 1.05ms (5% slower - within 10% threshold)
			MsPerOp: 1.05,
		},
	}

	regressions, err := CompareWithBaseline(currentResults, loaded, 10.0)
	if err != nil {
		t.Fatalf("Comparison failed: %v", err)
	}

	if len(regressions) != 0 {
		t.Errorf("Expected no regressions, got %d", len(regressions))
	}

	// Test: Regression detected (exceeds threshold)
	currentResults = []BenchmarkResult{
		{
			Name:    "BenchmarkFast",
			NsPerOp: 1200000, // 1.2ms (20% slower - exceeds 10% threshold)
			MsPerOp: 1.2,
		},
	}

	regressions, err = CompareWithBaseline(currentResults, loaded, 10.0)
	if err != nil {
		t.Fatalf("Comparison failed: %v", err)
	}

	if len(regressions) != 1 {
		t.Errorf("Expected 1 regression, got %d", len(regressions))
	}

	if len(regressions) > 0 {
		t.Logf("Detected regression: %s", regressions[0])
	}
}

// TestBaselinePersistence verifies baseline can be saved and loaded
func TestBaselinePersistence(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	results := []BenchmarkResult{
		{
			Name:        "BenchmarkTest1",
			NsPerOp:     1000000,
			MsPerOp:     1.0,
			BytesPerOp:  100,
			AllocsPerOp: 10,
		},
		{
			Name:        "BenchmarkTest2",
			NsPerOp:     2000000,
			MsPerOp:     2.0,
			BytesPerOp:  200,
			AllocsPerOp: 20,
		},
	}

	baselinePath := filepath.Join(helper.TempDir, "baseline.json")

	// Save baseline
	if err := SaveBaseline(results, baselinePath); err != nil {
		t.Fatalf("Failed to save baseline: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(baselinePath); os.IsNotExist(err) {
		t.Fatal("Baseline file was not created")
	}

	// Load baseline
	loaded, err := LoadBaseline(baselinePath)
	if err != nil {
		t.Fatalf("Failed to load baseline: %v", err)
	}

	// Verify contents
	if len(loaded.Results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(loaded.Results))
	}

	for _, result := range results {
		loadedResult, ok := loaded.Results[result.Name]
		if !ok {
			t.Errorf("Result %s not found in loaded baseline", result.Name)
			continue
		}

		if loadedResult.NsPerOp != result.NsPerOp {
			t.Errorf("NsPerOp mismatch for %s: expected %d, got %d",
				result.Name, result.NsPerOp, loadedResult.NsPerOp)
		}
	}
}

// TestNonexistentBaseline verifies handling of missing baseline
func TestNonexistentBaseline(t *testing.T) {
	baseline, err := LoadBaseline("/nonexistent/path/baseline.json")
	if err != nil {
		t.Fatalf("Expected no error for nonexistent baseline, got: %v", err)
	}

	if baseline != nil {
		t.Error("Expected nil baseline for nonexistent file")
	}
}

// BenchmarkRegressionCheck measures the overhead of regression detection
func BenchmarkRegressionCheck(b *testing.B) {
	helper := NewTestHelper(b)
	defer helper.Cleanup()

	// Create baseline with 100 benchmarks
	baselineResults := make([]BenchmarkResult, 100)
	for i := 0; i < 100; i++ {
		baselineResults[i] = BenchmarkResult{
			Name:    fmt.Sprintf("Benchmark%d", i),
			NsPerOp: int64((i + 1) * 1000000),
			MsPerOp: float64(i + 1),
		}
	}

	baselinePath := filepath.Join(helper.TempDir, "baseline.json")
	if err := SaveBaseline(baselineResults, baselinePath); err != nil {
		b.Fatalf("Failed to save baseline: %v", err)
	}

	baseline, err := LoadBaseline(baselinePath)
	if err != nil {
		b.Fatalf("Failed to load baseline: %v", err)
	}

	// Create current results (slightly slower)
	currentResults := make([]BenchmarkResult, 100)
	for i := 0; i < 100; i++ {
		currentResults[i] = BenchmarkResult{
			Name:    fmt.Sprintf("Benchmark%d", i),
			NsPerOp: int64((i+1)*1000000) + 50000, // 50μs slower
			MsPerOp: float64(i+1) + 0.05,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := CompareWithBaseline(currentResults, baseline, 10.0)
		if err != nil {
			b.Fatalf("Comparison failed: %v", err)
		}
	}
}

// TestReportGeneration verifies benchmark report generation
func TestReportGeneration(t *testing.T) {
	results := []BenchmarkResult{
		{
			Name:         "BenchmarkCLIStartup",
			NsPerOp:      50000000,
			MsPerOp:      50.0,
			PassedTarget: true,
			TargetValue:  100.0,
			ActualValue:  50.0,
		},
		{
			Name:         "BenchmarkMemoryIdle",
			NsPerOp:      0,
			MsPerOp:      0,
			MemAllocMB:   75.5,
			PassedTarget: true,
			TargetValue:  100.0,
			ActualValue:  75.5,
		},
		{
			Name:         "BenchmarkStreaming",
			NsPerOp:      30000000,
			MsPerOp:      30.0,
			PassedTarget: true,
			TargetValue:  50.0,
			ActualValue:  30.0,
		},
	}

	report := GenerateReport(results, nil)

	if report == nil {
		t.Fatal("Generated report is nil")
	}

	if len(report.Results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(report.Results))
	}

	if report.Summary.TotalBenchmarks != 3 {
		t.Errorf("Expected 3 total benchmarks, got %d", report.Summary.TotalBenchmarks)
	}

	if report.Summary.PassedTargets != 3 {
		t.Errorf("Expected 3 passed targets, got %d", report.Summary.PassedTargets)
	}

	if report.Summary.FailedTargets != 0 {
		t.Errorf("Expected 0 failed targets, got %d", report.Summary.FailedTargets)
	}

	t.Logf("Report summary: %d total, %d passed, %d failed",
		report.Summary.TotalBenchmarks,
		report.Summary.PassedTargets,
		report.Summary.FailedTargets)
}

// TestReportPersistence verifies report can be saved
func TestReportPersistence(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	results := []BenchmarkResult{
		{
			Name:         "BenchmarkTest",
			NsPerOp:      1000000,
			MsPerOp:      1.0,
			PassedTarget: true,
		},
	}

	report := GenerateReport(results, nil)
	reportPath := filepath.Join(helper.TempDir, "report.json")

	if err := SaveReport(report, reportPath); err != nil {
		t.Fatalf("Failed to save report: %v", err)
	}

	// Verify file exists and is valid JSON
	data, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("Failed to read report file: %v", err)
	}

	var loaded BenchmarkReport
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("Failed to unmarshal report: %v", err)
	}

	if len(loaded.Results) != 1 {
		t.Errorf("Expected 1 result in loaded report, got %d", len(loaded.Results))
	}
}

// TestTargetValidation verifies NFR target validation
func TestTargetValidation(t *testing.T) {
	tests := []struct {
		name         string
		actualMs     float64
		targetMs     float64
		shouldPass   bool
	}{
		{
			name:       "Well under target",
			actualMs:   50.0,
			targetMs:   100.0,
			shouldPass: true,
		},
		{
			name:       "Exactly at target",
			actualMs:   100.0,
			targetMs:   100.0,
			shouldPass: true,
		},
		{
			name:       "Slightly over target",
			actualMs:   101.0,
			targetMs:   100.0,
			shouldPass: false,
		},
		{
			name:       "Well over target",
			actualMs:   150.0,
			targetMs:   100.0,
			shouldPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BenchmarkResult{
				Name:         tt.name,
				MsPerOp:      tt.actualMs,
				TargetValue:  tt.targetMs,
				PassedTarget: tt.actualMs <= tt.targetMs,
			}

			if result.PassedTarget != tt.shouldPass {
				t.Errorf("Expected PassedTarget=%v, got %v (actual=%.2f, target=%.2f)",
					tt.shouldPass, result.PassedTarget, tt.actualMs, tt.targetMs)
			}
		})
	}
}
