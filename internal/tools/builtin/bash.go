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

// BashTool implements a tool for executing shell commands with security sandboxing.
// It provides timeout support, output capture, and command validation.
type BashTool struct {
	sandbox *Sandbox
}

// NewBashTool creates a new BashTool instance with the specified sandbox.
func NewBashTool(sandbox *Sandbox) *BashTool {
	return &BashTool{
		sandbox: sandbox,
	}
}

// Name returns the unique name of the tool.
func (t *BashTool) Name() string {
	return "bash"
}

// Description returns a human-readable description of what the tool does.
func (t *BashTool) Description() string {
	return "Executes shell commands with security sandboxing, timeout support, and separate stdout/stderr capture"
}

// Schema returns the JSON schema for the tool's input parameters.
func (t *BashTool) Schema() tools.ToolSchema {
	maxCommandLength := 65536 // 64KB
	defaultTimeout := 30
	maxTimeout := 300 // 5 minutes

	return tools.ToolSchema{
		Type: "object",
		Properties: map[string]tools.PropertyDef{
			"command": {
				Type:        "string",
				Description: "The shell command to execute",
				MaxLength:   &maxCommandLength,
			},
			"working_dir": {
				Type:        "string",
				Description: "Working directory for command execution (must be within allowed paths)",
			},
			"timeout": {
				Type:        "integer",
				Description: fmt.Sprintf("Execution timeout in seconds (default: %d, max: %d)", defaultTimeout, maxTimeout),
				Default:     defaultTimeout,
			},
			"capture_stderr": {
				Type:        "boolean",
				Description: "Whether to capture stderr separately (default: true)",
				Default:     true,
			},
		},
		Required: []string{"command"},
	}
}

// Execute runs the tool with the provided input and returns the result.
func (t *BashTool) Execute(ctx context.Context, input map[string]interface{}) (*tools.Result, error) {
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

	command = strings.TrimSpace(command)
	if command == "" {
		return nil, &tools.ErrInvalidInput{
			ToolName: t.Name(),
			Field:    "command",
			Reason:   "command cannot be empty",
		}
	}

	// Validate command against sandbox
	if err := t.sandbox.ValidateCommand(command); err != nil {
		// Add tool name if not already present
		if permErr, ok := err.(*tools.ErrPermissionDenied); ok && permErr.ToolName == "" {
			permErr.ToolName = t.Name()
		}
		return nil, err
	}

	// Extract working directory with default
	workingDir := t.sandbox.WorkingDirectory
	if workingDirRaw, exists := input["working_dir"]; exists {
		wd, ok := workingDirRaw.(string)
		if !ok {
			return nil, &tools.ErrInvalidInput{
				ToolName: t.Name(),
				Field:    "working_dir",
				Reason:   fmt.Sprintf("working_dir must be a string, got %T", workingDirRaw),
			}
		}

		// Resolve and validate working directory
		resolvedWD, err := t.sandbox.ResolveWorkingDirectory(wd)
		if err != nil {
			return nil, &tools.ErrInvalidInput{
				ToolName: t.Name(),
				Field:    "working_dir",
				Reason:   fmt.Sprintf("invalid working directory: %v", err),
			}
		}
		workingDir = resolvedWD
	} else if workingDir != "" {
		// Validate default working directory
		resolvedWD, err := t.sandbox.ResolveWorkingDirectory(workingDir)
		if err != nil {
			return nil, &tools.ErrExecutionFailed{
				ToolName: t.Name(),
				Reason:   fmt.Sprintf("default working directory is invalid: %v", err),
			}
		}
		workingDir = resolvedWD
	}

	// Extract timeout with default and validation
	timeout := 30 * time.Second
	if timeoutRaw, exists := input["timeout"]; exists {
		var timeoutSec int64
		switch v := timeoutRaw.(type) {
		case float64:
			timeoutSec = int64(v)
		case int:
			timeoutSec = int64(v)
		case int64:
			timeoutSec = v
		default:
			return nil, &tools.ErrInvalidInput{
				ToolName: t.Name(),
				Field:    "timeout",
				Reason:   fmt.Sprintf("timeout must be an integer, got %T", timeoutRaw),
			}
		}

		if timeoutSec <= 0 {
			return nil, &tools.ErrInvalidInput{
				ToolName: t.Name(),
				Field:    "timeout",
				Reason:   "timeout must be positive",
			}
		}

		if timeoutSec > 300 {
			return nil, &tools.ErrInvalidInput{
				ToolName: t.Name(),
				Field:    "timeout",
				Reason:   "timeout cannot exceed 300 seconds",
			}
		}

		timeout = time.Duration(timeoutSec) * time.Second
	}

	// Extract capture_stderr option
	captureStderr := true
	if captureStderrRaw, exists := input["capture_stderr"]; exists {
		cs, ok := captureStderrRaw.(bool)
		if !ok {
			return nil, &tools.ErrInvalidInput{
				ToolName: t.Name(),
				Field:    "capture_stderr",
				Reason:   fmt.Sprintf("capture_stderr must be a boolean, got %T", captureStderrRaw),
			}
		}
		captureStderr = cs
	}

	// Create command execution context with timeout
	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Execute the command
	result, err := t.executeCommand(execCtx, command, workingDir, captureStderr)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Category returns the category this tool belongs to.
func (t *BashTool) Category() tools.Category {
	return tools.CategorySystem
}

// RequiresConfirmation returns true if this tool requires user confirmation before execution.
func (t *BashTool) RequiresConfirmation() bool {
	return true // Executing arbitrary commands requires confirmation for safety
}

// executeCommand executes a shell command and captures its output.
func (t *BashTool) executeCommand(ctx context.Context, command, workingDir string, captureStderr bool) (*tools.Result, error) {
	// Create the command
	cmd := exec.CommandContext(ctx, "sh", "-c", command)

	// Set working directory if specified
	if workingDir != "" {
		cmd.Dir = workingDir
	}

	// Create buffers for output capture
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	if captureStderr {
		cmd.Stderr = &stderr
	} else {
		// If not capturing separately, combine stderr with stdout
		cmd.Stderr = &stdout
	}

	// Record start time
	startTime := time.Now()

	// Execute the command
	err := cmd.Run()
	executionDuration := time.Since(startTime)

	// Get output
	stdoutStr := stdout.String()
	stderrStr := stderr.String()

	// Check output size limits
	outputSize := int64(len(stdoutStr) + len(stderrStr))
	if err := t.sandbox.ValidateOutputSize(outputSize); err != nil {
		if outputErr, ok := err.(*tools.ErrOutputTooLarge); ok {
			outputErr.ToolName = t.Name()
		}
		return nil, err
	}

	// Build result metadata
	metadata := map[string]interface{}{
		"command":           command,
		"working_dir":       workingDir,
		"execution_time_ms": executionDuration.Milliseconds(),
		"output_size":       outputSize,
	}

	// Handle execution error
	if err != nil {
		// Check if it was a timeout
		if ctx.Err() == context.DeadlineExceeded {
			return nil, &tools.ErrTimeout{
				ToolName: t.Name(),
				Duration: executionDuration.String(),
			}
		}

		// Get exit code if available
		exitCode := -1
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		}

		metadata["exit_code"] = exitCode
		metadata["error"] = err.Error()

		// Build error output
		output := fmt.Sprintf("Command failed with exit code %d\n", exitCode)
		if stdoutStr != "" {
			output += fmt.Sprintf("\n--- STDOUT ---\n%s\n", stdoutStr)
		}
		if stderrStr != "" {
			output += fmt.Sprintf("\n--- STDERR ---\n%s\n", stderrStr)
		}

		return &tools.Result{
			Success:  false,
			Output:   output,
			Error:    err,
			Metadata: metadata,
		}, nil
	}

	// Success case
	metadata["exit_code"] = 0

	// Build success output
	var output string
	if captureStderr && stderrStr != "" {
		output = fmt.Sprintf("--- STDOUT ---\n%s\n\n--- STDERR ---\n%s", stdoutStr, stderrStr)
		metadata["has_stderr"] = true
	} else {
		output = stdoutStr
		metadata["has_stderr"] = false
	}

	return &tools.Result{
		Success:  true,
		Output:   output,
		Metadata: metadata,
	}, nil
}
