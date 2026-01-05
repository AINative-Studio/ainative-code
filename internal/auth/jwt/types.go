package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	// Issuer is the token issuer
	Issuer = "ainative-auth"

	// Audience is the intended token audience
	Audience = "ainative-code"

	// AccessTokenDuration is the lifetime of an access token
	AccessTokenDuration = 24 * time.Hour

	// RefreshTokenDuration is the lifetime of a refresh token
	RefreshTokenDuration = 7 * 24 * time.Hour

	// SigningMethod is the algorithm used for signing tokens
	SigningMethod = "RS256"
)

// AccessTokenClaims represents the claims in an access token.
type AccessTokenClaims struct {
	// Standard JWT claims
	jwt.RegisteredClaims

	// UserID is the unique identifier for the user
	UserID string `json:"user_id"`

	// Email is the user's email address
	Email string `json:"email"`

	// Roles are the user's authorization roles
	Roles []string `json:"roles"`
}

// RefreshTokenClaims represents the claims in a refresh token.
type RefreshTokenClaims struct {
	// Standard JWT claims
	jwt.RegisteredClaims

	// UserID is the unique identifier for the user
	UserID string `json:"user_id"`

	// SessionID is the unique identifier for the session
	SessionID string `json:"session_id"`
}

// TokenPair represents a pair of access and refresh tokens.
type TokenPair struct {
	// AccessToken is the JWT access token
	AccessToken string `json:"access_token"`

	// RefreshToken is the JWT refresh token
	RefreshToken string `json:"refresh_token"`

	// ExpiresIn is the access token expiration time in seconds
	ExpiresIn int64 `json:"expires_in"`

	// TokenType is the type of token (always "Bearer")
	TokenType string `json:"token_type"`
}

// ValidationResult represents the result of token validation.
type ValidationResult struct {
	// Valid indicates whether the token is valid
	Valid bool

	// Claims contains the validated token claims
	Claims interface{}

	// Error contains any validation error
	Error error

	// Expired indicates whether the token has expired
	Expired bool

	// ExpiresAt is when the token expires
	ExpiresAt time.Time
}

// PublicKeyCache represents a cache for public keys.
type PublicKeyCache struct {
	// Key is the public key in PEM format
	Key string

	// CachedAt is when the key was cached
	CachedAt time.Time

	// ExpiresAt is when the cached key expires
	ExpiresAt time.Time
}
