package auth

import (
	"testing"
	"time"
)

// TestAccessTokenIsExpired tests the AccessToken.IsExpired method.
func TestAccessTokenIsExpired(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		expiresAt time.Time
		want      bool
	}{
		{
			name:      "expired token (1 hour ago)",
			expiresAt: now.Add(-1 * time.Hour),
			want:      true,
		},
		{
			name:      "expired token (1 second ago)",
			expiresAt: now.Add(-1 * time.Second),
			want:      true,
		},
		{
			name:      "valid token (1 hour from now)",
			expiresAt: now.Add(1 * time.Hour),
			want:      false,
		},
		{
			name:      "valid token (1 second from now)",
			expiresAt: now.Add(1 * time.Second),
			want:      false,
		},
		{
			name:      "valid token (24 hours from now)",
			expiresAt: now.Add(24 * time.Hour),
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := &AccessToken{
				ExpiresAt: tt.expiresAt,
			}

			got := token.IsExpired()
			if got != tt.want {
				t.Errorf("AccessToken.IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestAccessTokenIsValid tests the AccessToken.IsValid method.
func TestAccessTokenIsValid(t *testing.T) {
	now := time.Now()
	validExpiry := now.Add(24 * time.Hour)
	expiredTime := now.Add(-1 * time.Hour)

	tests := []struct {
		name  string
		token *AccessToken
		want  bool
	}{
		{
			name: "valid token with all fields",
			token: &AccessToken{
				Raw:       "valid.jwt.token",
				ExpiresAt: validExpiry,
				UserID:    "user-123",
				Email:     "user@example.com",
				Roles:     []string{"user", "admin"},
				Issuer:    "ainative-auth",
				Audience:  "ainative-code",
			},
			want: true,
		},
		{
			name: "expired token",
			token: &AccessToken{
				Raw:       "expired.jwt.token",
				ExpiresAt: expiredTime,
				UserID:    "user-123",
				Email:     "user@example.com",
				Issuer:    "ainative-auth",
				Audience:  "ainative-code",
			},
			want: false,
		},
		{
			name: "invalid issuer",
			token: &AccessToken{
				Raw:       "invalid.jwt.token",
				ExpiresAt: validExpiry,
				UserID:    "user-123",
				Email:     "user@example.com",
				Issuer:    "wrong-issuer",
				Audience:  "ainative-code",
			},
			want: false,
		},
		{
			name: "invalid audience",
			token: &AccessToken{
				Raw:       "invalid.jwt.token",
				ExpiresAt: validExpiry,
				UserID:    "user-123",
				Email:     "user@example.com",
				Issuer:    "ainative-auth",
				Audience:  "wrong-audience",
			},
			want: false,
		},
		{
			name: "empty user ID",
			token: &AccessToken{
				Raw:       "invalid.jwt.token",
				ExpiresAt: validExpiry,
				UserID:    "",
				Email:     "user@example.com",
				Issuer:    "ainative-auth",
				Audience:  "ainative-code",
			},
			want: false,
		},
		{
			name: "empty issuer",
			token: &AccessToken{
				Raw:       "invalid.jwt.token",
				ExpiresAt: validExpiry,
				UserID:    "user-123",
				Email:     "user@example.com",
				Issuer:    "",
				Audience:  "ainative-code",
			},
			want: false,
		},
		{
			name: "empty audience",
			token: &AccessToken{
				Raw:       "invalid.jwt.token",
				ExpiresAt: validExpiry,
				UserID:    "user-123",
				Email:     "user@example.com",
				Issuer:    "ainative-auth",
				Audience:  "",
			},
			want: false,
		},
		{
			name: "multiple invalid fields",
			token: &AccessToken{
				Raw:       "invalid.jwt.token",
				ExpiresAt: expiredTime,
				UserID:    "",
				Email:     "",
				Issuer:    "wrong-issuer",
				Audience:  "wrong-audience",
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

// TestRefreshTokenIsExpired tests the RefreshToken.IsExpired method.
func TestRefreshTokenIsExpired(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		expiresAt time.Time
		want      bool
	}{
		{
			name:      "expired token (1 day ago)",
			expiresAt: now.Add(-24 * time.Hour),
			want:      true,
		},
		{
			name:      "expired token (1 second ago)",
			expiresAt: now.Add(-1 * time.Second),
			want:      true,
		},
		{
			name:      "valid token (1 day from now)",
			expiresAt: now.Add(24 * time.Hour),
			want:      false,
		},
		{
			name:      "valid token (1 second from now)",
			expiresAt: now.Add(1 * time.Second),
			want:      false,
		},
		{
			name:      "valid token (7 days from now)",
			expiresAt: now.Add(7 * 24 * time.Hour),
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := &RefreshToken{
				ExpiresAt: tt.expiresAt,
			}

			got := token.IsExpired()
			if got != tt.want {
				t.Errorf("RefreshToken.IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestRefreshTokenIsValid tests the RefreshToken.IsValid method.
func TestRefreshTokenIsValid(t *testing.T) {
	now := time.Now()
	validExpiry := now.Add(7 * 24 * time.Hour)
	expiredTime := now.Add(-1 * time.Hour)

	tests := []struct {
		name  string
		token *RefreshToken
		want  bool
	}{
		{
			name: "valid token with all fields",
			token: &RefreshToken{
				Raw:       "valid.refresh.token",
				ExpiresAt: validExpiry,
				UserID:    "user-123",
				SessionID: "session-456",
				Issuer:    "ainative-auth",
				Audience:  "ainative-code",
			},
			want: true,
		},
		{
			name: "expired token",
			token: &RefreshToken{
				Raw:       "expired.refresh.token",
				ExpiresAt: expiredTime,
				UserID:    "user-123",
				SessionID: "session-456",
				Issuer:    "ainative-auth",
				Audience:  "ainative-code",
			},
			want: false,
		},
		{
			name: "invalid issuer",
			token: &RefreshToken{
				Raw:       "invalid.refresh.token",
				ExpiresAt: validExpiry,
				UserID:    "user-123",
				SessionID: "session-456",
				Issuer:    "wrong-issuer",
				Audience:  "ainative-code",
			},
			want: false,
		},
		{
			name: "invalid audience",
			token: &RefreshToken{
				Raw:       "invalid.refresh.token",
				ExpiresAt: validExpiry,
				UserID:    "user-123",
				SessionID: "session-456",
				Issuer:    "ainative-auth",
				Audience:  "wrong-audience",
			},
			want: false,
		},
		{
			name: "empty user ID",
			token: &RefreshToken{
				Raw:       "invalid.refresh.token",
				ExpiresAt: validExpiry,
				UserID:    "",
				SessionID: "session-456",
				Issuer:    "ainative-auth",
				Audience:  "ainative-code",
			},
			want: false,
		},
		{
			name: "empty session ID",
			token: &RefreshToken{
				Raw:       "invalid.refresh.token",
				ExpiresAt: validExpiry,
				UserID:    "user-123",
				SessionID: "",
				Issuer:    "ainative-auth",
				Audience:  "ainative-code",
			},
			want: false,
		},
		{
			name: "empty issuer",
			token: &RefreshToken{
				Raw:       "invalid.refresh.token",
				ExpiresAt: validExpiry,
				UserID:    "user-123",
				SessionID: "session-456",
				Issuer:    "",
				Audience:  "ainative-code",
			},
			want: false,
		},
		{
			name: "empty audience",
			token: &RefreshToken{
				Raw:       "invalid.refresh.token",
				ExpiresAt: validExpiry,
				UserID:    "user-123",
				SessionID: "session-456",
				Issuer:    "ainative-auth",
				Audience:  "",
			},
			want: false,
		},
		{
			name: "multiple invalid fields",
			token: &RefreshToken{
				Raw:       "invalid.refresh.token",
				ExpiresAt: expiredTime,
				UserID:    "",
				SessionID: "",
				Issuer:    "wrong-issuer",
				Audience:  "wrong-audience",
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

// TestCallbackResultHasError tests the CallbackResult.HasError method.
func TestCallbackResultHasError(t *testing.T) {
	tests := []struct {
		name   string
		result *CallbackResult
		want   bool
	}{
		{
			name: "no error",
			result: &CallbackResult{
				Code:  "auth-code-123",
				State: "state-456",
				Error: "",
			},
			want: false,
		},
		{
			name: "with error code",
			result: &CallbackResult{
				Code:             "",
				State:            "state-456",
				Error:            "access_denied",
				ErrorDescription: "User denied access",
			},
			want: true,
		},
		{
			name: "error code only",
			result: &CallbackResult{
				Error: "invalid_request",
			},
			want: true,
		},
		{
			name: "empty result",
			result: &CallbackResult{
				Code:  "",
				State: "",
				Error: "",
			},
			want: false,
		},
		{
			name: "error with code and state",
			result: &CallbackResult{
				Code:             "auth-code-123",
				State:            "state-456",
				Error:            "server_error",
				ErrorDescription: "Internal server error",
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.result.HasError()
			if got != tt.want {
				t.Errorf("CallbackResult.HasError() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestDefaultClientOptions tests the DefaultClientOptions function.
func TestDefaultClientOptions(t *testing.T) {
	opts := DefaultClientOptions()

	// Verify default values
	if opts.ClientID != "ainative-code-cli" {
		t.Errorf("ClientID = %q, want %q", opts.ClientID, "ainative-code-cli")
	}

	if opts.AuthEndpoint != "https://auth.ainative.studio/oauth/authorize" {
		t.Errorf("AuthEndpoint = %q, want %q", opts.AuthEndpoint, "https://auth.ainative.studio/oauth/authorize")
	}

	if opts.TokenEndpoint != "https://auth.ainative.studio/oauth/token" {
		t.Errorf("TokenEndpoint = %q, want %q", opts.TokenEndpoint, "https://auth.ainative.studio/oauth/token")
	}

	if opts.RedirectURI != "http://localhost:8080/callback" {
		t.Errorf("RedirectURI = %q, want %q", opts.RedirectURI, "http://localhost:8080/callback")
	}

	if len(opts.Scopes) != 3 {
		t.Errorf("Scopes length = %d, want 3", len(opts.Scopes))
	} else {
		expectedScopes := []string{"read", "write", "offline_access"}
		for i, scope := range expectedScopes {
			if opts.Scopes[i] != scope {
				t.Errorf("Scopes[%d] = %q, want %q", i, opts.Scopes[i], scope)
			}
		}
	}

	if opts.Timeout != 30*time.Second {
		t.Errorf("Timeout = %v, want %v", opts.Timeout, 30*time.Second)
	}

	if opts.CallbackPort != 8080 {
		t.Errorf("CallbackPort = %d, want %d", opts.CallbackPort, 8080)
	}

	// Verify PublicKey is nil by default
	if opts.PublicKey != nil {
		t.Error("PublicKey should be nil by default")
	}
}

// TestAccessTokenValidation tests edge cases for AccessToken validation.
func TestAccessTokenValidation(t *testing.T) {
	now := time.Now()
	validExpiry := now.Add(24 * time.Hour)

	// Test that email and roles are not required for IsValid
	token := &AccessToken{
		Raw:       "valid.jwt.token",
		ExpiresAt: validExpiry,
		UserID:    "user-123",
		Email:     "", // Empty email
		Roles:     nil, // Nil roles
		Issuer:    "ainative-auth",
		Audience:  "ainative-code",
	}

	if !token.IsValid() {
		t.Error("Token should be valid even with empty email and nil roles")
	}

	// Test Raw field is not required for IsValid
	token2 := &AccessToken{
		Raw:       "", // Empty raw token
		ExpiresAt: validExpiry,
		UserID:    "user-123",
		Issuer:    "ainative-auth",
		Audience:  "ainative-code",
	}

	if !token2.IsValid() {
		t.Error("Token should be valid even with empty Raw field")
	}
}

// TestRefreshTokenValidation tests edge cases for RefreshToken validation.
func TestRefreshTokenValidation(t *testing.T) {
	now := time.Now()
	validExpiry := now.Add(7 * 24 * time.Hour)

	// Test that Raw field is not required for IsValid
	token := &RefreshToken{
		Raw:       "", // Empty raw token
		ExpiresAt: validExpiry,
		UserID:    "user-123",
		SessionID: "session-456",
		Issuer:    "ainative-auth",
		Audience:  "ainative-code",
	}

	if !token.IsValid() {
		t.Error("Token should be valid even with empty Raw field")
	}
}

// TestTokenResponseStructure tests the TokenResponse structure.
func TestTokenResponseStructure(t *testing.T) {
	now := time.Now()

	accessToken := &AccessToken{
		Raw:       "access.token.jwt",
		ExpiresAt: now.Add(24 * time.Hour),
		UserID:    "user-123",
		Email:     "user@example.com",
		Issuer:    "ainative-auth",
		Audience:  "ainative-code",
	}

	refreshToken := &RefreshToken{
		Raw:       "refresh.token.jwt",
		ExpiresAt: now.Add(7 * 24 * time.Hour),
		UserID:    "user-123",
		SessionID: "session-456",
		Issuer:    "ainative-auth",
		Audience:  "ainative-code",
	}

	response := &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    86400,
		TokenType:    "Bearer",
		Scope:        "read write offline_access",
	}

	// Verify structure can be created and fields accessed
	if response.AccessToken.UserID != "user-123" {
		t.Errorf("AccessToken.UserID = %q, want %q", response.AccessToken.UserID, "user-123")
	}

	if response.RefreshToken.SessionID != "session-456" {
		t.Errorf("RefreshToken.SessionID = %q, want %q", response.RefreshToken.SessionID, "session-456")
	}

	if response.ExpiresIn != 86400 {
		t.Errorf("ExpiresIn = %d, want %d", response.ExpiresIn, 86400)
	}

	if response.TokenType != "Bearer" {
		t.Errorf("TokenType = %q, want %q", response.TokenType, "Bearer")
	}

	// Test nil refresh token case
	response2 := &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: nil, // No refresh token
		ExpiresIn:    86400,
		TokenType:    "Bearer",
	}

	if response2.RefreshToken != nil {
		t.Error("RefreshToken should be nil")
	}
}

// TestClientOptionsAlias tests that Config is an alias for ClientOptions.
func TestClientOptionsAlias(t *testing.T) {
	// This test verifies that Config type exists as an alias
	var config *Config
	config = DefaultClientOptions()

	if config.ClientID != "ainative-code-cli" {
		t.Errorf("Config.ClientID = %q, want %q", config.ClientID, "ainative-code-cli")
	}
}
