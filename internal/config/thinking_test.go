package config

import (
	"testing"
	"time"

	"github.com/AINative-studio/ainative-code/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultExtendedThinkingConfig(t *testing.T) {
	cfg := DefaultExtendedThinkingConfig()

	require.NotNil(t, cfg)
	assert.True(t, cfg.Enabled)
	assert.False(t, cfg.AutoExpand)
	assert.Equal(t, 10, cfg.MaxDepth)
}

func TestValidateExtendedThinkingConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *ExtendedThinkingConfig
		expectError bool
		errorField  string
	}{
		{
			name:        "nil config is valid",
			config:      nil,
			expectError: false,
		},
		{
			name: "valid config with defaults",
			config: &ExtendedThinkingConfig{
				Enabled:    true,
				AutoExpand: false,
				MaxDepth:   10,
			},
			expectError: false,
		},
		{
			name: "valid config with max depth 1",
			config: &ExtendedThinkingConfig{
				Enabled:    true,
				AutoExpand: true,
				MaxDepth:   1,
			},
			expectError: false,
		},
		{
			name: "valid config with max depth 100",
			config: &ExtendedThinkingConfig{
				Enabled:    true,
				AutoExpand: false,
				MaxDepth:   100,
			},
			expectError: false,
		},
		{
			name: "invalid config with max depth 0",
			config: &ExtendedThinkingConfig{
				Enabled:    true,
				AutoExpand: false,
				MaxDepth:   0,
			},
			expectError: true,
			errorField:  "extended_thinking.max_depth",
		},
		{
			name: "invalid config with negative max depth",
			config: &ExtendedThinkingConfig{
				Enabled:    true,
				AutoExpand: false,
				MaxDepth:   -5,
			},
			expectError: true,
			errorField:  "extended_thinking.max_depth",
		},
		{
			name: "invalid config with max depth over 100",
			config: &ExtendedThinkingConfig{
				Enabled:    true,
				AutoExpand: false,
				MaxDepth:   101,
			},
			expectError: true,
			errorField:  "extended_thinking.max_depth",
		},
		{
			name: "valid config with disabled thinking",
			config: &ExtendedThinkingConfig{
				Enabled:    false,
				AutoExpand: false,
				MaxDepth:   5,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateExtendedThinkingConfig(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorField != "" {
					_, ok := err.(*errors.ConfigError)
					require.True(t, ok, "expected ConfigError")
					assert.Contains(t, err.Error(), tt.errorField)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIsExtendedThinkingEnabled(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected bool
	}{
		{
			name:     "nil config returns false",
			config:   nil,
			expected: false,
		},
		{
			name: "config without anthropic returns false",
			config: &Config{
				LLM: LLMConfig{
					DefaultProvider: "openai",
					Anthropic:       nil,
				},
			},
			expected: false,
		},
		{
			name: "config with anthropic but no extended thinking returns false",
			config: &Config{
				LLM: LLMConfig{
					DefaultProvider: "anthropic",
					Anthropic: &AnthropicConfig{
						APIKey:           "test-key",
						Model:            "claude-3-5-sonnet-20241022",
						ExtendedThinking: nil,
					},
				},
			},
			expected: false,
		},
		{
			name: "config with extended thinking enabled returns true",
			config: &Config{
				LLM: LLMConfig{
					DefaultProvider: "anthropic",
					Anthropic: &AnthropicConfig{
						APIKey: "test-key",
						Model:  "claude-3-5-sonnet-20241022",
						ExtendedThinking: &ExtendedThinkingConfig{
							Enabled:    true,
							AutoExpand: false,
							MaxDepth:   10,
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "config with extended thinking disabled returns false",
			config: &Config{
				LLM: LLMConfig{
					DefaultProvider: "anthropic",
					Anthropic: &AnthropicConfig{
						APIKey: "test-key",
						Model:  "claude-3-5-sonnet-20241022",
						ExtendedThinking: &ExtendedThinkingConfig{
							Enabled:    false,
							AutoExpand: false,
							MaxDepth:   10,
						},
					},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsExtendedThinkingEnabled(tt.config)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetExtendedThinkingConfig(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected *ExtendedThinkingConfig
	}{
		{
			name:     "nil config returns defaults",
			config:   nil,
			expected: DefaultExtendedThinkingConfig(),
		},
		{
			name: "config without anthropic returns defaults",
			config: &Config{
				LLM: LLMConfig{
					DefaultProvider: "openai",
					Anthropic:       nil,
				},
			},
			expected: DefaultExtendedThinkingConfig(),
		},
		{
			name: "config with anthropic but no extended thinking returns defaults",
			config: &Config{
				LLM: LLMConfig{
					DefaultProvider: "anthropic",
					Anthropic: &AnthropicConfig{
						APIKey:           "test-key",
						Model:            "claude-3-5-sonnet-20241022",
						ExtendedThinking: nil,
					},
				},
			},
			expected: DefaultExtendedThinkingConfig(),
		},
		{
			name: "config with custom extended thinking returns custom config",
			config: &Config{
				LLM: LLMConfig{
					DefaultProvider: "anthropic",
					Anthropic: &AnthropicConfig{
						APIKey: "test-key",
						Model:  "claude-3-5-sonnet-20241022",
						ExtendedThinking: &ExtendedThinkingConfig{
							Enabled:    false,
							AutoExpand: true,
							MaxDepth:   25,
						},
					},
				},
			},
			expected: &ExtendedThinkingConfig{
				Enabled:    false,
				AutoExpand: true,
				MaxDepth:   25,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetExtendedThinkingConfig(tt.config)
			require.NotNil(t, result)
			assert.Equal(t, tt.expected.Enabled, result.Enabled)
			assert.Equal(t, tt.expected.AutoExpand, result.AutoExpand)
			assert.Equal(t, tt.expected.MaxDepth, result.MaxDepth)
		})
	}
}

func TestShouldAutoExpandThinking(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected bool
	}{
		{
			name:     "nil config returns default (false)",
			config:   nil,
			expected: false,
		},
		{
			name: "config with auto expand enabled returns true",
			config: &Config{
				LLM: LLMConfig{
					DefaultProvider: "anthropic",
					Anthropic: &AnthropicConfig{
						APIKey: "test-key",
						Model:  "claude-3-5-sonnet-20241022",
						ExtendedThinking: &ExtendedThinkingConfig{
							Enabled:    true,
							AutoExpand: true,
							MaxDepth:   10,
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "config with auto expand disabled returns false",
			config: &Config{
				LLM: LLMConfig{
					DefaultProvider: "anthropic",
					Anthropic: &AnthropicConfig{
						APIKey: "test-key",
						Model:  "claude-3-5-sonnet-20241022",
						ExtendedThinking: &ExtendedThinkingConfig{
							Enabled:    true,
							AutoExpand: false,
							MaxDepth:   10,
						},
					},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShouldAutoExpandThinking(tt.config)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetMaxThinkingDepth(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected int
	}{
		{
			name:     "nil config returns default (10)",
			config:   nil,
			expected: 10,
		},
		{
			name: "config with custom max depth returns custom value",
			config: &Config{
				LLM: LLMConfig{
					DefaultProvider: "anthropic",
					Anthropic: &AnthropicConfig{
						APIKey: "test-key",
						Model:  "claude-3-5-sonnet-20241022",
						ExtendedThinking: &ExtendedThinkingConfig{
							Enabled:    true,
							AutoExpand: false,
							MaxDepth:   50,
						},
					},
				},
			},
			expected: 50,
		},
		{
			name: "config without extended thinking returns default",
			config: &Config{
				LLM: LLMConfig{
					DefaultProvider: "anthropic",
					Anthropic: &AnthropicConfig{
						APIKey:           "test-key",
						Model:            "claude-3-5-sonnet-20241022",
						ExtendedThinking: nil,
					},
				},
			},
			expected: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetMaxThinkingDepth(tt.config)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtendedThinkingConfigIntegration(t *testing.T) {
	// Test a complete config scenario
	cfg := &Config{
		App: AppConfig{
			Name:        "ainative-code",
			Version:     "1.0.0",
			Environment: "development",
			Debug:       true,
		},
		LLM: LLMConfig{
			DefaultProvider: "anthropic",
			Anthropic: &AnthropicConfig{
				APIKey:        "test-api-key",
				Model:         "claude-3-5-sonnet-20241022",
				MaxTokens:     4096,
				Temperature:   0.7,
				TopP:          0.9,
				TopK:          40,
				Timeout:       30 * time.Second,
				RetryAttempts: 3,
				APIVersion:    "2023-06-01",
				ExtendedThinking: &ExtendedThinkingConfig{
					Enabled:    true,
					AutoExpand: true,
					MaxDepth:   20,
				},
			},
		},
	}

	// Validate the config
	err := ValidateExtendedThinkingConfig(cfg.LLM.Anthropic.ExtendedThinking)
	assert.NoError(t, err)

	// Test helper functions
	assert.True(t, IsExtendedThinkingEnabled(cfg))
	assert.True(t, ShouldAutoExpandThinking(cfg))
	assert.Equal(t, 20, GetMaxThinkingDepth(cfg))

	// Get the config
	thinkingCfg := GetExtendedThinkingConfig(cfg)
	require.NotNil(t, thinkingCfg)
	assert.Equal(t, cfg.LLM.Anthropic.ExtendedThinking, thinkingCfg)
}
