package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Test helper: Generate RSA key pair for testing
func generateTestKeyPair(t *testing.T) (*rsa.PrivateKey, *rsa.PublicKey) {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key pair: %v", err)
	}
	return privateKey, &privateKey.PublicKey
}

// Test helper: Create a signed JWT token with custom claims
func createTestToken(t *testing.T, privateKey *rsa.PrivateKey, claims jwt.MapClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("Failed to sign test token: %v", err)
	}
	return tokenString
}

// TestParseAccessToken_Valid tests parsing of valid access tokens.
func TestParseAccessToken_Valid(t *testing.T) {
	privateKey, publicKey := generateTestKeyPair(t)

	tests := []struct {
		name   string
		claims jwt.MapClaims
	}{
		{
			name: "standard access token with roles",
			claims: jwt.MapClaims{
				"iss":   "ainative-auth",
				"aud":   []string{"ainative-code"},
				"sub":   "user-123",
				"email": "user@example.com",
				"roles": []interface{}{"admin", "user"},
				"exp":   time.Now().Add(1 * time.Hour).Unix(),
			},
		},
		{
			name: "access token without roles",
			claims: jwt.MapClaims{
				"iss":   "ainative-auth",
				"aud":   []string{"ainative-code"},
				"sub":   "user-456",
				"email": "test@example.com",
				"exp":   time.Now().Add(24 * time.Hour).Unix(),
			},
		},
		{
			name: "access token with single role",
			claims: jwt.MapClaims{
				"iss":   "ainative-auth",
				"aud":   []string{"ainative-code"},
				"sub":   "user-789",
				"email": "admin@example.com",
				"roles": "admin",
				"exp":   time.Now().Add(1 * time.Hour).Unix(),
			},
		},
		{
			name: "access token with empty roles array",
			claims: jwt.MapClaims{
				"iss":   "ainative-auth",
				"aud":   []string{"ainative-code"},
				"sub":   "user-abc",
				"email": "newuser@example.com",
				"roles": []interface{}{},
				"exp":   time.Now().Add(1 * time.Hour).Unix(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenString := createTestToken(t, privateKey, tt.claims)

			accessToken, err := ParseAccessToken(tokenString, publicKey)
			if err != nil {
				t.Fatalf("ParseAccessToken() unexpected error: %v", err)
			}

			// Verify token is not nil
			if accessToken == nil {
				t.Fatal("ParseAccessToken() returned nil token")
			}

			// Verify raw token is set
			if accessToken.Raw != tokenString {
				t.Error("AccessToken.Raw does not match input token string")
			}

			// Verify issuer
			if accessToken.Issuer != "ainative-auth" {
				t.Errorf("Issuer = %q, want %q", accessToken.Issuer, "ainative-auth")
			}

			// Verify audience
			if accessToken.Audience != "ainative-code" {
				t.Errorf("Audience = %q, want %q", accessToken.Audience, "ainative-code")
			}

			// Verify subject
			expectedSub := tt.claims["sub"].(string)
			if accessToken.UserID != expectedSub {
				t.Errorf("UserID = %q, want %q", accessToken.UserID, expectedSub)
			}

			// Verify email
			expectedEmail := tt.claims["email"].(string)
			if accessToken.Email != expectedEmail {
				t.Errorf("Email = %q, want %q", accessToken.Email, expectedEmail)
			}

			// Verify expiration time is set and in the future
			if accessToken.ExpiresAt.IsZero() {
				t.Error("ExpiresAt is zero")
			}
			if accessToken.IsExpired() {
				t.Error("Token should not be expired")
			}

			// Verify token is valid
			if !accessToken.IsValid() {
				t.Error("AccessToken.IsValid() = false, want true")
			}

			// Verify roles if present
			if rolesInterface, ok := tt.claims["roles"]; ok {
				switch roles := rolesInterface.(type) {
				case string:
					if len(accessToken.Roles) != 1 || accessToken.Roles[0] != roles {
						t.Errorf("Roles = %v, want [%s]", accessToken.Roles, roles)
					}
				case []interface{}:
					if len(accessToken.Roles) != len(roles) {
						t.Errorf("Roles length = %d, want %d", len(accessToken.Roles), len(roles))
					}
				}
			} else {
				// No roles claim, should have empty slice
				if len(accessToken.Roles) != 0 {
					t.Errorf("Roles = %v, want empty slice", accessToken.Roles)
				}
			}
		})
	}
}

// TestParseAccessToken_Errors tests error cases for access token parsing.
func TestParseAccessToken_Errors(t *testing.T) {
	privateKey, publicKey := generateTestKeyPair(t)
	wrongPrivateKey, _ := generateTestKeyPair(t)

	tests := []struct {
		name        string
		tokenString string
		publicKey   *rsa.PublicKey
		claims      jwt.MapClaims
		wrongKey    bool
		expectError error
	}{
		{
			name:        "nil public key",
			publicKey:   nil,
			expectError: ErrMissingPublicKey,
		},
		{
			name: "invalid signature",
			claims: jwt.MapClaims{
				"iss":   "ainative-auth",
				"aud":   []string{"ainative-code"},
				"sub":   "user-123",
				"email": "user@example.com",
				"exp":   time.Now().Add(1 * time.Hour).Unix(),
			},
			wrongKey:    true,
			expectError: ErrInvalidSignature,
		},
		{
			name:        "malformed token",
			tokenString: "not.a.valid.jwt",
			expectError: ErrTokenParseFailed,
		},
		{
			name:        "empty token",
			tokenString: "",
			expectError: ErrTokenParseFailed,
		},
		{
			name: "expired token",
			claims: jwt.MapClaims{
				"iss":   "ainative-auth",
				"aud":   []string{"ainative-code"},
				"sub":   "user-123",
				"email": "user@example.com",
				"exp":   time.Now().Add(-1 * time.Hour).Unix(),
			},
			expectError: ErrTokenExpired,
		},
		{
			name: "missing issuer",
			claims: jwt.MapClaims{
				"aud":   []string{"ainative-code"},
				"sub":   "user-123",
				"email": "user@example.com",
				"exp":   time.Now().Add(1 * time.Hour).Unix(),
			},
			expectError: ErrInvalidClaims,
		},
		{
			name: "invalid issuer",
			claims: jwt.MapClaims{
				"iss":   "wrong-issuer",
				"aud":   []string{"ainative-code"},
				"sub":   "user-123",
				"email": "user@example.com",
				"exp":   time.Now().Add(1 * time.Hour).Unix(),
			},
			expectError: ErrInvalidIssuer,
		},
		{
			name: "missing audience",
			claims: jwt.MapClaims{
				"iss":   "ainative-auth",
				"sub":   "user-123",
				"email": "user@example.com",
				"exp":   time.Now().Add(1 * time.Hour).Unix(),
			},
			expectError: ErrInvalidClaims,
		},
		{
			name: "empty audience array",
			claims: jwt.MapClaims{
				"iss":   "ainative-auth",
				"aud":   []string{},
				"sub":   "user-123",
				"email": "user@example.com",
				"exp":   time.Now().Add(1 * time.Hour).Unix(),
			},
			expectError: ErrInvalidAudience,
		},
		{
			name: "invalid audience",
			claims: jwt.MapClaims{
				"iss":   "ainative-auth",
				"aud":   []string{"wrong-audience"},
				"sub":   "user-123",
				"email": "user@example.com",
				"exp":   time.Now().Add(1 * time.Hour).Unix(),
			},
			expectError: ErrInvalidAudience,
		},
		{
			name: "missing subject",
			claims: jwt.MapClaims{
				"iss":   "ainative-auth",
				"aud":   []string{"ainative-code"},
				"email": "user@example.com",
				"exp":   time.Now().Add(1 * time.Hour).Unix(),
			},
			expectError: ErrInvalidClaims,
		},
		{
			name: "empty subject",
			claims: jwt.MapClaims{
				"iss":   "ainative-auth",
				"aud":   []string{"ainative-code"},
				"sub":   "",
				"email": "user@example.com",
				"exp":   time.Now().Add(1 * time.Hour).Unix(),
			},
			expectError: ErrInvalidClaims,
		},
		{
			name: "missing email",
			claims: jwt.MapClaims{
				"iss": "ainative-auth",
				"aud": []string{"ainative-code"},
				"sub": "user-123",
				"exp": time.Now().Add(1 * time.Hour).Unix(),
			},
			expectError: ErrInvalidClaims,
		},
		{
			name: "empty email",
			claims: jwt.MapClaims{
				"iss":   "ainative-auth",
				"aud":   []string{"ainative-code"},
				"sub":   "user-123",
				"email": "",
				"exp":   time.Now().Add(1 * time.Hour).Unix(),
			},
			expectError: ErrInvalidClaims,
		},
		{
			name: "missing expiration",
			claims: jwt.MapClaims{
				"iss":   "ainative-auth",
				"aud":   []string{"ainative-code"},
				"sub":   "user-123",
				"email": "user@example.com",
			},
			expectError: ErrInvalidClaims,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tokenString string
			var pubKey *rsa.PublicKey

			// Use provided token string if available
			if tt.tokenString != "" {
				tokenString = tt.tokenString
				pubKey = publicKey
			} else if tt.publicKey != nil {
				pubKey = tt.publicKey
			} else {
				pubKey = publicKey
			}

			// Create token if claims provided
			if tt.claims != nil {
				keyToUse := privateKey
				if tt.wrongKey {
					keyToUse = wrongPrivateKey
				}
				tokenString = createTestToken(t, keyToUse, tt.claims)
			}

			// Parse token
			accessToken, err := ParseAccessToken(tokenString, pubKey)

			// Verify error occurred
			if err == nil {
				t.Fatal("ParseAccessToken() expected error, got nil")
			}

			// Verify error type
			if !errors.Is(err, tt.expectError) {
				t.Errorf("ParseAccessToken() error = %v, want error type %v", err, tt.expectError)
			}

			// Verify nil token on error
			if accessToken != nil {
				t.Error("ParseAccessToken() should return nil token on error")
			}
		})
	}
}

// TestParseRefreshToken_Valid tests parsing of valid refresh tokens.
func TestParseRefreshToken_Valid(t *testing.T) {
	privateKey, publicKey := generateTestKeyPair(t)

	tests := []struct {
		name   string
		claims jwt.MapClaims
	}{
		{
			name: "standard refresh token",
			claims: jwt.MapClaims{
				"iss":        "ainative-auth",
				"aud":        []string{"ainative-code"},
				"sub":        "user-123",
				"session_id": "session-abc-123",
				"exp":        time.Now().Add(7 * 24 * time.Hour).Unix(),
			},
		},
		{
			name: "refresh token with long session ID",
			claims: jwt.MapClaims{
				"iss":        "ainative-auth",
				"aud":        []string{"ainative-code"},
				"sub":        "user-456",
				"session_id": "very-long-session-id-with-uuid-12345678-1234-1234-1234-123456789012",
				"exp":        time.Now().Add(7 * 24 * time.Hour).Unix(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenString := createTestToken(t, privateKey, tt.claims)

			refreshToken, err := ParseRefreshToken(tokenString, publicKey)
			if err != nil {
				t.Fatalf("ParseRefreshToken() unexpected error: %v", err)
			}

			// Verify token is not nil
			if refreshToken == nil {
				t.Fatal("ParseRefreshToken() returned nil token")
			}

			// Verify raw token is set
			if refreshToken.Raw != tokenString {
				t.Error("RefreshToken.Raw does not match input token string")
			}

			// Verify issuer
			if refreshToken.Issuer != "ainative-auth" {
				t.Errorf("Issuer = %q, want %q", refreshToken.Issuer, "ainative-auth")
			}

			// Verify audience
			if refreshToken.Audience != "ainative-code" {
				t.Errorf("Audience = %q, want %q", refreshToken.Audience, "ainative-code")
			}

			// Verify subject
			expectedSub := tt.claims["sub"].(string)
			if refreshToken.UserID != expectedSub {
				t.Errorf("UserID = %q, want %q", refreshToken.UserID, expectedSub)
			}

			// Verify session ID
			expectedSessionID := tt.claims["session_id"].(string)
			if refreshToken.SessionID != expectedSessionID {
				t.Errorf("SessionID = %q, want %q", refreshToken.SessionID, expectedSessionID)
			}

			// Verify expiration time is set and in the future
			if refreshToken.ExpiresAt.IsZero() {
				t.Error("ExpiresAt is zero")
			}
			if refreshToken.IsExpired() {
				t.Error("Token should not be expired")
			}

			// Verify token is valid
			if !refreshToken.IsValid() {
				t.Error("RefreshToken.IsValid() = false, want true")
			}
		})
	}
}

// TestParseRefreshToken_Errors tests error cases for refresh token parsing.
func TestParseRefreshToken_Errors(t *testing.T) {
	privateKey, publicKey := generateTestKeyPair(t)
	wrongPrivateKey, _ := generateTestKeyPair(t)

	tests := []struct {
		name        string
		tokenString string
		publicKey   *rsa.PublicKey
		claims      jwt.MapClaims
		wrongKey    bool
		expectError error
	}{
		{
			name:        "nil public key",
			publicKey:   nil,
			expectError: ErrMissingPublicKey,
		},
		{
			name: "invalid signature",
			claims: jwt.MapClaims{
				"iss":        "ainative-auth",
				"aud":        []string{"ainative-code"},
				"sub":        "user-123",
				"session_id": "session-abc",
				"exp":        time.Now().Add(7 * 24 * time.Hour).Unix(),
			},
			wrongKey:    true,
			expectError: ErrInvalidSignature,
		},
		{
			name:        "malformed token",
			tokenString: "invalid.jwt.token",
			expectError: ErrTokenParseFailed,
		},
		{
			name: "expired token",
			claims: jwt.MapClaims{
				"iss":        "ainative-auth",
				"aud":        []string{"ainative-code"},
				"sub":        "user-123",
				"session_id": "session-abc",
				"exp":        time.Now().Add(-1 * time.Hour).Unix(),
			},
			expectError: ErrTokenExpired,
		},
		{
			name: "missing session_id",
			claims: jwt.MapClaims{
				"iss": "ainative-auth",
				"aud": []string{"ainative-code"},
				"sub": "user-123",
				"exp": time.Now().Add(7 * 24 * time.Hour).Unix(),
			},
			expectError: ErrInvalidClaims,
		},
		{
			name: "empty session_id",
			claims: jwt.MapClaims{
				"iss":        "ainative-auth",
				"aud":        []string{"ainative-code"},
				"sub":        "user-123",
				"session_id": "",
				"exp":        time.Now().Add(7 * 24 * time.Hour).Unix(),
			},
			expectError: ErrInvalidClaims,
		},
		{
			name: "invalid issuer",
			claims: jwt.MapClaims{
				"iss":        "wrong-issuer",
				"aud":        []string{"ainative-code"},
				"sub":        "user-123",
				"session_id": "session-abc",
				"exp":        time.Now().Add(7 * 24 * time.Hour).Unix(),
			},
			expectError: ErrInvalidIssuer,
		},
		{
			name: "invalid audience",
			claims: jwt.MapClaims{
				"iss":        "ainative-auth",
				"aud":        []string{"wrong-audience"},
				"sub":        "user-123",
				"session_id": "session-abc",
				"exp":        time.Now().Add(7 * 24 * time.Hour).Unix(),
			},
			expectError: ErrInvalidAudience,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tokenString string
			var pubKey *rsa.PublicKey

			// Use provided token string if available
			if tt.tokenString != "" {
				tokenString = tt.tokenString
				pubKey = publicKey
			} else if tt.publicKey != nil {
				pubKey = tt.publicKey
			} else {
				pubKey = publicKey
			}

			// Create token if claims provided
			if tt.claims != nil {
				keyToUse := privateKey
				if tt.wrongKey {
					keyToUse = wrongPrivateKey
				}
				tokenString = createTestToken(t, keyToUse, tt.claims)
			}

			// Parse token
			refreshToken, err := ParseRefreshToken(tokenString, pubKey)

			// Verify error occurred
			if err == nil {
				t.Fatal("ParseRefreshToken() expected error, got nil")
			}

			// Verify error type
			if !errors.Is(err, tt.expectError) {
				t.Errorf("ParseRefreshToken() error = %v, want error type %v", err, tt.expectError)
			}

			// Verify nil token on error
			if refreshToken != nil {
				t.Error("ParseRefreshToken() should return nil token on error")
			}
		})
	}
}

// TestExtractStringClaim tests the extractStringClaim helper function.
func TestExtractStringClaim(t *testing.T) {
	tests := []struct {
		name      string
		claims    jwt.MapClaims
		key       string
		want      string
		wantError bool
	}{
		{
			name:   "valid string claim",
			claims: jwt.MapClaims{"email": "user@example.com"},
			key:    "email",
			want:   "user@example.com",
		},
		{
			name:      "missing claim",
			claims:    jwt.MapClaims{},
			key:       "email",
			wantError: true,
		},
		{
			name:      "empty string",
			claims:    jwt.MapClaims{"email": ""},
			key:       "email",
			wantError: true,
		},
		{
			name:      "wrong type (integer)",
			claims:    jwt.MapClaims{"email": 123},
			key:       "email",
			wantError: true,
		},
		{
			name:      "wrong type (boolean)",
			claims:    jwt.MapClaims{"email": true},
			key:       "email",
			wantError: true,
		},
		{
			name:      "wrong type (array)",
			claims:    jwt.MapClaims{"email": []string{"test@example.com"}},
			key:       "email",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractStringClaim(tt.claims, tt.key)

			if tt.wantError {
				if err == nil {
					t.Error("extractStringClaim() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("extractStringClaim() unexpected error: %v", err)
				}
				if got != tt.want {
					t.Errorf("extractStringClaim() = %q, want %q", got, tt.want)
				}
			}
		})
	}
}

// TestExtractStringSliceClaim tests the extractStringSliceClaim helper function.
func TestExtractStringSliceClaim(t *testing.T) {
	tests := []struct {
		name      string
		claims    jwt.MapClaims
		key       string
		want      []string
		wantError bool
	}{
		{
			name:   "valid string slice",
			claims: jwt.MapClaims{"roles": []interface{}{"admin", "user"}},
			key:    "roles",
			want:   []string{"admin", "user"},
		},
		{
			name:   "single string (converted to slice)",
			claims: jwt.MapClaims{"roles": "admin"},
			key:    "roles",
			want:   []string{"admin"},
		},
		{
			name:   "empty slice",
			claims: jwt.MapClaims{"roles": []interface{}{}},
			key:    "roles",
			want:   []string{},
		},
		{
			name:   "missing claim (returns empty slice)",
			claims: jwt.MapClaims{},
			key:    "roles",
			want:   []string{},
		},
		{
			name:      "wrong type (integer)",
			claims:    jwt.MapClaims{"roles": 123},
			key:       "roles",
			wantError: true,
		},
		{
			name:      "slice with non-string elements",
			claims:    jwt.MapClaims{"roles": []interface{}{"admin", 123, "user"}},
			key:       "roles",
			wantError: true,
		},
		{
			name:      "slice with mixed types",
			claims:    jwt.MapClaims{"roles": []interface{}{"admin", true}},
			key:       "roles",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractStringSliceClaim(tt.claims, tt.key)

			if tt.wantError {
				if err == nil {
					t.Error("extractStringSliceClaim() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("extractStringSliceClaim() unexpected error: %v", err)
				}
				if len(got) != len(tt.want) {
					t.Errorf("extractStringSliceClaim() length = %d, want %d", len(got), len(tt.want))
				}
				for i, v := range tt.want {
					if i >= len(got) || got[i] != v {
						t.Errorf("extractStringSliceClaim()[%d] = %q, want %q", i, got[i], v)
					}
				}
			}
		})
	}
}

// TestAccessToken_IsExpired tests the AccessToken.IsExpired method.
func TestAccessToken_IsExpired(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt time.Time
		want      bool
	}{
		{
			name:      "not expired (future)",
			expiresAt: time.Now().Add(1 * time.Hour),
			want:      false,
		},
		{
			name:      "expired (past)",
			expiresAt: time.Now().Add(-1 * time.Hour),
			want:      true,
		},
		{
			name:      "expired (just now)",
			expiresAt: time.Now().Add(-1 * time.Second),
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := &AccessToken{ExpiresAt: tt.expiresAt}
			got := token.IsExpired()
			if got != tt.want {
				t.Errorf("AccessToken.IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestAccessToken_IsValid tests the AccessToken.IsValid method.
func TestAccessToken_IsValid(t *testing.T) {
	tests := []struct {
		name  string
		token *AccessToken
		want  bool
	}{
		{
			name: "valid token",
			token: &AccessToken{
				ExpiresAt: time.Now().Add(1 * time.Hour),
				Issuer:    "ainative-auth",
				Audience:  "ainative-code",
				UserID:    "user-123",
			},
			want: true,
		},
		{
			name: "expired token",
			token: &AccessToken{
				ExpiresAt: time.Now().Add(-1 * time.Hour),
				Issuer:    "ainative-auth",
				Audience:  "ainative-code",
				UserID:    "user-123",
			},
			want: false,
		},
		{
			name: "wrong issuer",
			token: &AccessToken{
				ExpiresAt: time.Now().Add(1 * time.Hour),
				Issuer:    "wrong-issuer",
				Audience:  "ainative-code",
				UserID:    "user-123",
			},
			want: false,
		},
		{
			name: "wrong audience",
			token: &AccessToken{
				ExpiresAt: time.Now().Add(1 * time.Hour),
				Issuer:    "ainative-auth",
				Audience:  "wrong-audience",
				UserID:    "user-123",
			},
			want: false,
		},
		{
			name: "empty user ID",
			token: &AccessToken{
				ExpiresAt: time.Now().Add(1 * time.Hour),
				Issuer:    "ainative-auth",
				Audience:  "ainative-code",
				UserID:    "",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.token.IsValid()
			if got != tt.want {
				t.Errorf("AccessToken.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestRefreshToken_IsExpired tests the RefreshToken.IsExpired method.
func TestRefreshToken_IsExpired(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt time.Time
		want      bool
	}{
		{
			name:      "not expired (future)",
			expiresAt: time.Now().Add(7 * 24 * time.Hour),
			want:      false,
		},
		{
			name:      "expired (past)",
			expiresAt: time.Now().Add(-1 * time.Hour),
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := &RefreshToken{ExpiresAt: tt.expiresAt}
			got := token.IsExpired()
			if got != tt.want {
				t.Errorf("RefreshToken.IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestRefreshToken_IsValid tests the RefreshToken.IsValid method.
func TestRefreshToken_IsValid(t *testing.T) {
	tests := []struct {
		name  string
		token *RefreshToken
		want  bool
	}{
		{
			name: "valid token",
			token: &RefreshToken{
				ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
				Issuer:    "ainative-auth",
				Audience:  "ainative-code",
				UserID:    "user-123",
				SessionID: "session-abc",
			},
			want: true,
		},
		{
			name: "expired token",
			token: &RefreshToken{
				ExpiresAt: time.Now().Add(-1 * time.Hour),
				Issuer:    "ainative-auth",
				Audience:  "ainative-code",
				UserID:    "user-123",
				SessionID: "session-abc",
			},
			want: false,
		},
		{
			name: "wrong issuer",
			token: &RefreshToken{
				ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
				Issuer:    "wrong-issuer",
				Audience:  "ainative-code",
				UserID:    "user-123",
				SessionID: "session-abc",
			},
			want: false,
		},
		{
			name: "wrong audience",
			token: &RefreshToken{
				ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
				Issuer:    "ainative-auth",
				Audience:  "wrong-audience",
				UserID:    "user-123",
				SessionID: "session-abc",
			},
			want: false,
		},
		{
			name: "empty user ID",
			token: &RefreshToken{
				ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
				Issuer:    "ainative-auth",
				Audience:  "ainative-code",
				UserID:    "",
				SessionID: "session-abc",
			},
			want: false,
		},
		{
			name: "empty session ID",
			token: &RefreshToken{
				ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
				Issuer:    "ainative-auth",
				Audience:  "ainative-code",
				UserID:    "user-123",
				SessionID: "",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.token.IsValid()
			if got != tt.want {
				t.Errorf("RefreshToken.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}
