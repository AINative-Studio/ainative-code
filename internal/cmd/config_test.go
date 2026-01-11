package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// TestConfigCommand tests the config command initialization
func TestConfigCommand(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "config command exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if configCmd == nil {
				t.Fatal("configCmd should not be nil")
			}

			if configCmd.Use != "config" {
				t.Errorf("expected Use 'config', got %s", configCmd.Use)
			}

			if configCmd.Short == "" {
				t.Error("expected Short description to be set")
			}

			// Verify aliases
			if len(configCmd.Aliases) == 0 {
				t.Error("expected aliases to be set")
			}
		})
	}
}

// TestConfigSubcommands tests config subcommands exist
func TestConfigSubcommands(t *testing.T) {
	tests := []struct {
		name            string
		subcommand      *cobra.Command
		expectedUse     string
		expectedAliases []string
	}{
		{
			name:            "show subcommand exists",
			subcommand:      configShowCmd,
			expectedUse:     "show",
			expectedAliases: []string{"list", "ls"},
		},
		{
			name:        "set subcommand exists",
			subcommand:  configSetCmd,
			expectedUse: "set [key] [value]",
		},
		{
			name:        "get subcommand exists",
			subcommand:  configGetCmd,
			expectedUse: "get [key]",
		},
		{
			name:        "init subcommand exists",
			subcommand:  configInitCmd,
			expectedUse: "init",
		},
		{
			name:        "validate subcommand exists",
			subcommand:  configValidateCmd,
			expectedUse: "validate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.subcommand == nil {
				t.Fatalf("subcommand %s should not be nil", tt.name)
			}

			if tt.subcommand.Use != tt.expectedUse {
				t.Errorf("expected Use %q, got %q", tt.expectedUse, tt.subcommand.Use)
			}

			if tt.expectedAliases != nil {
				if len(tt.subcommand.Aliases) != len(tt.expectedAliases) {
					t.Errorf("expected %d aliases, got %d", len(tt.expectedAliases), len(tt.subcommand.Aliases))
				}
			}
		})
	}
}

// TestRunConfigShow tests the config show command
func TestRunConfigShow(t *testing.T) {
	tests := []struct {
		name         string
		setupViper   func()
		showSecrets  bool
		wantErr      bool
		checkOutput  func(t *testing.T, output string)
	}{
		{
			name: "shows empty configuration",
			setupViper: func() {
				viper.Reset()
			},
			showSecrets: false,
			wantErr:     false,
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "Current Configuration") {
					t.Error("expected output to contain 'Current Configuration'")
				}
				if !strings.Contains(output, "No configuration values set") {
					t.Error("expected output to contain 'No configuration values set'")
				}
			},
		},
		{
			name: "shows configuration with values (masked)",
			setupViper: func() {
				viper.Reset()
				viper.Set("provider", "openai")
				viper.Set("model", "gpt-4")
				viper.Set("verbose", true)
			},
			showSecrets: false,
			wantErr:     false,
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "Current Configuration") {
					t.Error("expected output to contain 'Current Configuration'")
				}
				if !strings.Contains(output, "openai") {
					t.Error("expected output to contain 'openai'")
				}
			},
		},
		{
			name: "masks API key by default",
			setupViper: func() {
				viper.Reset()
				viper.Set("provider", "openai")
				viper.Set("api_key", "sk-1234567890abcdefghijklmnopqrstuvwxyz123456")
			},
			showSecrets: false,
			wantErr:     false,
			checkOutput: func(t *testing.T, output string) {
				if strings.Contains(output, "sk-1234567890abcdefghijklmnopqrstuvwxyz123456") {
					t.Error("API key should be masked")
				}
				if !strings.Contains(output, "Sensitive values are masked") {
					t.Error("expected masking notice")
				}
			},
		},
		{
			name: "shows API key with --show-secrets flag",
			setupViper: func() {
				viper.Reset()
				viper.Set("provider", "openai")
				viper.Set("api_key", "sk-test123456")
			},
			showSecrets: true,
			wantErr:     false,
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "sk-test123456") {
					t.Error("API key should be visible with --show-secrets")
				}
				if !strings.Contains(output, "WARNING") {
					t.Error("expected security warning")
				}
			},
		},
		{
			name: "masks nested sensitive values",
			setupViper: func() {
				viper.Reset()
				viper.Set("llm.openai.api_key", "sk-openai123456789")
				viper.Set("llm.openai.model", "gpt-4")
				viper.Set("platform.authentication.client_secret", "secret123")
			},
			showSecrets: false,
			wantErr:     false,
			checkOutput: func(t *testing.T, output string) {
				if strings.Contains(output, "sk-openai123456789") {
					t.Error("OpenAI API key should be masked")
				}
				if strings.Contains(output, "secret123") {
					t.Error("client_secret should be masked")
				}
				if !strings.Contains(output, "gpt-4") {
					t.Error("non-sensitive model value should be visible")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupViper()

			// Capture stdout since the implementation uses fmt.Println
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Set the flag
			if tt.showSecrets {
				configShowCmd.Flags().Set("show-secrets", "true")
			} else {
				configShowCmd.Flags().Set("show-secrets", "false")
			}

			err := runConfigShow(configShowCmd, []string{})

			// Restore stdout and read captured output
			w.Close()
			os.Stdout = oldStdout
			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			if (err != nil) != tt.wantErr {
				t.Errorf("runConfigShow() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.checkOutput != nil {
				tt.checkOutput(t, output)
			}
		})
	}
}

// TestRunConfigSet tests the config set command
func TestRunConfigSet(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "sets provider",
			args:    []string{"provider", "openai"},
			wantErr: false,
		},
		{
			name:    "sets model",
			args:    []string{"model", "gpt-4"},
			wantErr: false,
		},
		{
			name:    "sets verbose",
			args:    []string{"verbose", "true"},
			wantErr: false,
		},
		{
			name:    "sets custom key",
			args:    []string{"custom.key", "custom-value"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use temp directory for config file
			tmpDir := t.TempDir()
			configFile := filepath.Join(tmpDir, ".ainative-code.yaml")

			// Reset viper
			viper.Reset()
			viper.SetConfigFile(configFile)

			var buf bytes.Buffer
			configSetCmd.SetOut(&buf)

			err := runConfigSet(configSetCmd, tt.args)

			if (err != nil) != tt.wantErr {
				t.Errorf("runConfigSet() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Verify value was set
			if !tt.wantErr {
				if viper.Get(tt.args[0]) != tt.args[1] {
					t.Errorf("expected value %s, got %v", tt.args[1], viper.Get(tt.args[0]))
				}
			}
		})
	}
}

// TestRunConfigGet tests the config get command
func TestRunConfigGet(t *testing.T) {
	tests := []struct {
		name       string
		setupViper func()
		args       []string
		wantErr    bool
		wantOutput string
	}{
		{
			name: "gets existing key",
			setupViper: func() {
				viper.Reset()
				viper.Set("provider", "openai")
			},
			args:       []string{"provider"},
			wantErr:    false,
			wantOutput: "provider: openai",
		},
		{
			name: "gets missing key",
			setupViper: func() {
				viper.Reset()
			},
			args:    []string{"nonexistent"},
			wantErr: true,
		},
		{
			name: "gets nested key",
			setupViper: func() {
				viper.Reset()
				viper.Set("database.path", "/path/to/db")
			},
			args:       []string{"database.path"},
			wantErr:    false,
			wantOutput: "database.path: /path/to/db",
		},
		{
			name: "gets empty string value - Issue #101",
			setupViper: func() {
				viper.Reset()
				viper.Set("provider", "")
			},
			args:       []string{"provider"},
			wantErr:    false,
			wantOutput: "provider: (empty)",
		},
		{
			name: "gets nil value - Issue #101",
			setupViper: func() {
				viper.Reset()
				viper.Set("model", nil)
			},
			args:       []string{"model"},
			wantErr:    true,
			wantOutput: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupViper()

			// Capture stdout since the implementation uses fmt.Printf
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := runConfigGet(configGetCmd, tt.args)

			// Restore stdout and read captured output
			w.Close()
			os.Stdout = oldStdout
			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := strings.TrimSpace(buf.String())

			if (err != nil) != tt.wantErr {
				t.Errorf("runConfigGet() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && tt.wantOutput != "" {
				if !strings.Contains(output, tt.wantOutput) {
					t.Errorf("runConfigGet() output = %q, want to contain %q", output, tt.wantOutput)
				}
			}
		})
	}
}

// TestRunConfigInit tests the config init command
func TestRunConfigInit(t *testing.T) {
	tests := []struct {
		name        string
		force       bool
		existingFile bool
		wantErr     bool
	}{
		{
			name:        "creates new config file",
			force:       false,
			existingFile: false,
			wantErr:     false,
		},
		{
			name:        "fails with existing file",
			force:       false,
			existingFile: true,
			wantErr:     true,
		},
		{
			name:        "overwrites with force flag",
			force:       true,
			existingFile: true,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use temp directory
			tmpDir := t.TempDir()

			// Mock home directory
			oldHome := os.Getenv("HOME")
			os.Setenv("HOME", tmpDir)
			defer os.Setenv("HOME", oldHome)

			configFile := filepath.Join(tmpDir, ".ainative-code.yaml")

			// Create existing file if needed
			if tt.existingFile {
				if err := os.WriteFile(configFile, []byte("existing: config\n"), 0644); err != nil {
					t.Fatalf("failed to create existing file: %v", err)
				}
			}

			// Reset viper
			viper.Reset()

			// Set force flag
			configInitCmd.Flags().Set("force", "false")
			if tt.force {
				configInitCmd.Flags().Set("force", "true")
			}

			var buf bytes.Buffer
			configInitCmd.SetOut(&buf)

			err := runConfigInit(configInitCmd, []string{})

			if (err != nil) != tt.wantErr {
				t.Errorf("runConfigInit() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Verify file was created if no error
			if !tt.wantErr {
				if _, err := os.Stat(configFile); os.IsNotExist(err) {
					t.Error("config file was not created")
				}
			}
		})
	}
}

// TestRunConfigValidate tests the config validate command
func TestRunConfigValidate(t *testing.T) {
	tests := []struct {
		name       string
		setupViper func()
		wantErr    bool
		wantOutput string
	}{
		{
			name: "validates valid configuration with legacy provider field",
			setupViper: func() {
				viper.Reset()
				viper.Set("provider", "openai")
				viper.Set("model", "gpt-4")
			},
			wantErr:    false,
			wantOutput: "Provider: openai (from provider)",
		},
		{
			name: "validates valid configuration with new llm.default_provider field",
			setupViper: func() {
				viper.Reset()
				viper.Set("llm.default_provider", "anthropic")
				viper.Set("llm.anthropic.model", "claude-3-5-sonnet-20241022")
			},
			wantErr:    false,
			wantOutput: "Provider: anthropic (from llm.default_provider)",
		},
		{
			name: "prefers llm.default_provider over legacy provider",
			setupViper: func() {
				viper.Reset()
				viper.Set("provider", "openai")
				viper.Set("llm.default_provider", "anthropic")
			},
			wantErr:    false,
			wantOutput: "Provider: anthropic (from llm.default_provider)",
		},
		{
			name: "fails with missing provider",
			setupViper: func() {
				viper.Reset()
				viper.Set("model", "gpt-4")
			},
			wantErr: true,
		},
		{
			name: "fails with empty provider - Issue #122",
			setupViper: func() {
				viper.Reset()
				viper.Set("provider", "")
				viper.Set("llm.default_provider", "anthropic")
			},
			wantErr:    false,
			wantOutput: "Provider: anthropic (from llm.default_provider)",
		},
		{
			name: "fails with empty root provider and no nested provider - Issue #122",
			setupViper: func() {
				viper.Reset()
				viper.Set("provider", "")
			},
			wantErr: true,
		},
		{
			name: "fails with invalid provider",
			setupViper: func() {
				viper.Reset()
				viper.Set("provider", "invalid-provider")
			},
			wantErr: true,
		},
		{
			name: "validates anthropic provider",
			setupViper: func() {
				viper.Reset()
				viper.Set("provider", "anthropic")
			},
			wantErr: false,
		},
		{
			name: "validates ollama provider",
			setupViper: func() {
				viper.Reset()
				viper.Set("provider", "ollama")
			},
			wantErr: false,
		},
		{
			name: "validates google provider",
			setupViper: func() {
				viper.Reset()
				viper.Set("llm.default_provider", "google")
			},
			wantErr: false,
		},
		{
			name: "validates bedrock provider",
			setupViper: func() {
				viper.Reset()
				viper.Set("llm.default_provider", "bedrock")
			},
			wantErr: false,
		},
		{
			name: "validates azure provider",
			setupViper: func() {
				viper.Reset()
				viper.Set("llm.default_provider", "azure")
			},
			wantErr: false,
		},
		{
			name: "validates meta_llama provider",
			setupViper: func() {
				viper.Reset()
				viper.Set("llm.default_provider", "meta_llama")
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupViper()

			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := runConfigValidate(configValidateCmd, []string{})

			// Restore stdout and read output
			w.Close()
			os.Stdout = oldStdout
			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			if (err != nil) != tt.wantErr {
				t.Errorf("runConfigValidate() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && tt.wantOutput != "" {
				if !strings.Contains(output, tt.wantOutput) {
					t.Errorf("runConfigValidate() output = %q, want to contain %q", output, tt.wantOutput)
				}
			}
		})
	}
}

// TestConfigCommandIntegration tests complete config command flow
func TestConfigCommandIntegration(t *testing.T) {
	// Use temp directory
	tmpDir := t.TempDir()

	// Mock home directory
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// Reset viper
	viper.Reset()

	// Test init
	configInitCmd.Flags().Set("force", "false")
	err := runConfigInit(configInitCmd, []string{})
	if err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Test set
	err = runConfigSet(configSetCmd, []string{"provider", "anthropic"})
	if err != nil {
		t.Fatalf("set failed: %v", err)
	}

	// Test get
	var getBuf bytes.Buffer
	configGetCmd.SetOut(&getBuf)
	err = runConfigGet(configGetCmd, []string{"provider"})
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}

	// Test show
	var showBuf bytes.Buffer
	configShowCmd.SetOut(&showBuf)
	err = runConfigShow(configShowCmd, []string{})
	if err != nil {
		t.Fatalf("show failed: %v", err)
	}

	// Test validate
	var validateBuf bytes.Buffer
	configValidateCmd.SetOut(&validateBuf)
	err = runConfigValidate(configValidateCmd, []string{})
	if err != nil {
		t.Fatalf("validate failed: %v", err)
	}
}

// TestConfigSetArgs tests that set command requires exactly 2 args
func TestConfigSetArgs(t *testing.T) {
	if configSetCmd.Args == nil {
		t.Fatal("configSetCmd.Args should not be nil")
	}

	// Test with correct number of args
	err := configSetCmd.Args(configSetCmd, []string{"key", "value"})
	if err != nil {
		t.Errorf("expected no error with 2 args, got %v", err)
	}

	// Test with wrong number of args
	err = configSetCmd.Args(configSetCmd, []string{"key"})
	if err == nil {
		t.Error("expected error with 1 arg")
	}

	err = configSetCmd.Args(configSetCmd, []string{"key", "value", "extra"})
	if err == nil {
		t.Error("expected error with 3 args")
	}
}

// TestConfigGetArgs tests that get command requires exactly 1 arg
func TestConfigGetArgs(t *testing.T) {
	if configGetCmd.Args == nil {
		t.Fatal("configGetCmd.Args should not be nil")
	}

	// Test with correct number of args
	err := configGetCmd.Args(configGetCmd, []string{"key"})
	if err != nil {
		t.Errorf("expected no error with 1 arg, got %v", err)
	}

	// Test with wrong number of args
	err = configGetCmd.Args(configGetCmd, []string{})
	if err == nil {
		t.Error("expected error with 0 args")
	}

	err = configGetCmd.Args(configGetCmd, []string{"key", "extra"})
	if err == nil {
		t.Error("expected error with 2 args")
	}
}

// TestConfigInitForceFlag tests the force flag
func TestConfigInitForceFlag(t *testing.T) {
	flag := configInitCmd.Flags().Lookup("force")
	if flag == nil {
		t.Fatal("force flag should exist")
	}

	if flag.Shorthand != "f" {
		t.Errorf("expected shorthand 'f', got %s", flag.Shorthand)
	}

	if flag.DefValue != "false" {
		t.Errorf("expected default value 'false', got %s", flag.DefValue)
	}
}

// TestValidateConfigKey tests the key validation function
func TestValidateConfigKey(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr bool
		errMsg  string
	}{
		// Valid keys
		{
			name:    "valid simple key",
			key:     "provider",
			wantErr: false,
		},
		{
			name:    "valid nested key",
			key:     "database.path",
			wantErr: false,
		},
		{
			name:    "valid key with underscores",
			key:     "api_key",
			wantErr: false,
		},
		{
			name:    "valid key with hyphens",
			key:     "max-tokens",
			wantErr: false,
		},
		{
			name:    "valid complex key",
			key:     "llm.openai.api_key",
			wantErr: false,
		},
		{
			name:    "valid key with numbers",
			key:     "timeout_ms_500",
			wantErr: false,
		},
		{
			name:    "valid mixed case key",
			key:     "OpenAI.ApiKey",
			wantErr: false,
		},

		// Invalid keys - empty and whitespace
		{
			name:    "empty key",
			key:     "",
			wantErr: true,
			errMsg:  "key name cannot be empty",
		},
		{
			name:    "whitespace only key - spaces",
			key:     "   ",
			wantErr: true,
			errMsg:  "key name cannot be whitespace only",
		},
		{
			name:    "whitespace only key - tabs",
			key:     "\t\t",
			wantErr: true,
			errMsg:  "key name cannot be whitespace only",
		},
		{
			name:    "whitespace only key - mixed",
			key:     " \t \n ",
			wantErr: true,
			errMsg:  "key name cannot be whitespace only",
		},

		// Invalid keys - invalid characters
		{
			name:    "key with spaces",
			key:     "api key",
			wantErr: true,
			errMsg:  "invalid character",
		},
		{
			name:    "key with slash",
			key:     "api/key",
			wantErr: true,
			errMsg:  "invalid character",
		},
		{
			name:    "key with equals",
			key:     "api=key",
			wantErr: true,
			errMsg:  "invalid character",
		},
		{
			name:    "key with special chars",
			key:     "api@key",
			wantErr: true,
			errMsg:  "invalid character",
		},
		{
			name:    "key with brackets",
			key:     "api[key]",
			wantErr: true,
			errMsg:  "invalid character",
		},

		// Invalid keys - dot issues
		{
			name:    "key starting with dot",
			key:     ".api_key",
			wantErr: true,
			errMsg:  "cannot start with a dot",
		},
		{
			name:    "key ending with dot",
			key:     "api_key.",
			wantErr: true,
			errMsg:  "cannot end with a dot",
		},
		{
			name:    "key with consecutive dots",
			key:     "api..key",
			wantErr: true,
			errMsg:  "cannot contain consecutive dots",
		},

		// Invalid keys - length
		{
			name:    "key too long",
			key:     strings.Repeat("a", 101),
			wantErr: true,
			errMsg:  "exceeds maximum length",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfigKey(tt.key)

			if (err != nil) != tt.wantErr {
				t.Errorf("validateConfigKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" {
				if err == nil || !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("validateConfigKey() error = %v, want error containing %q", err, tt.errMsg)
				}
			}
		})
	}
}

// TestRunConfigSetWithInvalidKeys tests config set with invalid keys
func TestRunConfigSetWithInvalidKeys(t *testing.T) {
	tests := []struct {
		name       string
		key        string
		value      string
		wantErr    bool
		errContains string
	}{
		{
			name:       "empty key rejected",
			key:        "",
			value:      "some-value",
			wantErr:    true,
			errContains: "key name cannot be empty",
		},
		{
			name:       "whitespace key rejected",
			key:        "   ",
			value:      "some-value",
			wantErr:    true,
			errContains: "key name cannot be whitespace only",
		},
		{
			name:       "key with spaces rejected",
			key:        "api key",
			value:      "some-value",
			wantErr:    true,
			errContains: "invalid character",
		},
		{
			name:       "key with special chars rejected",
			key:        "api@key",
			value:      "some-value",
			wantErr:    true,
			errContains: "invalid character",
		},
		{
			name:       "valid key accepted",
			key:        "valid_key",
			value:      "some-value",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use temp directory for config file
			tmpDir := t.TempDir()
			configFile := filepath.Join(tmpDir, ".ainative-code.yaml")

			// Reset viper
			viper.Reset()
			viper.SetConfigFile(configFile)

			var buf bytes.Buffer
			configSetCmd.SetOut(&buf)

			err := runConfigSet(configSetCmd, []string{tt.key, tt.value})

			if (err != nil) != tt.wantErr {
				t.Errorf("runConfigSet() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errContains != "" {
				if err == nil || !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("runConfigSet() error = %v, want error containing %q", err, tt.errContains)
				}
			}
		})
	}
}

// Benchmark tests for performance validation

// BenchmarkRunConfigShow benchmarks config show command
func BenchmarkRunConfigShow(b *testing.B) {
	viper.Reset()
	viper.Set("provider", "openai")
	viper.Set("model", "gpt-4")

	var buf bytes.Buffer
	configShowCmd.SetOut(&buf)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = runConfigShow(configShowCmd, []string{})
		buf.Reset()
	}
}

// BenchmarkRunConfigGet benchmarks config get command
func BenchmarkRunConfigGet(b *testing.B) {
	viper.Reset()
	viper.Set("provider", "openai")

	var buf bytes.Buffer
	configGetCmd.SetOut(&buf)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = runConfigGet(configGetCmd, []string{"provider"})
		buf.Reset()
	}
}

// BenchmarkRunConfigValidate benchmarks config validate command
func BenchmarkRunConfigValidate(b *testing.B) {
	viper.Reset()
	viper.Set("provider", "openai")

	var buf bytes.Buffer
	configValidateCmd.SetOut(&buf)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = runConfigValidate(configValidateCmd, []string{})
		buf.Reset()
	}
}
