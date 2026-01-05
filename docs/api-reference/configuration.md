# Configuration API Reference

**Import Path**: `github.com/AINative-studio/ainative-code/internal/config`

The config package provides configuration management with YAML file loading, environment variable overrides, and validation.

## Table of Contents

- [Config Structure](#config-structure)
- [Config Loading](#config-loading)
- [Environment Variables](#environment-variables)
- [Default Values](#default-values)
- [Validation](#validation)
- [Usage Examples](#usage-examples)

## Config Structure

### Config

```go
type Config struct {
    App         AppConfig         `mapstructure:"app" yaml:"app"`
    LLM         LLMConfig         `mapstructure:"llm" yaml:"llm"`
    Platform    PlatformConfig    `mapstructure:"platform" yaml:"platform"`
    Services    ServicesConfig    `mapstructure:"services" yaml:"services"`
    Tools       ToolsConfig       `mapstructure:"tools" yaml:"tools"`
    Performance PerformanceConfig `mapstructure:"performance" yaml:"performance"`
    Logging     LoggingConfig     `mapstructure:"logging" yaml:"logging"`
    Security    SecurityConfig    `mapstructure:"security" yaml:"security"`
}
```

See [internal/config/types.go](/Users/aideveloper/AINative-Code/internal/config/types.go) for complete type definitions.

### Key Configuration Sections

#### AppConfig

```go
type AppConfig struct {
    Name        string `mapstructure:"name" yaml:"name"`
    Version     string `mapstructure:"version" yaml:"version"`
    Environment string `mapstructure:"environment" yaml:"environment"`
    Debug       bool   `mapstructure:"debug" yaml:"debug"`
}
```

#### LLMConfig

```go
type LLMConfig struct {
    DefaultProvider string            `mapstructure:"default_provider" yaml:"default_provider"`
    Anthropic       *AnthropicConfig  `mapstructure:"anthropic,omitempty" yaml:"anthropic,omitempty"`
    OpenAI          *OpenAIConfig     `mapstructure:"openai,omitempty" yaml:"openai,omitempty"`
    Google          *GoogleConfig     `mapstructure:"google,omitempty" yaml:"google,omitempty"`
    Bedrock         *BedrockConfig    `mapstructure:"bedrock,omitempty" yaml:"bedrock,omitempty"`
    Azure           *AzureConfig      `mapstructure:"azure,omitempty" yaml:"azure,omitempty"`
    Ollama          *OllamaConfig     `mapstructure:"ollama,omitempty" yaml:"ollama,omitempty"`
    Fallback        *FallbackConfig   `mapstructure:"fallback,omitempty" yaml:"fallback,omitempty"`
}
```

## Config Loading

### Loader

```go
type Loader struct {
    // Contains filtered or unexported fields
}
```

#### NewLoader

```go
func NewLoader(opts ...LoaderOption) *Loader
```

Creates a new configuration loader with options.

**Options**:
- `WithConfigPaths(...string)` - Set custom config file search paths
- `WithConfigName(string)` - Set config file name (without extension)
- `WithConfigType(string)` - Set config file type (yaml, json, toml)
- `WithEnvPrefix(string)` - Set environment variable prefix
- `WithResolver(*Resolver)` - Set custom API key resolver

**Example**:

```go
loader := config.NewLoader(
    config.WithConfigPaths(".", "./configs", "$HOME/.ainative"),
    config.WithConfigName("config"),
    config.WithConfigType("yaml"),
    config.WithEnvPrefix("AINATIVE"),
)
```

#### Load

```go
func (l *Loader) Load() (*Config, error)
```

Loads configuration from all sources (file, environment, defaults).

**Search Order**:
1. Default values (hardcoded)
2. Configuration file (if found)
3. Environment variables (override)

**Example**:

```go
package main

import (
    "log"
    "github.com/AINative-studio/ainative-code/internal/config"
)

func main() {
    loader := config.NewLoader()
    cfg, err := loader.Load()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    log.Printf("Using provider: %s", cfg.LLM.DefaultProvider)
}
```

#### LoadFromFile

```go
func (l *Loader) LoadFromFile(filePath string) (*Config, error)
```

Loads configuration from a specific file path.

**Example**:

```go
cfg, err := loader.LoadFromFile("./configs/production.yaml")
if err != nil {
    log.Fatalf("Failed to load config: %v", err)
}
```

## Environment Variables

Environment variables override configuration file values using the pattern:

```
AINATIVE_<SECTION>_<SUBSECTION>_<KEY>
```

### Common Environment Variables

#### LLM Providers

```bash
# Anthropic
export AINATIVE_LLM_ANTHROPIC_API_KEY="sk-ant-..."
export AINATIVE_LLM_ANTHROPIC_MODEL="claude-3-5-sonnet-20241022"
export AINATIVE_LLM_ANTHROPIC_MAX_TOKENS=4096

# OpenAI
export AINATIVE_LLM_OPENAI_API_KEY="sk-..."
export AINATIVE_LLM_OPENAI_MODEL="gpt-4-turbo-preview"
export AINATIVE_LLM_OPENAI_ORGANIZATION="org-..."

# Google
export AINATIVE_LLM_GOOGLE_API_KEY="..."
export AINATIVE_LLM_GOOGLE_PROJECT_ID="my-project"

# Bedrock
export AINATIVE_LLM_BEDROCK_REGION="us-east-1"
export AINATIVE_LLM_BEDROCK_ACCESS_KEY_ID="..."
export AINATIVE_LLM_BEDROCK_SECRET_ACCESS_KEY="..."

# Azure
export AINATIVE_LLM_AZURE_API_KEY="..."
export AINATIVE_LLM_AZURE_ENDPOINT="https://..."
export AINATIVE_LLM_AZURE_DEPLOYMENT_NAME="gpt-4"

# Ollama
export AINATIVE_LLM_OLLAMA_BASE_URL="http://localhost:11434"
export AINATIVE_LLM_OLLAMA_MODEL="llama2"
```

#### Platform Authentication

```bash
export AINATIVE_PLATFORM_AUTHENTICATION_API_KEY="..."
export AINATIVE_PLATFORM_AUTHENTICATION_TOKEN="..."
export AINATIVE_PLATFORM_AUTHENTICATION_CLIENT_ID="..."
export AINATIVE_PLATFORM_AUTHENTICATION_CLIENT_SECRET="..."
```

#### Services

```bash
export AINATIVE_SERVICES_ZERODB_ENDPOINT="..."
export AINATIVE_SERVICES_ZERODB_USERNAME="..."
export AINATIVE_SERVICES_ZERODB_PASSWORD="..."

export AINATIVE_SERVICES_DESIGN_API_KEY="..."
export AINATIVE_SERVICES_STRAPI_API_KEY="..."
export AINATIVE_SERVICES_RLHF_API_KEY="..."
```

#### Application

```bash
export AINATIVE_APP_ENVIRONMENT="production"
export AINATIVE_APP_DEBUG="false"
export AINATIVE_LOGGING_LEVEL="info"
export AINATIVE_LOGGING_FORMAT="json"
```

## Default Values

The loader sets sensible defaults for all configuration values. Key defaults include:

### Application Defaults
- `app.name`: "ainative-code"
- `app.version`: "0.1.0"
- `app.environment`: "development"
- `app.debug`: false

### LLM Defaults
- `llm.default_provider`: "anthropic"
- `llm.anthropic.model`: "claude-3-5-sonnet-20241022"
- `llm.anthropic.max_tokens`: 4096
- `llm.anthropic.temperature`: 0.7
- `llm.anthropic.timeout`: 30s
- `llm.anthropic.retry_attempts`: 3

### Performance Defaults
- `performance.cache.enabled`: false
- `performance.rate_limit.enabled`: false
- `performance.concurrency.max_workers`: 10
- `performance.circuit_breaker.enabled`: false

### Logging Defaults
- `logging.level`: "info"
- `logging.format`: "json"
- `logging.output`: "stdout"
- `logging.max_size`: 100 (MB)
- `logging.max_backups`: 3

See loader.go `setDefaults()` method for complete list.

## Validation

### Validator

```go
type Validator struct {
    config *Config
}
```

#### NewValidator

```go
func NewValidator(cfg *Config) *Validator
```

#### Validate

```go
func (v *Validator) Validate() error
```

Validates the configuration and returns detailed errors.

**Validation Rules**:
- At least one LLM provider must be configured
- Required API keys for enabled providers
- Valid enum values for environment, log level, etc.
- Valid time durations and numeric ranges
- Valid paths for file-based configuration

**Example**:

```go
validator := config.NewValidator(cfg)
if err := validator.Validate(); err != nil {
    log.Fatalf("Invalid configuration: %v", err)
}
```

## Usage Examples

### Basic Configuration Loading

```go
package main

import (
    "fmt"
    "log"

    "github.com/AINative-studio/ainative-code/internal/config"
)

func main() {
    // Load configuration
    loader := config.NewLoader()
    cfg, err := loader.Load()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    // Use configuration
    fmt.Printf("App: %s v%s\n", cfg.App.Name, cfg.App.Version)
    fmt.Printf("Environment: %s\n", cfg.App.Environment)
    fmt.Printf("Provider: %s\n", cfg.LLM.DefaultProvider)

    if cfg.LLM.Anthropic != nil {
        fmt.Printf("Anthropic Model: %s\n", cfg.LLM.Anthropic.Model)
        fmt.Printf("Max Tokens: %d\n", cfg.LLM.Anthropic.MaxTokens)
    }
}
```

### Custom Config Paths

```go
// Load from custom locations
loader := config.NewLoader(
    config.WithConfigPaths(
        "/etc/ainative",
        "$HOME/.config/ainative",
        "./config",
    ),
)

cfg, err := loader.Load()
```

### Environment-Specific Config

```go
import "os"

// Determine environment
env := os.Getenv("ENV")
if env == "" {
    env = "development"
}

// Load environment-specific config
configFile := fmt.Sprintf("config.%s.yaml", env)
loader := config.NewLoader(config.WithConfigName(fmt.Sprintf("config.%s", env)))
cfg, err := loader.Load()
```

### Writing Configuration

```go
// Create configuration
cfg := &config.Config{
    App: config.AppConfig{
        Name:        "ainative-code",
        Version:     "1.0.0",
        Environment: "production",
    },
    LLM: config.LLMConfig{
        DefaultProvider: "anthropic",
        Anthropic: &config.AnthropicConfig{
            Model:     "claude-3-5-sonnet-20241022",
            MaxTokens: 4096,
            Temperature: 0.7,
        },
    },
}

// Write to file
if err := config.WriteConfig(cfg, "./config.yaml"); err != nil {
    log.Fatalf("Failed to write config: %v", err)
}
```

### Accessing Nested Configuration

```go
// Access provider config
if cfg.LLM.Anthropic != nil {
    fmt.Printf("API Key: %s...\n", cfg.LLM.Anthropic.APIKey[:10])
    fmt.Printf("Model: %s\n", cfg.LLM.Anthropic.Model)
    fmt.Printf("Timeout: %v\n", cfg.LLM.Anthropic.Timeout)
}

// Access tools config
if cfg.Tools.FileSystem != nil && cfg.Tools.FileSystem.Enabled {
    fmt.Printf("Allowed paths: %v\n", cfg.Tools.FileSystem.AllowedPaths)
    fmt.Printf("Max file size: %d bytes\n", cfg.Tools.FileSystem.MaxFileSize)
}

// Access performance config
fmt.Printf("Max workers: %d\n", cfg.Performance.Concurrency.MaxWorkers)
fmt.Printf("Cache enabled: %v\n", cfg.Performance.Cache.Enabled)
```

### Example YAML Configuration

```yaml
app:
  name: ainative-code
  version: 1.0.0
  environment: production
  debug: false

llm:
  default_provider: anthropic
  anthropic:
    api_key: "${ANTHROPIC_API_KEY}"
    model: claude-3-5-sonnet-20241022
    max_tokens: 4096
    temperature: 0.7
    top_p: 1.0
    timeout: 30s
    retry_attempts: 3

  fallback:
    enabled: true
    providers: [anthropic, openai]
    max_retries: 2
    retry_delay: 1s

platform:
  authentication:
    method: jwt
    client_id: ainative-code-cli
    timeout: 30s

services:
  zerodb:
    enabled: true
    endpoint: https://zerodb.ainative.studio
    database: ainative
    max_connections: 10
    timeout: 5s

tools:
  filesystem:
    enabled: true
    allowed_paths:
      - /tmp
      - ./workspace
    max_file_size: 104857600

  terminal:
    enabled: true
    timeout: 5m
    blocked_commands:
      - rm -rf /
      - dd

performance:
  cache:
    enabled: true
    type: memory
    ttl: 1h
    max_size: 100

  rate_limit:
    enabled: true
    requests_per_minute: 60
    burst_size: 10

  concurrency:
    max_workers: 10
    max_queue_size: 100

logging:
  level: info
  format: json
  output: stdout

security:
  encrypt_config: true
  tls_enabled: false
```

## Best Practices

### 1. Use Environment Variables for Secrets

```yaml
# config.yaml - Don't hardcode secrets
llm:
  anthropic:
    api_key: "${ANTHROPIC_API_KEY}"  # Use env var reference
    model: claude-3-5-sonnet-20241022
```

```bash
# Set in environment
export ANTHROPIC_API_KEY="sk-ant-..."
```

### 2. Separate Configs by Environment

```
configs/
├── config.development.yaml
├── config.staging.yaml
└── config.production.yaml
```

### 3. Validate After Loading

```go
cfg, err := loader.Load()
if err != nil {
    log.Fatal(err)
}

validator := config.NewValidator(cfg)
if err := validator.Validate(); err != nil {
    log.Fatalf("Invalid config: %v", err)
}
```

### 4. Document Configuration

```yaml
# config.yaml
llm:
  anthropic:
    # Temperature controls randomness (0.0 = deterministic, 1.0 = creative)
    temperature: 0.7

    # Maximum tokens in response (Claude 3.5 Sonnet supports up to 8192)
    max_tokens: 4096
```

### 5. Use Defaults Wisely

```go
// Only override what's necessary
loader := config.NewLoader()  // Uses sensible defaults
cfg, _ := loader.Load()

// Override specific values via env vars
os.Setenv("AINATIVE_LLM_ANTHROPIC_TEMPERATURE", "0.9")
```

## Related Documentation

- [Authentication](authentication.md) - Auth configuration
- [Providers](providers.md) - Provider configuration
- [Tools](tools.md) - Tool configuration
- [Core Packages](core-packages.md) - Using config in code
