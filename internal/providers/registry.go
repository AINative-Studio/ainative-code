package providers

import (
	"fmt"
	"sync"
)

// Registry manages multiple provider instances
type Registry struct {
	mu        sync.RWMutex
	providers map[string]Provider
	factories map[string]ProviderFactory
}

// NewRegistry creates a new provider registry
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]Provider),
		factories: make(map[string]ProviderFactory),
	}
}

// RegisterFactory registers a provider factory function
func (r *Registry) RegisterFactory(name string, factory ProviderFactory) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.factories[name]; exists {
		return fmt.Errorf("factory for provider %s already registered", name)
	}

	r.factories[name] = factory
	return nil
}

// Register registers a provider instance
func (r *Registry) Register(name string, provider Provider) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.providers[name]; exists {
		return fmt.Errorf("provider %s already registered", name)
	}

	r.providers[name] = provider
	return nil
}

// Get retrieves a provider by name
func (r *Registry) Get(name string) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, exists := r.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", name)
	}

	return provider, nil
}

// Create creates a new provider instance using a registered factory
func (r *Registry) Create(name string, config Config) (Provider, error) {
	r.mu.RLock()
	factory, exists := r.factories[name]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("factory for provider %s not found", name)
	}

	provider, err := factory(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider %s: %w", name, err)
	}

	// Auto-register the created provider
	if err := r.Register(name, provider); err != nil {
		// If registration fails, close the provider and return error
		provider.Close()
		return nil, err
	}

	return provider, nil
}

// List returns all registered provider names
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}

	return names
}

// Unregister removes a provider from the registry
func (r *Registry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	provider, exists := r.providers[name]
	if !exists {
		return fmt.Errorf("provider %s not found", name)
	}

	// Close the provider before unregistering
	if err := provider.Close(); err != nil {
		return fmt.Errorf("failed to close provider %s: %w", name, err)
	}

	delete(r.providers, name)
	return nil
}

// Close closes all registered providers
func (r *Registry) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var errs []error
	for name, provider := range r.providers {
		if err := provider.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close provider %s: %w", name, err))
		}
	}

	// Clear all providers after closing
	r.providers = make(map[string]Provider)

	if len(errs) > 0 {
		return fmt.Errorf("errors closing providers: %v", errs)
	}

	return nil
}

// Global registry instance
var globalRegistry = NewRegistry()

// GetGlobalRegistry returns the global registry instance
func GetGlobalRegistry() *Registry {
	return globalRegistry
}
