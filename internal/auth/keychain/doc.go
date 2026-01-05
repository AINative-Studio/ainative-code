// Package keychain provides cross-platform secure credential storage.
//
// This package provides a unified interface for storing and retrieving
// sensitive credentials (tokens, API keys) using OS-level secure storage:
//
//   - macOS: Keychain Services
//   - Linux: Secret Service (libsecret)
//   - Windows: Credential Manager (via Windows Credential API)
//
// Features:
//   - Store access tokens securely
//   - Store refresh tokens securely
//   - Store API keys and other credentials
//   - Retrieve credentials with error handling
//   - Delete credentials on logout
//   - Cross-platform unified API
//
// The package automatically detects the operating system and uses the
// appropriate backend. All storage is encrypted by the OS.
//
// Example usage:
//
//	import "github.com/AINative-studio/ainative-code/internal/auth/keychain"
//
//	// Get the platform-specific keychain
//	kc := keychain.Get()
//
//	// Store tokens
//	err := kc.SetAccessToken("my-access-token")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Retrieve tokens
//	token, err := kc.GetAccessToken()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Delete all credentials
//	err = kc.DeleteAll()
package keychain
