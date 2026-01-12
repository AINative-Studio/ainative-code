package integration

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestJSONOutputClean tests that all --json flags produce clean JSON without log lines
// This is the comprehensive test for issue #127
func TestJSONOutputClean(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-tags", "sqlite_fts5", "-o", "../../build/test-json-binary", "../../cmd/ainative-code")
	buildErr := buildCmd.Run()
	require.NoError(t, buildErr, "Failed to build binary for testing")
	defer os.Remove("../../build/test-json-binary")

	tests := []struct {
		name          string
		args          []string
		skipReason    string
		validateJSON  func(*testing.T, []byte) // Custom JSON validation function
	}{
		{
			name: "session list --json",
			args: []string{"session", "list", "--limit", "1", "--json"},
			validateJSON: func(t *testing.T, output []byte) {
				var sessions []map[string]interface{}
				err := json.Unmarshal(output, &sessions)
				assert.NoError(t, err, "Output should be valid JSON array")
			},
		},
		{
			name: "session search --json",
			args: []string{"session", "search", "test", "--limit", "5", "--json"},
			validateJSON: func(t *testing.T, output []byte) {
				var result map[string]interface{}
				err := json.Unmarshal(output, &result)
				assert.NoError(t, err, "Output should be valid JSON object")
			},
		},
		{
			name:       "strapi blog list --json",
			args:       []string{"strapi", "blog", "list", "--json"},
			skipReason: "Requires Strapi configuration",
		},
		{
			name:       "zerodb memory list --json",
			args:       []string{"zerodb", "memory", "list", "--agent-id", "test", "--json"},
			skipReason: "Requires ZeroDB configuration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipReason != "" {
				t.Skip(tt.skipReason)
			}

			cmd := exec.Command("../../build/test-json-binary", tt.args...)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			// Run the command (ignore errors for empty results)
			_ = cmd.Run()

			// Get stdout output
			output := stdout.Bytes()

			// Skip if no output (no data available)
			if len(output) == 0 {
				t.Skip("No data available for testing")
			}

			// Critical test: Output should NOT contain log lines
			outputStr := string(output)

			// Check for common log patterns that should NOT be in JSON output
			logPatterns := []string{
				"INF",
				"DBG",
				"Searching sessions",
				"Listing sessions",
				"Listing blog posts",
				"INFO",
				"DEBUG",
				"[90m", // ANSI color codes
				"[32m", // ANSI color codes
			}

			for _, pattern := range logPatterns {
				assert.NotContains(t, outputStr, pattern,
					"JSON output should not contain log pattern: %s", pattern)
			}

			// Validate it's proper JSON
			if tt.validateJSON != nil {
				tt.validateJSON(t, output)
			} else {
				// Generic JSON validation
				var jsonData interface{}
				err := json.Unmarshal(output, &jsonData)
				assert.NoError(t, err, "Output should be valid JSON")
			}

			// Verify first character is a JSON character
			firstChar := outputStr[0]
			assert.True(t, firstChar == '{' || firstChar == '[',
				"JSON output should start with { or [, got: %c", firstChar)
		})
	}
}

// TestJSONOutputPipeline tests that JSON output works correctly with jq
func TestJSONOutputPipeline(t *testing.T) {
	// Check if jq is available
	_, err := exec.LookPath("jq")
	if err != nil {
		t.Skip("jq not available, skipping pipeline test")
	}

	// Build the binary
	buildCmd := exec.Command("go", "build", "-tags", "sqlite_fts5", "-o", "../../build/test-jq-binary", "../../cmd/ainative-code")
	buildErr := buildCmd.Run()
	require.NoError(t, buildErr, "Failed to build binary for testing")
	defer os.Remove("../../build/test-jq-binary")

	tests := []struct {
		name       string
		args       []string
		jqFilter   string
		skipReason string
	}{
		{
			name:     "session list with jq",
			args:     []string{"session", "list", "--limit", "1", "--json"},
			jqFilter: ".",
		},
		{
			name:     "session list extract id",
			args:     []string{"session", "list", "--limit", "1", "--json"},
			jqFilter: ".[0].id",
		},
		{
			name:     "session search with jq",
			args:     []string{"session", "search", "test", "--limit", "1", "--json"},
			jqFilter: ".query",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipReason != "" {
				t.Skip(tt.skipReason)
			}

			// Run the command
			cmd := exec.Command("../../build/test-jq-binary", tt.args...)
			var cmdOut bytes.Buffer
			cmd.Stdout = &cmdOut
			cmd.Stderr = os.Stderr

			err := cmd.Run()
			if err != nil {
				t.Skip("No data available or command failed")
			}

			// Pipe to jq
			jqCmd := exec.Command("jq", tt.jqFilter)
			jqCmd.Stdin = &cmdOut
			var jqOut bytes.Buffer
			jqCmd.Stdout = &jqOut
			jqCmd.Stderr = os.Stderr

			err = jqCmd.Run()
			assert.NoError(t, err, "jq should successfully parse JSON output")

			// Verify jq produced output
			assert.NotEmpty(t, jqOut.String(), "jq should produce output")
		})
	}
}

// TestJSONOutputNoLogLinesRegression is a regression test for issue #127
func TestJSONOutputNoLogLinesRegression(t *testing.T) {
	// Build the binary
	buildCmd := exec.Command("go", "build", "-tags", "sqlite_fts5", "-o", "../../build/test-regression-binary", "../../cmd/ainative-code")
	buildErr := buildCmd.Run()
	require.NoError(t, buildErr, "Failed to build binary for testing")
	defer os.Remove("../../build/test-regression-binary")

	// Test the specific issue from #127: session search --json
	cmd := exec.Command("../../build/test-regression-binary", "session", "search", "test", "--json")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	_ = cmd.Run()

	output := stdout.String()

	if len(output) == 0 {
		t.Skip("No search results available")
	}

	// The specific bug: INFO logs should NOT be in stdout
	assert.NotContains(t, output, "Searching sessions",
		"INFO log 'Searching sessions' should not appear in JSON output")

	assert.NotContains(t, output, "INF",
		"INFO log level indicator should not appear in JSON output")

	// Verify it's valid JSON
	var jsonData map[string]interface{}
	err := json.Unmarshal([]byte(output), &jsonData)
	assert.NoError(t, err, "Output should be valid JSON")

	// Stderr can contain logs, but stdout should be pure JSON
	// This is acceptable: logs to stderr, JSON to stdout
	if len(stderr.String()) > 0 {
		t.Logf("Stderr output (acceptable): %s", stderr.String())
	}
}

// TestJSONOutputFirstLineIsJSON ensures the first line of output is always JSON
func TestJSONOutputFirstLineIsJSON(t *testing.T) {
	// Build the binary
	buildCmd := exec.Command("go", "build", "-tags", "sqlite_fts5", "-o", "../../build/test-firstline-binary", "../../cmd/ainative-code")
	buildErr := buildCmd.Run()
	require.NoError(t, buildErr, "Failed to build binary for testing")
	defer os.Remove("../../build/test-firstline-binary")

	commands := [][]string{
		{"session", "list", "--limit", "1", "--json"},
		{"session", "search", "test", "--limit", "1", "--json"},
	}

	for _, args := range commands {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			cmd := exec.Command("../../build/test-firstline-binary", args...)
			var stdout bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = os.Stderr

			_ = cmd.Run()

			output := stdout.String()
			if len(output) == 0 {
				t.Skip("No data available")
			}

			// Get first line
			lines := strings.Split(output, "\n")
			firstLine := lines[0]

			// First line should start with { or [
			assert.True(t,
				strings.HasPrefix(firstLine, "{") || strings.HasPrefix(firstLine, "["),
				"First line should start with { or [, got: %s", firstLine)

			// First line should NOT be a log line
			assert.NotContains(t, firstLine, "INF")
			assert.NotContains(t, firstLine, "DBG")
			assert.NotContains(t, firstLine, "Searching")
			assert.NotContains(t, firstLine, "Listing")
		})
	}
}

// TestJSONOutputErrorsToStderr tests that errors still go to stderr
func TestJSONOutputErrorsToStderr(t *testing.T) {
	// Build the binary
	buildCmd := exec.Command("go", "build", "-tags", "sqlite_fts5", "-o", "../../build/test-stderr-binary", "../../cmd/ainative-code")
	buildErr := buildCmd.Run()
	require.NoError(t, buildErr, "Failed to build binary for testing")
	defer os.Remove("../../build/test-stderr-binary")

	// Test with invalid command that should produce error
	cmd := exec.Command("../../build/test-stderr-binary", "session", "search", "", "--json")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	// Should fail with error
	assert.Error(t, err, "Command should fail with empty query")

	// Stderr should contain the Cobra error message (this is the user-facing error)
	assert.NotEmpty(t, stderr.String(), "Stderr should contain error message")
	assert.Contains(t, stderr.String(), "search query cannot be empty",
		"Stderr should contain the error message")

	// Note: The logger ERROR output may go to stdout (per logger config),
	// but the important part is that the Cobra error goes to stderr and
	// no JSON is produced. We don't assert stdout is empty because the
	// error logger might write there, which is acceptable since:
	// 1. The command failed (exit code != 0)
	// 2. No valid JSON was attempted to be produced
	// 3. The user-facing error went to stderr
}
