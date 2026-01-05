# Authentication API Reference

**Import Path**: `github.com/AINative-studio/ainative-code/internal/auth`

The auth package provides OAuth 2.0 authentication with PKCE flow, JWT token management, and OS keychain integration.

## Table of Contents

- [Client Interface](#client-interface)
- [Token Types](#token-types)
- [OAuth Flow](#oauth-flow)
- [JWT Validation](#jwt-validation)
- [Keychain Integration](#keychain-integration)
- [Usage Examples](#usage-examples)

## Client Interface

### Client

```go
type Client interface {
    Authenticate(ctx context.Context) (*TokenPair, error)
    RefreshToken(ctx context.Context, refreshToken *RefreshToken) (*TokenPair, error)
    GetStoredTokens(ctx context.Context) (*TokenPair, error)
    StoreTokens(ctx context.Context, tokens *TokenPair) error
    ClearTokens(ctx context.Context) error
    ValidateToken(ctx context.Context, token *AccessToken) bool
}
```

The Client interface defines OAuth 2.0 authentication operations with PKCE support.

## Token Types

### TokenPair

```go
type TokenPair struct {
    AccessToken  *AccessToken
    RefreshToken *RefreshToken
    ReceivedAt   time.Time
}
```

Represents a pair of access and refresh tokens.

**Methods**:

```go
func (tp *TokenPair) IsValid() bool
func (tp *TokenPair) NeedsRefresh() bool
```

**Example**:

```go
tokens, err := authClient.GetStoredTokens(ctx)
if err != nil || !tokens.IsValid() {
    // Re-authenticate
    tokens, err = authClient.Authenticate(ctx)
}

if tokens.NeedsRefresh() {
    tokens, err = authClient.RefreshToken(ctx, tokens.RefreshToken)
}
```

### AccessToken

```go
type AccessToken struct {
    Raw       string
    ExpiresAt time.Time
    UserID    string
    Email     string
    Roles     []string
    Issuer    string
    Audience  string
}
```

Represents a parsed and validated JWT access token.

**Methods**:

```go
func (t *AccessToken) IsExpired() bool
func (t *AccessToken) IsValid() bool
```

**Example**:

```go
if tokens.AccessToken.IsExpired() {
    log.Println("Access token has expired")
    // Refresh or re-authenticate
}

fmt.Printf("Authenticated as: %s (%s)\n",
    tokens.AccessToken.Email,
    tokens.AccessToken.UserID)
```

### RefreshToken

```go
type RefreshToken struct {
    Raw       string
    ExpiresAt time.Time
    UserID    string
    SessionID string
    Issuer    string
    Audience  string
}
```

Represents a parsed and validated JWT refresh token.

**Methods**:

```go
func (t *RefreshToken) IsExpired() bool
func (t *RefreshToken) IsValid() bool
```

### PKCEParams

```go
type PKCEParams struct {
    CodeVerifier  string
    CodeChallenge string
    Method        string  // Always "S256" for SHA-256
    State         string  // CSRF token
}
```

Contains parameters for OAuth 2.0 PKCE flow (RFC 7636).

### ClientOptions

```go
type ClientOptions struct {
    ClientID      string
    AuthEndpoint  string
    TokenEndpoint string
    RedirectURI   string
    Scopes        []string
    Timeout       time.Duration
    PublicKey     *rsa.PublicKey
    CallbackPort  int
}
```

Configuration for the OAuth client.

**Function**:

```go
func DefaultClientOptions() *ClientOptions
```

Returns default client options:
- ClientID: "ainative-code-cli"
- AuthEndpoint: "https://auth.ainative.studio/oauth/authorize"
- TokenEndpoint: "https://auth.ainative.studio/oauth/token"
- RedirectURI: "http://localhost:8080/callback"
- Scopes: ["read", "write", "offline_access"]
- Timeout: 30 seconds
- CallbackPort: 8080

## OAuth Flow

### Authenticate

```go
func (c Client) Authenticate(ctx context.Context) (*TokenPair, error)
```

Initiates the OAuth 2.0 Authorization Code Flow with PKCE.

**Process**:
1. Generates PKCE parameters (code verifier, challenge, state)
2. Builds authorization URL with required parameters
3. Opens the user's browser to the authorization URL
4. Starts local callback server to receive authorization code
5. Waits for user to complete authorization
6. Exchanges authorization code for tokens
7. Stores tokens securely in OS keychain
8. Returns the token pair

**Errors**:
- `ErrAuthorizationDenied` - User denies authorization
- `ErrAuthorizationTimeout` - User doesn't complete flow in time
- `ErrInvalidState` - State parameter doesn't match (CSRF attack)
- `ErrCodeExchangeFailed` - Token endpoint rejects the code
- `ErrCallbackServerStart` - Local server fails to start
- `ErrBrowserOpen` - Browser cannot be opened

**Example**:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

authClient, err := auth.NewClient(auth.DefaultClientOptions())
if err != nil {
    log.Fatalf("Failed to create auth client: %v", err)
}

tokens, err := authClient.Authenticate(ctx)
if err != nil {
    log.Fatalf("Authentication failed: %v", err)
}

fmt.Printf("Authenticated as: %s\n", tokens.AccessToken.Email)
```

### RefreshToken

```go
func (c Client) RefreshToken(ctx context.Context, refreshToken *RefreshToken) (*TokenPair, error)
```

Exchanges a refresh token for a new access token.

**Process**:
1. Sends refresh token to token endpoint
2. Receives new access token (and optionally new refresh token)
3. Updates tokens in OS keychain
4. Returns updated token pair

**Errors**:
- `ErrTokenExpired` - Refresh token has expired
- `ErrCodeExchangeFailed` - Token endpoint rejects refresh token
- `ErrHTTPRequest` - Network request fails
- `ErrHTTPResponse` - Server returns error status

**Example**:

```go
tokens, err := authClient.GetStoredTokens(ctx)
if err != nil {
    return authClient.Authenticate(ctx)
}

if tokens.AccessToken.IsExpired() {
    tokens, err = authClient.RefreshToken(ctx, tokens.RefreshToken)
    if err != nil {
        // Refresh failed, re-authenticate
        return authClient.Authenticate(ctx)
    }
}
```

## JWT Validation

### ValidateToken

```go
func (c Client) ValidateToken(ctx context.Context, token *AccessToken) bool
```

Checks if an access token is valid and not expired.

**Validation Steps**:
1. Checks token expiration time
2. Verifies JWT signature using public key
3. Validates standard claims (iss, aud, sub, exp)
4. Optionally performs introspection at server (if configured)

**Example**:

```go
tokens, _ := authClient.GetStoredTokens(ctx)
if !authClient.ValidateToken(ctx, tokens.AccessToken) {
    // Token invalid, refresh or re-authenticate
    tokens, err = authClient.RefreshToken(ctx, tokens.RefreshToken)
    if err != nil {
        tokens, err = authClient.Authenticate(ctx)
    }
}
```

## Keychain Integration

### GetStoredTokens

```go
func (c Client) GetStoredTokens(ctx context.Context) (*TokenPair, error)
```

Retrieves tokens from OS keychain.

**Process**:
1. Reads token data from secure keychain storage
2. Parses and validates JWT tokens
3. Returns token pair if found and valid

**Errors**:
- `ErrKeychainNotFound` - No tokens are stored
- `ErrKeychainAccess` - Keychain access is denied
- `ErrKeychainRetrieve` - Reading from keychain fails
- `ErrTokenParseFailed` - Stored token data is corrupted

**Example**:

```go
tokens, err := authClient.GetStoredTokens(ctx)
if errors.Is(err, auth.ErrKeychainNotFound) {
    // No stored tokens, authenticate
    return authClient.Authenticate(ctx)
}
if err != nil {
    return fmt.Errorf("failed to get tokens: %w", err)
}
```

### StoreTokens

```go
func (c Client) StoreTokens(ctx context.Context, tokens *TokenPair) error
```

Saves tokens to OS keychain.

**Process**:
1. Validates token format and expiration
2. Encrypts token data using OS keychain
3. Stores in platform-specific secure storage

**Errors**:
- `ErrKeychainAccess` - Keychain access is denied
- `ErrKeychainStore` - Writing to keychain fails

**Example**:

```go
tokens, err := authClient.Authenticate(ctx)
if err != nil {
    return err
}

// Tokens are automatically stored by Authenticate(),
// but you can manually store them if needed:
if err := authClient.StoreTokens(ctx, tokens); err != nil {
    log.Printf("Failed to store tokens: %v", err)
}
```

### ClearTokens

```go
func (c Client) ClearTokens(ctx context.Context) error
```

Removes tokens from OS keychain.

**Process**:
1. Deletes token data from OS keychain
2. Clears any in-memory token cache

**Errors**:
- `ErrKeychainAccess` - Keychain access is denied
- `ErrKeychainDelete` - Deletion fails

**Note**: This does NOT revoke tokens on the server. To fully logout, you should also call the server's token revocation endpoint.

**Example**:

```go
// Implement logout
if err := authClient.ClearTokens(ctx); err != nil {
    log.Printf("Failed to clear tokens: %v", err)
}
fmt.Println("Logged out successfully")
```

## Usage Examples

### Complete Authentication Flow

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/AINative-studio/ainative-code/internal/auth"
)

func main() {
    ctx := context.Background()

    // Create auth client
    authClient, err := auth.NewClient(auth.DefaultClientOptions())
    if err != nil {
        log.Fatalf("Failed to create auth client: %v", err)
    }

    // Try to get stored tokens
    tokens, err := authClient.GetStoredTokens(ctx)
    if err != nil || !tokens.IsValid() {
        log.Println("No valid stored tokens, authenticating...")

        // Authenticate user (opens browser)
        ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
        defer cancel()

        tokens, err = authClient.Authenticate(ctx)
        if err != nil {
            log.Fatalf("Authentication failed: %v", err)
        }

        log.Printf("Authenticated as: %s", tokens.AccessToken.Email)
    } else {
        log.Println("Using stored tokens")
    }

    // Check if token needs refresh
    if tokens.NeedsRefresh() {
        log.Println("Refreshing access token...")
        tokens, err = authClient.RefreshToken(ctx, tokens.RefreshToken)
        if err != nil {
            log.Printf("Token refresh failed: %v, re-authenticating", err)
            tokens, err = authClient.Authenticate(ctx)
            if err != nil {
                log.Fatalf("Re-authentication failed: %v", err)
            }
        }
    }

    // Use the access token
    fmt.Printf("Access token: %s...\n", tokens.AccessToken.Raw[:20])
    fmt.Printf("User ID: %s\n", tokens.AccessToken.UserID)
    fmt.Printf("Email: %s\n", tokens.AccessToken.Email)
    fmt.Printf("Roles: %v\n", tokens.AccessToken.Roles)
    fmt.Printf("Expires at: %s\n", tokens.AccessToken.ExpiresAt)
}
```

### Automatic Token Refresh

```go
type AuthenticatedClient struct {
    authClient auth.Client
    tokens     *auth.TokenPair
    mu         sync.RWMutex
}

func (c *AuthenticatedClient) GetAccessToken(ctx context.Context) (string, error) {
    c.mu.Lock()
    defer c.mu.Unlock()

    // Check if tokens need refresh
    if c.tokens == nil || c.tokens.NeedsRefresh() {
        // Try to refresh
        if c.tokens != nil && c.tokens.RefreshToken != nil {
            newTokens, err := c.authClient.RefreshToken(ctx, c.tokens.RefreshToken)
            if err == nil {
                c.tokens = newTokens
                return c.tokens.AccessToken.Raw, nil
            }
        }

        // Refresh failed or no refresh token, re-authenticate
        tokens, err := c.authClient.Authenticate(ctx)
        if err != nil {
            return "", err
        }
        c.tokens = tokens
    }

    return c.tokens.AccessToken.Raw, nil
}
```

### Custom OAuth Configuration

```go
import "crypto/rsa"

// Load custom public key for JWT validation
publicKey, err := loadPublicKey("path/to/public_key.pem")
if err != nil {
    log.Fatal(err)
}

// Create custom client options
options := &auth.ClientOptions{
    ClientID:      "my-custom-client",
    AuthEndpoint:  "https://custom-auth.example.com/oauth/authorize",
    TokenEndpoint: "https://custom-auth.example.com/oauth/token",
    RedirectURI:   "http://localhost:9000/callback",
    Scopes:        []string{"read", "write", "admin"},
    Timeout:       60 * time.Second,
    PublicKey:     publicKey,
    CallbackPort:  9000,
}

authClient, err := auth.NewClient(options)
if err != nil {
    log.Fatal(err)
}
```

### Token Introspection

```go
// Validate token and get details
if authClient.ValidateToken(ctx, tokens.AccessToken) {
    fmt.Printf("Token is valid\n")
    fmt.Printf("User: %s\n", tokens.AccessToken.Email)
    fmt.Printf("Roles: %v\n", tokens.AccessToken.Roles)
    fmt.Printf("Expires in: %v\n", time.Until(tokens.AccessToken.ExpiresAt))
} else {
    fmt.Println("Token is invalid or expired")
}
```

### Error Handling

```go
import "github.com/AINative-studio/ainative-code/internal/errors"

tokens, err := authClient.Authenticate(ctx)
if err != nil {
    // Check specific error types
    if errors.GetCode(err) == errors.ErrCodeAuthFailed {
        log.Println("Authentication failed - check credentials")
    } else if errors.GetCode(err) == errors.ErrCodeAuthExpiredToken {
        log.Println("Token expired - refreshing")
        tokens, err = authClient.RefreshToken(ctx, tokens.RefreshToken)
    } else {
        log.Printf("Unexpected error: %v", err)
    }
}
```

### Logout Implementation

```go
func logout(ctx context.Context, authClient auth.Client) error {
    // Clear local tokens
    if err := authClient.ClearTokens(ctx); err != nil {
        return fmt.Errorf("failed to clear tokens: %w", err)
    }

    // TODO: Call server's token revocation endpoint
    // This would revoke the refresh token server-side

    log.Println("Logged out successfully")
    return nil
}
```

## Best Practices

### 1. Always Use Context with Timeout

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

tokens, err := authClient.Authenticate(ctx)
```

### 2. Handle Token Refresh Gracefully

```go
if tokens.NeedsRefresh() {
    newTokens, err := authClient.RefreshToken(ctx, tokens.RefreshToken)
    if err != nil {
        // Refresh failed, re-authenticate
        newTokens, err = authClient.Authenticate(ctx)
    }
    tokens = newTokens
}
```

### 3. Store Tokens Securely

```go
// Tokens are automatically stored in OS keychain by Authenticate()
// and RefreshToken(), using platform-specific secure storage:
// - macOS: Keychain
// - Windows: Credential Manager
// - Linux: Secret Service API
```

### 4. Validate Tokens Before Use

```go
// Quick check
if tokens.AccessToken.IsExpired() {
    // Refresh or re-authenticate
}

// Full validation
if !authClient.ValidateToken(ctx, tokens.AccessToken) {
    // Token invalid, get new one
}
```

### 5. Implement Proper Error Recovery

```go
func ensureAuthenticated(ctx context.Context, authClient auth.Client) (*auth.TokenPair, error) {
    // Try stored tokens
    tokens, err := authClient.GetStoredTokens(ctx)
    if err == nil && tokens.IsValid() && !tokens.NeedsRefresh() {
        return tokens, nil
    }

    // Try refresh
    if err == nil && tokens.RefreshToken != nil && !tokens.RefreshToken.IsExpired() {
        tokens, err = authClient.RefreshToken(ctx, tokens.RefreshToken)
        if err == nil {
            return tokens, nil
        }
    }

    // Full re-authentication
    return authClient.Authenticate(ctx)
}
```

## Related Documentation

- [Configuration](configuration.md) - Auth configuration
- [Errors](errors.md) - Error handling
- [Core Packages](core-packages.md) - Client integration
- [Security](../security/) - Security best practices
