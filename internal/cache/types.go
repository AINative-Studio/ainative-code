package cache

import "time"

// CacheControl represents cache control directives for a prompt segment
type CacheControl struct {
	Type      string `json:"type"`      // "ephemeral" for Anthropic prompt caching
	Enabled   bool   `json:"enabled"`   // Whether caching is enabled for this segment
	CacheKey  string `json:"cache_key"` // Unique identifier for this cached content
	TTL       int    `json:"ttl"`       // Time-to-live in seconds (optional)
	Breakpoint bool  `json:"breakpoint"` // Whether this is a cache breakpoint
}

// CacheMetrics tracks cache performance metrics
type CacheMetrics struct {
	TotalRequests      int64         `json:"total_requests"`
	CacheHits          int64         `json:"cache_hits"`
	CacheMisses        int64         `json:"cache_misses"`
	BytesCached        int64         `json:"bytes_cached"`
	BytesSaved         int64         `json:"bytes_saved"`
	AverageHitRate     float64       `json:"average_hit_rate"`
	LastReset          time.Time     `json:"last_reset"`
	CacheByKey         map[string]*CacheKeyMetrics `json:"cache_by_key"`
}

// CacheKeyMetrics tracks metrics for a specific cache key
type CacheKeyMetrics struct {
	Key         string    `json:"key"`
	Hits        int64     `json:"hits"`
	Misses      int64     `json:"misses"`
	BytesCached int64     `json:"bytes_cached"`
	LastHit     time.Time `json:"last_hit"`
	Created     time.Time `json:"created"`
}

// Config represents caching configuration
type Config struct {
	Enabled           bool          `json:"enabled"`
	MinPromptLength   int           `json:"min_prompt_length"`   // Minimum length to cache (default: 1024)
	MaxPromptLength   int           `json:"max_prompt_length"`   // Maximum length to cache (default: 0 = unlimited)
	SystemPromptCache bool          `json:"system_prompt_cache"` // Cache system prompts
	ContextCache      bool          `json:"context_cache"`       // Cache conversation context
	TTL               time.Duration `json:"ttl"`                 // Default TTL (0 = use provider default)
	AutoCleanup       bool          `json:"auto_cleanup"`        // Automatically cleanup expired cache
	CleanupInterval   time.Duration `json:"cleanup_interval"`    // How often to run cleanup
}

// DefaultConfig returns the default caching configuration
func DefaultConfig() Config {
	return Config{
		Enabled:           true,
		MinPromptLength:   1024,
		MaxPromptLength:   0, // unlimited
		SystemPromptCache: true,
		ContextCache:      true,
		TTL:               5 * time.Minute,
		AutoCleanup:       true,
		CleanupInterval:   1 * time.Minute,
	}
}

// CacheableContent represents content that can be cached
type CacheableContent struct {
	Content     string        `json:"content"`
	Type        string        `json:"type"`        // "system", "context", "tools"
	Length      int           `json:"length"`
	CacheKey    string        `json:"cache_key"`
	Priority    int           `json:"priority"`    // Higher priority cached first
	Breakpoint  bool          `json:"breakpoint"`  // Mark as cache breakpoint
}

// CacheStatus represents the current status of a cache entry
type CacheStatus struct {
	Key        string    `json:"key"`
	Cached     bool      `json:"cached"`
	ExpiresAt  time.Time `json:"expires_at"`
	BytesSize  int64     `json:"bytes_size"`
	HitCount   int64     `json:"hit_count"`
	LastAccess time.Time `json:"last_access"`
}
