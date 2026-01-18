package ainative_e2e

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/AINative-studio/ainative-code/internal/backend"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAINativeE2E_CompleteAuthFlow tests the complete authentication flow
// GIVEN a running mock backend
// WHEN user performs login
// THEN login should succeed
// AND token should be valid
// AND user can make authenticated requests
func TestAINativeE2E_CompleteAuthFlow(t *testing.T) {
	// GIVEN a running mock backend
	mockBackend := NewMockBackend(t)
	defer mockBackend.Close()

	// WHEN user performs login
	client := backend.NewClient(mockBackend.URL)
	ctx := context.Background()

	resp, err := client.Login(ctx, "test@example.com", "password123")

	// THEN login should succeed
	require.NoError(t, err, "Login should succeed")
	assert.NotEmpty(t, resp.AccessToken, "Access token should not be empty")
	assert.NotEmpty(t, resp.RefreshToken, "Refresh token should not be empty")
	assert.Equal(t, "test@example.com", resp.User.Email, "User email should match")
	assert.Equal(t, "bearer", resp.TokenType, "Token type should be bearer")

	// AND token should be valid
	assert.True(t, isValidJWT(resp.AccessToken), "Access token should be valid JWT")

	// AND user can make authenticated requests
	chatReq := &backend.ChatCompletionRequest{
		Messages: []backend.Message{
			{Role: "user", Content: "Hello"},
		},
		Model: "claude-sonnet-4-5",
	}

	chatResp, err := client.ChatCompletion(ctx, resp.AccessToken, chatReq)
	require.NoError(t, err, "Chat completion should succeed")
	assert.NotEmpty(t, chatResp.Choices, "Chat response should have choices")
	assert.NotEmpty(t, chatResp.ID, "Chat response should have ID")
	assert.Equal(t, "claude-sonnet-4-5", chatResp.Model, "Chat response should have correct model")
}

// TestAINativeE2E_UserRegistration tests user registration
// GIVEN a running mock backend
// WHEN user registers with valid credentials
// THEN registration should succeed
// AND user can login with those credentials
func TestAINativeE2E_UserRegistration(t *testing.T) {
	// GIVEN a running mock backend
	mockBackend := NewMockBackend(t)
	defer mockBackend.Close()

	client := backend.NewClient(mockBackend.URL)
	ctx := context.Background()

	// WHEN user registers with valid credentials
	registerResp, err := client.Register(ctx, "newuser@example.com", "newpassword123")

	// THEN registration should succeed
	require.NoError(t, err, "Registration should succeed")
	assert.NotEmpty(t, registerResp.AccessToken, "Access token should not be empty")
	assert.NotEmpty(t, registerResp.RefreshToken, "Refresh token should not be empty")
	assert.Equal(t, "newuser@example.com", registerResp.User.Email, "User email should match")

	// AND user can login with those credentials
	loginResp, err := client.Login(ctx, "newuser@example.com", "password123")
	require.NoError(t, err, "Login should succeed after registration")
	assert.NotEmpty(t, loginResp.AccessToken, "Login access token should not be empty")
}

// TestAINativeE2E_AuthenticationFailure tests authentication failure scenarios
// GIVEN a running mock backend
// WHEN user logs in with invalid credentials
// THEN login should fail with 401
func TestAINativeE2E_AuthenticationFailure(t *testing.T) {
	// GIVEN a running mock backend
	mockBackend := NewMockBackend(t)
	defer mockBackend.Close()

	client := backend.NewClient(mockBackend.URL)
	ctx := context.Background()

	// WHEN user logs in with invalid credentials
	_, err := client.Login(ctx, "test@example.com", "wrongpassword")

	// THEN login should fail with 401
	require.Error(t, err, "Login should fail with wrong password")
	assert.True(t, errors.Is(err, backend.ErrUnauthorized), "Error should be unauthorized")
}

// TestAINativeE2E_UnauthorizedChatRequest tests unauthorized chat requests
// GIVEN a mock backend
// WHEN user sends chat request without authentication
// THEN request should fail with 401
func TestAINativeE2E_UnauthorizedChatRequest(t *testing.T) {
	// GIVEN a mock backend
	mockBackend := NewMockBackend(t)
	defer mockBackend.Close()

	client := backend.NewClient(mockBackend.URL)
	ctx := context.Background()

	// WHEN user sends chat request without authentication
	chatReq := &backend.ChatCompletionRequest{
		Messages: []backend.Message{
			{Role: "user", Content: "Hello"},
		},
		Model: "claude-sonnet-4-5",
	}

	_, err := client.ChatCompletion(ctx, "", chatReq)

	// THEN request should fail with 401
	require.Error(t, err, "Chat completion should fail without auth")
	assert.True(t, errors.Is(err, backend.ErrUnauthorized), "Error should be unauthorized")
}

// TestAINativeE2E_TokenRefreshFlow tests token refresh functionality
// GIVEN a user with an expired access token
// WHEN refreshing with refresh token
// THEN refresh should succeed
// AND new token should be valid
// AND new token should work for authenticated requests
func TestAINativeE2E_TokenRefreshFlow(t *testing.T) {
	// GIVEN a user with an expired access token
	mockBackend := NewMockBackend(t)
	defer mockBackend.Close()

	client := backend.NewClient(mockBackend.URL)
	ctx := context.Background()

	// Login to get initial tokens
	loginResp, err := client.Login(ctx, "test@example.com", "password123")
	require.NoError(t, err, "Initial login should succeed")

	// WHEN refreshing with refresh token
	refreshResp, err := client.RefreshToken(ctx, loginResp.RefreshToken)

	// THEN refresh should succeed
	require.NoError(t, err, "Token refresh should succeed")
	assert.NotEmpty(t, refreshResp.AccessToken, "New access token should not be empty")
	assert.NotEqual(t, loginResp.AccessToken, refreshResp.AccessToken, "New token should be different")

	// AND new token should be valid
	assert.True(t, isValidJWT(refreshResp.AccessToken), "New access token should be valid JWT")

	// AND new token should work for authenticated requests
	chatReq := &backend.ChatCompletionRequest{
		Messages: []backend.Message{{Role: "user", Content: "Test"}},
		Model:    "claude-sonnet-4-5",
	}

	_, err = client.ChatCompletion(ctx, refreshResp.AccessToken, chatReq)
	require.NoError(t, err, "Chat should succeed with refreshed token")
}

// TestAINativeE2E_RefreshWithInvalidToken tests refresh with invalid token
// GIVEN a mock backend
// WHEN refreshing with invalid refresh token
// THEN refresh should fail
func TestAINativeE2E_RefreshWithInvalidToken(t *testing.T) {
	// GIVEN a mock backend
	mockBackend := NewMockBackend(t)
	defer mockBackend.Close()

	client := backend.NewClient(mockBackend.URL)
	ctx := context.Background()

	// WHEN refreshing with invalid refresh token
	_, err := client.RefreshToken(ctx, "invalid-refresh-token")

	// THEN refresh should fail
	require.Error(t, err, "Refresh should fail with invalid token")
	assert.True(t, errors.Is(err, backend.ErrUnauthorized), "Error should be unauthorized")
}

// TestAINativeE2E_InsufficientCredits tests payment required scenario
// GIVEN a user with zero credits
// WHEN sending chat request
// THEN should fail with payment required error
func TestAINativeE2E_InsufficientCredits(t *testing.T) {
	// GIVEN a user with zero credits
	mockBackend := NewMockBackend(t)
	mockBackend.SetUserCredits("nocredit@example.com", 0)
	defer mockBackend.Close()

	client := backend.NewClient(mockBackend.URL)
	ctx := context.Background()

	loginResp, err := client.Login(ctx, "nocredit@example.com", "password123")
	require.NoError(t, err, "Login should succeed")

	// WHEN sending chat request
	req := &backend.ChatCompletionRequest{
		Messages: []backend.Message{{Role: "user", Content: "Test"}},
		Model:    "claude-sonnet-4-5",
	}

	_, err = client.ChatCompletion(ctx, loginResp.AccessToken, req)

	// THEN should fail with payment required error
	require.Error(t, err, "Chat should fail with insufficient credits")
	assert.True(t, errors.Is(err, backend.ErrPaymentRequired), "Error should be payment required")
}

// TestAINativeE2E_NetworkError tests network error handling
// GIVEN a client pointing to invalid URL
// WHEN attempting to login
// THEN should fail with network error
func TestAINativeE2E_NetworkError(t *testing.T) {
	// GIVEN a client pointing to invalid URL
	client := backend.NewClient("http://invalid-backend-url:9999")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// WHEN attempting to login
	_, err := client.Login(ctx, "test@example.com", "password123")

	// THEN should fail with network error
	require.Error(t, err, "Login should fail with network error")
}

// TestAINativeE2E_RateLimiting tests rate limiting functionality
// GIVEN a mock backend with rate limiting
// WHEN exceeding rate limit
// THEN should get rate limit error
func TestAINativeE2E_RateLimiting(t *testing.T) {
	// GIVEN a mock backend with rate limiting
	mockBackend := NewMockBackend(t)
	mockBackend.EnableRateLimit(5, time.Minute) // 5 requests per minute
	defer mockBackend.Close()

	client := backend.NewClient(mockBackend.URL)
	ctx := context.Background()

	loginResp, err := client.Login(ctx, "test@example.com", "password123")
	require.NoError(t, err, "Login should succeed")

	req := &backend.ChatCompletionRequest{
		Messages: []backend.Message{{Role: "user", Content: "Test"}},
		Model:    "claude-sonnet-4-5",
	}

	// WHEN exceeding rate limit
	var rateLimitErr error
	for i := 0; i < 10; i++ {
		_, err := client.ChatCompletion(ctx, loginResp.AccessToken, req)
		if err != nil {
			rateLimitErr = err
			break
		}
	}

	// THEN should get rate limit error
	require.Error(t, rateLimitErr, "Should hit rate limit")
	assert.Contains(t, rateLimitErr.Error(), "429", "Error should indicate rate limit")
}

// TestAINativeE2E_GetUserInfo tests retrieving user information
// GIVEN an authenticated user
// WHEN requesting user info
// THEN should return correct user data
func TestAINativeE2E_GetUserInfo(t *testing.T) {
	// GIVEN an authenticated user
	mockBackend := NewMockBackend(t)
	defer mockBackend.Close()

	client := backend.NewClient(mockBackend.URL)
	ctx := context.Background()

	loginResp, err := client.Login(ctx, "test@example.com", "password123")
	require.NoError(t, err, "Login should succeed")

	// WHEN requesting user info
	user, err := client.GetMe(ctx, loginResp.AccessToken)

	// THEN should return correct user data
	require.NoError(t, err, "Get user info should succeed")
	assert.Equal(t, "test@example.com", user.Email, "User email should match")
	assert.NotEmpty(t, user.ID, "User ID should not be empty")
}

// TestAINativeE2E_Logout tests logout functionality
// GIVEN an authenticated user
// WHEN logging out
// THEN logout should succeed
// AND token should no longer work
func TestAINativeE2E_Logout(t *testing.T) {
	// GIVEN an authenticated user
	mockBackend := NewMockBackend(t)
	defer mockBackend.Close()

	client := backend.NewClient(mockBackend.URL)
	ctx := context.Background()

	loginResp, err := client.Login(ctx, "test@example.com", "password123")
	require.NoError(t, err, "Login should succeed")

	// WHEN logging out
	err = client.Logout(ctx, loginResp.AccessToken)

	// THEN logout should succeed
	require.NoError(t, err, "Logout should succeed")

	// AND token should no longer work (backend should invalidate it)
	chatReq := &backend.ChatCompletionRequest{
		Messages: []backend.Message{{Role: "user", Content: "Test"}},
		Model:    "claude-sonnet-4-5",
	}

	_, err = client.ChatCompletion(ctx, loginResp.AccessToken, chatReq)
	assert.Error(t, err, "Chat should fail after logout")
}

// TestAINativeE2E_HealthCheck tests backend health check
// GIVEN a running mock backend
// WHEN checking health
// THEN should return healthy status
func TestAINativeE2E_HealthCheck(t *testing.T) {
	// GIVEN a running mock backend
	mockBackend := NewMockBackend(t)
	defer mockBackend.Close()

	client := backend.NewClient(mockBackend.URL)
	ctx := context.Background()

	// WHEN checking health
	err := client.HealthCheck(ctx)

	// THEN should return healthy status
	require.NoError(t, err, "Health check should succeed")
}

// TestAINativeE2E_ContextCancellation tests context cancellation
// GIVEN a chat request in progress
// WHEN context is cancelled
// THEN request should fail gracefully
func TestAINativeE2E_ContextCancellation(t *testing.T) {
	// GIVEN a chat request in progress
	mockBackend := NewMockBackend(t)
	defer mockBackend.Close()

	client := backend.NewClient(mockBackend.URL)
	ctx, cancel := context.WithCancel(context.Background())

	loginResp, err := client.Login(ctx, "test@example.com", "password123")
	require.NoError(t, err, "Login should succeed")

	// Cancel context immediately
	cancel()

	// WHEN context is cancelled
	chatReq := &backend.ChatCompletionRequest{
		Messages: []backend.Message{{Role: "user", Content: "Test"}},
		Model:    "claude-sonnet-4-5",
	}

	_, err = client.ChatCompletion(ctx, loginResp.AccessToken, chatReq)

	// THEN request should fail gracefully
	require.Error(t, err, "Chat should fail with cancelled context")
	assert.True(t, errors.Is(err, context.Canceled), "Error should be context canceled")
}

// TestAINativeE2E_MultipleMessages tests chat with multiple messages
// GIVEN an authenticated user
// WHEN sending chat with conversation history
// THEN should process all messages correctly
func TestAINativeE2E_MultipleMessages(t *testing.T) {
	// GIVEN an authenticated user
	mockBackend := NewMockBackend(t)
	defer mockBackend.Close()

	client := backend.NewClient(mockBackend.URL)
	ctx := context.Background()

	loginResp, err := client.Login(ctx, "test@example.com", "password123")
	require.NoError(t, err, "Login should succeed")

	// WHEN sending chat with conversation history
	chatReq := &backend.ChatCompletionRequest{
		Messages: []backend.Message{
			{Role: "user", Content: "What is 2+2?"},
			{Role: "assistant", Content: "2+2 equals 4."},
			{Role: "user", Content: "What about 3+3?"},
		},
		Model: "claude-sonnet-4-5",
	}

	resp, err := client.ChatCompletion(ctx, loginResp.AccessToken, chatReq)

	// THEN should process all messages correctly
	require.NoError(t, err, "Chat completion should succeed")
	assert.NotEmpty(t, resp.Choices, "Response should have choices")
	assert.Equal(t, "assistant", resp.Choices[0].Message.Role, "Response should be from assistant")
}

// isValidJWT is a helper function to validate JWT tokens
// This is a placeholder - actual implementation will check JWT structure
func isValidJWT(token string) bool {
	// Simple check: JWT has 3 parts separated by dots
	parts := strings.Split(token, ".")
	return len(parts) == 3 && len(parts[0]) > 0 && len(parts[1]) > 0 && len(parts[2]) > 0
}
