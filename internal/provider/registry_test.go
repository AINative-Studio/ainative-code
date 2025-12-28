package provider

import (
	"context"
	"sync"
	"testing"
)

// mockProvider implements the Provider interface for testing
type mockProvider struct {
	name       string
	closed     bool
	closeMutex sync.Mutex
	closeError error
}

func (m *mockProvider) Name() string {
	return m.name
}

func (m *mockProvider) Models() []string {
	return []string{"mock-model-1", "mock-model-2"}
}

func (m *mockProvider) Chat(ctx context.Context, messages []Message, opts ...ChatOption) (Response, error) {
	return Response{
		Content: "mock response",
		Model:   "mock-model",
		Usage: Usage{
			PromptTokens:     10,
			CompletionTokens: 20,
			TotalTokens:      30,
		},
	}, nil
}

func (m *mockProvider) Stream(ctx context.Context, messages []Message, opts ...StreamOption) (<-chan Event, error) {
	ch := make(chan Event, 1)
	ch <- Event{
		Type:    EventTypeContentDelta,
		Content: "mock stream",
	}
	close(ch)
	return ch, nil
}

func (m *mockProvider) Close() error {
	m.closeMutex.Lock()
	defer m.closeMutex.Unlock()
	m.closed = true
	return m.closeError
}

func (m *mockProvider) IsClosed() bool {
	m.closeMutex.Lock()
	defer m.closeMutex.Unlock()
	return m.closed
}

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()

	if registry == nil {
		t.Fatal("NewRegistry returned nil")
	}

	if registry.Count() != 0 {
		t.Errorf("NewRegistry should create empty registry, got count: %d", registry.Count())
	}

	if registry.providers == nil {
		t.Error("NewRegistry should initialize providers map")
	}
}

func TestRegistry_Register(t *testing.T) {
	tests := []struct {
		name          string
		providerName  string
		provider      Provider
		expectError   bool
		errorContains string
	}{
		{
			name:         "valid registration",
			providerName: "test-provider",
			provider:     &mockProvider{name: "test-provider"},
			expectError:  false,
		},
		{
			name:          "empty name",
			providerName:  "",
			provider:      &mockProvider{name: "test"},
			expectError:   true,
			errorContains: "cannot be empty",
		},
		{
			name:          "nil provider",
			providerName:  "test",
			provider:      nil,
			expectError:   true,
			errorContains: "cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := NewRegistry()
			err := registry.Register(tt.providerName, tt.provider)

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				} else if tt.errorContains != "" && !containsString(err.Error(), tt.errorContains) {
					t.Errorf("expected error to contain %q, got: %v", tt.errorContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if registry.Count() != 1 {
					t.Errorf("expected count 1, got: %d", registry.Count())
				}
			}
		})
	}
}

func TestRegistry_Register_Duplicate(t *testing.T) {
	registry := NewRegistry()
	provider1 := &mockProvider{name: "test-provider"}
	provider2 := &mockProvider{name: "test-provider"}

	// First registration should succeed
	err := registry.Register("test-provider", provider1)
	if err != nil {
		t.Fatalf("first registration failed: %v", err)
	}

	// Second registration with same name should fail
	err = registry.Register("test-provider", provider2)
	if err == nil {
		t.Error("expected error for duplicate registration, got nil")
	}
	if !containsString(err.Error(), "already registered") {
		t.Errorf("expected 'already registered' error, got: %v", err)
	}

	// Count should still be 1
	if registry.Count() != 1 {
		t.Errorf("expected count 1, got: %d", registry.Count())
	}
}

func TestRegistry_Get(t *testing.T) {
	registry := NewRegistry()
	provider := &mockProvider{name: "test-provider"}

	// Register provider
	err := registry.Register("test-provider", provider)
	if err != nil {
		t.Fatalf("failed to register provider: %v", err)
	}

	// Get existing provider
	retrieved, err := registry.Get("test-provider")
	if err != nil {
		t.Errorf("failed to get provider: %v", err)
	}
	if retrieved != provider {
		t.Error("retrieved provider does not match registered provider")
	}

	// Get non-existent provider
	_, err = registry.Get("non-existent")
	if err == nil {
		t.Error("expected error for non-existent provider, got nil")
	}
	if !containsString(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

func TestRegistry_List(t *testing.T) {
	registry := NewRegistry()

	// Empty registry
	names := registry.List()
	if len(names) != 0 {
		t.Errorf("expected empty list, got: %v", names)
	}

	// Register multiple providers
	providers := map[string]*mockProvider{
		"provider1": {name: "provider1"},
		"provider2": {name: "provider2"},
		"provider3": {name: "provider3"},
	}

	for name, provider := range providers {
		err := registry.Register(name, provider)
		if err != nil {
			t.Fatalf("failed to register %s: %v", name, err)
		}
	}

	// Get list
	names = registry.List()
	if len(names) != 3 {
		t.Errorf("expected 3 providers, got: %d", len(names))
	}

	// Verify all names are present
	nameMap := make(map[string]bool)
	for _, name := range names {
		nameMap[name] = true
	}

	for expectedName := range providers {
		if !nameMap[expectedName] {
			t.Errorf("expected provider %q in list, not found", expectedName)
		}
	}
}

func TestRegistry_Unregister(t *testing.T) {
	registry := NewRegistry()
	provider := &mockProvider{name: "test-provider"}

	// Register provider
	err := registry.Register("test-provider", provider)
	if err != nil {
		t.Fatalf("failed to register provider: %v", err)
	}

	// Unregister existing provider
	err = registry.Unregister("test-provider")
	if err != nil {
		t.Errorf("failed to unregister provider: %v", err)
	}

	if registry.Count() != 0 {
		t.Errorf("expected count 0 after unregister, got: %d", registry.Count())
	}

	// Unregister non-existent provider
	err = registry.Unregister("non-existent")
	if err == nil {
		t.Error("expected error for non-existent provider, got nil")
	}
	if !containsString(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

func TestRegistry_Close(t *testing.T) {
	registry := NewRegistry()

	// Create providers
	provider1 := &mockProvider{name: "provider1"}
	provider2 := &mockProvider{name: "provider2"}

	// Register providers
	registry.Register("provider1", provider1)
	registry.Register("provider2", provider2)

	// Close registry
	err := registry.Close()
	if err != nil {
		t.Errorf("unexpected error closing registry: %v", err)
	}

	// Verify providers were closed
	if !provider1.IsClosed() {
		t.Error("provider1 was not closed")
	}
	if !provider2.IsClosed() {
		t.Error("provider2 was not closed")
	}

	// Verify registry was cleared
	if registry.Count() != 0 {
		t.Errorf("expected count 0 after close, got: %d", registry.Count())
	}

	// Verify registry is empty
	names := registry.List()
	if len(names) != 0 {
		t.Errorf("expected empty list after close, got: %v", names)
	}
}

func TestRegistry_Close_WithErrors(t *testing.T) {
	registry := NewRegistry()

	// Create provider that returns error on close
	provider1 := &mockProvider{
		name:       "provider1",
		closeError: context.DeadlineExceeded,
	}
	provider2 := &mockProvider{name: "provider2"}

	// Register providers
	registry.Register("provider1", provider1)
	registry.Register("provider2", provider2)

	// Close registry
	err := registry.Close()
	if err == nil {
		t.Error("expected error when closing provider fails")
	}
	if !containsString(err.Error(), "errors closing providers") {
		t.Errorf("expected 'errors closing providers' message, got: %v", err)
	}

	// Registry should still be cleared even with errors
	if registry.Count() != 0 {
		t.Errorf("expected count 0 after close, got: %d", registry.Count())
	}
}

func TestRegistry_Count(t *testing.T) {
	registry := NewRegistry()

	// Initial count
	if registry.Count() != 0 {
		t.Errorf("expected count 0, got: %d", registry.Count())
	}

	// Add providers
	for i := 0; i < 5; i++ {
		name := string(rune('a' + i))
		provider := &mockProvider{name: name}
		registry.Register(name, provider)

		expected := i + 1
		if registry.Count() != expected {
			t.Errorf("after %d registrations, expected count %d, got: %d", expected, expected, registry.Count())
		}
	}

	// Remove providers
	for i := 0; i < 5; i++ {
		name := string(rune('a' + i))
		registry.Unregister(name)

		expected := 4 - i
		if registry.Count() != expected {
			t.Errorf("after %d unregistrations, expected count %d, got: %d", i+1, expected, registry.Count())
		}
	}
}

func TestRegistry_Has(t *testing.T) {
	registry := NewRegistry()
	provider := &mockProvider{name: "test-provider"}

	// Check non-existent provider
	if registry.Has("test-provider") {
		t.Error("Has returned true for non-existent provider")
	}

	// Register provider
	registry.Register("test-provider", provider)

	// Check existing provider
	if !registry.Has("test-provider") {
		t.Error("Has returned false for existing provider")
	}

	// Unregister provider
	registry.Unregister("test-provider")

	// Check after unregister
	if registry.Has("test-provider") {
		t.Error("Has returned true for unregistered provider")
	}
}

func TestRegistry_ConcurrentAccess(t *testing.T) {
	registry := NewRegistry()
	numGoroutines := 100
	numProvidersPerGoroutine := 10

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 3) // Register, Get, and List operations

	// Concurrent registrations
	for i := 0; i < numGoroutines; i++ {
		go func(base int) {
			defer wg.Done()
			for j := 0; j < numProvidersPerGoroutine; j++ {
				name := string(rune('a'+base)) + string(rune('0'+j))
				provider := &mockProvider{name: name}
				registry.Register(name, provider)
			}
		}(i)
	}

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		go func(base int) {
			defer wg.Done()
			for j := 0; j < numProvidersPerGoroutine; j++ {
				name := string(rune('a'+base)) + string(rune('0'+j))
				registry.Get(name) // Ignore errors as provider may not exist yet
			}
		}(i)
	}

	// Concurrent list operations
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			registry.List()
		}()
	}

	wg.Wait()

	// Verify final state - should have registered providers
	// (may have duplicates due to race, but that's okay for this test)
	if registry.Count() == 0 {
		t.Error("expected some providers to be registered")
	}
}

func TestRegistry_ConcurrentRegisterUnregister(t *testing.T) {
	registry := NewRegistry()
	numOperations := 100

	var wg sync.WaitGroup
	wg.Add(numOperations * 2)

	// Concurrent register/unregister of same provider
	providerName := "test-provider"

	for i := 0; i < numOperations; i++ {
		// Register
		go func() {
			defer wg.Done()
			provider := &mockProvider{name: providerName}
			registry.Register(providerName, provider)
		}()

		// Unregister
		go func() {
			defer wg.Done()
			registry.Unregister(providerName)
		}()
	}

	wg.Wait()

	// Final state is indeterminate, but operations should not panic
	// Just verify registry is still functional
	testProvider := &mockProvider{name: "final-test"}
	err := registry.Register("final-test", testProvider)
	if err != nil {
		t.Errorf("registry not functional after concurrent operations: %v", err)
	}
}

func TestRegistry_ConcurrentClose(t *testing.T) {
	registry := NewRegistry()

	// Register some providers
	for i := 0; i < 10; i++ {
		name := string(rune('a' + i))
		provider := &mockProvider{name: name}
		registry.Register(name, provider)
	}

	var wg sync.WaitGroup
	numCloseAttempts := 10
	wg.Add(numCloseAttempts)

	// Attempt to close registry concurrently
	for i := 0; i < numCloseAttempts; i++ {
		go func() {
			defer wg.Done()
			registry.Close()
		}()
	}

	wg.Wait()

	// Registry should be empty after close
	if registry.Count() != 0 {
		t.Errorf("expected count 0 after close, got: %d", registry.Count())
	}
}

func TestRegistry_GetAfterClose(t *testing.T) {
	registry := NewRegistry()
	provider := &mockProvider{name: "test-provider"}

	// Register provider
	registry.Register("test-provider", provider)

	// Close registry
	registry.Close()

	// Try to get provider after close
	_, err := registry.Get("test-provider")
	if err == nil {
		t.Error("expected error getting provider after close")
	}
	if !containsString(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

func TestRegistry_RegisterAfterClose(t *testing.T) {
	registry := NewRegistry()

	// Close empty registry
	registry.Close()

	// Try to register provider after close
	provider := &mockProvider{name: "test-provider"}
	err := registry.Register("test-provider", provider)
	if err != nil {
		t.Errorf("registration after close should work (registry was cleared): %v", err)
	}

	// Verify provider was registered
	if !registry.Has("test-provider") {
		t.Error("provider should be registered after close and re-register")
	}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())
}
