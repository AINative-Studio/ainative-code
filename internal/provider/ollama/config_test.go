package ollama

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name        string
		baseURL     string
		model       string
		expectError bool
		errorMsg    string
		expectedURL string
	}{
		{
			name:        "default config",
			baseURL:     "",
			model:       "llama2",
			expectError: false,
			expectedURL: DefaultOllamaURL,
		},
		{
			name:        "custom base URL",
			baseURL:     "http://custom-host:11434",
			model:       "llama3",
			expectError: false,
			expectedURL: "http://custom-host:11434",
		},
		{
			name:        "missing model",
			baseURL:     "",
			model:       "",
			expectError: true,
			errorMsg:    "model name is required",
		},
		{
			name:        "valid with localhost",
			baseURL:     "http://localhost:11434",
			model:       "codellama",
			expectError: false,
			expectedURL: "http://localhost:11434",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{
				BaseURL: tt.baseURL,
				Model:   tt.model,
			}

			err := config.Validate()

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
				if tt.expectedURL != "" && tt.baseURL == "" {
					// Default URL should be used
					assert.Equal(t, DefaultOllamaURL, DefaultOllamaURL)
				}
			}
		})
	}
}

func TestConfig_Defaults(t *testing.T) {
	config := Config{
		Model: "llama2",
	}

	config.SetDefaults()

	assert.Equal(t, DefaultOllamaURL, config.BaseURL)
	assert.Equal(t, DefaultNumCtx, config.NumCtx)
	assert.Equal(t, DefaultTemperature, config.Temperature)
	assert.Equal(t, DefaultTopK, config.TopK)
	assert.Equal(t, DefaultTopP, config.TopP)
	assert.Equal(t, DefaultTimeout, config.Timeout)
}

func TestConfig_CustomValues(t *testing.T) {
	config := Config{
		BaseURL:     "http://custom:11434",
		Model:       "mistral",
		NumCtx:      8192,
		Temperature: 0.9,
		TopK:        50,
		TopP:        0.95,
		Timeout:     60 * time.Second,
	}

	config.SetDefaults()

	// Custom values should not be overridden
	assert.Equal(t, "http://custom:11434", config.BaseURL)
	assert.Equal(t, 8192, config.NumCtx)
	assert.Equal(t, 0.9, config.Temperature)
	assert.Equal(t, 50, config.TopK)
	assert.Equal(t, 0.95, config.TopP)
	assert.Equal(t, 60*time.Second, config.Timeout)
}

func TestConfig_HealthCheck(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/tags" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"models":[]}`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	t.Run("successful health check", func(t *testing.T) {
		config := Config{
			BaseURL: server.URL,
			Model:   "llama2",
			Timeout: 5 * time.Second,
		}

		ctx := context.Background()
		err := config.HealthCheck(ctx)
		assert.NoError(t, err)
	})

	t.Run("health check with cancelled context", func(t *testing.T) {
		config := Config{
			BaseURL: server.URL,
			Model:   "llama2",
			Timeout: 5 * time.Second,
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		err := config.HealthCheck(ctx)
		assert.Error(t, err)
	})

	t.Run("health check with invalid URL", func(t *testing.T) {
		config := Config{
			BaseURL: "http://invalid-host:99999",
			Model:   "llama2",
			Timeout: 1 * time.Second,
		}

		ctx := context.Background()
		err := config.HealthCheck(ctx)
		assert.Error(t, err)
	})
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			config: Config{
				BaseURL: "http://localhost:11434",
				Model:   "llama2",
				NumCtx:  2048,
			},
			expectError: false,
		},
		{
			name: "missing model",
			config: Config{
				BaseURL: "http://localhost:11434",
			},
			expectError: true,
			errorMsg:    "model name is required",
		},
		{
			name: "invalid NumCtx (too low)",
			config: Config{
				BaseURL: "http://localhost:11434",
				Model:   "llama2",
				NumCtx:  -1,
			},
			expectError: true,
			errorMsg:    "NumCtx must be positive",
		},
		{
			name: "invalid temperature (too high)",
			config: Config{
				BaseURL:     "http://localhost:11434",
				Model:       "llama2",
				Temperature: 2.5,
			},
			expectError: true,
			errorMsg:    "temperature must be between 0 and 2",
		},
		{
			name: "invalid TopP",
			config: Config{
				BaseURL: "http://localhost:11434",
				Model:   "llama2",
				TopP:    1.5,
			},
			expectError: true,
			errorMsg:    "TopP must be between 0 and 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDefaultOllamaConfig(t *testing.T) {
	config := DefaultConfig()

	require.NotNil(t, config)
	assert.Equal(t, DefaultOllamaURL, config.BaseURL)
	assert.Equal(t, DefaultNumCtx, config.NumCtx)
	assert.Equal(t, DefaultTemperature, config.Temperature)
	assert.Equal(t, DefaultTopK, config.TopK)
	assert.Equal(t, DefaultTopP, config.TopP)
	assert.Equal(t, DefaultTimeout, config.Timeout)
	assert.NotNil(t, config.HTTPClient)
}
