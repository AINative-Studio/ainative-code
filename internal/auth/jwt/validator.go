package jwt

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"sync"
	"time"
)

const (
	// PublicKeyCacheTTL is the time-to-live for cached public keys
	PublicKeyCacheTTL = 5 * time.Minute

	// PublicKeyRefreshThreshold is when to refresh the key before expiry
	PublicKeyRefreshThreshold = 1 * time.Minute
)

// Validator handles JWT validation with public key caching.
type Validator struct {
	mu             sync.RWMutex
	publicKey      *rsa.PublicKey
	publicKeyPEM   string
	cachedAt       time.Time
	expiresAt      time.Time
	keyFetcher     KeyFetcher
	invalidateFunc func() error
}

// KeyFetcher is a function that fetches the public key from a remote source.
type KeyFetcher func() (string, error)

// NewValidator creates a new JWT validator with public key caching.
func NewValidator(keyFetcher KeyFetcher) *Validator {
	return &Validator{
		keyFetcher: keyFetcher,
	}
}

// SetInvalidateFunc sets a custom cache invalidation function.
func (v *Validator) SetInvalidateFunc(fn func() error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.invalidateFunc = fn
}

// ValidateAccessToken validates an access token using cached public key.
func (v *Validator) ValidateAccessToken(tokenString string) (*AccessTokenClaims, error) {
	publicKey, err := v.getPublicKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get public key: %w", err)
	}

	claims, err := ValidateAccessToken(tokenString, publicKey)
	if err != nil {
		// Invalidate cache on validation failure
		if invalidateErr := v.InvalidateCache(); invalidateErr != nil {
			// Log but don't fail on invalidation error
			fmt.Printf("warning: failed to invalidate cache: %v\n", invalidateErr)
		}
		return nil, err
	}

	return claims, nil
}

// ValidateRefreshToken validates a refresh token using cached public key.
func (v *Validator) ValidateRefreshToken(tokenString string) (*RefreshTokenClaims, error) {
	publicKey, err := v.getPublicKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get public key: %w", err)
	}

	claims, err := ValidateRefreshToken(tokenString, publicKey)
	if err != nil {
		// Invalidate cache on validation failure
		if invalidateErr := v.InvalidateCache(); invalidateErr != nil {
			fmt.Printf("warning: failed to invalidate cache: %v\n", invalidateErr)
		}
		return nil, err
	}

	return claims, nil
}

// ValidateToken validates a token and returns validation result.
func (v *Validator) ValidateToken(tokenString string) (*ValidationResult, error) {
	publicKey, err := v.getPublicKey()
	if err != nil {
		return &ValidationResult{
			Valid: false,
			Error: fmt.Errorf("failed to get public key: %w", err),
		}, nil
	}

	result, err := ValidateToken(tokenString, publicKey)
	if err != nil {
		return nil, err
	}

	// Invalidate cache if validation failed
	if !result.Valid && result.Error != nil {
		if invalidateErr := v.InvalidateCache(); invalidateErr != nil {
			fmt.Printf("warning: failed to invalidate cache: %v\n", invalidateErr)
		}
	}

	return result, nil
}

// getPublicKey returns the cached public key or fetches a new one if needed.
func (v *Validator) getPublicKey() (*rsa.PublicKey, error) {
	v.mu.RLock()

	// Check if we have a valid cached key
	now := time.Now()
	if v.publicKey != nil && now.Before(v.expiresAt) {
		key := v.publicKey
		v.mu.RUnlock()

		// Check if we should refresh in background (before expiry)
		if now.After(v.expiresAt.Add(-PublicKeyRefreshThreshold)) {
			go v.refreshKeyInBackground()
		}

		return key, nil
	}
	v.mu.RUnlock()

	// Need to fetch new key
	return v.refreshKey()
}

// refreshKey fetches and caches a new public key.
func (v *Validator) refreshKey() (*rsa.PublicKey, error) {
	v.mu.Lock()
	defer v.mu.Unlock()

	// Double-check after acquiring write lock
	now := time.Now()
	if v.publicKey != nil && now.Before(v.expiresAt) {
		return v.publicKey, nil
	}

	// Fetch new key
	if v.keyFetcher == nil {
		return nil, fmt.Errorf("no key fetcher configured")
	}

	pemData, err := v.keyFetcher()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch public key: %w", err)
	}

	// Parse PEM-encoded public key
	publicKey, err := parsePublicKeyPEM(pemData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	// Cache the key
	v.publicKey = publicKey
	v.publicKeyPEM = pemData
	v.cachedAt = now
	v.expiresAt = now.Add(PublicKeyCacheTTL)

	return publicKey, nil
}

// refreshKeyInBackground refreshes the public key in the background.
func (v *Validator) refreshKeyInBackground() {
	_, err := v.refreshKey()
	if err != nil {
		fmt.Printf("warning: background key refresh failed: %v\n", err)
	}
}

// InvalidateCache invalidates the cached public key.
func (v *Validator) InvalidateCache() error {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.publicKey = nil
	v.publicKeyPEM = ""
	v.cachedAt = time.Time{}
	v.expiresAt = time.Time{}

	// Call custom invalidation function if set
	if v.invalidateFunc != nil {
		return v.invalidateFunc()
	}

	return nil
}

// GetCacheInfo returns information about the cached key.
func (v *Validator) GetCacheInfo() *PublicKeyCacheInfo {
	v.mu.RLock()
	defer v.mu.RUnlock()

	return &PublicKeyCacheInfo{
		HasKey:    v.publicKey != nil,
		CachedAt:  v.cachedAt,
		ExpiresAt: v.expiresAt,
		TTL:       PublicKeyCacheTTL,
		IsValid:   v.publicKey != nil && time.Now().Before(v.expiresAt),
	}
}

// PublicKeyCacheInfo represents information about the cached public key.
type PublicKeyCacheInfo struct {
	HasKey    bool
	CachedAt  time.Time
	ExpiresAt time.Time
	TTL       time.Duration
	IsValid   bool
}

// parsePublicKeyPEM parses a PEM-encoded RSA public key.
func parsePublicKeyPEM(pemData string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	if block.Type != "PUBLIC KEY" && block.Type != "RSA PUBLIC KEY" {
		return nil, fmt.Errorf("invalid PEM type: %s", block.Type)
	}

	// Try parsing as PKIX format first
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		// Try parsing as PKCS1 format
		return x509.ParsePKCS1PublicKey(block.Bytes)
	}

	publicKey, ok := pubInterface.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}

	return publicKey, nil
}

// FormatPublicKeyPEM formats an RSA public key as PEM.
func FormatPublicKeyPEM(publicKey *rsa.PublicKey) (string, error) {
	pubBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return "", fmt.Errorf("failed to marshal public key: %w", err)
	}

	pemBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	}

	return string(pem.EncodeToMemory(pemBlock)), nil
}
