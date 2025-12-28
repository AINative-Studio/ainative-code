package provider

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// mockLogger implements logger.Logger for testing
type mockLogger struct {
	debugMessages []string
	infoMessages  []string
	warnMessages  []string
	errorMessages []string
}

func (m *mockLogger) Debug(msg string) {
	m.debugMessages = append(m.debugMessages, msg)
}

func (m *mockLogger) Info(msg string) {
	m.infoMessages = append(m.infoMessages, msg)
}

func (m *mockLogger) Warn(msg string) {
	m.warnMessages = append(m.warnMessages, msg)
}

func (m *mockLogger) Error(msg string) {
	m.errorMessages = append(m.errorMessages, msg)
}

func TestNewBaseProvider(t *testing.T) {
	tests := []struct {
		name           string
		config         BaseProviderConfig
		expectDefaults bool
	}{
		{
			name: "with custom config",
			config: BaseProviderConfig{
				Name: "test-provider",
				HTTPClient: &http.Client{
					Timeout: 30 * time.Second,
				},
				Logger: &mockLogger{},
				RetryConfig: RetryConfig{
					MaxRetries:     5,
					InitialBackoff: 2 * time.Second,
					MaxBackoff:     60 * time.Second,
					Multiplier:     3.0,
					RetryableStatusCodes: []int{429, 500},
				},
			},
			expectDefaults: false,
		},
		{
			name: "with defaults",
			config: BaseProviderConfig{
				Name: "test-provider",
			},
			expectDefaults: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewBaseProvider(tt.config)

			if provider == nil {
				t.Fatal("expected non-nil provider")
			}

			if provider.Name() != tt.config.Name {
				t.Errorf("expected name %q, got %q", tt.config.Name, provider.Name())
			}

			if provider.httpClient == nil {
				t.Error("expected non-nil HTTP client")
			}

			if tt.expectDefaults {
				if provider.httpClient.Timeout != 60*time.Second {
					t.Errorf("expected default timeout of 60s, got %v", provider.httpClient.Timeout)
				}

				if provider.retryConfig.MaxRetries != 3 {
					t.Errorf("expected default max retries of 3, got %d", provider.retryConfig.MaxRetries)
				}
			} else {
				if provider.retryConfig.MaxRetries != tt.config.RetryConfig.MaxRetries {
					t.Errorf("expected max retries %d, got %d", tt.config.RetryConfig.MaxRetries, provider.retryConfig.MaxRetries)
				}
			}
		})
	}
}

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()

	if config.MaxRetries != 3 {
		t.Errorf("expected MaxRetries=3, got %d", config.MaxRetries)
	}

	if config.InitialBackoff != 1*time.Second {
		t.Errorf("expected InitialBackoff=1s, got %v", config.InitialBackoff)
	}

	if config.MaxBackoff != 30*time.Second {
		t.Errorf("expected MaxBackoff=30s, got %v", config.MaxBackoff)
	}

	if config.Multiplier != 2.0 {
		t.Errorf("expected Multiplier=2.0, got %f", config.Multiplier)
	}

	expectedCodes := []int{429, 500, 502, 503, 504}
	if len(config.RetryableStatusCodes) != len(expectedCodes) {
		t.Errorf("expected %d retryable codes, got %d", len(expectedCodes), len(config.RetryableStatusCodes))
	}

	for i, code := range expectedCodes {
		if config.RetryableStatusCodes[i] != code {
			t.Errorf("expected code %d at index %d, got %d", code, i, config.RetryableStatusCodes[i])
		}
	}
}

func TestDoRequest_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))
	defer server.Close()

	provider := NewBaseProvider(BaseProviderConfig{
		Name: "test",
	})

	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	ctx := context.Background()
	resp, err := provider.DoRequest(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "success" {
		t.Errorf("expected body 'success', got %q", string(body))
	}
}

func TestDoRequest_RetryOnRetryableStatus(t *testing.T) {
	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		if attemptCount < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))
	defer server.Close()

	logger := &mockLogger{}
	provider := NewBaseProvider(BaseProviderConfig{
		Name:   "test",
		Logger: logger,
		RetryConfig: RetryConfig{
			MaxRetries:           3,
			InitialBackoff:       10 * time.Millisecond,
			MaxBackoff:           100 * time.Millisecond,
			Multiplier:           2.0,
			RetryableStatusCodes: []int{503},
		},
	})

	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	ctx := context.Background()
	resp, err := provider.DoRequest(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer resp.Body.Close()

	if attemptCount != 3 {
		t.Errorf("expected 3 attempts, got %d", attemptCount)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	// Check that retry logging occurred
	if len(logger.debugMessages) == 0 {
		t.Error("expected debug messages for retries")
	}
}

func TestDoRequest_ExhaustedRetries(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	logger := &mockLogger{}
	provider := NewBaseProvider(BaseProviderConfig{
		Name:   "test",
		Logger: logger,
		RetryConfig: RetryConfig{
			MaxRetries:           2,
			InitialBackoff:       1 * time.Millisecond,
			MaxBackoff:           10 * time.Millisecond,
			Multiplier:           2.0,
			RetryableStatusCodes: []int{503},
		},
	})

	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	ctx := context.Background()
	resp, err := provider.DoRequest(ctx, req)

	if err == nil {
		t.Fatal("expected error after exhausting retries")
	}

	if resp != nil {
		t.Error("expected nil response after exhausting retries")
	}

	if !strings.Contains(err.Error(), "request failed after") {
		t.Errorf("expected error message about failed retries, got %v", err)
	}

	// Check error logging
	if len(logger.errorMessages) == 0 {
		t.Error("expected error messages")
	}
}

func TestDoRequest_RateLimitError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "60")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	logger := &mockLogger{}
	provider := NewBaseProvider(BaseProviderConfig{
		Name:   "test",
		Logger: logger,
		RetryConfig: RetryConfig{
			MaxRetries:           1,
			InitialBackoff:       1 * time.Millisecond,
			MaxBackoff:           10 * time.Millisecond,
			Multiplier:           2.0,
			RetryableStatusCodes: []int{429},
		},
	})

	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	ctx := context.Background()
	_, err = provider.DoRequest(ctx, req)

	if err == nil {
		t.Fatal("expected rate limit error")
	}

	var rateLimitErr *RateLimitError
	if !errors.As(err, &rateLimitErr) {
		t.Errorf("expected RateLimitError, got %T: %v", err, err)
	}

	// Check warning logging for rate limit
	if len(logger.warnMessages) == 0 {
		t.Error("expected warning messages for rate limit")
	}
}

func TestDoRequest_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	provider := NewBaseProvider(BaseProviderConfig{
		Name: "test",
	})

	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err = provider.DoRequest(ctx, req)

	if err == nil {
		t.Fatal("expected error due to context cancellation")
	}

	if !strings.Contains(err.Error(), "request cancelled") && !strings.Contains(err.Error(), "context deadline exceeded") {
		t.Errorf("expected cancellation error, got %v", err)
	}
}

func TestDoRequest_ContextCancelledDuringBackoff(t *testing.T) {
	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	provider := NewBaseProvider(BaseProviderConfig{
		Name: "test",
		RetryConfig: RetryConfig{
			MaxRetries:           5,
			InitialBackoff:       500 * time.Millisecond,
			MaxBackoff:           5 * time.Second,
			Multiplier:           2.0,
			RetryableStatusCodes: []int{503},
		},
	})

	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err = provider.DoRequest(ctx, req)

	if err == nil {
		t.Fatal("expected error due to context cancellation during backoff")
	}

	if !strings.Contains(err.Error(), "cancelled during backoff") && !strings.Contains(err.Error(), "context deadline exceeded") {
		t.Errorf("expected backoff cancellation error, got %v", err)
	}

	// Should have attempted at least once before cancellation
	if attemptCount == 0 {
		t.Error("expected at least one attempt before cancellation")
	}
}

func TestShouldRetry(t *testing.T) {
	provider := NewBaseProvider(BaseProviderConfig{
		Name: "test",
		RetryConfig: RetryConfig{
			RetryableStatusCodes: []int{429, 500, 502, 503, 504},
		},
	})

	tests := []struct {
		statusCode    int
		shouldRetry   bool
	}{
		{200, false},
		{400, false},
		{401, false},
		{404, false},
		{429, true},
		{500, true},
		{502, true},
		{503, true},
		{504, true},
	}

	for _, tt := range tests {
		t.Run(http.StatusText(tt.statusCode), func(t *testing.T) {
			result := provider.shouldRetry(tt.statusCode)
			if result != tt.shouldRetry {
				t.Errorf("status %d: expected shouldRetry=%v, got %v", tt.statusCode, tt.shouldRetry, result)
			}
		})
	}
}

func TestCalculateBackoff(t *testing.T) {
	provider := NewBaseProvider(BaseProviderConfig{
		Name: "test",
		RetryConfig: RetryConfig{
			InitialBackoff: 1 * time.Second,
			MaxBackoff:     30 * time.Second,
			Multiplier:     2.0,
		},
	})

	tests := []struct {
		attempt     int
		minExpected time.Duration
		maxExpected time.Duration
	}{
		{0, 900 * time.Millisecond, 1100 * time.Millisecond},   // 1s ± 10%
		{1, 1800 * time.Millisecond, 2200 * time.Millisecond},  // 2s ± 10%
		{2, 3600 * time.Millisecond, 4400 * time.Millisecond},  // 4s ± 10%
		{3, 7200 * time.Millisecond, 8800 * time.Millisecond},  // 8s ± 10%
		{4, 14400 * time.Millisecond, 17600 * time.Millisecond}, // 16s ± 10%
		{5, 27000 * time.Millisecond, 33000 * time.Millisecond}, // 30s (capped) ± 10%
		{10, 27000 * time.Millisecond, 33000 * time.Millisecond}, // 30s (capped) ± 10%
	}

	for _, tt := range tests {
		t.Run(string(rune('0'+tt.attempt)), func(t *testing.T) {
			backoff := provider.calculateBackoff(tt.attempt)

			if backoff < tt.minExpected || backoff > tt.maxExpected {
				t.Errorf("attempt %d: expected backoff between %v and %v, got %v",
					tt.attempt, tt.minExpected, tt.maxExpected, backoff)
			}
		})
	}
}

func TestParseRetryAfter(t *testing.T) {
	tests := []struct {
		name          string
		headerValue   string
		expectedMin   int
		expectedMax   int
	}{
		{
			name:        "integer seconds",
			headerValue: "60",
			expectedMin: 60,
			expectedMax: 60,
		},
		{
			name:        "http date",
			headerValue: time.Now().Add(120 * time.Second).UTC().Format(http.TimeFormat),
			expectedMin: 119, // Allow some tolerance for processing time
			expectedMax: 121,
		},
		{
			name:        "empty header",
			headerValue: "",
			expectedMin: 0,
			expectedMax: 0,
		},
		{
			name:        "invalid format",
			headerValue: "invalid",
			expectedMin: 0,
			expectedMax: 0,
		},
	}

	provider := NewBaseProvider(BaseProviderConfig{Name: "test"})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{
				Header: http.Header{},
			}
			if tt.headerValue != "" {
				resp.Header.Set("Retry-After", tt.headerValue)
			}

			result := provider.parseRetryAfter(resp)

			if result < tt.expectedMin || result > tt.expectedMax {
				t.Errorf("expected result between %d and %d, got %d", tt.expectedMin, tt.expectedMax, result)
			}
		})
	}
}

func TestHandleHTTPError(t *testing.T) {
	provider := NewBaseProvider(BaseProviderConfig{Name: "test"})

	tests := []struct {
		name          string
		statusCode    int
		body          string
		retryAfter    string
		expectedError string
		errorType     interface{}
	}{
		{
			name:          "unauthorized",
			statusCode:    401,
			body:          "Invalid API key",
			expectedError: "authentication failed",
			errorType:     &AuthenticationError{},
		},
		{
			name:          "forbidden",
			statusCode:    403,
			body:          "Access denied",
			expectedError: "authentication failed",
			errorType:     &AuthenticationError{},
		},
		{
			name:          "rate limit with retry-after",
			statusCode:    429,
			body:          "Rate limited",
			retryAfter:    "60",
			expectedError: "rate limit exceeded",
			errorType:     &RateLimitError{},
		},
		{
			name:          "bad request",
			statusCode:    400,
			body:          "Invalid request",
			expectedError: "bad request",
			errorType:     &ProviderError{},
		},
		{
			name:          "internal server error",
			statusCode:    500,
			body:          "Server error",
			expectedError: "HTTP 500",
			errorType:     &ProviderError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{
				StatusCode: tt.statusCode,
				Header:     http.Header{},
			}
			if tt.retryAfter != "" {
				resp.Header.Set("Retry-After", tt.retryAfter)
			}

			err := provider.HandleHTTPError(resp, []byte(tt.body))

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if !strings.Contains(err.Error(), tt.expectedError) {
				t.Errorf("expected error to contain %q, got %q", tt.expectedError, err.Error())
			}

			// Check error type using type switch and errors.As
			switch tt.errorType.(type) {
			case *AuthenticationError:
				var authErr *AuthenticationError
				if !errors.As(err, &authErr) {
					t.Errorf("expected error type *AuthenticationError, got %T", err)
				}
			case *RateLimitError:
				var rateLimitErr *RateLimitError
				if !errors.As(err, &rateLimitErr) {
					t.Errorf("expected error type *RateLimitError, got %T", err)
				}
			case *ProviderError:
				var provErr *ProviderError
				if !errors.As(err, &provErr) {
					t.Errorf("expected error type *ProviderError, got %T", err)
				}
			default:
				t.Errorf("unexpected error type in test table: %T", tt.errorType)
			}
		})
	}
}

func TestValidateModel(t *testing.T) {
	provider := NewBaseProvider(BaseProviderConfig{Name: "test"})
	supportedModels := []string{"gpt-4", "gpt-3.5-turbo", "claude-3-opus"}

	tests := []struct {
		name        string
		model       string
		expectError bool
	}{
		{
			name:        "valid model",
			model:       "gpt-4",
			expectError: false,
		},
		{
			name:        "another valid model",
			model:       "claude-3-opus",
			expectError: false,
		},
		{
			name:        "invalid model",
			model:       "invalid-model",
			expectError: true,
		},
		{
			name:        "empty model",
			model:       "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := provider.ValidateModel(tt.model, supportedModels)

			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("expected no error, got %v", err)
			}

			if tt.expectError {
				var invalidModelErr *InvalidModelError
				if !errors.As(err, &invalidModelErr) {
					t.Errorf("expected InvalidModelError, got %T", err)
				}
			}
		})
	}
}

func TestClose(t *testing.T) {
	logger := &mockLogger{}
	provider := NewBaseProvider(BaseProviderConfig{
		Name:   "test",
		Logger: logger,
	})

	err := provider.Close()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Check that close was logged
	if len(logger.debugMessages) == 0 {
		t.Error("expected debug message for close")
	}
}

func TestLogRequest(t *testing.T) {
	logger := &mockLogger{}
	provider := NewBaseProvider(BaseProviderConfig{
		Name:   "test",
		Logger: logger,
	})

	req, _ := http.NewRequest(http.MethodGet, "https://api.example.com/v1/chat", nil)
	provider.LogRequest(req)

	if len(logger.debugMessages) == 0 {
		t.Error("expected debug message for request")
	}

	if !strings.Contains(logger.debugMessages[0], "GET") {
		t.Error("expected log message to contain HTTP method")
	}
}

func TestLogResponse(t *testing.T) {
	logger := &mockLogger{}
	provider := NewBaseProvider(BaseProviderConfig{
		Name:   "test",
		Logger: logger,
	})

	req, _ := http.NewRequest(http.MethodGet, "https://api.example.com/v1/chat", nil)
	resp := &http.Response{
		StatusCode: 200,
		Request:    req,
	}

	provider.LogResponse(resp)

	if len(logger.debugMessages) == 0 {
		t.Error("expected debug message for response")
	}

	if !strings.Contains(logger.debugMessages[0], "200") {
		t.Error("expected log message to contain status code")
	}
}

func TestLogRequest_NoLogger(t *testing.T) {
	provider := NewBaseProvider(BaseProviderConfig{
		Name: "test",
	})

	req, _ := http.NewRequest(http.MethodGet, "https://api.example.com/v1/chat", nil)

	// Should not panic
	provider.LogRequest(req)
}

func TestLogResponse_NoLogger(t *testing.T) {
	provider := NewBaseProvider(BaseProviderConfig{
		Name: "test",
	})

	req, _ := http.NewRequest(http.MethodGet, "https://api.example.com/v1/chat", nil)
	resp := &http.Response{
		StatusCode: 200,
		Request:    req,
	}

	// Should not panic
	provider.LogResponse(resp)
}
