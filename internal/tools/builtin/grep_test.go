package builtin

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/AINative-studio/ainative-code/internal/tools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGrepTool(t *testing.T) {
	sandbox := DefaultSandbox("/tmp")
	tool := NewGrepTool(sandbox)

	assert.NotNil(t, tool)
	assert.Equal(t, "grep", tool.Name())
	assert.NotEmpty(t, tool.Description())
	assert.Equal(t, tools.CategoryText, tool.Category())
	assert.False(t, tool.RequiresConfirmation())
}

func TestGrepTool_Schema(t *testing.T) {
	sandbox := DefaultSandbox("/tmp")
	tool := NewGrepTool(sandbox)

	schema := tool.Schema()

	assert.Equal(t, "object", schema.Type)
	assert.Contains(t, schema.Properties, "pattern")
	assert.Contains(t, schema.Properties, "path")
	assert.Contains(t, schema.Properties, "file_pattern")
	assert.Contains(t, schema.Properties, "recursive")
	assert.Contains(t, schema.Properties, "case_sensitive")
	assert.Equal(t, []string{"pattern"}, schema.Required)
}

func TestGrepTool_Execute_Success(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "grep-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test files
	testContent := `line 1: hello world
line 2: test pattern
line 3: another line
line 4: PATTERN in caps
line 5: final line`

	testFile := filepath.Join(tmpDir, "test.txt")
	err = os.WriteFile(testFile, []byte(testContent), 0644)
	require.NoError(t, err)

	sandbox := DefaultSandbox(tmpDir)
	tool := NewGrepTool(sandbox)

	tests := []struct {
		name           string
		input          map[string]interface{}
		expectMatches  bool
		outputContains string
	}{
		{
			name: "simple pattern match",
			input: map[string]interface{}{
				"pattern": "pattern",
				"path":    testFile,
			},
			expectMatches:  true,
			outputContains: "line 2",
		},
		{
			name: "case insensitive match",
			input: map[string]interface{}{
				"pattern":        "PATTERN",
				"path":           testFile,
				"case_sensitive": false,
			},
			expectMatches:  true,
			outputContains: "line 2",
		},
		{
			name: "case sensitive no match",
			input: map[string]interface{}{
				"pattern":        "HELLO",
				"path":           testFile,
				"case_sensitive": true,
			},
			expectMatches:  false,
			outputContains: "No matches",
		},
		{
			name: "regex pattern",
			input: map[string]interface{}{
				"pattern": "line \\d+:",
				"path":    testFile,
			},
			expectMatches: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := tool.Execute(ctx, tt.input)

			require.NoError(t, err)
			assert.NotNil(t, result)
			assert.True(t, result.Success)

			if tt.outputContains != "" {
				assert.Contains(t, result.Output, tt.outputContains)
			}
		})
	}
}

func TestGrepTool_Execute_DirectorySearch(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "grep-dir-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create multiple test files
	files := map[string]string{
		"file1.txt": "test pattern in file1",
		"file2.txt": "another pattern here",
		"file3.go":  "// pattern in go file",
		"file4.md":  "# pattern in markdown",
	}

	for name, content := range files {
		err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0644)
		require.NoError(t, err)
	}

	sandbox := DefaultSandbox(tmpDir)
	tool := NewGrepTool(sandbox)

	ctx := context.Background()

	t.Run("search all files", func(t *testing.T) {
		input := map[string]interface{}{
			"pattern":   "pattern",
			"path":      tmpDir,
			"recursive": false,
		}

		result, err := tool.Execute(ctx, input)

		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.Contains(t, result.Output, "file1.txt")
		assert.Contains(t, result.Metadata, "total_matches")
	})

	t.Run("search with file pattern", func(t *testing.T) {
		input := map[string]interface{}{
			"pattern":      "pattern",
			"path":         tmpDir,
			"file_pattern": "*.go",
			"recursive":    false,
		}

		result, err := tool.Execute(ctx, input)

		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.Contains(t, result.Output, "file3.go")
		assert.NotContains(t, result.Output, "file1.txt")
	})
}

func TestGrepTool_Execute_ContextLines(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "grep-context-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	testContent := `before1
before2
MATCH LINE
after1
after2`

	testFile := filepath.Join(tmpDir, "context.txt")
	err = os.WriteFile(testFile, []byte(testContent), 0644)
	require.NoError(t, err)

	sandbox := DefaultSandbox(tmpDir)
	tool := NewGrepTool(sandbox)

	ctx := context.Background()

	t.Run("with before context", func(t *testing.T) {
		input := map[string]interface{}{
			"pattern":        "MATCH",
			"path":           testFile,
			"context_before": 2,
		}

		result, err := tool.Execute(ctx, input)

		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.Contains(t, result.Output, "before1")
		assert.Contains(t, result.Output, "before2")
	})

	t.Run("with after context", func(t *testing.T) {
		input := map[string]interface{}{
			"pattern":       "MATCH",
			"path":          testFile,
			"context_after": 2,
		}

		result, err := tool.Execute(ctx, input)

		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.Contains(t, result.Output, "after1")
		assert.Contains(t, result.Output, "after2")
	})
}

func TestGrepTool_Execute_ValidationErrors(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "grep-validation-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	sandbox := DefaultSandbox(tmpDir)
	tool := NewGrepTool(sandbox)

	tests := []struct {
		name          string
		input         map[string]interface{}
		errorType     interface{}
		errorContains string
	}{
		{
			name:          "missing pattern",
			input:         map[string]interface{}{},
			errorType:     &tools.ErrInvalidInput{},
			errorContains: "pattern is required",
		},
		{
			name: "empty pattern",
			input: map[string]interface{}{
				"pattern": "",
			},
			errorType:     &tools.ErrInvalidInput{},
			errorContains: "cannot be empty",
		},
		{
			name: "invalid regex pattern",
			input: map[string]interface{}{
				"pattern": "[invalid(regex",
			},
			errorType:     &tools.ErrInvalidInput{},
			errorContains: "invalid regex",
		},
		{
			name: "invalid pattern type",
			input: map[string]interface{}{
				"pattern": 123,
			},
			errorType:     &tools.ErrInvalidInput{},
			errorContains: "must be a string",
		},
		{
			name: "context_before too large",
			input: map[string]interface{}{
				"pattern":        "test",
				"context_before": 20,
			},
			errorType:     &tools.ErrInvalidInput{},
			errorContains: "between 0 and 10",
		},
		{
			name: "context_after negative",
			input: map[string]interface{}{
				"pattern":       "test",
				"context_after": -1,
			},
			errorType:     &tools.ErrInvalidInput{},
			errorContains: "between 0 and 10",
		},
		{
			name: "max_matches too large",
			input: map[string]interface{}{
				"pattern":     "test",
				"max_matches": 20000,
			},
			errorType:     &tools.ErrInvalidInput{},
			errorContains: "between 1 and 10000",
		},
		{
			name: "max_files invalid",
			input: map[string]interface{}{
				"pattern":   "test",
				"max_files": 0,
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

func TestGrepTool_Execute_PathValidation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "grep-path-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	sandbox := DefaultSandbox(tmpDir)
	tool := NewGrepTool(sandbox)

	ctx := context.Background()

	t.Run("path outside sandbox", func(t *testing.T) {
		input := map[string]interface{}{
			"pattern": "test",
			"path":    "/etc/passwd",
		}

		result, err := tool.Execute(ctx, input)

		assert.Nil(t, result)
		assert.Error(t, err)
		assert.IsType(t, &tools.ErrPermissionDenied{}, err)
	})

	t.Run("non-existent path", func(t *testing.T) {
		input := map[string]interface{}{
			"pattern": "test",
			"path":    filepath.Join(tmpDir, "nonexistent"),
		}

		result, err := tool.Execute(ctx, input)

		assert.Nil(t, result)
		assert.Error(t, err)
		assert.IsType(t, &tools.ErrExecutionFailed{}, err)
		assert.Contains(t, err.Error(), "does not exist")
	})
}

func TestGrepTool_Execute_LineNumbers(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "grep-linenum-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	testContent := "line 1\nmatch here\nline 3"
	testFile := filepath.Join(tmpDir, "test.txt")
	err = os.WriteFile(testFile, []byte(testContent), 0644)
	require.NoError(t, err)

	sandbox := DefaultSandbox(tmpDir)
	tool := NewGrepTool(sandbox)

	ctx := context.Background()

	t.Run("with line numbers", func(t *testing.T) {
		input := map[string]interface{}{
			"pattern":           "match",
			"path":              testFile,
			"show_line_numbers": true,
		}

		result, err := tool.Execute(ctx, input)

		require.NoError(t, err)
		assert.Contains(t, result.Output, "2:")
	})

	t.Run("without line numbers", func(t *testing.T) {
		input := map[string]interface{}{
			"pattern":           "match",
			"path":              testFile,
			"show_line_numbers": false,
		}

		result, err := tool.Execute(ctx, input)

		require.NoError(t, err)
		assert.NotContains(t, result.Output, "2:")
	})
}

func TestGrepTool_Execute_MaxLimits(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "grep-limits-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create multiple files with matches
	for i := 0; i < 10; i++ {
		content := "match\nmatch\nmatch\n"
		filename := filepath.Join(tmpDir, fmt.Sprintf("file%d.txt", i))
		err := os.WriteFile(filename, []byte(content), 0644)
		require.NoError(t, err)
	}

	sandbox := DefaultSandbox(tmpDir)
	tool := NewGrepTool(sandbox)

	ctx := context.Background()

	t.Run("max matches limit", func(t *testing.T) {
		input := map[string]interface{}{
			"pattern":     "match",
			"path":        tmpDir,
			"max_matches": 5,
			"recursive":   false,
		}

		result, err := tool.Execute(ctx, input)

		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.Equal(t, true, result.Metadata["truncated"])
	})

	t.Run("max files limit", func(t *testing.T) {
		input := map[string]interface{}{
			"pattern":   "match",
			"path":      tmpDir,
			"max_files": 3,
			"recursive": false,
		}

		result, err := tool.Execute(ctx, input)

		require.NoError(t, err)
		assert.True(t, result.Success)
		filesSearched := result.Metadata["files_searched"].(int)
		assert.LessOrEqual(t, filesSearched, 3)
	})
}

func TestGrepTool_Execute_RecursiveSearch(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "grep-recursive-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create nested directory structure
	subDir := filepath.Join(tmpDir, "subdir")
	err = os.Mkdir(subDir, 0755)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(tmpDir, "root.txt"), []byte("match in root"), 0644)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(subDir, "sub.txt"), []byte("match in subdir"), 0644)
	require.NoError(t, err)

	sandbox := DefaultSandbox(tmpDir)
	tool := NewGrepTool(sandbox)

	ctx := context.Background()

	t.Run("recursive search", func(t *testing.T) {
		input := map[string]interface{}{
			"pattern":   "match",
			"path":      tmpDir,
			"recursive": true,
		}

		result, err := tool.Execute(ctx, input)

		require.NoError(t, err)
		assert.Contains(t, result.Output, "root.txt")
		assert.Contains(t, result.Output, "sub.txt")
	})

	t.Run("non-recursive search", func(t *testing.T) {
		input := map[string]interface{}{
			"pattern":   "match",
			"path":      tmpDir,
			"recursive": false,
		}

		result, err := tool.Execute(ctx, input)

		require.NoError(t, err)
		assert.Contains(t, result.Output, "root.txt")
		assert.NotContains(t, result.Output, "sub.txt")
	})
}
