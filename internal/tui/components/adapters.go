package components

import (
	tea "github.com/charmbracelet/bubbletea"
)

// ComponentAdapter provides a base implementation for components.
// It can be embedded in concrete component implementations to reduce boilerplate.
type ComponentAdapter struct {
	width      int
	height     int
	visible    bool
	focused    bool
	initialized bool
	mounted    bool
	state      LifecycleState
}

// NewComponentAdapter creates a new component adapter.
func NewComponentAdapter() *ComponentAdapter {
	return &ComponentAdapter{
		width:      0,
		height:     0,
		visible:    true,
		focused:    false,
		initialized: false,
		mounted:    false,
		state:      StateUninitialized,
	}
}

// Init initializes the component adapter.
func (c *ComponentAdapter) Init() tea.Cmd {
	c.initialized = true
	c.state = StateInitialized
	return nil
}

// Update handles messages (default implementation).
func (c *ComponentAdapter) Update(msg tea.Msg) (Component, tea.Cmd) {
	return c, nil
}

// View renders the component (default implementation).
func (c *ComponentAdapter) View() string {
	return ""
}

// SetSize implements Sizeable.
func (c *ComponentAdapter) SetSize(width, height int) {
	c.width = width
	c.height = height
}

// GetSize implements Sizeable.
func (c *ComponentAdapter) GetSize() (width, height int) {
	return c.width, c.height
}

// Focus implements Focusable.
func (c *ComponentAdapter) Focus() tea.Cmd {
	c.focused = true
	return nil
}

// Blur implements Focusable.
func (c *ComponentAdapter) Blur() {
	c.focused = false
}

// Focused implements Focusable.
func (c *ComponentAdapter) Focused() bool {
	return c.focused
}

// Show implements Stateful.
func (c *ComponentAdapter) Show() {
	c.visible = true
}

// Hide implements Stateful.
func (c *ComponentAdapter) Hide() {
	c.visible = false
}

// Toggle implements Stateful.
func (c *ComponentAdapter) Toggle() {
	c.visible = !c.visible
}

// IsVisible implements Stateful.
func (c *ComponentAdapter) IsVisible() bool {
	return c.visible
}

// OnInit implements Lifecycle.
func (c *ComponentAdapter) OnInit() tea.Cmd {
	c.state = StateInitializing
	return nil
}

// OnMount implements Lifecycle.
func (c *ComponentAdapter) OnMount() tea.Cmd {
	c.mounted = true
	c.state = StateMounted
	return nil
}

// OnUnmount implements Lifecycle.
func (c *ComponentAdapter) OnUnmount() tea.Cmd {
	c.mounted = false
	c.state = StateUnmounted
	return nil
}

// OnBeforeUpdate implements Lifecycle.
func (c *ComponentAdapter) OnBeforeUpdate(msg tea.Msg) bool {
	return true
}

// OnAfterUpdate implements Lifecycle.
func (c *ComponentAdapter) OnAfterUpdate(msg tea.Msg) tea.Cmd {
	return nil
}

// OnShow implements Lifecycle.
func (c *ComponentAdapter) OnShow() tea.Cmd {
	return nil
}

// OnHide implements Lifecycle.
func (c *ComponentAdapter) OnHide() tea.Cmd {
	return nil
}

// OnResize implements Lifecycle.
func (c *ComponentAdapter) OnResize(width, height int) tea.Cmd {
	c.width = width
	c.height = height
	return nil
}

// OnFocus implements Lifecycle.
func (c *ComponentAdapter) OnFocus() tea.Cmd {
	return nil
}

// OnBlur implements Lifecycle.
func (c *ComponentAdapter) OnBlur() tea.Cmd {
	return nil
}

// GetLifecycleState implements Lifecycle.
func (c *ComponentAdapter) GetLifecycleState() LifecycleState {
	return c.state
}

// IsInitialized implements Lifecycle.
func (c *ComponentAdapter) IsInitialized() bool {
	return c.initialized
}

// IsMounted implements Lifecycle.
func (c *ComponentAdapter) IsMounted() bool {
	return c.mounted
}

// PopupAdapter provides a base implementation for popup components.
type PopupAdapter struct {
	*ComponentAdapter
	x       int
	y       int
	zIndex  int
	modal   bool
	onClose func()
	config  PopupConfig
}

// NewPopupAdapter creates a new popup adapter with default configuration.
func NewPopupAdapter() *PopupAdapter {
	return &PopupAdapter{
		ComponentAdapter: NewComponentAdapter(),
		x:                0,
		y:                0,
		zIndex:           100,
		modal:            false,
		config:           DefaultPopupConfig(),
	}
}

// RenderPopup implements PopupComponent (default implementation).
func (p *PopupAdapter) RenderPopup() string {
	return ""
}

// GetPopupPosition implements PopupComponent.
func (p *PopupAdapter) GetPopupPosition() (x, y, width, height int) {
	return p.x, p.y, p.width, p.height
}

// SetPopupPosition implements PopupComponent.
func (p *PopupAdapter) SetPopupPosition(x, y int) {
	p.x = x
	p.y = y
}

// GetPopupDimensions implements PopupComponent.
func (p *PopupAdapter) GetPopupDimensions() (width, height int) {
	return p.width, p.height
}

// SetPopupDimensions implements PopupComponent.
func (p *PopupAdapter) SetPopupDimensions(width, height int) {
	p.width = width
	p.height = height
}

// IsModal implements PopupComponent.
func (p *PopupAdapter) IsModal() bool {
	return p.modal
}

// SetModal implements PopupComponent.
func (p *PopupAdapter) SetModal(modal bool) {
	p.modal = modal
}

// GetZIndex implements PopupComponent.
func (p *PopupAdapter) GetZIndex() int {
	return p.zIndex
}

// SetZIndex implements PopupComponent.
func (p *PopupAdapter) SetZIndex(zIndex int) {
	p.zIndex = zIndex
}

// Close implements PopupComponent.
func (p *PopupAdapter) Close() tea.Cmd {
	p.Hide()
	if p.onClose != nil {
		p.onClose()
	}
	return nil
}

// OnClose implements PopupComponent.
func (p *PopupAdapter) OnClose(callback func()) {
	p.onClose = callback
}

// GetConfig returns the popup configuration.
func (p *PopupAdapter) GetConfig() PopupConfig {
	return p.config
}

// SetConfig sets the popup configuration.
func (p *PopupAdapter) SetConfig(config PopupConfig) {
	p.config = config
	p.width = config.Width
	p.height = config.Height
	p.x = config.X
	p.y = config.Y
	p.zIndex = config.ZIndex
	p.modal = config.Modal
}
