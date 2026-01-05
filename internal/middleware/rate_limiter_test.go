package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/AINative-studio/ainative-code/internal/ratelimit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRateLimiterMiddleware_AllowedRequest(t *testing.T) {
	// Given a rate limiter middleware
	storage := ratelimit.NewMemoryStorage()
	defer storage.Close()

	config := ratelimit.Config{
		RequestsPerMinute: 10,
		TimeWindow:        1 * time.Minute,
	}

	limiter := ratelimit.NewLimiter(storage, config)
	middleware := NewRateLimiterMiddleware(limiter, RateLimiterConfig{})

	handler := middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))

	// When making an allowed request
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Then it should succeed
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "success", rec.Body.String())
	assert.Equal(t, "10", rec.Header().Get("X-RateLimit-Limit"))
	assert.Equal(t, "9", rec.Header().Get("X-RateLimit-Remaining"))
	assert.NotEmpty(t, rec.Header().Get("X-RateLimit-Reset"))
}

func TestRateLimiterMiddleware_BlockedRequest(t *testing.T) {
	// Given a rate limiter middleware with low limit
	storage := ratelimit.NewMemoryStorage()
	defer storage.Close()

	config := ratelimit.Config{
		RequestsPerMinute: 2,
		TimeWindow:        1 * time.Minute,
	}

	limiter := ratelimit.NewLimiter(storage, config)
	middleware := NewRateLimiterMiddleware(limiter, RateLimiterConfig{})

	handler := middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// When exhausting the rate limit
	req1 := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req1.RemoteAddr = "192.168.1.1:12345"
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)

	req2 := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req2.RemoteAddr = "192.168.1.1:12345"
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	req3 := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req3.RemoteAddr = "192.168.1.1:12345"
	rec3 := httptest.NewRecorder()
	handler.ServeHTTP(rec3, req3)

	// Then the third request should be blocked
	assert.Equal(t, http.StatusOK, rec1.Code)
	assert.Equal(t, http.StatusOK, rec2.Code)
	assert.Equal(t, http.StatusTooManyRequests, rec3.Code)
	assert.Equal(t, "0", rec3.Header().Get("X-RateLimit-Remaining"))
	assert.NotEmpty(t, rec3.Header().Get("Retry-After"))
}

func TestRateLimiterMiddleware_PerUserRateLimit(t *testing.T) {
	// Given a rate limiter middleware with per-user limits
	storage := ratelimit.NewMemoryStorage()
	defer storage.Close()

	config := ratelimit.Config{
		RequestsPerMinute: 2,
		TimeWindow:        1 * time.Minute,
		PerUser:           true,
	}

	limiter := ratelimit.NewLimiter(storage, config)

	userExtractor := func(r *http.Request) string {
		return r.Header.Get("X-User-ID")
	}

	middleware := NewRateLimiterMiddleware(limiter, RateLimiterConfig{
		PerUser:         true,
		UserIDExtractor: userExtractor,
		KeyGenerator:    PerUserKeyGenerator(userExtractor),
	})

	handler := middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// When making requests from different users
	req1 := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req1.Header.Set("X-User-ID", "user1")
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)

	req2 := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req2.Header.Set("X-User-ID", "user2")
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	req3 := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req3.Header.Set("X-User-ID", "user1")
	rec3 := httptest.NewRecorder()
	handler.ServeHTTP(rec3, req3)

	// Then each user should have their own limit
	assert.Equal(t, http.StatusOK, rec1.Code)
	assert.Equal(t, http.StatusOK, rec2.Code)
	assert.Equal(t, http.StatusOK, rec3.Code)
	assert.Equal(t, "1", rec1.Header().Get("X-RateLimit-Remaining"))
	assert.Equal(t, "1", rec2.Header().Get("X-RateLimit-Remaining"))
	assert.Equal(t, "0", rec3.Header().Get("X-RateLimit-Remaining"))
}

func TestRateLimiterMiddleware_SkipPaths(t *testing.T) {
	// Given a rate limiter middleware with skip paths
	storage := ratelimit.NewMemoryStorage()
	defer storage.Close()

	config := ratelimit.Config{
		RequestsPerMinute: 1,
		TimeWindow:        1 * time.Minute,
	}

	limiter := ratelimit.NewLimiter(storage, config)
	middleware := NewRateLimiterMiddleware(limiter, RateLimiterConfig{
		SkipPaths: []string{"/health", "/metrics"},
	})

	handler := middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// When making requests to skipped paths
	req1 := httptest.NewRequest(http.MethodGet, "/health", nil)
	req1.RemoteAddr = "192.168.1.1:12345"
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)

	req2 := httptest.NewRequest(http.MethodGet, "/health", nil)
	req2.RemoteAddr = "192.168.1.1:12345"
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	// Then they should not be rate limited
	assert.Equal(t, http.StatusOK, rec1.Code)
	assert.Equal(t, http.StatusOK, rec2.Code)
	assert.Empty(t, rec1.Header().Get("X-RateLimit-Limit"))
	assert.Empty(t, rec2.Header().Get("X-RateLimit-Limit"))
}

func TestRateLimiterMiddleware_CustomRateLimitHandler(t *testing.T) {
	// Given a rate limiter middleware with custom handler
	storage := ratelimit.NewMemoryStorage()
	defer storage.Close()

	config := ratelimit.Config{
		RequestsPerMinute: 1,
		TimeWindow:        1 * time.Minute,
	}

	limiter := ratelimit.NewLimiter(storage, config)

	customHandler := func(w http.ResponseWriter, r *http.Request, result *ratelimit.Result) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte("custom rate limit message"))
	}

	middleware := NewRateLimiterMiddleware(limiter, RateLimiterConfig{
		OnRateLimitExceeded: customHandler,
	})

	handler := middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// When exceeding rate limit
	req1 := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req1.RemoteAddr = "192.168.1.1:12345"
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)

	req2 := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req2.RemoteAddr = "192.168.1.1:12345"
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	// Then custom handler should be called
	assert.Equal(t, http.StatusOK, rec1.Code)
	assert.Equal(t, http.StatusTooManyRequests, rec2.Code)
	assert.Equal(t, "custom rate limit message", rec2.Body.String())
}

func TestRateLimiterMiddleware_Metrics(t *testing.T) {
	// Given a rate limiter middleware
	storage := ratelimit.NewMemoryStorage()
	defer storage.Close()

	config := ratelimit.Config{
		RequestsPerMinute: 2,
		TimeWindow:        1 * time.Minute,
	}

	limiter := ratelimit.NewLimiter(storage, config)
	middleware := NewRateLimiterMiddleware(limiter, RateLimiterConfig{})

	handler := middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// When making multiple requests
	req1 := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req1.RemoteAddr = "192.168.1.1:12345"
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)

	req2 := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req2.RemoteAddr = "192.168.1.1:12345"
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	req3 := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req3.RemoteAddr = "192.168.1.1:12345"
	rec3 := httptest.NewRecorder()
	handler.ServeHTTP(rec3, req3)

	// Then metrics should be recorded
	stats := middleware.GetMetrics()
	assert.Equal(t, uint64(3), stats.TotalRequests)
	assert.Equal(t, uint64(2), stats.AllowedRequests)
	assert.Equal(t, uint64(1), stats.BlockedRequests)
}

func TestRateLimiterMiddleware_IPExtraction(t *testing.T) {
	tests := []struct {
		name           string
		setupRequest   func(*http.Request)
		expectedIP     string
	}{
		{
			name: "from X-Forwarded-For",
			setupRequest: func(r *http.Request) {
				r.Header.Set("X-Forwarded-For", "203.0.113.1, 192.168.1.1")
			},
			expectedIP: "203.0.113.1",
		},
		{
			name: "from X-Real-IP",
			setupRequest: func(r *http.Request) {
				r.Header.Set("X-Real-IP", "203.0.113.2")
			},
			expectedIP: "203.0.113.2",
		},
		{
			name: "from RemoteAddr",
			setupRequest: func(r *http.Request) {
				r.RemoteAddr = "203.0.113.3:54321"
			},
			expectedIP: "203.0.113.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
			tt.setupRequest(req)

			ip := defaultIPExtractor(req)
			assert.Equal(t, tt.expectedIP, ip)
		})
	}
}

func TestRateLimiterMiddleware_APIKeyExtractor(t *testing.T) {
	tests := []struct {
		name          string
		setupRequest  func(*http.Request)
		expectedKey   string
	}{
		{
			name: "from Authorization Bearer",
			setupRequest: func(r *http.Request) {
				r.Header.Set("Authorization", "Bearer abc123")
			},
			expectedKey: "abc123",
		},
		{
			name: "from X-API-Key",
			setupRequest: func(r *http.Request) {
				r.Header.Set("X-API-Key", "xyz789")
			},
			expectedKey: "xyz789",
		},
		{
			name: "from query parameter",
			setupRequest: func(r *http.Request) {
				r.URL.RawQuery = "api_key=query123"
			},
			expectedKey: "query123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
			tt.setupRequest(req)

			key := APIKeyExtractor(req)
			assert.Equal(t, tt.expectedKey, key)
		})
	}
}

func TestRateLimiterMiddleware_ConcurrentRequests(t *testing.T) {
	// Given a rate limiter middleware
	storage := ratelimit.NewMemoryStorage()
	defer storage.Close()

	config := ratelimit.Config{
		RequestsPerMinute: 100,
		TimeWindow:        1 * time.Minute,
	}

	limiter := ratelimit.NewLimiter(storage, config)
	middleware := NewRateLimiterMiddleware(limiter, RateLimiterConfig{})

	handler := middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// When making concurrent requests
	numRequests := 50
	var wg sync.WaitGroup
	wg.Add(numRequests)

	allowedCount := 0
	blockedCount := 0
	var mu sync.Mutex

	for i := 0; i < numRequests; i++ {
		go func() {
			defer wg.Done()

			req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
			req.RemoteAddr = "192.168.1.1:12345"
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			mu.Lock()
			if rec.Code == http.StatusOK {
				allowedCount++
			} else {
				blockedCount++
			}
			mu.Unlock()
		}()
	}

	wg.Wait()

	// Then requests should be properly limited
	assert.Equal(t, numRequests, allowedCount+blockedCount)
	assert.True(t, allowedCount <= 100)
}

func TestRateLimiterMiddleware_Context(t *testing.T) {
	// Given a context with rate limit result
	result := &ratelimit.Result{
		Allowed:   true,
		Limit:     100,
		Remaining: 99,
	}

	ctx := WithContext(context.Background(), result)

	// When retrieving from context
	retrieved, ok := FromContext(ctx)

	// Then it should be available
	require.True(t, ok)
	assert.Equal(t, result.Allowed, retrieved.Allowed)
	assert.Equal(t, result.Limit, retrieved.Limit)
	assert.Equal(t, result.Remaining, retrieved.Remaining)
}

func TestRateLimiterMiddleware_ContextNotFound(t *testing.T) {
	// Given a context without rate limit result
	ctx := context.Background()

	// When retrieving from context
	_, ok := FromContext(ctx)

	// Then it should not be found
	assert.False(t, ok)
}

func TestPerEndpointKeyGenerator(t *testing.T) {
	userExtractor := func(r *http.Request) string {
		return r.Header.Get("X-User-ID")
	}

	keyGen := PerEndpointKeyGenerator(userExtractor)

	tests := []struct {
		name     string
		userID   string
		path     string
		expected string
	}{
		{
			name:     "with user ID",
			userID:   "user123",
			path:     "/api/test",
			expected: "ratelimit:user:user123:endpoint:/api/test",
		},
		{
			name:     "without user ID",
			userID:   "",
			path:     "/api/test",
			expected: "ratelimit:ip:192.168.1.1:endpoint:/api/test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			req.RemoteAddr = "192.168.1.1:12345"
			if tt.userID != "" {
				req.Header.Set("X-User-ID", tt.userID)
			}

			key := keyGen(req)
			assert.Equal(t, tt.expected, key)
		})
	}
}

func BenchmarkRateLimiterMiddleware_Handler(b *testing.B) {
	storage := ratelimit.NewMemoryStorage()
	defer storage.Close()

	config := ratelimit.Config{
		RequestsPerMinute: 1000000,
		TimeWindow:        1 * time.Minute,
	}

	limiter := ratelimit.NewLimiter(storage, config)
	middleware := NewRateLimiterMiddleware(limiter, RateLimiterConfig{})

	handler := middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}
