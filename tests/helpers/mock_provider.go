package helpers

import (
	"context"
	"errors"
	"io"
	"sync"
	"time"

	"github.com/AINative-studio/ainative-code/internal/providers"
)

// MockProvider is a mock implementation of the Provider interface for testing
type MockProvider struct {
	mu              sync.RWMutex
	name            string
	chatResponse    *providers.Response
	chatError       error
	streamEvents    []providers.Event
	streamError     error
	models          []providers.Model
	modelsError     error
	callCount       int
	lastChatRequest *providers.ChatRequest
	closed          bool
}

// NewMockProvider creates a new mock provider
func NewMockProvider(name string) *MockProvider {
	return &MockProvider{
		name: name,
		models: []providers.Model{
			{
				ID:       "test-model-v1",
				Name:     "Test Model v1",
				Provider: name,
				MaxTokens: 4096,
				Capabilities: []string{"chat", "streaming"},
			},
		},
	}
}

// SetChatResponse configures the response for Chat calls
func (m *MockProvider) SetChatResponse(resp *providers.Response, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.chatResponse = resp
	m.chatError = err
}

// SetStreamEvents configures the events for Stream calls
func (m *MockProvider) SetStreamEvents(events []providers.Event, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.streamEvents = events
	m.streamError = err
}

// SetModels configures the models returned by Models calls
func (m *MockProvider) SetModels(models []providers.Model, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.models = models
	m.modelsError = err
}

// GetCallCount returns the number of times Chat or Stream was called
func (m *MockProvider) GetCallCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.callCount
}

// GetLastChatRequest returns the last ChatRequest received
func (m *MockProvider) GetLastChatRequest() *providers.ChatRequest {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastChatRequest
}

// Chat implements the Provider interface
func (m *MockProvider) Chat(ctx context.Context, req *providers.ChatRequest, opts ...providers.Option) (*providers.Response, error) {
	m.mu.Lock()
	m.callCount++
	m.lastChatRequest = req
	m.mu.Unlock()

	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.chatError != nil {
		return nil, m.chatError
	}

	if m.chatResponse != nil {
		return m.chatResponse, nil
	}

	// Default response
	return &providers.Response{
		Content:      "This is a mock response",
		Model:        req.Model,
		Provider:     m.name,
		FinishReason: "stop",
		Usage: &providers.UsageInfo{
			PromptTokens:     10,
			CompletionTokens: 20,
			TotalTokens:      30,
		},
		CreatedAt: time.Now(),
	}, nil
}

// Stream implements the Provider interface
func (m *MockProvider) Stream(ctx context.Context, req *providers.StreamRequest, opts ...providers.Option) (<-chan providers.Event, error) {
	m.mu.Lock()
	m.callCount++
	m.mu.Unlock()

	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.streamError != nil {
		return nil, m.streamError
	}

	eventChan := make(chan providers.Event, len(m.streamEvents)+1)

	// Send events in a goroutine
	go func() {
		defer close(eventChan)

		// Send configured events or default events
		events := m.streamEvents
		if len(events) == 0 {
			// Default streaming events
			events = []providers.Event{
				{
					Type:      providers.EventMessageStart,
					Timestamp: time.Now(),
				},
				{
					Type:      providers.EventContentStart,
					Timestamp: time.Now(),
				},
				{
					Type:      providers.EventTextDelta,
					Data:      "This ",
					Timestamp: time.Now(),
				},
				{
					Type:      providers.EventTextDelta,
					Data:      "is ",
					Timestamp: time.Now(),
				},
				{
					Type:      providers.EventTextDelta,
					Data:      "a ",
					Timestamp: time.Now(),
				},
				{
					Type:      providers.EventTextDelta,
					Data:      "streaming ",
					Timestamp: time.Now(),
				},
				{
					Type:      providers.EventTextDelta,
					Data:      "response",
					Timestamp: time.Now(),
				},
				{
					Type:      providers.EventContentEnd,
					Timestamp: time.Now(),
				},
				{
					Type: providers.EventUsage,
					Usage: &providers.UsageInfo{
						PromptTokens:     10,
						CompletionTokens: 15,
						TotalTokens:      25,
					},
					Timestamp: time.Now(),
				},
				{
					Type:      providers.EventMessageStop,
					Timestamp: time.Now(),
				},
			}
		}

		for _, event := range events {
			select {
			case <-ctx.Done():
				return
			case eventChan <- event:
				// Small delay to simulate streaming
				time.Sleep(10 * time.Millisecond)
			}
		}
	}()

	return eventChan, nil
}

// Name implements the Provider interface
func (m *MockProvider) Name() string {
	return m.name
}

// Models implements the Provider interface
func (m *MockProvider) Models(ctx context.Context) ([]providers.Model, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.modelsError != nil {
		return nil, m.modelsError
	}

	return m.models, nil
}

// Close implements the Provider interface
func (m *MockProvider) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return errors.New("provider already closed")
	}

	m.closed = true
	return nil
}

// MockProviderWithErrors creates a mock provider that simulates various error conditions
type MockProviderWithErrors struct {
	*MockProvider
	authError      error
	rateLimitError error
	timeoutError   error
}

// NewMockProviderWithErrors creates a mock provider that can simulate errors
func NewMockProviderWithErrors(name string) *MockProviderWithErrors {
	return &MockProviderWithErrors{
		MockProvider: NewMockProvider(name),
	}
}

// SimulateAuthError configures the provider to return an authentication error
func (m *MockProviderWithErrors) SimulateAuthError() {
	m.authError = errors.New("authentication failed: invalid API key")
	m.SetChatResponse(nil, m.authError)
	m.SetStreamEvents(nil, m.authError)
}

// SimulateRateLimitError configures the provider to return a rate limit error
func (m *MockProviderWithErrors) SimulateRateLimitError() {
	m.rateLimitError = errors.New("rate limit exceeded: too many requests")
	m.SetChatResponse(nil, m.rateLimitError)
	m.SetStreamEvents(nil, m.rateLimitError)
}

// SimulateTimeoutError configures the provider to return a timeout error
func (m *MockProviderWithErrors) SimulateTimeoutError() {
	m.timeoutError = context.DeadlineExceeded
	m.SetChatResponse(nil, m.timeoutError)
	m.SetStreamEvents(nil, m.timeoutError)
}

// Reset clears all error simulations
func (m *MockProviderWithErrors) Reset() {
	m.authError = nil
	m.rateLimitError = nil
	m.timeoutError = nil
	m.SetChatResponse(nil, nil)
	m.SetStreamEvents(nil, nil)
}

var _ providers.Provider = (*MockProvider)(nil)
var _ io.Closer = (*MockProvider)(nil)
