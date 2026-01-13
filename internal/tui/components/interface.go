package components

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Component is the base interface that all TUI components must implement.
// It follows the Bubble Tea pattern of Init, Update, and View methods.
type Component interface {
	// Init initializes the component and returns an initial command
	Init() tea.Cmd

	// Update handles messages and updates component state
	Update(msg tea.Msg) (Component, tea.Cmd)

	// View renders the component as a string
	View() string
}

// Sizeable represents components that can be resized.
// Components implementing this interface can adapt to terminal size changes.
type Sizeable interface {
	Component

	// SetSize updates the component's dimensions
	SetSize(width, height int)

	// GetSize returns the current component dimensions
	GetSize() (width, height int)
}

// Focusable represents components that can receive keyboard focus.
// This is useful for input components, text areas, and interactive elements.
type Focusable interface {
	Component

	// Focus gives keyboard focus to the component
	Focus() tea.Cmd

	// Blur removes keyboard focus from the component
	Blur()

	// Focused returns true if the component currently has focus
	Focused() bool
}

// Stateful represents components with visibility and state management.
// This interface is particularly useful for overlays, popups, and toggleable components.
type Stateful interface {
	Component

	// Show makes the component visible
	Show()

	// Hide makes the component invisible
	Hide()

	// Toggle switches between visible and invisible states
	Toggle()

	// IsVisible returns true if the component is currently visible
	IsVisible() bool
}

// Scrollable represents components that support scrolling.
// This is useful for large content areas like viewports and lists.
type Scrollable interface {
	Component

	// ScrollUp scrolls the content up by the specified number of lines
	ScrollUp(lines int)

	// ScrollDown scrolls the content down by the specified number of lines
	ScrollDown(lines int)

	// ScrollToTop scrolls to the top of the content
	ScrollToTop()

	// ScrollToBottom scrolls to the bottom of the content
	ScrollToBottom()

	// GetScrollPosition returns the current scroll position
	GetScrollPosition() int

	// SetScrollPosition sets the scroll position
	SetScrollPosition(position int)
}

// Selectable represents components with selectable items.
// This is useful for lists, menus, and completion popups.
type Selectable interface {
	Component

	// SelectNext moves selection to the next item
	SelectNext()

	// SelectPrevious moves selection to the previous item
	SelectPrevious()

	// GetSelectedIndex returns the index of the currently selected item
	GetSelectedIndex() int

	// SetSelectedIndex sets the selected item by index
	SetSelectedIndex(index int)

	// GetSelectedValue returns the value of the currently selected item
	GetSelectedValue() interface{}
}

// Themeable represents components that support theming.
// This allows components to adapt their appearance based on theme settings.
type Themeable interface {
	Component

	// SetTheme applies a theme to the component
	SetTheme(theme Theme)

	// GetTheme returns the current theme
	GetTheme() Theme
}

// Theme represents a visual theme for components.
type Theme struct {
	Name            string
	PrimaryColor    string
	SecondaryColor  string
	BackgroundColor string
	TextColor       string
	BorderColor     string
	AccentColor     string
}

// Validatable represents components that can validate their content.
// This is useful for form inputs and data entry components.
type Validatable interface {
	Component

	// Validate checks if the component's content is valid
	Validate() error

	// IsValid returns true if the component's content is valid
	IsValid() bool

	// GetValidationError returns the current validation error, if any
	GetValidationError() error
}

// Configurable represents components with configuration options.
// This allows components to be customized after initialization.
type Configurable interface {
	Component

	// SetConfig applies configuration to the component
	SetConfig(config interface{}) error

	// GetConfig returns the current configuration
	GetConfig() interface{}
}

// Animatable represents components that support animations.
// This is useful for loading indicators, transitions, and visual effects.
type Animatable interface {
	Component

	// StartAnimation starts the animation
	StartAnimation() tea.Cmd

	// StopAnimation stops the animation
	StopAnimation()

	// IsAnimating returns true if the animation is running
	IsAnimating() bool

	// UpdateAnimation updates the animation state based on tick
	UpdateAnimation(tick int)
}

// Disposable represents components that need cleanup when destroyed.
// This ensures proper resource management and prevents memory leaks.
type Disposable interface {
	Component

	// Dispose cleans up component resources
	Dispose() error

	// IsDisposed returns true if the component has been disposed
	IsDisposed() bool
}

// Eventable represents components that can emit custom events.
// This enables pub-sub patterns and component communication.
type Eventable interface {
	Component

	// On registers an event handler for the specified event type
	On(eventType string, handler func(interface{}))

	// Off removes an event handler for the specified event type
	Off(eventType string)

	// Emit triggers an event with the specified type and data
	Emit(eventType string, data interface{})
}

// Cloneable represents components that can be cloned.
// This is useful for creating copies of components without affecting the original.
type Cloneable interface {
	Component

	// Clone creates a deep copy of the component
	Clone() Component
}

// Serializable represents components that can be serialized.
// This is useful for state persistence and debugging.
type Serializable interface {
	Component

	// Serialize converts the component state to a byte slice
	Serialize() ([]byte, error)

	// Deserialize restores component state from a byte slice
	Deserialize(data []byte) error
}
