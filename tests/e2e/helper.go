package e2e

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestHelper provides utilities for E2E tests
type TestHelper struct {
	t              *testing.T
	binaryPath     string
	workDir        string
	artifactsDir   string
	testName       string
	commandCounter int
	timeout        time.Duration
}

// CommandResult captures the result of a CLI command execution
type CommandResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Duration time.Duration
	Error    error
}

// NewTestHelper creates a new test helper instance
func NewTestHelper(t *testing.T) *TestHelper {
	t.Helper()

	// Build the binary if not already built
	binaryPath := buildBinary(t)

	// Create temporary working directory
	workDir := t.TempDir()

	// Create artifacts directory
	artifactsDir := filepath.Join("artifacts", t.Name())
	err := os.MkdirAll(artifactsDir, 0755)
	require.NoError(t, err, "failed to create artifacts directory")

	return &TestHelper{
		t:            t,
		binaryPath:   binaryPath,
		workDir:      workDir,
		artifactsDir: artifactsDir,
		testName:     t.Name(),
		timeout:      30 * time.Second, // Default timeout
	}
}

// SetTimeout sets the default timeout for commands
func (h *TestHelper) SetTimeout(timeout time.Duration) {
	h.timeout = timeout
}

// RunCommand executes a CLI command and returns the result
func (h *TestHelper) RunCommand(args ...string) *CommandResult {
	h.t.Helper()
	return h.RunCommandWithEnv(nil, args...)
}

// RunCommandWithEnv executes a CLI command with custom environment variables
func (h *TestHelper) RunCommandWithEnv(env map[string]string, args ...string) *CommandResult {
	h.t.Helper()

	h.commandCounter++
	startTime := time.Now()

	// Prepare command
	cmd := exec.Command(h.binaryPath, args...)
	cmd.Dir = h.workDir

	// Set environment variables
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("HOME=%s", h.workDir))
	for key, value := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	// Capture stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute command with timeout
	err := cmd.Start()
	if err != nil {
		return &CommandResult{
			Error:    fmt.Errorf("failed to start command: %w", err),
			Duration: time.Since(startTime),
		}
	}

	// Wait for command with timeout
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	var cmdErr error
	select {
	case <-time.After(h.timeout):
		if err := cmd.Process.Kill(); err != nil {
			h.t.Logf("failed to kill process: %v", err)
		}
		cmdErr = fmt.Errorf("command timed out after %v", h.timeout)
	case err := <-done:
		cmdErr = err
	}

	duration := time.Since(startTime)

	// Determine exit code
	exitCode := 0
	if cmdErr != nil {
		if exitError, ok := cmdErr.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else if cmdErr.Error() != "signal: killed" && !strings.Contains(cmdErr.Error(), "timed out") {
			exitCode = -1
		}
	}

	result := &CommandResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
		Duration: duration,
		Error:    cmdErr,
	}

	// Save artifact
	h.saveArtifact(args, result)

	return result
}

// AssertSuccess asserts that the command succeeded (exit code 0)
func (h *TestHelper) AssertSuccess(result *CommandResult, msgAndArgs ...interface{}) {
	h.t.Helper()
	require.Equal(h.t, 0, result.ExitCode, msgAndArgs...)
}

// AssertFailure asserts that the command failed (non-zero exit code)
func (h *TestHelper) AssertFailure(result *CommandResult, msgAndArgs ...interface{}) {
	h.t.Helper()
	require.NotEqual(h.t, 0, result.ExitCode, msgAndArgs...)
}

// AssertExitCode asserts that the command exited with a specific code
func (h *TestHelper) AssertExitCode(result *CommandResult, expectedCode int, msgAndArgs ...interface{}) {
	h.t.Helper()
	require.Equal(h.t, expectedCode, result.ExitCode, msgAndArgs...)
}

// AssertStdoutContains asserts that stdout contains the expected string
func (h *TestHelper) AssertStdoutContains(result *CommandResult, expected string, msgAndArgs ...interface{}) {
	h.t.Helper()
	require.Contains(h.t, result.Stdout, expected, msgAndArgs...)
}

// AssertStdoutNotContains asserts that stdout does not contain the string
func (h *TestHelper) AssertStdoutNotContains(result *CommandResult, notExpected string, msgAndArgs ...interface{}) {
	h.t.Helper()
	require.NotContains(h.t, result.Stdout, notExpected, msgAndArgs...)
}

// AssertStderrContains asserts that stderr contains the expected string
func (h *TestHelper) AssertStderrContains(result *CommandResult, expected string, msgAndArgs ...interface{}) {
	h.t.Helper()
	require.Contains(h.t, result.Stderr, expected, msgAndArgs...)
}

// AssertStderrNotContains asserts that stderr does not contain the string
func (h *TestHelper) AssertStderrNotContains(result *CommandResult, notExpected string, msgAndArgs ...interface{}) {
	h.t.Helper()
	require.NotContains(h.t, result.Stderr, notExpected, msgAndArgs...)
}

// WriteFile writes content to a file in the work directory
func (h *TestHelper) WriteFile(filename string, content string) {
	h.t.Helper()
	path := filepath.Join(h.workDir, filename)
	err := os.MkdirAll(filepath.Dir(path), 0755)
	require.NoError(h.t, err, "failed to create directory for file")
	err = os.WriteFile(path, []byte(content), 0644)
	require.NoError(h.t, err, "failed to write file")
}

// ReadFile reads content from a file in the work directory
func (h *TestHelper) ReadFile(filename string) string {
	h.t.Helper()
	path := filepath.Join(h.workDir, filename)
	content, err := os.ReadFile(path)
	require.NoError(h.t, err, "failed to read file")
	return string(content)
}

// FileExists checks if a file exists in the work directory
func (h *TestHelper) FileExists(filename string) bool {
	h.t.Helper()
	path := filepath.Join(h.workDir, filename)
	_, err := os.Stat(path)
	return err == nil
}

// GetWorkDir returns the temporary working directory
func (h *TestHelper) GetWorkDir() string {
	return h.workDir
}

// saveArtifact saves command output to an artifact file
func (h *TestHelper) saveArtifact(args []string, result *CommandResult) {
	// Create a sanitized filename from the command
	commandStr := strings.Join(args, "_")
	commandStr = strings.ReplaceAll(commandStr, "/", "_")
	commandStr = strings.ReplaceAll(commandStr, " ", "_")
	commandStr = strings.ReplaceAll(commandStr, "-", "_")
	if len(commandStr) > 100 {
		commandStr = commandStr[:100]
	}

	filename := fmt.Sprintf("command_%s.log", commandStr)
	filepath := filepath.Join(h.artifactsDir, filename)

	content := fmt.Sprintf("Exit Code: %d\nDuration: %v\n\n--- STDOUT ---\n%s\n\n--- STDERR ---\n%s\n",
		result.ExitCode,
		result.Duration,
		result.Stdout,
		result.Stderr,
	)

	if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
		h.t.Logf("Warning: failed to save artifact: %v", err)
	}
}

// buildBinary builds the ainative-code binary for testing
func buildBinary(t *testing.T) string {
	t.Helper()

	// Get project root
	projectRoot, err := filepath.Abs("../..")
	require.NoError(t, err, "failed to get project root")

	// Binary output path
	binaryPath := filepath.Join(projectRoot, "build", "ainative-code-e2e-test")

	// Build the binary
	cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/ainative-code")
	cmd.Dir = projectRoot

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		t.Fatalf("failed to build binary: %v\nStderr: %s", err, stderr.String())
	}

	t.Logf("Built test binary: %s", binaryPath)
	return binaryPath
}

// Cleanup removes test artifacts and temporary files
func (h *TestHelper) Cleanup() {
	// Test cleanup is handled by t.TempDir() automatically
	// Artifacts are preserved for debugging
}
