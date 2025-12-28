package errors

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestDebugMode(t *testing.T) {
	// Save original state
	originalDebugMode := DebugMode
	defer func() { DebugMode = originalDebugMode }()

	t.Run("EnableDebugMode", func(t *testing.T) {
		EnableDebugMode()
		if !IsDebugMode() {
			t.Error("expected debug mode to be enabled")
		}
	})

	t.Run("DisableDebugMode", func(t *testing.T) {
		DisableDebugMode()
		if IsDebugMode() {
			t.Error("expected debug mode to be disabled")
		}
	})
}

func TestFormat(t *testing.T) {
	t.Run("Format nil error", func(t *testing.T) {
		result := Format(nil)
		if result != "" {
			t.Errorf("expected empty string for nil error, got '%s'", result)
		}
	})

	t.Run("Format standard error", func(t *testing.T) {
		err := errors.New("standard error")
		result := Format(err)

		if result != "standard error" {
			t.Errorf("expected 'standard error', got '%s'", result)
		}
	})

	t.Run("Format BaseError", func(t *testing.T) {
		err := newError(ErrCodeConfigInvalid, "invalid config", SeverityHigh, false)
		result := Format(err)

		if !strings.Contains(result, "CONFIG_INVALID") {
			t.Errorf("expected error code in output: %s", result)
		}

		if !strings.Contains(result, "invalid config") {
			t.Errorf("expected error message in output: %s", result)
		}
	})

	t.Run("Format with stack trace in debug mode", func(t *testing.T) {
		EnableDebugMode()
		defer DisableDebugMode()

		err := newError(ErrCodeConfigInvalid, "test error", SeverityHigh, false)
		result := Format(err)

		if !strings.Contains(result, "Stack trace") {
			t.Errorf("expected stack trace in debug mode: %s", result)
		}
	})

	t.Run("Format without stack trace in production mode", func(t *testing.T) {
		DisableDebugMode()

		err := newError(ErrCodeConfigInvalid, "test error", SeverityHigh, false)
		result := Format(err)

		if strings.Contains(result, "Stack trace") {
			t.Errorf("expected no stack trace in production mode: %s", result)
		}
	})

	t.Run("Format with cause", func(t *testing.T) {
		originalErr := errors.New("original error")
		wrappedErr := Wrap(originalErr, ErrCodeConfigParse, "parse failed")
		result := Format(wrappedErr)

		if !strings.Contains(result, "Caused by:") {
			t.Errorf("expected cause information: %s", result)
		}

		if !strings.Contains(result, "original error") {
			t.Errorf("expected original error in output: %s", result)
		}
	})

	t.Run("Format with metadata in debug mode", func(t *testing.T) {
		EnableDebugMode()
		defer DisableDebugMode()

		err := newError(ErrCodeConfigInvalid, "test error", SeverityHigh, false)
		err.WithMetadata("key1", "value1")
		err.WithMetadata("key2", 123)

		result := Format(err)

		if !strings.Contains(result, "Metadata") {
			t.Errorf("expected metadata section: %s", result)
		}

		if !strings.Contains(result, "key1") || !strings.Contains(result, "value1") {
			t.Errorf("expected metadata in output: %s", result)
		}
	})
}

func TestFormatUser(t *testing.T) {
	t.Run("FormatUser nil error", func(t *testing.T) {
		result := FormatUser(nil)
		if result != "" {
			t.Errorf("expected empty string for nil error, got '%s'", result)
		}
	})

	t.Run("FormatUser standard error", func(t *testing.T) {
		err := errors.New("standard error")
		result := FormatUser(err)

		if result != "An unexpected error occurred. Please try again." {
			t.Errorf("unexpected user message: %s", result)
		}
	})

	t.Run("FormatUser with custom user message", func(t *testing.T) {
		err := NewConfigMissingError("api_key")
		result := FormatUser(err)

		if !strings.Contains(result, "api_key") {
			t.Errorf("expected config key in user message: %s", result)
		}

		// Should not contain technical details like error codes
		if strings.Contains(result, "CONFIG_MISSING") {
			t.Errorf("user message should not contain error codes: %s", result)
		}
	})
}

func TestToJSON(t *testing.T) {
	t.Run("ToJSON nil error", func(t *testing.T) {
		result, err := ToJSON(nil)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if result != nil {
			t.Errorf("expected nil result for nil error")
		}
	})

	t.Run("ToJSON standard error", func(t *testing.T) {
		err := errors.New("standard error")
		result, jsonErr := ToJSON(err)

		if jsonErr != nil {
			t.Fatalf("unexpected error: %v", jsonErr)
		}

		var response ErrorResponse
		if err := json.Unmarshal(result, &response); err != nil {
			t.Fatalf("failed to unmarshal JSON: %v", err)
		}

		if response.Code != "UNKNOWN_ERROR" {
			t.Errorf("expected code UNKNOWN_ERROR, got %s", response.Code)
		}

		if response.Message != "standard error" {
			t.Errorf("expected message 'standard error', got %s", response.Message)
		}
	})

	t.Run("ToJSON BaseError", func(t *testing.T) {
		err := newError(ErrCodeConfigInvalid, "invalid config", SeverityHigh, false)
		err.userMsg = "User-friendly message"
		err.WithMetadata("key1", "value1")

		result, jsonErr := ToJSON(err)
		if jsonErr != nil {
			t.Fatalf("unexpected error: %v", jsonErr)
		}

		var response ErrorResponse
		if err := json.Unmarshal(result, &response); err != nil {
			t.Fatalf("failed to unmarshal JSON: %v", err)
		}

		if response.Code != string(ErrCodeConfigInvalid) {
			t.Errorf("expected code %s, got %s", ErrCodeConfigInvalid, response.Code)
		}

		if response.Message != "invalid config" {
			t.Errorf("expected message 'invalid config', got %s", response.Message)
		}

		if response.UserMsg != "User-friendly message" {
			t.Errorf("expected user message 'User-friendly message', got %s", response.UserMsg)
		}

		if response.Severity != string(SeverityHigh) {
			t.Errorf("expected severity %s, got %s", SeverityHigh, response.Severity)
		}

		if response.Retryable {
			t.Error("expected retryable to be false")
		}

		if response.Metadata["key1"] != "value1" {
			t.Error("expected metadata to be included")
		}
	})

	t.Run("ToJSON includes stack in debug mode", func(t *testing.T) {
		EnableDebugMode()
		defer DisableDebugMode()

		err := newError(ErrCodeConfigInvalid, "test error", SeverityHigh, false)
		result, jsonErr := ToJSON(err)

		if jsonErr != nil {
			t.Fatalf("unexpected error: %v", jsonErr)
		}

		var response ErrorResponse
		if err := json.Unmarshal(result, &response); err != nil {
			t.Fatalf("failed to unmarshal JSON: %v", err)
		}

		if len(response.Stack) == 0 {
			t.Error("expected stack trace in debug mode")
		}
	})

	t.Run("ToJSON excludes stack in production mode", func(t *testing.T) {
		DisableDebugMode()

		err := newError(ErrCodeConfigInvalid, "test error", SeverityHigh, false)
		result, jsonErr := ToJSON(err)

		if jsonErr != nil {
			t.Fatalf("unexpected error: %v", jsonErr)
		}

		var response ErrorResponse
		if err := json.Unmarshal(result, &response); err != nil {
			t.Fatalf("failed to unmarshal JSON: %v", err)
		}

		if len(response.Stack) > 0 {
			t.Error("expected no stack trace in production mode")
		}
	})
}

func TestFromJSON(t *testing.T) {
	t.Run("FromJSON valid error", func(t *testing.T) {
		jsonData := `{
			"code": "CONFIG_INVALID",
			"message": "invalid configuration",
			"user_message": "Configuration error",
			"severity": "high",
			"retryable": false,
			"metadata": {"key": "value"}
		}`

		err := FromJSON([]byte(jsonData))
		if err == nil {
			t.Fatal("expected error to be created")
		}

		var baseErr *BaseError
		if !As(err, &baseErr) {
			t.Fatal("expected BaseError")
		}

		if baseErr.Code() != "CONFIG_INVALID" {
			t.Errorf("expected code CONFIG_INVALID, got %s", baseErr.Code())
		}

		if baseErr.message != "invalid configuration" {
			t.Errorf("expected message 'invalid configuration', got %s", baseErr.message)
		}

		if baseErr.UserMessage() != "Configuration error" {
			t.Errorf("expected user message 'Configuration error', got %s", baseErr.UserMessage())
		}

		if baseErr.Severity() != SeverityHigh {
			t.Errorf("expected severity %s, got %s", SeverityHigh, baseErr.Severity())
		}

		if baseErr.Metadata()["key"] != "value" {
			t.Error("expected metadata to be preserved")
		}
	})

	t.Run("FromJSON invalid JSON", func(t *testing.T) {
		jsonData := `{invalid json`
		err := FromJSON([]byte(jsonData))

		if err == nil {
			t.Error("expected error for invalid JSON")
		}
	})
}

func TestSummarize(t *testing.T) {
	t.Run("Summarize nil error", func(t *testing.T) {
		summary := Summarize(nil)
		if summary.Code != "" {
			t.Error("expected empty summary for nil error")
		}
	})

	t.Run("Summarize standard error", func(t *testing.T) {
		err := errors.New("standard error")
		summary := Summarize(err)

		if summary.Code != "UNKNOWN" {
			t.Errorf("expected code UNKNOWN, got %s", summary.Code)
		}

		if summary.Message != "standard error" {
			t.Errorf("expected message 'standard error', got %s", summary.Message)
		}
	})

	t.Run("Summarize BaseError", func(t *testing.T) {
		originalErr := errors.New("cause")
		err := Wrap(originalErr, ErrCodeConfigParse, "parse failed")

		summary := Summarize(err)

		if summary.Code != string(ErrCodeConfigParse) {
			t.Errorf("expected code %s, got %s", ErrCodeConfigParse, summary.Code)
		}

		if !summary.HasCause {
			t.Error("expected HasCause to be true")
		}

		if summary.StackSize == 0 {
			t.Error("expected StackSize to be > 0")
		}
	})
}

func TestChain(t *testing.T) {
	t.Run("Chain nil error", func(t *testing.T) {
		chain := Chain(nil)
		if chain != nil {
			t.Error("expected nil chain for nil error")
		}
	})

	t.Run("Chain single error", func(t *testing.T) {
		err := errors.New("single error")
		chain := Chain(err)

		if len(chain) != 1 {
			t.Errorf("expected chain length 1, got %d", len(chain))
		}

		if chain[0] != err {
			t.Error("expected chain to contain the error")
		}
	})

	t.Run("Chain wrapped errors", func(t *testing.T) {
		err1 := errors.New("original")
		err2 := Wrap(err1, ErrCodeConfigParse, "wrapped once")
		err3 := Wrap(err2, ErrCodeConfigValidation, "wrapped twice")

		chain := Chain(err3)

		if len(chain) != 3 {
			t.Errorf("expected chain length 3, got %d", len(chain))
		}
	})
}

func TestUnwrap(t *testing.T) {
	t.Run("Unwrap wrapped error", func(t *testing.T) {
		originalErr := errors.New("original")
		wrappedErr := Wrap(originalErr, ErrCodeConfigParse, "wrapped")

		unwrapped := Unwrap(wrappedErr)
		if unwrapped != originalErr {
			t.Error("expected to unwrap to original error")
		}
	})

	t.Run("Unwrap non-wrapped error", func(t *testing.T) {
		err := errors.New("not wrapped")
		unwrapped := Unwrap(err)

		if unwrapped != nil {
			t.Errorf("expected nil for non-wrapped error, got %v", unwrapped)
		}
	})
}

func TestRootCause(t *testing.T) {
	t.Run("RootCause single error", func(t *testing.T) {
		err := errors.New("root")
		root := RootCause(err)

		if root != err {
			t.Error("expected root cause to be the error itself")
		}
	})

	t.Run("RootCause wrapped errors", func(t *testing.T) {
		root := errors.New("root cause")
		err1 := Wrap(root, ErrCodeConfigParse, "level 1")
		err2 := Wrap(err1, ErrCodeConfigValidation, "level 2")
		err3 := Wrap(err2, ErrCodeConfigInvalid, "level 3")

		foundRoot := RootCause(err3)
		if foundRoot != root {
			t.Error("expected to find root cause")
		}
	})
}

func TestFormatChain(t *testing.T) {
	t.Run("FormatChain nil error", func(t *testing.T) {
		result := FormatChain(nil)
		if result != "" {
			t.Errorf("expected empty string for nil error, got '%s'", result)
		}
	})

	t.Run("FormatChain wrapped errors", func(t *testing.T) {
		err1 := errors.New("root error")
		err2 := Wrap(err1, ErrCodeConfigParse, "parse error")
		err3 := Wrap(err2, ErrCodeConfigValidation, "validation error")

		result := FormatChain(err3)

		if !strings.Contains(result, "CONFIG_VALIDATION") {
			t.Errorf("expected outer error in chain: %s", result)
		}

		if !strings.Contains(result, "CONFIG_PARSE") {
			t.Errorf("expected middle error in chain: %s", result)
		}

		if !strings.Contains(result, "Wrapped by:") {
			t.Errorf("expected wrapping indicator: %s", result)
		}
	})
}

func BenchmarkFormat(b *testing.B) {
	err := newError(ErrCodeConfigInvalid, "test error", SeverityHigh, false)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = Format(err)
	}
}

func BenchmarkToJSON(b *testing.B) {
	err := newError(ErrCodeConfigInvalid, "test error", SeverityHigh, false)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = ToJSON(err)
	}
}

func ExampleFormat() {
	err := NewConfigMissingError("api_key")
	formatted := Format(err)
	println(formatted)
}

func ExampleFormatUser() {
	err := NewAuthFailedError("openai", errors.New("invalid token"))
	userMsg := FormatUser(err)
	println(userMsg)
}
