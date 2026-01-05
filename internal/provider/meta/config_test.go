package meta

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, DefaultBaseURL, config.BaseURL)
	assert.Equal(t, ModelLlama4Maverick, config.Model)
	assert.Equal(t, 0.7, config.Temperature)
	assert.Equal(t, 0.9, config.TopP)
	assert.Equal(t, 2048, config.MaxTokens)
	assert.Equal(t, DefaultTimeout, config.Timeout)
	assert.NotNil(t, config.HTTPClient)
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: &Config{
				APIKey:  "test-key",
				BaseURL: DefaultBaseURL,
				Model:   ModelLlama4Maverick,
			},
			wantErr: false,
		},
		{
			name: "missing API key",
			config: &Config{
				BaseURL: DefaultBaseURL,
				Model:   ModelLlama4Maverick,
			},
			wantErr: true,
			errMsg:  "API key is required",
		},
		{
			name: "missing base URL",
			config: &Config{
				APIKey: "test-key",
				Model:  ModelLlama4Maverick,
			},
			wantErr: true,
			errMsg:  "base URL is required",
		},
		{
			name: "missing model",
			config: &Config{
				APIKey:  "test-key",
				BaseURL: DefaultBaseURL,
			},
			wantErr: true,
			errMsg:  "model is required",
		},
		{
			name: "invalid model",
			config: &Config{
				APIKey:  "test-key",
				BaseURL: DefaultBaseURL,
				Model:   "gpt-4",
			},
			wantErr: true,
			errMsg:  "invalid model",
		},
		{
			name: "temperature too low",
			config: &Config{
				APIKey:      "test-key",
				BaseURL:     DefaultBaseURL,
				Model:       ModelLlama4Maverick,
				Temperature: -0.1,
			},
			wantErr: true,
			errMsg:  "temperature must be between 0 and 2",
		},
		{
			name: "temperature too high",
			config: &Config{
				APIKey:      "test-key",
				BaseURL:     DefaultBaseURL,
				Model:       ModelLlama4Maverick,
				Temperature: 2.1,
			},
			wantErr: true,
			errMsg:  "temperature must be between 0 and 2",
		},
		{
			name: "top_p too low",
			config: &Config{
				APIKey:  "test-key",
				BaseURL: DefaultBaseURL,
				Model:   ModelLlama4Maverick,
				TopP:    -0.1,
			},
			wantErr: true,
			errMsg:  "top_p must be between 0 and 1",
		},
		{
			name: "top_p too high",
			config: &Config{
				APIKey:  "test-key",
				BaseURL: DefaultBaseURL,
				Model:   ModelLlama4Maverick,
				TopP:    1.1,
			},
			wantErr: true,
			errMsg:  "top_p must be between 0 and 1",
		},
		{
			name: "negative max_tokens",
			config: &Config{
				APIKey:    "test-key",
				BaseURL:   DefaultBaseURL,
				Model:     ModelLlama4Maverick,
				MaxTokens: -1,
			},
			wantErr: true,
			errMsg:  "max_tokens must be non-negative",
		},
		{
			name: "presence_penalty too low",
			config: &Config{
				APIKey:          "test-key",
				BaseURL:         DefaultBaseURL,
				Model:           ModelLlama4Maverick,
				PresencePenalty: -2.1,
			},
			wantErr: true,
			errMsg:  "presence_penalty must be between -2 and 2",
		},
		{
			name: "frequency_penalty too high",
			config: &Config{
				APIKey:           "test-key",
				BaseURL:          DefaultBaseURL,
				Model:            ModelLlama4Maverick,
				FrequencyPenalty: 2.1,
			},
			wantErr: true,
			errMsg:  "frequency_penalty must be between -2 and 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfig_SetDefaults(t *testing.T) {
	config := &Config{
		APIKey: "test-key",
	}

	config.SetDefaults()

	assert.Equal(t, DefaultBaseURL, config.BaseURL)
	assert.Equal(t, ModelLlama4Maverick, config.Model)
	assert.Equal(t, 0.7, config.Temperature)
	assert.Equal(t, 0.9, config.TopP)
	assert.Equal(t, 2048, config.MaxTokens)
	assert.Equal(t, DefaultTimeout, config.Timeout)
	assert.NotNil(t, config.HTTPClient)
}

func TestConfig_SetDefaults_PreservesExisting(t *testing.T) {
	config := &Config{
		APIKey:      "test-key",
		BaseURL:     "https://custom.url",
		Model:       ModelLlama33_8B,
		Temperature: 0.5,
		TopP:        0.95,
		MaxTokens:   1000,
		Timeout:     30 * time.Second,
	}

	config.SetDefaults()

	// Should preserve existing values
	assert.Equal(t, "https://custom.url", config.BaseURL)
	assert.Equal(t, ModelLlama33_8B, config.Model)
	assert.Equal(t, 0.5, config.Temperature)
	assert.Equal(t, 0.95, config.TopP)
	assert.Equal(t, 1000, config.MaxTokens)
	assert.Equal(t, 30*time.Second, config.Timeout)
}
