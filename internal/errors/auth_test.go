package errors

import (
	"errors"
	"strings"
	"testing"
)

func TestAuthenticationError(t *testing.T) {
	t.Run("NewAuthFailedError", func(t *testing.T) {
		originalErr := errors.New("invalid credentials")
		err := NewAuthFailedError("openai", originalErr)

		if err.Code() != ErrCodeAuthFailed {
			t.Errorf("expected code %s, got %s", ErrCodeAuthFailed, err.Code())
		}

		if err.Provider != "openai" {
			t.Errorf("expected Provider 'openai', got '%s'", err.Provider)
		}

		if err.Severity() != SeverityHigh {
			t.Errorf("expected severity %s, got %s", SeverityHigh, err.Severity())
		}

		if err.Unwrap() != originalErr {
			t.Error("expected error to wrap original error")
		}

		userMsg := err.UserMessage()
		if !strings.Contains(userMsg, "Authentication failed") {
			t.Errorf("user message should indicate auth failure: %s", userMsg)
		}
	})

	t.Run("NewInvalidTokenError", func(t *testing.T) {
		err := NewInvalidTokenError("anthropic")

		if err.Code() != ErrCodeAuthInvalidToken {
			t.Errorf("expected code %s, got %s", ErrCodeAuthInvalidToken, err.Code())
		}

		if err.Provider != "anthropic" {
			t.Errorf("expected Provider 'anthropic', got '%s'", err.Provider)
		}

		userMsg := err.UserMessage()
		if !strings.Contains(userMsg, "invalid") {
			t.Errorf("user message should mention invalid token: %s", userMsg)
		}
	})

	t.Run("NewExpiredTokenError", func(t *testing.T) {
		err := NewExpiredTokenError("google")

		if err.Code() != ErrCodeAuthExpiredToken {
			t.Errorf("expected code %s, got %s", ErrCodeAuthExpiredToken, err.Code())
		}

		if err.Provider != "google" {
			t.Errorf("expected Provider 'google', got '%s'", err.Provider)
		}

		if !err.IsRetryable() {
			t.Error("expired token error should be retryable (for refresh)")
		}

		userMsg := err.UserMessage()
		if !strings.Contains(userMsg, "expired") {
			t.Errorf("user message should mention expiration: %s", userMsg)
		}
	})

	t.Run("NewPermissionDeniedError", func(t *testing.T) {
		err := NewPermissionDeniedError("/api/users", "delete")

		if err.Code() != ErrCodeAuthPermissionDenied {
			t.Errorf("expected code %s, got %s", ErrCodeAuthPermissionDenied, err.Code())
		}

		if err.Resource != "/api/users" {
			t.Errorf("expected Resource '/api/users', got '%s'", err.Resource)
		}

		errMsg := err.Error()
		if !strings.Contains(errMsg, "delete") {
			t.Errorf("error message should contain action: %s", errMsg)
		}

		userMsg := err.UserMessage()
		if !strings.Contains(userMsg, "permission") {
			t.Errorf("user message should mention permission: %s", userMsg)
		}
	})

	t.Run("NewInvalidCredentialsError", func(t *testing.T) {
		err := NewInvalidCredentialsError("aws")

		if err.Code() != ErrCodeAuthInvalidCredentials {
			t.Errorf("expected code %s, got %s", ErrCodeAuthInvalidCredentials, err.Code())
		}

		if err.Provider != "aws" {
			t.Errorf("expected Provider 'aws', got '%s'", err.Provider)
		}

		userMsg := err.UserMessage()
		if !strings.Contains(userMsg, "incorrect") {
			t.Errorf("user message should mention incorrect credentials: %s", userMsg)
		}
	})

	t.Run("WithProvider", func(t *testing.T) {
		err := NewAuthenticationError(ErrCodeAuthFailed, "test")
		err.WithProvider("custom-provider")

		if err.Provider != "custom-provider" {
			t.Errorf("expected Provider 'custom-provider', got '%s'", err.Provider)
		}
	})

	t.Run("WithUserID", func(t *testing.T) {
		err := NewAuthenticationError(ErrCodeAuthFailed, "test")
		err.WithUserID("user123")

		if err.UserID != "user123" {
			t.Errorf("expected UserID 'user123', got '%s'", err.UserID)
		}
	})

	t.Run("WithResource", func(t *testing.T) {
		err := NewAuthenticationError(ErrCodeAuthFailed, "test")
		err.WithResource("/api/data")

		if err.Resource != "/api/data" {
			t.Errorf("expected Resource '/api/data', got '%s'", err.Resource)
		}
	})

	t.Run("Method chaining", func(t *testing.T) {
		err := NewAuthenticationError(ErrCodeAuthFailed, "test").
			WithProvider("test-provider").
			WithUserID("user456").
			WithResource("/api/test")

		if err.Provider != "test-provider" {
			t.Error("expected Provider to be set via chaining")
		}
		if err.UserID != "user456" {
			t.Error("expected UserID to be set via chaining")
		}
		if err.Resource != "/api/test" {
			t.Error("expected Resource to be set via chaining")
		}
	})

	t.Run("Retryability", func(t *testing.T) {
		// Expired tokens should be retryable
		expiredErr := NewExpiredTokenError("provider")
		if !expiredErr.IsRetryable() {
			t.Error("expired token error should be retryable")
		}

		// Invalid credentials should not be retryable
		invalidErr := NewInvalidCredentialsError("provider")
		if invalidErr.IsRetryable() {
			t.Error("invalid credentials error should not be retryable")
		}
	})
}

func TestAuthenticationErrorWrapping(t *testing.T) {
	t.Run("Wrap authentication error", func(t *testing.T) {
		authErr := NewInvalidTokenError("openai")
		wrappedErr := Wrap(authErr, ErrCodeAuthPermissionDenied, "cannot access resource")

		var baseErr *BaseError
		if !As(wrappedErr, &baseErr) {
			t.Fatal("expected BaseError")
		}

		// Check that we can still extract the original auth error
		var originalAuthErr *AuthenticationError
		if !As(wrappedErr, &originalAuthErr) {
			t.Fatal("expected to extract AuthenticationError from chain")
		}

		if originalAuthErr.Provider != "openai" {
			t.Errorf("expected Provider 'openai', got '%s'", originalAuthErr.Provider)
		}
	})
}

func BenchmarkNewAuthenticationError(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewInvalidTokenError("provider")
	}
}

func ExampleNewAuthFailedError() {
	originalErr := errors.New("connection timeout")
	err := NewAuthFailedError("openai", originalErr)
	println(err.Error())
	println(err.UserMessage())
}

func ExampleNewPermissionDeniedError() {
	err := NewPermissionDeniedError("/api/admin", "access")
	println(err.Resource)
	println(err.Severity())
}
