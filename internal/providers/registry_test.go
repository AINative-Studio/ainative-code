package providers

import (
	"context"
	"errors"
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

func (m *mockProvider) Chat(ctx context.Context, req *ChatRequest, opts ...Option) (*Response, error) {
	return &Response{Content: "mock response", Provider: m.name}, nil
}

func (m *mockProvider) Stream(ctx context.Context, req *StreamRequest, opts ...Option) (<-chan Event, error) {
	ch := make(chan Event)
	close(ch)
	return ch, nil
}

func (m *mockProvider) Name() string {
	return m.name
}

func (m *mockProvider) Models(ctx context.Context) ([]Model, error) {
	return []Model{{ID: "mock-model", Name: "Mock Model", Provider: m.name}}, nil
}

func (m *mockProvider) Close() error {
	m.closeMutex.Lock()
	defer m.closeMutex.Unlock()
	m.closed = true
	return m.closeError
}

// mockFactory creates a new mockProvider
func mockFactory(config Config) (Provider, error) {
	return &mockProvider{name: "mock-provider"}, nil
}

// failingFactory always returns an error
func failingFactory(config Config) (Provider, error) {
	return nil, errors.New("factory creation failed")
}

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()

	if registry == nil {
		t.Fatal("NewRegistry returned nil")
	}

	if registry.providers == nil {
		t.Error("registry.providers map is nil")
	}

	if registry.factories == nil {
		t.Error("registry.factories map is nil")
	}

	if len(registry.providers) != 0 {
		t.Errorf("new registry has %d providers, want 0", len(registry.providers))
	}

	if len(registry.factories) != 0 {
		t.Errorf("new registry has %d factories, want 0", len(registry.factories))
	}
}

func TestRegistry_RegisterFactory(t *testing.T) {
	registry := NewRegistry()

	// Test successful registration
	err := registry.RegisterFactory("test-provider", mockFactory)
	if err != nil {
		t.Errorf("RegisterFactory failed: %v", err)
	}

	// Verify factory was registered
	registry.mu.RLock()
	_, exists := registry.factories["test-provider"]
	registry.mu.RUnlock()

	if !exists {
		t.Error("factory was not registered")
	}
}

func TestRegistry_RegisterFactory_Duplicate(t *testing.T) {
	registry := NewRegistry()

	// Register factory first time
	err := registry.RegisterFactory("test-provider", mockFactory)
	if err != nil {
		t.Fatalf("First RegisterFactory failed: %v", err)
	}

	// Try to register again with same name
	err = registry.RegisterFactory("test-provider", mockFactory)
	if err == nil {
		t.Error("RegisterFactory should fail for duplicate name")
	}

	expectedErr := "factory for provider test-provider already registered"
	if err.Error() != expectedErr {
		t.Errorf("error = %v, want %v", err.Error(), expectedErr)
	}
}

func TestRegistry_Register(t *testing.T) {
	registry := NewRegistry()
	provider := &mockProvider{name: "test-provider"}

	// Test successful registration
	err := registry.Register("test-provider", provider)
	if err != nil {
		t.Errorf("Register failed: %v", err)
	}

	// Verify provider was registered
	registry.mu.RLock()
	_, exists := registry.providers["test-provider"]
	registry.mu.RUnlock()

	if !exists {
		t.Error("provider was not registered")
	}
}

func TestRegistry_Register_Duplicate(t *testing.T) {
	registry := NewRegistry()
	provider1 := &mockProvider{name: "test-provider-1"}
	provider2 := &mockProvider{name: "test-provider-2"}

	// Register provider first time
	err := registry.Register("test-provider", provider1)
	if err != nil {
		t.Fatalf("First Register failed: %v", err)
	}

	// Try to register again with same name
	err = registry.Register("test-provider", provider2)
	if err == nil {
		t.Error("Register should fail for duplicate name")
	}

	expectedErr := "provider test-provider already registered"
	if err.Error() != expectedErr {
		t.Errorf("error = %v, want %v", err.Error(), expectedErr)
	}
}

func TestRegistry_Get(t *testing.T) {
	registry := NewRegistry()
	provider := &mockProvider{name: "test-provider"}

	// Register a provider
	registry.Register("test-provider", provider)

	// Test successful retrieval
	retrieved, err := registry.Get("test-provider")
	if err != nil {
		t.Errorf("Get failed: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Get returned nil provider")
	}

	if retrieved.Name() != "test-provider" {
		t.Errorf("retrieved provider name = %v, want test-provider", retrieved.Name())
	}
}

func TestRegistry_Get_NotFound(t *testing.T) {
	registry := NewRegistry()

	// Try to get non-existent provider
	provider, err := registry.Get("non-existent")
	if err == nil {
		t.Error("Get should fail for non-existent provider")
	}

	if provider != nil {
		t.Error("Get should return nil provider for non-existent name")
	}

	expectedErr := "provider non-existent not found"
	if err.Error() != expectedErr {
		t.Errorf("error = %v, want %v", err.Error(), expectedErr)
	}
}

func TestRegistry_Create(t *testing.T) {
	registry := NewRegistry()

	// Register factory
	err := registry.RegisterFactory("test-provider", mockFactory)
	if err != nil {
		t.Fatalf("RegisterFactory failed: %v", err)
	}

	// Create provider using factory
	config := Config{APIKey: "test-key"}
	provider, err := registry.Create("test-provider", config)
	if err != nil {
		t.Errorf("Create failed: %v", err)
	}

	if provider == nil {
		t.Fatal("Create returned nil provider")
	}

	if provider.Name() != "mock-provider" {
		t.Errorf("provider name = %v, want mock-provider", provider.Name())
	}

	// Verify provider was auto-registered
	registered, err := registry.Get("test-provider")
	if err != nil {
		t.Errorf("provider was not auto-registered: %v", err)
	}

	if registered != provider {
		t.Error("auto-registered provider is not the same instance")
	}
}

func TestRegistry_Create_FactoryNotFound(t *testing.T) {
	registry := NewRegistry()

	// Try to create provider without registering factory
	config := Config{APIKey: "test-key"}
	provider, err := registry.Create("non-existent", config)
	if err == nil {
		t.Error("Create should fail when factory not found")
	}

	if provider != nil {
		t.Error("Create should return nil provider when factory not found")
	}

	expectedErr := "factory for provider non-existent not found"
	if err.Error() != expectedErr {
		t.Errorf("error = %v, want %v", err.Error(), expectedErr)
	}
}

func TestRegistry_Create_FactoryError(t *testing.T) {
	registry := NewRegistry()

	// Register failing factory
	err := registry.RegisterFactory("failing-provider", failingFactory)
	if err != nil {
		t.Fatalf("RegisterFactory failed: %v", err)
	}

	// Try to create provider with failing factory
	config := Config{APIKey: "test-key"}
	provider, err := registry.Create("failing-provider", config)
	if err == nil {
		t.Error("Create should fail when factory returns error")
	}

	if provider != nil {
		t.Error("Create should return nil provider when factory fails")
	}

	// Check that error message includes factory error
	if !contains(err.Error(), "factory creation failed") {
		t.Errorf("error should contain factory error, got: %v", err.Error())
	}
}

func TestRegistry_Create_RegistrationFailure(t *testing.T) {
	registry := NewRegistry()

	// Pre-register a provider
	existingProvider := &mockProvider{name: "existing"}
	registry.Register("test-provider", existingProvider)

	// Register factory with same name
	registry.RegisterFactory("test-provider", mockFactory)

	// Try to create - should fail because provider name already exists
	config := Config{APIKey: "test-key"}
	provider, err := registry.Create("test-provider", config)
	if err == nil {
		t.Error("Create should fail when provider name already registered")
	}

	if provider != nil {
		t.Error("Create should return nil when auto-registration fails")
	}
}

func TestRegistry_List(t *testing.T) {
	registry := NewRegistry()

	// Test empty list
	names := registry.List()
	if len(names) != 0 {
		t.Errorf("empty registry list has %d items, want 0", len(names))
	}

	// Register providers
	provider1 := &mockProvider{name: "provider1"}
	provider2 := &mockProvider{name: "provider2"}
	provider3 := &mockProvider{name: "provider3"}

	registry.Register("provider1", provider1)
	registry.Register("provider2", provider2)
	registry.Register("provider3", provider3)

	// Test list with multiple providers
	names = registry.List()
	if len(names) != 3 {
		t.Errorf("list has %d providers, want 3", len(names))
	}

	// Verify all names are present (order doesn't matter)
	nameSet := make(map[string]bool)
	for _, name := range names {
		nameSet[name] = true
	}

	if !nameSet["provider1"] || !nameSet["provider2"] || !nameSet["provider3"] {
		t.Errorf("list = %v, want all three providers", names)
	}
}

func TestRegistry_Unregister(t *testing.T) {
	registry := NewRegistry()
	provider := &mockProvider{name: "test-provider"}

	// Register provider
	registry.Register("test-provider", provider)

	// Verify it's registered
	_, err := registry.Get("test-provider")
	if err != nil {
		t.Fatalf("provider not registered: %v", err)
	}

	// Unregister provider
	err = registry.Unregister("test-provider")
	if err != nil {
		t.Errorf("Unregister failed: %v", err)
	}

	// Verify provider was closed
	if !provider.closed {
		t.Error("provider was not closed during unregister")
	}

	// Verify provider is no longer registered
	_, err = registry.Get("test-provider")
	if err == nil {
		t.Error("provider should not be found after unregister")
	}
}

func TestRegistry_Unregister_NotFound(t *testing.T) {
	registry := NewRegistry()

	// Try to unregister non-existent provider
	err := registry.Unregister("non-existent")
	if err == nil {
		t.Error("Unregister should fail for non-existent provider")
	}

	expectedErr := "provider non-existent not found"
	if err.Error() != expectedErr {
		t.Errorf("error = %v, want %v", err.Error(), expectedErr)
	}
}

func TestRegistry_Unregister_CloseError(t *testing.T) {
	registry := NewRegistry()
	provider := &mockProvider{
		name:       "test-provider",
		closeError: errors.New("close failed"),
	}

	// Register provider
	registry.Register("test-provider", provider)

	// Try to unregister - should fail due to close error
	err := registry.Unregister("test-provider")
	if err == nil {
		t.Error("Unregister should fail when Close returns error")
	}

	if !contains(err.Error(), "close failed") {
		t.Errorf("error should contain close error, got: %v", err.Error())
	}
}

func TestRegistry_Close(t *testing.T) {
	registry := NewRegistry()

	// Register multiple providers
	provider1 := &mockProvider{name: "provider1"}
	provider2 := &mockProvider{name: "provider2"}
	provider3 := &mockProvider{name: "provider3"}

	registry.Register("provider1", provider1)
	registry.Register("provider2", provider2)
	registry.Register("provider3", provider3)

	// Close registry
	err := registry.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// Verify all providers were closed
	if !provider1.closed {
		t.Error("provider1 was not closed")
	}
	if !provider2.closed {
		t.Error("provider2 was not closed")
	}
	if !provider3.closed {
		t.Error("provider3 was not closed")
	}

	// Verify registry is empty
	names := registry.List()
	if len(names) != 0 {
		t.Errorf("registry has %d providers after Close, want 0", len(names))
	}
}

func TestRegistry_Close_WithErrors(t *testing.T) {
	registry := NewRegistry()

	// Register providers with close errors
	provider1 := &mockProvider{name: "provider1", closeError: errors.New("error1")}
	provider2 := &mockProvider{name: "provider2"}
	provider3 := &mockProvider{name: "provider3", closeError: errors.New("error3")}

	registry.Register("provider1", provider1)
	registry.Register("provider2", provider2)
	registry.Register("provider3", provider3)

	// Close registry
	err := registry.Close()
	if err == nil {
		t.Error("Close should return error when provider Close fails")
	}

	// Verify error message contains information about failures
	if !contains(err.Error(), "errors closing providers") {
		t.Errorf("error should mention closing failures, got: %v", err.Error())
	}

	// Verify all providers were attempted to close
	if !provider1.closed {
		t.Error("provider1 was not closed despite error")
	}
	if !provider2.closed {
		t.Error("provider2 was not closed")
	}
	if !provider3.closed {
		t.Error("provider3 was not closed despite error")
	}

	// Verify registry is empty even with errors
	names := registry.List()
	if len(names) != 0 {
		t.Errorf("registry has %d providers after Close, want 0", len(names))
	}
}

func TestRegistry_Close_Empty(t *testing.T) {
	registry := NewRegistry()

	// Close empty registry
	err := registry.Close()
	if err != nil {
		t.Errorf("Close on empty registry failed: %v", err)
	}
}

func TestGetGlobalRegistry(t *testing.T) {
	registry := GetGlobalRegistry()

	if registry == nil {
		t.Fatal("GetGlobalRegistry returned nil")
	}

	// Verify it returns the same instance
	registry2 := GetGlobalRegistry()
	if registry != registry2 {
		t.Error("GetGlobalRegistry returned different instances")
	}
}

func TestRegistry_ThreadSafety(t *testing.T) {
	registry := NewRegistry()
	var wg sync.WaitGroup

	// Concurrent registrations
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			provider := &mockProvider{name: "provider"}
			registry.Register(string(rune('a'+index)), provider)
		}(i)
	}

	// Concurrent list operations
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			registry.List()
		}()
	}

	wg.Wait()

	// Verify no race conditions occurred and some providers were registered
	names := registry.List()
	if len(names) == 0 {
		t.Error("no providers were registered in concurrent test")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
