package jwt

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	// DefaultAPITimeout is the default timeout for API validation requests
	DefaultAPITimeout = 10 * time.Second

	// ValidationEndpoint is the API endpoint for token validation
	ValidationEndpoint = "/api/auth/validate"
)

// APIValidator validates tokens using the AINative API.
type APIValidator struct {
	baseURL    string
	httpClient *http.Client
	validator  *Validator
}

// NewAPIValidator creates a new API validator with a local validator fallback.
func NewAPIValidator(baseURL string, httpClient *http.Client) *APIValidator {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: DefaultAPITimeout,
		}
	}

	av := &APIValidator{
		baseURL:    baseURL,
		httpClient: httpClient,
	}

	// Create local validator with key fetcher from API
	av.validator = NewValidator(av.fetchPublicKey)

	return av
}

// ValidateAccessToken validates an access token using API and local validation.
func (av *APIValidator) ValidateAccessToken(ctx context.Context, tokenString string) (*AccessTokenClaims, error) {
	// Try API validation first
	result, err := av.validateTokenAPI(ctx, tokenString)
	if err != nil {
		// If API fails, try local validation as fallback
		return av.validator.ValidateAccessToken(tokenString)
	}

	if !result.Valid {
		return nil, fmt.Errorf("token validation failed: %s", result.Message)
	}

	// Cache public key if provided
	if result.PublicKey != "" {
		if err := av.cachePublicKey(result.PublicKey); err != nil {
			// Log error but don't fail validation
			fmt.Printf("warning: failed to cache public key: %v\n", err)
		}
	}

	// Parse claims from validated token
	return av.validator.ValidateAccessToken(tokenString)
}

// ValidateRefreshToken validates a refresh token using API and local validation.
func (av *APIValidator) ValidateRefreshToken(ctx context.Context, tokenString string) (*RefreshTokenClaims, error) {
	// Try API validation first
	result, err := av.validateTokenAPI(ctx, tokenString)
	if err != nil {
		// If API fails, try local validation as fallback
		return av.validator.ValidateRefreshToken(tokenString)
	}

	if !result.Valid {
		return nil, fmt.Errorf("token validation failed: %s", result.Message)
	}

	// Cache public key if provided
	if result.PublicKey != "" {
		if err := av.cachePublicKey(result.PublicKey); err != nil {
			fmt.Printf("warning: failed to cache public key: %v\n", err)
		}
	}

	// Parse claims from validated token
	return av.validator.ValidateRefreshToken(tokenString)
}

// ValidateToken validates a token using API and returns validation result.
func (av *APIValidator) ValidateToken(ctx context.Context, tokenString string) (*ValidationResult, error) {
	// Try API validation first
	apiResult, err := av.validateTokenAPI(ctx, tokenString)
	if err != nil {
		// If API fails, try local validation as fallback
		return av.validator.ValidateToken(tokenString)
	}

	// Convert API result to ValidationResult
	result := &ValidationResult{
		Valid:   apiResult.Valid,
		Expired: apiResult.Expired,
	}

	if !apiResult.Valid {
		result.Error = fmt.Errorf("%s", apiResult.Message)
	}

	if apiResult.ExpiresAt != nil {
		result.ExpiresAt = *apiResult.ExpiresAt
	}

	// Cache public key if provided
	if apiResult.PublicKey != "" {
		if err := av.cachePublicKey(apiResult.PublicKey); err != nil {
			fmt.Printf("warning: failed to cache public key: %v\n", err)
		}
	}

	// Get claims if validation succeeded
	if result.Valid {
		claims, err := av.validator.ValidateAccessToken(tokenString)
		if err == nil {
			result.Claims = claims
		}
	}

	return result, nil
}

// validateTokenAPI calls the API validation endpoint.
func (av *APIValidator) validateTokenAPI(ctx context.Context, tokenString string) (*APIValidationResponse, error) {
	// Prepare request payload
	reqPayload := APIValidationRequest{
		Token: tokenString,
	}

	reqBody, err := json.Marshal(reqPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := av.baseURL + ValidationEndpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Execute request
	resp, err := av.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Handle non-200 responses
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var result APIValidationResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// fetchPublicKey fetches the public key from the API.
func (av *APIValidator) fetchPublicKey() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultAPITimeout)
	defer cancel()

	url := av.baseURL + "/api/auth/public-key"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := av.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		PublicKey string `json:"public_key"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if result.PublicKey == "" {
		return "", fmt.Errorf("API returned empty public key")
	}

	return result.PublicKey, nil
}

// cachePublicKey caches a public key in the local validator.
func (av *APIValidator) cachePublicKey(pemData string) error {
	// Parse the public key to validate it
	_, err := parsePublicKeyPEM(pemData)
	if err != nil {
		return fmt.Errorf("invalid public key: %w", err)
	}

	// Invalidate current cache and let the validator fetch the new key
	// on next validation (it will use our fetchPublicKey function)
	return av.validator.InvalidateCache()
}

// GetValidator returns the underlying local validator.
func (av *APIValidator) GetValidator() *Validator {
	return av.validator
}

// InvalidateCache invalidates both API and local caches.
func (av *APIValidator) InvalidateCache() error {
	return av.validator.InvalidateCache()
}

// APIValidationRequest represents the request payload for token validation.
type APIValidationRequest struct {
	Token string `json:"token"`
}

// APIValidationResponse represents the response from the validation API.
type APIValidationResponse struct {
	Valid     bool       `json:"valid"`
	Message   string     `json:"message,omitempty"`
	Expired   bool       `json:"expired"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	PublicKey string     `json:"public_key,omitempty"`
}
