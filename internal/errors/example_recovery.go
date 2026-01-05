package errors

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// Example: Basic usage of ProviderRecoveryStrategy
func ExampleProviderRecoveryStrategy_basic() {
	// Create a new recovery strategy with defaults
	strategy := NewProviderRecoveryStrategy()

	// Configure logging
	strategy.Logger = func(message string) {
		fmt.Println("Recovery:", message)
	}

	// Example API call function
	makeAPICall := func(ctx context.Context) error {
		// Your API call logic here
		return nil
	}

	// Execute with recovery
	err := strategy.ExecuteWithRecovery(
		context.Background(),
		makeAPICall,
		func() int { return http.StatusOK },
		func() string { return "" },
	)

	if err != nil {
		fmt.Printf("API call failed: %v\n", err)
	}
}

// Example: Custom retry configuration
func ExampleProviderRecoveryStrategy_customConfig() {
	strategy := &ProviderRecoveryStrategy{
		MaxRetries:     5,                      // More retries
		InitialBackoff: 500 * time.Millisecond, // Shorter initial backoff
		MaxBackoff:     30 * time.Second,       // Lower max backoff
		Multiplier:     1.5,                    // Less aggressive exponential growth
		EnableJitter:   true,
		Logger:         func(msg string) { fmt.Println(msg) },
	}
	_ = strategy // Use the strategy variable

	// Use the custom strategy...
}

// Example: API key re-resolution on 401 errors
func ExampleProviderRecoveryStrategy_apiKeyResolution() {
	strategy := NewProviderRecoveryStrategy()

	// Configure API key resolution callback
	strategy.OnAPIKeyResolution = func(ctx context.Context) (string, error) {
		// Fetch a new API key from your key management system
		newKey := "new-api-key-from-vault"
		fmt.Println("Re-resolved API key")
		return newKey, nil
	}

	strategy.Logger = func(message string) {
		fmt.Println("Recovery:", message)
	}

	// Analyze a 401 error
	decision := strategy.AnalyzeError(
		context.Background(),
		fmt.Errorf("unauthorized"),
		0, // attempt number
		http.StatusUnauthorized,
		"", // retry-after header
	)

	fmt.Printf("Should retry: %v\n", decision.ShouldRetry)
	fmt.Printf("Action: %v\n", decision.Action)
	fmt.Printf("New API key: %s\n", decision.NewAPIKey)
}

// Example: Token reduction on token limit errors
func ExampleProviderRecoveryStrategy_tokenReduction() {
	strategy := NewProviderRecoveryStrategy()

	// The default token reduction callback reduces by 20%
	currentTokens := 1000
	newTokens := strategy.OnTokenReduction(currentTokens)
	fmt.Printf("Reduced tokens from %d to %d\n", currentTokens, newTokens)

	// Analyze a token limit error
	decision := strategy.AnalyzeError(
		context.Background(),
		fmt.Errorf("token limit exceeded"),
		0,
		http.StatusBadRequest,
		"",
	)

	fmt.Printf("Should retry: %v\n", decision.ShouldRetry)
	fmt.Printf("Action: %v\n", decision.Action)
}

// Example: Timeout increase on timeout errors
func ExampleProviderRecoveryStrategy_timeoutIncrease() {
	strategy := NewProviderRecoveryStrategy()

	// The default timeout increase callback increases by 50%
	currentTimeout := 10 * time.Second
	newTimeout := strategy.OnTimeoutIncrease(currentTimeout)
	fmt.Printf("Increased timeout from %v to %v\n", currentTimeout, newTimeout)

	// Analyze a timeout error
	decision := strategy.AnalyzeError(
		context.Background(),
		fmt.Errorf("gateway timeout"),
		0,
		http.StatusGatewayTimeout,
		"",
	)

	fmt.Printf("Should retry: %v\n", decision.ShouldRetry)
	fmt.Printf("Action: %v\n", decision.Action)
}

// Example: Rate limiting with Retry-After header
func ExampleProviderRecoveryStrategy_rateLimit() {
	strategy := NewProviderRecoveryStrategy()
	strategy.Logger = func(message string) {
		fmt.Println("Recovery:", message)
	}

	// Analyze a 429 rate limit error with Retry-After header
	decision := strategy.AnalyzeError(
		context.Background(),
		fmt.Errorf("rate limited"),
		0,
		http.StatusTooManyRequests,
		"30", // Retry after 30 seconds
	)

	fmt.Printf("Should retry: %v\n", decision.ShouldRetry)
	fmt.Printf("Retry after: %v\n", decision.RetryAfter)
}

// Example: Server error with exponential backoff
func ExampleProviderRecoveryStrategy_serverError() {
	strategy := NewProviderRecoveryStrategy()
	strategy.EnableJitter = false // Disable for predictable example

	// Analyze server errors at different attempts
	for attempt := 0; attempt < 3; attempt++ {
		decision := strategy.AnalyzeError(
			context.Background(),
			fmt.Errorf("internal server error"),
			attempt,
			http.StatusInternalServerError,
			"",
		)

		fmt.Printf("Attempt %d: Retry after %v\n", attempt, decision.RetryAfter)
	}
	// Output:
	// Attempt 0: Retry after 1s
	// Attempt 1: Retry after 2s
	// Attempt 2: Retry after 4s
}

// Example: Advanced circuit breaker with error-specific thresholds
func ExampleAdvancedCircuitBreaker_usage() {
	// Create circuit breaker
	acb := NewAdvancedCircuitBreaker(5, 30*time.Second)

	// Set different thresholds for different error types
	acb.SetErrorThreshold("timeout", 3)      // Open after 3 timeouts
	acb.SetErrorThreshold("rate_limit", 10)  // Open after 10 rate limits

	// Execute operations with error-type tracking
	err := acb.ExecuteWithErrorType(func() error {
		// Your operation that might timeout
		return fmt.Errorf("timeout")
	}, "timeout")

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		fmt.Printf("Circuit state: %v\n", acb.GetState())
	}
}

// Example: Complete provider integration pattern
func ExampleProviderRecoveryStrategy_integration() {
	// Create recovery strategy
	strategy := NewProviderRecoveryStrategy()

	// Track current state for recovery
	currentAPIKey := "original-api-key"
	currentMaxTokens := 1000
	currentTimeout := 30 * time.Second

	// Configure callbacks
	strategy.OnAPIKeyResolution = func(ctx context.Context) (string, error) {
		// Re-resolve from config or vault
		return "new-api-key", nil
	}

	strategy.Logger = func(message string) {
		fmt.Printf("[Provider Recovery] %s\n", message)
	}

	// Execute API call with recovery
	ctx := context.Background()

	var lastStatusCode int
	var lastRetryAfter string

	makeAPICall := func(ctx context.Context) error {
		// Simulate API call
		// In real usage, this would make an HTTP request
		lastStatusCode = http.StatusTooManyRequests
		lastRetryAfter = "5"
		return fmt.Errorf("rate limited")
	}

	err := strategy.ExecuteWithRecovery(
		ctx,
		makeAPICall,
		func() int { return lastStatusCode },
		func() string { return lastRetryAfter },
	)

	if err != nil {
		fmt.Printf("Final error: %v\n", err)
	}

	// Apply recovery actions based on decisions
	decision := strategy.AnalyzeError(ctx, err, 0, lastStatusCode, lastRetryAfter)

	switch decision.Action {
	case ActionResolveAPIKey:
		currentAPIKey = decision.NewAPIKey
		fmt.Printf("Updated API key: %s\n", currentAPIKey)

	case ActionReduceTokens:
		currentMaxTokens = strategy.OnTokenReduction(currentMaxTokens)
		fmt.Printf("Reduced max tokens to: %d\n", currentMaxTokens)

	case ActionIncreaseTimeout:
		currentTimeout = strategy.OnTimeoutIncrease(currentTimeout)
		fmt.Printf("Increased timeout to: %v\n", currentTimeout)
	}
}

// Example: Configuration-based setup
func ExampleProviderRecoveryStrategy_fromConfig() {
	// Example showing how to create strategy from config
	// This would typically be created from your config.RetryConfig

	type RetryConfig struct {
		MaxRetries             int
		InitialBackoff         time.Duration
		MaxBackoff             time.Duration
		Multiplier             float64
		EnableJitter           bool
		EnableAPIKeyResolution bool
		EnableTokenReduction   bool
		TokenReductionPercent  int
		EnableTimeoutIncrease  bool
		TimeoutIncreasePercent int
	}

	cfg := RetryConfig{
		MaxRetries:             3,
		InitialBackoff:         1 * time.Second,
		MaxBackoff:             60 * time.Second,
		Multiplier:             2.0,
		EnableJitter:           true,
		EnableAPIKeyResolution: true,
		EnableTokenReduction:   true,
		TokenReductionPercent:  20,
		EnableTimeoutIncrease:  true,
		TimeoutIncreasePercent: 50,
	}

	strategy := &ProviderRecoveryStrategy{
		MaxRetries:     cfg.MaxRetries,
		InitialBackoff: cfg.InitialBackoff,
		MaxBackoff:     cfg.MaxBackoff,
		Multiplier:     cfg.Multiplier,
		EnableJitter:   cfg.EnableJitter,
	}

	if cfg.EnableTokenReduction {
		strategy.OnTokenReduction = func(currentTokens int) int {
			reduction := float64(cfg.TokenReductionPercent) / 100.0
			return int(float64(currentTokens) * (1.0 - reduction))
		}
	}

	if cfg.EnableTimeoutIncrease {
		strategy.OnTimeoutIncrease = func(currentTimeout time.Duration) time.Duration {
			increase := float64(cfg.TimeoutIncreasePercent) / 100.0
			return time.Duration(float64(currentTimeout) * (1.0 + increase))
		}
	}

	fmt.Printf("Strategy configured with max retries: %d\n", strategy.MaxRetries)
}
