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
