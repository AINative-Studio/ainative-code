package provider

import (
	"errors"
	"fmt"
	"testing"
)

func TestProviderError_Error(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		model    string
		err      error
		expected string
	}{
		{
			name:     "with model",
			provider: "openai",
			model:    "gpt-4",
			err:      fmt.Errorf("connection failed"),
			expected: `provider "openai" (model "gpt-4"): connection failed`,
		},
		{
			name:     "without model",
			provider: "anthropic",
			model:    "",
			err:      fmt.Errorf("invalid request"),
			expected: `provider "anthropic": invalid request`,
		},
		{
			name:     "with wrapped error",
			provider: "google",
			model:    "gemini-pro",
			err:      fmt.Errorf("wrapped: %w", fmt.Errorf("original error")),
			expected: `provider "google" (model "gemini-pro"): wrapped: original error`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pe := &ProviderError{
				Provider: tt.provider,
				Model:    tt.model,
				Err:      tt.err,
			}

			if pe.Error() != tt.expected {
				t.Errorf("Error() = %q, expected %q", pe.Error(), tt.expected)
			}
		})
	}
}

func TestProviderError_Unwrap(t *testing.T) {
	originalErr := fmt.Errorf("original error")
	pe := &ProviderError{
		Provider: "test",
		Model:    "test-model",
		Err:      originalErr,
	}

	unwrapped := pe.Unwrap()
	if unwrapped != originalErr {
		t.Errorf("Unwrap() returned different error: got %v, expected %v", unwrapped, originalErr)
	}

	// Test error chain
	if !errors.Is(pe, originalErr) {
		t.Error("errors.Is should find original error in chain")
	}
}

func TestNewProviderError(t *testing.T) {
	provider := "openai"
	model := "gpt-4"
	err := fmt.Errorf("test error")

	pe := NewProviderError(provider, model, err)

	if pe.Provider != provider {
		t.Errorf("Provider = %q, expected %q", pe.Provider, provider)
	}

	if pe.Model != model {
		t.Errorf("Model = %q, expected %q", pe.Model, model)
	}

	if pe.Err != err {
		t.Errorf("Err = %v, expected %v", pe.Err, err)
	}
}

func TestRateLimitError_Error(t *testing.T) {
	tests := []struct {
		name         string
		provider     string
		retryAfter   int
		limitType    string
		currentUsage int
		limit        int
		expected     string
	}{
		{
			name:       "with retry after",
			provider:   "openai",
			retryAfter: 60,
			expected:   `rate limit exceeded for provider "openai": retry after 60 seconds`,
		},
		{
			name:         "with limit details",
			provider:     "anthropic",
			limitType:    "requests",
			currentUsage: 100,
			limit:        100,
			expected:     `rate limit exceeded for provider "anthropic": requests limit reached (100/100)`,
		},
		{
			name:     "basic message",
			provider: "google",
			expected: `rate limit exceeded for provider "google"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rle := &RateLimitError{
				Provider:     tt.provider,
				RetryAfter:   tt.retryAfter,
				LimitType:    tt.limitType,
				CurrentUsage: tt.currentUsage,
				Limit:        tt.limit,
			}

			if rle.Error() != tt.expected {
				t.Errorf("Error() = %q, expected %q", rle.Error(), tt.expected)
			}
		})
	}
}

func TestNewRateLimitError(t *testing.T) {
	provider := "openai"
	retryAfter := 120

	rle := NewRateLimitError(provider, retryAfter)

	if rle.Provider != provider {
		t.Errorf("Provider = %q, expected %q", rle.Provider, provider)
	}

	if rle.RetryAfter != retryAfter {
		t.Errorf("RetryAfter = %d, expected %d", rle.RetryAfter, retryAfter)
	}

	// Other fields should be zero/empty
	if rle.LimitType != "" {
		t.Errorf("LimitType should be empty, got %q", rle.LimitType)
	}

	if rle.CurrentUsage != 0 {
		t.Errorf("CurrentUsage should be 0, got %d", rle.CurrentUsage)
	}

	if rle.Limit != 0 {
		t.Errorf("Limit should be 0, got %d", rle.Limit)
	}
}

func TestNewRateLimitErrorWithDetails(t *testing.T) {
	provider := "anthropic"
	limitType := "tokens"
	currentUsage := 5000
	limit := 5000

	rle := NewRateLimitErrorWithDetails(provider, limitType, currentUsage, limit)

	if rle.Provider != provider {
		t.Errorf("Provider = %q, expected %q", rle.Provider, provider)
	}

	if rle.LimitType != limitType {
		t.Errorf("LimitType = %q, expected %q", rle.LimitType, limitType)
	}

	if rle.CurrentUsage != currentUsage {
		t.Errorf("CurrentUsage = %d, expected %d", rle.CurrentUsage, currentUsage)
	}

	if rle.Limit != limit {
		t.Errorf("Limit = %d, expected %d", rle.Limit, limit)
	}

	// RetryAfter should be zero
	if rle.RetryAfter != 0 {
		t.Errorf("RetryAfter should be 0, got %d", rle.RetryAfter)
	}
}

func TestAuthenticationError_Error(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		reason   string
		expected string
	}{
		{
			name:     "with reason",
			provider: "openai",
			reason:   "invalid API key",
			expected: `authentication failed for provider "openai": invalid API key`,
		},
		{
			name:     "without reason",
			provider: "anthropic",
			reason:   "",
			expected: `authentication failed for provider "anthropic"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ae := &AuthenticationError{
				Provider: tt.provider,
				Reason:   tt.reason,
			}

			if ae.Error() != tt.expected {
				t.Errorf("Error() = %q, expected %q", ae.Error(), tt.expected)
			}
		})
	}
}

func TestNewAuthenticationError(t *testing.T) {
	provider := "openai"
	reason := "expired token"

	ae := NewAuthenticationError(provider, reason)

	if ae.Provider != provider {
		t.Errorf("Provider = %q, expected %q", ae.Provider, provider)
	}

	if ae.Reason != reason {
		t.Errorf("Reason = %q, expected %q", ae.Reason, reason)
	}
}

func TestContextLengthError_Error(t *testing.T) {
	tests := []struct {
		name          string
		provider      string
		model         string
		requestTokens int
		maxTokens     int
		expected      string
	}{
		{
			name:          "with token counts",
			provider:      "openai",
			model:         "gpt-4",
			requestTokens: 10000,
			maxTokens:     8192,
			expected:      `context length exceeded for provider "openai" model "gpt-4": 10000 tokens exceeds maximum of 8192`,
		},
		{
			name:          "without token counts",
			provider:      "anthropic",
			model:         "claude-3-opus",
			requestTokens: 0,
			maxTokens:     0,
			expected:      `context length exceeded for provider "anthropic" model "claude-3-opus"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cle := &ContextLengthError{
				Provider:      tt.provider,
				Model:         tt.model,
				RequestTokens: tt.requestTokens,
				MaxTokens:     tt.maxTokens,
			}

			if cle.Error() != tt.expected {
				t.Errorf("Error() = %q, expected %q", cle.Error(), tt.expected)
			}
		})
	}
}

func TestNewContextLengthError(t *testing.T) {
	provider := "openai"
	model := "gpt-4"
	requestTokens := 10000
	maxTokens := 8192

	cle := NewContextLengthError(provider, model, requestTokens, maxTokens)

	if cle.Provider != provider {
		t.Errorf("Provider = %q, expected %q", cle.Provider, provider)
	}

	if cle.Model != model {
		t.Errorf("Model = %q, expected %q", cle.Model, model)
	}

	if cle.RequestTokens != requestTokens {
		t.Errorf("RequestTokens = %d, expected %d", cle.RequestTokens, requestTokens)
	}

	if cle.MaxTokens != maxTokens {
		t.Errorf("MaxTokens = %d, expected %d", cle.MaxTokens, maxTokens)
	}
}

func TestInvalidModelError_Error(t *testing.T) {
	tests := []struct {
		name            string
		provider        string
		model           string
		supportedModels []string
		expected        string
	}{
		{
			name:            "with supported models",
			provider:        "openai",
			model:           "gpt-5",
			supportedModels: []string{"gpt-4", "gpt-3.5-turbo"},
			expected:        `invalid model "gpt-5" for provider "openai": supported models are [gpt-4 gpt-3.5-turbo]`,
		},
		{
			name:            "without supported models",
			provider:        "anthropic",
			model:           "invalid-model",
			supportedModels: nil,
			expected:        `invalid model "invalid-model" for provider "anthropic"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ime := &InvalidModelError{
				Provider:        tt.provider,
				Model:           tt.model,
				SupportedModels: tt.supportedModels,
			}

			if ime.Error() != tt.expected {
				t.Errorf("Error() = %q, expected %q", ime.Error(), tt.expected)
			}
		})
	}
}

func TestNewInvalidModelError(t *testing.T) {
	provider := "openai"
	model := "invalid"
	supportedModels := []string{"gpt-4", "gpt-3.5-turbo"}

	ime := NewInvalidModelError(provider, model, supportedModels)

	if ime.Provider != provider {
		t.Errorf("Provider = %q, expected %q", ime.Provider, provider)
	}

	if ime.Model != model {
		t.Errorf("Model = %q, expected %q", ime.Model, model)
	}

	if len(ime.SupportedModels) != len(supportedModels) {
		t.Errorf("SupportedModels length = %d, expected %d", len(ime.SupportedModels), len(supportedModels))
	}

	for i, sm := range supportedModels {
		if ime.SupportedModels[i] != sm {
			t.Errorf("SupportedModels[%d] = %q, expected %q", i, ime.SupportedModels[i], sm)
		}
	}
}

func TestStreamingNotSupportedError_Error(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		model    string
		reason   string
		expected string
	}{
		{
			name:     "with model and reason",
			provider: "openai",
			model:    "gpt-4",
			reason:   "streaming disabled by admin",
			expected: `streaming not supported for provider "openai" model "gpt-4": streaming disabled by admin`,
		},
		{
			name:     "with model, no reason",
			provider: "anthropic",
			model:    "claude-3-opus",
			reason:   "",
			expected: `streaming not supported for provider "anthropic" model "claude-3-opus"`,
		},
		{
			name:     "without model",
			provider: "google",
			model:    "",
			reason:   "",
			expected: `streaming not supported for provider "google"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			snse := &StreamingNotSupportedError{
				Provider: tt.provider,
				Model:    tt.model,
				Reason:   tt.reason,
			}

			if snse.Error() != tt.expected {
				t.Errorf("Error() = %q, expected %q", snse.Error(), tt.expected)
			}
		})
	}
}

func TestNewStreamingNotSupportedError(t *testing.T) {
	provider := "openai"
	model := "gpt-4"
	reason := "feature not enabled"

	snse := NewStreamingNotSupportedError(provider, model, reason)

	if snse.Provider != provider {
		t.Errorf("Provider = %q, expected %q", snse.Provider, provider)
	}

	if snse.Model != model {
		t.Errorf("Model = %q, expected %q", snse.Model, model)
	}

	if snse.Reason != reason {
		t.Errorf("Reason = %q, expected %q", snse.Reason, reason)
	}
}

func TestErrorTypeAssertion(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		checkFn  func(error) bool
		expected bool
	}{
		{
			name: "ProviderError assertion",
			err:  NewProviderError("test", "test-model", fmt.Errorf("error")),
			checkFn: func(err error) bool {
				var pe *ProviderError
				return errors.As(err, &pe)
			},
			expected: true,
		},
		{
			name: "RateLimitError assertion",
			err:  NewRateLimitError("test", 60),
			checkFn: func(err error) bool {
				var rle *RateLimitError
				return errors.As(err, &rle)
			},
			expected: true,
		},
		{
			name: "AuthenticationError assertion",
			err:  NewAuthenticationError("test", "invalid key"),
			checkFn: func(err error) bool {
				var ae *AuthenticationError
				return errors.As(err, &ae)
			},
			expected: true,
		},
		{
			name: "ContextLengthError assertion",
			err:  NewContextLengthError("test", "model", 1000, 500),
			checkFn: func(err error) bool {
				var cle *ContextLengthError
				return errors.As(err, &cle)
			},
			expected: true,
		},
		{
			name: "InvalidModelError assertion",
			err:  NewInvalidModelError("test", "invalid", []string{"valid"}),
			checkFn: func(err error) bool {
				var ime *InvalidModelError
				return errors.As(err, &ime)
			},
			expected: true,
		},
		{
			name: "StreamingNotSupportedError assertion",
			err:  NewStreamingNotSupportedError("test", "model", "reason"),
			checkFn: func(err error) bool {
				var snse *StreamingNotSupportedError
				return errors.As(err, &snse)
			},
			expected: true,
		},
		{
			name: "wrong type assertion",
			err:  NewProviderError("test", "model", fmt.Errorf("error")),
			checkFn: func(err error) bool {
				var rle *RateLimitError
				return errors.As(err, &rle)
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.checkFn(tt.err)
			if result != tt.expected {
				t.Errorf("type assertion result = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestErrorUnwrapping(t *testing.T) {
	// Test error chain with wrapped ProviderError
	originalErr := fmt.Errorf("network timeout")
	wrappedErr := fmt.Errorf("request failed: %w", originalErr)
	providerErr := NewProviderError("openai", "gpt-4", wrappedErr)

	// Test errors.Is through the chain
	if !errors.Is(providerErr, originalErr) {
		t.Error("errors.Is should find original error through wrapped chain")
	}

	// Test direct unwrap
	unwrapped := providerErr.Unwrap()
	if unwrapped != wrappedErr {
		t.Errorf("Unwrap() = %v, expected %v", unwrapped, wrappedErr)
	}

	// Test unwrapping to get original
	if !errors.Is(unwrapped, originalErr) {
		t.Error("unwrapped error should contain original error")
	}
}

func TestErrorCreationWithNilValues(t *testing.T) {
	// Test that constructors handle edge cases

	// ProviderError with nil error
	pe := NewProviderError("test", "model", nil)
	if pe.Err != nil {
		t.Errorf("ProviderError.Err should be nil, got %v", pe.Err)
	}

	// InvalidModelError with nil supported models
	ime := NewInvalidModelError("test", "invalid", nil)
	if ime.SupportedModels != nil {
		t.Errorf("SupportedModels should be nil, got %v", ime.SupportedModels)
	}

	// InvalidModelError with empty supported models
	ime2 := NewInvalidModelError("test", "invalid", []string{})
	if len(ime2.SupportedModels) != 0 {
		t.Errorf("SupportedModels should be empty, got length %d", len(ime2.SupportedModels))
	}
}

func TestErrorMessageFormatting(t *testing.T) {
	// Test that error messages are properly formatted and consistent

	// ProviderError formatting
	pe := NewProviderError("openai", "gpt-4", fmt.Errorf("test"))
	peMsg := pe.Error()
	if !containsSubstring(peMsg, "openai") || !containsSubstring(peMsg, "gpt-4") {
		t.Errorf("ProviderError message missing provider or model: %q", peMsg)
	}

	// RateLimitError formatting with retry
	rle := NewRateLimitError("anthropic", 30)
	rleMsg := rle.Error()
	if !containsSubstring(rleMsg, "anthropic") || !containsSubstring(rleMsg, "30 seconds") {
		t.Errorf("RateLimitError message missing provider or retry time: %q", rleMsg)
	}

	// AuthenticationError formatting
	ae := NewAuthenticationError("google", "invalid credentials")
	aeMsg := ae.Error()
	if !containsSubstring(aeMsg, "google") || !containsSubstring(aeMsg, "invalid credentials") {
		t.Errorf("AuthenticationError message missing provider or reason: %q", aeMsg)
	}

	// ContextLengthError formatting
	cle := NewContextLengthError("openai", "gpt-4", 10000, 8192)
	cleMsg := cle.Error()
	if !containsSubstring(cleMsg, "openai") || !containsSubstring(cleMsg, "gpt-4") ||
		!containsSubstring(cleMsg, "10000") || !containsSubstring(cleMsg, "8192") {
		t.Errorf("ContextLengthError message missing details: %q", cleMsg)
	}

	// InvalidModelError formatting
	ime := NewInvalidModelError("anthropic", "invalid", []string{"claude-3-opus", "claude-3-sonnet"})
	imeMsg := ime.Error()
	if !containsSubstring(imeMsg, "anthropic") || !containsSubstring(imeMsg, "invalid") ||
		!containsSubstring(imeMsg, "claude-3-opus") {
		t.Errorf("InvalidModelError message missing details: %q", imeMsg)
	}

	// StreamingNotSupportedError formatting
	snse := NewStreamingNotSupportedError("openai", "gpt-4", "not available")
	snseMsg := snse.Error()
	if !containsSubstring(snseMsg, "openai") || !containsSubstring(snseMsg, "gpt-4") ||
		!containsSubstring(snseMsg, "not available") {
		t.Errorf("StreamingNotSupportedError message missing details: %q", snseMsg)
	}
}

func TestRateLimitErrorPriority(t *testing.T) {
	// Test that RetryAfter takes priority in error message
	rle := &RateLimitError{
		Provider:     "test",
		RetryAfter:   60,
		LimitType:    "requests",
		CurrentUsage: 100,
		Limit:        100,
	}

	msg := rle.Error()
	// Should prioritize RetryAfter over LimitType
	if !containsSubstring(msg, "retry after 60 seconds") {
		t.Errorf("expected RetryAfter message, got: %q", msg)
	}
	if containsSubstring(msg, "requests limit") {
		t.Errorf("should not include limit type when RetryAfter is set, got: %q", msg)
	}
}

func TestErrorEmptyStrings(t *testing.T) {
	// Test error messages with empty strings

	// ProviderError with empty provider
	pe := NewProviderError("", "model", fmt.Errorf("error"))
	peMsg := pe.Error()
	if !containsSubstring(peMsg, `provider ""`) {
		t.Errorf("expected empty provider quoted, got: %q", peMsg)
	}

	// AuthenticationError with empty reason
	ae := NewAuthenticationError("test", "")
	aeMsg := ae.Error()
	if containsSubstring(aeMsg, ":") && len(aeMsg) > len(`authentication failed for provider "test"`) {
		t.Errorf("should not include reason separator, got: %q", aeMsg)
	}

	// InvalidModelError with empty model
	ime := NewInvalidModelError("test", "", []string{"valid"})
	imeMsg := ime.Error()
	if !containsSubstring(imeMsg, `model ""`) {
		t.Errorf("expected empty model quoted, got: %q", imeMsg)
	}
}

// Helper function to check if a string contains a substring
func containsSubstring(s, substr string) bool {
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
