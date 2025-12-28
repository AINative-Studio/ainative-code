package errors

import (
	"errors"
	"fmt"
	"testing"
)

func TestBaseError(t *testing.T) {
	t.Run("Create base error", func(t *testing.T) {
		err := newError(ErrCodeConfigInvalid, "test error", SeverityHigh, false)
		if err == nil {
			t.Fatal("expected error to be created")
		}
		if err.Code() != ErrCodeConfigInvalid {
			t.Errorf("expected code %s, got %s", ErrCodeConfigInvalid, err.Code())
		}
		if err.Severity() != SeverityHigh {
			t.Errorf("expected severity %s, got %s", SeverityHigh, err.Severity())
		}
		if err.IsRetryable() {
			t.Error("expected error to not be retryable")
		}
	})

	t.Run("Error message formatting", func(t *testing.T) {
		err := newError(ErrCodeConfigInvalid, "test error", SeverityHigh, false)
		msg := err.Error()
		if msg != "[CONFIG_INVALID] test error" {
			t.Errorf("unexpected error message: %s", msg)
		}
	})

	t.Run("Stack trace capture", func(t *testing.T) {
		err := newError(ErrCodeConfigInvalid, "test error", SeverityHigh, false)
		stack := err.Stack()
		if len(stack) == 0 {
			t.Error("expected stack trace to be captured")
		}
	})

	t.Run("Metadata handling", func(t *testing.T) {
		err := newError(ErrCodeConfigInvalid, "test error", SeverityHigh, false)
		err.WithMetadata("key1", "value1")
		err.WithMetadata("key2", 123)

		metadata := err.Metadata()
		if metadata["key1"] != "value1" {
			t.Error("expected metadata key1 to be set")
		}
		if metadata["key2"] != 123 {
			t.Error("expected metadata key2 to be set")
		}
	})
}

func TestErrorWrapping(t *testing.T) {
	t.Run("Wrap error", func(t *testing.T) {
		originalErr := errors.New("original error")
		wrappedErr := Wrap(originalErr, ErrCodeConfigParse, "failed to parse config")

		if wrappedErr == nil {
			t.Fatal("expected wrapped error")
		}

		var baseErr *BaseError
		if !As(wrappedErr, &baseErr) {
			t.Fatal("expected BaseError")
		}

		if baseErr.Code() != ErrCodeConfigParse {
			t.Errorf("expected code %s, got %s", ErrCodeConfigParse, baseErr.Code())
		}

		if !errors.Is(wrappedErr, originalErr) {
			t.Error("wrapped error should contain original error")
		}
	})

	t.Run("Wrap nil error", func(t *testing.T) {
		wrappedErr := Wrap(nil, ErrCodeConfigParse, "test")
		if wrappedErr != nil {
			t.Error("wrapping nil should return nil")
		}
	})

	t.Run("Wrapf with formatting", func(t *testing.T) {
		originalErr := errors.New("original error")
		wrappedErr := Wrapf(originalErr, ErrCodeConfigParse, "failed to parse %s", "config.yaml")

		var baseErr *BaseError
		if !As(wrappedErr, &baseErr) {
			t.Fatal("expected BaseError")
		}

		if baseErr.message != "failed to parse config.yaml" {
			t.Errorf("unexpected message: %s", baseErr.message)
		}
	})

	t.Run("Unwrap error", func(t *testing.T) {
		originalErr := errors.New("original error")
		wrappedErr := Wrap(originalErr, ErrCodeConfigParse, "wrapped")

		var baseErr *BaseError
		if !As(wrappedErr, &baseErr) {
			t.Fatal("expected BaseError")
		}

		unwrapped := baseErr.Unwrap()
		if unwrapped != originalErr {
			t.Error("unwrap should return original error")
		}
	})
}

func TestErrorCode(t *testing.T) {
	t.Run("GetCode from error", func(t *testing.T) {
		err := newError(ErrCodeAuthFailed, "auth failed", SeverityHigh, false)
		code := GetCode(err)
		if code != ErrCodeAuthFailed {
			t.Errorf("expected code %s, got %s", ErrCodeAuthFailed, code)
		}
	})

	t.Run("GetCode from standard error", func(t *testing.T) {
		err := errors.New("standard error")
		code := GetCode(err)
		if code != "" {
			t.Errorf("expected empty code, got %s", code)
		}
	})

	t.Run("GetCode from wrapped error", func(t *testing.T) {
		originalErr := newError(ErrCodeAuthFailed, "auth failed", SeverityHigh, false)
		wrappedErr := Wrap(originalErr, ErrCodeConfigParse, "wrapped")

		code := GetCode(wrappedErr)
		if code != ErrCodeConfigParse {
			t.Errorf("expected code %s, got %s", ErrCodeConfigParse, code)
		}
	})
}

func TestIsRetryable(t *testing.T) {
	t.Run("Retryable error", func(t *testing.T) {
		err := newError(ErrCodeProviderTimeout, "timeout", SeverityMedium, true)
		if !IsRetryable(err) {
			t.Error("expected error to be retryable")
		}
	})

	t.Run("Non-retryable error", func(t *testing.T) {
		err := newError(ErrCodeConfigInvalid, "invalid config", SeverityHigh, false)
		if IsRetryable(err) {
			t.Error("expected error to not be retryable")
		}
	})

	t.Run("Standard error not retryable", func(t *testing.T) {
		err := errors.New("standard error")
		if IsRetryable(err) {
			t.Error("standard error should not be retryable")
		}
	})
}

func TestGetSeverity(t *testing.T) {
	tests := []struct {
		name     string
		severity Severity
	}{
		{"low severity", SeverityLow},
		{"medium severity", SeverityMedium},
		{"high severity", SeverityHigh},
		{"critical severity", SeverityCritical},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := newError(ErrCodeConfigInvalid, "test", tt.severity, false)
			severity := GetSeverity(err)
			if severity != tt.severity {
				t.Errorf("expected severity %s, got %s", tt.severity, severity)
			}
		})
	}

	t.Run("Standard error has low severity", func(t *testing.T) {
		err := errors.New("standard error")
		severity := GetSeverity(err)
		if severity != SeverityLow {
			t.Errorf("expected severity %s, got %s", SeverityLow, severity)
		}
	})
}

func TestErrorIs(t *testing.T) {
	t.Run("Is checks error type", func(t *testing.T) {
		targetErr := errors.New("target error")
		wrappedErr := Wrap(targetErr, ErrCodeConfigParse, "wrapped")

		if !Is(wrappedErr, targetErr) {
			t.Error("Is should find target error in chain")
		}
	})

	t.Run("Is with different errors", func(t *testing.T) {
		err1 := errors.New("error 1")
		err2 := errors.New("error 2")

		if Is(err1, err2) {
			t.Error("Is should not match different errors")
		}
	})
}

func TestErrorAs(t *testing.T) {
	t.Run("As extracts BaseError", func(t *testing.T) {
		err := newError(ErrCodeConfigInvalid, "test", SeverityHigh, false)

		var baseErr *BaseError
		if !As(err, &baseErr) {
			t.Fatal("As should extract BaseError")
		}

		if baseErr.Code() != ErrCodeConfigInvalid {
			t.Error("extracted error should have correct code")
		}
	})

	t.Run("As fails with standard error", func(t *testing.T) {
		err := errors.New("standard error")

		var baseErr *BaseError
		if As(err, &baseErr) {
			t.Error("As should fail with standard error")
		}
	})
}

func TestUserMessage(t *testing.T) {
	t.Run("Custom user message", func(t *testing.T) {
		err := newError(ErrCodeConfigInvalid, "technical message", SeverityHigh, false)
		err.userMsg = "User-friendly message"

		if err.UserMessage() != "User-friendly message" {
			t.Errorf("expected user message, got: %s", err.UserMessage())
		}
	})

	t.Run("Default to technical message", func(t *testing.T) {
		err := newError(ErrCodeConfigInvalid, "technical message", SeverityHigh, false)

		if err.UserMessage() != "technical message" {
			t.Errorf("expected technical message as default, got: %s", err.UserMessage())
		}
	})
}

func TestStackTrace(t *testing.T) {
	t.Run("Stack trace formatting", func(t *testing.T) {
		err := newError(ErrCodeConfigInvalid, "test", SeverityHigh, false)
		trace := err.StackTrace()

		if trace == "" {
			t.Error("expected non-empty stack trace")
		}

		if trace[:11] != "Stack trace" {
			t.Error("stack trace should start with 'Stack trace'")
		}
	})

	t.Run("Empty stack trace", func(t *testing.T) {
		err := &BaseError{
			code:    ErrCodeConfigInvalid,
			message: "test",
			stack:   []StackFrame{},
		}

		trace := err.StackTrace()
		if trace != "" {
			t.Error("expected empty stack trace")
		}
	})
}

func BenchmarkErrorCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = newError(ErrCodeConfigInvalid, "test error", SeverityHigh, false)
	}
}

func BenchmarkErrorWrapping(b *testing.B) {
	originalErr := errors.New("original error")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = Wrap(originalErr, ErrCodeConfigParse, "wrapped error")
	}
}

func BenchmarkGetCode(b *testing.B) {
	err := newError(ErrCodeAuthFailed, "auth failed", SeverityHigh, false)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = GetCode(err)
	}
}

func ExampleWrap() {
	originalErr := fmt.Errorf("connection refused")
	wrappedErr := Wrap(originalErr, ErrCodeDBConnection, "failed to connect to database")
	fmt.Println(wrappedErr)
	// Output: [DB_CONNECTION_FAILED] failed to connect to database: connection refused
}

func ExampleWrap_withConfigError() {
	err := NewConfigInvalidError("timeout", "must be a positive integer")
	fmt.Println(err.Code())
	fmt.Println(err.Severity())
	fmt.Println(err.IsRetryable())
	// Output:
	// CONFIG_INVALID
	// high
	// false
}
