# Tools API Reference

**Import Path**: `github.com/AINative-studio/ainative-code/internal/tools`

The tools package provides an extensible tool execution framework for LLM assistants, enabling filesystem operations, network requests, system commands, and database interactions.

## Table of Contents

- [Tool Interface](#tool-interface)
- [Tool Registry](#tool-registry)
- [Built-in Tools](#built-in-tools)
- [MCP Integration](#mcp-integration)
- [Custom Tool Development](#custom-tool-development)
- [Usage Examples](#usage-examples)

## Tool Interface

### Tool

```go
type Tool interface {
    Name() string
    Description() string
    Schema() ToolSchema
    Execute(ctx context.Context, input map[string]interface{}) (*Result, error)
    Category() Category
    RequiresConfirmation() bool
}
```

The Tool interface defines the contract that all tools must implement.

**Methods**:

- `Name` - Returns the unique name of the tool
- `Description` - Returns a human-readable description
- `Schema` - Returns the JSON schema for input parameters
- `Execute` - Runs the tool with provided input
- `Category` - Returns the category (filesystem, network, system, database, text)
- `RequiresConfirmation` - Returns true if user confirmation is needed

### Types

#### Category

```go
type Category string

const (
    CategoryFilesystem Category = "filesystem"
    CategoryNetwork    Category = "network"
    CategorySystem     Category = "system"
    CategoryDatabase   Category = "database"
    CategoryText       Category = "text"
)
```

#### ToolSchema

```go
type ToolSchema struct {
    Type       string                 `json:"type"`
    Properties map[string]PropertyDef `json:"properties"`
    Required   []string               `json:"required,omitempty"`
}
```

Defines the JSON schema for tool input validation.

#### PropertyDef

```go
type PropertyDef struct {
    Type        string      `json:"type"`
    Description string      `json:"description"`
    Enum        []string    `json:"enum,omitempty"`
    Default     interface{} `json:"default,omitempty"`
    MinLength   *int        `json:"minLength,omitempty"`
    MaxLength   *int        `json:"maxLength,omitempty"`
    Pattern     string      `json:"pattern,omitempty"`
}
```

Defines a property in the tool schema.

#### Result

```go
type Result struct {
    Success  bool                   `json:"success"`
    Output   string                 `json:"output"`
    Error    error                  `json:"error,omitempty"`
    Metadata map[string]interface{} `json:"metadata,omitempty"`
}
```

Represents the result of a tool execution.

#### ExecutionContext

```go
type ExecutionContext struct {
    Timeout          time.Duration
    WorkingDirectory string
    Environment      map[string]string
    AllowedPaths     []string
    MaxOutputSize    int64
    DryRun           bool
}
```

Contains settings for tool execution.

### Execution Options

```go
func WithTimeout(timeout time.Duration) ExecutionOption
func WithWorkingDirectory(dir string) ExecutionOption
func WithEnvironment(env map[string]string) ExecutionOption
func WithAllowedPaths(paths []string) ExecutionOption
func WithMaxOutputSize(size int64) ExecutionOption
func WithDryRun(dryRun bool) ExecutionOption
```

**Example**:

```go
execCtx := tools.NewExecutionContext(
    tools.WithTimeout(30 * time.Second),
    tools.WithWorkingDirectory("/tmp"),
    tools.WithMaxOutputSize(10 * 1024 * 1024), // 10MB
)
```

## Tool Registry

### Registry

The Registry manages tool registration and execution with thread safety.

#### NewRegistry

```go
func NewRegistry() *Registry
```

Creates a new Registry instance.

#### Register

```go
func (r *Registry) Register(tool Tool) error
```

Registers a new tool in the registry.

**Example**:

```go
registry := tools.NewRegistry()

tool := &MyCustomTool{
    name:        "my_tool",
    description: "Does something useful",
}

if err := registry.Register(tool); err != nil {
    log.Fatalf("Failed to register tool: %v", err)
}
```

#### Get

```go
func (r *Registry) Get(name string) (Tool, error)
```

Retrieves a tool by name.

**Example**:

```go
tool, err := registry.Get("file_read")
if err != nil {
    log.Fatalf("Tool not found: %v", err)
}
```

#### List

```go
func (r *Registry) List() []Tool
```

Returns all registered tools.

#### ListByCategory

```go
func (r *Registry) ListByCategory(category Category) []Tool
```

Returns all tools in a specific category.

**Example**:

```go
// Get all filesystem tools
fsTools := registry.ListByCategory(tools.CategoryFilesystem)
for _, tool := range fsTools {
    fmt.Printf("Tool: %s - %s\n", tool.Name(), tool.Description())
}
```

#### Execute

```go
func (r *Registry) Execute(ctx context.Context, toolName string, input map[string]interface{}, execCtx *ExecutionContext) (*Result, error)
```

Executes a tool with validation, timeout enforcement, and error handling.

**Example**:

```go
ctx := context.Background()

input := map[string]interface{}{
    "path": "/tmp/test.txt",
    "content": "Hello, World!",
}

execCtx := tools.NewExecutionContext(
    tools.WithTimeout(5 * time.Second),
)

result, err := registry.Execute(ctx, "file_write", input, execCtx)
if err != nil {
    log.Fatalf("Tool execution failed: %v", err)
}

if result.Success {
    fmt.Printf("Output: %s\n", result.Output)
}
```

#### Schemas

```go
func (r *Registry) Schemas() map[string]ToolSchema
```

Returns a map of tool names to their schemas.

## Built-in Tools

### Filesystem Tools

#### file_read

Reads content from a file.

**Schema**:
```json
{
    "type": "object",
    "properties": {
        "path": {
            "type": "string",
            "description": "Path to the file to read"
        }
    },
    "required": ["path"]
}
```

**Example**:

```go
result, err := registry.Execute(ctx, "file_read", map[string]interface{}{
    "path": "/tmp/data.txt",
}, nil)
```

#### file_write

Writes content to a file.

**Schema**:
```json
{
    "type": "object",
    "properties": {
        "path": {"type": "string", "description": "File path"},
        "content": {"type": "string", "description": "Content to write"}
    },
    "required": ["path", "content"]
}
```

#### file_delete

Deletes a file.

#### file_list

Lists files in a directory.

### System Tools

#### bash_execute

Executes a bash command.

**Schema**:
```json
{
    "type": "object",
    "properties": {
        "command": {"type": "string", "description": "Command to execute"},
        "timeout": {"type": "integer", "description": "Timeout in seconds"}
    },
    "required": ["command"]
}
```

**Example**:

```go
result, err := registry.Execute(ctx, "bash_execute", map[string]interface{}{
    "command": "ls -la /tmp",
    "timeout": 10,
}, execCtx)
```

## MCP Integration

**Import Path**: `github.com/AINative-studio/ainative-code/internal/mcp`

The MCP (Model Context Protocol) package provides integration with external tool servers.

### Types

#### Client

```go
type Client struct {
    // Contains filtered or unexported fields
}
```

MCP protocol client for communicating with tool servers.

#### Server

```go
type Server struct {
    Name    string
    URL     string
    Timeout time.Duration
    Headers map[string]string
}
```

MCP server configuration.

### Functions

#### NewClient

```go
func NewClient(server *Server) *Client
```

Creates a new MCP client for the given server.

**Example**:

```go
server := &mcp.Server{
    Name:    "my-mcp-server",
    URL:     "http://localhost:8080/mcp",
    Timeout: 30 * time.Second,
    Headers: map[string]string{
        "Authorization": "Bearer token",
    },
}

client := mcp.NewClient(server)
```

### Methods

#### ListTools

```go
func (c *Client) ListTools(ctx context.Context) ([]Tool, error)
```

Retrieves all available tools from the MCP server.

**Example**:

```go
tools, err := client.ListTools(ctx)
if err != nil {
    log.Fatalf("Failed to list tools: %v", err)
}

for _, tool := range tools {
    fmt.Printf("Tool: %s - %s\n", tool.Name, tool.Description)
}
```

#### CallTool

```go
func (c *Client) CallTool(ctx context.Context, name string, arguments map[string]interface{}) (*ToolResult, error)
```

Executes a tool on the MCP server.

**Example**:

```go
result, err := client.CallTool(ctx, "database_query", map[string]interface{}{
    "query": "SELECT * FROM users LIMIT 10",
})
if err != nil {
    log.Fatalf("Tool execution failed: %v", err)
}

fmt.Printf("Result: %s\n", result.Content)
```

#### Ping

```go
func (c *Client) Ping(ctx context.Context) error
```

Checks if the MCP server is reachable.

#### CheckHealth

```go
func (c *Client) CheckHealth(ctx context.Context) *HealthStatus
```

Checks the health status of the MCP server.

**Example**:

```go
health := client.CheckHealth(ctx)
if health.Healthy {
    fmt.Printf("Server is healthy (response time: %v)\n", health.ResponseTime)
} else {
    fmt.Printf("Server is unhealthy: %s\n", health.Error)
}
```

## Custom Tool Development

### Creating a Custom Tool

```go
package main

import (
    "context"
    "fmt"
    "strings"

    "github.com/AINative-studio/ainative-code/internal/tools"
)

// StringReverseTool reverses a string
type StringReverseTool struct{}

func (t *StringReverseTool) Name() string {
    return "string_reverse"
}

func (t *StringReverseTool) Description() string {
    return "Reverses a string"
}

func (t *StringReverseTool) Schema() tools.ToolSchema {
    return tools.ToolSchema{
        Type: "object",
        Properties: map[string]tools.PropertyDef{
            "text": {
                Type:        "string",
                Description: "The string to reverse",
            },
        },
        Required: []string{"text"},
    }
}

func (t *StringReverseTool) Execute(ctx context.Context, input map[string]interface{}) (*tools.Result, error) {
    // Extract and validate input
    text, ok := input["text"].(string)
    if !ok {
        return nil, fmt.Errorf("invalid input: 'text' must be a string")
    }

    // Perform the operation
    runes := []rune(text)
    for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
        runes[i], runes[j] = runes[j], runes[i]
    }
    reversed := string(runes)

    // Return result
    return &tools.Result{
        Success: true,
        Output:  reversed,
        Metadata: map[string]interface{}{
            "original_length": len(text),
            "reversed_length": len(reversed),
        },
    }, nil
}

func (t *StringReverseTool) Category() tools.Category {
    return tools.CategoryText
}

func (t *StringReverseTool) RequiresConfirmation() bool {
    return false
}

// Usage
func main() {
    registry := tools.NewRegistry()

    // Register the tool
    tool := &StringReverseTool{}
    if err := registry.Register(tool); err != nil {
        log.Fatal(err)
    }

    // Execute the tool
    ctx := context.Background()
    result, err := registry.Execute(ctx, "string_reverse", map[string]interface{}{
        "text": "Hello, World!",
    }, nil)

    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Reversed: %s\n", result.Output)
}
```

### Advanced Custom Tool with Validation

```go
type HTTPRequestTool struct {
    client *http.Client
}

func NewHTTPRequestTool() *HTTPRequestTool {
    return &HTTPRequestTool{
        client: &http.Client{
            Timeout: 30 * time.Second,
        },
    }
}

func (t *HTTPRequestTool) Name() string {
    return "http_request"
}

func (t *HTTPRequestTool) Description() string {
    return "Makes an HTTP request and returns the response"
}

func (t *HTTPRequestTool) Schema() tools.ToolSchema {
    methodEnum := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

    return tools.ToolSchema{
        Type: "object",
        Properties: map[string]tools.PropertyDef{
            "url": {
                Type:        "string",
                Description: "The URL to request",
                Pattern:     "^https?://",
            },
            "method": {
                Type:        "string",
                Description: "HTTP method",
                Enum:        methodEnum,
                Default:     "GET",
            },
            "headers": {
                Type:        "object",
                Description: "HTTP headers",
            },
            "body": {
                Type:        "string",
                Description: "Request body",
            },
        },
        Required: []string{"url"},
    }
}

func (t *HTTPRequestTool) Execute(ctx context.Context, input map[string]interface{}) (*tools.Result, error) {
    // Extract parameters
    url := input["url"].(string)
    method, _ := input["method"].(string)
    if method == "" {
        method = "GET"
    }

    // Create request
    req, err := http.NewRequestWithContext(ctx, method, url, nil)
    if err != nil {
        return &tools.Result{
            Success: false,
            Error:   err,
        }, nil
    }

    // Add headers
    if headers, ok := input["headers"].(map[string]interface{}); ok {
        for key, value := range headers {
            if strValue, ok := value.(string); ok {
                req.Header.Set(key, strValue)
            }
        }
    }

    // Execute request
    resp, err := t.client.Do(req)
    if err != nil {
        return &tools.Result{
            Success: false,
            Error:   err,
        }, nil
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return &tools.Result{
            Success: false,
            Error:   err,
        }, nil
    }

    return &tools.Result{
        Success: true,
        Output:  string(body),
        Metadata: map[string]interface{}{
            "status_code": resp.StatusCode,
            "headers":     resp.Header,
        },
    }, nil
}

func (t *HTTPRequestTool) Category() tools.Category {
    return tools.CategoryNetwork
}

func (t *HTTPRequestTool) RequiresConfirmation() bool {
    return true // Network requests should require confirmation
}
```

## Usage Examples

### Basic Tool Execution

```go
package main

import (
    "context"
    "log"

    "github.com/AINative-studio/ainative-code/internal/tools"
)

func main() {
    ctx := context.Background()
    registry := tools.NewRegistry()

    // Register built-in tools
    // (assume they're registered elsewhere)

    // Execute a file read
    result, err := registry.Execute(ctx, "file_read", map[string]interface{}{
        "path": "/tmp/data.txt",
    }, nil)

    if err != nil {
        log.Fatalf("Execution failed: %v", err)
    }

    if result.Success {
        log.Printf("File content: %s", result.Output)
    } else {
        log.Printf("Tool failed: %v", result.Error)
    }
}
```

### Tool Execution with Timeout

```go
execCtx := tools.NewExecutionContext(
    tools.WithTimeout(5 * time.Second),
    tools.WithMaxOutputSize(1024 * 1024), // 1MB
)

result, err := registry.Execute(ctx, "bash_execute", map[string]interface{}{
    "command": "sleep 10 && echo done",
}, execCtx)

if err != nil {
    // Will timeout after 5 seconds
    log.Printf("Execution timed out: %v", err)
}
```

### Safe File Operations

```go
// Restrict file operations to specific paths
execCtx := tools.NewExecutionContext(
    tools.WithAllowedPaths([]string{"/tmp", "/var/data"}),
)

// This will succeed
result, _ := registry.Execute(ctx, "file_read", map[string]interface{}{
    "path": "/tmp/safe.txt",
}, execCtx)

// This will fail (path not in AllowedPaths)
result, err := registry.Execute(ctx, "file_read", map[string]interface{}{
    "path": "/etc/passwd",
}, execCtx)
```

### Dry Run Mode

```go
// Test tool execution without actually running
execCtx := tools.NewExecutionContext(
    tools.WithDryRun(true),
)

result, err := registry.Execute(ctx, "file_delete", map[string]interface{}{
    "path": "/important/file.txt",
}, execCtx)

// Will return success without deleting the file
fmt.Printf("Dry run result: %s\n", result.Output)
```

### Listing Available Tools

```go
// List all tools
allTools := registry.List()
for _, tool := range allTools {
    fmt.Printf("Tool: %s (%s)\n", tool.Name(), tool.Category())
    fmt.Printf("  Description: %s\n", tool.Description())
    fmt.Printf("  Requires confirmation: %v\n", tool.RequiresConfirmation())
}

// List tools by category
fsTools := registry.ListByCategory(tools.CategoryFilesystem)
fmt.Printf("\nFilesystem tools: %d\n", len(fsTools))
```

## Best Practices

### 1. Always Validate Input

```go
func (t *MyTool) Execute(ctx context.Context, input map[string]interface{}) (*tools.Result, error) {
    // Validate required fields
    value, ok := input["required_field"].(string)
    if !ok {
        return &tools.Result{
            Success: false,
            Error:   fmt.Errorf("missing or invalid 'required_field'"),
        }, nil
    }

    // Validate value constraints
    if len(value) == 0 {
        return &tools.Result{
            Success: false,
            Error:   fmt.Errorf("'required_field' cannot be empty"),
        }, nil
    }

    // Proceed with execution...
}
```

### 2. Handle Context Cancellation

```go
func (t *MyTool) Execute(ctx context.Context, input map[string]interface{}) (*tools.Result, error) {
    // Check context before expensive operations
    if err := ctx.Err(); err != nil {
        return &tools.Result{
            Success: false,
            Error:   err,
        }, nil
    }

    // Perform work...
}
```

### 3. Provide Useful Metadata

```go
return &tools.Result{
    Success: true,
    Output:  output,
    Metadata: map[string]interface{}{
        "execution_time_ms": elapsed.Milliseconds(),
        "bytes_processed":   bytesProcessed,
        "items_count":       itemCount,
    },
}
```

### 4. Use Proper Error Handling

```go
// Return operational errors in Result
if err != nil {
    return &tools.Result{
        Success: false,
        Error:   err,
        Output:  partialOutput,
    }, nil
}

// Return system errors directly
if systemErr != nil {
    return nil, fmt.Errorf("system error: %w", systemErr)
}
```

### 5. Implement Schema Validation

```go
func (t *MyTool) Schema() tools.ToolSchema {
    minLen := 1
    maxLen := 1000

    return tools.ToolSchema{
        Type: "object",
        Properties: map[string]tools.PropertyDef{
            "text": {
                Type:        "string",
                Description: "Input text",
                MinLength:   &minLen,
                MaxLength:   &maxLen,
            },
        },
        Required: []string{"text"},
    }
}
```

## Related Documentation

- [MCP Integration](../user-guide/mcp-tools.md) - MCP tool servers
- [Configuration](configuration.md) - Tool configuration
- [Errors](errors.md) - Error handling
