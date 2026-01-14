package layout

import (
	"fmt"
	"sort"
)

// LayoutManager manages component positioning within a container
type LayoutManager interface {
	// RegisterComponent registers a component with its constraints
	RegisterComponent(id string, constraints Constraints) error
	// UnregisterComponent removes a component from management
	UnregisterComponent(id string) error
	// SetAvailableSize sets the container size available for layout
	SetAvailableSize(width, height int)
	// GetComponentBounds returns the calculated bounds for a component
	GetComponentBounds(id string) Rectangle
	// RecalculateLayout recalculates all component positions
	RecalculateLayout() error
	// IsDirty returns whether the layout needs recalculation
	IsDirty() bool
}

// Layout defines how components are arranged within a container
type Layout interface {
	// Calculate computes component bounds based on available space and constraints
	Calculate(available Rectangle, components []ComponentInfo) map[string]Rectangle
}

// ComponentInfo holds component metadata for layout calculation
type ComponentInfo struct {
	ID          string
	Constraints Constraints
	Order       int // Registration order for stable sorting
}

// Manager is the default implementation of LayoutManager
type Manager struct {
	components map[string]ComponentInfo
	bounds     map[string]Rectangle
	layout     Layout
	width      int
	height     int
	dirty      bool
	order      int // Counter for registration order
}

// NewManager creates a new layout manager with the specified layout algorithm
func NewManager(layout Layout) *Manager {
	return &Manager{
		components: make(map[string]ComponentInfo),
		bounds:     make(map[string]Rectangle),
		layout:     layout,
		dirty:      true,
		order:      0,
	}
}

// RegisterComponent registers a component with its layout constraints
func (m *Manager) RegisterComponent(id string, constraints Constraints) error {
	if id == "" {
		return fmt.Errorf("component id cannot be empty")
	}

	// Store component with registration order for stable sorting
	m.components[id] = ComponentInfo{
		ID:          id,
		Constraints: constraints,
		Order:       m.order,
	}
	m.order++
	m.dirty = true

	return nil
}

// UnregisterComponent removes a component from layout management
func (m *Manager) UnregisterComponent(id string) error {
	if _, exists := m.components[id]; !exists {
		return fmt.Errorf("component %q not found", id)
	}

	delete(m.components, id)
	delete(m.bounds, id)
	m.dirty = true

	return nil
}

// SetAvailableSize updates the container size and marks layout as dirty
func (m *Manager) SetAvailableSize(width, height int) {
	if m.width != width || m.height != height {
		m.width = width
		m.height = height
		m.dirty = true
	}
}

// GetComponentBounds returns the calculated bounds for a component
func (m *Manager) GetComponentBounds(id string) Rectangle {
	if m.dirty {
		_ = m.RecalculateLayout()
	}

	bounds, exists := m.bounds[id]
	if !exists {
		// Return zero rectangle if component not found
		return Rectangle{X: 0, Y: 0, Width: 0, Height: 0}
	}

	return bounds
}

// RecalculateLayout recalculates all component positions using the layout algorithm
func (m *Manager) RecalculateLayout() error {
	if !m.dirty {
		return nil
	}

	// Create available rectangle
	available := Rectangle{
		X:      0,
		Y:      0,
		Width:  m.width,
		Height: m.height,
	}

	// Convert map to sorted slice for deterministic ordering
	componentList := make([]ComponentInfo, 0, len(m.components))
	for _, comp := range m.components {
		componentList = append(componentList, comp)
	}

	// Sort by registration order for stable layout
	sort.Slice(componentList, func(i, j int) bool {
		return componentList[i].Order < componentList[j].Order
	})

	// Calculate layout
	m.bounds = m.layout.Calculate(available, componentList)

	m.dirty = false
	return nil
}

// IsDirty returns whether the layout needs recalculation
func (m *Manager) IsDirty() bool {
	return m.dirty
}

// GetComponentCount returns the number of registered components
func (m *Manager) GetComponentCount() int {
	return len(m.components)
}

// Clear removes all components and resets the manager
func (m *Manager) Clear() {
	m.components = make(map[string]ComponentInfo)
	m.bounds = make(map[string]Rectangle)
	m.dirty = true
	m.order = 0
}

// SetLayout changes the layout algorithm and marks as dirty
func (m *Manager) SetLayout(layout Layout) {
	m.layout = layout
	m.dirty = true
}
