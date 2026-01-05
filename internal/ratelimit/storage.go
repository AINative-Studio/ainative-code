package ratelimit

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Storage defines the interface for rate limit storage backends
type Storage interface {
	// Get retrieves the current count and reset time for a key
	Get(ctx context.Context, key string) (count int64, resetAt time.Time, err error)

	// Increment increments the counter for a key and returns the new count
	Increment(ctx context.Context, key string, window time.Duration) (count int64, resetAt time.Time, err error)

	// Reset resets the counter for a key
	Reset(ctx context.Context, key string) error

	// Close closes the storage backend
	Close() error
}

// MemoryStorage implements in-memory rate limiting storage
type MemoryStorage struct {
	mu      sync.RWMutex
	entries map[string]*entry
	ticker  *time.Ticker
	done    chan struct{}
}

type entry struct {
	count   int64
	resetAt time.Time
}

// NewMemoryStorage creates a new in-memory storage backend
func NewMemoryStorage() *MemoryStorage {
	s := &MemoryStorage{
		entries: make(map[string]*entry),
		ticker:  time.NewTicker(1 * time.Minute),
		done:    make(chan struct{}),
	}

	// Start cleanup goroutine
	go s.cleanup()

	return s
}

// Get retrieves the current count and reset time for a key
func (s *MemoryStorage) Get(ctx context.Context, key string) (int64, time.Time, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	e, exists := s.entries[key]
	if !exists {
		return 0, time.Time{}, nil
	}

	// Check if the entry has expired
	if time.Now().After(e.resetAt) {
		return 0, time.Time{}, nil
	}

	return e.count, e.resetAt, nil
}

// Increment increments the counter for a key and returns the new count
func (s *MemoryStorage) Increment(ctx context.Context, key string, window time.Duration) (int64, time.Time, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	e, exists := s.entries[key]

	// Create new entry or reset if expired
	if !exists || now.After(e.resetAt) {
		resetAt := now.Add(window)
		s.entries[key] = &entry{
			count:   1,
			resetAt: resetAt,
		}
		return 1, resetAt, nil
	}

	// Increment existing entry
	e.count++
	return e.count, e.resetAt, nil
}

// Reset resets the counter for a key
func (s *MemoryStorage) Reset(ctx context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.entries, key)
	return nil
}

// Close closes the storage backend
func (s *MemoryStorage) Close() error {
	close(s.done)
	s.ticker.Stop()
	return nil
}

// cleanup removes expired entries periodically
func (s *MemoryStorage) cleanup() {
	for {
		select {
		case <-s.ticker.C:
			s.mu.Lock()
			now := time.Now()
			for key, e := range s.entries {
				if now.After(e.resetAt) {
					delete(s.entries, key)
				}
			}
			s.mu.Unlock()
		case <-s.done:
			return
		}
	}
}

// RedisStorage implements Redis-based rate limiting storage
type RedisStorage struct {
	// Redis client would be injected here
	// For now, this is a placeholder for distributed rate limiting
	prefix string
}

// NewRedisStorage creates a new Redis storage backend
func NewRedisStorage(url string, prefix string) (*RedisStorage, error) {
	// TODO: Implement Redis connection
	// This is a placeholder for future Redis integration
	return nil, fmt.Errorf("redis storage not yet implemented")
}

// Get retrieves the current count and reset time for a key
func (s *RedisStorage) Get(ctx context.Context, key string) (int64, time.Time, error) {
	return 0, time.Time{}, fmt.Errorf("redis storage not yet implemented")
}

// Increment increments the counter for a key and returns the new count
func (s *RedisStorage) Increment(ctx context.Context, key string, window time.Duration) (int64, time.Time, error) {
	return 0, time.Time{}, fmt.Errorf("redis storage not yet implemented")
}

// Reset resets the counter for a key
func (s *RedisStorage) Reset(ctx context.Context, key string) error {
	return fmt.Errorf("redis storage not yet implemented")
}

// Close closes the storage backend
func (s *RedisStorage) Close() error {
	return nil
}
