package mcp

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRegistry_PersistenceAddServer(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.mcp.json")
	os.Setenv("MCP_CONFIG_PATH", configPath)
	defer os.Unsetenv("MCP_CONFIG_PATH")

	// Create registry
	registry := NewRegistry(1 * time.Minute)

	// Add a server
	server := &Server{
		Name:        "test-server",
		URL:         "http://localhost:3000",
		Timeout:     30 * time.Second,
		Headers:     map[string]string{"X-Test": "value"},
		Enabled:     true,
		Description: "Test server",
	}

	if err := registry.AddServer(server); err != nil {
		t.Fatalf("AddServer failed: %v", err)
	}

	// Verify server is in memory
	servers := registry.ListServers()
	if len(servers) != 1 {
		t.Fatalf("Expected 1 server in memory, got %d", len(servers))
	}

	// Verify server is persisted to disk
	cm := NewConfigManager(configPath)
	config, err := cm.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if len(config.MCPServers) != 1 {
		t.Fatalf("Expected 1 server in config file, got %d", len(config.MCPServers))
	}

	serverConfig, exists := config.MCPServers["test-server"]
	if !exists {
		t.Fatal("test-server should exist in config file")
	}
	if serverConfig.URL != "http://localhost:3000" {
		t.Errorf("Expected URL http://localhost:3000, got %s", serverConfig.URL)
	}
	if serverConfig.Timeout != "30s" {
		t.Errorf("Expected timeout 30s, got %s", serverConfig.Timeout)
	}
	if serverConfig.Description != "Test server" {
		t.Errorf("Expected description 'Test server', got %s", serverConfig.Description)
	}
}

func TestRegistry_PersistenceRemoveServer(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.mcp.json")
	os.Setenv("MCP_CONFIG_PATH", configPath)
	defer os.Unsetenv("MCP_CONFIG_PATH")

	// Create registry and add servers
	registry := NewRegistry(1 * time.Minute)

	server1 := &Server{
		Name:    "server1",
		URL:     "http://localhost:3000",
		Timeout: 30 * time.Second,
		Enabled: true,
	}
	server2 := &Server{
		Name:    "server2",
		URL:     "http://localhost:4000",
		Timeout: 30 * time.Second,
		Enabled: true,
	}

	if err := registry.AddServer(server1); err != nil {
		t.Fatalf("AddServer failed: %v", err)
	}
	if err := registry.AddServer(server2); err != nil {
		t.Fatalf("AddServer failed: %v", err)
	}

	// Remove server1
	if err := registry.RemoveServer("server1"); err != nil {
		t.Fatalf("RemoveServer failed: %v", err)
	}

	// Verify server is removed from memory
	servers := registry.ListServers()
	if len(servers) != 1 {
		t.Fatalf("Expected 1 server in memory, got %d", len(servers))
	}

	// Verify server is removed from disk
	cm := NewConfigManager(configPath)
	config, err := cm.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if len(config.MCPServers) != 1 {
		t.Fatalf("Expected 1 server in config file, got %d", len(config.MCPServers))
	}

	if _, exists := config.MCPServers["server1"]; exists {
		t.Error("server1 should not exist in config file")
	}
	if _, exists := config.MCPServers["server2"]; !exists {
		t.Error("server2 should exist in config file")
	}
}

func TestRegistry_LoadServersFromConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.mcp.json")

	// Create a config file with servers
	testConfig := `{
  "mcpServers": {
    "server1": {
      "url": "http://localhost:3000",
      "timeout": "30s",
      "enabled": true,
      "description": "First server"
    },
    "server2": {
      "url": "http://localhost:4000",
      "timeout": "60s",
      "enabled": false,
      "description": "Second server"
    }
  }
}`
	if err := os.WriteFile(configPath, []byte(testConfig), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Set environment variable
	os.Setenv("MCP_CONFIG_PATH", configPath)
	defer os.Unsetenv("MCP_CONFIG_PATH")

	// Create new registry (should load from config)
	registry := NewRegistry(1 * time.Minute)

	// Verify servers were loaded
	servers := registry.ListServers()
	if len(servers) != 2 {
		t.Fatalf("Expected 2 servers, got %d", len(servers))
	}

	// Verify server1
	client1, err := registry.GetServer("server1")
	if err != nil {
		t.Fatalf("GetServer failed for server1: %v", err)
	}
	server1 := client1.GetServer()
	if server1.URL != "http://localhost:3000" {
		t.Errorf("Expected URL http://localhost:3000, got %s", server1.URL)
	}
	if server1.Timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %s", server1.Timeout)
	}
	if server1.Enabled != true {
		t.Errorf("Expected enabled true, got %v", server1.Enabled)
	}

	// Verify server2
	client2, err := registry.GetServer("server2")
	if err != nil {
		t.Fatalf("GetServer failed for server2: %v", err)
	}
	server2 := client2.GetServer()
	if server2.URL != "http://localhost:4000" {
		t.Errorf("Expected URL http://localhost:4000, got %s", server2.URL)
	}
	if server2.Timeout != 60*time.Second {
		t.Errorf("Expected timeout 60s, got %s", server2.Timeout)
	}
	if server2.Enabled != false {
		t.Errorf("Expected enabled false, got %v", server2.Enabled)
	}
}

func TestRegistry_PersistenceWorkflow(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.mcp.json")
	os.Setenv("MCP_CONFIG_PATH", configPath)
	defer os.Unsetenv("MCP_CONFIG_PATH")

	// Step 1: Create registry and add server
	registry1 := NewRegistry(1 * time.Minute)
	server := &Server{
		Name:    "test-server",
		URL:     "http://localhost:3000",
		Timeout: 30 * time.Second,
		Enabled: true,
	}
	if err := registry1.AddServer(server); err != nil {
		t.Fatalf("AddServer failed: %v", err)
	}

	// Step 2: List servers (should show 1)
	servers := registry1.ListServers()
	if len(servers) != 1 {
		t.Fatalf("Expected 1 server, got %d", len(servers))
	}

	// Step 3: Simulate restart by creating new registry
	registry2 := NewRegistry(1 * time.Minute)

	// Step 4: List servers (should still show 1 - loaded from disk)
	servers = registry2.ListServers()
	if len(servers) != 1 {
		t.Fatalf("Expected 1 server after restart, got %d", len(servers))
	}

	// Step 5: Remove server
	if err := registry2.RemoveServer("test-server"); err != nil {
		t.Fatalf("RemoveServer failed: %v", err)
	}

	// Step 6: List servers (should show 0)
	servers = registry2.ListServers()
	if len(servers) != 0 {
		t.Fatalf("Expected 0 servers after removal, got %d", len(servers))
	}

	// Step 7: Simulate restart again
	registry3 := NewRegistry(1 * time.Minute)

	// Step 8: List servers (should show 0 - removal persisted)
	servers = registry3.ListServers()
	if len(servers) != 0 {
		t.Fatalf("Expected 0 servers after restart, got %d", len(servers))
	}
}

func TestRegistry_RollbackOnPersistenceFailure(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "readonly", "test.mcp.json")
	os.Setenv("MCP_CONFIG_PATH", configPath)
	defer os.Unsetenv("MCP_CONFIG_PATH")

	// Create readonly directory to simulate persistence failure
	readonlyDir := filepath.Join(tmpDir, "readonly")
	if err := os.Mkdir(readonlyDir, 0555); err != nil {
		t.Fatalf("Failed to create readonly directory: %v", err)
	}
	defer os.Chmod(readonlyDir, 0755) // Restore permissions for cleanup

	// Create registry (should not fail even if config can't be read)
	registry := NewRegistry(1 * time.Minute)

	// Try to add server (should fail and rollback)
	server := &Server{
		Name:    "test-server",
		URL:     "http://localhost:3000",
		Timeout: 30 * time.Second,
		Enabled: true,
	}

	err := registry.AddServer(server)
	if err == nil {
		t.Fatal("Expected AddServer to fail due to readonly directory")
	}

	// Verify server was NOT added to memory (rollback occurred)
	servers := registry.ListServers()
	if len(servers) != 0 {
		t.Fatalf("Expected 0 servers after failed add (rollback), got %d", len(servers))
	}
}

func TestRegistry_RemoveServerNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.mcp.json")
	os.Setenv("MCP_CONFIG_PATH", configPath)
	defer os.Unsetenv("MCP_CONFIG_PATH")

	registry := NewRegistry(1 * time.Minute)

	// Try to remove non-existent server
	err := registry.RemoveServer("nonexistent")
	if err == nil {
		t.Error("Expected error when removing non-existent server")
	}
	if err.Error() != "server nonexistent not found" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestRegistry_AddServerDuplicate(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.mcp.json")
	os.Setenv("MCP_CONFIG_PATH", configPath)
	defer os.Unsetenv("MCP_CONFIG_PATH")

	registry := NewRegistry(1 * time.Minute)

	// Add server
	server := &Server{
		Name:    "test-server",
		URL:     "http://localhost:3000",
		Timeout: 30 * time.Second,
		Enabled: true,
	}
	if err := registry.AddServer(server); err != nil {
		t.Fatalf("AddServer failed: %v", err)
	}

	// Try to add same server again
	err := registry.AddServer(server)
	if err == nil {
		t.Error("Expected error when adding duplicate server")
	}
	if err.Error() != "server test-server already registered" {
		t.Errorf("Unexpected error message: %v", err)
	}

	// Verify only one server exists
	servers := registry.ListServers()
	if len(servers) != 1 {
		t.Fatalf("Expected 1 server, got %d", len(servers))
	}
}
