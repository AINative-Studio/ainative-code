package zerodb

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AINative-studio/ainative-code/internal/client"
)

func TestStoreMemory(t *testing.T) {
	tests := []struct {
		name        string
		req         *MemoryStoreRequest
		serverResp  interface{}
		statusCode  int
		expectError bool
	}{
		{
			name: "successful store",
			req: &MemoryStoreRequest{
				AgentID:   "agent_123",
				Content:   "User prefers dark mode",
				Role:      "user",
				SessionID: "session_abc",
				Metadata: map[string]interface{}{
					"category": "preference",
				},
			},
			serverResp: EmbedAndStoreResponse{
				Success: true,
				Stored:  1,
				IDs:     []string{"memory_xyz"},
			},
			statusCode:  http.StatusOK,
			expectError: false,
		},
		{
			name: "missing agent_id",
			req: &MemoryStoreRequest{
				Content: "Some content",
			},
			expectError: true,
		},
		{
			name: "missing content",
			req: &MemoryStoreRequest{
				AgentID: "agent_123",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectError && tt.req != nil && (tt.req.AgentID == "" || tt.req.Content == "") {
				// Test validation errors
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}))
				defer server.Close()

				apiClient := client.New(client.WithBaseURL(server.URL))
				zdbClient := New(WithAPIClient(apiClient), WithProjectID("test-project"))

				_, err := zdbClient.StoreMemory(context.Background(), tt.req)
				require.Error(t, err)
				return
			}

			// Test successful cases
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/v1/public/test-project/embeddings/embed-and-store", r.URL.Path)
				assert.Equal(t, http.MethodPost, r.Method)

				w.WriteHeader(tt.statusCode)
				json.NewEncoder(w).Encode(tt.serverResp)
			}))
			defer server.Close()

			apiClient := client.New(client.WithBaseURL(server.URL))
			zdbClient := New(WithAPIClient(apiClient), WithProjectID("test-project"))

			memory, err := zdbClient.StoreMemory(context.Background(), tt.req)

			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, memory)
			assert.Contains(t, memory.ID, "memory_")
			assert.Equal(t, "agent_123", memory.AgentID)
			assert.Equal(t, "User prefers dark mode", memory.Content)
		})
	}
}

func TestRetrieveMemory(t *testing.T) {
	tests := []struct {
		name        string
		req         *MemoryRetrieveRequest
		serverResp  interface{}
		statusCode  int
		expectError bool
	}{
		{
			name: "successful retrieve",
			req: &MemoryRetrieveRequest{
				AgentID: "agent_123",
				Query:   "user preferences",
				Limit:   5,
			},
			serverResp: SearchEmbeddingsResponse{
				Results: []SearchResult{
					{
						ID:    "mem_1",
						Text:  "User prefers dark mode",
						Score: 0.95,
						Metadata: map[string]interface{}{
							"agent_id": "agent_123",
							"role":     "user",
						},
					},
					{
						ID:    "mem_2",
						Text:  "User wants email notifications",
						Score: 0.87,
						Metadata: map[string]interface{}{
							"agent_id": "agent_123",
							"role":     "user",
						},
					},
				},
				Total: 2,
			},
			statusCode:  http.StatusOK,
			expectError: false,
		},
		{
			name: "missing agent_id",
			req: &MemoryRetrieveRequest{
				Query: "test query",
			},
			expectError: true,
		},
		{
			name: "missing query",
			req: &MemoryRetrieveRequest{
				AgentID: "agent_123",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectError && tt.req != nil && (tt.req.AgentID == "" || tt.req.Query == "") {
				// Test validation errors
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}))
				defer server.Close()

				apiClient := client.New(client.WithBaseURL(server.URL))
				zdbClient := New(WithAPIClient(apiClient), WithProjectID("test-project"))

				_, err := zdbClient.RetrieveMemory(context.Background(), tt.req)
				require.Error(t, err)
				return
			}

			// Test successful cases
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/v1/public/test-project/embeddings/search", r.URL.Path)
				assert.Equal(t, http.MethodPost, r.Method)

				w.WriteHeader(tt.statusCode)
				json.NewEncoder(w).Encode(tt.serverResp)
			}))
			defer server.Close()

			apiClient := client.New(client.WithBaseURL(server.URL))
			zdbClient := New(WithAPIClient(apiClient), WithProjectID("test-project"))

			memories, err := zdbClient.RetrieveMemory(context.Background(), tt.req)

			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, memories, 2)
			assert.Equal(t, "mem_1", memories[0].ID)
			assert.Equal(t, 0.95, memories[0].Similarity)
			assert.Equal(t, "User prefers dark mode", memories[0].Content)
		})
	}
}

func TestClearMemory(t *testing.T) {
	tests := []struct {
		name        string
		req         *MemoryClearRequest
		expectError bool
	}{
		{
			name: "successful clear all",
			req: &MemoryClearRequest{
				AgentID: "agent_123",
			},
			expectError: false,
		},
		{
			name: "successful clear session",
			req: &MemoryClearRequest{
				AgentID:   "agent_123",
				SessionID: "session_abc",
			},
			expectError: false,
		},
		{
			name: "missing agent_id",
			req: &MemoryClearRequest{
				SessionID: "session_abc",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectError && tt.req != nil && tt.req.AgentID == "" {
				// Test validation errors
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}))
				defer server.Close()

				apiClient := client.New(client.WithBaseURL(server.URL))
				zdbClient := New(WithAPIClient(apiClient), WithProjectID("test-project"))

				_, err := zdbClient.ClearMemory(context.Background(), tt.req)
				require.Error(t, err)
				return
			}

			// Test successful cases - ClearMemory does search + delete
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/v1/public/test-project/embeddings/search" {
					// Return mock search results
					searchResp := SearchEmbeddingsResponse{
						Results: []SearchResult{
							{ID: "mem_1", Text: "Memory 1"},
							{ID: "mem_2", Text: "Memory 2"},
						},
						Total: 2,
					}
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(searchResp)
				} else if r.Method == http.MethodDelete {
					// Handle delete requests
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
				}
			}))
			defer server.Close()

			apiClient := client.New(client.WithBaseURL(server.URL))
			zdbClient := New(WithAPIClient(apiClient), WithProjectID("test-project"))

			resp, err := zdbClient.ClearMemory(context.Background(), tt.req)

			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, 2, resp.Deleted)
		})
	}
}

func TestListMemory(t *testing.T) {
	tests := []struct {
		name        string
		req         *MemoryListRequest
		serverResp  interface{}
		statusCode  int
		expectError bool
	}{
		{
			name: "successful list",
			req: &MemoryListRequest{
				AgentID: "agent_123",
				Limit:   50,
				Offset:  0,
			},
			serverResp: SearchEmbeddingsResponse{
				Results: []SearchResult{
					{
						ID:   "mem_1",
						Text: "First memory",
						Metadata: map[string]interface{}{
							"agent_id": "agent_123",
						},
					},
					{
						ID:   "mem_2",
						Text: "Second memory",
						Metadata: map[string]interface{}{
							"agent_id": "agent_123",
						},
					},
				},
				Total: 2,
			},
			statusCode:  http.StatusOK,
			expectError: false,
		},
		{
			name: "list with session filter",
			req: &MemoryListRequest{
				AgentID:   "agent_123",
				SessionID: "session_abc",
				Limit:     10,
			},
			serverResp: SearchEmbeddingsResponse{
				Results: []SearchResult{
					{
						ID:   "mem_1",
						Text: "Session memory",
						Metadata: map[string]interface{}{
							"agent_id":   "agent_123",
							"session_id": "session_abc",
						},
					},
				},
				Total: 1,
			},
			statusCode:  http.StatusOK,
			expectError: false,
		},
		{
			name: "missing agent_id",
			req: &MemoryListRequest{
				Limit: 10,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectError && tt.req != nil && tt.req.AgentID == "" {
				// Test validation errors
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}))
				defer server.Close()

				apiClient := client.New(client.WithBaseURL(server.URL))
				zdbClient := New(WithAPIClient(apiClient), WithProjectID("test-project"))

				_, _, err := zdbClient.ListMemory(context.Background(), tt.req)
				require.Error(t, err)
				return
			}

			// Test successful cases
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/v1/public/test-project/embeddings/search", r.URL.Path)
				assert.Equal(t, http.MethodPost, r.Method)

				w.WriteHeader(tt.statusCode)
				json.NewEncoder(w).Encode(tt.serverResp)
			}))
			defer server.Close()

			apiClient := client.New(client.WithBaseURL(server.URL))
			zdbClient := New(WithAPIClient(apiClient), WithProjectID("test-project"))

			memories, total, err := zdbClient.ListMemory(context.Background(), tt.req)

			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, memories)
			assert.Greater(t, total, 0)
			assert.Greater(t, len(memories), 0)
		})
	}
}
