package mcp

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigManager_LoadConfig(t *testing.T) {
	// Create temporary directory for test
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.mcp.json")

	// Test loading non-existent config
	cm := NewConfigManager(configPath)
	config, err := cm.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed for non-existent file: %v", err)
	}
	if config.MCPServers == nil {
		t.Fatal("MCPServers map should be initialized")
	}
	if len(config.MCPServers) != 0 {
		t.Fatalf("Expected empty config, got %d servers", len(config.MCPServers))
	}

	// Create a test config file
	testConfig := `{
  "mcpServers": {
    "test-server": {
      "url": "http://localhost:3000",
      "timeout": "30s",
      "enabled": true,
      "description": "Test server"
    }
  }
}`
	if err := os.WriteFile(configPath, []byte(testConfig), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Test loading existing config
	config, err = cm.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if len(config.MCPServers) != 1 {
		t.Fatalf("Expected 1 server, got %d", len(config.MCPServers))
	}
	server, exists := config.MCPServers["test-server"]
	if !exists {
		t.Fatal("Expected test-server to exist")
	}
	if server.URL != "http://localhost:3000" {
		t.Errorf("Expected URL http://localhost:3000, got %s", server.URL)
	}
	if server.Timeout != "30s" {
		t.Errorf("Expected timeout 30s, got %s", server.Timeout)
	}
}

func TestConfigManager_SaveConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.mcp.json")

	cm := NewConfigManager(configPath)

	// Create a config
	enabled := true
	config := &MCPConfig{
		MCPServers: map[string]ServerConfig{
			"test-server": {
				URL:         "http://localhost:3000",
				Timeout:     "30s",
				Enabled:     &enabled,
				Description: "Test server",
			},
		},
	}

	// Save config
	if err := cm.SaveConfig(config); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("Config file was not created")
	}

	// Load and verify
	loadedConfig, err := cm.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if len(loadedConfig.MCPServers) != 1 {
		t.Fatalf("Expected 1 server, got %d", len(loadedConfig.MCPServers))
	}
	server := loadedConfig.MCPServers["test-server"]
	if server.URL != "http://localhost:3000" {
		t.Errorf("Expected URL http://localhost:3000, got %s", server.URL)
	}
}

func TestConfigManager_AddServer(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.mcp.json")

	cm := NewConfigManager(configPath)

	// Add a server
	enabled := true
	serverConfig := ServerConfig{
		URL:         "http://localhost:3000",
		Timeout:     "30s",
		Enabled:     &enabled,
		Description: "Test server",
	}

	if err := cm.AddServer("test-server", serverConfig); err != nil {
		t.Fatalf("AddServer failed: %v", err)
	}

	// Verify server was added
	config, err := cm.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if len(config.MCPServers) != 1 {
		t.Fatalf("Expected 1 server, got %d", len(config.MCPServers))
	}
	server, exists := config.MCPServers["test-server"]
	if !exists {
		t.Fatal("Expected test-server to exist")
	}
	if server.URL != "http://localhost:3000" {
		t.Errorf("Expected URL http://localhost:3000, got %s", server.URL)
	}

	// Add another server
	serverConfig2 := ServerConfig{
		URL:         "http://localhost:4000",
		Timeout:     "60s",
		Enabled:     &enabled,
		Description: "Second test server",
	}

	if err := cm.AddServer("test-server-2", serverConfig2); err != nil {
		t.Fatalf("AddServer failed for second server: %v", err)
	}

	// Verify both servers exist
	config, err = cm.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if len(config.MCPServers) != 2 {
		t.Fatalf("Expected 2 servers, got %d", len(config.MCPServers))
	}
}

func TestConfigManager_RemoveServer(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.mcp.json")

	cm := NewConfigManager(configPath)

	// Add two servers
	enabled := true
	if err := cm.AddServer("server1", ServerConfig{URL: "http://localhost:3000", Enabled: &enabled}); err != nil {
		t.Fatalf("AddServer failed: %v", err)
	}
	if err := cm.AddServer("server2", ServerConfig{URL: "http://localhost:4000", Enabled: &enabled}); err != nil {
		t.Fatalf("AddServer failed: %v", err)
	}

	// Remove first server
	if err := cm.RemoveServer("server1"); err != nil {
		t.Fatalf("RemoveServer failed: %v", err)
	}

	// Verify server was removed
	config, err := cm.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if len(config.MCPServers) != 1 {
		t.Fatalf("Expected 1 server, got %d", len(config.MCPServers))
	}
	if _, exists := config.MCPServers["server1"]; exists {
		t.Error("server1 should have been removed")
	}
	if _, exists := config.MCPServers["server2"]; !exists {
		t.Error("server2 should still exist")
	}

	// Try to remove non-existent server
	if err := cm.RemoveServer("nonexistent"); err == nil {
		t.Error("Expected error when removing non-existent server")
	}
}

func TestConfigManager_GetServer(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.mcp.json")

	cm := NewConfigManager(configPath)

	// Add a server
	enabled := true
	serverConfig := ServerConfig{
		URL:         "http://localhost:3000",
		Timeout:     "30s",
		Enabled:     &enabled,
		Description: "Test server",
	}
	if err := cm.AddServer("test-server", serverConfig); err != nil {
		t.Fatalf("AddServer failed: %v", err)
	}

	// Get the server
	retrieved, err := cm.GetServer("test-server")
	if err != nil {
		t.Fatalf("GetServer failed: %v", err)
	}
	if retrieved.URL != "http://localhost:3000" {
		t.Errorf("Expected URL http://localhost:3000, got %s", retrieved.URL)
	}

	// Try to get non-existent server
	_, err = cm.GetServer("nonexistent")
	if err == nil {
		t.Error("Expected error when getting non-existent server")
	}
}

func TestConfigManager_ListServers(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.mcp.json")

	cm := NewConfigManager(configPath)

	// List empty config
	servers, err := cm.ListServers()
	if err != nil {
		t.Fatalf("ListServers failed: %v", err)
	}
	if len(servers) != 0 {
		t.Fatalf("Expected 0 servers, got %d", len(servers))
	}

	// Add servers
	enabled := true
	if err := cm.AddServer("server1", ServerConfig{URL: "http://localhost:3000", Enabled: &enabled}); err != nil {
		t.Fatalf("AddServer failed: %v", err)
	}
	if err := cm.AddServer("server2", ServerConfig{URL: "http://localhost:4000", Enabled: &enabled}); err != nil {
		t.Fatalf("AddServer failed: %v", err)
	}

	// List servers
	servers, err = cm.ListServers()
	if err != nil {
		t.Fatalf("ListServers failed: %v", err)
	}
	if len(servers) != 2 {
		t.Fatalf("Expected 2 servers, got %d", len(servers))
	}

	// Verify server names
	hasServer1 := false
	hasServer2 := false
	for _, name := range servers {
		if name == "server1" {
			hasServer1 = true
		}
		if name == "server2" {
			hasServer2 = true
		}
	}
	if !hasServer1 || !hasServer2 {
		t.Error("Expected both server1 and server2 in the list")
	}
}

func TestConfigManager_AtomicWrite(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.mcp.json")

	cm := NewConfigManager(configPath)

	// Add a server
	enabled := true
	if err := cm.AddServer("server1", ServerConfig{URL: "http://localhost:3000", Enabled: &enabled}); err != nil {
		t.Fatalf("AddServer failed: %v", err)
	}

	// Verify temp file is cleaned up
	tempFile := configPath + ".tmp"
	if _, err := os.Stat(tempFile); !os.IsNotExist(err) {
		t.Error("Temporary file should have been cleaned up")
	}
}
