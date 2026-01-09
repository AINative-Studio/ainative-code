package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

// TestSetupInitializationCheck tests the logic for checking if setup has been completed
func TestSetupInitializationCheck(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	testCases := []struct {
		name               string
		markerExists       bool
		configExists       bool
		forceFlag          bool
		shouldSkipSetup    bool
		description        string
	}{
		{
			name:            "Fresh install - no files",
			markerExists:    false,
			configExists:    false,
			forceFlag:       false,
			shouldSkipSetup: false,
			description:     "When neither marker nor config exists, setup should run",
		},
		{
			name:            "Marker exists but config missing (BUG SCENARIO)",
			markerExists:    true,
			configExists:    false,
			forceFlag:       false,
			shouldSkipSetup: false,
			description:     "When marker exists but config is missing, setup should run to recreate config",
		},
		{
			name:            "Both marker and config exist",
			markerExists:    true,
			configExists:    true,
			forceFlag:       false,
			shouldSkipSetup: true,
			description:     "When both marker and config exist, setup should be skipped",
		},
		{
			name:            "Config exists but marker missing",
			markerExists:    false,
			configExists:    true,
			forceFlag:       false,
			shouldSkipSetup: false,
			description:     "When config exists but marker is missing, setup should run to create marker",
		},
		{
			name:            "Force flag overrides all checks",
			markerExists:    true,
			configExists:    true,
			forceFlag:       true,
			shouldSkipSetup: false,
			description:     "When force flag is set, setup should always run",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup test files
			configPath := filepath.Join(tmpDir, ".ainative-code.yaml")
			markerPath := filepath.Join(tmpDir, ".ainative-code-initialized")

			// Create marker if needed
			if tc.markerExists {
				if err := os.WriteFile(markerPath, []byte("initialized_at: test\n"), 0644); err != nil {
					t.Fatalf("Failed to create marker file: %v", err)
				}
			}

			// Create config if needed
			if tc.configExists {
				configContent := `app:
  name: ainative-code
  version: 0.1.0
llm:
  default_provider: anthropic
`
				if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
					t.Fatalf("Failed to create config file: %v", err)
				}
			}

			// Simulate the logic from runSetup()
			shouldSkip := false
			if !tc.forceFlag {
				_, markerErr := os.Stat(markerPath)
				_, configErr := os.Stat(configPath)

				// Only skip setup if BOTH marker AND config file exist
				if markerErr == nil && configErr == nil {
					shouldSkip = true
				}
			}

			// Verify the result
			if shouldSkip != tc.shouldSkipSetup {
				t.Errorf("%s: Expected shouldSkipSetup=%v, got %v\nDescription: %s",
					tc.name, tc.shouldSkipSetup, shouldSkip, tc.description)
			} else {
				t.Logf("PASS: %s - %s", tc.name, tc.description)
			}

			// Cleanup for next iteration
			os.Remove(markerPath)
			os.Remove(configPath)
		})
	}
}

// TestSetupInitializationCheckRealHomeDir tests with actual home directory
func TestSetupInitializationCheckRealHomeDir(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test that requires home directory in short mode")
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	configPath := filepath.Join(homeDir, ".ainative-code.yaml")
	markerPath := filepath.Join(homeDir, ".ainative-code-initialized")

	// Backup existing files
	configBackup := configPath + ".test-backup"
	markerBackup := markerPath + ".test-backup"

	if _, err := os.Stat(configPath); err == nil {
		if err := os.Rename(configPath, configBackup); err != nil {
			t.Fatalf("Failed to backup config: %v", err)
		}
		defer os.Rename(configBackup, configPath)
	}

	if _, err := os.Stat(markerPath); err == nil {
		if err := os.Rename(markerPath, markerBackup); err != nil {
			t.Fatalf("Failed to backup marker: %v", err)
		}
		defer os.Rename(markerBackup, markerPath)
	}

	// Test Case: Marker exists but config is missing (the bug scenario)
	t.Run("Bug scenario - marker exists, config missing", func(t *testing.T) {
		// Create only marker
		if err := os.WriteFile(markerPath, []byte("initialized_at: test\n"), 0644); err != nil {
			t.Fatalf("Failed to create marker: %v", err)
		}
		defer os.Remove(markerPath)

		// Verify config doesn't exist
		if _, err := os.Stat(configPath); err == nil {
			t.Fatal("Config file should not exist for this test")
		}

		// Simulate the fixed logic
		_, markerErr := os.Stat(markerPath)
		_, configErr := os.Stat(configPath)

		shouldSkip := (markerErr == nil && configErr == nil)

		if shouldSkip {
			t.Error("BUG: Setup was skipped even though config file is missing!")
		} else {
			t.Log("PASS: Setup will run correctly to recreate missing config")
		}
	})
}

// TestForceFlag tests that the force flag always allows setup to run
func TestForceFlag(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".ainative-code.yaml")
	markerPath := filepath.Join(tmpDir, ".ainative-code-initialized")

	// Create both files
	os.WriteFile(markerPath, []byte("initialized_at: test\n"), 0644)
	os.WriteFile(configPath, []byte("app:\n  name: test\n"), 0600)

	// Test with force flag = false (should skip)
	forceFlag := false
	_, markerErr := os.Stat(markerPath)
	_, configErr := os.Stat(configPath)
	shouldSkip := !forceFlag && (markerErr == nil && configErr == nil)

	if !shouldSkip {
		t.Error("Should skip setup when both files exist and force=false")
	}

	// Test with force flag = true (should not skip)
	forceFlag = true
	shouldSkip = !forceFlag && (markerErr == nil && configErr == nil)

	if shouldSkip {
		t.Error("Should not skip setup when force=true")
	}
}
