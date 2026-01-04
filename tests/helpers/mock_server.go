package helpers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

// MockServerResponse represents a configured response for a specific endpoint
type MockServerResponse struct {
	StatusCode  int
	Body        interface{}
	Headers     map[string]string
	Delay       int // milliseconds
	CallCount   int
	LastRequest *http.Request
}

// MockServer is a configurable mock HTTP server for testing
type MockServer struct {
	mu        sync.RWMutex
	server    *httptest.Server
	responses map[string]*MockServerResponse
	t         *testing.T
}

// NewMockServer creates a new mock HTTP server
func NewMockServer(t *testing.T) *MockServer {
	t.Helper()

	ms := &MockServer{
		responses: make(map[string]*MockServerResponse),
		t:         t,
	}

	ms.server = httptest.NewServer(http.HandlerFunc(ms.handler))

	// Register cleanup
	t.Cleanup(func() {
		ms.Close()
	})

	return ms
}

// URL returns the base URL of the mock server
func (ms *MockServer) URL() string {
	return ms.server.URL
}

// Close shuts down the mock server
func (ms *MockServer) Close() {
	if ms.server != nil {
		ms.server.Close()
	}
}

// SetResponse configures a response for a specific path and method
func (ms *MockServer) SetResponse(method, path string, statusCode int, body interface{}) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	key := ms.makeKey(method, path)
	ms.responses[key] = &MockServerResponse{
		StatusCode: statusCode,
		Body:       body,
		Headers:    make(map[string]string),
	}
}

// SetResponseWithHeaders configures a response with custom headers
func (ms *MockServer) SetResponseWithHeaders(method, path string, statusCode int, body interface{}, headers map[string]string) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	key := ms.makeKey(method, path)
	ms.responses[key] = &MockServerResponse{
		StatusCode: statusCode,
		Body:       body,
		Headers:    headers,
	}
}

// SetStreamingResponse configures a server-sent events (SSE) streaming response
func (ms *MockServer) SetStreamingResponse(method, path string, events []string) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	key := ms.makeKey(method, path)

	// Combine events into SSE format
	var sseData strings.Builder
	for _, event := range events {
		sseData.WriteString("data: ")
		sseData.WriteString(event)
		sseData.WriteString("\n\n")
	}

	ms.responses[key] = &MockServerResponse{
		StatusCode: http.StatusOK,
		Body:       sseData.String(),
		Headers: map[string]string{
			"Content-Type":      "text/event-stream",
			"Cache-Control":     "no-cache",
			"Connection":        "keep-alive",
			"Transfer-Encoding": "chunked",
		},
	}
}

// GetCallCount returns the number of times an endpoint was called
func (ms *MockServer) GetCallCount(method, path string) int {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	key := ms.makeKey(method, path)
	if resp, ok := ms.responses[key]; ok {
		return resp.CallCount
	}
	return 0
}

// GetLastRequest returns the last request received for an endpoint
func (ms *MockServer) GetLastRequest(method, path string) *http.Request {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	key := ms.makeKey(method, path)
	if resp, ok := ms.responses[key]; ok {
		return resp.LastRequest
	}
	return nil
}

// Reset clears all configured responses
func (ms *MockServer) Reset() {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.responses = make(map[string]*MockServerResponse)
}

// handler is the HTTP handler for the mock server
func (ms *MockServer) handler(w http.ResponseWriter, r *http.Request) {
	ms.mu.Lock()
	key := ms.makeKey(r.Method, r.URL.Path)
	resp, ok := ms.responses[key]
	if ok {
		resp.CallCount++
		resp.LastRequest = r
	}
	ms.mu.Unlock()

	if !ok {
		ms.t.Logf("No mock response configured for %s %s", r.Method, r.URL.Path)
		http.NotFound(w, r)
		return
	}

	// Set custom headers
	for k, v := range resp.Headers {
		w.Header().Set(k, v)
	}

	// Set status code
	w.WriteHeader(resp.StatusCode)

	// Write body
	switch body := resp.Body.(type) {
	case string:
		fmt.Fprint(w, body)
	case []byte:
		w.Write(body)
	default:
		// Marshal as JSON
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(body); err != nil {
			ms.t.Errorf("Failed to encode response body: %v", err)
		}
	}
}

// makeKey creates a unique key for method and path combination
func (ms *MockServer) makeKey(method, path string) string {
	return fmt.Sprintf("%s:%s", strings.ToUpper(method), path)
}

// MockAnthropicServer creates a mock server that simulates Anthropic API
func MockAnthropicServer(t *testing.T) *MockServer {
	t.Helper()

	ms := NewMockServer(t)

	// Configure default Anthropic chat completion endpoint
	ms.SetResponse("POST", "/v1/messages", http.StatusOK, map[string]interface{}{
		"id":      "msg_test123",
		"type":    "message",
		"role":    "assistant",
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": "This is a mock Anthropic response",
			},
		},
		"model":        "claude-3-5-sonnet-20241022",
		"stop_reason":  "end_turn",
		"stop_sequence": nil,
		"usage": map[string]interface{}{
			"input_tokens":  10,
			"output_tokens": 20,
		},
	})

	return ms
}

// MockOpenAIServer creates a mock server that simulates OpenAI API
func MockOpenAIServer(t *testing.T) *MockServer {
	t.Helper()

	ms := NewMockServer(t)

	// Configure default OpenAI chat completion endpoint
	ms.SetResponse("POST", "/v1/chat/completions", http.StatusOK, map[string]interface{}{
		"id":      "chatcmpl-test123",
		"object":  "chat.completion",
		"created": 1234567890,
		"model":   "gpt-4",
		"choices": []map[string]interface{}{
			{
				"index": 0,
				"message": map[string]interface{}{
					"role":    "assistant",
					"content": "This is a mock OpenAI response",
				},
				"finish_reason": "stop",
			},
		},
		"usage": map[string]interface{}{
			"prompt_tokens":     10,
			"completion_tokens": 20,
			"total_tokens":      30,
		},
	})

	return ms
}

// MockStreamingAnthropicServer creates a mock server with SSE streaming
func MockStreamingAnthropicServer(t *testing.T) *MockServer {
	t.Helper()

	ms := NewMockServer(t)

	// Configure streaming events
	events := []string{
		`{"type":"message_start","message":{"id":"msg_test","type":"message","role":"assistant","content":[],"model":"claude-3-5-sonnet-20241022","stop_reason":null,"stop_sequence":null,"usage":{"input_tokens":10,"output_tokens":0}}}`,
		`{"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`,
		`{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hello"}}`,
		`{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":" world"}}`,
		`{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"!"}}`,
		`{"type":"content_block_stop","index":0}`,
		`{"type":"message_delta","delta":{"stop_reason":"end_turn","stop_sequence":null},"usage":{"output_tokens":5}}`,
		`{"type":"message_stop"}`,
	}

	ms.SetStreamingResponse("POST", "/v1/messages", events)

	return ms
}

// AssertEndpointCalled asserts that an endpoint was called at least once
func AssertEndpointCalled(t *testing.T, ms *MockServer, method, path string) {
	t.Helper()

	count := ms.GetCallCount(method, path)
	if count == 0 {
		t.Errorf("Expected %s %s to be called, but it was not", method, path)
	}
}

// AssertEndpointCalledTimes asserts that an endpoint was called a specific number of times
func AssertEndpointCalledTimes(t *testing.T, ms *MockServer, method, path string, expected int) {
	t.Helper()

	count := ms.GetCallCount(method, path)
	if count != expected {
		t.Errorf("Expected %s %s to be called %d times, but it was called %d times",
			method, path, expected, count)
	}
}
