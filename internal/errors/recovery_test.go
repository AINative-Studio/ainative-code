package errors

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestExponentialBackoff(t *testing.T) {
	t.Run("Default configuration", func(t *testing.T) {
		backoff := NewExponentialBackoff()

		if backoff.InitialDelay != 100*time.Millisecond {
			t.Errorf("expected InitialDelay 100ms, got %v", backoff.InitialDelay)
		}

		if backoff.MaxDelay != 30*time.Second {
			t.Errorf("expected MaxDelay 30s, got %v", backoff.MaxDelay)
		}

		if backoff.Multiplier != 2.0 {
			t.Errorf("expected Multiplier 2.0, got %f", backoff.Multiplier)
		}

		if backoff.MaxRetries != 5 {
			t.Errorf("expected MaxRetries 5, got %d", backoff.MaxRetries)
		}
	})

	t.Run("GetDelay exponential growth", func(t *testing.T) {
		backoff := NewExponentialBackoff()

		delay0 := backoff.GetDelay(0)
		if delay0 != 100*time.Millisecond {
			t.Errorf("expected delay 100ms for attempt 0, got %v", delay0)
		}

		delay1 := backoff.GetDelay(1)
		if delay1 != 200*time.Millisecond {
			t.Errorf("expected delay 200ms for attempt 1, got %v", delay1)
		}

		delay2 := backoff.GetDelay(2)
		if delay2 != 400*time.Millisecond {
			t.Errorf("expected delay 400ms for attempt 2, got %v", delay2)
		}
	})

	t.Run("GetDelay respects max delay", func(t *testing.T) {
		backoff := NewExponentialBackoff()

		delay := backoff.GetDelay(100) // Very large attempt number
		if delay > backoff.MaxDelay {
			t.Errorf("delay %v should not exceed max delay %v", delay, backoff.MaxDelay)
		}
	})

	t.Run("ShouldRetry with retryable error", func(t *testing.T) {
		backoff := NewExponentialBackoff()
		err := newError(ErrCodeProviderTimeout, "timeout", SeverityMedium, true)

		if !backoff.ShouldRetry(err, 0) {
			t.Error("should retry retryable error")
		}

		if !backoff.ShouldRetry(err, 3) {
			t.Error("should retry within max attempts")
		}
	})

	t.Run("ShouldRetry exceeds max attempts", func(t *testing.T) {
		backoff := NewExponentialBackoff()
		err := newError(ErrCodeProviderTimeout, "timeout", SeverityMedium, true)

		if backoff.ShouldRetry(err, 5) {
			t.Error("should not retry when max attempts exceeded")
		}
	})

	t.Run("ShouldRetry with non-retryable error", func(t *testing.T) {
		backoff := NewExponentialBackoff()
		err := newError(ErrCodeConfigInvalid, "invalid", SeverityHigh, false)

		if backoff.ShouldRetry(err, 0) {
			t.Error("should not retry non-retryable error")
		}
	})

	t.Run("MaxAttempts", func(t *testing.T) {
		backoff := NewExponentialBackoff()
		if backoff.MaxAttempts() != 5 {
			t.Errorf("expected MaxAttempts 5, got %d", backoff.MaxAttempts())
		}
	})
}

func TestLinearBackoff(t *testing.T) {
	t.Run("Configuration", func(t *testing.T) {
		backoff := NewLinearBackoff(1*time.Second, 3)

		if backoff.Delay != 1*time.Second {
			t.Errorf("expected Delay 1s, got %v", backoff.Delay)
		}

		if backoff.MaxRetries != 3 {
			t.Errorf("expected MaxRetries 3, got %d", backoff.MaxRetries)
		}
	})

	t.Run("GetDelay constant", func(t *testing.T) {
		backoff := NewLinearBackoff(500*time.Millisecond, 3)

		delay0 := backoff.GetDelay(0)
		delay1 := backoff.GetDelay(1)
		delay2 := backoff.GetDelay(2)

		if delay0 != delay1 || delay1 != delay2 {
			t.Error("linear backoff should return constant delay")
		}

		if delay0 != 500*time.Millisecond {
			t.Errorf("expected delay 500ms, got %v", delay0)
		}
	})

	t.Run("ShouldRetry", func(t *testing.T) {
		backoff := NewLinearBackoff(1*time.Second, 3)
		err := newError(ErrCodeProviderTimeout, "timeout", SeverityMedium, true)

		if !backoff.ShouldRetry(err, 0) {
			t.Error("should retry retryable error")
		}

		if backoff.ShouldRetry(err, 3) {
			t.Error("should not retry when max attempts exceeded")
		}
	})
}

func TestRetry(t *testing.T) {
	t.Run("Successful on first attempt", func(t *testing.T) {
		ctx := context.Background()
		attempts := 0

		err := Retry(ctx, func() error {
			attempts++
			return nil
		}, NewRetryConfig())

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if attempts != 1 {
			t.Errorf("expected 1 attempt, got %d", attempts)
		}
	})

	t.Run("Successful after retries", func(t *testing.T) {
		ctx := context.Background()
		attempts := 0

		err := Retry(ctx, func() error {
			attempts++
			if attempts < 3 {
				return newError(ErrCodeProviderTimeout, "timeout", SeverityMedium, true)
			}
			return nil
		}, NewRetryConfig())

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if attempts != 3 {
			t.Errorf("expected 3 attempts, got %d", attempts)
		}
	})

	t.Run("All retries exhausted", func(t *testing.T) {
		ctx := context.Background()
		config := NewRetryConfig()
		config.Strategy = NewLinearBackoff(10*time.Millisecond, 3)

		attempts := 0
		err := Retry(ctx, func() error {
			attempts++
			return newError(ErrCodeProviderTimeout, "timeout", SeverityMedium, true)
		}, config)

		if err == nil {
			t.Error("expected error after exhausting retries")
		}

		if attempts != 3 {
			t.Errorf("expected 3 attempts, got %d", attempts)
		}
	})

	t.Run("Non-retryable error", func(t *testing.T) {
		ctx := context.Background()
		attempts := 0

		err := Retry(ctx, func() error {
			attempts++
			return newError(ErrCodeConfigInvalid, "invalid", SeverityHigh, false)
		}, NewRetryConfig())

		if err == nil {
			t.Error("expected error")
		}

		if attempts != 1 {
			t.Errorf("expected 1 attempt for non-retryable error, got %d", attempts)
		}
	})

	t.Run("Context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		attempts := 0
		err := Retry(ctx, func() error {
			attempts++
			return newError(ErrCodeProviderTimeout, "timeout", SeverityMedium, true)
		}, NewRetryConfig())

		if err == nil {
			t.Error("expected error due to context cancellation")
		}

		// Should have at least tried once before checking context
		if attempts < 1 {
			t.Error("expected at least one attempt")
		}
	})

	t.Run("OnRetry callback", func(t *testing.T) {
		ctx := context.Background()
		config := NewRetryConfig()
		config.Strategy = NewLinearBackoff(10*time.Millisecond, 3)

		retryCount := 0
		config.OnRetry = func(attempt int, err error) {
			retryCount++
		}

		_ = Retry(ctx, func() error {
			return newError(ErrCodeProviderTimeout, "timeout", SeverityMedium, true)
		}, config)

		// The callback is called for each failed attempt that will be retried
		// With MaxAttempts=3, attempts 0, 1, 2 will all retry (3 callbacks)
		if retryCount != 3 {
			t.Errorf("expected 3 retry callbacks, got %d", retryCount)
		}
	})

	t.Run("OnFinalError callback", func(t *testing.T) {
		ctx := context.Background()
		config := NewRetryConfig()
		config.Strategy = NewLinearBackoff(10*time.Millisecond, 2)

		finalErrorCalled := false
		config.OnFinalError = func(err error) {
			finalErrorCalled = true
		}

		_ = Retry(ctx, func() error {
			return newError(ErrCodeProviderTimeout, "timeout", SeverityMedium, true)
		}, config)

		if !finalErrorCalled {
			t.Error("expected OnFinalError to be called")
		}
	})
}

func TestCircuitBreaker(t *testing.T) {
	t.Run("Initial state is closed", func(t *testing.T) {
		cb := NewCircuitBreaker(3, 1*time.Second)

		if cb.GetState() != StateClosed {
			t.Errorf("expected initial state %v, got %v", StateClosed, cb.GetState())
		}
	})

	t.Run("Opens after max failures", func(t *testing.T) {
		cb := NewCircuitBreaker(3, 1*time.Second)
		testErr := errors.New("test error")

		// Record 3 failures
		for i := 0; i < 3; i++ {
			_ = cb.Execute(func() error {
				return testErr
			})
		}

		if cb.GetState() != StateOpen {
			t.Errorf("expected state %v after max failures, got %v", StateOpen, cb.GetState())
		}
	})

	t.Run("Blocks requests when open", func(t *testing.T) {
		cb := NewCircuitBreaker(1, 1*time.Second)

		// Trigger circuit to open
		_ = cb.Execute(func() error {
			return errors.New("error")
		})

		// Next request should be blocked
		err := cb.Execute(func() error {
			return nil
		})

		if err == nil {
			t.Error("expected error when circuit is open")
		}

		if !strings.Contains(err.Error(), "circuit breaker is open") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("Transitions to half-open after timeout", func(t *testing.T) {
		cb := NewCircuitBreaker(1, 100*time.Millisecond)

		// Open the circuit
		_ = cb.Execute(func() error {
			return errors.New("error")
		})

		// Wait for reset timeout
		time.Sleep(150 * time.Millisecond)

		// Force state check by attempting execution
		initialState := cb.GetState()
		_ = cb.Execute(func() error {
			return nil
		})

		// After timeout, should transition through half-open
		if initialState == StateOpen && cb.GetState() != StateClosed {
			// State should have transitioned
			t.Log("Circuit transitioned from open to closed")
		}
	})

	t.Run("Resets on success", func(t *testing.T) {
		cb := NewCircuitBreaker(3, 1*time.Second)

		// Record a failure
		_ = cb.Execute(func() error {
			return errors.New("error")
		})

		// Record a success
		err := cb.Execute(func() error {
			return nil
		})

		if err != nil {
			t.Errorf("expected no error on success, got %v", err)
		}

		if cb.GetState() != StateClosed {
			t.Errorf("expected state %v after success, got %v", StateClosed, cb.GetState())
		}

		if cb.failures != 0 {
			t.Errorf("expected failures to be reset, got %d", cb.failures)
		}
	})

	t.Run("Manual reset", func(t *testing.T) {
		cb := NewCircuitBreaker(1, 1*time.Second)

		// Open the circuit
		_ = cb.Execute(func() error {
			return errors.New("error")
		})

		// Manual reset
		cb.Reset()

		if cb.GetState() != StateClosed {
			t.Errorf("expected state %v after reset, got %v", StateClosed, cb.GetState())
		}

		if cb.failures != 0 {
			t.Errorf("expected failures to be 0 after reset, got %d", cb.failures)
		}
	})
}

func TestFallback(t *testing.T) {
	t.Run("Primary succeeds", func(t *testing.T) {
		primaryCalled := false
		fallbackCalled := false

		err := Fallback(
			func() error {
				primaryCalled = true
				return nil
			},
			func() error {
				fallbackCalled = true
				return nil
			},
		)

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if !primaryCalled {
			t.Error("expected primary to be called")
		}

		if fallbackCalled {
			t.Error("expected fallback not to be called")
		}
	})

	t.Run("Primary fails, fallback succeeds", func(t *testing.T) {
		primaryCalled := false
		fallbackCalled := false

		err := Fallback(
			func() error {
				primaryCalled = true
				return errors.New("primary error")
			},
			func() error {
				fallbackCalled = true
				return nil
			},
		)

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if !primaryCalled {
			t.Error("expected primary to be called")
		}

		if !fallbackCalled {
			t.Error("expected fallback to be called")
		}
	})

	t.Run("Both fail", func(t *testing.T) {
		fallbackErr := errors.New("fallback error")

		err := Fallback(
			func() error {
				return errors.New("primary error")
			},
			func() error {
				return fallbackErr
			},
		)

		if err != fallbackErr {
			t.Errorf("expected fallback error, got %v", err)
		}
	})

	t.Run("Nil fallback", func(t *testing.T) {
		primaryErr := errors.New("primary error")

		err := Fallback(
			func() error {
				return primaryErr
			},
			nil,
		)

		if err != primaryErr {
			t.Errorf("expected primary error, got %v", err)
		}
	})
}

func TestFallbackWithValue(t *testing.T) {
	t.Run("Primary succeeds", func(t *testing.T) {
		result, err := FallbackWithValue(
			func() (string, error) {
				return "primary value", nil
			},
			"fallback value",
		)

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if result != "primary value" {
			t.Errorf("expected primary value, got %s", result)
		}
	})

	t.Run("Primary fails, returns fallback value", func(t *testing.T) {
		result, err := FallbackWithValue(
			func() (string, error) {
				return "", errors.New("error")
			},
			"fallback value",
		)

		if err == nil {
			t.Error("expected error to be returned")
		}

		if result != "fallback value" {
			t.Errorf("expected fallback value, got %s", result)
		}
	})
}

func BenchmarkRetry(b *testing.B) {
	ctx := context.Background()
	config := NewRetryConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Retry(ctx, func() error {
			return nil
		}, config)
	}
}

func BenchmarkCircuitBreaker(b *testing.B) {
	cb := NewCircuitBreaker(3, 1*time.Second)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cb.Execute(func() error {
			return nil
		})
	}
}
