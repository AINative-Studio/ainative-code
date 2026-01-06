package benchmark

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/AINative-studio/ainative-code/internal/cmd"
	"github.com/AINative-studio/ainative-code/internal/config"
	"github.com/AINative-studio/ainative-code/internal/logger"
)

const (
	// Target: CLI startup time < 100ms
	CLIStartupTargetMs = 100.0
)

// BenchmarkCLIStartup measures the time to initialize the CLI
func BenchmarkCLIStartup(b *testing.B) {
	// Disable logger output for benchmarking
	logger.Init()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Measure time to create root command
		start := time.Now()
		_ = cmd.NewRootCmd()
		elapsed := time.Since(start)

		// Record timing
		if i == 0 {
			b.ReportMetric(float64(elapsed.Nanoseconds())/1_000_000, "ms/startup")
		}
	}
	b.StopTimer()

	// Report result
	result := RecordResult(b, "CLIStartup", CLIStartupTargetMs)
	if !result.PassedTarget {
		b.Logf("WARNING: CLI startup time %.2fms exceeds target of %.2fms", result.MsPerOp, CLIStartupTargetMs)
	}
}

// BenchmarkCLIStartupWithConfig measures CLI initialization with configuration loading
func BenchmarkCLIStartupWithConfig(b *testing.B) {
	helper := NewTestHelper(b)
	defer helper.Cleanup()

	// Create a test config file
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

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()

		// Initialize config
		cfg := config.New()
		cfg.SetConfigFile(configFile)
		_, _ = cfg.Load()

		elapsed := time.Since(start)

		if i == 0 {
			b.ReportMetric(float64(elapsed.Nanoseconds())/1_000_000, "ms/config-load")
		}
	}
	b.StopTimer()
}

// BenchmarkCLIStartupWithLargeConfig tests initialization with a large config file
func BenchmarkCLIStartupWithLargeConfig(b *testing.B) {
	helper := NewTestHelper(b)
	defer helper.Cleanup()

	// Create a large config file (simulating many settings)
	configDir := filepath.Join(helper.TempDir, ".ainative")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		b.Fatalf("Failed to create config dir: %v", err)
	}

	configFile := filepath.Join(configDir, "config.yaml")
	configContent := `
default_provider: anthropic
default_model: claude-3-5-sonnet-20241022
log_level: debug
log_file: /tmp/ainative.log
session_storage: sqlite
database:
  path: ~/.ainative/sessions.db
  max_connections: 10
  timeout: 30s
providers:
  anthropic:
    api_key_source: env
    timeout: 60s
    max_retries: 3
  openai:
    api_key_source: env
    timeout: 60s
    max_retries: 3
ui:
  theme: default
  color_scheme: auto
  prompt_style: modern
  show_thinking: true
`
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		b.Fatalf("Failed to write config file: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()

		cfg := config.New()
		cfg.SetConfigFile(configFile)
		_, _ = cfg.Load()

		elapsed := time.Since(start)

		if i == 0 {
			b.ReportMetric(float64(elapsed.Nanoseconds())/1_000_000, "ms/large-config")
		}
	}
	b.StopTimer()
}

// BenchmarkCommandCreation measures the time to create various commands
func BenchmarkCommandCreation(b *testing.B) {
	logger.Init()

	b.Run("RootCommand", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = cmd.NewRootCmd()
		}
	})

	b.Run("ChatCommand", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = cmd.NewChatCmd()
		}
	})

	b.Run("SessionCommand", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = cmd.NewSessionCmd()
		}
	})

	b.Run("ConfigCommand", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = cmd.NewConfigCmd()
		}
	})
}

// BenchmarkCLIVersionCheck measures the time to check CLI version
func BenchmarkCLIVersionCheck(b *testing.B) {
	logger.Init()
	rootCmd := cmd.NewRootCmd()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate version command execution
		start := time.Now()
		_ = rootCmd.Version
		elapsed := time.Since(start)

		if i == 0 {
			b.ReportMetric(float64(elapsed.Nanoseconds())/1_000, "μs/version")
		}
	}
}

// BenchmarkCLIHelpGeneration measures help text generation time
func BenchmarkCLIHelpGeneration(b *testing.B) {
	logger.Init()
	rootCmd := cmd.NewRootCmd()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()
		_ = rootCmd.UsageString()
		elapsed := time.Since(start)

		if i == 0 {
			b.ReportMetric(float64(elapsed.Nanoseconds())/1_000_000, "ms/help")
		}
	}
}

// BenchmarkCLIBinaryLoad measures the overall binary initialization time
func BenchmarkCLIBinaryLoad(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()

		// Simulate complete CLI initialization
		logger.Init()
		rootCmd := cmd.NewRootCmd()
		rootCmd.SetContext(ctx)

		elapsed := time.Since(start)

		if i == 0 {
			b.ReportMetric(float64(elapsed.Nanoseconds())/1_000_000, "ms/binary-init")

			// Check against target
			actualMs := float64(elapsed.Nanoseconds()) / 1_000_000
			if actualMs > CLIStartupTargetMs {
				b.Logf("WARNING: Binary initialization time %.2fms exceeds target of %.2fms", actualMs, CLIStartupTargetMs)
			} else {
				b.Logf("SUCCESS: Binary initialization time %.2fms meets target of %.2fms", actualMs, CLIStartupTargetMs)
			}
		}
	}
}
