package ratelimit

import (
	"sync"
	"sync/atomic"
	"time"
)

// Metrics tracks rate limiting statistics
type Metrics struct {
	mu sync.RWMutex

	// Total requests
	totalRequests uint64

	// Allowed requests
	allowedRequests uint64

	// Blocked requests
	blockedRequests uint64

	// Requests by endpoint
	endpointRequests map[string]uint64

	// Requests by user
	userRequests map[string]uint64

	// Last reset time
	lastReset time.Time
}

// NewMetrics creates a new metrics tracker
func NewMetrics() *Metrics {
	return &Metrics{
		endpointRequests: make(map[string]uint64),
		userRequests:     make(map[string]uint64),
		lastReset:        time.Now(),
	}
}

// RecordRequest records a request
func (m *Metrics) RecordRequest(allowed bool, endpoint string, user string) {
	atomic.AddUint64(&m.totalRequests, 1)

	if allowed {
		atomic.AddUint64(&m.allowedRequests, 1)
	} else {
		atomic.AddUint64(&m.blockedRequests, 1)
	}

	// Record by endpoint
	if endpoint != "" {
		m.mu.Lock()
		m.endpointRequests[endpoint]++
		m.mu.Unlock()
	}

	// Record by user
	if user != "" {
		m.mu.Lock()
		m.userRequests[user]++
		m.mu.Unlock()
	}
}

// GetStats returns current statistics
func (m *Metrics) GetStats() Stats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return Stats{
		TotalRequests:    atomic.LoadUint64(&m.totalRequests),
		AllowedRequests:  atomic.LoadUint64(&m.allowedRequests),
		BlockedRequests:  atomic.LoadUint64(&m.blockedRequests),
		EndpointRequests: m.copyMap(m.endpointRequests),
		UserRequests:     m.copyMap(m.userRequests),
		LastReset:        m.lastReset,
	}
}

// Reset resets all metrics
func (m *Metrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	atomic.StoreUint64(&m.totalRequests, 0)
	atomic.StoreUint64(&m.allowedRequests, 0)
	atomic.StoreUint64(&m.blockedRequests, 0)
	m.endpointRequests = make(map[string]uint64)
	m.userRequests = make(map[string]uint64)
	m.lastReset = time.Now()
}

// GetBlockedRate returns the percentage of blocked requests
func (m *Metrics) GetBlockedRate() float64 {
	total := atomic.LoadUint64(&m.totalRequests)
	if total == 0 {
		return 0
	}
	blocked := atomic.LoadUint64(&m.blockedRequests)
	return float64(blocked) / float64(total) * 100
}

// GetTopEndpoints returns the top N endpoints by request count
func (m *Metrics) GetTopEndpoints(n int) []EndpointStat {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Convert map to slice
	stats := make([]EndpointStat, 0, len(m.endpointRequests))
	for endpoint, count := range m.endpointRequests {
		stats = append(stats, EndpointStat{
			Endpoint: endpoint,
			Count:    count,
		})
	}

	// Sort by count (descending)
	for i := 0; i < len(stats); i++ {
		for j := i + 1; j < len(stats); j++ {
			if stats[j].Count > stats[i].Count {
				stats[i], stats[j] = stats[j], stats[i]
			}
		}
	}

	// Return top N
	if n > len(stats) {
		n = len(stats)
	}
	return stats[:n]
}

// GetTopUsers returns the top N users by request count
func (m *Metrics) GetTopUsers(n int) []UserStat {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Convert map to slice
	stats := make([]UserStat, 0, len(m.userRequests))
	for user, count := range m.userRequests {
		stats = append(stats, UserStat{
			User:  user,
			Count: count,
		})
	}

	// Sort by count (descending)
	for i := 0; i < len(stats); i++ {
		for j := i + 1; j < len(stats); j++ {
			if stats[j].Count > stats[i].Count {
				stats[i], stats[j] = stats[j], stats[i]
			}
		}
	}

	// Return top N
	if n > len(stats) {
		n = len(stats)
	}
	return stats[:n]
}

func (m *Metrics) copyMap(src map[string]uint64) map[string]uint64 {
	dst := make(map[string]uint64, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

// Stats contains rate limiting statistics
type Stats struct {
	TotalRequests    uint64
	AllowedRequests  uint64
	BlockedRequests  uint64
	EndpointRequests map[string]uint64
	UserRequests     map[string]uint64
	LastReset        time.Time
}

// EndpointStat contains statistics for a single endpoint
type EndpointStat struct {
	Endpoint string
	Count    uint64
}

// UserStat contains statistics for a single user
type UserStat struct {
	User  string
	Count uint64
}
