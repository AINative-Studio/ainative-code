package ollama

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListModels(t *testing.T) {
	// Create test models response
	testModels := ollamaModelsResponse{
		Models: []ollamaModelInfo{
			{
				Name:       "llama2",
				ModifiedAt: time.Now(),
				Size:       7365960704,
				Digest:     "abc123",
				Details: ollamaModelDetails{
					Format:        "gguf",
					Family:        "llama",
					ParameterSize: "7B",
				},
			},
			{
				Name:       "llama3:8b",
				ModifiedAt: time.Now(),
				Size:       8638476288,
				Digest:     "def456",
				Details: ollamaModelDetails{
					Format:        "gguf",
					Family:        "llama",
					ParameterSize: "8B",
				},
			},
			{
				Name:       "codellama:13b",
				ModifiedAt: time.Now(),
				Size:       13800000000,
				Digest:     "ghi789",
				Details: ollamaModelDetails{
					Format:        "gguf",
					Family:        "llama",
					ParameterSize: "13B",
				},
			},
		},
	}

	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/tags" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(testModels)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	t.Run("successful model listing", func(t *testing.T) {
		config := &Config{
			BaseURL: server.URL,
			Model:   "llama2",
		}

		ctx := context.Background()
		models, err := ListModels(ctx, config)

		require.NoError(t, err)
		assert.Len(t, models, 3)
		assert.Equal(t, "llama2", models[0].Name)
		assert.Equal(t, "llama3:8b", models[1].Name)
		assert.Equal(t, "codellama:13b", models[2].Name)
		assert.Equal(t, "7B", models[0].ParameterSize)
		assert.Equal(t, "llama", models[0].Family)
	})

	t.Run("cancelled context", func(t *testing.T) {
		config := &Config{
			BaseURL: server.URL,
			Model:   "llama2",
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := ListModels(ctx, config)
		assert.Error(t, err)
	})
}

func TestGetModelInfo(t *testing.T) {
	testModels := ollamaModelsResponse{
		Models: []ollamaModelInfo{
			{
				Name:       "llama2",
				ModifiedAt: time.Now(),
				Size:       7365960704,
				Digest:     "abc123",
				Details: ollamaModelDetails{
					Format:        "gguf",
					Family:        "llama",
					ParameterSize: "7B",
				},
			},
			{
				Name:       "mistral:7b",
				ModifiedAt: time.Now(),
				Size:       7000000000,
				Digest:     "xyz789",
				Details: ollamaModelDetails{
					Format:        "gguf",
					Family:        "mistral",
					ParameterSize: "7B",
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/tags" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(testModels)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	config := &Config{
		BaseURL: server.URL,
		Model:   "llama2",
	}

	t.Run("get existing model info", func(t *testing.T) {
		ctx := context.Background()
		info, err := GetModelInfo(ctx, config, "llama2")

		require.NoError(t, err)
		assert.Equal(t, "llama2", info.Name)
		assert.Equal(t, "7B", info.ParameterSize)
		assert.Equal(t, "llama", info.Family)
		assert.Equal(t, int64(7365960704), info.Size)
	})

	t.Run("get non-existent model", func(t *testing.T) {
		ctx := context.Background()
		_, err := GetModelInfo(ctx, config, "unknown-model")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("get model with different name", func(t *testing.T) {
		ctx := context.Background()
		info, err := GetModelInfo(ctx, config, "mistral:7b")

		require.NoError(t, err)
		assert.Equal(t, "mistral:7b", info.Name)
		assert.Equal(t, "mistral", info.Family)
	})
}

func TestIsModelAvailable(t *testing.T) {
	testModels := ollamaModelsResponse{
		Models: []ollamaModelInfo{
			{Name: "llama2"},
			{Name: "llama3:8b"},
			{Name: "codellama"},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/tags" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(testModels)
		}
	}))
	defer server.Close()

	config := &Config{
		BaseURL: server.URL,
		Model:   "llama2",
	}

	ctx := context.Background()

	t.Run("available model", func(t *testing.T) {
		available, err := IsModelAvailable(ctx, config, "llama2")
		require.NoError(t, err)
		assert.True(t, available)
	})

	t.Run("available model with tag", func(t *testing.T) {
		available, err := IsModelAvailable(ctx, config, "llama3:8b")
		require.NoError(t, err)
		assert.True(t, available)
	})

	t.Run("unavailable model", func(t *testing.T) {
		available, err := IsModelAvailable(ctx, config, "unknown")
		require.NoError(t, err)
		assert.False(t, available)
	})
}

func TestGetSupportedModelNames(t *testing.T) {
	names := GetSupportedModelNames()

	assert.NotEmpty(t, names)
	assert.Contains(t, names, "llama2")
	assert.Contains(t, names, "llama3")
	assert.Contains(t, names, "codellama")
	assert.Contains(t, names, "mistral")
	assert.Contains(t, names, "mixtral")
	assert.Contains(t, names, "phi")
}

func TestGetPopularLlamaModels(t *testing.T) {
	models := GetPopularLlamaModels()

	assert.NotEmpty(t, models)
	assert.Contains(t, models, "llama2")
	assert.Contains(t, models, "llama2:13b")
	assert.Contains(t, models, "llama3")
	assert.Contains(t, models, "llama3:8b")
	assert.Contains(t, models, "codellama")
	assert.Contains(t, models, "codellama:13b")
}

func TestFormatModelSize(t *testing.T) {
	tests := []struct {
		name     string
		size     int64
		expected string
	}{
		{
			name:     "bytes",
			size:     512,
			expected: "512 B",
		},
		{
			name:     "kilobytes",
			size:     1024,
			expected: "1.0 KB",
		},
		{
			name:     "megabytes",
			size:     1024 * 1024,
			expected: "1.0 MB",
		},
		{
			name:     "gigabytes",
			size:     7365960704,
			expected: "6.9 GB",
		},
		{
			name:     "large gigabytes",
			size:     13800000000,
			expected: "12.9 GB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatModelSize(tt.size)
			assert.Equal(t, tt.expected, result)
		})
	}
}
