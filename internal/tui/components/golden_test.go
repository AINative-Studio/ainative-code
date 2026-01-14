package components

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// CheckGolden checks the actual output against a golden file.
// If UPDATE_GOLDEN environment variable is set to "true", it will update the golden file.
// The golden file is stored in testdata/{name}.golden relative to the test file.
func CheckGolden(t *testing.T, name string, actual string) {
	t.Helper()

	// Get the directory of the calling test file
	// Since we're in components package, golden files are in ../testdata
	goldenFile := filepath.Join("..", "testdata", name+".golden")

	// Check if we should update golden files
	if os.Getenv("UPDATE_GOLDEN") == "true" {
		// Ensure testdata directory exists
		testdataDir := filepath.Dir(goldenFile)
		err := os.MkdirAll(testdataDir, 0755)
		require.NoError(t, err, "Failed to create testdata directory")

		// Write new golden file
		err = os.WriteFile(goldenFile, []byte(actual), 0644)
		require.NoError(t, err, "Failed to write golden file %s", goldenFile)
		t.Logf("Updated golden file: %s", goldenFile)
		return
	}

	// Read expected output
	expected, err := os.ReadFile(goldenFile)
	require.NoError(t, err, "Failed to read golden file %s. Run with UPDATE_GOLDEN=true to create it.", goldenFile)

	// Compare
	assert.Equal(t, string(expected), actual, "Golden test %s failed. Run with UPDATE_GOLDEN=true to update.", name)
}

// CheckGoldenInDir checks the actual output against a golden file in a specific testdata directory.
// This allows tests in different packages to organize their golden files separately.
func CheckGoldenInDir(t *testing.T, testdataDir, name string, actual string) {
	t.Helper()

	goldenFile := filepath.Join(testdataDir, name+".golden")

	// Check if we should update golden files
	if os.Getenv("UPDATE_GOLDEN") == "true" {
		// Ensure testdata directory exists
		err := os.MkdirAll(testdataDir, 0755)
		require.NoError(t, err, "Failed to create testdata directory")

		// Write new golden file
		err = os.WriteFile(goldenFile, []byte(actual), 0644)
		require.NoError(t, err, "Failed to write golden file %s", goldenFile)
		t.Logf("Updated golden file: %s", goldenFile)
		return
	}

	// Read expected output
	expected, err := os.ReadFile(goldenFile)
	require.NoError(t, err, "Failed to read golden file %s. Run with UPDATE_GOLDEN=true to create it.", goldenFile)

	// Compare
	assert.Equal(t, string(expected), actual, "Golden test %s failed. Run with UPDATE_GOLDEN=true to update.", name)
}

// NormalizeForGolden normalizes strings for golden testing by removing non-deterministic elements.
// This includes timestamps, memory addresses, and other dynamic content.
func NormalizeForGolden(s string) string {
	// For now, return as-is. Can be extended to strip timestamps, etc. if needed
	return s
}
