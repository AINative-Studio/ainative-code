package tools

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrToolNotFound(t *testing.T) {
	tests := []struct {
		name         string
		toolName     string
		expectedMsg  string
	}{
		{
			name:        "basic tool not found",
			toolName:    "missing_tool",
			expectedMsg: "tool not found: missing_tool",
		},
		{
			name:        "empty tool name",
			toolName:    "",
			expectedMsg: "tool not found: ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &ErrToolNotFound{
				ToolName: tt.toolName,
			}

			assert.Equal(t, tt.expectedMsg, err.Error())
			assert.Contains(t, err.Error(), "tool not found")
		})
	}
}

func TestErrInvalidInput(t *testing.T) {
	tests := []struct {
		name         string
		toolName     string
		field        string
		reason       string
		expectedMsg  string
	}{
		{
			name:        "with field and reason",
			toolName:    "test_tool",
			field:       "input_field",
			reason:      "value is invalid",
			expectedMsg: "invalid input for tool test_tool: field input_field - value is invalid",
		},
		{
			name:        "without field",
			toolName:    "test_tool",
			field:       "",
			reason:      "general validation error",
			expectedMsg: "invalid input for tool test_tool: general validation error",
		},
		{
			name:        "empty tool name",
			toolName:    "",
			field:       "field1",
			reason:      "required",
			expectedMsg: "invalid input for tool : field field1 - required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &ErrInvalidInput{
				ToolName: tt.toolName,
				Field:    tt.field,
				Reason:   tt.reason,
			}

			assert.Equal(t, tt.expectedMsg, err.Error())
			assert.Contains(t, err.Error(), "invalid input")
		})
	}
}

func TestErrExecutionFailed(t *testing.T) {
	tests := []struct {
		name         string
		toolName     string
		reason       string
		cause        error
		expectedMsg  string
		shouldUnwrap bool
	}{
		{
			name:         "with cause",
			toolName:     "test_tool",
			reason:       "command failed",
			cause:        errors.New("exit code 1"),
			expectedMsg:  "execution failed for tool test_tool: command failed - caused by: exit code 1",
			shouldUnwrap: true,
		},
		{
			name:         "without cause",
			toolName:     "test_tool",
			reason:       "unknown error",
			cause:        nil,
			expectedMsg:  "execution failed for tool test_tool: unknown error",
			shouldUnwrap: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &ErrExecutionFailed{
				ToolName: tt.toolName,
				Reason:   tt.reason,
				Cause:    tt.cause,
			}

			assert.Equal(t, tt.expectedMsg, err.Error())
			assert.Contains(t, err.Error(), "execution failed")

			if tt.shouldUnwrap {
				unwrapped := err.Unwrap()
				assert.Equal(t, tt.cause, unwrapped)
				assert.True(t, errors.Is(err, tt.cause))
			} else {
				assert.Nil(t, err.Unwrap())
			}
		})
	}
}

func TestErrTimeout(t *testing.T) {
	tests := []struct {
		name         string
		toolName     string
		duration     string
		expectedMsg  string
	}{
		{
			name:        "30 seconds timeout",
			toolName:    "slow_tool",
			duration:    "30 seconds",
			expectedMsg: "tool slow_tool execution timed out after 30 seconds",
		},
		{
			name:        "5 minutes timeout",
			toolName:    "very_slow_tool",
			duration:    "5 minutes",
			expectedMsg: "tool very_slow_tool execution timed out after 5 minutes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &ErrTimeout{
				ToolName: tt.toolName,
				Duration: tt.duration,
			}

			assert.Equal(t, tt.expectedMsg, err.Error())
			assert.Contains(t, err.Error(), "timed out")
		})
	}
}

func TestErrPermissionDenied(t *testing.T) {
	tests := []struct {
		name         string
		toolName     string
		operation    string
		resource     string
		reason       string
		expectedMsg  string
	}{
		{
			name:        "with resource",
			toolName:    "file_tool",
			operation:   "read",
			resource:    "/etc/passwd",
			reason:      "path not in allowed list",
			expectedMsg: "permission denied for tool file_tool: cannot read resource /etc/passwd - path not in allowed list",
		},
		{
			name:        "without resource",
			toolName:    "command_tool",
			operation:   "execute",
			resource:    "",
			reason:      "command not allowed",
			expectedMsg: "permission denied for tool command_tool: cannot execute - command not allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &ErrPermissionDenied{
				ToolName:  tt.toolName,
				Operation: tt.operation,
				Resource:  tt.resource,
				Reason:    tt.reason,
			}

			assert.Equal(t, tt.expectedMsg, err.Error())
			assert.Contains(t, err.Error(), "permission denied")
		})
	}
}

func TestErrToolConflict(t *testing.T) {
	tests := []struct {
		name         string
		toolName     string
		expectedMsg  string
	}{
		{
			name:        "duplicate tool",
			toolName:    "existing_tool",
			expectedMsg: "tool conflict: a tool with name existing_tool is already registered",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &ErrToolConflict{
				ToolName: tt.toolName,
			}

			assert.Equal(t, tt.expectedMsg, err.Error())
			assert.Contains(t, err.Error(), "tool conflict")
		})
	}
}

func TestErrOutputTooLarge(t *testing.T) {
	tests := []struct {
		name         string
		toolName     string
		outputSize   int64
		maxSize      int64
		expectedMsg  string
	}{
		{
			name:        "output exceeds limit",
			toolName:    "large_output_tool",
			outputSize:  1000000,
			maxSize:     100000,
			expectedMsg: "output too large for tool large_output_tool: 1000000 bytes exceeds maximum of 100000 bytes",
		},
		{
			name:        "output way over limit",
			toolName:    "huge_tool",
			outputSize:  10485760, // 10MB
			maxSize:     1048576,  // 1MB
			expectedMsg: "output too large for tool huge_tool: 10485760 bytes exceeds maximum of 1048576 bytes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &ErrOutputTooLarge{
				ToolName:   tt.toolName,
				OutputSize: tt.outputSize,
				MaxSize:    tt.maxSize,
			}

			assert.Equal(t, tt.expectedMsg, err.Error())
			assert.Contains(t, err.Error(), "output too large")
			assert.Contains(t, err.Error(), "exceeds maximum")
		})
	}
}

func TestError_TypeAssertions(t *testing.T) {
	t.Run("ErrToolNotFound type assertion", func(t *testing.T) {
		var err error = &ErrToolNotFound{ToolName: "test"}
		var target *ErrToolNotFound
		assert.True(t, errors.As(err, &target))
		assert.Equal(t, "test", target.ToolName)
	})

	t.Run("ErrInvalidInput type assertion", func(t *testing.T) {
		var err error = &ErrInvalidInput{ToolName: "test", Field: "field1", Reason: "invalid"}
		var target *ErrInvalidInput
		assert.True(t, errors.As(err, &target))
		assert.Equal(t, "test", target.ToolName)
		assert.Equal(t, "field1", target.Field)
	})

	t.Run("ErrExecutionFailed type assertion", func(t *testing.T) {
		var err error = &ErrExecutionFailed{ToolName: "test", Reason: "failed"}
		var target *ErrExecutionFailed
		assert.True(t, errors.As(err, &target))
		assert.Equal(t, "test", target.ToolName)
	})

	t.Run("ErrTimeout type assertion", func(t *testing.T) {
		var err error = &ErrTimeout{ToolName: "test", Duration: "30s"}
		var target *ErrTimeout
		assert.True(t, errors.As(err, &target))
		assert.Equal(t, "test", target.ToolName)
	})

	t.Run("ErrPermissionDenied type assertion", func(t *testing.T) {
		var err error = &ErrPermissionDenied{ToolName: "test", Operation: "read"}
		var target *ErrPermissionDenied
		assert.True(t, errors.As(err, &target))
		assert.Equal(t, "test", target.ToolName)
	})

	t.Run("ErrToolConflict type assertion", func(t *testing.T) {
		var err error = &ErrToolConflict{ToolName: "test"}
		var target *ErrToolConflict
		assert.True(t, errors.As(err, &target))
		assert.Equal(t, "test", target.ToolName)
	})

	t.Run("ErrOutputTooLarge type assertion", func(t *testing.T) {
		var err error = &ErrOutputTooLarge{ToolName: "test", OutputSize: 1000, MaxSize: 100}
		var target *ErrOutputTooLarge
		assert.True(t, errors.As(err, &target))
		assert.Equal(t, "test", target.ToolName)
		assert.Equal(t, int64(1000), target.OutputSize)
	})
}

func TestError_Wrapping(t *testing.T) {
	t.Run("wrap ErrExecutionFailed", func(t *testing.T) {
		cause := errors.New("underlying error")
		execErr := &ErrExecutionFailed{
			ToolName: "test",
			Reason:   "failed",
			Cause:    cause,
		}

		assert.True(t, errors.Is(execErr, cause))
		unwrapped := errors.Unwrap(execErr)
		assert.Equal(t, cause, unwrapped)
	})

	t.Run("ErrExecutionFailed without cause", func(t *testing.T) {
		execErr := &ErrExecutionFailed{
			ToolName: "test",
			Reason:   "failed",
		}

		unwrapped := errors.Unwrap(execErr)
		assert.Nil(t, unwrapped)
	})
}
