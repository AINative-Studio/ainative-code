package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestFirstTimeUserOnboarding tests the complete first-time user experience
func TestFirstTimeUserOnboarding(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	h := NewTestHelper(t)
	defer h.Cleanup()

	t.Run("help command shows usage information", func(t *testing.T) {
		result := h.RunCommand("--help")
		h.AssertSuccess(result, "help command should succeed")
		h.AssertStdoutContains(result, "AINative Code", "help should show app name")
		h.AssertStdoutContains(result, "Usage:", "help should show usage")
		h.AssertStdoutContains(result, "chat", "help should list chat command")
		h.AssertStdoutContains(result, "config", "help should list config command")
		h.AssertStdoutContains(result, "session", "help should list session command")
	})

	t.Run("version command shows version information", func(t *testing.T) {
		result := h.RunCommand("version")
		h.AssertSuccess(result, "version command should succeed")
		assert.NotEmpty(t, result.Stdout, "version output should not be empty")
	})

	t.Run("config init creates configuration file", func(t *testing.T) {
		result := h.RunCommand("config", "init")
		h.AssertSuccess(result, "config init should succeed")
		h.AssertStdoutContains(result, "Configuration file created", "should confirm creation")
		h.AssertStdoutContains(result, ".ainative-code.yaml", "should show config file path")

		// Verify config file was created
		assert.True(t, h.FileExists(".ainative-code.yaml"), "config file should exist")
	})

	t.Run("config init with existing file requires force flag", func(t *testing.T) {
		// Ensure config exists first (init may have been run in previous test)
		_ = h.RunCommand("config", "init")
		// First init may succeed or fail if config already exists
		// What matters is that subsequent init without --force fails

		// Second init should fail when config already exists
		result := h.RunCommand("config", "init")
		h.AssertFailure(result, "config init should fail when file already exists without --force")
		h.AssertStderrContains(result, "already exists", "should show error about existing file")
		h.AssertStderrContains(result, "--force", "should suggest --force flag")
	})

	t.Run("config init with force flag overwrites existing file", func(t *testing.T) {
		// First init
		h.RunCommand("config", "init")

		// Second init with --force
		result := h.RunCommand("config", "init", "--force")
		h.AssertSuccess(result, "config init with --force should succeed")
		h.AssertStdoutContains(result, "Configuration file created", "should confirm creation")
	})

	t.Run("config validate validates configuration", func(t *testing.T) {
		// Initialize config first
		h.RunCommand("config", "init")

		result := h.RunCommand("config", "validate")
		h.AssertSuccess(result, "config validate should succeed with valid config")
		h.AssertStdoutContains(result, "Configuration is valid", "should confirm validation")
	})

	t.Run("config show displays current settings", func(t *testing.T) {
		// Initialize config first
		h.RunCommand("config", "init")

		result := h.RunCommand("config", "show")
		h.AssertSuccess(result, "config show should succeed")
		h.AssertStdoutContains(result, "Current Configuration", "should show configuration header")
		h.AssertStdoutContains(result, "provider", "should show provider setting")
	})

	t.Run("config set updates configuration values", func(t *testing.T) {
		// Initialize config first
		h.RunCommand("config", "init")

		result := h.RunCommand("config", "set", "provider", "anthropic")
		h.AssertSuccess(result, "config set should succeed")
		h.AssertStdoutContains(result, "Set provider = anthropic", "should confirm setting")

		// Verify the value was set
		result = h.RunCommand("config", "get", "provider")
		h.AssertSuccess(result, "config get should succeed")
		h.AssertStdoutContains(result, "anthropic", "should show updated value")
	})

	t.Run("config get retrieves configuration value", func(t *testing.T) {
		// Initialize config first
		h.RunCommand("config", "init")

		result := h.RunCommand("config", "get", "provider")
		h.AssertSuccess(result, "config get should succeed")
		h.AssertStdoutContains(result, "provider", "should show the key")
		assert.NotEmpty(t, result.Stdout, "should show the value")
	})

	t.Run("chat with single message", func(t *testing.T) {
		// Initialize config first
		h.RunCommand("config", "init")

		result := h.RunCommand("chat", "Hello")
		h.AssertSuccess(result, "chat with message should succeed")
		h.AssertStdoutContains(result, "Processing message", "should indicate processing")
	})

	t.Run("complete onboarding workflow", func(t *testing.T) {
		// Step 1: Check help
		result := h.RunCommand("--help")
		h.AssertSuccess(result, "help should work")

		// Step 2: Initialize config (use --force in case it exists from previous tests)
		result = h.RunCommand("config", "init", "--force")
		h.AssertSuccess(result, "config init should work")

		// Step 3: Validate config
		result = h.RunCommand("config", "validate")
		h.AssertSuccess(result, "config validate should work")

		// Step 4: Try a chat command
		result = h.RunCommand("chat", "Hello")
		h.AssertSuccess(result, "chat should work after setup")
	})
}
