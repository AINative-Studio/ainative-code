# Security Best Practices

## Overview

This document provides security best practices for authentication in AINative Code. Following these guidelines will help protect your account, tokens, and data from unauthorized access.

## Table of Contents

- [Token Security](#token-security)
- [Secure Storage](#secure-storage)
- [Network Security](#network-security)
- [Session Management](#session-management)
- [API Key Protection](#api-key-protection)
- [Audit and Monitoring](#audit-and-monitoring)
- [Multi-Factor Authentication](#multi-factor-authentication)
- [Development vs Production](#development-vs-production)
- [Incident Response](#incident-response)

---

## Token Security

### Never Share Tokens

**Why**: Tokens grant full access to your account and resources.

**Do**:
- ✓ Keep tokens private and personal
- ✓ Let CLI manage tokens automatically
- ✓ Use separate tokens per device/environment
- ✓ Revoke tokens when compromised

**Don't**:
- ✗ Share tokens with colleagues (use separate accounts)
- ✗ Send tokens via email, chat, or unencrypted channels
- ✗ Commit tokens to version control
- ✗ Log tokens in plaintext
- ✗ Include tokens in screenshots or documentation

### Token Rotation

**Automatic Rotation**: Access tokens expire after 1 hour and are automatically refreshed.

**Manual Rotation** (when needed):

```bash
# Force new tokens
ainative-code auth logout
ainative-code auth login
```

**When to rotate manually**:
- Suspect token compromise
- Employee departure
- Device lost or stolen
- Security policy requirement

### Token Expiration Settings

**Recommended Configuration**:

```yaml
# ~/.config/ainative-code/config.yaml
platform:
  keychain:
    # Auto-refresh before expiration
    auto_refresh: true
    refresh_threshold: 5m

  jwt:
    # Validate token on each use
    validate_on_use: true

    # Maximum token age before forced re-auth
    max_token_age: 24h
```

**Shorter lifetimes** = More secure but more frequent logins
**Longer lifetimes** = More convenient but higher risk if compromised

### Prevent Token Leakage

**1. Sanitize Logs**:

```go
// Bad - logs full token
log.Printf("Using token: %s", token)

// Good - logs only reference
log.Printf("Using token: %s...", token[:10])

// Better - logs nothing about token
log.Printf("Authenticated request to API")
```

**2. Exclude from Version Control**:

```gitignore
# .gitignore
.config/ainative-code/
.ainative/
*.token
tokens.json
credentials.*
```

**3. Clear Terminal History**:

```bash
# Don't include tokens in commands
# Bad
export ACCESS_TOKEN="eyJhbG..."

# Good - use keychain
ainative-code auth login

# Clear history if needed
history -c
```

---

## Secure Storage

### OS Keychain Usage

**Preferred**: Always use OS keychain for token storage.

**Platform-Specific Security**:

**macOS Keychain**:
```bash
# Verify keychain is locked when idle
security show-keychain-info ~/Library/Keychains/login.keychain-db

# Set auto-lock timeout (5 minutes of inactivity)
security set-keychain-settings -t 300 ~/Library/Keychains/login.keychain-db

# Require password after sleep
security set-keychain-settings -l ~/Library/Keychains/login.keychain-db
```

**Linux Secret Service**:
```bash
# Ensure keyring locks on logout
# Edit /etc/pam.d/login to include:
# session optional pam_gnome_keyring.so auto_start

# Lock keyring manually
gnome-keyring-daemon --lock
```

**Windows Credential Manager**:
```powershell
# Enable additional encryption
# Group Policy: Computer Configuration → Windows Settings →
# Security Settings → Local Policies → Security Options
# "System cryptography: Use FIPS compliant algorithms" → Enabled
```

### File Permissions

If using file-based storage (not recommended):

```bash
# Create config directory with restricted permissions
mkdir -p ~/.config/ainative-code
chmod 700 ~/.config/ainative-code

# Ensure token file is only readable by owner
touch ~/.config/ainative-code/tokens.enc
chmod 600 ~/.config/ainative-code/tokens.enc

# Verify permissions
ls -la ~/.config/ainative-code/
# Should show: drwx------ (700 for directory)
# Should show: -rw------- (600 for files)
```

### Encryption at Rest

**Configuration for encrypted file storage**:

```yaml
# ~/.config/ainative-code/config.yaml
platform:
  keychain:
    backend: "file"
    file_path: "~/.config/ainative-code/tokens.enc"

    # Encryption settings
    encryption:
      algorithm: "aes-256-gcm"
      key_derivation: "pbkdf2"
      iterations: 100000

      # Password from environment or keyring
      password_source: "env:AINATIVE_KEYRING_PASSWORD"
```

**Set encryption password**:

```bash
# Generate strong password
openssl rand -base64 32 > ~/.ainative/keyring.password
chmod 400 ~/.ainative/keyring.password

# Set environment variable
export AINATIVE_KEYRING_PASSWORD=$(cat ~/.ainative/keyring.password)

# Add to shell profile
echo 'export AINATIVE_KEYRING_PASSWORD=$(cat ~/.ainative/keyring.password)' >> ~/.bashrc
```

---

## Network Security

### HTTPS Everywhere

**Always use HTTPS** for authentication endpoints:

```yaml
# ~/.config/ainative-code/config.yaml
platform:
  oauth:
    # ✓ Good - HTTPS
    auth_endpoint: "https://auth.ainative.studio/oauth/authorize"
    token_endpoint: "https://auth.ainative.studio/oauth/token"

    # ✗ Bad - HTTP (will be rejected)
    # auth_endpoint: "http://auth.ainative.studio/oauth/authorize"
```

### Certificate Validation

**Do NOT disable certificate validation** in production:

```yaml
# Development only (if using self-signed certs)
http:
  insecure_skip_verify: false  # Keep false in production!

# Instead, add custom CA certificate
tls:
  ca_cert_path: "/path/to/corporate-ca.crt"
```

### Proxy Security

**If using corporate proxy**:

```yaml
http:
  proxy: "http://proxy.company.com:8080"

  # Exclude sensitive endpoints from proxy
  no_proxy:
    - "auth.ainative.studio"
    - "localhost"
    - "127.0.0.1"

  # Proxy authentication (if required)
  proxy_auth:
    username: "${PROXY_USER}"
    password: "${PROXY_PASS}"
```

**Avoid MITM attacks**:

```bash
# Verify SSL certificate fingerprint
openssl s_client -connect auth.ainative.studio:443 < /dev/null 2>/dev/null | \
  openssl x509 -fingerprint -noout

# Expected: SHA256 Fingerprint=XX:XX:XX:...
```

### Network Isolation

**Restrict callback server binding**:

```yaml
platform:
  oauth:
    # Only listen on localhost (not all interfaces)
    callback_host: "127.0.0.1"  # Not "0.0.0.0"
    callback_port: 8080
```

**Firewall rules**:

```bash
# macOS
sudo /usr/libexec/ApplicationFirewall/socketfilterfw --add /usr/local/bin/ainative-code
sudo /usr/libexec/ApplicationFirewall/socketfilterfw --unblockapp /usr/local/bin/ainative-code

# Linux iptables - Allow only localhost
sudo iptables -A INPUT -i lo -p tcp --dport 8080 -j ACCEPT
sudo iptables -A INPUT -p tcp --dport 8080 -j DROP

# UFW
sudo ufw allow from 127.0.0.1 to any port 8080
```

---

## Session Management

### Session Timeout Configuration

**Recommended Settings**:

```yaml
platform:
  session:
    # Access token lifetime (server-controlled)
    access_token_ttl: 1h

    # Refresh token lifetime (server-controlled)
    refresh_token_ttl: 30d

    # Maximum session duration before forced re-auth
    max_session_duration: 7d

    # Idle timeout - logout after inactivity
    idle_timeout: 4h

    # Session validation interval
    validation_interval: 15m
```

### Automatic Logout

**Implement idle timeout**:

```bash
# Add to shell profile
export AINATIVE_IDLE_TIMEOUT=4h

# Or in config
```

```yaml
platform:
  session:
    idle_timeout: 4h

    # Actions on idle timeout
    idle_action: "logout"  # or "refresh"
```

### Session Monitoring

**Track active sessions**:

```bash
# View current session info
ainative-code auth whoami

# Check token status
ainative-code auth token status

# View session history (server-side)
# Visit: https://ainative.studio/settings/sessions
```

### Multi-Device Sessions

**Best Practice**: Separate session per device

**Benefits**:
- Individual device revocation
- Better audit trails
- Isolated security breaches

**Implementation**:

```bash
# Device 1 (Work Laptop)
device1$ ainative-code auth login
# Session ID: sess-work-laptop-123

# Device 2 (Home Desktop)
device2$ ainative-code auth login
# Session ID: sess-home-desktop-456

# Revoke specific device
# Via web interface: https://ainative.studio/settings/sessions
# Select session → Revoke
```

---

## API Key Protection

### LLM Provider API Keys

**Secure storage for API keys**:

```yaml
# ~/.config/ainative-code/config.yaml
providers:
  anthropic:
    # Bad - plaintext in config
    # api_key: "sk-ant-api03-..."

    # Good - reference to secure storage
    api_key: "$(security find-generic-password -s anthropic-api-key -w)"

    # Or environment variable
    # api_key: "${ANTHROPIC_API_KEY}"
```

**Environment Variables**:

```bash
# Set in shell profile (better than config file)
export ANTHROPIC_API_KEY="sk-ant-api03-..."
export OPENAI_API_KEY="sk-proj-..."

# Or use password manager integration
export ANTHROPIC_API_KEY="$(pass show anthropic/api-key)"
export OPENAI_API_KEY="$(1password read "op://Personal/OpenAI/api-key")"
```

### API Key Rotation

**Regular rotation** (every 90 days):

```bash
# 1. Generate new key in provider dashboard
# 2. Update config or environment
export ANTHROPIC_API_KEY="sk-ant-api03-NEW-KEY"

# 3. Verify new key works
ainative-code chat "test message"

# 4. Revoke old key in provider dashboard
```

**Automate rotation**:

```bash
#!/bin/bash
# api-key-rotation.sh

# Check key age
KEY_DATE=$(security find-generic-password -s anthropic-api-key -g 2>&1 | grep "mdat" | cut -d'=' -f2)
DAYS_OLD=$(( ($(date +%s) - $(date -j -f "%Y%m%d%H%M%S" "$KEY_DATE" +%s)) / 86400 ))

if [ $DAYS_OLD -gt 90 ]; then
    echo "API key is $DAYS_OLD days old - rotation recommended"
    # Trigger rotation workflow
fi
```

### Scope Limitation

**Request minimum required scopes**:

```bash
# Bad - requesting all scopes
ainative-code auth login --scopes read,write,admin,delete

# Good - only what's needed
ainative-code auth login --scopes read,write
```

**Scope Mapping**:

| Scope | Permissions | Use Case |
|-------|-------------|----------|
| `read` | Read-only access | Querying data, viewing resources |
| `write` | Create/update resources | Generating code, uploading designs |
| `admin` | Administrative operations | User management, configuration |
| `delete` | Delete resources | Cleanup operations |
| `offline_access` | Refresh token | Long-running sessions |

---

## Audit and Monitoring

### Enable Audit Logging

**Configuration**:

```yaml
logging:
  level: info

  # Enable audit trail
  audit:
    enabled: true
    file: "~/.config/ainative-code/audit.log"

    # Log authentication events
    events:
      - "auth.login"
      - "auth.logout"
      - "auth.token.refresh"
      - "auth.token.revoke"
      - "auth.session.expired"
      - "auth.error"

    # Include details
    include:
      - timestamp
      - user_id
      - ip_address
      - user_agent
      - result
```

**Example Audit Log**:

```json
{
  "timestamp": "2025-01-05T10:30:00Z",
  "event": "auth.login",
  "user_id": "user-123",
  "email": "user@example.com",
  "ip_address": "192.168.1.100",
  "user_agent": "ainative-code/1.0.0",
  "result": "success"
}
```

### Monitor Suspicious Activity

**Watch for**:
- Failed login attempts (>3 in 5 minutes)
- Logins from new locations/devices
- Token refresh from multiple IPs
- Unusual API usage patterns

**Alerting**:

```yaml
alerts:
  # Email notification
  email: "security@example.com"

  # Triggers
  triggers:
    - event: "auth.failed_login"
      threshold: 3
      window: 5m

    - event: "auth.new_device"
      action: "notify"

    - event: "auth.unusual_location"
      action: "notify_and_lockout"
```

### Review Session Activity

**Regular audits**:

```bash
# Weekly session review
ainative-code auth sessions list

# Check for unknown devices
ainative-code auth sessions list --format json | \
  jq '.[] | select(.device_name | test("unknown"))'

# Revoke old sessions
ainative-code auth sessions prune --older-than 30d
```

---

## Multi-Factor Authentication

### Enable MFA (When Available)

**Server-Side MFA**:

1. Visit https://ainative.studio/settings/security
2. Enable two-factor authentication
3. Scan QR code with authenticator app
4. Save backup codes

**MFA with CLI**:

```bash
# Login will prompt for MFA code
ainative-code auth login
# Email: user@example.com
# Password: ********
# MFA Code: 123456
# ✓ Authentication successful!
```

### Backup Codes

**Store backup codes securely**:

```bash
# Print backup codes
ainative-code auth mfa backup-codes

# Save to encrypted file
ainative-code auth mfa backup-codes > backup-codes.txt
gpg --encrypt --recipient your@email.com backup-codes.txt
shred -u backup-codes.txt

# Store in password manager
# 1Password, LastPass, Bitwarden, etc.
```

### Recovery Procedures

**If device lost**:

1. Use backup code to login
2. Disable MFA
3. Re-enable MFA with new device
4. Generate new backup codes

```bash
# Login with backup code
ainative-code auth login --mfa-backup-code XXXXXXXX

# Reset MFA
ainative-code auth mfa reset

# Re-enable on new device
ainative-code auth mfa enable
```

---

## Development vs Production

### Separate Environments

**Configuration**:

```yaml
# ~/.config/ainative-code/config.dev.yaml (development)
platform:
  oauth:
    auth_endpoint: "https://auth.dev.ainative.studio/oauth/authorize"
    token_endpoint: "https://auth.dev.ainative.studio/oauth/token"

  environment: "development"

# ~/.config/ainative-code/config.prod.yaml (production)
platform:
  oauth:
    auth_endpoint: "https://auth.ainative.studio/oauth/authorize"
    token_endpoint: "https://auth.ainative.studio/oauth/token"

  environment: "production"
```

**Usage**:

```bash
# Development
export AINATIVE_ENV=development
ainative-code auth login

# Production
export AINATIVE_ENV=production
ainative-code auth login

# Separate credentials
# Dev credentials stored in separate keychain
```

### Development Best Practices

**1. Never Use Production Credentials in Dev**:

```bash
# Bad
export ANTHROPIC_API_KEY="sk-ant-api03-PRODUCTION-KEY"

# Good - use separate dev API key
export ANTHROPIC_API_KEY="sk-ant-api03-DEV-KEY"
```

**2. Use Mock Authentication in Tests**:

```go
// test.go
func TestWithMockAuth(t *testing.T) {
    // Don't use real OAuth in tests
    mockAuth := &MockAuthClient{
        AccessToken: "mock-token",
    }

    // Test with mock
    client := NewAPIClient(mockAuth)
    // ...
}
```

**3. Disable Sensitive Features in Dev**:

```yaml
# development config
features:
  audit_logging: false  # Don't clutter logs in dev
  mfa_required: false   # Ease testing
  token_rotation: true  # Still test rotation
```

---

## Incident Response

### Token Compromise Response

**If token compromised**:

**Immediate Actions** (< 5 minutes):

```bash
# 1. Revoke all tokens
ainative-code auth logout --all-devices

# 2. Change password
# Visit: https://ainative.studio/settings/security

# 3. Login with new credentials
ainative-code auth login
```

**Follow-up Actions** (< 24 hours):

```bash
# 4. Review audit logs
ainative-code auth audit --since 7d --format json > audit.json

# 5. Check for unauthorized activity
ainative-code audit analyze --file audit.json

# 6. Rotate API keys
# Update all LLM provider API keys

# 7. Enable MFA (if not already)
ainative-code auth mfa enable

# 8. Notify security team
# Send incident report with audit logs
```

### Breach Notification

**When to notify**:
- Confirmed unauthorized access
- Data exfiltration suspected
- Credentials leaked publicly
- Multiple failed compromise attempts

**How to report**:

```bash
# Generate incident report
ainative-code security incident-report \
  --type "token-compromise" \
  --severity "high" \
  --description "Access token found in public GitHub repo" \
  --evidence audit.json

# Submit to security team
# Email: security@ainative.studio
# Include: incident report, audit logs, timeline
```

### Recovery Checklist

- [ ] Revoke compromised tokens
- [ ] Change password
- [ ] Enable MFA
- [ ] Review audit logs
- [ ] Rotate API keys
- [ ] Check for unauthorized access
- [ ] Update security configuration
- [ ] Notify security team
- [ ] Document incident
- [ ] Implement preventive measures

---

## Security Checklist

### Daily
- [ ] Verify token status before important operations
- [ ] Review failed login attempts in logs

### Weekly
- [ ] Check active sessions
- [ ] Review audit logs for anomalies
- [ ] Verify MFA is enabled

### Monthly
- [ ] Rotate API keys
- [ ] Review and revoke unused sessions
- [ ] Update security configuration
- [ ] Test backup and recovery procedures

### Quarterly
- [ ] Comprehensive security audit
- [ ] Update authentication documentation
- [ ] Review and update access policies
- [ ] Security training/awareness

---

## Compliance Considerations

### Data Protection

**GDPR/Privacy Compliance**:

```yaml
privacy:
  # Minimize data collection
  collect_only_necessary: true

  # User data retention
  token_retention: 30d
  audit_log_retention: 90d

  # Data deletion
  allow_user_deletion: true

  # Consent management
  require_explicit_consent: true
```

### SOC 2 / ISO 27001

**Required Controls**:

- Encryption at rest and in transit
- Access logging and monitoring
- Regular security audits
- Incident response procedures
- Access control and authorization
- Secure credential management

**Implementation**:

```yaml
compliance:
  soc2:
    enabled: true

    controls:
      - encryption_at_rest
      - encryption_in_transit
      - access_logging
      - regular_audits
      - incident_response

  iso27001:
    enabled: true

    policies:
      - access_control_policy
      - cryptography_policy
      - incident_management_policy
```

---

## Additional Resources

- [Authentication Overview](README.md)
- [OAuth Flow](oauth-flow.md)
- [Troubleshooting](troubleshooting.md)
- [API Reference](api-reference.md)

### External Resources

- [OWASP Authentication Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html)
- [NIST Digital Identity Guidelines](https://pages.nist.gov/800-63-3/)
- [OAuth 2.0 Security Best Current Practice](https://datatracker.ietf.org/doc/html/draft-ietf-oauth-security-topics)
- [JWT Best Practices](https://datatracker.ietf.org/doc/html/rfc8725)

---

**Remember**: Security is a continuous process. Regularly review and update your security practices to protect against evolving threats.
