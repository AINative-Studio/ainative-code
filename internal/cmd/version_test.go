package cmd

import (
	"bytes"
	"encoding/json"
	"runtime"
	"strings"
	"testing"
)

// TestVersionCommand tests the version command initialization
func TestVersionCommand(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "version command exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if versionCmd == nil {
				t.Fatal("versionCmd should not be nil")
			}

			if versionCmd.Use != "version" {
				t.Errorf("expected Use 'version', got %s", versionCmd.Use)
			}

			if versionCmd.Short == "" {
				t.Error("expected Short description to be set")
			}

			if versionCmd.Long == "" {
				t.Error("expected Long description to be set")
			}

			// Verify aliases
			expectedAliases := []string{"v", "ver"}
			if len(versionCmd.Aliases) != len(expectedAliases) {
				t.Errorf("expected %d aliases, got %d", len(expectedAliases), len(versionCmd.Aliases))
			}

			for i, alias := range expectedAliases {
				if i >= len(versionCmd.Aliases) || versionCmd.Aliases[i] != alias {
					t.Errorf("expected alias %s at index %d", alias, i)
				}
			}
		})
	}
}

// TestVersionFlags tests the version command flags
func TestVersionFlags(t *testing.T) {
	tests := []struct {
		name      string
		flagName  string
		shorthand string
	}{
		{
			name:      "short flag exists",
			flagName:  "short",
			shorthand: "s",
		},
		{
			name:      "json flag exists",
			flagName:  "json",
			shorthand: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := versionCmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Errorf("flag %s should exist", tt.flagName)
				return
			}

			if tt.shorthand != "" && flag.Shorthand != tt.shorthand {
				t.Errorf("expected shorthand %s, got %s", tt.shorthand, flag.Shorthand)
			}
		})
	}
}

// TestRunVersionNormal tests normal version output
func TestRunVersionNormal(t *testing.T) {
	// Set version info
	SetVersion("1.0.0")
	SetCommit("abc123")
	SetBuildDate("2024-01-01")
	SetBuiltBy("test")

	var buf bytes.Buffer
	versionCmd.SetOut(&buf)

	// Reset flags
	versionCmd.Flags().Set("short", "false")
	versionCmd.Flags().Set("json", "false")

	runVersion(versionCmd, []string{})

	output := buf.String()

	// Verify output contains expected information
	expectedStrings := []string{
		"AINative Code",
		"1.0.0",
		"Commit:",
		"abc123",
		"Built:",
		"2024-01-01",
		"Built by:",
		"test",
		"Go version:",
		runtime.Version(),
		"Platform:",
		runtime.GOOS,
		runtime.GOARCH,
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("expected output to contain %q, got:\n%s", expected, output)
		}
	}
}

// TestRunVersionShort tests short version output
func TestRunVersionShort(t *testing.T) {
	// Set version
	SetVersion("1.2.3")

	var buf bytes.Buffer
	versionCmd.SetOut(&buf)

	// Set short flag
	versionCmd.Flags().Set("short", "true")
	versionCmd.Flags().Set("json", "false")

	runVersion(versionCmd, []string{})

	output := strings.TrimSpace(buf.String())

	// Should only output version number
	if output != "1.2.3" {
		t.Errorf("expected output '1.2.3', got %q", output)
	}

	// Should not contain other info
	if strings.Contains(output, "Commit:") {
		t.Error("short output should not contain 'Commit:'")
	}
	if strings.Contains(output, "Built:") {
		t.Error("short output should not contain 'Built:'")
	}
}

// TestRunVersionJSON tests JSON version output
func TestRunVersionJSON(t *testing.T) {
	// Set version info
	SetVersion("2.0.0")
	SetCommit("def456")
	SetBuildDate("2024-02-01")
	SetBuiltBy("ci")

	var buf bytes.Buffer
	versionCmd.SetOut(&buf)

	// Set json flag
	versionCmd.Flags().Set("short", "false")
	versionCmd.Flags().Set("json", "true")

	runVersion(versionCmd, []string{})

	output := buf.String()

	// Parse JSON
	var versionInfo map[string]string
	if err := json.Unmarshal([]byte(output), &versionInfo); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nOutput: %s", err, output)
	}

	// Verify fields
	tests := []struct {
		field    string
		expected string
	}{
		{
			field:    "version",
			expected: "2.0.0",
		},
		{
			field:    "commit",
			expected: "def456",
		},
		{
			field:    "buildDate",
			expected: "2024-02-01",
		},
		{
			field:    "builtBy",
			expected: "ci",
		},
		{
			field:    "goVersion",
			expected: runtime.Version(),
		},
	}

	for _, tt := range tests {
		if versionInfo[tt.field] != tt.expected {
			t.Errorf("expected %s to be %q, got %q", tt.field, tt.expected, versionInfo[tt.field])
		}
	}

	// Verify platform field format
	expectedPlatform := runtime.GOOS + "/" + runtime.GOARCH
	if versionInfo["platform"] != expectedPlatform {
		t.Errorf("expected platform %q, got %q", expectedPlatform, versionInfo["platform"])
	}
}

// TestGetVersion tests the GetVersion function
func TestGetVersion(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected string
	}{
		{
			name:     "returns version",
			version:  "1.0.0",
			expected: "1.0.0",
		},
		{
			name:     "returns dev version",
			version:  "dev",
			expected: "dev",
		},
		{
			name:     "returns semantic version",
			version:  "1.2.3-beta.1",
			expected: "1.2.3-beta.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetVersion(tt.version)

			result := GetVersion()

			if result != tt.expected {
				t.Errorf("GetVersion() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestGetCommit tests the GetCommit function
func TestGetCommit(t *testing.T) {
	tests := []struct {
		name     string
		commit   string
		expected string
	}{
		{
			name:     "returns full commit hash",
			commit:   "abc123def456",
			expected: "abc123def456",
		},
		{
			name:     "returns short commit hash",
			commit:   "abc123",
			expected: "abc123",
		},
		{
			name:     "returns none",
			commit:   "none",
			expected: "none",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetCommit(tt.commit)

			result := GetCommit()

			if result != tt.expected {
				t.Errorf("GetCommit() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestGetBuildDate tests the GetBuildDate function
func TestGetBuildDate(t *testing.T) {
	tests := []struct {
		name      string
		buildDate string
		expected  string
	}{
		{
			name:      "returns build date",
			buildDate: "2024-01-15T10:30:00Z",
			expected:  "2024-01-15T10:30:00Z",
		},
		{
			name:      "returns simple date",
			buildDate: "2024-01-15",
			expected:  "2024-01-15",
		},
		{
			name:      "returns unknown",
			buildDate: "unknown",
			expected:  "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetBuildDate(tt.buildDate)

			result := GetBuildDate()

			if result != tt.expected {
				t.Errorf("GetBuildDate() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestSetters tests all setter functions
func TestSetters(t *testing.T) {
	tests := []struct {
		name      string
		version   string
		commit    string
		buildDate string
		builtBy   string
	}{
		{
			name:      "sets all values",
			version:   "3.0.0",
			commit:    "xyz789",
			buildDate: "2024-03-01",
			builtBy:   "github-actions",
		},
		{
			name:      "sets dev values",
			version:   "dev",
			commit:    "none",
			buildDate: "unknown",
			builtBy:   "manual",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetVersion(tt.version)
			SetCommit(tt.commit)
			SetBuildDate(tt.buildDate)
			SetBuiltBy(tt.builtBy)

			if GetVersion() != tt.version {
				t.Errorf("version = %q, want %q", GetVersion(), tt.version)
			}
			if GetCommit() != tt.commit {
				t.Errorf("commit = %q, want %q", GetCommit(), tt.commit)
			}
			if GetBuildDate() != tt.buildDate {
				t.Errorf("buildDate = %q, want %q", GetBuildDate(), tt.buildDate)
			}

			// Verify by running version command
			var buf bytes.Buffer
			versionCmd.SetOut(&buf)
			versionCmd.Flags().Set("short", "false")
			versionCmd.Flags().Set("json", "false")

			runVersion(versionCmd, []string{})

			output := buf.String()

			if !strings.Contains(output, tt.version) {
				t.Errorf("output should contain version %q", tt.version)
			}
			if !strings.Contains(output, tt.commit) {
				t.Errorf("output should contain commit %q", tt.commit)
			}
			if !strings.Contains(output, tt.buildDate) {
				t.Errorf("output should contain buildDate %q", tt.buildDate)
			}
			if !strings.Contains(output, tt.builtBy) {
				t.Errorf("output should contain builtBy %q", tt.builtBy)
			}
		})
	}
}

// TestVersionOutputFormats tests all output formats work together
func TestVersionOutputFormats(t *testing.T) {
	// Set version info
	SetVersion("1.5.0")
	SetCommit("test123")
	SetBuildDate("2024-01-20")
	SetBuiltBy("test-runner")

	tests := []struct {
		name      string
		shortFlag bool
		jsonFlag  bool
		validate  func(*testing.T, string)
	}{
		{
			name:      "normal format",
			shortFlag: false,
			jsonFlag:  false,
			validate: func(t *testing.T, output string) {
				if !strings.Contains(output, "AINative Code") {
					t.Error("normal format should contain 'AINative Code'")
				}
				if !strings.Contains(output, "1.5.0") {
					t.Error("normal format should contain version")
				}
			},
		},
		{
			name:      "short format",
			shortFlag: true,
			jsonFlag:  false,
			validate: func(t *testing.T, output string) {
				output = strings.TrimSpace(output)
				if output != "1.5.0" {
					t.Errorf("short format should only output version, got %q", output)
				}
			},
		},
		{
			name:      "JSON format",
			shortFlag: false,
			jsonFlag:  true,
			validate: func(t *testing.T, output string) {
				var data map[string]string
				if err := json.Unmarshal([]byte(output), &data); err != nil {
					t.Errorf("JSON format should be valid JSON: %v", err)
				}
				if data["version"] != "1.5.0" {
					t.Error("JSON should contain version field")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			versionCmd.SetOut(&buf)

			versionCmd.Flags().Set("short", "false")
			versionCmd.Flags().Set("json", "false")

			if tt.shortFlag {
				versionCmd.Flags().Set("short", "true")
			}
			if tt.jsonFlag {
				versionCmd.Flags().Set("json", "true")
			}

			runVersion(versionCmd, []string{})

			tt.validate(t, buf.String())
		})
	}
}

// TestVersionDefaultValues tests default version values
func TestVersionDefaultValues(t *testing.T) {
	// Note: These are the package-level defaults
	// In a real build, these would be set by ldflags

	var buf bytes.Buffer
	versionCmd.SetOut(&buf)

	versionCmd.Flags().Set("short", "false")
	versionCmd.Flags().Set("json", "false")

	runVersion(versionCmd, []string{})

	output := buf.String()

	// Should contain some output
	if output == "" {
		t.Error("version output should not be empty")
	}

	// Should contain Go version
	if !strings.Contains(output, runtime.Version()) {
		t.Error("output should contain Go version")
	}

	// Should contain platform
	if !strings.Contains(output, runtime.GOOS) {
		t.Error("output should contain OS")
	}
	if !strings.Contains(output, runtime.GOARCH) {
		t.Error("output should contain architecture")
	}
}

// Benchmark tests for performance validation

// BenchmarkRunVersionNormal benchmarks normal version output
func BenchmarkRunVersionNormal(b *testing.B) {
	SetVersion("1.0.0")
	SetCommit("abc123")
	SetBuildDate("2024-01-01")

	var buf bytes.Buffer
	versionCmd.SetOut(&buf)

	versionCmd.Flags().Set("short", "false")
	versionCmd.Flags().Set("json", "false")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runVersion(versionCmd, []string{})
		buf.Reset()
	}
}

// BenchmarkRunVersionShort benchmarks short version output
func BenchmarkRunVersionShort(b *testing.B) {
	SetVersion("1.0.0")

	var buf bytes.Buffer
	versionCmd.SetOut(&buf)

	versionCmd.Flags().Set("short", "true")
	versionCmd.Flags().Set("json", "false")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runVersion(versionCmd, []string{})
		buf.Reset()
	}
}

// BenchmarkRunVersionJSON benchmarks JSON version output
func BenchmarkRunVersionJSON(b *testing.B) {
	SetVersion("1.0.0")
	SetCommit("abc123")
	SetBuildDate("2024-01-01")

	var buf bytes.Buffer
	versionCmd.SetOut(&buf)

	versionCmd.Flags().Set("short", "false")
	versionCmd.Flags().Set("json", "true")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runVersion(versionCmd, []string{})
		buf.Reset()
	}
}

// BenchmarkGetVersion benchmarks GetVersion function
func BenchmarkGetVersion(b *testing.B) {
	SetVersion("1.0.0")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetVersion()
	}
}

// BenchmarkGetCommit benchmarks GetCommit function
func BenchmarkGetCommit(b *testing.B) {
	SetCommit("abc123")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetCommit()
	}
}
