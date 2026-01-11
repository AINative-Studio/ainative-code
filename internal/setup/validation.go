package setup

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Validator handles validation of API keys and connections
type Validator struct {
	httpClient *http.Client
}

// NewValidator creates a new validator instance
func NewValidator() *Validator {
	return &Validator{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// ValidateAnthropicKey validates an Anthropic API key
func (v *Validator) ValidateAnthropicKey(ctx context.Context, apiKey string) error {
	if apiKey == "" {
		return fmt.Errorf("API key cannot be empty")
	}

	if !strings.HasPrefix(apiKey, "sk-ant-") {
		return fmt.Errorf("invalid API key format: must start with 'sk-ant-'")
	}

	// Basic format validation
	// Anthropic API keys are typically 100+ characters
	if len(apiKey) < 20 {
		return fmt.Errorf("API key appears to be too short")
	}

	// Note: We don't actually test the API key here to avoid rate limits and network dependencies
	// The key will be validated when actually used
	return nil
}

// ValidateOpenAIKey validates an OpenAI API key
func (v *Validator) ValidateOpenAIKey(ctx context.Context, apiKey string) error {
	if apiKey == "" {
		return fmt.Errorf("API key cannot be empty")
	}

	if !strings.HasPrefix(apiKey, "sk-") {
		return fmt.Errorf("invalid API key format: must start with 'sk-'")
	}

	// Test the API key by making a request to the models endpoint
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.openai.com/v1/models", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := v.httpClient.Do(req)
	if err != nil {
		// Network error is not a validation failure
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("invalid API key: authentication failed")
	}

	return nil
}

// ValidateGoogleKey validates a Google API key
func (v *Validator) ValidateGoogleKey(ctx context.Context, apiKey string) error {
	if apiKey == "" {
		return fmt.Errorf("API key cannot be empty")
	}

	// Test the API key by making a request to the Gemini API
	testURL := fmt.Sprintf("https://generativelanguage.googleapis.com/v1/models?key=%s", apiKey)

	req, err := http.NewRequestWithContext(ctx, "GET", testURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := v.httpClient.Do(req)
	if err != nil {
		// Network error is not a validation failure
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("invalid API key: authentication failed")
	}

	if resp.StatusCode == http.StatusBadRequest {
		return fmt.Errorf("invalid API key format")
	}

	return nil
}

// ValidateOllamaConnection validates connection to an Ollama server
func (v *Validator) ValidateOllamaConnection(ctx context.Context, baseURL string) error {
	if baseURL == "" {
		return fmt.Errorf("base URL cannot be empty")
	}

	// Validate URL format
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("URL must use http or https scheme")
	}

	// Test connection to Ollama server
	testURL := fmt.Sprintf("%s/api/tags", baseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", testURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to Ollama server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Ollama server returned status %d", resp.StatusCode)
	}

	return nil
}

// ValidateOllamaModel validates that a model exists in Ollama
func (v *Validator) ValidateOllamaModel(ctx context.Context, baseURL, modelName string) error {
	if modelName == "" {
		return fmt.Errorf("model name cannot be empty")
	}

	// This is a soft validation - we just check the format
	// Actual model existence will be checked at runtime
	if strings.Contains(modelName, " ") {
		return fmt.Errorf("model name cannot contain spaces")
	}

	return nil
}

// ValidateMetaLlamaKey validates a Meta Llama API key
func (v *Validator) ValidateMetaLlamaKey(ctx context.Context, apiKey string) error {
	if apiKey == "" {
		return fmt.Errorf("API key cannot be empty")
	}

	// Basic format validation
	if len(apiKey) < 20 {
		return fmt.Errorf("API key appears to be too short")
	}

	// Note: We don't actually test the API key here to avoid rate limits and network dependencies
	// The key will be validated when actually used
	return nil
}

// ValidateAINativeKey validates an AINative platform API key
func (v *Validator) ValidateAINativeKey(ctx context.Context, apiKey string) error {
	if apiKey == "" {
		return fmt.Errorf("API key cannot be empty")
	}

	// Basic format validation
	if len(apiKey) < 20 {
		return fmt.Errorf("API key appears to be too short")
	}

	// In a real implementation, we would test the key against the AINative API
	// For now, we'll just do basic format validation

	return nil
}

// ValidateProviderConfig validates the complete provider configuration
func (v *Validator) ValidateProviderConfig(ctx context.Context, provider string, selections map[string]interface{}) error {
	switch provider {
	case "anthropic":
		apiKey, ok := selections["anthropic_api_key"].(string)
		if !ok || apiKey == "" {
			return fmt.Errorf("Anthropic API key is required")
		}
		return v.ValidateAnthropicKey(ctx, apiKey)

	case "openai":
		apiKey, ok := selections["openai_api_key"].(string)
		if !ok || apiKey == "" {
			return fmt.Errorf("OpenAI API key is required")
		}
		return v.ValidateOpenAIKey(ctx, apiKey)

	case "google":
		apiKey, ok := selections["google_api_key"].(string)
		if !ok || apiKey == "" {
			return fmt.Errorf("Google API key is required")
		}
		return v.ValidateGoogleKey(ctx, apiKey)

	case "ollama":
		baseURL, ok := selections["ollama_url"].(string)
		if !ok || baseURL == "" {
			baseURL = "http://localhost:11434"
		}
		if err := v.ValidateOllamaConnection(ctx, baseURL); err != nil {
			return err
		}

		modelName, ok := selections["ollama_model"].(string)
		if !ok || modelName == "" {
			return fmt.Errorf("Ollama model name is required")
		}
		return v.ValidateOllamaModel(ctx, baseURL, modelName)

	case "meta_llama", "meta":
		apiKey, ok := selections["meta_llama_api_key"].(string)
		if !ok || apiKey == "" {
			return fmt.Errorf("Meta Llama API key is required")
		}
		return v.ValidateMetaLlamaKey(ctx, apiKey)

	default:
		return fmt.Errorf("unsupported provider: %s", provider)
	}
}

// ValidateAll performs comprehensive validation of all user inputs
func (v *Validator) ValidateAll(ctx context.Context, selections map[string]interface{}) error {
	// Validate provider selection
	provider, ok := selections["provider"].(string)
	if !ok || provider == "" {
		return fmt.Errorf("provider selection is required")
	}

	// Validate provider-specific configuration
	if err := v.ValidateProviderConfig(ctx, provider, selections); err != nil {
		return fmt.Errorf("provider configuration validation failed: %w", err)
	}

	// Validate AINative configuration if enabled
	if loginEnabled, ok := selections["ainative_login"].(bool); ok && loginEnabled {
		apiKey, ok := selections["ainative_api_key"].(string)
		if !ok || apiKey == "" {
			return fmt.Errorf("AINative API key is required when platform login is enabled")
		}
		if err := v.ValidateAINativeKey(ctx, apiKey); err != nil {
			return fmt.Errorf("AINative API key validation failed: %w", err)
		}
	}

	// Validate Strapi configuration if enabled
	if strapiEnabled, ok := selections["strapi_enabled"].(bool); ok && strapiEnabled {
		strapiURL, ok := selections["strapi_url"].(string)
		if !ok || strapiURL == "" {
			return fmt.Errorf("Strapi URL is required when Strapi is enabled")
		}
		if err := v.ValidateStrapiURL(ctx, strapiURL); err != nil {
			return fmt.Errorf("Strapi URL validation failed: %w", err)
		}

		// API key is optional, but validate if provided
		if strapiAPIKey, ok := selections["strapi_api_key"].(string); ok && strapiAPIKey != "" {
			if err := v.ValidateStrapiConnection(ctx, strapiURL, strapiAPIKey); err != nil {
				return fmt.Errorf("Strapi connection validation failed: %w", err)
			}
		}
	}

	// Validate ZeroDB configuration if enabled
	if zeroDBEnabled, ok := selections["zerodb_enabled"].(bool); ok && zeroDBEnabled {
		projectID, ok := selections["zerodb_project_id"].(string)
		if !ok || projectID == "" {
			return fmt.Errorf("ZeroDB Project ID is required when ZeroDB is enabled")
		}
		if err := v.ValidateZeroDBProjectID(projectID); err != nil {
			return fmt.Errorf("ZeroDB Project ID validation failed: %w", err)
		}

		// Endpoint is optional
		if endpoint, ok := selections["zerodb_endpoint"].(string); ok && endpoint != "" {
			if err := v.ValidateZeroDBEndpoint(endpoint); err != nil {
				return fmt.Errorf("ZeroDB endpoint validation failed: %w", err)
			}
		}
	}

	return nil
}

// ValidateStrapiURL validates a Strapi instance URL
func (v *Validator) ValidateStrapiURL(ctx context.Context, strapiURL string) error {
	if strapiURL == "" {
		return fmt.Errorf("Strapi URL cannot be empty")
	}

	// Validate URL format
	parsedURL, err := url.Parse(strapiURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("URL must use http or https scheme")
	}

	return nil
}

// ValidateStrapiConnection makes a REAL API call to validate Strapi connection
func (v *Validator) ValidateStrapiConnection(ctx context.Context, baseURL, apiKey string) error {
	// Test the /api endpoint
	testURL := strings.TrimRight(baseURL, "/") + "/api"

	req, err := http.NewRequestWithContext(ctx, "GET", testURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to Strapi: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("Strapi authentication failed: invalid API key")
	}

	if resp.StatusCode >= 500 {
		return fmt.Errorf("Strapi server error: status %d", resp.StatusCode)
	}

	// Status 200-399 are considered successful
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return nil
	}

	return fmt.Errorf("unexpected Strapi response: status %d", resp.StatusCode)
}

// ValidateZeroDBProjectID validates a ZeroDB project ID format
func (v *Validator) ValidateZeroDBProjectID(projectID string) error {
	if projectID == "" {
		return fmt.Errorf("ZeroDB Project ID cannot be empty")
	}

	// Project IDs should be alphanumeric with hyphens/underscores
	if len(projectID) < 3 {
		return fmt.Errorf("Project ID is too short")
	}

	// Basic format validation
	for _, c := range projectID {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
			return fmt.Errorf("Project ID contains invalid characters")
		}
	}

	return nil
}

// ValidateZeroDBEndpoint validates a ZeroDB endpoint URL
func (v *Validator) ValidateZeroDBEndpoint(endpoint string) error {
	if endpoint == "" {
		// Empty endpoint is allowed (will use default)
		return nil
	}

	// Validate URL format
	parsedURL, err := url.Parse(endpoint)
	if err != nil {
		return fmt.Errorf("invalid endpoint URL format: %w", err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("endpoint URL must use http or https scheme")
	}

	return nil
}

// SanitizeAPIKey sanitizes an API key for display
func SanitizeAPIKey(apiKey string) string {
	if len(apiKey) <= 8 {
		return "***"
	}
	return apiKey[:4] + "..." + apiKey[len(apiKey)-4:]
}
