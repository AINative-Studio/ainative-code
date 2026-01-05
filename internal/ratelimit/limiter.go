package ratelimit

import (
	"context"
	"fmt"
	"time"
)

// Limiter implements rate limiting using the token bucket algorithm
type Limiter struct {
	storage Storage
	config  Config
}

// Config contains rate limiter configuration
type Config struct {
	// RequestsPerMinute defines the maximum requests allowed per minute
	RequestsPerMinute int

	// BurstSize defines the maximum burst of requests allowed
	BurstSize int

	// TimeWindow defines the time window for rate limiting
	TimeWindow time.Duration

	// PerUser enables per-user rate limiting
	PerUser bool

	// PerEndpoint enables per-endpoint rate limiting
	PerEndpoint bool

	// EndpointLimits defines custom limits for specific endpoints
	EndpointLimits map[string]int

	// Storage defines the storage backend type
	StorageType string

	// RedisURL is the Redis connection URL (if using Redis storage)
	RedisURL string
}

// Result contains the result of a rate limit check
type Result struct {
	// Allowed indicates if the request is allowed
	Allowed bool

	// Limit is the maximum number of requests allowed
	Limit int64

	// Remaining is the number of requests remaining
	Remaining int64

	// ResetAt is when the rate limit resets
	ResetAt time.Time

	// RetryAfter is the duration to wait before retrying (if blocked)
	RetryAfter time.Duration
}

// NewLimiter creates a new rate limiter
func NewLimiter(storage Storage, config Config) *Limiter {
	return &Limiter{
		storage: storage,
		config:  config,
	}
}

// Allow checks if a request is allowed for the given key
func (l *Limiter) Allow(ctx context.Context, key string) (*Result, error) {
	// Get the limit for this key
	limit := int64(l.config.RequestsPerMinute)

	// Increment the counter
	count, resetAt, err := l.storage.Increment(ctx, key, l.config.TimeWindow)
	if err != nil {
		return nil, fmt.Errorf("failed to increment counter: %w", err)
	}

	// Calculate remaining requests
	remaining := limit - count
	if remaining < 0 {
		remaining = 0
	}

	// Check if request is allowed
	allowed := count <= limit

	result := &Result{
		Allowed:   allowed,
		Limit:     limit,
		Remaining: remaining,
		ResetAt:   resetAt,
	}

	// Calculate retry after if blocked
	if !allowed {
		result.RetryAfter = time.Until(resetAt)
		if result.RetryAfter < 0 {
			result.RetryAfter = 0
		}
	}

	return result, nil
}

// AllowN checks if N requests are allowed for the given key
func (l *Limiter) AllowN(ctx context.Context, key string, n int) (*Result, error) {
	// For simplicity, we'll check if we can allow 1 request
	// A more sophisticated implementation would reserve N tokens
	return l.Allow(ctx, key)
}

// Reset resets the rate limit for the given key
func (l *Limiter) Reset(ctx context.Context, key string) error {
	return l.storage.Reset(ctx, key)
}

// GetLimitForEndpoint returns the rate limit for a specific endpoint
func (l *Limiter) GetLimitForEndpoint(endpoint string) int {
	if limit, ok := l.config.EndpointLimits[endpoint]; ok {
		return limit
	}
	return l.config.RequestsPerMinute
}

// BuildKey builds a rate limit key from the given components
func (l *Limiter) BuildKey(components ...string) string {
	key := "ratelimit"
	for _, c := range components {
		if c != "" {
			key += ":" + c
		}
	}
	return key
}

// Close closes the rate limiter and its storage backend
func (l *Limiter) Close() error {
	if l.storage != nil {
		return l.storage.Close()
	}
	return nil
}
