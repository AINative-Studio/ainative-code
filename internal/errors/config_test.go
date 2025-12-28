package errors

import (
	"errors"
	"strings"
	"testing"
)

func TestConfigError(t *testing.T) {
	t.Run("NewConfigInvalidError", func(t *testing.T) {
		err := NewConfigInvalidError("api_key", "must not be empty")

		if err.Code() != ErrCodeConfigInvalid {
			t.Errorf("expected code %s, got %s", ErrCodeConfigInvalid, err.Code())
		}

		if err.ConfigKey != "api_key" {
			t.Errorf("expected ConfigKey 'api_key', got '%s'", err.ConfigKey)
		}

		if err.Severity() != SeverityHigh {
			t.Errorf("expected severity %s, got %s", SeverityHigh, err.Severity())
		}

		userMsg := err.UserMessage()
		if !strings.Contains(userMsg, "api_key") {
			t.Errorf("user message should contain config key: %s", userMsg)
		}
	})

	t.Run("NewConfigMissingError", func(t *testing.T) {
		err := NewConfigMissingError("database_url")

		if err.Code() != ErrCodeConfigMissing {
			t.Errorf("expected code %s, got %s", ErrCodeConfigMissing, err.Code())
		}

		if err.ConfigKey != "database_url" {
			t.Errorf("expected ConfigKey 'database_url', got '%s'", err.ConfigKey)
		}

		userMsg := err.UserMessage()
		if !strings.Contains(userMsg, "database_url") {
			t.Errorf("user message should contain config key: %s", userMsg)
		}
	})

	t.Run("NewConfigParseError", func(t *testing.T) {
		originalErr := errors.New("invalid YAML syntax")
		err := NewConfigParseError("/path/to/config.yaml", originalErr)

		if err.Code() != ErrCodeConfigParse {
			t.Errorf("expected code %s, got %s", ErrCodeConfigParse, err.Code())
		}

		if err.ConfigPath != "/path/to/config.yaml" {
			t.Errorf("expected ConfigPath '/path/to/config.yaml', got '%s'", err.ConfigPath)
		}

		if err.Severity() != SeverityCritical {
			t.Errorf("expected severity %s, got %s", SeverityCritical, err.Severity())
		}

		if err.Unwrap() != originalErr {
			t.Error("expected error to wrap original error")
		}
	})

	t.Run("NewConfigValidationError", func(t *testing.T) {
		err := NewConfigValidationError("port", "must be between 1 and 65535")

		if err.Code() != ErrCodeConfigValidation {
			t.Errorf("expected code %s, got %s", ErrCodeConfigValidation, err.Code())
		}

		if err.ConfigKey != "port" {
			t.Errorf("expected ConfigKey 'port', got '%s'", err.ConfigKey)
		}

		userMsg := err.UserMessage()
		if !strings.Contains(userMsg, "port") || !strings.Contains(userMsg, "1 and 65535") {
			t.Errorf("user message should contain validation details: %s", userMsg)
		}
	})

	t.Run("WithPath", func(t *testing.T) {
		err := NewConfigInvalidError("test_key", "invalid value")
		err.WithPath("/etc/config.yaml")

		if err.ConfigPath != "/etc/config.yaml" {
			t.Errorf("expected ConfigPath '/etc/config.yaml', got '%s'", err.ConfigPath)
		}
	})

	t.Run("Error message format", func(t *testing.T) {
		err := NewConfigInvalidError("api_key", "must not be empty")
		errMsg := err.Error()

		if !strings.Contains(errMsg, "CONFIG_INVALID") {
			t.Errorf("error message should contain code: %s", errMsg)
		}

		if !strings.Contains(errMsg, "api_key") {
			t.Errorf("error message should contain config key: %s", errMsg)
		}
	})

	t.Run("Retryability", func(t *testing.T) {
		err := NewConfigInvalidError("api_key", "invalid")
		if err.IsRetryable() {
			t.Error("config errors should not be retryable")
		}
	})
}

func TestConfigErrorChaining(t *testing.T) {
	t.Run("Wrap config error", func(t *testing.T) {
		configErr := NewConfigMissingError("database_url")
		wrappedErr := Wrap(configErr, ErrCodeConfigValidation, "configuration validation failed")

		var baseErr *BaseError
		if !As(wrappedErr, &baseErr) {
			t.Fatal("expected BaseError")
		}

		if baseErr.Code() != ErrCodeConfigValidation {
			t.Errorf("expected outer code %s, got %s", ErrCodeConfigValidation, baseErr.Code())
		}

		// Check that we can still extract the original config error
		var originalConfigErr *ConfigError
		if !As(wrappedErr, &originalConfigErr) {
			t.Fatal("expected to extract ConfigError from chain")
		}

		if originalConfigErr.ConfigKey != "database_url" {
			t.Errorf("expected ConfigKey 'database_url', got '%s'", originalConfigErr.ConfigKey)
		}
	})
}

func BenchmarkNewConfigError(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewConfigInvalidError("api_key", "invalid value")
	}
}

func ExampleNewConfigInvalidError() {
	err := NewConfigInvalidError("timeout", "must be a positive integer")
	println(err.Error())
	println(err.UserMessage())
}

func ExampleNewConfigMissingError() {
	err := NewConfigMissingError("api_key")
	println(err.ConfigKey)
	println(err.Severity())
}
