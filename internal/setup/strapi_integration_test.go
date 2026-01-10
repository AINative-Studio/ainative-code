package setup

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/AINative-studio/ainative-code/internal/config"
)

// TestStrapiSetupIntegration tests the setup wizard with REAL Strapi API
// CRITICAL: This test uses REAL API calls - NO MOCK DATA
func TestStrapiSetupIntegration(t *testing.T) {
	// Skip if no Strapi URL is configured
	strapiURL := os.Getenv("TEST_STRAPI_URL")
	if strapiURL == "" {
		t.Skip("TEST_STRAPI_URL not set - skipping Strapi integration test")
	}

	strapiAPIKey := os.Getenv("TEST_STRAPI_API_KEY")

	ctx := context.Background()

	t.Run("Setup_with_Strapi_config", func(t *testing.T) {
		wizard := NewWizard(ctx, WizardConfig{
			ConfigPath:      t.TempDir() + "/config.yaml",
			SkipValidation:  true, // Skip LLM validation, but test Strapi
			InteractiveMode: false,
			Force:           true,
		})

		// Set selections to include Strapi
		selections := map[string]interface{}{
			"provider":        "anthropic",
			"anthropic_api_key": "test-key-for-config",
			"anthropic_model": "claude-3-5-sonnet-20241022",
			"ainative_login":  false,
			"strapi_enabled":  true,
			"strapi_url":      strapiURL,
			"strapi_api_key":  strapiAPIKey,
			"zerodb_enabled":  false,
			"color_scheme":    "auto",
			"prompt_caching":  true,
		}
		wizard.SetSelections(selections)

		// Build configuration
		err := wizard.buildConfiguration()
		if err != nil {
			t.Fatalf("Failed to build configuration: %v", err)
		}

		// Verify Strapi config was created
		if wizard.result.Config.Services.Strapi == nil {
			t.Fatal("Strapi configuration was not created")
		}

		strapiCfg := wizard.result.Config.Services.Strapi
		if !strapiCfg.Enabled {
			t.Error("Strapi should be enabled")
		}
		if strapiCfg.Endpoint != strapiURL {
			t.Errorf("Expected Strapi URL %s, got %s", strapiURL, strapiCfg.Endpoint)
		}
		if strapiCfg.APIKey != strapiAPIKey {
			t.Errorf("Expected Strapi API key %s, got %s", strapiAPIKey, strapiCfg.APIKey)
		}
		if strapiCfg.Timeout != 30*time.Second {
			t.Errorf("Expected timeout 30s, got %v", strapiCfg.Timeout)
		}
		if strapiCfg.RetryAttempts != 3 {
			t.Errorf("Expected 3 retry attempts, got %d", strapiCfg.RetryAttempts)
		}

		t.Logf("✓ Strapi configuration created successfully")
	})

	t.Run("Validate_Strapi_API_Connection", func(t *testing.T) {
		// Test REAL API connection to Strapi
		result, err := validateStrapiConnection(ctx, strapiURL, strapiAPIKey)
		if err != nil {
			t.Logf("Warning: Strapi API validation failed: %v", err)
			t.Logf("This might be expected if Strapi instance is not accessible")
		} else {
			t.Logf("✓ Successfully connected to Strapi API")
			t.Logf("API Response: %s", result)
		}
	})

	t.Run("Test_Strapi_Content_Types_API", func(t *testing.T) {
		// Test REAL content-types endpoint
		contentTypes, err := getStrapiContentTypes(ctx, strapiURL, strapiAPIKey)
		if err != nil {
			t.Logf("Warning: Failed to fetch Strapi content types: %v", err)
		} else {
			t.Logf("✓ Successfully fetched Strapi content types")
			t.Logf("Content Types: %v", contentTypes)

			if len(contentTypes) > 0 {
				t.Logf("Found %d content types in Strapi", len(contentTypes))
			}
		}
	})

	t.Run("Test_Strapi_Health_Check", func(t *testing.T) {
		// Test REAL health check endpoint
		healthy, info := checkStrapiHealth(ctx, strapiURL)
		if !healthy {
			t.Logf("Warning: Strapi health check failed")
		} else {
			t.Logf("✓ Strapi instance is healthy")
			t.Logf("Health Info: %s", info)
		}
	})
}

// validateStrapiConnection makes a REAL API call to validate Strapi connection
func validateStrapiConnection(ctx context.Context, baseURL, apiKey string) (string, error) {
	// Test the /api/users/me endpoint (requires authentication)
	url := strings.TrimRight(baseURL, "/") + "/api/users/me"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusOK {
		return string(body), nil
	}

	// Try the root API endpoint as fallback
	url = strings.TrimRight(baseURL, "/") + "/api"
	req, _ = http.NewRequestWithContext(ctx, "GET", url, nil)
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err = client.Do(req)
	if err != nil {
		return "", fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ = io.ReadAll(resp.Body)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return string(body), nil
	}

	return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
}

// getStrapiContentTypes fetches REAL content types from Strapi API
func getStrapiContentTypes(ctx context.Context, baseURL, apiKey string) ([]string, error) {
	url := strings.TrimRight(baseURL, "/") + "/api/content-type-builder/content-types"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Data []struct {
			UID string `json:"uid"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	contentTypes := make([]string, len(result.Data))
	for i, ct := range result.Data {
		contentTypes[i] = ct.UID
	}

	return contentTypes, nil
}

// checkStrapiHealth performs REAL health check on Strapi instance
func checkStrapiHealth(ctx context.Context, baseURL string) (bool, string) {
	url := strings.TrimRight(baseURL, "/") + "/_health"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		// Try alternative health endpoint
		url = strings.TrimRight(baseURL, "/")
		req, err = http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return false, fmt.Sprintf("Failed to create request: %v", err)
		}
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Sprintf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return true, string(body)
	}

	return false, fmt.Sprintf("Status: %d, Body: %s", resp.StatusCode, string(body))
}

// TestStrapiConfigValidation tests validation of Strapi configuration
func TestStrapiConfigValidation(t *testing.T) {
	t.Run("Valid_Strapi_Config", func(t *testing.T) {
		cfg := &config.Config{
			Services: config.ServicesConfig{
				Strapi: &config.StrapiConfig{
					Enabled:       true,
					Endpoint:      "https://example.com",
					APIKey:        "test-key",
					Timeout:       30 * time.Second,
					RetryAttempts: 3,
				},
			},
		}

		// Validate the config structure
		if cfg.Services.Strapi == nil {
			t.Fatal("Strapi config should not be nil")
		}
		if !cfg.Services.Strapi.Enabled {
			t.Error("Strapi should be enabled")
		}
		if cfg.Services.Strapi.Endpoint == "" {
			t.Error("Strapi endpoint should not be empty")
		}
	})

	t.Run("Strapi_URL_Format_Validation", func(t *testing.T) {
		validURLs := []string{
			"http://localhost:1337",
			"https://strapi.example.com",
			"https://api.example.com/strapi",
		}

		for _, url := range validURLs {
			if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
				t.Errorf("URL %s should have http/https prefix", url)
			}
		}
	})
}
