package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/AINative-studio/ainative-code/internal/ratelimit"
)

// RateLimiterMiddleware provides HTTP rate limiting middleware
type RateLimiterMiddleware struct {
	limiter *ratelimit.Limiter
	metrics *ratelimit.Metrics
	config  RateLimiterConfig
}

// RateLimiterConfig contains middleware configuration
type RateLimiterConfig struct {
	// PerUser enables per-user rate limiting
	PerUser bool

	// PerEndpoint enables per-endpoint rate limiting
	PerEndpoint bool

	// UserIDExtractor extracts user ID from request
	UserIDExtractor func(*http.Request) string

	// IPExtractor extracts IP address from request
	IPExtractor func(*http.Request) string

	// KeyGenerator generates a rate limit key from request
	KeyGenerator func(*http.Request) string

	// OnRateLimitExceeded is called when rate limit is exceeded
	OnRateLimitExceeded func(http.ResponseWriter, *http.Request, *ratelimit.Result)

	// SkipPaths defines paths that should skip rate limiting
	SkipPaths []string

	// EndpointLimits defines custom limits for specific endpoints
	EndpointLimits map[string]int
}

// NewRateLimiterMiddleware creates a new rate limiter middleware
func NewRateLimiterMiddleware(limiter *ratelimit.Limiter, config RateLimiterConfig) *RateLimiterMiddleware {
	// Set defaults
	if config.IPExtractor == nil {
		config.IPExtractor = defaultIPExtractor
	}
	if config.KeyGenerator == nil {
		config.KeyGenerator = defaultKeyGenerator
	}
	if config.OnRateLimitExceeded == nil {
		config.OnRateLimitExceeded = defaultRateLimitExceededHandler
	}

	return &RateLimiterMiddleware{
		limiter: limiter,
		metrics: ratelimit.NewMetrics(),
		config:  config,
	}
}

// Handler returns an HTTP middleware handler
func (m *RateLimiterMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if path should skip rate limiting
		if m.shouldSkipPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Generate rate limit key
		key := m.config.KeyGenerator(r)

		// Check rate limit
		ctx := r.Context()
		result, err := m.limiter.Allow(ctx, key)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Set rate limit headers
		m.setRateLimitHeaders(w, result)

		// Extract user and endpoint for metrics
		userID := ""
		if m.config.UserIDExtractor != nil {
			userID = m.config.UserIDExtractor(r)
		}
		endpoint := r.URL.Path

		// Record metrics
		m.metrics.RecordRequest(result.Allowed, endpoint, userID)

		// Check if request is allowed
		if !result.Allowed {
			m.config.OnRateLimitExceeded(w, r, result)
			return
		}

		// Continue with next handler
		next.ServeHTTP(w, r)
	})
}

// Middleware returns a middleware function that can wrap http.HandlerFunc
func (m *RateLimiterMiddleware) Middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m.Handler(http.HandlerFunc(next)).ServeHTTP(w, r)
	}
}

// GetMetrics returns the current metrics
func (m *RateLimiterMiddleware) GetMetrics() ratelimit.Stats {
	return m.metrics.GetStats()
}

// ResetMetrics resets all metrics
func (m *RateLimiterMiddleware) ResetMetrics() {
	m.metrics.Reset()
}

func (m *RateLimiterMiddleware) shouldSkipPath(path string) bool {
	for _, skipPath := range m.config.SkipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

func (m *RateLimiterMiddleware) setRateLimitHeaders(w http.ResponseWriter, result *ratelimit.Result) {
	w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", result.Limit))
	w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", result.Remaining))
	if !result.ResetAt.IsZero() {
		w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", result.ResetAt.Unix()))
	}
	if !result.Allowed && result.RetryAfter > 0 {
		w.Header().Set("Retry-After", fmt.Sprintf("%d", int(result.RetryAfter.Seconds())))
	}
}

// defaultIPExtractor extracts IP address from request
func defaultIPExtractor(r *http.Request) string {
	// Check X-Forwarded-For header
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

// defaultKeyGenerator generates a rate limit key from request
func defaultKeyGenerator(r *http.Request) string {
	// Use IP address as default key
	ip := defaultIPExtractor(r)
	return fmt.Sprintf("ratelimit:ip:%s", ip)
}

// defaultRateLimitExceededHandler handles rate limit exceeded responses
func defaultRateLimitExceededHandler(w http.ResponseWriter, r *http.Request, result *ratelimit.Result) {
	w.WriteHeader(http.StatusTooManyRequests)
	fmt.Fprintf(w, "Rate limit exceeded. Retry after %d seconds.\n", int(result.RetryAfter.Seconds()))
}

// PerUserKeyGenerator generates a rate limit key per user
func PerUserKeyGenerator(userIDExtractor func(*http.Request) string) func(*http.Request) string {
	return func(r *http.Request) string {
		userID := userIDExtractor(r)
		if userID == "" {
			// Fall back to IP-based rate limiting if no user ID
			ip := defaultIPExtractor(r)
			return fmt.Sprintf("ratelimit:ip:%s", ip)
		}
		return fmt.Sprintf("ratelimit:user:%s", userID)
	}
}

// PerEndpointKeyGenerator generates a rate limit key per endpoint
func PerEndpointKeyGenerator(userIDExtractor func(*http.Request) string) func(*http.Request) string {
	return func(r *http.Request) string {
		userID := userIDExtractor(r)
		endpoint := r.URL.Path
		if userID == "" {
			ip := defaultIPExtractor(r)
			return fmt.Sprintf("ratelimit:ip:%s:endpoint:%s", ip, endpoint)
		}
		return fmt.Sprintf("ratelimit:user:%s:endpoint:%s", userID, endpoint)
	}
}

// APIKeyExtractor extracts API key from request
func APIKeyExtractor(r *http.Request) string {
	// Check Authorization header
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}

	// Check X-API-Key header
	apiKey := r.Header.Get("X-API-Key")
	if apiKey != "" {
		return apiKey
	}

	// Check query parameter
	return r.URL.Query().Get("api_key")
}

// WithContext adds rate limit result to request context
func WithContext(ctx context.Context, result *ratelimit.Result) context.Context {
	return context.WithValue(ctx, rateLimitContextKey, result)
}

// FromContext retrieves rate limit result from context
func FromContext(ctx context.Context) (*ratelimit.Result, bool) {
	result, ok := ctx.Value(rateLimitContextKey).(*ratelimit.Result)
	return result, ok
}

type contextKey string

const rateLimitContextKey contextKey = "ratelimit"
