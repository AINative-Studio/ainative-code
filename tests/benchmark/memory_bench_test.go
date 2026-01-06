package benchmark

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/AINative-studio/ainative-code/internal/cmd"
	"github.com/AINative-studio/ainative-code/internal/config"
	"github.com/AINative-studio/ainative-code/internal/database"
	"github.com/AINative-studio/ainative-code/internal/logger"
)

const (
	// Target: Memory usage at idle < 100MB
	MemoryIdleTargetMB = 100.0
)

// BenchmarkMemoryAtStartup measures memory usage immediately after CLI initialization
func BenchmarkMemoryAtStartup(b *testing.B) {
	logger.Init()

	// Force garbage collection before measurement
	runtime.GC()
	time.Sleep(100 * time.Millisecond)

	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Initialize CLI
		rootCmd := cmd.NewRootCmd()
		_ = rootCmd

		// Force garbage collection
		runtime.GC()
		time.Sleep(50 * time.Millisecond)

		// Measure memory
		var m2 runtime.MemStats
		runtime.ReadMemStats(&m2)

		allocMB := BytesToMB(m2.Alloc - m1.Alloc)

		if i == 0 {
			b.ReportMetric(allocMB, "MB/alloc")

			if allocMB > MemoryIdleTargetMB {
				b.Logf("WARNING: Startup memory %.2fMB exceeds target of %.2fMB", allocMB, MemoryIdleTargetMB)
			} else {
				b.Logf("SUCCESS: Startup memory %.2fMB meets target of %.2fMB", allocMB, MemoryIdleTargetMB)
			}
		}
	}
}

// BenchmarkMemoryAtIdle measures memory usage during idle state
func BenchmarkMemoryAtIdle(b *testing.B) {
	logger.Init()

	// Initialize CLI
	rootCmd := cmd.NewRootCmd()
	_ = rootCmd

	// Force garbage collection
	runtime.GC()
	time.Sleep(100 * time.Millisecond)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		allocMB := BytesToMB(m.Alloc)
		heapMB := BytesToMB(m.HeapAlloc)
		sysMB := BytesToMB(m.Sys)

		if i == 0 {
			b.ReportMetric(allocMB, "MB/alloc")
			b.ReportMetric(heapMB, "MB/heap")
			b.ReportMetric(sysMB, "MB/sys")

			b.Logf("Memory at idle - Alloc: %.2fMB, Heap: %.2fMB, Sys: %.2fMB", allocMB, heapMB, sysMB)

			if allocMB > MemoryIdleTargetMB {
				b.Logf("WARNING: Idle memory %.2fMB exceeds target of %.2fMB", allocMB, MemoryIdleTargetMB)
			}
		}

		// Sleep to simulate idle state
		time.Sleep(100 * time.Millisecond)
	}
}

// BenchmarkMemoryGrowthOverTime measures memory growth during prolonged operation
func BenchmarkMemoryGrowthOverTime(b *testing.B) {
	logger.Init()
	rootCmd := cmd.NewRootCmd()
	_ = rootCmd

	// Initial measurement
	runtime.GC()
	time.Sleep(100 * time.Millisecond)

	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)
	initialMB := BytesToMB(m1.Alloc)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate some activity
		_ = cmd.NewChatCmd()
		_ = cmd.NewSessionCmd()

		// Periodic measurement
		if i%10 == 0 {
			var m2 runtime.MemStats
			runtime.ReadMemStats(&m2)
			currentMB := BytesToMB(m2.Alloc)
			growthMB := currentMB - initialMB

			if i == 0 {
				b.ReportMetric(growthMB, "MB/growth")
				b.Logf("Memory growth after %d iterations: %.2fMB", i, growthMB)
			}
		}
	}
}

// BenchmarkMemoryWithDatabase measures memory with database operations
func BenchmarkMemoryWithDatabase(b *testing.B) {
	helper := NewTestHelper(b)
	defer helper.Cleanup()

	// Create test database
	dbPath := filepath.Join(helper.TempDir, "test.db")
	cfg := database.DefaultConfig(dbPath)

	// Force GC before test
	runtime.GC()
	time.Sleep(100 * time.Millisecond)

	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db, err := database.Initialize(cfg)
		if err != nil {
			b.Fatalf("Failed to initialize database: %v", err)
		}

		// Measure memory with database loaded
		runtime.GC()
		time.Sleep(50 * time.Millisecond)

		var m2 runtime.MemStats
		runtime.ReadMemStats(&m2)

		allocMB := BytesToMB(m2.Alloc - m1.Alloc)

		if i == 0 {
			b.ReportMetric(allocMB, "MB/with-db")
			b.Logf("Memory with database: %.2fMB", allocMB)
		}

		db.Close()
	}
}

// BenchmarkMemoryWithConfig measures memory with configuration loaded
func BenchmarkMemoryWithConfig(b *testing.B) {
	helper := NewTestHelper(b)
	defer helper.Cleanup()

	// Create config file
	configDir := filepath.Join(helper.TempDir, ".ainative")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		b.Fatalf("Failed to create config dir: %v", err)
	}

	configFile := filepath.Join(configDir, "config.yaml")
	configContent := `
default_provider: anthropic
default_model: claude-3-5-sonnet-20241022
log_level: info
`
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		b.Fatalf("Failed to write config file: %v", err)
	}

	runtime.GC()
	time.Sleep(100 * time.Millisecond)

	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cfg := config.New()
		cfg.SetConfigFile(configFile)
		if _, err := cfg.Load(); err != nil {
			b.Fatalf("Failed to load config: %v", err)
		}

		runtime.GC()
		time.Sleep(50 * time.Millisecond)

		var m2 runtime.MemStats
		runtime.ReadMemStats(&m2)

		allocMB := BytesToMB(m2.Alloc - m1.Alloc)

		if i == 0 {
			b.ReportMetric(allocMB, "MB/with-config")
			b.Logf("Memory with config: %.2fMB", allocMB)
		}
	}
}

// BenchmarkMemoryAllocations measures allocation patterns
func BenchmarkMemoryAllocations(b *testing.B) {
	logger.Init()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var m1 runtime.MemStats
		runtime.ReadMemStats(&m1)

		// Create various commands
		_ = cmd.NewRootCmd()
		_ = cmd.NewChatCmd()
		_ = cmd.NewSessionCmd()

		var m2 runtime.MemStats
		runtime.ReadMemStats(&m2)

		mallocs := m2.Mallocs - m1.Mallocs
		allocations := m2.TotalAlloc - m1.TotalAlloc

		if i == 0 {
			b.ReportMetric(float64(mallocs), "mallocs")
			b.ReportMetric(BytesToMB(allocations), "MB/total-alloc")
		}
	}
}

// BenchmarkMemoryLeakDetection runs operations repeatedly to detect memory leaks
func BenchmarkMemoryLeakDetection(b *testing.B) {
	ctx := context.Background()
	logger.Init()

	// Initial baseline
	runtime.GC()
	time.Sleep(200 * time.Millisecond)

	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)
	baseline := BytesToMB(m1.Alloc)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Perform operations
		rootCmd := cmd.NewRootCmd()
		rootCmd.SetContext(ctx)

		// Force cleanup
		runtime.GC()

		// Check memory periodically
		if i > 0 && i%100 == 0 {
			time.Sleep(100 * time.Millisecond)
			var m2 runtime.MemStats
			runtime.ReadMemStats(&m2)
			current := BytesToMB(m2.Alloc)
			growth := current - baseline

			if i == 100 {
				b.ReportMetric(growth, "MB/leak-check")
				b.Logf("Memory after %d iterations: baseline=%.2fMB, current=%.2fMB, growth=%.2fMB", i, baseline, current, growth)

				// Significant growth might indicate a leak
				if growth > 10.0 {
					b.Logf("WARNING: Possible memory leak detected - growth: %.2fMB", growth)
				}
			}
		}
	}
}

// BenchmarkMemoryGarbageCollection measures GC impact on memory
func BenchmarkMemoryGarbageCollection(b *testing.B) {
	logger.Init()
	rootCmd := cmd.NewRootCmd()
	_ = rootCmd

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()

		runtime.GC()

		elapsed := time.Since(start)

		if i == 0 {
			b.ReportMetric(float64(elapsed.Nanoseconds())/1_000_000, "ms/gc")
		}
	}
}
