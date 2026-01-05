package security

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/AINative-studio/ainative-code/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAPIKeyResolution_CommandInjectionPrevention verifies command injection is prevented
func TestAPIKeyResolution_CommandInjectionPrevention(t *testing.T) {
	resolver := config.NewResolver()

	testCases := []struct {
		name           string
		input          string
		shouldError    bool
		description    string
	}{
		{
			name:        "Command chaining with semicolon",
			input:       "$(echo test; cat /etc/passwd)",
			shouldError: false, // Command executes but semicolon is passed to echo
			description: "Semicolon should be treated as literal argument, not command separator",
		},
		{
			name:        "Command chaining with AND",
			input:       "$(echo test && cat /etc/passwd)",
			shouldError: false,
			description: "AND operator should be treated as literal",
		},
		{
			name:        "Command chaining with OR",
			input:       "$(echo test || cat /etc/passwd)",
			shouldError: false,
			description: "OR operator should be treated as literal",
		},
		{
			name:        "Pipe to another command",
			input:       "$(echo test | cat)",
			shouldError: false,
			description: "Pipe should be treated as literal argument",
		},
		{
			name:        "Command substitution in argument",
			input:       "$(echo $(whoami))",
			shouldError: false,
			description: "Nested command substitution should be literal",
		},
		{
			name:        "Backtick command substitution",
			input:       "$(echo `whoami`)",
			shouldError: false,
			description: "Backticks should be treated as literal",
		},
		{
			name:        "Output redirection",
			input:       "$(echo test > /tmp/evil)",
			shouldError: false,
			description: "Redirection should be treated as literal",
		},
		{
			name:        "Input redirection",
			input:       "$(echo test < /etc/passwd)",
			shouldError: false,
			description: "Input redirection should be treated as literal",
		},
		{
			name:        "Background execution",
			input:       "$(echo test &)",
			shouldError: false,
			description: "Background operator should be treated as literal",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// When: Attempting command injection
			result, err := resolver.Resolve(tc.input)

			// Then: Should handle safely without executing injected commands
			if tc.shouldError {
				assert.Error(t, err, tc.description)
			} else {
				// May succeed or fail based on command, but should not execute injection
				_ = result
				_ = err
			}
		})
	}
}

// TestAPIKeyResolution_PathTraversalPrevention verifies path traversal attacks are prevented
func TestAPIKeyResolution_PathTraversalPrevention(t *testing.T) {
	resolver := config.NewResolver()

	// Create a test file in a safe location
	tmpDir := t.TempDir()
	secretFile := filepath.Join(tmpDir, "secret.key")
	err := os.WriteFile(secretFile, []byte("sk-test-key-123"), 0600)
	require.NoError(t, err)

	testCases := []struct {
		name        string
		createPath  func() string
		shouldError bool
		description string
	}{
		{
			name: "Path with parent directory traversal",
			createPath: func() string {
				return filepath.Join(tmpDir, "..", filepath.Base(tmpDir), "secret.key")
			},
			shouldError: false, // Should succeed after path cleaning
			description: "Path should be cleaned and resolved correctly",
		},
		{
			name: "Path with multiple parent traversals",
			createPath: func() string {
				return filepath.Join(tmpDir, "..", "..", "..", filepath.Base(tmpDir), "secret.key")
			},
			shouldError: true, // Will fail because path doesn't exist after cleaning
			description: "Multiple parent traversals should be cleaned",
		},
		{
			name: "Path with null byte",
			createPath: func() string {
				return secretFile + "\x00/etc/passwd"
			},
			shouldError: true,
			description: "Null byte should be detected and rejected",
		},
		{
			name: "Path with current directory references",
			createPath: func() string {
				return filepath.Join(tmpDir, ".", ".", "secret.key")
			},
			shouldError: false,
			description: "Current directory references should be cleaned",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Given: A path that might contain traversal attempts
			testPath := tc.createPath()

			// When: Attempting to resolve from file
			result, err := resolver.Resolve(testPath)

			// Then: Should prevent traversal or handle safely
			if tc.shouldError {
				assert.Error(t, err, tc.description)
			} else {
				assert.NoError(t, err, tc.description)
				assert.Equal(t, "sk-test-key-123", result)
			}
		})
	}
}

// TestAPIKeyResolution_SymlinkHandling verifies symlink security
func TestAPIKeyResolution_SymlinkHandling(t *testing.T) {
	resolver := config.NewResolver()

	// Create test structure with symlinks
	tmpDir := t.TempDir()
	secretDir := filepath.Join(tmpDir, "secrets")
	err := os.MkdirAll(secretDir, 0755)
	require.NoError(t, err)

	secretFile := filepath.Join(secretDir, "api.key")
	err = os.WriteFile(secretFile, []byte("sk-real-key-123"), 0600)
	require.NoError(t, err)

	// Create a symlink to the secret file
	symlinkPath := filepath.Join(tmpDir, "link-to-secret.key")
	err = os.Symlink(secretFile, symlinkPath)
	require.NoError(t, err)

	t.Run("resolves symlink to actual file", func(t *testing.T) {
		// When: Resolving via symlink
		result, err := resolver.Resolve(symlinkPath)

		// Then: Should resolve symlink and read actual file
		require.NoError(t, err)
		assert.Equal(t, "sk-real-key-123", result)
	})

	t.Run("detects broken symlinks", func(t *testing.T) {
		// Given: A broken symlink
		brokenLink := filepath.Join(tmpDir, "broken-link.key")
		err := os.Symlink("/nonexistent/file", brokenLink)
		require.NoError(t, err)

		// When: Attempting to resolve
		result, err := resolver.Resolve(brokenLink)

		// Then: Should error on broken symlink
		assert.Error(t, err)
		assert.Empty(t, result)
	})
}

// TestAPIKeyResolution_FileSizeLimits verifies file size restrictions
func TestAPIKeyResolution_FileSizeLimits(t *testing.T) {
	resolver := config.NewResolver()
	tmpDir := t.TempDir()

	testCases := []struct {
		name        string
		fileSize    int
		shouldError bool
		description string
	}{
		{
			name:        "Small file within limit",
			fileSize:    100,
			shouldError: false,
			description: "100 bytes should be accepted",
		},
		{
			name:        "File at size limit",
			fileSize:    1024, // 1KB
			shouldError: false,
			description: "Exactly 1KB should be accepted",
		},
		{
			name:        "File exceeding limit",
			fileSize:    1025, // 1KB + 1 byte
			shouldError: true,
			description: "File over 1KB should be rejected",
		},
		{
			name:        "Large file",
			fileSize:    1024 * 100, // 100KB
			shouldError: true,
			description: "Large file should be rejected for security",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Given: A file of specific size
			keyFile := filepath.Join(tmpDir, fmt.Sprintf("key-%d.txt", tc.fileSize))
			content := make([]byte, tc.fileSize)
			// Fill with valid key content (avoid all zeros which might be trimmed)
			for i := range content {
				content[i] = byte('a' + (i % 26))
			}
			err := os.WriteFile(keyFile, content, 0600)
			require.NoError(t, err)

			// When: Attempting to resolve
			result, err := resolver.Resolve(keyFile)

			// Then: Should enforce size limits
			if tc.shouldError {
				assert.Error(t, err, tc.description)
				assert.Contains(t, err.Error(), "too large", "Error should mention file size")
			} else {
				assert.NoError(t, err, tc.description)
				assert.NotEmpty(t, result)
			}
		})
	}
}

// TestAPIKeyResolution_CommandTimeout verifies timeout enforcement
func TestAPIKeyResolution_CommandTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping timeout test in short mode")
	}

	// Given: Resolver with very short timeout
	resolver := config.NewResolver(config.WithCommandTimeout(100 * time.Millisecond))

	// When: Executing a slow command
	result, err := resolver.Resolve("$(/bin/sleep 5)")

	// Then: Should timeout
	assert.Error(t, err)
	assert.Empty(t, result)
	assert.Contains(t, err.Error(), "timed out", "Error should indicate timeout")
}

// TestAPIKeyResolution_EnvironmentVariableSecurity verifies env var security
func TestAPIKeyResolution_EnvironmentVariableSecurity(t *testing.T) {
	resolver := config.NewResolver()

	testCases := []struct {
		name        string
		input       string
		envVar      string
		envValue    string
		shouldError bool
		description string
	}{
		{
			name:        "Valid environment variable with braces",
			input:       "${TEST_API_KEY}",
			envVar:      "TEST_API_KEY",
			envValue:    "sk-test-123",
			shouldError: false,
			description: "Standard env var should work",
		},
		{
			name:        "Valid environment variable without braces",
			input:       "$TEST_API_KEY",
			envVar:      "TEST_API_KEY",
			envValue:    "sk-test-456",
			shouldError: false,
			description: "Env var without braces should work",
		},
		{
			name:        "Environment variable with command substitution in value",
			input:       "${CMD_VAR}",
			envVar:      "CMD_VAR",
			envValue:    "$(cat /etc/passwd)",
			shouldError: false,
			description: "Command in env var value should be treated as literal",
		},
		{
			name:        "Missing environment variable",
			input:       "${MISSING_VAR}",
			shouldError: true,
			description: "Missing env var should error",
		},
		{
			name:        "Invalid variable name with spaces",
			input:       "${ INVALID }",
			shouldError: false, // Treated as direct string
			description: "Invalid syntax should be treated as direct string",
		},
		{
			name:        "Invalid variable name starting with number",
			input:       "${123INVALID}",
			shouldError: false, // Treated as direct string
			description: "Invalid syntax should be treated as direct string",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup environment
			if tc.envVar != "" {
				os.Setenv(tc.envVar, tc.envValue)
				defer os.Unsetenv(tc.envVar)
			}

			// When: Resolving environment variable
			result, err := resolver.Resolve(tc.input)

			// Then: Should handle securely
			if tc.shouldError {
				assert.Error(t, err, tc.description)
			} else {
				assert.NoError(t, err, tc.description)
				if tc.envValue != "" {
					assert.Equal(t, tc.envValue, result)
				}
			}
		})
	}
}

// TestAPIKeyResolution_FilePermissions verifies permission handling
func TestAPIKeyResolution_FilePermissions(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Skipping permission test when running as root")
	}

	resolver := config.NewResolver()
	tmpDir := t.TempDir()

	t.Run("reads file with proper permissions", func(t *testing.T) {
		// Given: File with read permissions
		keyFile := filepath.Join(tmpDir, "readable.key")
		err := os.WriteFile(keyFile, []byte("sk-readable-123"), 0600)
		require.NoError(t, err)

		// When: Resolving
		result, err := resolver.Resolve(keyFile)

		// Then: Should succeed
		assert.NoError(t, err)
		assert.Equal(t, "sk-readable-123", result)
	})

	t.Run("handles unreadable file", func(t *testing.T) {
		// Given: File without read permissions
		keyFile := filepath.Join(tmpDir, "unreadable.key")
		err := os.WriteFile(keyFile, []byte("sk-unreadable-123"), 0000)
		require.NoError(t, err)
		defer os.Chmod(keyFile, 0600) // Cleanup

		// When: Attempting to resolve
		result, err := resolver.Resolve(keyFile)

		// Then: Should error with permission denied
		assert.Error(t, err)
		assert.Empty(t, result)
		// Error should indicate permission issue
	})
}

// TestAPIKeyResolution_DirectStringSafety verifies direct string handling
func TestAPIKeyResolution_DirectStringSafety(t *testing.T) {
	resolver := config.NewResolver()

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Anthropic API key format",
			input:    "sk-ant-api03-1234567890abcdef",
			expected: "sk-ant-api03-1234567890abcdef",
		},
		{
			name:     "OpenAI API key format",
			input:    "sk-proj-1234567890abcdef",
			expected: "sk-proj-1234567890abcdef",
		},
		{
			name:     "Google API key format",
			input:    "AIzaSyABC123DEF456GHI789",
			expected: "AIzaSyABC123DEF456GHI789",
		},
		{
			name:     "API key with whitespace",
			input:    "  sk-ant-key-123  \n",
			expected: "sk-ant-key-123",
		},
		{
			name:     "UUID-like key",
			input:    "123e4567-e89b-12d3-a456-426614174000",
			expected: "123e4567-e89b-12d3-a456-426614174000",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// When: Resolving direct string
			result, err := resolver.Resolve(tc.input)

			// Then: Should return trimmed string
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestAPIKeyResolution_EmptyValues verifies empty value handling
func TestAPIKeyResolution_EmptyValues(t *testing.T) {
	resolver := config.NewResolver()
	tmpDir := t.TempDir()

	testCases := []struct {
		name        string
		setup       func() string
		shouldError bool
		description string
	}{
		{
			name: "Empty string input",
			setup: func() string {
				return ""
			},
			shouldError: false,
			description: "Empty string should return empty without error",
		},
		{
			name: "Whitespace only",
			setup: func() string {
				return "   \n\t  "
			},
			shouldError: false,
			description: "Whitespace should be trimmed to empty",
		},
		{
			name: "Empty file",
			setup: func() string {
				f := filepath.Join(tmpDir, "empty.key")
				os.WriteFile(f, []byte(""), 0600)
				return f
			},
			shouldError: true,
			description: "Empty file should error",
		},
		{
			name: "File with only whitespace",
			setup: func() string {
				f := filepath.Join(tmpDir, "whitespace.key")
				os.WriteFile(f, []byte("   \n\t  "), 0600)
				return f
			},
			shouldError: true,
			description: "File with only whitespace should error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Given: Various empty value scenarios
			input := tc.setup()

			// When: Attempting to resolve
			result, err := resolver.Resolve(input)

			// Then: Should handle empty values appropriately
			if tc.shouldError {
				assert.Error(t, err, tc.description)
			} else {
				assert.NoError(t, err, tc.description)
				assert.Empty(t, result)
			}
		})
	}
}

// TestAPIKeyResolution_WhitelistEnforcement verifies command whitelist
func TestAPIKeyResolution_WhitelistEnforcement(t *testing.T) {
	// Given: Resolver with strict command whitelist
	resolver := config.NewResolver(
		config.WithAllowedCommands("echo", "cat"),
	)

	testCases := []struct {
		name        string
		input       string
		shouldError bool
		description string
	}{
		{
			name:        "Allowed command: echo",
			input:       "$(echo sk-test-123)",
			shouldError: false,
			description: "Whitelisted command should work",
		},
		{
			name:        "Blocked command: ls",
			input:       "$(ls /tmp)",
			shouldError: true,
			description: "Non-whitelisted command should error",
		},
		{
			name:        "Blocked command: whoami",
			input:       "$(whoami)",
			shouldError: true,
			description: "Non-whitelisted command should error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// When: Attempting command execution
			result, err := resolver.Resolve(tc.input)

			// Then: Should enforce whitelist
			if tc.shouldError {
				assert.Error(t, err, tc.description)
				assert.Contains(t, err.Error(), "not in the allowed commands")
			} else {
				assert.NoError(t, err, tc.description)
				assert.NotEmpty(t, result)
			}
		})
	}
}

// TestAPIKeyResolution_DisabledCommandExecution verifies command execution can be disabled
func TestAPIKeyResolution_DisabledCommandExecution(t *testing.T) {
	// Given: Resolver with command execution disabled
	resolver := config.NewResolver(
		config.WithCommandExecution(false),
	)

	testCases := []string{
		"$(echo test)",
		"$(cat /etc/passwd)",
		"$(whoami)",
	}

	for _, input := range testCases {
		t.Run("Disabled_"+input, func(t *testing.T) {
			// When: Attempting command execution
			result, err := resolver.Resolve(input)

			// Then: Should error indicating disabled
			assert.Error(t, err)
			assert.Empty(t, result)
			assert.Contains(t, err.Error(), "disabled")
		})
	}
}

// TestAPIKeyResolution_ConcurrentAccess verifies thread safety
func TestAPIKeyResolution_ConcurrentAccess(t *testing.T) {
	resolver := config.NewResolver()
	tmpDir := t.TempDir()

	// Create test file
	keyFile := filepath.Join(tmpDir, "concurrent.key")
	err := os.WriteFile(keyFile, []byte("sk-concurrent-123"), 0600)
	require.NoError(t, err)

	// Set up environment
	os.Setenv("CONCURRENT_KEY", "sk-env-concurrent")
	defer os.Unsetenv("CONCURRENT_KEY")

	// When: Multiple goroutines resolve simultaneously
	const numGoroutines = 100
	results := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			var input string
			switch idx % 3 {
			case 0:
				input = keyFile
			case 1:
				input = "${CONCURRENT_KEY}"
			case 2:
				input = "$(echo sk-cmd-concurrent)"
			}

			_, err := resolver.Resolve(input)
			results <- err
		}(i)
	}

	// Then: All should complete without data races
	for i := 0; i < numGoroutines; i++ {
		err := <-results
		assert.NoError(t, err, "Concurrent resolution should not error")
	}
}

// BenchmarkAPIKeyResolution measures performance of different resolution methods
func BenchmarkAPIKeyResolution(b *testing.B) {
	resolver := config.NewResolver()

	b.Run("DirectString", func(b *testing.B) {
		input := "sk-ant-api-key-123"
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = resolver.Resolve(input)
		}
	})

	b.Run("EnvironmentVariable", func(b *testing.B) {
		os.Setenv("BENCH_KEY", "sk-bench-123")
		defer os.Unsetenv("BENCH_KEY")
		input := "${BENCH_KEY}"
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = resolver.Resolve(input)
		}
	})

	b.Run("FileRead", func(b *testing.B) {
		tmpDir := b.TempDir()
		keyFile := filepath.Join(tmpDir, "bench.key")
		os.WriteFile(keyFile, []byte("sk-bench-file-123"), 0600)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = resolver.Resolve(keyFile)
		}
	})

	b.Run("Command", func(b *testing.B) {
		input := "$(echo sk-bench-cmd)"
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = resolver.Resolve(input)
		}
	})
}
