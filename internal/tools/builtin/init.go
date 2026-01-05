// Package builtin provides built-in tools for the tool execution framework.
package builtin

import (
	"os"
	"path/filepath"

	"github.com/AINative-studio/ainative-code/internal/tools"
)

// RegisterCoreTools registers all core built-in tools with the provided registry.
// This includes bash execution, file operations, grep, and search/replace functionality.
//
// Parameters:
//   - registry: The tool registry to register tools with
//   - workingDir: The working directory for tool execution (used for sandboxing)
//   - allowedPaths: Paths that tools are allowed to access (for security sandboxing)
//
// Returns an error if any tool registration fails.
func RegisterCoreTools(registry *tools.Registry, workingDir string, allowedPaths []string) error {
	// Create sandbox for all tools
	sandbox := &Sandbox{
		AllowedPaths:     allowedPaths,
		AllowedCommands:  DefaultAllowedCommands(),
		DeniedCommands:   DefaultDeniedCommands(),
		WorkingDirectory: workingDir,
		MaxFileSize:      100 * 1024 * 1024, // 100MB
		MaxOutputSize:    10 * 1024 * 1024,  // 10MB
		AuditLog:         true,
	}

	// Register bash tool
	bashTool := NewBashTool(sandbox)
	if err := registry.Register(bashTool); err != nil {
		return err
	}

	// Register read file tool (already exists in codebase)
	readFileTool := NewReadFileTool(allowedPaths)
	if err := registry.Register(readFileTool); err != nil {
		return err
	}

	// Register write file tool (already exists in codebase)
	writeFileTool := NewWriteFileTool(allowedPaths)
	if err := registry.Register(writeFileTool); err != nil {
		return err
	}

	// Register grep tool
	grepTool := NewGrepTool(sandbox)
	if err := registry.Register(grepTool); err != nil {
		return err
	}

	// Register search/replace tool
	searchReplaceTool := NewSearchReplaceTool(sandbox)
	if err := registry.Register(searchReplaceTool); err != nil {
		return err
	}

	return nil
}

// RegisterCoreToolsWithDefaults registers all core tools with default configuration.
// It uses the current working directory and allows access to it recursively.
//
// This is a convenience function for quick setup. For production use, prefer
// RegisterCoreTools with explicit configuration.
func RegisterCoreToolsWithDefaults(registry *tools.Registry) error {
	workingDir, err := os.Getwd()
	if err != nil {
		// Fallback to temp directory if current directory can't be determined
		workingDir = os.TempDir()
	}

	allowedPaths := []string{workingDir}

	return RegisterCoreTools(registry, workingDir, allowedPaths)
}

// NewRegistryWithCoreTools creates a new registry and registers all core tools with default settings.
// This is the simplest way to get a fully configured registry for immediate use.
//
// Example:
//
//	registry, err := builtin.NewRegistryWithCoreTools()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Execute a tool
//	result, err := registry.Execute(ctx, "bash", map[string]interface{}{
//	    "command": "ls -la",
//	}, nil)
func NewRegistryWithCoreTools() (*tools.Registry, error) {
	registry := tools.NewRegistry()
	err := RegisterCoreToolsWithDefaults(registry)
	if err != nil {
		return nil, err
	}
	return registry, nil
}

// GetDefaultSandbox returns a sandbox instance with default security settings
// for the given working directory and allowed paths.
func GetDefaultSandbox(workingDir string, allowedPaths []string) *Sandbox {
	if workingDir == "" {
		workingDir, _ = os.Getwd()
		if workingDir == "" {
			workingDir = os.TempDir()
		}
	}

	if len(allowedPaths) == 0 {
		allowedPaths = []string{workingDir}
	}

	return &Sandbox{
		AllowedPaths:     allowedPaths,
		AllowedCommands:  DefaultAllowedCommands(),
		DeniedCommands:   DefaultDeniedCommands(),
		WorkingDirectory: workingDir,
		MaxFileSize:      100 * 1024 * 1024, // 100MB
		MaxOutputSize:    10 * 1024 * 1024,  // 10MB
		AuditLog:         true,
	}
}

// ExpandAllowedPaths expands the allowed paths to include commonly needed directories
// relative to the working directory.
func ExpandAllowedPaths(workingDir string) []string {
	paths := []string{workingDir}

	// Add common subdirectories if they exist
	commonDirs := []string{
		"src",
		"internal",
		"pkg",
		"cmd",
		"test",
		"tests",
		"docs",
		"scripts",
	}

	for _, dir := range commonDirs {
		fullPath := filepath.Join(workingDir, dir)
		if info, err := os.Stat(fullPath); err == nil && info.IsDir() {
			paths = append(paths, fullPath)
		}
	}

	return paths
}
