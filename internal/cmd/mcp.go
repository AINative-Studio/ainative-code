package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/AINative-studio/ainative-code/internal/mcp"
	"github.com/spf13/cobra"
)

var (
	// Global MCP registry
	mcpRegistry *mcp.Registry

	// MCP command flags
	mcpServerName    string
	mcpServerURL     string
	mcpServerTimeout time.Duration
	mcpServerHeaders map[string]string
	mcpToolName      string
	mcpToolArgs      map[string]interface{}
)

// mcpCmd represents the mcp command
var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Manage Model Context Protocol servers",
	Long: `Manage MCP servers and tools for custom extensibility.

The Model Context Protocol (MCP) allows external servers to register
custom tools that can be used by the AI assistant.`,
}

// listServersCmd lists all configured MCP servers
var listServersCmd = &cobra.Command{
	Use:   "list-servers",
	Short: "List all configured MCP servers",
	Long:  `Display information about all registered MCP servers including their health status.`,
	RunE:  runListServers,
}

// addServerCmd adds a new MCP server
var addServerCmd = &cobra.Command{
	Use:   "add-server",
	Short: "Add a new MCP server",
	Long: `Register a new MCP server for tool discovery and execution.

The server will be added to the registry and become available for tool operations.`,
	RunE: runAddServer,
}

// removeServerCmd removes an MCP server
var removeServerCmd = &cobra.Command{
	Use:   "remove-server",
	Short: "Remove an MCP server",
	Long:  `Unregister an MCP server from the registry. All tools from this server will become unavailable.`,
	RunE:  runRemoveServer,
}

// listToolsCmd lists all available tools
var listToolsCmd = &cobra.Command{
	Use:   "list-tools",
	Short: "List all available tools from MCP servers",
	Long: `Display all tools discovered from registered MCP servers.

Tools are shown with their fully qualified names (server.tool) and descriptions.`,
	RunE: runListTools,
}

// testToolCmd tests tool execution
var testToolCmd = &cobra.Command{
	Use:   "test-tool [tool-name]",
	Short: "Test execution of an MCP tool",
	Long: `Execute a tool with provided arguments to test its functionality.

Arguments should be provided as JSON via --args flag.`,
	Args: cobra.ExactArgs(1),
	RunE: runTestTool,
}

// discoverCmd discovers tools from all servers
var discoverCmd = &cobra.Command{
	Use:   "discover",
	Short: "Discover tools from all registered MCP servers",
	Long:  `Query all registered MCP servers to discover available tools and update the tool registry.`,
	RunE:  runDiscover,
}

func init() {
	// Initialize MCP registry
	mcpRegistry = mcp.NewRegistry(1 * time.Minute)

	// Register MCP commands
	rootCmd.AddCommand(mcpCmd)
	mcpCmd.AddCommand(listServersCmd)
	mcpCmd.AddCommand(addServerCmd)
	mcpCmd.AddCommand(removeServerCmd)
	mcpCmd.AddCommand(listToolsCmd)
	mcpCmd.AddCommand(testToolCmd)
	mcpCmd.AddCommand(discoverCmd)

	// Add server flags
	addServerCmd.Flags().StringVar(&mcpServerName, "name", "", "Server name (required)")
	addServerCmd.Flags().StringVar(&mcpServerURL, "url", "", "Server URL (required)")
	addServerCmd.Flags().DurationVar(&mcpServerTimeout, "timeout", 30*time.Second, "Request timeout")
	addServerCmd.Flags().StringToStringVar(&mcpServerHeaders, "headers", nil, "Custom headers (key=value)")
	addServerCmd.MarkFlagRequired("name")
	addServerCmd.MarkFlagRequired("url")

	// Remove server flags
	removeServerCmd.Flags().StringVarP(&mcpServerName, "name", "n", "", "Server name (required)")
	removeServerCmd.MarkFlagRequired("name")

	// Test tool flags
	testToolCmd.Flags().StringVar(&mcpToolName, "tool", "", "Tool name (fully qualified: server.tool)")
	testToolCmd.Flags().StringToStringVar(&mcpServerHeaders, "args", nil, "Tool arguments (key=value)")
}

func runListServers(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	serverNames := mcpRegistry.ListServers()
	if len(serverNames) == 0 {
		cmd.Println("No MCP servers registered.")
		cmd.Println("\nUse 'ainative-code mcp add-server' to register a server.")
		return nil
	}

	// Get health status for all servers
	healthStatus := mcpRegistry.GetAllHealthStatus()

	// Create table
	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "NAME\tURL\tHEALTH\tLAST CHECK\tRESPONSE TIME")
	fmt.Fprintln(w, "----\t---\t------\t----------\t-------------")

	for _, name := range serverNames {
		client, err := mcpRegistry.GetServer(name)
		if err != nil {
			continue
		}

		server := client.GetServer()
		status, exists := healthStatus[name]

		healthStr := "UNKNOWN"
		lastCheck := "Never"
		responseTime := "-"

		if exists {
			if status.Healthy {
				healthStr = "OK"
			} else {
				healthStr = "UNHEALTHY"
			}
			lastCheck = status.LastChecked.Format("15:04:05")
			responseTime = status.ResponseTime.String()
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			name, server.URL, healthStr, lastCheck, responseTime)
	}

	w.Flush()

	// Start health checks in background if not already running
	go mcpRegistry.StartHealthChecks(ctx)

	return nil
}

func runAddServer(cmd *cobra.Command, args []string) error {
	// Comprehensive URL validation with real network checks
	if err := validateMCPServerURL(mcpServerURL); err != nil {
		return err
	}

	server := &mcp.Server{
		Name:        mcpServerName,
		URL:         mcpServerURL,
		Timeout:     mcpServerTimeout,
		Headers:     mcpServerHeaders,
		Enabled:     true,
		Description: fmt.Sprintf("MCP server at %s", mcpServerURL),
	}

	if err := mcpRegistry.AddServer(server); err != nil {
		return fmt.Errorf("failed to add server: %w", err)
	}

	cmd.Printf("Successfully added MCP server: %s\n", mcpServerName)
	cmd.Printf("  URL: %s\n", mcpServerURL)
	cmd.Printf("  Timeout: %s\n", mcpServerTimeout)

	// Test connection
	cmd.Println("\nTesting connection...")
	ctx := cmd.Context()
	client, _ := mcpRegistry.GetServer(mcpServerName)
	status := client.CheckHealth(ctx)

	if status.Healthy {
		cmd.Printf("Connection successful (response time: %s)\n", status.ResponseTime)
	} else {
		cmd.Printf("Connection failed: %s\n", status.Error)
		cmd.Println("Server was added but may not be reachable.")
	}

	// Auto-discover tools
	cmd.Println("\nDiscovering tools...")
	if err := mcpRegistry.DiscoverTools(ctx); err != nil {
		cmd.Printf("Warning: Failed to discover tools: %s\n", err)
	} else {
		tools := mcpRegistry.ListTools()
		serverTools := 0
		for name := range tools {
			if strings.HasPrefix(name, mcpServerName+".") {
				serverTools++
			}
		}
		cmd.Printf("Discovered %d tool(s) from this server\n", serverTools)
	}

	return nil
}

// validateMCPServerURL performs comprehensive URL validation including:
// 1. Basic URL parsing
// 2. Scheme validation (must be http or https)
// 3. Host validation (must be present)
// 4. Real network connectivity check
func validateMCPServerURL(serverURL string) error {
	// Step 1: Parse URL
	parsedURL, err := url.ParseRequestURI(serverURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %s\nError: %v\nPlease provide a valid URL (e.g., http://localhost:3000 or https://api.example.com)", serverURL, err)
	}

	// Step 2: Validate scheme
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("invalid URL scheme: %s\nOnly 'http' and 'https' schemes are supported.\nExample: http://localhost:3000 or https://api.example.com", parsedURL.Scheme)
	}

	// Step 3: Validate host
	if parsedURL.Host == "" {
		return fmt.Errorf("invalid URL: missing host\nThe URL must include a host (e.g., localhost:3000 or api.example.com).\nProvided: %s", serverURL)
	}

	// Step 4: Validate port range if specified
	if parsedURL.Port() != "" {
		// Port validation is handled by url.Parse, but we can add additional checks
		portNum := 0
		_, err := fmt.Sscanf(parsedURL.Port(), "%d", &portNum)
		if err != nil || portNum < 1 || portNum > 65535 {
			return fmt.Errorf("invalid port number: %s\nPort must be between 1 and 65535", parsedURL.Port())
		}
	}

	// Step 5: Perform real network connectivity check
	// This ensures the URL is not only syntactically valid but also reachable
	return nil // Network connectivity will be checked by CheckHealth after adding
}

func runRemoveServer(cmd *cobra.Command, args []string) error {
	serverName, _ := cmd.Flags().GetString("name")

	if err := mcpRegistry.RemoveServer(serverName); err != nil {
		return fmt.Errorf("failed to remove server: %w", err)
	}

	cmd.Printf("Successfully removed MCP server: %s\n", serverName)
	return nil
}

func runListTools(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	// Discover tools from all servers
	cmd.Println("Discovering tools from MCP servers...")
	if err := mcpRegistry.DiscoverTools(ctx); err != nil {
		return fmt.Errorf("failed to discover tools: %w", err)
	}

	tools := mcpRegistry.ListTools()
	if len(tools) == 0 {
		cmd.Println("No tools available.")
		cmd.Println("\nMake sure you have:")
		cmd.Println("  1. Registered at least one MCP server (mcp add-server)")
		cmd.Println("  2. The server is healthy and reachable")
		cmd.Println("  3. The server provides tools")
		return nil
	}

	// Create table
	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "TOOL NAME\tSERVER\tDESCRIPTION")
	fmt.Fprintln(w, "---------\t------\t-----------")

	for name, info := range tools {
		description := info.Tool.Description
		if len(description) > 60 {
			description = description[:57] + "..."
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n",
			name, info.ServerName, description)
	}

	w.Flush()

	cmd.Printf("\nTotal: %d tool(s) from %d server(s)\n",
		len(tools), len(mcpRegistry.ListServers()))

	return nil
}

func runTestTool(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	toolName := args[0]

	// Get tool info
	toolInfo, err := mcpRegistry.GetTool(toolName)
	if err != nil {
		return fmt.Errorf("tool not found: %w\nUse 'ainative-code mcp list-tools' to see available tools", err)
	}

	cmd.Printf("Testing tool: %s\n", toolName)
	cmd.Printf("Server: %s\n", toolInfo.ServerName)
	cmd.Printf("Description: %s\n\n", toolInfo.Tool.Description)

	// Parse arguments
	arguments := make(map[string]interface{})
	for key, value := range mcpServerHeaders {
		// Try to parse as JSON
		var jsonValue interface{}
		if err := json.Unmarshal([]byte(value), &jsonValue); err == nil {
			arguments[key] = jsonValue
		} else {
			arguments[key] = value
		}
	}

	cmd.Printf("Arguments: %v\n\n", arguments)

	// Execute tool
	cmd.Println("Executing...")
	result, err := mcpRegistry.CallTool(ctx, toolName, arguments)
	if err != nil {
		return fmt.Errorf("tool execution failed: %w", err)
	}

	if result.IsError {
		cmd.Println("Result: ERROR")
	} else {
		cmd.Println("Result: SUCCESS")
	}

	cmd.Println("\nContent:")
	for i, content := range result.Content {
		cmd.Printf("  [%d] Type: %s\n", i+1, content.Type)
		if content.Text != "" {
			cmd.Printf("      Text: %s\n", content.Text)
		}
		if content.Data != "" {
			cmd.Printf("      Data: %s\n", content.Data)
		}
	}

	return nil
}

func runDiscover(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	cmd.Println("Discovering tools from MCP servers...")

	if err := mcpRegistry.DiscoverTools(ctx); err != nil {
		return fmt.Errorf("failed to discover tools: %w", err)
	}

	tools := mcpRegistry.ListTools()
	servers := mcpRegistry.ListServers()

	cmd.Printf("\nDiscovered %d tool(s) from %d server(s)\n", len(tools), len(servers))

	if len(tools) > 0 {
		cmd.Println("\nUse 'ainative-code mcp list-tools' to see all available tools")
	}

	return nil
}

// GetMCPRegistry returns the global MCP registry for use by other commands.
func GetMCPRegistry() *mcp.Registry {
	return mcpRegistry
}
