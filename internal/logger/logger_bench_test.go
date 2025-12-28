package logger

import (
	"context"
	"io"
	"path/filepath"
	"testing"

	"github.com/rs/zerolog"
)

// BenchmarkLoggerSimpleMessage benchmarks simple message logging
func BenchmarkLoggerSimpleMessage(b *testing.B) {
	config := &Config{
		Level:  InfoLevel,
		Format: JSONFormat,
		Output: filepath.Join(b.TempDir(), "bench.log"),
	}

	logger, err := New(config)
	if err != nil {
		b.Fatalf("Failed to create logger: %v", err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("benchmark message")
		}
	})
}

// BenchmarkLoggerFormattedMessage benchmarks formatted message logging
func BenchmarkLoggerFormattedMessage(b *testing.B) {
	config := &Config{
		Level:  InfoLevel,
		Format: JSONFormat,
		Output: filepath.Join(b.TempDir(), "bench.log"),
	}

	logger, err := New(config)
	if err != nil {
		b.Fatalf("Failed to create logger: %v", err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Infof("benchmark message with %d args and %s strings", 42, "formatted")
		}
	})
}

// BenchmarkLoggerStructuredFields benchmarks structured field logging
func BenchmarkLoggerStructuredFields(b *testing.B) {
	config := &Config{
		Level:  InfoLevel,
		Format: JSONFormat,
		Output: filepath.Join(b.TempDir(), "bench.log"),
	}

	logger, err := New(config)
	if err != nil {
		b.Fatalf("Failed to create logger: %v", err)
	}

	fields := map[string]interface{}{
		"user_id":    "user123",
		"session_id": "session456",
		"count":      42,
		"active":     true,
		"timestamp":  "2025-01-01T00:00:00Z",
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.InfoWithFields("benchmark message", fields)
		}
	})
}

// BenchmarkLoggerContextAware benchmarks context-aware logging
func BenchmarkLoggerContextAware(b *testing.B) {
	config := &Config{
		Level:  InfoLevel,
		Format: JSONFormat,
		Output: filepath.Join(b.TempDir(), "bench.log"),
	}

	logger, err := New(config)
	if err != nil {
		b.Fatalf("Failed to create logger: %v", err)
	}

	ctx := context.Background()
	ctx = WithRequestID(ctx, "req-123")
	ctx = WithSessionID(ctx, "sess-456")
	ctx = WithUserID(ctx, "user-789")

	ctxLogger := logger.WithContext(ctx)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ctxLogger.Info("benchmark message")
		}
	})
}

// BenchmarkLoggerDisabledLevel benchmarks logging with disabled level (should be very fast)
func BenchmarkLoggerDisabledLevel(b *testing.B) {
	config := &Config{
		Level:  ErrorLevel, // Set to error, so debug won't log
		Format: JSONFormat,
		Output: filepath.Join(b.TempDir(), "bench.log"),
	}

	logger, err := New(config)
	if err != nil {
		b.Fatalf("Failed to create logger: %v", err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Debug("this should not be logged")
		}
	})
}

// BenchmarkLoggerJSONFormat benchmarks JSON output format
func BenchmarkLoggerJSONFormat(b *testing.B) {
	config := &Config{
		Level:  InfoLevel,
		Format: JSONFormat,
		Output: filepath.Join(b.TempDir(), "bench.log"),
	}

	logger, err := New(config)
	if err != nil {
		b.Fatalf("Failed to create logger: %v", err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("benchmark message")
		}
	})
}

// BenchmarkLoggerTextFormat benchmarks text output format
func BenchmarkLoggerTextFormat(b *testing.B) {
	config := &Config{
		Level:  InfoLevel,
		Format: TextFormat,
		Output: filepath.Join(b.TempDir(), "bench.log"),
	}

	logger, err := New(config)
	if err != nil {
		b.Fatalf("Failed to create logger: %v", err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("benchmark message")
		}
	})
}

// BenchmarkLoggerWithCaller benchmarks logging with caller information
func BenchmarkLoggerWithCaller(b *testing.B) {
	config := &Config{
		Level:        InfoLevel,
		Format:       JSONFormat,
		Output:       filepath.Join(b.TempDir(), "bench.log"),
		EnableCaller: true,
	}

	logger, err := New(config)
	if err != nil {
		b.Fatalf("Failed to create logger: %v", err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("benchmark message")
		}
	})
}

// BenchmarkLoggerWithStackTrace benchmarks error logging with stack traces
func BenchmarkLoggerWithStackTrace(b *testing.B) {
	config := &Config{
		Level:            ErrorLevel,
		Format:           JSONFormat,
		Output:           filepath.Join(b.TempDir(), "bench.log"),
		EnableStackTrace: true,
	}

	logger, err := New(config)
	if err != nil {
		b.Fatalf("Failed to create logger: %v", err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Error("benchmark error message")
		}
	})
}

// BenchmarkLoggerWithRotation benchmarks logging with rotation enabled
func BenchmarkLoggerWithRotation(b *testing.B) {
	config := &Config{
		Level:          InfoLevel,
		Format:         JSONFormat,
		Output:         filepath.Join(b.TempDir(), "bench.log"),
		EnableRotation: true,
		MaxSize:        100,
		MaxBackups:     3,
		MaxAge:         28,
		Compress:       false,
	}

	logger, err := New(config)
	if err != nil {
		b.Fatalf("Failed to create logger: %v", err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("benchmark message")
		}
	})
}

// BenchmarkLoggerDiscardOutput benchmarks pure overhead (no I/O)
func BenchmarkLoggerDiscardOutput(b *testing.B) {
	// Create logger with discard output for pure overhead measurement
	logger := &Logger{
		logger: zerolog.New(io.Discard).With().Timestamp().Logger(),
		config: &Config{
			Level:  InfoLevel,
			Format: JSONFormat,
		},
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("benchmark message")
		}
	})
}

// BenchmarkContextOperations benchmarks context operations
func BenchmarkContextOperations(b *testing.B) {
	ctx := context.Background()

	b.Run("WithRequestID", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = WithRequestID(ctx, "req-123")
		}
	})

	b.Run("GetRequestID", func(b *testing.B) {
		ctx = WithRequestID(ctx, "req-123")
		for i := 0; i < b.N; i++ {
			_, _ = GetRequestID(ctx)
		}
	})

	b.Run("WithAllIDs", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ctx := WithRequestID(ctx, "req-123")
			ctx = WithSessionID(ctx, "sess-456")
			_ = WithUserID(ctx, "user-789")
		}
	})
}

// BenchmarkLoggerComparison compares different logging scenarios
func BenchmarkLoggerComparison(b *testing.B) {
	benchmarks := []struct {
		name   string
		config *Config
		logFn  func(*Logger)
	}{
		{
			name: "Minimal-JSON",
			config: &Config{
				Level:  InfoLevel,
				Format: JSONFormat,
				Output: filepath.Join(b.TempDir(), "minimal.log"),
			},
			logFn: func(l *Logger) { l.Info("msg") },
		},
		{
			name: "Structured-5Fields",
			config: &Config{
				Level:  InfoLevel,
				Format: JSONFormat,
				Output: filepath.Join(b.TempDir(), "structured.log"),
			},
			logFn: func(l *Logger) {
				l.InfoWithFields("msg", map[string]interface{}{
					"f1": "v1", "f2": "v2", "f3": "v3", "f4": "v4", "f5": "v5",
				})
			},
		},
		{
			name: "Context-3IDs",
			config: &Config{
				Level:  InfoLevel,
				Format: JSONFormat,
				Output: filepath.Join(b.TempDir(), "context.log"),
			},
			logFn: func(l *Logger) {
				ctx := WithRequestID(context.Background(), "req")
				ctx = WithSessionID(ctx, "sess")
				ctx = WithUserID(ctx, "user")
				l.WithContext(ctx).Info("msg")
			},
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			logger, err := New(bm.config)
			if err != nil {
				b.Fatalf("Failed to create logger: %v", err)
			}

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					bm.logFn(logger)
				}
			})
		})
	}
}

// BenchmarkMemoryAllocation benchmarks memory allocations
func BenchmarkMemoryAllocation(b *testing.B) {
	config := &Config{
		Level:  InfoLevel,
		Format: JSONFormat,
		Output: filepath.Join(b.TempDir(), "mem.log"),
	}

	logger, err := New(config)
	if err != nil {
		b.Fatalf("Failed to create logger: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		logger.InfoWithFields("message", map[string]interface{}{
			"user_id":    "user123",
			"session_id": "session456",
			"count":      42,
		})
	}
}
