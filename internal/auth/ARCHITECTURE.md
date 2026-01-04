# OAuth 2.0 PKCE Authentication Architecture

**Package**: `internal/auth`
**Task**: TASK-044 (Issue #32) - OAuth 2.0 PKCE Flow
**Dependency**: TASK-040 (Issue #28) - JWT Token Structures
**Status**: In Design

---

## Executive Summary

This package implements OAuth 2.0 authorization code flow with PKCE (Proof Key for Code Exchange) for secure authentication with the AINative platform, plus JWT token management for access/refresh tokens.

**Design Philosophy**: Follow proven patterns from TASK-023 (LLM Provider Interface) which achieved 100% test coverage through:
- Provider interface pattern
- Functional options pattern
- Thread-safe operations
- Context-first design
- Comprehensive error handling

---

## OAuth 2.0 PKCE Flow Sequence

```
┌─────────────┐                                      ┌──────────────┐
│  CLI Client │                                      │  AINative    │
│             │                                      │  Auth Server │
└──────┬──────┘                                      └──────┬───────┘
       │                                                     │
       │ 1. Generate PKCE params                            │
       │    - code_verifier (random 43-128 chars)           │
       │    - code_challenge = SHA256(verifier)             │
       │    - state (CSRF token)                            │
       │                                                     │
       │ 2. Build authorization URL                         │
       │    + client_id, redirect_uri, code_challenge       │
       │                                                     │
       │ 3. Open browser ──────────────────────────────────>│
       │                                                     │
       │                                    4. User login    │
       │                                    & authorize      │
       │                                                     │
       │<───────────────── 5. Redirect to callback ─────────│
       │    http://localhost:8080/callback?code=XXX&state=YYY
       │                                                     │
┌──────▼──────┐                                             │
│  Callback   │                                             │
│  Server     │                                             │
│ :8080       │                                             │
└──────┬──────┘                                             │
       │ 6. Validate state (CSRF check)                     │
       │                                                     │
       │ 7. Exchange code for tokens ──────────────────────>│
       │    POST /oauth/token                               │
       │    + code, code_verifier, client_id, redirect_uri  │
       │                                                     │
       │<───────────────── 8. Return tokens ────────────────│
       │    {                                               │
       │      "access_token": "eyJhbG...",                  │
       │      "refresh_token": "eyJhbG...",                 │
       │      "expires_in": 86400                           │
       │    }                                               │
       │                                                     │
       │ 9. Parse & validate JWT tokens                     │
       │    - Verify RS256 signature                        │
       │    - Check expiry                                  │
       │    - Extract claims                                │
       │                                                     │
       │ 10. Store in OS keychain                           │
       │     - macOS: Keychain                              │
       │     - Windows: Credential Manager                  │
       │     - Linux: Secret Service                        │
       │                                                     │
```

---

## Component Architecture

### 1. Client Interface (Following Provider Pattern)

```go
// Client manages OAuth 2.0 PKCE authentication flow
type Client interface {
    // Authorize initiates OAuth flow, opens browser, returns tokens
    Authorize(ctx context.Context, opts ...Option) (*TokenResponse, error)

    // ExchangeCode exchanges authorization code for tokens
    ExchangeCode(ctx context.Context, code string, opts ...Option) (*TokenResponse, error)

    // RefreshToken obtains new access token using refresh token
    RefreshToken(ctx context.Context, refreshToken string, opts ...Option) (*TokenResponse, error)

    // ValidateToken verifies JWT signature and claims
    ValidateToken(ctx context.Context, token string) (*AccessToken, error)

    // Close releases resources
    Close() error
}
```

**Design Decisions**:
- Context-first parameters for cancellation/timeout
- Variadic options for flexible configuration
- io.Closer interface for resource cleanup
- Returns structured token types (not raw strings)

### 2. PKCE Generator

```go
// PKCEParams contains PKCE flow parameters
type PKCEParams struct {
    CodeVerifier  string // 43-128 random chars
    CodeChallenge string // SHA256(verifier), base64url
    Method        string // "S256"
}

// GeneratePKCE creates PKCE parameters for OAuth flow
func GeneratePKCE() (*PKCEParams, error)
```

**Security**:
- Cryptographically secure random generator (`crypto/rand`)
- Code verifier: 128 characters (maximum security)
- SHA-256 hashing for challenge
- Base64URL encoding (RFC 4648 Section 5, no padding)

### 3. Callback Server

```go
// CallbackServer handles OAuth redirect
type CallbackServer struct {
    server   *http.Server
    port     int
    codeChan chan *CallbackResult
    errChan  chan error
}

// Start begins listening for OAuth callback
func (s *CallbackServer) Start(ctx context.Context) error

// Wait blocks until callback received or context cancelled
func (s *CallbackServer) Wait(ctx context.Context) (*CallbackResult, error)

// Close shuts down server
func (s *CallbackServer) Close() error
```

**Design**:
- Single-use server (closes after one callback)
- Channel-based async communication
- Context cancellation support
- Graceful shutdown with timeout

### 4. JWT Token Management

```go
// AccessToken represents parsed access token
type AccessToken struct {
    Raw       string
    ExpiresAt time.Time
    UserID    string
    Email     string
    Roles     []string
}

// RefreshToken represents parsed refresh token
type RefreshToken struct {
    Raw       string
    ExpiresAt time.Time
    UserID    string
    SessionID string
}

// ParseAccessToken parses and validates access token
func ParseAccessToken(tokenString string, publicKey *rsa.PublicKey) (*AccessToken, error)

// ParseRefreshToken parses and validates refresh token
func ParseRefreshToken(tokenString string, publicKey *rsa.PublicKey) (*RefreshToken, error)
```

**Validation**:
- RS256 signature verification
- Expiry checking
- Issuer validation ("ainative-auth")
- Audience validation ("ainative-code")

### 5. Keychain Storage

```go
// KeychainStorage manages secure token storage
type KeychainStorage interface {
    Store(key string, value []byte) error
    Retrieve(key string) ([]byte, error)
    Delete(key string) error
    Clear() error
}

// Platform implementations
type macOSKeychain struct {}
type windowsCredentialManager struct {}
type linuxSecretService struct {}
```

**Security**:
- OS-native secure storage
- Encrypted at rest
- Per-user isolation
- No plaintext token files

### 6. Functional Options

```go
type Option func(*ClientOptions)

func WithClientID(id string) Option
func WithAuthEndpoint(url string) Option
func WithTokenEndpoint(url string) Option
func WithRedirectURI(uri string) Option
func WithScopes(scopes ...string) Option
func WithTimeout(timeout time.Duration) Option
func WithPublicKey(key *rsa.PublicKey) Option
```

**Benefits**:
- Backward compatible API evolution
- Optional parameters without overloading
- Composable configuration
- Type-safe

---

## Error Handling

### Error Types

```go
var (
    ErrInvalidState      = errors.New("invalid state parameter (CSRF check failed)")
    ErrCodeExchangeFailed = errors.New("failed to exchange code for token")
    ErrTokenExpired      = errors.New("token has expired")
    ErrInvalidSignature  = errors.New("invalid token signature")
    ErrInvalidClaims     = errors.New("invalid token claims")
    ErrKeychainAccess    = errors.New("keychain access denied")
)
```

### Error Wrapping

All errors use `fmt.Errorf` with `%w` verb for error chains:
```go
return nil, fmt.Errorf("failed to generate PKCE challenge: %w", err)
```

---

## Security Considerations

### 1. PKCE (RFC 7636)

**Threat**: Authorization code interception attack
**Mitigation**: Code verifier known only to client, server validates challenge

### 2. CSRF Protection

**Threat**: Cross-site request forgery
**Mitigation**: State parameter validation (cryptographically random, single-use)

### 3. Token Storage

**Threat**: Token theft from disk
**Mitigation**: OS keychain (encrypted storage, access control)

### 4. JWT Validation

**Threat**: Token forgery
**Mitigation**: RS256 signature verification with public key

### 5. Localhost Callback

**Threat**: Malicious app on same machine intercepting callback
**Mitigation**: State validation, single-use callback server

---

## Configuration Integration

### Config Schema Addition

```yaml
platform:
  oauth:
    client_id: "ainative-code-cli"
    auth_endpoint: "https://auth.ainative.studio/oauth/authorize"
    token_endpoint: "https://auth.ainative.studio/oauth/token"
    redirect_uri: "http://localhost:8080/callback"
    scopes:
      - "read"
      - "write"
      - "offline_access"
    timeout: 30s

  jwt:
    issuer: "ainative-auth"
    audience: "ainative-code"
    public_key_path: "~/.ainative/jwt-public.pem"
```

---

## Testing Strategy

### Target Coverage: 80%+ (Required by Issue #32)

Following TASK-023's 100% coverage achievement:

**1. Unit Tests**
- PKCE generation (valid length, base64url encoding)
- JWT parsing (valid/expired/invalid signature)
- State generation (uniqueness, length)
- URL building (proper encoding, all parameters)

**2. Integration Tests**
- Mock OAuth server for end-to-end flow
- Mock keychain for storage operations
- Timeout scenarios
- Error paths (network failures, invalid responses)

**3. Security Tests**
- Invalid state rejection
- Expired token rejection
- Invalid signature rejection
- Code verifier mismatch

**4. Concurrency Tests**
- Thread-safe token storage
- Concurrent refresh operations

---

## File Structure

```
internal/auth/
├── ARCHITECTURE.md        # This document
├── doc.go                # Package documentation
├── types.go              # Core type definitions
├── interface.go          # Client interface
├── options.go            # Functional options
├── client.go             # Client implementation
├── pkce.go               # PKCE generation
├── jwt.go                # JWT parsing/validation
├── callback.go           # Callback server
├── keychain.go           # Keychain storage interface
├── keychain_darwin.go    # macOS implementation
├── keychain_windows.go   # Windows implementation
├── keychain_linux.go     # Linux implementation
├── errors.go             # Error definitions
├── types_test.go         # Type tests
├── pkce_test.go          # PKCE tests
├── jwt_test.go           # JWT tests
├── callback_test.go      # Callback server tests
├── keychain_test.go      # Keychain tests
├── client_test.go        # Integration tests
└── README.md             # Usage documentation
```

---

## Dependencies

### Standard Library
- `context` - Cancellation/timeout
- `crypto/rand` - Secure random generation
- `crypto/sha256` - PKCE challenge hashing
- `crypto/rsa` - JWT signature verification
- `encoding/base64` - Base64URL encoding
- `encoding/json` - JSON parsing
- `net/http` - Callback server
- `time` - Token expiry

### External
- `github.com/golang-jwt/jwt/v5` - JWT operations ✅ (installed in previous session)

### Platform-Specific
- macOS: `github.com/keybase/go-keychain`
- Windows: `github.com/danieljoos/wincred`
- Linux: `github.com/zalando/go-keyring`

---

## Implementation Phases

### Phase 1: Core Types & PKCE ✅ (Next)
- types.go
- pkce.go
- pkce_test.go
- errors.go

### Phase 2: JWT Operations
- jwt.go
- jwt_test.go

### Phase 3: OAuth Client
- client.go
- options.go
- callback.go
- client_test.go
- callback_test.go

### Phase 4: Keychain Integration
- keychain.go
- keychain_*.go (platform-specific)
- keychain_test.go

### Phase 5: Integration & Documentation
- README.md
- Integration tests
- Configuration updates

---

## Performance Considerations

- **PKCE Generation**: ~1ms (crypto/rand + SHA256)
- **JWT Validation**: ~2ms (RSA signature verification)
- **Callback Server**: Single-use, minimal overhead
- **Keychain Operations**: Platform-dependent (10-100ms)

---

## References

- **RFC 7636**: Proof Key for Code Exchange (PKCE)
- **RFC 6749**: OAuth 2.0 Authorization Framework
- **RFC 7519**: JSON Web Tokens (JWT)
- **RFC 4648**: Base64 Encoding (Section 5: Base64URL)
- **TASK-023**: LLM Provider Interface (reference implementation, 100% coverage)

---

**Design Status**: ✅ Complete
**Next Step**: Implement Phase 1 (Core Types & PKCE)
**Target Coverage**: 80%+ (TASK-023 achieved 100%)
