package toast

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// ShowToastMsg triggers a new toast notification
type ShowToastMsg struct {
	Config ToastConfig
}

// DismissToastMsg dismisses a specific toast by ID
type DismissToastMsg struct {
	ID string
}

// DismissAllToastsMsg dismisses all visible toasts
type DismissAllToastsMsg struct{}

// ToastExpiredMsg is fired when a toast auto-dismisses
type ToastExpiredMsg struct {
	ID string
}

// ToastActionMsg is fired when an action button is clicked
type ToastActionMsg struct {
	ToastID string
	Action  *ToastAction
}

// ToastTickMsg is sent periodically to update toast state
type ToastTickMsg struct {
	ID string
}

// Command helper functions

// ShowInfo creates a command to show an info toast
func ShowInfo(message string) tea.Cmd {
	return func() tea.Msg {
		config := DefaultToastConfig(ToastInfo)
		config.Message = message
		return ShowToastMsg{Config: config}
	}
}

// ShowInfoWithTitle creates a command to show an info toast with a title
func ShowInfoWithTitle(title, message string) tea.Cmd {
	return func() tea.Msg {
		config := DefaultToastConfig(ToastInfo)
		config.Title = title
		config.Message = message
		return ShowToastMsg{Config: config}
	}
}

// ShowSuccess creates a command to show a success toast
func ShowSuccess(message string) tea.Cmd {
	return func() tea.Msg {
		config := DefaultToastConfig(ToastSuccess)
		config.Message = message
		return ShowToastMsg{Config: config}
	}
}

// ShowSuccessWithTitle creates a command to show a success toast with a title
func ShowSuccessWithTitle(title, message string) tea.Cmd {
	return func() tea.Msg {
		config := DefaultToastConfig(ToastSuccess)
		config.Title = title
		config.Message = message
		return ShowToastMsg{Config: config}
	}
}

// ShowWarning creates a command to show a warning toast
func ShowWarning(message string) tea.Cmd {
	return func() tea.Msg {
		config := DefaultToastConfig(ToastWarning)
		config.Message = message
		return ShowToastMsg{Config: config}
	}
}

// ShowWarningWithTitle creates a command to show a warning toast with a title
func ShowWarningWithTitle(title, message string) tea.Cmd {
	return func() tea.Msg {
		config := DefaultToastConfig(ToastWarning)
		config.Title = title
		config.Message = message
		return ShowToastMsg{Config: config}
	}
}

// ShowError creates a command to show an error toast
func ShowError(message string) tea.Cmd {
	return func() tea.Msg {
		config := DefaultToastConfig(ToastError)
		config.Message = message
		return ShowToastMsg{Config: config}
	}
}

// ShowErrorWithTitle creates a command to show an error toast with a title
func ShowErrorWithTitle(title, message string) tea.Cmd {
	return func() tea.Msg {
		config := DefaultToastConfig(ToastError)
		config.Title = title
		config.Message = message
		return ShowToastMsg{Config: config}
	}
}

// ShowLoading creates a command to show a loading toast
func ShowLoading(message string) tea.Cmd {
	return func() tea.Msg {
		config := DefaultToastConfig(ToastLoading)
		config.Message = message
		return ShowToastMsg{Config: config}
	}
}

// ShowLoadingWithTitle creates a command to show a loading toast with a title
func ShowLoadingWithTitle(title, message string) tea.Cmd {
	return func() tea.Msg {
		config := DefaultToastConfig(ToastLoading)
		config.Title = title
		config.Message = message
		return ShowToastMsg{Config: config}
	}
}

// ShowCustomToast creates a command to show a custom toast with full configuration
func ShowCustomToast(config ToastConfig) tea.Cmd {
	return func() tea.Msg {
		return ShowToastMsg{Config: config}
	}
}

// DismissToast creates a command to dismiss a specific toast
func DismissToast(id string) tea.Cmd {
	return func() tea.Msg {
		return DismissToastMsg{ID: id}
	}
}

// DismissAllToasts creates a command to dismiss all toasts
func DismissAllToasts() tea.Cmd {
	return func() tea.Msg {
		return DismissAllToastsMsg{}
	}
}

// ShowTemporaryToast creates a toast that auto-dismisses after the specified duration
func ShowTemporaryToast(toastType ToastType, message string, duration time.Duration) tea.Cmd {
	return func() tea.Msg {
		config := DefaultToastConfig(toastType)
		config.Message = message
		config.Duration = duration
		return ShowToastMsg{Config: config}
	}
}

// ShowPersistentToast creates a toast that requires manual dismissal
func ShowPersistentToast(toastType ToastType, message string) tea.Cmd {
	return func() tea.Msg {
		config := DefaultToastConfig(toastType)
		config.Message = message
		config.Duration = 0 // No auto-dismiss
		return ShowToastMsg{Config: config}
	}
}

// ShowToastWithAction creates a toast with an action button
func ShowToastWithAction(toastType ToastType, message string, actionLabel string, actionCmd tea.Cmd) tea.Cmd {
	return func() tea.Msg {
		config := DefaultToastConfig(toastType)
		config.Message = message
		config.Action = &ToastAction{
			Label:   actionLabel,
			Command: actionCmd,
		}
		return ShowToastMsg{Config: config}
	}
}

// ShowUndoableToast creates a toast with an undo action
func ShowUndoableToast(message string, undoCmd tea.Cmd) tea.Cmd {
	return ShowToastWithAction(ToastInfo, message, "Undo", undoCmd)
}

// ShowProgressToast creates a loading toast for tracking progress
func ShowProgressToast(message string) tea.Cmd {
	return func() tea.Msg {
		config := DefaultToastConfig(ToastLoading)
		config.Message = message
		config.Dismissible = false // Can't be dismissed manually
		return ShowToastMsg{Config: config}
	}
}

// Builder pattern for complex toast configurations

// ToastBuilder helps build complex toast configurations
type ToastBuilder struct {
	config ToastConfig
}

// NewToastBuilder creates a new toast builder
func NewToastBuilder(toastType ToastType) *ToastBuilder {
	return &ToastBuilder{
		config: DefaultToastConfig(toastType),
	}
}

// WithMessage sets the toast message
func (b *ToastBuilder) WithMessage(message string) *ToastBuilder {
	b.config.Message = message
	return b
}

// WithTitle sets the toast title
func (b *ToastBuilder) WithTitle(title string) *ToastBuilder {
	b.config.Title = title
	return b
}

// WithDuration sets the auto-dismiss duration
func (b *ToastBuilder) WithDuration(duration time.Duration) *ToastBuilder {
	b.config.Duration = duration
	return b
}

// WithPosition sets the toast position
func (b *ToastBuilder) WithPosition(position ToastPosition) *ToastBuilder {
	b.config.Position = position
	return b
}

// WithIcon sets a custom icon
func (b *ToastBuilder) WithIcon(icon string) *ToastBuilder {
	b.config.Icon = icon
	return b
}

// WithAction adds an action button
func (b *ToastBuilder) WithAction(label string, cmd tea.Cmd) *ToastBuilder {
	b.config.Action = &ToastAction{
		Label:   label,
		Command: cmd,
	}
	return b
}

// Dismissible sets whether the toast can be manually dismissed
func (b *ToastBuilder) Dismissible(dismissible bool) *ToastBuilder {
	b.config.Dismissible = dismissible
	return b
}

// Persistent makes the toast require manual dismissal
func (b *ToastBuilder) Persistent() *ToastBuilder {
	b.config.Duration = 0
	return b
}

// Build returns the configured toast command
func (b *ToastBuilder) Build() tea.Cmd {
	return ShowCustomToast(b.config)
}

// Convenience methods for common toast patterns

// QuickInfo shows a quick info message (2 seconds)
func QuickInfo(message string) tea.Cmd {
	return ShowTemporaryToast(ToastInfo, message, 2*time.Second)
}

// QuickSuccess shows a quick success message (2 seconds)
func QuickSuccess(message string) tea.Cmd {
	return ShowTemporaryToast(ToastSuccess, message, 2*time.Second)
}

// LongWarning shows a warning that stays for 7 seconds
func LongWarning(message string) tea.Cmd {
	return ShowTemporaryToast(ToastWarning, message, 7*time.Second)
}

// CriticalError shows a persistent error that must be dismissed manually
func CriticalError(message string) tea.Cmd {
	return ShowPersistentToast(ToastError, message)
}

// SaveNotification shows a save confirmation with undo option
func SaveNotification(filename string, undoCmd tea.Cmd) tea.Cmd {
	return NewToastBuilder(ToastSuccess).
		WithMessage("Saved " + filename).
		WithAction("Undo", undoCmd).
		WithDuration(5 * time.Second).
		Build()
}

// DeleteNotification shows a delete confirmation with undo option
func DeleteNotification(itemName string, undoCmd tea.Cmd) tea.Cmd {
	return NewToastBuilder(ToastWarning).
		WithMessage("Deleted " + itemName).
		WithAction("Undo", undoCmd).
		WithDuration(5 * time.Second).
		Build()
}

// NetworkErrorNotification shows a network error toast
func NetworkErrorNotification(retryCmd tea.Cmd) tea.Cmd {
	return NewToastBuilder(ToastError).
		WithTitle("Network Error").
		WithMessage("Could not connect to server").
		WithAction("Retry", retryCmd).
		WithDuration(10 * time.Second).
		Build()
}
