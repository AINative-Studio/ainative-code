# Authentication API Reference

## Overview

This document describes the authentication API endpoints used by AINative Code CLI. These endpoints implement OAuth 2.0 Authorization Code Flow with PKCE and JWT token management.

**Base URL**: `https://auth.ainative.studio`

## Table of Contents

- [Authorization Endpoint](#authorization-endpoint)
- [Token Endpoint](#token-endpoint)
- [Token Refresh](#token-refresh)
- [Token Revocation](#token-revocation)
- [User Info Endpoint](#user-info-endpoint)
- [JWKS Endpoint](#jwks-endpoint)
- [Error Responses](#error-responses)
- [Rate Limiting](#rate-limiting)

---

## Authorization Endpoint

Initiates the OAuth 2.0 authorization flow with PKCE.

### Request

```http
GET /oauth/authorize HTTP/1.1
Host: auth.ainative.studio
```

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `response_type` | string | Yes | Must be `code` |
| `client_id` | string | Yes | OAuth client identifier (e.g., `ainative-code-cli`) |
| `redirect_uri` | string | Yes | Callback URL (must be pre-registered) |
| `code_challenge` | string | Yes | Base64url-encoded SHA-256 hash of code verifier |
| `code_challenge_method` | string | Yes | Must be `S256` |
| `state` | string | Yes | CSRF protection token (random string) |
| `scope` | string | No | Space-separated list of scopes (default: `read write`) |

**Example Request**:

```http
GET /oauth/authorize
  ?response_type=code
  &client_id=ainative-code-cli
  &redirect_uri=http%3A%2F%2Flocalhost%3A8080%2Fcallback
  &code_challenge=E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM
  &code_challenge_method=S256
  &state=xcoiv98y2kd22vusuye3kch
  &scope=read+write+offline_access HTTP/1.1
Host: auth.ainative.studio
```

### Response

**Success** - Redirect to `redirect_uri` with authorization code:

```http
HTTP/1.1 302 Found
Location: http://localhost:8080/callback?code=SplxlOBeZQQYbYS6WxSbIA&state=xcoiv98y2kd22vusuye3kch
```

**User Denies** - Redirect with error:

```http
HTTP/1.1 302 Found
Location: http://localhost:8080/callback?error=access_denied&error_description=User+denied+authorization&state=xcoiv98y2kd22vusuye3kch
```

---

## Token Endpoint

Exchanges authorization code for access and refresh tokens.

### Request

```http
POST /oauth/token HTTP/1.1
Host: auth.ainative.studio
Content-Type: application/x-www-form-urlencoded
```

**Parameters** (Authorization Code Grant):

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `grant_type` | string | Yes | Must be `authorization_code` |
| `code` | string | Yes | Authorization code from callback |
| `code_verifier` | string | Yes | Original PKCE code verifier |
| `client_id` | string | Yes | OAuth client identifier |
| `redirect_uri` | string | Yes | Must match authorization request |

**Example Request**:

```http
POST /oauth/token HTTP/1.1
Host: auth.ainative.studio
Content-Type: application/x-www-form-urlencoded

grant_type=authorization_code
&code=SplxlOBeZQQYbYS6WxSbIA
&code_verifier=dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk...
&client_id=ainative-code-cli
&redirect_uri=http%3A%2F%2Flocalhost%3A8080%2Fcallback
```

### Response

**Success** (200 OK):

```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "scope": "read write offline_access"
}
```

**Response Fields**:

| Field | Type | Description |
|-------|------|-------------|
| `access_token` | string | JWT access token for API authentication |
| `refresh_token` | string | JWT refresh token for obtaining new access tokens |
| `token_type` | string | Always `Bearer` |
| `expires_in` | integer | Access token lifetime in seconds (3600 = 1 hour) |
| `scope` | string | Granted scopes (space-separated) |

**Access Token Claims**:

```json
{
  "iss": "ainative-auth",
  "aud": "ainative-code",
  "sub": "user-123",
  "exp": 1704678000,
  "iat": 1704674400,
  "email": "user@example.com",
  "roles": ["user", "developer"]
}
```

**Refresh Token Claims**:

```json
{
  "iss": "ainative-auth",
  "aud": "ainative-code",
  "sub": "user-123",
  "exp": 1707270000,
  "iat": 1704674400,
  "session_id": "sess-abc123"
}
```

**Error Responses**:

See [Error Responses](#error-responses) section below.

---

## Token Refresh

Obtains a new access token using a refresh token.

### Request

```http
POST /oauth/token HTTP/1.1
Host: auth.ainative.studio
Content-Type: application/x-www-form-urlencoded
```

**Parameters** (Refresh Token Grant):

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `grant_type` | string | Yes | Must be `refresh_token` |
| `refresh_token` | string | Yes | Valid refresh token |
| `client_id` | string | Yes | OAuth client identifier |
| `scope` | string | No | Requested scopes (must be subset of original) |

**Example Request**:

```http
POST /oauth/token HTTP/1.1
Host: auth.ainative.studio
Content-Type: application/x-www-form-urlencoded

grant_type=refresh_token
&refresh_token=eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...
&client_id=ainative-code-cli
```

### Response

**Success** (200 OK):

```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "scope": "read write offline_access"
}
```

**Note**: Server may issue a new refresh token. Always store the latest refresh token.

**Error Responses**:

```json
{
  "error": "invalid_grant",
  "error_description": "Refresh token has expired"
}
```

```json
{
  "error": "invalid_grant",
  "error_description": "Refresh token has been revoked"
}
```

---

## Token Revocation

Revokes an access or refresh token.

### Request

```http
POST /oauth/revoke HTTP/1.1
Host: auth.ainative.studio
Content-Type: application/x-www-form-urlencoded
Authorization: Basic <base64(client_id:client_secret)>
```

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `token` | string | Yes | Access token or refresh token to revoke |
| `token_type_hint` | string | No | `access_token` or `refresh_token` |

**Example Request**:

```http
POST /oauth/revoke HTTP/1.1
Host: auth.ainative.studio
Content-Type: application/x-www-form-urlencoded

token=eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...
&token_type_hint=refresh_token
```

### Response

**Success** (200 OK):

```json
{
  "revoked": true
}
```

**Note**: Endpoint returns 200 even if token is already invalid (per OAuth 2.0 spec).

---

## User Info Endpoint

Retrieves information about the authenticated user.

### Request

```http
GET /oauth/userinfo HTTP/1.1
Host: auth.ainative.studio
Authorization: Bearer <access_token>
```

**Example Request**:

```http
GET /oauth/userinfo HTTP/1.1
Host: auth.ainative.studio
Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...
```

### Response

**Success** (200 OK):

```json
{
  "sub": "user-123",
  "email": "user@example.com",
  "email_verified": true,
  "name": "John Doe",
  "picture": "https://avatar.ainative.studio/user-123.jpg",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-12-01T00:00:00Z",
  "roles": ["user", "developer"]
}
```

**Response Fields**:

| Field | Type | Description |
|-------|------|-------------|
| `sub` | string | User unique identifier |
| `email` | string | User email address |
| `email_verified` | boolean | Whether email is verified |
| `name` | string | User display name |
| `picture` | string | Avatar URL |
| `created_at` | string | Account creation timestamp (ISO 8601) |
| `updated_at` | string | Last update timestamp (ISO 8601) |
| `roles` | array | User roles |

**Error Responses**:

```json
{
  "error": "invalid_token",
  "error_description": "Token has expired"
}
```

---

## JWKS Endpoint

Provides public keys for JWT signature verification.

### Request

```http
GET /.well-known/jwks.json HTTP/1.1
Host: auth.ainative.studio
```

### Response

**Success** (200 OK):

```json
{
  "keys": [
    {
      "kty": "RSA",
      "use": "sig",
      "kid": "2024-01-key",
      "alg": "RS256",
      "n": "0vx7agoebGcQSuuPiLJXZptN9nndrQmbXEps2aiAFbWhM78LhWx4cbbfAAtVT86zwu1RK7aPFFxuhDR1L6tSoc_BJECPebWKRXjBZCiFV4n3oknjhMstn64tZ_2W-5JsGY4Hc5n9yBXArwl93lqt7_RN5w6Cf0h4QyQ5v-65YGjQR0_FDW2QvzqY368QQMicAtaSqzs8KJZgnYb9c7d0zgdAZHzu6qMQvRL5hajrn1n91CbOpbISD08qNLyrdkt-bFTWhAI4vMQFh6WeZu0fM4lFd2NcRwr3XPksINHaQ-G_xBniIqbw0Ls1jF44-csFCur-kEgU8awapJzKnqDKgw",
      "e": "AQAB"
    }
  ]
}
```

**Key Fields**:

| Field | Type | Description |
|-------|------|-------------|
| `kty` | string | Key type (`RSA`) |
| `use` | string | Key usage (`sig` for signature) |
| `kid` | string | Key ID for key rotation |
| `alg` | string | Algorithm (`RS256`) |
| `n` | string | RSA modulus (base64url) |
| `e` | string | RSA exponent (base64url) |

**Usage Example**:

```go
import (
    "encoding/json"
    "crypto/rsa"
    "github.com/golang-jwt/jwt/v5"
)

// Fetch JWKS
resp, _ := http.Get("https://auth.ainative.studio/.well-known/jwks.json")
var jwks struct {
    Keys []json.RawMessage `json:"keys"`
}
json.NewDecoder(resp.Body).Decode(&jwks)

// Parse key
var key *rsa.PublicKey
// ... convert JWK to RSA public key

// Verify token
token, _ := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
    return key, nil
})
```

---

## Error Responses

### OAuth Error Response Format

```json
{
  "error": "error_code",
  "error_description": "Human-readable error description",
  "error_uri": "https://docs.ainative.studio/errors/error_code"
}
```

### Common Error Codes

**Authorization Endpoint Errors**:

| Error Code | Description | User Action |
|------------|-------------|-------------|
| `invalid_request` | Missing or invalid parameter | Check authorization URL parameters |
| `unauthorized_client` | Client not authorized | Verify client_id is registered |
| `access_denied` | User denied authorization | User must approve authorization |
| `unsupported_response_type` | Invalid response_type | Must use `response_type=code` |
| `invalid_scope` | Invalid or unknown scope | Check requested scopes |
| `server_error` | Server internal error | Retry or contact support |

**Token Endpoint Errors**:

| Error Code | Description | User Action |
|------------|-------------|-------------|
| `invalid_request` | Missing required parameter | Check all required parameters |
| `invalid_client` | Invalid client_id | Verify client configuration |
| `invalid_grant` | Invalid or expired authorization code | Retry authorization flow |
| `unauthorized_client` | Client not authorized for grant type | Contact support |
| `unsupported_grant_type` | Invalid grant_type | Use `authorization_code` or `refresh_token` |
| `invalid_scope` | Requested scope invalid or exceeds granted | Request valid scopes |

**Token Validation Errors**:

| Error Code | Description | User Action |
|------------|-------------|-------------|
| `invalid_token` | Token is malformed | Re-authenticate |
| `token_expired` | Token has expired | Refresh token or re-authenticate |
| `invalid_signature` | Token signature invalid | Re-authenticate |
| `invalid_issuer` | Token issuer mismatch | Verify token source |
| `invalid_audience` | Token audience mismatch | Verify token intended for this client |

**Example Error Responses**:

**Invalid authorization code**:
```json
{
  "error": "invalid_grant",
  "error_description": "Authorization code has expired or already been used"
}
```

**Expired refresh token**:
```json
{
  "error": "invalid_grant",
  "error_description": "Refresh token has expired. Please re-authenticate."
}
```

**Invalid PKCE verifier**:
```json
{
  "error": "invalid_grant",
  "error_description": "Code verifier does not match code challenge"
}
```

**Invalid scope**:
```json
{
  "error": "invalid_scope",
  "error_description": "Requested scope 'admin' exceeds granted scopes"
}
```

---

## Rate Limiting

All authentication endpoints implement rate limiting to prevent abuse.

### Rate Limit Headers

```http
HTTP/1.1 200 OK
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1704678000
```

**Header Descriptions**:

| Header | Description |
|--------|-------------|
| `X-RateLimit-Limit` | Maximum requests per window |
| `X-RateLimit-Remaining` | Requests remaining in current window |
| `X-RateLimit-Reset` | Unix timestamp when limit resets |

### Rate Limits by Endpoint

| Endpoint | Limit | Window | Scope |
|----------|-------|--------|-------|
| `/oauth/authorize` | 10 requests | 1 minute | Per IP |
| `/oauth/token` | 20 requests | 1 minute | Per client_id |
| `/oauth/revoke` | 10 requests | 1 minute | Per client_id |
| `/oauth/userinfo` | 100 requests | 1 minute | Per access token |
| `/.well-known/jwks.json` | 100 requests | 1 minute | Per IP |

### Rate Limit Exceeded Response

```http
HTTP/1.1 429 Too Many Requests
Content-Type: application/json
Retry-After: 60

{
  "error": "rate_limit_exceeded",
  "error_description": "Too many requests. Please try again in 60 seconds.",
  "retry_after": 60
}
```

### Best Practices

**1. Implement Exponential Backoff**:

```go
func retryWithBackoff(fn func() error) error {
    backoff := 1 * time.Second
    maxRetries := 5

    for i := 0; i < maxRetries; i++ {
        err := fn()
        if err == nil {
            return nil
        }

        if isRateLimitError(err) {
            time.Sleep(backoff)
            backoff *= 2
            continue
        }

        return err
    }

    return fmt.Errorf("max retries exceeded")
}
```

**2. Cache JWKS Response**:

```go
// Cache JWKS for 24 hours
var jwksCache struct {
    keys      []JWK
    expiresAt time.Time
}

func getJWKS() ([]JWK, error) {
    if time.Now().Before(jwksCache.expiresAt) {
        return jwksCache.keys, nil
    }

    // Fetch fresh JWKS
    keys, err := fetchJWKS()
    if err != nil {
        return nil, err
    }

    jwksCache.keys = keys
    jwksCache.expiresAt = time.Now().Add(24 * time.Hour)

    return keys, nil
}
```

**3. Respect Retry-After Header**:

```go
if resp.StatusCode == 429 {
    retryAfter, _ := strconv.Atoi(resp.Header.Get("Retry-After"))
    time.Sleep(time.Duration(retryAfter) * time.Second)
    // Retry request
}
```

---

## Security Considerations

### HTTPS Only

**All endpoints must be accessed via HTTPS**. HTTP requests will be rejected:

```http
HTTP/1.1 400 Bad Request
Content-Type: application/json

{
  "error": "invalid_request",
  "error_description": "HTTPS required for OAuth endpoints"
}
```

### PKCE Required

Authorization code flow **requires PKCE**. Requests without PKCE will be rejected:

```http
HTTP/1.1 400 Bad Request
Content-Type: application/json

{
  "error": "invalid_request",
  "error_description": "code_challenge and code_challenge_method are required"
}
```

### State Parameter Required

State parameter is **mandatory** for CSRF protection:

```http
HTTP/1.1 400 Bad Request
Content-Type: application/json

{
  "error": "invalid_request",
  "error_description": "state parameter is required"
}
```

### Token Binding

Refresh tokens are bound to:
- **Client ID**: Cannot be used by different client
- **Session**: Invalidated when session expires or is revoked
- **User**: Invalidated when user changes password or revokes access

---

## Client Implementation Example

Complete Go client implementation:

```go
package main

import (
    "context"
    "crypto/rsa"
    "encoding/json"
    "fmt"
    "net/http"
    "net/url"
    "strings"
    "time"

    "github.com/AINative-studio/ainative-code/internal/auth"
)

type OAuthClient struct {
    clientID    string
    authURL     string
    tokenURL    string
    redirectURI string
    publicKey   *rsa.PublicKey
}

func (c *OAuthClient) Authorize(ctx context.Context) (*auth.TokenPair, error) {
    // Generate PKCE parameters
    pkce, err := auth.GeneratePKCE()
    if err != nil {
        return nil, err
    }

    // Build authorization URL
    authURL := fmt.Sprintf("%s?%s", c.authURL, url.Values{
        "response_type":         {"code"},
        "client_id":             {c.clientID},
        "redirect_uri":          {c.redirectURI},
        "code_challenge":        {pkce.CodeChallenge},
        "code_challenge_method": {"S256"},
        "state":                 {pkce.State},
        "scope":                 {"read write offline_access"},
    }.Encode())

    // Open browser and start callback server
    // ... (implementation details)

    // Exchange code for tokens
    return c.exchangeCode(ctx, code, pkce.CodeVerifier)
}

func (c *OAuthClient) exchangeCode(ctx context.Context, code, verifier string) (*auth.TokenPair, error) {
    data := url.Values{
        "grant_type":    {"authorization_code"},
        "code":          {code},
        "code_verifier": {verifier},
        "client_id":     {c.clientID},
        "redirect_uri":  {c.redirectURI},
    }

    req, _ := http.NewRequestWithContext(ctx, "POST", c.tokenURL,
        strings.NewReader(data.Encode()))
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != 200 {
        var errResp struct {
            Error            string `json:"error"`
            ErrorDescription string `json:"error_description"`
        }
        json.NewDecoder(resp.Body).Decode(&errResp)
        return nil, fmt.Errorf("%s: %s", errResp.Error, errResp.ErrorDescription)
    }

    var tokenResp struct {
        AccessToken  string `json:"access_token"`
        RefreshToken string `json:"refresh_token"`
        ExpiresIn    int64  `json:"expires_in"`
    }
    json.NewDecoder(resp.Body).Decode(&tokenResp)

    // Parse and validate tokens
    accessToken, _ := auth.ParseAccessToken(tokenResp.AccessToken, c.publicKey)
    refreshToken, _ := auth.ParseRefreshToken(tokenResp.RefreshToken, c.publicKey)

    return &auth.TokenPair{
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
        ReceivedAt:   time.Now(),
    }, nil
}
```

---

## Related Documentation

- [Authentication Overview](README.md)
- [OAuth Flow](oauth-flow.md)
- [User Guide](user-guide.md)
- [Security Best Practices](security-best-practices.md)

## External References

- [RFC 6749: OAuth 2.0 Authorization Framework](https://tools.ietf.org/html/rfc6749)
- [RFC 7636: Proof Key for Code Exchange (PKCE)](https://tools.ietf.org/html/rfc7636)
- [RFC 7519: JSON Web Tokens (JWT)](https://tools.ietf.org/html/rfc7519)
- [RFC 7517: JSON Web Key (JWK)](https://tools.ietf.org/html/rfc7517)
