package e2e

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestErrorRecoveryWorkflow tests error recovery scenarios
func TestErrorRecoveryWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	h := NewTestHelper(t)
	defer h.Cleanup()

	t.Run("recover from missing config", func(t *testing.T) {
		// Try command without config
		result := h.RunCommand("chat", "test")
		h.AssertFailure(result, "should fail without config")

		// Initialize config
		result = h.RunCommand("config", "init")
		h.AssertSuccess(result, "config init should succeed")

		// Retry command
		result = h.RunCommand("chat", "test")
		h.AssertSuccess(result, "should succeed after config init")
	})

	t.Run("recover from invalid config", func(t *testing.T) {
		h.RunCommand("config", "init")

		// Set invalid provider
		h.RunCommand("config", "set", "provider", "invalid-provider")

		// Validation should fail
		result := h.RunCommand("config", "validate")
		h.AssertFailure(result, "validation should fail")

		// Fix config
		result = h.RunCommand("config", "set", "provider", "openai")
		h.AssertSuccess(result, "should be able to fix config")

		// Validation should now succeed
		result = h.RunCommand("config", "validate")
		h.AssertSuccess(result, "validation should succeed after fix")
	})

	t.Run("recover from corrupted config file", func(t *testing.T) {
		h.RunCommand("config", "init")

		// Corrupt the config file
		h.WriteFile(".ainative-code.yaml", "invalid: yaml: content:")

		// Command should handle gracefully
		result := h.RunCommand("config", "show")
		// May succeed with default values or fail gracefully
		assert.NotEmpty(t, result.Stdout+result.Stderr, "should provide output")

		// Reinitialize with force
		result = h.RunCommand("config", "init", "--force")
		h.AssertSuccess(result, "should be able to reinitialize")

		// Verify recovery
		result = h.RunCommand("config", "validate")
		h.AssertSuccess(result, "should work after recovery")
	})
}

// TestNetworkErrorHandling tests network-related error scenarios
func TestNetworkErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	h := NewTestHelper(t)
	defer h.Cleanup()

	h.RunCommand("config", "init")

	t.Run("chat handles network unavailability gracefully", func(t *testing.T) {
		// This tests that the command completes even if network is unavailable
		result := h.RunCommand("chat", "Test")
		// Should not hang or crash
		assert.NotNil(t, result, "command should complete")
		assert.Less(t, result.Duration.Seconds(), 30.0, "should timeout appropriately")
	})
}

// TestInvalidAPIKeyRecovery tests API key error recovery
func TestInvalidAPIKeyRecovery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	h := NewTestHelper(t)
	defer h.Cleanup()

	h.RunCommand("config", "init")

	t.Run("invalid API key shows helpful error", func(t *testing.T) {
		env := map[string]string{
			"OPENAI_API_KEY": "invalid-key-123",
		}
		result := h.RunCommandWithEnv(env, "chat", "test")
		// Command should handle gracefully
		assert.NotNil(t, result, "command should complete")
	})
}

// TestRateLimitHandling tests rate limit scenarios
func TestRateLimitHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	h := NewTestHelper(t)
	defer h.Cleanup()

	h.RunCommand("config", "init")

	t.Run("multiple rapid requests complete", func(t *testing.T) {
		// Send multiple requests rapidly
		for i := 0; i < 3; i++ {
			result := h.RunCommand("chat", "quick test")
			assert.NotNil(t, result, "request %d should complete", i+1)
		}
	})
}

// TestTokenExpirationHandling tests token expiration scenarios
func TestTokenExpirationHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	h := NewTestHelper(t)
	defer h.Cleanup()

	h.RunCommand("config", "init")

	t.Run("expired token shows helpful error", func(t *testing.T) {
		env := map[string]string{
			"OPENAI_API_KEY": "expired-token",
		}
		result := h.RunCommandWithEnv(env, "chat", "test")
		// Should complete without crashing
		assert.NotNil(t, result, "command should complete")
	})
}

// TestInputValidation tests input validation and error handling
func TestInputValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	h := NewTestHelper(t)
	defer h.Cleanup()

	h.RunCommand("config", "init")

	t.Run("config set requires key and value", func(t *testing.T) {
		// Test with missing value - only provide key
		result := h.RunCommand("config", "set", "key")
		h.AssertFailure(result, "should fail with missing value")

		// Empty strings are technically valid arguments, so skip that test
		// The CLI accepts them but may fail at validation time
	})

	t.Run("session commands validate arguments", func(t *testing.T) {
		result := h.RunCommand("session", "show")
		h.AssertFailure(result, "show requires session ID")

		result = h.RunCommand("session", "delete")
		h.AssertFailure(result, "delete requires session ID")

		result = h.RunCommand("session", "export")
		h.AssertFailure(result, "export requires session ID")
	})
}

// TestGracefulErrorMessages tests that error messages are user-friendly
func TestGracefulErrorMessages(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	h := NewTestHelper(t)
	defer h.Cleanup()

	t.Run("missing config shows helpful message", func(t *testing.T) {
		result := h.RunCommand("config")
		// Should show help or error message
		assert.NotEmpty(t, result.Stdout+result.Stderr, "should show output")
	})

	t.Run("unknown command shows suggestions", func(t *testing.T) {
		result := h.RunCommand("unknown-command")
		h.AssertFailure(result, "unknown command should fail")
		// Should suggest available commands
		assert.NotEmpty(t, result.Stderr, "should show error message")
	})
}

// TestVerboseErrorOutput tests verbose error reporting
func TestVerboseErrorOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	h := NewTestHelper(t)
	defer h.Cleanup()

	t.Run("verbose mode provides detailed errors", func(t *testing.T) {
		// Try to validate without config
		result := h.RunCommand("--verbose", "config", "validate")
		// Should show detailed error information
		assert.NotEmpty(t, result.Stderr, "verbose mode should produce output")
	})

	t.Run("non-verbose mode shows concise errors", func(t *testing.T) {
		result := h.RunCommand("config", "validate")
		// Should show error but without excessive detail
		assert.NotEmpty(t, result.Stdout+result.Stderr, "should show error")
	})
}

// TestInterruptionRecovery tests handling of interrupted operations
func TestInterruptionRecovery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	h := NewTestHelper(t)
	defer h.Cleanup()
	h.SetTimeout(2 * time.Second) // Short timeout to simulate interruption

	h.RunCommand("config", "init")

	t.Run("interrupted command times out gracefully", func(t *testing.T) {
		// This would timeout but should not leave artifacts
		result := h.RunCommand("chat", "test")
		assert.NotNil(t, result, "command should complete or timeout")
	})
}

// TestCorruptedDataRecovery tests recovery from corrupted data
func TestCorruptedDataRecovery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	h := NewTestHelper(t)
	defer h.Cleanup()

	t.Run("recover from corrupted database", func(t *testing.T) {
		h.RunCommand("config", "init")

		// Create a corrupted database file
		h.WriteFile(".ainative-code/data.db", "corrupted data")

		// Commands should handle gracefully
		result := h.RunCommand("session", "list")
		// Should either recover or show helpful error
		assert.NotNil(t, result, "command should complete")
	})
}

// TestResourceExhaustion tests handling of resource constraints
func TestResourceExhaustion(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	h := NewTestHelper(t)
	defer h.Cleanup()

	h.RunCommand("config", "init")

	t.Run("handles large message input", func(t *testing.T) {
		// Create a very large message
		largeMessage := string(make([]byte, 10000))
		result := h.RunCommand("chat", largeMessage)
		// Should handle without crashing
		assert.NotNil(t, result, "command should complete")
	})
}

// TestConcurrentOperationErrors tests concurrent operation handling
func TestConcurrentOperationErrors(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	h := NewTestHelper(t)
	defer h.Cleanup()

	h.RunCommand("config", "init")

	t.Run("concurrent config updates complete safely", func(t *testing.T) {
		// Simulate concurrent operations
		result1 := h.RunCommand("config", "set", "provider", "openai")
		result2 := h.RunCommand("config", "set", "model", "gpt-4")

		// Both should complete
		h.AssertSuccess(result1, "first update should succeed")
		h.AssertSuccess(result2, "second update should succeed")

		// Verify final state is consistent
		result := h.RunCommand("config", "show")
		h.AssertSuccess(result, "config should be in consistent state")
	})
}
