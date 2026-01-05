package meta

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/AINative-studio/ainative-code/internal/provider"
)

// MetaProvider implements the Provider interface for Meta LLAMA API
type MetaProvider struct {
	config *Config
	client *http.Client
}

// NewMetaProvider creates a new Meta LLAMA provider instance
func NewMetaProvider(config *Config) (*MetaProvider, error) {
	if config == nil {
		config = DefaultConfig()
	}

	config.SetDefaults()

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &MetaProvider{
		config: config,
		client: config.HTTPClient,
	}, nil
}

// Name returns the provider's name
func (p *MetaProvider) Name() string {
	return "meta"
}

// Models returns the list of supported model identifiers
func (p *MetaProvider) Models() []string {
	return []string{
		ModelLlama4Maverick,
		ModelLlama4Scout,
		ModelLlama33_70B,
		ModelLlama33_8B,
	}
}

// Chat sends a complete chat request and waits for the full response
func (p *MetaProvider) Chat(ctx context.Context, messages []provider.Message, opts ...provider.ChatOption) (provider.Response, error) {
	// Apply options
	options := &provider.ChatOptions{}
	for _, opt := range opts {
		opt(options)
	}

	// Build request
	req := p.buildRequest(messages, options)

	// Create HTTP request
	httpReq, err := p.createHTTPRequest(ctx, req)
	if err != nil {
		return provider.Response{}, fmt.Errorf("failed to create request: %w", err)
	}

	// Execute request
	resp, err := p.client.Do(httpReq)
	if err != nil {
		return provider.Response{}, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Handle errors
	if err := handleHTTPError(resp); err != nil {
		return provider.Response{}, err
	}

	// Parse response
	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return provider.Response{}, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert to provider.Response
	return p.convertResponse(chatResp), nil
}

// Stream sends a streaming chat request and returns a channel for events
func (p *MetaProvider) Stream(ctx context.Context, messages []provider.Message, opts ...provider.StreamOption) (<-chan provider.Event, error) {
	// Apply options
	options := &provider.ChatOptions{}
	provider.ApplyStreamOptions(options, opts...)

	req := p.buildRequest(messages, options)
	req.Stream = true

	// Create HTTP request
	httpReq, err := p.createHTTPRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Execute request
	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// Handle errors
	if err := handleHTTPError(resp); err != nil {
		resp.Body.Close()
		return nil, err
	}

	// Create event channel and start streaming
	eventChan := make(chan provider.Event, 10)
	go p.processStream(ctx, resp.Body, eventChan)

	return eventChan, nil
}

// Close releases any resources held by the provider
func (p *MetaProvider) Close() error {
	// No resources to release for Meta provider
	return nil
}

// buildRequest builds a Meta API request from messages and options
func (p *MetaProvider) buildRequest(messages []provider.Message, options *provider.ChatOptions) *ChatRequest {
	req := &ChatRequest{
		Model:            p.config.Model,
		Messages:         make([]Message, len(messages)),
		Temperature:      p.config.Temperature,
		TopP:             p.config.TopP,
		MaxTokens:        p.config.MaxTokens,
		PresencePenalty:  p.config.PresencePenalty,
		FrequencyPenalty: p.config.FrequencyPenalty,
		Stop:             p.config.Stop,
	}

	// Convert messages
	for i, msg := range messages {
		req.Messages[i] = Message{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// Apply options
	if options != nil {
		if options.Model != "" {
			req.Model = options.Model
		}
		if options.Temperature > 0 {
			req.Temperature = options.Temperature
		}
		if options.TopP > 0 {
			req.TopP = options.TopP
		}
		if options.MaxTokens > 0 {
			req.MaxTokens = options.MaxTokens
		}
		if len(options.StopSequences) > 0 {
			req.Stop = options.StopSequences
		}
	}

	return req
}

// createHTTPRequest creates an HTTP request for the Meta API
func (p *MetaProvider) createHTTPRequest(ctx context.Context, req *ChatRequest) (*http.Request, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/chat/completions", p.config.BaseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.config.APIKey))

	return httpReq, nil
}

// convertResponse converts a Meta API response to provider.Response
func (p *MetaProvider) convertResponse(resp ChatResponse) provider.Response {
	content := ""
	if len(resp.Choices) > 0 {
		content = resp.Choices[0].Message.Content
	}

	return provider.Response{
		Content: content,
		Usage: provider.Usage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
		Model: resp.Model,
	}
}

// processStream processes streaming responses and sends events to the channel
func (p *MetaProvider) processStream(ctx context.Context, body io.ReadCloser, eventChan chan<- provider.Event) {
	defer close(eventChan)
	defer body.Close()

	// Send start event
	eventChan <- provider.Event{
		Type: provider.EventTypeContentStart,
	}

	buffer := make([]byte, 4096)

	for {
		select {
		case <-ctx.Done():
			eventChan <- provider.Event{
				Type:  provider.EventTypeError,
				Error: ctx.Err(),
			}
			return
		default:
		}

		// Read next chunk
		n, err := body.Read(buffer)
		if err != nil && err != io.EOF {
			eventChan <- provider.Event{
				Type:  provider.EventTypeError,
				Error: fmt.Errorf("stream read error: %w", err),
			}
			return
		}

		if n == 0 {
			break
		}

		// Process SSE data
		data := buffer[:n]
		lines := bytes.Split(data, []byte("\n"))

		for _, line := range lines {
			line = bytes.TrimSpace(line)
			if len(line) == 0 || !bytes.HasPrefix(line, []byte("data: ")) {
				continue
			}

			// Extract JSON data
			jsonData := bytes.TrimPrefix(line, []byte("data: "))
			if bytes.Equal(jsonData, []byte("[DONE]")) {
				eventChan <- provider.Event{
					Type: provider.EventTypeContentEnd,
					Done: true,
				}
				return
			}

			// Parse stream chunk
			var chunk StreamResponse
			if err := json.Unmarshal(jsonData, &chunk); err != nil {
				continue // Skip malformed chunks
			}

			// Send content delta
			if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
				eventChan <- provider.Event{
					Type:    provider.EventTypeContentDelta,
					Content: chunk.Choices[0].Delta.Content,
				}
			}
		}

		if err == io.EOF {
			break
		}
	}

	// Send end event
	eventChan <- provider.Event{
		Type: provider.EventTypeContentEnd,
		Done: true,
	}
}
