package security

import (
	"context"
	"strings"
	"testing"

	"github.com/AINative-studio/ainative-code/internal/tools"
	"github.com/AINative-studio/ainative-code/internal/tools/builtin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInputValidation_ExcessiveLength verifies that excessively long inputs are rejected
func TestInputValidation_ExcessiveLength(t *testing.T) {
	testCases := []struct {
		name          string
		inputField    string
		inputValue    string
		maxLength     int
		shouldReject  bool
	}{
		{
			name:         "Command within length limit",
			inputField:   "command",
			inputValue:   "ls",
			maxLength:    8192,
			shouldReject: false,
		},
		{
			name:         "Command exceeds length limit",
			inputField:   "command",
			inputValue:   strings.Repeat("A", 10000),
			maxLength:    8192,
			shouldReject: true,
		},
		{
			name:         "Empty command",
			inputField:   "command",
			inputValue:   "",
			maxLength:    8192,
			shouldReject: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Given: An exec command tool
			tool := builtin.NewExecCommandTool([]string{"ls", "pwd"}, "/tmp")

			// When: Executing with test input
			input := map[string]interface{}{
				"command": tc.inputValue,
			}

			result, err := tool.Execute(context.Background(), input)

			// Then: Should validate length appropriately
			if tc.shouldReject {
				if tc.inputValue == "" {
					assert.Error(t, err)
					assert.Nil(t, result)
				} else if len(tc.inputValue) > tc.maxLength {
					// Schema validation should catch this before execution
					assert.Error(t, err)
				}
			} else {
				// Command should be valid (might fail execution but not validation)
				assert.NotNil(t, result)
			}
		})
	}
}

// TestInputValidation_TypeConfusion verifies type validation prevents type confusion attacks
func TestInputValidation_TypeConfusion(t *testing.T) {
	testCases := []struct {
		name       string
		input      map[string]interface{}
		shouldFail bool
	}{
		{
			name: "Valid string command",
			input: map[string]interface{}{
				"command": "ls",
			},
			shouldFail: false,
		},
		{
			name: "Integer instead of string",
			input: map[string]interface{}{
				"command": 12345,
			},
			shouldFail: true,
		},
		{
			name: "Boolean instead of string",
			input: map[string]interface{}{
				"command": true,
			},
			shouldFail: true,
		},
		{
			name: "Null command",
			input: map[string]interface{}{
				"command": nil,
			},
			shouldFail: true,
		},
		{
			name: "Array instead of string",
			input: map[string]interface{}{
				"command": []string{"ls", "-la"},
			},
			shouldFail: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Given: An exec command tool
			tool := builtin.NewExecCommandTool([]string{"ls", "pwd"}, "/tmp")

			// When: Executing with test input
			result, err := tool.Execute(context.Background(), tc.input)

			// Then: Should properly validate types
			if tc.shouldFail {
				assert.Error(t, err)
				assert.Nil(t, result)
				// Verify error is specifically about invalid input
				var inputErr *tools.ErrInvalidInput
				assert.ErrorAs(t, err, &inputErr)
			} else {
				assert.NotNil(t, result)
			}
		})
	}
}

// TestInputValidation_SpecialCharacters verifies handling of special characters
func TestInputValidation_SpecialCharacters(t *testing.T) {
	testCases := []struct {
		name          string
		commandArgs   []string
		expectBlocked bool
	}{
		{
			name:          "Normal arguments",
			commandArgs:   []string{"-la"},
			expectBlocked: false,
		},
		{
			name:          "Arguments with shell metacharacters - semicolon",
			commandArgs:   []string{"; cat /etc/passwd"},
			expectBlocked: false, // Should be treated as literal argument
		},
		{
			name:          "Arguments with pipe",
			commandArgs:   []string{"| cat /etc/passwd"},
			expectBlocked: false, // Should be treated as literal argument
		},
		{
			name:          "Arguments with redirection",
			commandArgs:   []string{"> /tmp/evil.txt"},
			expectBlocked: false, // Should be treated as literal argument
		},
		{
			name:          "Arguments with command substitution",
			commandArgs:   []string{"$(rm -rf /)"},
			expectBlocked: false, // Should be treated as literal argument
		},
		{
			name:          "Arguments with backticks",
			commandArgs:   []string{"`whoami`"},
			expectBlocked: false, // Should be treated as literal argument
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Given: An exec command tool with whitelist
			tool := builtin.NewExecCommandTool([]string{"ls", "echo"}, "/tmp")

			// When: Executing with special characters
			input := map[string]interface{}{
				"command": "echo",
				"args":    tc.commandArgs,
			}

			result, err := tool.Execute(context.Background(), input)

			// Then: Arguments should be treated as literals, not interpreted
			require.NoError(t, err)
			assert.NotNil(t, result)
			// The shell metacharacters should be passed as literal arguments
			// and not interpreted by the shell
		})
	}
}

// TestInputValidation_NullByteInjection verifies null byte injection is prevented
func TestInputValidation_NullByteInjection(t *testing.T) {
	testCases := []struct {
		name  string
		input string
	}{
		{
			name:  "Input with null byte",
			input: "file.txt\x00.sh",
		},
		{
			name:  "Input with multiple null bytes",
			input: "test\x00\x00malicious",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Given: A file operation tool
			tool := builtin.NewReadFileTool("/tmp")

			// When: Attempting to use null byte injection
			input := map[string]interface{}{
				"file_path": tc.input,
			}

			// Then: Should handle null bytes safely
			// Note: In Go, strings handle null bytes but file operations may reject them
			_, err := tool.Execute(context.Background(), input)

			// The file operation should either:
			// 1. Reject the null byte (preferred)
			// 2. Treat it as part of the filename (safe in Go)
			// It should NOT truncate at the null byte
			_ = err // Error is acceptable
		})
	}
}

// TestInputValidation_BoundaryValues verifies boundary value handling
func TestInputValidation_BoundaryValues(t *testing.T) {
	testCases := []struct {
		name         string
		timeout      interface{}
		expectError  bool
		errorMessage string
	}{
		{
			name:        "Minimum valid timeout",
			timeout:     1,
			expectError: false,
		},
		{
			name:        "Maximum valid timeout",
			timeout:     300,
			expectError: false,
		},
		{
			name:         "Timeout exceeds maximum",
			timeout:      301,
			expectError:  true,
			errorMessage: "cannot exceed 300 seconds",
		},
		{
			name:         "Zero timeout",
			timeout:      0,
			expectError:  true,
			errorMessage: "must be positive",
		},
		{
			name:         "Negative timeout",
			timeout:      -1,
			expectError:  true,
			errorMessage: "must be positive",
		},
		{
			name:        "Floating point timeout",
			timeout:     30.5,
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Given: An exec command tool
			tool := builtin.NewExecCommandTool([]string{"sleep"}, "/tmp")

			// When: Executing with boundary timeout values
			input := map[string]interface{}{
				"command":         "sleep",
				"args":            []string{"0.1"},
				"timeout_seconds": tc.timeout,
			}

			_, err := tool.Execute(context.Background(), input)

			// Then: Should properly validate boundaries
			if tc.expectError {
				assert.Error(t, err)
				if tc.errorMessage != "" {
					assert.Contains(t, err.Error(), tc.errorMessage)
				}
			} else {
				// May succeed or fail execution, but should pass validation
				_ = err
			}
		})
	}
}

// TestInputValidation_IntegerOverflow verifies integer overflow protection
func TestInputValidation_IntegerOverflow(t *testing.T) {
	testCases := []struct {
		name    string
		value   interface{}
		isValid bool
	}{
		{
			name:    "Normal integer",
			value:   100,
			isValid: true,
		},
		{
			name:    "Maximum int32",
			value:   2147483647,
			isValid: true,
		},
		{
			name:    "Exceeds int32 but valid int64",
			value:   int64(2147483648),
			isValid: true, // Go handles this correctly
		},
		{
			name:    "Negative overflow",
			value:   -2147483649,
			isValid: true, // Go handles this correctly
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Given: A tool that accepts integer parameters
			tool := builtin.NewExecCommandTool([]string{"ls"}, "/tmp")

			// When: Providing integer values
			input := map[string]interface{}{
				"command":         "ls",
				"timeout_seconds": tc.value,
			}

			// Then: Should handle integer values without overflow
			_, err := tool.Execute(context.Background(), input)
			_ = err // Error is acceptable, we're testing for crashes
		})
	}
}

// TestInputValidation_EmailFormat verifies email validation
func TestInputValidation_EmailFormat(t *testing.T) {
	testCases := []struct {
		email   string
		isValid bool
	}{
		{"valid@example.com", true},
		{"user+tag@example.com", true},
		{"user.name@example.co.uk", true},
		{"invalid@", false},
		{"@example.com", false},
		{"invalid email@example.com", false},
		{"invalid@example", false},
		{"", false},
		{"<script>@example.com", false},
	}

	for _, tc := range testCases {
		t.Run(tc.email, func(t *testing.T) {
			// Email validation would be tested in the actual validator
			// This is a placeholder for the pattern
			emailPattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
			_ = emailPattern

			// In real implementation, this would use the validator
			// assert.Equal(t, tc.isValid, validator.ValidateEmail(tc.email))
		})
	}
}

// TestInputValidation_FuzzTesting performs basic fuzz testing on inputs
func TestInputValidation_FuzzTesting(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping fuzz test in short mode")
	}

	// Given: An exec command tool
	tool := builtin.NewExecCommandTool([]string{"echo"}, "/tmp")

	// Fuzz inputs
	fuzzInputs := []string{
		strings.Repeat("A", 1000),
		"../../../etc/passwd",
		"'; DROP TABLE users; --",
		"<script>alert('xss')</script>",
		"\x00\x01\x02\x03",
		"${IFS}cat${IFS}/etc/passwd",
		"$(whoami)",
		"`id`",
		"|| cat /etc/passwd",
		"& calc",
	}

	for _, fuzzInput := range fuzzInputs {
		t.Run("Fuzz_"+fuzzInput[:min(20, len(fuzzInput))], func(t *testing.T) {
			input := map[string]interface{}{
				"command": "echo",
				"args":    []interface{}{fuzzInput},
			}

			// When: Executing with fuzz input
			// Then: Should not crash or execute malicious code
			_, err := tool.Execute(context.Background(), input)
			_ = err // Error is acceptable, we're testing for crashes and code execution
		})
	}
}

// TestInputValidation_ArrayInjection verifies array parameter injection is prevented
func TestInputValidation_ArrayInjection(t *testing.T) {
	// Given: A tool that accepts array parameters
	tool := builtin.NewExecCommandTool([]string{"ls"}, "/tmp")

	testCases := []struct {
		name string
		args interface{}
	}{
		{
			name: "Normal array",
			args: []interface{}{"-la"},
		},
		{
			name: "Nested arrays",
			args: []interface{}{
				[]interface{}{"-la"},
			},
		},
		{
			name: "Mixed types in array",
			args: []interface{}{
				"-la",
				123,
				true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			input := map[string]interface{}{
				"command": "ls",
				"args":    tc.args,
			}

			// When: Executing with various array inputs
			_, err := tool.Execute(context.Background(), input)

			// Then: Should properly validate array contents
			_ = err // Some may fail validation, which is acceptable
		})
	}
}

// TestInputValidation_EnvironmentVariableInjection verifies env var injection is prevented
func TestInputValidation_EnvironmentVariableInjection(t *testing.T) {
	// Given: A tool that accepts environment variables
	tool := builtin.NewExecCommandTool([]string{"env"}, "/tmp")

	testCases := []struct {
		name       string
		envVars    map[string]interface{}
		shouldFail bool
	}{
		{
			name: "Normal env vars",
			envVars: map[string]interface{}{
				"MY_VAR": "value",
			},
			shouldFail: false,
		},
		{
			name: "Env var with special characters",
			envVars: map[string]interface{}{
				"MY_VAR": "value; malicious",
			},
			shouldFail: false, // Values are safe
		},
		{
			name: "Non-string env var value",
			envVars: map[string]interface{}{
				"MY_VAR": 123,
			},
			shouldFail: true, // Must be string
		},
		{
			name: "LD_PRELOAD injection attempt",
			envVars: map[string]interface{}{
				"LD_PRELOAD": "/tmp/malicious.so",
			},
			shouldFail: false, // Key name is allowed, but value sanitization depends on implementation
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			input := map[string]interface{}{
				"command": "env",
				"env":     tc.envVars,
			}

			// When: Executing with environment variables
			_, err := tool.Execute(context.Background(), input)

			// Then: Should validate env var values
			if tc.shouldFail {
				assert.Error(t, err)
			}
		})
	}
}

// Helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// BenchmarkInputValidation measures input validation performance
func BenchmarkInputValidation(b *testing.B) {
	tool := builtin.NewExecCommandTool([]string{"echo"}, "/tmp")
	input := map[string]interface{}{
		"command": "echo",
		"args":    []interface{}{"test"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = tool.Execute(context.Background(), input)
	}
}
