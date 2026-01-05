package gemini

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AINative-studio/ainative-code/internal/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewGeminiProvider tests the provider creation
func TestNewGeminiProvider(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			config: Config{
				APIKey: "test-api-key",
			},
			expectError: false,
		},
		{
			name: "valid config with custom base URL",
			config: Config{
				APIKey:  "test-api-key",
				BaseURL: "https://custom.api.com",
			},
			expectError: false,
		},
		{
			name:        "missing API key",
			config:      Config{},
			expectError: true,
			errorMsg:    "API key is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewGeminiProvider(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, p)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, p)
				assert.Equal(t, "gemini", p.Name())
			}
		})
	}
}

// TestModels tests that the provider returns supported models
func TestModels(t *testing.T) {
	p, err := NewGeminiProvider(Config{APIKey: "test-key"})
	require.NoError(t, err)

	models := p.Models()
	assert.NotEmpty(t, models)
	assert.Contains(t, models, "gemini-pro")
	assert.Contains(t, models, "gemini-pro-vision")
	assert.Contains(t, models, "gemini-ultra")
	assert.Contains(t, models, "gemini-1.5-pro")
}

// TestConvertMessages tests message conversion to Gemini format
func TestConvertMessages(t *testing.T) {
	p, err := NewGeminiProvider(Config{APIKey: "test-key"})
	require.NoError(t, err)

	tests := []struct {
		name               string
		messages           []provider.Message
		systemPrompt       string
		expectedContents   int
		expectedSystem     bool
		expectedFirstRole  string
	}{
		{
			name: "user message only",
			messages: []provider.Message{
				{Role: "user", Content: "Hello"},
			},
			expectedContents:  1,
			expectedSystem:    false,
			expectedFirstRole: "user",
		},
		{
			name: "user and assistant messages",
			messages: []provider.Message{
				{Role: "user", Content: "Hello"},
				{Role: "assistant", Content: "Hi there"},
			},
			expectedContents:  2,
			expectedSystem:    false,
			expectedFirstRole: "user",
		},
		{
			name: "with system message",
			messages: []provider.Message{
				{Role: "system", Content: "You are helpful"},
				{Role: "user", Content: "Hello"},
			},
			expectedContents:  1,
			expectedSystem:    true,
			expectedFirstRole: "user",
		},
		{
			name: "with system prompt",
			messages: []provider.Message{
				{Role: "user", Content: "Hello"},
			},
			systemPrompt:      "You are helpful",
			expectedContents:  1,
			expectedSystem:    true,
			expectedFirstRole: "user",
		},
		{
			name: "multiple system messages",
			messages: []provider.Message{
				{Role: "system", Content: "Part 1"},
				{Role: "system", Content: "Part 2"},
				{Role: "user", Content: "Hello"},
			},
			expectedContents:  1,
			expectedSystem:    true,
			expectedFirstRole: "user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contents, systemInstruction := p.convertMessages(tt.messages, tt.systemPrompt)

			assert.Equal(t, tt.expectedContents, len(contents))

			if tt.expectedSystem {
				assert.NotNil(t, systemInstruction)
				assert.NotEmpty(t, systemInstruction.Parts)
			} else {
				if systemInstruction != nil {
					// System instruction can be nil or empty
					assert.Empty(t, systemInstruction.Parts)
				}
			}

			if len(contents) > 0 {
				assert.Equal(t, tt.expectedFirstRole, contents[0].Role)
			}
		})
	}
}

// TestConvertMessagesRoleMapping tests that assistant role is converted to model
func TestConvertMessagesRoleMapping(t *testing.T) {
	p, err := NewGeminiProvider(Config{APIKey: "test-key"})
	require.NoError(t, err)

	messages := []provider.Message{
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi"},
		{Role: "user", Content: "How are you?"},
	}

	contents, _ := p.convertMessages(messages, "")

	assert.Equal(t, 3, len(contents))
	assert.Equal(t, "user", contents[0].Role)
	assert.Equal(t, "model", contents[1].Role) // assistant -> model
	assert.Equal(t, "user", contents[2].Role)
}

// TestChat tests the Chat method with a mock server
func TestChat(t *testing.T) {
	// Create mock server
	mockResponse := geminiResponse{
		Candidates: []candidate{
			{
				Content: geminiContent{
					Parts: []geminiPart{
						{Text: "This is a test response"},
					},
				},
				FinishReason: "STOP",
			},
		},
		UsageMetadata: &usageMetadata{
			PromptTokenCount:     10,
			CandidatesTokenCount: 20,
			TotalTokenCount:      30,
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		assert.Equal(t, "POST", r.Method)
		assert.Contains(t, r.URL.Path, "generateContent")
		assert.Contains(t, r.URL.RawQuery, "key=")

		// Read and verify request body
		body, _ := io.ReadAll(r.Body)
		var req geminiRequest
		json.Unmarshal(body, &req)
		assert.NotEmpty(t, req.Contents)

		// Send response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	// Create provider with mock server
	p, err := NewGeminiProvider(Config{
		APIKey:  "test-key",
		BaseURL: server.URL,
	})
	require.NoError(t, err)

	// Test Chat
	ctx := context.Background()
	messages := []provider.Message{
		{Role: "user", Content: "Hello"},
	}

	resp, err := p.Chat(ctx, messages, provider.WithModel("gemini-pro"))
	require.NoError(t, err)

	assert.Equal(t, "This is a test response", resp.Content)
	assert.Equal(t, "gemini-pro", resp.Model)
	assert.Equal(t, 10, resp.Usage.PromptTokens)
	assert.Equal(t, 20, resp.Usage.CompletionTokens)
	assert.Equal(t, 30, resp.Usage.TotalTokens)
}

// TestChatWithOptions tests Chat with various options
func TestChatWithOptions(t *testing.T) {
	mockResponse := geminiResponse{
		Candidates: []candidate{
			{
				Content: geminiContent{
					Parts: []geminiPart{{Text: "Response"}},
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req geminiRequest
		json.Unmarshal(body, &req)

		// Verify options were applied
		assert.NotNil(t, req.GenerationConfig)
		assert.Equal(t, 100, req.GenerationConfig.MaxOutputTokens)
		assert.NotNil(t, req.GenerationConfig.Temperature)
		assert.Equal(t, 0.7, *req.GenerationConfig.Temperature)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	p, err := NewGeminiProvider(Config{
		APIKey:  "test-key",
		BaseURL: server.URL,
	})
	require.NoError(t, err)

	ctx := context.Background()
	messages := []provider.Message{
		{Role: "user", Content: "Hello"},
	}

	_, err = p.Chat(ctx, messages,
		provider.WithModel("gemini-pro"),
		provider.WithMaxTokens(100),
		provider.WithTemperature(0.7),
	)
	require.NoError(t, err)
}

// TestChatInvalidModel tests Chat with an invalid model
func TestChatInvalidModel(t *testing.T) {
	p, err := NewGeminiProvider(Config{APIKey: "test-key"})
	require.NoError(t, err)

	ctx := context.Background()
	messages := []provider.Message{
		{Role: "user", Content: "Hello"},
	}

	_, err = p.Chat(ctx, messages, provider.WithModel("invalid-model"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid model")
}

// TestChatSafetyBlock tests handling of safety blocks
func TestChatSafetyBlock(t *testing.T) {
	mockResponse := geminiResponse{
		PromptFeedback: &promptFeedback{
			BlockReason: "SAFETY",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	p, err := NewGeminiProvider(Config{
		APIKey:  "test-key",
		BaseURL: server.URL,
	})
	require.NoError(t, err)

	ctx := context.Background()
	messages := []provider.Message{
		{Role: "user", Content: "Hello"},
	}

	_, err = p.Chat(ctx, messages, provider.WithModel("gemini-pro"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "blocked")
}

// TestStream tests the Stream method
func TestStream(t *testing.T) {
	// Mock SSE stream
	sseData := `data: {"candidates":[{"content":{"parts":[{"text":"Hello"}]}}]}

data: {"candidates":[{"content":{"parts":[{"text":" world!"}]},"finishReason":"STOP"}]}

`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "streamGenerateContent")

		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(sseData))
	}))
	defer server.Close()

	p, err := NewGeminiProvider(Config{
		APIKey:  "test-key",
		BaseURL: server.URL,
	})
	require.NoError(t, err)

	ctx := context.Background()
	messages := []provider.Message{
		{Role: "user", Content: "Hello"},
	}

	eventChan, err := p.Stream(ctx, messages, provider.StreamWithModel("gemini-pro"))
	require.NoError(t, err)
	require.NotNil(t, eventChan)

	// Collect events
	var events []provider.Event
	for event := range eventChan {
		events = append(events, event)
	}

	// Verify events
	assert.NotEmpty(t, events)

	// Find content start event
	hasStart := false
	for _, e := range events {
		if e.Type == provider.EventTypeContentStart {
			hasStart = true
			break
		}
	}
	assert.True(t, hasStart, "should have content start event")

	// Find content delta events
	var deltaContent string
	for _, e := range events {
		if e.Type == provider.EventTypeContentDelta {
			deltaContent += e.Content
		}
	}
	assert.Equal(t, "Hello world!", deltaContent)

	// Find content end event
	hasEnd := false
	for _, e := range events {
		if e.Type == provider.EventTypeContentEnd {
			hasEnd = true
			assert.True(t, e.Done)
			break
		}
	}
	assert.True(t, hasEnd, "should have content end event")
}

// TestStreamWithContext tests stream cancellation
func TestStreamWithContext(t *testing.T) {
	// Create a server that streams slowly
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		// Write one chunk
		w.Write([]byte(`data: {"candidates":[{"content":{"parts":[{"text":"Hello"}]}}]}` + "\n\n"))
		w.(http.Flusher).Flush()

		// Sleep to allow cancellation
		time.Sleep(100 * time.Millisecond)

		// Try to write another chunk (will fail if context cancelled)
		w.Write([]byte(`data: {"candidates":[{"content":{"parts":[{"text":" world"}]}}]}` + "\n\n"))
	}))
	defer server.Close()

	p, err := NewGeminiProvider(Config{
		APIKey:  "test-key",
		BaseURL: server.URL,
	})
	require.NoError(t, err)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	messages := []provider.Message{
		{Role: "user", Content: "Hello"},
	}

	eventChan, err := p.Stream(ctx, messages, provider.StreamWithModel("gemini-pro"))
	require.NoError(t, err)

	// Collect events until channel closes
	var hasError bool
	for event := range eventChan {
		if event.Type == provider.EventTypeError {
			hasError = true
			assert.ErrorIs(t, event.Error, context.DeadlineExceeded)
		}
	}

	assert.True(t, hasError, "should have error event for cancelled context")
}

// TestStreamSafetyBlock tests handling of safety blocks in streaming
func TestStreamSafetyBlock(t *testing.T) {
	sseData := `data: {"promptFeedback":{"blockReason":"SAFETY"}}

`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(sseData))
	}))
	defer server.Close()

	p, err := NewGeminiProvider(Config{
		APIKey:  "test-key",
		BaseURL: server.URL,
	})
	require.NoError(t, err)

	ctx := context.Background()
	messages := []provider.Message{
		{Role: "user", Content: "Hello"},
	}

	eventChan, err := p.Stream(ctx, messages, provider.StreamWithModel("gemini-pro"))
	require.NoError(t, err)

	// Look for error event
	var hasError bool
	for event := range eventChan {
		if event.Type == provider.EventTypeError {
			hasError = true
			assert.Contains(t, event.Error.Error(), "blocked")
		}
	}

	assert.True(t, hasError, "should have error for safety block")
}

// TestClose tests the Close method
func TestClose(t *testing.T) {
	p, err := NewGeminiProvider(Config{APIKey: "test-key"})
	require.NoError(t, err)

	err = p.Close()
	assert.NoError(t, err)
}

// TestParseResponseMultiPart tests parsing responses with multiple parts
func TestParseResponseMultiPart(t *testing.T) {
	p, err := NewGeminiProvider(Config{APIKey: "test-key"})
	require.NoError(t, err)

	mockResponse := geminiResponse{
		Candidates: []candidate{
			{
				Content: geminiContent{
					Parts: []geminiPart{
						{Text: "Part 1"},
						{Text: "Part 2"},
						{Text: "Part 3"},
					},
				},
			},
		},
	}

	body, _ := json.Marshal(mockResponse)
	resp, err := p.parseResponse(body, "gemini-pro")
	require.NoError(t, err)

	// Parts should be joined with newlines
	assert.Equal(t, "Part 1\nPart 2\nPart 3", resp.Content)
}

// TestParseResponseNoCandidates tests handling of responses with no candidates
func TestParseResponseNoCandidates(t *testing.T) {
	p, err := NewGeminiProvider(Config{APIKey: "test-key"})
	require.NoError(t, err)

	mockResponse := geminiResponse{
		Candidates: []candidate{},
	}

	body, _ := json.Marshal(mockResponse)
	_, err = p.parseResponse(body, "gemini-pro")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no candidates")
}

// TestBuildRequestURL tests URL construction
func TestBuildRequestURL(t *testing.T) {
	p, err := NewGeminiProvider(Config{
		APIKey:  "test-key",
		BaseURL: "https://test.api.com",
	})
	require.NoError(t, err)

	ctx := context.Background()
	messages := []provider.Message{
		{Role: "user", Content: "Hello"},
	}

	options := provider.DefaultChatOptions()
	options.Model = "gemini-pro"

	// Test non-streaming URL
	req, err := p.buildRequest(ctx, messages, options, false)
	require.NoError(t, err)
	assert.Contains(t, req.URL.String(), "generateContent")
	assert.NotContains(t, req.URL.String(), "streamGenerateContent")
	assert.Contains(t, req.URL.RawQuery, "key=test-key")

	// Test streaming URL
	req, err = p.buildRequest(ctx, messages, options, true)
	require.NoError(t, err)
	assert.Contains(t, req.URL.String(), "streamGenerateContent")
	assert.Contains(t, req.URL.RawQuery, "alt=sse")
}

// TestSystemPromptHandling tests various system prompt scenarios
func TestSystemPromptHandling(t *testing.T) {
	p, err := NewGeminiProvider(Config{APIKey: "test-key"})
	require.NoError(t, err)

	tests := []struct {
		name             string
		messages         []provider.Message
		systemPrompt     string
		expectSystemInst bool
	}{
		{
			name:             "no system",
			messages:         []provider.Message{{Role: "user", Content: "Hi"}},
			systemPrompt:     "",
			expectSystemInst: false,
		},
		{
			name:             "system in messages",
			messages:         []provider.Message{{Role: "system", Content: "Be helpful"}, {Role: "user", Content: "Hi"}},
			systemPrompt:     "",
			expectSystemInst: true,
		},
		{
			name:             "system prompt only",
			messages:         []provider.Message{{Role: "user", Content: "Hi"}},
			systemPrompt:     "Be helpful",
			expectSystemInst: true,
		},
		{
			name:             "both system message and prompt",
			messages:         []provider.Message{{Role: "system", Content: "Part 1"}, {Role: "user", Content: "Hi"}},
			systemPrompt:     "Part 2",
			expectSystemInst: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, sysInst := p.convertMessages(tt.messages, tt.systemPrompt)
			if tt.expectSystemInst {
				assert.NotNil(t, sysInst)
				assert.NotEmpty(t, sysInst.Parts)
			} else {
				if sysInst != nil {
					assert.Empty(t, sysInst.Parts)
				}
			}
		})
	}
}
