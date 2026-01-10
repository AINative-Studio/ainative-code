package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateInputFile(t *testing.T) {
	// Create temporary test directory
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		setup     func() string // Returns file path to test
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid readable file",
			setup: func() string {
				path := filepath.Join(tmpDir, "valid.json")
				os.WriteFile(path, []byte(`{"test": "data"}`), 0644)
				return path
			},
			wantError: false,
		},
		{
			name: "empty file path",
			setup: func() string {
				return ""
			},
			wantError: true,
			errorMsg:  "file path cannot be empty",
		},
		{
			name: "nonexistent file",
			setup: func() string {
				return filepath.Join(tmpDir, "nonexistent.json")
			},
			wantError: true,
			errorMsg:  "file not found",
		},
		{
			name: "directory instead of file",
			setup: func() string {
				dirPath := filepath.Join(tmpDir, "testdir")
				os.Mkdir(dirPath, 0755)
				return dirPath
			},
			wantError: true,
			errorMsg:  "path is a directory, not a file",
		},
		{
			name: "empty file",
			setup: func() string {
				path := filepath.Join(tmpDir, "empty.json")
				os.WriteFile(path, []byte{}, 0644)
				return path
			},
			wantError: true,
			errorMsg:  "file is empty",
		},
		{
			name: "unreadable file (permission denied)",
			setup: func() string {
				path := filepath.Join(tmpDir, "unreadable.json")
				os.WriteFile(path, []byte(`{"test": "data"}`), 0644)
				os.Chmod(path, 0000)
				return path
			},
			wantError: true,
			errorMsg:  "cannot read file (permission denied)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup()

			// Cleanup: restore permissions for cleanup
			if tt.name == "unreadable file (permission denied)" {
				defer os.Chmod(path, 0644)
			}

			err := validateInputFile(path)

			if tt.wantError {
				if err == nil {
					t.Errorf("validateInputFile() expected error but got nil")
					return
				}
				if tt.errorMsg != "" && !stringContainsHelper(err.Error(), tt.errorMsg) {
					t.Errorf("validateInputFile() error = %v, want error containing %q", err, tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validateInputFile() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestValidateOutputPath(t *testing.T) {
	// Create temporary test directory
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		setup     func() string // Returns file path to test
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid writable directory",
			setup: func() string {
				return filepath.Join(tmpDir, "output.json")
			},
			wantError: false,
		},
		{
			name: "empty file path",
			setup: func() string {
				return ""
			},
			wantError: true,
			errorMsg:  "output file path cannot be empty",
		},
		{
			name: "nonexistent directory",
			setup: func() string {
				return filepath.Join(tmpDir, "nonexistent", "output.json")
			},
			wantError: true,
			errorMsg:  "output directory does not exist",
		},
		{
			name: "read-only directory",
			setup: func() string {
				dirPath := filepath.Join(tmpDir, "readonly")
				os.Mkdir(dirPath, 0755)
				os.Chmod(dirPath, 0555)
				return filepath.Join(dirPath, "output.json")
			},
			wantError: true,
			errorMsg:  "permission denied",
		},
		{
			name: "parent is file not directory",
			setup: func() string {
				filePath := filepath.Join(tmpDir, "notadir")
				os.WriteFile(filePath, []byte("test"), 0644)
				return filepath.Join(filePath, "output.json")
			},
			wantError: true,
			errorMsg:  "parent path is not a directory",
		},
		{
			name: "existing writable file",
			setup: func() string {
				path := filepath.Join(tmpDir, "existing.json")
				os.WriteFile(path, []byte(`{"old": "data"}`), 0644)
				return path
			},
			wantError: false,
		},
		{
			name: "existing unwritable file",
			setup: func() string {
				path := filepath.Join(tmpDir, "readonly_file.json")
				os.WriteFile(path, []byte(`{"old": "data"}`), 0644)
				os.Chmod(path, 0444)
				return path
			},
			wantError: true,
			errorMsg:  "output file exists but is not writable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup()

			// Cleanup: restore permissions for cleanup
			switch tt.name {
			case "read-only directory":
				defer os.Chmod(filepath.Dir(path), 0755)
			case "existing unwritable file":
				defer os.Chmod(path, 0644)
			}

			err := validateOutputPath(path)

			if tt.wantError {
				if err == nil {
					t.Errorf("validateOutputPath() expected error but got nil")
					return
				}
				if tt.errorMsg != "" && !stringContainsHelper(err.Error(), tt.errorMsg) {
					t.Errorf("validateOutputPath() error = %v, want error containing %q", err, tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validateOutputPath() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestValidateInputFile_EdgeCases(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("symlink to valid file", func(t *testing.T) {
		targetPath := filepath.Join(tmpDir, "target.json")
		os.WriteFile(targetPath, []byte(`{"test": "data"}`), 0644)

		symlinkPath := filepath.Join(tmpDir, "link.json")
		if err := os.Symlink(targetPath, symlinkPath); err != nil {
			t.Skip("Cannot create symlinks on this system")
		}

		err := validateInputFile(symlinkPath)
		if err != nil {
			t.Errorf("validateInputFile() with symlink failed: %v", err)
		}
	})

	t.Run("symlink to nonexistent file", func(t *testing.T) {
		symlinkPath := filepath.Join(tmpDir, "broken_link.json")
		if err := os.Symlink("/nonexistent/file.json", symlinkPath); err != nil {
			t.Skip("Cannot create symlinks on this system")
		}

		err := validateInputFile(symlinkPath)
		if err == nil {
			t.Error("validateInputFile() with broken symlink should fail")
		}
	})

	t.Run("large file readable", func(t *testing.T) {
		path := filepath.Join(tmpDir, "large.json")
		// Create a file larger than 1KB
		data := make([]byte, 2048)
		for i := range data {
			data[i] = 'x'
		}
		os.WriteFile(path, data, 0644)

		err := validateInputFile(path)
		if err != nil {
			t.Errorf("validateInputFile() with large file failed: %v", err)
		}
	})
}

func TestValidateOutputPath_EdgeCases(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("deeply nested path", func(t *testing.T) {
		// Create nested directories
		nestedDir := filepath.Join(tmpDir, "a", "b", "c", "d")
		os.MkdirAll(nestedDir, 0755)

		path := filepath.Join(nestedDir, "output.json")
		err := validateOutputPath(path)
		if err != nil {
			t.Errorf("validateOutputPath() with nested path failed: %v", err)
		}
	})

	t.Run("path with spaces", func(t *testing.T) {
		dirPath := filepath.Join(tmpDir, "dir with spaces")
		os.Mkdir(dirPath, 0755)

		path := filepath.Join(dirPath, "output file.json")
		err := validateOutputPath(path)
		if err != nil {
			t.Errorf("validateOutputPath() with spaces failed: %v", err)
		}
	})

	t.Run("path with special characters", func(t *testing.T) {
		dirPath := filepath.Join(tmpDir, "dir-with_special.chars")
		os.Mkdir(dirPath, 0755)

		path := filepath.Join(dirPath, "output-file_123.json")
		err := validateOutputPath(path)
		if err != nil {
			t.Errorf("validateOutputPath() with special chars failed: %v", err)
		}
	})
}

// Helper function to check if a string contains a substring
func stringContainsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
