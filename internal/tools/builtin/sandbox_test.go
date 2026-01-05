package builtin

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/AINative-studio/ainative-code/internal/tools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultSandbox(t *testing.T) {
	workDir := "/tmp/test"
	sandbox := DefaultSandbox(workDir)

	assert.NotNil(t, sandbox)
	assert.Equal(t, []string{workDir}, sandbox.AllowedPaths)
	assert.NotEmpty(t, sandbox.AllowedCommands)
	assert.NotEmpty(t, sandbox.DeniedCommands)
	assert.Equal(t, workDir, sandbox.WorkingDirectory)
	assert.Equal(t, int64(100*1024*1024), sandbox.MaxFileSize)
	assert.Equal(t, int64(10*1024*1024), sandbox.MaxOutputSize)
	assert.True(t, sandbox.AuditLog)
}

func TestDefaultAllowedCommands(t *testing.T) {
	commands := DefaultAllowedCommands()

	assert.NotEmpty(t, commands)
	assert.Contains(t, commands, "ls")
	assert.Contains(t, commands, "git")
	assert.Contains(t, commands, "go")
	assert.Contains(t, commands, "grep")
	assert.NotContains(t, commands, "rm")
	assert.NotContains(t, commands, "shutdown")
}

func TestDefaultDeniedCommands(t *testing.T) {
	commands := DefaultDeniedCommands()

	assert.NotEmpty(t, commands)
	assert.Contains(t, commands, "rm")
	assert.Contains(t, commands, "shutdown")
	assert.Contains(t, commands, "chmod")
	assert.Contains(t, commands, "dd")
	assert.NotContains(t, commands, "ls")
	assert.NotContains(t, commands, "git")
}

func TestSandbox_ValidatePath(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "sandbox-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name        string
		allowedPath []string
		testPath    string
		expectError bool
		errorType   interface{}
	}{
		{
			name:        "path within allowed directory",
			allowedPath: []string{tmpDir},
			testPath:    filepath.Join(tmpDir, "test.txt"),
			expectError: false,
		},
		{
			name:        "path outside allowed directory",
			allowedPath: []string{tmpDir},
			testPath:    "/etc/passwd",
			expectError: true,
			errorType:   &tools.ErrPermissionDenied{},
		},
		{
			name:        "directory traversal attempt",
			allowedPath: []string{tmpDir},
			testPath:    filepath.Join(tmpDir, "..", "..", "etc", "passwd"),
			expectError: true,
			errorType:   &tools.ErrPermissionDenied{},
		},
		{
			name:        "no allowed paths configured",
			allowedPath: []string{},
			testPath:    tmpDir,
			expectError: true,
			errorType:   &tools.ErrPermissionDenied{},
		},
		{
			name:        "multiple allowed paths - first match",
			allowedPath: []string{tmpDir, "/tmp"},
			testPath:    filepath.Join(tmpDir, "file.txt"),
			expectError: false,
		},
		{
			name:        "relative path within allowed directory",
			allowedPath: []string{tmpDir},
			testPath:    "test.txt",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sandbox := &Sandbox{
				AllowedPaths:     tt.allowedPath,
				WorkingDirectory: tmpDir,
				AuditLog:         false, // Disable for cleaner test output
			}

			err := sandbox.ValidatePath(tt.testPath)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorType != nil {
					assert.IsType(t, tt.errorType, err)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSandbox_ValidateCommand(t *testing.T) {
	tests := []struct {
		name           string
		allowedCmds    []string
		deniedCmds     []string
		testCommand    string
		expectError    bool
		errorType      interface{}
		errorContains  string
	}{
		{
			name:        "allowed command",
			allowedCmds: []string{"ls", "git"},
			deniedCmds:  []string{"rm"},
			testCommand: "ls -la",
			expectError: false,
		},
		{
			name:          "denied command - explicit blacklist",
			allowedCmds:   []string{"ls", "git", "rm"},
			deniedCmds:    []string{"rm"},
			testCommand:   "rm -rf /",
			expectError:   true,
			errorType:     &tools.ErrPermissionDenied{},
			errorContains: "explicitly denied",
		},
		{
			name:          "dangerous pattern - rm -rf /",
			allowedCmds:   []string{"ls", "git"},
			deniedCmds:    []string{},
			testCommand:   "rm -rf /",
			expectError:   true,
			errorType:     &tools.ErrPermissionDenied{},
			errorContains: "delete root filesystem",
		},
		{
			name:          "dangerous pattern - fork bomb",
			allowedCmds:   []string{"ls"},
			deniedCmds:    []string{},
			testCommand:   ":(){ :|:& };:",
			expectError:   true,
			errorType:     &tools.ErrPermissionDenied{},
			errorContains: "fork bomb",
		},
		{
			name:          "command not in whitelist",
			allowedCmds:   []string{"ls", "git"},
			deniedCmds:    []string{},
			testCommand:   "cat /etc/passwd",
			expectError:   true,
			errorType:     &tools.ErrPermissionDenied{},
			errorContains: "not in the allowed list",
		},
		{
			name:          "empty command",
			allowedCmds:   []string{"ls"},
			deniedCmds:    []string{},
			testCommand:   "",
			expectError:   true,
			errorType:     &tools.ErrInvalidInput{},
			errorContains: "cannot be empty",
		},
		{
			name:        "command with arguments",
			allowedCmds: []string{"git"},
			deniedCmds:  []string{},
			testCommand: "git status --short",
			expectError: false,
		},
		{
			name:          "dangerous pattern - chmod 777",
			allowedCmds:   []string{"chmod"},
			deniedCmds:    []string{},
			testCommand:   "chmod 777 /tmp/file",
			expectError:   true,
			errorType:     &tools.ErrPermissionDenied{},
			errorContains: "unsafe permission modification",
		},
		{
			name:          "dangerous pattern - disk device access",
			allowedCmds:   []string{"dd"},
			deniedCmds:    []string{},
			testCommand:   "dd if=/dev/zero of=/dev/sda",
			expectError:   true,
			errorType:     &tools.ErrPermissionDenied{},
			errorContains: "direct disk device access",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sandbox := &Sandbox{
				AllowedCommands: tt.allowedCmds,
				DeniedCommands:  tt.deniedCmds,
				AuditLog:        false,
			}

			err := sandbox.ValidateCommand(tt.testCommand)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorType != nil {
					assert.IsType(t, tt.errorType, err)
				}
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSandbox_ResolveWorkingDirectory(t *testing.T) {
	// Create temporary directories for testing
	tmpDir, err := os.MkdirTemp("", "sandbox-wd-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	subDir := filepath.Join(tmpDir, "subdir")
	err = os.Mkdir(subDir, 0755)
	require.NoError(t, err)

	tests := []struct {
		name          string
		workingDir    string
		allowedPaths  []string
		inputDir      string
		expectError   bool
		expectedDir   string
		errorContains string
	}{
		{
			name:         "use default working directory",
			workingDir:   tmpDir,
			allowedPaths: []string{tmpDir},
			inputDir:     "",
			expectError:  false,
			expectedDir:  tmpDir,
		},
		{
			name:         "valid subdirectory",
			workingDir:   tmpDir,
			allowedPaths: []string{tmpDir},
			inputDir:     subDir,
			expectError:  false,
			expectedDir:  subDir,
		},
		{
			name:          "directory outside allowed paths",
			workingDir:    tmpDir,
			allowedPaths:  []string{tmpDir},
			inputDir:      "/etc",
			expectError:   true,
			errorContains: "not allowed",
		},
		{
			name:          "non-existent directory",
			workingDir:    tmpDir,
			allowedPaths:  []string{tmpDir},
			inputDir:      filepath.Join(tmpDir, "nonexistent"),
			expectError:   true,
			errorContains: "does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sandbox := &Sandbox{
				WorkingDirectory: tt.workingDir,
				AllowedPaths:     tt.allowedPaths,
				AuditLog:         false,
			}

			result, err := sandbox.ResolveWorkingDirectory(tt.inputDir)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedDir, result)
			}
		})
	}
}

func TestSandbox_ValidateFileSize(t *testing.T) {
	tests := []struct {
		name        string
		maxFileSize int64
		testSize    int64
		expectError bool
	}{
		{
			name:        "size within limit",
			maxFileSize: 1024 * 1024,
			testSize:    512 * 1024,
			expectError: false,
		},
		{
			name:        "size exceeds limit",
			maxFileSize: 1024 * 1024,
			testSize:    2 * 1024 * 1024,
			expectError: true,
		},
		{
			name:        "no limit set",
			maxFileSize: 0,
			testSize:    1024 * 1024 * 1024,
			expectError: false,
		},
		{
			name:        "exact limit",
			maxFileSize: 1024,
			testSize:    1024,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sandbox := &Sandbox{
				MaxFileSize: tt.maxFileSize,
			}

			err := sandbox.ValidateFileSize(tt.testSize)

			if tt.expectError {
				assert.Error(t, err)
				assert.IsType(t, &tools.ErrExecutionFailed{}, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSandbox_ValidateOutputSize(t *testing.T) {
	tests := []struct {
		name          string
		maxOutputSize int64
		testSize      int64
		expectError   bool
	}{
		{
			name:          "size within limit",
			maxOutputSize: 10 * 1024,
			testSize:      5 * 1024,
			expectError:   false,
		},
		{
			name:          "size exceeds limit",
			maxOutputSize: 10 * 1024,
			testSize:      20 * 1024,
			expectError:   true,
		},
		{
			name:          "no limit set",
			maxOutputSize: 0,
			testSize:      1024 * 1024,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sandbox := &Sandbox{
				MaxOutputSize: tt.maxOutputSize,
			}

			err := sandbox.ValidateOutputSize(tt.testSize)

			if tt.expectError {
				assert.Error(t, err)
				assert.IsType(t, &tools.ErrOutputTooLarge{}, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMatchesCommand(t *testing.T) {
	tests := []struct {
		name        string
		baseCommand string
		pattern     string
		expected    bool
	}{
		{
			name:        "exact match",
			baseCommand: "ls",
			pattern:     "ls",
			expected:    true,
		},
		{
			name:        "no match",
			baseCommand: "ls",
			pattern:     "cat",
			expected:    false,
		},
		{
			name:        "path pattern - basename match",
			baseCommand: "git",
			pattern:     "/usr/bin/git",
			expected:    true,
		},
		{
			name:        "path pattern - no match",
			baseCommand: "ls",
			pattern:     "/usr/bin/cat",
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesCommand(tt.baseCommand, tt.pattern)
			assert.Equal(t, tt.expected, result)
		})
	}
}
