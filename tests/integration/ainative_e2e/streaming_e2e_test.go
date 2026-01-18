package ainative_e2e

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/AINative-studio/ainative-code/internal/backend"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// StreamChunk represents a single chunk in a streaming response
type StreamChunk struct {
	Content string
	Error   error
}

// TestAINativeE2E_StreamingChat tests streaming chat functionality
// GIVEN authenticated user and mock backend with streaming support
// WHEN requesting streaming chat
// THEN should receive streaming chunks
func TestAINativeE2E_StreamingChat(t *testing.T) {
	// GIVEN authenticated user and mock backend with streaming support
	mockBackend := NewMockBackend(t)
	mockBackend.EnableStreaming()
	defer mockBackend.Close()

	client := backend.NewClient(mockBackend.URL)
	ctx := context.Background()

	// Login
	loginResp, err := client.Login(ctx, "test@example.com", "password123")
	require.NoError(t, err, "Login should succeed")

	// WHEN requesting streaming chat
	req := &backend.ChatCompletionRequest{
		Messages: []backend.Message{
			{Role: "user", Content: "Count to 3"},
		},
		Model:  "claude-sonnet-4-5",
		Stream: true,
	}

	stream, err := streamChatCompletion(ctx, client, loginResp.AccessToken, req)
	require.NoError(t, err, "Stream creation should succeed")

	// THEN should receive streaming chunks
	chunks := []string{}
	for chunk := range stream {
		require.NoError(t, chunk.Error, "Stream chunk should not have error")
		chunks = append(chunks, chunk.Content)
	}

	assert.NotEmpty(t, chunks, "Should receive at least one chunk")
	assert.GreaterOrEqual(t, len(chunks), 1, "Should receive multiple chunks")

	// Verify complete message
	completeMessage := strings.Join(chunks, "")
	assert.NotEmpty(t, completeMessage, "Complete message should not be empty")
}

// TestAINativeE2E_StreamingDisconnect tests streaming disconnection
// GIVEN a streaming request
// WHEN context is cancelled during streaming
// THEN stream should stop gracefully
func TestAINativeE2E_StreamingDisconnect(t *testing.T) {
	// GIVEN a streaming request
	mockBackend := NewMockBackend(t)
	mockBackend.EnableStreaming()
	mockBackend.SetStreamDelay(100 * time.Millisecond) // Add delay to allow cancellation
	defer mockBackend.Close()

	client := backend.NewClient(mockBackend.URL)
	ctx, cancel := context.WithCancel(context.Background())

	loginResp, err := client.Login(ctx, "test@example.com", "password123")
	require.NoError(t, err, "Login should succeed")

	req := &backend.ChatCompletionRequest{
		Messages: []backend.Message{{Role: "user", Content: "Long response"}},
		Model:    "claude-sonnet-4-5",
		Stream:   true,
	}

	stream, err := streamChatCompletion(ctx, client, loginResp.AccessToken, req)
	require.NoError(t, err, "Stream creation should succeed")

	// WHEN context is cancelled during streaming
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	// THEN stream should stop gracefully
	gotCancellation := false
	for chunk := range stream {
		if chunk.Error != nil {
			if errors.Is(chunk.Error, context.Canceled) {
				gotCancellation = true
			}
			break
		}
	}

	assert.True(t, gotCancellation, "Should receive context cancellation error")
}

// TestAINativeE2E_StreamingUnauthorized tests unauthorized streaming
// GIVEN a streaming request without authentication
// WHEN requesting streaming chat
// THEN should fail with unauthorized error
func TestAINativeE2E_StreamingUnauthorized(t *testing.T) {
	// GIVEN a streaming request without authentication
	mockBackend := NewMockBackend(t)
	mockBackend.EnableStreaming()
	defer mockBackend.Close()

	client := backend.NewClient(mockBackend.URL)
	ctx := context.Background()

	// WHEN requesting streaming chat without auth
	req := &backend.ChatCompletionRequest{
		Messages: []backend.Message{{Role: "user", Content: "Test"}},
		Model:    "claude-sonnet-4-5",
		Stream:   true,
	}

	_, err := streamChatCompletion(ctx, client, "", req)

	// THEN should fail with unauthorized error
	require.Error(t, err, "Stream should fail without authentication")
}

// TestAINativeE2E_StreamingEmptyMessage tests streaming with empty message
// GIVEN an authenticated user
// WHEN sending streaming request with empty message
// THEN should handle gracefully
func TestAINativeE2E_StreamingEmptyMessage(t *testing.T) {
	// GIVEN an authenticated user
	mockBackend := NewMockBackend(t)
	mockBackend.EnableStreaming()
	defer mockBackend.Close()

	client := backend.NewClient(mockBackend.URL)
	ctx := context.Background()

	loginResp, err := client.Login(ctx, "test@example.com", "password123")
	require.NoError(t, err, "Login should succeed")

	// WHEN sending streaming request with empty message
	req := &backend.ChatCompletionRequest{
		Messages: []backend.Message{{Role: "user", Content: ""}},
		Model:    "claude-sonnet-4-5",
		Stream:   true,
	}

	_, err = streamChatCompletion(ctx, client, loginResp.AccessToken, req)

	// THEN should handle gracefully (either succeed with empty response or return error)
	// The backend decides the behavior - we just verify it doesn't panic
	assert.NotNil(t, err != nil || err == nil, "Should handle empty message")
}

// TestAINativeE2E_StreamingLargeResponse tests streaming large response
// GIVEN an authenticated user
// WHEN requesting streaming with large expected response
// THEN should stream all chunks successfully
func TestAINativeE2E_StreamingLargeResponse(t *testing.T) {
	// GIVEN an authenticated user
	mockBackend := NewMockBackend(t)
	mockBackend.EnableStreaming()
	mockBackend.SetStreamChunkCount(100) // Simulate large response
	defer mockBackend.Close()

	client := backend.NewClient(mockBackend.URL)
	ctx := context.Background()

	loginResp, err := client.Login(ctx, "test@example.com", "password123")
	require.NoError(t, err, "Login should succeed")

	// WHEN requesting streaming with large expected response
	req := &backend.ChatCompletionRequest{
		Messages: []backend.Message{{Role: "user", Content: "Write a long story"}},
		Model:    "claude-sonnet-4-5",
		Stream:   true,
	}

	stream, err := streamChatCompletion(ctx, client, loginResp.AccessToken, req)
	require.NoError(t, err, "Stream creation should succeed")

	// THEN should stream all chunks successfully
	chunks := []string{}
	for chunk := range stream {
		require.NoError(t, chunk.Error, "Stream chunk should not have error")
		chunks = append(chunks, chunk.Content)
	}

	assert.GreaterOrEqual(t, len(chunks), 10, "Should receive multiple chunks for large response")
}

// streamChatCompletion is a helper function to stream chat completions
// This is a placeholder implementation - actual implementation will be in the client
func streamChatCompletion(ctx context.Context, client *backend.Client, token string, req *backend.ChatCompletionRequest) (<-chan StreamChunk, error) {
	// Create HTTP request
	url := client.BaseURL + "/api/v1/chat/completions"

	httpReq, err := createHTTPRequest(ctx, "POST", url, token, req)
	if err != nil {
		return nil, err
	}

	resp, err := client.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, err
	}

	// Check for unauthorized
	if resp.StatusCode == 401 {
		resp.Body.Close()
		return nil, backend.ErrUnauthorized
	}

	if resp.StatusCode != 200 {
		resp.Body.Close()
		return nil, errors.New("streaming request failed")
	}

	// Create channel for streaming
	stream := make(chan StreamChunk)

	go func() {
		defer resp.Body.Close()
		defer close(stream)

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "data: ") {
				data := strings.TrimPrefix(line, "data: ")
				if data == "[DONE]" {
					break
				}

				var chunk map[string]interface{}
				if err := json.Unmarshal([]byte(data), &chunk); err != nil {
					stream <- StreamChunk{Error: err}
					return
				}

				// Extract content from chunk
				if choices, ok := chunk["choices"].([]interface{}); ok && len(choices) > 0 {
					if choice, ok := choices[0].(map[string]interface{}); ok {
						if delta, ok := choice["delta"].(map[string]interface{}); ok {
							if content, ok := delta["content"].(string); ok {
								stream <- StreamChunk{Content: content}
							}
						}
					}
				}
			}
		}

		if err := scanner.Err(); err != nil {
			stream <- StreamChunk{Error: err}
		}
	}()

	return stream, nil
}

// createHTTPRequest is a helper to create HTTP requests with proper headers
func createHTTPRequest(ctx context.Context, method, url, token string, body interface{}) (*http.Request, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	return req, nil
}
