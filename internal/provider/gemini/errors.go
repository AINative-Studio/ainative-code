package gemini

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/AINative-studio/ainative-code/internal/provider"
)

// handleAPIError converts Gemini API errors to provider errors
func (g *GeminiProvider) handleAPIError(resp *http.Response, body []byte, model string) error {
	var apiErr geminiError
	if err := json.Unmarshal(body, &apiErr); err != nil {
		return g.HandleHTTPError(resp, body)
	}

	return g.convertAPIError(&apiErr, resp.StatusCode, model)
}

// convertAPIError converts a Gemini API error to a provider error
func (g *GeminiProvider) convertAPIError(apiErr *geminiError, statusCode int, model string) error {
	errCode := apiErr.Error.Code
	errMsg := apiErr.Error.Message
	errStatus := apiErr.Error.Status

	switch {
	case statusCode == http.StatusUnauthorized || statusCode == http.StatusForbidden:
		return provider.NewAuthenticationError("gemini", errMsg)

	case statusCode == http.StatusTooManyRequests:
		return provider.NewRateLimitError("gemini", 0)

	case statusCode == http.StatusBadRequest:
		// Check for various Gemini-specific errors
		if strings.Contains(errMsg, "API key") || strings.Contains(errMsg, "invalid key") {
			return provider.NewAuthenticationError("gemini", errMsg)
		}

		// Check for model errors
		if strings.Contains(errMsg, "model") && strings.Contains(errMsg, "not found") {
			return provider.NewInvalidModelError("gemini", model, supportedModels)
		}

		// Check for token limit errors
		if strings.Contains(errMsg, "token") && (strings.Contains(errMsg, "limit") || strings.Contains(errMsg, "exceed")) {
			return provider.NewContextLengthError("gemini", model, 0, 0)
		}

		// Check for content length errors
		if strings.Contains(errMsg, "content") && strings.Contains(errMsg, "too long") {
			return provider.NewContextLengthError("gemini", model, 0, 0)
		}

		return provider.NewProviderError("gemini", model, fmt.Errorf("invalid request: %s", errMsg))

	case statusCode == http.StatusNotFound:
		// Model not found
		if strings.Contains(errMsg, "model") {
			return provider.NewInvalidModelError("gemini", model, supportedModels)
		}
		return provider.NewProviderError("gemini", model, fmt.Errorf("not found: %s", errMsg))

	case errCode == 400:
		// Additional BadRequest handling based on error code
		if strings.Contains(errStatus, "INVALID_ARGUMENT") {
			return provider.NewProviderError("gemini", model, fmt.Errorf("invalid argument: %s", errMsg))
		}
		return provider.NewProviderError("gemini", model, fmt.Errorf("bad request: %s", errMsg))

	case errCode == 403:
		// Permission denied
		return provider.NewAuthenticationError("gemini", errMsg)

	case errCode == 404:
		// Resource not found
		return provider.NewInvalidModelError("gemini", model, supportedModels)

	case errCode == 429:
		// Rate limit
		return provider.NewRateLimitError("gemini", 0)

	case errCode == 500, errCode == 503:
		// Server errors
		return provider.NewProviderError("gemini", model, fmt.Errorf("server error (%d): %s", errCode, errMsg))

	default:
		// Generic error
		return provider.NewProviderError("gemini", model, fmt.Errorf("%s (code %d): %s", errStatus, errCode, errMsg))
	}
}
