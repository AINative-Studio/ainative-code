package provider

import (
	"fmt"
	"sync"
)

// Registry manages registered LLM providers in a thread-safe manner
type Registry struct {
	mu        sync.RWMutex
	providers map[string]Provider
}

// NewRegistry creates a new provider registry
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]Provider),
	}
}

// Register adds a provider to the registry
// Returns an error if a provider with the same name is already registered
func (r *Registry) Register(name string, p Provider) error {
	if name == "" {
		return fmt.Errorf("provider name cannot be empty")
	}
	if p == nil {
		return fmt.Errorf("provider cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.providers[name]; exists {
		return fmt.Errorf("provider %q is already registered", name)
	}

	r.providers[name] = p
	return nil
}

// Get retrieves a provider by name
// Returns an error if the provider is not found
func (r *Registry) Get(name string) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, exists := r.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider %q not found", name)
	}

	return p, nil
}

// List returns the names of all registered providers
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
// Returns an error if the provider is not found
func (r *Registry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.providers[name]; !exists {
		return fmt.Errorf("provider %q not found", name)
	}

	delete(r.providers, name)
	return nil
}

// Close closes all registered providers and clears the registry
func (r *Registry) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var errs []error
	for name, p := range r.providers {
		if err := p.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close provider %q: %w", name, err))
		}
	}

	// Clear the registry
	r.providers = make(map[string]Provider)

	if len(errs) > 0 {
		return fmt.Errorf("errors closing providers: %v", errs)
	}

	return nil
}

// Count returns the number of registered providers
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.providers)
}

// Has checks if a provider with the given name is registered
func (r *Registry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.providers[name]
	return exists
}
