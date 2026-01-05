package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// Manager manages prompt caching
type Manager struct {
	config  Config
	metrics *CacheMetrics
	cache   map[string]*CacheStatus
	mu      sync.RWMutex
	stopCh  chan struct{}
	wg      sync.WaitGroup
}

// NewManager creates a new cache manager
func NewManager(config Config) *Manager {
	manager := &Manager{
		config: config,
		metrics: &CacheMetrics{
			LastReset:  time.Now(),
			CacheByKey: make(map[string]*CacheKeyMetrics),
		},
		cache:  make(map[string]*CacheStatus),
		stopCh: make(chan struct{}),
	}

	// Start auto-cleanup if enabled
	if config.AutoCleanup && config.CleanupInterval > 0 {
		manager.startCleanup()
	}

	return manager
}

// ShouldCache determines if content should be cached and returns cache control
func (m *Manager) ShouldCache(content *CacheableContent) *CacheControl {
	if !m.config.Enabled {
		return nil
	}

	// Check minimum length
	if content.Length < m.config.MinPromptLength {
		return nil
	}

	// Check maximum length (if set)
	if m.config.MaxPromptLength > 0 && content.Length > m.config.MaxPromptLength {
		return nil
	}

	// Check type-specific caching rules
	switch content.Type {
	case "system":
		if !m.config.SystemPromptCache {
			return nil
		}
	case "context":
		if !m.config.ContextCache {
			return nil
		}
	}

	// Generate cache key if not provided
	cacheKey := content.CacheKey
	if cacheKey == "" {
		cacheKey = m.generateCacheKey(content.Content, content.Type)
	}

	// Create cache control
	control := &CacheControl{
		Type:      "ephemeral",
		Enabled:   true,
		CacheKey:  cacheKey,
		Breakpoint: content.Breakpoint,
	}

	// Add TTL if configured
	if m.config.TTL > 0 {
		control.TTL = int(m.config.TTL.Seconds())
	}

	return control
}

// GenerateCacheKey creates a unique cache key for content
func (m *Manager) generateCacheKey(content, contentType string) string {
	h := sha256.New()
	h.Write([]byte(contentType))
	h.Write([]byte(":"))
	h.Write([]byte(content))
	return hex.EncodeToString(h.Sum(nil))[:16] // Use first 16 chars
}

// RecordCacheHit records a cache hit for metrics
func (m *Manager) RecordCacheHit(cacheKey string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.metrics.TotalRequests++
	m.metrics.CacheHits++

	// Update key-specific metrics
	if keyMetrics, exists := m.metrics.CacheByKey[cacheKey]; exists {
		keyMetrics.Hits++
		keyMetrics.LastHit = time.Now()
	} else {
		m.metrics.CacheByKey[cacheKey] = &CacheKeyMetrics{
			Key:     cacheKey,
			Hits:    1,
			Misses:  0,
			LastHit: time.Now(),
			Created: time.Now(),
		}
	}

	// Update cache status
	if status, exists := m.cache[cacheKey]; exists {
		status.HitCount++
		status.LastAccess = time.Now()
	}

	m.updateHitRate()
}

// RecordCacheMiss records a cache miss for metrics
func (m *Manager) RecordCacheMiss(cacheKey string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.metrics.TotalRequests++
	m.metrics.CacheMisses++

	// Update key-specific metrics
	if keyMetrics, exists := m.metrics.CacheByKey[cacheKey]; exists {
		keyMetrics.Misses++
	} else {
		m.metrics.CacheByKey[cacheKey] = &CacheKeyMetrics{
			Key:     cacheKey,
			Hits:    0,
			Misses:  1,
			Created: time.Now(),
		}
	}

	m.updateHitRate()
}

// RecordCached records that content was cached
func (m *Manager) RecordCached(cacheKey string, bytesSize int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.metrics.BytesCached += bytesSize

	// Update key metrics
	if keyMetrics, exists := m.metrics.CacheByKey[cacheKey]; exists {
		keyMetrics.BytesCached = bytesSize
	}

	// Create or update cache status
	expiresAt := time.Now().Add(m.config.TTL)
	if m.config.TTL == 0 {
		expiresAt = time.Now().Add(5 * time.Minute) // Default 5 min
	}

	m.cache[cacheKey] = &CacheStatus{
		Key:        cacheKey,
		Cached:     true,
		ExpiresAt:  expiresAt,
		BytesSize:  bytesSize,
		HitCount:   0,
		LastAccess: time.Now(),
	}
}

// RecordBytesSaved records bytes saved due to caching
func (m *Manager) RecordBytesSaved(bytes int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.metrics.BytesSaved += bytes
}

// GetStats returns current cache statistics
func (m *Manager) GetStats() *CacheMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Create a copy to avoid concurrent access issues
	stats := &CacheMetrics{
		TotalRequests:  m.metrics.TotalRequests,
		CacheHits:      m.metrics.CacheHits,
		CacheMisses:    m.metrics.CacheMisses,
		BytesCached:    m.metrics.BytesCached,
		BytesSaved:     m.metrics.BytesSaved,
		AverageHitRate: m.metrics.AverageHitRate,
		LastReset:      m.metrics.LastReset,
		CacheByKey:     make(map[string]*CacheKeyMetrics),
	}

	// Copy key metrics
	for k, v := range m.metrics.CacheByKey {
		keyMetricsCopy := *v
		stats.CacheByKey[k] = &keyMetricsCopy
	}

	return stats
}

// GetCacheStatus returns the status of a specific cache key
func (m *Manager) GetCacheStatus(cacheKey string) *CacheStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if status, exists := m.cache[cacheKey]; exists {
		statusCopy := *status
		return &statusCopy
	}

	return nil
}

// IsCached checks if content is currently cached
func (m *Manager) IsCached(cacheKey string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status, exists := m.cache[cacheKey]
	if !exists {
		return false
	}

	// Check if expired
	if time.Now().After(status.ExpiresAt) {
		return false
	}

	return status.Cached
}

// InvalidateCache invalidates a specific cache key
func (m *Manager) InvalidateCache(cacheKey string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.cache, cacheKey)
}

// ClearCache clears all cached entries
func (m *Manager) ClearCache() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.cache = make(map[string]*CacheStatus)
}

// ResetMetrics resets all cache metrics
func (m *Manager) ResetMetrics() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.metrics = &CacheMetrics{
		LastReset:  time.Now(),
		CacheByKey: make(map[string]*CacheKeyMetrics),
	}
}

// updateHitRate calculates and updates the average hit rate
func (m *Manager) updateHitRate() {
	if m.metrics.TotalRequests == 0 {
		m.metrics.AverageHitRate = 0
		return
	}

	m.metrics.AverageHitRate = float64(m.metrics.CacheHits) / float64(m.metrics.TotalRequests)
}

// startCleanup starts the background cleanup process
func (m *Manager) startCleanup() {
	m.wg.Add(1)
	go m.cleanupLoop()
}

// cleanupLoop runs periodic cleanup of expired cache entries
func (m *Manager) cleanupLoop() {
	defer m.wg.Done()

	interval := m.config.CleanupInterval
	if interval <= 0 {
		interval = 1 * time.Minute // Default interval
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.cleanup()
		case <-m.stopCh:
			return
		}
	}
}

// cleanup removes expired cache entries
func (m *Manager) cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for key, status := range m.cache {
		if now.After(status.ExpiresAt) {
			delete(m.cache, key)
		}
	}
}

// Stop stops the cache manager and cleanup processes
func (m *Manager) Stop() {
	close(m.stopCh)
	m.wg.Wait()
}

// GetConfig returns the current configuration
func (m *Manager) GetConfig() Config {
	return m.config
}

// UpdateConfig updates the cache configuration
func (m *Manager) UpdateConfig(config Config) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.config = config
}

// GetCacheKeys returns all currently cached keys
func (m *Manager) GetCacheKeys() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	keys := make([]string, 0, len(m.cache))
	for key := range m.cache {
		keys = append(keys, key)
	}

	return keys
}

// GetCacheSize returns the total size of cached content in bytes
func (m *Manager) GetCacheSize() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var total int64
	for _, status := range m.cache {
		total += status.BytesSize
	}

	return total
}

// FormatStats formats cache statistics as a human-readable string
func (m *Manager) FormatStats() string {
	stats := m.GetStats()

	return fmt.Sprintf(`Cache Statistics:
  Total Requests: %d
  Cache Hits: %d
  Cache Misses: %d
  Hit Rate: %.2f%%
  Bytes Cached: %d
  Bytes Saved: %d
  Active Keys: %d
  Last Reset: %s`,
		stats.TotalRequests,
		stats.CacheHits,
		stats.CacheMisses,
		stats.AverageHitRate*100,
		stats.BytesCached,
		stats.BytesSaved,
		len(stats.CacheByKey),
		stats.LastReset.Format(time.RFC3339))
}
