package provider

import (
	"context"
	"errors"
	"testing"
)

// Test 1: User Preference Selection
func TestSelector_SelectByUserPreference(t *testing.T) {
	// GIVEN a selector with multiple providers and user prefers Anthropic
	selector := NewSelector(
		WithProviders("anthropic", "openai", "google"),
		WithUserPreference("anthropic"),
	)

	// WHEN selecting a provider
	provider, err := selector.Select(context.Background(), nil)

	// THEN it should return Anthropic
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if provider == nil {
		t.Fatal("expected provider, got nil")
	}
	if provider.Name != "anthropic" {
		t.Errorf("expected anthropic, got %s", provider.Name)
	}
}

func TestSelector_SelectByUserPreference_NotAvailable(t *testing.T) {
	// GIVEN a selector where preferred provider is not available
	selector := NewSelector(
		WithProviders("openai", "google"),
		WithUserPreference("anthropic"),
	)

	// WHEN selecting a provider
	provider, err := selector.Select(context.Background(), nil)

	// THEN it should fallback to first available
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if provider.Name == "anthropic" {
		t.Error("should not return unavailable provider")
	}
	// Should return first available provider (openai)
	if provider.Name != "openai" {
		t.Errorf("expected openai as fallback, got %s", provider.Name)
	}
}

// Test 2: Credit-Aware Selection
func TestSelector_SelectByCreditBalance(t *testing.T) {
	// GIVEN a user with low credits
	user := &User{
		Email:   "test@example.com",
		Credits: 10,
		Tier:    "free",
	}

	selector := NewSelector(
		WithProviders("anthropic", "openai"),
		WithCreditThreshold(50),
	)

	// WHEN selecting a provider
	provider, err := selector.Select(context.Background(), user)

	// THEN it should warn about low credits
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !provider.LowCreditWarning {
		t.Error("expected low credit warning")
	}
}

func TestSelector_SelectByCreditBalance_NoCredits(t *testing.T) {
	// GIVEN a user with zero credits
	user := &User{
		Email:   "test@example.com",
		Credits: 0,
		Tier:    "free",
	}

	selector := NewSelector(
		WithProviders("anthropic", "openai"),
	)

	// WHEN selecting a provider
	_, err := selector.Select(context.Background(), user)

	// THEN it should return an error
	if err == nil {
		t.Fatal("expected error for zero credits, got nil")
	}
	if !errors.Is(err, ErrInsufficientCredits) {
		t.Errorf("expected ErrInsufficientCredits, got %v", err)
	}
}

func TestSelector_SelectByCreditBalance_SufficientCredits(t *testing.T) {
	// GIVEN a user with sufficient credits
	user := &User{
		Email:   "test@example.com",
		Credits: 100,
		Tier:    "pro",
	}

	selector := NewSelector(
		WithProviders("anthropic", "openai"),
		WithCreditThreshold(50),
	)

	// WHEN selecting a provider
	provider, err := selector.Select(context.Background(), user)

	// THEN it should not warn about low credits
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if provider.LowCreditWarning {
		t.Error("expected no low credit warning for sufficient credits")
	}
}

// Test 3: Model Capability Matching
func TestSelector_SelectByModelCapability_Vision(t *testing.T) {
	// GIVEN a request requiring vision capabilities
	req := &SelectionRequest{
		RequiresVision: true,
		Model:          "auto",
	}

	selector := NewSelector(
		WithProviders("anthropic", "openai"),
	)

	// WHEN selecting a provider
	provider, err := selector.Select(context.Background(), nil, req)

	// THEN it should return a provider with vision support
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !provider.SupportsVision {
		t.Error("expected provider with vision support")
	}
}

func TestSelector_SelectByModelCapability_FunctionCalling(t *testing.T) {
	// GIVEN a request requiring function calling
	req := &SelectionRequest{
		RequiresFunctionCalling: true,
		Model:                   "auto",
	}

	selector := NewSelector(
		WithProviders("anthropic", "openai", "google"),
	)

	// WHEN selecting a provider
	provider, err := selector.Select(context.Background(), nil, req)

	// THEN it should return a provider with function calling support
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !provider.SupportsFunctionCalling {
		t.Error("expected provider with function calling support")
	}
}

func TestSelector_SelectByModelCapability_Streaming(t *testing.T) {
	// GIVEN a request requiring streaming
	req := &SelectionRequest{
		RequiresStreaming: true,
		Model:             "auto",
	}

	selector := NewSelector(
		WithProviders("anthropic", "openai"),
	)

	// WHEN selecting a provider
	provider, err := selector.Select(context.Background(), nil, req)

	// THEN it should return a provider with streaming support
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !provider.SupportsStreaming {
		t.Error("expected provider with streaming support")
	}
}

func TestSelector_SelectByModelCapability_MultipleRequirements(t *testing.T) {
	// GIVEN a request requiring multiple capabilities
	req := &SelectionRequest{
		RequiresVision:          true,
		RequiresFunctionCalling: true,
		RequiresStreaming:       true,
		Model:                   "auto",
	}

	selector := NewSelector(
		WithProviders("anthropic", "openai", "google"),
	)

	// WHEN selecting a provider
	provider, err := selector.Select(context.Background(), nil, req)

	// THEN it should return a provider meeting all requirements
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !provider.SupportsVision {
		t.Error("expected provider with vision support")
	}
	if !provider.SupportsFunctionCalling {
		t.Error("expected provider with function calling support")
	}
	if !provider.SupportsStreaming {
		t.Error("expected provider with streaming support")
	}
}

func TestSelector_SelectByModelCapability_NoMatchingProvider(t *testing.T) {
	// GIVEN a request with requirements no provider can meet
	// (This is a hypothetical test - in reality all our providers support these)
	req := &SelectionRequest{
		RequiresVision: true,
		Model:          "auto",
	}

	// Create selector with empty providers list to simulate no match
	selector := NewSelector(
		WithProviders(),
	)

	// WHEN selecting a provider
	_, err := selector.Select(context.Background(), nil, req)

	// THEN it should return an error
	if err == nil {
		t.Fatal("expected error when no provider available, got nil")
	}
	if !errors.Is(err, ErrNoProviderAvailable) {
		t.Errorf("expected ErrNoProviderAvailable, got %v", err)
	}
}

// Test 4: Fallback Logic
func TestSelector_PreferredProviderWithCapabilityMismatch(t *testing.T) {
	// GIVEN a selector with user preference but requirement doesn't match
	// Note: In our case, all providers support all capabilities
	// This test demonstrates the fallback logic structure
	req := &SelectionRequest{
		RequiresVision: true,
		Model:          "auto",
	}

	selector := NewSelector(
		WithProviders("anthropic", "openai"),
		WithUserPreference("anthropic"),
	)

	// WHEN selecting a provider
	provider, err := selector.Select(context.Background(), nil, req)

	// THEN it should still select a provider that meets requirements
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !provider.SupportsVision {
		t.Error("expected provider with vision support")
	}
}

// Test 5: Provider Availability Check
func TestSelector_CheckProviderAvailability(t *testing.T) {
	// GIVEN a selector
	selector := NewSelector(
		WithProviders("anthropic", "openai", "google"),
	)

	// WHEN checking if provider is available
	available := selector.IsAvailable("anthropic")

	// THEN it should return true
	if !available {
		t.Error("expected anthropic to be available")
	}
}

func TestSelector_CheckProviderAvailability_Unavailable(t *testing.T) {
	// GIVEN a selector
	selector := NewSelector(
		WithProviders("anthropic", "openai"),
	)

	// WHEN checking if unavailable provider exists
	available := selector.IsAvailable("cohere")

	// THEN it should return false
	if available {
		t.Error("expected cohere to be unavailable")
	}
}

func TestSelector_CheckProviderAvailability_EmptyList(t *testing.T) {
	// GIVEN a selector with no providers
	selector := NewSelector(
		WithProviders(),
	)

	// WHEN checking if any provider is available
	available := selector.IsAvailable("anthropic")

	// THEN it should return false
	if available {
		t.Error("expected no providers to be available")
	}
}

// Test 6: Default Behavior
func TestSelector_DefaultSelection_NoPreference(t *testing.T) {
	// GIVEN a selector without user preference
	selector := NewSelector(
		WithProviders("anthropic", "openai", "google"),
	)

	// WHEN selecting a provider
	provider, err := selector.Select(context.Background(), nil)

	// THEN it should return the first available provider
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if provider.Name != "anthropic" {
		t.Errorf("expected first provider (anthropic), got %s", provider.Name)
	}
}

func TestSelector_DefaultSelection_WithUser(t *testing.T) {
	// GIVEN a selector with a user who has credits
	user := &User{
		Email:   "test@example.com",
		Credits: 75,
		Tier:    "pro",
	}

	selector := NewSelector(
		WithProviders("anthropic", "openai"),
		WithCreditThreshold(50),
	)

	// WHEN selecting a provider
	provider, err := selector.Select(context.Background(), user)

	// THEN it should return provider without warning
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if provider.LowCreditWarning {
		t.Error("expected no warning for credits above threshold")
	}
}

// Test 7: Edge Cases
func TestSelector_NoProvidersConfigured(t *testing.T) {
	// GIVEN a selector with no providers
	selector := NewSelector(
		WithProviders(),
	)

	// WHEN selecting a provider
	_, err := selector.Select(context.Background(), nil)

	// THEN it should return an error
	if err == nil {
		t.Fatal("expected error when no providers configured, got nil")
	}
	if !errors.Is(err, ErrNoProviderAvailable) {
		t.Errorf("expected ErrNoProviderAvailable, got %v", err)
	}
}

func TestSelector_NilUser(t *testing.T) {
	// GIVEN a selector with nil user
	selector := NewSelector(
		WithProviders("anthropic", "openai"),
	)

	// WHEN selecting a provider
	provider, err := selector.Select(context.Background(), nil)

	// THEN it should work without error
	if err != nil {
		t.Fatalf("expected no error with nil user, got %v", err)
	}
	if provider == nil {
		t.Fatal("expected provider, got nil")
	}
}

// Test 8: Credit Threshold Configuration
func TestSelector_CustomCreditThreshold(t *testing.T) {
	// GIVEN a selector with custom credit threshold
	user := &User{
		Email:   "test@example.com",
		Credits: 75,
		Tier:    "pro",
	}

	selector := NewSelector(
		WithProviders("anthropic"),
		WithCreditThreshold(100),
	)

	// WHEN selecting a provider
	provider, err := selector.Select(context.Background(), user)

	// THEN it should warn because credits (75) < threshold (100)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !provider.LowCreditWarning {
		t.Error("expected low credit warning with custom threshold")
	}
}

func TestSelector_ZeroCreditThreshold(t *testing.T) {
	// GIVEN a selector with zero credit threshold
	user := &User{
		Email:   "test@example.com",
		Credits: 10,
		Tier:    "free",
	}

	selector := NewSelector(
		WithProviders("anthropic"),
		WithCreditThreshold(0),
	)

	// WHEN selecting a provider
	provider, err := selector.Select(context.Background(), user)

	// THEN it should not warn (any credits above 0 is fine)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if provider.LowCreditWarning {
		t.Error("expected no warning with zero threshold and positive credits")
	}
}

// Test 9: Provider Capabilities Validation
func TestSelector_ProviderHasCorrectCapabilities(t *testing.T) {
	// GIVEN a selector
	selector := NewSelector(
		WithProviders("anthropic"),
	)

	// WHEN selecting a provider
	provider, err := selector.Select(context.Background(), nil)

	// THEN it should have expected capabilities
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Validate Anthropic capabilities
	if provider.Name != "anthropic" {
		t.Errorf("expected anthropic, got %s", provider.Name)
	}
	if provider.DisplayName == "" {
		t.Error("expected non-empty display name")
	}
	if provider.MaxTokens <= 0 {
		t.Error("expected positive max tokens")
	}
}

// Test 10: User Preference Override with Capabilities
func TestSelector_UserPreferenceWithCapabilityCheck(t *testing.T) {
	// GIVEN user prefers a provider and request has capability requirements
	user := &User{
		Email:   "test@example.com",
		Credits: 100,
		Tier:    "pro",
	}

	req := &SelectionRequest{
		RequiresVision:          true,
		RequiresFunctionCalling: true,
	}

	selector := NewSelector(
		WithProviders("anthropic", "openai", "google"),
		WithUserPreference("google"),
	)

	// WHEN selecting a provider
	provider, err := selector.Select(context.Background(), user, req)

	// THEN it should return preferred provider if it meets requirements
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	// Google supports both vision and function calling
	if provider.Name != "google" {
		t.Errorf("expected preferred provider (google), got %s", provider.Name)
	}
	if !provider.SupportsVision || !provider.SupportsFunctionCalling {
		t.Error("expected provider to meet all capability requirements")
	}
}
