# Migration Guide: Upgrading to AINative Code v1.0

This guide helps you migrate from beta/development versions to the official v1.0 release of AINative Code.

---

## Table of Contents

1. [Overview](#overview)
2. [Before You Upgrade](#before-you-upgrade)
3. [Upgrade Process](#upgrade-process)
4. [Configuration Changes](#configuration-changes)
5. [Breaking Changes](#breaking-changes)
6. [Deprecated Features](#deprecated-features)
7. [New Features to Adopt](#new-features-to-adopt)
8. [Migration Checklist](#migration-checklist)
9. [Troubleshooting](#troubleshooting)
10. [Rollback Instructions](#rollback-instructions)

---

## Overview

AINative Code v1.0 is the first production release. If you've been using development builds or beta versions, this guide will help you upgrade smoothly.

**Key Changes**:
- Stable configuration format
- Enhanced authentication system
- Expanded platform integrations
- Production-ready infrastructure

**Migration Time**: 15-30 minutes for most users

---

## Before You Upgrade

### 1. Backup Your Data

**Configuration**:
```bash
# Backup your configuration
cp ~/.config/ainative-code/config.yaml ~/.config/ainative-code/config.yaml.backup

# Backup session data
cp ~/.local/share/ainative-code/sessions.db ~/.local/share/ainative-code/sessions.db.backup
```

**Environment Variables**:
```bash
# Save current environment variables
env | grep AINATIVE_CODE > ~/ainative-env-backup.txt
```

### 2. Document Your Setup

Record your current setup:
- Which LLM providers you're using
- API keys and credentials location
- Custom configuration settings
- Installed MCP servers
- Active sessions you want to preserve

### 3. Check Version

```bash
# Check your current version
ainative-code --version

# View current configuration
ainative-code config show
```

### 4. Review Release Notes

Read the [v1.0 Release Notes](v1.0-release-notes.md) to understand new features and changes.

---

## Upgrade Process

### Option 1: Homebrew (macOS - Recommended)

```bash
# Update Homebrew
brew update

# Upgrade AINative Code
brew upgrade ainative-code

# Verify installation
ainative-code --version
```

### Option 2: Direct Download

#### macOS

```bash
# Download latest version
# Apple Silicon
curl -LO https://github.com/AINative-studio/ainative-code/releases/latest/download/ainative-code-darwin-arm64

# Intel
curl -LO https://github.com/AINative-studio/ainative-code/releases/latest/download/ainative-code-darwin-amd64

# Backup old binary
sudo mv /usr/local/bin/ainative-code /usr/local/bin/ainative-code.old

# Install new binary
chmod +x ainative-code-darwin-*
sudo mv ainative-code-darwin-* /usr/local/bin/ainative-code

# Verify
ainative-code --version
```

#### Linux

```bash
# Download for your architecture
curl -LO https://github.com/AINative-studio/ainative-code/releases/latest/download/ainative-code-linux-amd64

# Backup and install
sudo mv /usr/local/bin/ainative-code /usr/local/bin/ainative-code.old
chmod +x ainative-code-linux-amd64
sudo mv ainative-code-linux-amd64 /usr/local/bin/ainative-code

# Verify
ainative-code --version
```

#### Windows

```powershell
# Download latest release
Invoke-WebRequest -Uri "https://github.com/AINative-studio/ainative-code/releases/latest/download/ainative-code-windows-amd64.exe" -OutFile "ainative-code-new.exe"

# Backup old version
Move-Item C:\Windows\System32\ainative-code.exe C:\Windows\System32\ainative-code-old.exe -Force

# Install new version
Move-Item ainative-code-new.exe C:\Windows\System32\ainative-code.exe

# Verify
ainative-code --version
```

### Option 3: Docker

```bash
# Pull latest image
docker pull ghcr.io/ainative-studio/ainative-code:1.0.0

# Update your docker-compose.yml or run commands
docker run -it --rm ghcr.io/ainative-studio/ainative-code:1.0.0
```

### Option 4: Build from Source

```bash
# Clone repository
git clone https://github.com/AINative-studio/ainative-code.git
cd ainative-code

# Checkout v1.0.0
git checkout v1.0.0

# Build
make build

# Install
make install

# Verify
ainative-code --version
```

---

## Configuration Changes

### Configuration File Migration

v1.0 introduces an enhanced configuration schema. Most beta configurations will work, but you should update to the new format.

#### Old Format (Beta)
```yaml
# Beta configuration
llm:
  provider: anthropic
  api_key: sk-ant-xxx

ainative:
  endpoint: https://api.ainative.studio
  token: xxx
```

#### New Format (v1.0)
```yaml
# v1.0 configuration with expanded options
app:
  name: ainative-code
  version: 1.0.0
  environment: production
  debug: false

# LLM Provider Configuration
providers:
  anthropic:
    api_key: "${ANTHROPIC_API_KEY}"
    model: claude-3-5-sonnet-20241022
    max_tokens: 4096
    temperature: 0.7
    enable_streaming: true
    enable_cache: true

  openai:
    api_key: "${OPENAI_API_KEY}"
    model: gpt-4
    max_tokens: 4096

# AINative Platform
ainative:
  auth:
    method: jwt  # jwt, oauth2, or api_key
    token_cache: ~/.config/ainative-code/tokens.json
    auto_refresh: true

  organization:
    org_id: "${AINATIVE_ORG_ID}"
    workspace_id: "${AINATIVE_WORKSPACE_ID}"

  zerodb:
    endpoint: https://api.zerodb.ainative.studio
    project_id: "${ZERODB_PROJECT_ID}"
    ssl_mode: require
    connection_pool:
      max_connections: 10
      idle_timeout: 300

# Performance Settings
cache:
  enabled: true
  type: memory
  ttl: 3600
  max_size: 1000

rate_limit:
  enabled: true
  requests_per_minute: 60
  burst_size: 10

# Logging
logging:
  level: info
  format: json
  output: ~/.local/share/ainative-code/logs/app.log
  enable_rotation: true
  max_size: 100
  max_backups: 10
  max_age: 30
  compress: true

# Security
security:
  encryption:
    enabled: true
    key: "${AINATIVE_ENCRYPTION_KEY}"

  tls:
    min_version: "1.2"
    verify_certificates: true
```

### Migrate Configuration

Run the migration helper:

```bash
# Automatically migrate configuration
ainative-code config migrate --from ~/.config/ainative-code/config.yaml.backup

# Or initialize fresh config
ainative-code init
```

### Environment Variables

New environment variable names in v1.0:

| Old Variable | New Variable | Notes |
|--------------|--------------|-------|
| `AINATIVE_API_KEY` | `AINATIVE_CODE_AUTH_TOKEN` | For AINative platform |
| `CLAUDE_API_KEY` | `AINATIVE_CODE_ANTHROPIC_API_KEY` | Claude API key |
| `OPENAI_KEY` | `AINATIVE_CODE_OPENAI_API_KEY` | OpenAI API key |
| `AINATIVE_ENDPOINT` | `AINATIVE_CODE_ZERODB_ENDPOINT` | ZeroDB endpoint |
| `LOG_LEVEL` | `AINATIVE_CODE_LOGGING_LEVEL` | Log level |

**Update your .env or shell profile**:

```bash
# Old
export AINATIVE_API_KEY=xxx
export CLAUDE_API_KEY=sk-ant-xxx

# New (v1.0)
export AINATIVE_CODE_AUTH_TOKEN=xxx
export AINATIVE_CODE_ANTHROPIC_API_KEY=sk-ant-xxx
```

---

## Breaking Changes

### 1. Command Structure Changes

Some commands have been reorganized for better consistency:

| Old Command | New Command | Notes |
|-------------|-------------|-------|
| `ainative-code db query` | `ainative-code zerodb table query` | ZeroDB NoSQL operations |
| `ainative-code vector search` | `ainative-code zerodb vector search` | Vector operations |
| `ainative-code auth token` | `ainative-code auth status` | Show auth status |
| `ainative-code memory store` | `ainative-code zerodb memory store` | Agent memory |

**Migration Script**:

```bash
#!/bin/bash
# Alias old commands to new ones

alias ainative-db='ainative-code zerodb table'
alias ainative-vector='ainative-code zerodb vector'
alias ainative-memory='ainative-code zerodb memory'
```

### 2. Configuration Key Changes

Renamed configuration keys:

| Old Key | New Key |
|---------|---------|
| `llm.provider` | `providers.{provider_name}` |
| `ainative.endpoint` | `ainative.zerodb.endpoint` |
| `cache.ttl_seconds` | `cache.ttl` |
| `log.level` | `logging.level` |
| `log.format` | `logging.format` |

### 3. API Changes

If you're using AINative Code as a library:

```go
// Old (Beta)
import "github.com/ainative/code/pkg/llm"
client := llm.NewClient("anthropic", apiKey)

// New (v1.0)
import "github.com/AINative-studio/ainative-code/internal/providers/anthropic"
import "github.com/AINative-studio/ainative-code/internal/config"

cfg := &config.Config{...}
client := anthropic.NewClient(cfg)
```

### 4. Session Storage

Session database schema has been updated. Sessions will be automatically migrated on first run.

```bash
# Manually trigger migration if needed
ainative-code session migrate
```

---

## Deprecated Features

The following features are deprecated and will be removed in v2.0:

### 1. Legacy Authentication (Removed in v1.0)

**Old**: Simple API key auth
**New**: JWT/OAuth 2.0 with keychain storage

**Migration**:
```bash
# Re-authenticate using new flow
ainative-code auth logout
ainative-code auth login
```

### 2. Inline API Keys (Deprecated)

**Old**: API keys in config file
**New**: Environment variables or OS keychain

**Migration**:
```bash
# Move API keys to environment
export AINATIVE_CODE_ANTHROPIC_API_KEY="sk-ant-xxx"

# Or store in keychain
ainative-code config set anthropic.api_key --keychain
```

### 3. Legacy Log Format (Deprecated)

**Old**: Plain text logs
**New**: Structured JSON logging

**Migration**:
Update your log parsing tools to handle JSON format, or set:
```yaml
logging:
  format: text  # Temporary compatibility
```

---

## New Features to Adopt

### 1. OS Keychain Integration

Store credentials securely:

```bash
# Store API key in keychain
ainative-code auth keychain set anthropic-key

# Retrieve from keychain
ainative-code auth keychain get anthropic-key

# List stored credentials
ainative-code auth keychain list
```

### 2. Prompt Caching

Enable caching for cost savings:

```bash
# Enable in config
ainative-code config set providers.anthropic.enable_cache true

# Or per-request
ainative-code chat --enable-cache "Your prompt here"
```

### 3. Extended Thinking

Use Claude's extended thinking:

```bash
# Enable extended thinking
ainative-code chat --enable-thinking "Complex problem..."

# Configure thinking budget
ainative-code config set providers.anthropic.thinking_budget 10000
```

### 4. MCP Server Integration

Install and use MCP servers:

```bash
# List available servers
ainative-code mcp marketplace list

# Install a server
ainative-code mcp server install @modelcontextprotocol/server-filesystem

# Start server
ainative-code mcp server start filesystem

# Use server resources
ainative-code mcp resource list --server filesystem
```

### 5. Quantum Vector Operations

Use quantum features for vector operations:

```bash
# Compress vectors
ainative-code zerodb quantum compress --vector-id vec_123 --compression-ratio 0.5

# Create knowledge graphs
ainative-code zerodb quantum entangle --vector-id-1 vec_a --vector-id-2 vec_b

# Enhanced search
ainative-code zerodb quantum search --query-vector '[...]' --use-quantum-boost
```

---

## Migration Checklist

Use this checklist to ensure a smooth migration:

### Pre-Migration
- [ ] Backup configuration files
- [ ] Backup session database
- [ ] Document current setup
- [ ] Export critical sessions
- [ ] Note all installed MCP servers

### During Migration
- [ ] Upgrade binary/package
- [ ] Verify new version installed
- [ ] Run configuration migration
- [ ] Update environment variables
- [ ] Re-authenticate with AINative platform
- [ ] Test LLM provider connections

### Post-Migration
- [ ] Verify all commands work
- [ ] Check session history preserved
- [ ] Test each LLM provider
- [ ] Verify ZeroDB connectivity
- [ ] Test MCP servers
- [ ] Review new configuration
- [ ] Enable new features (caching, thinking)
- [ ] Update scripts/aliases
- [ ] Remove old backup files

### Validation
- [ ] Run: `ainative-code auth status`
- [ ] Run: `ainative-code config validate`
- [ ] Run: `ainative-code chat "test message"`
- [ ] Run: `ainative-code session list`
- [ ] Run: `ainative-code zerodb project info`

---

## Troubleshooting

### Issue: "Configuration file format invalid"

**Solution**:
```bash
# Validate configuration
ainative-code config validate

# See specific errors
ainative-code config validate --verbose

# Auto-fix common issues
ainative-code config migrate --auto-fix
```

### Issue: "Authentication failed"

**Solution**:
```bash
# Clear auth cache
rm -rf ~/.config/ainative-code/tokens.json

# Re-authenticate
ainative-code auth login

# Check status
ainative-code auth status
```

### Issue: "Command not found"

**Solution**:
```bash
# Check version
ainative-code --version

# Verify PATH
echo $PATH | grep ainative-code

# Reinstall if needed
brew reinstall ainative-code  # macOS
```

### Issue: "Sessions not appearing"

**Solution**:
```bash
# Check database
ls -lh ~/.local/share/ainative-code/sessions.db

# Manually migrate
ainative-code session migrate --force

# Check session list
ainative-code session list --all
```

### Issue: "API keys not working"

**Solution**:
```bash
# Check environment variables
env | grep AINATIVE_CODE

# Test API key
ainative-code config test anthropic

# Store in keychain instead
ainative-code auth keychain set anthropic-key
```

### Issue: "MCP servers not loading"

**Solution**:
```bash
# List installed servers
ainative-code mcp server list

# Reinstall problematic server
ainative-code mcp server uninstall filesystem
ainative-code mcp server install @modelcontextprotocol/server-filesystem

# Check logs
ainative-code mcp server logs filesystem
```

---

## Rollback Instructions

If you need to rollback to a previous version:

### Rollback Binary

```bash
# macOS/Linux
sudo mv /usr/local/bin/ainative-code.old /usr/local/bin/ainative-code

# Windows
Move-Item C:\Windows\System32\ainative-code-old.exe C:\Windows\System32\ainative-code.exe -Force
```

### Restore Configuration

```bash
# Restore backup
cp ~/.config/ainative-code/config.yaml.backup ~/.config/ainative-code/config.yaml

# Restore session database
cp ~/.local/share/ainative-code/sessions.db.backup ~/.local/share/ainative-code/sessions.db
```

### Rollback with Homebrew

```bash
# List available versions
brew list --versions ainative-code

# Install specific version
brew unlink ainative-code
brew install ainative-code@0.9.0
```

---

## Getting Help

If you encounter issues during migration:

1. **Check Documentation**: [docs.ainative.studio/code](https://docs.ainative.studio/code)
2. **Search Issues**: [GitHub Issues](https://github.com/AINative-studio/ainative-code/issues)
3. **Ask Community**: [GitHub Discussions](https://github.com/AINative-studio/ainative-code/discussions)
4. **Contact Support**: support@ainative.studio

When reporting migration issues, include:
- Source version (before upgrade)
- Target version (after upgrade)
- Operating system and version
- Configuration file (redact secrets)
- Error messages
- Steps to reproduce

---

## Post-Migration Best Practices

After successful migration:

### 1. Review Security

```bash
# Use keychain for secrets
ainative-code auth keychain set anthropic-key
ainative-code auth keychain set openai-key

# Enable encryption
ainative-code config set security.encryption.enabled true

# Review TLS settings
ainative-code config get security.tls
```

### 2. Optimize Performance

```bash
# Enable caching
ainative-code config set cache.enabled true

# Configure rate limiting
ainative-code config set rate_limit.enabled true
ainative-code config set rate_limit.requests_per_minute 100

# Set up connection pooling
ainative-code config set ainative.zerodb.connection_pool.max_connections 20
```

### 3. Configure Logging

```bash
# Set appropriate log level
ainative-code config set logging.level info

# Enable log rotation
ainative-code config set logging.enable_rotation true
ainative-code config set logging.max_size 100
ainative-code config set logging.max_backups 10
```

### 4. Test New Features

```bash
# Try prompt caching
ainative-code chat --enable-cache "Tell me about Go"

# Try extended thinking
ainative-code chat --enable-thinking "Solve this complex algorithm problem"

# Try quantum compression
ainative-code zerodb quantum compress --vector-id vec_123 --compression-ratio 0.6
```

---

## Next Steps

After migration:

1. **Read Release Notes**: Review [v1.0 Release Notes](v1.0-release-notes.md) for all new features
2. **Explore New Features**: Try MCP servers, quantum operations, extended thinking
3. **Update Scripts**: Update any automation scripts to use new commands
4. **Share Feedback**: Let us know how the migration went
5. **Stay Updated**: Watch the repository for v1.1 announcements

---

## Migration Support

Need help with migration?

- **Documentation**: This guide and [v1.0 Release Notes](v1.0-release-notes.md)
- **Examples**: Check [docs/examples/](../examples/) for updated examples
- **Support**: support@ainative.studio
- **Community**: [GitHub Discussions](https://github.com/AINative-studio/ainative-code/discussions)

We're here to help make your migration smooth!

---

**Copyright © 2024 AINative Studio. All rights reserved.**
