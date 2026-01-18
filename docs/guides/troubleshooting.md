# Troubleshooting AINative Integration

## Quick Diagnostics

Before troubleshooting specific issues, run these diagnostic commands:

```bash
# Check authentication status
ainative-code auth whoami

# Test with verbose mode
ainative-code chat-ainative -m "test" --verbose

# Check backend connectivity
curl http://localhost:8000/health
```

## Common Errors

### Authentication Errors

#### Error: "not authenticated"

**Full Error:**
```
Error: not authenticated: no access token found
```

**Cause:** No valid access token in configuration

**Solution:**
```bash
ainative-code auth login-backend \
  --email your-email@example.com \
  --password your-password
```

**Prevention:** Ensure you log in before using chat commands

---

#### Error: "unauthorized" (401)

**Full Error:**
```
Error: unauthorized (401)
```

**Cause:** Access token has expired or is invalid

**Solution 1 - Refresh token:**
```bash
ainative-code auth refresh-backend
```

**Solution 2 - Re-login if refresh fails:**
```bash
ainative-code auth login-backend \
  --email your-email@example.com \
  --password your-password
```

**Why this happens:**
- Access tokens expire after 15-30 minutes
- Token was invalidated server-side
- Token was corrupted in config file

---

#### Error: "invalid credentials"

**Full Error:**
```
Error: login failed: invalid credentials
```

**Cause:** Incorrect email or password

**Solutions:**
1. Verify your email address
2. Verify your password (check for typos)
3. Reset password at https://ainative.studio if forgotten
4. Ensure account is activated (check email for activation link)

---

### Network and Connection Errors

#### Error: "connection refused"

**Full Error:**
```
Error: Post "http://localhost:8000/v1/chat/completions": dial tcp 127.0.0.1:8000: connect: connection refused
```

**Cause:** Python backend is not running

**Solution - Start the backend:**
```bash
cd /Users/aideveloper/AINative-Code/python-backend
uvicorn app.main:app --reload
```

**Verify backend is running:**
```bash
curl http://localhost:8000/health
```

**Expected response:**
```json
{"status": "healthy"}
```

---

#### Error: "timeout"

**Full Error:**
```
Error: context deadline exceeded (timeout)
```

**Causes:**
- Backend is overloaded or slow
- Network connectivity issues
- Large/complex request taking too long

**Solution 1 - Increase timeout in config:**
```yaml
ainative:
  timeout: 180  # 3 minutes
```

**Solution 2 - Simplify request:**
- Reduce prompt length
- Use a faster model (GPT-3.5, Claude Sonnet)
- Break down complex requests

**Solution 3 - Check backend logs:**
```bash
# In the backend terminal, check for errors
tail -f /path/to/backend/logs
```

---

#### Error: "no such host"

**Full Error:**
```
Error: Post "http://api.ainative.studio/v1/chat/completions": dial tcp: lookup api.ainative.studio: no such host
```

**Cause:** DNS resolution failure or incorrect backend URL

**Solution:**
1. Check your backend URL in config
2. Verify internet connectivity
3. Try using IP address instead of hostname

**Config fix:**
```yaml
# Use localhost for development
backend_url: "http://localhost:8000"

# Or use IP address
backend_url: "http://127.0.0.1:8000"
```

---

### Credit and Payment Errors

#### Error: "payment required" (402)

**Full Error:**
```
Error: payment required (402): insufficient credits
```

**Cause:** Insufficient credits in account

**Solution:**
1. Check current balance:
   ```bash
   ainative-code auth whoami
   ```

2. Add credits at https://ainative.studio/billing

3. Verify credits were added and try again

**Temporary workaround:** Use a different account with available credits

---

### Provider Errors

#### Error: "no provider available"

**Full Error:**
```
Error: no provider available for request
```

**Causes:**
- Preferred provider is not configured
- All providers are unavailable
- Provider selection failed

**Solution 1 - Set preferred provider:**
```yaml
ainative:
  preferred_provider: anthropic
```

**Solution 2 - Use auto-provider:**
```bash
ainative-code chat-ainative -m "test" --auto-provider
```

**Solution 3 - Manually specify provider:**
```bash
ainative-code chat-ainative -m "test" --provider anthropic
```

---

#### Error: "provider not found"

**Full Error:**
```
Error: provider not found: invalidprovider
```

**Cause:** Invalid provider name specified

**Valid providers:**
- `anthropic`
- `openai`
- `google`

**Solution:**
```bash
ainative-code chat-ainative -m "test" --provider anthropic
```

---

### Request Errors

#### Error: "message cannot be empty"

**Full Error:**
```
Error: message cannot be empty
```

**Cause:** No message provided to chat command

**Solution:**
```bash
# Correct usage
ainative-code chat-ainative --message "Your question here"

# Or short form
ainative-code chat-ainative -m "Your question here"
```

---

#### Error: "model not found"

**Full Error:**
```
Error: model not found: invalid-model
```

**Cause:** Invalid model name specified

**Valid models by provider:**

**Anthropic:**
- `claude-sonnet-4-5`
- `claude-opus-4`

**OpenAI:**
- `gpt-4`
- `gpt-4-turbo`
- `gpt-3.5-turbo`

**Google:**
- `gemini-pro`
- `gemini-ultra`

**Solution:**
```bash
ainative-code chat-ainative \
  --message "test" \
  --model claude-sonnet-4-5
```

---

### Configuration Errors

#### Error: "config file not found"

**Full Error:**
```
Warning: Could not save config: config file not found
```

**Cause:** Config directory doesn't exist

**Solution:**
```bash
# Create config directory
mkdir -p ~/.config/ainative-code

# Initialize config
ainative-code config init
```

---

#### Error: "invalid YAML"

**Full Error:**
```
Error: error unmarshaling YAML: yaml: line 5: did not find expected key
```

**Cause:** Syntax error in config.yaml

**Solution:**
1. Check YAML syntax (indentation, colons, quotes)
2. Validate YAML at https://www.yamllint.com/
3. Restore from backup or recreate config

**Valid YAML example:**
```yaml
backend_url: "http://localhost:8000"
ainative:
  preferred_provider: anthropic
  fallback_enabled: true
```

---

## Debugging Steps

### Enable Debug Mode

Set environment variable for verbose logging:

```bash
export DEBUG=1
ainative-code chat-ainative -m "test" --verbose
```

### Check Configuration

View current configuration:

```bash
# View all config
cat ~/.config/ainative-code/config.yaml

# View specific setting
ainative-code config get ainative.preferred_provider
```

### Test Backend Connection

Verify backend is accessible:

```bash
# Health check
curl http://localhost:8000/health

# Test login endpoint
curl -X POST http://localhost:8000/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"test"}'
```

### Check Logs

Backend logs provide detailed error information:

```bash
# If running with uvicorn
# Logs appear in the terminal where backend is running

# Check for errors
grep -i error /path/to/backend/logs/*.log
```

### Verify Token Status

Check if tokens are stored and valid:

```bash
# View authentication status
ainative-code auth whoami

# View token expiration
ainative-code auth token status
```

---

## Advanced Troubleshooting

### Clear All State

If you encounter persistent issues, clear all state:

```bash
# Logout (clears tokens)
ainative-code auth logout-backend

# Remove config (backup first!)
mv ~/.config/ainative-code/config.yaml ~/.config/ainative-code/config.yaml.backup

# Reinitialize
ainative-code config init

# Login again
ainative-code auth login-backend -e your@email.com -p password
```

### Network Debugging

Capture network traffic:

```bash
# Use tcpdump (macOS/Linux)
sudo tcpdump -i lo0 port 8000 -A

# Or use mitmproxy
mitmproxy --mode reverse:http://localhost:8000

# Then set proxy
export HTTP_PROXY=http://localhost:8080
```

### Backend Health Check

Comprehensive backend health check:

```bash
# Check if backend is running
curl -f http://localhost:8000/health || echo "Backend not responding"

# Check auth endpoint
curl -X POST http://localhost:8000/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","password":"test"}' \
  2>&1 | head -20

# Check chat endpoint
curl -X POST http://localhost:8000/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"messages":[{"role":"user","content":"test"}]}'
```

---

## Getting Help

### Self-Service Resources

1. **Documentation**: https://github.com/AINative-Studio/ainative-code/docs
2. **API Reference**: [API Documentation](../api/ainative-provider.md)
3. **Examples**: [Code Examples](../examples/)

### Community Support

1. **GitHub Issues**: https://github.com/AINative-Studio/ainative-code/issues
   - Search existing issues first
   - Provide verbose output and logs
   - Include system information

2. **GitHub Discussions**: https://github.com/AINative-Studio/ainative-code/discussions
   - Ask questions
   - Share solutions
   - Request features

### Direct Support

**Email:** support@ainative.studio

**Include in support requests:**
- CLI version (`ainative-code version`)
- Operating system
- Backend version
- Full error message
- Steps to reproduce
- Output from `--verbose` mode

---

## Issue Reporting Template

When reporting issues, include:

```markdown
**Environment:**
- OS: macOS 14.0
- CLI Version: v1.0.0
- Backend Version: v1.0.0

**Command:**
```bash
ainative-code chat-ainative -m "test" --verbose
```

**Expected Behavior:**
Receive response from AI

**Actual Behavior:**
Error: connection refused

**Logs/Output:**
```
[paste verbose output here]
```

**Steps to Reproduce:**
1. Start backend
2. Login with auth login-backend
3. Run chat-ainative command
4. See error

**Additional Context:**
Backend is running on custom port 9000
```

---

## Common Solutions Summary

| Issue | Quick Fix |
|-------|-----------|
| Not authenticated | `ainative-code auth login-backend -e email -p pass` |
| Token expired | `ainative-code auth refresh-backend` |
| Backend not running | `cd python-backend && uvicorn app.main:app --reload` |
| Connection refused | Check backend is on port 8000 |
| Invalid provider | Use `--provider anthropic` or `--auto-provider` |
| Timeout | Increase timeout in config or use faster model |
| Config not loading | Check `~/.config/ainative-code/config.yaml` syntax |
| Message empty | Use `-m "your message"` or `--message "text"` |

---

## Next Steps

- [Getting Started Guide](ainative-getting-started.md) - Setup instructions
- [Authentication Guide](authentication.md) - Manage credentials
- [Hosted Inference Guide](hosted-inference.md) - Use chat features
- [Provider Configuration Guide](provider-configuration.md) - Configure providers
