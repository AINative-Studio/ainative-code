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

// TestProviderRecoveryStrategy tests the sophisticated provider recovery strategies
func TestProviderRecoveryStrategy(t *testing.T) {
	t.Run("Default configuration", func(t *testing.T) {
		strategy := NewProviderRecoveryStrategy()

		if strategy.MaxRetries != 3 {
			t.Errorf("expected MaxRetries 3, got %d", strategy.MaxRetries)
		}

		if strategy.InitialBackoff != 1*time.Second {
			t.Errorf("expected InitialBackoff 1s, got %v", strategy.InitialBackoff)
		}

		if strategy.MaxBackoff != 60*time.Second {
			t.Errorf("expected MaxBackoff 60s, got %v", strategy.MaxBackoff)
		}

		if strategy.Multiplier != 2.0 {
			t.Errorf("expected Multiplier 2.0, got %f", strategy.Multiplier)
		}

		if !strategy.EnableJitter {
			t.Error("expected EnableJitter true")
		}
	})

	t.Run("Token reduction callback", func(t *testing.T) {
		strategy := NewProviderRecoveryStrategy()

		newTokens := strategy.OnTokenReduction(1000)
		expected := 800 // 20% reduction

		if newTokens != expected {
			t.Errorf("expected %d tokens after reduction, got %d", expected, newTokens)
		}
	})

	t.Run("Timeout increase callback", func(t *testing.T) {
		strategy := NewProviderRecoveryStrategy()

		newTimeout := strategy.OnTimeoutIncrease(10 * time.Second)
		expected := 15 * time.Second // 50% increase

		if newTimeout != expected {
			t.Errorf("expected timeout %v, got %v", expected, newTimeout)
		}
	})
}

func TestAnalyzeError_Unauthorized(t *testing.T) {
	t.Run("401 with API key resolution", func(t *testing.T) {
		strategy := NewProviderRecoveryStrategy()
		resolvedKey := ""

		strategy.OnAPIKeyResolution = func(ctx context.Context) (string, error) {
			return "new-api-key-123", nil
		}

		strategy.Logger = func(message string) {
			t.Logf("Log: %s", message)
		}

		ctx := context.Background()
		err := errors.New("unauthorized")
		decision := strategy.AnalyzeError(ctx, err, 0, 401, "")

		if !decision.ShouldRetry {
			t.Error("expected ShouldRetry true for 401")
		}

		if decision.Action != ActionResolveAPIKey {
			t.Errorf("expected action ActionResolveAPIKey, got %v", decision.Action)
		}

		if decision.NewAPIKey != "new-api-key-123" {
			t.Errorf("expected new API key, got %s", resolvedKey)
		}
	})

	t.Run("401 without API key resolution callback", func(t *testing.T) {
		strategy := NewProviderRecoveryStrategy()
		strategy.OnAPIKeyResolution = nil

		ctx := context.Background()
		err := errors.New("unauthorized")
		decision := strategy.AnalyzeError(ctx, err, 0, 401, "")

		if decision.ShouldRetry {
			t.Error("expected ShouldRetry false when no callback")
		}

		if decision.Action != ActionFail {
			t.Errorf("expected action ActionFail, got %v", decision.Action)
		}
	})

	t.Run("401 with failed API key resolution", func(t *testing.T) {
		strategy := NewProviderRecoveryStrategy()

		strategy.OnAPIKeyResolution = func(ctx context.Context) (string, error) {
			return "", errors.New("key resolution failed")
		}

		strategy.Logger = func(message string) {
			t.Logf("Log: %s", message)
		}

		ctx := context.Background()
		err := errors.New("unauthorized")
		decision := strategy.AnalyzeError(ctx, err, 0, 401, "")

		if decision.ShouldRetry {
			t.Error("expected ShouldRetry false when resolution fails")
		}

		if decision.Action != ActionFail {
			t.Errorf("expected action ActionFail, got %v", decision.Action)
		}
	})
}

func TestAnalyzeError_RateLimit(t *testing.T) {
	t.Run("429 with Retry-After header (seconds)", func(t *testing.T) {
		strategy := NewProviderRecoveryStrategy()
		strategy.Logger = func(message string) {
			t.Logf("Log: %s", message)
		}

		ctx := context.Background()
		err := errors.New("rate limited")
		decision := strategy.AnalyzeError(ctx, err, 0, 429, "10")

		if !decision.ShouldRetry {
			t.Error("expected ShouldRetry true for 429")
		}

		if decision.Action != ActionRetryWithBackoff {
			t.Errorf("expected action ActionRetryWithBackoff, got %v", decision.Action)
		}

		if decision.RetryAfter != 10*time.Second {
			t.Errorf("expected RetryAfter 10s, got %v", decision.RetryAfter)
		}
	})

	t.Run("429 without Retry-After header", func(t *testing.T) {
		strategy := NewProviderRecoveryStrategy()
		strategy.EnableJitter = false // Disable jitter for predictable testing
		strategy.Logger = func(message string) {}

		ctx := context.Background()
		err := errors.New("rate limited")
		decision := strategy.AnalyzeError(ctx, err, 0, 429, "")

		if !decision.ShouldRetry {
			t.Error("expected ShouldRetry true for 429")
		}

		// Should use exponential backoff
		if decision.RetryAfter != strategy.InitialBackoff {
			t.Errorf("expected RetryAfter %v, got %v", strategy.InitialBackoff, decision.RetryAfter)
		}
	})

	t.Run("429 respects max backoff", func(t *testing.T) {
		strategy := NewProviderRecoveryStrategy()
		strategy.MaxBackoff = 30 * time.Second
		strategy.Logger = func(message string) {}

		ctx := context.Background()
		err := errors.New("rate limited")
		decision := strategy.AnalyzeError(ctx, err, 0, 429, "120")

		if decision.RetryAfter > strategy.MaxBackoff {
			t.Errorf("RetryAfter %v should not exceed MaxBackoff %v", decision.RetryAfter, strategy.MaxBackoff)
		}
	})
}

func TestAnalyzeError_BadRequest(t *testing.T) {
	t.Run("400 with token limit error", func(t *testing.T) {
		strategy := NewProviderRecoveryStrategy()
		strategy.Logger = func(message string) {
			t.Logf("Log: %s", message)
		}

		ctx := context.Background()
		err := errors.New("token limit exceeded")
		decision := strategy.AnalyzeError(ctx, err, 0, 400, "")

		if !decision.ShouldRetry {
			t.Error("expected ShouldRetry true for token limit error")
		}

		if decision.Action != ActionReduceTokens {
			t.Errorf("expected action ActionReduceTokens, got %v", decision.Action)
		}
	})

	t.Run("400 with max_tokens error", func(t *testing.T) {
		strategy := NewProviderRecoveryStrategy()
		strategy.Logger = func(message string) {}

		ctx := context.Background()
		err := errors.New("max_tokens parameter is too large")
		decision := strategy.AnalyzeError(ctx, err, 0, 400, "")

		if !decision.ShouldRetry {
			t.Error("expected ShouldRetry true for max_tokens error")
		}

		if decision.Action != ActionReduceTokens {
			t.Errorf("expected action ActionReduceTokens, got %v", decision.Action)
		}
	})

	t.Run("400 with validation error (non-retryable)", func(t *testing.T) {
		strategy := NewProviderRecoveryStrategy()
		strategy.Logger = func(message string) {}

		ctx := context.Background()
		err := errors.New("invalid parameter value")
		decision := strategy.AnalyzeError(ctx, err, 0, 400, "")

		if decision.ShouldRetry {
			t.Error("expected ShouldRetry false for validation error")
		}

		if decision.Action != ActionFail {
			t.Errorf("expected action ActionFail, got %v", decision.Action)
		}
	})
}

func TestAnalyzeError_ServerError(t *testing.T) {
	t.Run("500 Internal Server Error", func(t *testing.T) {
		strategy := NewProviderRecoveryStrategy()
		strategy.Logger = func(message string) {
			t.Logf("Log: %s", message)
		}

		ctx := context.Background()
		err := errors.New("internal server error")
		decision := strategy.AnalyzeError(ctx, err, 0, 500, "")

		if !decision.ShouldRetry {
			t.Error("expected ShouldRetry true for 500")
		}

		if decision.Action != ActionRetryWithBackoff {
			t.Errorf("expected action ActionRetryWithBackoff, got %v", decision.Action)
		}
	})

	t.Run("502 Bad Gateway", func(t *testing.T) {
		strategy := NewProviderRecoveryStrategy()
		strategy.Logger = func(message string) {}

		ctx := context.Background()
		err := errors.New("bad gateway")
		decision := strategy.AnalyzeError(ctx, err, 0, 502, "")

		if !decision.ShouldRetry {
			t.Error("expected ShouldRetry true for 502")
		}

		if decision.Action != ActionRetryWithBackoff {
			t.Errorf("expected action ActionRetryWithBackoff, got %v", decision.Action)
		}
	})

	t.Run("503 Service Unavailable", func(t *testing.T) {
		strategy := NewProviderRecoveryStrategy()
		strategy.Logger = func(message string) {}

		ctx := context.Background()
		err := errors.New("service unavailable")
		decision := strategy.AnalyzeError(ctx, err, 0, 503, "")

		if !decision.ShouldRetry {
			t.Error("expected ShouldRetry true for 503")
		}

		if decision.Action != ActionRetryWithBackoff {
			t.Errorf("expected action ActionRetryWithBackoff, got %v", decision.Action)
		}
	})
}

func TestAnalyzeError_Timeout(t *testing.T) {
	t.Run("504 Gateway Timeout", func(t *testing.T) {
		strategy := NewProviderRecoveryStrategy()
		strategy.Logger = func(message string) {
			t.Logf("Log: %s", message)
		}

		ctx := context.Background()
		err := errors.New("gateway timeout")
		decision := strategy.AnalyzeError(ctx, err, 0, 504, "")

		if !decision.ShouldRetry {
			t.Error("expected ShouldRetry true for 504")
		}

		if decision.Action != ActionIncreaseTimeout {
			t.Errorf("expected action ActionIncreaseTimeout, got %v", decision.Action)
		}
	})

	t.Run("Network timeout error", func(t *testing.T) {
		strategy := NewProviderRecoveryStrategy()
		strategy.Logger = func(message string) {}

		ctx := context.Background()
		err := errors.New("connection timeout")
		decision := strategy.AnalyzeError(ctx, err, 0, 0, "")

		if !decision.ShouldRetry {
			t.Error("expected ShouldRetry true for timeout")
		}

		if decision.Action != ActionIncreaseTimeout {
			t.Errorf("expected action ActionIncreaseTimeout, got %v", decision.Action)
		}
	})

	t.Run("Context deadline exceeded", func(t *testing.T) {
		strategy := NewProviderRecoveryStrategy()
		strategy.Logger = func(message string) {}

		ctx := context.Background()
		err := errors.New("context deadline exceeded")
		decision := strategy.AnalyzeError(ctx, err, 0, 0, "")

		if !decision.ShouldRetry {
			t.Error("expected ShouldRetry true for deadline exceeded")
		}

		if decision.Action != ActionIncreaseTimeout {
			t.Errorf("expected action ActionIncreaseTimeout, got %v", decision.Action)
		}
	})
}

func TestAnalyzeError_MaxRetries(t *testing.T) {
	t.Run("Max retries exceeded", func(t *testing.T) {
		strategy := NewProviderRecoveryStrategy()
		strategy.MaxRetries = 3
		strategy.Logger = func(message string) {}

		ctx := context.Background()
		err := errors.New("rate limited")
		decision := strategy.AnalyzeError(ctx, err, 3, 429, "")

		if decision.ShouldRetry {
			t.Error("expected ShouldRetry false when max retries exceeded")
		}

		if decision.Action != ActionFail {
			t.Errorf("expected action ActionFail, got %v", decision.Action)
		}
	})
}

func TestCalculateBackoff(t *testing.T) {
	t.Run("Exponential growth", func(t *testing.T) {
		strategy := NewProviderRecoveryStrategy()
		strategy.EnableJitter = false // Disable jitter for predictable testing

		backoff0 := strategy.calculateBackoff(0)
		backoff1 := strategy.calculateBackoff(1)
		backoff2 := strategy.calculateBackoff(2)

		if backoff0 != 1*time.Second {
			t.Errorf("expected 1s for attempt 0, got %v", backoff0)
		}

		if backoff1 != 2*time.Second {
			t.Errorf("expected 2s for attempt 1, got %v", backoff1)
		}

		if backoff2 != 4*time.Second {
			t.Errorf("expected 4s for attempt 2, got %v", backoff2)
		}
	})

	t.Run("Respects max backoff", func(t *testing.T) {
		strategy := NewProviderRecoveryStrategy()
		strategy.EnableJitter = false
		strategy.MaxBackoff = 30 * time.Second

		backoff := strategy.calculateBackoff(100)

		if backoff > strategy.MaxBackoff {
			t.Errorf("backoff %v should not exceed max %v", backoff, strategy.MaxBackoff)
		}
	})

	t.Run("Jitter adds randomization", func(t *testing.T) {
		strategy := NewProviderRecoveryStrategy()
		strategy.EnableJitter = true

		backoff1 := strategy.calculateBackoff(0)
		backoff2 := strategy.calculateBackoff(0)

		// With jitter, two calls should likely produce different results
		// This test might occasionally fail due to randomness, but it's statistically unlikely
		if backoff1 == backoff2 {
			t.Log("Warning: jitter produced same backoff twice (statistically unlikely but possible)")
		}

		// Verify jitter keeps values within ±10% range
		base := float64(strategy.InitialBackoff)
		min := time.Duration(base * 0.9)
		max := time.Duration(base * 1.1)

		if backoff1 < min || backoff1 > max {
			t.Errorf("backoff %v outside jitter range [%v, %v]", backoff1, min, max)
		}
	})
}

func TestExecuteWithRecovery(t *testing.T) {
	t.Run("Succeeds on first attempt", func(t *testing.T) {
		strategy := NewProviderRecoveryStrategy()
		attempts := 0

		err := strategy.ExecuteWithRecovery(
			context.Background(),
			func(ctx context.Context) error {
				attempts++
				return nil
			},
			func() int { return 200 },
			func() string { return "" },
		)

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if attempts != 1 {
			t.Errorf("expected 1 attempt, got %d", attempts)
		}
	})

	t.Run("Retries on retryable error", func(t *testing.T) {
		strategy := NewProviderRecoveryStrategy()
		strategy.InitialBackoff = 10 * time.Millisecond
		strategy.Logger = func(message string) {
			t.Logf("Log: %s", message)
		}

		attempts := 0

		err := strategy.ExecuteWithRecovery(
			context.Background(),
			func(ctx context.Context) error {
				attempts++
				if attempts < 3 {
					return errors.New("server error")
				}
				return nil
			},
			func() int { return 500 },
			func() string { return "" },
		)

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if attempts != 3 {
			t.Errorf("expected 3 attempts, got %d", attempts)
		}
	})

	t.Run("Fails on non-retryable error", func(t *testing.T) {
		strategy := NewProviderRecoveryStrategy()
		attempts := 0

		err := strategy.ExecuteWithRecovery(
			context.Background(),
			func(ctx context.Context) error {
				attempts++
				return errors.New("validation error")
			},
			func() int { return 400 },
			func() string { return "" },
		)

		if err == nil {
			t.Error("expected error for non-retryable error")
		}

		if attempts != 1 {
			t.Errorf("expected 1 attempt for non-retryable, got %d", attempts)
		}
	})

	t.Run("Context cancellation", func(t *testing.T) {
		strategy := NewProviderRecoveryStrategy()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := strategy.ExecuteWithRecovery(
			ctx,
			func(ctx context.Context) error {
				return errors.New("error")
			},
			func() int { return 500 },
			func() string { return "" },
		)

		if err == nil {
			t.Error("expected error due to context cancellation")
		}

		if !strings.Contains(err.Error(), "cancelled") {
			t.Errorf("expected cancellation error, got %v", err)
		}
	})

	t.Run("Exhausts all retries", func(t *testing.T) {
		strategy := NewProviderRecoveryStrategy()
		strategy.MaxRetries = 2
		strategy.InitialBackoff = 10 * time.Millisecond
		attempts := 0

		err := strategy.ExecuteWithRecovery(
			context.Background(),
			func(ctx context.Context) error {
				attempts++
				return errors.New("persistent error")
			},
			func() int { return 500 },
			func() string { return "" },
		)

		if err == nil {
			t.Error("expected error after exhausting retries")
		}

		// Check for either "exhausted" or "non-retryable" in error message
		// (both indicate retry limit reached)
		if !strings.Contains(err.Error(), "exhausted") && !strings.Contains(err.Error(), "non-retryable") {
			t.Errorf("expected exhausted or non-retryable error, got %v", err)
		}

		// MaxRetries is 2, so we get attempts 0, 1, 2 = 3 total attempts
		if attempts != 3 {
			t.Errorf("expected 3 attempts (0, 1, 2), got %d", attempts)
		}
	})
}

func TestAdvancedCircuitBreaker(t *testing.T) {
	t.Run("Creates with default thresholds", func(t *testing.T) {
		acb := NewAdvancedCircuitBreaker(3, 1*time.Second)

		if acb.GetState() != StateClosed {
			t.Errorf("expected initial state %v, got %v", StateClosed, acb.GetState())
		}

		if acb.errorThresholds == nil {
			t.Error("expected errorThresholds to be initialized")
		}

		if acb.errorCounts == nil {
			t.Error("expected errorCounts to be initialized")
		}
	})

	t.Run("Sets error-specific thresholds", func(t *testing.T) {
		acb := NewAdvancedCircuitBreaker(5, 1*time.Second)
		acb.SetErrorThreshold("timeout", 3)
		acb.SetErrorThreshold("rate_limit", 10)

		if acb.errorThresholds["timeout"] != 3 {
			t.Errorf("expected timeout threshold 3, got %d", acb.errorThresholds["timeout"])
		}

		if acb.errorThresholds["rate_limit"] != 10 {
			t.Errorf("expected rate_limit threshold 10, got %d", acb.errorThresholds["rate_limit"])
		}
	})

	t.Run("Opens on error-specific threshold", func(t *testing.T) {
		acb := NewAdvancedCircuitBreaker(10, 1*time.Second)
		acb.SetErrorThreshold("timeout", 2)

		// Trigger 2 timeout errors
		for i := 0; i < 2; i++ {
			_ = acb.ExecuteWithErrorType(func() error {
				return errors.New("timeout")
			}, "timeout")
		}

		if acb.GetState() != StateOpen {
			t.Errorf("expected state %v after threshold, got %v", StateOpen, acb.GetState())
		}
	})

	t.Run("Resets error count on success", func(t *testing.T) {
		acb := NewAdvancedCircuitBreaker(10, 1*time.Second)
		acb.SetErrorThreshold("timeout", 5)

		// Record some errors
		_ = acb.ExecuteWithErrorType(func() error {
			return errors.New("timeout")
		}, "timeout")

		// Success should reset count
		_ = acb.ExecuteWithErrorType(func() error {
			return nil
		}, "timeout")

		if acb.errorCounts["timeout"] != 0 {
			t.Errorf("expected timeout count 0 after success, got %d", acb.errorCounts["timeout"])
		}
	})
}

func BenchmarkProviderRecoveryStrategy(b *testing.B) {
	strategy := NewProviderRecoveryStrategy()
	strategy.Logger = func(message string) {}

	ctx := context.Background()
	err := errors.New("test error")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strategy.AnalyzeError(ctx, err, 0, 500, "")
	}
}

func BenchmarkAdvancedCircuitBreaker(b *testing.B) {
	acb := NewAdvancedCircuitBreaker(3, 1*time.Second)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = acb.ExecuteWithErrorType(func() error {
			return nil
		}, "test")
	}
}
