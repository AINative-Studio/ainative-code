package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/AINative-studio/ainative-code/internal/logger"
	"github.com/AINative-studio/ainative-code/internal/provider"
)

const (
	// AnthropicAPIURL is the base URL for the Anthropic API
	AnthropicAPIURL = "https://api.anthropic.com/v1"

	// AnthropicAPIVersion is the API version header value
	AnthropicAPIVersion = "2023-06-01"
)

// Supported Claude models (as of January 2026)
// Note: Claude 3.5 series was retired on January 5, 2026
var supportedModels = []string{
	// Claude 4.5 series (current, recommended)
	"claude-sonnet-4-5-20250929", // Recommended: Best balance of intelligence, speed, and cost
	"claude-haiku-4-5-20251001",  // Fast and cost-effective
	"claude-opus-4-1",            // Premium model for complex tasks

	// Model aliases (auto-update to latest version)
	"claude-sonnet-4-5",
	"claude-haiku-4-5",

	// Legacy Claude 3.x models (deprecated/retired - kept for backwards compatibility)
	// These will likely fail with not_found_error from the API
	"claude-3-5-sonnet-20241022", // RETIRED: January 5, 2026
	"claude-3-5-haiku-20241022",  // RETIRED: January 5, 2026
	"claude-3-opus-20240229",     // DEPRECATED: Use claude-opus-4-1 instead
	"claude-3-haiku-20240307",    // DEPRECATED: Use claude-haiku-4-5 instead
	"claude-3-sonnet-20240229",   // DEPRECATED: Use claude-sonnet-4-5 instead
}

// AnthropicProvider implements the Provider interface for Anthropic's Claude API
type AnthropicProvider struct {
	*provider.BaseProvider
	apiKey  string
	baseURL string
}

// Config contains configuration for the Anthropic provider
type Config struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
	Logger     logger.LoggerInterface
}

// NewAnthropicProvider creates a new Anthropic provider instance
func NewAnthropicProvider(config Config) (*AnthropicProvider, error) {
	if config.APIKey == "" {
		return nil, provider.NewAuthenticationError("anthropic", "API key is required")
	}

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = AnthropicAPIURL
	}

	baseProvider := provider.NewBaseProvider(provider.BaseProviderConfig{
		Name:       "anthropic",
		HTTPClient: config.HTTPClient,
		Logger:     config.Logger,
		RetryConfig: provider.DefaultRetryConfig(),
	})

	return &AnthropicProvider{
		BaseProvider: baseProvider,
		apiKey:       config.APIKey,
		baseURL:      baseURL,
	}, nil
}

// Name returns the provider name
func (a *AnthropicProvider) Name() string {
	return a.BaseProvider.Name()
}

// Models returns the list of supported models
func (a *AnthropicProvider) Models() []string {
	models := make([]string, len(supportedModels))
	copy(models, supportedModels)
	return models
}

// Chat sends a chat request to the Anthropic API
func (a *AnthropicProvider) Chat(ctx context.Context, messages []provider.Message, opts ...provider.ChatOption) (provider.Response, error) {
	// Apply options
	options := provider.DefaultChatOptions()
	provider.ApplyChatOptions(options, opts...)

	// Validate model
	if err := a.ValidateModel(options.Model, supportedModels); err != nil {
		return provider.Response{}, err
	}

	// Build request
	req, err := a.buildRequest(ctx, messages, options, false)
	if err != nil {
		return provider.Response{}, provider.NewProviderError("anthropic", options.Model, err)
	}

	// Execute request
	resp, err := a.DoRequest(ctx, req)
	if err != nil {
		return provider.Response{}, err
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return provider.Response{}, provider.NewProviderError("anthropic", options.Model, fmt.Errorf("failed to read response: %w", err))
	}

	// Handle errors
	if resp.StatusCode != http.StatusOK {
		return provider.Response{}, a.handleAPIError(resp, body, options.Model)
	}

	// Parse response
	return a.parseResponse(body, options.Model)
}

// Stream sends a streaming chat request to the Anthropic API
func (a *AnthropicProvider) Stream(ctx context.Context, messages []provider.Message, opts ...provider.StreamOption) (<-chan provider.Event, error) {
	// Apply options
	options := provider.DefaultChatOptions()
	provider.ApplyStreamOptions(options, opts...)

	// Validate model
	if err := a.ValidateModel(options.Model, supportedModels); err != nil {
		return nil, err
	}

	// Build request
	req, err := a.buildRequest(ctx, messages, options, true)
	if err != nil {
		return nil, provider.NewProviderError("anthropic", options.Model, err)
	}

	// Execute request
	resp, err := a.DoRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	// Handle errors
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, a.handleAPIError(resp, body, options.Model)
	}

	// Create event channel
	eventChan := make(chan provider.Event)

	// Start streaming goroutine
	go a.streamResponse(ctx, resp.Body, eventChan, options.Model)

	return eventChan, nil
}

// Close releases resources held by the provider
func (a *AnthropicProvider) Close() error {
	return a.BaseProvider.Close()
}

// buildRequest constructs an HTTP request for the Anthropic API
func (a *AnthropicProvider) buildRequest(ctx context.Context, messages []provider.Message, options *provider.ChatOptions, stream bool) (*http.Request, error) {
	// Convert messages to Anthropic format
	apiMessages, systemPrompt := a.convertMessages(messages, options.SystemPrompt)

	// Build request body
	reqBody := anthropicRequest{
		Model:       options.Model,
		Messages:    apiMessages,
		MaxTokens:   options.MaxTokens,
		Stream:      stream,
	}

	// Add optional fields
	if systemPrompt != "" {
		reqBody.System = systemPrompt
	}
	if options.Temperature > 0 {
		reqBody.Temperature = &options.Temperature
	}
	if options.TopP > 0 && options.TopP < 1.0 {
		reqBody.TopP = &options.TopP
	}
	if len(options.StopSequences) > 0 {
		reqBody.StopSequences = options.StopSequences
	}

	// Marshal request body
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/messages", a.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", a.apiKey)
	req.Header.Set("anthropic-version", AnthropicAPIVersion)

	return req, nil
}

// convertMessages converts provider messages to Anthropic API format
func (a *AnthropicProvider) convertMessages(messages []provider.Message, systemPrompt string) ([]anthropicMessage, string) {
	var apiMessages []anthropicMessage
	var extractedSystem string

	for _, msg := range messages {
		// Extract system messages separately
		if msg.Role == "system" {
			if extractedSystem != "" {
				extractedSystem += "\n\n"
			}
			extractedSystem += msg.Content
			continue
		}

		apiMessages = append(apiMessages, anthropicMessage{
			Role: msg.Role,
			Content: []anthropicContent{
				{
					Type: "text",
					Text: msg.Content,
				},
			},
		})
	}

	// Combine extracted system messages with provided system prompt
	finalSystem := extractedSystem
	if systemPrompt != "" {
		if finalSystem != "" {
			finalSystem += "\n\n"
		}
		finalSystem += systemPrompt
	}

	return apiMessages, finalSystem
}

// parseResponse parses the Anthropic API response
func (a *AnthropicProvider) parseResponse(body []byte, model string) (provider.Response, error) {
	var apiResp anthropicResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return provider.Response{}, provider.NewProviderError("anthropic", model, fmt.Errorf("failed to parse response: %w", err))
	}

	// Extract text content
	var content string
	for _, block := range apiResp.Content {
		if block.Type == "text" {
			if content != "" {
				content += "\n"
			}
			content += block.Text
		}
	}

	return provider.Response{
		Content: content,
		Model:   apiResp.Model,
		Usage: provider.Usage{
			PromptTokens:     apiResp.Usage.InputTokens,
			CompletionTokens: apiResp.Usage.OutputTokens,
			TotalTokens:      apiResp.Usage.InputTokens + apiResp.Usage.OutputTokens,
		},
	}, nil
}

// eventResult holds the result of reading an SSE event
type eventResult struct {
	event *sseEvent
	err   error
}

// streamResponse handles streaming SSE responses from the Anthropic API
func (a *AnthropicProvider) streamResponse(ctx context.Context, body io.ReadCloser, eventChan chan<- provider.Event, model string) {
	defer close(eventChan)
	defer body.Close()

	reader := newSSEReader(body)
	var currentText string

	for {
		// Run readEvent in goroutine to allow context cancellation
		resultChan := make(chan eventResult, 1)
		go func() {
			event, err := reader.readEvent()
			resultChan <- eventResult{event: event, err: err}
		}()

		// Wait for either context cancellation or event result
		var event *sseEvent
		var err error
		select {
		case <-ctx.Done():
			eventChan <- provider.Event{
				Type:  provider.EventTypeError,
				Error: ctx.Err(),
			}
			return
		case result := <-resultChan:
			event = result.event
			err = result.err
		}

		// Handle read errors
		if err != nil {
			if err != io.EOF {
				eventChan <- provider.Event{
					Type:  provider.EventTypeError,
					Error: provider.NewProviderError("anthropic", model, err),
				}
			}
			return
		}

		// Handle different event types
		switch event.eventType {
		case "message_start":
			eventChan <- provider.Event{
				Type: provider.EventTypeContentStart,
			}

		case "content_block_delta":
			var delta contentBlockDelta
			if err := json.Unmarshal([]byte(event.data), &delta); err != nil {
				continue
			}
			if delta.Delta.Type == "text_delta" {
				currentText += delta.Delta.Text
				eventChan <- provider.Event{
					Type:    provider.EventTypeContentDelta,
					Content: delta.Delta.Text,
				}
			}

		case "message_delta":
			// Handle usage updates if needed
			continue

		case "message_stop":
			eventChan <- provider.Event{
				Type:    provider.EventTypeContentEnd,
				Content: currentText,
				Done:    true,
			}
			return

		case "error":
			var errResp anthropicError
			if err := json.Unmarshal([]byte(event.data), &errResp); err != nil {
				eventChan <- provider.Event{
					Type:  provider.EventTypeError,
					Error: provider.NewProviderError("anthropic", model, fmt.Errorf("stream error: %s", event.data)),
				}
			} else {
				eventChan <- provider.Event{
					Type:  provider.EventTypeError,
					Error: a.convertAPIError(&errResp, model),
				}
			}
			return
		}
	}
}

// handleAPIError converts Anthropic API errors to provider errors
func (a *AnthropicProvider) handleAPIError(resp *http.Response, body []byte, model string) error {
	var apiErr anthropicError
	if err := json.Unmarshal(body, &apiErr); err != nil {
		return a.HandleHTTPError(resp, body)
	}

	return a.convertAPIError(&apiErr, model)
}

// convertAPIError converts an Anthropic API error to a provider error
func (a *AnthropicProvider) convertAPIError(apiErr *anthropicError, model string) error {
	switch apiErr.Error.Type {
	case "authentication_error", "permission_error":
		return provider.NewAuthenticationError("anthropic", apiErr.Error.Message)

	case "rate_limit_error":
		return provider.NewRateLimitError("anthropic", 0)

	case "not_found_error":
		// Model not found - likely deprecated/retired
		// Provide helpful message suggesting current models
		return provider.NewInvalidModelError("anthropic", model, []string{
			"claude-sonnet-4-5-20250929 (recommended)",
			"claude-haiku-4-5-20251001",
			"claude-opus-4-1",
			"claude-sonnet-4-5",
			"claude-haiku-4-5",
		})

	case "invalid_request_error":
		// Check for context length errors
		if strings.Contains(apiErr.Error.Message, "prompt is too long") ||
			strings.Contains(apiErr.Error.Message, "max_tokens") {
			return provider.NewContextLengthError("anthropic", model, 0, 0)
		}
		return provider.NewProviderError("anthropic", model, fmt.Errorf("invalid request: %s", apiErr.Error.Message))

	default:
		return provider.NewProviderError("anthropic", model, fmt.Errorf("%s: %s", apiErr.Error.Type, apiErr.Error.Message))
	}
}
