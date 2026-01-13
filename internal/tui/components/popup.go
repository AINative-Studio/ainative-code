package components

import (
	tea "github.com/charmbracelet/bubbletea"
)

// PopupComponent represents a component that can be displayed as an overlay popup.
// Popups are temporary UI elements that appear on top of the main content.
// They typically handle user interactions like completions, hover info, and navigation.
type PopupComponent interface {
	Component
	Stateful
	Sizeable

	// RenderPopup renders the popup content without applying overlay positioning.
	// The returned string contains the popup's visual content.
	RenderPopup() string

	// GetPopupPosition returns the preferred position for the popup.
	// Returns (x, y, width, height) coordinates relative to the screen.
	GetPopupPosition() (x, y, width, height int)

	// SetPopupPosition sets the popup's position on the screen.
	SetPopupPosition(x, y int)

	// GetPopupDimensions returns the popup's dimensions.
	GetPopupDimensions() (width, height int)

	// SetPopupDimensions sets the popup's dimensions.
	SetPopupDimensions(width, height int)

	// IsModal returns true if the popup should block interaction with underlying content.
	IsModal() bool

	// SetModal sets whether the popup blocks underlying content interaction.
	SetModal(modal bool)

	// GetZIndex returns the popup's z-index for stacking order.
	// Higher values appear on top of lower values.
	GetZIndex() int

	// SetZIndex sets the popup's z-index.
	SetZIndex(zIndex int)

	// Close closes the popup and performs any necessary cleanup.
	Close() tea.Cmd

	// OnClose registers a callback to be called when the popup closes.
	OnClose(callback func())
}

// PopupAlignment specifies how a popup should be aligned relative to a reference point.
type PopupAlignment int

const (
	// AlignTopLeft aligns popup's top-left corner to reference point
	AlignTopLeft PopupAlignment = iota

	// AlignTopCenter aligns popup's top edge center to reference point
	AlignTopCenter

	// AlignTopRight aligns popup's top-right corner to reference point
	AlignTopRight

	// AlignCenterLeft aligns popup's left edge center to reference point
	AlignCenterLeft

	// AlignCenter centers popup at reference point
	AlignCenter

	// AlignCenterRight aligns popup's right edge center to reference point
	AlignCenterRight

	// AlignBottomLeft aligns popup's bottom-left corner to reference point
	AlignBottomLeft

	// AlignBottomCenter aligns popup's bottom edge center to reference point
	AlignBottomCenter

	// AlignBottomRight aligns popup's bottom-right corner to reference point
	AlignBottomRight
)

// PopupStyle defines the visual style of a popup.
type PopupStyle struct {
	BorderStyle      string // "rounded", "double", "single", "none"
	BorderColor      string
	BackgroundColor  string
	Shadow           bool
	Padding          int
	Margin           int
	Transparency     float64 // 0.0 (opaque) to 1.0 (transparent)
	BlurBackground   bool    // Whether to blur the background content
}

// PopupConfig contains configuration for popup behavior.
type PopupConfig struct {
	Width            int
	Height           int
	X                int
	Y                int
	Alignment        PopupAlignment
	Style            PopupStyle
	Modal            bool
	ZIndex           int
	CloseOnEscape    bool
	CloseOnClickOut  bool
	AnimateOpen      bool
	AnimateClose     bool
	FocusOnOpen      bool
}

// DefaultPopupConfig returns a default popup configuration.
func DefaultPopupConfig() PopupConfig {
	return PopupConfig{
		Width:           60,
		Height:          20,
		X:               0,
		Y:               0,
		Alignment:       AlignCenter,
		Modal:           false,
		ZIndex:          100,
		CloseOnEscape:   true,
		CloseOnClickOut: false,
		AnimateOpen:     false,
		AnimateClose:    false,
		FocusOnOpen:     true,
		Style: PopupStyle{
			BorderStyle:     "rounded",
			BorderColor:     "62",
			BackgroundColor: "235",
			Shadow:          false,
			Padding:         1,
			Margin:          0,
			Transparency:    0.0,
			BlurBackground:  false,
		},
	}
}

// SelectablePopup represents a popup with selectable items.
// This is useful for completion popups, dropdown menus, and picker dialogs.
type SelectablePopup interface {
	PopupComponent
	Selectable

	// GetItems returns all items in the popup
	GetItems() []interface{}

	// SetItems sets the items in the popup
	SetItems(items []interface{})

	// GetItemCount returns the number of items
	GetItemCount() int

	// ClearItems removes all items
	ClearItems()

	// FilterItems filters items based on a predicate
	FilterItems(filter func(item interface{}) bool)

	// SortItems sorts items using a comparator
	SortItems(less func(i, j interface{}) bool)
}

// ScrollablePopup represents a popup with scrollable content.
// This is useful for popups with long content that doesn't fit in the visible area.
type ScrollablePopup interface {
	PopupComponent
	Scrollable

	// GetVisibleItemCount returns the number of items visible at once
	GetVisibleItemCount() int

	// SetVisibleItemCount sets the number of items visible at once
	SetVisibleItemCount(count int)

	// GetScrollOffset returns the current scroll offset
	GetScrollOffset() int

	// SetScrollOffset sets the current scroll offset
	SetScrollOffset(offset int)

	// CanScrollUp returns true if scrolling up is possible
	CanScrollUp() bool

	// CanScrollDown returns true if scrolling down is possible
	CanScrollDown() bool
}

// FilterablePopup represents a popup with filterable content.
// This is useful for search dialogs and filtered lists.
type FilterablePopup interface {
	PopupComponent

	// GetFilter returns the current filter text
	GetFilter() string

	// SetFilter sets the filter text and filters content
	SetFilter(filter string)

	// ClearFilter removes the filter and shows all content
	ClearFilter()

	// IsFiltered returns true if a filter is active
	IsFiltered() bool
}

// ConfirmationPopup represents a popup that asks for user confirmation.
// This is useful for destructive actions and important decisions.
type ConfirmationPopup interface {
	PopupComponent

	// GetMessage returns the confirmation message
	GetMessage() string

	// SetMessage sets the confirmation message
	SetMessage(message string)

	// GetButtons returns the button labels
	GetButtons() []string

	// SetButtons sets the button labels
	SetButtons(buttons []string)

	// GetDefaultButton returns the index of the default button
	GetDefaultButton() int

	// SetDefaultButton sets the default button index
	SetDefaultButton(index int)

	// GetSelectedButton returns the index of the currently selected button
	GetSelectedButton() int

	// Confirm triggers the confirmation action
	Confirm() tea.Cmd

	// Cancel triggers the cancel action
	Cancel() tea.Cmd
}

// NotificationPopup represents a temporary notification popup.
// This is useful for status messages, alerts, and toast notifications.
type NotificationPopup interface {
	PopupComponent

	// GetMessage returns the notification message
	GetMessage() string

	// SetMessage sets the notification message
	SetMessage(message string)

	// GetType returns the notification type (info, success, warning, error)
	GetType() string

	// SetType sets the notification type
	SetType(notificationType string)

	// GetDuration returns how long the notification should be displayed
	GetDuration() int

	// SetDuration sets how long the notification should be displayed (in milliseconds)
	SetDuration(duration int)

	// AutoClose returns true if the notification closes automatically
	AutoClose() bool

	// SetAutoClose sets whether the notification closes automatically
	SetAutoClose(autoClose bool)
}

// PopupManager manages multiple popups and their stacking order.
type PopupManager interface {
	// AddPopup adds a popup to the manager
	AddPopup(popup PopupComponent) error

	// RemovePopup removes a popup from the manager
	RemovePopup(popup PopupComponent) error

	// GetActivePopup returns the currently active (topmost) popup
	GetActivePopup() PopupComponent

	// GetPopups returns all managed popups sorted by z-index
	GetPopups() []PopupComponent

	// CloseAll closes all managed popups
	CloseAll() tea.Cmd

	// HasPopups returns true if any popups are open
	HasPopups() bool

	// GetPopupCount returns the number of open popups
	GetPopupCount() int

	// BringToFront brings a popup to the front by increasing its z-index
	BringToFront(popup PopupComponent)

	// SendToBack sends a popup to the back by decreasing its z-index
	SendToBack(popup PopupComponent)
}
