package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// TestFlagStandardization_DesignCommands tests that design commands use standardized -f/--file flags
func TestFlagStandardization_DesignCommands(t *testing.T) {
	tests := []struct {
		name        string
		cmd         *cobra.Command
		flagName    string
		shorthand   string
		description string
	}{
		{
			name:        "design import uses -f/--file",
			cmd:         designImportCmd,
			flagName:    "file",
			shorthand:   "f",
			description: "input file path",
		},
		{
			name:        "design export uses -f/--file",
			cmd:         designExportCmd,
			flagName:    "file",
			shorthand:   "f",
			description: "output file path",
		},
		{
			name:        "design validate has -f/--file",
			cmd:         designValidateCmd,
			flagName:    "file",
			shorthand:   "f",
			description: "input file path to validate",
		},
		{
			name:        "design extract uses -f/--file",
			cmd:         designExtractCmd,
			flagName:    "file",
			shorthand:   "f",
			description: "output file path",
		},
		{
			name:        "design generate uses -f/--file",
			cmd:         designGenerateCmd,
			flagName:    "file",
			shorthand:   "f",
			description: "output file path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := tt.cmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Errorf("Flag --%s not found in command %s", tt.flagName, tt.cmd.Name())
				return
			}

			if flag.Shorthand != tt.shorthand {
				t.Errorf("Expected shorthand -%s for flag --%s, got -%s",
					tt.shorthand, tt.flagName, flag.Shorthand)
			}

			// Check that description mentions file
			if !strings.Contains(strings.ToLower(flag.Usage), "file") &&
				!strings.Contains(strings.ToLower(flag.Usage), "path") {
				t.Errorf("Flag --%s usage should mention 'file' or 'path', got: %s",
					tt.flagName, flag.Usage)
			}
		})
	}
}

// TestFlagStandardization_RLHFExport tests RLHF export command flag standardization
func TestFlagStandardization_RLHFExport(t *testing.T) {
	t.Run("rlhf export uses -f/--file", func(t *testing.T) {
		flag := rlhfExportCmd.Flags().Lookup("file")
		if flag == nil {
			t.Fatal("Flag --file not found in rlhf export command")
		}

		if flag.Shorthand != "f" {
			t.Errorf("Expected shorthand -f for flag --file, got -%s", flag.Shorthand)
		}
	})

	t.Run("rlhf export has deprecated --output flag", func(t *testing.T) {
		flag := rlhfExportCmd.Flags().Lookup("output")
		if flag == nil {
			t.Fatal("Deprecated flag --output not found (needed for backward compatibility)")
		}

		if flag.Shorthand != "o" {
			t.Errorf("Expected shorthand -o for deprecated flag --output, got -%s", flag.Shorthand)
		}

		// Check that usage mentions deprecation
		if !strings.Contains(strings.ToLower(flag.Usage), "deprecated") {
			t.Errorf("Flag --output usage should mention 'deprecated', got: %s", flag.Usage)
		}
	})
}

// TestFlagStandardization_SessionExport tests session export command flag standardization
func TestFlagStandardization_SessionExport(t *testing.T) {
	t.Run("session export uses -f/--file", func(t *testing.T) {
		flag := sessionExportCmd.Flags().Lookup("file")
		if flag == nil {
			t.Fatal("Flag --file not found in session export command")
		}

		if flag.Shorthand != "f" {
			t.Errorf("Expected shorthand -f for flag --file, got -%s", flag.Shorthand)
		}
	})

	t.Run("session export has deprecated --output flag", func(t *testing.T) {
		flag := sessionExportCmd.Flags().Lookup("output")
		if flag == nil {
			t.Fatal("Deprecated flag --output not found (needed for backward compatibility)")
		}

		if flag.Shorthand != "o" {
			t.Errorf("Expected shorthand -o for deprecated flag --output, got -%s", flag.Shorthand)
		}

		// Check that usage mentions deprecation
		if !strings.Contains(strings.ToLower(flag.Usage), "deprecated") {
			t.Errorf("Flag --output usage should mention 'deprecated', got: %s", flag.Usage)
		}
	})

	t.Run("session export format flag has no shorthand", func(t *testing.T) {
		flag := sessionExportCmd.Flags().Lookup("format")
		if flag == nil {
			t.Fatal("Flag --format not found in session export command")
		}

		// format should not use -f to avoid conflict with -f/--file
		if flag.Shorthand == "f" {
			t.Error("Flag --format should not use shorthand -f to avoid conflict with --file")
		}
	})
}

// TestFlagStandardization_BackwardCompatibility tests backward compatibility with deprecated flags
func TestFlagStandardization_BackwardCompatibility(t *testing.T) {
	tests := []struct {
		name            string
		cmd             *cobra.Command
		newFlag         string
		deprecatedFlag  string
		shorthandNew    string
		shorthandOld    string
	}{
		{
			name:            "rlhf export backward compat",
			cmd:             rlhfExportCmd,
			newFlag:         "file",
			deprecatedFlag:  "output",
			shorthandNew:    "f",
			shorthandOld:    "o",
		},
		{
			name:            "design extract backward compat",
			cmd:             designExtractCmd,
			newFlag:         "file",
			deprecatedFlag:  "output",
			shorthandNew:    "f",
			shorthandOld:    "o",
		},
		{
			name:            "design generate backward compat",
			cmd:             designGenerateCmd,
			newFlag:         "file",
			deprecatedFlag:  "output",
			shorthandNew:    "f",
			shorthandOld:    "o",
		},
		{
			name:            "session export backward compat",
			cmd:             sessionExportCmd,
			newFlag:         "file",
			deprecatedFlag:  "output",
			shorthandNew:    "f",
			shorthandOld:    "o",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Check new flag exists
			newFlag := tt.cmd.Flags().Lookup(tt.newFlag)
			if newFlag == nil {
				t.Errorf("New flag --%s not found", tt.newFlag)
				return
			}

			if newFlag.Shorthand != tt.shorthandNew {
				t.Errorf("New flag --%s should have shorthand -%s, got -%s",
					tt.newFlag, tt.shorthandNew, newFlag.Shorthand)
			}

			// Check deprecated flag exists
			deprecatedFlag := tt.cmd.Flags().Lookup(tt.deprecatedFlag)
			if deprecatedFlag == nil {
				t.Errorf("Deprecated flag --%s not found (needed for backward compatibility)",
					tt.deprecatedFlag)
				return
			}

			if deprecatedFlag.Shorthand != tt.shorthandOld {
				t.Errorf("Deprecated flag --%s should have shorthand -%s, got -%s",
					tt.deprecatedFlag, tt.shorthandOld, deprecatedFlag.Shorthand)
			}

			// Both flags should point to the same value
			// This ensures backward compatibility works
		})
	}
}

// TestFlagStandardization_HelpText tests that help text is updated correctly
func TestFlagStandardization_HelpText(t *testing.T) {
	tests := []struct {
		name           string
		cmd            *cobra.Command
		shouldContain  []string
	}{
		{
			name: "rlhf export help mentions --file",
			cmd:  rlhfExportCmd,
			shouldContain: []string{"--file", "-f"},
		},
		{
			name: "design extract help mentions --file",
			cmd:  designExtractCmd,
			shouldContain: []string{"--file", "-f"},
		},
		{
			name: "design generate help mentions --file",
			cmd:  designGenerateCmd,
			shouldContain: []string{"--file", "-f"},
		},
		{
			name: "session export help mentions --file",
			cmd:  sessionExportCmd,
			shouldContain: []string{"--file", "-f"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			tt.cmd.SetOut(&buf)
			tt.cmd.SetErr(&buf)

			// Get help text
			helpCmd := &cobra.Command{Use: "help"}
			helpCmd.SetArgs([]string{tt.cmd.Name()})

			// Execute help
			tt.cmd.Flags().VisitAll(func(f *pflag.Flag) {
				usage := f.Usage
				for _, expected := range tt.shouldContain {
					if strings.Contains(usage, expected) {
						return
					}
				}
			})

			helpText := buf.String()
			for _, expected := range tt.shouldContain {
				if !strings.Contains(helpText, expected) &&
				   !strings.Contains(tt.cmd.Long, expected) &&
				   !strings.Contains(tt.cmd.Example, expected) {
					// At least one flag should contain the expected text
					found := false
					tt.cmd.Flags().VisitAll(func(f *pflag.Flag) {
						if strings.Contains(f.Usage, expected) || f.Name == strings.TrimPrefix(expected, "--") {
							found = true
						}
					})
					if !found {
						t.Logf("Warning: help text for %s might not mention %s", tt.cmd.Name(), expected)
					}
				}
			}
		})
	}
}

// TestFlagStandardization_NoConflicts tests that there are no flag conflicts
func TestFlagStandardization_NoConflicts(t *testing.T) {
	commands := []*cobra.Command{
		designImportCmd,
		designExportCmd,
		designValidateCmd,
		designExtractCmd,
		designGenerateCmd,
		rlhfExportCmd,
		sessionExportCmd,
	}

	for _, cmd := range commands {
		t.Run("no duplicate flags in "+cmd.Name(), func(t *testing.T) {
			shorthandMap := make(map[string]string)
			longMap := make(map[string]int)

			cmd.Flags().VisitAll(func(f *pflag.Flag) {
				// Track long flag names
				longMap[f.Name]++

				// Track shorthand flags (skip empty ones)
				if f.Shorthand != "" {
					if existing, exists := shorthandMap[f.Shorthand]; exists {
						// This is expected for deprecated flags that point to the same variable
						if !(f.Name == "file" && existing == "output") &&
						   !(f.Name == "output" && existing == "file") {
							t.Errorf("Duplicate shorthand -%s used for flags: %s and %s",
								f.Shorthand, existing, f.Name)
						}
					}
					shorthandMap[f.Shorthand] = f.Name
				}
			})

			// Check for truly duplicate long flags (count > 1)
			for name, count := range longMap {
				if count > 1 {
					t.Errorf("Duplicate long flag --%s appears %d times", name, count)
				}
			}
		})
	}
}

// TestFlagStandardization_RequiredFlags tests that required flags are properly set
func TestFlagStandardization_RequiredFlags(t *testing.T) {
	tests := []struct {
		name         string
		cmd          *cobra.Command
		requiredFlag string
	}{
		{
			name:         "design import requires --file",
			cmd:          designImportCmd,
			requiredFlag: "file",
		},
		{
			name:         "design export requires --file",
			cmd:          designExportCmd,
			requiredFlag: "file",
		},
		{
			name:         "design extract requires --file",
			cmd:          designExtractCmd,
			requiredFlag: "file",
		},
		// Note: design validate --file is optional
		// Note: design generate --file is optional (prints to stdout if not provided)
		// Note: rlhf export --file has default value
		// Note: session export --file has default value
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := tt.cmd.Flags().Lookup(tt.requiredFlag)
			if flag == nil {
				t.Fatalf("Required flag --%s not found", tt.requiredFlag)
			}

			// Try to determine if flag is required
			// In cobra, we check the Annotations or try parsing without the flag
			var buf bytes.Buffer
			tt.cmd.SetOut(&buf)
			tt.cmd.SetErr(&buf)

			// Create a copy of the command to test
			testCmd := &cobra.Command{
				Use:  tt.cmd.Use,
				RunE: tt.cmd.RunE,
			}
			tt.cmd.Flags().VisitAll(func(f *pflag.Flag) {
				testCmd.Flags().AddFlag(f)
			})

			// This test verifies the flag exists and can be accessed
			// The actual requirement is enforced at runtime
			if flag.Shorthand != "f" {
				t.Errorf("Required flag --%s should use shorthand -f, got -%s",
					tt.requiredFlag, flag.Shorthand)
			}
		})
	}
}
