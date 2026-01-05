package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/AINative-studio/ainative-code/internal/logger"
	"github.com/AINative-studio/ainative-code/internal/provider"
)

const (
	// GeminiAPIURL is the base URL for the Gemini API
	GeminiAPIURL = "https://generativelanguage.googleapis.com/v1beta"

	// Default API version
	defaultAPIVersion = "v1beta"
)

// Supported Gemini models
var supportedModels = []string{
	"gemini-pro",
	"gemini-pro-vision",
	"gemini-ultra",
	"gemini-1.5-pro",
	"gemini-1.5-pro-latest",
	"gemini-1.5-flash",
	"gemini-1.5-flash-latest",
}

// GeminiProvider implements the Provider interface for Google's Gemini API
type GeminiProvider struct {
	*provider.BaseProvider
	apiKey  string
	baseURL string
}

// Config contains configuration for the Gemini provider
type Config struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
	Logger     logger.LoggerInterface
}

// NewGeminiProvider creates a new Gemini provider instance
func NewGeminiProvider(config Config) (*GeminiProvider, error) {
	if config.APIKey == "" {
		return nil, provider.NewAuthenticationError("gemini", "API key is required")
	}

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = GeminiAPIURL
	}

	baseProvider := provider.NewBaseProvider(provider.BaseProviderConfig{
		Name:        "gemini",
		HTTPClient:  config.HTTPClient,
		Logger:      config.Logger,
		RetryConfig: provider.DefaultRetryConfig(),
	})

	return &GeminiProvider{
		BaseProvider: baseProvider,
		apiKey:       config.APIKey,
		baseURL:      baseURL,
	}, nil
}

// Name returns the provider name
func (g *GeminiProvider) Name() string {
	return g.BaseProvider.Name()
}

// Models returns the list of supported models
func (g *GeminiProvider) Models() []string {
	models := make([]string, len(supportedModels))
	copy(models, supportedModels)
	return models
}

// Chat sends a chat request to the Gemini API
func (g *GeminiProvider) Chat(ctx context.Context, messages []provider.Message, opts ...provider.ChatOption) (provider.Response, error) {
	// Apply options
	options := provider.DefaultChatOptions()
	provider.ApplyChatOptions(options, opts...)

	// Validate model
	if err := g.ValidateModel(options.Model, supportedModels); err != nil {
		return provider.Response{}, err
	}

	// Build request
	req, err := g.buildRequest(ctx, messages, options, false)
	if err != nil {
		return provider.Response{}, provider.NewProviderError("gemini", options.Model, err)
	}

	// Execute request
	resp, err := g.DoRequest(ctx, req)
	if err != nil {
		return provider.Response{}, err
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return provider.Response{}, provider.NewProviderError("gemini", options.Model, fmt.Errorf("failed to read response: %w", err))
	}

	// Handle errors
	if resp.StatusCode != http.StatusOK {
		return provider.Response{}, g.handleAPIError(resp, body, options.Model)
	}

	// Parse response
	return g.parseResponse(body, options.Model)
}

// Stream sends a streaming chat request to the Gemini API
func (g *GeminiProvider) Stream(ctx context.Context, messages []provider.Message, opts ...provider.StreamOption) (<-chan provider.Event, error) {
	// Apply options
	options := provider.DefaultChatOptions()
	provider.ApplyStreamOptions(options, opts...)

	// Validate model
	if err := g.ValidateModel(options.Model, supportedModels); err != nil {
		return nil, err
	}

	// Build request
	req, err := g.buildRequest(ctx, messages, options, true)
	if err != nil {
		return nil, provider.NewProviderError("gemini", options.Model, err)
	}

	// Execute request
	resp, err := g.DoRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	// Handle errors
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, g.handleAPIError(resp, body, options.Model)
	}

	// Create event channel
	eventChan := make(chan provider.Event)

	// Start streaming goroutine
	go g.streamResponse(ctx, resp.Body, eventChan, options.Model)

	return eventChan, nil
}

// Close releases resources held by the provider
func (g *GeminiProvider) Close() error {
	return g.BaseProvider.Close()
}

// buildRequest constructs an HTTP request for the Gemini API
func (g *GeminiProvider) buildRequest(ctx context.Context, messages []provider.Message, options *provider.ChatOptions, stream bool) (*http.Request, error) {
	// Convert messages to Gemini format
	contents, systemInstruction := g.convertMessages(messages, options.SystemPrompt)

	// Build request body
	reqBody := geminiRequest{
		Contents: contents,
	}

	// Add system instruction if present
	if systemInstruction != nil {
		reqBody.SystemInstruction = systemInstruction
	}

	// Build generation config
	genConfig := &generationConfig{}
	if options.MaxTokens > 0 {
		genConfig.MaxOutputTokens = options.MaxTokens
	}
	if options.Temperature > 0 {
		genConfig.Temperature = &options.Temperature
	}
	if options.TopP > 0 && options.TopP < 1.0 {
		genConfig.TopP = &options.TopP
	}
	if len(options.StopSequences) > 0 {
		genConfig.StopSequences = options.StopSequences
	}

	// TopK is Gemini-specific, can be set via metadata
	if topKStr, ok := options.Metadata["topK"]; ok {
		if topK, err := strconv.Atoi(topKStr); err == nil && topK > 0 {
			genConfig.TopK = &topK
		}
	}

	reqBody.GenerationConfig = genConfig

	// Note: Safety settings would need to be passed through a different mechanism
	// or added to options as a typed field for proper handling

	// Marshal request body
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Build URL
	method := "generateContent"
	if stream {
		method = "streamGenerateContent?alt=sse"
	}
	url := fmt.Sprintf("%s/models/%s:%s", g.baseURL, options.Model, method)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Add API key as query parameter (Gemini's authentication method)
	q := req.URL.Query()
	q.Add("key", g.apiKey)
	req.URL.RawQuery = q.Encode()

	return req, nil
}

// convertMessages converts provider messages to Gemini API format
func (g *GeminiProvider) convertMessages(messages []provider.Message, systemPrompt string) ([]geminiContent, *geminiContent) {
	var contents []geminiContent
	var systemInstruction *geminiContent

	// Handle system prompt
	if systemPrompt != "" {
		systemInstruction = &geminiContent{
			Parts: []geminiPart{
				{Text: systemPrompt},
			},
		}
	}

	// Convert messages
	for _, msg := range messages {
		// Extract system messages for system instruction
		if msg.Role == "system" {
			if systemInstruction == nil {
				systemInstruction = &geminiContent{
					Parts: []geminiPart{{Text: msg.Content}},
				}
			} else {
				// Append to existing system instruction
				systemInstruction.Parts = append(systemInstruction.Parts, geminiPart{Text: "\n\n" + msg.Content})
			}
			continue
		}

		// Convert role (Gemini uses "model" instead of "assistant")
		role := msg.Role
		if role == "assistant" {
			role = "model"
		}

		contents = append(contents, geminiContent{
			Role: role,
			Parts: []geminiPart{
				{Text: msg.Content},
			},
		})
	}

	return contents, systemInstruction
}

// parseResponse parses the Gemini API response
func (g *GeminiProvider) parseResponse(body []byte, model string) (provider.Response, error) {
	var apiResp geminiResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return provider.Response{}, provider.NewProviderError("gemini", model, fmt.Errorf("failed to parse response: %w", err))
	}

	// Check for prompt feedback (safety blocks)
	if apiResp.PromptFeedback != nil && apiResp.PromptFeedback.BlockReason != "" {
		return provider.Response{}, provider.NewProviderError("gemini", model,
			fmt.Errorf("prompt blocked: %s", apiResp.PromptFeedback.BlockReason))
	}

	// Check if we have candidates
	if len(apiResp.Candidates) == 0 {
		return provider.Response{}, provider.NewProviderError("gemini", model, fmt.Errorf("no candidates in response"))
	}

	// Extract content from first candidate
	candidate := apiResp.Candidates[0]

	// Check if response was blocked
	if candidate.FinishReason == "SAFETY" {
		return provider.Response{}, provider.NewProviderError("gemini", model,
			fmt.Errorf("response blocked due to safety settings"))
	}

	// Extract text content
	var content string
	for _, part := range candidate.Content.Parts {
		if part.Text != "" {
			if content != "" {
				content += "\n"
			}
			content += part.Text
		}
	}

	// Build usage information
	usage := provider.Usage{}
	if apiResp.UsageMetadata != nil {
		usage.PromptTokens = apiResp.UsageMetadata.PromptTokenCount
		usage.CompletionTokens = apiResp.UsageMetadata.CandidatesTokenCount
		usage.TotalTokens = apiResp.UsageMetadata.TotalTokenCount
	}

	return provider.Response{
		Content: content,
		Model:   model,
		Usage:   usage,
	}, nil
}
