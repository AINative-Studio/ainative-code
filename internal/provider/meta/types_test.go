package meta

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSupportedModels(t *testing.T) {
	models := GetSupportedModels()

	assert.Len(t, models, 4)

	// Check Llama 4 Maverick
	maverick := models[0]
	assert.Equal(t, ModelLlama4Maverick, maverick.ID)
	assert.Equal(t, "Llama 4 Maverick", maverick.Name)
	assert.Equal(t, "400B total, 17B active", maverick.ParameterCount)
	assert.Equal(t, "Mixture of Experts (128 experts)", maverick.Architecture)
	assert.Equal(t, 8192, maverick.MaxTokens)
	assert.True(t, maverick.Recommended)

	// Check Llama 4 Scout
	scout := models[1]
	assert.Equal(t, ModelLlama4Scout, scout.ID)
	assert.Equal(t, "109B total, 17B active", scout.ParameterCount)
	assert.False(t, scout.Recommended)

	// Check Llama 3.3 70B
	llama33_70b := models[2]
	assert.Equal(t, ModelLlama33_70B, llama33_70b.ID)
	assert.Equal(t, "70B", llama33_70b.ParameterCount)
	assert.Equal(t, "Dense transformer", llama33_70b.Architecture)

	// Check Llama 3.3 8B
	llama33_8b := models[3]
	assert.Equal(t, ModelLlama33_8B, llama33_8b.ID)
	assert.Equal(t, "8B", llama33_8b.ParameterCount)
}

func TestIsValidModel(t *testing.T) {
	tests := []struct {
		name     string
		modelID  string
		expected bool
	}{
		{
			name:     "Llama 4 Maverick",
			modelID:  ModelLlama4Maverick,
			expected: true,
		},
		{
			name:     "Llama 4 Scout",
			modelID:  ModelLlama4Scout,
			expected: true,
		},
		{
			name:     "Llama 3.3 70B",
			modelID:  ModelLlama33_70B,
			expected: true,
		},
		{
			name:     "Llama 3.3 8B",
			modelID:  ModelLlama33_8B,
			expected: true,
		},
		{
			name:     "invalid model",
			modelID:  "gpt-4",
			expected: false,
		},
		{
			name:     "empty string",
			modelID:  "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsValidModel(tt.modelID))
		})
	}
}

func TestConstants(t *testing.T) {
	assert.Equal(t, "https://api.llama.com/compat/v1", DefaultBaseURL)
	assert.Equal(t, "Llama-4-Maverick-17B-128E-Instruct-FP8", ModelLlama4Maverick)
	assert.Equal(t, "Llama-4-Scout-17B-16E", ModelLlama4Scout)
	assert.Equal(t, "Llama-3.3-70B-Instruct", ModelLlama33_70B)
	assert.Equal(t, "Llama-3.3-8B-Instruct", ModelLlama33_8B)
}
