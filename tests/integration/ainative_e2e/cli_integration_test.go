package ainative_e2e

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCLI_AINativeAuthLogin tests CLI authentication login
// GIVEN a mock backend
// WHEN running auth login command
// THEN command should succeed
// AND token should be saved
func TestCLI_AINativeAuthLogin(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CLI integration test in short mode")
	}

	// GIVEN a mock backend
	mockBackend := NewMockBackend(t)
	defer mockBackend.Close()

	// Setup test config
	configDir := setupTestConfig(t, mockBackend.URL)
	defer os.RemoveAll(configDir)

	// WHEN running auth login command
	cmd := exec.Command(
		getBinaryPath(),
		"auth", "login-backend",
		"--email", "test@example.com",
		"--password", "password123",
		"--config", filepath.Join(configDir, "config.yaml"),
	)

	output, err := cmd.CombinedOutput()

	// THEN command should succeed
	require.NoError(t, err, "Login command should succeed")
	assert.Contains(t, string(output), "Successfully logged in", "Output should confirm login")

	// AND token should be saved
	config := loadTestConfig(t, configDir)
	assert.NotEmpty(t, config["access_token"], "Access token should be saved")
}

// TestCLI_AINativeChatCommand tests CLI chat command
// GIVEN an authenticated user
// WHEN running chat command
// THEN chat should succeed
func TestCLI_AINativeChatCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CLI integration test in short mode")
	}

	// GIVEN an authenticated user
	mockBackend := NewMockBackend(t)
	defer mockBackend.Close()

	configDir := setupAuthenticatedTestConfig(t, mockBackend.URL)
	defer os.RemoveAll(configDir)

	// WHEN running chat command
	cmd := exec.Command(
		getBinaryPath(),
		"chat-ainative",
		"--message", "Hello",
		"--auto-provider",
		"--config", filepath.Join(configDir, "config.yaml"),
	)

	output, err := cmd.CombinedOutput()

	// THEN chat should succeed
	require.NoError(t, err, "Chat command should succeed")
	assert.Contains(t, string(output), "Assistant:", "Output should contain assistant response")
}

// TestCLI_AINativeChat_NotAuthenticated tests chat without authentication
// GIVEN no authentication
// WHEN running chat command
// THEN should fail with authentication error
func TestCLI_AINativeChat_NotAuthenticated(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CLI integration test in short mode")
	}

	// GIVEN no authentication
	configDir := setupTestConfig(t, "http://localhost:8000")
	defer os.RemoveAll(configDir)

	// Clear tokens
	clearTestTokens(t, configDir)

	// WHEN running chat command
	cmd := exec.Command(
		getBinaryPath(),
		"chat-ainative",
		"--message", "Hello",
		"--config", filepath.Join(configDir, "config.yaml"),
	)

	output, err := cmd.CombinedOutput()

	// THEN should fail with authentication error
	require.Error(t, err, "Chat should fail without authentication")
	assert.Contains(t, string(output), "not authenticated", "Output should indicate auth required")
}

// TestCLI_AINativeLogout tests logout command
// GIVEN an authenticated user
// WHEN running logout command
// THEN should clear tokens
func TestCLI_AINativeLogout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CLI integration test in short mode")
	}

	// GIVEN an authenticated user
	mockBackend := NewMockBackend(t)
	defer mockBackend.Close()

	configDir := setupAuthenticatedTestConfig(t, mockBackend.URL)
	defer os.RemoveAll(configDir)

	// WHEN running logout command
	cmd := exec.Command(
		getBinaryPath(),
		"auth", "logout",
		"--config", filepath.Join(configDir, "config.yaml"),
	)

	output, err := cmd.CombinedOutput()

	// THEN should clear tokens
	require.NoError(t, err, "Logout command should succeed")
	assert.Contains(t, string(output), "Successfully logged out", "Output should confirm logout")

	// Verify tokens are cleared
	config := loadTestConfig(t, configDir)
	assert.Empty(t, config["access_token"], "Access token should be cleared")
}

// TestCLI_AINativeProviderSelection tests provider selection
// GIVEN an authenticated user
// WHEN running chat with specific provider
// THEN should use specified provider
func TestCLI_AINativeProviderSelection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CLI integration test in short mode")
	}

	// GIVEN an authenticated user
	mockBackend := NewMockBackend(t)
	defer mockBackend.Close()

	configDir := setupAuthenticatedTestConfig(t, mockBackend.URL)
	defer os.RemoveAll(configDir)

	// WHEN running chat with specific provider
	cmd := exec.Command(
		getBinaryPath(),
		"chat-ainative",
		"--message", "Test",
		"--provider", "anthropic",
		"--config", filepath.Join(configDir, "config.yaml"),
	)

	output, err := cmd.CombinedOutput()

	// THEN should use specified provider
	require.NoError(t, err, "Chat command should succeed")
	assert.NotEmpty(t, string(output), "Should receive response")
}

// TestCLI_AINativeStreamingChat tests streaming chat command
// GIVEN an authenticated user
// WHEN running chat with streaming
// THEN should stream response
func TestCLI_AINativeStreamingChat(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CLI integration test in short mode")
	}

	// GIVEN an authenticated user
	mockBackend := NewMockBackend(t)
	mockBackend.EnableStreaming()
	defer mockBackend.Close()

	configDir := setupAuthenticatedTestConfig(t, mockBackend.URL)
	defer os.RemoveAll(configDir)

	// WHEN running chat with streaming
	cmd := exec.Command(
		getBinaryPath(),
		"chat-ainative",
		"--message", "Count to 3",
		"--stream",
		"--config", filepath.Join(configDir, "config.yaml"),
	)

	output, err := cmd.CombinedOutput()

	// THEN should stream response
	require.NoError(t, err, "Streaming chat command should succeed")
	assert.NotEmpty(t, string(output), "Should receive streamed response")
}

// TestCLI_AINativeJSONOutput tests JSON output format
// GIVEN an authenticated user
// WHEN running chat with JSON output
// THEN should output valid JSON
func TestCLI_AINativeJSONOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CLI integration test in short mode")
	}

	// GIVEN an authenticated user
	mockBackend := NewMockBackend(t)
	defer mockBackend.Close()

	configDir := setupAuthenticatedTestConfig(t, mockBackend.URL)
	defer os.RemoveAll(configDir)

	// WHEN running chat with JSON output
	cmd := exec.Command(
		getBinaryPath(),
		"chat-ainative",
		"--message", "Test",
		"--output", "json",
		"--config", filepath.Join(configDir, "config.yaml"),
	)

	output, err := cmd.CombinedOutput()

	// THEN should output valid JSON
	require.NoError(t, err, "Chat command should succeed")

	var jsonOutput map[string]interface{}
	err = json.Unmarshal(output, &jsonOutput)
	require.NoError(t, err, "Output should be valid JSON")
	assert.NotEmpty(t, jsonOutput, "JSON output should not be empty")
}

// TestCLI_AINativeHelp tests help command
// GIVEN the CLI binary
// WHEN running help command
// THEN should show help information
func TestCLI_AINativeHelp(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CLI integration test in short mode")
	}

	// WHEN running help command
	cmd := exec.Command(getBinaryPath(), "--help")
	output, err := cmd.CombinedOutput()

	// THEN should show help information
	require.NoError(t, err, "Help command should succeed")
	assert.Contains(t, string(output), "ainative-code", "Help should contain command name")
	assert.Contains(t, string(output), "Available Commands", "Help should show available commands")
}

// TestCLI_AINativeVersion tests version command
// GIVEN the CLI binary
// WHEN running version command
// THEN should show version information
func TestCLI_AINativeVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CLI integration test in short mode")
	}

	// WHEN running version command
	cmd := exec.Command(getBinaryPath(), "version")
	output, err := cmd.CombinedOutput()

	// THEN should show version information
	require.NoError(t, err, "Version command should succeed")
	assert.NotEmpty(t, string(output), "Version output should not be empty")
}

// Helper functions

// getBinaryPath returns the path to the CLI binary
func getBinaryPath() string {
	// Try to find the binary in common locations
	possiblePaths := []string{
		"./ainative-code",
		"../../../ainative-code",
		"/Users/aideveloper/AINative-Code/ainative-code",
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// Default to assuming it's in PATH
	return "ainative-code"
}

// setupTestConfig creates a temporary config directory
func setupTestConfig(t *testing.T, backendURL string) string {
	t.Helper()

	configDir, err := os.MkdirTemp("", "ainative-test-*")
	require.NoError(t, err, "Should create temp config dir")

	configPath := filepath.Join(configDir, "config.yaml")
	configContent := `
backend_url: ` + backendURL + `
log_level: info
`
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err, "Should write config file")

	return configDir
}

// setupAuthenticatedTestConfig creates a config with valid tokens
func setupAuthenticatedTestConfig(t *testing.T, backendURL string) string {
	t.Helper()

	configDir := setupTestConfig(t, backendURL)

	// Add valid token to config
	configPath := filepath.Join(configDir, "config.yaml")
	token := generateTestToken("test@example.com")

	configContent := `
backend_url: ` + backendURL + `
log_level: info
access_token: ` + token + `
refresh_token: ` + token + `
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err, "Should write authenticated config")

	return configDir
}

// loadTestConfig loads the test configuration
func loadTestConfig(t *testing.T, configDir string) map[string]string {
	t.Helper()

	configPath := filepath.Join(configDir, "config.yaml")
	content, err := os.ReadFile(configPath)
	require.NoError(t, err, "Should read config file")

	config := make(map[string]string)
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			config[key] = value
		}
	}

	return config
}

// clearTestTokens removes tokens from config
func clearTestTokens(t *testing.T, configDir string) {
	t.Helper()

	configPath := filepath.Join(configDir, "config.yaml")
	configContent := `
backend_url: http://localhost:8000
log_level: info
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err, "Should clear tokens from config")
}

// generateTestToken generates a test JWT token
func generateTestToken(email string) string {
	// Simple test token - will be replaced by proper fixture
	return "test-jwt-token-" + email
}
