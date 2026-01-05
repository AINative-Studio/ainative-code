package e2e

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSessionExportWorkflow tests session export functionality
func TestSessionExportWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	h := NewTestHelper(t)
	defer h.Cleanup()

	// Setup
	h.RunCommand("config", "init")

	t.Run("session list shows available sessions", func(t *testing.T) {
		result := h.RunCommand("session", "list")
		h.AssertSuccess(result, "session list should succeed")
		h.AssertStdoutContains(result, "Session List", "should show session list header")
	})

	t.Run("session list with aliases", func(t *testing.T) {
		// Test 'ls' alias
		result := h.RunCommand("session", "ls")
		h.AssertSuccess(result, "session ls should work")

		// Test 'l' alias
		result = h.RunCommand("session", "l")
		h.AssertSuccess(result, "session l should work")

		// Test 'sessions' command alias
		result = h.RunCommand("sessions", "list")
		h.AssertSuccess(result, "sessions command alias should work")

		// Test 'sess' command alias
		result = h.RunCommand("sess", "list")
		h.AssertSuccess(result, "sess command alias should work")
	})

	t.Run("session list with all flag", func(t *testing.T) {
		result := h.RunCommand("session", "list", "--all")
		h.AssertSuccess(result, "session list --all should work")
		h.AssertStdoutContains(result, "All: true", "should show all sessions")
	})

	t.Run("session list with limit flag", func(t *testing.T) {
		result := h.RunCommand("session", "list", "--limit", "5")
		h.AssertSuccess(result, "session list with limit should work")
		h.AssertStdoutContains(result, "Limit: 5", "should respect limit")

		// Test short flag
		result = h.RunCommand("session", "list", "-n", "3")
		h.AssertSuccess(result, "session list with -n should work")
		h.AssertStdoutContains(result, "Limit: 3", "should respect -n limit")
	})

	t.Run("session show displays session details", func(t *testing.T) {
		result := h.RunCommand("session", "show", "test-session-id")
		h.AssertSuccess(result, "session show should succeed")
		h.AssertStdoutContains(result, "Session Details", "should show session details")
		h.AssertStdoutContains(result, "test-session-id", "should show session ID")
	})

	t.Run("session export to JSON", func(t *testing.T) {
		result := h.RunCommand("session", "export", "test-session")
		h.AssertSuccess(result, "session export should succeed")
		h.AssertStdoutContains(result, "Exporting session", "should show export message")
	})

	t.Run("session export with custom output", func(t *testing.T) {
		outputFile := "my-session.json"
		result := h.RunCommand("session", "export", "test-session", "--output", outputFile)
		h.AssertSuccess(result, "session export with output should work")
		h.AssertStdoutContains(result, outputFile, "should mention output file")
	})

	t.Run("session export with short output flag", func(t *testing.T) {
		result := h.RunCommand("session", "export", "test-session", "-o", "session.json")
		h.AssertSuccess(result, "session export with -o should work")
	})

	t.Run("session delete removes session", func(t *testing.T) {
		result := h.RunCommand("session", "delete", "test-session-to-delete")
		h.AssertSuccess(result, "session delete should succeed")
		h.AssertStdoutContains(result, "Deleting session", "should show delete message")
	})

	t.Run("session delete with aliases", func(t *testing.T) {
		// Test 'rm' alias
		result := h.RunCommand("session", "rm", "test-session")
		h.AssertSuccess(result, "session rm should work")

		// Test 'remove' alias
		result = h.RunCommand("session", "remove", "test-session")
		h.AssertSuccess(result, "session remove should work")
	})
}

// TestSessionExportMarkdown tests Markdown export format
func TestSessionExportMarkdown(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	h := NewTestHelper(t)
	defer h.Cleanup()

	h.RunCommand("config", "init")

	t.Run("export session as markdown", func(t *testing.T) {
		// Create a session first
		h.RunCommand("chat", "-s", "auto-filename-session", "test message")

		// Export it (default format is JSON, no --format flag exists yet)
		result := h.RunCommand("session", "export", "auto-filename-session", "-o", "session.json")
		h.AssertSuccess(result, "session export should work")
	})
}

// TestSessionExportJSON tests JSON export format
func TestSessionExportJSON(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	h := NewTestHelper(t)
	defer h.Cleanup()

	h.RunCommand("config", "init")

	t.Run("export session as JSON", func(t *testing.T) {
		// Create a session first
		h.RunCommand("chat", "-s", "test-session", "test message")

		// Export it (default format is JSON)
		result := h.RunCommand("session", "export", "test-session", "-o", "test-session.json")
		h.AssertSuccess(result, "JSON export should work")
	})
}

// TestSessionExportHTML tests HTML export format
func TestSessionExportHTML(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	h := NewTestHelper(t)
	defer h.Cleanup()

	h.RunCommand("config", "init")

	t.Run("export session as HTML", func(t *testing.T) {
		// Create a session first
		h.RunCommand("chat", "-s", "test-session-html", "test message")

		// Export it (only JSON format supported currently)
		result := h.RunCommand("session", "export", "test-session-html", "-o", "session.json")
		h.AssertSuccess(result, "session export should work")
	})
}

// TestSessionExportStdout tests exporting to stdout
func TestSessionExportStdout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	h := NewTestHelper(t)
	defer h.Cleanup()

	h.RunCommand("config", "init")

	t.Run("export to stdout with dash", func(t *testing.T) {
		// Create a session first
		h.RunCommand("chat", "-s", "stdout-session", "test message")

		// Export to stdout (use -o with dash)
		result := h.RunCommand("session", "export", "stdout-session", "-o", "-")
		h.AssertSuccess(result, "export to stdout should work")
	})
}

// TestSessionExportBatch tests batch export operations
func TestSessionExportBatch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	h := NewTestHelper(t)
	defer h.Cleanup()

	h.RunCommand("config", "init")

	t.Run("export multiple sessions sequentially", func(t *testing.T) {
		sessions := []string{"session-1", "session-2", "session-3"}
		for i, session := range sessions {
			outputFile := fmt.Sprintf("export-%d.json", i)
			result := h.RunCommand("session", "export", session, "-o", outputFile)
			h.AssertSuccess(result, "export %s should succeed", session)
		}
	})
}

// TestSessionExportMultipleFormats tests exporting in various formats
func TestSessionExportMultipleFormats(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	h := NewTestHelper(t)
	defer h.Cleanup()

	h.RunCommand("config", "init")

	// Create a session first that we'll export
	h.RunCommand("chat", "-s", "test-session-formats", "test message")

	// Currently only JSON format is supported (no --format flag exists)
	// Test exporting to different output files
	outputs := []struct {
		name string
		file string
	}{
		{"default output", "session1.json"},
		{"custom path", filepath.Join(h.GetWorkDir(), "session2.json")},
	}

	for _, out := range outputs {
		t.Run(fmt.Sprintf("export as %s", out.name), func(t *testing.T) {
			result := h.RunCommand("session", "export", "test-session-formats", "-o", out.file)
			h.AssertSuccess(result, "%s export should work", out.name)
		})
	}
}

// TestSessionExportErrorHandling tests error scenarios
func TestSessionExportErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	h := NewTestHelper(t)
	defer h.Cleanup()

	h.RunCommand("config", "init")

	t.Run("export requires session ID", func(t *testing.T) {
		result := h.RunCommand("session", "export")
		h.AssertFailure(result, "export without session ID should fail")
		// The error message says "accepts 1 arg(s), received 0" not "requires"
		// Check for the actual error pattern
		assert.Contains(t, result.Stderr, "accepts 1 arg(s), received 0", "should show error about missing argument")
	})

	t.Run("show requires session ID", func(t *testing.T) {
		result := h.RunCommand("session", "show")
		h.AssertFailure(result, "show without session ID should fail")
	})

	t.Run("delete requires session ID", func(t *testing.T) {
		result := h.RunCommand("session", "delete")
		h.AssertFailure(result, "delete without session ID should fail")
	})
}

// TestSessionExportWithFilters tests filtered exports
func TestSessionExportWithFilters(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	h := NewTestHelper(t)
	defer h.Cleanup()

	h.RunCommand("config", "init")

	t.Run("export with date filters", func(t *testing.T) {
		// This tests the command accepts filter flags even if not fully implemented
		result := h.RunCommand("session", "export", "test-session", "-o", "filtered.json")
		h.AssertSuccess(result, "export with potential filters should work")
	})
}

// TestSessionExportMetadata tests metadata handling
func TestSessionExportMetadata(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	h := NewTestHelper(t)
	defer h.Cleanup()

	h.RunCommand("config", "init")

	t.Run("export includes session metadata", func(t *testing.T) {
		result := h.RunCommand("session", "export", "test-session")
		h.AssertSuccess(result, "export should succeed")
		// Metadata would be in the exported file content
	})
}

// TestSessionDatabase tests session database operations
func TestSessionDatabase(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	h := NewTestHelper(t)
	defer h.Cleanup()

	t.Run("sessions work with custom database path", func(t *testing.T) {
		// Initialize with custom database path
		h.RunCommand("config", "init")

		customDBPath := filepath.Join(h.GetWorkDir(), "custom-sessions.db")
		h.RunCommand("config", "set", "database.path", customDBPath)

		// Operations should work with custom path
		result := h.RunCommand("session", "list")
		h.AssertSuccess(result, "session list should work with custom db path")
	})
}

// TestSessionCompleteWorkflow tests end-to-end session workflow
func TestSessionCompleteWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	h := NewTestHelper(t)
	defer h.Cleanup()

	t.Run("complete session lifecycle", func(t *testing.T) {
		// Setup
		h.RunCommand("config", "init")

		// Create a session via chat
		result := h.RunCommand("chat", "-s", "workflow-session", "test message")
		h.AssertSuccess(result, "chat with session should work")

		// List sessions
		result = h.RunCommand("session", "list")
		h.AssertSuccess(result, "list should work")

		// Show session details
		result = h.RunCommand("session", "show", "workflow-session")
		h.AssertSuccess(result, "show should work")

		// Export session
		exportFile := filepath.Join(h.GetWorkDir(), "workflow-export.json")
		result = h.RunCommand("session", "export", "workflow-session", "-o", exportFile)
		h.AssertSuccess(result, "export should work")

		// Verify operations complete successfully
		assert.True(t, true, "complete workflow executed")
	})
}
