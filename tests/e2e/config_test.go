package e2e

import (
	"testing"
)

// TestConfigEnvironmentVariables tests environment variable handling
func TestConfigEnvironmentVariables(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	h := NewTestHelper(t)
	defer h.Cleanup()

	t.Run("environment variables override config file", func(t *testing.T) {
		// Set config file values
		h.RunCommand("config", "init")
		h.RunCommand("config", "set", "provider", "openai")

		// Override with environment variable
		env := map[string]string{
			"AINATIVE_PROVIDER": "anthropic",
		}
		result := h.RunCommandWithEnv(env, "chat", "test")
		h.AssertSuccess(result, "env var should override config")
	})

	t.Run("multiple environment variables work together", func(t *testing.T) {
		env := map[string]string{
			"AINATIVE_PROVIDER": "openai",
			"AINATIVE_MODEL":    "gpt-4",
			"AINATIVE_VERBOSE":  "true",
		}
		result := h.RunCommandWithEnv(env, "chat", "test")
		h.AssertSuccess(result, "multiple env vars should work")
	})
}

// TestConfigFileFormats tests different config file formats
func TestConfigFileFormats(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	h := NewTestHelper(t)
	defer h.Cleanup()

	t.Run("YAML config file works", func(t *testing.T) {
		yamlConfig := `provider: openai
model: gpt-4
verbose: false`
		h.WriteFile(".ainative-code.yaml", yamlConfig)

		result := h.RunCommand("config", "show")
		h.AssertSuccess(result, "YAML config should work")
		h.AssertStdoutContains(result, "openai", "should show openai")
	})

	t.Run("config with nested values", func(t *testing.T) {
		yamlConfig := `provider: openai
model: gpt-4
database:
  path: /tmp/test.db
  auto_save: true`
		h.WriteFile(".ainative-code.yaml", yamlConfig)

		result := h.RunCommand("config", "show")
		h.AssertSuccess(result, "nested config should work")
	})
}

// TestVersionCommand tests version command
func TestVersionCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	h := NewTestHelper(t)
	defer h.Cleanup()

	t.Run("version command shows version info", func(t *testing.T) {
		result := h.RunCommand("version")
		h.AssertSuccess(result, "version command should succeed")
	})

	t.Run("version flag shows version", func(t *testing.T) {
		result := h.RunCommand("--version")
		h.AssertSuccess(result, "--version flag should work")
	})
}

// TestHelpCommand tests help command
func TestHelpCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	h := NewTestHelper(t)
	defer h.Cleanup()

	t.Run("help command shows usage", func(t *testing.T) {
		result := h.RunCommand("help")
		h.AssertSuccess(result, "help command should succeed")
		h.AssertStdoutContains(result, "Usage:", "should show usage")
	})

	t.Run("help flag shows usage", func(t *testing.T) {
		result := h.RunCommand("--help")
		h.AssertSuccess(result, "--help flag should work")
		h.AssertStdoutContains(result, "Usage:", "should show usage")
	})

	t.Run("help for specific command", func(t *testing.T) {
		result := h.RunCommand("chat", "--help")
		h.AssertSuccess(result, "chat help should work")
		h.AssertStdoutContains(result, "chat", "should show chat help")
	})

	t.Run("help for config command", func(t *testing.T) {
		result := h.RunCommand("config", "--help")
		h.AssertSuccess(result, "config help should work")
		h.AssertStdoutContains(result, "config", "should show config help")
	})

	t.Run("help for session command", func(t *testing.T) {
		result := h.RunCommand("session", "--help")
		h.AssertSuccess(result, "session help should work")
		h.AssertStdoutContains(result, "session", "should show session help")
	})
}
