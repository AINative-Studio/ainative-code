package auth

import (
	"crypto/rsa"
	"time"
)

// PKCEParams contains parameters for OAuth 2.0 PKCE flow (RFC 7636).
//
// PKCE (Proof Key for Code Exchange) prevents authorization code interception
// attacks by requiring the client to prove possession of a code verifier that
// corresponds to the code challenge sent in the authorization request.
type PKCEParams struct {
	// CodeVerifier is a cryptographically random string (43-128 characters)
	// using the characters [A-Z] / [a-z] / [0-9] / "-" / "." / "_" / "~"
	CodeVerifier string

	// CodeChallenge is the SHA-256 hash of CodeVerifier, base64url-encoded
	// without padding (RFC 4648 Section 5)
	CodeChallenge string

	// Method is the code challenge method, always "S256" for SHA-256
	Method string

	// State is a CSRF token that must be validated in the callback
	State string
}

// AccessToken represents a parsed and validated JWT access token.
//
// Access tokens are short-lived (24 hours) and used to authenticate
// API requests to AINative services.
type AccessToken struct {
	// Raw is the original JWT token string
	Raw string

	// ExpiresAt is when the token expires (from "exp" claim)
	ExpiresAt time.Time

	// UserID is the unique user identifier (from "sub" claim)
	UserID string

	// Email is the user's email address (from "email" claim)
	Email string

	// Roles are the user's roles (from "roles" claim)
	Roles []string

	// Issuer is the token issuer (from "iss" claim, should be "ainative-auth")
	Issuer string

	// Audience is the intended audience (from "aud" claim, should be "ainative-code")
	Audience string
}

// IsExpired returns true if the access token has expired.
func (t *AccessToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// IsValid performs basic validation checks on the access token.
//
// Returns true if:
//   - Token is not expired
//   - Issuer is "ainative-auth"
//   - Audience is "ainative-code"
//   - UserID is not empty
func (t *AccessToken) IsValid() bool {
	if t.IsExpired() {
		return false
	}
	if t.Issuer != "ainative-auth" {
		return false
	}
	if t.Audience != "ainative-code" {
		return false
	}
	if t.UserID == "" {
		return false
	}
	return true
}

// RefreshToken represents a parsed and validated JWT refresh token.
//
// Refresh tokens are long-lived (7 days) and used to obtain new
// access tokens without re-authentication.
type RefreshToken struct {
	// Raw is the original JWT token string
	Raw string

	// ExpiresAt is when the token expires (from "exp" claim)
	ExpiresAt time.Time

	// UserID is the unique user identifier (from "sub" claim)
	UserID string

	// SessionID is the session identifier (from "session_id" claim)
	SessionID string

	// Issuer is the token issuer (from "iss" claim, should be "ainative-auth")
	Issuer string

	// Audience is the intended audience (from "aud" claim, should be "ainative-code")
	Audience string
}

// IsExpired returns true if the refresh token has expired.
func (t *RefreshToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// IsValid performs basic validation checks on the refresh token.
//
// Returns true if:
//   - Token is not expired
//   - Issuer is "ainative-auth"
//   - Audience is "ainative-code"
//   - UserID is not empty
//   - SessionID is not empty
func (t *RefreshToken) IsValid() bool {
	if t.IsExpired() {
		return false
	}
	if t.Issuer != "ainative-auth" {
		return false
	}
	if t.Audience != "ainative-code" {
		return false
	}
	if t.UserID == "" {
		return false
	}
	if t.SessionID == "" {
		return false
	}
	return true
}

// TokenResponse contains the OAuth 2.0 token response.
//
// This is returned by the token endpoint after successful code exchange
// or refresh token usage.
type TokenResponse struct {
	// AccessToken is the parsed access token
	AccessToken *AccessToken

	// RefreshToken is the parsed refresh token (may be nil)
	RefreshToken *RefreshToken

	// ExpiresIn is the number of seconds until access token expires
	ExpiresIn int64

	// TokenType is the type of token, always "Bearer" for JWT
	TokenType string

	// Scope is the granted scopes (space-separated string)
	Scope string
}

// CallbackResult contains the result of the OAuth callback.
//
// After the user authorizes the application, the authorization server
// redirects to the callback URL with the authorization code and state.
type CallbackResult struct {
	// Code is the authorization code to exchange for tokens
	Code string

	// State is the CSRF token that must match the original request
	State string

	// Error is an error code if authorization failed
	Error string

	// ErrorDescription is a human-readable error description
	ErrorDescription string
}

// HasError returns true if the callback contains an error.
func (r *CallbackResult) HasError() bool {
	return r.Error != ""
}

// ClientOptions contains configuration for the OAuth client.
//
// Use functional options (WithXxx functions) to configure the client:
//
//	client := auth.NewClient(
//	    auth.WithClientID("ainative-code-cli"),
//	    auth.WithAuthEndpoint("https://auth.ainative.studio/oauth/authorize"),
//	    auth.WithTokenEndpoint("https://auth.ainative.studio/oauth/token"),
//	)
type ClientOptions struct {
	// ClientID is the OAuth 2.0 client identifier
	ClientID string

	// AuthEndpoint is the authorization endpoint URL
	// Default: https://auth.ainative.studio/oauth/authorize
	AuthEndpoint string

	// TokenEndpoint is the token endpoint URL
	// Default: https://auth.ainative.studio/oauth/token
	TokenEndpoint string

	// RedirectURI is the callback URL for OAuth redirect
	// Default: http://localhost:8080/callback
	RedirectURI string

	// Scopes are the OAuth 2.0 scopes to request
	// Default: ["read", "write", "offline_access"]
	Scopes []string

	// Timeout is the timeout for HTTP requests
	// Default: 30 seconds
	Timeout time.Duration

	// PublicKey is the RSA public key for JWT signature verification
	// This should be fetched from the auth server's JWKS endpoint
	PublicKey *rsa.PublicKey

	// CallbackPort is the port for the local callback server
	// Default: 8080
	CallbackPort int
}

// DefaultClientOptions returns the default client options.
func DefaultClientOptions() *ClientOptions {
	return &ClientOptions{
		ClientID:      "ainative-code-cli",
		AuthEndpoint:  "https://auth.ainative.studio/oauth/authorize",
		TokenEndpoint: "https://auth.ainative.studio/oauth/token",
		RedirectURI:   "http://localhost:8080/callback",
		Scopes:        []string{"read", "write", "offline_access"},
		Timeout:       30 * time.Second,
		CallbackPort:  8080,
	}
}

// Config is an alias for ClientOptions for backward compatibility.
type Config = ClientOptions
