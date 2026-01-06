package azure

import (
	"fmt"
	"net/http"
	"time"

	"github.com/AINative-studio/ainative-code/internal/logger"
)

const (
	// DefaultAPIVersion is the default Azure OpenAI API version
	DefaultAPIVersion = "2024-02-15-preview"

	// DefaultTimeout is the default HTTP request timeout
	DefaultTimeout = 60 * time.Second
)

// Popular Azure OpenAI deployment names (commonly used)
const (
	// GPT-4 models
	DeploymentGPT4         = "gpt-4"
	DeploymentGPT4_32k     = "gpt-4-32k"
	DeploymentGPT4Turbo    = "gpt-4-turbo"
	DeploymentGPT4O        = "gpt-4o"
	DeploymentGPT4OMini    = "gpt-4o-mini"

	// GPT-3.5 models
	DeploymentGPT35Turbo   = "gpt-35-turbo"
	DeploymentGPT35Turbo16k = "gpt-35-turbo-16k"
)

// Config holds configuration for the Azure OpenAI provider
type Config struct {
	// Endpoint is the Azure OpenAI resource endpoint
	// Format: https://{resource-name}.openai.azure.com
	// Required
	Endpoint string

	// APIKey is the Azure OpenAI API key
	// Required
	APIKey string

	// Deployment is the deployment ID/name for the model
	// This is configured in Azure Portal and maps to a specific model version
	// Required
	Deployment string

	// APIVersion is the Azure OpenAI API version
	// Default: "2024-02-15-preview"
	// Optional - will use DefaultAPIVersion if not specified
	APIVersion string

	// HTTPClient is the HTTP client to use for requests
	// Optional - a default client will be created if not provided
	HTTPClient *http.Client

	// Timeout is the request timeout
	// Default: 60 seconds
	// Optional
	Timeout time.Duration

	// Logger for debugging and error logging
	// Optional
	Logger logger.LoggerInterface

	// MaxRetries is the maximum number of retry attempts for failed requests
	// Default: 3
	// Optional
	MaxRetries int

	// RetryDelay is the initial delay between retries
	// Default: 1 second
	// Optional
	RetryDelay time.Duration
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Endpoint == "" {
		return fmt.Errorf("azure endpoint is required (e.g., https://your-resource.openai.azure.com)")
	}

	if c.APIKey == "" {
		return fmt.Errorf("azure API key is required")
	}

	if c.Deployment == "" {
		return fmt.Errorf("azure deployment ID is required")
	}

	// Validate timeout if provided
	if c.Timeout < 0 {
		return fmt.Errorf("timeout must be non-negative")
	}

	// Validate max retries if provided
	if c.MaxRetries < 0 {
		return fmt.Errorf("max retries must be non-negative")
	}

	return nil
}

// SetDefaults sets default values for unspecified fields
func (c *Config) SetDefaults() {
	if c.APIVersion == "" {
		c.APIVersion = DefaultAPIVersion
	}

	if c.Timeout == 0 {
		c.Timeout = DefaultTimeout
	}

	if c.MaxRetries == 0 {
		c.MaxRetries = 3
	}

	if c.RetryDelay == 0 {
		c.RetryDelay = 1 * time.Second
	}

	if c.HTTPClient == nil {
		c.HTTPClient = &http.Client{
			Timeout: c.Timeout,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				MaxIdleConnsPerHost: 5,
				IdleConnTimeout:     90 * time.Second,
			},
		}
	}
}

// DefaultConfig returns a Config with sensible defaults
// Note: Endpoint, APIKey, and Deployment must still be set
func DefaultConfig() *Config {
	return &Config{
		APIVersion: DefaultAPIVersion,
		Timeout:    DefaultTimeout,
		MaxRetries: 3,
		RetryDelay: 1 * time.Second,
		HTTPClient: &http.Client{
			Timeout: DefaultTimeout,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				MaxIdleConnsPerHost: 5,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

// IsValidDeployment checks if a deployment name is in the list of well-known deployments
// This is a helper function - custom deployment names are also valid
func IsValidDeployment(deployment string) bool {
	knownDeployments := []string{
		DeploymentGPT4,
		DeploymentGPT4_32k,
		DeploymentGPT4Turbo,
		DeploymentGPT4O,
		DeploymentGPT4OMini,
		DeploymentGPT35Turbo,
		DeploymentGPT35Turbo16k,
	}

	for _, known := range knownDeployments {
		if deployment == known {
			return true
		}
	}

	// Custom deployments are valid, so we return true
	// This function just indicates if it's a well-known deployment
	return false
}
