package bedrock

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/AINative-studio/ainative-code/internal/provider"
)

const (
	// BedrockRuntimeEndpoint is the default Bedrock Runtime endpoint pattern
	BedrockRuntimeEndpoint = "https://bedrock-runtime.%s.amazonaws.com"
)

// Supported Claude models on Bedrock
var supportedModels = []string{
	"anthropic.claude-3-5-sonnet-20241022-v2:0",
	"anthropic.claude-3-opus-20240229-v1:0",
	"anthropic.claude-3-sonnet-20240229-v1:0",
	"anthropic.claude-3-haiku-20240307-v1:0",
	"anthropic.claude-v2",
	"anthropic.claude-instant-v1",
}

// BedrockProvider implements the Provider interface for AWS Bedrock
type BedrockProvider struct {
	*provider.BaseProvider
	region   string
	endpoint string
	signer   *awsSigner
}

// NewBedrockProvider creates a new Bedrock provider instance
func NewBedrockProvider(config Config) (*BedrockProvider, error) {
	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, provider.NewAuthenticationError("bedrock", err.Error())
	}

	// Determine endpoint
	endpoint := config.Endpoint
	if endpoint == "" {
		endpoint = fmt.Sprintf(BedrockRuntimeEndpoint, config.Region)
	}

	// Create AWS signer
	signer := newAWSSigner(config.Region, config.AccessKey, config.SecretKey, config.SessionToken)

	// Create base provider
	baseProvider := provider.NewBaseProvider(provider.BaseProviderConfig{
		Name:       "bedrock",
		HTTPClient: config.HTTPClient,
		Logger:     config.Logger,
		RetryConfig: provider.DefaultRetryConfig(),
	})

	return &BedrockProvider{
		BaseProvider: baseProvider,
		region:       config.Region,
		endpoint:     endpoint,
		signer:       signer,
	}, nil
}

// Name returns the provider name
func (b *BedrockProvider) Name() string {
	return b.BaseProvider.Name()
}

// Models returns the list of supported models
func (b *BedrockProvider) Models() []string {
	models := make([]string, len(supportedModels))
	copy(models, supportedModels)
	return models
}

// Chat sends a chat request to the Bedrock API
func (b *BedrockProvider) Chat(ctx context.Context, messages []provider.Message, opts ...provider.ChatOption) (provider.Response, error) {
	// Apply options
	options := provider.DefaultChatOptions()
	provider.ApplyChatOptions(options, opts...)

	// Validate model
	if err := b.ValidateModel(options.Model, supportedModels); err != nil {
		return provider.Response{}, err
	}

	// Build request
	req, err := b.buildRequest(ctx, messages, options, false)
	if err != nil {
		return provider.Response{}, provider.NewProviderError("bedrock", options.Model, err)
	}

	// Execute request
	resp, err := b.DoRequest(ctx, req)
	if err != nil {
		return provider.Response{}, err
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return provider.Response{}, provider.NewProviderError("bedrock", options.Model, fmt.Errorf("failed to read response: %w", err))
	}

	// Handle errors
	if resp.StatusCode != http.StatusOK {
		return provider.Response{}, parseBedrockError(resp.StatusCode, body, options.Model)
	}

	// Parse response
	return b.parseResponse(body, options.Model)
}

// Stream sends a streaming chat request to the Bedrock API
func (b *BedrockProvider) Stream(ctx context.Context, messages []provider.Message, opts ...provider.StreamOption) (<-chan provider.Event, error) {
	// Apply options
	options := provider.DefaultChatOptions()
	provider.ApplyStreamOptions(options, opts...)

	// Validate model
	if err := b.ValidateModel(options.Model, supportedModels); err != nil {
		return nil, err
	}

	// Build request
	req, err := b.buildRequest(ctx, messages, options, true)
	if err != nil {
		return nil, provider.NewProviderError("bedrock", options.Model, err)
	}

	// Execute request
	resp, err := b.DoRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	// Handle errors
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, parseBedrockError(resp.StatusCode, body, options.Model)
	}

	// Create event channel
	eventChan := make(chan provider.Event)

	// Start streaming goroutine
	go parseStreamingEvents(ctx, resp.Body, eventChan, options.Model)

	return eventChan, nil
}

// Close releases resources held by the provider
func (b *BedrockProvider) Close() error {
	return b.BaseProvider.Close()
}

// buildRequest constructs an HTTP request for the Bedrock API
func (b *BedrockProvider) buildRequest(ctx context.Context, messages []provider.Message, options *provider.ChatOptions, stream bool) (*http.Request, error) {
	// Build request body
	reqBody := buildBedrockRequest(messages, options)

	// Marshal request body
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Build URL
	url := b.buildInvokeURL(options.Model, stream)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if stream {
		req.Header.Set("Accept", "application/vnd.amazon.eventstream")
	}

	// Sign request with AWS Signature V4
	if err := b.signer.signRequest(req, jsonBody, time.Now()); err != nil {
		return nil, fmt.Errorf("failed to sign request: %w", err)
	}

	return req, nil
}

// buildInvokeURL constructs the invoke URL for a model
func (b *BedrockProvider) buildInvokeURL(model string, stream bool) string {
	path := fmt.Sprintf("/model/%s/invoke", model)
	if stream {
		path = fmt.Sprintf("/model/%s/invoke-with-response-stream", model)
	}
	return b.endpoint + path
}

// parseResponse parses the Bedrock API response
func (b *BedrockProvider) parseResponse(body []byte, model string) (provider.Response, error) {
	var bedrockResp bedrockResponse
	if err := json.Unmarshal(body, &bedrockResp); err != nil {
		return provider.Response{}, provider.NewProviderError("bedrock", model, fmt.Errorf("failed to parse response: %w", err))
	}

	return parseBedrockResponse(&bedrockResp, model), nil
}
