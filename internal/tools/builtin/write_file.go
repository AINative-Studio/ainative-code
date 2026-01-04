// Package builtin provides built-in tools for the tool execution framework.
package builtin

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AINative-studio/ainative-code/internal/tools"
)

// WriteFileTool implements a tool for writing file contents with path sandboxing.
type WriteFileTool struct {
	allowedPaths []string
}

// NewWriteFileTool creates a new WriteFileTool instance.
func NewWriteFileTool(allowedPaths []string) *WriteFileTool {
	return &WriteFileTool{
		allowedPaths: allowedPaths,
	}
}

// Name returns the unique name of the tool.
func (t *WriteFileTool) Name() string {
	return "write_file"
}

// Description returns a human-readable description of what the tool does.
func (t *WriteFileTool) Description() string {
	return "Writes content to a file on the filesystem with path sandboxing and permission checks"
}

// Schema returns the JSON schema for the tool's input parameters.
func (t *WriteFileTool) Schema() tools.ToolSchema {
	maxPathLength := 4096
	maxContentLength := 10 * 1024 * 1024 // 10MB
	return tools.ToolSchema{
		Type: "object",
		Properties: map[string]tools.PropertyDef{
			"path": {
				Type:        "string",
				Description: "The absolute or relative path to the file to write",
				MaxLength:   &maxPathLength,
			},
			"content": {
				Type:        "string",
				Description: "The content to write to the file",
				MaxLength:   &maxContentLength,
			},
			"mode": {
				Type:        "string",
				Description: "Write mode: 'overwrite' to replace file, 'append' to add to end (default: overwrite)",
				Enum:        []string{"overwrite", "append"},
				Default:     "overwrite",
			},
			"create_dirs": {
				Type:        "boolean",
				Description: "Whether to create parent directories if they don't exist (default: false)",
				Default:     false,
			},
			"permissions": {
				Type:        "string",
				Description: "File permissions in octal format (e.g., '0644', default: '0644')",
				Pattern:     "^0[0-7]{3}$",
				Default:     "0644",
			},
		},
		Required: []string{"path", "content"},
	}
}

// Execute runs the tool with the provided input and returns the result.
func (t *WriteFileTool) Execute(ctx context.Context, input map[string]interface{}) (*tools.Result, error) {
	// Extract and validate path
	pathRaw, ok := input["path"]
	if !ok {
		return nil, &tools.ErrInvalidInput{
			ToolName: t.Name(),
			Field:    "path",
			Reason:   "path is required",
		}
	}

	path, ok := pathRaw.(string)
	if !ok {
		return nil, &tools.ErrInvalidInput{
			ToolName: t.Name(),
			Field:    "path",
			Reason:   fmt.Sprintf("path must be a string, got %T", pathRaw),
		}
	}

	// Resolve to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, &tools.ErrInvalidInput{
			ToolName: t.Name(),
			Field:    "path",
			Reason:   fmt.Sprintf("cannot resolve path: %v", err),
		}
	}

	// Clean the path to prevent directory traversal attacks
	absPath = filepath.Clean(absPath)

	// Check path sandboxing
	if len(t.allowedPaths) > 0 {
		if err := t.validatePathAccess(absPath); err != nil {
			return nil, err
		}
	}

	// Extract content
	contentRaw, ok := input["content"]
	if !ok {
		return nil, &tools.ErrInvalidInput{
			ToolName: t.Name(),
			Field:    "content",
			Reason:   "content is required",
		}
	}

	content, ok := contentRaw.(string)
	if !ok {
		return nil, &tools.ErrInvalidInput{
			ToolName: t.Name(),
			Field:    "content",
			Reason:   fmt.Sprintf("content must be a string, got %T", contentRaw),
		}
	}

	// Extract mode parameter with default
	mode := "overwrite"
	if modeRaw, exists := input["mode"]; exists {
		mode, ok = modeRaw.(string)
		if !ok {
			return nil, &tools.ErrInvalidInput{
				ToolName: t.Name(),
				Field:    "mode",
				Reason:   fmt.Sprintf("mode must be a string, got %T", modeRaw),
			}
		}
	}

	// Extract create_dirs parameter with default
	createDirs := false
	if createDirsRaw, exists := input["create_dirs"]; exists {
		createDirs, ok = createDirsRaw.(bool)
		if !ok {
			return nil, &tools.ErrInvalidInput{
				ToolName: t.Name(),
				Field:    "create_dirs",
				Reason:   fmt.Sprintf("create_dirs must be a boolean, got %T", createDirsRaw),
			}
		}
	}

	// Extract permissions parameter with default
	permissionsStr := "0644"
	if permissionsRaw, exists := input["permissions"]; exists {
		permissionsStr, ok = permissionsRaw.(string)
		if !ok {
			return nil, &tools.ErrInvalidInput{
				ToolName: t.Name(),
				Field:    "permissions",
				Reason:   fmt.Sprintf("permissions must be a string, got %T", permissionsRaw),
			}
		}
	}

	// Parse permissions
	var permissions os.FileMode
	_, err = fmt.Sscanf(permissionsStr, "%o", &permissions)
	if err != nil {
		return nil, &tools.ErrInvalidInput{
			ToolName: t.Name(),
			Field:    "permissions",
			Reason:   fmt.Sprintf("invalid permissions format: %s (expected octal like '0644')", permissionsStr),
		}
	}

	// Check if parent directory exists
	parentDir := filepath.Dir(absPath)
	if _, err := os.Stat(parentDir); os.IsNotExist(err) {
		if !createDirs {
			return nil, &tools.ErrExecutionFailed{
				ToolName: t.Name(),
				Reason:   fmt.Sprintf("parent directory does not exist: %s (use create_dirs=true to create it)", parentDir),
				Cause:    err,
			}
		}

		// Validate parent directory is also within allowed paths
		if len(t.allowedPaths) > 0 {
			if err := t.validatePathAccess(parentDir); err != nil {
				return nil, &tools.ErrPermissionDenied{
					ToolName:  t.Name(),
					Operation: "create directory",
					Resource:  parentDir,
					Reason:    "parent directory is not within allowed paths",
				}
			}
		}

		// Create parent directories
		if err := os.MkdirAll(parentDir, 0755); err != nil {
			if os.IsPermission(err) {
				return nil, &tools.ErrPermissionDenied{
					ToolName:  t.Name(),
					Operation: "create directory",
					Resource:  parentDir,
					Reason:    "insufficient permissions to create parent directories",
				}
			}
			return nil, &tools.ErrExecutionFailed{
				ToolName: t.Name(),
				Reason:   fmt.Sprintf("failed to create parent directories: %s", parentDir),
				Cause:    err,
			}
		}
	} else if err != nil {
		// Some other error occurred
		if os.IsPermission(err) {
			return nil, &tools.ErrPermissionDenied{
				ToolName:  t.Name(),
				Operation: "access",
				Resource:  parentDir,
				Reason:    "insufficient permissions to access parent directory",
			}
		}
		return nil, &tools.ErrExecutionFailed{
			ToolName: t.Name(),
			Reason:   fmt.Sprintf("cannot access parent directory: %s", parentDir),
			Cause:    err,
		}
	}

	// Check if file already exists
	var existingSize int64
	var isNewFile bool
	if fileInfo, err := os.Stat(absPath); err == nil {
		// File exists
		isNewFile = false
		existingSize = fileInfo.Size()

		// Check if it's a directory
		if fileInfo.IsDir() {
			return nil, &tools.ErrInvalidInput{
				ToolName: t.Name(),
				Field:    "path",
				Reason:   fmt.Sprintf("path is a directory, not a file: %s", absPath),
			}
		}
	} else if os.IsNotExist(err) {
		// File doesn't exist, will be created
		isNewFile = true
		existingSize = 0
	} else {
		// Some other error
		if os.IsPermission(err) {
			return nil, &tools.ErrPermissionDenied{
				ToolName:  t.Name(),
				Operation: "access",
				Resource:  absPath,
				Reason:    "insufficient permissions to access file",
			}
		}
		return nil, &tools.ErrExecutionFailed{
			ToolName: t.Name(),
			Reason:   fmt.Sprintf("cannot stat file: %s", absPath),
			Cause:    err,
		}
	}

	// Write the file
	var writeErr error
	var bytesWritten int

	switch mode {
	case "overwrite":
		// Write file with atomic operation using temp file
		tempPath := absPath + ".tmp"
		writeErr = os.WriteFile(tempPath, []byte(content), permissions)
		if writeErr == nil {
			// Atomically replace original file
			writeErr = os.Rename(tempPath, absPath)
			if writeErr != nil {
				// Clean up temp file on failure
				os.Remove(tempPath)
			} else {
				bytesWritten = len(content)
			}
		}

	case "append":
		// Open file in append mode
		file, err := os.OpenFile(absPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, permissions)
		if err != nil {
			writeErr = err
		} else {
			defer file.Close()
			n, err := file.WriteString(content)
			bytesWritten = n
			writeErr = err
		}

	default:
		return nil, &tools.ErrInvalidInput{
			ToolName: t.Name(),
			Field:    "mode",
			Reason:   fmt.Sprintf("unsupported mode: %s (must be 'overwrite' or 'append')", mode),
		}
	}

	// Handle write errors
	if writeErr != nil {
		if os.IsPermission(writeErr) {
			return nil, &tools.ErrPermissionDenied{
				ToolName:  t.Name(),
				Operation: "write",
				Resource:  absPath,
				Reason:    "insufficient permissions to write file",
			}
		}
		return nil, &tools.ErrExecutionFailed{
			ToolName: t.Name(),
			Reason:   fmt.Sprintf("failed to write file: %s", absPath),
			Cause:    writeErr,
		}
	}

	// Get final file info
	fileInfo, err := os.Stat(absPath)
	if err != nil {
		// File was written but we can't stat it - still report success
		return &tools.Result{
			Success: true,
			Output:  fmt.Sprintf("Successfully wrote %d bytes to %s", bytesWritten, absPath),
			Metadata: map[string]interface{}{
				"path":          absPath,
				"bytes_written": bytesWritten,
				"mode":          mode,
				"is_new_file":   isNewFile,
			},
		}, nil
	}

	// Build result with metadata
	result := &tools.Result{
		Success: true,
		Output:  fmt.Sprintf("Successfully wrote %d bytes to %s", bytesWritten, absPath),
		Metadata: map[string]interface{}{
			"path":                absPath,
			"bytes_written":       bytesWritten,
			"mode":                mode,
			"is_new_file":         isNewFile,
			"previous_size_bytes": existingSize,
			"final_size_bytes":    fileInfo.Size(),
			"permissions":         fileInfo.Mode().String(),
			"modified_time":       fileInfo.ModTime().Format("2006-01-02T15:04:05Z07:00"),
		},
	}

	return result, nil
}

// Category returns the category this tool belongs to.
func (t *WriteFileTool) Category() tools.Category {
	return tools.CategoryFilesystem
}

// RequiresConfirmation returns true if this tool requires user confirmation before execution.
func (t *WriteFileTool) RequiresConfirmation() bool {
	return true // Writing files modifies state and should require confirmation
}

// validatePathAccess checks if the given path is within the allowed paths.
func (t *WriteFileTool) validatePathAccess(requestedPath string) error {
	// If no allowed paths configured, deny all access
	if len(t.allowedPaths) == 0 {
		return &tools.ErrPermissionDenied{
			ToolName:  t.Name(),
			Operation: "write",
			Resource:  requestedPath,
			Reason:    "no allowed paths configured, file access denied",
		}
	}

	// Check if requested path is within any allowed path
	for _, allowedPath := range t.allowedPaths {
		// Resolve allowed path to absolute
		absAllowedPath, err := filepath.Abs(allowedPath)
		if err != nil {
			continue // Skip invalid allowed paths
		}

		// Clean the allowed path
		absAllowedPath = filepath.Clean(absAllowedPath)

		// Check if requested path is under allowed path
		relPath, err := filepath.Rel(absAllowedPath, requestedPath)
		if err != nil {
			continue
		}

		// If the relative path doesn't start with "..", it's within the allowed path
		if !strings.HasPrefix(relPath, "..") {
			return nil // Access granted
		}
	}

	// Path not in any allowed path
	return &tools.ErrPermissionDenied{
		ToolName:  t.Name(),
		Operation: "write",
		Resource:  requestedPath,
		Reason:    fmt.Sprintf("path is not within allowed paths: %v", t.allowedPaths),
	}
}
