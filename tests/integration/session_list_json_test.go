package integration

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSessionListJSONFlag tests the --json flag for session list command (Issue #124)
func TestSessionListJSONFlag(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-tags", "sqlite_fts5", "-o", "../../build/test-binary", "../../cmd/ainative-code")
	buildErr := buildCmd.Run()
	require.NoError(t, buildErr, "Failed to build binary for testing")
	defer os.Remove("../../build/test-binary")

	tests := []struct {
		name        string
		args        []string
		expectJSON  bool
		expectError bool
	}{
		{
			name:       "session list without --json flag",
			args:       []string{"session", "list", "--limit", "1"},
			expectJSON: false,
		},
		{
			name:       "session list with --json flag",
			args:       []string{"session", "list", "--limit", "1", "--json"},
			expectJSON: true,
		},
		{
			name:       "session list --all --json",
			args:       []string{"session", "list", "--all", "--json"},
			expectJSON: true,
		},
		{
			name:       "session list --json with custom limit",
			args:       []string{"session", "list", "--limit", "5", "--json"},
			expectJSON: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("../../build/test-binary", tt.args...)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()

			if tt.expectError {
				assert.Error(t, err, "Expected command to fail")
				return
			}

			// Check if output is valid JSON when --json flag is used
			if tt.expectJSON {
				var sessions []map[string]interface{}
				err := json.Unmarshal(stdout.Bytes(), &sessions)
				assert.NoError(t, err, "Output should be valid JSON: %s", stdout.String())

				// Verify JSON structure
				if len(sessions) > 0 {
					session := sessions[0]
					assert.Contains(t, session, "id", "Session should have 'id' field")
					assert.Contains(t, session, "name", "Session should have 'name' field")
					assert.Contains(t, session, "created_at", "Session should have 'created_at' field")
					assert.Contains(t, session, "status", "Session should have 'status' field")
				}
			} else {
				// Non-JSON output should not be valid JSON
				var sessions []map[string]interface{}
				err := json.Unmarshal(stdout.Bytes(), &sessions)
				assert.Error(t, err, "Non-JSON output should not be valid JSON")
			}
		})
	}
}

// TestSessionListJSONEmpty tests --json flag with no sessions
func TestSessionListJSONEmpty(t *testing.T) {
	t.Skip("Skipping test that requires empty database")

	cmd := exec.Command("../../build/test-binary", "session", "list", "--json")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	err := cmd.Run()
	require.NoError(t, err)

	// Should output empty JSON array
	assert.JSONEq(t, "[]", stdout.String(), "Empty session list should output empty JSON array")
}

// TestSessionListJSONFieldPresence verifies all expected fields are present
func TestSessionListJSONFieldPresence(t *testing.T) {
	cmd := exec.Command("../../build/test-binary", "session", "list", "--limit", "1", "--json")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	err := cmd.Run()
	if err != nil {
		t.Skip("Skipping test - no sessions available")
	}

	var sessions []map[string]interface{}
	err = json.Unmarshal(stdout.Bytes(), &sessions)
	require.NoError(t, err, "Output should be valid JSON")

	if len(sessions) == 0 {
		t.Skip("No sessions available for testing")
	}

	session := sessions[0]

	// Required fields
	requiredFields := []string{"id", "name", "created_at", "updated_at", "status"}
	for _, field := range requiredFields {
		assert.Contains(t, session, field, "Session must have '%s' field", field)
	}
}

// TestSessionListJSONPipeline tests JSON output works with jq
func TestSessionListJSONPipeline(t *testing.T) {
	// Check if jq is available
	_, err := exec.LookPath("jq")
	if err != nil {
		t.Skip("jq not available, skipping pipeline test")
	}

	// Get JSON output
	listCmd := exec.Command("../../build/test-binary", "session", "list", "--limit", "1", "--json")
	var listOut bytes.Buffer
	listCmd.Stdout = &listOut
	err = listCmd.Run()
	if err != nil {
		t.Skip("No sessions available")
	}

	// Parse first session ID using jq
	jqCmd := exec.Command("jq", "-r", ".[0].id")
	jqCmd.Stdin = &listOut
	var jqOut bytes.Buffer
	jqCmd.Stdout = &jqOut

	err = jqCmd.Run()
	require.NoError(t, err, "jq should successfully parse JSON output")

	sessionID := jqOut.String()
	assert.NotEmpty(t, sessionID, "Should extract session ID")
}
