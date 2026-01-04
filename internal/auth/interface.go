package auth

import (
	"context"
	"time"
)

// Client defines the interface for OAuth 2.0 authentication operations.
//
// This interface abstracts the OAuth 2.0 Authorization Code Flow with PKCE,
// providing methods for user authentication, token management, and secure
// storage integration.
//
// Implementation Requirements:
//   - Thread-safe operations
//   - Context-aware for cancellation and timeouts
//   - Secure token storage via OS keychain
//   - Automatic token refresh when possible
//   - PKCE (RFC 7636) support for all authorization flows
//
// Typical Usage Flow:
//
//	// 1. Create client
//	client, err := NewClient(config, options...)
//
//	// 2. Check for existing valid tokens
//	tokens, err := client.GetStoredTokens(ctx)
//	if err == nil && tokens.IsValid() {
//	    return tokens, nil
//	}
//
//	// 3. Authenticate user if needed
//	tokens, err = client.Authenticate(ctx)
//
//	// 4. Use tokens for API requests
//	// If access token expires, refresh it
//	if tokens.AccessToken.IsExpired() {
//	    tokens, err = client.RefreshToken(ctx, tokens.RefreshToken)
//	}
type Client interface {
	// Authenticate initiates the OAuth 2.0 Authorization Code Flow with PKCE.
	//
	// This method:
	//   1. Generates PKCE parameters (code verifier, challenge, state)
	//   2. Builds authorization URL with required parameters
	//   3. Opens the user's browser to the authorization URL
	//   4. Starts local callback server to receive authorization code
	//   5. Waits for user to complete authorization
	//   6. Exchanges authorization code for tokens
	//   7. Stores tokens securely in OS keychain
	//   8. Returns the token pair
	//
	// The authorization flow respects context cancellation and timeouts.
	// Default timeout is 5 minutes, but can be configured via options.
	//
	// Returns:
	//   - ErrAuthorizationDenied if user denies authorization
	//   - ErrAuthorizationTimeout if user doesn't complete flow in time
	//   - ErrInvalidState if state parameter doesn't match (CSRF attack)
	//   - ErrCodeExchangeFailed if token endpoint rejects the code
	//   - ErrCallbackServerStart if local server fails to start
	//   - ErrBrowserOpen if browser cannot be opened
	//
	// Example:
	//
	//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	//	defer cancel()
	//
	//	tokens, err := client.Authenticate(ctx)
	//	if err != nil {
	//	    return fmt.Errorf("authentication failed: %w", err)
	//	}
	//	fmt.Printf("Authenticated as: %s\n", tokens.AccessToken.Email)
	Authenticate(ctx context.Context) (*TokenPair, error)

	// RefreshToken exchanges a refresh token for a new access token.
	//
	// This method:
	//   1. Sends refresh token to token endpoint
	//   2. Receives new access token (and optionally new refresh token)
	//   3. Updates tokens in OS keychain
	//   4. Returns updated token pair
	//
	// If the refresh token is expired or revoked, this method returns an error
	// and the client should re-authenticate using Authenticate().
	//
	// Returns:
	//   - ErrTokenExpired if refresh token has expired
	//   - ErrCodeExchangeFailed if token endpoint rejects refresh token
	//   - ErrHTTPRequest if network request fails
	//   - ErrHTTPResponse if server returns error status
	//
	// Example:
	//
	//	tokens, err := client.GetStoredTokens(ctx)
	//	if err != nil {
	//	    return client.Authenticate(ctx)
	//	}
	//
	//	if tokens.AccessToken.IsExpired() {
	//	    tokens, err = client.RefreshToken(ctx, tokens.RefreshToken)
	//	    if err != nil {
	//	        // Refresh failed, re-authenticate
	//	        return client.Authenticate(ctx)
	//	    }
	//	}
	RefreshToken(ctx context.Context, refreshToken *RefreshToken) (*TokenPair, error)

	// GetStoredTokens retrieves tokens from OS keychain.
	//
	// This method:
	//   1. Reads token data from secure keychain storage
	//   2. Parses and validates JWT tokens
	//   3. Returns token pair if found and valid
	//
	// The returned tokens may be expired. Callers should check
	// token.AccessToken.IsExpired() and refresh if needed.
	//
	// Returns:
	//   - ErrKeychainNotFound if no tokens are stored
	//   - ErrKeychainAccess if keychain access is denied
	//   - ErrKeychainRetrieve if reading from keychain fails
	//   - ErrTokenParseFailed if stored token data is corrupted
	//
	// Example:
	//
	//	tokens, err := client.GetStoredTokens(ctx)
	//	if errors.Is(err, auth.ErrKeychainNotFound) {
	//	    // No stored tokens, authenticate
	//	    return client.Authenticate(ctx)
	//	}
	//	if err != nil {
	//	    return fmt.Errorf("failed to get tokens: %w", err)
	//	}
	GetStoredTokens(ctx context.Context) (*TokenPair, error)

	// StoreTokens saves tokens to OS keychain.
	//
	// This method:
	//   1. Validates token format and expiration
	//   2. Encrypts token data using OS keychain
	//   3. Stores in platform-specific secure storage
	//
	// Tokens are stored with service identifier "ainative-code"
	// and account identifier derived from user ID.
	//
	// Returns:
	//   - ErrKeychainAccess if keychain access is denied
	//   - ErrKeychainStore if writing to keychain fails
	//
	// Example:
	//
	//	tokens, err := client.Authenticate(ctx)
	//	if err != nil {
	//	    return err
	//	}
	//	// Tokens are automatically stored by Authenticate(),
	//	// but you can manually store them if needed:
	//	if err := client.StoreTokens(ctx, tokens); err != nil {
	//	    log.Printf("Failed to store tokens: %v", err)
	//	}
	StoreTokens(ctx context.Context, tokens *TokenPair) error

	// ClearTokens removes tokens from OS keychain.
	//
	// This method:
	//   1. Deletes token data from OS keychain
	//   2. Clears any in-memory token cache
	//
	// Use this method to implement logout functionality.
	//
	// Returns:
	//   - ErrKeychainAccess if keychain access is denied
	//   - ErrKeychainDelete if deletion fails
	//
	// Note: This does NOT revoke tokens on the server. To fully logout,
	// you should also call the server's token revocation endpoint.
	//
	// Example:
	//
	//	if err := client.ClearTokens(ctx); err != nil {
	//	    log.Printf("Failed to clear tokens: %v", err)
	//	}
	ClearTokens(ctx context.Context) error

	// ValidateToken checks if an access token is valid and not expired.
	//
	// This method:
	//   1. Checks token expiration time
	//   2. Verifies JWT signature using public key
	//   3. Validates standard claims (iss, aud, sub, exp)
	//   4. Optionally performs introspection at server (if configured)
	//
	// Returns true if token is valid and can be used for API requests.
	// Returns false if token is expired, has invalid signature, or fails
	// server-side validation.
	//
	// This is a convenience method. Callers can also check
	// token.IsExpired() directly for simple expiration checks.
	//
	// Example:
	//
	//	tokens, _ := client.GetStoredTokens(ctx)
	//	if !client.ValidateToken(ctx, tokens.AccessToken) {
	//	    // Token invalid, refresh or re-authenticate
	//	    tokens, err = client.RefreshToken(ctx, tokens.RefreshToken)
	//	}
	ValidateToken(ctx context.Context, token *AccessToken) bool
}

// TokenPair represents a pair of access and refresh tokens.
//
// Both tokens are JWT tokens signed with RS256. The access token is
// short-lived (typically 1 hour) and used for API authentication.
// The refresh token is long-lived (typically 30 days) and used to
// obtain new access tokens without re-authentication.
type TokenPair struct {
	// AccessToken is the JWT access token used for API authentication.
	AccessToken *AccessToken

	// RefreshToken is the JWT refresh token used to obtain new access tokens.
	RefreshToken *RefreshToken

	// ReceivedAt is when this token pair was received from the server.
	// Used for calculating token age and refresh timing.
	ReceivedAt time.Time
}

// IsValid checks if the token pair is valid and usable.
//
// A token pair is considered valid if:
//   - Both access and refresh tokens are non-nil
//   - Access token is not expired (or expires in more than 1 minute)
//   - Refresh token is not expired
//
// This method provides a quick check for token validity.
// For more detailed validation, use Client.ValidateToken().
func (tp *TokenPair) IsValid() bool {
	if tp == nil || tp.AccessToken == nil || tp.RefreshToken == nil {
		return false
	}

	// Access token should not be expired (with 1 minute buffer)
	if tp.AccessToken.IsExpired() {
		return false
	}

	// Refresh token should not be expired
	if tp.RefreshToken.IsExpired() {
		return false
	}

	return true
}

// NeedsRefresh checks if the access token should be refreshed.
//
// Returns true if the access token:
//   - Is expired, OR
//   - Will expire within the next 5 minutes
//
// This provides a safety buffer to avoid using tokens that might
// expire during a long-running API request.
func (tp *TokenPair) NeedsRefresh() bool {
	if tp == nil || tp.AccessToken == nil {
		return true
	}

	// Refresh if token expires within 5 minutes
	refreshThreshold := 5 * time.Minute
	return time.Until(tp.AccessToken.ExpiresAt) < refreshThreshold
}
