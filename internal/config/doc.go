// Package config provides comprehensive configuration management for AINative Code.
//
// The config package supports loading configuration from multiple sources with
// proper precedence, validation, and type safety.
//
// # Configuration Sources
//
// Configuration is loaded from the following sources in order of precedence:
//   1. Command-line flags (highest priority)
//   2. Environment variables
//   3. Configuration file
//   4. Default values (lowest priority)
//
// # Basic Usage
//
//	loader := config.NewLoader()
//	cfg, err := loader.Load()
//	if err != nil {
//	    log.Fatalf("Failed to load config: %v", err)
//	}
//
// # Loading from Specific File
//
//	loader := config.NewLoader()
//	cfg, err := loader.LoadFromFile("/path/to/config.yaml")
//	if err != nil {
//	    log.Fatalf("Failed to load config: %v", err)
//	}
//
// # Custom Loader Options
//
//	loader := config.NewLoader(
//	    config.WithConfigName("myconfig"),
//	    config.WithConfigType("yaml"),
//	    config.WithEnvPrefix("MYAPP"),
//	)
//
// # Environment Variables
//
// All configuration values can be set via environment variables using the
// configured prefix (default: AINATIVE_). Nested keys use underscores:
//
//	export AINATIVE_APP_ENVIRONMENT=production
//	export AINATIVE_LLM_DEFAULT_PROVIDER=anthropic
//	export AINATIVE_LLM_ANTHROPIC_API_KEY=sk-ant-...
//
// # Validation
//
// All loaded configurations are automatically validated. The validator checks:
//   - Required fields are present
//   - Values are within valid ranges
//   - URLs and paths are properly formatted
//   - Dependencies between fields are consistent
//
// Example validation error:
//
//	Configuration validation failed:
//	  - llm.anthropic.api_key: Anthropic API key is required
//	  - services.zerodb.endpoint: endpoint is required
//
// # Configuration Structure
//
// The configuration is organized into major sections:
//   - App: General application settings
//   - LLM: Language model provider configurations
//   - Platform: AINative platform settings
//   - Services: External service endpoints
//   - Tools: Tool-specific configurations
//   - Performance: Performance optimization settings
//   - Logging: Logging configuration
//   - Security: Security settings
//
// # LLM Providers
//
// Supported LLM providers:
//   - Anthropic Claude
//   - OpenAI
//   - Google Gemini
//   - AWS Bedrock
//   - Azure OpenAI
//   - Ollama (local models)
//
// Each provider has specific configuration requirements. See the documentation
// for details on configuring each provider.
//
// # Fallback Support
//
// The config package supports automatic fallback to alternative LLM providers:
//
//	llm:
//	  default_provider: anthropic
//	  fallback:
//	    enabled: true
//	    providers:
//	      - anthropic
//	      - openai
//	      - ollama
//
// # Security Best Practices
//
//   - Never commit secrets to configuration files
//   - Use environment variables for sensitive data
//   - Enable configuration encryption in production
//   - Use TLS for network communications
//   - Implement proper secret rotation
//
// # Example Configuration
//
// See examples/config.yaml for a complete configuration example with all
// available options and detailed comments.
//
// # Further Documentation
//
// For detailed configuration documentation, see docs/configuration.md
package config
