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
			serverResp: MemoryStoreResponse{
				Memory: &Memory{
					ID:        "mem_xyz",
					AgentID:   "agent_123",
					Content:   "User prefers dark mode",
					Role:      "user",
					SessionID: "session_abc",
				},
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
				assert.Equal(t, "/api/v1/projects/test-project/memory/store", r.URL.Path)
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
			assert.Equal(t, "mem_xyz", memory.ID)
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
			serverResp: MemoryRetrieveResponse{
				Memories: []*Memory{
					{
						ID:         "mem_1",
						AgentID:    "agent_123",
						Content:    "User prefers dark mode",
						Similarity: 0.95,
					},
					{
						ID:         "mem_2",
						AgentID:    "agent_123",
						Content:    "User wants email notifications",
						Similarity: 0.87,
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
				assert.Equal(t, "/api/v1/projects/test-project/memory/retrieve", r.URL.Path)
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
		})
	}
}

func TestClearMemory(t *testing.T) {
	tests := []struct {
		name        string
		req         *MemoryClearRequest
		serverResp  interface{}
		statusCode  int
		expectError bool
	}{
		{
			name: "successful clear all",
			req: &MemoryClearRequest{
				AgentID: "agent_123",
			},
			serverResp: MemoryClearResponse{
				Deleted: 10,
				Message: "All memories cleared",
			},
			statusCode:  http.StatusOK,
			expectError: false,
		},
		{
			name: "successful clear session",
			req: &MemoryClearRequest{
				AgentID:   "agent_123",
				SessionID: "session_abc",
			},
			serverResp: MemoryClearResponse{
				Deleted: 5,
				Message: "Session memories cleared",
			},
			statusCode:  http.StatusOK,
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

			// Test successful cases
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/v1/projects/test-project/memory/clear", r.URL.Path)
				assert.Equal(t, http.MethodPost, r.Method)

				w.WriteHeader(tt.statusCode)
				json.NewEncoder(w).Encode(tt.serverResp)
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
			assert.Greater(t, resp.Deleted, 0)
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
			serverResp: MemoryListResponse{
				Memories: []*Memory{
					{
						ID:      "mem_1",
						AgentID: "agent_123",
						Content: "First memory",
					},
					{
						ID:      "mem_2",
						AgentID: "agent_123",
						Content: "Second memory",
					},
				},
				Total:  2,
				Limit:  50,
				Offset: 0,
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
			serverResp: MemoryListResponse{
				Memories: []*Memory{
					{
						ID:        "mem_1",
						AgentID:   "agent_123",
						SessionID: "session_abc",
						Content:   "Session memory",
					},
				},
				Total:  1,
				Limit:  10,
				Offset: 0,
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
				assert.Equal(t, "/api/v1/projects/test-project/memory/list", r.URL.Path)
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
