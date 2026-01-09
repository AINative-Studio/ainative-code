package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/AINative-studio/ainative-code/internal/logger"
	llmprovider "github.com/AINative-studio/ainative-code/internal/provider"
)

var (
	chatSessionID string
	chatSystemMsg string
	chatStream    bool
)

// chatCmd represents the chat command
var chatCmd = &cobra.Command{
	Use:   "chat [message]",
	Short: "Start interactive AI chat session",
	Long: `Start an interactive chat session with an AI model.

The chat command provides an interactive interface to communicate with AI models.
You can either provide a message as an argument for a single interaction, or
run without arguments to enter interactive mode.

Examples:
  # Start interactive chat mode
  ainative-code chat

  # Send a single message
  ainative-code chat "Explain how to use goroutines"

  # Continue a previous session
  ainative-code chat --session-id abc123

  # Use a specific model
  ainative-code chat --provider openai --model gpt-4`,
	Aliases: []string{"c", "ask"},
	RunE:    runChat,
}

func init() {
	rootCmd.AddCommand(chatCmd)

	// Chat-specific flags
	chatCmd.Flags().StringVarP(&chatSessionID, "session-id", "s", "", "resume a previous chat session")
	chatCmd.Flags().StringVar(&chatSystemMsg, "system", "", "custom system message")
	chatCmd.Flags().BoolVar(&chatStream, "stream", true, "stream responses in real-time")
}

func runChat(cmd *cobra.Command, args []string) error {
	providerName := GetProvider()
	modelName := GetModel()

	logger.DebugEvent().
		Str("provider", providerName).
		Str("model", modelName).
		Str("session_id", chatSessionID).
		Msg("Starting chat command")

	// Check if provider is configured
	if providerName == "" {
		return fmt.Errorf("AI provider not configured. Use --provider flag or set in config file")
	}

	// Set default model if not specified
	if modelName == "" {
		modelName = getDefaultModel(providerName)
		logger.DebugEvent().
			Str("provider", providerName).
			Str("model", modelName).
			Msg("Using default model for provider")
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Initialize provider
	aiProvider, err := initializeProvider(ctx, providerName, modelName)
	if err != nil {
		return fmt.Errorf("failed to initialize AI provider: %w", err)
	}
	defer aiProvider.Close()

	if len(args) > 0 {
		// Single message mode
		message := args[0]
		return runSingleMessage(ctx, aiProvider, modelName, message)
	}

	// Interactive mode
	return runInteractiveChat(ctx, aiProvider, modelName)
}

// runSingleMessage processes a single message and prints the response
func runSingleMessage(ctx context.Context, aiProvider llmprovider.Provider, modelName, message string) error {
	logger.InfoEvent().
		Str("model", modelName).
		Str("message", message).
		Msg("Processing single message")

	// Prepare messages
	messages := []llmprovider.Message{
		{
			Role:    "user",
			Content: message,
		},
	}

	// Add system message if provided
	var opts []llmprovider.ChatOption
	opts = append(opts, llmprovider.WithModel(modelName))

	if chatSystemMsg != "" {
		opts = append(opts, llmprovider.WithSystemPrompt(chatSystemMsg))
	}

	// Check if streaming is enabled
	if chatStream {
		return streamSingleMessage(ctx, aiProvider, messages, opts)
	}

	// Non-streaming response
	resp, err := aiProvider.Chat(ctx, messages, opts...)
	if err != nil {
		return fmt.Errorf("chat request failed: %w", err)
	}

	// Print response
	fmt.Println(resp.Content)

	// Print usage stats if verbose
	if GetVerbose() {
		fmt.Fprintf(os.Stderr, "\n---\n")
		fmt.Fprintf(os.Stderr, "Model: %s\n", resp.Model)
		fmt.Fprintf(os.Stderr, "Tokens - Prompt: %d, Completion: %d, Total: %d\n",
			resp.Usage.PromptTokens,
			resp.Usage.CompletionTokens,
			resp.Usage.TotalTokens)
	}

	return nil
}

// streamSingleMessage streams a single message response
func streamSingleMessage(ctx context.Context, aiProvider llmprovider.Provider, messages []llmprovider.Message, opts []llmprovider.ChatOption) error {
	// Convert ChatOptions to StreamOptions
	streamOpts := make([]llmprovider.StreamOption, len(opts))
	for i, opt := range opts {
		streamOpts[i] = llmprovider.StreamOption(opt)
	}

	eventChan, err := aiProvider.Stream(ctx, messages, streamOpts...)
	if err != nil {
		return fmt.Errorf("failed to start stream: %w", err)
	}

	// Process streaming events
	for event := range eventChan {
		switch event.Type {
		case llmprovider.EventTypeContentDelta:
			// Print delta content without newline
			fmt.Print(event.Content)
		case llmprovider.EventTypeError:
			return fmt.Errorf("streaming error: %w", event.Error)
		case llmprovider.EventTypeContentEnd:
			// Print final newline
			fmt.Println()
		}
	}

	return nil
}

// runInteractiveChat starts an interactive chat session
func runInteractiveChat(ctx context.Context, aiProvider llmprovider.Provider, modelName string) error {
	logger.Info("Starting interactive chat mode")

	fmt.Println("Interactive Chat Mode")
	fmt.Printf("Provider: %s\n", GetProvider())
	fmt.Printf("Model: %s\n", modelName)
	fmt.Println("Type 'exit' or 'quit' to end the session")
	fmt.Println("---")

	// Conversation history
	var messages []llmprovider.Message

	// Add system message if provided
	if chatSystemMsg != "" {
		fmt.Printf("System: %s\n---\n", chatSystemMsg)
	}

	// Simple interactive loop (basic implementation)
	// TODO: Replace with bubbletea for better UX
	for {
		fmt.Print("\nYou: ")

		// Read user input
		var input string
		_, err := fmt.Scanln(&input)
		if err != nil {
			// Handle multi-word input
			input = readLine()
		}

		input = strings.TrimSpace(input)

		// Check for exit commands
		if input == "exit" || input == "quit" {
			fmt.Println("Goodbye!")
			return nil
		}

		if input == "" {
			continue
		}

		// Add user message to history
		messages = append(messages, llmprovider.Message{
			Role:    "user",
			Content: input,
		})

		// Prepare options
		opts := []llmprovider.ChatOption{
			llmprovider.WithModel(modelName),
		}

		if chatSystemMsg != "" {
			opts = append(opts, llmprovider.WithSystemPrompt(chatSystemMsg))
		}

		// Get AI response
		fmt.Print("\nAssistant: ")

		if chatStream {
			// Stream response
			streamOpts := make([]llmprovider.StreamOption, len(opts))
			for i, opt := range opts {
				streamOpts[i] = llmprovider.StreamOption(opt)
			}

			eventChan, err := aiProvider.Stream(ctx, messages, streamOpts...)
			if err != nil {
				return fmt.Errorf("failed to start stream: %w", err)
			}

			var fullResponse string
			for event := range eventChan {
				switch event.Type {
				case llmprovider.EventTypeContentDelta:
					fmt.Print(event.Content)
					fullResponse += event.Content
				case llmprovider.EventTypeError:
					return fmt.Errorf("streaming error: %w", event.Error)
				case llmprovider.EventTypeContentEnd:
					fmt.Println()
				}
			}

			// Add assistant response to history
			messages = append(messages, llmprovider.Message{
				Role:    "assistant",
				Content: fullResponse,
			})
		} else {
			// Non-streaming response
			resp, err := aiProvider.Chat(ctx, messages, opts...)
			if err != nil {
				return fmt.Errorf("chat request failed: %w", err)
			}

			fmt.Println(resp.Content)

			// Add assistant response to history
			messages = append(messages, llmprovider.Message{
				Role:    "assistant",
				Content: resp.Content,
			})
		}
	}
}

// readLine reads a full line of input including spaces
func readLine() string {
	var line strings.Builder
	var char byte
	for {
		_, err := fmt.Scanf("%c", &char)
		if err != nil || char == '\n' {
			break
		}
		line.WriteByte(char)
	}
	return line.String()
}

// getDefaultModel returns the default model for a given provider
func getDefaultModel(providerName string) string {
	switch providerName {
	case "openai":
		return "gpt-4"
	case "anthropic":
		return "claude-3-5-sonnet-20241022"
	case "ollama":
		return "llama2"
	default:
		return ""
	}
}
