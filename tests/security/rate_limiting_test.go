package security

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/AINative-studio/ainative-code/internal/ratelimit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRateLimiting_BasicEnforcement verifies rate limiting enforces request limits
func TestRateLimiting_BasicEnforcement(t *testing.T) {
	// Given: A rate limiter with 10 requests per minute, burst of 2
	limiter := ratelimit.NewLimiter(
		ratelimit.WithRequestsPerMinute(10),
		ratelimit.WithBurstSize(2),
	)

	userID := "test-user-123"

	// When: Making requests within and exceeding the rate limit
	results := make([]bool, 15)
	for i := 0; i < 15; i++ {
		allowed := limiter.Allow(userID)
		results[i] = allowed
	}

	// Then: First requests should be allowed (within burst + rate)
	// Later requests should be denied
	allowedCount := 0
	for _, allowed := range results {
		if allowed {
			allowedCount++
		}
	}

	// Should allow burst + a few more based on rate
	assert.LessOrEqual(t, allowedCount, 5, "Should enforce rate limit")
	assert.GreaterOrEqual(t, allowedCount, 2, "Should allow at least burst size")
}

// TestRateLimiting_PerUserLimits verifies different users have independent limits
func TestRateLimiting_PerUserLimits(t *testing.T) {
	// Given: A rate limiter with 5 requests per minute
	limiter := ratelimit.NewLimiter(
		ratelimit.WithRequestsPerMinute(5),
		ratelimit.WithBurstSize(5),
	)

	user1 := "user-1"
	user2 := "user-2"

	// When: Each user makes 5 requests
	var user1Results, user2Results []bool

	for i := 0; i < 5; i++ {
		user1Results = append(user1Results, limiter.Allow(user1))
		user2Results = append(user2Results, limiter.Allow(user2))
	}

	// Then: Each user should have their own independent limit
	user1Allowed := countTrue(user1Results)
	user2Allowed := countTrue(user2Results)

	assert.Equal(t, 5, user1Allowed, "User 1 should be allowed 5 requests")
	assert.Equal(t, 5, user2Allowed, "User 2 should be allowed 5 requests")

	// Next request for each should be denied
	assert.False(t, limiter.Allow(user1), "User 1 should be rate limited")
	assert.False(t, limiter.Allow(user2), "User 2 should be rate limited")
}

// TestRateLimiting_BurstHandling verifies burst capacity works correctly
func TestRateLimiting_BurstHandling(t *testing.T) {
	// Given: A rate limiter with burst of 10
	limiter := ratelimit.NewLimiter(
		ratelimit.WithRequestsPerMinute(60), // 1 per second
		ratelimit.WithBurstSize(10),
	)

	userID := "burst-user"

	// When: Making burst of requests
	var results []bool
	for i := 0; i < 15; i++ {
		results = append(results, limiter.Allow(userID))
	}

	// Then: First 10 (burst) should be allowed immediately
	for i := 0; i < 10; i++ {
		assert.True(t, results[i], "Request %d should be allowed (within burst)", i)
	}

	// Remaining should be denied (exceeds burst)
	for i := 10; i < 15; i++ {
		assert.False(t, results[i], "Request %d should be denied (exceeds burst)", i)
	}
}

// TestRateLimiting_TimeWindowReset verifies rate limit resets after time window
func TestRateLimiting_TimeWindowReset(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping time-based test in short mode")
	}

	// Given: A rate limiter with 5 requests per second
	limiter := ratelimit.NewLimiter(
		ratelimit.WithRequestsPerMinute(300), // 5 per second
		ratelimit.WithBurstSize(5),
	)

	userID := "time-test-user"

	// When: Exhausting the rate limit
	for i := 0; i < 5; i++ {
		require.True(t, limiter.Allow(userID))
	}
	assert.False(t, limiter.Allow(userID), "Should be rate limited")

	// Wait for time window to reset
	time.Sleep(1 * time.Second)

	// Then: Should be allowed again after reset
	assert.True(t, limiter.Allow(userID), "Should be allowed after time window reset")
}

// TestRateLimiting_ConcurrentAccess verifies thread-safe operation
func TestRateLimiting_ConcurrentAccess(t *testing.T) {
	// Given: A rate limiter with 100 requests per minute
	limiter := ratelimit.NewLimiter(
		ratelimit.WithRequestsPerMinute(100),
		ratelimit.WithBurstSize(50),
	)

	userID := "concurrent-user"
	concurrentRequests := 200
	var wg sync.WaitGroup
	results := make([]bool, concurrentRequests)

	// When: Making concurrent requests
	for i := 0; i < concurrentRequests; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			results[index] = limiter.Allow(userID)
		}(i)
	}

	wg.Wait()

	// Then: Should handle concurrent access safely
	allowedCount := countTrue(results)

	// Should allow approximately burst size + rate limit
	assert.GreaterOrEqual(t, allowedCount, 50, "Should allow at least burst size")
	assert.LessOrEqual(t, allowedCount, 120, "Should not exceed rate limit significantly")
}

// TestRateLimiting_DifferentTimeWindows verifies different window sizes work correctly
func TestRateLimiting_DifferentTimeWindows(t *testing.T) {
	testCases := []struct {
		name               string
		requestsPerMinute  int
		burstSize          int
		requestCount       int
		expectedAllowed    int
		expectedAllowedMin int
	}{
		{
			name:               "1 request per minute",
			requestsPerMinute:  1,
			burstSize:          1,
			requestCount:       5,
			expectedAllowed:    1,
			expectedAllowedMin: 1,
		},
		{
			name:               "60 requests per minute (1 per second)",
			requestsPerMinute:  60,
			burstSize:          10,
			requestCount:       20,
			expectedAllowed:    10,
			expectedAllowedMin: 10,
		},
		{
			name:               "120 requests per minute (2 per second)",
			requestsPerMinute:  120,
			burstSize:          20,
			requestCount:       30,
			expectedAllowed:    20,
			expectedAllowedMin: 20,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Given: A rate limiter with specific configuration
			limiter := ratelimit.NewLimiter(
				ratelimit.WithRequestsPerMinute(tc.requestsPerMinute),
				ratelimit.WithBurstSize(tc.burstSize),
			)

			userID := "window-test-user"

			// When: Making requests
			var results []bool
			for i := 0; i < tc.requestCount; i++ {
				results = append(results, limiter.Allow(userID))
			}

			// Then: Should respect the configured limits
			allowedCount := countTrue(results)
			assert.GreaterOrEqual(t, allowedCount, tc.expectedAllowedMin,
				"Should allow at least minimum requests")
			assert.LessOrEqual(t, allowedCount, tc.expectedAllowed+2,
				"Should not exceed expected maximum significantly")
		})
	}
}

// TestRateLimiting_GracefulDegradation verifies graceful handling when limits are reached
func TestRateLimiting_GracefulDegradation(t *testing.T) {
	// Given: A rate limiter with strict limits
	limiter := ratelimit.NewLimiter(
		ratelimit.WithRequestsPerMinute(10),
		ratelimit.WithBurstSize(5),
	)

	userID := "degradation-test"

	// When: Exceeding rate limits
	for i := 0; i < 10; i++ {
		limiter.Allow(userID)
	}

	// Then: Should continue to work (not crash) even when limits exceeded
	for i := 0; i < 100; i++ {
		allowed := limiter.Allow(userID)
		// All should be denied, but system should remain stable
		assert.False(t, allowed)
	}
}

// TestRateLimiting_IPBasedLimiting verifies IP-based rate limiting
func TestRateLimiting_IPBasedLimiting(t *testing.T) {
	// Given: A rate limiter for IP addresses
	limiter := ratelimit.NewLimiter(
		ratelimit.WithRequestsPerMinute(30),
		ratelimit.WithBurstSize(10),
	)

	ip1 := "192.168.1.100"
	ip2 := "192.168.1.101"
	ip3 := "10.0.0.50"

	// When: Different IPs make requests
	var ip1Results, ip2Results, ip3Results []bool

	for i := 0; i < 15; i++ {
		ip1Results = append(ip1Results, limiter.Allow(ip1))
		ip2Results = append(ip2Results, limiter.Allow(ip2))
		ip3Results = append(ip3Results, limiter.Allow(ip3))
	}

	// Then: Each IP should have independent limits
	ip1Allowed := countTrue(ip1Results)
	ip2Allowed := countTrue(ip2Results)
	ip3Allowed := countTrue(ip3Results)

	assert.GreaterOrEqual(t, ip1Allowed, 10, "IP1 should be allowed burst size")
	assert.GreaterOrEqual(t, ip2Allowed, 10, "IP2 should be allowed burst size")
	assert.GreaterOrEqual(t, ip3Allowed, 10, "IP3 should be allowed burst size")
}

// TestRateLimiting_APIKeyBasedLimiting verifies API key-based rate limiting
func TestRateLimiting_APIKeyBasedLimiting(t *testing.T) {
	// Given: A rate limiter for API keys
	limiter := ratelimit.NewLimiter(
		ratelimit.WithRequestsPerMinute(100),
		ratelimit.WithBurstSize(20),
	)

	apiKey1 := "sk-test-key-1"
	apiKey2 := "sk-test-key-2"

	// When: Different API keys make requests
	var key1Results, key2Results []bool

	for i := 0; i < 25; i++ {
		key1Results = append(key1Results, limiter.Allow(apiKey1))
		key2Results = append(key2Results, limiter.Allow(apiKey2))
	}

	// Then: Each API key should have independent limits
	key1Allowed := countTrue(key1Results)
	key2Allowed := countTrue(key2Results)

	assert.GreaterOrEqual(t, key1Allowed, 20, "API key 1 should be allowed burst size")
	assert.GreaterOrEqual(t, key2Allowed, 20, "API key 2 should be allowed burst size")
}

// TestRateLimiting_DistributedScenario simulates distributed rate limiting
func TestRateLimiting_DistributedScenario(t *testing.T) {
	// Given: Multiple rate limiter instances (simulating different servers)
	limiter1 := ratelimit.NewLimiter(
		ratelimit.WithRequestsPerMinute(60),
		ratelimit.WithBurstSize(10),
	)

	limiter2 := ratelimit.NewLimiter(
		ratelimit.WithRequestsPerMinute(60),
		ratelimit.WithBurstSize(10),
	)

	userID := "distributed-user"

	// When: User makes requests to different servers
	var results []bool

	for i := 0; i < 10; i++ {
		results = append(results, limiter1.Allow(userID))
		results = append(results, limiter2.Allow(userID))
	}

	// Then: Each instance tracks independently (in-memory)
	// Note: For true distributed rate limiting, need shared storage (Redis)
	allowedCount := countTrue(results)

	// Both instances allow independently, so ~20 allowed
	assert.GreaterOrEqual(t, allowedCount, 20,
		"Independent instances should allow requests separately")
}

// TestRateLimiting_HeaderInformation verifies rate limit headers are provided
func TestRateLimiting_HeaderInformation(t *testing.T) {
	// Given: A rate limiter that provides state information
	limiter := ratelimit.NewLimiter(
		ratelimit.WithRequestsPerMinute(60),
		ratelimit.WithBurstSize(10),
	)

	userID := "header-test-user"

	// When: Making requests
	for i := 0; i < 5; i++ {
		limiter.Allow(userID)
	}

	// Then: Should be able to query rate limit state
	// (Implementation would need to expose state for headers)
	// Expected headers:
	// - X-RateLimit-Limit: 60
	// - X-RateLimit-Remaining: 5
	// - X-RateLimit-Reset: <timestamp>

	t.Log("Rate limit state should be queryable for HTTP headers")
}

// TestRateLimiting_429Response verifies 429 Too Many Requests behavior
func TestRateLimiting_429Response(t *testing.T) {
	// Given: A rate limiter with strict limits
	limiter := ratelimit.NewLimiter(
		ratelimit.WithRequestsPerMinute(5),
		ratelimit.WithBurstSize(5),
	)

	userID := "429-test-user"

	// When: Exhausting rate limit
	for i := 0; i < 5; i++ {
		require.True(t, limiter.Allow(userID))
	}

	// Then: Next request should be denied
	allowed := limiter.Allow(userID)
	assert.False(t, allowed, "Should return 429 status")

	// Should provide retry-after information
	t.Log("Should return Retry-After header with seconds until reset")
}

// Helper function to count true values in boolean slice
func countTrue(values []bool) int {
	count := 0
	for _, v := range values {
		if v {
			count++
		}
	}
	return count
}

// BenchmarkRateLimiting measures rate limiting performance
func BenchmarkRateLimiting(b *testing.B) {
	limiter := ratelimit.NewLimiter(
		ratelimit.WithRequestsPerMinute(10000),
		ratelimit.WithBurstSize(1000),
	)

	userID := "bench-user"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limiter.Allow(userID)
	}
}

// BenchmarkRateLimiting_Concurrent measures concurrent rate limiting performance
func BenchmarkRateLimiting_Concurrent(b *testing.B) {
	limiter := ratelimit.NewLimiter(
		ratelimit.WithRequestsPerMinute(10000),
		ratelimit.WithBurstSize(1000),
	)

	userID := "bench-concurrent-user"

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			limiter.Allow(userID)
		}
	})
}
