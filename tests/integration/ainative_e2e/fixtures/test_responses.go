package fixtures

import "github.com/AINative-studio/ainative-code/internal/backend"

// GetDefaultChatResponse returns a default chat completion response for testing
func GetDefaultChatResponse() *backend.ChatCompletionResponse {
	return &backend.ChatCompletionResponse{
		ID:    "chatcmpl-test-123",
		Model: "claude-sonnet-4-5",
		Choices: []backend.Choice{
			{
				Message: backend.Message{
					Role:    "assistant",
					Content: "This is a test response from the mock backend server.",
				},
				Index: 0,
			},
		},
		Usage: backend.Usage{
			PromptTokens:     10,
			CompletionTokens: 15,
			TotalTokens:      25,
		},
	}
}

// GetStreamingChatChunks returns sample streaming chunks for testing
func GetStreamingChatChunks() []string {
	return []string{
		"This ",
		"is ",
		"a ",
		"streaming ",
		"response ",
		"from ",
		"the ",
		"mock ",
		"backend.",
	}
}

// GetLargeStreamingChunks returns a large number of chunks for testing
func GetLargeStreamingChunks(count int) []string {
	chunks := make([]string, count)
	for i := 0; i < count; i++ {
		chunks[i] = "chunk "
	}
	return chunks
}

// GetDefaultUser returns a default user for testing
func GetDefaultUser(email string) *backend.User {
	return &backend.User{
		ID:    "user-test-" + email,
		Email: email,
	}
}

// GetTokenResponse returns a default token response for testing
func GetTokenResponse(email string) *backend.TokenResponse {
	return &backend.TokenResponse{
		AccessToken:  CreateValidTokenForEmail(email),
		RefreshToken: CreateRefreshToken(email),
		TokenType:    "bearer",
		User: backend.User{
			ID:    "user-test-" + email,
			Email: email,
		},
	}
}

// GetHealthResponse returns a default health response for testing
func GetHealthResponse() map[string]string {
	return map[string]string{
		"status": "healthy",
	}
}

// GetErrorResponse returns an error response for testing
func GetErrorResponse(message string) map[string]string {
	return map[string]string{
		"detail": message,
	}
}
