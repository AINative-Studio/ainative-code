package errors

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestToolExecutionError(t *testing.T) {
	t.Run("NewToolNotFoundError", func(t *testing.T) {
		err := NewToolNotFoundError("git")

		if err.Code() != ErrCodeToolNotFound {
			t.Errorf("expected code %s, got %s", ErrCodeToolNotFound, err.Code())
		}

		if err.ToolName != "git" {
			t.Errorf("expected ToolName 'git', got '%s'", err.ToolName)
		}

		if err.Severity() != SeverityHigh {
			t.Errorf("expected severity %s, got %s", SeverityHigh, err.Severity())
		}

		userMsg := err.UserMessage()
		if !strings.Contains(userMsg, "not available") {
			t.Errorf("user message should mention tool not available: %s", userMsg)
		}
	})

	t.Run("NewToolExecutionFailedError", func(t *testing.T) {
		originalErr := errors.New("command failed")
		output := "fatal: not a git repository"
		err := NewToolExecutionFailedError("git", 128, output, originalErr)

		if err.Code() != ErrCodeToolExecutionFailed {
			t.Errorf("expected code %s, got %s", ErrCodeToolExecutionFailed, err.Code())
		}

		if err.ToolName != "git" {
			t.Errorf("expected ToolName 'git', got '%s'", err.ToolName)
		}

		if err.ExitCode != 128 {
			t.Errorf("expected ExitCode 128, got %d", err.ExitCode)
		}

		if err.Output != output {
			t.Errorf("expected Output '%s', got '%s'", output, err.Output)
		}

		if err.Unwrap() != originalErr {
			t.Error("expected error to wrap original error")
		}
	})

	t.Run("NewToolTimeoutError", func(t *testing.T) {
		timeout := 30 * time.Second
		err := NewToolTimeoutError("terraform", timeout)

		if err.Code() != ErrCodeToolTimeout {
			t.Errorf("expected code %s, got %s", ErrCodeToolTimeout, err.Code())
		}

		if err.ToolName != "terraform" {
			t.Errorf("expected ToolName 'terraform', got '%s'", err.ToolName)
		}

		if !err.IsRetryable() {
			t.Error("timeout error should be retryable")
		}

		errMsg := err.Error()
		if !strings.Contains(errMsg, "30s") {
			t.Errorf("error message should contain timeout duration: %s", errMsg)
		}
	})

	t.Run("NewToolInvalidInputError", func(t *testing.T) {
		err := NewToolInvalidInputError("docker", "image", "must not be empty")

		if err.Code() != ErrCodeToolInvalidInput {
			t.Errorf("expected code %s, got %s", ErrCodeToolInvalidInput, err.Code())
		}

		if err.ToolName != "docker" {
			t.Errorf("expected ToolName 'docker', got '%s'", err.ToolName)
		}

		if err.Severity() != SeverityLow {
			t.Errorf("expected severity %s, got %s", SeverityLow, err.Severity())
		}

		errMsg := err.Error()
		if !strings.Contains(errMsg, "image") || !strings.Contains(errMsg, "must not be empty") {
			t.Errorf("error message should contain parameter and reason: %s", errMsg)
		}
	})

	t.Run("NewToolPermissionDeniedError", func(t *testing.T) {
		err := NewToolPermissionDeniedError("kubectl", "/etc/kubernetes/config")

		if err.Code() != ErrCodeToolPermissionDenied {
			t.Errorf("expected code %s, got %s", ErrCodeToolPermissionDenied, err.Code())
		}

		if err.ToolName != "kubectl" {
			t.Errorf("expected ToolName 'kubectl', got '%s'", err.ToolName)
		}

		if err.Severity() != SeverityHigh {
			t.Errorf("expected severity %s, got %s", SeverityHigh, err.Severity())
		}

		errMsg := err.Error()
		if !strings.Contains(errMsg, "Permission denied") {
			t.Errorf("error message should mention permission denied: %s", errMsg)
		}
	})

	t.Run("WithPath", func(t *testing.T) {
		err := NewToolExecutionError(ErrCodeToolExecutionFailed, "test", "tool")
		err.WithPath("/usr/bin/tool")

		if err.ToolPath != "/usr/bin/tool" {
			t.Errorf("expected ToolPath '/usr/bin/tool', got '%s'", err.ToolPath)
		}
	})

	t.Run("WithParameter", func(t *testing.T) {
		err := NewToolExecutionError(ErrCodeToolExecutionFailed, "test", "tool")
		err.WithParameter("timeout", 30)
		err.WithParameter("retry", true)

		if err.Parameters["timeout"] != 30 {
			t.Error("expected parameter 'timeout' to be set")
		}
		if err.Parameters["retry"] != true {
			t.Error("expected parameter 'retry' to be set")
		}
	})

	t.Run("WithExitCode", func(t *testing.T) {
		err := NewToolExecutionError(ErrCodeToolExecutionFailed, "test", "tool")
		err.WithExitCode(1)

		if err.ExitCode != 1 {
			t.Errorf("expected ExitCode 1, got %d", err.ExitCode)
		}
	})

	t.Run("WithOutput", func(t *testing.T) {
		output := "error: file not found"
		err := NewToolExecutionError(ErrCodeToolExecutionFailed, "test", "tool")
		err.WithOutput(output)

		if err.Output != output {
			t.Errorf("expected Output '%s', got '%s'", output, err.Output)
		}
	})

	t.Run("GetOutput truncation", func(t *testing.T) {
		longOutput := strings.Repeat("x", 200)
		err := NewToolExecutionError(ErrCodeToolExecutionFailed, "test", "tool")
		err.WithOutput(longOutput)

		truncated := err.GetOutput(50)
		// 50 chars + "... (truncated)" = 65 chars total
		if len(truncated) != 65 {
			t.Errorf("expected truncated output length 65, got %d", len(truncated))
		}
		if !strings.Contains(truncated, "truncated") {
			t.Error("expected truncated message")
		}
	})

	t.Run("GetOutput no truncation", func(t *testing.T) {
		shortOutput := "short output"
		err := NewToolExecutionError(ErrCodeToolExecutionFailed, "test", "tool")
		err.WithOutput(shortOutput)

		result := err.GetOutput(100)
		if result != shortOutput {
			t.Errorf("expected full output '%s', got '%s'", shortOutput, result)
		}
	})

	t.Run("Method chaining", func(t *testing.T) {
		err := NewToolExecutionError(ErrCodeToolExecutionFailed, "test", "tool").
			WithPath("/usr/bin/tool").
			WithExitCode(2).
			WithOutput("error output").
			WithParameter("verbose", true)

		if err.ToolPath != "/usr/bin/tool" {
			t.Error("expected ToolPath to be set via chaining")
		}
		if err.ExitCode != 2 {
			t.Error("expected ExitCode to be set via chaining")
		}
		if err.Output != "error output" {
			t.Error("expected Output to be set via chaining")
		}
		if err.Parameters["verbose"] != true {
			t.Error("expected parameter to be set via chaining")
		}
	})

	t.Run("Retryability", func(t *testing.T) {
		// Timeout should be retryable
		timeoutErr := NewToolTimeoutError("tool", 30*time.Second)
		if !timeoutErr.IsRetryable() {
			t.Error("timeout error should be retryable")
		}

		// Execution failure should not be retryable by default
		execErr := NewToolExecutionFailedError("tool", 1, "error", nil)
		if execErr.IsRetryable() {
			t.Error("execution failure should not be retryable by default")
		}

		// Not found should not be retryable
		notFoundErr := NewToolNotFoundError("tool")
		if notFoundErr.IsRetryable() {
			t.Error("not found error should not be retryable")
		}
	})
}

func TestToolErrorWrapping(t *testing.T) {
	t.Run("Wrap tool error", func(t *testing.T) {
		toolErr := NewToolNotFoundError("git")
		wrappedErr := Wrap(toolErr, ErrCodeToolExecutionFailed, "failed to execute git command")

		var baseErr *BaseError
		if !As(wrappedErr, &baseErr) {
			t.Fatal("expected BaseError")
		}

		// Check that we can still extract the original tool error
		var originalToolErr *ToolExecutionError
		if !As(wrappedErr, &originalToolErr) {
			t.Fatal("expected to extract ToolExecutionError from chain")
		}

		if originalToolErr.ToolName != "git" {
			t.Errorf("expected ToolName 'git', got '%s'", originalToolErr.ToolName)
		}
	})
}

func BenchmarkNewToolError(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewToolNotFoundError("git")
	}
}

func ExampleNewToolExecutionFailedError() {
	originalErr := errors.New("exit status 1")
	err := NewToolExecutionFailedError("git", 1, "fatal: not a git repository", originalErr)
	println(err.Error())
	println(err.ExitCode)
	println(err.GetOutput(100))
}

func ExampleNewToolTimeoutError() {
	err := NewToolTimeoutError("terraform", 30*time.Second)
	println(err.ToolName)
	println(err.IsRetryable())
}
