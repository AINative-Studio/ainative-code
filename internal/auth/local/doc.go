// Package local provides local authentication for offline operation.
//
// This package implements a fallback authentication system that works
// without network connectivity. It uses:
//
//   - SQLite for local credential storage
//   - Bcrypt for secure password hashing (12 rounds)
//   - Session management for local tokens
//   - JWT-compatible token generation for offline use
//
// The local auth system is Tier 3 in the authentication hierarchy:
//   - Tier 1: Local JWT validation (fast)
//   - Tier 2: API token validation (network-dependent)
//   - Tier 3: Local authentication fallback (offline)
//
// Features:
//   - Register new local users
//   - Authenticate with username/password
//   - Generate local access/refresh tokens
//   - Session management
//   - Password validation with bcrypt
//
// Example usage:
//
//	import "github.com/AINative-studio/ainative-code/internal/auth/local"
//
//	// Create local auth store
//	store, err := local.NewStore("path/to/db.sqlite")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Register user
//	err = store.Register("user@example.com", "secure-password")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Authenticate
//	session, err := store.Authenticate("user@example.com", "secure-password")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Use session tokens
//	fmt.Printf("Access Token: %s\n", session.AccessToken)
package local
