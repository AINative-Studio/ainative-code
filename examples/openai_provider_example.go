package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/AINative-studio/ainative-code/internal/provider"
	"github.com/AINative-studio/ainative-code/internal/provider/openai"
)

func main() {
	// Example 1: Basic Chat Completion
	basicChatExample()

	// Example 2: Streaming Chat
	streamingChatExample()

	// Example 3: Multi-Provider Setup
	multiProviderExample()

	// Example 4: Advanced Configuration
	advancedConfigExample()
}

// basicChatExample demonstrates basic chat completion with OpenAI
func basicChatExample() {
	fmt.Println("\n=== Example 1: Basic Chat Completion ===")

	// Create OpenAI provider
	p, err := openai.NewOpenAIProvider(openai.Config{
		APIKey: os.Getenv("OPENAI_API_KEY"),
	})
	if err != nil {
		log.Fatalf("Failed to create provider: %v", err)
	}
	defer p.Close()

	// Prepare messages
	messages := []provider.Message{
		{Role: "system", Content: "You are a helpful AI assistant."},
		{Role: "user", Content: "What is the capital of France?"},
	}

	// Send chat request
	ctx := context.Background()
	resp, err := p.Chat(ctx, messages,
		provider.WithModel("gpt-4"),
		provider.WithMaxTokens(100),
		provider.WithTemperature(0.7),
	)
	if err != nil {
		log.Fatalf("Chat failed: %v", err)
	}

	fmt.Printf("Model: %s\n", resp.Model)
	fmt.Printf("Response: %s\n", resp.Content)
	fmt.Printf("Tokens used: %d (prompt: %d, completion: %d)\n",
		resp.Usage.TotalTokens,
		resp.Usage.PromptTokens,
		resp.Usage.CompletionTokens,
	)
}

// streamingChatExample demonstrates streaming responses
func streamingChatExample() {
	fmt.Println("\n=== Example 2: Streaming Chat ===")

	// Create OpenAI provider
	p, err := openai.NewOpenAIProvider(openai.Config{
		APIKey: os.Getenv("OPENAI_API_KEY"),
	})
	if err != nil {
		log.Fatalf("Failed to create provider: %v", err)
	}
	defer p.Close()

	// Prepare messages
	messages := []provider.Message{
		{Role: "user", Content: "Count from 1 to 5"},
	}

	// Stream chat request
	ctx := context.Background()
	eventChan, err := p.Stream(ctx, messages,
		provider.StreamWithModel("gpt-3.5-turbo"),
		provider.StreamWithMaxTokens(50),
	)
	if err != nil {
		log.Fatalf("Stream failed: %v", err)
	}

	// Process streaming events
	fmt.Print("Streaming response: ")
	for event := range eventChan {
		switch event.Type {
		case provider.EventTypeContentStart:
			// Stream started
		case provider.EventTypeContentDelta:
			// Incremental content
			fmt.Print(event.Content)
		case provider.EventTypeContentEnd:
			// Stream completed
			fmt.Println("\n[Stream complete]")
		case provider.EventTypeError:
			// Error occurred
			log.Printf("Stream error: %v", event.Error)
		}
	}
}

// multiProviderExample demonstrates using multiple providers together
func multiProviderExample() {
	fmt.Println("\n=== Example 3: Multi-Provider Setup ===")

	// Create provider registry
	registry := provider.NewRegistry()
	defer registry.Close()

	// Register OpenAI provider
	openaiProvider, err := openai.NewOpenAIProvider(openai.Config{
		APIKey: os.Getenv("OPENAI_API_KEY"),
	})
	if err != nil {
		log.Fatalf("Failed to create OpenAI provider: %v", err)
	}

	err = registry.Register("openai", openaiProvider)
	if err != nil {
		log.Fatalf("Failed to register provider: %v", err)
	}

	// Note: In a real application, you would also register Anthropic, Google, etc.
	// registry.Register("anthropic", anthropicProvider)
	// registry.Register("google", googleProvider)

	// List available providers
	fmt.Printf("Available providers: %v\n", registry.List())

	// Use a specific provider
	p, err := registry.Get("openai")
	if err != nil {
		log.Fatalf("Provider not found: %v", err)
	}

	ctx := context.Background()
	messages := []provider.Message{
		{Role: "user", Content: "Hello from multi-provider setup!"},
	}

	resp, err := p.Chat(ctx, messages, provider.WithModel("gpt-3.5-turbo"))
	if err != nil {
		log.Fatalf("Chat failed: %v", err)
	}

	fmt.Printf("Response from %s: %s\n", p.Name(), resp.Content)
}

// advancedConfigExample demonstrates advanced configuration options
func advancedConfigExample() {
	fmt.Println("\n=== Example 4: Advanced Configuration ===")

	// Create provider with advanced options
	p, err := openai.NewOpenAIProvider(openai.Config{
		APIKey:       os.Getenv("OPENAI_API_KEY"),
		Organization: os.Getenv("OPENAI_ORG_ID"), // Optional organization ID
		BaseURL:      "https://api.openai.com/v1", // Can be customized for proxies
	})
	if err != nil {
		log.Fatalf("Failed to create provider: %v", err)
	}
	defer p.Close()

	// Use advanced chat options
	messages := []provider.Message{
		{Role: "user", Content: "Explain quantum computing in simple terms."},
	}

	ctx := context.Background()
	resp, err := p.Chat(ctx, messages,
		provider.WithModel("gpt-4-turbo-preview"),
		provider.WithMaxTokens(200),
		provider.WithTemperature(0.8),
		provider.WithTopP(0.95),
		provider.WithStopSequences("\n\n", "END"),
	)
	if err != nil {
		log.Fatalf("Chat failed: %v", err)
	}

	fmt.Printf("Response: %s\n", resp.Content)

	// List supported models
	fmt.Printf("\nSupported models: %v\n", p.Models())
}

// Example: Error Handling
func errorHandlingExample() {
	fmt.Println("\n=== Example 5: Error Handling ===")

	p, err := openai.NewOpenAIProvider(openai.Config{
		APIKey: "invalid-key",
	})
	if err != nil {
		log.Fatalf("Failed to create provider: %v", err)
	}
	defer p.Close()

	ctx := context.Background()
	messages := []provider.Message{
		{Role: "user", Content: "Test"},
	}

	_, err = p.Chat(ctx, messages, provider.WithModel("gpt-4"))
	if err != nil {
		// Check error type
		switch {
		case isAuthenticationError(err):
			fmt.Println("Authentication error: Check your API key")
		case isRateLimitError(err):
			fmt.Println("Rate limit error: Please retry later")
		case isContextLengthError(err):
			fmt.Println("Context length error: Reduce message length")
		default:
			fmt.Printf("Other error: %v\n", err)
		}
	}
}

// Helper functions to check error types
func isAuthenticationError(err error) bool {
	return err != nil && (err.Error() == "authentication" ||
		err.Error() == "invalid_api_key")
}

func isRateLimitError(err error) bool {
	return err != nil && err.Error() == "rate_limit"
}

func isContextLengthError(err error) bool {
	return err != nil && err.Error() == "context_length"
}
