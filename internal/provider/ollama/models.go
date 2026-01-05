package ollama

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Popular LLAMA model identifiers supported by Ollama
var llamaModels = []string{
	"llama2",        // Meta's LLAMA 2 (7B default)
	"llama2:7b",     // LLAMA 2 7B
	"llama2:13b",    // LLAMA 2 13B
	"llama2:70b",    // LLAMA 2 70B
	"llama3",        // Meta's LLAMA 3 (8B default)
	"llama3:8b",     // LLAMA 3 8B
	"llama3:70b",    // LLAMA 3 70B
	"codellama",     // Meta's Code LLAMA (7B default)
	"codellama:7b",  // Code LLAMA 7B
	"codellama:13b", // Code LLAMA 13B
	"codellama:34b", // Code LLAMA 34B
	"codellama:70b", // Code LLAMA 70B
}

// Other popular models supported by Ollama
var otherPopularModels = []string{
	"mistral",       // Mistral 7B
	"mistral:7b",    // Mistral 7B explicit
	"mixtral",       // Mixtral 8x7B
	"mixtral:8x7b",  // Mixtral 8x7B explicit
	"phi",           // Microsoft Phi
	"phi:2.7b",      // Phi 2.7B
	"neural-chat",   // Neural Chat 7B
	"starling-lm",   // Starling LM 7B
	"orca-mini",     // Orca Mini
	"vicuna",        // Vicuna
	"gemma",         // Google Gemma
}

// ListModels fetches the list of available models from Ollama
func ListModels(ctx context.Context, config *Config) ([]ModelInfo, error) {
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = DefaultOllamaURL
	}

	client := config.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: config.Timeout}
	}

	url := fmt.Sprintf("%s/api/tags", baseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, NewOllamaConnectionError(baseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, parseOllamaError(resp.StatusCode, body, "")
	}

	var modelsResp ollamaModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&modelsResp); err != nil {
		return nil, fmt.Errorf("failed to parse models response: %w", err)
	}

	// Convert to public ModelInfo
	models := make([]ModelInfo, len(modelsResp.Models))
	for i, model := range modelsResp.Models {
		models[i] = convertToModelInfo(model)
	}

	return models, nil
}

// GetModelInfo retrieves information about a specific model
func GetModelInfo(ctx context.Context, config *Config, modelName string) (*ModelInfo, error) {
	models, err := ListModels(ctx, config)
	if err != nil {
		return nil, err
	}

	for _, model := range models {
		if model.Name == modelName {
			return &model, nil
		}
	}

	return nil, fmt.Errorf("model '%s' not found in available models", modelName)
}

// IsModelAvailable checks if a specific model is available in Ollama
func IsModelAvailable(ctx context.Context, config *Config, modelName string) (bool, error) {
	models, err := ListModels(ctx, config)
	if err != nil {
		return false, err
	}

	for _, model := range models {
		if model.Name == modelName {
			return true, nil
		}
	}

	return false, nil
}

// GetSupportedModelNames returns a list of well-known model names
// These are models commonly used with Ollama
func GetSupportedModelNames() []string {
	supported := make([]string, 0, len(llamaModels)+len(otherPopularModels))
	supported = append(supported, llamaModels...)
	supported = append(supported, otherPopularModels...)
	return supported
}

// GetPopularLlamaModels returns specifically the LLAMA model variants
func GetPopularLlamaModels() []string {
	models := make([]string, len(llamaModels))
	copy(models, llamaModels)
	return models
}

// formatModelSize formats a size in bytes to human-readable format
func formatModelSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"KB", "MB", "GB", "TB"}
	if exp >= len(units) {
		exp = len(units) - 1
	}

	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), units[exp])
}
