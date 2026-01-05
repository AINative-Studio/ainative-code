// Package mcp implements the Model Context Protocol for custom tool registration.
//
// The Model Context Protocol (MCP) is a JSON-RPC based protocol that enables
// external servers to register and provide custom tools to the AI assistant.
// This allows for extensibility and integration with external services.
//
// Features:
//   - JSON-RPC 2.0 protocol implementation
//   - Tool discovery from MCP servers
//   - Dynamic tool schema parsing
//   - Tool execution delegation
//   - Error handling and recovery
//   - Server health monitoring
//   - Configuration management
//
// Protocol Flow:
//   1. Client connects to MCP server
//   2. Client requests available tools (list_tools)
//   3. Server responds with tool schemas
//   4. Client executes tool (call_tool)
//   5. Server executes and returns result
//
// Example usage:
//
//	import "github.com/AINative-studio/ainative-code/internal/mcp"
//
//	// Create MCP client
//	client, err := mcp.NewClient("http://localhost:8080")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Discover tools
//	tools, err := client.ListTools(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Execute tool
//	result, err := client.CallTool(ctx, "my_tool", map[string]interface{}{
//	    "param1": "value1",
//	})
package mcp
