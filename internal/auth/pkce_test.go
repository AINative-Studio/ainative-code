package auth

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"
	"testing"
)

// TestGeneratePKCE tests the main PKCE generation function.
func TestGeneratePKCE(t *testing.T) {
	pkce, err := GeneratePKCE()
	if err != nil {
		t.Fatalf("GeneratePKCE() failed: %v", err)
	}

	// Verify PKCEParams is not nil
	if pkce == nil {
		t.Fatal("GeneratePKCE() returned nil params")
	}

	// Verify code verifier length (RFC 7636: 43-128 characters)
	if len(pkce.CodeVerifier) != codeVerifierLength {
		t.Errorf("CodeVerifier length = %d, want %d", len(pkce.CodeVerifier), codeVerifierLength)
	}

	// Verify code verifier uses only allowed characters
	for i, char := range pkce.CodeVerifier {
		if !strings.ContainsRune(allowedChars, char) {
			t.Errorf("CodeVerifier contains invalid character at position %d: %c", i, char)
		}
	}

	// Verify code challenge length (SHA-256 -> 32 bytes -> 43 base64url chars)
	const expectedChallengeLength = 43
	if len(pkce.CodeChallenge) != expectedChallengeLength {
		t.Errorf("CodeChallenge length = %d, want %d", len(pkce.CodeChallenge), expectedChallengeLength)
	}

	// Verify code challenge is valid base64url
	if _, err := base64.RawURLEncoding.DecodeString(pkce.CodeChallenge); err != nil {
		t.Errorf("CodeChallenge is not valid base64url: %v", err)
	}

	// Verify method is S256
	if pkce.Method != "S256" {
		t.Errorf("Method = %q, want %q", pkce.Method, "S256")
	}

	// Verify state is not empty
	if pkce.State == "" {
		t.Error("State is empty")
	}

	// Verify state is valid base64url
	if _, err := base64.RawURLEncoding.DecodeString(pkce.State); err != nil {
		t.Errorf("State is not valid base64url: %v", err)
	}

	// Verify code challenge matches verifier
	expectedChallenge := generateCodeChallenge(pkce.CodeVerifier)
	if pkce.CodeChallenge != expectedChallenge {
		t.Error("CodeChallenge does not match expected challenge for verifier")
	}
}

// TestGeneratePKCEUniqueness verifies multiple PKCE generations produce unique values.
func TestGeneratePKCEUniqueness(t *testing.T) {
	const iterations = 100

	verifiers := make(map[string]bool)
	challenges := make(map[string]bool)
	states := make(map[string]bool)

	for i := 0; i < iterations; i++ {
		pkce, err := GeneratePKCE()
		if err != nil {
			t.Fatalf("GeneratePKCE() iteration %d failed: %v", i, err)
		}

		// Check verifier uniqueness
		if verifiers[pkce.CodeVerifier] {
			t.Errorf("Duplicate code verifier generated at iteration %d", i)
		}
		verifiers[pkce.CodeVerifier] = true

		// Check challenge uniqueness
		if challenges[pkce.CodeChallenge] {
			t.Errorf("Duplicate code challenge generated at iteration %d", i)
		}
		challenges[pkce.CodeChallenge] = true

		// Check state uniqueness
		if states[pkce.State] {
			t.Errorf("Duplicate state generated at iteration %d", i)
		}
		states[pkce.State] = true
	}

	// Verify we got expected number of unique values
	if len(verifiers) != iterations {
		t.Errorf("Generated %d unique verifiers, want %d", len(verifiers), iterations)
	}
	if len(challenges) != iterations {
		t.Errorf("Generated %d unique challenges, want %d", len(challenges), iterations)
	}
	if len(states) != iterations {
		t.Errorf("Generated %d unique states, want %d", len(states), iterations)
	}
}

// TestCodeChallengeDeterminism verifies the same verifier always produces the same challenge.
func TestCodeChallengeDeterminism(t *testing.T) {
	verifier := "test-verifier-123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678"

	// Generate challenge multiple times
	challenge1 := generateCodeChallenge(verifier)
	challenge2 := generateCodeChallenge(verifier)
	challenge3 := generateCodeChallenge(verifier)

	// All challenges should be identical
	if challenge1 != challenge2 || challenge1 != challenge3 {
		t.Error("generateCodeChallenge is not deterministic")
	}

	// Verify challenge is computed correctly
	hash := sha256.Sum256([]byte(verifier))
	expected := base64.RawURLEncoding.EncodeToString(hash[:])
	if challenge1 != expected {
		t.Errorf("Challenge = %q, want %q", challenge1, expected)
	}
}

// TestValidateCodeVerifier tests code verifier validation.
func TestValidateCodeVerifier(t *testing.T) {
	tests := []struct {
		name      string
		verifier  string
		wantError bool
		errType   error
	}{
		{
			name:      "valid minimum length (43 chars)",
			verifier:  "abcdefghijklmnopqrstuvwxyz0123456789-._~abc",
			wantError: false,
		},
		{
			name:      "valid medium length (64 chars)",
			verifier:  "abcdefghijklmnopqrstuvwxyz0123456789-._~ABCDEFGHIJKLMNOPQRSTUVW",
			wantError: false,
		},
		{
			name:      "valid maximum length (128 chars)",
			verifier:  strings.Repeat("a", 128),
			wantError: false,
		},
		{
			name:      "valid with all allowed characters",
			verifier:  "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~0123456789-._~0123456789",
			wantError: false,
		},
		{
			name:      "too short (42 chars)",
			verifier:  strings.Repeat("a", 42),
			wantError: true,
			errType:   ErrInvalidCodeVerifier,
		},
		{
			name:      "too long (129 chars)",
			verifier:  strings.Repeat("a", 129),
			wantError: true,
			errType:   ErrInvalidCodeVerifier,
		},
		{
			name:      "empty string",
			verifier:  "",
			wantError: true,
			errType:   ErrInvalidCodeVerifier,
		},
		{
			name:      "contains invalid character (space)",
			verifier:  "abcdefghijklmnopqrstuvwxyz 0123456789abc",
			wantError: true,
			errType:   ErrInvalidCodeVerifier,
		},
		{
			name:      "contains invalid character (@)",
			verifier:  "abcdefghijklmnopqrstuvwxyz@0123456789abc",
			wantError: true,
			errType:   ErrInvalidCodeVerifier,
		},
		{
			name:      "contains invalid character (!)",
			verifier:  "abcdefghijklmnopqrstuvwxyz!0123456789abc",
			wantError: true,
			errType:   ErrInvalidCodeVerifier,
		},
		{
			name:      "contains invalid character (=)",
			verifier:  "abcdefghijklmnopqrstuvwxyz=0123456789abc",
			wantError: true,
			errType:   ErrInvalidCodeVerifier,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCodeVerifier(tt.verifier)

			if tt.wantError {
				if err == nil {
					t.Error("ValidateCodeVerifier() expected error, got nil")
				} else if tt.errType != nil && !errors.Is(err, tt.errType) {
					t.Errorf("ValidateCodeVerifier() error = %v, want error type %v", err, tt.errType)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateCodeVerifier() unexpected error: %v", err)
				}
			}
		})
	}
}

// TestValidateCodeChallenge tests code challenge validation.
func TestValidateCodeChallenge(t *testing.T) {
	// Generate a valid challenge for testing
	validVerifier := strings.Repeat("a", 128)
	validChallenge := generateCodeChallenge(validVerifier)

	tests := []struct {
		name      string
		challenge string
		wantError bool
		errType   error
	}{
		{
			name:      "valid challenge",
			challenge: validChallenge,
			wantError: false,
		},
		{
			name:      "valid base64url (43 chars)",
			challenge: "E3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			wantError: true, // wrong length (64 chars)
			errType:   ErrInvalidCodeChallenge,
		},
		{
			name:      "too short (42 chars)",
			challenge: strings.Repeat("a", 42),
			wantError: true,
			errType:   ErrInvalidCodeChallenge,
		},
		{
			name:      "too long (44 chars)",
			challenge: strings.Repeat("a", 44),
			wantError: true,
			errType:   ErrInvalidCodeChallenge,
		},
		{
			name:      "empty string",
			challenge: "",
			wantError: true,
			errType:   ErrInvalidCodeChallenge,
		},
		{
			name:      "invalid base64url (contains =)",
			challenge: "abcdefghijklmnopqrstuvwxyz01234567890123=",
			wantError: true,
			errType:   ErrInvalidCodeChallenge,
		},
		{
			name:      "invalid base64url (contains space)",
			challenge: "abcdefghijklmnopqrstuvwxyz0123456789012 ",
			wantError: true,
			errType:   ErrInvalidCodeChallenge,
		},
		{
			name:      "invalid base64url (contains invalid char)",
			challenge: "abcdefghijklmnopqrstuvwxyz0123456789012!",
			wantError: true,
			errType:   ErrInvalidCodeChallenge,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCodeChallenge(tt.challenge)

			if tt.wantError {
				if err == nil {
					t.Error("ValidateCodeChallenge() expected error, got nil")
				} else if tt.errType != nil && !errors.Is(err, tt.errType) {
					t.Errorf("ValidateCodeChallenge() error = %v, want error type %v", err, tt.errType)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateCodeChallenge() unexpected error: %v", err)
				}
			}
		})
	}
}

// TestGenerateCodeVerifier tests internal code verifier generation.
func TestGenerateCodeVerifier(t *testing.T) {
	verifier, err := generateCodeVerifier()
	if err != nil {
		t.Fatalf("generateCodeVerifier() failed: %v", err)
	}

	// Verify length
	if len(verifier) != codeVerifierLength {
		t.Errorf("verifier length = %d, want %d", len(verifier), codeVerifierLength)
	}

	// Verify only allowed characters
	for i, char := range verifier {
		if !strings.ContainsRune(allowedChars, char) {
			t.Errorf("verifier contains invalid character at position %d: %c", i, char)
		}
	}

	// Verify validation passes
	if err := ValidateCodeVerifier(verifier); err != nil {
		t.Errorf("Generated verifier failed validation: %v", err)
	}
}

// TestGenerateCodeChallenge tests internal code challenge generation.
func TestGenerateCodeChallenge(t *testing.T) {
	tests := []struct {
		name     string
		verifier string
	}{
		{
			name:     "minimum length verifier",
			verifier: strings.Repeat("a", 43),
		},
		{
			name:     "medium length verifier",
			verifier: strings.Repeat("b", 64),
		},
		{
			name:     "maximum length verifier",
			verifier: strings.Repeat("c", 128),
		},
		{
			name:     "mixed characters",
			verifier: "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~0123456789-._~0123456789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			challenge := generateCodeChallenge(tt.verifier)

			// Verify challenge length (SHA-256 -> 32 bytes -> 43 base64url chars)
			if len(challenge) != 43 {
				t.Errorf("challenge length = %d, want 43", len(challenge))
			}

			// Verify challenge is valid base64url
			decoded, err := base64.RawURLEncoding.DecodeString(challenge)
			if err != nil {
				t.Errorf("challenge is not valid base64url: %v", err)
			}

			// Verify decoded hash is 32 bytes (SHA-256 output)
			if len(decoded) != 32 {
				t.Errorf("decoded hash length = %d, want 32", len(decoded))
			}

			// Verify challenge matches manual computation
			hash := sha256.Sum256([]byte(tt.verifier))
			expected := base64.RawURLEncoding.EncodeToString(hash[:])
			if challenge != expected {
				t.Errorf("challenge = %q, want %q", challenge, expected)
			}

			// Verify validation passes
			if err := ValidateCodeChallenge(challenge); err != nil {
				t.Errorf("Generated challenge failed validation: %v", err)
			}
		})
	}
}

// TestGenerateState tests internal state generation.
func TestGenerateState(t *testing.T) {
	state, err := generateState()
	if err != nil {
		t.Fatalf("generateState() failed: %v", err)
	}

	// Verify state is not empty
	if state == "" {
		t.Error("generateState() returned empty string")
	}

	// Verify state is valid base64url
	decoded, err := base64.RawURLEncoding.DecodeString(state)
	if err != nil {
		t.Errorf("state is not valid base64url: %v", err)
	}

	// Verify decoded state is stateLength bytes
	if len(decoded) != stateLength {
		t.Errorf("decoded state length = %d, want %d", len(decoded), stateLength)
	}
}

// TestGenerateStateUniqueness verifies multiple state generations produce unique values.
func TestGenerateStateUniqueness(t *testing.T) {
	const iterations = 100
	states := make(map[string]bool)

	for i := 0; i < iterations; i++ {
		state, err := generateState()
		if err != nil {
			t.Fatalf("generateState() iteration %d failed: %v", i, err)
		}

		if states[state] {
			t.Errorf("Duplicate state generated at iteration %d", i)
		}
		states[state] = true
	}

	// Verify we got expected number of unique values
	if len(states) != iterations {
		t.Errorf("Generated %d unique states, want %d", len(states), iterations)
	}
}

// TestPKCEIntegration tests the full PKCE flow integration.
func TestPKCEIntegration(t *testing.T) {
	// Generate PKCE parameters
	pkce, err := GeneratePKCE()
	if err != nil {
		t.Fatalf("GeneratePKCE() failed: %v", err)
	}

	// Validate verifier
	if err := ValidateCodeVerifier(pkce.CodeVerifier); err != nil {
		t.Errorf("Generated verifier failed validation: %v", err)
	}

	// Validate challenge
	if err := ValidateCodeChallenge(pkce.CodeChallenge); err != nil {
		t.Errorf("Generated challenge failed validation: %v", err)
	}

	// Verify challenge matches verifier
	expectedChallenge := generateCodeChallenge(pkce.CodeVerifier)
	if pkce.CodeChallenge != expectedChallenge {
		t.Error("Challenge does not match verifier")
	}

	// Verify method is S256
	if pkce.Method != "S256" {
		t.Errorf("Method = %q, want S256", pkce.Method)
	}
}
