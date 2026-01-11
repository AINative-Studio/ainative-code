package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestRegistry creates a registry with isolated config for testing
func setupTestRegistry(t *testing.T) *Registry {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.mcp.json")
	os.Setenv("MCP_CONFIG_PATH", configPath)
	t.Cleanup(func() {
		os.Unsetenv("MCP_CONFIG_PATH")
	})
	return NewRegistry(0)
}

func TestNewRegistry(t *testing.T) {
	registry := setupTestRegistry(t)
	assert.NotNil(t, registry)
	assert.NotNil(t, registry.servers)
	assert.NotNil(t, registry.tools)
	assert.NotNil(t, registry.healthStatus)
}

func TestNewRegistry_DefaultInterval(t *testing.T) {
	registry := setupTestRegistry(t)
	assert.Equal(t, 1*time.Minute, registry.checkInterval)
}

func TestAddServer(t *testing.T) {
	registry := setupTestRegistry(t)

	server := &Server{
		Name:    "test-server",
		URL:     "http://localhost:8080",
		Enabled: true,
	}

	err := registry.AddServer(server)
	assert.NoError(t, err)
	assert.Len(t, registry.servers, 1)
}

func TestAddServer_Duplicate(t *testing.T) {
	registry := setupTestRegistry(t)

	server := &Server{
		Name:    "test-server",
		URL:     "http://localhost:8080",
		Enabled: true,
	}

	err := registry.AddServer(server)
	require.NoError(t, err)

	err = registry.AddServer(server)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")
}

func TestRemoveServer(t *testing.T) {
	registry := setupTestRegistry(t)

	server := &Server{
		Name:    "test-server",
		URL:     "http://localhost:8080",
		Enabled: true,
	}

	err := registry.AddServer(server)
	require.NoError(t, err)

	err = registry.RemoveServer("test-server")
	assert.NoError(t, err)
	assert.Len(t, registry.servers, 0)
}

func TestRemoveServer_NotFound(t *testing.T) {
	registry := setupTestRegistry(t)

	err := registry.RemoveServer("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestRemoveServer_RemovesTools(t *testing.T) {
	registry := setupTestRegistry(t)

	server := &Server{
		Name:    "test-server",
		URL:     "http://localhost:8080",
		Enabled: true,
	}

	err := registry.AddServer(server)
	require.NoError(t, err)

	// Manually add a tool
	registry.tools["test-server.tool1"] = &ToolInfo{
		Tool:       Tool{Name: "tool1"},
		ServerName: "test-server",
	}

	err = registry.RemoveServer("test-server")
	assert.NoError(t, err)
	assert.Len(t, registry.tools, 0)
}

func TestRegistry_GetServer(t *testing.T) {
	registry := setupTestRegistry(t)

	server := &Server{
		Name:    "test-server",
		URL:     "http://localhost:8080",
		Enabled: true,
	}

	err := registry.AddServer(server)
	require.NoError(t, err)

	client, err := registry.GetServer("test-server")
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, server, client.GetServer())
}

func TestRegistry_GetServer_NotFound(t *testing.T) {
	registry := setupTestRegistry(t)

	client, err := registry.GetServer("nonexistent")
	assert.Error(t, err)
	assert.Nil(t, client)
}

func TestListServers(t *testing.T) {
	registry := setupTestRegistry(t)

	servers := []*Server{
		{Name: "server1", URL: "http://localhost:8080", Enabled: true},
		{Name: "server2", URL: "http://localhost:8081", Enabled: true},
	}

	for _, server := range servers {
		err := registry.AddServer(server)
		require.NoError(t, err)
	}

	names := registry.ListServers()
	assert.Len(t, names, 2)
	assert.Contains(t, names, "server1")
	assert.Contains(t, names, "server2")
}

func TestDiscoverTools(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		resp := JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      1,
			Result: ListToolsResult{
				Tools: []Tool{
					{Name: "tool1", Description: "Tool 1"},
					{Name: "tool2", Description: "Tool 2"},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}

	mockServer := httptest.NewServer(http.HandlerFunc(handler))
	defer mockServer.Close()

	registry := setupTestRegistry(t)
	server := &Server{
		Name:    "test-server",
		URL:     mockServer.URL,
		Enabled: true,
	}

	err := registry.AddServer(server)
	require.NoError(t, err)

	err = registry.DiscoverTools(context.Background())
	assert.NoError(t, err)
	assert.Len(t, registry.tools, 2)
	assert.Contains(t, registry.tools, "test-server.tool1")
	assert.Contains(t, registry.tools, "test-server.tool2")
}

func TestDiscoverTools_SkipsDisabledServers(t *testing.T) {
	registry := setupTestRegistry(t)

	server := &Server{
		Name:    "disabled-server",
		URL:     "http://localhost:8080",
		Enabled: false,
	}

	err := registry.AddServer(server)
	require.NoError(t, err)

	err = registry.DiscoverTools(context.Background())
	assert.NoError(t, err)
	assert.Len(t, registry.tools, 0)
}

func TestGetTool(t *testing.T) {
	registry := setupTestRegistry(t)

	toolInfo := &ToolInfo{
		Tool:       Tool{Name: "tool1", Description: "Test Tool"},
		ServerName: "test-server",
	}
	registry.tools["test-server.tool1"] = toolInfo

	result, err := registry.GetTool("test-server.tool1")
	assert.NoError(t, err)
	assert.Equal(t, toolInfo, result)
}

func TestGetTool_NotFound(t *testing.T) {
	registry := setupTestRegistry(t)

	result, err := registry.GetTool("nonexistent")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "not found")
}

func TestRegistry_ListTools(t *testing.T) {
	registry := setupTestRegistry(t)

	registry.tools["server1.tool1"] = &ToolInfo{
		Tool:       Tool{Name: "tool1"},
		ServerName: "server1",
	}
	registry.tools["server2.tool2"] = &ToolInfo{
		Tool:       Tool{Name: "tool2"},
		ServerName: "server2",
	}

	tools := registry.ListTools()
	assert.Len(t, tools, 2)
	assert.Contains(t, tools, "server1.tool1")
	assert.Contains(t, tools, "server2.tool2")
}

func TestRegistry_CallTool(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		resp := JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      1,
			Result: ToolResult{
				Content: []ResultContent{
					{Type: "text", Text: "Result"},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}

	mockServer := httptest.NewServer(http.HandlerFunc(handler))
	defer mockServer.Close()

	registry := setupTestRegistry(t)

	server := &Server{
		Name:    "test-server",
		URL:     mockServer.URL,
		Enabled: true,
	}

	err := registry.AddServer(server)
	require.NoError(t, err)

	registry.tools["test-server.tool1"] = &ToolInfo{
		Tool:       Tool{Name: "tool1"},
		ServerName: "test-server",
	}

	result, err := registry.CallTool(context.Background(), "test-server.tool1", nil)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Content, 1)
	assert.Equal(t, "Result", result.Content[0].Text)
}

func TestRegistry_CallTool_ToolNotFound(t *testing.T) {
	registry := setupTestRegistry(t)

	result, err := registry.CallTool(context.Background(), "nonexistent.tool", nil)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "not found")
}

func TestGetHealthStatus(t *testing.T) {
	registry := setupTestRegistry(t)

	status := &HealthStatus{
		Healthy:     true,
		LastChecked: time.Now(),
	}
	registry.healthStatus["test-server"] = status

	result, err := registry.GetHealthStatus("test-server")
	assert.NoError(t, err)
	assert.Equal(t, status, result)
}

func TestGetHealthStatus_NotFound(t *testing.T) {
	registry := setupTestRegistry(t)

	result, err := registry.GetHealthStatus("nonexistent")
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestGetAllHealthStatus(t *testing.T) {
	registry := setupTestRegistry(t)

	registry.healthStatus["server1"] = &HealthStatus{Healthy: true}
	registry.healthStatus["server2"] = &HealthStatus{Healthy: false}

	status := registry.GetAllHealthStatus()
	assert.Len(t, status, 2)
	assert.Contains(t, status, "server1")
	assert.Contains(t, status, "server2")
}

func TestHealthChecks(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		resp := JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      1,
			Result:  "pong",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}

	mockServer := httptest.NewServer(http.HandlerFunc(handler))
	defer mockServer.Close()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.mcp.json")
	os.Setenv("MCP_CONFIG_PATH", configPath)
	defer os.Unsetenv("MCP_CONFIG_PATH")

	registry := NewRegistry(100 * time.Millisecond)

	server := &Server{
		Name:    "test-server",
		URL:     mockServer.URL,
		Enabled: true,
	}

	err := registry.AddServer(server)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	registry.StartHealthChecks(ctx)
	defer registry.StopHealthChecks()

	// Wait for at least one health check
	time.Sleep(200 * time.Millisecond)

	status, err := registry.GetHealthStatus("test-server")
	require.NoError(t, err)
	assert.True(t, status.Healthy)
	assert.False(t, status.LastChecked.IsZero())
}

func TestHealthChecks_Stop(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.mcp.json")
	os.Setenv("MCP_CONFIG_PATH", configPath)
	defer os.Unsetenv("MCP_CONFIG_PATH")

	registry := NewRegistry(100 * time.Millisecond)

	server := &Server{
		Name:    "test-server",
		URL:     "http://localhost:9999",
		Enabled: true,
		Timeout: 1 * time.Second,
	}

	err := registry.AddServer(server)
	require.NoError(t, err)

	ctx := context.Background()
	registry.StartHealthChecks(ctx)

	// Give it time to start
	time.Sleep(50 * time.Millisecond)

	// Stop health checks
	done := make(chan struct{})
	go func() {
		registry.StopHealthChecks()
		close(done)
	}()

	select {
	case <-done:
		// Successfully stopped
	case <-time.After(2 * time.Second):
		t.Fatal("Health checks did not stop in time")
	}
}

func TestPerformHealthChecks_ConcurrentServers(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		resp := JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      1,
			Result:  "pong",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}

	mockServer := httptest.NewServer(http.HandlerFunc(handler))
	defer mockServer.Close()

	registry := setupTestRegistry(t)

	// Add multiple servers
	for i := 1; i <= 5; i++ {
		server := &Server{
			Name:    "server" + string(rune('0'+i)),
			URL:     mockServer.URL,
			Enabled: true,
		}
		err := registry.AddServer(server)
		require.NoError(t, err)
	}

	// Perform health checks
	registry.performHealthChecks(context.Background())

	// Verify all servers were checked
	allStatus := registry.GetAllHealthStatus()
	assert.Len(t, allStatus, 5)

	for _, status := range allStatus {
		assert.True(t, status.Healthy)
	}
}
