package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/AINative-studio/ainative-code/internal/logger"
	"github.com/AINative-studio/ainative-code/internal/provider"
)

const (
	// OllamaAPIVersion is the Ollama API endpoint version
	OllamaAPIVersion = "v1"
)

// OllamaProvider implements the Provider interface for Ollama
type OllamaProvider struct {
	*provider.BaseProvider
	config *Config
}

// NewOllamaProvider creates a new Ollama provider instance
func NewOllamaProvider(config Config) (*OllamaProvider, error) {
	// Set defaults
	config.SetDefaults()

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid ollama config: %w", err)
	}

	// Create base provider
	baseProvider := provider.NewBaseProvider(provider.BaseProviderConfig{
		Name:       "ollama",
		HTTPClient: config.HTTPClient,
		Logger:     config.Logger,
		RetryConfig: provider.RetryConfig{
			MaxRetries:     2, // Fewer retries for local server
			InitialBackoff: 500 * 1000000,  // 500ms
			MaxBackoff:     2000 * 1000000,  // 2s
			Multiplier:     2.0,
			RetryableStatusCodes: []int{
				http.StatusServiceUnavailable,
				http.StatusTooManyRequests,
			},
		},
	})

	return &OllamaProvider{
		BaseProvider: baseProvider,
		config:       &config,
	}, nil
}

// Name returns the provider name
func (o *OllamaProvider) Name() string {
	return o.BaseProvider.Name()
}

// Models returns the list of well-known supported models
func (o *OllamaProvider) Models() []string {
	return GetSupportedModelNames()
}

// Chat sends a chat request to Ollama
func (o *OllamaProvider) Chat(ctx context.Context, messages []provider.Message, opts ...provider.ChatOption) (provider.Response, error) {
	// Apply options
	options := provider.DefaultChatOptions()
	provider.ApplyChatOptions(options, opts...)

	// Use config model if not specified
	if options.Model == "" {
		options.Model = o.config.Model
	}

	// Build request
	req, err := o.buildChatRequest(ctx, messages, options, false)
	if err != nil {
		return provider.Response{}, provider.NewProviderError("ollama", options.Model, err)
	}

	// Execute request
	resp, err := o.DoRequest(ctx, req)
	if err != nil {
		// Check if it's a connection error
		if resp == nil {
			return provider.Response{}, NewOllamaConnectionError(o.config.BaseURL, err)
		}
		return provider.Response{}, err
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return provider.Response{}, provider.NewProviderError("ollama", options.Model, fmt.Errorf("failed to read response: %w", err))
	}

	// Handle errors
	if resp.StatusCode != http.StatusOK {
		return provider.Response{}, parseOllamaError(resp.StatusCode, body, options.Model)
	}

	// Parse response
	return o.parseResponse(body, options.Model)
}

// Stream sends a streaming chat request to Ollama
func (o *OllamaProvider) Stream(ctx context.Context, messages []provider.Message, opts ...provider.StreamOption) (<-chan provider.Event, error) {
	// Apply options
	options := provider.DefaultChatOptions()
	provider.ApplyStreamOptions(options, opts...)

	// Use config model if not specified
	if options.Model == "" {
		options.Model = o.config.Model
	}

	// Build request
	req, err := o.buildChatRequest(ctx, messages, options, true)
	if err != nil {
		return nil, provider.NewProviderError("ollama", options.Model, err)
	}

	// Execute request
	resp, err := o.DoRequest(ctx, req)
	if err != nil {
		// Check if it's a connection error
		if resp == nil {
			return nil, NewOllamaConnectionError(o.config.BaseURL, err)
		}
		return nil, err
	}

	// Handle errors
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, parseOllamaError(resp.StatusCode, body, options.Model)
	}

	// Create event channel
	eventChan := make(chan provider.Event, 100)

	// Start streaming goroutine
	go handleStreamResponse(ctx, resp.Body, eventChan, options.Model)

	return eventChan, nil
}

// Close releases resources held by the provider
func (o *OllamaProvider) Close() error {
	return o.BaseProvider.Close()
}

// buildChatRequest constructs an HTTP request for the Ollama chat API
func (o *OllamaProvider) buildChatRequest(ctx context.Context, messages []provider.Message, options *provider.ChatOptions, stream bool) (*http.Request, error) {
	// Build Ollama request
	ollamaReq := buildOllamaRequest(o.config, messages, options, stream)

	// Marshal request body
	jsonBody, err := json.Marshal(ollamaReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/api/chat", o.config.BaseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if stream {
		req.Header.Set("Accept", "application/x-ndjson")
	}

	return req, nil
}

// parseResponse parses the Ollama API response
func (o *OllamaProvider) parseResponse(body []byte, model string) (provider.Response, error) {
	var ollamaResp ollamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return provider.Response{}, provider.NewProviderError("ollama", model, fmt.Errorf("failed to parse response: %w", err))
	}

	// Check for errors in response
	if !ollamaResp.Done {
		return provider.Response{}, provider.NewProviderError("ollama", model, fmt.Errorf("incomplete response received"))
	}

	return provider.Response{
		Content: ollamaResp.Message.Content,
		Model:   ollamaResp.Model,
		Usage: provider.Usage{
			PromptTokens:     ollamaResp.PromptEvalCount,
			CompletionTokens: ollamaResp.EvalCount,
			TotalTokens:      ollamaResp.PromptEvalCount + ollamaResp.EvalCount,
		},
	}, nil
}

// HealthCheck verifies connectivity to Ollama server
func (o *OllamaProvider) HealthCheck(ctx context.Context) error {
	return o.config.HealthCheck(ctx)
}

// ListAvailableModels fetches the list of models from the Ollama server
func (o *OllamaProvider) ListAvailableModels(ctx context.Context) ([]ModelInfo, error) {
	return ListModels(ctx, o.config)
}

// GetModelInfo retrieves information about a specific model
func (o *OllamaProvider) GetModelInfo(ctx context.Context, modelName string) (*ModelInfo, error) {
	return GetModelInfo(ctx, o.config, modelName)
}

// IsModelAvailable checks if a model is currently available
func (o *OllamaProvider) IsModelAvailable(ctx context.Context, modelName string) (bool, error) {
	return IsModelAvailable(ctx, o.config, modelName)
}

// NewOllamaProviderWithLogger creates a provider with a custom logger
func NewOllamaProviderWithLogger(config Config, log logger.LoggerInterface) (*OllamaProvider, error) {
	config.Logger = log
	return NewOllamaProvider(config)
}

// NewOllamaProviderForModel creates a provider configured for a specific model
func NewOllamaProviderForModel(model string) (*OllamaProvider, error) {
	config := Config{
		Model: model,
	}
	return NewOllamaProvider(config)
}
