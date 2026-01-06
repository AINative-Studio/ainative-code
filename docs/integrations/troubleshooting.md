# Integration Troubleshooting Guide

## Overview

This guide provides solutions to common issues when working with AINative Code integrations including ZeroDB, Design Tokens, Strapi CMS, RLHF, and authentication systems.

## Table of Contents

1. [General Issues](#general-issues)
2. [ZeroDB Integration](#zerodb-integration)
3. [Design Token Integration](#design-token-integration)
4. [Strapi CMS Integration](#strapi-cms-integration)
5. [RLHF Feedback](#rlhf-feedback)
6. [Authentication](#authentication)
7. [Network and Connectivity](#network-and-connectivity)
8. [Performance Issues](#performance-issues)

## General Issues

### Integration Not Available

**Problem:** Feature/integration not accessible

**Solutions:**

```bash
# 1. Check if service is enabled
ainative-code config get services.zerodb.enabled
ainative-code config get services.strapi.enabled

# 2. Enable service
ainative-code config set services.zerodb.enabled true

# 3. Verify authentication
ainative-code auth whoami

# 4. Check version
ainative-code --version
# Ensure you have the latest version

# 5. Reinstall if needed
brew upgrade ainative-code
```

### MCP Server Not Responding

**Problem:** MCP tools not working

**Solutions:**

```bash
# 1. Check MCP configuration
cat ~/.mcp.json

# 2. Restart MCP servers
# Kill any running instances
pkill -f mcp-server

# 3. Check MCP server logs
# Location varies by installation

# 4. Test MCP connection
curl http://localhost:3000/health  # Adjust port

# 5. Reinstall MCP servers if needed
npm install -g @ainative/mcp-server
```

## ZeroDB Integration

### Connection Failures

**Problem:** Cannot connect to ZeroDB

**Solutions:**

```bash
# 1. Check credentials
echo $ZERODB_PROJECT_ID
echo $ZERODB_API_KEY

# 2. Test connection
/zerodb-project-info

# 3. Verify endpoint
ainative-code config get services.zerodb.endpoint

# 4. Check network connectivity
ping zerodb.ainative.studio

# 5. Verify SSL/TLS
curl -v https://zerodb.ainative.studio/api/health
```

### Vector Dimension Mismatch

**Problem:** "Vector dimension mismatch" error

**Solutions:**

```javascript
// All vectors must be exactly 1536 dimensions

// Check vector length
if (vector.length !== 1536) {
  console.error(`Invalid dimension: ${vector.length}, expected 1536`);
}

// Pad or truncate if needed (not recommended)
function normalizeVector(vector) {
  if (vector.length === 1536) return vector;

  if (vector.length < 1536) {
    // Pad with zeros
    return [...vector, ...new Array(1536 - vector.length).fill(0)];
  } else {
    // Truncate
    return vector.slice(0, 1536);
  }
}

// Better: Use correct embedding model
// OpenAI text-embedding-ada-002 produces 1536 dimensions
```

### Query Performance Issues

**Problem:** Slow vector search queries

**Solutions:**

```bash
# 1. Add namespace filter
--namespace "documentation"

# 2. Increase similarity threshold
--similarity-threshold 0.8  # Higher = fewer results

# 3. Reduce top_k
--top-k 5  # Instead of 100

# 4. Add metadata filters
--metadata-filter '{"type": "doc"}'

# 5. Check project stats
/zerodb-project-stats
# Look for large vector counts

# 6. Consider partitioning data
# Create separate namespaces for different datasets
```

### PostgreSQL Instance Issues

**Problem:** Cannot connect to PostgreSQL instance

**Solutions:**

```bash
# 1. Check instance status
/zerodb-postgres-status

# 2. Wait if provisioning
# Takes 2-3 minutes

# 3. Get connection details
/zerodb-postgres-connection

# 4. Test connection with psql
psql "$CONNECTION_STRING"

# 5. Check SSL requirement
# Connection string should include sslmode=require

# 6. Verify credentials
# Ensure using correct credential type (primary, readonly, admin)

# 7. Restart instance if needed
/zerodb-postgres-restart
```

## Design Token Integration

### Figma Extraction Fails

**Problem:** Cannot extract tokens from Figma

**Solutions:**

```bash
# 1. Verify Figma token
echo $FIGMA_TOKEN

# 2. Test token
curl -H "X-Figma-Token: $FIGMA_TOKEN" \
  https://api.figma.com/v1/me

# 3. Check file ID
# URL: https://www.figma.com/file/ABC123/Design
# File ID is: ABC123

# 4. Verify file permissions
# Token must have access to the file

# 5. Check Figma API status
# https://status.figma.com/

# 6. Try with --verbose flag
ainative-code design extract \
  --source figma \
  --file-id "ABC123" \
  --verbose
```

### Token Validation Errors

**Problem:** Tokens fail validation on upload

**Solutions:**

```json
// Common validation issues:

// 1. Invalid color format
{
  "name": "primary-color",
  "value": "blue",  // ✗ Wrong
  "type": "color"
}
// Fix: Use hex/rgb
{
  "name": "primary-color",
  "value": "#0000FF",  // ✓ Correct
  "type": "color"
}

// 2. Missing units
{
  "name": "spacing-md",
  "value": "16",  // ✗ Wrong
  "type": "spacing"
}
// Fix: Add unit
{
  "name": "spacing-md",
  "value": "16px",  // ✓ Correct
  "type": "spacing"
}

// 3. Invalid token type
{
  "name": "my-token",
  "value": "#000",
  "type": "colour"  // ✗ Wrong (British spelling)
}
// Fix: Use correct type
{
  "name": "my-token",
  "value": "#000",
  "type": "color"  // ✓ Correct
}
```

**Validate before upload:**

```bash
ainative-code design upload \
  --tokens tokens.json \
  --validate-only
```

### Code Generation Issues

**Problem:** Generated code not working

**Solutions:**

```bash
# 1. Check input format
cat tokens.json | jq

# 2. Regenerate with clean input
ainative-code design generate \
  --input tokens.json \
  --format css \
  --output styles/tokens.css \
  --force  # Overwrite existing

# 3. Verify output syntax
npx prettier --check styles/tokens.css

# 4. Try different format
ainative-code design generate \
  --input tokens.json \
  --format typescript \
  --output src/tokens.ts

# 5. Check for name conflicts
# Token names must be unique
```

## Strapi CMS Integration

### Authentication Failures

**Problem:** 401 Unauthorized when accessing Strapi

**Solutions:**

```bash
# 1. Verify API token
echo $STRAPI_API_KEY

# 2. Check token permissions in Strapi
# Settings > API Tokens > View permissions

# 3. Regenerate token
# In Strapi admin: Settings > API Tokens > Create New

# 4. Test authentication
curl https://cms.ainative.studio/api/blog-posts \
  -H "Authorization: Bearer $STRAPI_API_KEY"

# 5. Reconfigure
ainative-code strapi config \
  --url "https://cms.ainative.studio" \
  --token "new-token-here"
```

### Content Not Found

**Problem:** 404 errors when accessing content

**Solutions:**

```bash
# 1. List all posts to find correct ID
ainative-code strapi blog list --json | \
  jq '.data[] | {id, title}'

# 2. Check content type exists
ainative-code strapi content types

# 3. Verify URL
ainative-code strapi config

# 4. Check Strapi permissions
# Strapi admin > Settings > Roles & Permissions

# 5. Test with Strapi admin interface
# Verify content exists in Strapi
```

### File Upload Issues

**Problem:** Cannot upload content from file

**Solutions:**

```bash
# 1. Check file exists
ls -la article.md

# 2. Use absolute path
ainative-code strapi blog create \
  --title "Title" \
  --content @/full/path/article.md

# 3. Check file encoding
file article.md  # Should be UTF-8 text

# 4. Verify file size
du -h article.md  # Strapi may have limits

# 5. Test with inline content first
ainative-code strapi blog create \
  --title "Test" \
  --content "# Test Content"
```

### Markdown Rendering Issues

**Problem:** Markdown not rendering correctly in Strapi

**Solutions:**

```bash
# 1. Validate markdown syntax
npx markdownlint article.md

# 2. Preview locally
npx marked article.md > preview.html

# 3. Check for special characters
# Escape backticks, quotes in content

# 4. Use proper code block syntax
# ```language
# code here
# ```

# 5. Test in Strapi admin
# Paste content directly to verify rendering
```

## RLHF Feedback

### Feedback Submission Fails

**Problem:** Cannot submit feedback

**Solutions:**

```bash
# 1. Check authentication
ainative-code auth whoami

# 2. Verify RLHF is enabled
ainative-code config get services.rlhf.enabled

# 3. Enable RLHF
ainative-code config set services.rlhf.enabled true

# 4. Check rate limits
# Wait 1 minute and retry

# 5. Test with simple feedback
/zerodb-rlhf-feedback
# Select thumbs_up with simple comment
```

### Missing Interaction ID

**Problem:** Don't have interaction ID for feedback

**Solutions:**

```javascript
// Track interaction IDs automatically
let lastInteractionId = null;

async function chat(message) {
  const response = await aiChat(message);
  lastInteractionId = response.interaction_id;
  return response;
}

async function provideFeedback(type, comment) {
  if (!lastInteractionId) {
    console.error('No recent interaction to provide feedback on');
    return;
  }

  await submitFeedback(lastInteractionId, type, comment);
}

// Usage
await chat("How do I use async/await?");
await provideFeedback('thumbs_up', 'Great explanation');
```

### Feedback Not Appearing

**Problem:** Submitted feedback not showing in stats

**Solutions:**

```bash
# 1. Wait for processing (5 minutes)

# 2. Check feedback list
ainative-code rlhf list --limit 10

# 3. Verify submission
# Look for confirmation message after submit

# 4. Check feedback ID
# Note the ID returned when submitting

# 5. Re-submit if needed
/zerodb-rlhf-feedback
```

## Authentication

### OAuth Login Failures

**Problem:** Cannot complete OAuth login

**Solutions:**

```bash
# 1. Check port availability
lsof -i :8080

# 2. Kill process using port
kill -9 <PID>

# 3. Use different port
ainative-code auth login --port 8081

# 4. Manual URL method
ainative-code auth login --print-url
# Copy URL and open in browser

# 5. Check firewall
sudo ufw status  # Linux
# Ensure port 8080 allowed

# 6. Clear browser cache
# Try incognito/private mode

# 7. Check callback URL
# Should be http://localhost:8080/callback
```

### API Key Invalid

**Problem:** "Invalid API key" errors

**Solutions:**

```bash
# 1. Check key format
# Anthropic: sk-ant-api03-...
# OpenAI: sk-...

# 2. Remove whitespace
export ANTHROPIC_API_KEY=$(echo $ANTHROPIC_API_KEY | tr -d '[:space:]')

# 3. Test key directly
curl https://api.anthropic.com/v1/messages \
  -H "x-api-key: $ANTHROPIC_API_KEY" \
  -H "anthropic-version: 2023-06-01" \
  -H "content-type: application/json" \
  -d '{"model":"claude-3-haiku-20240307","max_tokens":10,"messages":[{"role":"user","content":"Hi"}]}'

# 4. Regenerate key
# Visit provider console and create new key

# 5. Check for expiration
# Some keys may have expiration dates
```

### Token Expired

**Problem:** "Token expired" errors

**Solutions:**

```bash
# 1. Refresh token
ainative-code auth token refresh

# 2. Check token status
ainative-code auth token status

# 3. Re-login if refresh fails
ainative-code auth logout
ainative-code auth login

# 4. Enable auto-refresh
ainative-code config set platform.authentication.auto_refresh true
```

## Network and Connectivity

### Timeout Errors

**Problem:** Connection timeouts

**Solutions:**

```bash
# 1. Increase timeout
ainative-code config set services.zerodb.timeout 60s

# 2. Check network connectivity
ping api.ainative.studio

# 3. Test DNS resolution
nslookup api.ainative.studio

# 4. Check proxy settings
echo $HTTP_PROXY
echo $HTTPS_PROXY

# 5. Bypass proxy if needed
unset HTTP_PROXY
unset HTTPS_PROXY

# 6. Try different network
# Switch between WiFi/ethernet
```

### SSL/TLS Errors

**Problem:** SSL certificate verification failures

**Solutions:**

```bash
# 1. Update CA certificates
sudo update-ca-certificates  # Linux
brew install ca-certificates  # macOS

# 2. Check system date/time
date
# Incorrect time causes cert validation failures

# 3. Test with curl
curl -v https://api.ainative.studio

# 4. Temporary workaround (not recommended)
export SSL_VERIFY=false

# 5. Use custom cert bundle
export SSL_CERT_FILE=/path/to/ca-bundle.crt
```

### Rate Limiting

**Problem:** "Rate limit exceeded" errors

**Solutions:**

```bash
# 1. Wait before retrying
sleep 60  # Wait 1 minute

# 2. Check current usage
ainative-code analytics providers

# 3. Use different provider
ainative-code chat --provider openai

# 4. Enable request throttling
ainative-code config set performance.rate_limit.enabled true
ainative-code config set performance.rate_limit.requests_per_minute 30

# 5. Batch operations
# Instead of many small requests, use fewer larger ones

# 6. Upgrade plan
# Contact support for higher rate limits
```

## Performance Issues

### Slow Responses

**Problem:** Integration operations are slow

**Solutions:**

```bash
# 1. Use faster models
ainative-code chat --model claude-3-haiku-20240307

# 2. Reduce response size
ainative-code config set llm.max_tokens 1024

# 3. Enable caching
ainative-code config set performance.cache.enabled true

# 4. Check network latency
ping api.ainative.studio

# 5. Monitor system resources
top  # Check CPU/memory

# 6. Close unnecessary applications

# 7. Use local caching
ainative-code config set performance.cache.ttl 3600
```

### High Memory Usage

**Problem:** Excessive memory consumption

**Solutions:**

```bash
# 1. Reduce cache size
ainative-code config set performance.cache.max_size 50  # MB

# 2. Limit concurrent operations
ainative-code config set performance.concurrency.max_workers 5

# 3. Clear cache
rm -rf ~/.cache/ainative-code/*

# 4. Reduce max_tokens
ainative-code config set llm.max_tokens 2048

# 5. Monitor memory
top -p $(pgrep ainative-code)
```

## Getting Help

### Diagnostic Information

```bash
# Generate diagnostic report
ainative-code diagnostics --output diagnostics.zip

# Contents:
# - Configuration (secrets redacted)
# - Recent logs
# - System information
# - Integration status
```

### Enable Debug Logging

```bash
# Set debug level
export AINATIVE_LOGGING_LEVEL=debug

# Or in config
ainative-code config set logging.level debug

# View logs
tail -f ~/.config/ainative-code/logs/app.log
```

### Contact Support

If issues persist:

1. **Check Documentation**: Review integration guides
2. **Search Issues**: https://github.com/AINative-studio/ainative-code/issues
3. **Community**: https://community.ainative.studio
4. **Support**: support@ainative.studio

**Include in report:**
- AINative Code version
- Operating system
- Steps to reproduce
- Error messages
- Diagnostic report

## Next Steps

- [ZeroDB Integration](zerodb-integration.md)
- [Design Token Integration](design-token-integration.md)
- [Strapi CMS Integration](strapi-integration.md)
- [RLHF Feedback System](rlhf-integration.md)
- [Authentication Setup](authentication-setup.md)

## Resources

- [General Troubleshooting](../user-guide/troubleshooting.md)
- [Security Best Practices](../security/security-best-practices.md)
- [Performance Tuning](../configuration.md#performance)
