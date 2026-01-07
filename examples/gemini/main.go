package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/AINative-studio/ainative-code/internal/provider"
	"github.com/AINative-studio/ainative-code/internal/provider/gemini"
)

func main() {
	// Get API key from environment
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		log.Fatal("GOOGLE_API_KEY environment variable is required")
	}

	// Create Gemini provider
	config := gemini.Config{
		APIKey: apiKey,
	}

	geminiprovider, err := gemini.NewGeminiProvider(config)
	if err != nil {
		log.Fatalf("Failed to create provider: %v", err)
	}
	defer geminiprovider.Close()

	fmt.Println("=== Gemini Provider Examples ===\n")

	// Example 1: Simple chat
	example1SimplChat(geminiprovider)

	// Example 2: Streaming chat
	example2StreamingChat(geminiprovider)

	// Example 3: Multi-turn conversation
	example3MultiTurn(geminiprovider)

	// Example 4: With system prompt
	example4SystemPrompt(geminiprovider)

	// Example 5: Different models
	example5DifferentModels(geminiprovider)
}

func example1SimplChat(p *gemini.GeminiProvider) {
	fmt.Println("Example 1: Simple Chat Request")
	fmt.Println("-------------------------------")

	ctx := context.Background()
	messages := []provider.Message{
		{Role: "user", Content: "What are the three laws of robotics?"},
	}

	response, err := p.Chat(ctx, messages,
		provider.WithModel("gemini-pro"),
		provider.WithMaxTokens(200),
		provider.WithTemperature(0.7),
	)

	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Response: %s\n", response.Content)
	fmt.Printf("Tokens: %d prompt + %d completion = %d total\n\n",
		response.Usage.PromptTokens,
		response.Usage.CompletionTokens,
		response.Usage.TotalTokens)
}

func example2StreamingChat(p *gemini.GeminiProvider) {
	fmt.Println("Example 2: Streaming Chat")
	fmt.Println("-------------------------")

	ctx := context.Background()
	messages := []provider.Message{
		{Role: "user", Content: "Write a haiku about artificial intelligence"},
	}

	eventChan, err := p.Stream(ctx, messages,
		provider.StreamWithModel("gemini-pro"),
		provider.StreamWithTemperature(0.8),
	)

	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	fmt.Print("Response: ")
	for event := range eventChan {
		switch event.Type {
		case provider.EventTypeContentDelta:
			fmt.Print(event.Content)
		case provider.EventTypeContentEnd:
			fmt.Println("\n")
		case provider.EventTypeError:
			log.Printf("Stream error: %v\n", event.Error)
		}
	}
}

func example3MultiTurn(p *gemini.GeminiProvider) {
	fmt.Println("Example 3: Multi-Turn Conversation")
	fmt.Println("----------------------------------")

	ctx := context.Background()
	messages := []provider.Message{
		{Role: "user", Content: "What is the capital of France?"},
		{Role: "assistant", Content: "The capital of France is Paris."},
		{Role: "user", Content: "What is the population of that city?"},
	}

	response, err := p.Chat(ctx, messages,
		provider.WithModel("gemini-pro"),
		provider.WithMaxTokens(150),
	)

	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Response: %s\n\n", response.Content)
}

func example4SystemPrompt(p *gemini.GeminiProvider) {
	fmt.Println("Example 4: Using System Prompt")
	fmt.Println("------------------------------")

	ctx := context.Background()
	messages := []provider.Message{
		{Role: "user", Content: "Explain quantum entanglement"},
	}

	response, err := p.Chat(ctx, messages,
		provider.WithModel("gemini-pro"),
		provider.WithSystemPrompt("You are a physics teacher explaining concepts to high school students. Use simple analogies and avoid jargon."),
		provider.WithMaxTokens(200),
		provider.WithTemperature(0.5),
	)

	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Response: %s\n\n", response.Content)
}

func example5DifferentModels(p *gemini.GeminiProvider) {
	fmt.Println("Example 5: Different Models")
	fmt.Println("---------------------------")

	ctx := context.Background()
	prompt := "In one sentence, what is Go programming language?"

	models := []string{"gemini-pro", "gemini-1.5-pro", "gemini-1.5-flash"}

	for _, model := range models {
		messages := []provider.Message{
			{Role: "user", Content: prompt},
		}

		response, err := p.Chat(ctx, messages,
			provider.WithModel(model),
			provider.WithMaxTokens(100),
		)

		if err != nil {
			log.Printf("Error with model %s: %v\n", model, err)
			continue
		}

		fmt.Printf("%s: %s\n", model, response.Content)
	}
	fmt.Println()
}
