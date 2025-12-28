package errors

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestProviderError(t *testing.T) {
	t.Run("NewProviderUnavailableError", func(t *testing.T) {
		originalErr := errors.New("connection refused")
		err := NewProviderUnavailableError("openai", originalErr)

		if err.Code() != ErrCodeProviderUnavailable {
			t.Errorf("expected code %s, got %s", ErrCodeProviderUnavailable, err.Code())
		}

		if err.ProviderName != "openai" {
			t.Errorf("expected ProviderName 'openai', got '%s'", err.ProviderName)
		}

		if !err.IsRetryable() {
			t.Error("unavailable provider error should be retryable")
		}

		if err.Unwrap() != originalErr {
			t.Error("expected error to wrap original error")
		}

		userMsg := err.UserMessage()
		if !strings.Contains(userMsg, "unavailable") {
			t.Errorf("user message should mention unavailability: %s", userMsg)
		}
	})

	t.Run("NewProviderTimeoutError", func(t *testing.T) {
		timeout := 30 * time.Second
		err := NewProviderTimeoutError("anthropic", "claude-3", timeout)

		if err.Code() != ErrCodeProviderTimeout {
			t.Errorf("expected code %s, got %s", ErrCodeProviderTimeout, err.Code())
		}

		if err.ProviderName != "anthropic" {
			t.Errorf("expected ProviderName 'anthropic', got '%s'", err.ProviderName)
		}

		if err.Model != "claude-3" {
			t.Errorf("expected Model 'claude-3', got '%s'", err.Model)
		}

		if !err.IsRetryable() {
			t.Error("timeout error should be retryable")
		}

		errMsg := err.Error()
		if !strings.Contains(errMsg, "30s") {
			t.Errorf("error message should contain timeout duration: %s", errMsg)
		}
	})

	t.Run("NewProviderRateLimitError", func(t *testing.T) {
		retryAfter := 60 * time.Second
		err := NewProviderRateLimitError("openai", retryAfter)

		if err.Code() != ErrCodeProviderRateLimit {
			t.Errorf("expected code %s, got %s", ErrCodeProviderRateLimit, err.Code())
		}

		if err.ProviderName != "openai" {
			t.Errorf("expected ProviderName 'openai', got '%s'", err.ProviderName)
		}

		if !err.IsRetryable() {
			t.Error("rate limit error should be retryable")
		}

		if err.RetryAfter == nil {
			t.Fatal("expected RetryAfter to be set")
		}

		if *err.RetryAfter != retryAfter {
			t.Errorf("expected RetryAfter %v, got %v", retryAfter, *err.RetryAfter)
		}

		delay := err.GetRetryDelay()
		if delay != retryAfter {
			t.Errorf("expected delay %v, got %v", retryAfter, delay)
		}

		userMsg := err.UserMessage()
		if !strings.Contains(userMsg, "Rate limit") {
			t.Errorf("user message should mention rate limit: %s", userMsg)
		}
	})

	t.Run("NewProviderInvalidResponseError", func(t *testing.T) {
		originalErr := errors.New("JSON parse error")
		err := NewProviderInvalidResponseError("google", "malformed JSON", originalErr)

		if err.Code() != ErrCodeProviderInvalidResponse {
			t.Errorf("expected code %s, got %s", ErrCodeProviderInvalidResponse, err.Code())
		}

		if err.ProviderName != "google" {
			t.Errorf("expected ProviderName 'google', got '%s'", err.ProviderName)
		}

		if !err.IsRetryable() {
			t.Error("invalid response error should be retryable")
		}

		if err.Unwrap() != originalErr {
			t.Error("expected error to wrap original error")
		}

		errMsg := err.Error()
		if !strings.Contains(errMsg, "malformed JSON") {
			t.Errorf("error message should contain reason: %s", errMsg)
		}
	})

	t.Run("NewProviderNotFoundError", func(t *testing.T) {
		err := NewProviderNotFoundError("unknown-provider")

		if err.Code() != ErrCodeProviderNotFound {
			t.Errorf("expected code %s, got %s", ErrCodeProviderNotFound, err.Code())
		}

		if err.ProviderName != "unknown-provider" {
			t.Errorf("expected ProviderName 'unknown-provider', got '%s'", err.ProviderName)
		}

		if err.Severity() != SeverityCritical {
			t.Errorf("expected severity %s, got %s", SeverityCritical, err.Severity())
		}

		userMsg := err.UserMessage()
		if !strings.Contains(userMsg, "not available") {
			t.Errorf("user message should mention provider not available: %s", userMsg)
		}
	})

	t.Run("WithModel", func(t *testing.T) {
		err := NewProviderError(ErrCodeProviderTimeout, "timeout", "openai")
		err.WithModel("gpt-4")

		if err.Model != "gpt-4" {
			t.Errorf("expected Model 'gpt-4', got '%s'", err.Model)
		}
	})

	t.Run("WithRequestID", func(t *testing.T) {
		err := NewProviderError(ErrCodeProviderTimeout, "timeout", "openai")
		err.WithRequestID("req-123456")

		if err.RequestID != "req-123456" {
			t.Errorf("expected RequestID 'req-123456', got '%s'", err.RequestID)
		}
	})

	t.Run("WithStatusCode", func(t *testing.T) {
		err := NewProviderError(ErrCodeProviderTimeout, "timeout", "openai")
		err.WithStatusCode(503)

		if err.StatusCode != 503 {
			t.Errorf("expected StatusCode 503, got %d", err.StatusCode)
		}
	})

	t.Run("Method chaining", func(t *testing.T) {
		err := NewProviderError(ErrCodeProviderTimeout, "test", "openai").
			WithModel("gpt-4").
			WithRequestID("req-abc").
			WithStatusCode(500)

		if err.Model != "gpt-4" {
			t.Error("expected Model to be set via chaining")
		}
		if err.RequestID != "req-abc" {
			t.Error("expected RequestID to be set via chaining")
		}
		if err.StatusCode != 500 {
			t.Error("expected StatusCode to be set via chaining")
		}
	})

	t.Run("ShouldRetry", func(t *testing.T) {
		// Retryable errors
		timeoutErr := NewProviderTimeoutError("provider", "model", 30*time.Second)
		if !timeoutErr.ShouldRetry() {
			t.Error("timeout error should be retryable")
		}

		rateLimitErr := NewProviderRateLimitError("provider", 60*time.Second)
		if !rateLimitErr.ShouldRetry() {
			t.Error("rate limit error should be retryable")
		}

		// Non-retryable errors
		notFoundErr := NewProviderNotFoundError("provider")
		if notFoundErr.ShouldRetry() {
			t.Error("not found error should not be retryable")
		}
	})

	t.Run("GetRetryDelay with no RetryAfter", func(t *testing.T) {
		err := NewProviderTimeoutError("provider", "model", 30*time.Second)
		delay := err.GetRetryDelay()

		if delay != 0 {
			t.Errorf("expected delay 0 when RetryAfter not set, got %v", delay)
		}
	})
}

func TestProviderErrorWrapping(t *testing.T) {
	t.Run("Wrap provider error", func(t *testing.T) {
		providerErr := NewProviderTimeoutError("openai", "gpt-4", 30*time.Second)
		wrappedErr := Wrap(providerErr, ErrCodeProviderInvalidResponse, "invalid response after retry")

		var baseErr *BaseError
		if !As(wrappedErr, &baseErr) {
			t.Fatal("expected BaseError")
		}

		// Check that we can still extract the original provider error
		var originalProviderErr *ProviderError
		if !As(wrappedErr, &originalProviderErr) {
			t.Fatal("expected to extract ProviderError from chain")
		}

		if originalProviderErr.Model != "gpt-4" {
			t.Errorf("expected Model 'gpt-4', got '%s'", originalProviderErr.Model)
		}
	})
}

func BenchmarkNewProviderError(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewProviderTimeoutError("openai", "gpt-4", 30*time.Second)
	}
}

func ExampleNewProviderRateLimitError() {
	retryAfter := 60 * time.Second
	err := NewProviderRateLimitError("openai", retryAfter)
	println(err.Error())
	println(err.GetRetryDelay())
	println(err.ShouldRetry())
}

func ExampleNewProviderTimeoutError() {
	err := NewProviderTimeoutError("anthropic", "claude-3", 30*time.Second)
	println(err.ProviderName)
	println(err.Model)
	println(err.IsRetryable())
}
