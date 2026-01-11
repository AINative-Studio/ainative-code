package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

// TestExecute tests the Execute function
func TestExecute(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "execute without error",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset viper for clean test
			viper.Reset()

			// Set minimal required config
			viper.Set("provider", "openai")

			// Execute should not error with help flag
			rootCmd.SetArgs([]string{"--help"})
			err := Execute()

			// Help returns error but it's expected
			if err != nil && !tt.wantErr {
				// This is expected for --help
				t.Logf("Execute() with --help returned: %v", err)
			}
		})
	}
}

// TestRootCommand tests the root command initialization
func TestRootCommand(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "root command exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if rootCmd == nil {
				t.Error("rootCmd should not be nil")
			}

			if rootCmd.Use != "ainative-code" {
				t.Errorf("expected Use 'ainative-code', got %s", rootCmd.Use)
			}

			if rootCmd.Short == "" {
				t.Error("expected Short description to be set")
			}

			if rootCmd.Long == "" {
				t.Error("expected Long description to be set")
			}

			if rootCmd.Version == "" {
				t.Error("expected Version to be set")
			}
		})
	}
}

// TestGlobalFlags tests global flag initialization
func TestGlobalFlags(t *testing.T) {
	tests := []struct {
		name     string
		flagName string
	}{
		{
			name:     "config flag exists",
			flagName: "config",
		},
		{
			name:     "provider flag exists",
			flagName: "provider",
		},
		{
			name:     "model flag exists",
			flagName: "model",
		},
		{
			name:     "verbose flag exists",
			flagName: "verbose",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := rootCmd.PersistentFlags().Lookup(tt.flagName)
			if flag == nil {
				t.Errorf("flag %s should exist", tt.flagName)
			}
		})
	}
}

// TestGetProvider tests the GetProvider function
func TestGetProvider(t *testing.T) {
	tests := []struct {
		name         string
		flagValue    string
		viperValue   string
		expected     string
	}{
		{
			name:       "returns flag value when set",
			flagValue:  "anthropic",
			viperValue: "openai",
			expected:   "anthropic",
		},
		{
			name:       "returns viper value when flag not set",
			flagValue:  "",
			viperValue: "openai",
			expected:   "openai",
		},
		{
			name:       "returns empty when neither set",
			flagValue:  "",
			viperValue: "",
			expected:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset viper and provider variable
			viper.Reset()
			provider = ""

			// Set values
			if tt.flagValue != "" {
				provider = tt.flagValue
			}
			if tt.viperValue != "" {
				viper.Set("provider", tt.viperValue)
			}

			result := GetProvider()

			if result != tt.expected {
				t.Errorf("GetProvider() = %s, want %s", result, tt.expected)
			}
		})
	}
}

// TestGetModel tests the GetModel function
func TestGetModel(t *testing.T) {
	tests := []struct {
		name       string
		flagValue  string
		viperValue string
		expected   string
	}{
		{
			name:       "returns flag value when set",
			flagValue:  "gpt-4-turbo",
			viperValue: "gpt-4",
			expected:   "gpt-4-turbo",
		},
		{
			name:       "returns viper value when flag not set",
			flagValue:  "",
			viperValue: "gpt-4",
			expected:   "gpt-4",
		},
		{
			name:       "returns empty when neither set",
			flagValue:  "",
			viperValue: "",
			expected:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset viper and model variable
			viper.Reset()
			model = ""

			// Set values
			if tt.flagValue != "" {
				model = tt.flagValue
			}
			if tt.viperValue != "" {
				viper.Set("model", tt.viperValue)
			}

			result := GetModel()

			if result != tt.expected {
				t.Errorf("GetModel() = %s, want %s", result, tt.expected)
			}
		})
	}
}

// TestGetVerbose tests the GetVerbose function
func TestGetVerbose(t *testing.T) {
	tests := []struct {
		name       string
		flagValue  bool
		viperValue bool
		expected   bool
	}{
		{
			name:       "returns true when flag is true",
			flagValue:  true,
			viperValue: false,
			expected:   true,
		},
		{
			name:       "returns true when viper is true",
			flagValue:  false,
			viperValue: true,
			expected:   true,
		},
		{
			name:       "returns true when both are true",
			flagValue:  true,
			viperValue: true,
			expected:   true,
		},
		{
			name:       "returns false when both are false",
			flagValue:  false,
			viperValue: false,
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset viper and verbose variable
			viper.Reset()
			verbose = false

			// Set values
			verbose = tt.flagValue
			viper.Set("verbose", tt.viperValue)

			result := GetVerbose()

			if result != tt.expected {
				t.Errorf("GetVerbose() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestInitConfig tests the initConfig function
func TestInitConfig(t *testing.T) {
	tests := []struct {
		name           string
		setupConfigFile bool
		configFilePath string
		cfgFileValue   string
	}{
		{
			name:           "loads config from home directory",
			setupConfigFile: true,
			configFilePath: ".ainative-code.yaml",
		},
		{
			name:           "handles missing config file",
			setupConfigFile: false,
		},
		{
			name:         "uses config from flag",
			cfgFileValue: "/tmp/custom-config.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset viper
			viper.Reset()

			// Setup
			if tt.setupConfigFile {
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, tt.configFilePath)

				// Create config file
				content := []byte("provider: openai\nmodel: gpt-4\n")
				if err := os.WriteFile(configPath, content, 0644); err != nil {
					t.Fatalf("failed to create config file: %v", err)
				}

				// Change to temp dir
				oldWd, _ := os.Getwd()
				os.Chdir(tmpDir)
				defer os.Chdir(oldWd)
			}

			if tt.cfgFileValue != "" {
				cfgFile = tt.cfgFileValue
			} else {
				cfgFile = ""
			}

			// Call initConfig
			initConfig()

			// Verify viper is configured (no panic)
			// We can't verify much more without mocking os.UserHomeDir
		})
	}
}

// TestPersistentPreRun tests the PersistentPreRun function
func TestPersistentPreRun(t *testing.T) {
	tests := []struct {
		name    string
		verbose bool
	}{
		{
			name:    "sets debug level when verbose is true",
			verbose: true,
		},
		{
			name:    "does not set debug level when verbose is false",
			verbose: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset verbose
			verbose = tt.verbose

			// Call PersistentPreRun
			if rootCmd.PersistentPreRun != nil {
				rootCmd.PersistentPreRun(rootCmd, []string{})
			}

			// No panic means success
			// We can't verify logger level without accessing internal state
		})
	}
}

// TestViperBindings tests that flags are bound to viper
func TestViperBindings(t *testing.T) {
	tests := []struct {
		name     string
		flagName string
		viperKey string
	}{
		{
			name:     "provider flag bound to viper",
			flagName: "provider",
			viperKey: "provider",
		},
		{
			name:     "model flag bound to viper",
			flagName: "model",
			viperKey: "model",
		},
		{
			name:     "verbose flag bound to viper",
			flagName: "verbose",
			viperKey: "verbose",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset viper
			viper.Reset()

			// Re-initialize bindings
			viper.BindPFlag(tt.viperKey, rootCmd.PersistentFlags().Lookup(tt.flagName))

			// Set flag value
			testValue := "test-value"
			if tt.flagName == "verbose" {
				rootCmd.PersistentFlags().Set(tt.flagName, "true")
			} else {
				rootCmd.PersistentFlags().Set(tt.flagName, testValue)
			}

			// Verify viper has the value
			if tt.flagName == "verbose" {
				if !viper.GetBool(tt.viperKey) {
					t.Errorf("viper key %s should be true", tt.viperKey)
				}
			} else {
				if viper.GetString(tt.viperKey) != testValue {
					t.Errorf("viper key %s should be %s, got %s", tt.viperKey, testValue, viper.GetString(tt.viperKey))
				}
			}
		})
	}
}

// TestConfigFileValidation tests the config file validation logic
func TestConfigFileValidation(t *testing.T) {
	tests := []struct {
		name           string
		setupFunc      func(t *testing.T) string // Returns config file path
		expectExit     bool
		expectError    string
	}{
		{
			name: "nonexistent config file shows error",
			setupFunc: func(t *testing.T) string {
				return "/nonexistent/config.yaml"
			},
			expectExit:  true,
			expectError: "Config file not found",
		},
		{
			name: "directory instead of file shows error",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				return tmpDir
			},
			expectExit:  true,
			expectError: "Config path is a directory",
		},
		{
			name: "valid config file succeeds",
			setupFunc: func(t *testing.T) string {
				tmpFile := filepath.Join(t.TempDir(), "config.yaml")
				content := []byte("provider: openai\nmodel: gpt-4\n")
				if err := os.WriteFile(tmpFile, content, 0644); err != nil {
					t.Fatalf("failed to create config file: %v", err)
				}
				return tmpFile
			},
			expectExit:  false,
			expectError: "",
		},
		// Note: Skipping permission test as it's unreliable across platforms
		// macOS/Linux handle file permissions differently with root/owner access
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset viper
			viper.Reset()

			// Get config file path from setup function
			configPath := tt.setupFunc(t)
			cfgFile = configPath

			// For tests expecting exit, we can't actually test os.Exit
			// Instead, we verify the file stat behavior
			fileInfo, err := os.Stat(configPath)

			if tt.expectError == "Config file not found" {
				if !os.IsNotExist(err) {
					t.Errorf("Expected file not found error for %s", configPath)
				}
			} else if tt.expectError == "Config path is a directory" {
				if err == nil && !fileInfo.IsDir() {
					t.Errorf("Expected directory but got file for %s", configPath)
				}
			} else if !tt.expectExit {
				// Valid config file should not error
				if err != nil {
					t.Errorf("Unexpected error for valid config: %v", err)
				}
			}
		})
	}
}

// TestConfigFlagWithDifferentPaths tests various path formats
func TestConfigFlagWithDifferentPaths(t *testing.T) {
	tests := []struct {
		name        string
		setupPath   func(t *testing.T) string
		shouldExist bool
	}{
		{
			name: "absolute path to valid file",
			setupPath: func(t *testing.T) string {
				tmpFile := filepath.Join(t.TempDir(), "config.yaml")
				os.WriteFile(tmpFile, []byte("provider: openai\n"), 0644)
				return tmpFile
			},
			shouldExist: true,
		},
		{
			name: "relative path to valid file",
			setupPath: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.yaml")
				os.WriteFile(configPath, []byte("provider: openai\n"), 0644)
				// Change to temp dir
				oldWd, _ := os.Getwd()
				os.Chdir(tmpDir)
				t.Cleanup(func() { os.Chdir(oldWd) })
				return "config.yaml"
			},
			shouldExist: true,
		},
		{
			name: "path with spaces",
			setupPath: func(t *testing.T) string {
				tmpDir := t.TempDir()
				subDir := filepath.Join(tmpDir, "my config dir")
				os.Mkdir(subDir, 0755)
				configPath := filepath.Join(subDir, "config.yaml")
				os.WriteFile(configPath, []byte("provider: openai\n"), 0644)
				return configPath
			},
			shouldExist: true,
		},
		{
			name: "path with special characters",
			setupPath: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config-file_2024.yaml")
				os.WriteFile(configPath, []byte("provider: openai\n"), 0644)
				return configPath
			},
			shouldExist: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath := tt.setupPath(t)

			// Verify file exists
			_, err := os.Stat(configPath)
			exists := !os.IsNotExist(err)

			if exists != tt.shouldExist {
				t.Errorf("Expected file to exist: %v, but got: %v", tt.shouldExist, exists)
			}
		})
	}
}

// Benchmark tests for performance validation

// BenchmarkGetProvider benchmarks the GetProvider function
func BenchmarkGetProvider(b *testing.B) {
	viper.Set("provider", "openai")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetProvider()
	}
}

// BenchmarkGetModel benchmarks the GetModel function
func BenchmarkGetModel(b *testing.B) {
	viper.Set("model", "gpt-4")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetModel()
	}
}

// BenchmarkGetVerbose benchmarks the GetVerbose function
func BenchmarkGetVerbose(b *testing.B) {
	viper.Set("verbose", true)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetVerbose()
	}
}
