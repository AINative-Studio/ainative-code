package backend

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is an HTTP client for communicating with the Python backend
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	Timeout    time.Duration
}

// ClientOption is a function that configures a Client
type ClientOption func(*Client)

// WithTimeout returns a ClientOption that sets the client timeout
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.Timeout = timeout
		c.HTTPClient.Timeout = timeout
	}
}

// NewClient creates a new HTTP client for the backend API
func NewClient(baseURL string, opts ...ClientOption) *Client {
	client := &Client{
		BaseURL: baseURL,
		Timeout: 30 * time.Second,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

// Login authenticates a user with email and password
func (c *Client) Login(ctx context.Context, email, password string) (*TokenResponse, error) {
	reqBody := LoginRequest{
		Email:    email,
		Password: password,
	}

	var resp TokenResponse
	err := c.doRequest(ctx, "POST", "/api/v1/auth/login", "", reqBody, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// Register creates a new user account
func (c *Client) Register(ctx context.Context, email, password string) (*TokenResponse, error) {
	reqBody := RegisterRequest{
		Email:    email,
		Password: password,
	}

	var resp TokenResponse
	err := c.doRequest(ctx, "POST", "/api/v1/auth/register", "", reqBody, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// RefreshToken refreshes an access token using a refresh token
func (c *Client) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	reqBody := RefreshTokenRequest{
		RefreshToken: refreshToken,
	}

	var resp TokenResponse
	err := c.doRequest(ctx, "POST", "/api/v1/auth/refresh", "", reqBody, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// Logout logs out the current user
func (c *Client) Logout(ctx context.Context, accessToken string) error {
	return c.doRequest(ctx, "POST", "/api/v1/auth/logout", accessToken, nil, nil)
}

// GetMe retrieves the current user's information
func (c *Client) GetMe(ctx context.Context, accessToken string) (*User, error) {
	var user User
	err := c.doRequest(ctx, "GET", "/api/v1/auth/me", accessToken, nil, &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// ChatCompletion sends a chat completion request to the backend
func (c *Client) ChatCompletion(ctx context.Context, accessToken string, req *ChatCompletionRequest) (*ChatCompletionResponse, error) {
	var resp ChatCompletionResponse
	err := c.doRequest(ctx, "POST", "/api/v1/chat/completions", accessToken, req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// HealthCheck checks if the backend is healthy
func (c *Client) HealthCheck(ctx context.Context) error {
	var resp HealthResponse
	err := c.doRequest(ctx, "GET", "/health", "", nil, &resp)
	if err != nil {
		return err
	}

	return nil
}

// doRequest is a helper method that performs HTTP requests
func (c *Client) doRequest(ctx context.Context, method, path, token string, body, result interface{}) error {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Handle error status codes
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return ErrUnauthorized
	case http.StatusPaymentRequired:
		return ErrPaymentRequired
	case http.StatusBadRequest:
		return ErrBadRequest
	case http.StatusNotFound:
		return ErrNotFound
	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
		return ErrServerError
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}
