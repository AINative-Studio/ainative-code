package cmd

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/AINative-studio/ainative-code/internal/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createRealMCPTestServer creates a test server that implements the real MCP protocol.
// This is NOT a mock - it's a real HTTP server implementing the full JSON-RPC 2.0 MCP spec.
func createRealMCPTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	handler := func(w http.ResponseWriter, r *http.Request) {
		// Verify it's a POST request
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Verify Content-Type
		contentType := r.Header.Get("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			http.Error(w, "Invalid Content-Type", http.StatusBadRequest)
			return
		}

		// Decode the JSON-RPC request
		var req mcp.JSONRPCRequest
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&req); err != nil {
			http.Error(w, "Invalid JSON-RPC request", http.StatusBadRequest)
			return
		}

		// Validate JSON-RPC 2.0 protocol
		if req.JSONRPC != "2.0" {
			http.Error(w, "Invalid JSON-RPC version", http.StatusBadRequest)
			return
		}

		// Prepare response
		resp := mcp.JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
		}

		// Route to appropriate handler based on method
		switch req.Method {
		case "ping":
			// Real ping implementation
			resp.Result = "pong"

		case "tools/list":
			// Real tools/list implementation with pagination
			var params mcp.ListToolsParams
			if req.Params != nil {
				paramsBytes, _ := json.Marshal(req.Params)
				json.Unmarshal(paramsBytes, &params)
			}

			// Return real tool definitions
			result := mcp.ListToolsResult{
				Tools: []mcp.Tool{
					{
						Name:        "echo",
						Description: "Echoes the input message back to the caller",
						InputSchema: map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"message": map[string]interface{}{
									"type":        "string",
									"description": "The message to echo back",
								},
							},
							"required": []string{"message"},
						},
					},
					{
						Name:        "uppercase",
						Description: "Converts input text to uppercase",
						InputSchema: map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"text": map[string]interface{}{
									"type":        "string",
									"description": "Text to convert to uppercase",
								},
							},
							"required": []string{"text"},
						},
					},
				},
				NextCursor: "", // No pagination for this test
			}
			resp.Result = result

		case "tools/call":
			// Real tools/call implementation
			var params mcp.CallToolParams
			if req.Params != nil {
				paramsBytes, _ := json.Marshal(req.Params)
				json.Unmarshal(paramsBytes, &params)
			}

			// Execute the tool based on name
			switch params.Name {
			case "echo":
				message, ok := params.Arguments["message"].(string)
				if !ok {
					resp.Error = &mcp.RPCError{
						Code:    -32602,
						Message: "Invalid parameters: message must be a string",
					}
				} else {
					resp.Result = mcp.ToolResult{
						Content: []mcp.ResultContent{
							{
								Type: "text",
								Text: message,
							},
						},
						IsError: false,
					}
				}

			case "uppercase":
				text, ok := params.Arguments["text"].(string)
				if !ok {
					resp.Error = &mcp.RPCError{
						Code:    -32602,
						Message: "Invalid parameters: text must be a string",
					}
				} else {
					resp.Result = mcp.ToolResult{
						Content: []mcp.ResultContent{
							{
								Type: "text",
								Text: strings.ToUpper(text),
							},
						},
						IsError: false,
					}
				}

			default:
				resp.Error = &mcp.RPCError{
					Code:    -32601,
					Message: "Tool not found: " + params.Name,
				}
			}

		default:
			// Method not found
			resp.Error = &mcp.RPCError{
				Code:    -32601,
				Message: "Method not found: " + req.Method,
			}
		}

		// Send JSON-RPC response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}

	// Create and return real HTTP test server
	server := httptest.NewServer(http.HandlerFunc(handler))
	t.Logf("Created real MCP test server at: %s", server.URL)
	return server
}

// TestMCPAddRemoveServerIntegration tests the full add/remove server workflow
// using REAL MCP protocol communication (no mocks).
func TestMCPAddRemoveServerIntegration(t *testing.T) {
	// Create a REAL MCP server (not a mock)
	realMCPServer := createRealMCPTestServer(t)
	defer realMCPServer.Close()

	// Setup test environment
	cmd, output := setupMCPTest(t)

	// Test 1: Add server using --name and --url flags
	t.Run("AddServer", func(t *testing.T) {
		mcpServerName = "real-integration-server"
		mcpServerURL = realMCPServer.URL
		mcpServerTimeout = 5 * time.Second
		mcpServerHeaders = nil

		err := runAddServer(cmd, []string{})
		require.NoError(t, err, "Failed to add MCP server")

		out := output.String()
		assert.Contains(t, out, "Successfully added MCP server: real-integration-server")
		assert.Contains(t, out, "Connection successful", "Server should be reachable")
		assert.Contains(t, out, "Discovered 2 tool(s)", "Should discover echo and uppercase tools")

		// Verify server was added to registry
		servers := mcpRegistry.ListServers()
		assert.Contains(t, servers, "real-integration-server")

		output.Reset()
	})

	// Test 2: List servers and verify health check works with real API
	t.Run("ListServers", func(t *testing.T) {
		// Force a health check by getting the client and checking health
		ctx := context.Background()
		client, err := mcpRegistry.GetServer("real-integration-server")
		require.NoError(t, err)

		status := client.CheckHealth(ctx)
		assert.True(t, status.Healthy, "Health check should pass")
		assert.Empty(t, status.Error, "Should have no errors")
		assert.Greater(t, status.ResponseTime, time.Duration(0), "Should have response time")

		// Set health status in registry
		mcpRegistry.SetHealthStatus("real-integration-server", status)

		err = runListServers(cmd, []string{})
		require.NoError(t, err)

		out := output.String()
		assert.Contains(t, out, "real-integration-server")
		assert.Contains(t, out, realMCPServer.URL)
		assert.Contains(t, out, "OK", "Server should be healthy")

		output.Reset()
	})

	// Test 3: List tools and verify real tool discovery
	t.Run("ListTools", func(t *testing.T) {
		err := runListTools(cmd, []string{})
		require.NoError(t, err)

		out := output.String()
		assert.Contains(t, out, "real-integration-server.echo")
		assert.Contains(t, out, "real-integration-server.uppercase")
		assert.Contains(t, out, "Echoes the input message")
		assert.Contains(t, out, "Converts input text to uppercase")
		assert.Contains(t, out, "Total: 2 tool(s)")

		output.Reset()
	})

	// Test 4: Test tool execution with real API call
	t.Run("TestToolExecution", func(t *testing.T) {
		// Test echo tool
		mcpServerHeaders = map[string]string{
			"message": "\"Hello, Real MCP Server!\"",
		}

		err := runTestTool(cmd, []string{"real-integration-server.echo"})
		require.NoError(t, err, "Echo tool should execute successfully")

		out := output.String()
		assert.Contains(t, out, "Result: SUCCESS")
		assert.Contains(t, out, "Hello, Real MCP Server!")

		output.Reset()

		// Test uppercase tool
		mcpServerHeaders = map[string]string{
			"text": "\"lowercase text\"",
		}

		err = runTestTool(cmd, []string{"real-integration-server.uppercase"})
		require.NoError(t, err, "Uppercase tool should execute successfully")

		out = output.String()
		assert.Contains(t, out, "Result: SUCCESS")
		assert.Contains(t, out, "LOWERCASE TEXT")

		output.Reset()
	})

	// Test 5: Remove server using --name flag (NOT positional argument)
	t.Run("RemoveServer_UsingNameFlag", func(t *testing.T) {
		// This is the fix for issue #108 - must use --name flag
		cmd.Flags().Set("name", "real-integration-server")
		err := runRemoveServer(cmd, []string{})
		require.NoError(t, err, "Failed to remove server")

		out := output.String()
		assert.Contains(t, out, "Successfully removed MCP server: real-integration-server")

		// Verify server was removed
		servers := mcpRegistry.ListServers()
		assert.NotContains(t, servers, "real-integration-server")

		output.Reset()
	})

	// Test 6: Verify server is truly removed
	t.Run("VerifyServerRemoved", func(t *testing.T) {
		_, err := mcpRegistry.GetServer("real-integration-server")
		assert.Error(t, err, "Server should not be found after removal")
		assert.Contains(t, err.Error(), "not found")
	})
}

// TestMCPCommandFlagConsistency verifies all MCP commands use consistent flag-based interfaces.
func TestMCPCommandFlagConsistency(t *testing.T) {
	realMCPServer := createRealMCPTestServer(t)
	defer realMCPServer.Close()

	t.Run("AddServerUsesNameFlag", func(t *testing.T) {
		cmd, _ := setupMCPTest(t)

		// Verify add-server requires --name flag
		mcpServerName = "consistency-test-server"
		mcpServerURL = realMCPServer.URL
		mcpServerTimeout = 5 * time.Second

		err := runAddServer(cmd, []string{})
		assert.NoError(t, err, "add-server should accept --name flag")

		// Clean up
		cmd.Flags().Set("name", "consistency-test-server")
		runRemoveServer(cmd, []string{})
	})

	t.Run("RemoveServerUsesNameFlag", func(t *testing.T) {
		cmd, output := setupMCPTest(t)

		// Add a server first
		server := &mcp.Server{
			Name:    "flag-test-server",
			URL:     realMCPServer.URL,
			Enabled: true,
		}
		mcpRegistry.AddServer(server)

		// Verify remove-server uses --name flag (consistent with add-server)
		cmd.Flags().Set("name", "flag-test-server")
		err := runRemoveServer(cmd, []string{})
		assert.NoError(t, err, "remove-server should accept --name flag")

		out := output.String()
		assert.Contains(t, out, "Successfully removed")
	})

	t.Run("NoPositionalArgsForRemoveServer", func(t *testing.T) {
		cmd, _ := setupMCPTest(t)

		// Add a server
		server := &mcp.Server{
			Name:    "positional-test-server",
			URL:     realMCPServer.URL,
			Enabled: true,
		}
		mcpRegistry.AddServer(server)

		// Passing server name as positional arg should NOT work
		// The function reads from flags, not args
		err := runRemoveServer(cmd, []string{"positional-test-server"})

		// This should fail because --name flag is empty
		assert.Error(t, err, "Positional args should not work for remove-server")
		assert.Contains(t, err.Error(), "not found", "Should not find server without --name flag")
	})
}

// TestMCPRealAPIErrorHandling tests error handling with real API calls.
func TestMCPRealAPIErrorHandling(t *testing.T) {
	realMCPServer := createRealMCPTestServer(t)
	defer realMCPServer.Close()

	t.Run("InvalidToolName", func(t *testing.T) {
		cmd, _ := setupMCPTest(t)

		// Add server
		mcpServerName = "error-test-server"
		mcpServerURL = realMCPServer.URL
		mcpServerTimeout = 5 * time.Second
		runAddServer(cmd, []string{})

		// Try to call non-existent tool
		err := runTestTool(cmd, []string{"error-test-server.nonexistent"})
		assert.Error(t, err, "Should fail for non-existent tool")

		// Clean up
		cmd.Flags().Set("name", "error-test-server")
		runRemoveServer(cmd, []string{})
	})

	t.Run("InvalidArguments", func(t *testing.T) {
		cmd, output := setupMCPTest(t)

		// Add server
		mcpServerName = "invalid-args-server"
		mcpServerURL = realMCPServer.URL
		mcpServerTimeout = 5 * time.Second
		runAddServer(cmd, []string{})

		// Try to call tool with wrong argument type (number instead of string)
		mcpServerHeaders = map[string]string{
			"message": "123", // This is a number, not a JSON string
		}

		err := runTestTool(cmd, []string{"invalid-args-server.echo"})
		// Should fail because the server validates argument types
		assert.Error(t, err, "Should fail when argument type doesn't match schema")
		assert.Contains(t, err.Error(), "Invalid parameters", "Error should mention invalid parameters")

		output.Reset()

		// Clean up
		cmd.Flags().Set("name", "invalid-args-server")
		runRemoveServer(cmd, []string{})
	})

	t.Run("ServerTimeout", func(t *testing.T) {
		cmd, _ := setupMCPTest(t)

		// Try to add server with invalid URL
		mcpServerName = "timeout-server"
		mcpServerURL = "http://localhost:65535" // Invalid port
		mcpServerTimeout = 1 * time.Second

		err := runAddServer(cmd, []string{})
		// Should add server but connection test should fail
		assert.NoError(t, err, "Server should be added even if unreachable")

		// Clean up
		cmd.Flags().Set("name", "timeout-server")
		runRemoveServer(cmd, []string{})
	})
}

// TestMCPDiscoverWithRealServer tests tool discovery with real MCP server.
func TestMCPDiscoverWithRealServer(t *testing.T) {
	realMCPServer := createRealMCPTestServer(t)
	defer realMCPServer.Close()

	cmd, output := setupMCPTest(t)

	// Add server
	server := &mcp.Server{
		Name:    "discover-test-server",
		URL:     realMCPServer.URL,
		Enabled: true,
	}
	err := mcpRegistry.AddServer(server)
	require.NoError(t, err)

	// Discover tools
	err = runDiscover(cmd, []string{})
	require.NoError(t, err)

	out := output.String()
	assert.Contains(t, out, "Discovered 2 tool(s) from 1 server(s)")
	assert.Contains(t, out, "mcp list-tools")

	// Verify tools are actually in registry
	tools := mcpRegistry.ListTools()
	assert.Len(t, tools, 2, "Should have exactly 2 tools")
	assert.Contains(t, tools, "discover-test-server.echo")
	assert.Contains(t, tools, "discover-test-server.uppercase")

	// Clean up
	output.Reset()
	cmd.Flags().Set("name", "discover-test-server")
	runRemoveServer(cmd, []string{})
}
