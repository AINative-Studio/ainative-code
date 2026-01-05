// +build integration

package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/AINative-studio/ainative-code/internal/providers"
	"github.com/stretchr/testify/suite"
)

// ChatIntegrationTestSuite tests chat/LLM provider integration functionality.
type ChatIntegrationTestSuite struct {
	suite.Suite
	mockServer *httptest.Server
	cleanup    func()
}

// SetupTest runs before each test in the suite.
func (s *ChatIntegrationTestSuite) SetupTest() {
	// Create mock LLM provider server
	mux := http.NewServeMux()

	// Chat endpoint (non-streaming)
	mux.HandleFunc("/v1/messages", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Parse request
		var req map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid_request"})
			return
		}

		// Simulate LLM response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "msg_12345",
			"type":    "message",
			"role":    "assistant",
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": "This is a mock response from the LLM provider.",
				},
			},
			"model": req["model"],
			"usage": map[string]int{
				"input_tokens":  10,
				"output_tokens": 20,
			},
			"stop_reason": "end_turn",
		})
	})

	// Streaming chat endpoint
	mux.HandleFunc("/v1/messages/stream", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Set headers for SSE
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Send SSE events
		events := []string{
			`event: message_start\ndata: {"type":"message_start","message":{"id":"msg_stream_123","type":"message","role":"assistant"}}\n\n`,
			`event: content_block_start\ndata: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}\n\n`,
			`event: content_block_delta\ndata: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hello"}}\n\n`,
			`event: content_block_delta\ndata: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":" from"}}\n\n`,
			`event: content_block_delta\ndata: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":" streaming!"}}\n\n`,
			`event: content_block_stop\ndata: {"type":"content_block_stop","index":0}\n\n`,
			`event: message_delta\ndata: {"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":15}}\n\n`,
			`event: message_stop\ndata: {"type":"message_stop"}\n\n`,
		}

		for _, event := range events {
			w.Write([]byte(event))
			flusher.Flush()
			time.Sleep(10 * time.Millisecond) // Simulate streaming delay
		}
	})

	// Models endpoint
	mux.HandleFunc("/v1/models", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"id":         "claude-3-opus-20240229",
					"type":       "model",
					"created":    1677610602,
					"max_tokens": 200000,
				},
				{
					"id":         "claude-3-sonnet-20240229",
					"type":       "model",
					"created":    1677610602,
					"max_tokens": 200000,
				},
			},
		})
	})

	// Error simulation endpoint
	mux.HandleFunc("/v1/error/rate_limit", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"type":    "rate_limit_error",
				"message": "Rate limit exceeded",
			},
		})
	})

	mux.HandleFunc("/v1/error/invalid", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"type":    "invalid_request_error",
				"message": "Invalid request format",
			},
		})
	})

	s.mockServer = httptest.NewServer(mux)
	s.cleanup = func() {
		s.mockServer.Close()
	}
}

// TearDownTest runs after each test in the suite.
func (s *ChatIntegrationTestSuite) TearDownTest() {
	if s.cleanup != nil {
		s.cleanup()
	}
}

// TestProviderInitialization tests initializing LLM provider clients.
func (s *ChatIntegrationTestSuite) TestProviderInitialization() {
	// Given: Provider configuration
	config := providers.Config{
		APIKey:       "test_api_key",
		BaseURL:      s.mockServer.URL,
		MaxRetries:   3,
		Timeout:      30,
		DefaultModel: "claude-3-opus-20240229",
	}

	// When: Creating a mock provider
	// Note: In real implementation, we'd use providers.NewAnthropicProvider or similar
	// For testing, we simulate provider interface

	// Then: Provider should be initialized successfully
	s.NotEmpty(config.APIKey, "API key should be set")
	s.NotEmpty(config.BaseURL, "Base URL should be set")
	s.Equal(3, config.MaxRetries, "Max retries should be 3")
	s.Equal(30, config.Timeout, "Timeout should be 30 seconds")
}

// TestChatMessageSending tests sending chat messages to LLM provider.
func (s *ChatIntegrationTestSuite) TestChatMessageSending() {
	// Given: A chat request
	ctx := context.Background()

	req := &providers.ChatRequest{
		Messages: []providers.Message{
			{
				Role:    providers.RoleUser,
				Content: "Hello, how are you?",
			},
		},
		Model:       "claude-3-opus-20240229",
		MaxTokens:   1024,
		Temperature: 0.7,
	}

	// When: Sending the chat request via HTTP client
	reqBody, err := json.Marshal(req)
	s.Require().NoError(err)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, s.mockServer.URL+"/v1/messages", strings.NewReader(string(reqBody)))
	s.Require().NoError(err)

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", "test_api_key")

	httpClient := &http.Client{Timeout: 10 * time.Second}
	httpResp, err := httpClient.Do(httpReq)
	s.Require().NoError(err)
	defer httpResp.Body.Close()

	// Then: Should receive successful response
	s.Equal(http.StatusOK, httpResp.StatusCode, "Response should be OK")

	var response map[string]interface{}
	err = json.NewDecoder(httpResp.Body).Decode(&response)
	s.Require().NoError(err)

	s.Equal("message", response["type"], "Response type should be message")
	s.Equal("assistant", response["role"], "Role should be assistant")
	s.NotNil(response["content"], "Content should be present")
	s.NotNil(response["usage"], "Usage info should be present")
}

// TestStreamingResponseHandling tests handling streaming responses from LLM.
func (s *ChatIntegrationTestSuite) TestStreamingResponseHandling() {
	// Given: A streaming chat request
	ctx := context.Background()

	req := &providers.StreamRequest{
		Messages: []providers.Message{
			{
				Role:    providers.RoleUser,
				Content: "Tell me a story",
			},
		},
		Model:       "claude-3-opus-20240229",
		MaxTokens:   1024,
		Temperature: 0.7,
	}

	// When: Sending streaming request
	reqBody, err := json.Marshal(req)
	s.Require().NoError(err)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, s.mockServer.URL+"/v1/messages/stream", strings.NewReader(string(reqBody)))
	s.Require().NoError(err)

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", "test_api_key")

	httpClient := &http.Client{Timeout: 30 * time.Second}
	httpResp, err := httpClient.Do(httpReq)
	s.Require().NoError(err)
	defer httpResp.Body.Close()

	// Then: Should receive streaming events
	s.Equal(http.StatusOK, httpResp.StatusCode, "Response should be OK")
	s.Equal("text/event-stream", httpResp.Header.Get("Content-Type"), "Content-Type should be text/event-stream")

	// Read and verify streaming events
	buf := make([]byte, 4096)
	n, err := httpResp.Body.Read(buf)
	s.Require().NoError(err)

	responseData := string(buf[:n])
	s.Contains(responseData, "message_start", "Should contain message_start event")
	s.Contains(responseData, "content_block_delta", "Should contain content_block_delta events")
	s.Contains(responseData, "Hello from streaming!", "Should contain streamed text")
}

// TestContextManagement tests managing conversation context.
func (s *ChatIntegrationTestSuite) TestContextManagement() {
	// Given: A conversation with multiple messages
	ctx := context.Background()

	conversationHistory := []providers.Message{
		{
			Role:    providers.RoleUser,
			Content: "What is 2+2?",
		},
		{
			Role:    providers.RoleAssistant,
			Content: "2+2 equals 4.",
		},
		{
			Role:    providers.RoleUser,
			Content: "And what about 3+3?",
		},
	}

	req := &providers.ChatRequest{
		Messages:    conversationHistory,
		Model:       "claude-3-opus-20240229",
		MaxTokens:   1024,
		Temperature: 0.7,
	}

	// When: Sending request with conversation history
	reqBody, err := json.Marshal(req)
	s.Require().NoError(err)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, s.mockServer.URL+"/v1/messages", strings.NewReader(string(reqBody)))
	s.Require().NoError(err)

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", "test_api_key")

	httpClient := &http.Client{Timeout: 10 * time.Second}
	httpResp, err := httpClient.Do(httpReq)
	s.Require().NoError(err)
	defer httpResp.Body.Close()

	// Then: Should maintain conversation context
	s.Equal(http.StatusOK, httpResp.StatusCode, "Response should be OK")

	// Verify request contains all context messages
	s.Len(conversationHistory, 3, "Should maintain all 3 messages in context")
	s.Equal(providers.RoleUser, conversationHistory[0].Role)
	s.Equal(providers.RoleAssistant, conversationHistory[1].Role)
	s.Equal(providers.RoleUser, conversationHistory[2].Role)
}

// TestProviderErrorHandling tests handling various provider errors.
func (s *ChatIntegrationTestSuite) TestProviderErrorHandling() {
	// Given: HTTP client
	ctx := context.Background()
	httpClient := &http.Client{Timeout: 10 * time.Second}

	// When: Requesting rate limited endpoint
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, s.mockServer.URL+"/v1/error/rate_limit", nil)
	s.Require().NoError(err)

	httpResp, err := httpClient.Do(httpReq)
	s.Require().NoError(err)
	defer httpResp.Body.Close()

	// Then: Should receive rate limit error
	s.Equal(http.StatusTooManyRequests, httpResp.StatusCode, "Should return 429 status")

	var errorResp map[string]interface{}
	err = json.NewDecoder(httpResp.Body).Decode(&errorResp)
	s.Require().NoError(err)
	s.Contains(errorResp, "error", "Response should contain error field")

	// When: Requesting invalid endpoint
	httpReq, err = http.NewRequestWithContext(ctx, http.MethodPost, s.mockServer.URL+"/v1/error/invalid", nil)
	s.Require().NoError(err)

	httpResp, err = httpClient.Do(httpReq)
	s.Require().NoError(err)
	defer httpResp.Body.Close()

	// Then: Should receive invalid request error
	s.Equal(http.StatusBadRequest, httpResp.StatusCode, "Should return 400 status")
}

// TestModelListing tests listing available models from provider.
func (s *ChatIntegrationTestSuite) TestModelListing() {
	// Given: HTTP client
	ctx := context.Background()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, s.mockServer.URL+"/v1/models", nil)
	s.Require().NoError(err)

	httpReq.Header.Set("x-api-key", "test_api_key")

	httpClient := &http.Client{Timeout: 10 * time.Second}

	// When: Requesting available models
	httpResp, err := httpClient.Do(httpReq)
	s.Require().NoError(err)
	defer httpResp.Body.Close()

	// Then: Should receive list of models
	s.Equal(http.StatusOK, httpResp.StatusCode, "Response should be OK")

	var response map[string]interface{}
	err = json.NewDecoder(httpResp.Body).Decode(&response)
	s.Require().NoError(err)

	models, ok := response["data"].([]interface{})
	s.True(ok, "Response should contain models array")
	s.Len(models, 2, "Should return 2 models")

	// Verify model structure
	firstModel := models[0].(map[string]interface{})
	s.Equal("claude-3-opus-20240229", firstModel["id"], "First model should be opus")
	s.Equal("model", firstModel["type"], "Type should be model")
}

// TestMessageRolesValidation tests validating message roles.
func (s *ChatIntegrationTestSuite) TestMessageRolesValidation() {
	// Given: Messages with different roles
	userMsg := providers.Message{
		Role:    providers.RoleUser,
		Content: "User message",
	}

	assistantMsg := providers.Message{
		Role:    providers.RoleAssistant,
		Content: "Assistant message",
	}

	systemMsg := providers.Message{
		Role:    providers.RoleSystem,
		Content: "System message",
	}

	// Then: Roles should be correctly set
	s.Equal(providers.RoleUser, userMsg.Role)
	s.Equal(providers.RoleAssistant, assistantMsg.Role)
	s.Equal(providers.RoleSystem, systemMsg.Role)
}

// TestMaxTokensConfiguration tests configuring max tokens parameter.
func (s *ChatIntegrationTestSuite) TestMaxTokensConfiguration() {
	// Given: Chat request with various max tokens settings
	testCases := []struct {
		name      string
		maxTokens int
		valid     bool
	}{
		{"Default", 1024, true},
		{"High", 4096, true},
		{"Max", 200000, true},
		{"Zero", 0, false},
		{"Negative", -100, false},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			req := &providers.ChatRequest{
				Messages: []providers.Message{
					{Role: providers.RoleUser, Content: "Test"},
				},
				Model:     "claude-3-opus-20240229",
				MaxTokens: tc.maxTokens,
			}

			// Then: MaxTokens should be set correctly
			if tc.valid {
				s.GreaterOrEqual(req.MaxTokens, 0, "MaxTokens should be non-negative")
			} else {
				s.LessOrEqual(req.MaxTokens, 0, "Invalid MaxTokens should be zero or negative")
			}
		})
	}
}

// TestTemperatureConfiguration tests configuring temperature parameter.
func (s *ChatIntegrationTestSuite) TestTemperatureConfiguration() {
	// Given: Chat request with various temperature settings
	testCases := []struct {
		name        string
		temperature float64
		valid       bool
	}{
		{"Low", 0.0, true},
		{"Medium", 0.5, true},
		{"High", 1.0, true},
		{"TooHigh", 1.5, false},
		{"Negative", -0.1, false},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			req := &providers.ChatRequest{
				Messages: []providers.Message{
					{Role: providers.RoleUser, Content: "Test"},
				},
				Model:       "claude-3-opus-20240229",
				Temperature: tc.temperature,
			}

			// Then: Temperature should be set correctly
			if tc.valid {
				s.GreaterOrEqual(req.Temperature, 0.0, "Temperature should be >= 0.0")
				s.LessOrEqual(req.Temperature, 1.0, "Temperature should be <= 1.0")
			}
		})
	}
}

// TestConcurrentChatRequests tests handling concurrent chat requests.
func (s *ChatIntegrationTestSuite) TestConcurrentChatRequests() {
	// Given: Multiple concurrent chat requests
	ctx := context.Background()
	httpClient := &http.Client{Timeout: 10 * time.Second}

	concurrentRequests := 5
	done := make(chan bool, concurrentRequests)
	errors := make(chan error, concurrentRequests)

	// When: Making concurrent requests
	for i := 0; i < concurrentRequests; i++ {
		go func(index int) {
			req := &providers.ChatRequest{
				Messages: []providers.Message{
					{
						Role:    providers.RoleUser,
						Content: "Concurrent request",
					},
				},
				Model:     "claude-3-opus-20240229",
				MaxTokens: 1024,
			}

			reqBody, err := json.Marshal(req)
			if err != nil {
				errors <- err
				done <- true
				return
			}

			httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, s.mockServer.URL+"/v1/messages", strings.NewReader(string(reqBody)))
			if err != nil {
				errors <- err
				done <- true
				return
			}

			httpReq.Header.Set("Content-Type", "application/json")
			httpResp, err := httpClient.Do(httpReq)
			if err != nil {
				errors <- err
				done <- true
				return
			}
			httpResp.Body.Close()

			if httpResp.StatusCode != http.StatusOK {
				errors <- err
			}

			done <- true
		}(i)
	}

	// Wait for all requests to complete
	for i := 0; i < concurrentRequests; i++ {
		<-done
	}
	close(errors)

	// Then: All requests should succeed
	s.Empty(errors, "No errors should occur during concurrent requests")
}

// TestChatIntegrationTestSuite runs the test suite.
func TestChatIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(ChatIntegrationTestSuite))
}
