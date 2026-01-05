package openai

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
	// OpenAIAPIURL is the base URL for the OpenAI API
	OpenAIAPIURL = "https://api.openai.com/v1"

	// Default API version
	defaultAPIVersion = "v1"
)

// Supported OpenAI models
var supportedModels = []string{
	// GPT-4 Turbo models
	"gpt-4-turbo-preview",
	"gpt-4-0125-preview",
	"gpt-4-1106-preview",

	// GPT-4 models
	"gpt-4",
	"gpt-4-0613",
	"gpt-4-32k",
	"gpt-4-32k-0613",

	// GPT-3.5 Turbo models
	"gpt-3.5-turbo",
	"gpt-3.5-turbo-0125",
	"gpt-3.5-turbo-1106",
	"gpt-3.5-turbo-16k",
	"gpt-3.5-turbo-16k-0613",
}

// OpenAIProvider implements the Provider interface for OpenAI's GPT models
type OpenAIProvider struct {
	*provider.BaseProvider
	apiKey       string
	baseURL      string
	organization string
}

// Config contains configuration for the OpenAI provider
type Config struct {
	APIKey       string
	BaseURL      string
	Organization string // Optional organization ID
	HTTPClient   *http.Client
	Logger       logger.LoggerInterface
}

// NewOpenAIProvider creates a new OpenAI provider instance
func NewOpenAIProvider(config Config) (*OpenAIProvider, error) {
	if config.APIKey == "" {
		return nil, provider.NewAuthenticationError("openai", "API key is required")
	}

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = OpenAIAPIURL
	}

	baseProvider := provider.NewBaseProvider(provider.BaseProviderConfig{
		Name:        "openai",
		HTTPClient:  config.HTTPClient,
		Logger:      config.Logger,
		RetryConfig: provider.DefaultRetryConfig(),
	})

	return &OpenAIProvider{
		BaseProvider: baseProvider,
		apiKey:       config.APIKey,
		baseURL:      baseURL,
		organization: config.Organization,
	}, nil
}

// Name returns the provider name
func (o *OpenAIProvider) Name() string {
	return o.BaseProvider.Name()
}

// Models returns the list of supported models
func (o *OpenAIProvider) Models() []string {
	models := make([]string, len(supportedModels))
	copy(models, supportedModels)
	return models
}

// Chat sends a chat request to the OpenAI API
func (o *OpenAIProvider) Chat(ctx context.Context, messages []provider.Message, opts ...provider.ChatOption) (provider.Response, error) {
	// Apply options
	options := provider.DefaultChatOptions()
	provider.ApplyChatOptions(options, opts...)

	// Validate model
	if err := o.ValidateModel(options.Model, supportedModels); err != nil {
		return provider.Response{}, err
	}

	// Build request
	req, err := o.buildRequest(ctx, messages, options, false)
	if err != nil {
		return provider.Response{}, provider.NewProviderError("openai", options.Model, err)
	}

	// Execute request
	resp, err := o.DoRequest(ctx, req)
	if err != nil {
		return provider.Response{}, err
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return provider.Response{}, provider.NewProviderError("openai", options.Model, fmt.Errorf("failed to read response: %w", err))
	}

	// Handle errors
	if resp.StatusCode != http.StatusOK {
		return provider.Response{}, o.handleAPIError(resp, body, options.Model)
	}

	// Parse response
	return o.parseResponse(body, options.Model)
}

// Stream sends a streaming chat request to the OpenAI API
func (o *OpenAIProvider) Stream(ctx context.Context, messages []provider.Message, opts ...provider.StreamOption) (<-chan provider.Event, error) {
	// Apply options
	options := provider.DefaultChatOptions()
	provider.ApplyStreamOptions(options, opts...)

	// Validate model
	if err := o.ValidateModel(options.Model, supportedModels); err != nil {
		return nil, err
	}

	// Build request
	req, err := o.buildRequest(ctx, messages, options, true)
	if err != nil {
		return nil, provider.NewProviderError("openai", options.Model, err)
	}

	// Execute request
	resp, err := o.DoRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	// Handle errors
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, o.handleAPIError(resp, body, options.Model)
	}

	// Create event channel
	eventChan := make(chan provider.Event)

	// Start streaming goroutine
	go o.streamResponse(ctx, resp.Body, eventChan, options.Model)

	return eventChan, nil
}

// Close releases resources held by the provider
func (o *OpenAIProvider) Close() error {
	return o.BaseProvider.Close()
}

// buildRequest constructs an HTTP request for the OpenAI API
func (o *OpenAIProvider) buildRequest(ctx context.Context, messages []provider.Message, options *provider.ChatOptions, stream bool) (*http.Request, error) {
	// Convert messages to OpenAI format
	apiMessages := o.convertMessages(messages, options.SystemPrompt)

	// Build request body
	reqBody := openAIRequest{
		Model:    options.Model,
		Messages: apiMessages,
		Stream:   stream,
	}

	// Add optional fields
	if options.MaxTokens > 0 {
		reqBody.MaxTokens = options.MaxTokens
	}
	if options.Temperature > 0 {
		reqBody.Temperature = &options.Temperature
	}
	if options.TopP > 0 && options.TopP < 1.0 {
		reqBody.TopP = &options.TopP
	}
	if len(options.StopSequences) > 0 {
		reqBody.Stop = options.StopSequences
	}

	// Marshal request body
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/chat/completions", o.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", o.apiKey))

	if o.organization != "" {
		req.Header.Set("OpenAI-Organization", o.organization)
	}

	return req, nil
}

// convertMessages converts provider messages to OpenAI API format
func (o *OpenAIProvider) convertMessages(messages []provider.Message, systemPrompt string) []openAIMessage {
	var apiMessages []openAIMessage

	// Add system prompt if provided
	if systemPrompt != "" {
		apiMessages = append(apiMessages, openAIMessage{
			Role:    "system",
			Content: systemPrompt,
		})
	}

	// Convert messages
	for _, msg := range messages {
		// OpenAI supports system messages in the messages array
		apiMessages = append(apiMessages, openAIMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	return apiMessages
}

// parseResponse parses the OpenAI API response
func (o *OpenAIProvider) parseResponse(body []byte, model string) (provider.Response, error) {
	var apiResp openAIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return provider.Response{}, provider.NewProviderError("openai", model, fmt.Errorf("failed to parse response: %w", err))
	}

	// Check if we have choices
	if len(apiResp.Choices) == 0 {
		return provider.Response{}, provider.NewProviderError("openai", model, fmt.Errorf("no choices in response"))
	}

	// Extract content from first choice
	content := ""
	if apiResp.Choices[0].Message.Content != nil {
		switch v := apiResp.Choices[0].Message.Content.(type) {
		case string:
			content = v
		case []interface{}:
			// Handle multi-modal content
			for _, part := range v {
				if partMap, ok := part.(map[string]interface{}); ok {
					if text, ok := partMap["text"].(string); ok {
						content += text
					}
				}
			}
		}
	}

	return provider.Response{
		Content: content,
		Model:   apiResp.Model,
		Usage: provider.Usage{
			PromptTokens:     apiResp.Usage.PromptTokens,
			CompletionTokens: apiResp.Usage.CompletionTokens,
			TotalTokens:      apiResp.Usage.TotalTokens,
		},
	}, nil
}

// eventResult holds the result of reading an SSE event
type eventResult struct {
	event *streamEvent
	err   error
}

// streamResponse handles streaming SSE responses from the OpenAI API
func (o *OpenAIProvider) streamResponse(ctx context.Context, body io.ReadCloser, eventChan chan<- provider.Event, model string) {
	defer close(eventChan)
	defer body.Close()

	reader := newSSEReader(body)
	var currentText string

	// Send start event
	eventChan <- provider.Event{
		Type: provider.EventTypeContentStart,
	}

	for {
		// Run readEvent in goroutine to allow context cancellation
		resultChan := make(chan eventResult, 1)
		go func() {
			event, err := reader.readEvent()
			resultChan <- eventResult{event: event, err: err}
		}()

		// Wait for either context cancellation or event result
		var event *streamEvent
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
					Error: provider.NewProviderError("openai", model, err),
				}
			}
			return
		}

		// OpenAI uses "[DONE]" to signal end of stream
		if event.data == "[DONE]" {
			eventChan <- provider.Event{
				Type:    provider.EventTypeContentEnd,
				Content: currentText,
				Done:    true,
			}
			return
		}

		// Parse the chunk
		var chunk openAIStreamResponse
		if err := json.Unmarshal([]byte(event.data), &chunk); err != nil {
			// Skip unparseable chunks
			continue
		}

		// Process choices
		if len(chunk.Choices) > 0 {
			choice := chunk.Choices[0]

			// Check for finish
			if choice.FinishReason != nil && *choice.FinishReason != "" {
				eventChan <- provider.Event{
					Type:    provider.EventTypeContentEnd,
					Content: currentText,
					Done:    true,
				}
				return
			}

			// Extract delta content
			if choice.Delta.Content != "" {
				currentText += choice.Delta.Content
				eventChan <- provider.Event{
					Type:    provider.EventTypeContentDelta,
					Content: choice.Delta.Content,
				}
			}
		}
	}
}

// handleAPIError converts OpenAI API errors to provider errors
func (o *OpenAIProvider) handleAPIError(resp *http.Response, body []byte, model string) error {
	var apiErr openAIError
	if err := json.Unmarshal(body, &apiErr); err != nil {
		return o.HandleHTTPError(resp, body)
	}

	return o.convertAPIError(&apiErr, resp.StatusCode, model)
}

// convertAPIError converts an OpenAI API error to a provider error
func (o *OpenAIProvider) convertAPIError(apiErr *openAIError, statusCode int, model string) error {
	errType := apiErr.Error.Type
	errMsg := apiErr.Error.Message

	switch {
	case statusCode == http.StatusUnauthorized || statusCode == http.StatusForbidden:
		return provider.NewAuthenticationError("openai", errMsg)

	case statusCode == http.StatusTooManyRequests:
		return provider.NewRateLimitError("openai", 0)

	case statusCode == http.StatusBadRequest:
		// Check for context length errors
		if strings.Contains(errMsg, "maximum context length") ||
			strings.Contains(errMsg, "context_length_exceeded") ||
			strings.Contains(errType, "context_length_exceeded") {
			return provider.NewContextLengthError("openai", model, 0, 0)
		}
		return provider.NewProviderError("openai", model, fmt.Errorf("invalid request: %s", errMsg))

	case statusCode == http.StatusNotFound:
		// Model not found
		if strings.Contains(errMsg, "model") {
			return provider.NewInvalidModelError("openai", model, supportedModels)
		}
		return provider.NewProviderError("openai", model, fmt.Errorf("not found: %s", errMsg))

	default:
		return provider.NewProviderError("openai", model, fmt.Errorf("%s: %s", errType, errMsg))
	}
}
