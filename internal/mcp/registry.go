package mcp

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Registry manages multiple MCP servers and their tools.
type Registry struct {
	mu            sync.RWMutex
	servers       map[string]*Client
	tools         map[string]*ToolInfo
	healthStatus  map[string]*HealthStatus
	checkInterval time.Duration
	stopChan      chan struct{}
	wg            sync.WaitGroup
}

// ToolInfo contains information about a tool and its source server.
type ToolInfo struct {
	Tool       Tool
	ServerName string
}

// NewRegistry creates a new MCP server registry.
func NewRegistry(checkInterval time.Duration) *Registry {
	if checkInterval == 0 {
		checkInterval = 1 * time.Minute
	}

	return &Registry{
		servers:       make(map[string]*Client),
		tools:         make(map[string]*ToolInfo),
		healthStatus:  make(map[string]*HealthStatus),
		checkInterval: checkInterval,
		stopChan:      make(chan struct{}),
	}
}

// AddServer adds an MCP server to the registry.
func (r *Registry) AddServer(server *Server) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.servers[server.Name]; exists {
		return fmt.Errorf("server %s already registered", server.Name)
	}

	client := NewClient(server)
	r.servers[server.Name] = client

	return nil
}

// RemoveServer removes an MCP server from the registry.
func (r *Registry) RemoveServer(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.servers[name]; !exists {
		return fmt.Errorf("server %s not found", name)
	}

	delete(r.servers, name)
	delete(r.healthStatus, name)

	// Remove tools from this server
	for toolName, toolInfo := range r.tools {
		if toolInfo.ServerName == name {
			delete(r.tools, toolName)
		}
	}

	return nil
}

// GetServer returns a server client by name.
func (r *Registry) GetServer(name string) (*Client, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	client, exists := r.servers[name]
	if !exists {
		return nil, fmt.Errorf("server %s not found", name)
	}

	return client, nil
}

// ListServers returns all registered server names.
func (r *Registry) ListServers() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.servers))
	for name := range r.servers {
		names = append(names, name)
	}

	return names
}

// DiscoverTools discovers all tools from all registered servers.
func (r *Registry) DiscoverTools(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Clear existing tools
	r.tools = make(map[string]*ToolInfo)

	// Discover tools from each server
	for name, client := range r.servers {
		if !client.server.Enabled {
			continue
		}

		tools, err := client.ListTools(ctx)
		if err != nil {
			return fmt.Errorf("failed to list tools from server %s: %w", name, err)
		}

		for _, tool := range tools {
			// Use server name prefix to avoid naming conflicts
			toolKey := fmt.Sprintf("%s.%s", name, tool.Name)
			r.tools[toolKey] = &ToolInfo{
				Tool:       tool,
				ServerName: name,
			}
		}
	}

	return nil
}

// GetTool returns a tool by its fully qualified name (server.tool).
func (r *Registry) GetTool(name string) (*ToolInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	toolInfo, exists := r.tools[name]
	if !exists {
		return nil, fmt.Errorf("tool %s not found", name)
	}

	return toolInfo, nil
}

// ListTools returns all discovered tools.
func (r *Registry) ListTools() map[string]*ToolInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return a copy to avoid concurrent modification
	tools := make(map[string]*ToolInfo, len(r.tools))
	for name, info := range r.tools {
		tools[name] = info
	}

	return tools
}

// CallTool executes a tool by its fully qualified name.
func (r *Registry) CallTool(ctx context.Context, name string, arguments map[string]interface{}) (*ToolResult, error) {
	// Get tool info
	toolInfo, err := r.GetTool(name)
	if err != nil {
		return nil, err
	}

	// Get server client
	client, err := r.GetServer(toolInfo.ServerName)
	if err != nil {
		return nil, err
	}

	// Execute tool
	return client.CallTool(ctx, toolInfo.Tool.Name, arguments)
}

// GetHealthStatus returns the health status of a server.
func (r *Registry) GetHealthStatus(name string) (*HealthStatus, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	status, exists := r.healthStatus[name]
	if !exists {
		return nil, fmt.Errorf("no health status for server %s", name)
	}

	return status, nil
}

// GetAllHealthStatus returns health status for all servers.
func (r *Registry) GetAllHealthStatus() map[string]*HealthStatus {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return a copy
	status := make(map[string]*HealthStatus, len(r.healthStatus))
	for name, s := range r.healthStatus {
		status[name] = s
	}

	return status
}

// StartHealthChecks starts periodic health checks for all servers.
func (r *Registry) StartHealthChecks(ctx context.Context) {
	r.wg.Add(1)
	go r.healthCheckLoop(ctx)
}

// StopHealthChecks stops the health check background process.
func (r *Registry) StopHealthChecks() {
	close(r.stopChan)
	r.wg.Wait()
}

// healthCheckLoop runs periodic health checks on all servers.
func (r *Registry) healthCheckLoop(ctx context.Context) {
	defer r.wg.Done()

	ticker := time.NewTicker(r.checkInterval)
	defer ticker.Stop()

	// Run initial health check
	r.performHealthChecks(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-r.stopChan:
			return
		case <-ticker.C:
			r.performHealthChecks(ctx)
		}
	}
}

// performHealthChecks checks health of all servers.
func (r *Registry) performHealthChecks(ctx context.Context) {
	r.mu.RLock()
	servers := make(map[string]*Client, len(r.servers))
	for name, client := range r.servers {
		servers[name] = client
	}
	r.mu.RUnlock()

	// Check each server
	var wg sync.WaitGroup
	statusChan := make(chan struct {
		name   string
		status *HealthStatus
	}, len(servers))

	for name, client := range servers {
		if !client.server.Enabled {
			continue
		}

		wg.Add(1)
		go func(n string, c *Client) {
			defer wg.Done()
			status := c.CheckHealth(ctx)
			statusChan <- struct {
				name   string
				status *HealthStatus
			}{n, status}
		}(name, client)
	}

	// Wait for all health checks to complete
	go func() {
		wg.Wait()
		close(statusChan)
	}()

	// Collect results
	for result := range statusChan {
		r.mu.Lock()
		r.healthStatus[result.name] = result.status
		r.mu.Unlock()
	}
}
