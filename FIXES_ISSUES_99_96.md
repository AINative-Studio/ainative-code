# Fixes for Issues #99 and #96

## Summary
Fixed critical issues where the chat command was not properly calling AI providers and failing with "AI provider not configured" errors even after successful setup.

## Issues Fixed

### Issue #99: Chat Command Not Calling AI
**Problem**: The chat command would print "Processing message" but never actually call the AI provider's Chat() or Stream() method.

**Root Cause**: The chat command in `/internal/cmd/chat.go` was properly implemented with provider initialization and API calls, but the provider initialization in `utils.go` was incomplete and didn't support all provider types.

**Fix**: Enhanced `initializeProvider()` function to properly support all provider types:
- OpenAI
- Anthropic
- Meta Llama
- Ollama
- Google/Gemini

### Issue #96: AI Provider Not Configured Error
**Problem**: Chat fails with "AI provider not configured" even after successful setup wizard.

**Root Cause**: The `getAPIKey()` function in `utils.go` was only checking for:
1. Provider-specific environment variables
2. Generic `api_key` viper config (flat structure)
3. Keychain

It was NOT checking the nested config structure that the setup wizard creates:
```yaml
llm:
  anthropic:
    api_key: "sk-..."
```

**Fix**: Updated `getAPIKey()` to check nested configuration paths in the correct order:
1. Provider-specific environment variables (e.g., `ANTHROPIC_API_KEY`)
2. Generic `AINATIVE_CODE_API_KEY` environment variable
3. **NEW**: Nested provider config (e.g., `llm.anthropic.api_key`)
4. Generic `api_key` field (backward compatibility)
5. System keychain

## Files Modified

### 1. `/internal/cmd/utils.go`

#### Changes to `getAPIKey()`:
```go
// Added support for nested config structure
var configKey string
switch providerName {
case "anthropic":
    configKey = "llm.anthropic.api_key"
case "openai":
    configKey = "llm.openai.api_key"
case "google":
    configKey = "llm.google.api_key"
case "meta_llama", "meta":
    configKey = "llm.meta_llama.api_key"
}

if configKey != "" {
    if key := viper.GetString(configKey); key != "" {
        // Use nested config key
        return key, nil
    }
}
```

#### Changes to `initializeProvider()`:
- Added imports for `ollama` and `gemini` providers
- Added cases for all supported providers:
  - `ollama`: Uses base URL from config, no API key required
  - `google`/`gemini`: Uses Google API key
  - Improved error messages to list all supported providers

### 2. New Test Files

#### `/internal/cmd/utils_integration_test.go`
Comprehensive unit tests covering:
- API key retrieval from different sources
- Provider initialization logic
- Configuration flow (nested and flat formats)
- Backward compatibility

All tests passing:
```
=== RUN   TestGetAPIKey
--- PASS: TestGetAPIKey (0.01s)
=== RUN   TestProviderConfigurationFlow
--- PASS: TestProviderConfigurationFlow (0.00s)
```

#### `/test-config.yaml`
Example configuration file demonstrating:
- Nested LLM configuration structure
- Flat configuration for backward compatibility
- All supported providers

#### `/test-chat-integration.sh`
Integration test script that verifies:
- Application builds successfully
- Config file loading works
- Provider initialization works
- Error handling for missing API keys
- All providers are recognized

## Configuration Format

The system now supports both configuration formats:

### Nested Format (Recommended - Created by Setup Wizard)
```yaml
llm:
  default_provider: anthropic
  anthropic:
    api_key: "sk-ant-..."
    model: "claude-3-5-sonnet-20241022"
    max_tokens: 4096
    temperature: 0.7
```

### Flat Format (Backward Compatible)
```yaml
provider: anthropic
model: claude-3-5-sonnet-20241022
api_key: "sk-ant-..."
```

## API Key Priority Order

The system now checks for API keys in this order:

1. **Provider-specific environment variable** (highest priority)
   - `ANTHROPIC_API_KEY`
   - `OPENAI_API_KEY`
   - `META_LLAMA_API_KEY`
   - `GOOGLE_API_KEY`

2. **Generic environment variable**
   - `AINATIVE_CODE_API_KEY`

3. **Nested configuration** (NEW - fixes issue #96)
   - `llm.anthropic.api_key`
   - `llm.openai.api_key`
   - `llm.google.api_key`
   - `llm.meta_llama.api_key`

4. **Generic config field** (backward compatibility)
   - `api_key`

5. **System keychain** (lowest priority)

## Supported Providers

After this fix, the following providers are fully supported:

1. **OpenAI** - GPT-4, GPT-3.5, etc.
2. **Anthropic** - Claude 3.5 Sonnet, Claude 3 Opus, etc.
3. **Meta Llama** - Meta's Llama models
4. **Ollama** - Local LLM hosting (no API key needed)
5. **Google/Gemini** - Google's Gemini models

## Testing

### Unit Tests
```bash
go test ./internal/cmd -v -run TestGetAPIKey
go test ./internal/cmd -v -run TestProviderConfigurationFlow
```

### Integration Tests
```bash
./test-chat-integration.sh
```

### Manual Testing with Real API
```bash
# Set your API key
export ANTHROPIC_API_KEY='your-key-here'

# Test single message
./bin/ainative-code chat "Hello, how are you?"

# Test with streaming
./bin/ainative-code chat --stream "Explain quantum computing"

# Test with different provider
export OPENAI_API_KEY='your-openai-key'
./bin/ainative-code --provider openai chat "Hello"
```

## Verification Steps

To verify these fixes work:

1. **Build the application**
   ```bash
   go build -o bin/ainative-code ./cmd/ainative-code
   ```

2. **Run setup wizard** (creates nested config)
   ```bash
   ./bin/ainative-code setup
   ```

3. **Test chat command**
   ```bash
   ./bin/ainative-code chat "Test message"
   ```

Expected behavior:
- ✓ Provider is initialized successfully
- ✓ API key is found from config or environment
- ✓ Chat request is sent to the AI provider
- ✓ Response is received and displayed
- ✗ No "Processing message" without actual API call
- ✗ No "AI provider not configured" errors

## Error Handling

Improved error messages:

### Before:
```
AI provider not configured
```

### After:
```
Error: no API key found for provider anthropic. Set ANTHROPIC_API_KEY or
AINATIVE_CODE_API_KEY environment variable, or run 'ainative-code setup'
```

### Unsupported Provider:
```
Error: unsupported provider: xyz. Supported providers: openai, anthropic,
meta_llama, ollama, google/gemini
```

## Backward Compatibility

All changes maintain backward compatibility:

1. Flat config format still works
2. Generic `api_key` field still checked
3. Existing environment variables still work
4. Priority order ensures no breaking changes

## Performance Impact

- Minimal: Only added config key lookups (negligible overhead)
- No changes to actual API call paths
- No additional dependencies

## Security Considerations

- API keys still checked in same security-conscious order
- Keychain integration unchanged
- No API keys logged (proper logging in place)
- Config file should still use 0600 permissions

## Future Improvements

Potential enhancements for future versions:

1. Add Azure OpenAI support to initializeProvider()
2. Add AWS Bedrock support to initializeProvider()
3. Cache provider instances for multiple calls
4. Add config validation on startup
5. Support multiple API keys for fallback/rotation

## Related Files

- `/internal/cmd/chat.go` - Chat command implementation (already correct)
- `/internal/cmd/root.go` - Root command with provider flags
- `/internal/cmd/config.go` - Config management commands
- `/internal/setup/wizard.go` - Setup wizard (creates nested config)
- `/internal/config/types.go` - Configuration type definitions
- `/internal/provider/*/` - Provider implementations

## References

- Issue #99: Chat command doesn't call AI
- Issue #96: Provider not configured error
- Pull Request: [Link to PR]
