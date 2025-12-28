package errors_test

import (
	"context"
	"fmt"
	"time"

	"github.com/AINative-studio/ainative-code/internal/errors"
)

// Example demonstrates basic error creation and handling
func Example_basicUsage() {
	// Create a configuration error
	err := errors.NewConfigMissingError("api_key")

	// Get technical error message (for logging)
	fmt.Println("Technical:", err.Error())

	// Get user-friendly message (for UI display)
	fmt.Println("User:", err.UserMessage())

	// Check error properties
	fmt.Println("Code:", err.Code())
	fmt.Println("Severity:", err.Severity())
	fmt.Println("Retryable:", err.IsRetryable())

	// Output:
	// Technical: [CONFIG_MISSING] Required configuration 'api_key' is missing
	// User: Configuration error: Required setting 'api_key' is not configured. Please check your configuration file.
	// Code: CONFIG_MISSING
	// Severity: high
	// Retryable: false
}

// Example demonstrates error wrapping
func Example_errorWrapping() {
	// Simulate a lower-level error
	dbErr := fmt.Errorf("connection refused")

	// Wrap it with context
	wrappedErr := errors.Wrap(dbErr, errors.ErrCodeDBConnection, "failed to connect to database")

	fmt.Println(wrappedErr)

	// Output:
	// [DB_CONNECTION_FAILED] failed to connect to database: connection refused
}

// Example demonstrates retry with exponential backoff
func Example_retryWithBackoff() {
	ctx := context.Background()

	// Configure retry strategy
	config := errors.NewRetryConfig()
	config.Strategy = errors.NewLinearBackoff(10*time.Millisecond, 3)

	callCount := 0
	config.OnRetry = func(attempt int, err error) {
		fmt.Printf("Retry attempt %d failed\n", attempt+1)
	}

	// Simulate an operation that fails twice then succeeds
	err := errors.Retry(ctx, func() error {
		callCount++
		if callCount < 3 {
			return errors.NewProviderTimeoutError("api", "model", 30*time.Second)
		}
		return nil
	}, config)

	if err != nil {
		fmt.Println("Failed after all retries")
	} else {
		fmt.Printf("Success after %d total attempts\n", callCount)
	}

	// Output:
	// Retry attempt 1 failed
	// Retry attempt 2 failed
	// Success after 3 total attempts
}

// Example demonstrates circuit breaker pattern
func Example_circuitBreaker() {
	// Create circuit breaker (max 2 failures, 100ms reset)
	cb := errors.NewCircuitBreaker(2, 100*time.Millisecond)

	// Simulate failing operation
	for i := 0; i < 3; i++ {
		err := cb.Execute(func() error {
			return fmt.Errorf("service unavailable")
		})

		if err != nil {
			if cb.GetState() == errors.StateOpen {
				fmt.Println("Circuit opened")
				break
			}
		}
	}

	// Circuit is now open
	err := cb.Execute(func() error {
		return nil
	})

	if err != nil {
		fmt.Println("Request blocked:", err.Error())
	}

	// Output:
	// Circuit opened
	// Request blocked: circuit breaker is open: too many failures
}

// Example demonstrates provider error with metadata
func Example_providerErrorWithMetadata() {
	err := errors.NewProviderTimeoutError("openai", "gpt-4", 30*time.Second)
	err.WithMetadata("request_id", "req-123456")
	err.WithMetadata("user_id", "user-789")

	fmt.Println("Provider:", err.ProviderName)
	fmt.Println("Model:", err.Model)
	fmt.Println("Retryable:", err.IsRetryable())

	metadata := err.Metadata()
	fmt.Println("Request ID:", metadata["request_id"])

	// Output:
	// Provider: openai
	// Model: gpt-4
	// Retryable: true
	// Request ID: req-123456
}

// Example demonstrates authentication error
func Example_authenticationError() {
	err := errors.NewExpiredTokenError("google")

	fmt.Println("Error:", err.Error())
	fmt.Println("Provider:", err.Provider)
	fmt.Println("Can retry:", err.IsRetryable())
	fmt.Println("User message:", err.UserMessage())

	// Output:
	// Error: [AUTH_EXPIRED_TOKEN] Authentication token expired for provider 'google'
	// Provider: google
	// Can retry: true
	// User message: Authentication error: Your session has expired. Please log in again.
}

// Example demonstrates database error handling
func Example_databaseError() {
	// Not found error
	notFoundErr := errors.NewDBNotFoundError("users", "id=123")
	fmt.Println("Not found:", notFoundErr.UserMessage())

	// Duplicate entry error
	dupErr := errors.NewDBDuplicateError("users", "email", "test@example.com")
	fmt.Println("Duplicate:", dupErr.UserMessage())

	// Output:
	// Not found: Not found: The requested resource does not exist.
	// Duplicate: Duplicate entry: A record with this email already exists.
}

// Example demonstrates tool execution error
func Example_toolExecutionError() {
	err := errors.NewToolNotFoundError("git")
	err.WithPath("/usr/bin/git")

	fmt.Println("Error:", err.Error())
	fmt.Println("Tool:", err.ToolName)
	fmt.Println("Path:", err.ToolPath)
	fmt.Println("User message:", err.UserMessage())

	// Output:
	// Error: [TOOL_NOT_FOUND] Tool 'git' not found
	// Tool: git
	// Path: /usr/bin/git
	// User message: Tool error: 'git' is not available or not installed. Please verify the tool is properly configured.
}

// Example demonstrates fallback pattern
func Example_fallback() {
	// Try primary operation, fall back to alternative
	err := errors.Fallback(
		func() error {
			return fmt.Errorf("primary service unavailable")
		},
		func() error {
			fmt.Println("Using fallback service")
			return nil
		},
	)

	if err == nil {
		fmt.Println("Operation succeeded")
	}

	// Output:
	// Using fallback service
	// Operation succeeded
}

// Example demonstrates error severity checking
func Example_errorSeverity() {
	criticalErr := errors.NewDBConnectionError("postgres", fmt.Errorf("refused"))
	highErr := errors.NewConfigMissingError("database_url")
	mediumErr := errors.NewDBDuplicateError("users", "email", "test@example.com")
	lowErr := errors.NewDBNotFoundError("products", "id=999")

	fmt.Println("Critical:", errors.GetSeverity(criticalErr))
	fmt.Println("High:", errors.GetSeverity(highErr))
	fmt.Println("Medium:", errors.GetSeverity(mediumErr))
	fmt.Println("Low:", errors.GetSeverity(lowErr))

	// Output:
	// Critical: critical
	// High: high
	// Medium: medium
	// Low: low
}

// Example demonstrates error code extraction
func Example_errorCodes() {
	configErr := errors.NewConfigInvalidError("timeout", "must be positive")
	authErr := errors.NewAuthFailedError("provider", nil)
	providerErr := errors.NewProviderRateLimitError("openai", 60*time.Second)

	fmt.Println("Config:", errors.GetCode(configErr))
	fmt.Println("Auth:", errors.GetCode(authErr))
	fmt.Println("Provider:", errors.GetCode(providerErr))

	// Output:
	// Config: CONFIG_INVALID
	// Auth: AUTH_FAILED
	// Provider: PROVIDER_RATE_LIMIT
}
