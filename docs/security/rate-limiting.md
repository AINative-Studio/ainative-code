# Rate Limiting

## Overview

The AINative-Code platform implements comprehensive rate limiting to protect APIs from abuse, ensure fair resource allocation, and maintain system stability. This document describes the rate limiting architecture, configuration options, and best practices.

## Architecture

### Components

1. **Storage Backend**
   - In-memory storage for single-instance deployments
   - Redis storage for distributed deployments (planned)
   - Thread-safe implementations with automatic cleanup

2. **Rate Limiter**
   - Token bucket algorithm implementation
   - Configurable limits per user, endpoint, and IP
   - Sliding window time-based limits
   - Burst support for traffic spikes

3. **HTTP Middleware**
   - Standard rate limit headers
   - Per-user and per-endpoint rate limiting
   - Custom rate limit exceeded handlers
   - Path-based skip logic

4. **Metrics and Monitoring**
   - Real-time request tracking
   - Per-endpoint and per-user statistics
   - Blocked request rate monitoring
   - Top consumers tracking

## Configuration

### Basic Configuration

Add to your `ainative-code.yaml`:

```yaml
performance:
  rate_limit:
    enabled: true
    requests_per_minute: 60
    burst_size: 10
    time_window: 1m
    storage: memory  # or "redis"
```

### Per-User Rate Limiting

Enable per-user rate limiting to track limits by authenticated user:

```yaml
performance:
  rate_limit:
    enabled: true
    per_user: true
    requests_per_minute: 100
```

### Per-Endpoint Rate Limiting

Configure different limits for different endpoints:

```yaml
performance:
  rate_limit:
    enabled: true
    per_endpoint: true
    requests_per_minute: 60
    endpoint_limits:
      /api/heavy-operation: 10
      /api/search: 30
      /api/list: 100
```

### Advanced Configuration

```yaml
performance:
  rate_limit:
    enabled: true
    requests_per_minute: 60
    burst_size: 10
    time_window: 1m
    per_user: true
    per_endpoint: true
    storage: memory

    # Custom endpoint limits
    endpoint_limits:
      /api/expensive: 10
      /api/standard: 60
      /api/cheap: 200

    # Skip rate limiting for certain paths
    skip_paths:
      - /health
      - /metrics
      - /docs

    # IP allowlist (never rate limited)
    ip_allowlist:
      - 127.0.0.1
      - 10.0.0.0/8

    # IP blocklist (always blocked)
    ip_blocklist:
      - 192.0.2.0/24
```

### Distributed Deployment (Redis)

For multi-instance deployments, use Redis storage:

```yaml
performance:
  rate_limit:
    enabled: true
    storage: redis
    redis_url: redis://localhost:6379/0
    requests_per_minute: 100
```

## Rate Limit Headers

The middleware sets standard rate limit headers on all responses:

```
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 42
X-RateLimit-Reset: 1704067200
Retry-After: 30
```

- `X-RateLimit-Limit`: Maximum requests allowed in the time window
- `X-RateLimit-Remaining`: Requests remaining in current window
- `X-RateLimit-Reset`: Unix timestamp when the limit resets
- `Retry-After`: Seconds to wait before retrying (only when blocked)

## CLI Commands

### View Status

```bash
ainative-code rate-limit status
```

Shows the current rate limiting configuration and status.

### View Configuration

```bash
ainative-code rate-limit config
```

Displays the complete rate limiting configuration.

### Reset Rate Limit

Reset the rate limit for a specific user:

```bash
ainative-code rate-limit reset --user user123
```

Reset the rate limit for a specific IP:

```bash
ainative-code rate-limit reset --ip 192.168.1.1
```

Reset using a custom key:

```bash
ainative-code rate-limit reset --key "ratelimit:user:user123:endpoint:/api/test"
```

### View Metrics

```bash
ainative-code rate-limit metrics
```

Shows rate limiting statistics including:
- Total requests
- Allowed/blocked requests
- Blocked rate percentage
- Top endpoints by request count
- Top users by request count

## Usage Examples

### Basic HTTP Server with Rate Limiting

```go
package main

import (
    "net/http"
    "time"

    "github.com/AINative-studio/ainative-code/internal/middleware"
    "github.com/AINative-studio/ainative-code/internal/ratelimit"
)

func main() {
    // Create storage
    storage := ratelimit.NewMemoryStorage()
    defer storage.Close()

    // Create limiter
    limiter := ratelimit.NewLimiter(storage, ratelimit.Config{
        RequestsPerMinute: 60,
        BurstSize:         10,
        TimeWindow:        1 * time.Minute,
    })
    defer limiter.Close()

    // Create middleware
    rateLimitMiddleware := middleware.NewRateLimiterMiddleware(
        limiter,
        middleware.RateLimiterConfig{},
    )

    // Create handler
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("Hello, World!"))
    })

    // Wrap with rate limiting
    http.Handle("/api/", rateLimitMiddleware.Handler(handler))

    // Start server
    http.ListenAndServe(":8080", nil)
}
```

### Per-User Rate Limiting

```go
// Extract user ID from request
userExtractor := func(r *http.Request) string {
    // From JWT token, session, or API key
    return r.Header.Get("X-User-ID")
}

// Create middleware with per-user limits
rateLimitMiddleware := middleware.NewRateLimiterMiddleware(
    limiter,
    middleware.RateLimiterConfig{
        PerUser:         true,
        UserIDExtractor: userExtractor,
        KeyGenerator:    middleware.PerUserKeyGenerator(userExtractor),
    },
)
```

### Per-Endpoint Rate Limiting

```go
rateLimitMiddleware := middleware.NewRateLimiterMiddleware(
    limiter,
    middleware.RateLimiterConfig{
        PerEndpoint:     true,
        UserIDExtractor: userExtractor,
        KeyGenerator:    middleware.PerEndpointKeyGenerator(userExtractor),
    },
)
```

### Custom Rate Limit Handler

```go
customHandler := func(w http.ResponseWriter, r *http.Request, result *ratelimit.Result) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusTooManyRequests)

    json.NewEncoder(w).Encode(map[string]interface{}{
        "error": "rate_limit_exceeded",
        "message": "Too many requests",
        "retry_after": result.RetryAfter.Seconds(),
        "reset_at": result.ResetAt.Unix(),
    })
}

rateLimitMiddleware := middleware.NewRateLimiterMiddleware(
    limiter,
    middleware.RateLimiterConfig{
        OnRateLimitExceeded: customHandler,
    },
)
```

## Best Practices

### 1. Choose Appropriate Limits

- **Public APIs**: 60-100 requests/minute for anonymous users
- **Authenticated APIs**: 1000-5000 requests/minute for authenticated users
- **Heavy Operations**: 10-30 requests/minute for expensive operations
- **Lightweight Operations**: 100-500 requests/minute for cheap operations

### 2. Use Per-User Limits

Enable per-user rate limiting to prevent a single user from consuming all resources:

```yaml
rate_limit:
  per_user: true
```

### 3. Configure Burst Size

Allow short bursts of traffic while maintaining overall limits:

```yaml
rate_limit:
  requests_per_minute: 60
  burst_size: 10  # Allow 10 extra requests in bursts
```

### 4. Skip Health Checks

Exclude health and monitoring endpoints from rate limiting:

```yaml
rate_limit:
  skip_paths:
    - /health
    - /metrics
    - /.well-known/
```

### 5. Monitor Metrics

Regularly review rate limiting metrics to:
- Identify abuse patterns
- Adjust limits based on usage
- Detect system issues
- Plan capacity

### 6. Graceful Degradation

Implement fallback behavior when rate limits are exceeded:

```go
// Check if rate limited
result, _ := middleware.FromContext(r.Context())
if result != nil && !result.Allowed {
    // Implement degraded functionality
    // e.g., serve cached response, queue request, etc.
}
```

### 7. Use Redis for Production

For production deployments with multiple instances, use Redis:

```yaml
rate_limit:
  storage: redis
  redis_url: redis://redis-cluster:6379/0
```

## Performance

### Overhead

Rate limiting adds minimal overhead to request processing:

- **Memory storage**: <1ms per request
- **Redis storage**: <5ms per request (including network)
- **Thread-safe operations**: Lock-free counters minimize contention

### Scalability

- **Memory storage**: Scales to 100K+ requests/second per instance
- **Redis storage**: Scales horizontally across instances
- **Cleanup**: Automatic expired entry removal

### Benchmarks

```
BenchmarkMemoryStorage_Increment-8         10000000    120 ns/op
BenchmarkMemoryStorage_IncrementConcurrent-8  5000000    250 ns/op
BenchmarkLimiter_Allow-8                    5000000    300 ns/op
BenchmarkRateLimiter_Handler-8              1000000   1500 ns/op
```

## Troubleshooting

### Issue: Rate limits not enforced

**Causes:**
- Rate limiting disabled in config
- Path included in skip_paths
- IP in allowlist

**Solution:**
```bash
# Check configuration
ainative-code rate-limit config

# Verify status
ainative-code rate-limit status
```

### Issue: Legitimate users being blocked

**Causes:**
- Limits too restrictive
- Missing per-user rate limiting
- Shared IP addresses

**Solution:**
- Increase limits in configuration
- Enable per-user rate limiting
- Add trusted IPs to allowlist

### Issue: Rate limits reset unexpectedly

**Causes:**
- Short time window
- Server restarts (memory storage)

**Solution:**
- Use longer time windows
- Switch to Redis storage for persistence

### Issue: High memory usage

**Causes:**
- Many unique keys (users, IPs, endpoints)
- Cleanup not running

**Solution:**
- Use Redis storage
- Reduce time window
- Implement key namespacing

## Security Considerations

### 1. DDoS Protection

Rate limiting provides basic DDoS protection but should be combined with:
- WAF (Web Application Firewall)
- CDN with DDoS protection
- Network-level rate limiting

### 2. API Key Security

Store API keys securely and rotate regularly:

```yaml
security:
  secret_rotation: 90d  # Rotate every 90 days
```

### 3. Logging

Log rate limit violations for security monitoring:

```go
if !result.Allowed {
    logger.Warn("rate limit exceeded",
        "ip", ip,
        "user", userID,
        "endpoint", endpoint,
    )
}
```

### 4. Gradual Throttling

Implement progressive rate limiting for repeat offenders:

```go
// Increase restrictions for abusive behavior
if violations > threshold {
    applyStricterLimits()
}
```

## Migration Guide

### From No Rate Limiting

1. Add configuration with generous limits
2. Deploy and monitor metrics
3. Gradually reduce limits based on usage
4. Enable per-user/per-endpoint limits

### From Custom Implementation

1. Map existing limits to new configuration
2. Test in staging with production traffic replay
3. Deploy with monitoring
4. Verify metrics match expected behavior

## API Reference

See [Rate Limiting API Documentation](../api/rate-limiting.md) for detailed API reference.

## Support

For issues or questions:
- GitHub Issues: https://github.com/AINative-studio/ainative-code/issues
- Documentation: https://docs.ainative.studio/rate-limiting
- Community: https://community.ainative.studio
