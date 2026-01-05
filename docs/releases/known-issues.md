# Known Issues - AINative Code v1.0

This document lists known limitations, platform-specific issues, and workarounds for AINative Code v1.0.

**Last Updated**: January 4, 2026

---

## Table of Contents

1. [Current Limitations](#current-limitations)
2. [Platform-Specific Issues](#platform-specific-issues)
3. [Performance Considerations](#performance-considerations)
4. [Integration Issues](#integration-issues)
5. [Workarounds](#workarounds)
6. [Planned Fixes](#planned-fixes)
7. [Reporting New Issues](#reporting-new-issues)

---

## Current Limitations

### 1. File Operations

**Issue**: Large file handling limitations

**Description**:
- Files larger than 100MB may cause performance degradation
- Very large codebases (>10,000 files) may slow down context gathering

**Impact**: Medium
**Affected Versions**: v1.0.0
**Workaround**:
```bash
# Limit context scope with .gitignore
echo "large-data-dir/" >> .gitignore

# Use file filters
ainative-code chat --exclude "*.log,*.db"
```

**Status**: Will be improved in v1.1 with streaming file processing

---

### 2. Session Management

**Issue**: Cross-device session sync not yet implemented

**Description**:
- Sessions are stored locally in SQLite
- Cannot resume sessions on different machines without manual export/import

**Impact**: Low
**Affected Versions**: v1.0.0
**Workaround**:
```bash
# Export session on machine A
ainative-code session export session_123 > session.json

# Import session on machine B
ainative-code session import session.json
```

**Status**: Cloud sync planned for v1.2

---

### 3. MCP Server Limitations

**Issue**: Not all MCP features supported in v1.0

**Description**:
- Some advanced MCP features like sampling are not yet implemented
- Server-to-server communication not supported

**Impact**: Low
**Affected Versions**: v1.0.0
**Workaround**: Use direct API calls for unsupported features

**Status**: Full MCP compliance planned for v1.1

---

### 4. Prompt Caching Availability

**Issue**: Prompt caching only available for Claude

**Description**:
- Anthropic Claude 3.5 Sonnet supports prompt caching
- Other providers (OpenAI, Gemini, etc.) do not have caching yet

**Impact**: Medium
**Affected Versions**: v1.0.0
**Workaround**: Use Claude for long contexts requiring caching

**Status**: Provider-dependent, will add as providers enable caching

---

### 5. Extended Thinking Token Limits

**Issue**: Extended thinking has token budget constraints

**Description**:
- Extended thinking limited to configured token budget
- May not complete for extremely complex problems

**Impact**: Low
**Affected Versions**: v1.0.0
**Workaround**:
```bash
# Increase thinking budget
ainative-code config set providers.anthropic.thinking_budget 20000

# Or split complex problems into smaller parts
```

**Status**: Working as designed

---

### 6. Concurrent Request Limits

**Issue**: Rate limiting on concurrent requests

**Description**:
- Default configuration limits concurrent requests to prevent API throttling
- Batch operations may be slower than expected

**Impact**: Low
**Affected Versions**: v1.0.0
**Workaround**:
```yaml
# Adjust in config.yaml
rate_limit:
  enabled: true
  requests_per_minute: 100  # Increase if your API allows
  burst_size: 20
```

**Status**: Configurable, working as designed

---

## Platform-Specific Issues

### macOS

#### Issue: Keychain Access Permission Prompt

**Description**:
- First-time keychain access requires user approval
- May prompt for password on each access in some macOS versions

**Impact**: Low
**Affected Versions**: All macOS versions
**Workaround**:
```bash
# Grant "Always Allow" when prompted
# Or use environment variables instead
export AINATIVE_CODE_ANTHROPIC_API_KEY="sk-ant-xxx"
```

**Status**: macOS security requirement, no fix planned

---

#### Issue: Gatekeeper Warning on First Run

**Description**:
- Downloaded binary may trigger Gatekeeper warning
- "ainative-code cannot be opened because it is from an unidentified developer"

**Impact**: Low
**Affected Versions**: v1.0.0
**Workaround**:
```bash
# Remove quarantine attribute
xattr -d com.apple.quarantine /usr/local/bin/ainative-code

# Or use Homebrew installation (recommended)
brew install ainative-studio/tap/ainative-code
```

**Status**: Code signing for macOS planned for v1.1

---

#### Issue: Terminal.app Color Support

**Description**:
- Some color schemes may not display correctly in default Terminal.app
- Extended characters may not render properly

**Impact**: Low
**Affected Versions**: All versions on macOS < 12
**Workaround**: Use iTerm2, Warp, or Alacritty for better color support

**Status**: Terminal.app limitation, no fix possible

---

### Linux

#### Issue: Keychain Integration on Headless Systems

**Description**:
- OS keychain integration requires a running desktop environment
- Fails on headless Linux servers

**Impact**: Medium
**Affected Versions**: v1.0.0
**Workaround**:
```bash
# Use environment variables instead
export AINATIVE_CODE_ANTHROPIC_API_KEY="sk-ant-xxx"

# Or use encrypted config file
ainative-code config set security.encryption.enabled true
```

**Status**: Expected behavior, environment variables recommended for servers

---

#### Issue: Unicode Characters on Some Distros

**Description**:
- Some Linux distributions have limited Unicode support in terminal
- Box drawing characters may appear broken

**Impact**: Low
**Affected Versions**: All versions on older distros
**Workaround**:
```bash
# Set locale
export LANG=en_US.UTF-8
export LC_ALL=en_US.UTF-8

# Or disable fancy UI
ainative-code config set ui.ascii_only true
```

**Status**: Terminal/locale issue, workaround available

---

### Windows

#### Issue: Windows Defender SmartScreen Warning

**Description**:
- Downloaded executable may trigger SmartScreen warning
- "Windows protected your PC" message

**Impact**: Low
**Affected Versions**: v1.0.0
**Workaround**: Click "More info" → "Run anyway"

**Status**: Code signing for Windows planned for v1.1

---

#### Issue: Path Length Limitations

**Description**:
- Windows MAX_PATH (260 characters) can cause issues with deep directory structures
- May affect session storage in nested paths

**Impact**: Low
**Affected Versions**: All versions on Windows < 10 (v1803)
**Workaround**:
```bash
# Enable long paths in Windows 10+
# Run as Administrator in PowerShell:
New-ItemProperty -Path "HKLM:\SYSTEM\CurrentControlSet\Control\FileSystem" -Name "LongPathsEnabled" -Value 1 -PropertyType DWORD -Force

# Or use shorter config paths
ainative-code config set --config-dir C:\ainative
```

**Status**: Windows limitation, long path support available in Win10+

---

#### Issue: PowerShell Execution Policy

**Description**:
- PowerShell may block execution of downloaded scripts
- Affects automated installation

**Impact**: Low
**Affected Versions**: All versions
**Workaround**:
```powershell
# Temporarily allow execution
Set-ExecutionPolicy -Scope Process -ExecutionPolicy Bypass

# Or use manual installation instead
```

**Status**: Windows security feature, working as designed

---

## Performance Considerations

### 1. Large Context Windows

**Issue**: Performance degradation with very large contexts

**Description**:
- Contexts over 100,000 tokens may cause slowdowns
- Memory usage increases linearly with context size

**Impact**: Medium
**Workaround**:
```bash
# Limit context size
ainative-code config set providers.anthropic.max_tokens 100000

# Use prompt caching for repeated contexts
ainative-code chat --enable-cache
```

**Status**: Inherent LLM API limitation, optimization planned for v1.1

---

### 2. SQLite Contention on High Load

**Issue**: Session database locking under concurrent access

**Description**:
- Multiple concurrent sessions may experience brief delays
- SQLite has limitations with write concurrency

**Impact**: Low
**Workaround**:
```bash
# Reduce concurrent operations
# Or use separate session files for parallel work
```

**Status**: SQLite limitation, PostgreSQL backend planned for v1.2

---

### 3. Initial Cold Start

**Issue**: First command execution slower than subsequent runs

**Description**:
- Binary needs to initialize on first run
- Configuration parsing and validation occurs

**Impact**: Low (< 1 second)
**Workaround**: None needed, subsequent runs are fast

**Status**: Expected behavior

---

## Integration Issues

### 1. ZeroDB API Rate Limits

**Issue**: ZeroDB free tier has rate limits

**Description**:
- Free tier: 100 requests/minute
- Batch operations may hit limits

**Impact**: Medium for free tier users
**Workaround**:
```bash
# Configure rate limiting
ainative-code config set rate_limit.requests_per_minute 90

# Or upgrade to paid tier
```

**Status**: API limitation, working as designed

---

### 2. Google Analytics Quota

**Issue**: GA4 API has daily quota limits

**Description**:
- Standard quota: 25,000 requests/day
- May be exhausted with frequent queries

**Impact**: Low
**Workaround**:
```bash
# Cache analytics data
ainative-code ga get-data --cache-ttl 3600

# Use quick reports instead of raw data
ainative-code ga quick-report
```

**Status**: Google API limitation, caching helps

---

### 3. Strapi CMS Authentication Timeout

**Issue**: Strapi auth tokens expire after inactivity

**Description**:
- Long idle periods require re-authentication
- Default timeout: 30 minutes

**Impact**: Low
**Workaround**: Auto-refresh tokens are enabled by default in v1.0

**Status**: Resolved in v1.0 with auto-refresh

---

### 4. Figma API Token Scope

**Issue**: Design token extraction requires specific Figma permissions

**Description**:
- Personal access token needs file read permissions
- Organization files may require additional approval

**Impact**: Low
**Workaround**: Ensure Figma token has correct scopes:
```bash
# Required scopes:
# - file:read
# - images:read
```

**Status**: Figma API requirement, documented

---

## Workarounds

### Memory Usage Optimization

If experiencing high memory usage:

```bash
# Reduce context window
ainative-code config set providers.anthropic.max_tokens 50000

# Disable caching if not needed
ainative-code config set cache.enabled false

# Limit concurrent operations
ainative-code config set concurrency.max_workers 2
```

### Network Timeout Issues

If experiencing timeout errors:

```bash
# Increase timeouts
ainative-code config set http.timeout 300

# Enable retries
ainative-code config set http.max_retries 5

# Use circuit breaker
ainative-code config set circuit_breaker.enabled true
```

### Terminal Display Issues

If TUI display appears broken:

```bash
# Use ASCII-only mode
ainative-code config set ui.ascii_only true

# Disable colors
ainative-code config set ui.colors.enabled false

# Or use JSON output
ainative-code chat --json
```

---

## Planned Fixes

### v1.1 (Q1 2026)

- [ ] Code signing for macOS and Windows binaries
- [ ] Improved file streaming for large files
- [ ] Full MCP protocol compliance
- [ ] Enhanced session export/import
- [ ] Batch operation optimization

### v1.2 (Q2 2026)

- [ ] Cloud session sync
- [ ] PostgreSQL backend option
- [ ] Advanced caching strategies
- [ ] WebSocket streaming
- [ ] Multi-language documentation

### v2.0 (Q3 2026)

- [ ] Plugin system for extensibility
- [ ] Distributed session storage
- [ ] Advanced analytics and monitoring
- [ ] Team collaboration features
- [ ] Enterprise SSO support

See [roadmap.md](roadmap.md) for complete roadmap.

---

## Reporting New Issues

If you encounter an issue not listed here:

### Before Reporting

1. **Check Documentation**: [docs.ainative.studio/code](https://docs.ainative.studio/code)
2. **Search Existing Issues**: [GitHub Issues](https://github.com/AINative-studio/ainative-code/issues)
3. **Try Workarounds**: Review workarounds in this document
4. **Update to Latest**: Ensure you're running the latest version

### How to Report

**Create a GitHub Issue** with:

1. **Clear Title**: Brief description of the issue
2. **Environment**:
   - AINative Code version (`ainative-code --version`)
   - Operating system and version
   - Go version (if building from source)
   - Terminal emulator
3. **Description**: What happened vs what you expected
4. **Steps to Reproduce**: Detailed steps to reproduce
5. **Logs**: Relevant error messages or logs
6. **Configuration**: Your config (redact secrets)

**Example**:

```markdown
## Issue: Chat command hangs with large files

**Environment**:
- AINative Code v1.0.0
- macOS 14.2 (Sonoma)
- iTerm2 3.4.23

**Description**:
Chat command hangs when processing files larger than 50MB

**Steps to Reproduce**:
1. Run `ainative-code chat`
2. Include large file: `@large-file.json`
3. Command hangs indefinitely

**Expected**: Should process or show error
**Actual**: Hangs without feedback

**Logs**:
```
[ERROR] context gathering timeout after 30s
```

**Configuration**:
```yaml
providers:
  anthropic:
    max_tokens: 100000
```
```

### Priority Levels

When reporting, suggest priority:

- **P0 (Critical)**: Crashes, data loss, security vulnerabilities
- **P1 (High)**: Major features broken, no workaround
- **P2 (Medium)**: Features partially broken, workaround exists
- **P3 (Low)**: Minor issues, cosmetic problems

---

## Getting Help

If you need assistance with known issues:

1. **Documentation**: Read relevant docs at [docs.ainative.studio/code](https://docs.ainative.studio/code)
2. **Community**: Ask in [GitHub Discussions](https://github.com/AINative-studio/ainative-code/discussions)
3. **Support**: Email support@ainative.studio
4. **FAQ**: Check [v1.0 Release Notes](v1.0-release-notes.md#faq)

### Support Response Times

- **Critical Issues (P0)**: 24 hours
- **High Priority (P1)**: 48 hours
- **Medium Priority (P2)**: 5 business days
- **Low Priority (P3)**: Best effort

---

## Issue Status Tracking

Track issue resolution:

- **Open**: Issue reported and under investigation
- **Confirmed**: Issue reproduced and confirmed
- **In Progress**: Fix in development
- **Fixed**: Fix available in release
- **Won't Fix**: Issue is by design or not feasible
- **Duplicate**: Duplicate of another issue

Check [GitHub Issues](https://github.com/AINative-studio/ainative-code/issues) for current status.

---

## Contributing Fixes

Want to help fix issues?

1. **Comment on Issue**: Let us know you're working on it
2. **Fork Repository**: Create your fork
3. **Create Branch**: `fix/issue-123`
4. **Implement Fix**: With tests
5. **Submit PR**: Reference the issue number

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for guidelines.

---

**Note**: This document is regularly updated. Check back for the latest known issues and workarounds.

**Last Updated**: January 4, 2026
**Version**: v1.0.0

---

**Copyright © 2024 AINative Studio. All rights reserved.**
