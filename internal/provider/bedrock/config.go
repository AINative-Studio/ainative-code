package bedrock

import (
	"fmt"
	"net/http"
	"os"

	"github.com/AINative-studio/ainative-code/internal/logger"
)

// Config contains configuration for the Bedrock provider
type Config struct {
	// AWS Configuration
	Region       string // AWS region (default: us-east-1)
	AccessKey    string // AWS access key ID
	SecretKey    string // AWS secret access key
	SessionToken string // AWS session token (optional, for temporary credentials)

	// Model Configuration
	Model       string  // Default model to use
	MaxTokens   int     // Default max tokens
	Temperature float64 // Default temperature
	TopP        float64 // Default top_p

	// Optional Configuration
	Endpoint   string                 // Custom endpoint (for testing or VPC endpoints)
	HTTPClient *http.Client           // Custom HTTP client
	Logger     logger.LoggerInterface // Logger instance
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Access key is required
	if c.AccessKey == "" {
		return fmt.Errorf("AccessKey is required")
	}

	// Secret key is required
	if c.SecretKey == "" {
		return fmt.Errorf("SecretKey is required")
	}

	// Region defaults to us-east-1
	if c.Region == "" {
		c.Region = "us-east-1"
	}

	return nil
}

// LoadFromEnvironment loads configuration from environment variables
// Returns a partial config that should be merged with user-provided config
func LoadFromEnvironment() *Config {
	config := &Config{}

	// Load AWS credentials from environment
	if accessKey := os.Getenv("AWS_ACCESS_KEY_ID"); accessKey != "" {
		config.AccessKey = accessKey
	}

	if secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY"); secretKey != "" {
		config.SecretKey = secretKey
	}

	if sessionToken := os.Getenv("AWS_SESSION_TOKEN"); sessionToken != "" {
		config.SessionToken = sessionToken
	}

	if region := os.Getenv("AWS_REGION"); region != "" {
		config.Region = region
	} else if region := os.Getenv("AWS_DEFAULT_REGION"); region != "" {
		config.Region = region
	}

	return config
}

// MergeWithEnvironment merges the config with environment variables
// Environment variables take precedence if not already set in config
func (c *Config) MergeWithEnvironment() {
	envConfig := LoadFromEnvironment()

	// Only use environment values if not already set
	if c.AccessKey == "" && envConfig.AccessKey != "" {
		c.AccessKey = envConfig.AccessKey
	}

	if c.SecretKey == "" && envConfig.SecretKey != "" {
		c.SecretKey = envConfig.SecretKey
	}

	if c.SessionToken == "" && envConfig.SessionToken != "" {
		c.SessionToken = envConfig.SessionToken
	}

	if c.Region == "" && envConfig.Region != "" {
		c.Region = envConfig.Region
	}
}

// DefaultConfig returns a Config with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Region:      "us-east-1",
		MaxTokens:   1024,
		Temperature: 0.7,
		TopP:        1.0,
	}
}
