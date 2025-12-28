package tui

import tea "github.com/charmbracelet/bubbletea"

// Init initializes the TUI and returns initial commands
// This function is called once when the Bubble Tea program starts
func (m Model) Init() tea.Cmd {
	// Return a batch of initial commands
	return tea.Batch(
		// Signal that the TUI is ready to display
		SendReady(),
		// Start listening for terminal size changes
		tea.EnterAltScreen,
	)
}
