package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

// PKCE constants following RFC 7636 specifications
const (
	// codeVerifierLength is the length of the code verifier.
	// RFC 7636 allows 43-128 characters; we use 128 for maximum security.
	codeVerifierLength = 128

	// stateLength is the length of the state parameter in bytes.
	// 32 bytes = 256 bits provides strong CSRF protection.
	stateLength = 32

	// allowedChars contains the unreserved characters for code verifier.
	// RFC 7636 Section 4.1: [A-Z] / [a-z] / [0-9] / "-" / "." / "_" / "~"
	allowedChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~"
)

// GeneratePKCE creates PKCE parameters for OAuth 2.0 PKCE flow.
//
// This function generates:
//   - Code verifier: 128-character cryptographically random string
//   - Code challenge: SHA-256 hash of verifier, base64url-encoded (no padding)
//   - Challenge method: "S256"
//   - State: 32-byte cryptographically random CSRF token, base64url-encoded
//
// The PKCE flow prevents authorization code interception attacks by requiring
// the client to prove possession of the code verifier during token exchange.
//
// Returns ErrPKCEGeneration if cryptographic random generation fails.
//
// Example:
//
//	pkce, err := GeneratePKCE()
//	if err != nil {
//	    return fmt.Errorf("PKCE generation failed: %w", err)
//	}
//	// Use pkce.CodeChallenge in authorization URL
//	// Use pkce.CodeVerifier in token exchange
//	// Use pkce.State for CSRF validation
func GeneratePKCE() (*PKCEParams, error) {
	// Generate code verifier
	verifier, err := generateCodeVerifier()
	if err != nil {
		return nil, fmt.Errorf("failed to generate code verifier: %w", err)
	}

	// Generate code challenge from verifier
	challenge := generateCodeChallenge(verifier)

	// Generate state parameter for CSRF protection
	state, err := generateState()
	if err != nil {
		return nil, fmt.Errorf("failed to generate state parameter: %w", err)
	}

	return &PKCEParams{
		CodeVerifier:  verifier,
		CodeChallenge: challenge,
		Method:        "S256",
		State:         state,
	}, nil
}

// generateCodeVerifier creates a cryptographically random code verifier.
//
// RFC 7636 Section 4.1 specifies:
//   - Minimum length: 43 characters
//   - Maximum length: 128 characters
//   - Allowed characters: [A-Z] / [a-z] / [0-9] / "-" / "." / "_" / "~"
//
// We use 128 characters (maximum) for strongest security.
//
// Returns ErrPKCEGeneration if random generation fails.
func generateCodeVerifier() (string, error) {
	// Generate random bytes
	randomBytes := make([]byte, codeVerifierLength)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("%w: %v", ErrPKCEGeneration, err)
	}

	// Map random bytes to allowed characters
	verifier := make([]byte, codeVerifierLength)
	for i, b := range randomBytes {
		verifier[i] = allowedChars[int(b)%len(allowedChars)]
	}

	return string(verifier), nil
}

// generateCodeChallenge creates a code challenge from the code verifier.
//
// RFC 7636 Section 4.2 specifies S256 method:
//   - code_challenge = BASE64URL(SHA256(ASCII(code_verifier)))
//   - Base64URL encoding without padding (RFC 4648 Section 5)
//
// The server will compute the same hash from the verifier sent during
// token exchange and compare it to the challenge sent in authorization.
func generateCodeChallenge(verifier string) string {
	// Compute SHA-256 hash of verifier
	hash := sha256.Sum256([]byte(verifier))

	// Encode as base64url without padding
	challenge := base64.RawURLEncoding.EncodeToString(hash[:])

	return challenge
}

// generateState creates a cryptographically random state parameter.
//
// The state parameter provides CSRF protection by ensuring the authorization
// callback originates from the same client that initiated the flow.
//
// We generate 32 random bytes (256 bits) and encode as base64url for:
//   - Strong entropy against brute-force attacks
//   - URL-safe representation
//   - Compact size
//
// Returns ErrPKCEGeneration if random generation fails.
func generateState() (string, error) {
	// Generate random bytes
	randomBytes := make([]byte, stateLength)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("%w: %v", ErrPKCEGeneration, err)
	}

	// Encode as base64url without padding
	state := base64.RawURLEncoding.EncodeToString(randomBytes)

	return state, nil
}

// ValidateCodeVerifier checks if a code verifier meets RFC 7636 requirements.
//
// Returns ErrInvalidCodeVerifier if:
//   - Length is not between 43-128 characters
//   - Contains characters outside allowed set
//
// This function can be used to validate code verifiers before using them
// in token exchange requests.
func ValidateCodeVerifier(verifier string) error {
	// Check length constraints (RFC 7636 Section 4.1)
	if len(verifier) < 43 || len(verifier) > 128 {
		return fmt.Errorf("%w: length must be 43-128 characters, got %d",
			ErrInvalidCodeVerifier, len(verifier))
	}

	// Validate allowed characters
	for i, char := range verifier {
		valid := false
		for _, allowed := range allowedChars {
			if char == allowed {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("%w: invalid character at position %d: %c",
				ErrInvalidCodeVerifier, i, char)
		}
	}

	return nil
}

// ValidateCodeChallenge checks if a code challenge is valid.
//
// Returns ErrInvalidCodeChallenge if:
//   - Not a valid base64url-encoded string
//   - Length doesn't match SHA-256 output (43 characters when base64url-encoded)
//
// Note: This performs format validation only. To verify a challenge matches
// a verifier, use generateCodeChallenge(verifier) and compare the result.
func ValidateCodeChallenge(challenge string) error {
	// SHA-256 produces 32 bytes, which encodes to 43 base64url characters (no padding)
	const expectedLength = 43

	if len(challenge) != expectedLength {
		return fmt.Errorf("%w: expected %d characters for SHA-256 base64url encoding, got %d",
			ErrInvalidCodeChallenge, expectedLength, len(challenge))
	}

	// Attempt to decode to verify valid base64url
	if _, err := base64.RawURLEncoding.DecodeString(challenge); err != nil {
		return fmt.Errorf("%w: invalid base64url encoding: %v",
			ErrInvalidCodeChallenge, err)
	}

	return nil
}
