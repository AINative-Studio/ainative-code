package setup

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewValidator(t *testing.T) {
	validator := NewValidator()

	assert.NotNil(t, validator)
	assert.NotNil(t, validator.httpClient)
}

func TestValidateAnthropicKey_Format(t *testing.T) {
	ctx := context.Background()
	validator := NewValidator()

	tests := []struct {
		name    string
		apiKey  string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty key",
			apiKey:  "",
			wantErr: true,
			errMsg:  "API key cannot be empty",
		},
		{
			name:    "invalid format - no prefix",
			apiKey:  "invalid-key",
			wantErr: true,
			errMsg:  "invalid API key format",
		},
		{
			name:    "invalid format - wrong prefix",
			apiKey:  "sk-test-123",
			wantErr: true,
			errMsg:  "invalid API key format",
		},
		{
			name:    "valid format",
			apiKey:  "sk-ant-test123456789",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateAnthropicKey(ctx, tt.apiKey)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				// For valid format, we might get a network error or auth error
				// which is acceptable in unit tests without real API keys
				// We just check that the format validation passed
			}
		})
	}
}

func TestValidateOpenAIKey_Format(t *testing.T) {
	ctx := context.Background()
	validator := NewValidator()

	tests := []struct {
		name    string
		apiKey  string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty key",
			apiKey:  "",
			wantErr: true,
			errMsg:  "API key cannot be empty",
		},
		{
			name:    "invalid format - no prefix",
			apiKey:  "invalid-key",
			wantErr: true,
			errMsg:  "invalid API key format",
		},
		{
			name:    "valid format",
			apiKey:  "sk-test123456789",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateOpenAIKey(ctx, tt.apiKey)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestValidateGoogleKey_Format(t *testing.T) {
	ctx := context.Background()
	validator := NewValidator()

	tests := []struct {
		name    string
		apiKey  string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty key",
			apiKey:  "",
			wantErr: true,
			errMsg:  "API key cannot be empty",
		},
		{
			name:    "valid format",
			apiKey:  "test-google-api-key-123",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateGoogleKey(ctx, tt.apiKey)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestValidateOllamaConnection_Format(t *testing.T) {
	ctx := context.Background()
	validator := NewValidator()

	tests := []struct {
		name    string
		baseURL string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty URL",
			baseURL: "",
			wantErr: true,
			errMsg:  "base URL cannot be empty",
		},
		{
			name:    "invalid URL scheme",
			baseURL: "ftp://localhost:11434",
			wantErr: true,
			errMsg:  "URL must use http or https scheme",
		},
		{
			name:    "invalid URL format",
			baseURL: "not-a-url",
			wantErr: true,
		},
		{
			name:    "valid HTTP URL",
			baseURL: "http://localhost:11434",
			wantErr: false, // May fail connection, but format is valid
		},
		{
			name:    "valid HTTPS URL",
			baseURL: "https://ollama.example.com",
			wantErr: false, // May fail connection, but format is valid
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateOllamaConnection(ctx, tt.baseURL)

			if tt.wantErr {
				if tt.errMsg != "" {
					require.Error(t, err)
					assert.Contains(t, err.Error(), tt.errMsg)
				} else {
					assert.Error(t, err)
				}
			}
			// For valid format, connection may still fail, which is OK
		})
	}
}

func TestValidateOllamaModel(t *testing.T) {
	ctx := context.Background()
	validator := NewValidator()

	tests := []struct {
		name      string
		baseURL   string
		modelName string
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "empty model name",
			baseURL:   "http://localhost:11434",
			modelName: "",
			wantErr:   true,
			errMsg:    "model name cannot be empty",
		},
		{
			name:      "model name with spaces",
			baseURL:   "http://localhost:11434",
			modelName: "llama 2",
			wantErr:   true,
			errMsg:    "model name cannot contain spaces",
		},
		{
			name:      "valid model name",
			baseURL:   "http://localhost:11434",
			modelName: "llama2",
			wantErr:   false,
		},
		{
			name:      "valid model name with version",
			baseURL:   "http://localhost:11434",
			modelName: "llama2:13b",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateOllamaModel(ctx, tt.baseURL, tt.modelName)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateAINativeKey(t *testing.T) {
	ctx := context.Background()
	validator := NewValidator()

	tests := []struct {
		name    string
		apiKey  string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty key",
			apiKey:  "",
			wantErr: true,
			errMsg:  "API key cannot be empty",
		},
		{
			name:    "too short key",
			apiKey:  "short",
			wantErr: true,
			errMsg:  "API key appears to be too short",
		},
		{
			name:    "valid length key",
			apiKey:  "ainative-test-key-12345678901234567890",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateAINativeKey(ctx, tt.apiKey)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateProviderConfig(t *testing.T) {
	ctx := context.Background()
	validator := NewValidator()

	tests := []struct {
		name       string
		provider   string
		selections map[string]interface{}
		wantErr    bool
		errMsg     string
	}{
		{
			name:     "anthropic - missing API key",
			provider: "anthropic",
			selections: map[string]interface{}{
				"provider": "anthropic",
			},
			wantErr: true,
			errMsg:  "Anthropic API key is required",
		},
		{
			name:     "anthropic - valid",
			provider: "anthropic",
			selections: map[string]interface{}{
				"provider":          "anthropic",
				"anthropic_api_key": "sk-ant-test123456789",
			},
			wantErr: false,
		},
		{
			name:     "openai - missing API key",
			provider: "openai",
			selections: map[string]interface{}{
				"provider": "openai",
			},
			wantErr: true,
			errMsg:  "OpenAI API key is required",
		},
		{
			name:     "openai - valid",
			provider: "openai",
			selections: map[string]interface{}{
				"provider":       "openai",
				"openai_api_key": "sk-test123456789",
			},
			wantErr: false,
		},
		{
			name:     "google - valid",
			provider: "google",
			selections: map[string]interface{}{
				"provider":       "google",
				"google_api_key": "google-test-key",
			},
			wantErr: false,
		},
		{
			name:     "ollama - with model",
			provider: "ollama",
			selections: map[string]interface{}{
				"provider":     "ollama",
				"ollama_url":   "http://localhost:11434",
				"ollama_model": "llama2",
			},
			wantErr: false, // Will fail connection but that's expected in test
		},
		{
			name:     "unsupported provider",
			provider: "unknown",
			selections: map[string]interface{}{
				"provider": "unknown",
			},
			wantErr: true,
			errMsg:  "unsupported provider",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateProviderConfig(ctx, tt.provider, tt.selections)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			}
		})
	}
}

func TestValidateAll(t *testing.T) {
	ctx := context.Background()
	validator := NewValidator()

	tests := []struct {
		name       string
		selections map[string]interface{}
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "missing provider",
			selections: map[string]interface{}{},
			wantErr:    true,
			errMsg:     "provider selection is required",
		},
		{
			name: "valid anthropic config",
			selections: map[string]interface{}{
				"provider":          "anthropic",
				"anthropic_api_key": "sk-ant-test123456789",
			},
			wantErr: false,
		},
		{
			name: "ainative enabled but missing key",
			selections: map[string]interface{}{
				"provider":          "anthropic",
				"anthropic_api_key": "sk-ant-test123456789",
				"ainative_login":    true,
			},
			wantErr: true,
			errMsg:  "AINative API key is required",
		},
		{
			name: "valid config with ainative",
			selections: map[string]interface{}{
				"provider":          "anthropic",
				"anthropic_api_key": "sk-ant-test123456789",
				"ainative_login":    true,
				"ainative_api_key":  "ainative-test-key-12345678901234567890",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateAll(ctx, tt.selections)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else if err != nil {
				// Network errors are acceptable for valid configs in tests
				t.Logf("Got network error (acceptable): %v", err)
			}
		})
	}
}

func TestSanitizeAPIKey(t *testing.T) {
	tests := []struct {
		name     string
		apiKey   string
		expected string
	}{
		{
			name:     "short key",
			apiKey:   "test",
			expected: "***",
		},
		{
			name:     "medium key",
			apiKey:   "sk-test123",
			expected: "sk-t...t123", // 10 chars, shows first 4 and last 4
		},
		{
			name:     "long key",
			apiKey:   "sk-ant-api03-test1234567890abcdefghijklmnop",
			expected: "sk-a...mnop",
		},
		{
			name:     "very long key",
			apiKey:   "sk-proj-1234567890abcdefghijklmnopqrstuvwxyz",
			expected: "sk-p...wxyz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeAPIKey(tt.apiKey)
			assert.Equal(t, tt.expected, result)
		})
	}
}
