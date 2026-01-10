# GitHub Issue #116 - Setup Wizard Visual Flow

## Interactive Setup Wizard Flow Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                  AINative Code Setup Wizard                  │
│                   ainative-code setup                        │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│  Step 1: Welcome Screen                                     │
│  ┌────────────────────────────────────────────────────┐    │
│  │ Welcome to AINative Code!                          │    │
│  │ Let's set up your AI-powered development          │    │
│  │ environment.                                       │    │
│  └────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│  Step 2: LLM Provider Selection                             │
│  ┌────────────────────────────────────────────────────┐    │
│  │ Which LLM provider would you like to use?         │    │
│  │                                                    │    │
│  │  > Anthropic (Claude)                             │    │
│  │    OpenAI (GPT)                                   │    │
│  │    Google (Gemini)                                │    │
│  │    Meta (Llama)                                   │    │
│  │    Ollama (Local)                                 │    │
│  └────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│  Step 3: API Key Configuration                              │
│  ┌────────────────────────────────────────────────────┐    │
│  │ Enter your Anthropic API key:                     │    │
│  │ _____________________________________________     │    │
│  │                                                    │    │
│  │ Get your API key from:                            │    │
│  │ https://console.anthropic.com/                    │    │
│  └────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│  Step 4: Model Selection                                    │
│  ┌────────────────────────────────────────────────────┐    │
│  │ Which Claude model would you like to use?         │    │
│  │                                                    │    │
│  │  > claude-3-5-sonnet-20241022 (Recommended)      │    │
│  │    claude-3-opus-20240229                         │    │
│  │    claude-3-sonnet-20240229                       │    │
│  │    claude-3-haiku-20240307                        │    │
│  └────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│  Step 5: Extended Thinking                                  │
│  ┌────────────────────────────────────────────────────┐    │
│  │ Enable extended thinking mode for complex         │    │
│  │ reasoning?                                         │    │
│  │                                                    │    │
│  │ [yes / no]                                        │    │
│  └────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│  Step 6: AINative Platform (Optional)                       │
│  ┌────────────────────────────────────────────────────┐    │
│  │ Would you like to connect to the AINative        │    │
│  │ platform?                                         │    │
│  │ (Optional - enables advanced features like       │    │
│  │  ZeroDB, Design tools, etc.)                     │    │
│  │                                                    │    │
│  │ [yes / no]                                        │    │
│  └────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
                              │
                ┌─────────────┴─────────────┐
                │ If yes                     │ If no
                ▼                            ▼
┌──────────────────────────┐    ┌──────────────────────────┐
│ Enter AINative API Key   │    │ Skip to Step 7           │
└──────────────────────────┘    └──────────────────────────┘
                │                            │
                └─────────────┬─────────────┘
                              ▼
┌─────────────────────────────────────────────────────────────┐
│  ✨ Step 7: Strapi CMS Integration (Optional) ✨           │
│  ┌────────────────────────────────────────────────────┐    │
│  │ Would you like to configure Strapi CMS?           │    │
│  │ (Optional - enables headless CMS features)        │    │
│  │                                                    │    │
│  │ [yes / no]                                        │    │
│  └────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
                              │
                ┌─────────────┴─────────────┐
                │ If yes                     │ If no
                ▼                            ▼
┌──────────────────────────┐    ┌──────────────────────────┐
│ Strapi Configuration     │    │ Skip to Step 8           │
│ ┌──────────────────────┐ │    └──────────────────────────┘
│ │ Enter Strapi URL:    │ │                  │
│ │ __________________   │ │                  │
│ │                      │ │                  │
│ │ Example:             │ │                  │
│ │ https://your-strapi  │ │                  │
│ │ -instance.com        │ │                  │
│ └──────────────────────┘ │                  │
│                          │                  │
│ ┌──────────────────────┐ │                  │
│ │ Enter Strapi API Key │ │                  │
│ │ (optional):          │ │                  │
│ │ __________________   │ │                  │
│ └──────────────────────┘ │                  │
└──────────────────────────┘                  │
                │                             │
                └─────────────┬───────────────┘
                              ▼
┌─────────────────────────────────────────────────────────────┐
│  ✨ Step 8: ZeroDB Integration (Optional) ✨               │
│  ┌────────────────────────────────────────────────────┐    │
│  │ Would you like to configure ZeroDB?                │    │
│  │ (Optional - enables vector storage, quantum        │    │
│  │  operations, and memory)                           │    │
│  │                                                    │    │
│  │ [yes / no]                                        │    │
│  └────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
                              │
                ┌─────────────┴─────────────┐
                │ If yes                     │ If no
                ▼                            ▼
┌──────────────────────────┐    ┌──────────────────────────┐
│ ZeroDB Configuration     │    │ Skip to Step 9           │
│ ┌──────────────────────┐ │    └──────────────────────────┘
│ │ Enter ZeroDB         │ │                  │
│ │ Project ID:          │ │                  │
│ │ __________________   │ │                  │
│ │                      │ │                  │
│ │ Get your Project ID  │ │                  │
│ │ from ZeroDB          │ │                  │
│ │ dashboard            │ │                  │
│ └──────────────────────┘ │                  │
│                          │                  │
│ ┌──────────────────────┐ │                  │
│ │ Enter ZeroDB         │ │                  │
│ │ endpoint URL         │ │                  │
│ │ (optional):          │ │                  │
│ │ __________________   │ │                  │
│ │                      │ │                  │
│ │ Leave empty to use   │ │                  │
│ │ default endpoint     │ │                  │
│ └──────────────────────┘ │                  │
└──────────────────────────┘                  │
                │                             │
                └─────────────┬───────────────┘
                              ▼
┌─────────────────────────────────────────────────────────────┐
│  Step 9: Color Scheme                                       │
│  ┌────────────────────────────────────────────────────┐    │
│  │ Choose your preferred color scheme:                │    │
│  │                                                    │    │
│  │  > Auto (Match terminal)                          │    │
│  │    Light                                          │    │
│  │    Dark                                           │    │
│  └────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│  Step 10: Prompt Caching                                    │
│  ┌────────────────────────────────────────────────────┐    │
│  │ Enable prompt caching for faster responses?        │    │
│  │ (Caches common prompts to reduce latency and      │    │
│  │  costs)                                            │    │
│  │                                                    │    │
│  │ [yes / no]                                        │    │
│  └────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│  Step 11: Configuration Summary                             │
│  ┌────────────────────────────────────────────────────┐    │
│  │ Configuration Summary                              │    │
│  │                                                    │    │
│  │ LLM Provider: anthropic                           │    │
│  │ Model: claude-3-5-sonnet-20241022                 │    │
│  │ Extended Thinking: true                           │    │
│  │ AINative Platform: true                           │    │
│  │ Strapi CMS: true                                  │    │
│  │   URL: https://strapi.example.com                 │    │
│  │ ZeroDB: true                                      │    │
│  │   Project ID: my-project-123                      │    │
│  │   Endpoint: https://zerodb.example.com            │    │
│  │                                                    │    │
│  │ Confirm and save this configuration? [y/n]:       │    │
│  └────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│  Step 12: Validation (if not skipped)                       │
│  ┌────────────────────────────────────────────────────┐    │
│  │ Validating configuration...                        │    │
│  │                                                    │    │
│  │ ✓ API key validation successful                   │    │
│  │ ✓ Provider connection verified                    │    │
│  │ ✓ Configuration validated                         │    │
│  └────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│  Step 13: Save Configuration                                │
│  ┌────────────────────────────────────────────────────┐    │
│  │ Writing configuration file...                      │    │
│  │                                                    │    │
│  │ ✓ Configuration saved to:                         │    │
│  │   ~/.ainative-code.yaml                           │    │
│  │                                                    │    │
│  │ ✓ Initialization marker created:                  │    │
│  │   ~/.ainative-code-initialized                    │    │
│  └────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│  ✅ Setup Complete!                                         │
│  ┌────────────────────────────────────────────────────┐    │
│  │ Your configuration has been saved to:              │    │
│  │   ~/.ainative-code.yaml                           │    │
│  │                                                    │    │
│  │ Next steps:                                       │    │
│  │   1. Start a chat session:                        │    │
│  │      ainative-code chat                           │    │
│  │   2. View configuration:                          │    │
│  │      ainative-code config show                    │    │
│  │   3. Check version:                               │    │
│  │      ainative-code version                        │    │
│  │                                                    │    │
│  │ For help, run: ainative-code --help              │    │
│  └────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
```

## Configuration File Structure

```yaml
# ~/.ainative-code.yaml

app:
  name: ainative-code
  version: 0.1.0
  environment: development
  debug: false

llm:
  default_provider: anthropic
  anthropic:
    api_key: "sk-ant-api03-..."
    model: "claude-3-5-sonnet-20241022"
    max_tokens: 4096
    temperature: 0.7
    extended_thinking:
      enabled: true
      auto_expand: false
      max_depth: 5

platform:
  authentication:
    method: api_key
    api_key: "your-ainative-api-key"
    timeout: 10s

services:
  # ✨ Optional Strapi Configuration ✨
  strapi:
    enabled: true
    endpoint: "https://strapi.example.com"
    api_key: "your-strapi-api-key"
    timeout: 30s
    retry_attempts: 3

  # ✨ Optional ZeroDB Configuration ✨
  zerodb:
    enabled: true
    project_id: "my-project-123"
    endpoint: "https://zerodb.example.com"
    database: "default"
    ssl: true
    ssl_mode: "require"
    max_connections: 10
    idle_connections: 5
    conn_max_lifetime: 1h
    timeout: 30s
    retry_attempts: 3
    retry_delay: 1s

performance:
  cache:
    enabled: false
    type: memory
    ttl: 1h
    max_size: 100
  concurrency:
    max_workers: 10
    max_queue_size: 100
    worker_timeout: 5m

logging:
  level: info
  format: json
  output: stdout
  max_size: 100
  max_backups: 3
  max_age: 7
  compress: true

security:
  encrypt_config: false
  tls_enabled: false
```

## Key Features

### 1. Optional Configuration
Both Strapi and ZeroDB configurations are **completely optional**:
- Users can skip by answering "no"
- Setup completes successfully without them
- Can be added later by re-running `ainative-code setup --force`

### 2. Validation
- API keys are validated unless `--skip-validation` flag is used
- Configuration structure is validated before saving
- Proper error messages for invalid inputs

### 3. Non-Interactive Mode
```bash
# Set environment variables
export ANTHROPIC_API_KEY="sk-ant-..."
export ZERODB_PROJECT_ID="project-123"
export STRAPI_URL="https://strapi.example.com"

# Run in non-interactive mode
ainative-code setup --non-interactive
```

### 4. Force Re-run
```bash
# Re-run setup to update configuration
ainative-code setup --force
```

## Testing

### Unit Tests
- `internal/setup/wizard_test.go` - Wizard logic tests
- `internal/setup/prompts_test.go` - Prompt UI tests
- `internal/config/validator_test.go` - Config validation tests

### Integration Tests (Existing)
- `internal/setup/strapi_integration_test.go` - Strapi API tests
- `internal/setup/zerodb_integration_test.go` - ZeroDB config tests

### Production Integration Tests (NEW)
- `tests/integration/zerodb_production_api_test.go` - Real API tests
- Uses PRODUCTION credentials from .env
- Makes REAL HTTP requests to https://api.ainative.studio
- NO MOCK DATA

### Test Execution
```bash
# Run all setup tests
go test ./internal/setup/... -v

# Run production integration tests
./test_zerodb_production.sh
```

## Command Line Interface

```bash
# Interactive setup (default)
ainative-code setup

# Skip validation for faster setup
ainative-code setup --skip-validation

# Force re-run (overwrites existing config)
ainative-code setup --force

# Custom config path
ainative-code setup --config ~/my-config.yaml

# Non-interactive mode
ainative-code setup --non-interactive

# Get help
ainative-code setup --help
```

## Related Files

### Core Implementation
- `/Users/aideveloper/AINative-Code/internal/setup/wizard.go`
- `/Users/aideveloper/AINative-Code/internal/setup/prompts.go`
- `/Users/aideveloper/AINative-Code/internal/config/types.go`
- `/Users/aideveloper/AINative-Code/internal/cmd/setup.go`

### Tests
- `/Users/aideveloper/AINative-Code/tests/integration/zerodb_production_api_test.go`
- `/Users/aideveloper/AINative-Code/internal/setup/strapi_integration_test.go`
- `/Users/aideveloper/AINative-Code/internal/setup/zerodb_integration_test.go`

### Documentation
- `/Users/aideveloper/AINative-Code/ISSUE_116_COMPREHENSIVE_REPORT.md`
- `/Users/aideveloper/AINative-Code/ISSUE_116_EXECUTIVE_SUMMARY.md`
- `/Users/aideveloper/AINative-Code/ISSUE_116_VISUAL_FLOW.md` (this file)

---

✅ **GitHub Issue #116 - COMPLETE**

The setup wizard properly implements optional Strapi and ZeroDB configuration fields, and comprehensive production integration tests verify the implementation works with real APIs.
