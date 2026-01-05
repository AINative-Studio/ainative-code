package lsp

import (
	"time"
)

// LanguageServerConfig represents configuration for a language server
type LanguageServerConfig struct {
	// Language identifier (e.g., "go", "python", "typescript")
	Language string

	// Command to start the language server
	Command string

	// Arguments to pass to the language server
	Args []string

	// Environment variables for the language server process
	Env map[string]string

	// InitializationOptions sent during initialization
	InitializationOptions interface{}

	// Timeout for initialization
	InitTimeout time.Duration

	// Timeout for requests
	RequestTimeout time.Duration

	// Enable/disable specific capabilities
	EnableCompletion  bool
	EnableHover       bool
	EnableDefinition  bool
	EnableReferences  bool

	// Auto-restart on failure
	AutoRestart bool

	// Maximum restart attempts
	MaxRestarts int

	// Health check interval
	HealthCheckInterval time.Duration
}

// DefaultConfig returns default configuration for a language
func DefaultConfig(language string) *LanguageServerConfig {
	config := &LanguageServerConfig{
		Language:            language,
		InitTimeout:         30 * time.Second,
		RequestTimeout:      10 * time.Second,
		EnableCompletion:    true,
		EnableHover:         true,
		EnableDefinition:    true,
		EnableReferences:    true,
		AutoRestart:         true,
		MaxRestarts:         3,
		HealthCheckInterval: 60 * time.Second,
		Env:                 make(map[string]string),
	}

	// Set language-specific defaults
	switch language {
	case "go":
		config.Command = "gopls"
		config.Args = []string{"serve"}
		config.InitializationOptions = map[string]interface{}{
			"usePlaceholders": true,
			"completeUnimported": true,
		}
	case "python":
		config.Command = "pylsp"
		config.Args = []string{}
		config.InitializationOptions = map[string]interface{}{
			"pylsp": map[string]interface{}{
				"plugins": map[string]interface{}{
					"pycodestyle": map[string]interface{}{"enabled": false},
					"pylint":      map[string]interface{}{"enabled": false},
				},
			},
		}
	case "typescript":
		config.Command = "typescript-language-server"
		config.Args = []string{"--stdio"}
	case "javascript":
		config.Command = "typescript-language-server"
		config.Args = []string{"--stdio"}
	case "rust":
		config.Command = "rust-analyzer"
		config.Args = []string{}
	case "java":
		config.Command = "jdtls"
		config.Args = []string{}
	case "cpp":
		config.Command = "clangd"
		config.Args = []string{}
	case "c":
		config.Command = "clangd"
		config.Args = []string{}
	}

	return config
}

// Validate checks if the configuration is valid
func (c *LanguageServerConfig) Validate() error {
	if c.Language == "" {
		return &JSONRPCError{
			Code:    InvalidParams,
			Message: "language identifier is required",
		}
	}

	if c.Command == "" {
		return &JSONRPCError{
			Code:    InvalidParams,
			Message: "language server command is required",
		}
	}

	if c.InitTimeout <= 0 {
		return &JSONRPCError{
			Code:    InvalidParams,
			Message: "initialization timeout must be positive",
		}
	}

	if c.RequestTimeout <= 0 {
		return &JSONRPCError{
			Code:    InvalidParams,
			Message: "request timeout must be positive",
		}
	}

	if c.MaxRestarts < 0 {
		return &JSONRPCError{
			Code:    InvalidParams,
			Message: "max restarts cannot be negative",
		}
	}

	return nil
}

// Clone creates a deep copy of the configuration
func (c *LanguageServerConfig) Clone() *LanguageServerConfig {
	clone := *c

	// Deep copy slices and maps
	if c.Args != nil {
		clone.Args = make([]string, len(c.Args))
		copy(clone.Args, c.Args)
	}

	if c.Env != nil {
		clone.Env = make(map[string]string, len(c.Env))
		for k, v := range c.Env {
			clone.Env[k] = v
		}
	}

	return &clone
}

// Merge merges another configuration into this one (other takes precedence)
func (c *LanguageServerConfig) Merge(other *LanguageServerConfig) {
	if other.Language != "" {
		c.Language = other.Language
	}
	if other.Command != "" {
		c.Command = other.Command
	}
	if len(other.Args) > 0 {
		c.Args = other.Args
	}
	if len(other.Env) > 0 {
		if c.Env == nil {
			c.Env = make(map[string]string)
		}
		for k, v := range other.Env {
			c.Env[k] = v
		}
	}
	if other.InitializationOptions != nil {
		c.InitializationOptions = other.InitializationOptions
	}
	if other.InitTimeout > 0 {
		c.InitTimeout = other.InitTimeout
	}
	if other.RequestTimeout > 0 {
		c.RequestTimeout = other.RequestTimeout
	}
	if other.HealthCheckInterval > 0 {
		c.HealthCheckInterval = other.HealthCheckInterval
	}
}

// SupportedLanguages returns a list of languages with default configurations
func SupportedLanguages() []string {
	return []string{
		"go",
		"python",
		"typescript",
		"javascript",
		"rust",
		"java",
		"cpp",
		"c",
	}
}
