package keychain

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/99designs/keyring"
	"github.com/AINative-studio/ainative-code/internal/auth/jwt"
)

const (
	// ServiceName is the identifier for this application in the keychain
	ServiceName = "ainative-code"

	// AccessTokenKey is the key for storing access tokens
	AccessTokenKey = "access_token"

	// RefreshTokenKey is the key for storing refresh tokens
	RefreshTokenKey = "refresh_token"

	// TokenPairKey is the key for storing the complete token pair
	TokenPairKey = "token_pair"

	// APIKeyKey is the key for storing API keys
	APIKeyKey = "api_key"

	// UserEmailKey is the key for storing user email
	UserEmailKey = "user_email"
)

// Keychain provides secure credential storage using OS-level services.
type Keychain interface {
	// SetAccessToken stores an access token
	SetAccessToken(token string) error

	// GetAccessToken retrieves the access token
	GetAccessToken() (string, error)

	// SetRefreshToken stores a refresh token
	SetRefreshToken(token string) error

	// GetRefreshToken retrieves the refresh token
	GetRefreshToken() (string, error)

	// SetTokenPair stores both access and refresh tokens
	SetTokenPair(tokens *jwt.TokenPair) error

	// GetTokenPair retrieves both access and refresh tokens
	GetTokenPair() (*jwt.TokenPair, error)

	// SetAPIKey stores an API key
	SetAPIKey(key string) error

	// GetAPIKey retrieves the API key
	GetAPIKey() (string, error)

	// SetUserEmail stores the user's email
	SetUserEmail(email string) error

	// GetUserEmail retrieves the user's email
	GetUserEmail() (string, error)

	// Delete removes a specific key
	Delete(key string) error

	// DeleteAll removes all stored credentials
	DeleteAll() error

	// Exists checks if a key exists
	Exists(key string) bool
}

// keychainImpl is the default implementation using 99designs/keyring.
type keychainImpl struct {
	ring keyring.Keyring
}

var (
	// defaultKeychain is the singleton instance
	defaultKeychain Keychain
)

// Get returns the platform-specific keychain instance.
func Get() Keychain {
	if defaultKeychain == nil {
		defaultKeychain = newKeychain()
	}
	return defaultKeychain
}

// New creates a new keychain instance with custom configuration.
func New(config keyring.Config) (Keychain, error) {
	ring, err := keyring.Open(config)
	if err != nil {
		return nil, fmt.Errorf("failed to open keyring: %w", err)
	}

	return &keychainImpl{
		ring: ring,
	}, nil
}

// newKeychain creates a new keychain with default configuration.
func newKeychain() Keychain {
	config := keyring.Config{
		ServiceName: ServiceName,

		// Allowed backends in order of preference
		AllowedBackends: []keyring.BackendType{
			keyring.KeychainBackend,        // macOS Keychain
			keyring.SecretServiceBackend,   // Linux Secret Service
			keyring.WinCredBackend,         // Windows Credential Manager
			keyring.FileBackend,            // Fallback encrypted file
		},
	}

	ring, err := keyring.Open(config)
	if err != nil {
		// If all backends fail, return a no-op implementation
		return &noopKeychain{}
	}

	return &keychainImpl{
		ring: ring,
	}
}

// SetAccessToken stores an access token.
func (k *keychainImpl) SetAccessToken(token string) error {
	item := keyring.Item{
		Key:  AccessTokenKey,
		Data: []byte(token),
	}

	if err := k.ring.Set(item); err != nil {
		return fmt.Errorf("failed to store access token: %w", err)
	}

	return nil
}

// GetAccessToken retrieves the access token.
func (k *keychainImpl) GetAccessToken() (string, error) {
	item, err := k.ring.Get(AccessTokenKey)
	if err != nil {
		if err == keyring.ErrKeyNotFound {
			return "", fmt.Errorf("access token not found")
		}
		return "", fmt.Errorf("failed to retrieve access token: %w", err)
	}

	return string(item.Data), nil
}

// SetRefreshToken stores a refresh token.
func (k *keychainImpl) SetRefreshToken(token string) error {
	item := keyring.Item{
		Key:  RefreshTokenKey,
		Data: []byte(token),
	}

	if err := k.ring.Set(item); err != nil {
		return fmt.Errorf("failed to store refresh token: %w", err)
	}

	return nil
}

// GetRefreshToken retrieves the refresh token.
func (k *keychainImpl) GetRefreshToken() (string, error) {
	item, err := k.ring.Get(RefreshTokenKey)
	if err != nil {
		if err == keyring.ErrKeyNotFound {
			return "", fmt.Errorf("refresh token not found")
		}
		return "", fmt.Errorf("failed to retrieve refresh token: %w", err)
	}

	return string(item.Data), nil
}

// SetTokenPair stores both access and refresh tokens as a JSON object.
func (k *keychainImpl) SetTokenPair(tokens *jwt.TokenPair) error {
	if tokens == nil {
		return fmt.Errorf("tokens cannot be nil")
	}

	data, err := json.Marshal(tokens)
	if err != nil {
		return fmt.Errorf("failed to marshal tokens: %w", err)
	}

	item := keyring.Item{
		Key:  TokenPairKey,
		Data: data,
	}

	if err := k.ring.Set(item); err != nil {
		return fmt.Errorf("failed to store token pair: %w", err)
	}

	// Also store individually for convenience
	if err := k.SetAccessToken(tokens.AccessToken); err != nil {
		return err
	}

	if err := k.SetRefreshToken(tokens.RefreshToken); err != nil {
		return err
	}

	return nil
}

// GetTokenPair retrieves both access and refresh tokens.
func (k *keychainImpl) GetTokenPair() (*jwt.TokenPair, error) {
	item, err := k.ring.Get(TokenPairKey)
	if err != nil {
		if err == keyring.ErrKeyNotFound {
			return nil, fmt.Errorf("token pair not found")
		}
		return nil, fmt.Errorf("failed to retrieve token pair: %w", err)
	}

	var tokens jwt.TokenPair
	if err := json.Unmarshal(item.Data, &tokens); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tokens: %w", err)
	}

	return &tokens, nil
}

// SetAPIKey stores an API key.
func (k *keychainImpl) SetAPIKey(key string) error {
	item := keyring.Item{
		Key:  APIKeyKey,
		Data: []byte(key),
	}

	if err := k.ring.Set(item); err != nil {
		return fmt.Errorf("failed to store API key: %w", err)
	}

	return nil
}

// GetAPIKey retrieves the API key.
func (k *keychainImpl) GetAPIKey() (string, error) {
	item, err := k.ring.Get(APIKeyKey)
	if err != nil {
		if err == keyring.ErrKeyNotFound {
			return "", fmt.Errorf("API key not found")
		}
		return "", fmt.Errorf("failed to retrieve API key: %w", err)
	}

	return string(item.Data), nil
}

// SetUserEmail stores the user's email.
func (k *keychainImpl) SetUserEmail(email string) error {
	item := keyring.Item{
		Key:  UserEmailKey,
		Data: []byte(email),
	}

	if err := k.ring.Set(item); err != nil {
		return fmt.Errorf("failed to store user email: %w", err)
	}

	return nil
}

// GetUserEmail retrieves the user's email.
func (k *keychainImpl) GetUserEmail() (string, error) {
	item, err := k.ring.Get(UserEmailKey)
	if err != nil {
		if err == keyring.ErrKeyNotFound {
			return "", fmt.Errorf("user email not found")
		}
		return "", fmt.Errorf("failed to retrieve user email: %w", err)
	}

	return string(item.Data), nil
}

// Delete removes a specific key from the keychain.
func (k *keychainImpl) Delete(key string) error {
	if err := k.ring.Remove(key); err != nil {
		// Ignore if key doesn't exist (already deleted)
		if err == keyring.ErrKeyNotFound {
			return nil
		}
		// FileBackend may return os.ErrNotExist for missing keys
		if err.Error() == "remove" || strings.Contains(err.Error(), "no such file") {
			return nil
		}
		return fmt.Errorf("failed to delete key %s: %w", key, err)
	}

	return nil
}

// DeleteAll removes all stored credentials.
func (k *keychainImpl) DeleteAll() error {
	keys := []string{
		AccessTokenKey,
		RefreshTokenKey,
		TokenPairKey,
		APIKeyKey,
		UserEmailKey,
	}

	var errs []error
	for _, key := range keys {
		if err := k.Delete(key); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to delete all credentials: %v", errs)
	}

	return nil
}

// Exists checks if a key exists in the keychain.
func (k *keychainImpl) Exists(key string) bool {
	_, err := k.ring.Get(key)
	return err == nil
}

// noopKeychain is a no-op implementation when no backends are available.
type noopKeychain struct{}

func (n *noopKeychain) SetAccessToken(token string) error       { return fmt.Errorf("no keychain available") }
func (n *noopKeychain) GetAccessToken() (string, error)         { return "", fmt.Errorf("no keychain available") }
func (n *noopKeychain) SetRefreshToken(token string) error      { return fmt.Errorf("no keychain available") }
func (n *noopKeychain) GetRefreshToken() (string, error)        { return "", fmt.Errorf("no keychain available") }
func (n *noopKeychain) SetTokenPair(tokens *jwt.TokenPair) error { return fmt.Errorf("no keychain available") }
func (n *noopKeychain) GetTokenPair() (*jwt.TokenPair, error)   { return nil, fmt.Errorf("no keychain available") }
func (n *noopKeychain) SetAPIKey(key string) error              { return fmt.Errorf("no keychain available") }
func (n *noopKeychain) GetAPIKey() (string, error)              { return "", fmt.Errorf("no keychain available") }
func (n *noopKeychain) SetUserEmail(email string) error         { return fmt.Errorf("no keychain available") }
func (n *noopKeychain) GetUserEmail() (string, error)           { return "", fmt.Errorf("no keychain available") }
func (n *noopKeychain) Delete(key string) error                 { return nil }
func (n *noopKeychain) DeleteAll() error                        { return nil }
func (n *noopKeychain) Exists(key string) bool                  { return false }
