package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AINative-studio/ainative-code/internal/mcp"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupMCPTest(t *testing.T) (*cobra.Command, *bytes.Buffer) {
	t.Helper()

	// Reset global registry
	mcpRegistry = mcp.NewRegistry(1 * time.Minute)

	// Create test command with context
	cmd := &cobra.Command{
		Use: "test",
	}

	// Initialize flags for remove-server command (must match mcp.go line 115-116)
	cmd.Flags().StringP("name", "n", "", "Server name (required)")

	// Set a background context for the command to prevent nil context panics
	ctx := context.Background()
	cmd.SetContext(ctx)

	// Create output buffer
	output := new(bytes.Buffer)
	cmd.SetOut(output)
	cmd.SetErr(output)

	return cmd, output
}

func createMockMCPServer(t *testing.T) *httptest.Server {
	t.Helper()

	handler := func(w http.ResponseWriter, r *http.Request) {
		var req mcp.JSONRPCRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		resp := mcp.JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
		}

		switch req.Method {
		case "ping":
			resp.Result = "pong"

		case "tools/list":
			resp.Result = mcp.ListToolsResult{
				Tools: []mcp.Tool{
					{
						Name:        "test_tool",
						Description: "A test tool for integration testing",
						InputSchema: map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"message": map[string]interface{}{
									"type":        "string",
									"description": "Message to process",
								},
							},
						},
					},
				},
			}

		case "tools/call":
			resp.Result = mcp.ToolResult{
				Content: []mcp.ResultContent{
					{
						Type: "text",
						Text: "Tool executed successfully",
					},
				},
				IsError: false,
			}

		default:
			resp.Error = &mcp.RPCError{
				Code:    -32601,
				Message: "Method not found",
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}

	return httptest.NewServer(http.HandlerFunc(handler))
}

func TestRunAddServer(t *testing.T) {
	cmd, output := setupMCPTest(t)
	mockServer := createMockMCPServer(t)
	defer mockServer.Close()

	// Set flags
	mcpServerName = "test-server"
	mcpServerURL = mockServer.URL
	mcpServerTimeout = 5 * time.Second
	mcpServerHeaders = nil

	err := runAddServer(cmd, []string{})
	assert.NoError(t, err)

	// Verify output
	out := output.String()
	assert.Contains(t, out, "Successfully added MCP server: test-server")
	assert.Contains(t, out, "Connection successful")
	assert.Contains(t, out, "Discovered 1 tool(s)")

	// Verify server was added to registry
	servers := mcpRegistry.ListServers()
	assert.Contains(t, servers, "test-server")
}

func TestRunAddServer_InvalidURL(t *testing.T) {
	cmd, output := setupMCPTest(t)

	mcpServerName = "invalid-server"
	mcpServerURL = "http://invalid-server:99999"
	mcpServerTimeout = 1 * time.Second
	mcpServerHeaders = nil

	err := runAddServer(cmd, []string{})
	assert.NoError(t, err) // Server is added even if unreachable

	out := output.String()
	assert.Contains(t, out, "Connection failed")
	assert.Contains(t, out, "may not be reachable")
}

func TestRunRemoveServer(t *testing.T) {
	cmd, output := setupMCPTest(t)
	mockServer := createMockMCPServer(t)
	defer mockServer.Close()

	// Add a server first
	server := &mcp.Server{
		Name:    "test-server",
		URL:     mockServer.URL,
		Enabled: true,
	}
	err := mcpRegistry.AddServer(server)
	require.NoError(t, err)

	// Remove it using --name flag (consistent with add-server)
	cmd.Flags().Set("name", "test-server")
	err = runRemoveServer(cmd, []string{})
	assert.NoError(t, err)

	out := output.String()
	assert.Contains(t, out, "Successfully removed")

	// Verify it was removed
	servers := mcpRegistry.ListServers()
	assert.NotContains(t, servers, "test-server")
}

func TestRunRemoveServer_NotFound(t *testing.T) {
	cmd, _ := setupMCPTest(t)

	// Use --name flag consistently
	cmd.Flags().Set("name", "nonexistent")
	err := runRemoveServer(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestRunListServers(t *testing.T) {
	cmd, output := setupMCPTest(t)
	mockServer := createMockMCPServer(t)
	defer mockServer.Close()

	// Add a server
	server := &mcp.Server{
		Name:    "test-server",
		URL:     mockServer.URL,
		Enabled: true,
	}
	err := mcpRegistry.AddServer(server)
	require.NoError(t, err)

	// Run health check and manually set it in the registry
	ctx := context.Background()
	client, _ := mcpRegistry.GetServer("test-server")
	status := client.CheckHealth(ctx)

	// Set the health status in the registry so it's available when runListServers is called
	mcpRegistry.SetHealthStatus("test-server", status)

	err = runListServers(cmd, []string{})
	assert.NoError(t, err)

	out := output.String()
	assert.Contains(t, out, "test-server")
	assert.Contains(t, out, mockServer.URL)
	assert.Contains(t, out, "OK")
}

func TestRunListServers_Empty(t *testing.T) {
	cmd, output := setupMCPTest(t)

	err := runListServers(cmd, []string{})
	assert.NoError(t, err)

	out := output.String()
	assert.Contains(t, out, "No MCP servers registered")
	assert.Contains(t, out, "mcp add-server")
}

func TestRunListTools(t *testing.T) {
	cmd, output := setupMCPTest(t)
	mockServer := createMockMCPServer(t)
	defer mockServer.Close()

	// Add a server
	server := &mcp.Server{
		Name:    "test-server",
		URL:     mockServer.URL,
		Enabled: true,
	}
	err := mcpRegistry.AddServer(server)
	require.NoError(t, err)

	err = runListTools(cmd, []string{})
	assert.NoError(t, err)

	out := output.String()
	assert.Contains(t, out, "test-server.test_tool")
	assert.Contains(t, out, "A test tool")
	assert.Contains(t, out, "Total: 1 tool(s)")
}

func TestRunListTools_Empty(t *testing.T) {
	cmd, output := setupMCPTest(t)

	err := runListTools(cmd, []string{})
	assert.NoError(t, err)

	out := output.String()
	assert.Contains(t, out, "No tools available")
	assert.Contains(t, out, "mcp add-server")
}

func TestRunTestTool(t *testing.T) {
	cmd, output := setupMCPTest(t)
	mockServer := createMockMCPServer(t)
	defer mockServer.Close()

	// Add a server
	server := &mcp.Server{
		Name:    "test-server",
		URL:     mockServer.URL,
		Enabled: true,
	}
	err := mcpRegistry.AddServer(server)
	require.NoError(t, err)

	// Discover tools
	err = mcpRegistry.DiscoverTools(context.Background())
	require.NoError(t, err)

	// Test the tool
	mcpServerHeaders = map[string]string{
		"message": "\"Hello, World!\"",
	}

	err = runTestTool(cmd, []string{"test-server.test_tool"})
	assert.NoError(t, err)

	out := output.String()
	assert.Contains(t, out, "Testing tool: test-server.test_tool")
	assert.Contains(t, out, "Result: SUCCESS")
	assert.Contains(t, out, "Tool executed successfully")
}

func TestRunTestTool_NotFound(t *testing.T) {
	cmd, _ := setupMCPTest(t)

	err := runTestTool(cmd, []string{"nonexistent.tool"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tool not found")
}

func TestRunDiscover(t *testing.T) {
	cmd, output := setupMCPTest(t)
	mockServer := createMockMCPServer(t)
	defer mockServer.Close()

	// Add a server
	server := &mcp.Server{
		Name:    "test-server",
		URL:     mockServer.URL,
		Enabled: true,
	}
	err := mcpRegistry.AddServer(server)
	require.NoError(t, err)

	err = runDiscover(cmd, []string{})
	assert.NoError(t, err)

	out := output.String()
	assert.Contains(t, out, "Discovered 1 tool(s) from 1 server(s)")
	assert.Contains(t, out, "mcp list-tools")
}

func TestRunDiscover_NoServers(t *testing.T) {
	cmd, output := setupMCPTest(t)

	err := runDiscover(cmd, []string{})
	assert.NoError(t, err)

	out := output.String()
	assert.Contains(t, out, "Discovered 0 tool(s) from 0 server(s)")
}

func TestGetMCPRegistry(t *testing.T) {
	// Reset registry
	mcpRegistry = mcp.NewRegistry(1 * time.Minute)

	registry := GetMCPRegistry()
	assert.NotNil(t, registry)
	assert.Equal(t, mcpRegistry, registry)
}

func TestMCPCommands_Integration(t *testing.T) {
	cmd, output := setupMCPTest(t)
	mockServer := createMockMCPServer(t)
	defer mockServer.Close()

	// Step 1: Add server
	mcpServerName = "integration-server"
	mcpServerURL = mockServer.URL
	mcpServerTimeout = 5 * time.Second
	mcpServerHeaders = nil

	err := runAddServer(cmd, []string{})
	assert.NoError(t, err)
	output.Reset()

	// Step 2: List servers
	err = runListServers(cmd, []string{})
	assert.NoError(t, err)
	assert.Contains(t, output.String(), "integration-server")
	output.Reset()

	// Step 3: List tools
	err = runListTools(cmd, []string{})
	assert.NoError(t, err)
	assert.Contains(t, output.String(), "integration-server.test_tool")
	output.Reset()

	// Step 4: Test tool
	mcpServerHeaders = map[string]string{"message": "\"test\""}
	err = runTestTool(cmd, []string{"integration-server.test_tool"})
	assert.NoError(t, err)
	assert.Contains(t, output.String(), "SUCCESS")
	output.Reset()

	// Step 5: Remove server using --name flag
	cmd.Flags().Set("name", "integration-server")
	err = runRemoveServer(cmd, []string{})
	assert.NoError(t, err)
	assert.Contains(t, output.String(), "Successfully removed")

	// Step 6: Verify removal
	servers := mcpRegistry.ListServers()
	assert.NotContains(t, servers, "integration-server")
}
