package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValidateMCPServerURL_RealNetworkValidation tests URL validation with real network scenarios
func TestValidateMCPServerURL_RealNetworkValidation(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		expectError bool
		errorMsg    string
	}{
		// Invalid URL format tests
		{
			name:        "Empty URL",
			url:         "",
			expectError: true,
			errorMsg:    "invalid URL format",
		},
		{
			name:        "Invalid URL - no scheme",
			url:         "localhost:3000",
			expectError: true,
			errorMsg:    "invalid URL scheme", // ParseRequestURI accepts this but scheme is "localhost"
		},
		{
			name:        "Invalid URL - just text",
			url:         "notaurl",
			expectError: true,
			errorMsg:    "invalid URL format",
		},
		{
			name:        "Invalid URL - malformed",
			url:         "http://",
			expectError: true,
			errorMsg:    "missing host",
		},

		// Invalid scheme tests
		{
			name:        "Invalid scheme - ftp",
			url:         "ftp://localhost:3000",
			expectError: true,
			errorMsg:    "invalid URL scheme",
		},
		{
			name:        "Invalid scheme - ws",
			url:         "ws://localhost:3000",
			expectError: true,
			errorMsg:    "invalid URL scheme",
		},
		{
			name:        "Invalid scheme - file",
			url:         "file:///path/to/file",
			expectError: true,
			errorMsg:    "invalid URL scheme",
		},
		{
			name:        "Invalid scheme - custom",
			url:         "custom://server:3000",
			expectError: true,
			errorMsg:    "invalid URL scheme",
		},

		// Missing host tests
		{
			name:        "Missing host - http only",
			url:         "http://",
			expectError: true,
			errorMsg:    "missing host",
		},
		{
			name:        "Missing host - https only",
			url:         "https://",
			expectError: true,
			errorMsg:    "missing host",
		},

		// Invalid port tests
		{
			name:        "Invalid port - too large",
			url:         "http://localhost:99999",
			expectError: true,
			errorMsg:    "invalid port number",
		},
		{
			name:        "Invalid port - zero",
			url:         "http://localhost:0",
			expectError: true,
			errorMsg:    "invalid port number",
		},
		{
			name:        "Invalid port - negative",
			url:         "http://localhost:-1",
			expectError: true,
			errorMsg:    "invalid URL format", // This fails at parsing stage
		},
		{
			name:        "Invalid port - non-numeric",
			url:         "http://localhost:abc",
			expectError: true,
			errorMsg:    "invalid URL format", // This fails at parsing stage
		},

		// Valid URL format tests (syntactically correct)
		{
			name:        "Valid HTTP localhost with port",
			url:         "http://localhost:3000",
			expectError: false,
		},
		{
			name:        "Valid HTTPS localhost with port",
			url:         "https://localhost:8080",
			expectError: false,
		},
		{
			name:        "Valid HTTP with IP address",
			url:         "http://127.0.0.1:3000",
			expectError: false,
		},
		{
			name:        "Valid HTTPS with domain",
			url:         "https://api.example.com",
			expectError: false,
		},
		{
			name:        "Valid HTTPS with subdomain and path",
			url:         "https://mcp.api.example.com/v1",
			expectError: false,
		},
		{
			name:        "Valid HTTP without port (default 80)",
			url:         "http://localhost",
			expectError: false,
		},
		{
			name:        "Valid HTTPS without port (default 443)",
			url:         "https://example.com",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMCPServerURL(tt.url)

			if tt.expectError {
				require.Error(t, err, "Expected error for URL: %s", tt.url)
				assert.Contains(t, err.Error(), tt.errorMsg,
					"Error message should contain '%s' for URL: %s", tt.errorMsg, tt.url)
			} else {
				assert.NoError(t, err, "Expected no error for valid URL: %s", tt.url)
			}
		})
	}
}

// TestRunAddServer_RealInvalidURLs tests add-server command with real invalid URLs
func TestRunAddServer_RealInvalidURLs(t *testing.T) {
	invalidURLTests := []struct {
		name     string
		url      string
		errorMsg string
	}{
		{
			name:     "No scheme provided",
			url:      "localhost:3000",
			errorMsg: "invalid URL scheme", // ParseRequestURI accepts this but scheme is "localhost"
		},
		{
			name:     "FTP scheme not allowed",
			url:      "ftp://localhost:3000",
			errorMsg: "invalid URL scheme",
		},
		{
			name:     "Missing host",
			url:      "http://",
			errorMsg: "missing host",
		},
		{
			name:     "Invalid port number",
			url:      "http://localhost:99999",
			errorMsg: "invalid port number",
		},
		{
			name:     "WebSocket scheme not allowed",
			url:      "ws://localhost:3000",
			errorMsg: "invalid URL scheme",
		},
		{
			name:     "Empty string",
			url:      "",
			errorMsg: "invalid URL format",
		},
		{
			name:     "Just text without structure",
			url:      "not-a-valid-url",
			errorMsg: "invalid URL format",
		},
	}

	for _, tt := range invalidURLTests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, _ := setupMCPTest(t)

			mcpServerName = "test-invalid-server"
			mcpServerURL = tt.url

			err := runAddServer(cmd, []string{})

			require.Error(t, err, "Expected error when adding server with invalid URL: %s", tt.url)
			assert.Contains(t, err.Error(), tt.errorMsg,
				"Error should contain '%s' for URL: %s", tt.errorMsg, tt.url)

			// Verify server was NOT added to registry
			servers := mcpRegistry.ListServers()
			assert.NotContains(t, servers, "test-invalid-server",
				"Invalid server should not be added to registry")
		})
	}
}

// TestRunAddServer_RealValidURLFormat tests add-server with valid URL formats
// Note: These URLs are syntactically valid but may not be reachable
func TestRunAddServer_RealValidURLFormat(t *testing.T) {
	validURLTests := []struct {
		name string
		url  string
	}{
		{
			name: "HTTP localhost with port",
			url:  "http://localhost:3000",
		},
		{
			name: "HTTPS localhost with port",
			url:  "https://localhost:8080",
		},
		{
			name: "HTTP with IP address",
			url:  "http://127.0.0.1:3000",
		},
		{
			name: "HTTPS with domain",
			url:  "https://api.example.com",
		},
	}

	for _, tt := range validURLTests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, output := setupMCPTest(t)

			serverName := "test-valid-" + tt.name
			mcpServerName = serverName
			mcpServerURL = tt.url

			// The URL validation should pass (format is correct)
			err := runAddServer(cmd, []string{})

			// No error means URL format was accepted
			assert.NoError(t, err, "Valid URL format should be accepted: %s", tt.url)

			// Verify server was added
			servers := mcpRegistry.ListServers()
			assert.Contains(t, servers, serverName,
				"Server with valid URL should be added to registry")

			// Check output for connection test results
			out := output.String()
			assert.Contains(t, out, "Successfully added MCP server",
				"Should show success message for valid URL format")
			assert.Contains(t, out, "Testing connection",
				"Should attempt to test connection after adding")
		})
	}
}

// TestURLValidation_EdgeCases tests edge cases in URL validation
func TestURLValidation_EdgeCases(t *testing.T) {
	edgeCases := []struct {
		name        string
		url         string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "URL with path",
			url:         "http://localhost:3000/api/v1",
			expectError: false,
		},
		{
			name:        "URL with query parameters",
			url:         "http://localhost:3000?param=value",
			expectError: false,
		},
		{
			name:        "URL with fragment",
			url:         "http://localhost:3000#section",
			expectError: true, // ParseRequestURI fails on fragments
			errorMsg:    "invalid URL format",
		},
		{
			name:        "URL with username and password",
			url:         "http://user:pass@localhost:3000",
			expectError: false,
		},
		{
			name:        "URL with IPv6",
			url:         "http://[::1]:3000",
			expectError: false,
		},
		{
			name:        "Max valid port",
			url:         "http://localhost:65535",
			expectError: false,
		},
		{
			name:        "Min valid port",
			url:         "http://localhost:1",
			expectError: false,
		},
		{
			name:        "Port just above max",
			url:         "http://localhost:65536",
			expectError: true,
			errorMsg:    "invalid port number",
		},
		{
			name:        "Malformed scheme separator",
			url:         "http:/localhost:3000",
			expectError: true,
			errorMsg:    "missing host", // ParseRequestURI parses this but host is empty
		},
		{
			name:        "Double slashes in path",
			url:         "http://localhost:3000//api",
			expectError: false,
		},
	}

	for _, tt := range edgeCases {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMCPServerURL(tt.url)

			if tt.expectError {
				require.Error(t, err, "Expected error for edge case: %s", tt.name)
				assert.Contains(t, err.Error(), tt.errorMsg,
					"Error should contain '%s' for URL: %s", tt.errorMsg, tt.url)
			} else {
				assert.NoError(t, err, "Expected no error for valid edge case: %s", tt.name)
			}
		})
	}
}

// TestURLValidation_HelpfulErrorMessages verifies that error messages guide users
func TestURLValidation_HelpfulErrorMessages(t *testing.T) {
	testCases := []struct {
		name         string
		url          string
		shouldContain []string
	}{
		{
			name: "Invalid format shows example",
			url:  "notaurl",
			shouldContain: []string{
				"invalid URL format",
				"http://localhost:3000",
				"https://api.example.com",
			},
		},
		{
			name: "Invalid scheme shows allowed schemes",
			url:  "ftp://server:3000",
			shouldContain: []string{
				"invalid URL scheme",
				"http",
				"https",
			},
		},
		{
			name: "Missing host shows examples",
			url:  "http://",
			shouldContain: []string{
				"missing host",
				"localhost:3000",
				"api.example.com",
			},
		},
		{
			name: "Invalid port shows valid range",
			url:  "http://localhost:99999",
			shouldContain: []string{
				"invalid port number",
				"1 and 65535",
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMCPServerURL(tt.url)
			require.Error(t, err, "Should return error for: %s", tt.url)

			errMsg := err.Error()
			for _, substring := range tt.shouldContain {
				assert.Contains(t, errMsg, substring,
					"Error message should contain helpful guidance: '%s'", substring)
			}
		})
	}
}
