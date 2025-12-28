package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewResolver(t *testing.T) {
	t.Run("creates resolver with defaults", func(t *testing.T) {
		resolver := NewResolver()

		assert.NotNil(t, resolver)
		assert.Equal(t, 5*time.Second, resolver.commandTimeout)
		assert.Empty(t, resolver.allowedCommands)
		assert.True(t, resolver.enableCommandExecution)
	})

	t.Run("creates resolver with custom options", func(t *testing.T) {
		resolver := NewResolver(
			WithCommandTimeout(10*time.Second),
			WithAllowedCommands("echo", "cat"),
			WithCommandExecution(false),
		)

		assert.Equal(t, 10*time.Second, resolver.commandTimeout)
		assert.Equal(t, []string{"echo", "cat"}, resolver.allowedCommands)
		assert.False(t, resolver.enableCommandExecution)
	})
}

func TestResolver_ResolveDirectString(t *testing.T) {
	resolver := NewResolver()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple API key",
			input:    "sk-ant-api-key-123",
			expected: "sk-ant-api-key-123",
		},
		{
			name:     "OpenAI API key",
			input:    "sk-proj-abc123def456",
			expected: "sk-proj-abc123def456",
		},
		{
			name:     "Google API key",
			input:    "AIzaSyABC123DEF456",
			expected: "AIzaSyABC123DEF456",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "string with whitespace",
			input:    "  sk-ant-key-123  ",
			expected: "sk-ant-key-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := resolver.Resolve(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestResolver_ResolveEnvVar(t *testing.T) {
	resolver := NewResolver()

	t.Run("resolves existing environment variable", func(t *testing.T) {
		// Set test environment variable
		os.Setenv("TEST_API_KEY", "sk-test-key-123")
		defer os.Unsetenv("TEST_API_KEY")

		result, err := resolver.Resolve("${TEST_API_KEY}")
		require.NoError(t, err)
		assert.Equal(t, "sk-test-key-123", result)
	})

	t.Run("resolves environment variable with underscores", func(t *testing.T) {
		os.Setenv("OPENAI_API_KEY", "sk-openai-key-456")
		defer os.Unsetenv("OPENAI_API_KEY")

		result, err := resolver.Resolve("${OPENAI_API_KEY}")
		require.NoError(t, err)
		assert.Equal(t, "sk-openai-key-456", result)
	})

	t.Run("errors on missing environment variable", func(t *testing.T) {
		result, err := resolver.Resolve("${NONEXISTENT_VAR}")
		assert.Error(t, err)
		assert.Empty(t, result)
		assert.Contains(t, err.Error(), "NONEXISTENT_VAR")
	})

	t.Run("invalid environment variable name", func(t *testing.T) {
		// Invalid names should be treated as direct strings
		result, err := resolver.Resolve("${123INVALID}")
		require.NoError(t, err)
		assert.Equal(t, "${123INVALID}", result)
	})

	t.Run("environment variable with spaces is treated as direct string", func(t *testing.T) {
		result, err := resolver.Resolve("${ INVALID VAR }")
		require.NoError(t, err)
		assert.Equal(t, "${ INVALID VAR }", result)
	})
}

func TestResolver_ResolveFilePath(t *testing.T) {
	resolver := NewResolver()

	t.Run("resolves file with absolute path", func(t *testing.T) {
		// Create temporary file
		tmpDir := t.TempDir()
		keyFile := filepath.Join(tmpDir, "api-key.txt")
		err := os.WriteFile(keyFile, []byte("sk-file-key-123"), 0600)
		require.NoError(t, err)

		result, err := resolver.Resolve(keyFile)
		require.NoError(t, err)
		assert.Equal(t, "sk-file-key-123", result)
	})

	t.Run("resolves file with relative path", func(t *testing.T) {
		tmpDir := t.TempDir()
		keyFile := filepath.Join(tmpDir, "secrets", "key.txt")
		err := os.MkdirAll(filepath.Dir(keyFile), 0755)
		require.NoError(t, err)
		err = os.WriteFile(keyFile, []byte("sk-relative-key-456"), 0600)
		require.NoError(t, err)

		result, err := resolver.Resolve(keyFile)
		require.NoError(t, err)
		assert.Equal(t, "sk-relative-key-456", result)
	})

	t.Run("resolves file with home directory", func(t *testing.T) {
		// Create temporary home directory structure
		tmpDir := t.TempDir()
		keyFile := filepath.Join(tmpDir, ".secrets", "api.key")
		err := os.MkdirAll(filepath.Dir(keyFile), 0755)
		require.NoError(t, err)
		err = os.WriteFile(keyFile, []byte("sk-home-key-789"), 0600)
		require.NoError(t, err)

		// Mock home directory
		originalHome := os.Getenv("HOME")
		os.Setenv("HOME", tmpDir)
		defer os.Setenv("HOME", originalHome)

		result, err := resolver.Resolve("~/.secrets/api.key")
		require.NoError(t, err)
		assert.Equal(t, "sk-home-key-789", result)
	})

	t.Run("trims whitespace from file content", func(t *testing.T) {
		tmpDir := t.TempDir()
		keyFile := filepath.Join(tmpDir, "key.txt")
		err := os.WriteFile(keyFile, []byte("\n  sk-whitespace-key  \n"), 0600)
		require.NoError(t, err)

		result, err := resolver.Resolve(keyFile)
		require.NoError(t, err)
		assert.Equal(t, "sk-whitespace-key", result)
	})

	t.Run("errors on nonexistent file", func(t *testing.T) {
		result, err := resolver.Resolve("/nonexistent/path/to/key.txt")
		assert.Error(t, err)
		assert.Empty(t, result)
		assert.Contains(t, err.Error(), "does not exist")
	})

	t.Run("errors on directory instead of file", func(t *testing.T) {
		tmpDir := t.TempDir()
		result, err := resolver.Resolve(tmpDir)
		assert.Error(t, err)
		assert.Empty(t, result)
		assert.Contains(t, err.Error(), "directory")
	})

	t.Run("errors on empty file", func(t *testing.T) {
		tmpDir := t.TempDir()
		keyFile := filepath.Join(tmpDir, "empty.txt")
		err := os.WriteFile(keyFile, []byte(""), 0600)
		require.NoError(t, err)

		result, err := resolver.Resolve(keyFile)
		assert.Error(t, err)
		assert.Empty(t, result)
		assert.Contains(t, err.Error(), "empty")
	})

	t.Run("errors on file too large", func(t *testing.T) {
		tmpDir := t.TempDir()
		keyFile := filepath.Join(tmpDir, "large.txt")

		// Create a file larger than 1MB
		largeContent := make([]byte, 2*1024*1024)
		err := os.WriteFile(keyFile, largeContent, 0600)
		require.NoError(t, err)

		result, err := resolver.Resolve(keyFile)
		assert.Error(t, err)
		assert.Empty(t, result)
		assert.Contains(t, err.Error(), "too large")
	})

	t.Run("recognizes common key file extensions", func(t *testing.T) {
		tmpDir := t.TempDir()
		extensions := []string{".txt", ".key", ".pem", ".secret", ".env"}

		for _, ext := range extensions {
			keyFile := filepath.Join(tmpDir, "api"+ext)
			err := os.WriteFile(keyFile, []byte("sk-ext-key-"+ext), 0600)
			require.NoError(t, err)

			result, err := resolver.Resolve(keyFile)
			require.NoError(t, err, "failed for extension %s", ext)
			assert.Equal(t, "sk-ext-key-"+ext, result)
		}
	})
}

func TestResolver_ResolveCommand(t *testing.T) {
	resolver := NewResolver()

	t.Run("executes simple command", func(t *testing.T) {
		result, err := resolver.Resolve("$(/bin/echo sk-cmd-key-123)")
		require.NoError(t, err)
		assert.Equal(t, "sk-cmd-key-123", result)
	})

	t.Run("executes command with arguments", func(t *testing.T) {
		// Create temporary file for testing
		tmpDir := t.TempDir()
		keyFile := filepath.Join(tmpDir, "key.txt")
		err := os.WriteFile(keyFile, []byte("sk-cat-key-456"), 0600)
		require.NoError(t, err)

		result, err := resolver.Resolve(fmt.Sprintf("$(cat %s)", keyFile))
		require.NoError(t, err)
		assert.Equal(t, "sk-cat-key-456", result)
	})

	t.Run("trims whitespace from command output", func(t *testing.T) {
		result, err := resolver.Resolve("$(/bin/echo   sk-trim-key  )")
		require.NoError(t, err)
		assert.Equal(t, "sk-trim-key", result)
	})

	t.Run("errors when command execution is disabled", func(t *testing.T) {
		disabledResolver := NewResolver(WithCommandExecution(false))
		result, err := disabledResolver.Resolve("$(/bin/echo test)")
		assert.Error(t, err)
		assert.Empty(t, result)
		assert.Contains(t, err.Error(), "disabled")
	})

	t.Run("errors on empty command", func(t *testing.T) {
		result, err := resolver.Resolve("$()")
		assert.Error(t, err)
		assert.Empty(t, result)
		assert.Contains(t, err.Error(), "empty")
	})

	t.Run("errors on command not in whitelist", func(t *testing.T) {
		restrictedResolver := NewResolver(WithAllowedCommands("cat", "grep"))
		result, err := restrictedResolver.Resolve("$(/bin/echo test)")
		assert.Error(t, err)
		assert.Empty(t, result)
		assert.Contains(t, err.Error(), "not in the allowed commands")
	})

	t.Run("allows command in whitelist", func(t *testing.T) {
		restrictedResolver := NewResolver(WithAllowedCommands("/bin/echo", "cat"))
		result, err := restrictedResolver.Resolve("$(/bin/echo allowed-key)")
		require.NoError(t, err)
		assert.Equal(t, "allowed-key", result)
	})

	t.Run("errors on command failure", func(t *testing.T) {
		result, err := resolver.Resolve("$(cat /nonexistent/file)")
		assert.Error(t, err)
		assert.Empty(t, result)
		assert.Contains(t, err.Error(), "execution failed")
	})

	t.Run("errors on command timeout", func(t *testing.T) {
		fastResolver := NewResolver(WithCommandTimeout(100 * time.Millisecond))
		result, err := fastResolver.Resolve("$(/bin/sleep 5)")
		assert.Error(t, err)
		assert.Empty(t, result)
		assert.Contains(t, err.Error(), "timed out")
	})

	t.Run("errors on empty command output", func(t *testing.T) {
		// Create a script that outputs nothing
		tmpDir := t.TempDir()
		scriptFile := filepath.Join(tmpDir, "empty.sh")
		err := os.WriteFile(scriptFile, []byte("#!/bin/sh\n"), 0755)
		require.NoError(t, err)

		result, err := resolver.Resolve(fmt.Sprintf("$(%s)", scriptFile))
		assert.Error(t, err)
		assert.Empty(t, result)
		if err != nil {
			assert.Contains(t, err.Error(), "empty output")
		}
	})
}

func TestResolver_ResolutionPrecedence(t *testing.T) {
	resolver := NewResolver()

	t.Run("command takes precedence over other patterns", func(t *testing.T) {
		result, err := resolver.Resolve("$(/bin/echo from-command)")
		require.NoError(t, err)
		assert.Equal(t, "from-command", result)
	})

	t.Run("env var takes precedence over file path", func(t *testing.T) {
		os.Setenv("TEST_KEY", "from-env")
		defer os.Unsetenv("TEST_KEY")

		result, err := resolver.Resolve("${TEST_KEY}")
		require.NoError(t, err)
		assert.Equal(t, "from-env", result)
	})

	t.Run("file path is checked before treating as direct string", func(t *testing.T) {
		tmpDir := t.TempDir()
		keyFile := filepath.Join(tmpDir, "key.txt")
		err := os.WriteFile(keyFile, []byte("from-file"), 0600)
		require.NoError(t, err)

		result, err := resolver.Resolve(keyFile)
		require.NoError(t, err)
		assert.Equal(t, "from-file", result)
	})
}

func TestResolver_ResolveAll(t *testing.T) {
	resolver := NewResolver()

	t.Run("resolves multiple values", func(t *testing.T) {
		// Setup environment
		os.Setenv("ANTHROPIC_KEY", "sk-ant-env-key")
		defer os.Unsetenv("ANTHROPIC_KEY")

		tmpDir := t.TempDir()
		keyFile := filepath.Join(tmpDir, "openai.key")
		err := os.WriteFile(keyFile, []byte("sk-openai-file-key"), 0600)
		require.NoError(t, err)

		values := map[string]string{
			"anthropic": "${ANTHROPIC_KEY}",
			"openai":    keyFile,
			"google":    "AIzaSyDirect123",
			"command":   "$(/bin/echo sk-cmd-key)",
		}

		results, err := resolver.ResolveAll(values)
		require.NoError(t, err)
		assert.Len(t, results, 4)
		assert.Equal(t, "sk-ant-env-key", results["anthropic"])
		assert.Equal(t, "sk-openai-file-key", results["openai"])
		assert.Equal(t, "AIzaSyDirect123", results["google"])
		assert.Equal(t, "sk-cmd-key", results["command"])
	})

	t.Run("returns error if any resolution fails", func(t *testing.T) {
		values := map[string]string{
			"valid":   "sk-direct-key",
			"invalid": "${MISSING_ENV_VAR}",
		}

		results, err := resolver.ResolveAll(values)
		assert.Error(t, err)
		assert.Nil(t, results)
		assert.Contains(t, err.Error(), "MISSING_ENV_VAR")
	})

	t.Run("handles empty map", func(t *testing.T) {
		results, err := resolver.ResolveAll(map[string]string{})
		require.NoError(t, err)
		assert.Empty(t, results)
	})
}

func TestIsFilePath(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		// Positive cases
		{name: "home directory path", value: "~/secrets/key.txt", expected: true},
		{name: "absolute path", value: "/etc/secrets/api.key", expected: true},
		{name: "relative path with ./", value: "./config/key.txt", expected: true},
		{name: "relative path with ../", value: "../secrets/key.txt", expected: true},
		{name: "path with subdirectory", value: "config/secrets/key.txt", expected: true},
		{name: "windows path", value: "C:\\secrets\\key.txt", expected: true},
		{name: ".txt extension", value: "api-key.txt", expected: true},
		{name: ".key extension", value: "api.key", expected: true},
		{name: ".pem extension", value: "cert.pem", expected: true},
		{name: ".secret extension", value: "my.secret", expected: true},
		{name: ".env extension", value: "config.env", expected: true},

		// Negative cases
		{name: "plain API key", value: "sk-ant-api-key-123", expected: false},
		{name: "environment variable", value: "${API_KEY}", expected: false},
		{name: "command", value: "$(pass show key)", expected: false},
		{name: "single word", value: "apikey", expected: false},
		{name: "UUID", value: "123e4567-e89b-12d3-a456-426614174000", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isFilePath(tt.value)
			assert.Equal(t, tt.expected, result, "value: %s", tt.value)
		})
	}
}

func TestResolver_ComplexScenarios(t *testing.T) {
	resolver := NewResolver()

	t.Run("resolves configuration with mixed sources", func(t *testing.T) {
		// Setup
		os.Setenv("PROD_ANTHROPIC_KEY", "sk-prod-ant-key")
		defer os.Unsetenv("PROD_ANTHROPIC_KEY")

		tmpDir := t.TempDir()
		googleKeyFile := filepath.Join(tmpDir, "google.key")
		err := os.WriteFile(googleKeyFile, []byte("AIzaSyGoogleKey123"), 0600)
		require.NoError(t, err)

		testCases := []struct {
			input    string
			expected string
		}{
			{"${PROD_ANTHROPIC_KEY}", "sk-prod-ant-key"},
			{googleKeyFile, "AIzaSyGoogleKey123"},
			{"sk-ant-direct-key", "sk-ant-direct-key"},
			{"$(/bin/echo sk-cmd-key)", "sk-cmd-key"},
		}

		for _, tc := range testCases {
			result, err := resolver.Resolve(tc.input)
			require.NoError(t, err, "failed for input: %s", tc.input)
			assert.Equal(t, tc.expected, result)
		}
	})

	t.Run("handles multiline file content", func(t *testing.T) {
		tmpDir := t.TempDir()
		keyFile := filepath.Join(tmpDir, "multiline.key")
		content := `
sk-ant-key-line1
sk-ant-key-line2
sk-ant-key-line3
`
		err := os.WriteFile(keyFile, []byte(content), 0600)
		require.NoError(t, err)

		result, err := resolver.Resolve(keyFile)
		require.NoError(t, err)
		// Should trim all whitespace and return all content
		assert.Contains(t, result, "sk-ant-key-line1")
	})
}
