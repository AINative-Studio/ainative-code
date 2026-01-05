package builtin

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/AINative-studio/ainative-code/internal/tools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSearchReplaceTool(t *testing.T) {
	sandbox := DefaultSandbox("/tmp")
	tool := NewSearchReplaceTool(sandbox)

	assert.NotNil(t, tool)
	assert.Equal(t, "search_replace", tool.Name())
	assert.NotEmpty(t, tool.Description())
	assert.Equal(t, tools.CategoryText, tool.Category())
	assert.True(t, tool.RequiresConfirmation())
}

func TestSearchReplaceTool_Schema(t *testing.T) {
	sandbox := DefaultSandbox("/tmp")
	tool := NewSearchReplaceTool(sandbox)

	schema := tool.Schema()

	assert.Equal(t, "object", schema.Type)
	assert.Contains(t, schema.Properties, "pattern")
	assert.Contains(t, schema.Properties, "replacement")
	assert.Contains(t, schema.Properties, "path")
	assert.Contains(t, schema.Properties, "dry_run")
	assert.Contains(t, schema.Properties, "backup")
	assert.Equal(t, []string{"pattern", "replacement"}, schema.Required)
}

func TestSearchReplaceTool_Execute_Success(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "replace-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test file
	testContent := "Hello world\nHello again\nGoodbye world"
	testFile := filepath.Join(tmpDir, "test.txt")
	err = os.WriteFile(testFile, []byte(testContent), 0644)
	require.NoError(t, err)

	sandbox := DefaultSandbox(tmpDir)
	tool := NewSearchReplaceTool(sandbox)

	ctx := context.Background()

	t.Run("simple replacement", func(t *testing.T) {
		input := map[string]interface{}{
			"pattern":     "Hello",
			"replacement": "Hi",
			"path":        testFile,
			"dry_run":     true, // Use dry run for test
		}

		result, err := tool.Execute(ctx, input)

		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.Contains(t, result.Output, "2")
		assert.Equal(t, 2, result.Metadata["total_replacements"])
		assert.Equal(t, true, result.Metadata["dry_run"])
	})

	t.Run("regex replacement with capture groups", func(t *testing.T) {
		input := map[string]interface{}{
			"pattern":     "(\\w+) world",
			"replacement": "world $1",
			"path":        testFile,
			"dry_run":     true,
		}

		result, err := tool.Execute(ctx, input)

		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.Equal(t, 2, result.Metadata["total_replacements"])
	})

	t.Run("case insensitive replacement", func(t *testing.T) {
		input := map[string]interface{}{
			"pattern":        "HELLO",
			"replacement":    "Hi",
			"path":           testFile,
			"case_sensitive": false,
			"dry_run":        true,
		}

		result, err := tool.Execute(ctx, input)

		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.Equal(t, 2, result.Metadata["total_replacements"])
	})
}

func TestSearchReplaceTool_Execute_ActualReplacement(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "replace-actual-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test file
	testContent := "foo bar foo"
	testFile := filepath.Join(tmpDir, "test.txt")
	err = os.WriteFile(testFile, []byte(testContent), 0644)
	require.NoError(t, err)

	sandbox := DefaultSandbox(tmpDir)
	tool := NewSearchReplaceTool(sandbox)

	ctx := context.Background()

	t.Run("actual file modification", func(t *testing.T) {
		input := map[string]interface{}{
			"pattern":     "foo",
			"replacement": "bar",
			"path":        testFile,
			"dry_run":     false,
			"backup":      true,
		}

		result, err := tool.Execute(ctx, input)

		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.Equal(t, 2, result.Metadata["total_replacements"])
		assert.Equal(t, 1, result.Metadata["files_changed"])

		// Verify file was modified
		modifiedContent, err := os.ReadFile(testFile)
		require.NoError(t, err)
		assert.Equal(t, "bar bar bar", string(modifiedContent))

		// Verify backup was created
		backupFile := testFile + ".bak"
		_, err = os.Stat(backupFile)
		assert.NoError(t, err)

		// Verify backup contains original content
		backupContent, err := os.ReadFile(backupFile)
		require.NoError(t, err)
		assert.Equal(t, testContent, string(backupContent))
	})
}

func TestSearchReplaceTool_Execute_MultipleFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "replace-multi-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create multiple test files
	files := map[string]string{
		"file1.txt": "test pattern here",
		"file2.txt": "another pattern match",
		"file3.go":  "// pattern in code",
	}

	for name, content := range files {
		err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0644)
		require.NoError(t, err)
	}

	sandbox := DefaultSandbox(tmpDir)
	tool := NewSearchReplaceTool(sandbox)

	ctx := context.Background()

	t.Run("replace in all txt files", func(t *testing.T) {
		input := map[string]interface{}{
			"pattern":      "pattern",
			"replacement":  "PATTERN",
			"path":         tmpDir,
			"file_pattern": "*.txt",
			"dry_run":      true,
		}

		result, err := tool.Execute(ctx, input)

		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.Equal(t, 2, result.Metadata["files_matched"])
	})
}

func TestSearchReplaceTool_Execute_ValidationErrors(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "replace-validation-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	sandbox := DefaultSandbox(tmpDir)
	tool := NewSearchReplaceTool(sandbox)

	tests := []struct {
		name          string
		input         map[string]interface{}
		errorType     interface{}
		errorContains string
	}{
		{
			name:          "missing pattern",
			input:         map[string]interface{}{"replacement": "test"},
			errorType:     &tools.ErrInvalidInput{},
			errorContains: "pattern is required",
		},
		{
			name:          "missing replacement",
			input:         map[string]interface{}{"pattern": "test"},
			errorType:     &tools.ErrInvalidInput{},
			errorContains: "replacement is required",
		},
		{
			name: "empty pattern",
			input: map[string]interface{}{
				"pattern":     "",
				"replacement": "test",
			},
			errorType:     &tools.ErrInvalidInput{},
			errorContains: "cannot be empty",
		},
		{
			name: "invalid regex",
			input: map[string]interface{}{
				"pattern":     "[invalid(regex",
				"replacement": "test",
			},
			errorType:     &tools.ErrInvalidInput{},
			errorContains: "invalid regex",
		},
		{
			name: "invalid pattern type",
			input: map[string]interface{}{
				"pattern":     123,
				"replacement": "test",
			},
			errorType:     &tools.ErrInvalidInput{},
			errorContains: "must be a string",
		},
		{
			name: "invalid replacement type",
			input: map[string]interface{}{
				"pattern":     "test",
				"replacement": 123,
			},
			errorType:     &tools.ErrInvalidInput{},
			errorContains: "must be a string",
		},
		{
			name: "max_files too large",
			input: map[string]interface{}{
				"pattern":     "test",
				"replacement": "replace",
				"max_files":   2000,
			},
			errorType:     &tools.ErrInvalidInput{},
			errorContains: "between 1 and 1000",
		},
		{
			name: "max_files zero",
			input: map[string]interface{}{
				"pattern":     "test",
				"replacement": "replace",
				"max_files":   0,
			},
			errorType:     &tools.ErrInvalidInput{},
			errorContains: "between 1 and 1000",
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

func TestSearchReplaceTool_Execute_PathValidation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "replace-path-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	sandbox := DefaultSandbox(tmpDir)
	tool := NewSearchReplaceTool(sandbox)

	ctx := context.Background()

	t.Run("path outside sandbox", func(t *testing.T) {
		input := map[string]interface{}{
			"pattern":     "test",
			"replacement": "replace",
			"path":        "/etc/passwd",
		}

		result, err := tool.Execute(ctx, input)

		assert.Nil(t, result)
		assert.Error(t, err)
		assert.IsType(t, &tools.ErrPermissionDenied{}, err)
	})

	t.Run("non-existent path", func(t *testing.T) {
		input := map[string]interface{}{
			"pattern":     "test",
			"replacement": "replace",
			"path":        filepath.Join(tmpDir, "nonexistent"),
		}

		result, err := tool.Execute(ctx, input)

		assert.Nil(t, result)
		assert.Error(t, err)
		assert.IsType(t, &tools.ErrExecutionFailed{}, err)
		assert.Contains(t, err.Error(), "does not exist")
	})
}

func TestSearchReplaceTool_Execute_NoMatches(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "replace-nomatch-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	testContent := "foo bar baz"
	testFile := filepath.Join(tmpDir, "test.txt")
	err = os.WriteFile(testFile, []byte(testContent), 0644)
	require.NoError(t, err)

	sandbox := DefaultSandbox(tmpDir)
	tool := NewSearchReplaceTool(sandbox)

	ctx := context.Background()
	input := map[string]interface{}{
		"pattern":     "notfound",
		"replacement": "replace",
		"path":        testFile,
		"dry_run":     true,
	}

	result, err := tool.Execute(ctx, input)

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Contains(t, result.Output, "No matches found")
	assert.Equal(t, 0, result.Metadata["total_replacements"])
}

func TestSearchReplaceTool_Execute_RecursiveSearch(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "replace-recursive-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create nested directory structure
	subDir := filepath.Join(tmpDir, "subdir")
	err = os.Mkdir(subDir, 0755)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(tmpDir, "root.txt"), []byte("pattern in root"), 0644)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(subDir, "sub.txt"), []byte("pattern in subdir"), 0644)
	require.NoError(t, err)

	sandbox := DefaultSandbox(tmpDir)
	tool := NewSearchReplaceTool(sandbox)

	ctx := context.Background()

	t.Run("recursive search", func(t *testing.T) {
		input := map[string]interface{}{
			"pattern":     "pattern",
			"replacement": "PATTERN",
			"path":        tmpDir,
			"recursive":   true,
			"dry_run":     true,
		}

		result, err := tool.Execute(ctx, input)

		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.Equal(t, 2, result.Metadata["files_matched"])
	})

	t.Run("non-recursive search", func(t *testing.T) {
		input := map[string]interface{}{
			"pattern":     "pattern",
			"replacement": "PATTERN",
			"path":        tmpDir,
			"recursive":   false,
			"dry_run":     true,
		}

		result, err := tool.Execute(ctx, input)

		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.Equal(t, 1, result.Metadata["files_matched"])
	})
}

func TestSearchReplaceTool_Execute_BackupOption(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "replace-backup-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	testContent := "foo bar"
	testFile := filepath.Join(tmpDir, "test.txt")
	err = os.WriteFile(testFile, []byte(testContent), 0644)
	require.NoError(t, err)

	sandbox := DefaultSandbox(tmpDir)
	tool := NewSearchReplaceTool(sandbox)

	ctx := context.Background()

	t.Run("with backup", func(t *testing.T) {
		input := map[string]interface{}{
			"pattern":     "foo",
			"replacement": "baz",
			"path":        testFile,
			"backup":      true,
			"dry_run":     false,
		}

		result, err := tool.Execute(ctx, input)

		require.NoError(t, err)
		assert.True(t, result.Success)

		// Verify backup exists
		backupFile := testFile + ".bak"
		_, err = os.Stat(backupFile)
		assert.NoError(t, err)
	})

	// Create a new test file for the no-backup test
	testFile2 := filepath.Join(tmpDir, "test2.txt")
	err = os.WriteFile(testFile2, []byte(testContent), 0644)
	require.NoError(t, err)

	t.Run("without backup", func(t *testing.T) {
		input := map[string]interface{}{
			"pattern":     "foo",
			"replacement": "baz",
			"path":        testFile2,
			"backup":      false,
			"dry_run":     false,
		}

		result, err := tool.Execute(ctx, input)

		require.NoError(t, err)
		assert.True(t, result.Success)

		// Verify no backup exists
		backupFile := testFile2 + ".bak"
		_, err = os.Stat(backupFile)
		assert.True(t, os.IsNotExist(err))
	})
}

func TestSearchReplaceTool_Execute_FilePermissions(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "replace-perms-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	testContent := "foo bar"
	testFile := filepath.Join(tmpDir, "test.txt")
	err = os.WriteFile(testFile, []byte(testContent), 0755) // Executable permissions
	require.NoError(t, err)

	sandbox := DefaultSandbox(tmpDir)
	tool := NewSearchReplaceTool(sandbox)

	ctx := context.Background()
	input := map[string]interface{}{
		"pattern":     "foo",
		"replacement": "baz",
		"path":        testFile,
		"backup":      false,
		"dry_run":     false,
	}

	result, err := tool.Execute(ctx, input)

	require.NoError(t, err)
	assert.True(t, result.Success)

	// Verify permissions are preserved
	fileInfo, err := os.Stat(testFile)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0755), fileInfo.Mode().Perm())
}
