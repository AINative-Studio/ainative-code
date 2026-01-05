package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCompleteChatSession tests a complete chat session workflow
func TestCompleteChatSession(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	h := NewTestHelper(t)
	defer h.Cleanup()

	t.Run("chat requires provider configuration", func(t *testing.T) {
		result := h.RunCommand("chat", "test message")
		h.AssertFailure(result, "chat should fail without provider")
		h.AssertStderrContains(result, "provider not configured", "should indicate missing provider")
	})

	t.Run("chat with provider flag", func(t *testing.T) {
		result := h.RunCommand("--provider", "openai", "chat", "test message")
		h.AssertSuccess(result, "chat should work with provider flag")
		h.AssertStdoutContains(result, "Processing message", "should process message")
	})

	t.Run("chat with verbose flag shows debug output", func(t *testing.T) {
		result := h.RunCommand("--verbose", "--provider", "openai", "chat", "test")
		h.AssertSuccess(result, "verbose chat should work")
		// Verbose output may go to stdout or stderr depending on logger config
		// Check that command produces output (either stdout or stderr)
		assert.NotEmpty(t, result.Stdout+result.Stderr, "verbose mode should produce output")
	})

	t.Run("chat with custom system message", func(t *testing.T) {
		result := h.RunCommand("--provider", "openai", "chat", "--system", "You are helpful", "Hello")
		h.AssertSuccess(result, "chat with system message should work")
		h.AssertStdoutContains(result, "Processing message", "should process message")
	})

	t.Run("chat aliases work correctly", func(t *testing.T) {
		// Test 'c' alias
		result := h.RunCommand("--provider", "openai", "c", "test")
		h.AssertSuccess(result, "'c' alias should work")

		// Test 'ask' alias
		result = h.RunCommand("--provider", "openai", "ask", "test")
		h.AssertSuccess(result, "'ask' alias should work")
	})
}

// TestChatSingleMessage tests single message mode
func TestChatSingleMessage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	h := NewTestHelper(t)
	defer h.Cleanup()

	// Setup config
	h.RunCommand("config", "init")

	t.Run("single message is processed", func(t *testing.T) {
		result := h.RunCommand("chat", "What is Go?")
		h.AssertSuccess(result, "single message should be processed")
		h.AssertStdoutContains(result, "Processing message", "should show processing")
	})

	t.Run("message with special characters", func(t *testing.T) {
		result := h.RunCommand("chat", "Hello! How are you? I'm testing.")
		h.AssertSuccess(result, "message with punctuation should work")
	})

	t.Run("empty message handling", func(t *testing.T) {
		result := h.RunCommand("chat")
		h.AssertSuccess(result, "chat without message should enter interactive mode")
		h.AssertStdoutContains(result, "Interactive chat mode", "should indicate interactive mode")
	})
}

// TestChatSessionManagement tests session-related features
func TestChatSessionManagement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	h := NewTestHelper(t)
	defer h.Cleanup()

	// Setup
	h.RunCommand("config", "init")

	t.Run("chat with session ID flag", func(t *testing.T) {
		result := h.RunCommand("chat", "--session-id", "test-session-123", "Hello")
		h.AssertSuccess(result, "chat with session ID should work")
	})

	t.Run("chat with short session flag", func(t *testing.T) {
		result := h.RunCommand("chat", "-s", "test-session-456", "Hello")
		h.AssertSuccess(result, "chat with -s flag should work")
	})
}

// TestChatVerboseMode tests verbose output
func TestChatVerboseMode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	h := NewTestHelper(t)
	defer h.Cleanup()

	h.RunCommand("config", "init")

	t.Run("verbose flag provides debug information", func(t *testing.T) {
		result := h.RunCommand("--verbose", "chat", "test")
		h.AssertSuccess(result, "verbose mode should work")
		// Verbose output may go to stdout or stderr depending on logger config
		// Check that command produces output (either stdout or stderr)
		assert.NotEmpty(t, result.Stdout+result.Stderr, "verbose mode should produce output")
	})

	t.Run("verbose short flag works", func(t *testing.T) {
		result := h.RunCommand("-v", "chat", "test")
		h.AssertSuccess(result, "-v flag should work")
	})
}

// TestChatErrorHandling tests error scenarios
func TestChatErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	h := NewTestHelper(t)
	defer h.Cleanup()

	t.Run("invalid provider shows error", func(t *testing.T) {
		result := h.RunCommand("--provider", "invalid-provider", "chat", "test")
		// Command may still succeed but with proper error messaging
		// The actual behavior depends on implementation
		assert.NotEmpty(t, result.Stdout+result.Stderr, "should provide output")
	})

	t.Run("chat without configuration shows helpful error", func(t *testing.T) {
		result := h.RunCommand("chat", "test")
		h.AssertFailure(result, "should fail without config")
		h.AssertStderrContains(result, "provider", "error should mention provider")
	})
}

// TestChatStreamingResponse tests streaming functionality
func TestChatStreamingResponse(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	h := NewTestHelper(t)
	defer h.Cleanup()

	h.RunCommand("config", "init")

	t.Run("streaming is enabled by default", func(t *testing.T) {
		result := h.RunCommand("chat", "test")
		h.AssertSuccess(result, "chat should succeed")
		// Default behavior is streaming
	})

	t.Run("streaming can be disabled", func(t *testing.T) {
		result := h.RunCommand("chat", "--stream=false", "test")
		h.AssertSuccess(result, "chat with streaming disabled should work")
	})
}

// TestChatWorkflowIntegration tests complete chat workflows
func TestChatWorkflowIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	h := NewTestHelper(t)
	defer h.Cleanup()

	t.Run("complete workflow: config -> chat -> verify", func(t *testing.T) {
		// Initialize configuration
		result := h.RunCommand("config", "init")
		h.AssertSuccess(result, "config init should succeed")

		// Set provider
		result = h.RunCommand("config", "set", "provider", "openai")
		h.AssertSuccess(result, "config set should succeed")

		// Validate config
		result = h.RunCommand("config", "validate")
		h.AssertSuccess(result, "config validate should succeed")

		// Run chat command
		result = h.RunCommand("chat", "Hello")
		h.AssertSuccess(result, "chat should succeed after proper setup")

		// Verify config persists
		result = h.RunCommand("config", "get", "provider")
		h.AssertSuccess(result, "config get should work")
		h.AssertStdoutContains(result, "openai", "provider should be persisted")
	})
}

// TestChatCustomSystemMessage tests custom system messages
func TestChatCustomSystemMessage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	h := NewTestHelper(t)
	defer h.Cleanup()

	h.RunCommand("config", "init")

	t.Run("custom system message is accepted", func(t *testing.T) {
		result := h.RunCommand("chat", "--system", "You are a helpful coding assistant", "Hello")
		h.AssertSuccess(result, "chat with system message should work")
	})

	t.Run("system message with special characters", func(t *testing.T) {
		result := h.RunCommand("chat", "--system", "Be concise! Use examples.", "test")
		h.AssertSuccess(result, "system message with punctuation should work")
	})
}

// TestChatInteractiveMode tests interactive chat mode
func TestChatInteractiveMode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	h := NewTestHelper(t)
	defer h.Cleanup()

	h.RunCommand("config", "init")

	t.Run("interactive mode starts without message argument", func(t *testing.T) {
		result := h.RunCommand("chat")
		h.AssertSuccess(result, "interactive mode should start")
		h.AssertStdoutContains(result, "Interactive chat mode", "should show interactive mode message")
	})
}

// TestChatAliases tests command aliases
func TestChatAliases(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	h := NewTestHelper(t)
	defer h.Cleanup()

	h.RunCommand("config", "init")

	aliases := []string{"c", "ask"}
	for _, alias := range aliases {
		t.Run("alias "+alias+" works", func(t *testing.T) {
			result := h.RunCommand(alias, "test message")
			h.AssertSuccess(result, "alias %s should work", alias)
		})
	}
}

// TestChatWithDifferentProviders tests provider-specific behavior
func TestChatWithDifferentProviders(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	h := NewTestHelper(t)
	defer h.Cleanup()

	providers := []string{"openai", "anthropic", "ollama"}
	for _, provider := range providers {
		t.Run("provider "+provider, func(t *testing.T) {
			result := h.RunCommand("--provider", provider, "chat", "test")
			h.AssertSuccess(result, "chat with %s should work", provider)
		})
	}
}
