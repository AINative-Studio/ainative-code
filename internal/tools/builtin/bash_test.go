package builtin

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/AINative-studio/ainative-code/internal/tools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBashTool(t *testing.T) {
	sandbox := DefaultSandbox("/tmp")
	tool := NewBashTool(sandbox)

	assert.NotNil(t, tool)
	assert.Equal(t, "bash", tool.Name())
	assert.NotEmpty(t, tool.Description())
	assert.Equal(t, tools.CategorySystem, tool.Category())
	assert.True(t, tool.RequiresConfirmation())
}

func TestBashTool_Schema(t *testing.T) {
	sandbox := DefaultSandbox("/tmp")
	tool := NewBashTool(sandbox)

	schema := tool.Schema()

	assert.Equal(t, "object", schema.Type)
	assert.Contains(t, schema.Properties, "command")
	assert.Contains(t, schema.Properties, "working_dir")
	assert.Contains(t, schema.Properties, "timeout")
	assert.Contains(t, schema.Properties, "capture_stderr")
	assert.Equal(t, []string{"command"}, schema.Required)

	// Verify command property
	cmdProp := schema.Properties["command"]
	assert.Equal(t, "string", cmdProp.Type)
	assert.NotNil(t, cmdProp.MaxLength)
	assert.Equal(t, 65536, *cmdProp.MaxLength)
}

func TestBashTool_Execute_Success(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "bash-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	sandbox := DefaultSandbox(tmpDir)
	tool := NewBashTool(sandbox)

	tests := []struct {
		name           string
		input          map[string]interface{}
		expectSuccess  bool
		outputContains string
	}{
		{
			name: "simple echo command",
			input: map[string]interface{}{
				"command": "echo 'Hello, World!'",
			},
			expectSuccess:  true,
			outputContains: "Hello, World!",
		},
		{
			name: "pwd command",
			input: map[string]interface{}{
				"command":     "pwd",
				"working_dir": tmpDir,
			},
			expectSuccess:  true,
			outputContains: tmpDir,
		},
		{
			name: "command with stderr",
			input: map[string]interface{}{
				"command":        "echo 'stdout'; echo 'stderr' >&2",
				"capture_stderr": true,
			},
			expectSuccess:  true,
			outputContains: "STDERR",
		},
		{
			name: "ls command",
			input: map[string]interface{}{
				"command": "ls -la",
			},
			expectSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := tool.Execute(ctx, tt.input)

			require.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.expectSuccess, result.Success)

			if tt.outputContains != "" {
				assert.Contains(t, result.Output, tt.outputContains)
			}

			// Verify metadata
			assert.NotNil(t, result.Metadata)
			assert.Contains(t, result.Metadata, "exit_code")
			assert.Equal(t, 0, result.Metadata["exit_code"])
			assert.Contains(t, result.Metadata, "execution_time_ms")
		})
	}
}

func TestBashTool_Execute_CommandFailure(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "bash-fail-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	sandbox := DefaultSandbox(tmpDir)
	tool := NewBashTool(sandbox)

	ctx := context.Background()
	input := map[string]interface{}{
		"command": "exit 1",
	}

	result, err := tool.Execute(ctx, input)

	require.NoError(t, err) // Tool execution succeeds but command fails
	assert.NotNil(t, result)
	assert.False(t, result.Success)
	assert.NotNil(t, result.Error)
	assert.Contains(t, result.Metadata, "exit_code")
	assert.Equal(t, 1, result.Metadata["exit_code"])
}

func TestBashTool_Execute_ValidationErrors(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "bash-validation-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	sandbox := DefaultSandbox(tmpDir)
	tool := NewBashTool(sandbox)

	tests := []struct {
		name          string
		input         map[string]interface{}
		errorType     interface{}
		errorContains string
	}{
		{
			name:          "missing command",
			input:         map[string]interface{}{},
			errorType:     &tools.ErrInvalidInput{},
			errorContains: "command is required",
		},
		{
			name: "empty command",
			input: map[string]interface{}{
				"command": "   ",
			},
			errorType:     &tools.ErrInvalidInput{},
			errorContains: "cannot be empty",
		},
		{
			name: "invalid command type",
			input: map[string]interface{}{
				"command": 123,
			},
			errorType:     &tools.ErrInvalidInput{},
			errorContains: "must be a string",
		},
		{
			name: "denied command",
			input: map[string]interface{}{
				"command": "rm -rf /",
			},
			errorType:     &tools.ErrPermissionDenied{},
			errorContains: "denied",
		},
		{
			name: "invalid timeout type",
			input: map[string]interface{}{
				"command": "echo test",
				"timeout": "not-a-number",
			},
			errorType:     &tools.ErrInvalidInput{},
			errorContains: "timeout must be an integer",
		},
		{
			name: "negative timeout",
			input: map[string]interface{}{
				"command": "echo test",
				"timeout": -5,
			},
			errorType:     &tools.ErrInvalidInput{},
			errorContains: "timeout must be positive",
		},
		{
			name: "timeout too large",
			input: map[string]interface{}{
				"command": "echo test",
				"timeout": 500,
			},
			errorType:     &tools.ErrInvalidInput{},
			errorContains: "cannot exceed 300",
		},
		{
			name: "invalid working_dir type",
			input: map[string]interface{}{
				"command":     "echo test",
				"working_dir": 123,
			},
			errorType:     &tools.ErrInvalidInput{},
			errorContains: "working_dir must be a string",
		},
		{
			name: "working_dir outside sandbox",
			input: map[string]interface{}{
				"command":     "echo test",
				"working_dir": "/etc",
			},
			errorType:     &tools.ErrInvalidInput{},
			errorContains: "invalid working directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := tool.Execute(ctx, tt.input)

			assert.Nil(t, result)
			assert.Error(t, err)
			assert.IsType(t, tt.errorType, err)
			if tt.errorContains != "" {
				assert.Contains(t, err.Error(), tt.errorContains)
			}
		})
	}
}

func TestBashTool_Execute_Timeout(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "bash-timeout-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	sandbox := DefaultSandbox(tmpDir)
	tool := NewBashTool(sandbox)

	ctx := context.Background()
	input := map[string]interface{}{
		"command": "sleep 5",
		"timeout": 1, // 1 second timeout
	}

	result, err := tool.Execute(ctx, input)

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.IsType(t, &tools.ErrTimeout{}, err)
}

func TestBashTool_Execute_ContextCancellation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "bash-cancel-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	sandbox := DefaultSandbox(tmpDir)
	tool := NewBashTool(sandbox)

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel after a short delay
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	input := map[string]interface{}{
		"command": "sleep 10",
		"timeout": 30,
	}

	result, err := tool.Execute(ctx, input)

	// Should get timeout error when context is cancelled
	assert.Nil(t, result)
	assert.Error(t, err)
}

func TestBashTool_Execute_WorkingDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "bash-wd-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a subdirectory
	subDir := filepath.Join(tmpDir, "subdir")
	err = os.Mkdir(subDir, 0755)
	require.NoError(t, err)

	// Create a test file in the subdirectory
	testFile := filepath.Join(subDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(t, err)

	sandbox := DefaultSandbox(tmpDir)
	tool := NewBashTool(sandbox)

	ctx := context.Background()
	input := map[string]interface{}{
		"command":     "ls",
		"working_dir": subDir,
	}

	result, err := tool.Execute(ctx, input)

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Contains(t, result.Output, "test.txt")
	assert.Equal(t, subDir, result.Metadata["working_dir"])
}

func TestBashTool_Execute_StderrCapture(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "bash-stderr-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	sandbox := DefaultSandbox(tmpDir)
	tool := NewBashTool(sandbox)

	ctx := context.Background()

	t.Run("capture stderr separately", func(t *testing.T) {
		input := map[string]interface{}{
			"command":        "echo 'to stdout'; echo 'to stderr' >&2",
			"capture_stderr": true,
		}

		result, err := tool.Execute(ctx, input)

		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.Contains(t, result.Output, "STDOUT")
		assert.Contains(t, result.Output, "STDERR")
		assert.Contains(t, result.Output, "to stdout")
		assert.Contains(t, result.Output, "to stderr")
		assert.Equal(t, true, result.Metadata["has_stderr"])
	})

	t.Run("combine stderr with stdout", func(t *testing.T) {
		input := map[string]interface{}{
			"command":        "echo 'to stdout'; echo 'to stderr' >&2",
			"capture_stderr": false,
		}

		result, err := tool.Execute(ctx, input)

		require.NoError(t, err)
		assert.True(t, result.Success)
		// Should not have separate sections
		assert.NotContains(t, result.Output, "--- STDOUT ---")
		assert.Equal(t, false, result.Metadata["has_stderr"])
	})
}

func TestBashTool_Execute_OutputSizeLimit(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "bash-size-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create sandbox with small output limit
	sandbox := DefaultSandbox(tmpDir)
	sandbox.MaxOutputSize = 100 // 100 bytes

	tool := NewBashTool(sandbox)

	ctx := context.Background()
	input := map[string]interface{}{
		"command": "head -c 1000 /dev/zero | base64", // Generate more than 100 bytes
	}

	result, err := tool.Execute(ctx, input)

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.IsType(t, &tools.ErrOutputTooLarge{}, err)
}

func TestBashTool_Execute_SecurityPatterns(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "bash-security-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	sandbox := DefaultSandbox(tmpDir)
	tool := NewBashTool(sandbox)

	dangerousCommands := []string{
		"rm -rf /",
		"rm -rf /*",
		":(){ :|:& };:",
		"dd if=/dev/zero of=/dev/sda",
		"chmod 777 /tmp/file",
		"mkfs /dev/sda1",
	}

	ctx := context.Background()

	for _, cmd := range dangerousCommands {
		t.Run(cmd, func(t *testing.T) {
			input := map[string]interface{}{
				"command": cmd,
			}

			result, err := tool.Execute(ctx, input)

			assert.Nil(t, result)
			assert.Error(t, err)
			assert.IsType(t, &tools.ErrPermissionDenied{}, err)
		})
	}
}

func TestBashTool_Execute_MetadataCompleteness(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "bash-metadata-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	sandbox := DefaultSandbox(tmpDir)
	sandbox.AuditLog = false // Disable for cleaner test output
	tool := NewBashTool(sandbox)

	ctx := context.Background()
	input := map[string]interface{}{
		"command": "echo 'test'",
	}

	result, err := tool.Execute(ctx, input)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Metadata)

	// Verify all expected metadata fields
	expectedFields := []string{
		"command",
		"working_dir",
		"execution_time_ms",
		"output_size",
		"exit_code",
		"has_stderr",
	}

	for _, field := range expectedFields {
		assert.Contains(t, result.Metadata, field, "Missing metadata field: %s", field)
	}

	// Verify metadata types and values
	assert.IsType(t, "", result.Metadata["command"])
	assert.IsType(t, "", result.Metadata["working_dir"])
	assert.IsType(t, int64(0), result.Metadata["execution_time_ms"])
	assert.IsType(t, int64(0), result.Metadata["output_size"])
	assert.IsType(t, 0, result.Metadata["exit_code"])
	assert.IsType(t, false, result.Metadata["has_stderr"])
}

func TestBashTool_Execute_ComplexCommands(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "bash-complex-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	sandbox := DefaultSandbox(tmpDir)
	tool := NewBashTool(sandbox)

	tests := []struct {
		name           string
		command        string
		expectSuccess  bool
		outputContains string
	}{
		{
			name:           "pipe command",
			command:        "echo 'hello world' | grep hello",
			expectSuccess:  true,
			outputContains: "hello",
		},
		{
			name:           "command with redirects",
			command:        "echo 'test' > /dev/null && echo 'success'",
			expectSuccess:  true,
			outputContains: "success",
		},
		{
			name:          "command with variables",
			command:       "VAR=test; echo $VAR",
			expectSuccess: true,
			outputContains: "test",
		},
		{
			name:           "multiple commands",
			command:        "echo 'first' && echo 'second'",
			expectSuccess:  true,
			outputContains: "second",
		},
	}

	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := map[string]interface{}{
				"command": tt.command,
			}

			result, err := tool.Execute(ctx, input)

			require.NoError(t, err)
			assert.Equal(t, tt.expectSuccess, result.Success)
			if tt.outputContains != "" {
				assert.Contains(t, strings.ToLower(result.Output), strings.ToLower(tt.outputContains))
			}
		})
	}
}
