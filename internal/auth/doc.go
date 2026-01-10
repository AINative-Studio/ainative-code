// Package auth provides OAuth 2.0 authentication with PKCE flow
// and JWT token management for AINative Code.
//
// This package implements:
//   - OAuth 2.0 Authorization Code Flow with PKCE (RFC 7636)
//   - JWT token parsing and validation (RS256)
//   - Secure token storage using OS keychain
//   - Local callback server for OAuth redirect
//
// # OAuth 2.0 PKCE Flow
//
// The PKCE (Proof Key for Code Exchange) flow prevents authorization code
// interception attacks by generating a code verifier and challenge:
//
//  1. Generate PKCE parameters (code verifier, code challenge, state)
//  2. Build authorization URL with challenge
//  3. Open browser for user authorization
//  4. Start local callback server
//  5. Receive authorization code via redirect
//  6. Validate state parameter (CSRF protection)
//  7. Exchange code + verifier for tokens
//  8. Parse and validate JWT tokens
//  9. Store tokens in OS keychain
//
// # Example Usage
//
//	client := auth.NewClient(
//	    auth.WithClientID("ainative-code-cli"),
//	    auth.WithAuthEndpoint("https://api.ainative.studio/v1/auth/login"),
//	    auth.WithTokenEndpoint("https://api.ainative.studio/v1/auth/token"),
//	)
//	defer client.Close()
//
//	tokens, err := client.Authorize(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Security Measures
//
// The package implements multiple security measures:
//   - PKCE prevents code interception (RFC 7636)
//   - State parameter prevents CSRF attacks
//   - RS256 signature verification prevents token forgery
//   - OS keychain provides encrypted token storage
//   - Single-use callback server prevents replay attacks
//
// # Token Management
//
// JWT tokens are parsed and validated with:
//   - RS256 signature verification using public key
//   - Expiry checking (access: 24h, refresh: 7d)
//   - Issuer validation ("ainative-auth")
//   - Audience validation ("ainative-code")
//
// # PKCE Generation
//
// PKCE parameters are generated using cryptographically secure random:
//   - Code verifier: 128 characters (A-Z, a-z, 0-9, -, ., _, ~)
//   - Code challenge: SHA-256 hash of verifier, base64url encoded
//   - Challenge method: "S256"
//
// # Architecture
//
// The package follows the Provider Interface pattern from TASK-023,
// using functional options for flexible configuration and thread-safe
// operations with context-based cancellation support.
//
// For detailed architecture documentation, see internal/auth/ARCHITECTURE.md.
package auth
