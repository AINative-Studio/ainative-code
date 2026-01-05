package mcp

import "time"

// Tool represents a tool available from an MCP server.
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// ToolResult represents the result of a tool execution.
type ToolResult struct {
	Content []ResultContent `json:"content"`
	IsError bool            `json:"isError,omitempty"`
}

// ResultContent represents content in a tool result.
type ResultContent struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
	Data string `json:"data,omitempty"`
}

// Server represents an MCP server configuration.
type Server struct {
	Name        string            `json:"name"`
	URL         string            `json:"url"`
	Timeout     time.Duration     `json:"timeout"`
	Headers     map[string]string `json:"headers,omitempty"`
	Enabled     bool              `json:"enabled"`
	Description string            `json:"description,omitempty"`
}

// JSONRPCRequest represents a JSON-RPC 2.0 request.
type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response.
type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

// RPCError represents a JSON-RPC error.
type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Error implements the error interface for RPCError.
func (e *RPCError) Error() string {
	return e.Message
}

// ListToolsParams represents parameters for list_tools request.
type ListToolsParams struct {
	Cursor string `json:"cursor,omitempty"`
}

// ListToolsResult represents the result of list_tools request.
type ListToolsResult struct {
	Tools      []Tool `json:"tools"`
	NextCursor string `json:"nextCursor,omitempty"`
}

// CallToolParams represents parameters for call_tool request.
type CallToolParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// HealthStatus represents the health status of an MCP server.
type HealthStatus struct {
	Healthy      bool          `json:"healthy"`
	LastChecked  time.Time     `json:"lastChecked"`
	ResponseTime time.Duration `json:"responseTime"`
	Error        string        `json:"error,omitempty"`
}
