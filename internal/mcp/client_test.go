package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	server := &Server{
		Name:    "test-server",
		URL:     "http://localhost:8080",
		Timeout: 10 * time.Second,
		Enabled: true,
	}

	client := NewClient(server)
	assert.NotNil(t, client)
	assert.Equal(t, server, client.server)
	assert.NotNil(t, client.httpClient)
}

func TestNewClient_DefaultTimeout(t *testing.T) {
	server := &Server{
		Name:    "test-server",
		URL:     "http://localhost:8080",
		Enabled: true,
	}

	client := NewClient(server)
	assert.Equal(t, 30*time.Second, client.httpClient.Timeout)
}

func TestListTools(t *testing.T) {
	// Create mock server
	handler := func(w http.ResponseWriter, r *http.Request) {
		var req JSONRPCRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Equal(t, "tools/list", req.Method)

		resp := JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: ListToolsResult{
				Tools: []Tool{
					{
						Name:        "test_tool",
						Description: "A test tool",
						InputSchema: map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"message": map[string]interface{}{
									"type": "string",
								},
							},
						},
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}

	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	client := NewClient(&Server{
		Name: "test",
		URL:  server.URL,
	})

	tools, err := client.ListTools(context.Background())
	require.NoError(t, err)
	require.Len(t, tools, 1)
	assert.Equal(t, "test_tool", tools[0].Name)
	assert.Equal(t, "A test tool", tools[0].Description)
}

func TestListTools_Pagination(t *testing.T) {
	page := 0
	handler := func(w http.ResponseWriter, r *http.Request) {
		var req JSONRPCRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		resp := JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
		}

		if page == 0 {
			resp.Result = ListToolsResult{
				Tools: []Tool{
					{Name: "tool1", Description: "Tool 1"},
				},
				NextCursor: "page2",
			}
			page++
		} else {
			resp.Result = ListToolsResult{
				Tools: []Tool{
					{Name: "tool2", Description: "Tool 2"},
				},
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}

	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	client := NewClient(&Server{
		Name: "test",
		URL:  server.URL,
	})

	tools, err := client.ListTools(context.Background())
	require.NoError(t, err)
	require.Len(t, tools, 2)
	assert.Equal(t, "tool1", tools[0].Name)
	assert.Equal(t, "tool2", tools[1].Name)
}

func TestCallTool(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		var req JSONRPCRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Equal(t, "tools/call", req.Method)

		resp := JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: ToolResult{
				Content: []ResultContent{
					{
						Type: "text",
						Text: "Hello, World!",
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}

	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	client := NewClient(&Server{
		Name: "test",
		URL:  server.URL,
	})

	result, err := client.CallTool(context.Background(), "test_tool", map[string]interface{}{
		"message": "hello",
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Content, 1)
	assert.Equal(t, "text", result.Content[0].Type)
	assert.Equal(t, "Hello, World!", result.Content[0].Text)
}

func TestCallTool_Error(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		var req JSONRPCRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		resp := JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &RPCError{
				Code:    -32600,
				Message: "Tool not found",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}

	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	client := NewClient(&Server{
		Name: "test",
		URL:  server.URL,
	})

	result, err := client.CallTool(context.Background(), "nonexistent", nil)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Tool not found")
}

func TestPing(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		var req JSONRPCRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Equal(t, "ping", req.Method)

		resp := JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  "pong",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}

	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	client := NewClient(&Server{
		Name: "test",
		URL:  server.URL,
	})

	err := client.Ping(context.Background())
	assert.NoError(t, err)
}

func TestCheckHealth(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		resp := JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      1,
			Result:  "pong",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}

	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	client := NewClient(&Server{
		Name: "test",
		URL:  server.URL,
	})

	status := client.CheckHealth(context.Background())
	assert.True(t, status.Healthy)
	assert.Empty(t, status.Error)
	assert.Greater(t, status.ResponseTime, time.Duration(0))
}

func TestCheckHealth_Unhealthy(t *testing.T) {
	client := NewClient(&Server{
		Name:    "test",
		URL:     "http://invalid-server:9999",
		Timeout: 1 * time.Second,
	})

	status := client.CheckHealth(context.Background())
	assert.False(t, status.Healthy)
	assert.NotEmpty(t, status.Error)
}

func TestCall_HTTPError(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}

	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	client := NewClient(&Server{
		Name: "test",
		URL:  server.URL,
	})

	var result interface{}
	err := client.call(context.Background(), "test", nil, &result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP error: 500")
}

func TestCall_InvalidJSON(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json"))
	}

	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	client := NewClient(&Server{
		Name: "test",
		URL:  server.URL,
	})

	var result interface{}
	err := client.call(context.Background(), "test", nil, &result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse response")
}

func TestCall_ContextCanceled(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
	}

	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	client := NewClient(&Server{
		Name: "test",
		URL:  server.URL,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	var result interface{}
	err := client.call(ctx, "test", nil, &result)
	assert.Error(t, err)
}

func TestCall_CustomHeaders(t *testing.T) {
	headerReceived := false
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Custom-Header") == "test-value" {
			headerReceived = true
		}

		resp := JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      1,
			Result:  "ok",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}

	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	client := NewClient(&Server{
		Name: "test",
		URL:  server.URL,
		Headers: map[string]string{
			"X-Custom-Header": "test-value",
		},
	})

	var result interface{}
	err := client.call(context.Background(), "test", nil, &result)
	assert.NoError(t, err)
	assert.True(t, headerReceived)
}

func TestGetServer(t *testing.T) {
	server := &Server{
		Name: "test",
		URL:  "http://localhost:8080",
	}

	client := NewClient(server)
	assert.Equal(t, server, client.GetServer())
}
