package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// MCPConfig represents the structure of .mcp.json file
type MCPConfig struct {
	MCPServers map[string]ServerConfig `json:"mcpServers"`
}

// ServerConfig represents a single server configuration in .mcp.json
// This matches the Claude Desktop format for MCP server configuration
type ServerConfig struct {
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	// For HTTP-based servers (our internal format)
	URL         string            `json:"url,omitempty"`
	Timeout     string            `json:"timeout,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	Enabled     *bool             `json:"enabled,omitempty"`
	Description string            `json:"description,omitempty"`
}

// ConfigManager handles reading and writing MCP configuration files
type ConfigManager struct {
	configPath string
	mu         sync.RWMutex
}

// NewConfigManager creates a new config manager
func NewConfigManager(configPath string) *ConfigManager {
	if configPath == "" {
		// Default to .mcp.json in user's home directory
		home, err := os.UserHomeDir()
		if err == nil {
			configPath = filepath.Join(home, ".mcp.json")
		} else {
			configPath = ".mcp.json"
		}
	}
	return &ConfigManager{
		configPath: configPath,
	}
}

// LoadConfig loads the MCP configuration from disk
func (cm *ConfigManager) LoadConfig() (*MCPConfig, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// Check if file exists
	if _, err := os.Stat(cm.configPath); os.IsNotExist(err) {
		// Return empty config if file doesn't exist
		return &MCPConfig{
			MCPServers: make(map[string]ServerConfig),
		}, nil
	}

	// Read file
	data, err := os.ReadFile(cm.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse JSON
	var config MCPConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Initialize map if nil
	if config.MCPServers == nil {
		config.MCPServers = make(map[string]ServerConfig)
	}

	return &config, nil
}

// SaveConfig saves the MCP configuration to disk
func (cm *ConfigManager) SaveConfig(config *MCPConfig) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Ensure directory exists
	dir := filepath.Dir(cm.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal to JSON with pretty printing
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to temporary file first for atomic operation
	tempFile := cm.configPath + ".tmp"
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp config file: %w", err)
	}

	// Rename temp file to actual config file (atomic on POSIX systems)
	if err := os.Rename(tempFile, cm.configPath); err != nil {
		os.Remove(tempFile) // Clean up temp file on error
		return fmt.Errorf("failed to save config file: %w", err)
	}

	return nil
}

// AddServer adds or updates a server in the configuration
func (cm *ConfigManager) AddServer(name string, serverConfig ServerConfig) error {
	// Load current config
	config, err := cm.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Add or update server
	config.MCPServers[name] = serverConfig

	// Save config
	if err := cm.SaveConfig(config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// RemoveServer removes a server from the configuration
func (cm *ConfigManager) RemoveServer(name string) error {
	// Load current config
	config, err := cm.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Check if server exists
	if _, exists := config.MCPServers[name]; !exists {
		return fmt.Errorf("server %s not found in configuration", name)
	}

	// Remove server
	delete(config.MCPServers, name)

	// Save config
	if err := cm.SaveConfig(config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// GetServer retrieves a server configuration by name
func (cm *ConfigManager) GetServer(name string) (*ServerConfig, error) {
	config, err := cm.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	serverConfig, exists := config.MCPServers[name]
	if !exists {
		return nil, fmt.Errorf("server %s not found", name)
	}

	return &serverConfig, nil
}

// ListServers returns all server names in the configuration
func (cm *ConfigManager) ListServers() ([]string, error) {
	config, err := cm.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	names := make([]string, 0, len(config.MCPServers))
	for name := range config.MCPServers {
		names = append(names, name)
	}

	return names, nil
}

// GetConfigPath returns the path to the configuration file
func (cm *ConfigManager) GetConfigPath() string {
	return cm.configPath
}
