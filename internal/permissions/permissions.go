// Package permissions provides tool execution permission management
package permissions

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Manager handles tool execution permissions
type Manager struct {
	allowedTools map[string]bool
	yolo         bool // Skip all prompts
}

// NewManager creates a new permission manager
func NewManager(allowedTools []string, yolo bool) *Manager {
	allowed := make(map[string]bool)
	for _, tool := range allowedTools {
		allowed[tool] = true
	}

	return &Manager{
		allowedTools: allowed,
		yolo:         yolo,
	}
}

// CheckPermission checks if a tool can be executed
// Returns true if allowed, false if denied
func (m *Manager) CheckPermission(toolName string, args map[string]interface{}) (bool, error) {
	// YOLO mode: allow everything
	if m.yolo {
		return true, nil
	}

	// Check if tool is in allowed list
	if m.allowedTools[toolName] {
		return true, nil
	}

	// Ask user for permission
	return m.promptUser(toolName, args)
}

// promptUser displays a permission prompt and gets user response
func (m *Manager) promptUser(toolName string, args map[string]interface{}) (bool, error) {
	fmt.Println()
	fmt.Println(titleStyle.Render("🔐 Tool Permission Required"))
	fmt.Println()
	fmt.Println(infoStyle.Render(fmt.Sprintf("Tool: %s", toolName)))

	if len(args) > 0 {
		fmt.Println(infoStyle.Render("Arguments:"))
		for key, value := range args {
			fmt.Println(infoStyle.Render(fmt.Sprintf("  %s: %v", key, value)))
		}
	}

	fmt.Println()
	fmt.Println(promptStyle.Render("Allow this tool to execute? [y/N/always/never]: "))

	var response string
	fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))

	switch response {
	case "y", "yes":
		return true, nil
	case "always":
		// Add to allowed list for this session
		m.allowedTools[toolName] = true
		fmt.Println(successStyle.Render(fmt.Sprintf("✓ Tool '%s' will always be allowed in this session", toolName)))
		return true, nil
	case "never":
		// Explicitly deny
		fmt.Println(errorStyle.Render(fmt.Sprintf("✗ Tool '%s' denied", toolName)))
		return false, nil
	default:
		// Default to deny
		fmt.Println(warnStyle.Render("Permission denied"))
		return false, nil
	}
}

// Styles for permission prompts
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")). // Blue
			MarginBottom(1)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("250")) // Light gray

	promptStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("226")) // Yellow

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")) // Green

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")) // Red

	warnStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")) // Orange
)

// PermissionPromptModel is a bubbletea model for permission prompts
type PermissionPromptModel struct {
	toolName string
	args     map[string]interface{}
	choice   string
	quitting bool
	list     list.Model
}

type permissionItem struct {
	title, desc string
	value       string
}

func (i permissionItem) Title() string       { return i.title }
func (i permissionItem) Description() string { return i.desc }
func (i permissionItem) FilterValue() string { return i.title }

// NewPermissionPromptModel creates a new permission prompt model
func NewPermissionPromptModel(toolName string, args map[string]interface{}) PermissionPromptModel {
	items := []list.Item{
		permissionItem{
			title: "Allow Once",
			desc:  "Execute this tool one time",
			value: "yes",
		},
		permissionItem{
			title: "Always Allow (This Session)",
			desc:  "Allow this tool for the rest of the session",
			value: "always",
		},
		permissionItem{
			title: "Deny",
			desc:  "Do not execute this tool",
			value: "no",
		},
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = fmt.Sprintf("🔐 Permission Required: %s", toolName)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(true)

	return PermissionPromptModel{
		toolName: toolName,
		args:     args,
		list:     l,
	}
}

func (m PermissionPromptModel) Init() tea.Cmd {
	return nil
}

func (m PermissionPromptModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			// Get selected item
			if i, ok := m.list.SelectedItem().(permissionItem); ok {
				m.choice = i.value
				m.quitting = true
				return m, tea.Quit
			}
		case "ctrl+c", "q", "esc":
			m.choice = "no"
			m.quitting = true
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := lipgloss.NewStyle().GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m PermissionPromptModel) View() string {
	if m.quitting {
		return ""
	}

	argsView := ""
	if len(m.args) > 0 {
		argsView = "\n\nArguments:\n"
		for key, value := range m.args {
			argsView += fmt.Sprintf("  %s: %v\n", key, value)
		}
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		argsView,
		"\n",
		m.list.View(),
	)
}

// GetChoice returns the user's choice
func (m PermissionPromptModel) GetChoice() string {
	return m.choice
}

// PromptWithTUI displays a permission prompt using bubbletea
func PromptWithTUI(toolName string, args map[string]interface{}) (bool, error) {
	m := NewPermissionPromptModel(toolName, args)
	p := tea.NewProgram(m)

	finalModel, err := p.Run()
	if err != nil {
		return false, fmt.Errorf("permission prompt failed: %w", err)
	}

	result := finalModel.(PermissionPromptModel)
	choice := result.GetChoice()

	return choice == "yes" || choice == "always", nil
}
