// Package builtin provides built-in tools for the tool execution framework.
package builtin

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/AINative-studio/ainative-code/internal/tools"
)

// ExecCommandTool implements a tool for executing shell commands with security restrictions.
type ExecCommandTool struct {
	allowedCommands []string
	workingDir      string
}

// NewExecCommandTool creates a new ExecCommandTool instance.
// If allowedCommands is empty, all commands are allowed (use with caution).
// workingDir sets the working directory for command execution.
func NewExecCommandTool(allowedCommands []string, workingDir string) *ExecCommandTool {
	return &ExecCommandTool{
		allowedCommands: allowedCommands,
		workingDir:      workingDir,
	}
}

// Name returns the unique name of the tool.
func (t *ExecCommandTool) Name() string {
	return "exec_command"
}

// Description returns a human-readable description of what the tool does.
func (t *ExecCommandTool) Description() string {
	return "Executes shell commands with security restrictions, timeout handling, and output capture"
}

// Schema returns the JSON schema for the tool's input parameters.
func (t *ExecCommandTool) Schema() tools.ToolSchema {
	maxCommandLength := 8192
	return tools.ToolSchema{
		Type: "object",
		Properties: map[string]tools.PropertyDef{
			"command": {
				Type:        "string",
				Description: "The command to execute (e.g., 'ls', 'git', 'npm')",
				MaxLength:   &maxCommandLength,
			},
			"args": {
				Type:        "array",
				Description: "Command arguments as an array of strings (e.g., ['status', '--short'])",
			},
			"working_dir": {
				Type:        "string",
				Description: "Working directory for command execution (overrides default if provided)",
			},
			"timeout_seconds": {
				Type:        "integer",
				Description: "Command timeout in seconds (default: 30, max: 300)",
			},
			"capture_stderr": {
				Type:        "boolean",
				Description: "Whether to capture stderr separately (default: true)",
				Default:     true,
			},
			"env": {
				Type:        "object",
				Description: "Environment variables to set for the command (key-value pairs)",
			},
		},
		Required: []string{"command"},
	}
}

// Execute runs the tool with the provided input and returns the result.
func (t *ExecCommandTool) Execute(ctx context.Context, input map[string]interface{}) (*tools.Result, error) {
	// Extract and validate command
	commandRaw, ok := input["command"]
	if !ok {
		return nil, &tools.ErrInvalidInput{
			ToolName: t.Name(),
			Field:    "command",
			Reason:   "command is required",
		}
	}

	command, ok := commandRaw.(string)
	if !ok {
		return nil, &tools.ErrInvalidInput{
			ToolName: t.Name(),
			Field:    "command",
			Reason:   fmt.Sprintf("command must be a string, got %T", commandRaw),
		}
	}

	if command == "" {
		return nil, &tools.ErrInvalidInput{
			ToolName: t.Name(),
			Field:    "command",
			Reason:   "command cannot be empty",
		}
	}

	// Validate command against allowed list
	if len(t.allowedCommands) > 0 {
		if err := t.validateCommand(command); err != nil {
			return nil, err
		}
	}

	// Extract args parameter
	var args []string
	if argsRaw, exists := input["args"]; exists {
		argsSlice, ok := argsRaw.([]interface{})
		if !ok {
			return nil, &tools.ErrInvalidInput{
				ToolName: t.Name(),
				Field:    "args",
				Reason:   fmt.Sprintf("args must be an array, got %T", argsRaw),
			}
		}

		// Convert []interface{} to []string
		args = make([]string, len(argsSlice))
		for i, argRaw := range argsSlice {
			argStr, ok := argRaw.(string)
			if !ok {
				return nil, &tools.ErrInvalidInput{
					ToolName: t.Name(),
					Field:    "args",
					Reason:   fmt.Sprintf("argument at index %d must be a string, got %T", i, argRaw),
				}
			}
			args[i] = argStr
		}
	}

	// Extract working_dir parameter
	workingDir := t.workingDir
	if workingDirRaw, exists := input["working_dir"]; exists {
		var ok bool
		workingDir, ok = workingDirRaw.(string)
		if !ok {
			return nil, &tools.ErrInvalidInput{
				ToolName: t.Name(),
				Field:    "working_dir",
				Reason:   fmt.Sprintf("working_dir must be a string, got %T", workingDirRaw),
			}
		}
	}

	// Extract timeout_seconds parameter with default
	timeoutSeconds := 30
	if timeoutRaw, exists := input["timeout_seconds"]; exists {
		switch v := timeoutRaw.(type) {
		case float64:
			timeoutSeconds = int(v)
		case int:
			timeoutSeconds = v
		case int64:
			timeoutSeconds = int(v)
		default:
			return nil, &tools.ErrInvalidInput{
				ToolName: t.Name(),
				Field:    "timeout_seconds",
				Reason:   fmt.Sprintf("timeout_seconds must be an integer, got %T", timeoutRaw),
			}
		}

		// Validate timeout bounds
		if timeoutSeconds <= 0 {
			return nil, &tools.ErrInvalidInput{
				ToolName: t.Name(),
				Field:    "timeout_seconds",
				Reason:   "timeout_seconds must be positive",
			}
		}
		if timeoutSeconds > 300 { // 5 minutes max
			return nil, &tools.ErrInvalidInput{
				ToolName: t.Name(),
				Field:    "timeout_seconds",
				Reason:   "timeout_seconds cannot exceed 300 seconds (5 minutes)",
			}
		}
	}

	// Extract capture_stderr parameter with default
	captureStderr := true
	if captureStderrRaw, exists := input["capture_stderr"]; exists {
		var ok bool
		captureStderr, ok = captureStderrRaw.(bool)
		if !ok {
			return nil, &tools.ErrInvalidInput{
				ToolName: t.Name(),
				Field:    "capture_stderr",
				Reason:   fmt.Sprintf("capture_stderr must be a boolean, got %T", captureStderrRaw),
			}
		}
	}

	// Extract env parameter
	var envVars []string
	if envRaw, exists := input["env"]; exists {
		envMap, ok := envRaw.(map[string]interface{})
		if !ok {
			return nil, &tools.ErrInvalidInput{
				ToolName: t.Name(),
				Field:    "env",
				Reason:   fmt.Sprintf("env must be an object, got %T", envRaw),
			}
		}

		// Convert map to KEY=VALUE format
		for key, valueRaw := range envMap {
			value, ok := valueRaw.(string)
			if !ok {
				return nil, &tools.ErrInvalidInput{
					ToolName: t.Name(),
					Field:    "env",
					Reason:   fmt.Sprintf("environment variable '%s' must have a string value, got %T", key, valueRaw),
				}
			}
			envVars = append(envVars, fmt.Sprintf("%s=%s", key, value))
		}
	}

	// Create command context with timeout
	cmdCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	// Create command
	cmd := exec.CommandContext(cmdCtx, command, args...)

	// Set working directory if provided
	if workingDir != "" {
		cmd.Dir = workingDir
	}

	// Set environment variables if provided
	if len(envVars) > 0 {
		cmd.Env = append(cmd.Environ(), envVars...)
	}

	// Prepare output buffers
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout

	if captureStderr {
		cmd.Stderr = &stderr
	} else {
		// Combine stdout and stderr
		cmd.Stderr = &stdout
	}

	// Record start time
	startTime := time.Now()

	// Execute command
	execErr := cmd.Run()

	// Calculate execution duration
	duration := time.Since(startTime)

	// Get exit code
	exitCode := 0
	if execErr != nil {
		// Check if it was a timeout
		if cmdCtx.Err() == context.DeadlineExceeded {
			return nil, &tools.ErrTimeout{
				ToolName: t.Name(),
				Duration: fmt.Sprintf("%d seconds", timeoutSeconds),
			}
		}

		// Check if it was a cancellation
		if cmdCtx.Err() == context.Canceled {
			return nil, &tools.ErrExecutionFailed{
				ToolName: t.Name(),
				Reason:   "command execution was cancelled",
				Cause:    cmdCtx.Err(),
			}
		}

		// Get exit code from error
		if exitError, ok := execErr.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			// Command failed to start or other error
			return nil, &tools.ErrExecutionFailed{
				ToolName: t.Name(),
				Reason:   fmt.Sprintf("failed to execute command: %s", command),
				Cause:    execErr,
			}
		}
	}

	// Build output string
	var output strings.Builder
	output.WriteString(fmt.Sprintf("Command: %s %s\n", command, strings.Join(args, " ")))
	output.WriteString(fmt.Sprintf("Exit Code: %d\n", exitCode))
	output.WriteString(fmt.Sprintf("Duration: %s\n", duration))
	output.WriteString("\n--- STDOUT ---\n")
	output.WriteString(stdout.String())

	if captureStderr && stderr.Len() > 0 {
		output.WriteString("\n--- STDERR ---\n")
		output.WriteString(stderr.String())
	}

	// Build metadata
	metadata := map[string]interface{}{
		"command":      command,
		"args":         args,
		"exit_code":    exitCode,
		"duration_ms":  duration.Milliseconds(),
		"stdout_bytes": stdout.Len(),
		"stderr_bytes": stderr.Len(),
	}

	if workingDir != "" {
		metadata["working_dir"] = workingDir
	}

	if len(envVars) > 0 {
		metadata["env_vars"] = envVars
	}

	// Determine success based on exit code
	success := exitCode == 0

	result := &tools.Result{
		Success:  success,
		Output:   output.String(),
		Metadata: metadata,
	}

	return result, nil
}

// Category returns the category this tool belongs to.
func (t *ExecCommandTool) Category() tools.Category {
	return tools.CategorySystem
}

// RequiresConfirmation returns true if this tool requires user confirmation before execution.
func (t *ExecCommandTool) RequiresConfirmation() bool {
	return true // Executing commands can modify state and should require confirmation
}

// validateCommand checks if the given command is in the allowed list.
func (t *ExecCommandTool) validateCommand(command string) error {
	// If no allowed commands configured, deny all access
	if len(t.allowedCommands) == 0 {
		return &tools.ErrPermissionDenied{
			ToolName:  t.Name(),
			Operation: "execute",
			Resource:  command,
			Reason:    "no allowed commands configured, command execution denied",
		}
	}

	// Check if command is in allowed list
	for _, allowedCmd := range t.allowedCommands {
		if command == allowedCmd {
			return nil // Command allowed
		}
	}

	// Command not in allowed list
	return &tools.ErrPermissionDenied{
		ToolName:  t.Name(),
		Operation: "execute",
		Resource:  command,
		Reason:    fmt.Sprintf("command '%s' is not in allowed commands: %v", command, t.allowedCommands),
	}
}
