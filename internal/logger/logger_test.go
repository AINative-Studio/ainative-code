package logger

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name:    "default config",
			config:  DefaultConfig(),
			wantErr: false,
		},
		{
			name:    "nil config",
			config:  nil,
			wantErr: false,
		},
		{
			name: "json format",
			config: &Config{
				Level:  InfoLevel,
				Format: JSONFormat,
				Output: "stdout",
			},
			wantErr: false,
		},
		{
			name: "text format",
			config: &Config{
				Level:  DebugLevel,
				Format: TextFormat,
				Output: "stdout",
			},
			wantErr: false,
		},
		{
			name: "with caller enabled",
			config: &Config{
				Level:        InfoLevel,
				Format:       JSONFormat,
				Output:       "stdout",
				EnableCaller: true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := New(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && logger == nil {
				t.Error("New() returned nil logger")
			}
		})
	}
}

func TestLogLevels(t *testing.T) {
	tests := []struct {
		name     string
		logLevel LogLevel
		logFunc  func(*Logger)
		expected string
	}{
		{
			name:     "debug level",
			logLevel: DebugLevel,
			logFunc: func(l *Logger) {
				l.Debug("debug message")
			},
			expected: "debug message",
		},
		{
			name:     "info level",
			logLevel: InfoLevel,
			logFunc: func(l *Logger) {
				l.Info("info message")
			},
			expected: "info message",
		},
		{
			name:     "warn level",
			logLevel: WarnLevel,
			logFunc: func(l *Logger) {
				l.Warn("warn message")
			},
			expected: "warn message",
		},
		{
			name:     "error level",
			logLevel: ErrorLevel,
			logFunc: func(l *Logger) {
				l.Error("error message")
			},
			expected: "error message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary file for logging
			tmpFile := filepath.Join(t.TempDir(), "test.log")

			config := &Config{
				Level:  tt.logLevel,
				Format: JSONFormat,
				Output: tmpFile,
			}

			logger, err := New(config)
			if err != nil {
				t.Fatalf("Failed to create logger: %v", err)
			}

			// Execute the log function
			tt.logFunc(logger)

			// Read the log file
			content, err := os.ReadFile(tmpFile)
			if err != nil {
				t.Fatalf("Failed to read log file: %v", err)
			}

			// Verify the message was logged
			if !strings.Contains(string(content), tt.expected) {
				t.Errorf("Log output does not contain expected message. Got: %s", string(content))
			}
		})
	}
}

func TestFormattedLogging(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test.log")

	config := &Config{
		Level:  DebugLevel,
		Format: JSONFormat,
		Output: tmpFile,
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Test formatted logging
	logger.Debugf("debug %s", "formatted")
	logger.Infof("info %d", 123)
	logger.Warnf("warn %v", true)
	logger.Errorf("error %s", "formatted")

	// Read the log file
	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(content)

	// Verify all formatted messages
	expectedMessages := []string{
		"debug formatted",
		"info 123",
		"warn true",
		"error formatted",
	}

	for _, msg := range expectedMessages {
		if !strings.Contains(logContent, msg) {
			t.Errorf("Log output does not contain expected message: %s", msg)
		}
	}
}

func TestStructuredLogging(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test.log")

	config := &Config{
		Level:  InfoLevel,
		Format: JSONFormat,
		Output: tmpFile,
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Test structured logging
	fields := map[string]interface{}{
		"user_id":    "user123",
		"session_id": "session456",
		"count":      42,
		"active":     true,
	}

	logger.InfoWithFields("structured message", fields)

	// Read the log file
	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	// Parse JSON
	var logEntry map[string]interface{}
	if err := json.Unmarshal(content, &logEntry); err != nil {
		t.Fatalf("Failed to parse JSON log: %v", err)
	}

	// Verify fields
	if logEntry["message"] != "structured message" {
		t.Errorf("Expected message 'structured message', got '%v'", logEntry["message"])
	}

	if logEntry["user_id"] != "user123" {
		t.Errorf("Expected user_id 'user123', got '%v'", logEntry["user_id"])
	}

	if logEntry["session_id"] != "session456" {
		t.Errorf("Expected session_id 'session456', got '%v'", logEntry["session_id"])
	}

	if logEntry["count"] != float64(42) {
		t.Errorf("Expected count 42, got '%v'", logEntry["count"])
	}

	if logEntry["active"] != true {
		t.Errorf("Expected active true, got '%v'", logEntry["active"])
	}
}

func TestContextAwareLogging(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test.log")

	config := &Config{
		Level:  InfoLevel,
		Format: JSONFormat,
		Output: tmpFile,
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Create context with IDs
	ctx := context.Background()
	ctx = WithRequestID(ctx, "req-123")
	ctx = WithSessionID(ctx, "sess-456")
	ctx = WithUserID(ctx, "user-789")

	// Create context-aware logger
	ctxLogger := logger.WithContext(ctx)
	ctxLogger.Info("context message")

	// Read the log file
	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	// Parse JSON
	var logEntry map[string]interface{}
	if err := json.Unmarshal(content, &logEntry); err != nil {
		t.Fatalf("Failed to parse JSON log: %v", err)
	}

	// Verify context fields
	if logEntry["request_id"] != "req-123" {
		t.Errorf("Expected request_id 'req-123', got '%v'", logEntry["request_id"])
	}

	if logEntry["session_id"] != "sess-456" {
		t.Errorf("Expected session_id 'sess-456', got '%v'", logEntry["session_id"])
	}

	if logEntry["user_id"] != "user-789" {
		t.Errorf("Expected user_id 'user-789', got '%v'", logEntry["user_id"])
	}
}

func TestContextHelpers(t *testing.T) {
	ctx := context.Background()

	// Test WithRequestID and GetRequestID
	ctx = WithRequestID(ctx, "req-123")
	requestID, ok := GetRequestID(ctx)
	if !ok || requestID != "req-123" {
		t.Errorf("Expected request_id 'req-123', got '%s' (ok=%v)", requestID, ok)
	}

	// Test WithSessionID and GetSessionID
	ctx = WithSessionID(ctx, "sess-456")
	sessionID, ok := GetSessionID(ctx)
	if !ok || sessionID != "sess-456" {
		t.Errorf("Expected session_id 'sess-456', got '%s' (ok=%v)", sessionID, ok)
	}

	// Test WithUserID and GetUserID
	ctx = WithUserID(ctx, "user-789")
	userID, ok := GetUserID(ctx)
	if !ok || userID != "user-789" {
		t.Errorf("Expected user_id 'user-789', got '%s' (ok=%v)", userID, ok)
	}

	// Test missing values
	emptyCtx := context.Background()
	_, ok = GetRequestID(emptyCtx)
	if ok {
		t.Error("Expected GetRequestID to return false for empty context")
	}
}

func TestOutputFormats(t *testing.T) {
	tests := []struct {
		name   string
		format OutputFormat
	}{
		{
			name:   "json format",
			format: JSONFormat,
		},
		{
			name:   "text format",
			format: TextFormat,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := filepath.Join(t.TempDir(), "test.log")

			config := &Config{
				Level:  InfoLevel,
				Format: tt.format,
				Output: tmpFile,
			}

			logger, err := New(config)
			if err != nil {
				t.Fatalf("Failed to create logger: %v", err)
			}

			logger.Info("test message")

			// Verify file was created and has content
			content, err := os.ReadFile(tmpFile)
			if err != nil {
				t.Fatalf("Failed to read log file: %v", err)
			}

			if len(content) == 0 {
				t.Error("Log file is empty")
			}

			if !strings.Contains(string(content), "test message") {
				t.Errorf("Log does not contain expected message")
			}
		})
	}
}

func TestLogRotation(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	config := &Config{
		Level:          InfoLevel,
		Format:         JSONFormat,
		Output:         logFile,
		EnableRotation: true,
		MaxSize:        1, // 1 MB
		MaxBackups:     3,
		MaxAge:         1,
		Compress:       false,
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Write a log message
	logger.Info("rotation test")

	// Verify file was created
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Error("Log file was not created")
	}
}

func TestErrorWithErr(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test.log")

	config := &Config{
		Level:  ErrorLevel,
		Format: JSONFormat,
		Output: tmpFile,
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Test error logging with error object
	testErr := os.ErrNotExist
	logger.ErrorWithErr("file not found", testErr)

	// Read the log file
	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(content)

	// Verify error message
	if !strings.Contains(logContent, "file not found") {
		t.Error("Log does not contain expected message")
	}

	// Parse JSON to verify error field
	var logEntry map[string]interface{}
	if err := json.Unmarshal(content, &logEntry); err != nil {
		t.Fatalf("Failed to parse JSON log: %v", err)
	}

	if logEntry["error"] == nil {
		t.Error("Log entry does not contain error field")
	}
}

func TestEnableCaller(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test.log")

	config := &Config{
		Level:        InfoLevel,
		Format:       JSONFormat,
		Output:       tmpFile,
		EnableCaller: true,
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	logger.Info("caller test")

	// Read the log file
	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	// Parse JSON
	var logEntry map[string]interface{}
	if err := json.Unmarshal(content, &logEntry); err != nil {
		t.Fatalf("Failed to parse JSON log: %v", err)
	}

	// Verify caller field exists
	if logEntry["caller"] == nil {
		t.Error("Log entry does not contain caller field")
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Level != InfoLevel {
		t.Errorf("Expected default level InfoLevel, got %s", config.Level)
	}

	if config.Format != TextFormat {
		t.Errorf("Expected default format TextFormat, got %s", config.Format)
	}

	if config.Output != "stderr" {
		t.Errorf("Expected default output stderr, got %s", config.Output)
	}

	if config.EnableRotation {
		t.Error("Expected EnableRotation to be false by default")
	}

	if config.MaxSize != 100 {
		t.Errorf("Expected MaxSize 100, got %d", config.MaxSize)
	}

	if config.MaxBackups != 3 {
		t.Errorf("Expected MaxBackups 3, got %d", config.MaxBackups)
	}

	if config.MaxAge != 28 {
		t.Errorf("Expected MaxAge 28, got %d", config.MaxAge)
	}

	if !config.Compress {
		t.Error("Expected Compress to be true by default")
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		level    LogLevel
		wantErr  bool
		expected string
	}{
		{
			name:     "debug",
			level:    DebugLevel,
			wantErr:  false,
			expected: "debug",
		},
		{
			name:     "info",
			level:    InfoLevel,
			wantErr:  false,
			expected: "info",
		},
		{
			name:     "warn",
			level:    WarnLevel,
			wantErr:  false,
			expected: "warn",
		},
		{
			name:     "error",
			level:    ErrorLevel,
			wantErr:  false,
			expected: "error",
		},
		{
			name:     "invalid",
			level:    LogLevel("invalid"),
			wantErr:  true,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level, err := parseLogLevel(tt.level)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseLogLevel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && level.String() != tt.expected {
				t.Errorf("parseLogLevel() = %v, want %v", level.String(), tt.expected)
			}
		})
	}
}

// TestLoggerOutputsToStderr verifies that the logger writes to stderr by default
func TestLoggerOutputsToStderr(t *testing.T) {
	// Test that default config uses stderr
	config := DefaultConfig()
	if config.Output != "stderr" {
		t.Errorf("DefaultConfig should output to stderr, got %s", config.Output)
	}

	// Test that logger can be explicitly configured to use stderr
	stderrConfig := &Config{
		Level:  InfoLevel,
		Format: JSONFormat,
		Output: "stderr",
	}

	logger, err := New(stderrConfig)
	if err != nil {
		t.Fatalf("Failed to create logger with stderr output: %v", err)
	}

	if logger == nil {
		t.Fatal("Logger should not be nil")
	}

	// Verify the config is stored correctly
	if logger.config.Output != "stderr" {
		t.Errorf("Logger config should have Output=stderr, got %s", logger.config.Output)
	}
}

// TestLoggerStdoutVsStderr verifies that stdout and stderr outputs work correctly
func TestLoggerStdoutVsStderr(t *testing.T) {
	tests := []struct {
		name   string
		output string
	}{
		{
			name:   "stdout output",
			output: "stdout",
		},
		{
			name:   "stderr output",
			output: "stderr",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Level:  InfoLevel,
				Format: JSONFormat,
				Output: tt.output,
			}

			logger, err := New(config)
			if err != nil {
				t.Fatalf("Failed to create logger with %s output: %v", tt.output, err)
			}

			if logger == nil {
				t.Fatalf("Logger should not be nil for %s output", tt.output)
			}

			if logger.config.Output != tt.output {
				t.Errorf("Logger config should have Output=%s, got %s", tt.output, logger.config.Output)
			}

			// Verify logger can write without errors
			logger.Info("test message")
		})
	}
}

// Benchmark to verify performance is acceptable
func BenchmarkLogger(b *testing.B) {
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
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message")
	}
}
