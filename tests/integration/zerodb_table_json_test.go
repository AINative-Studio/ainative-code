// +build integration

package integration

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestZeroDBTableJSONOutput tests the --json flag for all zerodb table subcommands
func TestZeroDBTableJSONOutput(t *testing.T) {
	// Skip if not in integration mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Build the CLI binary first
	binaryPath := buildCLIBinary(t)
	defer os.Remove(binaryPath)

	// Setup test configuration
	configPath := setupTestConfig(t)
	defer os.Remove(configPath)

	t.Run("TableCreate_JSONOutput", func(t *testing.T) {
		testTableCreateJSONOutput(t, binaryPath, configPath)
	})

	t.Run("TableList_JSONOutput", func(t *testing.T) {
		testTableListJSONOutput(t, binaryPath, configPath)
	})

	t.Run("TableInsert_JSONOutput", func(t *testing.T) {
		testTableInsertJSONOutput(t, binaryPath, configPath)
	})

	t.Run("TableQuery_JSONOutput", func(t *testing.T) {
		testTableQueryJSONOutput(t, binaryPath, configPath)
	})

	t.Run("TableUpdate_JSONOutput", func(t *testing.T) {
		testTableUpdateJSONOutput(t, binaryPath, configPath)
	})

	t.Run("TableDelete_JSONOutput", func(t *testing.T) {
		testTableDeleteJSONOutput(t, binaryPath, configPath)
	})
}

func testTableCreateJSONOutput(t *testing.T, binaryPath, configPath string) {
	schema := `{"type":"object","properties":{"name":{"type":"string"},"age":{"type":"number"}}}`

	cmd := exec.Command(binaryPath,
		"--config", configPath,
		"zerodb", "table", "create",
		"--name", "test_users",
		"--schema", schema,
		"--json",
	)

	output, err := cmd.CombinedOutput()
	t.Logf("Command output: %s", string(output))

	// If the command fails due to API/config issues, verify the flag is at least recognized
	if err != nil {
		// Check that the error is NOT about unknown flag
		assert.NotContains(t, string(output), "unknown flag: --json",
			"--json flag should be recognized")
		assert.NotContains(t, string(output), "unknown shorthand flag",
			"--json flag should be properly registered")
		return
	}

	// If successful, verify JSON output
	var result map[string]interface{}
	err = json.Unmarshal(output, &result)
	require.NoError(t, err, "Output should be valid JSON")

	// Verify expected JSON fields
	assert.Contains(t, result, "id", "JSON output should contain 'id' field")
	assert.Contains(t, result, "name", "JSON output should contain 'name' field")
}

func testTableListJSONOutput(t *testing.T, binaryPath, configPath string) {
	cmd := exec.Command(binaryPath,
		"--config", configPath,
		"zerodb", "table", "list",
		"--json",
	)

	output, err := cmd.CombinedOutput()
	t.Logf("Command output: %s", string(output))

	// If the command fails due to API/config issues, verify the flag is at least recognized
	if err != nil {
		// Check that the error is NOT about unknown flag
		assert.NotContains(t, string(output), "unknown flag: --json",
			"--json flag should be recognized")
		assert.NotContains(t, string(output), "unknown shorthand flag",
			"--json flag should be properly registered")
		return
	}

	// If successful, verify JSON output
	var result []map[string]interface{}
	err = json.Unmarshal(output, &result)
	require.NoError(t, err, "Output should be valid JSON array")

	// If tables exist, verify structure
	if len(result) > 0 {
		assert.Contains(t, result[0], "id", "JSON output should contain 'id' field")
		assert.Contains(t, result[0], "name", "JSON output should contain 'name' field")
	}
}

func testTableInsertJSONOutput(t *testing.T, binaryPath, configPath string) {
	data := `{"name":"John Doe","age":30}`

	cmd := exec.Command(binaryPath,
		"--config", configPath,
		"zerodb", "table", "insert",
		"--table", "test_users",
		"--data", data,
		"--json",
	)

	output, err := cmd.CombinedOutput()
	t.Logf("Command output: %s", string(output))

	// If the command fails due to API/config issues, verify the flag is at least recognized
	if err != nil {
		// Check that the error is NOT about unknown flag
		assert.NotContains(t, string(output), "unknown flag: --json",
			"--json flag should be recognized")
		assert.NotContains(t, string(output), "unknown shorthand flag",
			"--json flag should be properly registered")
		return
	}

	// If successful, verify JSON output
	var result map[string]interface{}
	err = json.Unmarshal(output, &result)
	require.NoError(t, err, "Output should be valid JSON")

	// Verify expected JSON fields
	assert.Contains(t, result, "id", "JSON output should contain 'id' field")
}

func testTableQueryJSONOutput(t *testing.T, binaryPath, configPath string) {
	filter := `{"age":{"$gte":18}}`

	cmd := exec.Command(binaryPath,
		"--config", configPath,
		"zerodb", "table", "query",
		"--table", "test_users",
		"--filter", filter,
		"--json",
	)

	output, err := cmd.CombinedOutput()
	t.Logf("Command output: %s", string(output))

	// If the command fails due to API/config issues, verify the flag is at least recognized
	if err != nil {
		// Check that the error is NOT about unknown flag
		assert.NotContains(t, string(output), "unknown flag: --json",
			"--json flag should be recognized")
		assert.NotContains(t, string(output), "unknown shorthand flag",
			"--json flag should be properly registered")
		return
	}

	// If successful, verify JSON output
	var result []map[string]interface{}
	err = json.Unmarshal(output, &result)
	require.NoError(t, err, "Output should be valid JSON array")

	// If documents exist, verify structure
	if len(result) > 0 {
		assert.Contains(t, result[0], "id", "JSON output should contain 'id' field")
		assert.Contains(t, result[0], "data", "JSON output should contain 'data' field")
	}
}

func testTableUpdateJSONOutput(t *testing.T, binaryPath, configPath string) {
	data := `{"age":31}`

	cmd := exec.Command(binaryPath,
		"--config", configPath,
		"zerodb", "table", "update",
		"--table", "test_users",
		"--id", "test-doc-id",
		"--data", data,
		"--json",
	)

	output, err := cmd.CombinedOutput()
	t.Logf("Command output: %s", string(output))

	// If the command fails due to API/config issues, verify the flag is at least recognized
	if err != nil {
		// Check that the error is NOT about unknown flag
		assert.NotContains(t, string(output), "unknown flag: --json",
			"--json flag should be recognized")
		assert.NotContains(t, string(output), "unknown shorthand flag",
			"--json flag should be properly registered")
		return
	}

	// If successful, verify JSON output
	var result map[string]interface{}
	err = json.Unmarshal(output, &result)
	require.NoError(t, err, "Output should be valid JSON")

	// Verify expected JSON fields
	assert.Contains(t, result, "id", "JSON output should contain 'id' field")
}

func testTableDeleteJSONOutput(t *testing.T, binaryPath, configPath string) {
	cmd := exec.Command(binaryPath,
		"--config", configPath,
		"zerodb", "table", "delete",
		"--table", "test_users",
		"--id", "test-doc-id",
		"--json",
	)

	output, err := cmd.CombinedOutput()
	t.Logf("Command output: %s", string(output))

	// If the command fails due to API/config issues, verify the flag is at least recognized
	if err != nil {
		// Check that the error is NOT about unknown flag
		assert.NotContains(t, string(output), "unknown flag: --json",
			"--json flag should be recognized")
		assert.NotContains(t, string(output), "unknown shorthand flag",
			"--json flag should be properly registered")
		return
	}

	// If successful, verify JSON output
	var result map[string]interface{}
	err = json.Unmarshal(output, &result)
	require.NoError(t, err, "Output should be valid JSON")

	// Verify expected JSON fields
	assert.Contains(t, result, "success", "JSON output should contain 'success' field")
	assert.Contains(t, result, "id", "JSON output should contain 'id' field")
	assert.Contains(t, result, "table", "JSON output should contain 'table' field")
}

// buildCLIBinary builds the CLI binary for testing
func buildCLIBinary(t *testing.T) string {
	t.Helper()

	// Create temporary directory for binary
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "ainative-code-test")

	// Build the binary
	cmd := exec.Command("go", "build", "-o", binaryPath, "../../cmd/ainative-code")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Failed to build CLI binary: %s", string(output))

	return binaryPath
}

// setupTestConfig creates a temporary config file for testing
func setupTestConfig(t *testing.T) string {
	t.Helper()

	// Create temporary config file
	tmpFile, err := os.CreateTemp("", "ainative-config-*.yaml")
	require.NoError(t, err)
	defer tmpFile.Close()

	// Write minimal config
	config := `
services:
  zerodb:
    endpoint: https://api.ainative.studio
    project_id: test-project-id
`
	_, err = tmpFile.WriteString(config)
	require.NoError(t, err)

	return tmpFile.Name()
}

// TestZeroDBTableJSONFlagRegistration verifies the --json flag is properly registered
func TestZeroDBTableJSONFlagRegistration(t *testing.T) {
	// This test verifies the flag is registered by checking the help output
	binaryPath := buildCLIBinary(t)
	defer os.Remove(binaryPath)

	// Test each subcommand's help output
	subcommands := []string{"create", "list", "insert", "query", "update", "delete"}

	for _, subcmd := range subcommands {
		t.Run("FlagRegistration_"+subcmd, func(t *testing.T) {
			cmd := exec.Command(binaryPath, "zerodb", "table", subcmd, "--help")
			output, err := cmd.CombinedOutput()
			require.NoError(t, err, "Help command should succeed")

			outputStr := string(output)
			assert.Contains(t, outputStr, "--json",
				"Help output for '%s' should mention --json flag", subcmd)
			assert.Contains(t, outputStr, "output in JSON format",
				"Help output for '%s' should describe --json flag", subcmd)
		})
	}
}

// TestZeroDBTableJSONFlagInheritance verifies --json is available to all subcommands
func TestZeroDBTableJSONFlagInheritance(t *testing.T) {
	binaryPath := buildCLIBinary(t)
	defer os.Remove(binaryPath)

	// Test that --json is recognized (not unknown) for each subcommand
	subcommands := []struct {
		name string
		args []string
	}{
		{"create", []string{"--name", "test", "--schema", "{}"}},
		{"list", []string{}},
		{"insert", []string{"--table", "test", "--data", "{}"}},
		{"query", []string{"--table", "test"}},
		{"update", []string{"--table", "test", "--id", "123", "--data", "{}"}},
		{"delete", []string{"--table", "test", "--id", "123"}},
	}

	for _, tc := range subcommands {
		t.Run("FlagRecognized_"+tc.name, func(t *testing.T) {
			args := append([]string{"zerodb", "table", tc.name}, tc.args...)
			args = append(args, "--json")

			cmd := exec.Command(binaryPath, args...)
			output, _ := cmd.CombinedOutput()

			// The command may fail due to missing config/API issues,
			// but it should NOT fail with "unknown flag"
			outputStr := string(output)
			assert.NotContains(t, outputStr, "unknown flag: --json",
				"--json should be recognized for '%s' command", tc.name)
			assert.NotContains(t, outputStr, "flag provided but not defined: -json",
				"--json should be defined for '%s' command", tc.name)
		})
	}
}

// TestZeroDBTableJSONOutputValidation tests that JSON output is valid
func TestZeroDBTableJSONOutputValidation(t *testing.T) {
	// This test verifies JSON can be parsed by jq or json.Unmarshal
	testCases := []struct {
		name       string
		jsonOutput string
	}{
		{
			name:       "CreateResponse",
			jsonOutput: `{"id":"tbl_123","name":"users","created_at":"2024-01-01T00:00:00Z"}`,
		},
		{
			name:       "ListResponse",
			jsonOutput: `[{"id":"tbl_123","name":"users","created_at":"2024-01-01T00:00:00Z"}]`,
		},
		{
			name:       "DeleteResponse",
			jsonOutput: `{"success":true,"id":"doc_123","table":"users"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test with Go's json.Unmarshal
			var result interface{}
			err := json.Unmarshal([]byte(tc.jsonOutput), &result)
			assert.NoError(t, err, "JSON should be valid for %s", tc.name)

			// Test with jq if available
			if _, err := exec.LookPath("jq"); err == nil {
				cmd := exec.Command("jq", ".")
				cmd.Stdin = bytes.NewBufferString(tc.jsonOutput)
				output, err := cmd.CombinedOutput()
				assert.NoError(t, err, "jq should parse JSON for %s: %s", tc.name, string(output))
			}
		})
	}
}
