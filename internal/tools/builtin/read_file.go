// Package builtin provides built-in tools for the tool execution framework.
package builtin

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ainative/ainative-code/internal/tools"
)

// ReadFileTool implements a tool for reading file contents with path sandboxing.
type ReadFileTool struct {
	allowedPaths []string
}

// NewReadFileTool creates a new ReadFileTool instance.
func NewReadFileTool(allowedPaths []string) *ReadFileTool {
	return &ReadFileTool{
		allowedPaths: allowedPaths,
	}
}

// Name returns the unique name of the tool.
func (t *ReadFileTool) Name() string {
	return "read_file"
}

// Description returns a human-readable description of what the tool does.
func (t *ReadFileTool) Description() string {
	return "Reads the contents of a file from the filesystem with path sandboxing and permission checks"
}

// Schema returns the JSON schema for the tool's input parameters.
func (t *ReadFileTool) Schema() tools.ToolSchema {
	maxPathLength := 4096
	return tools.ToolSchema{
		Type: "object",
		Properties: map[string]tools.PropertyDef{
			"path": {
				Type:        "string",
				Description: "The absolute or relative path to the file to read",
				MaxLength:   &maxPathLength,
			},
			"max_size": {
				Type:        "integer",
				Description: "Maximum file size to read in bytes (default: 10MB, max: 100MB)",
			},
			"encoding": {
				Type:        "string",
				Description: "File encoding to use for reading (default: utf-8)",
				Enum:        []string{"utf-8", "ascii", "binary"},
				Default:     "utf-8",
			},
		},
		Required: []string{"path"},
	}
}

// Execute runs the tool with the provided input and returns the result.
func (t *ReadFileTool) Execute(ctx context.Context, input map[string]interface{}) (*tools.Result, error) {
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

	// Extract max_size parameter with default
	maxSize := int64(10 * 1024 * 1024) // 10MB default
	if maxSizeRaw, exists := input["max_size"]; exists {
		// Handle JSON unmarshaling float64 for all numbers
		switch v := maxSizeRaw.(type) {
		case float64:
			maxSize = int64(v)
		case int:
			maxSize = int64(v)
		case int64:
			maxSize = v
		default:
			return nil, &tools.ErrInvalidInput{
				ToolName: t.Name(),
				Field:    "max_size",
				Reason:   fmt.Sprintf("max_size must be an integer, got %T", maxSizeRaw),
			}
		}

		// Validate max_size bounds
		if maxSize <= 0 {
			return nil, &tools.ErrInvalidInput{
				ToolName: t.Name(),
				Field:    "max_size",
				Reason:   "max_size must be positive",
			}
		}
		if maxSize > 100*1024*1024 { // 100MB absolute maximum
			return nil, &tools.ErrInvalidInput{
				ToolName: t.Name(),
				Field:    "max_size",
				Reason:   "max_size cannot exceed 100MB",
			}
		}
	}

	// Extract encoding parameter with default
	encoding := "utf-8"
	if encodingRaw, exists := input["encoding"]; exists {
		var ok bool
		encoding, ok = encodingRaw.(string)
		if !ok {
			return nil, &tools.ErrInvalidInput{
				ToolName: t.Name(),
				Field:    "encoding",
				Reason:   fmt.Sprintf("encoding must be a string, got %T", encodingRaw),
			}
		}
	}

	// Check if file exists
	fileInfo, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &tools.ErrExecutionFailed{
				ToolName: t.Name(),
				Reason:   fmt.Sprintf("file does not exist: %s", absPath),
				Cause:    err,
			}
		}
		if os.IsPermission(err) {
			return nil, &tools.ErrPermissionDenied{
				ToolName:  t.Name(),
				Operation: "read",
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

	// Check if it's a directory
	if fileInfo.IsDir() {
		return nil, &tools.ErrInvalidInput{
			ToolName: t.Name(),
			Field:    "path",
			Reason:   fmt.Sprintf("path is a directory, not a file: %s", absPath),
		}
	}

	// Check file size
	fileSize := fileInfo.Size()
	if fileSize > maxSize {
		return nil, &tools.ErrExecutionFailed{
			ToolName: t.Name(),
			Reason:   fmt.Sprintf("file size %d bytes exceeds max_size %d bytes", fileSize, maxSize),
		}
	}

	// Read file contents
	content, err := os.ReadFile(absPath)
	if err != nil {
		if os.IsPermission(err) {
			return nil, &tools.ErrPermissionDenied{
				ToolName:  t.Name(),
				Operation: "read",
				Resource:  absPath,
				Reason:    "insufficient permissions to read file",
			}
		}
		return nil, &tools.ErrExecutionFailed{
			ToolName: t.Name(),
			Reason:   fmt.Sprintf("failed to read file: %s", absPath),
			Cause:    err,
		}
	}

	// Handle encoding
	var output string
	switch encoding {
	case "utf-8", "ascii":
		output = string(content)
	case "binary":
		// For binary, return hex representation
		output = fmt.Sprintf("Binary content (%d bytes): %x", len(content), content)
	default:
		// Should not happen due to enum validation, but handle defensively
		output = string(content)
	}

	// Build result with metadata
	result := &tools.Result{
		Success: true,
		Output:  output,
		Metadata: map[string]interface{}{
			"path":         absPath,
			"size_bytes":   fileSize,
			"encoding":     encoding,
			"permissions":  fileInfo.Mode().String(),
			"modified_time": fileInfo.ModTime().Format("2006-01-02T15:04:05Z07:00"),
		},
	}

	return result, nil
}

// Category returns the category this tool belongs to.
func (t *ReadFileTool) Category() tools.Category {
	return tools.CategoryFilesystem
}

// RequiresConfirmation returns true if this tool requires user confirmation before execution.
func (t *ReadFileTool) RequiresConfirmation() bool {
	return false // Reading files is generally safe and doesn't require confirmation
}

// validatePathAccess checks if the given path is within the allowed paths.
func (t *ReadFileTool) validatePathAccess(requestedPath string) error {
	// If no allowed paths configured, deny all access
	if len(t.allowedPaths) == 0 {
		return &tools.ErrPermissionDenied{
			ToolName:  t.Name(),
			Operation: "read",
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
		Operation: "read",
		Resource:  requestedPath,
		Reason:    fmt.Sprintf("path is not within allowed paths: %v", t.allowedPaths),
	}
}
