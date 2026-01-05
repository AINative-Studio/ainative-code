// Package builtin provides built-in tools for the tool execution framework.
package builtin

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AINative-studio/ainative-code/internal/tools"
)

// Sandbox provides security sandboxing for tool operations.
// It enforces path restrictions, command whitelisting, and resource limits.
type Sandbox struct {
	// AllowedPaths defines paths that tools are allowed to access
	AllowedPaths []string

	// AllowedCommands defines whitelisted commands for bash execution
	AllowedCommands []string

	// DeniedCommands defines explicitly blacklisted commands
	DeniedCommands []string

	// WorkingDirectory is the base directory for relative path operations
	WorkingDirectory string

	// MaxFileSize is the maximum file size that can be read/written (in bytes)
	MaxFileSize int64

	// MaxOutputSize is the maximum output size for command execution (in bytes)
	MaxOutputSize int64

	// AuditLog enables logging of all sandbox operations
	AuditLog bool
}

// DefaultSandbox returns a Sandbox with sensible default security settings.
func DefaultSandbox(workingDir string) *Sandbox {
	return &Sandbox{
		AllowedPaths:     []string{workingDir},
		AllowedCommands:  DefaultAllowedCommands(),
		DeniedCommands:   DefaultDeniedCommands(),
		WorkingDirectory: workingDir,
		MaxFileSize:      100 * 1024 * 1024, // 100MB
		MaxOutputSize:    10 * 1024 * 1024,  // 10MB
		AuditLog:         true,
	}
}

// DefaultAllowedCommands returns a whitelist of safe commands for execution.
func DefaultAllowedCommands() []string {
	return []string{
		// File operations (safe variants)
		"ls", "cat", "grep", "find", "head", "tail", "wc",
		"file", "stat", "du", "diff", "sort", "uniq",

		// Text processing
		"sed", "awk", "cut", "tr", "paste", "column",

		// Version control
		"git",

		// Build tools
		"make", "go", "npm", "yarn", "pip", "cargo",
		"docker", "kubectl",

		// Testing
		"pytest", "jest", "mocha", "cargo test",

		// Utilities
		"echo", "printf", "date", "env", "which", "whereis",
		"pwd", "basename", "dirname", "realpath",
	}
}

// DefaultDeniedCommands returns a blacklist of dangerous commands.
func DefaultDeniedCommands() []string {
	return []string{
		// Destructive file operations
		"rm", "rmdir", "dd", "shred",

		// System modification
		"chmod", "chown", "chgrp", "useradd", "usermod", "userdel",
		"groupadd", "groupmod", "groupdel",

		// Package management (can modify system)
		"apt", "apt-get", "yum", "dnf", "pacman", "brew install",

		// Network operations (potential security risk)
		"nc", "netcat", "telnet", "curl", "wget",

		// System control
		"shutdown", "reboot", "init", "systemctl", "service",
		"kill", "killall", "pkill",

		// Shell manipulation
		"exec", "eval", "source",

		// Disk operations
		"mount", "umount", "fdisk", "mkfs", "fsck",

		// Kernel modules
		"modprobe", "insmod", "rmmod",
	}
}

// ValidatePath checks if a path is allowed by the sandbox.
// It resolves the path to an absolute path and checks against allowed paths.
func (s *Sandbox) ValidatePath(path string) error {
	// Resolve to absolute path
	absPath, err := s.resolveAbsolutePath(path)
	if err != nil {
		return &tools.ErrInvalidInput{
			Field:  "path",
			Reason: fmt.Sprintf("cannot resolve path: %v", err),
		}
	}

	// Clean the path to prevent directory traversal
	absPath = filepath.Clean(absPath)

	// Check if path is within allowed paths
	if len(s.AllowedPaths) == 0 {
		return &tools.ErrPermissionDenied{
			Operation: "access",
			Resource:  absPath,
			Reason:    "no allowed paths configured",
		}
	}

	for _, allowedPath := range s.AllowedPaths {
		absAllowedPath, err := filepath.Abs(allowedPath)
		if err != nil {
			continue // Skip invalid allowed paths
		}

		absAllowedPath = filepath.Clean(absAllowedPath)

		// Check if requested path is under allowed path
		relPath, err := filepath.Rel(absAllowedPath, absPath)
		if err != nil {
			continue
		}

		// If the relative path doesn't start with "..", it's within the allowed path
		if !strings.HasPrefix(relPath, "..") && !strings.HasPrefix(relPath, string(filepath.Separator)) {
			// Log if audit is enabled
			if s.AuditLog {
				fmt.Printf("[AUDIT] Path access granted: %s (within %s)\n", absPath, absAllowedPath)
			}
			return nil // Access granted
		}
	}

	// Path not in any allowed path
	return &tools.ErrPermissionDenied{
		Operation: "access",
		Resource:  absPath,
		Reason:    fmt.Sprintf("path is outside allowed paths: %v", s.AllowedPaths),
	}
}

// ValidateCommand checks if a command is allowed by the sandbox.
// It checks against both whitelist and blacklist.
func (s *Sandbox) ValidateCommand(command string) error {
	// Extract the base command (first word)
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return &tools.ErrInvalidInput{
			Field:  "command",
			Reason: "command cannot be empty",
		}
	}

	baseCommand := parts[0]

	// Check blacklist first (takes precedence)
	for _, deniedCmd := range s.DeniedCommands {
		if matchesCommand(baseCommand, deniedCmd) || strings.Contains(command, deniedCmd) {
			return &tools.ErrPermissionDenied{
				Operation: "execute",
				Resource:  command,
				Reason:    fmt.Sprintf("command '%s' is explicitly denied", deniedCmd),
			}
		}
	}

	// Check for dangerous patterns
	if err := s.checkDangerousPatterns(command); err != nil {
		return err
	}

	// Check whitelist
	if len(s.AllowedCommands) > 0 {
		allowed := false
		for _, allowedCmd := range s.AllowedCommands {
			if matchesCommand(baseCommand, allowedCmd) {
				allowed = true
				break
			}
		}

		if !allowed {
			return &tools.ErrPermissionDenied{
				Operation: "execute",
				Resource:  command,
				Reason:    fmt.Sprintf("command '%s' is not in the allowed list", baseCommand),
			}
		}
	}

	// Log if audit is enabled
	if s.AuditLog {
		fmt.Printf("[AUDIT] Command execution granted: %s\n", command)
	}

	return nil
}

// checkDangerousPatterns checks for dangerous command patterns.
func (s *Sandbox) checkDangerousPatterns(command string) error {
	dangerousPatterns := []struct {
		pattern string
		reason  string
	}{
		{"rm -rf /", "attempt to delete root filesystem"},
		{"rm -rf /*", "attempt to delete root filesystem"},
		{":(){ :|:& };:", "fork bomb detected"},
		{"> /dev/sda", "attempt to overwrite disk device"},
		{"mkfs", "attempt to format filesystem"},
		{"/dev/sd", "direct disk device access"},
		{"chmod -R 777", "attempt to make everything world-writable"},
		{"chmod 777", "unsafe permission modification"},
	}

	for _, dp := range dangerousPatterns {
		if strings.Contains(command, dp.pattern) {
			return &tools.ErrPermissionDenied{
				Operation: "execute",
				Resource:  command,
				Reason:    dp.reason,
			}
		}
	}

	return nil
}

// ResolveWorkingDirectory resolves a working directory path with sandbox validation.
func (s *Sandbox) ResolveWorkingDirectory(dir string) (string, error) {
	if dir == "" {
		dir = s.WorkingDirectory
	}

	absDir, err := s.resolveAbsolutePath(dir)
	if err != nil {
		return "", fmt.Errorf("cannot resolve working directory: %w", err)
	}

	// Validate the directory is within allowed paths
	if err := s.ValidatePath(absDir); err != nil {
		return "", fmt.Errorf("working directory not allowed: %w", err)
	}

	// Verify it's actually a directory
	info, err := os.Stat(absDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("working directory does not exist: %s", absDir)
		}
		return "", fmt.Errorf("cannot stat working directory: %w", err)
	}

	if !info.IsDir() {
		return "", fmt.Errorf("path is not a directory: %s", absDir)
	}

	return absDir, nil
}

// ValidateFileSize checks if a file size is within the allowed limit.
func (s *Sandbox) ValidateFileSize(size int64) error {
	if s.MaxFileSize > 0 && size > s.MaxFileSize {
		return &tools.ErrExecutionFailed{
			Reason: fmt.Sprintf("file size %d bytes exceeds maximum allowed size %d bytes", size, s.MaxFileSize),
		}
	}
	return nil
}

// ValidateOutputSize checks if output size is within the allowed limit.
func (s *Sandbox) ValidateOutputSize(size int64) error {
	if s.MaxOutputSize > 0 && size > s.MaxOutputSize {
		return &tools.ErrOutputTooLarge{
			OutputSize: size,
			MaxSize:    s.MaxOutputSize,
		}
	}
	return nil
}

// resolveAbsolutePath resolves a path to an absolute path.
// If the path is relative, it's resolved relative to WorkingDirectory.
func (s *Sandbox) resolveAbsolutePath(path string) (string, error) {
	if filepath.IsAbs(path) {
		return path, nil
	}

	// Relative path - resolve relative to working directory
	if s.WorkingDirectory == "" {
		return filepath.Abs(path)
	}

	return filepath.Join(s.WorkingDirectory, path), nil
}

// matchesCommand checks if a base command matches an allowed/denied command pattern.
func matchesCommand(baseCommand, pattern string) bool {
	// Exact match
	if baseCommand == pattern {
		return true
	}

	// Check if pattern is a path and matches the basename
	if strings.Contains(pattern, "/") {
		return filepath.Base(pattern) == baseCommand
	}

	return false
}
