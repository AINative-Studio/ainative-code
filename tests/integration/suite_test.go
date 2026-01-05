// +build integration

// Package integration provides integration tests for AINative-Code.
//
// These tests verify critical user workflows across multiple components:
//   - OAuth login flow
//   - Session persistence and resume
//   - ZeroDB operations
//   - Design token extraction
//   - Strapi content management
//   - RLHF feedback submission
//
// Integration tests use real database instances (in-memory SQLite) and
// mock HTTP servers for external APIs. They verify end-to-end functionality
// without requiring actual external service dependencies.
//
// Running Integration Tests:
//   - Run all: go test -v -tags=integration ./tests/integration/...
//   - Run specific suite: go test -v -tags=integration ./tests/integration -run TestSessionIntegrationTestSuite
//   - With coverage: go test -v -tags=integration -coverprofile=coverage.out ./tests/integration/...
//
// Test Environment:
//   - Uses in-memory SQLite databases for fast execution
//   - Mock HTTP servers for external API calls
//   - Parallel execution where safe
//   - Automatic cleanup after each test
//
// Performance Requirements:
//   - Total suite runtime: < 5 minutes
//   - Individual test timeout: 30 seconds
//   - Code coverage: >= 80%
package integration

import (
	"testing"
	"time"
)

// TestMain is the entry point for integration tests.
// It can be used for global setup/teardown if needed.
func TestMain(m *testing.M) {
	// Set test timeout
	// Note: Individual tests should have their own timeouts
	// This is a global failsafe
	// No global setup needed currently - each suite manages its own resources

	// Run tests
	m.Run()
}

// Helper function to enforce test timeouts
func withTimeout(t *testing.T, timeout time.Duration, fn func()) {
	t.Helper()

	done := make(chan bool)
	go func() {
		fn()
		done <- true
	}()

	select {
	case <-done:
		// Test completed successfully
	case <-time.After(timeout):
		t.Fatal("Test exceeded timeout of", timeout)
	}
}
