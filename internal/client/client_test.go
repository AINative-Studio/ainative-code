package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AINative-studio/ainative-code/internal/auth"
	"github.com/AINative-studio/ainative-code/internal/client"
)

// mockAuthClient implements auth.Client interface for testing
type mockAuthClient struct {
	tokens         *auth.TokenPair
	refreshCalled  bool
	refreshError   error
	shouldRefresh  bool
}

func (m *mockAuthClient) Authenticate(ctx context.Context) (*auth.TokenPair, error) {
	return m.tokens, nil
}

func (m *mockAuthClient) GetStoredTokens(ctx context.Context) (*auth.TokenPair, error) {
	if m.tokens == nil {
		return nil, client.ErrNoAuthClient
	}
	return m.tokens, nil
}

func (m *mockAuthClient) RefreshToken(ctx context.Context, refreshToken *auth.RefreshToken) (*auth.TokenPair, error) {
	m.refreshCalled = true
	if m.refreshError != nil {
		return nil, m.refreshError
	}

	// Simulate successful refresh by updating access token
	m.tokens.AccessToken = &auth.AccessToken{
		Raw:       "new-access-token",
		ExpiresAt: time.Now().Add(1 * time.Hour),
		UserID:    "test-user",
		Email:     "test@example.com",
		Issuer:    "ainative-auth",
		Audience:  "ainative-code",
	}

	return m.tokens, nil
}

func (m *mockAuthClient) StoreTokens(ctx context.Context, tokens *auth.TokenPair) error {
	m.tokens = tokens
	return nil
}

func (m *mockAuthClient) ClearTokens(ctx context.Context) error {
	m.tokens = nil
	return nil
}

func (m *mockAuthClient) ValidateToken(ctx context.Context, token *auth.AccessToken) bool {
	return token != nil && !token.IsExpired()
}

func newMockAuthClient(accessToken, refreshToken string) *mockAuthClient {
	tokens := &auth.TokenPair{
		AccessToken: &auth.AccessToken{
			Raw:       accessToken,
			ExpiresAt: time.Now().Add(1 * time.Hour),
			UserID:    "test-user",
			Email:     "test@example.com",
			Issuer:    "ainative-auth",
			Audience:  "ainative-code",
		},
		ReceivedAt: time.Now(),
	}

	if refreshToken != "" {
		tokens.RefreshToken = &auth.RefreshToken{
			Raw:       refreshToken,
			ExpiresAt: time.Now().Add(24 * time.Hour),
			UserID:    "test-user",
			SessionID: "test-session",
			Issuer:    "ainative-auth",
			Audience:  "ainative-code",
		}
	}

	return &mockAuthClient{
		tokens: tokens,
	}
}

// TestClientBasicGet tests basic GET request functionality
func TestClientBasicGet(t *testing.T) {
	expectedResponse := map[string]interface{}{
		"status": "success",
		"data":   "test data",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/test", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expectedResponse)
	}))
	defer server.Close()

	c := client.New(
		client.WithBaseURL(server.URL),
		client.WithTimeout(5*time.Second),
	)

	resp, err := c.Get(context.Background(), "/api/test")
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(resp, &result)
	require.NoError(t, err)
	assert.Equal(t, "success", result["status"])
}

// TestClientBasicPost tests basic POST request functionality
func TestClientBasicPost(t *testing.T) {
	requestBody := map[string]interface{}{
		"name":  "Test User",
		"email": "test@example.com",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/users", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var received map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&received)
		require.NoError(t, err)
		assert.Equal(t, "Test User", received["name"])

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"id":     "user-123",
			"status": "created",
		})
	}))
	defer server.Close()

	c := client.New(
		client.WithBaseURL(server.URL),
	)

	resp, err := c.Post(context.Background(), "/api/users", requestBody)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(resp, &result)
	require.NoError(t, err)
	assert.Equal(t, "user-123", result["id"])
}

// TestClientWithJWTAuthentication tests JWT bearer token injection
func TestClientWithJWTAuthentication(t *testing.T) {
	mockAuth := newMockAuthClient("test-access-token", "test-refresh-token")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify Authorization header is present
		authHeader := r.Header.Get("Authorization")
		assert.Equal(t, "Bearer test-access-token", authHeader)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "authenticated"})
	}))
	defer server.Close()

	c := client.New(
		client.WithBaseURL(server.URL),
		client.WithAuthClient(mockAuth),
	)

	resp, err := c.Get(context.Background(), "/api/protected")
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(resp, &result)
	require.NoError(t, err)
	assert.Equal(t, "authenticated", result["status"])
}

// TestClientSkipAuth tests skipping authentication for specific requests
func TestClientSkipAuth(t *testing.T) {
	mockAuth := newMockAuthClient("test-access-token", "test-refresh-token")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify Authorization header is NOT present
		authHeader := r.Header.Get("Authorization")
		assert.Empty(t, authHeader)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "public"})
	}))
	defer server.Close()

	c := client.New(
		client.WithBaseURL(server.URL),
		client.WithAuthClient(mockAuth),
	)

	resp, err := c.Get(context.Background(), "/api/public", client.WithSkipAuth())
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(resp, &result)
	require.NoError(t, err)
	assert.Equal(t, "public", result["status"])
}

// TestClientTokenRefreshOn401 tests automatic token refresh when receiving 401
func TestClientTokenRefreshOn401(t *testing.T) {
	mockAuth := newMockAuthClient("expired-token", "valid-refresh-token")
	requestCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		authHeader := r.Header.Get("Authorization")

		if requestCount == 1 {
			// First request with expired token returns 401
			assert.Equal(t, "Bearer expired-token", authHeader)
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":"token expired"}`))
			return
		}

		// Second request after token refresh should succeed
		assert.Equal(t, "Bearer new-access-token", authHeader)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	}))
	defer server.Close()

	c := client.New(
		client.WithBaseURL(server.URL),
		client.WithAuthClient(mockAuth),
		client.WithMaxRetries(3),
	)

	resp, err := c.Get(context.Background(), "/api/protected")
	require.NoError(t, err)

	// Verify token refresh was called
	assert.True(t, mockAuth.refreshCalled, "Token refresh should have been called")

	// Verify the request eventually succeeded
	var result map[string]interface{}
	err = json.Unmarshal(resp, &result)
	require.NoError(t, err)
	assert.Equal(t, "success", result["status"])

	// Verify request was retried
	assert.Equal(t, 2, requestCount, "Should have made 2 requests (initial + retry after refresh)")
}

// TestClientRetryOnServerError tests retry logic for 5xx errors
func TestClientRetryOnServerError(t *testing.T) {
	requestCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++

		if requestCount < 3 {
			// First two requests fail with 503
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"error":"service unavailable"}`))
			return
		}

		// Third request succeeds
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	}))
	defer server.Close()

	c := client.New(
		client.WithBaseURL(server.URL),
		client.WithMaxRetries(3),
	)

	resp, err := c.Get(context.Background(), "/api/flaky")
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(resp, &result)
	require.NoError(t, err)
	assert.Equal(t, "success", result["status"])

	// Verify retry happened
	assert.Equal(t, 3, requestCount, "Should have made 3 requests (2 failures + 1 success)")
}

// TestClientRetryOnRateLimited tests retry logic for 429 rate limiting
func TestClientRetryOnRateLimited(t *testing.T) {
	requestCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++

		if requestCount == 1 {
			// First request is rate limited
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":"rate limited"}`))
			return
		}

		// Second request succeeds
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	}))
	defer server.Close()

	c := client.New(
		client.WithBaseURL(server.URL),
		client.WithMaxRetries(3),
	)

	resp, err := c.Get(context.Background(), "/api/rate-limited")
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(resp, &result)
	require.NoError(t, err)
	assert.Equal(t, "success", result["status"])

	assert.Equal(t, 2, requestCount, "Should have retried once")
}

// TestClientMaxRetriesExceeded tests that retries stop after max attempts
func TestClientMaxRetriesExceeded(t *testing.T) {
	requestCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		// Always return 503
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"error":"always failing"}`))
	}))
	defer server.Close()

	c := client.New(
		client.WithBaseURL(server.URL),
		client.WithMaxRetries(2), // Max 2 retries = 3 total attempts
	)

	_, err := c.Get(context.Background(), "/api/always-failing")
	require.Error(t, err)
	// After max retries, should return the HTTP error directly (not wrapped)
	assert.Contains(t, err.Error(), "HTTP 503")

	// Should have made 3 attempts (initial + 2 retries)
	assert.Equal(t, 3, requestCount)
}

// TestClientNoRetryOn400Errors tests that 4xx errors (except 401) are not retried
func TestClientNoRetryOn400Errors(t *testing.T) {
	requestCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"bad request"}`))
	}))
	defer server.Close()

	c := client.New(
		client.WithBaseURL(server.URL),
		client.WithMaxRetries(3),
	)

	_, err := c.Get(context.Background(), "/api/bad-request")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP 400")

	// Should only make 1 request (no retries for 400)
	assert.Equal(t, 1, requestCount)
}

// TestClientCustomHeaders tests adding custom headers to requests
func TestClientCustomHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "custom-value", r.Header.Get("X-Custom-Header"))
		assert.Equal(t, "another-value", r.Header.Get("X-Another-Header"))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	c := client.New(
		client.WithBaseURL(server.URL),
	)

	_, err := c.Get(context.Background(), "/api/test",
		client.WithHeader("X-Custom-Header", "custom-value"),
		client.WithHeader("X-Another-Header", "another-value"),
	)
	require.NoError(t, err)
}

// TestClientQueryParameters tests adding query parameters to requests
func TestClientQueryParameters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		assert.Equal(t, "value1", query.Get("param1"))
		assert.Equal(t, "value2", query.Get("param2"))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	c := client.New(
		client.WithBaseURL(server.URL),
	)

	_, err := c.Get(context.Background(), "/api/test",
		client.WithQueryParam("param1", "value1"),
		client.WithQueryParam("param2", "value2"),
	)
	require.NoError(t, err)
}

// TestClientAllHTTPMethods tests all HTTP methods (GET, POST, PUT, PATCH, DELETE)
func TestClientAllHTTPMethods(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE"}

	for _, expectedMethod := range methods {
		t.Run(expectedMethod, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, expectedMethod, r.Method)

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]string{"method": expectedMethod})
			}))
			defer server.Close()

			c := client.New(
				client.WithBaseURL(server.URL),
			)

			var err error
			var resp []byte

			switch expectedMethod {
			case "GET":
				resp, err = c.Get(context.Background(), "/api/test")
			case "POST":
				resp, err = c.Post(context.Background(), "/api/test", map[string]string{"data": "test"})
			case "PUT":
				resp, err = c.Put(context.Background(), "/api/test", map[string]string{"data": "test"})
			case "PATCH":
				resp, err = c.Patch(context.Background(), "/api/test", map[string]string{"data": "test"})
			case "DELETE":
				resp, err = c.Delete(context.Background(), "/api/test")
			}

			require.NoError(t, err)

			var result map[string]interface{}
			err = json.Unmarshal(resp, &result)
			require.NoError(t, err)
			assert.Equal(t, expectedMethod, result["method"])
		})
	}
}

// TestClientTimeout tests request timeout functionality
func TestClientTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow server
		time.Sleep(2 * time.Second)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	c := client.New(
		client.WithBaseURL(server.URL),
		client.WithTimeout(100*time.Millisecond), // Very short timeout
	)

	_, err := c.Get(context.Background(), "/api/slow")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}

// TestClientContextCancellation tests that context cancellation is respected
func TestClientContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow server
		time.Sleep(2 * time.Second)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	c := client.New(
		client.WithBaseURL(server.URL),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := c.Get(ctx, "/api/slow")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}

// TestClientWithCustomHTTPClient tests using a custom HTTP client
func TestClientWithCustomHTTPClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	customHTTPClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	c := client.New(
		client.WithBaseURL(server.URL),
		client.WithHTTPClient(customHTTPClient),
	)

	resp, err := c.Get(context.Background(), "/api/test")
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(resp, &result)
	require.NoError(t, err)
	assert.Equal(t, "ok", result["status"])
}
