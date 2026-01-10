package cmd

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRunAddServer_RealNetworkConnectivity tests that URL validation includes real network checks
func TestRunAddServer_RealNetworkConnectivity(t *testing.T) {
	// Create a REAL HTTP server that responds to MCP protocol
	realMCPServer := createRealMCPServer(t)
	defer realMCPServer.Close()

	t.Run("Valid reachable MCP server", func(t *testing.T) {
		cmd, output := setupMCPTest(t)

		mcpServerName = "real-mcp-server"
		mcpServerURL = realMCPServer.URL // This is a REAL URL that's actually reachable
		mcpServerTimeout = 5 * time.Second

		err := runAddServer(cmd, []string{})
		assert.NoError(t, err, "Should successfully add reachable MCP server")

		out := output.String()
		assert.Contains(t, out, "Successfully added MCP server")
		assert.Contains(t, out, "Connection successful", "Should verify real connectivity")

		// Verify the server is in registry and can be reached
		servers := mcpRegistry.ListServers()
		assert.Contains(t, servers, "real-mcp-server")

		// Verify real health check
		client, err := mcpRegistry.GetServer("real-mcp-server")
		require.NoError(t, err)

		status := client.CheckHealth(context.Background())
		assert.True(t, status.Healthy, "Real server should be healthy")
		assert.NotZero(t, status.ResponseTime, "Should have real response time")
	})

	t.Run("Valid URL but unreachable server", func(t *testing.T) {
		cmd, output := setupMCPTest(t)

		// Use a valid URL format but unreachable address
		// Port 9999 on localhost is likely not in use
		mcpServerName = "unreachable-server"
		mcpServerURL = "http://localhost:19999" // Valid format but likely unreachable
		mcpServerTimeout = 1 * time.Second

		err := runAddServer(cmd, []string{})
		// No error during add - URL format is valid
		assert.NoError(t, err)

		out := output.String()
		assert.Contains(t, out, "Successfully added MCP server")
		// Connection test should fail due to unreachable server
		assert.Contains(t, out, "Connection failed", "Should detect unreachable server")
		assert.Contains(t, out, "may not be reachable", "Should warn about reachability")
	})

	t.Run("Valid URL but wrong protocol response", func(t *testing.T) {
		// Create a server that responds with non-MCP protocol
		wrongProtocolServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("This is not an MCP server"))
		}))
		defer wrongProtocolServer.Close()

		cmd, output := setupMCPTest(t)

		mcpServerName = "wrong-protocol-server"
		mcpServerURL = wrongProtocolServer.URL // Real server but wrong protocol
		mcpServerTimeout = 5 * time.Second

		err := runAddServer(cmd, []string{})
		assert.NoError(t, err, "URL is valid and reachable")

		out := output.String()
		assert.Contains(t, out, "Successfully added MCP server")
		// Should detect protocol mismatch during health check
		assert.Contains(t, out, "Connection failed", "Should detect non-MCP response")
	})
}

// createRealMCPServer creates a REAL HTTP server that implements MCP protocol
func createRealMCPServer(t *testing.T) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This is a REAL server responding to actual HTTP requests
		w.Header().Set("Content-Type", "application/json")

		// For simplicity, respond to ping requests
		// Real MCP servers would handle full JSON-RPC
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"pong"}`))
	}))
}

// TestRealServerScenarios tests various real-world server scenarios
func TestRealServerScenarios(t *testing.T) {
	t.Run("Server with timeout", func(t *testing.T) {
		// Create a server that delays response
		slowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(3 * time.Second) // Delay response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"pong"}`))
		}))
		defer slowServer.Close()

		cmd, output := setupMCPTest(t)

		mcpServerName = "slow-server"
		mcpServerURL = slowServer.URL
		mcpServerTimeout = 1 * time.Second // Timeout before server responds

		err := runAddServer(cmd, []string{})
		assert.NoError(t, err, "Server is added despite slow response")

		out := output.String()
		assert.Contains(t, out, "Successfully added MCP server")
		// Health check should timeout
		assert.Contains(t, out, "Connection failed", "Should timeout on slow server")
	})

	t.Run("Server returns HTTP error", func(t *testing.T) {
		// Create a server that returns HTTP errors
		errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
		}))
		defer errorServer.Close()

		cmd, output := setupMCPTest(t)

		mcpServerName = "error-server"
		mcpServerURL = errorServer.URL
		mcpServerTimeout = 5 * time.Second

		err := runAddServer(cmd, []string{})
		assert.NoError(t, err, "Server is added despite errors")

		out := output.String()
		assert.Contains(t, out, "Successfully added MCP server")
		assert.Contains(t, out, "Connection failed", "Should detect server errors")
	})

	t.Run("Server with TLS/HTTPS", func(t *testing.T) {
		// Create a TLS server (uses httptest's built-in TLS)
		tlsServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"pong"}`))
		}))
		defer tlsServer.Close()

		cmd, output := setupMCPTest(t)

		mcpServerName = "tls-server"
		mcpServerURL = tlsServer.URL // This will be an https:// URL
		mcpServerTimeout = 5 * time.Second

		err := runAddServer(cmd, []string{})
		assert.NoError(t, err, "Should accept HTTPS URLs")

		out := output.String()
		assert.Contains(t, out, "Successfully added MCP server")
		// Note: httptest TLS server uses self-signed cert, may fail cert validation
	})
}

// TestNetworkErrorScenarios tests various network error conditions
func TestNetworkErrorScenarios(t *testing.T) {
	testCases := []struct {
		name           string
		url            string
		expectAddError bool // Whether add-server should fail
		expectUnhealthy bool // Whether health check should fail
	}{
		{
			name:            "Non-routable IP address",
			url:             "http://192.0.2.1:3000", // TEST-NET-1, non-routable
			expectAddError:  false,
			expectUnhealthy: true,
		},
		{
			name:            "Invalid hostname",
			url:             "http://this-domain-definitely-does-not-exist-12345.com",
			expectAddError:  false,
			expectUnhealthy: true,
		},
		{
			name:            "Localhost with unused port",
			url:             "http://localhost:18273", // Random high port likely unused
			expectAddError:  false,
			expectUnhealthy: true,
		},
		{
			name:            "Localhost port 1 (requires root)",
			url:             "http://localhost:1",
			expectAddError:  false,
			expectUnhealthy: true, // Likely no service running
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd, output := setupMCPTest(t)

			mcpServerName = "test-network-error"
			mcpServerURL = tc.url
			mcpServerTimeout = 2 * time.Second

			err := runAddServer(cmd, []string{})

			if tc.expectAddError {
				require.Error(t, err, "Expected add-server to fail for: %s", tc.url)
			} else {
				assert.NoError(t, err, "URL format is valid, should be added: %s", tc.url)

				if tc.expectUnhealthy {
					out := output.String()
					assert.Contains(t, out, "Connection failed",
						"Health check should fail for unreachable server: %s", tc.url)
				}
			}
		})
	}
}

// TestURLValidationBeforeNetworkCheck ensures URL validation happens before network calls
func TestURLValidationBeforeNetworkCheck(t *testing.T) {
	// These should fail at validation stage, not network stage
	invalidURLs := []string{
		"ftp://localhost:3000",    // Invalid scheme
		"http://",                  // Missing host
		"localhost:3000",           // Missing scheme
		"http://localhost:99999",   // Invalid port
	}

	for _, invalidURL := range invalidURLs {
		t.Run(invalidURL, func(t *testing.T) {
			cmd, _ := setupMCPTest(t)

			mcpServerName = "invalid-url-test"
			mcpServerURL = invalidURL
			mcpServerTimeout = 5 * time.Second

			startTime := time.Now()
			err := runAddServer(cmd, []string{})
			elapsed := time.Since(startTime)

			require.Error(t, err, "Should fail validation for: %s", invalidURL)

			// Should fail quickly (validation, not network timeout)
			assert.Less(t, elapsed, 1*time.Second,
				"Validation should fail quickly without network calls")

			// Verify server was NOT added
			servers := mcpRegistry.ListServers()
			assert.NotContains(t, servers, "invalid-url-test")
		})
	}
}

// TestRealMCPServerDiscovery tests tool discovery on real servers
func TestRealMCPServerDiscovery(t *testing.T) {
	// Create a real MCP server with tools
	mcpServerWithTools := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Simple MCP server that responds with a test tool
		response := `{
			"jsonrpc": "2.0",
			"id": 1,
			"result": {
				"tools": [
					{
						"name": "real_tool_1",
						"description": "A real tool from a real server",
						"inputSchema": {
							"type": "object",
							"properties": {
								"param": {"type": "string"}
							}
						}
					},
					{
						"name": "real_tool_2",
						"description": "Another real tool",
						"inputSchema": {
							"type": "object"
						}
					}
				]
			}
		}`
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer mcpServerWithTools.Close()

	cmd, output := setupMCPTest(t)

	mcpServerName = "tools-server"
	mcpServerURL = mcpServerWithTools.URL
	mcpServerTimeout = 5 * time.Second

	err := runAddServer(cmd, []string{})
	assert.NoError(t, err)

	out := output.String()
	assert.Contains(t, out, "Successfully added MCP server")
	assert.Contains(t, out, "Connection successful")
	assert.Contains(t, out, "Discovered 2 tool(s)", "Should discover real tools from real server")

	// Verify tools are accessible
	tools := mcpRegistry.ListTools()
	assert.Contains(t, tools, "tools-server.real_tool_1")
	assert.Contains(t, tools, "tools-server.real_tool_2")
}

// TestConcurrentRealServerAccess tests thread safety with real network calls
func TestConcurrentRealServerAccess(t *testing.T) {
	realServer := createRealMCPServer(t)
	defer realServer.Close()

	// Add server once
	cmd, _ := setupMCPTest(t)
	mcpServerName = "concurrent-test-server"
	mcpServerURL = realServer.URL
	mcpServerTimeout = 5 * time.Second

	err := runAddServer(cmd, []string{})
	require.NoError(t, err)

	// Make concurrent real health checks
	const concurrency = 10
	done := make(chan bool, concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			client, err := mcpRegistry.GetServer("concurrent-test-server")
			if err != nil {
				done <- false
				return
			}

			status := client.CheckHealth(context.Background())
			done <- status.Healthy
		}()
	}

	// Collect results
	successCount := 0
	for i := 0; i < concurrency; i++ {
		if <-done {
			successCount++
		}
	}

	assert.Greater(t, successCount, 0, "At least some concurrent checks should succeed")
}
