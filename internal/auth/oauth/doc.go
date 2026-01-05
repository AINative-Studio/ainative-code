// Package oauth provides OAuth 2.0 authentication with PKCE (Proof Key for Code Exchange).
//
// This package implements the OAuth 2.0 Authorization Code Flow with PKCE for
// secure authentication without client secrets. It includes:
//
//   - Code verifier generation (cryptographically random, 43-128 characters)
//   - Code challenge generation (SHA-256 hash, base64url encoded)
//   - Authorization URL construction
//   - Local callback server for authorization code receipt
//   - Token exchange (authorization code for access/refresh tokens)
//   - Token storage in OS keychain
//
// OAuth 2.0 PKCE Flow:
//
//  1. Generate random code verifier (43-128 chars)
//  2. Create code challenge: BASE64URL(SHA256(code_verifier))
//  3. Redirect user to authorization URL with challenge
//  4. User authorizes on provider's site
//  5. Provider redirects to callback URL with authorization code
//  6. Exchange code + verifier for access/refresh tokens
//  7. Store tokens securely in OS keychain
//
// Example usage:
//
//	import "github.com/AINative-studio/ainative-code/internal/auth/oauth"
//
//	// Create OAuth client
//	client := oauth.NewClient(oauth.Config{
//	    AuthURL:      "https://auth.example.com/authorize",
//	    TokenURL:     "https://auth.example.com/token",
//	    ClientID:     "your-client-id",
//	    RedirectURL:  "http://localhost:8080/callback",
//	    Scopes:       []string{"read", "write"},
//	})
//
//	// Start authentication flow
//	tokens, err := client.Authenticate(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Tokens are automatically stored in OS keychain
//	fmt.Printf("Access token: %s\n", tokens.AccessToken)
package oauth
