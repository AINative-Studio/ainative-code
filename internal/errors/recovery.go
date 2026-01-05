package errors

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// RetryStrategy defines the strategy for retrying failed operations
type RetryStrategy interface {
	// ShouldRetry determines if an error should trigger a retry
	ShouldRetry(err error, attempt int) bool

	// GetDelay calculates the delay before the next retry attempt
	GetDelay(attempt int) time.Duration

	// MaxAttempts returns the maximum number of retry attempts
	MaxAttempts() int
}

// ExponentialBackoff implements exponential backoff retry strategy
type ExponentialBackoff struct {
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
	MaxRetries   int
}

// NewExponentialBackoff creates a new exponential backoff strategy with sensible defaults
func NewExponentialBackoff() *ExponentialBackoff {
	return &ExponentialBackoff{
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
		MaxRetries:   5,
	}
}

// ShouldRetry determines if the error is retryable
func (e *ExponentialBackoff) ShouldRetry(err error, attempt int) bool {
	if attempt >= e.MaxRetries {
		return false
	}
	return IsRetryable(err)
}

// GetDelay calculates the exponential backoff delay
func (e *ExponentialBackoff) GetDelay(attempt int) time.Duration {
	if attempt <= 0 {
		return e.InitialDelay
	}

	delay := float64(e.InitialDelay) * math.Pow(e.Multiplier, float64(attempt))
	if delay > float64(e.MaxDelay) {
		return e.MaxDelay
	}
	return time.Duration(delay)
}

// MaxAttempts returns the maximum number of retry attempts
func (e *ExponentialBackoff) MaxAttempts() int {
	return e.MaxRetries
}

// LinearBackoff implements linear backoff retry strategy
type LinearBackoff struct {
	Delay      time.Duration
	MaxRetries int
}

// NewLinearBackoff creates a new linear backoff strategy
func NewLinearBackoff(delay time.Duration, maxRetries int) *LinearBackoff {
	return &LinearBackoff{
		Delay:      delay,
		MaxRetries: maxRetries,
	}
}

// ShouldRetry determines if the error is retryable
func (l *LinearBackoff) ShouldRetry(err error, attempt int) bool {
	if attempt >= l.MaxRetries {
		return false
	}
	return IsRetryable(err)
}

// GetDelay returns a constant delay
func (l *LinearBackoff) GetDelay(attempt int) time.Duration {
	return l.Delay
}

// MaxAttempts returns the maximum number of retry attempts
func (l *LinearBackoff) MaxAttempts() int {
	return l.MaxRetries
}

// RetryConfig holds configuration for retry operations
type RetryConfig struct {
	Strategy      RetryStrategy
	OnRetry       func(attempt int, err error)
	OnFinalError  func(err error)
}

// NewRetryConfig creates a default retry configuration
func NewRetryConfig() *RetryConfig {
	return &RetryConfig{
		Strategy: NewExponentialBackoff(),
		OnRetry: func(attempt int, err error) {
			// Default: do nothing, can be overridden
		},
		OnFinalError: func(err error) {
			// Default: do nothing, can be overridden
		},
	}
}

// Retry executes a function with retry logic
func Retry(ctx context.Context, fn func() error, config *RetryConfig) error {
	if config == nil {
		config = NewRetryConfig()
	}

	var lastErr error
	for attempt := 0; attempt < config.Strategy.MaxAttempts(); attempt++ {
		// Execute the function
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if we should retry
		if !config.Strategy.ShouldRetry(err, attempt) {
			break
		}

		// Call retry callback
		if config.OnRetry != nil {
			config.OnRetry(attempt, err)
		}

		// Calculate delay for next retry
		delay := config.Strategy.GetDelay(attempt)

		// Wait with context cancellation support
		select {
		case <-ctx.Done():
			return fmt.Errorf("retry cancelled: %w", ctx.Err())
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	// All retries exhausted
	if config.OnFinalError != nil {
		config.OnFinalError(lastErr)
	}

	return Wrap(lastErr, ErrCodeToolExecutionFailed, "all retry attempts exhausted")
}

// RetryWithRecovery executes a function with retry and recovery logic
func RetryWithRecovery(ctx context.Context, fn func() error, config *RetryConfig, recovery func(error) error) error {
	err := Retry(ctx, fn, config)
	if err != nil && recovery != nil {
		// Attempt recovery
		return recovery(err)
	}
	return err
}

// CircuitBreaker implements the circuit breaker pattern for fault tolerance
type CircuitBreaker struct {
	maxFailures     int
	resetTimeout    time.Duration
	failures        int
	lastFailureTime time.Time
	state           CircuitState
}

// CircuitState represents the state of a circuit breaker
type CircuitState int

const (
	// StateClosed allows requests through
	StateClosed CircuitState = iota
	// StateOpen blocks requests
	StateOpen
	// StateHalfOpen allows limited requests to test recovery
	StateHalfOpen
)

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        StateClosed,
	}
}

// Execute runs a function through the circuit breaker
func (cb *CircuitBreaker) Execute(fn func() error) error {
	// Check if circuit should transition to half-open
	if cb.state == StateOpen && time.Since(cb.lastFailureTime) >= cb.resetTimeout {
		cb.state = StateHalfOpen
	}

	// Block requests if circuit is open
	if cb.state == StateOpen {
		return fmt.Errorf("circuit breaker is open: too many failures")
	}

	// Execute the function
	err := fn()

	if err != nil {
		cb.recordFailure()
		return err
	}

	// Success - reset circuit breaker
	cb.recordSuccess()
	return nil
}

// recordFailure records a failure and potentially opens the circuit
func (cb *CircuitBreaker) recordFailure() {
	cb.failures++
	cb.lastFailureTime = time.Now()

	if cb.failures >= cb.maxFailures {
		cb.state = StateOpen
	}
}

// recordSuccess records a successful call and resets the circuit
func (cb *CircuitBreaker) recordSuccess() {
	cb.failures = 0
	cb.state = StateClosed
}

// GetState returns the current circuit state
func (cb *CircuitBreaker) GetState() CircuitState {
	return cb.state
}

// Reset manually resets the circuit breaker
func (cb *CircuitBreaker) Reset() {
	cb.failures = 0
	cb.state = StateClosed
}

// Fallback executes a primary function and falls back to an alternative on error
func Fallback(primary func() error, fallback func() error) error {
	err := primary()
	if err != nil && fallback != nil {
		return fallback()
	}
	return err
}

// FallbackWithValue executes a primary function and returns a fallback value on error
func FallbackWithValue[T any](primary func() (T, error), fallbackValue T) (T, error) {
	result, err := primary()
	if err != nil {
		return fallbackValue, err
	}
	return result, nil
}

// ProviderRecoveryStrategy implements sophisticated recovery strategies for provider API failures
type ProviderRecoveryStrategy struct {
	MaxRetries          int
	InitialBackoff      time.Duration
	MaxBackoff          time.Duration
	Multiplier          float64
	EnableJitter        bool
	OnAPIKeyResolution  func(ctx context.Context) (string, error)
	OnTokenReduction    func(currentTokens int) int
	OnTimeoutIncrease   func(currentTimeout time.Duration) time.Duration
	Logger              func(message string)
}

// RecoveryAction represents an action to take for error recovery
type RecoveryAction int

const (
	// ActionRetry indicates a simple retry with backoff
	ActionRetry RecoveryAction = iota
	// ActionRetryWithBackoff indicates exponential backoff retry
	ActionRetryWithBackoff
	// ActionResolveAPIKey indicates re-resolution of API key before retry
	ActionResolveAPIKey
	// ActionReduceTokens indicates reducing max_tokens before retry
	ActionReduceTokens
	// ActionIncreaseTimeout indicates increasing timeout before retry
	ActionIncreaseTimeout
	// ActionCircuitBreak indicates circuit breaker activation
	ActionCircuitBreak
	// ActionFail indicates non-retryable error
	ActionFail
)

// RecoveryDecision contains the decision on how to handle an error
type RecoveryDecision struct {
	Action         RecoveryAction
	RetryAfter     time.Duration
	ShouldRetry    bool
	NewAPIKey      string
	TokenReduction int
	NewTimeout     time.Duration
	Message        string
}

// NewProviderRecoveryStrategy creates a new provider recovery strategy with defaults
func NewProviderRecoveryStrategy() *ProviderRecoveryStrategy {
	return &ProviderRecoveryStrategy{
		MaxRetries:     3,
		InitialBackoff: 1 * time.Second,
		MaxBackoff:     60 * time.Second,
		Multiplier:     2.0,
		EnableJitter:   true,
		OnTokenReduction: func(currentTokens int) int {
			// Reduce by 20%
			return int(float64(currentTokens) * 0.8)
		},
		OnTimeoutIncrease: func(currentTimeout time.Duration) time.Duration {
			// Increase by 50%
			return time.Duration(float64(currentTimeout) * 1.5)
		},
		Logger: func(message string) {
			// Default no-op logger
		},
	}
}

// AnalyzeError analyzes an error and returns a recovery decision
func (p *ProviderRecoveryStrategy) AnalyzeError(ctx context.Context, err error, attempt int, statusCode int, retryAfterHeader string) RecoveryDecision {
	// Check if max retries exceeded
	if attempt >= p.MaxRetries {
		return RecoveryDecision{
			Action:      ActionFail,
			ShouldRetry: false,
			Message:     fmt.Sprintf("max retries (%d) exceeded", p.MaxRetries),
		}
	}

	// Analyze based on status code
	switch statusCode {
	case http.StatusUnauthorized: // 401
		return p.handleUnauthorized(ctx, attempt)

	case http.StatusBadRequest: // 400
		return p.handleBadRequest(ctx, err, attempt)

	case http.StatusTooManyRequests: // 429
		return p.handleRateLimit(ctx, attempt, retryAfterHeader)

	case http.StatusInternalServerError, // 500
		http.StatusBadGateway,          // 502
		http.StatusServiceUnavailable:  // 503
		return p.handleServerError(ctx, attempt, statusCode)

	case http.StatusGatewayTimeout, http.StatusRequestTimeout: // 504, 408
		return p.handleTimeout(ctx, attempt)

	default:
		// Check for specific error types
		if strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "deadline exceeded") {
			return p.handleTimeout(ctx, attempt)
		}

		// Non-retryable error
		return RecoveryDecision{
			Action:      ActionFail,
			ShouldRetry: false,
			Message:     fmt.Sprintf("non-retryable error: status %d", statusCode),
		}
	}
}

// handleUnauthorized handles 401 Unauthorized errors
func (p *ProviderRecoveryStrategy) handleUnauthorized(ctx context.Context, attempt int) RecoveryDecision {
	decision := RecoveryDecision{
		Action:      ActionResolveAPIKey,
		ShouldRetry: true,
		RetryAfter:  p.calculateBackoff(attempt),
		Message:     "unauthorized - attempting to re-resolve API key",
	}

	// Attempt to re-resolve API key
	if p.OnAPIKeyResolution != nil {
		newKey, err := p.OnAPIKeyResolution(ctx)
		if err != nil {
			p.Logger(fmt.Sprintf("failed to re-resolve API key: %v", err))
			decision.Action = ActionFail
			decision.ShouldRetry = false
			decision.Message = "failed to re-resolve API key"
		} else {
			decision.NewAPIKey = newKey
			p.Logger("successfully re-resolved API key")
		}
	} else {
		// No API key resolution callback, fail
		decision.Action = ActionFail
		decision.ShouldRetry = false
		decision.Message = "no API key resolution callback configured"
	}

	return decision
}

// handleBadRequest handles 400 Bad Request errors
func (p *ProviderRecoveryStrategy) handleBadRequest(ctx context.Context, err error, attempt int) RecoveryDecision {
	errMsg := strings.ToLower(err.Error())

	// Check if it's a token limit error
	if strings.Contains(errMsg, "token") && (strings.Contains(errMsg, "limit") ||
		strings.Contains(errMsg, "exceeded") || strings.Contains(errMsg, "maximum") ||
		strings.Contains(errMsg, "too long") || strings.Contains(errMsg, "max_tokens")) {

		decision := RecoveryDecision{
			Action:      ActionReduceTokens,
			ShouldRetry: true,
			RetryAfter:  p.calculateBackoff(attempt),
			Message:     "token limit exceeded - reducing max_tokens by 20%",
		}

		// Calculate new token count if callback is provided
		if p.OnTokenReduction != nil {
			// We don't have the current token count here, but the callback will handle it
			p.Logger("reducing token count for retry")
		}

		return decision
	}

	// Other 400 errors are typically not retryable (validation errors, etc.)
	return RecoveryDecision{
		Action:      ActionFail,
		ShouldRetry: false,
		Message:     "bad request - validation error",
	}
}

// handleRateLimit handles 429 Rate Limited errors
func (p *ProviderRecoveryStrategy) handleRateLimit(ctx context.Context, attempt int, retryAfterHeader string) RecoveryDecision {
	decision := RecoveryDecision{
		Action:      ActionRetryWithBackoff,
		ShouldRetry: true,
		Message:     "rate limited - applying exponential backoff",
	}

	// Parse Retry-After header if present
	if retryAfterHeader != "" {
		if seconds, err := strconv.Atoi(retryAfterHeader); err == nil {
			decision.RetryAfter = time.Duration(seconds) * time.Second
			decision.Message = fmt.Sprintf("rate limited - respecting Retry-After: %d seconds", seconds)
		} else if retryTime, err := http.ParseTime(retryAfterHeader); err == nil {
			decision.RetryAfter = time.Until(retryTime)
			decision.Message = fmt.Sprintf("rate limited - waiting until %v", retryTime)
		}
	}

	// If no Retry-After header, use exponential backoff
	if decision.RetryAfter == 0 {
		decision.RetryAfter = p.calculateBackoff(attempt)
	}

	// Cap at max backoff
	if decision.RetryAfter > p.MaxBackoff {
		decision.RetryAfter = p.MaxBackoff
	}

	p.Logger(fmt.Sprintf("rate limit retry after %v", decision.RetryAfter))

	return decision
}

// handleServerError handles 500/502/503 server errors
func (p *ProviderRecoveryStrategy) handleServerError(ctx context.Context, attempt int, statusCode int) RecoveryDecision {
	decision := RecoveryDecision{
		Action:      ActionRetryWithBackoff,
		ShouldRetry: true,
		RetryAfter:  p.calculateBackoff(attempt),
		Message:     fmt.Sprintf("server error %d - retrying with exponential backoff", statusCode),
	}

	p.Logger(fmt.Sprintf("server error %d, retry after %v", statusCode, decision.RetryAfter))

	return decision
}

// handleTimeout handles timeout errors
func (p *ProviderRecoveryStrategy) handleTimeout(ctx context.Context, attempt int) RecoveryDecision {
	decision := RecoveryDecision{
		Action:      ActionIncreaseTimeout,
		ShouldRetry: true,
		RetryAfter:  p.calculateBackoff(attempt),
		Message:     "timeout - increasing timeout and retrying",
	}

	// Increase timeout if callback is provided
	if p.OnTimeoutIncrease != nil {
		p.Logger("increasing timeout for retry")
	}

	return decision
}

// calculateBackoff calculates exponential backoff with optional jitter
func (p *ProviderRecoveryStrategy) calculateBackoff(attempt int) time.Duration {
	// Calculate base exponential backoff
	backoff := float64(p.InitialBackoff) * math.Pow(p.Multiplier, float64(attempt))

	// Cap at max backoff
	if backoff > float64(p.MaxBackoff) {
		backoff = float64(p.MaxBackoff)
	}

	// Apply jitter if enabled (±10% randomization)
	if p.EnableJitter {
		jitter := 0.9 + (0.2 * rand.Float64())
		backoff = backoff * jitter
	}

	return time.Duration(backoff)
}

// ExecuteWithRecovery executes a function with sophisticated error recovery
func (p *ProviderRecoveryStrategy) ExecuteWithRecovery(
	ctx context.Context,
	fn func(ctx context.Context) error,
	getStatusCode func() int,
	getRetryAfter func() string,
) error {
	var lastErr error

	for attempt := 0; attempt <= p.MaxRetries; attempt++ {
		// Check context cancellation
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("execution cancelled: %w", err)
		}

		// Log attempt
		if attempt > 0 {
			p.Logger(fmt.Sprintf("retry attempt %d/%d", attempt, p.MaxRetries))
		}

		// Execute the function
		err := fn(ctx)
		if err == nil {
			return nil // Success
		}

		lastErr = err

		// Analyze the error and get recovery decision
		statusCode := getStatusCode()
		retryAfter := getRetryAfter()
		decision := p.AnalyzeError(ctx, err, attempt, statusCode, retryAfter)

		// Log decision
		p.Logger(fmt.Sprintf("recovery decision: %s", decision.Message))

		// Check if we should retry
		if !decision.ShouldRetry {
			return fmt.Errorf("non-retryable error after %d attempts: %w", attempt+1, lastErr)
		}

		// Don't sleep after the last attempt
		if attempt < p.MaxRetries {
			// Wait with context cancellation support
			select {
			case <-time.After(decision.RetryAfter):
				// Continue to next retry
			case <-ctx.Done():
				return fmt.Errorf("execution cancelled during backoff: %w", ctx.Err())
			}
		}
	}

	return fmt.Errorf("all recovery attempts exhausted after %d retries: %w", p.MaxRetries, lastErr)
}

// AdvancedCircuitBreaker extends the basic circuit breaker with error-specific behavior
type AdvancedCircuitBreaker struct {
	*CircuitBreaker
	errorThresholds map[string]int // Different thresholds for different error types
	errorCounts     map[string]int // Track counts by error type
}

// NewAdvancedCircuitBreaker creates an advanced circuit breaker
func NewAdvancedCircuitBreaker(maxFailures int, resetTimeout time.Duration) *AdvancedCircuitBreaker {
	return &AdvancedCircuitBreaker{
		CircuitBreaker:  NewCircuitBreaker(maxFailures, resetTimeout),
		errorThresholds: make(map[string]int),
		errorCounts:     make(map[string]int),
	}
}

// SetErrorThreshold sets a specific threshold for an error type
func (acb *AdvancedCircuitBreaker) SetErrorThreshold(errorType string, threshold int) {
	acb.errorThresholds[errorType] = threshold
}

// ExecuteWithErrorType executes a function with error-type-specific circuit breaking
func (acb *AdvancedCircuitBreaker) ExecuteWithErrorType(fn func() error, errorType string) error {
	// Check circuit state
	err := acb.Execute(fn)

	// Track error-specific counts
	if err != nil {
		acb.errorCounts[errorType]++

		// Check if error-specific threshold is exceeded
		if threshold, exists := acb.errorThresholds[errorType]; exists {
			if acb.errorCounts[errorType] >= threshold {
				acb.state = StateOpen
				return fmt.Errorf("circuit breaker opened due to %s errors: %w", errorType, err)
			}
		}
	} else {
		// Reset error-specific count on success
		acb.errorCounts[errorType] = 0
	}

	return err
}
