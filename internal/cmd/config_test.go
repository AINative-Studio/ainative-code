package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

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
		subcommand      *Command
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
		name       string
		setupViper func()
		wantErr    bool
	}{
		{
			name: "shows empty configuration",
			setupViper: func() {
				viper.Reset()
			},
			wantErr: false,
		},
		{
			name: "shows configuration with values",
			setupViper: func() {
				viper.Reset()
				viper.Set("provider", "openai")
				viper.Set("model", "gpt-4")
				viper.Set("verbose", true)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupViper()

			var buf bytes.Buffer
			configShowCmd.SetOut(&buf)

			err := runConfigShow(configShowCmd, []string{})

			if (err != nil) != tt.wantErr {
				t.Errorf("runConfigShow() error = %v, wantErr %v", err, tt.wantErr)
			}

			output := buf.String()
			if !strings.Contains(output, "Current Configuration") {
				t.Error("expected output to contain 'Current Configuration'")
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
	}{
		{
			name: "gets existing key",
			setupViper: func() {
				viper.Reset()
				viper.Set("provider", "openai")
			},
			args:    []string{"provider"},
			wantErr: false,
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
			args:    []string{"database.path"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupViper()

			var buf bytes.Buffer
			configGetCmd.SetOut(&buf)

			err := runConfigGet(configGetCmd, tt.args)

			if (err != nil) != tt.wantErr {
				t.Errorf("runConfigGet() error = %v, wantErr %v", err, tt.wantErr)
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
	}{
		{
			name: "validates valid configuration",
			setupViper: func() {
				viper.Reset()
				viper.Set("provider", "openai")
				viper.Set("model", "gpt-4")
			},
			wantErr: false,
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupViper()

			var buf bytes.Buffer
			configValidateCmd.SetOut(&buf)

			err := runConfigValidate(configValidateCmd, []string{})

			if (err != nil) != tt.wantErr {
				t.Errorf("runConfigValidate() error = %v, wantErr %v", err, tt.wantErr)
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
