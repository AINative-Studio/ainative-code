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
				Success:             true,
				VectorsStored:       1,
				EmbeddingsGenerated: 1,
				Model:               "BAAI/bge-small-en-v1.5",
				Dimensions:          384,
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
						VectorID:       "mem_1",
						Document:       "User prefers dark mode",
						Score:          0.95,
						VectorMetadata: map[string]interface{}{
							"agent_id": "agent_123",
							"role":     "user",
						},
						Namespace: "agent_memories",
						CreatedAt: "2024-01-15T10:00:00Z",
					},
					{
						VectorID:       "mem_2",
						Document:       "User wants email notifications",
						Score:          0.87,
						VectorMetadata: map[string]interface{}{
							"agent_id": "agent_123",
							"role":     "user",
						},
						Namespace: "agent_memories",
						CreatedAt: "2024-01-15T10:05:00Z",
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
							{VectorID: "mem_1", Document: "Memory 1", Namespace: "agent_memories", CreatedAt: "2024-01-15T10:00:00Z"},
							{VectorID: "mem_2", Document: "Memory 2", Namespace: "agent_memories", CreatedAt: "2024-01-15T10:01:00Z"},
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
						VectorID: "mem_1",
						Document: "First memory",
						VectorMetadata: map[string]interface{}{
							"agent_id": "agent_123",
						},
						Namespace: "agent_memories",
						CreatedAt: "2024-01-15T10:00:00Z",
					},
					{
						VectorID: "mem_2",
						Document: "Second memory",
						VectorMetadata: map[string]interface{}{
							"agent_id": "agent_123",
						},
						Namespace: "agent_memories",
						CreatedAt: "2024-01-15T10:01:00Z",
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
						VectorID: "mem_1",
						Document: "Session memory",
						VectorMetadata: map[string]interface{}{
							"agent_id":   "agent_123",
							"session_id": "session_abc",
						},
						Namespace: "agent_memories",
						CreatedAt: "2024-01-15T10:00:00Z",
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
