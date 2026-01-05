package ollama

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/AINative-studio/ainative-code/internal/logger"
)

const (
	// DefaultOllamaURL is the default Ollama API endpoint
	DefaultOllamaURL = "http://localhost:11434"

	// DefaultNumCtx is the default context window size
	DefaultNumCtx = 2048

	// DefaultTemperature is the default sampling temperature
	DefaultTemperature = 0.8

	// DefaultTopK is the default top-k sampling value
	DefaultTopK = 40

	// DefaultTopP is the default top-p (nucleus) sampling value
	DefaultTopP = 0.9

	// DefaultTimeout is the default request timeout
	DefaultTimeout = 120 * time.Second
)

// Config contains configuration for the Ollama provider
type Config struct {
	// BaseURL is the Ollama API endpoint (default: http://localhost:11434)
	BaseURL string

	// Model is the name of the model to use (e.g., "llama2", "llama3", "codellama")
	Model string

	// NumCtx is the context window size (default: 2048)
	NumCtx int

	// Temperature controls randomness in generation (0.0 to 2.0, default: 0.8)
	Temperature float64

	// TopK limits the next token selection to the K most likely tokens (default: 40)
	TopK int

	// TopP controls nucleus sampling (0.0 to 1.0, default: 0.9)
	TopP float64

	// Timeout is the request timeout duration (default: 120s)
	Timeout time.Duration

	// HTTPClient is the HTTP client to use (optional)
	HTTPClient *http.Client

	// Logger is the logger instance (optional)
	Logger logger.LoggerInterface
}

// DefaultConfig returns a Config with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		BaseURL:     DefaultOllamaURL,
		NumCtx:      DefaultNumCtx,
		Temperature: DefaultTemperature,
		TopK:        DefaultTopK,
		TopP:        DefaultTopP,
		Timeout:     DefaultTimeout,
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

// SetDefaults sets default values for unspecified fields
func (c *Config) SetDefaults() {
	if c.BaseURL == "" {
		c.BaseURL = DefaultOllamaURL
	}
	if c.NumCtx == 0 {
		c.NumCtx = DefaultNumCtx
	}
	if c.Temperature == 0 {
		c.Temperature = DefaultTemperature
	}
	if c.TopK == 0 {
		c.TopK = DefaultTopK
	}
	if c.TopP == 0 {
		c.TopP = DefaultTopP
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

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Model == "" {
		return fmt.Errorf("model name is required")
	}

	if c.NumCtx < 0 {
		return fmt.Errorf("NumCtx must be positive")
	}

	if c.Temperature < 0 || c.Temperature > 2.0 {
		return fmt.Errorf("temperature must be between 0 and 2")
	}

	if c.TopP < 0 || c.TopP > 1.0 {
		return fmt.Errorf("TopP must be between 0 and 1")
	}

	if c.TopK < 0 {
		return fmt.Errorf("TopK must be positive")
	}

	return nil
}

// HealthCheck verifies connectivity to the Ollama server
func (c *Config) HealthCheck(ctx context.Context) error {
	baseURL := c.BaseURL
	if baseURL == "" {
		baseURL = DefaultOllamaURL
	}

	client := c.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}

	// Check the /api/tags endpoint which lists available models
	url := fmt.Sprintf("%s/api/tags", baseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("ollama server not reachable at %s: %w", baseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ollama server returned unexpected status: %d", resp.StatusCode)
	}

	return nil
}
