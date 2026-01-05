package bedrock

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/AINative-studio/ainative-code/internal/provider"
)

// bedrockErrorResponse represents an error response from Bedrock
type bedrockErrorResponse struct {
	Type    string `json:"__type,omitempty"`
	Message string `json:"message"`
}

// parseBedrockError parses a Bedrock error response and returns appropriate error type
func parseBedrockError(statusCode int, body []byte, model string) error {
	// Try to parse error response
	errResp, err := parseErrorResponse(body)
	if err != nil {
		// If parsing fails, return generic error
		return provider.NewProviderError("bedrock", model, fmt.Errorf("HTTP %d: %s", statusCode, string(body)))
	}

	message := errResp.Message

	// Check for authentication errors
	if isAuthenticationError(statusCode, message) {
		return provider.NewAuthenticationError("bedrock", message)
	}

	// Check for throttling errors
	if isThrottlingError(statusCode, message) {
		return provider.NewRateLimitError("bedrock", 0)
	}

	// Check for context length errors
	if isContextLengthError(message) {
		requested, max := extractContextLengthInfo(message)
		return provider.NewContextLengthError("bedrock", model, requested, max)
	}

	// Check for validation errors
	if isValidationError(statusCode, message) {
		return provider.NewProviderError("bedrock", model, fmt.Errorf("validation error: %s", message))
	}

	// Check for model not found
	if statusCode == http.StatusNotFound {
		return provider.NewProviderError("bedrock", model, fmt.Errorf("model not found: %s", message))
	}

	// Generic error
	return provider.NewProviderError("bedrock", model, fmt.Errorf("%s", message))
}

// parseErrorResponse parses the error response body
func parseErrorResponse(body []byte) (*bedrockErrorResponse, error) {
	var errResp bedrockErrorResponse
	if err := json.Unmarshal(body, &errResp); err != nil {
		return nil, err
	}
	return &errResp, nil
}

// isAuthenticationError checks if the error is an authentication error
func isAuthenticationError(statusCode int, message string) bool {
	if statusCode == http.StatusUnauthorized || statusCode == http.StatusForbidden {
		return true
	}

	// Check message for authentication-related keywords
	lowerMessage := strings.ToLower(message)
	authKeywords := []string{
		"security token",
		"access denied",
		"unauthorized",
		"forbidden",
		"authentication",
		"credentials",
		"invalid key",
		"expired",
	}

	for _, keyword := range authKeywords {
		if strings.Contains(lowerMessage, keyword) {
			return true
		}
	}

	return false
}

// isThrottlingError checks if the error is a throttling error
func isThrottlingError(statusCode int, message string) bool {
	if statusCode == http.StatusTooManyRequests {
		return true
	}

	// Check message for throttling-related keywords
	lowerMessage := strings.ToLower(message)
	throttlingKeywords := []string{
		"throttl",
		"rate limit",
		"rate exceed",
		"too many requests",
	}

	for _, keyword := range throttlingKeywords {
		if strings.Contains(lowerMessage, keyword) {
			return true
		}
	}

	return false
}

// isContextLengthError checks if the error is a context length error
func isContextLengthError(message string) bool {
	lowerMessage := strings.ToLower(message)
	contextKeywords := []string{
		"prompt is too long",
		"exceeds maximum",
		"too many tokens",
		"context length",
		"context window",
		"maximum context",
	}

	for _, keyword := range contextKeywords {
		if strings.Contains(lowerMessage, keyword) {
			return true
		}
	}

	return false
}

// isValidationError checks if the error is a validation error
func isValidationError(statusCode int, message string) bool {
	if statusCode != http.StatusBadRequest {
		return false
	}

	// Don't treat context length errors as validation errors
	if isContextLengthError(message) {
		return false
	}

	// Check for validation-related keywords
	lowerMessage := strings.ToLower(message)
	validationKeywords := []string{
		"validation",
		"invalid request",
		"missing",
		"required",
		"invalid parameter",
	}

	for _, keyword := range validationKeywords {
		if strings.Contains(lowerMessage, keyword) {
			return true
		}
	}

	return true // All other 400 errors are validation errors
}

// extractContextLengthInfo extracts token counts from context length error messages
func extractContextLengthInfo(message string) (requested int, max int) {
	// Try to extract numbers from common patterns
	// Pattern 1: "has X tokens but the model only supports Y"
	re1 := regexp.MustCompile(`has (\d+) tokens.*?supports (\d+)`)
	if matches := re1.FindStringSubmatch(message); len(matches) == 3 {
		requested, _ = strconv.Atoi(matches[1])
		max, _ = strconv.Atoi(matches[2])
		return
	}

	// Pattern 2: "requested X, maximum Y tokens"
	re2 := regexp.MustCompile(`requested (\d+).*?maximum (\d+)`)
	if matches := re2.FindStringSubmatch(message); len(matches) == 3 {
		requested, _ = strconv.Atoi(matches[1])
		max, _ = strconv.Atoi(matches[2])
		return
	}

	// Pattern 3: "X tokens" (just requested)
	re3 := regexp.MustCompile(`(\d+) tokens`)
	if matches := re3.FindStringSubmatch(message); len(matches) == 2 {
		requested, _ = strconv.Atoi(matches[1])
		return
	}

	return 0, 0
}
