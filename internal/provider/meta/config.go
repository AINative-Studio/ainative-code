package meta

import (
	"fmt"
	"net/http"
	"time"
)

// Config holds configuration for the Meta LLAMA provider
type Config struct {
	// APIKey is the Meta LLAMA API key (required)
	// Format: LLM|<app_id>|<token>
	APIKey string

	// BaseURL is the Meta LLAMA API endpoint
	// Default: https://api.llama.com/compat/v1
	BaseURL string

	// Model is the Meta LLAMA model to use
	// Options: Llama-4-Maverick-17B-128E-Instruct-FP8, Llama-4-Scout-17B-16E,
	//          Llama-3.3-70B-Instruct, Llama-3.3-8B-Instruct
	Model string

	// Temperature controls randomness (0.0 to 2.0, default: 0.7)
	Temperature float64

	// TopP controls nucleus sampling (0.0 to 1.0, default: 0.9)
	TopP float64

	// MaxTokens is the maximum number of tokens to generate
	MaxTokens int

	// Timeout is the request timeout
	Timeout time.Duration

	// HTTPClient is the HTTP client to use (optional)
	HTTPClient *http.Client

	// PresencePenalty reduces repetition (-2.0 to 2.0, default: 0.0)
	PresencePenalty float64

	// FrequencyPenalty reduces repetition of token sequences (-2.0 to 2.0, default: 0.0)
	FrequencyPenalty float64

	// Stop sequences where the API will stop generating
	Stop []string
}

// DefaultConfig returns a Config with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		BaseURL:          DefaultBaseURL,
		Model:            ModelLlama4Maverick,
		Temperature:      0.7,
		TopP:             0.9,
		MaxTokens:        2048,
		Timeout:          DefaultTimeout,
		PresencePenalty:  0.0,
		FrequencyPenalty: 0.0,
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

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.APIKey == "" {
		return fmt.Errorf("API key is required")
	}

	if c.BaseURL == "" {
		return fmt.Errorf("base URL is required")
	}

	if c.Model == "" {
		return fmt.Errorf("model is required")
	}

	if !IsValidModel(c.Model) {
		return fmt.Errorf("invalid model: %s (supported: %s, %s, %s, %s)",
			c.Model, ModelLlama4Maverick, ModelLlama4Scout, ModelLlama33_70B, ModelLlama33_8B)
	}

	if c.Temperature < 0 || c.Temperature > 2.0 {
		return fmt.Errorf("temperature must be between 0 and 2")
	}

	if c.TopP < 0 || c.TopP > 1.0 {
		return fmt.Errorf("top_p must be between 0 and 1")
	}

	if c.MaxTokens < 0 {
		return fmt.Errorf("max_tokens must be non-negative")
	}

	if c.PresencePenalty < -2.0 || c.PresencePenalty > 2.0 {
		return fmt.Errorf("presence_penalty must be between -2 and 2")
	}

	if c.FrequencyPenalty < -2.0 || c.FrequencyPenalty > 2.0 {
		return fmt.Errorf("frequency_penalty must be between -2 and 2")
	}

	return nil
}

// SetDefaults sets default values for unspecified fields
func (c *Config) SetDefaults() {
	if c.BaseURL == "" {
		c.BaseURL = DefaultBaseURL
	}
	if c.Model == "" {
		c.Model = ModelLlama4Maverick
	}
	if c.Temperature == 0 {
		c.Temperature = 0.7
	}
	if c.TopP == 0 {
		c.TopP = 0.9
	}
	if c.MaxTokens == 0 {
		c.MaxTokens = 2048
	}
	if c.Timeout == 0 {
		c.Timeout = DefaultTimeout
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
