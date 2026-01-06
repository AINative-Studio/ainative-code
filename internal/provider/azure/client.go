package azure

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/AINative-studio/ainative-code/internal/provider"
)

// AzureProvider implements the Provider interface for Azure OpenAI
type AzureProvider struct {
	*provider.BaseProvider
	config Config
}

// NewAzureProvider creates a new Azure OpenAI provider instance
func NewAzureProvider(config Config) (*AzureProvider, error) {
	// Validate and set defaults
	if err := config.Validate(); err != nil {
		return nil, err
	}
	config.SetDefaults()

	baseProvider := provider.NewBaseProvider(provider.BaseProviderConfig{
		Name:       "azure",
		HTTPClient: config.HTTPClient,
		Logger:     config.Logger,
		RetryConfig: provider.RetryConfig{
			MaxRetries:           config.MaxRetries,
			InitialBackoff:       config.RetryDelay,
			MaxBackoff:           config.RetryDelay * 10,
			Multiplier:           1.5,
			RetryableStatusCodes: []int{429, 500, 502, 503, 504},
		},
	})

	return &AzureProvider{
		BaseProvider: baseProvider,
		config:       config,
	}, nil
}

// Name returns the provider name
func (a *AzureProvider) Name() string {
	return "azure"
}

// Models returns the list of available models (based on deployment)
func (a *AzureProvider) Models() []string {
	// For Azure, the deployment name is what matters
	// Return the configured deployment as the "model"
	return []string{a.config.Deployment}
}

// Chat sends a chat request to Azure OpenAI
func (a *AzureProvider) Chat(ctx context.Context, messages []provider.Message, opts ...provider.ChatOption) (provider.Response, error) {
	// Apply options
	options := provider.DefaultChatOptions()
	provider.ApplyChatOptions(options, opts...)

	// Build request
	req, err := a.buildRequest(ctx, messages, options, false)
	if err != nil {
		return provider.Response{}, err
	}

	// Execute request with retry
	resp, err := a.DoRequest(ctx, req)
	if err != nil {
		return provider.Response{}, err
	}
	defer resp.Body.Close()

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return provider.Response{}, a.handleError(resp, body)
	}

	// Parse response
	var azureResp azureResponse
	if err := json.NewDecoder(resp.Body).Decode(&azureResp); err != nil {
		return provider.Response{}, provider.NewProviderError("azure", a.config.Deployment, fmt.Errorf("failed to decode response: %w", err))
	}

	// Convert to provider response
	return a.convertResponse(azureResp), nil
}

// Stream sends a streaming chat request to Azure OpenAI
func (a *AzureProvider) Stream(ctx context.Context, messages []provider.Message, opts ...provider.StreamOption) (<-chan provider.Event, error) {
	// Apply stream options
	options := provider.DefaultChatOptions()
	provider.ApplyStreamOptions(options, opts...)

	// Build streaming request
	req, err := a.buildRequest(ctx, messages, options, true)
	if err != nil {
		return nil, err
	}

	// Execute request
	resp, err := a.DoRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, a.handleError(resp, body)
	}

	// Create event channel
	eventChan := make(chan provider.Event)

	// Start streaming goroutine
	go a.streamResponse(ctx, resp.Body, eventChan)

	return eventChan, nil
}

// Close releases resources
func (a *AzureProvider) Close() error {
	return a.BaseProvider.Close()
}

// buildRequest constructs an HTTP request for Azure OpenAI
func (a *AzureProvider) buildRequest(ctx context.Context, messages []provider.Message, options *provider.ChatOptions, stream bool) (*http.Request, error) {
	// Convert messages to Azure format
	azureMessages := make([]azureMessage, len(messages))
	for i, msg := range messages {
		azureMessages[i] = azureMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// Build request body
	reqBody := azureRequest{
		Messages:  azureMessages,
		MaxTokens: options.MaxTokens,
		Stream:    stream,
	}

	// Set optional parameters
	if options.Temperature != 0 {
		reqBody.Temperature = &options.Temperature
	}
	if options.TopP != 0 {
		reqBody.TopP = &options.TopP
	}

	if len(options.StopSequences) > 0 {
		reqBody.Stop = options.StopSequences
	}

	// Marshal request
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, provider.NewProviderError("azure", a.config.Deployment, fmt.Errorf("failed to marshal request: %w", err))
	}

	// Build URL: https://{endpoint}/openai/deployments/{deployment}/chat/completions?api-version={version}
	url := fmt.Sprintf("%s/openai/deployments/%s/chat/completions?api-version=%s",
		strings.TrimSuffix(a.config.Endpoint, "/"),
		a.config.Deployment,
		a.config.APIVersion,
	)

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, provider.NewProviderError("azure", a.config.Deployment, fmt.Errorf("failed to create request: %w", err))
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", a.config.APIKey)

	return req, nil
}

// convertResponse converts Azure response to provider response
func (a *AzureProvider) convertResponse(azureResp azureResponse) provider.Response {
	if len(azureResp.Choices) == 0 {
		return provider.Response{}
	}

	choice := azureResp.Choices[0]
	return provider.Response{
		Content: choice.Message.Content.(string),
		Usage: provider.Usage{
			PromptTokens:     azureResp.Usage.PromptTokens,
			CompletionTokens: azureResp.Usage.CompletionTokens,
			TotalTokens:      azureResp.Usage.TotalTokens,
		},
		Model: azureResp.Model,
	}
}

// streamResponse processes the streaming response
func (a *AzureProvider) streamResponse(ctx context.Context, body io.ReadCloser, eventChan chan<- provider.Event) {
	defer close(eventChan)
	defer body.Close()

	reader := &sseReader{reader: body}

	for {
		select {
		case <-ctx.Done():
			eventChan <- provider.Event{
				Error: ctx.Err(),
			}
			return
		default:
		}

		// Read next SSE event
		event, err := reader.ReadEvent()
		if err != nil {
			if err != io.EOF {
				eventChan <- provider.Event{
					Error: provider.NewProviderError("azure", a.config.Deployment, err),
				}
			}
			return
		}

		// Skip non-data events
		if !strings.HasPrefix(event, "data: ") {
			continue
		}

		// Extract data
		data := strings.TrimPrefix(event, "data: ")
		data = strings.TrimSpace(data)

		// Check for done signal
		if data == "[DONE]" {
			return
		}

		// Parse chunk
		var chunk azureStreamResponse
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			eventChan <- provider.Event{
				Error: provider.NewProviderError("azure", a.config.Deployment, fmt.Errorf("failed to parse chunk: %w", err)),
			}
			return
		}

		// Send content
		if len(chunk.Choices) > 0 {
			delta := chunk.Choices[0].Delta
			if delta.Content != "" {
				eventChan <- provider.Event{
					Content: delta.Content,
				}
			}

			// Check for finish
			if chunk.Choices[0].FinishReason != nil {
				return
			}
		}
	}
}

// handleError processes Azure OpenAI error responses
func (a *AzureProvider) handleError(resp *http.Response, body []byte) error {
	var azureErr azureError
	if err := json.Unmarshal(body, &azureErr); err != nil {
		return provider.NewProviderError("azure", a.config.Deployment, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body)))
	}

	message := azureErr.Error.Message
	if message == "" {
		message = fmt.Sprintf("Azure OpenAI error: %s", azureErr.Error.Type)
	}

	switch resp.StatusCode {
	case 401:
		return provider.NewAuthenticationError("azure", message)
	case 429:
		return provider.NewRateLimitError("azure", 0) // 0 = unknown retry-after
	case 400:
		return provider.NewProviderError("azure", a.config.Deployment, fmt.Errorf("validation error: %s", message))
	default:
		return provider.NewProviderError("azure", a.config.Deployment, fmt.Errorf("HTTP %d: %s", resp.StatusCode, message))
	}
}

// sseReader reads Server-Sent Events from a stream
type sseReader struct {
	reader io.Reader
	buffer []byte
}

// ReadEvent reads the next SSE event
func (r *sseReader) ReadEvent() (string, error) {
	buf := make([]byte, 4096)
	for {
		n, err := r.reader.Read(buf)
		if n > 0 {
			r.buffer = append(r.buffer, buf[:n]...)

			// Look for double newline (end of event)
			if idx := bytes.Index(r.buffer, []byte("\n\n")); idx >= 0 {
				event := string(r.buffer[:idx])
				r.buffer = r.buffer[idx+2:]
				return event, nil
			}
		}
		if err != nil {
			return "", err
		}
	}
}
