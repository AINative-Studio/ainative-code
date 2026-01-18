package backend

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Test 1: Client Initialization
func TestNewClient(t *testing.T) {
	// GIVEN a base URL
	baseURL := "http://localhost:8000"

	// WHEN creating a new client
	client := NewClient(baseURL)

	// THEN client should be properly initialized
	if client == nil {
		t.Fatal("expected client to be initialized, got nil")
	}
	if client.BaseURL != baseURL {
		t.Errorf("expected BaseURL %s, got %s", baseURL, client.BaseURL)
	}
	if client.Timeout == 0 {
		t.Error("expected non-zero timeout")
	}
	if client.HTTPClient == nil {
		t.Error("expected HTTPClient to be initialized, got nil")
	}
}

func TestNewClient_WithCustomTimeout(t *testing.T) {
	// GIVEN a base URL and custom timeout
	baseURL := "http://localhost:8000"
	timeout := 60 * time.Second

	// WHEN creating a new client with custom timeout
	client := NewClient(baseURL, WithTimeout(timeout))

	// THEN timeout should be set correctly
	if client.Timeout != timeout {
		t.Errorf("expected timeout %v, got %v", timeout, client.Timeout)
	}
	if client.HTTPClient.Timeout != timeout {
		t.Errorf("expected HTTPClient timeout %v, got %v", timeout, client.HTTPClient.Timeout)
	}
}

// Test 2: Login Method
func TestClient_Login_Success(t *testing.T) {
	// GIVEN a mock HTTP server returning successful login response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.URL.Path != "/api/v1/auth/login" {
			t.Errorf("expected path /api/v1/auth/login, got %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("expected POST method, got %s", r.Method)
		}

		// Verify Content-Type
		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", contentType)
		}

		// Verify request body
		var reqBody LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}
		if reqBody.Email != "test@example.com" {
			t.Errorf("expected email test@example.com, got %s", reqBody.Email)
		}
		if reqBody.Password != "password123" {
			t.Errorf("expected password password123, got %s", reqBody.Password)
		}

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token":  "eyJhbGciOiJIUzI1NiIs...",
			"refresh_token": "eyJhbGciOiJIUzI1NiIs...",
			"token_type":    "bearer",
			"user": map[string]interface{}{
				"id":    "123",
				"email": "test@example.com",
			},
		})
	}))
	defer server.Close()

	// WHEN calling Login
	client := NewClient(server.URL)
	resp, err := client.Login(context.Background(), "test@example.com", "password123")

	// THEN login should succeed
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.AccessToken == "" {
		t.Error("expected access token, got empty string")
	}
	if resp.RefreshToken == "" {
		t.Error("expected refresh token, got empty string")
	}
	if resp.TokenType != "bearer" {
		t.Errorf("expected token type bearer, got %s", resp.TokenType)
	}
	if resp.User.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", resp.User.Email)
	}
	if resp.User.ID != "123" {
		t.Errorf("expected user ID 123, got %s", resp.User.ID)
	}
}

func TestClient_Login_InvalidCredentials(t *testing.T) {
	// GIVEN a mock HTTP server returning 401 Unauthorized
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"detail": "Incorrect email or password",
		})
	}))
	defer server.Close()

	// WHEN calling Login with invalid credentials
	client := NewClient(server.URL)
	_, err := client.Login(context.Background(), "test@example.com", "wrongpassword")

	// THEN it should return an authentication error
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}

// Test 3: Register Method
func TestClient_Register_Success(t *testing.T) {
	// GIVEN a mock HTTP server returning successful registration response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.URL.Path != "/api/v1/auth/register" {
			t.Errorf("expected path /api/v1/auth/register, got %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("expected POST method, got %s", r.Method)
		}

		// Verify request body
		var reqBody RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token":  "eyJhbGciOiJIUzI1NiIs...",
			"refresh_token": "eyJhbGciOiJIUzI1NiIs...",
			"token_type":    "bearer",
			"user": map[string]interface{}{
				"id":    "456",
				"email": "newuser@example.com",
			},
		})
	}))
	defer server.Close()

	// WHEN calling Register
	client := NewClient(server.URL)
	resp, err := client.Register(context.Background(), "newuser@example.com", "password123")

	// THEN registration should succeed
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.AccessToken == "" {
		t.Error("expected access token, got empty string")
	}
	if resp.User.Email != "newuser@example.com" {
		t.Errorf("expected email newuser@example.com, got %s", resp.User.Email)
	}
}

// Test 4: Refresh Token Method
func TestClient_RefreshToken_Success(t *testing.T) {
	// GIVEN a mock HTTP server returning successful refresh response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.URL.Path != "/api/v1/auth/refresh" {
			t.Errorf("expected path /api/v1/auth/refresh, got %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("expected POST method, got %s", r.Method)
		}

		// Verify request body
		var reqBody RefreshTokenRequest
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}
		if reqBody.RefreshToken != "old_refresh_token" {
			t.Errorf("expected refresh token old_refresh_token, got %s", reqBody.RefreshToken)
		}

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token":  "new_access_token",
			"refresh_token": "new_refresh_token",
			"token_type":    "bearer",
		})
	}))
	defer server.Close()

	// WHEN calling RefreshToken
	client := NewClient(server.URL)
	resp, err := client.RefreshToken(context.Background(), "old_refresh_token")

	// THEN refresh should succeed
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.AccessToken != "new_access_token" {
		t.Errorf("expected access token new_access_token, got %s", resp.AccessToken)
	}
	if resp.RefreshToken != "new_refresh_token" {
		t.Errorf("expected refresh token new_refresh_token, got %s", resp.RefreshToken)
	}
}

// Test 5: Logout Method
func TestClient_Logout_Success(t *testing.T) {
	// GIVEN a mock HTTP server returning successful logout response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.URL.Path != "/api/v1/auth/logout" {
			t.Errorf("expected path /api/v1/auth/logout, got %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("expected POST method, got %s", r.Method)
		}

		// Verify Authorization header
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			t.Error("expected Authorization header with Bearer token")
		}

		// Return success
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// WHEN calling Logout
	client := NewClient(server.URL)
	err := client.Logout(context.Background(), "access_token")

	// THEN logout should succeed
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// Test 6: Chat Completion Method
func TestClient_ChatCompletion_Success(t *testing.T) {
	// GIVEN a mock HTTP server returning chat completion response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.URL.Path != "/api/v1/chat/completions" {
			t.Errorf("expected path /api/v1/chat/completions, got %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("expected POST method, got %s", r.Method)
		}

		// Verify Authorization header
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			t.Error("expected Authorization header with Bearer token")
		}
		if authHeader != "Bearer token123" {
			t.Errorf("expected Authorization header 'Bearer token123', got %s", authHeader)
		}

		// Verify request body
		var reqBody ChatCompletionRequest
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}
		if len(reqBody.Messages) == 0 {
			t.Error("expected at least one message")
		}
		if reqBody.Model != "claude-sonnet-4-5" {
			t.Errorf("expected model claude-sonnet-4-5, got %s", reqBody.Model)
		}

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":    "chatcmpl-123",
			"model": "claude-sonnet-4-5",
			"choices": []map[string]interface{}{
				{
					"message": map[string]string{
						"role":    "assistant",
						"content": "Hello! How can I help you?",
					},
				},
			},
		})
	}))
	defer server.Close()

	// WHEN calling ChatCompletion
	client := NewClient(server.URL)
	req := &ChatCompletionRequest{
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
		Model: "claude-sonnet-4-5",
	}
	resp, err := client.ChatCompletion(context.Background(), "token123", req)

	// THEN chat completion should succeed
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.ID != "chatcmpl-123" {
		t.Errorf("expected ID chatcmpl-123, got %s", resp.ID)
	}
	if len(resp.Choices) == 0 {
		t.Fatal("expected at least one choice")
	}
	if resp.Choices[0].Message.Content == "" {
		t.Error("expected message content, got empty string")
	}
	if resp.Choices[0].Message.Content != "Hello! How can I help you?" {
		t.Errorf("expected content 'Hello! How can I help you?', got %s", resp.Choices[0].Message.Content)
	}
}

func TestClient_ChatCompletion_InsufficientCredits(t *testing.T) {
	// GIVEN a mock HTTP server returning 402 Payment Required
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusPaymentRequired)
		json.NewEncoder(w).Encode(map[string]string{
			"detail": "Insufficient credits",
		})
	}))
	defer server.Close()

	// WHEN calling ChatCompletion without sufficient credits
	client := NewClient(server.URL)
	req := &ChatCompletionRequest{
		Messages: []Message{{Role: "user", Content: "test"}},
		Model:    "claude-sonnet-4-5",
	}
	_, err := client.ChatCompletion(context.Background(), "token123", req)

	// THEN it should return a payment required error
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrPaymentRequired) {
		t.Errorf("expected ErrPaymentRequired, got %v", err)
	}
}

func TestClient_ChatCompletion_WithOptionalParams(t *testing.T) {
	// GIVEN a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request body contains optional parameters
		var reqBody ChatCompletionRequest
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}
		if reqBody.Temperature != 0.7 {
			t.Errorf("expected temperature 0.7, got %f", reqBody.Temperature)
		}
		if reqBody.MaxTokens != 1000 {
			t.Errorf("expected max_tokens 1000, got %d", reqBody.MaxTokens)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":    "chatcmpl-123",
			"model": "claude-sonnet-4-5",
			"choices": []map[string]interface{}{
				{
					"message": map[string]string{
						"role":    "assistant",
						"content": "Response",
					},
				},
			},
		})
	}))
	defer server.Close()

	// WHEN calling ChatCompletion with optional parameters
	client := NewClient(server.URL)
	req := &ChatCompletionRequest{
		Messages:    []Message{{Role: "user", Content: "test"}},
		Model:       "claude-sonnet-4-5",
		Temperature: 0.7,
		MaxTokens:   1000,
	}
	_, err := client.ChatCompletion(context.Background(), "token123", req)

	// THEN request should succeed
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// Test 7: Error Handling
func TestClient_NetworkError(t *testing.T) {
	// GIVEN a client pointing to invalid URL
	client := NewClient("http://invalid-url-that-does-not-exist:9999")

	// WHEN making a request
	_, err := client.Login(context.Background(), "test@example.com", "password")

	// THEN it should return a network error
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_TimeoutError(t *testing.T) {
	// GIVEN a mock server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// WHEN calling with a short timeout
	client := NewClient(server.URL, WithTimeout(10*time.Millisecond))
	_, err := client.Login(context.Background(), "test@example.com", "password")

	// THEN it should return a timeout error
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}

func TestClient_ServerError(t *testing.T) {
	// GIVEN a mock HTTP server returning 500 Internal Server Error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"detail": "Internal server error",
		})
	}))
	defer server.Close()

	// WHEN making a request
	client := NewClient(server.URL)
	_, err := client.Login(context.Background(), "test@example.com", "password")

	// THEN it should return a server error
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrServerError) {
		t.Errorf("expected ErrServerError, got %v", err)
	}
}

func TestClient_BadGatewayError(t *testing.T) {
	// GIVEN a mock HTTP server returning 502 Bad Gateway
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()

	// WHEN making a request
	client := NewClient(server.URL)
	_, err := client.Login(context.Background(), "test@example.com", "password")

	// THEN it should return a server error
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrServerError) {
		t.Errorf("expected ErrServerError, got %v", err)
	}
}

func TestClient_ServiceUnavailableError(t *testing.T) {
	// GIVEN a mock HTTP server returning 503 Service Unavailable
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	// WHEN making a request
	client := NewClient(server.URL)
	_, err := client.Login(context.Background(), "test@example.com", "password")

	// THEN it should return a server error
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrServerError) {
		t.Errorf("expected ErrServerError, got %v", err)
	}
}

func TestClient_UnexpectedStatusCode(t *testing.T) {
	// GIVEN a mock HTTP server returning an unexpected status code
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot) // 418
	}))
	defer server.Close()

	// WHEN making a request
	client := NewClient(server.URL)
	_, err := client.Login(context.Background(), "test@example.com", "password")

	// THEN it should return an error
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_InvalidJSON(t *testing.T) {
	// GIVEN a mock HTTP server returning invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	// WHEN making a request
	client := NewClient(server.URL)
	_, err := client.Login(context.Background(), "test@example.com", "password")

	// THEN it should return a decode error
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// Test 8: GetMe Method
func TestClient_GetMe_Success(t *testing.T) {
	// GIVEN a mock HTTP server returning user info
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.URL.Path != "/api/v1/auth/me" {
			t.Errorf("expected path /api/v1/auth/me, got %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("expected GET method, got %s", r.Method)
		}

		// Verify Authorization header
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			t.Error("expected Authorization header with Bearer token")
		}

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":    "123",
			"email": "test@example.com",
		})
	}))
	defer server.Close()

	// WHEN calling GetMe
	client := NewClient(server.URL)
	user, err := client.GetMe(context.Background(), "access_token")

	// THEN it should succeed
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.ID != "123" {
		t.Errorf("expected user ID 123, got %s", user.ID)
	}
	if user.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", user.Email)
	}
}

// Test 9: Health Check Method
func TestClient_HealthCheck_Success(t *testing.T) {
	// GIVEN a mock HTTP server returning health status
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.URL.Path != "/health" {
			t.Errorf("expected path /health, got %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("expected GET method, got %s", r.Method)
		}

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
		})
	}))
	defer server.Close()

	// WHEN calling HealthCheck
	client := NewClient(server.URL)
	err := client.HealthCheck(context.Background())

	// THEN it should succeed
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestClient_HealthCheck_Failure(t *testing.T) {
	// GIVEN a mock HTTP server returning unhealthy status
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	// WHEN calling HealthCheck
	client := NewClient(server.URL)
	err := client.HealthCheck(context.Background())

	// THEN it should return an error
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
