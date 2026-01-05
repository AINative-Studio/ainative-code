# Troubleshooting Guide

This guide helps you diagnose and resolve common issues with AINative Code.

## Table of Contents

1. [Installation Issues](#installation-issues)
2. [Configuration Problems](#configuration-problems)
3. [Authentication Errors](#authentication-errors)
4. [Provider Issues](#provider-issues)
5. [Session Management](#session-management)
6. [Performance Issues](#performance-issues)
7. [Tool Execution](#tool-execution)
8. [Platform Integration](#platform-integration)
9. [Debug Mode](#debug-mode)
10. [Getting Help](#getting-help)

## Installation Issues

### Command Not Found

**Problem:** `ainative-code: command not found` after installation

**Solutions:**

```bash
# Check if binary exists
which ainative-code
ls -la /usr/local/bin/ainative-code

# Add to PATH if needed
export PATH="$PATH:/usr/local/bin"
echo 'export PATH="$PATH:/usr/local/bin"' >> ~/.bashrc  # or ~/.zshrc

# Reload shell config
source ~/.bashrc  # or source ~/.zshrc

# Verify installation location
brew --prefix ainative-code  # macOS with Homebrew
```

### Permission Denied

**Problem:** Permission errors when running commands

**Solutions:**

```bash
# Make binary executable
chmod +x /usr/local/bin/ainative-code

# Fix ownership
sudo chown $USER /usr/local/bin/ainative-code

# For manual installation
sudo mv ainative-code /usr/local/bin/
sudo chmod +x /usr/local/bin/ainative-code
```

### Binary Won't Execute (macOS)

**Problem:** "ainative-code cannot be opened because the developer cannot be verified"

**Solutions:**

```bash
# Remove quarantine attribute
xattr -d com.apple.quarantine /usr/local/bin/ainative-code

# Or: System Preferences > Security & Privacy > Allow anyway
```

### Build from Source Fails

**Problem:** Compilation errors when building

**Solutions:**

```bash
# Verify Go version
go version  # Should be 1.21 or higher

# Update Go if needed
brew upgrade go  # macOS
# Or download from golang.org

# Clean and rebuild
make clean
go mod tidy
go mod verify
make build

# Check for dependency issues
go mod download
go build -v ./...
```

## Configuration Problems

### Configuration Not Loading

**Problem:** Settings not being applied

**Solutions:**

```bash
# Check which config file is being used
ainative-code --verbose config show

# Verify config file exists
ls -la ~/.config/ainative-code/config.yaml

# Check config file syntax
cat ~/.config/ainative-code/config.yaml | yq eval

# Re-initialize if corrupted
mv ~/.config/ainative-code/config.yaml ~/.config/ainative-code/config.yaml.bak
ainative-code init

# Validate configuration
ainative-code config validate
```

### Environment Variables Not Working

**Problem:** Environment variables not being read

**Solutions:**

```bash
# Verify variables are set
env | grep AINATIVE
env | grep ANTHROPIC
env | grep OPENAI

# Check variable naming
export AINATIVE_LLM_DEFAULT_PROVIDER=anthropic  # Correct
export ainative.llm.default_provider=anthropic  # Wrong

# Use source or export
source .env  # Loads variables
export $(cat .env | xargs)  # Alternative

# Test with explicit variable
ANTHROPIC_API_KEY=sk-ant-... ainative-code chat "test"
```

### Invalid Configuration Values

**Problem:** Configuration validation errors

**Solutions:**

```bash
# Check error message for specific issue
ainative-code config validate

# Common issues:
# 1. Invalid YAML syntax
cat config.yaml | yq eval

# 2. Wrong data types
# temperature should be float, not string
temperature: 0.7  # Correct
temperature: "0.7"  # Wrong

# 3. Missing required fields
llm:
  anthropic:
    api_key: "..."  # Required
    model: "..."     # Required

# 4. Invalid enum values
environment: development  # Valid: development, staging, production
environment: dev  # Invalid
```

## Authentication Errors

### API Key Invalid

**Problem:** Invalid API key errors

**Solutions:**

```bash
# Verify API key format
# Anthropic: sk-ant-api03-...
# OpenAI: sk-...

# Check for whitespace
export ANTHROPIC_API_KEY=$(echo $ANTHROPIC_API_KEY | tr -d '[:space:]')

# Test API key directly
curl https://api.anthropic.com/v1/messages \
  -H "x-api-key: $ANTHROPIC_API_KEY" \
  -H "anthropic-version: 2023-06-01" \
  -H "content-type: application/json" \
  -d '{
    "model": "claude-3-haiku-20240307",
    "max_tokens": 10,
    "messages": [{"role": "user", "content": "Hi"}]
  }'

# Regenerate key if invalid
# Visit provider console and create new key
```

### OAuth Login Fails

**Problem:** Cannot complete OAuth login

**Solutions:**

```bash
# Check callback URL is accessible
curl http://localhost:8080/health

# Use different port if 8080 is busy
ainative-code auth login --redirect-url http://localhost:8081/callback

# Check firewall settings
sudo ufw status  # Linux
# Ensure port 8080 is allowed

# Clear browser cache/cookies
# Try incognito/private browsing

# Manual URL if browser doesn't open
# Copy the URL shown and paste in browser
```

### Token Expired

**Problem:** Access token expired errors

**Solutions:**

```bash
# Refresh token
ainative-code auth token refresh

# Check token status
ainative-code auth token status

# Re-login if refresh fails
ainative-code auth logout
ainative-code auth login

# Enable auto-refresh
ainative-code config set platform.authentication.auto_refresh true
```

### Keychain Access Denied

**Problem:** Cannot access OS keychain

**Solutions:**

**macOS:**
```bash
# Grant terminal access
# System Preferences > Security & Privacy > Privacy > Automation

# Or use file-based storage
ainative-code config set security.use_keychain false
```

**Linux:**
```bash
# Install keyring
sudo apt install gnome-keyring  # Ubuntu
sudo dnf install gnome-keyring  # Fedora

# Start keyring daemon
eval $(gnome-keyring-daemon --start)

# Or use file-based storage
ainative-code config set security.use_keychain false
```

## Provider Issues

### Rate Limited

**Problem:** Rate limit exceeded errors

**Solutions:**

```bash
# Check current usage
ainative-code analytics providers

# Wait before retrying
sleep 60

# Use different provider
ainative-code chat --provider openai  # Switch from anthropic

# Enable fallback
llm:
  fallback:
    enabled: true
    providers:
      - anthropic
      - openai

# Reduce request rate
performance:
  rate_limit:
    enabled: true
    requests_per_minute: 30
```

### Slow Responses

**Problem:** Responses taking too long

**Solutions:**

```bash
# Use faster model
ainative-code chat --model claude-3-haiku-20240307

# Reduce max_tokens
llm:
  anthropic:
    max_tokens: 1024  # Lower for faster response

# Enable streaming
ainative-code chat --stream

# Check network
ping api.anthropic.com

# Try different provider
ainative-code chat --provider openai --model gpt-3.5-turbo
```

### Connection Timeout

**Problem:** Connection timeouts

**Solutions:**

```bash
# Increase timeout
llm:
  anthropic:
    timeout: 600s  # 10 minutes

# Check network connectivity
ping api.anthropic.com
curl -I https://api.anthropic.com

# Check proxy settings
export HTTP_PROXY=http://proxy:8080
export HTTPS_PROXY=http://proxy:8080

# Try different endpoint (if available)
llm:
  anthropic:
    base_url: "https://api.anthropic.com"  # Try different region
```

### Model Not Available

**Problem:** Model not found or not available

**Solutions:**

```bash
# Verify model name
# Anthropic: claude-3-5-sonnet-20241022 (with date)
# OpenAI: gpt-4-turbo-preview

# Check model availability
curl https://api.anthropic.com/v1/models \
  -H "x-api-key: $ANTHROPIC_API_KEY"

# Use default model
ainative-code config unset llm.anthropic.model  # Uses provider default

# Check for newer model versions
# Visit provider documentation
```

## Session Management

### Session Not Found

**Problem:** Cannot find or resume session

**Solutions:**

```bash
# List all sessions
ainative-code session list --all

# Search for session
ainative-code session list --search "keyword"

# Check session ID
# IDs are case-sensitive: abc123def456

# Verify database
ls -la ~/.config/ainative-code/sessions.db

# Sync from ZeroDB if enabled
ainative-code session sync --pull
```

### Session Export Fails

**Problem:** Cannot export session

**Solutions:**

```bash
# Check permissions
ls -la ~/exports/
mkdir -p ~/exports/

# Use full path
ainative-code session export abc123 -o $(pwd)/session.json

# Check disk space
df -h

# Export as different format
ainative-code session export abc123 --format markdown -o session.md
```

### Session Database Corrupted

**Problem:** Database errors

**Solutions:**

```bash
# Backup current database
cp ~/.config/ainative-code/sessions.db ~/.config/ainative-code/sessions.db.bak

# Try to recover
sqlite3 ~/.config/ainative-code/sessions.db ".recover" | \
  sqlite3 ~/.config/ainative-code/sessions-recovered.db

# Restore from ZeroDB
ainative-code session sync --pull --force

# Last resort: rebuild database
mv ~/.config/ainative-code/sessions.db ~/.config/ainative-code/sessions.db.old
# Sessions will be recreated on next use
```

## Performance Issues

### High Memory Usage

**Problem:** Excessive memory consumption

**Solutions:**

```bash
# Check memory usage
top -p $(pgrep ainative-code)

# Reduce cache size
performance:
  cache:
    max_size: 50  # MB (default: 100)

# Limit concurrent operations
performance:
  concurrency:
    max_workers: 5  # (default: 10)

# Clear cache
rm -rf ~/.cache/ainative-code/*

# Reduce max_tokens
llm:
  anthropic:
    max_tokens: 2048  # Lower limit
```

### Slow Startup

**Problem:** Application takes long to start

**Solutions:**

```bash
# Disable unnecessary services
services:
  zerodb:
    enabled: false  # If not using
  strapi:
    enabled: false

# Reduce health check frequency
mcp:
  health_check_interval: 5m  # (default: 1m)

# Disable cache on startup
performance:
  cache:
    enabled: false

# Profile startup
time ainative-code --version
```

### High CPU Usage

**Problem:** High CPU utilization

**Solutions:**

```bash
# Check what's running
top -p $(pgrep ainative-code)

# Reduce concurrent operations
performance:
  concurrency:
    max_workers: 3

# Disable background tasks
mcp:
  health_checks_enabled: false

# Use simpler models
ainative-code chat --model claude-3-haiku-20240307
```

## Tool Execution

### Tool Permission Denied

**Problem:** Tools cannot execute

**Solutions:**

```bash
# Check tool permissions
ls -la ~/.config/ainative-code/tools/

# Make scripts executable
chmod +x ~/.config/ainative-code/tools/*.sh

# Check allowed paths
tools:
  filesystem:
    allowed_paths:
      - /workspace  # Ensure path is allowed

# Verify tool configuration
ainative-code tools list
```

### MCP Server Not Responding

**Problem:** MCP server connection issues

**Solutions:**

```bash
# Check MCP server status
ainative-code mcp list-servers

# Test server directly
curl http://localhost:3000/health

# Restart MCP server
# (depends on your setup)

# Remove and re-add server
ainative-code mcp remove-server mytools
ainative-code mcp add-server --name mytools --url http://localhost:3000

# Check server logs
# (location depends on your server implementation)
```

### Command Timeout

**Problem:** Commands timing out

**Solutions:**

```bash
# Increase timeout
tools:
  terminal:
    timeout: 600s  # 10 minutes

# Run in background
ainative-code chat "Run long-running script in background"

# Split into smaller commands
# Instead of one long command, break into steps
```

## Platform Integration

### ZeroDB Connection Fails

**Problem:** Cannot connect to ZeroDB

**Solutions:**

```bash
# Check ZeroDB status
ainative-code zerodb ping

# Verify endpoint
ainative-code config get services.zerodb.endpoint

# Test connection
psql $ZERODB_ENDPOINT

# Check SSL settings
services:
  zerodb:
    ssl: true
    ssl_mode: require  # or: prefer, disable

# Check authentication
ainative-code auth whoami

# Verify network access
telnet zerodb.ainative.studio 5432
```

### Strapi API Errors

**Problem:** Strapi integration issues

**Solutions:**

```bash
# Check Strapi endpoint
curl https://cms.ainative.studio/api/health

# Verify API token
ainative-code config get services.strapi.api_key

# Test authentication
curl https://cms.ainative.studio/api/blog-posts \
  -H "Authorization: Bearer $STRAPI_TOKEN"

# Check permissions
# Ensure your Strapi API token has correct permissions
```

### Design Token Sync Fails

**Problem:** Cannot sync design tokens

**Solutions:**

```bash
# Check Figma token
echo $FIGMA_TOKEN

# Verify file ID
# Should be from Figma file URL: figma.com/file/FILE_ID/...

# Test manually
ainative-code design extract --source figma --file-id "..." --output test.json

# Check output directory
mkdir -p ./design-tokens
ainative-code design extract ... --output ./design-tokens/tokens.json
```

## Debug Mode

### Enable Verbose Logging

```bash
# Run with verbose flag
ainative-code --verbose chat

# Or enable in config
logging:
  level: debug
  format: console

# View logs
tail -f ~/.config/ainative-code/logs/app.log
```

### Capture Debug Information

```bash
# Enable all debug options
export AINATIVE_APP_DEBUG=true
export AINATIVE_LOGGING_LEVEL=debug

# Run command and capture output
ainative-code --verbose chat "test" 2>&1 | tee debug.log

# Include system information
ainative-code version --verbose
ainative-code config show > config-debug.txt
env | grep AINATIVE >> config-debug.txt
```

### Check System Status

```bash
# Verify all components
ainative-code health

# Expected output:
# Component           Status    Details
# ─────────────────   ───────   ─────────────
# Config              OK        Loaded from ~/.config/...
# Authentication      OK        Logged in as user@example.com
# LLM Provider        OK        anthropic (claude-3-5-sonnet)
# ZeroDB              OK        Connected
# MCP Servers         OK        2 servers, 15 tools
# Tools               OK        4 enabled
```

## Common Error Messages

### "Provider not configured"

**Solution:**

```bash
ainative-code config set llm.default_provider anthropic
export ANTHROPIC_API_KEY="sk-ant-..."
```

### "Session database locked"

**Solution:**

```bash
# Close other ainative-code instances
pkill ainative-code

# Remove lock file
rm ~/.config/ainative-code/sessions.db-lock
```

### "Tool execution blocked"

**Solution:**

```bash
# Add to allowed commands
tools:
  terminal:
    allowed_commands:
      - git
      - npm
      - your-command
```

### "Rate limit exceeded"

**Solution:**

```bash
# Wait and retry
sleep 60
ainative-code chat --provider openai "retry message"
```

### "Certificate verification failed"

**Solution:**

```bash
# Update CA certificates
sudo update-ca-certificates  # Linux
brew install ca-certificates  # macOS

# Or set custom cert
export SSL_CERT_FILE=/path/to/ca-bundle.crt

# Temporary workaround (not recommended)
export SSL_VERIFY=false
```

## Logs and Diagnostics

### Log Locations

```bash
# Application logs
~/.config/ainative-code/logs/app.log

# Audit logs
~/.config/ainative-code/logs/audit.log

# Tool execution logs
~/.config/ainative-code/logs/tools.log

# Error logs
~/.config/ainative-code/logs/error.log
```

### View Recent Errors

```bash
# Last 50 error lines
grep ERROR ~/.config/ainative-code/logs/app.log | tail -50

# Today's errors
grep "$(date +%Y-%m-%d)" ~/.config/ainative-code/logs/error.log

# Specific error type
grep "authentication failed" ~/.config/ainative-code/logs/app.log
```

### Generate Diagnostic Report

```bash
# Create diagnostic bundle
ainative-code diagnostics --output diagnostics.zip

# Contents:
# - Configuration (with secrets redacted)
# - Recent logs
# - System information
# - Provider status
# - Session database stats
```

## Getting Help

If you've tried the above solutions and still have issues:

### 1. Check Documentation

- [Installation Guide](installation.md)
- [Configuration Guide](configuration.md)
- [FAQ](faq.md)

### 2. Search Existing Issues

[GitHub Issues](https://github.com/AINative-studio/ainative-code/issues)

### 3. Community Support

[GitHub Discussions](https://github.com/AINative-studio/ainative-code/discussions)

### 4. Create a Bug Report

Include:
- AINative Code version: `ainative-code --version`
- Operating system and version
- Steps to reproduce
- Error messages (check logs)
- Configuration (redact secrets!)
- Diagnostic report (if applicable)

```bash
# Generate bug report
ainative-code bug-report --output bug-report.md

# Review and redact sensitive information
# Submit to: https://github.com/AINative-studio/ainative-code/issues/new
```

### 5. Contact Support

For enterprise/commercial support:
- Email: support@ainative.studio
- Include diagnostic report
- Provide account/organization details

## Recovery Procedures

### Reset Configuration

```bash
# Backup current config
cp ~/.config/ainative-code/config.yaml ~/.config/ainative-code/config.yaml.bak

# Reset to defaults
rm ~/.config/ainative-code/config.yaml
ainative-code init

# Restore API keys
export ANTHROPIC_API_KEY="..."
```

### Clean Reinstall

```bash
# 1. Backup important data
ainative-code session export --all -o sessions-backup.json

# 2. Uninstall
brew uninstall ainative-code  # macOS
sudo rm /usr/local/bin/ainative-code  # Manual installation

# 3. Remove config and data
mv ~/.config/ainative-code ~/.config/ainative-code.bak

# 4. Reinstall
brew install ainative-studio/tap/ainative-code

# 5. Reconfigure
ainative-code init
export ANTHROPIC_API_KEY="..."

# 6. Restore sessions
ainative-code session import -i sessions-backup.json
```

### Factory Reset

```bash
# WARNING: This deletes all data

# 1. Logout
ainative-code auth logout

# 2. Remove all data
rm -rf ~/.config/ainative-code
rm -rf ~/.cache/ainative-code
rm -rf ~/.local/share/ainative-code

# 3. Start fresh
ainative-code init
```

## Preventive Measures

### Regular Backups

```bash
# Weekly session backup
0 0 * * 0 ainative-code session export --all -o ~/backups/sessions-$(date +\%Y\%m\%d).json

# Config backup
cp ~/.config/ainative-code/config.yaml ~/backups/config-$(date +\%Y\%m\%d).yaml
```

### Monitor Resources

```bash
# Check disk space
df -h ~/.config/ainative-code

# Monitor API usage
ainative-code analytics cost --month current

# Review logs regularly
tail -f ~/.config/ainative-code/logs/error.log
```

### Keep Updated

```bash
# Check for updates
ainative-code version --check-updates

# Update
brew upgrade ainative-code  # macOS
# Or download latest release
```

## Next Steps

- [FAQ](faq.md) - Frequently asked questions
- [Getting Started](getting-started.md) - Basic usage
- [Configuration Guide](configuration.md) - Detailed configuration
