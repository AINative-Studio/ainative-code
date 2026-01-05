package ratelimit

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryStorage_Basic(t *testing.T) {
	storage := NewMemoryStorage()
	defer storage.Close()

	ctx := context.Background()
	window := 1 * time.Minute

	count, resetAt, err := storage.Increment(ctx, "test", window)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
	assert.False(t, resetAt.IsZero())
}

func TestLimiter_Basic(t *testing.T) {
	storage := NewMemoryStorage()
	defer storage.Close()

	config := Config{
		RequestsPerMinute: 10,
		TimeWindow:        1 * time.Minute,
	}

	limiter := NewLimiter(storage, config)
	ctx := context.Background()

	result, err := limiter.Allow(ctx, "test-key")
	require.NoError(t, err)
	assert.True(t, result.Allowed)
	assert.Equal(t, int64(10), result.Limit)
}

func TestMetrics_Basic(t *testing.T) {
	metrics := NewMetrics()
	metrics.RecordRequest(true, "/api/test", "user1")

	stats := metrics.GetStats()
	assert.Equal(t, uint64(1), stats.TotalRequests)
	assert.Equal(t, uint64(1), stats.AllowedRequests)
}

func TestMemoryStorage_Increment(t *testing.T) {
	storage := NewMemoryStorage()
	defer storage.Close()

	ctx := context.Background()
	window := 1 * time.Minute

	count1, _, err1 := storage.Increment(ctx, "test-key", window)
	require.NoError(t, err1)
	assert.Equal(t, int64(1), count1)

	count2, _, err2 := storage.Increment(ctx, "test-key", window)
	require.NoError(t, err2)
	assert.Equal(t, int64(2), count2)
}

func TestMemoryStorage_Get(t *testing.T) {
	storage := NewMemoryStorage()
	defer storage.Close()

	ctx := context.Background()
	window := 1 * time.Minute

	storage.Increment(ctx, "test-key", window)

	count, resetAt, err := storage.Get(ctx, "test-key")
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
	assert.False(t, resetAt.IsZero())
}

func TestMemoryStorage_Reset(t *testing.T) {
	storage := NewMemoryStorage()
	defer storage.Close()

	ctx := context.Background()
	window := 1 * time.Minute

	storage.Increment(ctx, "test-key", window)
	err := storage.Reset(ctx, "test-key")
	require.NoError(t, err)

	count, _, _ := storage.Get(ctx, "test-key")
	assert.Equal(t, int64(0), count)
}

func TestLimiter_Allow(t *testing.T) {
	storage := NewMemoryStorage()
	defer storage.Close()

	config := Config{
		RequestsPerMinute: 5,
		TimeWindow:        1 * time.Minute,
	}

	limiter := NewLimiter(storage, config)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		result, err := limiter.Allow(ctx, "test-key")
		require.NoError(t, err)
		assert.True(t, result.Allowed)
	}

	result, err := limiter.Allow(ctx, "test-key")
	require.NoError(t, err)
	assert.False(t, result.Allowed)
}

func TestLimiter_Reset(t *testing.T) {
	storage := NewMemoryStorage()
	defer storage.Close()

	config := Config{
		RequestsPerMinute: 2,
		TimeWindow:        1 * time.Minute,
	}

	limiter := NewLimiter(storage, config)
	ctx := context.Background()

	limiter.Allow(ctx, "test-key")
	limiter.Allow(ctx, "test-key")

	err := limiter.Reset(ctx, "test-key")
	require.NoError(t, err)

	result, err := limiter.Allow(ctx, "test-key")
	require.NoError(t, err)
	assert.True(t, result.Allowed)
}

func TestLimiter_BuildKey(t *testing.T) {
	storage := NewMemoryStorage()
	defer storage.Close()

	limiter := NewLimiter(storage, Config{})

	key := limiter.BuildKey("user", "123")
	assert.Equal(t, "ratelimit:user:123", key)
}

func TestLimiter_GetLimitForEndpoint(t *testing.T) {
	storage := NewMemoryStorage()
	defer storage.Close()

	config := Config{
		RequestsPerMinute: 60,
		EndpointLimits: map[string]int{
			"/api/heavy": 10,
		},
	}

	limiter := NewLimiter(storage, config)

	assert.Equal(t, 10, limiter.GetLimitForEndpoint("/api/heavy"))
	assert.Equal(t, 60, limiter.GetLimitForEndpoint("/api/normal"))
}

func TestLimiter_AllowN(t *testing.T) {
	storage := NewMemoryStorage()
	defer storage.Close()

	config := Config{
		RequestsPerMinute: 10,
		TimeWindow:        1 * time.Minute,
	}

	limiter := NewLimiter(storage, config)
	ctx := context.Background()

	result, err := limiter.AllowN(ctx, "test-key", 5)
	require.NoError(t, err)
	assert.True(t, result.Allowed)
}

func TestMetrics_RecordRequest(t *testing.T) {
	metrics := NewMetrics()

	metrics.RecordRequest(true, "/api/test", "user123")
	metrics.RecordRequest(true, "/api/test", "user456")
	metrics.RecordRequest(false, "/api/test", "user123")

	stats := metrics.GetStats()
	assert.Equal(t, uint64(3), stats.TotalRequests)
	assert.Equal(t, uint64(2), stats.AllowedRequests)
	assert.Equal(t, uint64(1), stats.BlockedRequests)
}

func TestMetrics_GetBlockedRate(t *testing.T) {
	metrics := NewMetrics()

	for i := 0; i < 7; i++ {
		metrics.RecordRequest(true, "", "")
	}
	for i := 0; i < 3; i++ {
		metrics.RecordRequest(false, "", "")
	}

	rate := metrics.GetBlockedRate()
	assert.Equal(t, 30.0, rate)
}

func TestMetrics_Reset(t *testing.T) {
	metrics := NewMetrics()
	metrics.RecordRequest(true, "/api/test", "user123")

	metrics.Reset()

	stats := metrics.GetStats()
	assert.Equal(t, uint64(0), stats.TotalRequests)
}

func TestMetrics_GetTopEndpoints(t *testing.T) {
	metrics := NewMetrics()

	metrics.RecordRequest(true, "/api/heavy", "")
	metrics.RecordRequest(true, "/api/heavy", "")
	metrics.RecordRequest(true, "/api/medium", "")

	top := metrics.GetTopEndpoints(2)
	assert.Len(t, top, 2)
	assert.Equal(t, "/api/heavy", top[0].Endpoint)
	assert.Equal(t, uint64(2), top[0].Count)
}

func TestMetrics_GetTopUsers(t *testing.T) {
	metrics := NewMetrics()

	metrics.RecordRequest(true, "", "alice")
	metrics.RecordRequest(true, "", "alice")
	metrics.RecordRequest(true, "", "bob")

	top := metrics.GetTopUsers(2)
	assert.Len(t, top, 2)
	assert.Equal(t, "alice", top[0].User)
	assert.Equal(t, uint64(2), top[0].Count)
}

func TestMemoryStorage_WindowExpiration(t *testing.T) {
	storage := NewMemoryStorage()
	defer storage.Close()

	ctx := context.Background()
	window := 100 * time.Millisecond

	count1, _, _ := storage.Increment(ctx, "test-key", window)
	assert.Equal(t, int64(1), count1)

	time.Sleep(150 * time.Millisecond)

	count2, _, _ := storage.Increment(ctx, "test-key", window)
	assert.Equal(t, int64(1), count2)
}

func TestLimiter_RetryAfter(t *testing.T) {
	storage := NewMemoryStorage()
	defer storage.Close()

	config := Config{
		RequestsPerMinute: 1,
		TimeWindow:        1 * time.Minute,
	}

	limiter := NewLimiter(storage, config)
	ctx := context.Background()

	limiter.Allow(ctx, "test-key")

	result, err := limiter.Allow(ctx, "test-key")
	require.NoError(t, err)
	assert.False(t, result.Allowed)
	assert.True(t, result.RetryAfter > 0)
}
