# Creating Tools Guide

## Overview

This guide explains how to create custom tools for AINative Code. Tools allow LLMs to perform actions like file operations, code execution, API calls, and more.

## Tool Interface

### Interface Definition

```go
// Location: internal/tools/interface.go
type Tool interface {
    // Name returns the unique identifier for the tool
    Name() string

    // Description returns a human-readable description of the tool's purpose
    Description() string

    // Schema returns the JSON schema defining the tool's input parameters
    Schema() ToolSchema

    // Execute runs the tool with the given input and returns the result
    Execute(ctx context.Context, input map[string]interface{}) (string, error)
}
```

### Core Types

```go
// ToolSchema defines the structure for tool input validation
type ToolSchema struct {
    Type        string                     `json:"type"`
    Properties  map[string]PropertySchema  `json:"properties"`
    Required    []string                   `json:"required,omitempty"`
    Description string                     `json:"description,omitempty"`
}

// PropertySchema defines a single input parameter
type PropertySchema struct {
    Type        string                     `json:"type"`
    Description string                     `json:"description,omitempty"`
    Enum        []interface{}             `json:"enum,omitempty"`
    Default     interface{}               `json:"default,omitempty"`
    Format      string                     `json:"format,omitempty"`
    Minimum     *float64                   `json:"minimum,omitempty"`
    Maximum     *float64                   `json:"maximum,omitempty"`
    MinLength   *int                       `json:"minLength,omitempty"`
    MaxLength   *int                       `json:"maxLength,omitempty"`
}
```

## Creating a Custom Tool

### Step 1: Define Tool Struct

```go
package mytools

import (
    "context"
    "fmt"

    "github.com/AINative-studio/ainative-code/internal/tools"
)

// MyTool implements a custom tool
type MyTool struct {
    // Configuration fields
    config MyToolConfig
}

type MyToolConfig struct {
    // Configuration options
    WorkingDir string
    Timeout    int
}
```

### Step 2: Implement Name and Description

```go
// Name returns the tool's unique identifier
func (t *MyTool) Name() string {
    return "my_tool"
}

// Description returns what the tool does
func (t *MyTool) Description() string {
    return "Performs a specific operation with the given parameters"
}
```

### Step 3: Define Schema

```go
// Schema returns the tool's input schema
func (t *MyTool) Schema() tools.ToolSchema {
    return tools.ToolSchema{
        Type: "object",
        Properties: map[string]tools.PropertySchema{
            "input_text": {
                Type:        "string",
                Description: "The text to process",
                MinLength:   intPtr(1),
                MaxLength:   intPtr(1000),
            },
            "mode": {
                Type:        "string",
                Description: "Processing mode",
                Enum:        []interface{}{"fast", "thorough", "detailed"},
                Default:     "fast",
            },
            "count": {
                Type:        "number",
                Description: "Number of iterations",
                Minimum:     float64Ptr(1),
                Maximum:     float64Ptr(100),
                Default:     10,
            },
        },
        Required: []string{"input_text"},
        Description: "Parameters for my tool execution",
    }
}

// Helper functions for pointers
func intPtr(i int) *int          { return &i }
func float64Ptr(f float64) *float64 { return &f }
```

### Step 4: Implement Execute

```go
// Execute runs the tool with validated input
func (t *MyTool) Execute(ctx context.Context, input map[string]interface{}) (string, error) {
    // 1. Extract and validate input
    inputText, ok := input["input_text"].(string)
    if !ok || inputText == "" {
        return "", fmt.Errorf("invalid or missing input_text")
    }

    mode := "fast" // default
    if m, ok := input["mode"].(string); ok {
        mode = m
    }

    count := 10 // default
    if c, ok := input["count"].(float64); ok {
        count = int(c)
    }

    // 2. Check context cancellation
    select {
    case <-ctx.Done():
        return "", ctx.Err()
    default:
    }

    // 3. Execute tool logic
    result, err := t.performOperation(ctx, inputText, mode, count)
    if err != nil {
        return "", fmt.Errorf("operation failed: %w", err)
    }

    // 4. Return formatted result
    return result, nil
}

func (t *MyTool) performOperation(ctx context.Context, text string, mode string, count int) (string, error) {
    // Your tool's actual logic here
    return fmt.Sprintf("Processed '%s' in %s mode %d times", text, mode, count), nil
}
```

## Tool Examples

### Example 1: File Reader Tool

```go
package builtin

import (
    "context"
    "fmt"
    "os"
    "path/filepath"

    "github.com/AINative-studio/ainative-code/internal/tools"
)

type FileReaderTool struct {
    workingDir string
}

func NewFileReaderTool(workingDir string) *FileReaderTool {
    return &FileReaderTool{workingDir: workingDir}
}

func (t *FileReaderTool) Name() string {
    return "read_file"
}

func (t *FileReaderTool) Description() string {
    return "Reads the contents of a file from the filesystem"
}

func (t *FileReaderTool) Schema() tools.ToolSchema {
    return tools.ToolSchema{
        Type: "object",
        Properties: map[string]tools.PropertySchema{
            "path": {
                Type:        "string",
                Description: "Path to the file to read (relative or absolute)",
            },
        },
        Required:    []string{"path"},
        Description: "Read file parameters",
    }
}

func (t *FileReaderTool) Execute(ctx context.Context, input map[string]interface{}) (string, error) {
    path, ok := input["path"].(string)
    if !ok {
        return "", fmt.Errorf("invalid path parameter")
    }

    // Resolve path
    if !filepath.IsAbs(path) {
        path = filepath.Join(t.workingDir, path)
    }

    // Security: prevent directory traversal
    cleanPath := filepath.Clean(path)
    if !filepath.IsAbs(cleanPath) || filepath.Dir(cleanPath) == ".." {
        return "", fmt.Errorf("invalid file path")
    }

    // Read file
    content, err := os.ReadFile(cleanPath)
    if err != nil {
        return "", fmt.Errorf("failed to read file: %w", err)
    }

    return string(content), nil
}
```

### Example 2: HTTP Request Tool

```go
package builtin

import (
    "context"
    "fmt"
    "io"
    "net/http"
    "time"

    "github.com/AINative-studio/ainative-code/internal/tools"
)

type HTTPRequestTool struct {
    client  *http.Client
    timeout time.Duration
}

func NewHTTPRequestTool(timeout time.Duration) *HTTPRequestTool {
    return &HTTPRequestTool{
        client: &http.Client{
            Timeout: timeout,
        },
        timeout: timeout,
    }
}

func (t *HTTPRequestTool) Name() string {
    return "http_request"
}

func (t *HTTPRequestTool) Description() string {
    return "Makes an HTTP request to a specified URL"
}

func (t *HTTPRequestTool) Schema() tools.ToolSchema {
    return tools.ToolSchema{
        Type: "object",
        Properties: map[string]tools.PropertySchema{
            "url": {
                Type:        "string",
                Description: "The URL to request",
                Format:      "uri",
            },
            "method": {
                Type:        "string",
                Description: "HTTP method",
                Enum:        []interface{}{"GET", "POST", "PUT", "DELETE"},
                Default:     "GET",
            },
        },
        Required: []string{"url"},
    }
}

func (t *HTTPRequestTool) Execute(ctx context.Context, input map[string]interface{}) (string, error) {
    url, ok := input["url"].(string)
    if !ok {
        return "", fmt.Errorf("invalid url parameter")
    }

    method := "GET"
    if m, ok := input["method"].(string); ok {
        method = m
    }

    // Create request
    req, err := http.NewRequestWithContext(ctx, method, url, nil)
    if err != nil {
        return "", fmt.Errorf("failed to create request: %w", err)
    }

    // Execute request
    resp, err := t.client.Do(req)
    if err != nil {
        return "", fmt.Errorf("request failed: %w", err)
    }
    defer resp.Body.Close()

    // Read response
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", fmt.Errorf("failed to read response: %w", err)
    }

    return fmt.Sprintf("Status: %d\nBody: %s", resp.StatusCode, string(body)), nil
}
```

### Example 3: Code Execution Tool

```go
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

type CodeExecutionTool struct {
    timeout    time.Duration
    allowedCmd []string
}

func NewCodeExecutionTool(timeout time.Duration) *CodeExecutionTool {
    return &CodeExecutionTool{
        timeout:    timeout,
        allowedCmd: []string{"python3", "node", "go", "bash"},
    }
}

func (t *CodeExecutionTool) Name() string {
    return "execute_code"
}

func (t *CodeExecutionTool) Description() string {
    return "Executes code in a specified language and returns the output"
}

func (t *CodeExecutionTool) Schema() tools.ToolSchema {
    return tools.ToolSchema{
        Type: "object",
        Properties: map[string]tools.PropertySchema{
            "language": {
                Type:        "string",
                Description: "Programming language",
                Enum:        []interface{}{"python", "javascript", "go", "bash"},
            },
            "code": {
                Type:        "string",
                Description: "Code to execute",
            },
        },
        Required: []string{"language", "code"},
    }
}

func (t *CodeExecutionTool) Execute(ctx context.Context, input map[string]interface{}) (string, error) {
    language, ok := input["language"].(string)
    if !ok {
        return "", fmt.Errorf("invalid language parameter")
    }

    code, ok := input["code"].(string)
    if !ok {
        return "", fmt.Errorf("invalid code parameter")
    }

    // Map language to command
    var cmd string
    switch language {
    case "python":
        cmd = "python3"
    case "javascript":
        cmd = "node"
    case "go":
        cmd = "go"
    case "bash":
        cmd = "bash"
    default:
        return "", fmt.Errorf("unsupported language: %s", language)
    }

    // Security: check allowed commands
    allowed := false
    for _, allowedCmd := range t.allowedCmd {
        if cmd == allowedCmd {
            allowed = true
            break
        }
    }
    if !allowed {
        return "", fmt.Errorf("command not allowed: %s", cmd)
    }

    // Create context with timeout
    execCtx, cancel := context.WithTimeout(ctx, t.timeout)
    defer cancel()

    // Execute code
    var stdout, stderr bytes.Buffer
    execCmd := exec.CommandContext(execCtx, cmd, "-c", code)
    execCmd.Stdout = &stdout
    execCmd.Stderr = &stderr

    err := execCmd.Run()

    // Build result
    result := strings.Builder{}
    if stdout.Len() > 0 {
        result.WriteString("STDOUT:\n")
        result.WriteString(stdout.String())
        result.WriteString("\n")
    }
    if stderr.Len() > 0 {
        result.WriteString("STDERR:\n")
        result.WriteString(stderr.String())
        result.WriteString("\n")
    }
    if err != nil {
        result.WriteString(fmt.Sprintf("ERROR: %v\n", err))
    }

    return result.String(), nil
}
```

## Tool Registration

### Register Tool

```go
package main

import (
    "github.com/AINative-studio/ainative-code/internal/tools"
    "github.com/AINative-studio/ainative-code/internal/tools/builtin"
)

func initializeTools() *tools.Registry {
    registry := tools.NewRegistry()

    // Register built-in tools
    registry.Register(builtin.NewFileReaderTool("."))
    registry.Register(builtin.NewHTTPRequestTool(30 * time.Second))
    registry.Register(builtin.NewCodeExecutionTool(60 * time.Second))

    // Register custom tools
    registry.Register(NewMyTool(MyToolConfig{
        WorkingDir: ".",
        Timeout:    30,
    }))

    return registry
}
```

### Get and Execute Tools

```go
func executeTool(registry *tools.Registry, name string, input map[string]interface{}) (string, error) {
    // Get tool by name
    tool, ok := registry.Get(name)
    if !ok {
        return "", fmt.Errorf("tool not found: %s", name)
    }

    // Validate input against schema
    if err := tools.ValidateInput(tool.Schema(), input); err != nil {
        return "", fmt.Errorf("invalid input: %w", err)
    }

    // Execute tool
    ctx := context.Background()
    result, err := tool.Execute(ctx, input)
    if err != nil {
        return "", fmt.Errorf("tool execution failed: %w", err)
    }

    return result, nil
}
```

## MCP Server Development

### What is MCP?

Model Context Protocol (MCP) is a standard for exposing tools to LLMs. An MCP server provides tools over a standard interface.

### Basic MCP Server

```go
package mcp

import (
    "context"
    "encoding/json"
    "fmt"
    "io"

    "github.com/AINative-studio/ainative-code/internal/tools"
)

// Server implements an MCP server
type Server struct {
    registry *tools.Registry
}

func NewServer(registry *tools.Registry) *Server {
    return &Server{registry: registry}
}

// Request represents an MCP request
type Request struct {
    Method string                 `json:"method"`
    Params map[string]interface{} `json:"params"`
}

// Response represents an MCP response
type Response struct {
    Result interface{} `json:"result,omitempty"`
    Error  *Error      `json:"error,omitempty"`
}

type Error struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
}

// HandleRequest processes an MCP request
func (s *Server) HandleRequest(ctx context.Context, req Request) Response {
    switch req.Method {
    case "tools/list":
        return s.handleListTools()
    case "tools/call":
        return s.handleCallTool(ctx, req.Params)
    default:
        return Response{
            Error: &Error{
                Code:    -32601,
                Message: fmt.Sprintf("method not found: %s", req.Method),
            },
        }
    }
}

func (s *Server) handleListTools() Response {
    tools := s.registry.List()

    toolDescs := make([]map[string]interface{}, len(tools))
    for i, tool := range tools {
        toolDescs[i] = map[string]interface{}{
            "name":        tool.Name(),
            "description": tool.Description(),
            "inputSchema": tool.Schema(),
        }
    }

    return Response{Result: toolDescs}
}

func (s *Server) handleCallTool(ctx context.Context, params map[string]interface{}) Response {
    // Extract tool name and arguments
    name, ok := params["name"].(string)
    if !ok {
        return Response{
            Error: &Error{Code: -32602, Message: "missing tool name"},
        }
    }

    arguments, ok := params["arguments"].(map[string]interface{})
    if !ok {
        arguments = make(map[string]interface{})
    }

    // Get tool
    tool, ok := s.registry.Get(name)
    if !ok {
        return Response{
            Error: &Error{Code: -32602, Message: fmt.Sprintf("tool not found: %s", name)},
        }
    }

    // Execute tool
    result, err := tool.Execute(ctx, arguments)
    if err != nil {
        return Response{
            Error: &Error{Code: -32603, Message: err.Error()},
        }
    }

    return Response{Result: result}
}

// Serve runs the MCP server on stdio
func (s *Server) Serve(ctx context.Context, reader io.Reader, writer io.Writer) error {
    decoder := json.NewDecoder(reader)
    encoder := json.NewEncoder(writer)

    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
        }

        var req Request
        if err := decoder.Decode(&req); err != nil {
            if err == io.EOF {
                return nil
            }
            return fmt.Errorf("failed to decode request: %w", err)
        }

        resp := s.HandleRequest(ctx, req)
        if err := encoder.Encode(resp); err != nil {
            return fmt.Errorf("failed to encode response: %w", err)
        }
    }
}
```

## Testing Tools

### Unit Tests

```go
package mytools

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestMyTool_Execute(t *testing.T) {
    tool := &MyTool{
        config: MyToolConfig{
            WorkingDir: ".",
            Timeout:    30,
        },
    }

    tests := []struct {
        name    string
        input   map[string]interface{}
        want    string
        wantErr bool
    }{
        {
            name: "valid input",
            input: map[string]interface{}{
                "input_text": "test",
                "mode":       "fast",
                "count":      5,
            },
            want:    "Processed 'test' in fast mode 5 times",
            wantErr: false,
        },
        {
            name: "missing required field",
            input: map[string]interface{}{
                "mode": "fast",
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := tool.Execute(context.Background(), tt.input)

            if tt.wantErr {
                assert.Error(t, err)
                return
            }

            require.NoError(t, err)
            assert.Equal(t, tt.want, got)
        })
    }
}

func TestMyTool_Schema(t *testing.T) {
    tool := &MyTool{}
    schema := tool.Schema()

    assert.Equal(t, "object", schema.Type)
    assert.Contains(t, schema.Properties, "input_text")
    assert.Contains(t, schema.Required, "input_text")
}
```

## Best Practices

### 1. Input Validation

Always validate input thoroughly:

```go
func (t *Tool) Execute(ctx context.Context, input map[string]interface{}) (string, error) {
    // Type checking
    value, ok := input["param"].(string)
    if !ok {
        return "", fmt.Errorf("invalid param type")
    }

    // Range checking
    if len(value) > 1000 {
        return "", fmt.Errorf("param too long")
    }

    // Business logic validation
    if !isValid(value) {
        return "", fmt.Errorf("param validation failed")
    }

    // Continue execution...
}
```

### 2. Security Considerations

- Validate and sanitize all inputs
- Prevent directory traversal attacks
- Limit resource usage (timeout, memory)
- Sandbox code execution when possible
- Never trust user input

### 3. Error Handling

Provide clear, actionable error messages:

```go
if err != nil {
    return "", fmt.Errorf("failed to read file %s: %w", path, err)
}
```

### 4. Context Usage

Respect context cancellation:

```go
select {
case <-ctx.Done():
    return "", ctx.Err()
default:
    // Continue
}
```

### 5. Documentation

Document your tool clearly:

```go
// MyTool performs X operation on Y data.
// It supports modes: fast, thorough, and detailed.
//
// Example:
//   result, err := tool.Execute(ctx, map[string]interface{}{
//       "input_text": "hello",
//       "mode": "fast",
//   })
```

## Resources

- [Tool Interface](../../internal/tools/interface.go)
- [Tool Registry](../../internal/tools/registry.go)
- [Built-in Tools](../../internal/tools/builtin/)
- [MCP Server](../../internal/mcp/)

---

**Last Updated**: 2025-01-05
