package oauth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"math/big"
)

const (
	// PKCECodeVerifierLength is the length of the code verifier (43-128 characters per RFC 7636)
	PKCECodeVerifierLength = 64

	// PKCECodeVerifierMinLength is the minimum length for a code verifier
	PKCECodeVerifierMinLength = 43

	// PKCECodeVerifierMaxLength is the maximum length for a code verifier
	PKCECodeVerifierMaxLength = 128

	// PKCEChallengeMethod is the method used to generate the code challenge
	PKCEChallengeMethod = "S256" // SHA-256
)

// PKCECodePair represents a code verifier and its corresponding challenge.
type PKCECodePair struct {
	Verifier        string
	Challenge       string
	ChallengeMethod string
}

// GeneratePKCECodePair generates a code verifier and challenge for PKCE.
//
// The code verifier is a cryptographically random string of 43-128 characters
// using the characters [A-Z], [a-z], [0-9], "-", ".", "_", and "~".
//
// The code challenge is the Base64URL-encoded SHA-256 hash of the verifier.
//
// See RFC 7636 for PKCE specification.
func GeneratePKCECodePair() (*PKCECodePair, error) {
	verifier, err := generateCodeVerifier(PKCECodeVerifierLength)
	if err != nil {
		return nil, fmt.Errorf("failed to generate code verifier: %w", err)
	}

	challenge := generateCodeChallenge(verifier)

	return &PKCECodePair{
		Verifier:        verifier,
		Challenge:       challenge,
		ChallengeMethod: PKCEChallengeMethod,
	}, nil
}

// GeneratePKCECodePairWithLength generates a PKCE code pair with custom verifier length.
func GeneratePKCECodePairWithLength(length int) (*PKCECodePair, error) {
	if length < PKCECodeVerifierMinLength || length > PKCECodeVerifierMaxLength {
		return nil, fmt.Errorf("code verifier length must be between %d and %d characters",
			PKCECodeVerifierMinLength, PKCECodeVerifierMaxLength)
	}

	verifier, err := generateCodeVerifier(length)
	if err != nil {
		return nil, fmt.Errorf("failed to generate code verifier: %w", err)
	}

	challenge := generateCodeChallenge(verifier)

	return &PKCECodePair{
		Verifier:        verifier,
		Challenge:       challenge,
		ChallengeMethod: PKCEChallengeMethod,
	}, nil
}

// generateCodeVerifier generates a cryptographically random code verifier.
//
// The verifier uses unreserved characters from RFC 3986:
// [A-Z] / [a-z] / [0-9] / "-" / "." / "_" / "~"
func generateCodeVerifier(length int) (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~"
	charsetLength := big.NewInt(int64(len(charset)))

	b := make([]byte, length)
	for i := range b {
		n, err := rand.Int(rand.Reader, charsetLength)
		if err != nil {
			return "", fmt.Errorf("failed to generate random number: %w", err)
		}
		b[i] = charset[n.Int64()]
	}

	return string(b), nil
}

// generateCodeChallenge generates a code challenge from a verifier.
//
// The challenge is the Base64URL-encoded SHA-256 hash of the verifier,
// without padding.
func generateCodeChallenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

// ValidateCodeVerifier validates that a code verifier meets PKCE requirements.
func ValidateCodeVerifier(verifier string) error {
	if len(verifier) < PKCECodeVerifierMinLength {
		return fmt.Errorf("code verifier is too short: %d characters (minimum %d)",
			len(verifier), PKCECodeVerifierMinLength)
	}

	if len(verifier) > PKCECodeVerifierMaxLength {
		return fmt.Errorf("code verifier is too long: %d characters (maximum %d)",
			len(verifier), PKCECodeVerifierMaxLength)
	}

	// Validate characters
	const allowedChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~"
	for _, c := range verifier {
		valid := false
		for _, allowed := range allowedChars {
			if c == allowed {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("code verifier contains invalid character: %c", c)
		}
	}

	return nil
}

// VerifyCodeChallenge verifies that a challenge matches a verifier.
func VerifyCodeChallenge(verifier, challenge string) bool {
	expectedChallenge := generateCodeChallenge(verifier)
	return challenge == expectedChallenge
}
