package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/AINative-studio/ainative-code/internal/backend"
	"github.com/AINative-studio/ainative-code/internal/logger"
	llmprovider "github.com/AINative-studio/ainative-code/internal/provider"
	"github.com/AINative-studio/ainative-code/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	// Validate message early if single message mode to avoid unnecessary API calls
	if len(args) > 0 {
		message := args[0]
		// Validate message is not empty or only whitespace
		if strings.TrimSpace(message) == "" {
			return fmt.Errorf("Error: message cannot be empty")
		}
	}

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

// runInteractiveChat starts an interactive chat session with TUI
func runInteractiveChat(ctx context.Context, aiProvider llmprovider.Provider, modelName string) error {
	logger.Info("Starting interactive chat mode with TUI")

	// Initialize TUI model
	model := tui.NewModel()

	// Add initial system message if provided
	if chatSystemMsg != "" {
		model.AddMessage("system", chatSystemMsg)
	}

	// Create bubbletea program with alt screen
	p := tea.NewProgram(
		&interactiveChatModel{
			tuiModel:  model,
			provider:  aiProvider,
			modelName: modelName,
			ctx:       ctx,
			messages:  []llmprovider.Message{},
			systemMsg: chatSystemMsg,
		},
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	// Run the program
	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("error running TUI: %w", err)
	}

	// Check if there was an error during execution
	if chatModel, ok := finalModel.(*interactiveChatModel); ok {
		if chatModel.err != nil {
			return chatModel.err
		}
	}

	return nil
}

// interactiveChatModel wraps the TUI model with chat functionality
type interactiveChatModel struct {
	tuiModel         tui.Model
	provider         llmprovider.Provider
	modelName        string
	ctx              context.Context
	messages         []llmprovider.Message
	systemMsg        string
	err              error
	waitingForAI     bool
	lastUserInput    string
	streamingContent string
}

// Init initializes the interactive chat model
func (m *interactiveChatModel) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		tui.SendReady(),
	)
}

// Update handles messages and updates the model
func (m *interactiveChatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Handle custom streaming messages first
	switch msg := msg.(type) {
	case streamStartMsg:
		// Start processing stream events
		m.tuiModel.SetStreaming(true)
		m.streamingContent = "" // Reset streaming content
		return m, m.handleStreamEvents(msg.eventChan)

	case streamChunkMsg:
		// Track the streaming content
		m.streamingContent += msg.content

		// Add chunk to TUI
		// Send the chunk as a stream chunk message to TUI
		tuiModel, tuiCmd := m.tuiModel.Update(tui.SendStreamChunk(msg.content)())
		if tuiModelTyped, ok := tuiModel.(tui.Model); ok {
			m.tuiModel = tuiModelTyped
		}
		// Continue processing stream with the event channel
		return m, tea.Batch(tuiCmd, m.handleStreamEvents(msg.eventChan))

	case streamCompleteMsg:
		// Streaming complete
		m.tuiModel.SetStreaming(false)
		m.waitingForAI = false

		// Add assistant response to history
		// Use streaming content if we were streaming, otherwise use msg.content (non-streaming)
		content := msg.content
		if content == "" && m.streamingContent != "" {
			content = m.streamingContent
		}

		if content != "" {
			m.messages = append(m.messages, llmprovider.Message{
				Role:    "assistant",
				Content: content,
			})
		}

		// Clear streaming content
		m.streamingContent = ""

		// Send stream done to TUI
		tuiModel, tuiCmd := m.tuiModel.Update(tui.SendStreamDone()())
		if tuiModelTyped, ok := tuiModel.(tui.Model); ok {
			m.tuiModel = tuiModelTyped
		}
		return m, tuiCmd

	case streamErrorMsg:
		// Handle streaming error
		m.err = msg.err
		m.tuiModel.SetError(msg.err)
		m.tuiModel.SetStreaming(false)
		m.waitingForAI = false

		// Send error to TUI
		tuiModel, tuiCmd := m.tuiModel.Update(tui.SendError(msg.err)())
		if tuiModelTyped, ok := tuiModel.(tui.Model); ok {
			m.tuiModel = tuiModelTyped
		}
		return m, tuiCmd
	}

	// Capture user input before it's cleared by TUI
	var userInput string
	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "enter" && !m.tuiModel.IsStreaming() {
		userInput = strings.TrimSpace(m.tuiModel.GetUserInput())
	}

	// Let the TUI model handle the message
	tuiModel, tuiCmd := m.tuiModel.Update(msg)
	if tuiModelTyped, ok := tuiModel.(tui.Model); ok {
		m.tuiModel = tuiModelTyped
	}
	if tuiCmd != nil {
		cmds = append(cmds, tuiCmd)
	}

	// If we captured user input, process it now
	if userInput != "" && userInput != m.lastUserInput {
		m.lastUserInput = userInput

		// Add to conversation history
		m.messages = append(m.messages, llmprovider.Message{
			Role:    "user",
			Content: userInput,
		})

		// Start streaming AI response
		m.waitingForAI = true
		cmds = append(cmds, m.streamAIResponse())
	}

	// Check if user wants to quit
	if m.tuiModel.IsQuitting() {
		return m, tea.Quit
	}

	return m, tea.Batch(cmds...)
}

// View renders the chat interface
func (m *interactiveChatModel) View() string {
	return m.tuiModel.View()
}

// streamAIResponse streams the AI response and sends updates to the TUI
func (m *interactiveChatModel) streamAIResponse() tea.Cmd {
	return func() tea.Msg {
		// Prepare options
		opts := []llmprovider.ChatOption{
			llmprovider.WithModel(m.modelName),
		}

		if m.systemMsg != "" {
			opts = append(opts, llmprovider.WithSystemPrompt(m.systemMsg))
		}

		if chatStream {
			// Convert to stream options
			streamOpts := make([]llmprovider.StreamOption, len(opts))
			for i, opt := range opts {
				streamOpts[i] = llmprovider.StreamOption(opt)
			}

			// Start streaming
			eventChan, err := m.provider.Stream(m.ctx, m.messages, streamOpts...)
			if err != nil {
				return streamErrorMsg{err: err}
			}

			// Process streaming events in this goroutine
			// Send each chunk as a separate message to the TUI
			return streamStartMsg{eventChan: eventChan}
		}

		// Non-streaming response
		resp, err := m.provider.Chat(m.ctx, m.messages, opts...)
		if err != nil {
			return streamErrorMsg{err: err}
		}

		return streamCompleteMsg{content: resp.Content}
	}
}

// Custom message types for streaming
type streamStartMsg struct {
	eventChan <-chan llmprovider.Event
}

type streamChunkMsg struct {
	content   string
	eventChan <-chan llmprovider.Event
}

type streamCompleteMsg struct {
	content string
}

type streamErrorMsg struct {
	err error
}

// handleStreamEvents creates a command that processes stream events
func (m *interactiveChatModel) handleStreamEvents(eventChan <-chan llmprovider.Event) tea.Cmd {
	return func() tea.Msg {
		// Read the next event from the channel
		event, ok := <-eventChan
		if !ok {
			// Channel closed, streaming is done
			return streamCompleteMsg{}
		}

		switch event.Type {
		case llmprovider.EventTypeContentDelta:
			return streamChunkMsg{content: event.Content, eventChan: eventChan}
		case llmprovider.EventTypeError:
			return streamErrorMsg{err: event.Error}
		case llmprovider.EventTypeContentEnd:
			// Stream is complete, return done message
			return streamCompleteMsg{}
		}

		// Continue processing
		return m.handleStreamEvents(eventChan)()
	}
}

// getDefaultModel returns the default model for a given provider
func getDefaultModel(providerName string) string {
	switch providerName {
	case "openai":
		return "gpt-4"
	case "anthropic":
		// Updated to Claude 4.5 series (Claude 3.5 retired January 5, 2026)
		return "claude-sonnet-4-5-20250929"
	case "ollama":
		return "llama2"
	default:
		return ""
	}
}

// newChatAINativeCmd creates a new chat command using AINative backend
func newChatAINativeCmd() *cobra.Command {
	var (
		message      string
		autoProvider bool
		modelName    string
		verbose      bool
	)

	cmd := &cobra.Command{
		Use:   "chat-ainative",
		Short: "Chat using AINative backend API",
		Long: `Send chat messages using the AINative backend with intelligent provider selection.

This command integrates with:
- AINative backend API for chat completions
- Provider selector for intelligent provider routing
- Credit management and warnings`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			// Validate message
			if message == "" && len(args) == 0 {
				return fmt.Errorf("Error: message cannot be empty")
			}

			// Use message from args if not provided via flag
			if message == "" && len(args) > 0 {
				message = strings.TrimSpace(args[0])
			}

			// Validate message is not empty or whitespace only
			if strings.TrimSpace(message) == "" {
				return fmt.Errorf("Error: message cannot be empty")
			}

			// Check authentication
			accessToken := viper.GetString("access_token")
			if accessToken == "" {
				return fmt.Errorf("not authenticated. Please run 'ainative-code auth login' first")
			}

			// Get backend URL from config
			backendURL := viper.GetString("backend_url")
			if backendURL == "" {
				backendURL = "http://localhost:8000"
			}

			// Provider selection logic
			var selectedModel string
			if autoProvider {
				// Use provider selector
				selector := llmprovider.NewSelector(
					llmprovider.WithProviders("anthropic", "openai", "google"),
					llmprovider.WithUserPreference(viper.GetString("preferred_provider")),
					llmprovider.WithCreditThreshold(50),
					llmprovider.WithFallback(viper.GetBool("fallback_enabled")),
				)

				user := &llmprovider.User{
					Email:   viper.GetString("user_email"),
					Credits: viper.GetInt("credits"),
					Tier:    viper.GetString("tier"),
				}

				provider, err := selector.Select(ctx, user)
				if err != nil {
					return fmt.Errorf("provider selection failed: %w", err)
				}

				// Display low credit warning if applicable
				if provider.LowCreditWarning {
					fmt.Fprintf(cmd.ErrOrStderr(), "Warning: Low credit balance (%d credits remaining)\n", user.Credits)
				}

				// Use provider's default model if not specified
				if modelName == "" {
					selectedModel = getDefaultModel(provider.Name)
				} else {
					selectedModel = modelName
				}
			} else {
				// Use specified model or default
				if modelName == "" {
					selectedModel = "claude-sonnet-4-5-20250929"
				} else {
					selectedModel = modelName
				}
			}

			// Create backend client
			client := backend.NewClient(backendURL)

			// Prepare chat request
			req := &backend.ChatCompletionRequest{
				Messages: []backend.Message{
					{
						Role:    "user",
						Content: message,
					},
				},
				Model: selectedModel,
			}

			// Send chat completion request
			resp, err := client.ChatCompletion(ctx, accessToken, req)
			if err != nil {
				return fmt.Errorf("chat request failed: %w", err)
			}

			// Display response
			if len(resp.Choices) > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "%s\n", resp.Choices[0].Message.Content)
			}

			// Display usage stats if verbose
			if verbose {
				fmt.Fprintf(cmd.ErrOrStderr(), "\nModel: %s\n", resp.Model)
				fmt.Fprintf(cmd.ErrOrStderr(), "Tokens - Prompt: %d, Completion: %d, Total: %d\n",
					resp.Usage.PromptTokens,
					resp.Usage.CompletionTokens,
					resp.Usage.TotalTokens)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&message, "message", "m", "", "Message to send (required)")
	cmd.Flags().BoolVar(&autoProvider, "auto-provider", false, "Auto-select provider based on preferences")
	cmd.Flags().StringVar(&modelName, "model", "", "Model to use (default: auto-selected)")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "Display usage statistics")

	return cmd
}
