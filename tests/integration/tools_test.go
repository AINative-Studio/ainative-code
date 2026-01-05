// +build integration

package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/AINative-studio/ainative-code/internal/tools"
	"github.com/AINative-studio/ainative-code/internal/tools/builtin"
	"github.com/stretchr/testify/suite"
)

// ToolsIntegrationTestSuite tests tool execution functionality.
type ToolsIntegrationTestSuite struct {
	suite.Suite
	tmpDir  string
	cleanup func()
}

// SetupTest runs before each test in the suite.
func (s *ToolsIntegrationTestSuite) SetupTest() {
	// Create temporary directory for test files
	tmpDir := s.T().TempDir()
	s.tmpDir = tmpDir

	s.cleanup = func() {
		// Cleanup is handled automatically by testing.T.TempDir()
	}
}

// TearDownTest runs after each test in the suite.
func (s *ToolsIntegrationTestSuite) TearDownTest() {
	if s.cleanup != nil {
		s.cleanup()
	}
}

// TestBashCommandExecution tests executing bash commands.
func (s *ToolsIntegrationTestSuite) TestBashCommandExecution() {
	// Given: An ExecCommandTool with allowed commands
	allowedCommands := []string{"echo", "ls", "pwd"}
	tool := builtin.NewExecCommandTool(allowedCommands, s.tmpDir)

	ctx := context.Background()

	// When: Executing an echo command
	input := map[string]interface{}{
		"command": "echo",
		"args":    []interface{}{"Hello", "World"},
	}

	result, err := tool.Execute(ctx, input)

	// Then: Command should execute successfully
	s.Require().NoError(err, "Command execution should succeed")
	s.NotNil(result, "Result should not be nil")
	s.True(result.Success, "Command should succeed")
	s.Contains(result.Output, "Hello World", "Output should contain echoed text")
	s.Contains(result.Output, "Exit Code: 0", "Exit code should be 0")
}

// TestCommandWithWorkingDirectory tests executing commands in specific working directory.
func (s *ToolsIntegrationTestSuite) TestCommandWithWorkingDirectory() {
	// Given: A tool and a test directory
	allowedCommands := []string{"pwd"}
	tool := builtin.NewExecCommandTool(allowedCommands, s.tmpDir)

	ctx := context.Background()

	// When: Executing pwd command
	input := map[string]interface{}{
		"command":     "pwd",
		"working_dir": s.tmpDir,
	}

	result, err := tool.Execute(ctx, input)

	// Then: Should execute in specified directory
	s.Require().NoError(err, "Command execution should succeed")
	s.True(result.Success, "Command should succeed")
	s.Contains(result.Output, s.tmpDir, "Output should contain working directory path")
}

// TestCommandWithTimeout tests command timeout handling.
func (s *ToolsIntegrationTestSuite) TestCommandWithTimeout() {
	// Given: A tool configured with short timeout
	allowedCommands := []string{"sleep"}
	tool := builtin.NewExecCommandTool(allowedCommands, s.tmpDir)

	ctx := context.Background()

	// When: Executing a command that exceeds timeout
	input := map[string]interface{}{
		"command":         "sleep",
		"args":            []interface{}{"10"}, // 10 seconds
		"timeout_seconds": 1,                    // 1 second timeout
	}

	result, err := tool.Execute(ctx, input)

	// Then: Should timeout
	s.Error(err, "Command should timeout")
	s.Nil(result, "Result should be nil on timeout")
	s.IsType(&tools.ErrTimeout{}, err, "Error should be ErrTimeout type")
}

// TestCommandPermissionDenied tests blocking unauthorized commands.
func (s *ToolsIntegrationTestSuite) TestCommandPermissionDenied() {
	// Given: A tool with restricted commands
	allowedCommands := []string{"echo", "ls"}
	tool := builtin.NewExecCommandTool(allowedCommands, s.tmpDir)

	ctx := context.Background()

	// When: Attempting to execute unauthorized command
	input := map[string]interface{}{
		"command": "rm", // Not in allowed list
		"args":    []interface{}{"-rf", "/"},
	}

	result, err := tool.Execute(ctx, input)

	// Then: Should be denied
	s.Error(err, "Unauthorized command should be denied")
	s.Nil(result, "Result should be nil for denied command")
	s.IsType(&tools.ErrPermissionDenied{}, err, "Error should be ErrPermissionDenied type")
}

// TestCommandWithEnvironmentVariables tests setting environment variables.
func (s *ToolsIntegrationTestSuite) TestCommandWithEnvironmentVariables() {
	// Given: A tool and environment variables
	allowedCommands := []string{"printenv", "sh"}
	tool := builtin.NewExecCommandTool(allowedCommands, s.tmpDir)

	ctx := context.Background()

	// When: Executing command with custom environment
	input := map[string]interface{}{
		"command": "sh",
		"args":    []interface{}{"-c", "echo $TEST_VAR"},
		"env": map[string]interface{}{
			"TEST_VAR": "test_value_123",
		},
	}

	result, err := tool.Execute(ctx, input)

	// Then: Environment variable should be set
	s.Require().NoError(err, "Command execution should succeed")
	s.True(result.Success, "Command should succeed")
	s.Contains(result.Output, "test_value_123", "Output should contain environment variable value")
}

// TestFileReadOperation tests reading files.
func (s *ToolsIntegrationTestSuite) TestFileReadOperation() {
	// Given: A file with test content
	testFilePath := filepath.Join(s.tmpDir, "test_read.txt")
	testContent := "This is test content for reading.\nLine 2\nLine 3"
	err := os.WriteFile(testFilePath, []byte(testContent), 0644)
	s.Require().NoError(err, "Failed to create test file")

	readTool := builtin.NewReadFileTool()
	ctx := context.Background()

	// When: Reading the file
	input := map[string]interface{}{
		"path": testFilePath,
	}

	result, err := readTool.Execute(ctx, input)

	// Then: Should read file successfully
	s.Require().NoError(err, "File read should succeed")
	s.NotNil(result, "Result should not be nil")
	s.True(result.Success, "Operation should succeed")
	s.Contains(result.Output, testContent, "Output should contain file content")
}

// TestFileWriteOperation tests writing files.
func (s *ToolsIntegrationTestSuite) TestFileWriteOperation() {
	// Given: A write file tool
	writeTool := builtin.NewWriteFileTool()
	ctx := context.Background()

	testFilePath := filepath.Join(s.tmpDir, "test_write.txt")
	testContent := "This is newly written content."

	// When: Writing to a file
	input := map[string]interface{}{
		"path":    testFilePath,
		"content": testContent,
	}

	result, err := writeTool.Execute(ctx, input)

	// Then: Should write successfully
	s.Require().NoError(err, "File write should succeed")
	s.NotNil(result, "Result should not be nil")
	s.True(result.Success, "Operation should succeed")

	// Verify file was written
	writtenContent, err := os.ReadFile(testFilePath)
	s.Require().NoError(err, "Should be able to read written file")
	s.Equal(testContent, string(writtenContent), "Written content should match")
}

// TestFileOverwriteOperation tests overwriting existing files.
func (s *ToolsIntegrationTestSuite) TestFileOverwriteOperation() {
	// Given: An existing file
	testFilePath := filepath.Join(s.tmpDir, "test_overwrite.txt")
	originalContent := "Original content"
	err := os.WriteFile(testFilePath, []byte(originalContent), 0644)
	s.Require().NoError(err)

	writeTool := builtin.NewWriteFileTool()
	ctx := context.Background()

	// When: Overwriting the file
	newContent := "New overwritten content"
	input := map[string]interface{}{
		"path":    testFilePath,
		"content": newContent,
	}

	result, err := writeTool.Execute(ctx, input)

	// Then: Should overwrite successfully
	s.Require().NoError(err, "File overwrite should succeed")
	s.True(result.Success, "Operation should succeed")

	// Verify content was overwritten
	writtenContent, err := os.ReadFile(testFilePath)
	s.Require().NoError(err)
	s.Equal(newContent, string(writtenContent), "Content should be overwritten")
	s.NotContains(string(writtenContent), originalContent, "Original content should be replaced")
}

// TestToolResultCapture tests capturing tool execution results.
func (s *ToolsIntegrationTestSuite) TestToolResultCapture() {
	// Given: A command tool
	allowedCommands := []string{"echo"}
	tool := builtin.NewExecCommandTool(allowedCommands, s.tmpDir)

	ctx := context.Background()

	// When: Executing a command and capturing results
	input := map[string]interface{}{
		"command": "echo",
		"args":    []interface{}{"Result capture test"},
	}

	result, err := tool.Execute(ctx, input)

	// Then: Result should contain complete metadata
	s.Require().NoError(err, "Execution should succeed")
	s.NotNil(result, "Result should not be nil")

	// Verify result structure
	s.True(result.Success, "Success flag should be set")
	s.NotEmpty(result.Output, "Output should not be empty")
	s.NotNil(result.Metadata, "Metadata should not be nil")

	// Verify metadata contains expected fields
	metadata := result.Metadata
	s.Contains(metadata, "command", "Metadata should contain command")
	s.Contains(metadata, "exit_code", "Metadata should contain exit code")
	s.Contains(metadata, "duration_ms", "Metadata should contain duration")
	s.Equal(0, metadata["exit_code"], "Exit code should be 0")
}

// TestToolErrorHandling tests tool error handling.
func (s *ToolsIntegrationTestSuite) TestToolErrorHandling() {
	// Given: A read file tool
	readTool := builtin.NewReadFileTool()
	ctx := context.Background()

	// When: Attempting to read non-existent file
	input := map[string]interface{}{
		"path": filepath.Join(s.tmpDir, "non_existent_file.txt"),
	}

	result, err := readTool.Execute(ctx, input)

	// Then: Should return error
	s.Error(err, "Should error on non-existent file")
	s.Nil(result, "Result should be nil on error")
}

// TestCommandStderrCapture tests capturing stderr output separately.
func (s *ToolsIntegrationTestSuite) TestCommandStderrCapture() {
	// Given: A command that outputs to stderr
	allowedCommands := []string{"sh"}
	tool := builtin.NewExecCommandTool(allowedCommands, s.tmpDir)

	ctx := context.Background()

	// When: Executing command that writes to stderr
	input := map[string]interface{}{
		"command":        "sh",
		"args":           []interface{}{"-c", "echo 'stdout message'; echo 'stderr message' >&2"},
		"capture_stderr": true,
	}

	result, err := tool.Execute(ctx, input)

	// Then: Should capture both stdout and stderr
	s.Require().NoError(err, "Command should succeed")
	s.True(result.Success, "Operation should succeed")
	s.Contains(result.Output, "stdout message", "Should contain stdout")
	s.Contains(result.Output, "stderr message", "Should contain stderr")
	s.Contains(result.Output, "--- STDOUT ---", "Should label stdout section")
	s.Contains(result.Output, "--- STDERR ---", "Should label stderr section")
}

// TestCommandExitCodeHandling tests handling non-zero exit codes.
func (s *ToolsIntegrationTestSuite) TestCommandExitCodeHandling() {
	// Given: A command that exits with non-zero code
	allowedCommands := []string{"sh"}
	tool := builtin.NewExecCommandTool(allowedCommands, s.tmpDir)

	ctx := context.Background()

	// When: Executing command that fails
	input := map[string]interface{}{
		"command": "sh",
		"args":    []interface{}{"-c", "exit 42"},
	}

	result, err := tool.Execute(ctx, input)

	// Then: Should capture exit code
	s.Require().NoError(err, "Execution itself should succeed (command ran)")
	s.NotNil(result, "Result should not be nil")
	s.False(result.Success, "Success should be false for non-zero exit")
	s.Contains(result.Output, "Exit Code: 42", "Should show exit code 42")
	s.Equal(42, result.Metadata["exit_code"], "Metadata should contain exit code 42")
}

// TestToolRegistry tests registering and retrieving tools.
func (s *ToolsIntegrationTestSuite) TestToolRegistry() {
	// Given: A tool registry
	registry := tools.NewRegistry()

	// When: Registering tools
	execTool := builtin.NewExecCommandTool([]string{"echo"}, s.tmpDir)
	readTool := builtin.NewReadFileTool()
	writeTool := builtin.NewWriteFileTool()

	err := registry.Register(execTool)
	s.Require().NoError(err, "Should register exec tool")

	err = registry.Register(readTool)
	s.Require().NoError(err, "Should register read tool")

	err = registry.Register(writeTool)
	s.Require().NoError(err, "Should register write tool")

	// Then: Should be able to retrieve tools
	retrievedExec, err := registry.Get("exec_command")
	s.Require().NoError(err, "Should retrieve exec tool")
	s.Equal(execTool.Name(), retrievedExec.Name())

	retrievedRead, err := registry.Get("read_file")
	s.Require().NoError(err, "Should retrieve read tool")
	s.Equal(readTool.Name(), retrievedRead.Name())

	// When: Listing all tools
	allTools := registry.List()

	// Then: Should return all registered tools
	s.Len(allTools, 3, "Should have 3 registered tools")
}

// TestSandboxingVerification tests that tools respect sandboxing constraints.
func (s *ToolsIntegrationTestSuite) TestSandboxingVerification() {
	// Given: A tool with working directory constraint
	allowedCommands := []string{"ls"}
	tool := builtin.NewExecCommandTool(allowedCommands, s.tmpDir)

	ctx := context.Background()

	// When: Executing command in sandboxed directory
	input := map[string]interface{}{
		"command":     "ls",
		"working_dir": s.tmpDir,
	}

	result, err := tool.Execute(ctx, input)

	// Then: Should execute successfully in sandbox
	s.Require().NoError(err, "Command should succeed in sandbox")
	s.True(result.Success, "Operation should succeed")

	// Verify working directory was used
	s.Contains(result.Metadata, "working_dir", "Metadata should track working directory")
	s.Equal(s.tmpDir, result.Metadata["working_dir"], "Working directory should match sandbox")
}

// TestConcurrentToolExecution tests executing tools concurrently.
func (s *ToolsIntegrationTestSuite) TestConcurrentToolExecution() {
	// Given: Multiple tools
	allowedCommands := []string{"echo"}
	tool := builtin.NewExecCommandTool(allowedCommands, s.tmpDir)

	ctx := context.Background()
	concurrentOps := 10
	done := make(chan bool, concurrentOps)
	errors := make(chan error, concurrentOps)

	// When: Executing tools concurrently
	for i := 0; i < concurrentOps; i++ {
		go func(index int) {
			input := map[string]interface{}{
				"command": "echo",
				"args":    []interface{}{"Concurrent execution"},
			}

			_, err := tool.Execute(ctx, input)
			if err != nil {
				errors <- err
			}
			done <- true
		}(i)
	}

	// Wait for all operations to complete
	for i := 0; i < concurrentOps; i++ {
		<-done
	}
	close(errors)

	// Then: All operations should succeed
	s.Empty(errors, "No errors should occur during concurrent execution")
}

// TestInvalidToolInput tests handling invalid tool inputs.
func (s *ToolsIntegrationTestSuite) TestInvalidToolInput() {
	// Given: An exec command tool
	allowedCommands := []string{"echo"}
	tool := builtin.NewExecCommandTool(allowedCommands, s.tmpDir)

	ctx := context.Background()

	// When: Providing invalid input (missing command)
	input := map[string]interface{}{
		"args": []interface{}{"test"},
		// Missing "command" field
	}

	result, err := tool.Execute(ctx, input)

	// Then: Should return invalid input error
	s.Error(err, "Should error on invalid input")
	s.Nil(result, "Result should be nil")
	s.IsType(&tools.ErrInvalidInput{}, err, "Error should be ErrInvalidInput type")

	// When: Providing invalid command type
	input2 := map[string]interface{}{
		"command": 123, // Should be string
	}

	result, err = tool.Execute(ctx, input2)

	// Then: Should return invalid input error
	s.Error(err, "Should error on invalid command type")
	s.Nil(result, "Result should be nil")
	s.IsType(&tools.ErrInvalidInput{}, err, "Error should be ErrInvalidInput type")
}

// TestToolSchemaValidation tests tool schema definitions.
func (s *ToolsIntegrationTestSuite) TestToolSchemaValidation() {
	// Given: Various tools
	execTool := builtin.NewExecCommandTool([]string{"echo"}, s.tmpDir)
	readTool := builtin.NewReadFileTool()
	writeTool := builtin.NewWriteFileTool()

	// When: Getting tool schemas
	execSchema := execTool.Schema()
	readSchema := readTool.Schema()
	writeSchema := writeTool.Schema()

	// Then: Schemas should be properly defined
	s.Equal("object", execSchema.Type, "Exec schema should be object type")
	s.Contains(execSchema.Required, "command", "Command should be required")
	s.NotNil(execSchema.Properties, "Schema should have properties")

	s.Equal("object", readSchema.Type, "Read schema should be object type")
	s.NotNil(readSchema.Properties, "Read schema should have properties")

	s.Equal("object", writeSchema.Type, "Write schema should be object type")
	s.NotNil(writeSchema.Properties, "Write schema should have properties")
}

// TestToolsIntegrationTestSuite runs the test suite.
func TestToolsIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(ToolsIntegrationTestSuite))
}
