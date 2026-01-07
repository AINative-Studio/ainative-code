package rlhf

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FeedbackPromptModel represents the feedback prompt UI component
type FeedbackPromptModel struct {
	interactionID string
	ratingInput   textinput.Model
	commentInput  textinput.Model
	currentField  int
	rating        int
	comment       string
	submitted     bool
	dismissed     bool
	width         int
	height        int
}

// FeedbackResult represents the result of a feedback prompt
type FeedbackResult struct {
	InteractionID string
	Rating        int
	Comment       string
	Dismissed     bool
}

// NewFeedbackPromptModel creates a new feedback prompt model
func NewFeedbackPromptModel(interactionID string) FeedbackPromptModel {
	ratingInput := textinput.New()
	ratingInput.Placeholder = "1-5"
	ratingInput.CharLimit = 1
	ratingInput.Width = 5
	ratingInput.Focus()

	commentInput := textinput.New()
	commentInput.Placeholder = "Optional feedback..."
	commentInput.CharLimit = 500
	commentInput.Width = 50

	return FeedbackPromptModel{
		interactionID: interactionID,
		ratingInput:   ratingInput,
		commentInput:  commentInput,
		currentField:  0,
	}
}

// Init initializes the model
func (m FeedbackPromptModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages
func (m FeedbackPromptModel) Update(msg tea.Msg) (FeedbackPromptModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.dismissed = true
			return m, tea.Quit

		case "enter":
			if m.currentField == 0 {
				// Move to comment field
				m.ratingInput.Blur()
				m.commentInput.Focus()
				m.currentField = 1
				return m, textinput.Blink
			} else {
				// Submit feedback
				m.submitted = true
				return m, tea.Quit
			}

		case "tab", "shift+tab":
			if m.currentField == 0 {
				m.ratingInput.Blur()
				m.commentInput.Focus()
				m.currentField = 1
			} else {
				m.commentInput.Blur()
				m.ratingInput.Focus()
				m.currentField = 0
			}
			return m, textinput.Blink
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	// Update the active input
	if m.currentField == 0 {
		m.ratingInput, cmd = m.ratingInput.Update(msg)
	} else {
		m.commentInput, cmd = m.commentInput.Update(msg)
	}

	return m, cmd
}

// View renders the feedback prompt
func (m FeedbackPromptModel) View() string {
	if m.submitted || m.dismissed {
		return ""
	}

	var sb strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")).
		Padding(0, 1)

	sb.WriteString(titleStyle.Render("How would you rate this response?"))
	sb.WriteString("\n\n")

	// Rating field
	fieldStyle := lipgloss.NewStyle().Padding(0, 1)
	sb.WriteString(fieldStyle.Render("Rating (1-5): " + m.ratingInput.View()))
	sb.WriteString("\n\n")

	// Comment field
	sb.WriteString(fieldStyle.Render("Comment: " + m.commentInput.View()))
	sb.WriteString("\n\n")

	// Instructions
	instructionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Padding(0, 1)

	sb.WriteString(instructionStyle.Render("Tab to switch fields • Enter to submit • Esc to skip"))

	// Wrap in a box
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("12")).
		Padding(1, 2).
		Width(60)

	return boxStyle.Render(sb.String())
}

// GetResult returns the feedback result
func (m FeedbackPromptModel) GetResult() *FeedbackResult {
	if m.dismissed {
		return &FeedbackResult{
			InteractionID: m.interactionID,
			Dismissed:     true,
		}
	}

	// Parse rating
	rating := 0
	if m.ratingInput.Value() != "" {
		fmt.Sscanf(m.ratingInput.Value(), "%d", &rating)
	}

	// Validate rating
	if rating < 1 || rating > 5 {
		rating = 0
	}

	return &FeedbackResult{
		InteractionID: m.interactionID,
		Rating:        rating,
		Comment:       m.commentInput.Value(),
		Dismissed:     false,
	}
}

// IsSubmitted returns whether the feedback was submitted
func (m FeedbackPromptModel) IsSubmitted() bool {
	return m.submitted
}

// IsDismissed returns whether the prompt was dismissed
func (m FeedbackPromptModel) IsDismissed() bool {
	return m.dismissed
}
