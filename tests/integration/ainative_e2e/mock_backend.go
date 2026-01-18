package ainative_e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/AINative-studio/ainative-code/internal/backend"
	"github.com/AINative-studio/ainative-code/tests/integration/ainative_e2e/fixtures"
)

// MockBackend represents a mock HTTP server that simulates the Python backend
type MockBackend struct {
	*httptest.Server
	t                *testing.T
	streamingEnabled bool
	streamDelay      time.Duration
	streamChunkCount int
	failingProviders map[string]bool
	userCredits      map[string]int
	rateLimiters     map[string]*RateLimiter
	loggedOutTokens  map[string]bool
	mu               sync.Mutex
}

// NewMockBackend creates a new mock backend server for testing
func NewMockBackend(t *testing.T) *MockBackend {
	mb := &MockBackend{
		t:                t,
		streamingEnabled: false,
		streamDelay:      0,
		streamChunkCount: 10,
		failingProviders: make(map[string]bool),
		userCredits:      make(map[string]int),
		rateLimiters:     make(map[string]*RateLimiter),
		loggedOutTokens:  make(map[string]bool),
	}

	// Set default credits for test users
	mb.userCredits["test@example.com"] = 1000
	mb.userCredits["nocredit@example.com"] = 0

	mux := http.NewServeMux()

	// Auth endpoints
	mux.HandleFunc("/api/v1/auth/login", mb.handleLogin)
	mux.HandleFunc("/api/v1/auth/register", mb.handleRegister)
	mux.HandleFunc("/api/v1/auth/logout", mb.handleLogout)
	mux.HandleFunc("/api/v1/auth/refresh", mb.handleRefresh)
	mux.HandleFunc("/api/v1/auth/me", mb.handleGetMe)

	// Chat endpoint
	mux.HandleFunc("/api/v1/chat/completions", mb.handleChatCompletion)

	// Health check
	mux.HandleFunc("/health", mb.handleHealth)

	mb.Server = httptest.NewServer(mux)
	return mb
}

// handleLogin handles login requests
func (mb *MockBackend) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req backend.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Check credentials
	if req.Password != "password123" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(fixtures.GetErrorResponse("Incorrect email or password"))
		return
	}

	// Generate tokens
	resp := fixtures.GetTokenResponse(req.Email)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// handleRegister handles registration requests
func (mb *MockBackend) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req backend.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set default credits for new user
	mb.mu.Lock()
	mb.userCredits[req.Email] = 1000
	mb.mu.Unlock()

	// Generate tokens
	resp := fixtures.GetTokenResponse(req.Email)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// handleLogout handles logout requests
func (mb *MockBackend) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract token from header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(fixtures.GetErrorResponse("Authorization header required"))
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")

	// Mark token as logged out
	mb.mu.Lock()
	mb.loggedOutTokens[token] = true
	mb.mu.Unlock()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Successfully logged out"})
}

// handleRefresh handles token refresh requests
func (mb *MockBackend) handleRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req backend.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate refresh token
	email := fixtures.ExtractEmailFromToken(req.RefreshToken)
	if email == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(fixtures.GetErrorResponse("Invalid refresh token"))
		return
	}

	// Generate new tokens
	resp := fixtures.GetTokenResponse(email)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// handleGetMe handles get user info requests
func (mb *MockBackend) handleGetMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Verify auth
	email := mb.extractEmailFromAuth(r)
	if email == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(fixtures.GetErrorResponse("Unauthorized"))
		return
	}

	user := fixtures.GetDefaultUser(email)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

// handleChatCompletion handles chat completion requests
func (mb *MockBackend) handleChatCompletion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Verify auth
	email := mb.extractEmailFromAuth(r)
	if email == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(fixtures.GetErrorResponse("Unauthorized"))
		return
	}

	// Check if token is logged out
	authHeader := r.Header.Get("Authorization")
	token := strings.TrimPrefix(authHeader, "Bearer ")
	mb.mu.Lock()
	if mb.loggedOutTokens[token] {
		mb.mu.Unlock()
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(fixtures.GetErrorResponse("Token has been revoked"))
		return
	}
	mb.mu.Unlock()

	// Check rate limit
	if mb.isRateLimited(email) {
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(fixtures.GetErrorResponse("Rate limit exceeded"))
		return
	}

	// Check credits
	mb.mu.Lock()
	credits := mb.userCredits[email]
	mb.mu.Unlock()

	if credits <= 0 {
		w.WriteHeader(http.StatusPaymentRequired)
		json.NewEncoder(w).Encode(fixtures.GetErrorResponse("Insufficient credits"))
		return
	}

	var req backend.ChatCompletionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Handle streaming
	if req.Stream && mb.streamingEnabled {
		mb.handleStreamingChat(w, r, req)
		return
	}

	// Non-streaming response
	resp := fixtures.GetDefaultChatResponse()
	resp.Model = req.Model

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// handleStreamingChat handles streaming chat responses
func (mb *MockBackend) handleStreamingChat(w http.ResponseWriter, r *http.Request, req backend.ChatCompletionRequest) {
	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	chunks := fixtures.GetStreamingChatChunks()
	if mb.streamChunkCount > 0 {
		chunks = fixtures.GetLargeStreamingChunks(mb.streamChunkCount)
	}

	for _, chunk := range chunks {
		// Check if request context is done
		select {
		case <-r.Context().Done():
			return
		default:
		}

		// Apply delay if set
		if mb.streamDelay > 0 {
			time.Sleep(mb.streamDelay)
		}

		// Send chunk
		chunkData := map[string]interface{}{
			"id":    "chatcmpl-stream-123",
			"model": req.Model,
			"choices": []map[string]interface{}{
				{
					"index": 0,
					"delta": map[string]string{
						"content": chunk,
					},
				},
			},
		}

		data, _ := json.Marshal(chunkData)
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()
	}

	// Send done signal
	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()
}

// handleHealth handles health check requests
func (mb *MockBackend) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(fixtures.GetHealthResponse())
}

// extractEmailFromAuth extracts email from authorization header
func (mb *MockBackend) extractEmailFromAuth(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	return fixtures.ExtractEmailFromToken(token)
}

// isRateLimited checks if the user has exceeded rate limits
func (mb *MockBackend) isRateLimited(identifier string) bool {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	// Check if rate limiting is enabled
	config, hasConfig := mb.rateLimiters["_config"]
	if !hasConfig {
		return false
	}

	// Get or create per-user rate limiter
	limiter, exists := mb.rateLimiters[identifier]
	if !exists {
		// Create new rate limiter for this user with same config
		limiter = NewRateLimiter(config.maxTokens, config.window)
		mb.rateLimiters[identifier] = limiter
	}

	return !limiter.Allow()
}

// Configuration methods

// EnableStreaming enables streaming support
func (mb *MockBackend) EnableStreaming() {
	mb.mu.Lock()
	defer mb.mu.Unlock()
	mb.streamingEnabled = true
}

// SetStreamDelay sets the delay between streaming chunks
func (mb *MockBackend) SetStreamDelay(delay time.Duration) {
	mb.mu.Lock()
	defer mb.mu.Unlock()
	mb.streamDelay = delay
}

// SetStreamChunkCount sets the number of chunks for streaming
func (mb *MockBackend) SetStreamChunkCount(count int) {
	mb.mu.Lock()
	defer mb.mu.Unlock()
	mb.streamChunkCount = count
}

// SetUserCredits sets the credits for a user
func (mb *MockBackend) SetUserCredits(email string, credits int) {
	mb.mu.Lock()
	defer mb.mu.Unlock()
	mb.userCredits[email] = credits
}

// SetPrimaryProviderFailing marks a provider as failing
func (mb *MockBackend) SetPrimaryProviderFailing(provider string) {
	mb.mu.Lock()
	defer mb.mu.Unlock()
	mb.failingProviders[provider] = true
}

// EnableRateLimit enables rate limiting with specified requests per window
func (mb *MockBackend) EnableRateLimit(requests int, window time.Duration) {
	mb.mu.Lock()
	defer mb.mu.Unlock()
	// Create rate limiter for all users - we'll use email as key
	// Store it with a special key that we'll use to initialize per-user limiters
	mb.rateLimiters["_config"] = NewRateLimiter(requests, window)
	mb.rateLimiters["_max_tokens"] = &RateLimiter{maxTokens: requests, window: window}
}

// RateLimiter implements a simple token bucket rate limiter
type RateLimiter struct {
	tokens    int
	maxTokens int
	refillAt  time.Time
	window    time.Duration
	mu        sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(maxTokens int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		tokens:    maxTokens,
		maxTokens: maxTokens,
		refillAt:  time.Now().Add(window),
		window:    window,
	}
}

// Allow checks if a request is allowed under the rate limit
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// Refill tokens if window has passed
	if now.After(rl.refillAt) {
		rl.tokens = rl.maxTokens
		rl.refillAt = now.Add(rl.window)
	}

	// Check if tokens available
	if rl.tokens > 0 {
		rl.tokens--
		return true
	}

	return false
}
