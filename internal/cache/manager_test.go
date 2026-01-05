package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	config := DefaultConfig()
	manager := NewManager(config)

	assert.NotNil(t, manager)
	assert.NotNil(t, manager.metrics)
	assert.NotNil(t, manager.cache)
	assert.Equal(t, config, manager.config)
}

func TestShouldCache(t *testing.T) {
	tests := []struct {
		name       string
		config     Config
		content    *CacheableContent
		shouldCache bool
	}{
		{
			name:   "caching disabled",
			config: Config{Enabled: false},
			content: &CacheableContent{
				Content: "test content",
				Type:    "system",
				Length:  2000,
			},
			shouldCache: false,
		},
		{
			name: "content too short",
			config: Config{
				Enabled:         true,
				MinPromptLength: 1024,
			},
			content: &CacheableContent{
				Content: "short",
				Type:    "system",
				Length:  10,
			},
			shouldCache: false,
		},
		{
			name: "content too long",
			config: Config{
				Enabled:         true,
				MinPromptLength: 100,
				MaxPromptLength: 500,
			},
			content: &CacheableContent{
				Content: "very long content",
				Type:    "system",
				Length:  1000,
			},
			shouldCache: false,
		},
		{
			name: "system prompt caching disabled",
			config: Config{
				Enabled:           true,
				MinPromptLength:   100,
				SystemPromptCache: false,
			},
			content: &CacheableContent{
				Content: "system prompt",
				Type:    "system",
				Length:  200,
			},
			shouldCache: false,
		},
		{
			name: "context caching disabled",
			config: Config{
				Enabled:         true,
				MinPromptLength: 100,
				ContextCache:    false,
			},
			content: &CacheableContent{
				Content: "context",
				Type:    "context",
				Length:  200,
			},
			shouldCache: false,
		},
		{
			name: "valid cacheable content",
			config: Config{
				Enabled:           true,
				MinPromptLength:   100,
				SystemPromptCache: true,
				TTL:               5 * time.Minute,
			},
			content: &CacheableContent{
				Content: "cacheable system prompt",
				Type:    "system",
				Length:  200,
			},
			shouldCache: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewManager(tt.config)
			control := manager.ShouldCache(tt.content)

			if tt.shouldCache {
				assert.NotNil(t, control)
				assert.True(t, control.Enabled)
				assert.Equal(t, "ephemeral", control.Type)
				assert.NotEmpty(t, control.CacheKey)
			} else {
				assert.Nil(t, control)
			}
		})
	}
}

func TestGenerateCacheKey(t *testing.T) {
	manager := NewManager(DefaultConfig())

	key1 := manager.generateCacheKey("content1", "system")
	key2 := manager.generateCacheKey("content2", "system")
	key3 := manager.generateCacheKey("content1", "context")

	// Keys should be different for different content
	assert.NotEqual(t, key1, key2)

	// Keys should be different for same content but different types
	assert.NotEqual(t, key1, key3)

	// Keys should be consistent
	key4 := manager.generateCacheKey("content1", "system")
	assert.Equal(t, key1, key4)

	// Keys should be 16 characters
	assert.Len(t, key1, 16)
}

func TestRecordCacheHit(t *testing.T) {
	manager := NewManager(DefaultConfig())
	cacheKey := "test-key"

	// Record cache hit
	manager.RecordCacheHit(cacheKey)

	stats := manager.GetStats()
	assert.Equal(t, int64(1), stats.TotalRequests)
	assert.Equal(t, int64(1), stats.CacheHits)
	assert.Equal(t, int64(0), stats.CacheMisses)
	assert.Equal(t, 1.0, stats.AverageHitRate)

	// Check key-specific metrics
	keyMetrics, exists := stats.CacheByKey[cacheKey]
	require.True(t, exists)
	assert.Equal(t, int64(1), keyMetrics.Hits)
	assert.Equal(t, int64(0), keyMetrics.Misses)
}

func TestRecordCacheMiss(t *testing.T) {
	manager := NewManager(DefaultConfig())
	cacheKey := "test-key"

	// Record cache miss
	manager.RecordCacheMiss(cacheKey)

	stats := manager.GetStats()
	assert.Equal(t, int64(1), stats.TotalRequests)
	assert.Equal(t, int64(0), stats.CacheHits)
	assert.Equal(t, int64(1), stats.CacheMisses)
	assert.Equal(t, 0.0, stats.AverageHitRate)

	// Check key-specific metrics
	keyMetrics, exists := stats.CacheByKey[cacheKey]
	require.True(t, exists)
	assert.Equal(t, int64(0), keyMetrics.Hits)
	assert.Equal(t, int64(1), keyMetrics.Misses)
}

func TestRecordCached(t *testing.T) {
	manager := NewManager(DefaultConfig())
	cacheKey := "test-key"
	bytesSize := int64(1024)

	// Record cached content
	manager.RecordCached(cacheKey, bytesSize)

	stats := manager.GetStats()
	assert.Equal(t, bytesSize, stats.BytesCached)

	// Check cache status
	status := manager.GetCacheStatus(cacheKey)
	require.NotNil(t, status)
	assert.True(t, status.Cached)
	assert.Equal(t, bytesSize, status.BytesSize)
	assert.True(t, status.ExpiresAt.After(time.Now()))
}

func TestRecordBytesSaved(t *testing.T) {
	manager := NewManager(DefaultConfig())

	manager.RecordBytesSaved(512)
	manager.RecordBytesSaved(256)

	stats := manager.GetStats()
	assert.Equal(t, int64(768), stats.BytesSaved)
}

func TestCacheHitRate(t *testing.T) {
	manager := NewManager(DefaultConfig())

	// Record 3 hits and 1 miss
	manager.RecordCacheHit("key1")
	manager.RecordCacheHit("key2")
	manager.RecordCacheMiss("key3")
	manager.RecordCacheHit("key1")

	stats := manager.GetStats()
	assert.Equal(t, int64(4), stats.TotalRequests)
	assert.Equal(t, int64(3), stats.CacheHits)
	assert.Equal(t, int64(1), stats.CacheMisses)
	assert.InDelta(t, 0.75, stats.AverageHitRate, 0.01)
}

func TestIsCached(t *testing.T) {
	config := DefaultConfig()
	config.TTL = 100 * time.Millisecond
	manager := NewManager(config)

	cacheKey := "test-key"

	// Not cached initially
	assert.False(t, manager.IsCached(cacheKey))

	// After recording as cached
	manager.RecordCached(cacheKey, 1024)
	assert.True(t, manager.IsCached(cacheKey))

	// After expiration
	time.Sleep(150 * time.Millisecond)
	assert.False(t, manager.IsCached(cacheKey))
}

func TestInvalidateCache(t *testing.T) {
	manager := NewManager(DefaultConfig())
	cacheKey := "test-key"

	// Cache some content
	manager.RecordCached(cacheKey, 1024)
	assert.True(t, manager.IsCached(cacheKey))

	// Invalidate
	manager.InvalidateCache(cacheKey)
	assert.False(t, manager.IsCached(cacheKey))
}

func TestClearCache(t *testing.T) {
	manager := NewManager(DefaultConfig())

	// Cache multiple items
	manager.RecordCached("key1", 1024)
	manager.RecordCached("key2", 2048)
	manager.RecordCached("key3", 512)

	assert.Len(t, manager.GetCacheKeys(), 3)

	// Clear all
	manager.ClearCache()
	assert.Len(t, manager.GetCacheKeys(), 0)
}

func TestResetMetrics(t *testing.T) {
	manager := NewManager(DefaultConfig())

	// Generate some metrics
	manager.RecordCacheHit("key1")
	manager.RecordCacheMiss("key2")
	manager.RecordBytesSaved(1024)

	stats := manager.GetStats()
	assert.Greater(t, stats.TotalRequests, int64(0))

	// Reset
	manager.ResetMetrics()

	stats = manager.GetStats()
	assert.Equal(t, int64(0), stats.TotalRequests)
	assert.Equal(t, int64(0), stats.CacheHits)
	assert.Equal(t, int64(0), stats.CacheMisses)
	assert.Equal(t, int64(0), stats.BytesSaved)
}

func TestGetCacheSize(t *testing.T) {
	manager := NewManager(DefaultConfig())

	manager.RecordCached("key1", 1024)
	manager.RecordCached("key2", 2048)
	manager.RecordCached("key3", 512)

	totalSize := manager.GetCacheSize()
	assert.Equal(t, int64(3584), totalSize)
}

func TestGetCacheKeys(t *testing.T) {
	manager := NewManager(DefaultConfig())

	manager.RecordCached("key1", 1024)
	manager.RecordCached("key2", 2048)

	keys := manager.GetCacheKeys()
	assert.Len(t, keys, 2)
	assert.Contains(t, keys, "key1")
	assert.Contains(t, keys, "key2")
}

func TestAutoCleanup(t *testing.T) {
	config := DefaultConfig()
	config.TTL = 50 * time.Millisecond
	config.CleanupInterval = 30 * time.Millisecond
	manager := NewManager(config)
	defer manager.Stop()

	// Cache some content
	manager.RecordCached("key1", 1024)
	assert.Len(t, manager.GetCacheKeys(), 1)

	// Wait for expiration and cleanup
	time.Sleep(150 * time.Millisecond)

	// Should be cleaned up
	assert.Len(t, manager.GetCacheKeys(), 0)
}

func TestUpdateConfig(t *testing.T) {
	manager := NewManager(DefaultConfig())

	newConfig := Config{
		Enabled:         false,
		MinPromptLength: 2048,
	}

	manager.UpdateConfig(newConfig)

	assert.Equal(t, newConfig, manager.GetConfig())
}

func TestFormatStats(t *testing.T) {
	manager := NewManager(DefaultConfig())

	manager.RecordCacheHit("key1")
	manager.RecordCacheMiss("key2")
	manager.RecordBytesSaved(1024)

	stats := manager.FormatStats()
	assert.Contains(t, stats, "Cache Statistics")
	assert.Contains(t, stats, "Total Requests: 2")
	assert.Contains(t, stats, "Cache Hits: 1")
	assert.Contains(t, stats, "Cache Misses: 1")
}

func TestConcurrentAccess(t *testing.T) {
	manager := NewManager(DefaultConfig())

	// Concurrent cache hits
	done := make(chan bool)
	for i := 0; i < 100; i++ {
		go func(n int) {
			manager.RecordCacheHit("key1")
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 100; i++ {
		<-done
	}

	stats := manager.GetStats()
	assert.Equal(t, int64(100), stats.CacheHits)
}

func BenchmarkShouldCache(b *testing.B) {
	manager := NewManager(DefaultConfig())
	content := &CacheableContent{
		Content: "test content for benchmarking",
		Type:    "system",
		Length:  2000,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.ShouldCache(content)
	}
}

func BenchmarkGenerateCacheKey(b *testing.B) {
	manager := NewManager(DefaultConfig())
	content := "test content for cache key generation benchmark"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.generateCacheKey(content, "system")
	}
}

func BenchmarkRecordCacheHit(b *testing.B) {
	manager := NewManager(DefaultConfig())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.RecordCacheHit("test-key")
	}
}
