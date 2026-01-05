package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"
	"time"
)

// Client represents an MCP protocol client.
type Client struct {
	server     *Server
	httpClient *http.Client
	requestID  atomic.Uint64
}

// NewClient creates a new MCP client for the given server.
func NewClient(server *Server) *Client {
	timeout := server.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &Client{
		server: server,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// ListTools retrieves all available tools from the MCP server.
func (c *Client) ListTools(ctx context.Context) ([]Tool, error) {
	var allTools []Tool
	var cursor string

	for {
		params := ListToolsParams{
			Cursor: cursor,
		}

		var result ListToolsResult
		if err := c.call(ctx, "tools/list", params, &result); err != nil {
			return nil, fmt.Errorf("failed to list tools: %w", err)
		}

		allTools = append(allTools, result.Tools...)

		// Check if there are more pages
		if result.NextCursor == "" {
			break
		}
		cursor = result.NextCursor
	}

	return allTools, nil
}

// CallTool executes a tool on the MCP server.
func (c *Client) CallTool(ctx context.Context, name string, arguments map[string]interface{}) (*ToolResult, error) {
	params := CallToolParams{
		Name:      name,
		Arguments: arguments,
	}

	var result ToolResult
	if err := c.call(ctx, "tools/call", params, &result); err != nil {
		return nil, fmt.Errorf("failed to call tool %s: %w", name, err)
	}

	return &result, nil
}

// Ping checks if the MCP server is reachable.
func (c *Client) Ping(ctx context.Context) error {
	var result interface{}
	if err := c.call(ctx, "ping", nil, &result); err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}
	return nil
}

// CheckHealth checks the health status of the MCP server.
func (c *Client) CheckHealth(ctx context.Context) *HealthStatus {
	start := time.Now()
	status := &HealthStatus{
		LastChecked: start,
	}

	err := c.Ping(ctx)
	status.ResponseTime = time.Since(start)

	if err != nil {
		status.Healthy = false
		status.Error = err.Error()
	} else {
		status.Healthy = true
	}

	return status
}

// call performs a JSON-RPC call to the MCP server.
func (c *Client) call(ctx context.Context, method string, params interface{}, result interface{}) error {
	// Generate unique request ID
	requestID := c.requestID.Add(1)

	// Build JSON-RPC request
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      requestID,
		Method:  method,
		Params:  params,
	}

	// Marshal request
	reqBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.server.URL, bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	for key, value := range c.server.Headers {
		httpReq.Header.Set(key, value)
	}

	// Send request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP error: %d %s", resp.StatusCode, string(respBody))
	}

	// Parse JSON-RPC response
	var rpcResp JSONRPCResponse
	if err := json.Unmarshal(respBody, &rpcResp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for JSON-RPC error
	if rpcResp.Error != nil {
		return rpcResp.Error
	}

	// Unmarshal result into target
	if result != nil && rpcResp.Result != nil {
		resultBytes, err := json.Marshal(rpcResp.Result)
		if err != nil {
			return fmt.Errorf("failed to marshal result: %w", err)
		}

		if err := json.Unmarshal(resultBytes, result); err != nil {
			return fmt.Errorf("failed to unmarshal result: %w", err)
		}
	}

	return nil
}

// GetServer returns the server configuration.
func (c *Client) GetServer() *Server {
	return c.server
}
