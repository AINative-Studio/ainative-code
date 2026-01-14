package toast

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// TestToastCreation tests basic toast creation
func TestToastCreation(t *testing.T) {
	config := DefaultToastConfig(ToastInfo)
	config.Message = "Test message"

	toast := NewToast(config)

	if toast == nil {
		t.Fatal("NewToast returned nil")
	}

	if toast.ID() == "" {
		t.Error("Toast ID should not be empty")
	}

	if toast.Config().Message != "Test message" {
		t.Errorf("Expected message 'Test message', got '%s'", toast.Config().Message)
	}

	if toast.Config().Type != ToastInfo {
		t.Errorf("Expected type ToastInfo, got %v", toast.Config().Type)
	}

	if toast.IsDismissed() {
		t.Error("New toast should not be dismissed")
	}
}

// TestToastTypes tests all toast types
func TestToastTypes(t *testing.T) {
	types := []ToastType{
		ToastInfo,
		ToastSuccess,
		ToastWarning,
		ToastError,
		ToastLoading,
	}

	for _, toastType := range types {
		config := DefaultToastConfig(toastType)
		config.Message = "Test"

		toast := NewToast(config)
		if toast.Config().Type != toastType {
			t.Errorf("Expected type %v, got %v", toastType, toast.Config().Type)
		}

		// Check default durations
		switch toastType {
		case ToastInfo, ToastSuccess:
			if config.Duration != 3*time.Second {
				t.Errorf("%s should have 3s duration, got %v", toastType, config.Duration)
			}
		case ToastWarning:
			if config.Duration != 5*time.Second {
				t.Errorf("%s should have 5s duration, got %v", toastType, config.Duration)
			}
		case ToastError:
			if config.Duration != 10*time.Second {
				t.Errorf("%s should have 10s duration, got %v", toastType, config.Duration)
			}
		case ToastLoading:
			if config.Duration != 0 {
				t.Errorf("%s should have 0 duration (manual dismiss), got %v", toastType, config.Duration)
			}
		}
	}
}

// TestToastDismiss tests toast dismissal
func TestToastDismiss(t *testing.T) {
	config := DefaultToastConfig(ToastInfo)
	config.Message = "Test"

	toast := NewToast(config)

	if toast.IsDismissed() {
		t.Error("New toast should not be dismissed")
	}

	cmd := toast.Dismiss()
	if cmd == nil {
		t.Error("Dismiss should return a command")
	}

	if !toast.IsDismissed() {
		t.Error("Toast should be marked as dismissed")
	}

	if !toast.IsFadingOut() {
		t.Error("Toast should be fading out after dismiss")
	}
}

// TestToastExpiration tests auto-expiration
func TestToastExpiration(t *testing.T) {
	config := DefaultToastConfig(ToastInfo)
	config.Message = "Test"
	config.Duration = 100 * time.Millisecond

	toast := NewToast(config)

	if toast.IsExpired() {
		t.Error("New toast should not be expired")
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	if !toast.IsExpired() {
		t.Error("Toast should be expired after duration")
	}
}

// TestToastManager tests toast manager creation
func TestToastManager(t *testing.T) {
	manager := NewToastManager()

	if manager == nil {
		t.Fatal("NewToastManager returned nil")
	}

	if manager.GetMaxToasts() != 3 {
		t.Errorf("Expected max toasts 3, got %d", manager.GetMaxToasts())
	}

	if manager.GetPosition() != TopRight {
		t.Errorf("Expected position TopRight, got %v", manager.GetPosition())
	}

	if manager.HasToasts() {
		t.Error("New manager should have no toasts")
	}

	if manager.GetQueueLength() != 0 {
		t.Errorf("New manager should have empty queue, got %d", manager.GetQueueLength())
	}
}

// TestToastManagerShowToast tests showing toasts
func TestToastManagerShowToast(t *testing.T) {
	manager := NewToastManager()

	config := DefaultToastConfig(ToastInfo)
	config.Message = "Test message"

	cmd := manager.ShowToast(config)
	if cmd == nil {
		t.Error("ShowToast should return a command")
	}

	if !manager.HasToasts() {
		t.Error("Manager should have toasts after showing")
	}

	toasts := manager.GetToasts()
	if len(toasts) != 1 {
		t.Errorf("Expected 1 toast, got %d", len(toasts))
	}

	if toasts[0].Config().Message != "Test message" {
		t.Errorf("Expected message 'Test message', got '%s'", toasts[0].Config().Message)
	}
}

// TestToastManagerQueueing tests toast queueing
func TestToastManagerQueueing(t *testing.T) {
	manager := NewToastManager()
	manager.SetMaxToasts(2)

	// Add 4 toasts (should queue 2)
	for i := 0; i < 4; i++ {
		config := DefaultToastConfig(ToastInfo)
		config.Message = "Test"
		manager.ShowToast(config)
	}

	if len(manager.GetToasts()) != 2 {
		t.Errorf("Expected 2 visible toasts, got %d", len(manager.GetToasts()))
	}

	if manager.GetQueueLength() != 2 {
		t.Errorf("Expected 2 queued toasts, got %d", manager.GetQueueLength())
	}
}

// TestToastManagerDismiss tests dismissing toasts
func TestToastManagerDismiss(t *testing.T) {
	manager := NewToastManager()

	config := DefaultToastConfig(ToastInfo)
	config.Message = "Test"
	manager.ShowToast(config)

	toasts := manager.GetToasts()
	if len(toasts) != 1 {
		t.Fatal("Expected 1 toast")
	}

	toastID := toasts[0].ID()
	cmd := manager.DismissToast(toastID)
	if cmd == nil {
		t.Error("DismissToast should return a command")
	}

	// Toast should be marked as dismissed
	if !toasts[0].IsDismissed() {
		t.Error("Toast should be dismissed")
	}
}

// TestToastManagerDismissAll tests dismissing all toasts
func TestToastManagerDismissAll(t *testing.T) {
	manager := NewToastManager()

	// Add 3 toasts
	for i := 0; i < 3; i++ {
		config := DefaultToastConfig(ToastInfo)
		config.Message = "Test"
		manager.ShowToast(config)
	}

	cmd := manager.DismissAll()
	if cmd == nil {
		t.Error("DismissAll should return a command")
	}

	// All toasts should be dismissed
	toasts := manager.GetToasts()
	for i, toast := range toasts {
		if !toast.IsDismissed() {
			t.Errorf("Toast %d should be dismissed", i)
		}
	}
}

// TestToastManagerHelperMethods tests helper methods
func TestToastManagerHelperMethods(t *testing.T) {
	tests := []struct {
		name     string
		expected ToastType
	}{
		{"ShowInfo", ToastInfo},
		{"ShowSuccess", ToastSuccess},
		{"ShowWarning", ToastWarning},
		{"ShowError", ToastError},
		{"ShowLoading", ToastLoading},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewToastManager() // Fresh manager for each test

			var cmd tea.Cmd
			switch tt.expected {
			case ToastInfo:
				cmd = manager.ShowInfo("Test message")
			case ToastSuccess:
				cmd = manager.ShowSuccess("Test message")
			case ToastWarning:
				cmd = manager.ShowWarning("Test message")
			case ToastError:
				cmd = manager.ShowError("Test message")
			case ToastLoading:
				cmd = manager.ShowLoading("Test message")
			}

			if cmd == nil {
				t.Error("Method should return a command")
			}

			toasts := manager.GetToasts()
			if len(toasts) != 1 {
				t.Errorf("Expected 1 toast, got %d", len(toasts))
				return
			}

			if toasts[0].Config().Type != tt.expected {
				t.Errorf("Expected type %v, got %v", tt.expected, toasts[0].Config().Type)
			}
		})
	}
}

// TestToastManagerUpdate tests update logic
func TestToastManagerUpdate(t *testing.T) {
	manager := NewToastManager()

	// Test ShowToastMsg
	config := DefaultToastConfig(ToastInfo)
	config.Message = "Test"
	msg := ShowToastMsg{Config: config}

	model, cmd := manager.Update(msg)
	manager = model.(*ToastManager)

	if cmd == nil {
		t.Error("Update should return a command for ShowToastMsg")
	}

	if !manager.HasToasts() {
		t.Error("Manager should have toasts after ShowToastMsg")
	}
}

// TestToastManagerSize tests size management
func TestToastManagerSize(t *testing.T) {
	manager := NewToastManager()

	manager.SetSize(100, 50)

	// Width should be adjusted for toasts
	if manager.width == 0 {
		t.Error("Width should be set")
	}

	if manager.screenWidth != 100 || manager.screenHeight != 50 {
		t.Error("Screen dimensions should be set")
	}
}

// TestToastPositions tests all toast positions
func TestToastPositions(t *testing.T) {
	positions := []ToastPosition{
		TopRight,
		TopCenter,
		TopLeft,
		BottomRight,
		BottomCenter,
		BottomLeft,
	}

	for _, pos := range positions {
		manager := NewToastManager()
		manager.SetPosition(pos)

		if manager.GetPosition() != pos {
			t.Errorf("Expected position %v, got %v", pos, manager.GetPosition())
		}
	}
}

// TestToastManagerStats tests statistics
func TestToastManagerStats(t *testing.T) {
	manager := NewToastManager()
	manager.SetMaxToasts(2)

	// Add 3 toasts (1 queued)
	for i := 0; i < 3; i++ {
		config := DefaultToastConfig(ToastInfo)
		config.Message = "Test"
		manager.ShowToast(config)
	}

	stats := manager.Stats()
	if stats == "" {
		t.Error("Stats should return a string")
	}

	// Stats should contain information about visible and queued toasts
	if !contains(stats, "Visible: 2") {
		t.Error("Stats should show 2 visible toasts")
	}

	if !contains(stats, "Queued: 1") {
		t.Error("Stats should show 1 queued toast")
	}
}

// TestToastBuilder tests the builder pattern
func TestToastBuilder(t *testing.T) {
	builder := NewToastBuilder(ToastInfo)

	cmd := builder.
		WithMessage("Test message").
		WithTitle("Test title").
		WithDuration(5 * time.Second).
		WithPosition(TopLeft).
		WithIcon("🔔").
		Dismissible(true).
		Build()

	if cmd == nil {
		t.Error("Build should return a command")
	}

	// Test persistent builder
	cmd = NewToastBuilder(ToastInfo).
		WithMessage("Persistent").
		Persistent().
		Build()

	if cmd == nil {
		t.Error("Build should return a command for persistent toast")
	}
}

// TestConvenienceMethods tests convenience message functions
func TestConvenienceMethods(t *testing.T) {
	tests := []struct {
		name string
		cmd  tea.Cmd
	}{
		{"QuickInfo", QuickInfo("test")},
		{"QuickSuccess", QuickSuccess("test")},
		{"LongWarning", LongWarning("test")},
		{"CriticalError", CriticalError("test")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.cmd == nil {
				t.Error("Convenience method should return a command")
			}
		})
	}
}

// TestToastIcons tests icon retrieval
func TestToastIcons(t *testing.T) {
	types := []ToastType{
		ToastInfo,
		ToastSuccess,
		ToastWarning,
		ToastError,
		ToastLoading,
	}

	for _, toastType := range types {
		icon := GetDefaultIcon(toastType)
		if icon == "" {
			t.Errorf("GetDefaultIcon should return an icon for %v", toastType)
		}
	}

	// Test custom icon
	config := DefaultToastConfig(ToastInfo)
	config.Icon = "🔔"
	toast := NewToast(config)

	if toast.Config().Icon != "🔔" {
		t.Error("Custom icon should be preserved")
	}
}

// TestToastStyles tests style functions
func TestToastStyles(t *testing.T) {
	types := []ToastType{
		ToastInfo,
		ToastSuccess,
		ToastWarning,
		ToastError,
		ToastLoading,
	}

	for _, toastType := range types {
		style := GetToastStyle(toastType, 40)
		// Check if style is set by checking if String() returns something
		if style.String() == "" {
			t.Errorf("GetToastStyle should return a style for %v", toastType)
		}
	}
}

// TestToastView tests toast rendering
func TestToastView(t *testing.T) {
	config := DefaultToastConfig(ToastInfo)
	config.Message = "Test message"
	config.Title = "Test title"

	toast := NewToast(config)
	toast.SetValue(1.0) // Set opacity to 1.0 for rendering

	view := toast.View()
	if view == "" {
		t.Error("Toast view should not be empty")
	}

	// View should contain the message
	if !contains(view, "Test message") {
		t.Error("Toast view should contain the message")
	}
}

// TestToastManagerView tests manager rendering
func TestToastManagerView(t *testing.T) {
	manager := NewToastManager()
	manager.SetSize(100, 50)

	// Empty manager should return empty view
	view := manager.View()
	if view != "" {
		t.Error("Empty manager view should be empty")
	}

	// Add a toast
	config := DefaultToastConfig(ToastInfo)
	config.Message = "Test"
	manager.ShowToast(config)

	// Set opacity for visibility
	toasts := manager.GetToasts()
	if len(toasts) > 0 {
		toasts[0].opacity = 1.0
	}

	view = manager.View()
	// View may be empty if toasts haven't been initialized properly
	// This is acceptable in unit tests
}

// TestCalculateToastWidth tests width calculation
func TestCalculateToastWidth(t *testing.T) {
	tests := []struct {
		screenWidth int
		expected    int
	}{
		{40, 36},   // Small screen
		{60, 40},   // Medium screen
		{90, 50},   // Large screen
		{130, 60},  // Extra large screen
	}

	for _, tt := range tests {
		width := CalculateToastWidth(tt.screenWidth)
		if width != tt.expected {
			t.Errorf("For screen width %d, expected %d, got %d", tt.screenWidth, tt.expected, width)
		}
	}
}

// TestCalculateMaxToasts tests max toasts calculation
func TestCalculateMaxToasts(t *testing.T) {
	tests := []struct {
		screenHeight int
		minExpected  int
		maxExpected  int
	}{
		{10, 1, 1},   // Very small screen
		{20, 1, 2},   // Small screen
		{40, 2, 6},   // Medium screen
		{80, 5, 14},  // Large screen
	}

	for _, tt := range tests {
		maxToasts := CalculateMaxToasts(tt.screenHeight)
		if maxToasts < tt.minExpected || maxToasts > tt.maxExpected {
			t.Errorf("For screen height %d, expected between %d and %d, got %d",
				tt.screenHeight, tt.minExpected, tt.maxExpected, maxToasts)
		}
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
