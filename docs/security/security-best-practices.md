# Security Best Practices Guide - AINative-Code

**Version:** 1.0
**Last Updated:** January 4, 2026
**Audience:** Developers, DevOps, Security Engineers

## Table of Contents

1. [Authentication and Authorization](#authentication-and-authorization)
2. [API Key and Secret Management](#api-key-and-secret-management)
3. [Input Validation and Sanitization](#input-validation-and-sanitization)
4. [Database Security](#database-security)
5. [Tool Execution Security](#tool-execution-security)
6. [Network Security](#network-security)
7. [Logging and Monitoring](#logging-and-monitoring)
8. [Dependency Management](#dependency-management)
9. [Deployment Security](#deployment-security)
10. [Incident Response](#incident-response)

---

## Authentication and Authorization

### JWT Token Handling

**DO:**
- ✅ Use RS256 (RSA with SHA-256) for JWT signing
- ✅ Validate ALL claims: `iss`, `aud`, `sub`, `exp`
- ✅ Use separate access and refresh tokens
- ✅ Bind refresh tokens to sessions
- ✅ Set appropriate expiration times (access: 15min, refresh: 7 days)
- ✅ Validate tokens on every request

**DON'T:**
- ❌ Use HS256 (symmetric) algorithms in production
- ❌ Skip expiration validation
- ❌ Store tokens in localStorage (use httpOnly cookies)
- ❌ Accept tokens without signature verification
- ❌ Log token contents

**Example Implementation:**
```go
// Good: Comprehensive JWT validation
token, err := auth.ParseAccessToken(tokenString, publicKey)
if err != nil {
    return fmt.Errorf("invalid access token: %w", err)
}

// Validate additional business logic
if !token.HasRole("admin") {
    return errors.New("insufficient permissions")
}
```

### Session Management

**Best Practices:**
1. **Generate secure session IDs:** Use crypto/rand, minimum 128 bits
2. **Rotate session IDs:** After login, privilege escalation
3. **Implement session timeouts:** Idle: 30min, absolute: 24hr
4. **Secure session storage:** Redis with encryption, or encrypted cookies
5. **Session invalidation:** On logout, password change, suspicious activity

**Example:**
```go
// Generate secure session ID
import "crypto/rand"

func GenerateSessionID() (string, error) {
    b := make([]byte, 32) // 256 bits
    if _, err := rand.Read(b); err != nil {
        return "", err
    }
    return base64.URLEncoding.EncodeToString(b), nil
}
```

### Password Security (Local Auth)

**Requirements:**
- ✅ Use bcrypt with cost factor ≥ 12
- ✅ Enforce minimum 12 characters
- ✅ Require mix of character types
- ✅ Check against common password lists
- ✅ Implement rate limiting on login attempts
- ✅ Never log or store passwords in plaintext

**Example:**
```go
import "golang.org/x/crypto/bcrypt"

// Hash password before storage
func HashPassword(password string) (string, error) {
    hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
    if err != nil {
        return "", err
    }
    return string(hash), nil
}

// Verify password during authentication
func VerifyPassword(hash, password string) error {
    return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
```

---

## API Key and Secret Management

### Storage Locations

**Priority Order (most to least secure):**

1. **System Keychain** (Recommended for development)
   - macOS: Keychain Access
   - Windows: Credential Manager
   - Linux: GNOME Keyring / KWallet

2. **External Secret Managers** (Recommended for production)
   - AWS Secrets Manager
   - HashiCorp Vault
   - Google Secret Manager
   - Azure Key Vault

3. **Environment Variables** (Acceptable for containers)
   - Kubernetes Secrets
   - Docker Secrets
   - `.env` files (never commit to git)

4. **Encrypted Configuration Files** (Last resort)
   - Use AES-256-GCM
   - Store encryption key separately
   - Rotate keys regularly

**Example Configuration:**
```yaml
llm:
  anthropic:
    # GOOD: Environment variable
    api_key: ${ANTHROPIC_API_KEY}

    # GOOD: Password manager (1Password)
    api_key: $(op read op://Private/Anthropic/api_key)

    # GOOD: AWS Secrets Manager
    api_key: $(aws secretsmanager get-secret-value --secret-id anthropic-key --query SecretString --output text)

    # BAD: Hardcoded (NEVER DO THIS)
    # api_key: sk-ant-api-xxxxx
```

### Secret Rotation

**Rotation Schedule:**
- **API Keys:** Every 90 days
- **Database Passwords:** Every 90 days
- **JWT Signing Keys:** Every 180 days
- **Encryption Keys:** Every 365 days
- **Immediately:** On suspected compromise

**Rotation Process:**
1. Generate new secret
2. Deploy new secret to all systems
3. Update application configuration
4. Grace period (both old and new work)
5. Revoke old secret
6. Verify no failures
7. Document in audit log

### Secret Scanning

**Pre-commit Hook (recommended):**
```bash
# .git/hooks/pre-commit
#!/bin/bash
gitleaks protect --staged --verbose

if [ $? -ne 0 ]; then
    echo "❌ Secrets detected! Commit blocked."
    exit 1
fi
```

**.gitleaksignore:**
```
# Ignore test fixtures
tests/**/*:generic-api-key
tests/fixtures/**/*:*
.ainative/settings.local.json:jwt

# Document exceptions
# tests/e2e/error_recovery_test.go contains intentional invalid API keys for testing
```

---

## Input Validation and Sanitization

### Validation Principles

**Always Validate:**
1. **Type:** Ensure correct Go type
2. **Format:** Regex patterns, email, URL, UUID
3. **Range:** Min/max length, numeric bounds
4. **Whitelist:** Allowed values, characters
5. **Business Logic:** Cross-field validation

**Example Tool Schema:**
```go
func (t *MyTool) Schema() tools.ToolSchema {
    maxLength := 1000
    minLength := 1

    return tools.ToolSchema{
        Type: "object",
        Properties: map[string]tools.PropertySchema{
            "email": {
                Type:        "string",
                Description: "User email address",
                Pattern:     `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
                MaxLength:   &maxLength,
            },
            "age": {
                Type:        "integer",
                Description: "User age",
                Minimum:     18,
                Maximum:     120,
            },
            "role": {
                Type:        "string",
                Description: "User role",
                Enum:        []interface{}{"admin", "user", "guest"},
            },
        },
        Required: []string{"email", "role"},
    }
}
```

### Path Traversal Prevention

**Always Sanitize File Paths:**
```go
func SanitizePath(userPath string) (string, error) {
    // Clean the path
    cleaned := filepath.Clean(userPath)

    // Prevent absolute paths (if not allowed)
    if filepath.IsAbs(cleaned) {
        return "", errors.New("absolute paths not allowed")
    }

    // Prevent path traversal
    if strings.Contains(cleaned, "..") {
        return "", errors.New("path traversal detected")
    }

    // Ensure path is within allowed directory
    allowedDir := "/var/app/uploads"
    fullPath := filepath.Join(allowedDir, cleaned)
    if !strings.HasPrefix(fullPath, allowedDir) {
        return "", errors.New("path outside allowed directory")
    }

    return fullPath, nil
}
```

### Command Injection Prevention

**Whitelist Approach (Recommended):**
```go
var allowedCommands = map[string]bool{
    "git":    true,
    "npm":    true,
    "go":     true,
    "docker": true,
}

func ExecuteCommand(cmd string, args []string) error {
    if !allowedCommands[cmd] {
        return fmt.Errorf("command %s not allowed", cmd)
    }

    // Use CommandContext for timeout protection
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    command := exec.CommandContext(ctx, cmd, args...)
    return command.Run()
}
```

### SQL Injection Prevention

**Always use parameterized queries:**
```go
// GOOD: Parameterized query
func GetUserByEmail(email string) (*User, error) {
    var user User
    err := db.QueryRow(
        "SELECT id, email, name FROM users WHERE email = ?",
        email, // Parameter binding
    ).Scan(&user.ID, &user.Email, &user.Name)
    return &user, err
}

// BAD: String concatenation (NEVER DO THIS)
func GetUserByEmailBad(email string) (*User, error) {
    query := "SELECT * FROM users WHERE email = '" + email + "'"
    // ❌ Vulnerable to: email = "x' OR '1'='1"
    return db.Query(query)
}
```

---

## Database Security

### Connection Security

**Best Practices:**
1. **Use TLS for database connections**
2. **Limit connection privileges** (principle of least privilege)
3. **Use connection pooling with limits**
4. **Enable query logging** (without sensitive data)
5. **Regular backups with encryption**

**Example Configuration:**
```go
config := &database.ConnectionConfig{
    Driver:       "sqlite3",
    DSN:          "file:ainative.db?_journal=WAL&_timeout=5000",
    MaxOpenConns: 25,
    MaxIdleConns: 5,
    MaxLifetime:  5 * time.Minute,
    // For PostgreSQL/MySQL, enable TLS:
    // TLS: &tls.Config{MinVersion: tls.VersionTLS12},
}
```

### Data Encryption

**Encryption at Rest:**
```go
import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
)

func EncryptSensitiveData(data []byte, key []byte) ([]byte, error) {
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }

    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }

    nonce := make([]byte, gcm.NonceSize())
    if _, err := rand.Read(nonce); err != nil {
        return nil, err
    }

    ciphertext := gcm.Seal(nonce, nonce, data, nil)
    return ciphertext, nil
}
```

### Backup Security

**Backup Checklist:**
- [ ] Encrypted backups (AES-256)
- [ ] Secure backup storage (separate from production)
- [ ] Regular backup testing
- [ ] Access logging for backups
- [ ] Retention policy (e.g., 30 days)
- [ ] Offsite backup copies

---

## Tool Execution Security

### Sandboxing Principles

**Implementation:**
```go
type ExecCommandTool struct {
    allowedCommands []string // Whitelist
    workingDir      string   // Restrict to specific directory
    maxTimeout      time.Duration
    maxOutputSize   int64
}

func (t *ExecCommandTool) Execute(cmd string, args []string) error {
    // 1. Validate command against whitelist
    if !t.isAllowed(cmd) {
        return errors.New("command not allowed")
    }

    // 2. Set timeout
    ctx, cancel := context.WithTimeout(context.Background(), t.maxTimeout)
    defer cancel()

    // 3. Create command in restricted environment
    command := exec.CommandContext(ctx, cmd, args...)
    command.Dir = t.workingDir

    // 4. Limit output size
    var out bytes.Buffer
    command.Stdout = io.LimitReader(&out, t.maxOutputSize)

    // 5. Execute
    return command.Run()
}
```

### Resource Limits

**Enforce Limits:**
- **Timeout:** Maximum execution time
- **Memory:** Process memory limit (via cgroups or ulimit)
- **File Size:** Maximum file read/write size
- **Network:** Bandwidth limits for network operations
- **CPU:** CPU time limits

**Example:**
```go
const (
    MaxCommandTimeout   = 5 * time.Minute
    MaxFileSize        = 100 * 1024 * 1024 // 100MB
    MaxOutputSize      = 10 * 1024 * 1024  // 10MB
)
```

---

## Network Security

### TLS/HTTPS Configuration

**Minimum TLS Configuration:**
```go
import (
    "crypto/tls"
    "net/http"
)

func NewSecureServer() *http.Server {
    return &http.Server{
        Addr:    ":443",
        Handler: router,

        // TLS configuration
        TLSConfig: &tls.Config{
            MinVersion: tls.VersionTLS12,
            CipherSuites: []uint16{
                tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
                tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
            },
            PreferServerCipherSuites: true,
        },

        // Timeouts (Slowloris protection)
        ReadTimeout:       5 * time.Second,
        ReadHeaderTimeout: 5 * time.Second,
        WriteTimeout:      10 * time.Second,
        IdleTimeout:       120 * time.Second,
    }
}
```

### HTTPS Enforcement Middleware

```go
func HTTPSRedirectMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.TLS == nil && r.Header.Get("X-Forwarded-Proto") != "https" {
            target := "https://" + r.Host + r.URL.Path
            if r.URL.RawQuery != "" {
                target += "?" + r.URL.RawQuery
            }
            http.Redirect(w, r, target, http.StatusMovedPermanently)
            return
        }

        // Add HSTS header
        w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
        next.ServeHTTP(w, r)
    })
}
```

### Rate Limiting

**Implementation:**
```go
func RateLimitMiddleware(limiter *ratelimit.Limiter) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Get user identifier (IP, API key, user ID)
            identifier := getUserIdentifier(r)

            // Check rate limit
            allowed, remaining, resetTime := limiter.Allow(identifier)

            // Add rate limit headers
            w.Header().Set("X-RateLimit-Limit", "60")
            w.Header().Set("X-RateLimit-Remaining", fmt.Sprint(remaining))
            w.Header().Set("X-RateLimit-Reset", fmt.Sprint(resetTime.Unix()))

            if !allowed {
                w.WriteHeader(http.StatusTooManyRequests)
                json.NewEncoder(w).Encode(map[string]string{
                    "error": "rate limit exceeded",
                    "retry_after": resetTime.Format(time.RFC3339),
                })
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}
```

### CORS Configuration

```go
func CORSMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Allow specific origins (never use "*" in production)
        allowedOrigins := []string{
            "https://ainative.studio",
            "https://app.ainative.studio",
        }

        origin := r.Header.Get("Origin")
        if contains(allowedOrigins, origin) {
            w.Header().Set("Access-Control-Allow-Origin", origin)
            w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
            w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
            w.Header().Set("Access-Control-Max-Age", "86400")
        }

        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }

        next.ServeHTTP(w, r)
    })
}
```

---

## Logging and Monitoring

### Secure Logging Practices

**DO:**
- ✅ Log security events (authentication, authorization failures)
- ✅ Include context (user ID, IP, timestamp, action)
- ✅ Use structured logging (JSON format)
- ✅ Log to centralized system (ELK, Splunk, CloudWatch)
- ✅ Set up alerts for suspicious patterns

**DON'T:**
- ❌ Log passwords, API keys, tokens
- ❌ Log full credit card numbers, SSNs
- ❌ Log session IDs
- ❌ Log PII without redaction
- ❌ Log to world-readable files

**Example:**
```go
import "github.com/rs/zerolog/log"

// GOOD: Security event logging
log.Info().
    Str("user_id", userID).
    Str("action", "login").
    Str("ip", req.RemoteAddr).
    Str("user_agent", req.UserAgent()).
    Msg("User logged in successfully")

// BAD: Logging sensitive data
log.Info().
    Str("password", password).           // ❌ NEVER log passwords
    Str("api_key", apiKey).               // ❌ NEVER log API keys
    Str("token", token).                  // ❌ NEVER log tokens
    Msg("Authentication attempt")
```

### Security Event Monitoring

**Events to Monitor:**
1. **Failed login attempts** (> 5 in 5 minutes)
2. **Privilege escalation attempts**
3. **Unusual API usage patterns**
4. **SQL injection attempts**
5. **Path traversal attempts**
6. **Rate limit violations**
7. **Geographic anomalies**
8. **Suspicious file access patterns**

**Example Alert Configuration:**
```yaml
alerts:
  - name: "Brute Force Attack"
    condition: "failed_logins > 10 in 5m"
    severity: "high"
    action: "block_ip"

  - name: "SQL Injection Attempt"
    condition: "sql_keywords in request_params"
    severity: "critical"
    action: "alert_security_team"

  - name: "Unusual Geographic Access"
    condition: "country_change and time < 1h"
    severity: "medium"
    action: "require_mfa"
```

---

## Dependency Management

### Dependency Security Checklist

**Regular Tasks:**
- [ ] Weekly: `go mod tidy` to remove unused dependencies
- [ ] Weekly: `go mod verify` to check integrity
- [ ] Monthly: Update dependencies with `go get -u ./...`
- [ ] Monthly: Run `govulncheck ./...`
- [ ] Quarterly: Audit all direct dependencies
- [ ] Immediately: Patch critical vulnerabilities

**Example Workflow:**
```bash
# Weekly security check
go mod verify
go mod tidy
go list -m all | grep -v "indirect"

# Monthly vulnerability scan
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...

# Update dependencies (test thoroughly!)
go get -u ./...
go mod tidy
go test ./...
```

### Vulnerability Response Process

**When vulnerability is discovered:**

1. **Assessment (0-4 hours)**
   - Determine if vulnerability affects our code
   - Check if we use the vulnerable function
   - Assess CVSS score and exploitability

2. **Prioritization**
   - Critical (CVSS 9.0-10.0): Immediate hotfix
   - High (CVSS 7.0-8.9): Fix within 24 hours
   - Medium (CVSS 4.0-6.9): Fix within 1 week
   - Low (CVSS 0.1-3.9): Fix in next release

3. **Remediation**
   - Update to patched version
   - If no patch: implement workaround or remove dependency
   - Test thoroughly
   - Deploy to production

4. **Documentation**
   - Update CHANGELOG
   - Document in security advisory
   - Notify affected users (if applicable)

---

## Deployment Security

### Production Deployment Checklist

**Before First Production Deployment:**
- [ ] TLS/HTTPS enabled and enforced
- [ ] Rate limiting enabled
- [ ] All secrets in secure storage (not env vars in plain text)
- [ ] Database backups configured and tested
- [ ] Monitoring and alerting configured
- [ ] Incident response plan documented
- [ ] Security contact information published
- [ ] WAF configured (if using cloud provider)
- [ ] DDoS protection enabled
- [ ] Security headers configured

**Security Headers:**
```go
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Prevent clickjacking
        w.Header().Set("X-Frame-Options", "DENY")

        // XSS protection
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("X-XSS-Protection", "1; mode=block")

        // HSTS (HTTPS only)
        w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")

        // CSP (Content Security Policy)
        w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; object-src 'none'")

        // Referrer policy
        w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

        // Permissions policy
        w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

        next.ServeHTTP(w, r)
    })
}
```

### Environment-Specific Configuration

**Development:**
```yaml
security:
  tls_enabled: false
  debug: true
  rate_limit:
    enabled: false
```

**Staging:**
```yaml
security:
  tls_enabled: true
  debug: false
  rate_limit:
    enabled: true
    requests_per_minute: 120
```

**Production:**
```yaml
security:
  tls_enabled: true
  tls_min_version: "1.2"
  debug: false
  rate_limit:
    enabled: true
    requests_per_minute: 60
    burst_size: 10
  cors:
    enabled: true
    allowed_origins:
      - "https://ainative.studio"
  hsts_enabled: true
  csrf_protection: true
```

---

## Incident Response

### Security Incident Types

**Classification:**
1. **Data Breach:** Unauthorized access to sensitive data
2. **DoS/DDoS Attack:** Service unavailability
3. **Malware Infection:** Compromised systems
4. **Account Takeover:** Unauthorized access to user accounts
5. **Insider Threat:** Malicious or negligent employee action
6. **Supply Chain Attack:** Compromised dependency or vendor

### Incident Response Plan

**Phase 1: Detection and Analysis (0-1 hour)**
```markdown
1. Alert received from monitoring system
2. Security team notified via PagerDuty/on-call
3. Initial triage:
   - Verify alert is real (not false positive)
   - Determine severity level
   - Identify affected systems
4. Begin incident log (who, what, when, where, why)
5. Escalate to appropriate level based on severity
```

**Phase 2: Containment (1-4 hours)**
```markdown
1. Short-term containment:
   - Isolate affected systems
   - Block malicious IPs/users
   - Revoke compromised credentials
   - Enable additional logging

2. Long-term containment:
   - Apply temporary patches
   - Implement additional monitoring
   - Prepare for recovery phase
```

**Phase 3: Eradication (4-24 hours)**
```markdown
1. Identify and remove root cause
2. Patch vulnerabilities
3. Scan for persistence mechanisms
4. Verify all systems are clean
5. Update security controls
```

**Phase 4: Recovery (24-72 hours)**
```markdown
1. Restore systems from clean backups
2. Re-enable services gradually
3. Monitor for signs of recurring issues
4. Validate functionality
5. Communicate with stakeholders
```

**Phase 5: Post-Incident (1-2 weeks)**
```markdown
1. Complete incident report
2. Lessons learned meeting
3. Update documentation
4. Implement preventive measures
5. Security training for team
```

### Communication Templates

**Internal Alert (Slack/Teams):**
```
🚨 SECURITY INCIDENT - P1
Status: ACTIVE
Affected: API authentication system
Impact: Users unable to login
Lead: @security-lead
War room: #incident-2026-01-04
Updates: Every 30 minutes
```

**Customer Communication (if required):**
```
Subject: Security Incident Notification

Dear Valued Customer,

We are writing to inform you of a security incident that may have affected your account.

What Happened:
[Brief description]

What Information Was Involved:
[Specific data types]

What We're Doing:
[Remediation steps]

What You Should Do:
[Customer actions]

We take security seriously and sincerely apologize for any inconvenience.

Contact: security@ainative.studio
Reference: INC-2026-01-04-001
```

---

## Quick Reference Checklists

### Daily Security Tasks
- [ ] Review security alerts
- [ ] Check failed login attempts
- [ ] Monitor rate limit violations
- [ ] Review unusual API usage

### Weekly Security Tasks
- [ ] Run `go mod verify`
- [ ] Review security logs
- [ ] Check for dependency updates
- [ ] Test backup restoration
- [ ] Review access logs

### Monthly Security Tasks
- [ ] Run `govulncheck`
- [ ] Update dependencies
- [ ] Review and rotate API keys
- [ ] Security awareness training
- [ ] Review and update firewall rules
- [ ] Test incident response plan

### Quarterly Security Tasks
- [ ] Full security audit
- [ ] Penetration testing
- [ ] Update threat model
- [ ] Review all access permissions
- [ ] Update security documentation
- [ ] Disaster recovery drill

---

## Additional Resources

### External Links
- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [CWE Top 25](https://cwe.mitre.org/top25/)
- [Go Security Best Practices](https://go.dev/doc/security/best-practices)
- [NIST Cybersecurity Framework](https://www.nist.gov/cyberframework)

### Internal Documentation
- [Security Audit Report](./security-audit-report.md)
- [Vulnerability Remediation Guide](./vulnerability-remediation.md)
- [Incident Response Runbook](../operations/incident-response.md)

### Security Contacts
- Security Team: security@ainative.studio
- Bug Bounty: https://ainative.studio/security/bounty
- GPG Key: https://ainative.studio/security/gpg-key.asc

---

**Document Version:** 1.0
**Last Reviewed:** January 4, 2026
**Next Review:** April 4, 2026 (Quarterly)
**Maintained By:** Security Engineering Team
