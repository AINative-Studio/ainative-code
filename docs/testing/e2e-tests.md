# End-to-End (E2E) Testing Guide

## Overview

This document describes the E2E testing framework for AINative-Code CLI application. E2E tests simulate real user interactions by spawning actual CLI processes and verifying their behavior end-to-end.

## Test Architecture

### Framework Components

1. **TestHelper** (`tests/e2e/helper.go`)
   - Core utility for running CLI commands
   - Manages temporary test environments
   - Handles artifact collection
   - Provides assertion helpers

2. **Test Suites**
   - `onboarding_test.go` - First-time user onboarding workflows
   - `chat_test.go` - Chat command functionality
   - `session_test.go` - Session management and export
   - `provider_test.go` - Multi-provider switching
   - `error_recovery_test.go` - Error handling and recovery
   - `config_test.go` - Configuration management

### Key Features

- **Isolated Test Environments**: Each test runs in its own temporary directory
- **Artifact Collection**: Command outputs are saved for debugging
- **Timeout Management**: Configurable timeouts prevent hanging tests
- **Cross-Platform**: Tests run on Linux, macOS, and Windows
- **CI Integration**: Automated testing via GitHub Actions

## Running Tests

### Local Development

#### Run all E2E tests
```bash
make test-e2e
```

#### Run E2E tests in short mode (skips long-running tests)
```bash
make test-e2e-short
```

#### Run E2E tests with verbose output
```bash
make test-e2e-verbose
```

#### Run specific test
```bash
cd tests/e2e
go test -v -run TestFirstTimeUserOnboarding
```

#### Run all tests (unit + integration + E2E)
```bash
make test-all
```

### Clean Test Artifacts

```bash
make test-e2e-clean
```

## Test Scenarios Covered

### 1. First-Time User Onboarding
- **File**: `onboarding_test.go`
- **Tests**:
  - Help and version commands
  - Configuration initialization
  - Configuration validation
  - First chat interaction
  - Complete onboarding workflow

### 2. Chat Sessions
- **File**: `chat_test.go`
- **Tests**:
  - Single message mode
  - Interactive mode
  - Session management
  - Verbose mode
  - Custom system messages
  - Command aliases
  - Error handling
  - Streaming responses
  - Provider-specific behavior

### 3. Session Export
- **File**: `session_test.go`
- **Tests**:
  - Session listing
  - Session details viewing
  - Export to JSON, Markdown, HTML
  - Export to stdout
  - Batch export operations
  - Session deletion
  - Database operations
  - Complete session lifecycle

### 4. Multi-Provider Switching
- **File**: `provider_test.go`
- **Tests**:
  - Switch between OpenAI, Anthropic, Ollama
  - Provider flag overrides
  - Environment variable configuration
  - Provider validation
  - Provider persistence
  - Custom endpoints
  - Provider performance

### 5. Error Recovery
- **File**: `error_recovery_test.go`
- **Tests**:
  - Missing configuration recovery
  - Invalid configuration recovery
  - Corrupted config file handling
  - Network error handling
  - Invalid API key handling
  - Rate limit handling
  - Input validation
  - Graceful error messages
  - Verbose error output
  - Interruption recovery
  - Corrupted data recovery
  - Resource exhaustion
  - Concurrent operation errors

### 6. Configuration Management
- **File**: `config_test.go`
- **Tests**:
  - Environment variable handling
  - Config file formats (YAML)
  - Nested configuration values
  - Help command
  - Version command

## Writing New E2E Tests

### Basic Test Structure

```go
func TestMyFeature(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping E2E test in short mode")
    }

    h := NewTestHelper(t)
    defer h.Cleanup()

    t.Run("specific scenario", func(t *testing.T) {
        result := h.RunCommand("command", "arg1", "arg2")
        h.AssertSuccess(result, "command should succeed")
        h.AssertStdoutContains(result, "expected output", "should show expected text")
    })
}
```

### TestHelper API

#### Running Commands
```go
// Basic command execution
result := h.RunCommand("chat", "Hello")

// With environment variables
env := map[string]string{"OPENAI_API_KEY": "test-key"}
result := h.RunCommandWithEnv(env, "chat", "Hello")

// Set custom timeout
h.SetTimeout(60 * time.Second)
```

#### Assertions
```go
// Success/Failure
h.AssertSuccess(result, "should succeed")
h.AssertFailure(result, "should fail")
h.AssertExitCode(result, 1, "should exit with code 1")

// Output checking
h.AssertStdoutContains(result, "text", "should contain text")
h.AssertStdoutNotContains(result, "text", "should not contain text")
h.AssertStderrContains(result, "error", "should show error")
```

#### File Operations
```go
// Write files
h.WriteFile(".ainative-code.yaml", yamlContent)

// Read files
content := h.ReadFile("output.json")

// Check existence
exists := h.FileExists("config.yaml")

// Get work directory
workDir := h.GetWorkDir()
```

### CommandResult Structure

```go
type CommandResult struct {
    Stdout   string        // Command stdout
    Stderr   string        // Command stderr
    ExitCode int           // Exit code (0 = success)
    Duration time.Duration // Execution time
    Error    error         // Error if command failed
}
```

## Test Artifacts

### Artifact Collection

Test artifacts are automatically saved to `tests/e2e/artifacts/<TestName>/`:
- Command stdout and stderr
- Exit codes and duration
- Named by command arguments

### Artifact Format

```
Exit Code: 0
Duration: 15.234ms

--- STDOUT ---
<command output>

--- STDERR ---
<error output>
```

### Viewing Artifacts

After test failures, artifacts help debug issues:

```bash
ls tests/e2e/artifacts/TestFirstTimeUserOnboarding/
cat tests/e2e/artifacts/TestFirstTimeUserOnboarding/command_config_init.log
```

## CI/CD Integration

### GitHub Actions Workflow

E2E tests run automatically on:
- Push to main, develop, or phase-* branches
- Pull requests to main or develop
- Manual workflow dispatch

**File**: `.github/workflows/e2e-tests.yml`

### CI Configuration

- **Platforms**: Ubuntu, macOS, Windows
- **Timeout**: 15 minutes
- **Go Version**: 1.25.5
- **Artifact Retention**: 7 days
- **On Failure**: Uploads artifacts for debugging

### Running in CI

Tests automatically run with:
- 10-minute timeout per test suite
- Artifact collection on failure
- Cross-platform validation

## Performance Requirements

### Runtime Constraints

- **Total E2E Suite**: < 10 minutes
- **Individual Test**: < 30 seconds (default)
- **Long Tests**: < 60 seconds (explicitly marked)

### Timeout Configuration

```go
// Default timeout (30 seconds)
h := NewTestHelper(t)

// Custom timeout
h.SetTimeout(60 * time.Second)
```

## Best Practices

### 1. Test Isolation
- Each test gets its own temporary directory
- No shared state between tests
- Clean up is automatic via `defer h.Cleanup()`

### 2. Fast Tests
- Use `-short` flag to skip slow tests during development
- Keep individual tests under 30 seconds
- Use `h.SetTimeout()` for longer operations

### 3. Meaningful Assertions
- Include descriptive messages
- Test both success and failure cases
- Verify actual behavior, not just exit codes

### 4. Artifact Usage
- Artifacts are saved automatically
- Use them for debugging failures
- They're preserved in CI for 7 days

### 5. Environment Variables
- Test with different env var configurations
- Verify precedence (flag > env > config)
- Clean up env vars between tests

### 6. Error Messages
- Test that error messages are helpful
- Verify they guide users to solutions
- Check both verbose and non-verbose modes

## Debugging Failed Tests

### Local Debugging

1. **Run specific test**:
   ```bash
   cd tests/e2e
   go test -v -run TestSpecificName
   ```

2. **Check artifacts**:
   ```bash
   cat tests/e2e/artifacts/TestName/command_*.log
   ```

3. **Run with verbose**:
   ```bash
   make test-e2e-verbose
   ```

### CI Debugging

1. **Download artifacts** from GitHub Actions
2. **Check test output** in the workflow logs
3. **Review uploaded artifacts** for failed commands
4. **Replicate locally** using the same Go version

## Common Issues

### Binary Not Found
**Problem**: Test can't find the binary to execute

**Solution**: Ensure you run `make build` or `make test-e2e` (which builds automatically)

### Timeout Errors
**Problem**: Tests timeout during execution

**Solution**:
- Increase timeout: `h.SetTimeout(60 * time.Second)`
- Check for hanging commands
- Verify network connectivity for API calls

### File Permission Errors
**Problem**: Can't write to test directories

**Solution**: Ensure proper permissions in temporary directories

### Flaky Tests
**Problem**: Tests pass sometimes, fail others

**Solution**:
- Check for race conditions
- Verify test isolation
- Look for time-dependent assertions
- Check for external dependencies

## Test Coverage

Current E2E test coverage includes:

- ✅ First-time user onboarding (12 scenarios)
- ✅ Chat command functionality (30+ scenarios)
- ✅ Session management (25+ scenarios)
- ✅ Provider switching (15+ scenarios)
- ✅ Error recovery (20+ scenarios)
- ✅ Configuration management (15+ scenarios)

**Total**: 100+ E2E test scenarios

## Future Enhancements

Potential additions to the E2E test suite:

1. **Performance Testing**
   - Response time benchmarks
   - Concurrent operation tests
   - Load testing scenarios

2. **Integration Testing**
   - ZeroDB integration
   - Strapi CMS integration
   - Design tokens integration
   - RLHF feedback collection

3. **UI/UX Testing**
   - Interactive mode testing
   - TUI (bubbletea) testing
   - Color output validation

4. **Security Testing**
   - API key handling
   - Sensitive data masking
   - Secure storage validation

## Resources

- **Go Testing Package**: https://pkg.go.dev/testing
- **Testify Assertions**: https://pkg.go.dev/github.com/stretchr/testify
- **GitHub Actions**: https://docs.github.com/en/actions
- **Makefile Documentation**: `make help`

## Support

For questions or issues with E2E tests:
1. Check this documentation
2. Review existing test examples in `tests/e2e/`
3. Check test artifacts for debugging
4. Create an issue with test logs and artifacts
