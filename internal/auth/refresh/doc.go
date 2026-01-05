// Package refresh provides automatic token refresh functionality.
//
// This package implements background token monitoring and automatic refresh
// to ensure tokens remain valid without user intervention. It monitors token
// expiration times and proactively refreshes tokens before they expire.
//
// Features:
//   - Background goroutine for expiration monitoring
//   - Configurable refresh threshold (default: 5 minutes before expiry)
//   - Automatic token refresh using OAuth client
//   - Token storage via callback function
//   - Refresh failure handling with re-authentication callbacks
//   - Graceful shutdown support
//
// Example usage:
//
//	import (
//	    "github.com/AINative-studio/ainative-code/internal/auth/refresh"
//	    "github.com/AINative-studio/ainative-code/internal/auth/oauth"
//	)
//
//	// Create OAuth client
//	oauthClient := oauth.NewClient(config)
//
//	// Create refresh manager
//	manager := refresh.NewManager(refresh.Config{
//	    OAuthClient:   oauthClient,
//	    TokenStore:    storeTokenFunc,
//	    OnRefreshFail: handleRefreshFailure,
//	})
//
//	// Start monitoring with current tokens
//	manager.Start(ctx, tokens)
//
//	// Manager automatically refreshes tokens before expiry
//	// Call Stop() for graceful shutdown
//	defer manager.Stop()
package refresh
