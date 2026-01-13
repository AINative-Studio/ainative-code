package components

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Lifecycle represents the complete lifecycle of a component.
// Components implementing this interface can hook into various lifecycle events
// for initialization, mounting, updating, and cleanup operations.
type Lifecycle interface {
	Component

	// OnInit is called when the component is first initialized.
	// Use this for setting up initial state and configuration.
	OnInit() tea.Cmd

	// OnMount is called when the component is mounted to the DOM/UI tree.
	// Use this for operations that require the component to be rendered.
	OnMount() tea.Cmd

	// OnUnmount is called when the component is about to be removed.
	// Use this for cleanup operations like closing connections or timers.
	OnUnmount() tea.Cmd

	// OnBeforeUpdate is called before the component processes an update.
	// Return false to cancel the update.
	OnBeforeUpdate(msg tea.Msg) bool

	// OnAfterUpdate is called after the component has processed an update.
	// Use this for side effects that depend on the updated state.
	OnAfterUpdate(msg tea.Msg) tea.Cmd

	// OnShow is called when the component becomes visible.
	OnShow() tea.Cmd

	// OnHide is called when the component becomes hidden.
	OnHide() tea.Cmd

	// OnResize is called when the component's size changes.
	OnResize(width, height int) tea.Cmd

	// OnFocus is called when the component receives focus.
	OnFocus() tea.Cmd

	// OnBlur is called when the component loses focus.
	OnBlur() tea.Cmd

	// GetLifecycleState returns the current lifecycle state.
	GetLifecycleState() LifecycleState

	// IsInitialized returns true if the component has been initialized.
	IsInitialized() bool

	// IsMounted returns true if the component is currently mounted.
	IsMounted() bool
}

// LifecycleState represents the current state in a component's lifecycle.
type LifecycleState int

const (
	// StateUninitialized indicates the component has not been initialized
	StateUninitialized LifecycleState = iota

	// StateInitializing indicates the component is being initialized
	StateInitializing

	// StateInitialized indicates the component has been initialized
	StateInitialized

	// StateMounting indicates the component is being mounted
	StateMounting

	// StateMounted indicates the component is mounted and active
	StateMounted

	// StateUpdating indicates the component is processing an update
	StateUpdating

	// StateUnmounting indicates the component is being unmounted
	StateUnmounting

	// StateUnmounted indicates the component has been unmounted
	StateUnmounted

	// StateError indicates the component encountered an error
	StateError

	// StateDisposed indicates the component has been disposed
	StateDisposed
)

// String returns the string representation of a LifecycleState.
func (s LifecycleState) String() string {
	switch s {
	case StateUninitialized:
		return "uninitialized"
	case StateInitializing:
		return "initializing"
	case StateInitialized:
		return "initialized"
	case StateMounting:
		return "mounting"
	case StateMounted:
		return "mounted"
	case StateUpdating:
		return "updating"
	case StateUnmounting:
		return "unmounting"
	case StateUnmounted:
		return "unmounted"
	case StateError:
		return "error"
	case StateDisposed:
		return "disposed"
	default:
		return "unknown"
	}
}

// LifecycleHooks provides a way to register lifecycle hooks without implementing the full interface.
type LifecycleHooks struct {
	OnInitFunc         func() tea.Cmd
	OnMountFunc        func() tea.Cmd
	OnUnmountFunc      func() tea.Cmd
	OnBeforeUpdateFunc func(msg tea.Msg) bool
	OnAfterUpdateFunc  func(msg tea.Msg) tea.Cmd
	OnShowFunc         func() tea.Cmd
	OnHideFunc         func() tea.Cmd
	OnResizeFunc       func(width, height int) tea.Cmd
	OnFocusFunc        func() tea.Cmd
	OnBlurFunc         func() tea.Cmd
}

// ExecuteInit executes the OnInit hook if defined.
func (h *LifecycleHooks) ExecuteInit() tea.Cmd {
	if h.OnInitFunc != nil {
		return h.OnInitFunc()
	}
	return nil
}

// ExecuteMount executes the OnMount hook if defined.
func (h *LifecycleHooks) ExecuteMount() tea.Cmd {
	if h.OnMountFunc != nil {
		return h.OnMountFunc()
	}
	return nil
}

// ExecuteUnmount executes the OnUnmount hook if defined.
func (h *LifecycleHooks) ExecuteUnmount() tea.Cmd {
	if h.OnUnmountFunc != nil {
		return h.OnUnmountFunc()
	}
	return nil
}

// ExecuteBeforeUpdate executes the OnBeforeUpdate hook if defined.
// Returns true if the update should proceed, false to cancel.
func (h *LifecycleHooks) ExecuteBeforeUpdate(msg tea.Msg) bool {
	if h.OnBeforeUpdateFunc != nil {
		return h.OnBeforeUpdateFunc(msg)
	}
	return true
}

// ExecuteAfterUpdate executes the OnAfterUpdate hook if defined.
func (h *LifecycleHooks) ExecuteAfterUpdate(msg tea.Msg) tea.Cmd {
	if h.OnAfterUpdateFunc != nil {
		return h.OnAfterUpdateFunc(msg)
	}
	return nil
}

// ExecuteShow executes the OnShow hook if defined.
func (h *LifecycleHooks) ExecuteShow() tea.Cmd {
	if h.OnShowFunc != nil {
		return h.OnShowFunc()
	}
	return nil
}

// ExecuteHide executes the OnHide hook if defined.
func (h *LifecycleHooks) ExecuteHide() tea.Cmd {
	if h.OnHideFunc != nil {
		return h.OnHideFunc()
	}
	return nil
}

// ExecuteResize executes the OnResize hook if defined.
func (h *LifecycleHooks) ExecuteResize(width, height int) tea.Cmd {
	if h.OnResizeFunc != nil {
		return h.OnResizeFunc(width, height)
	}
	return nil
}

// ExecuteFocus executes the OnFocus hook if defined.
func (h *LifecycleHooks) ExecuteFocus() tea.Cmd {
	if h.OnFocusFunc != nil {
		return h.OnFocusFunc()
	}
	return nil
}

// ExecuteBlur executes the OnBlur hook if defined.
func (h *LifecycleHooks) ExecuteBlur() tea.Cmd {
	if h.OnBlurFunc != nil {
		return h.OnBlurFunc()
	}
	return nil
}

// LifecycleManager manages the lifecycle of multiple components.
type LifecycleManager interface {
	// RegisterComponent registers a component with the lifecycle manager.
	RegisterComponent(name string, component Lifecycle) error

	// UnregisterComponent unregisters a component from the lifecycle manager.
	UnregisterComponent(name string) error

	// GetComponent returns a registered component by name.
	GetComponent(name string) (Lifecycle, error)

	// InitializeAll initializes all registered components.
	InitializeAll() []tea.Cmd

	// MountAll mounts all registered components.
	MountAll() []tea.Cmd

	// UnmountAll unmounts all registered components.
	UnmountAll() []tea.Cmd

	// GetRegisteredComponents returns the names of all registered components.
	GetRegisteredComponents() []string

	// HasComponent returns true if a component with the given name is registered.
	HasComponent(name string) bool
}

// LifecycleEvent represents an event in a component's lifecycle.
type LifecycleEvent struct {
	Type      LifecycleEventType
	Component string
	Timestamp int64
	Data      interface{}
}

// LifecycleEventType represents the type of lifecycle event.
type LifecycleEventType int

const (
	// EventInit represents an initialization event
	EventInit LifecycleEventType = iota

	// EventMount represents a mount event
	EventMount

	// EventUnmount represents an unmount event
	EventUnmount

	// EventUpdate represents an update event
	EventUpdate

	// EventShow represents a show event
	EventShow

	// EventHide represents a hide event
	EventHide

	// EventResize represents a resize event
	EventResize

	// EventFocus represents a focus event
	EventFocus

	// EventBlur represents a blur event
	EventBlur

	// EventError represents an error event
	EventError

	// EventDispose represents a dispose event
	EventDispose
)

// String returns the string representation of a LifecycleEventType.
func (e LifecycleEventType) String() string {
	switch e {
	case EventInit:
		return "init"
	case EventMount:
		return "mount"
	case EventUnmount:
		return "unmount"
	case EventUpdate:
		return "update"
	case EventShow:
		return "show"
	case EventHide:
		return "hide"
	case EventResize:
		return "resize"
	case EventFocus:
		return "focus"
	case EventBlur:
		return "blur"
	case EventError:
		return "error"
	case EventDispose:
		return "dispose"
	default:
		return "unknown"
	}
}

// LifecycleObserver can observe lifecycle events from components.
type LifecycleObserver interface {
	// OnLifecycleEvent is called when a lifecycle event occurs.
	OnLifecycleEvent(event LifecycleEvent)
}

// LifecycleEventBus manages lifecycle event distribution to observers.
type LifecycleEventBus interface {
	// Subscribe adds an observer to receive lifecycle events.
	Subscribe(observer LifecycleObserver)

	// Unsubscribe removes an observer from receiving lifecycle events.
	Unsubscribe(observer LifecycleObserver)

	// Publish publishes a lifecycle event to all observers.
	Publish(event LifecycleEvent)

	// GetObserverCount returns the number of subscribed observers.
	GetObserverCount() int

	// Clear removes all observers.
	Clear()
}

// Mountable represents components that can be mounted and unmounted.
// This is a simplified version of Lifecycle for components that only need
// mount/unmount hooks.
type Mountable interface {
	Component

	// Mount mounts the component.
	Mount() tea.Cmd

	// Unmount unmounts the component.
	Unmount() tea.Cmd

	// IsMounted returns true if the component is currently mounted.
	IsMounted() bool
}

// Initializable represents components that need explicit initialization.
// This is useful for components with expensive setup operations.
type Initializable interface {
	Component

	// Initialize initializes the component.
	Initialize() tea.Cmd

	// IsInitialized returns true if the component has been initialized.
	IsInitialized() bool

	// GetInitError returns any error that occurred during initialization.
	GetInitError() error
}

// Resettable represents components that can be reset to their initial state.
type Resettable interface {
	Component

	// Reset resets the component to its initial state.
	Reset() tea.Cmd

	// CanReset returns true if the component can be reset.
	CanReset() bool
}

// Reloadable represents components that can reload their content.
type Reloadable interface {
	Component

	// Reload reloads the component's content.
	Reload() tea.Cmd

	// IsReloading returns true if the component is currently reloading.
	IsReloading() bool

	// GetLastReloadTime returns the timestamp of the last reload.
	GetLastReloadTime() int64
}
