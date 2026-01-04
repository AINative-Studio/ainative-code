package auth

import (
	"testing"
	"time"
)

// TestTokenPair_IsValid tests the TokenPair.IsValid method with table-driven tests.
func TestTokenPair_IsValid(t *testing.T) {
	now := time.Now()
	validAccessToken := &AccessToken{
		Raw:       "valid.access.token",
		ExpiresAt: now.Add(1 * time.Hour),
		UserID:    "user-123",
		Email:     "user@example.com",
		Issuer:    "ainative-auth",
		Audience:  "ainative-code",
	}
	validRefreshToken := &RefreshToken{
		Raw:       "valid.refresh.token",
		ExpiresAt: now.Add(7 * 24 * time.Hour),
		UserID:    "user-123",
		SessionID: "session-456",
		Issuer:    "ainative-auth",
		Audience:  "ainative-code",
	}
	expiredAccessToken := &AccessToken{
		Raw:       "expired.access.token",
		ExpiresAt: now.Add(-1 * time.Hour),
		UserID:    "user-123",
		Email:     "user@example.com",
		Issuer:    "ainative-auth",
		Audience:  "ainative-code",
	}
	expiredRefreshToken := &RefreshToken{
		Raw:       "expired.refresh.token",
		ExpiresAt: now.Add(-1 * time.Hour),
		UserID:    "user-123",
		SessionID: "session-456",
		Issuer:    "ainative-auth",
		Audience:  "ainative-code",
	}

	tests := []struct {
		name      string
		tokenPair *TokenPair
		want      bool
	}{
		{
			name: "valid token pair",
			tokenPair: &TokenPair{
				AccessToken:  validAccessToken,
				RefreshToken: validRefreshToken,
				ReceivedAt:   now,
			},
			want: true,
		},
		{
			name:      "nil token pair",
			tokenPair: nil,
			want:      false,
		},
		{
			name: "nil access token",
			tokenPair: &TokenPair{
				AccessToken:  nil,
				RefreshToken: validRefreshToken,
				ReceivedAt:   now,
			},
			want: false,
		},
		{
			name: "nil refresh token",
			tokenPair: &TokenPair{
				AccessToken:  validAccessToken,
				RefreshToken: nil,
				ReceivedAt:   now,
			},
			want: false,
		},
		{
			name: "expired access token",
			tokenPair: &TokenPair{
				AccessToken:  expiredAccessToken,
				RefreshToken: validRefreshToken,
				ReceivedAt:   now,
			},
			want: false,
		},
		{
			name: "expired refresh token",
			tokenPair: &TokenPair{
				AccessToken:  validAccessToken,
				RefreshToken: expiredRefreshToken,
				ReceivedAt:   now,
			},
			want: false,
		},
		{
			name: "both tokens expired",
			tokenPair: &TokenPair{
				AccessToken:  expiredAccessToken,
				RefreshToken: expiredRefreshToken,
				ReceivedAt:   now,
			},
			want: false,
		},
		{
			name: "access token expiring in 30 seconds (edge case)",
			tokenPair: &TokenPair{
				AccessToken: &AccessToken{
					Raw:       "soon.expired.token",
					ExpiresAt: now.Add(30 * time.Second),
					UserID:    "user-123",
					Email:     "user@example.com",
					Issuer:    "ainative-auth",
					Audience:  "ainative-code",
				},
				RefreshToken: validRefreshToken,
				ReceivedAt:   now,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.tokenPair.IsValid()
			if got != tt.want {
				t.Errorf("TokenPair.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestTokenPair_NeedsRefresh tests the TokenPair.NeedsRefresh method with table-driven tests.
func TestTokenPair_NeedsRefresh(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		tokenPair *TokenPair
		want      bool
	}{
		{
			name:      "nil token pair",
			tokenPair: nil,
			want:      true,
		},
		{
			name: "nil access token",
			tokenPair: &TokenPair{
				AccessToken: nil,
				RefreshToken: &RefreshToken{
					Raw:       "valid.refresh.token",
					ExpiresAt: now.Add(7 * 24 * time.Hour),
					UserID:    "user-123",
					SessionID: "session-456",
					Issuer:    "ainative-auth",
					Audience:  "ainative-code",
				},
			},
			want: true,
		},
		{
			name: "access token expires in 1 hour (no refresh needed)",
			tokenPair: &TokenPair{
				AccessToken: &AccessToken{
					Raw:       "valid.access.token",
					ExpiresAt: now.Add(1 * time.Hour),
					UserID:    "user-123",
					Email:     "user@example.com",
					Issuer:    "ainative-auth",
					Audience:  "ainative-code",
				},
				ReceivedAt: now,
			},
			want: false,
		},
		{
			name: "access token expires in 6 minutes (no refresh needed)",
			tokenPair: &TokenPair{
				AccessToken: &AccessToken{
					Raw:       "valid.access.token",
					ExpiresAt: now.Add(6 * time.Minute),
					UserID:    "user-123",
					Email:     "user@example.com",
					Issuer:    "ainative-auth",
					Audience:  "ainative-code",
				},
				ReceivedAt: now,
			},
			want: false,
		},
		{
			name: "access token expires in exactly 5 minutes (threshold - refresh needed)",
			tokenPair: &TokenPair{
				AccessToken: &AccessToken{
					Raw:       "valid.access.token",
					ExpiresAt: now.Add(5 * time.Minute),
					UserID:    "user-123",
					Email:     "user@example.com",
					Issuer:    "ainative-auth",
					Audience:  "ainative-code",
				},
				ReceivedAt: now,
			},
			want: true,
		},
		{
			name: "access token expires in 4 minutes (refresh needed)",
			tokenPair: &TokenPair{
				AccessToken: &AccessToken{
					Raw:       "soon.expired.token",
					ExpiresAt: now.Add(4 * time.Minute),
					UserID:    "user-123",
					Email:     "user@example.com",
					Issuer:    "ainative-auth",
					Audience:  "ainative-code",
				},
				ReceivedAt: now,
			},
			want: true,
		},
		{
			name: "access token expires in 1 minute (refresh needed)",
			tokenPair: &TokenPair{
				AccessToken: &AccessToken{
					Raw:       "soon.expired.token",
					ExpiresAt: now.Add(1 * time.Minute),
					UserID:    "user-123",
					Email:     "user@example.com",
					Issuer:    "ainative-auth",
					Audience:  "ainative-code",
				},
				ReceivedAt: now,
			},
			want: true,
		},
		{
			name: "access token expired (refresh needed)",
			tokenPair: &TokenPair{
				AccessToken: &AccessToken{
					Raw:       "expired.access.token",
					ExpiresAt: now.Add(-1 * time.Hour),
					UserID:    "user-123",
					Email:     "user@example.com",
					Issuer:    "ainative-auth",
					Audience:  "ainative-code",
				},
				ReceivedAt: now,
			},
			want: true,
		},
		{
			name: "access token expires in 30 seconds (refresh needed)",
			tokenPair: &TokenPair{
				AccessToken: &AccessToken{
					Raw:       "soon.expired.token",
					ExpiresAt: now.Add(30 * time.Second),
					UserID:    "user-123",
					Email:     "user@example.com",
					Issuer:    "ainative-auth",
					Audience:  "ainative-code",
				},
				ReceivedAt: now,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.tokenPair.NeedsRefresh()
			if got != tt.want {
				t.Errorf("TokenPair.NeedsRefresh() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestTokenPair_EdgeCases tests edge cases and boundary conditions.
func TestTokenPair_EdgeCases(t *testing.T) {
	now := time.Now()

	t.Run("exactly at 5 minute threshold", func(t *testing.T) {
		tp := &TokenPair{
			AccessToken: &AccessToken{
				Raw:       "token",
				ExpiresAt: now.Add(5*time.Minute + 1*time.Millisecond),
				UserID:    "user-123",
				Email:     "user@example.com",
				Issuer:    "ainative-auth",
				Audience:  "ainative-code",
			},
			ReceivedAt: now,
		}

		// Just over 5 minutes should not need refresh
		if tp.NeedsRefresh() {
			t.Error("Token expiring in just over 5 minutes should not need refresh")
		}
	})

	t.Run("just under 5 minute threshold", func(t *testing.T) {
		tp := &TokenPair{
			AccessToken: &AccessToken{
				Raw:       "token",
				ExpiresAt: now.Add(5*time.Minute - 1*time.Millisecond),
				UserID:    "user-123",
				Email:     "user@example.com",
				Issuer:    "ainative-auth",
				Audience:  "ainative-code",
			},
			ReceivedAt: now,
		}

		// Just under 5 minutes should need refresh
		if !tp.NeedsRefresh() {
			t.Error("Token expiring in just under 5 minutes should need refresh")
		}
	})

	t.Run("empty token pair", func(t *testing.T) {
		tp := &TokenPair{}

		if tp.IsValid() {
			t.Error("Empty TokenPair should not be valid")
		}

		if !tp.NeedsRefresh() {
			t.Error("Empty TokenPair should need refresh")
		}
	})
}
