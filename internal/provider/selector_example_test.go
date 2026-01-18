package provider_test

import (
	"context"
	"fmt"

	"github.com/AINative-studio/ainative-code/internal/provider"
)

// Example demonstrates basic provider selection by user preference
func ExampleSelector_Select_userPreference() {
	// Create a selector with user's preferred provider
	selector := provider.NewSelector(
		provider.WithProviders("anthropic", "openai", "google"),
		provider.WithUserPreference("anthropic"),
	)

	// Select provider
	selectedProvider, err := selector.Select(context.Background(), nil)
	if err != nil {
		panic(err)
	}

	fmt.Println(selectedProvider.Name)
	fmt.Println(selectedProvider.DisplayName)
	// Output:
	// anthropic
	// Anthropic Claude
}

// Example demonstrates credit-aware provider selection
func ExampleSelector_Select_creditAware() {
	// User with low credits
	user := &provider.User{
		Email:   "test@example.com",
		Credits: 10,
		Tier:    "free",
	}

	// Create selector with credit threshold
	selector := provider.NewSelector(
		provider.WithProviders("anthropic", "openai"),
		provider.WithCreditThreshold(50),
	)

	// Select provider
	selectedProvider, err := selector.Select(context.Background(), user)
	if err != nil {
		panic(err)
	}

	fmt.Println("Provider:", selectedProvider.Name)
	fmt.Println("Low Credit Warning:", selectedProvider.LowCreditWarning)
	// Output:
	// Provider: anthropic
	// Low Credit Warning: true
}

// Example demonstrates capability-based provider selection
func ExampleSelector_Select_capabilities() {
	// Request requiring vision support
	req := &provider.SelectionRequest{
		RequiresVision: true,
		Model:          "auto",
	}

	selector := provider.NewSelector(
		provider.WithProviders("anthropic", "openai", "google"),
	)

	// Select provider that supports vision
	selectedProvider, err := selector.Select(context.Background(), nil, req)
	if err != nil {
		panic(err)
	}

	fmt.Println("Supports Vision:", selectedProvider.SupportsVision)
	// Output:
	// Supports Vision: true
}

// Example demonstrates checking provider availability
func ExampleSelector_IsAvailable() {
	selector := provider.NewSelector(
		provider.WithProviders("anthropic", "openai"),
	)

	// Check if providers are available
	fmt.Println("Anthropic:", selector.IsAvailable("anthropic"))
	fmt.Println("Google:", selector.IsAvailable("google"))
	// Output:
	// Anthropic: true
	// Google: false
}

// Example demonstrates advanced selection with multiple requirements
func ExampleSelector_Select_advanced() {
	// User with sufficient credits
	user := &provider.User{
		Email:   "premium@example.com",
		Credits: 1000,
		Tier:    "pro",
	}

	// Request with multiple capability requirements
	req := &provider.SelectionRequest{
		RequiresVision:          true,
		RequiresFunctionCalling: true,
		RequiresStreaming:       true,
		Model:                   "auto",
	}

	// Create selector with preferred provider
	selector := provider.NewSelector(
		provider.WithProviders("anthropic", "openai", "google"),
		provider.WithUserPreference("google"),
		provider.WithCreditThreshold(100),
	)

	// Select provider
	selectedProvider, err := selector.Select(context.Background(), user, req)
	if err != nil {
		panic(err)
	}

	fmt.Println("Selected:", selectedProvider.Name)
	fmt.Println("Vision:", selectedProvider.SupportsVision)
	fmt.Println("Function Calling:", selectedProvider.SupportsFunctionCalling)
	fmt.Println("Streaming:", selectedProvider.SupportsStreaming)
	fmt.Println("Low Credit Warning:", selectedProvider.LowCreditWarning)
	// Output:
	// Selected: google
	// Vision: true
	// Function Calling: true
	// Streaming: true
	// Low Credit Warning: false
}
