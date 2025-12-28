package config

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/AINative-studio/ainative-code/internal/errors"
)

// Resolver handles dynamic resolution of API keys and secrets from multiple sources
type Resolver struct {
	// commandTimeout is the maximum time allowed for command execution
	commandTimeout time.Duration

	// allowedCommands is a whitelist of commands allowed for execution
	// If empty, all commands are allowed (use with caution)
	allowedCommands []string

	// enableCommandExecution controls whether command execution is allowed
	enableCommandExecution bool
}

// ResolverOption is a functional option for configuring the Resolver
type ResolverOption func(*Resolver)

// NewResolver creates a new API key resolver with default settings
func NewResolver(opts ...ResolverOption) *Resolver {
	r := &Resolver{
		commandTimeout:         5 * time.Second,
		allowedCommands:        []string{},
		enableCommandExecution: true,
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// WithCommandTimeout sets the timeout for command execution
func WithCommandTimeout(timeout time.Duration) ResolverOption {
	return func(r *Resolver) {
		r.commandTimeout = timeout
	}
}

// WithAllowedCommands sets a whitelist of allowed commands
func WithAllowedCommands(commands ...string) ResolverOption {
	return func(r *Resolver) {
		r.allowedCommands = commands
	}
}

// WithCommandExecution enables or disables command execution
func WithCommandExecution(enabled bool) ResolverOption {
	return func(r *Resolver) {
		r.enableCommandExecution = enabled
	}
}

// Resolve dynamically resolves an API key from various sources
//
// Supported formats:
//   - Direct string: "sk-ant-api-key-123"
//   - Command execution: "$(pass show anthropic)"
//   - Environment variable: "${OPENAI_API_KEY}"
//   - File path: "~/secrets/api-key.txt" or "/path/to/key.txt"
//
// Resolution priority:
//   1. Command execution (if matches pattern)
//   2. Environment variable (if matches pattern)
//   3. File path (if looks like a path)
//   4. Direct string (fallback)
func (r *Resolver) Resolve(value string) (string, error) {
	if value == "" {
		return "", nil
	}

	// Trim whitespace
	value = strings.TrimSpace(value)

	// Check for command execution: $(command)
	if matched, result, err := r.resolveCommand(value); matched {
		if err != nil {
			return "", err
		}
		return result, nil
	}

	// Check for environment variable: ${VAR_NAME}
	if matched, result, err := r.resolveEnvVar(value); matched {
		if err != nil {
			return "", err
		}
		return result, nil
	}

	// Check for file path: ~/path or /path or ./path
	if matched, result, err := r.resolveFilePath(value); matched {
		if err != nil {
			return "", err
		}
		return result, nil
	}

	// Return as direct string
	return value, nil
}

// resolveCommand handles command execution pattern: $(command)
func (r *Resolver) resolveCommand(value string) (bool, string, error) {
	// Pattern: $(command args...)
	pattern := regexp.MustCompile(`^\$\((.*)\)$`)
	matches := pattern.FindStringSubmatch(value)

	if len(matches) != 2 {
		return false, "", nil
	}

	if !r.enableCommandExecution {
		return true, "", errors.NewSecurityError(
			errors.ErrCodeSecurityViolation,
			"Command execution is disabled",
			"command_execution",
		)
	}

	command := strings.TrimSpace(matches[1])
	if command == "" {
		return true, "", errors.NewConfigInvalidError(
			"api_key_command",
			"command cannot be empty",
		)
	}

	// Parse command and arguments
	parts := strings.Fields(command)
	cmdName := parts[0]
	cmdArgs := parts[1:]

	// Check if command is in whitelist (if whitelist is configured)
	if len(r.allowedCommands) > 0 {
		allowed := false
		for _, allowedCmd := range r.allowedCommands {
			if cmdName == allowedCmd {
				allowed = true
				break
			}
		}
		if !allowed {
			return true, "", errors.NewSecurityError(
				errors.ErrCodeSecurityViolation,
				fmt.Sprintf("command '%s' is not in the allowed commands list", cmdName),
				"command_execution",
			)
		}
	}

	// Execute command with timeout
	ctx, cancel := context.WithTimeout(context.Background(), r.commandTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, cmdName, cmdArgs...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return true, "", errors.NewConfigInvalidError(
				"api_key_command",
				fmt.Sprintf("command execution timed out after %v", r.commandTimeout),
			)
		}
		return true, "", errors.NewConfigInvalidError(
			"api_key_command",
			fmt.Sprintf("command execution failed: %v - output: %s", err, string(output)),
		)
	}

	result := strings.TrimSpace(string(output))
	if result == "" {
		return true, "", errors.NewConfigInvalidError(
			"api_key_command",
			"command returned empty output",
		)
	}

	return true, result, nil
}

// resolveEnvVar handles environment variable pattern: ${VAR_NAME}
func (r *Resolver) resolveEnvVar(value string) (bool, string, error) {
	// Pattern: ${VAR_NAME}
	pattern := regexp.MustCompile(`^\$\{([A-Za-z_][A-Za-z0-9_]*)\}$`)
	matches := pattern.FindStringSubmatch(value)

	if len(matches) != 2 {
		return false, "", nil
	}

	varName := matches[1]
	result := os.Getenv(varName)

	if result == "" {
		return true, "", errors.NewConfigMissingError(
			fmt.Sprintf("environment variable %s", varName),
		)
	}

	return true, result, nil
}

// resolveFilePath handles file path pattern: ~/path, /path, ./path
func (r *Resolver) resolveFilePath(value string) (bool, string, error) {
	// Check if value looks like a file path
	if !isFilePath(value) {
		return false, "", nil
	}

	// Expand home directory
	expandedPath := value
	if strings.HasPrefix(value, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return true, "", errors.NewConfigInvalidError(
				"api_key_file",
				fmt.Sprintf("failed to get home directory: %v", err),
			)
		}
		expandedPath = filepath.Join(homeDir, value[2:])
	}

	// Expand environment variables in path
	expandedPath = os.ExpandEnv(expandedPath)

	// Check if file exists
	fileInfo, err := os.Stat(expandedPath)
	if err != nil {
		if os.IsNotExist(err) {
			return true, "", errors.NewConfigInvalidError(
				"api_key_file",
				fmt.Sprintf("file does not exist: %s", expandedPath),
			)
		}
		return true, "", errors.NewConfigInvalidError(
			"api_key_file",
			fmt.Sprintf("failed to access file: %v", err),
		)
	}

	// Check if it's a directory
	if fileInfo.IsDir() {
		return true, "", errors.NewConfigInvalidError(
			"api_key_file",
			fmt.Sprintf("path is a directory, not a file: %s", expandedPath),
		)
	}

	// Check file size (prevent reading huge files)
	maxSize := int64(1024 * 1024) // 1MB max
	if fileInfo.Size() > maxSize {
		return true, "", errors.NewConfigInvalidError(
			"api_key_file",
			fmt.Sprintf("file is too large (max %d bytes): %s", maxSize, expandedPath),
		)
	}

	// Read file content
	content, err := os.ReadFile(expandedPath)
	if err != nil {
		return true, "", errors.NewConfigInvalidError(
			"api_key_file",
			fmt.Sprintf("failed to read file: %v", err),
		)
	}

	result := strings.TrimSpace(string(content))
	if result == "" {
		return true, "", errors.NewConfigInvalidError(
			"api_key_file",
			fmt.Sprintf("file is empty: %s", expandedPath),
		)
	}

	return true, result, nil
}

// isFilePath checks if a value looks like a file path
func isFilePath(value string) bool {
	// Paths starting with ~, /, ./, or ../
	if strings.HasPrefix(value, "~/") ||
		strings.HasPrefix(value, "/") ||
		strings.HasPrefix(value, "./") ||
		strings.HasPrefix(value, "../") {
		return true
	}

	// Paths containing path separators (for relative paths)
	if strings.Contains(value, "/") || strings.Contains(value, "\\") {
		return true
	}

	// Common file extensions for key files
	extensions := []string{".txt", ".key", ".pem", ".secret", ".env"}
	for _, ext := range extensions {
		if strings.HasSuffix(value, ext) {
			return true
		}
	}

	return false
}

// ResolveAll resolves multiple values and returns them as a map
func (r *Resolver) ResolveAll(values map[string]string) (map[string]string, error) {
	result := make(map[string]string, len(values))

	for key, value := range values {
		resolved, err := r.Resolve(value)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve %s: %w", key, err)
		}
		result[key] = resolved
	}

	return result, nil
}
