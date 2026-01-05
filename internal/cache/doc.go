// Package cache implements ephemeral prompt caching for improved performance.
//
// This package provides caching mechanisms for AI prompts to reduce costs
// and improve response times through Anthropic's prompt caching feature.
//
// Features:
//   - Cache control headers for system prompts
//   - Cache control headers for large context
//   - Automatic cache key generation
//   - Cache hit/miss metrics tracking
//   - Configurable cache behavior
//   - TTL-based cache expiration
//
// Prompt caching works by marking portions of prompts as cacheable,
// allowing the AI provider to reuse processed prompts across requests.
//
// Example usage:
//
//	import "github.com/AINative-studio/ainative-code/internal/cache"
//
//	// Create cache manager
//	manager := cache.NewManager(cache.Config{
//	    Enabled:           true,
//	    MinPromptLength:   1024,
//	    SystemPromptCache: true,
//	})
//
//	// Mark content as cacheable
//	cacheControl := manager.ShouldCache(content)
//	if cacheControl != nil {
//	    // Add cache control to API request
//	}
//
//	// Track cache metrics
//	manager.RecordCacheHit("system_prompt")
//	stats := manager.GetStats()
package cache
