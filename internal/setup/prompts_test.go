package setup

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestArrowKeyNavigation(t *testing.T) {
	tests := []struct {
		name          string
		currentStep   Step
		initialCursor int
		keyPress      string
		expectedCursor int
	}{
		{
			name:          "Down arrow on provider selection from first option",
			currentStep:   StepProvider,
			initialCursor: 0,
			keyPress:      "down",
			expectedCursor: 1,
		},
		{
			name:          "Up arrow on provider selection from second option",
			currentStep:   StepProvider,
			initialCursor: 1,
			keyPress:      "up",
			expectedCursor: 0,
		},
		{
			name:          "Down arrow on last option should stay at last",
			currentStep:   StepProvider,
			initialCursor: 4, // Last provider option (Ollama) - 5 providers total (0-4)
			keyPress:      "down",
			expectedCursor: 4,
		},
		{
			name:          "Up arrow on first option should stay at first",
			currentStep:   StepProvider,
			initialCursor: 0,
			keyPress:      "up",
			expectedCursor: 0,
		},
		{
			name:          "Down arrow with 'j' key (vim-style)",
			currentStep:   StepAnthropicModel,
			initialCursor: 0,
			keyPress:      "j",
			expectedCursor: 1,
		},
		{
			name:          "Up arrow with 'k' key (vim-style)",
			currentStep:   StepAnthropicModel,
			initialCursor: 1,
			keyPress:      "k",
			expectedCursor: 0,
		},
		{
			name:          "Navigate through OpenAI models",
			currentStep:   StepOpenAIModel,
			initialCursor: 0,
			keyPress:      "down",
			expectedCursor: 1,
		},
		{
			name:          "Navigate through Google models",
			currentStep:   StepGoogleModel,
			initialCursor: 0,
			keyPress:      "down",
			expectedCursor: 1,
		},
		{
			name:          "Navigate through color schemes",
			currentStep:   StepColorScheme,
			initialCursor: 1,
			keyPress:      "up",
			expectedCursor: 0,
		},
		{
			name:          "Navigate through Meta Llama models",
			currentStep:   StepMetaLlamaModel,
			initialCursor: 0,
			keyPress:      "down",
			expectedCursor: 1,
		},
		{
			name:          "Navigate to last Meta Llama model",
			currentStep:   StepMetaLlamaModel,
			initialCursor: 3,
			keyPress:      "down",
			expectedCursor: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewPromptModel()
			m.currentStep = tt.currentStep
			m.cursor = tt.initialCursor

			// Simulate key press
			msg := tea.KeyMsg{Type: tea.KeyRunes}
			switch tt.keyPress {
			case "up":
				msg.Type = tea.KeyUp
			case "down":
				msg.Type = tea.KeyDown
			case "k":
				msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
			case "j":
				msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
			}

			// Update the model
			updatedModel, _ := m.Update(msg)
			updated := updatedModel.(PromptModel)

			if updated.cursor != tt.expectedCursor {
				t.Errorf("Expected cursor to be %d, got %d", tt.expectedCursor, updated.cursor)
			}
		})
	}
}

func TestArrowKeysDoNotWorkOnTextInputSteps(t *testing.T) {
	textInputSteps := []Step{
		StepAnthropicAPIKey,
		StepOpenAIAPIKey,
		StepGoogleAPIKey,
		StepOllamaURL,
		StepOllamaModel,
		StepMetaLlamaAPIKey,
		StepAINativeAPIKey,
	}

	for _, step := range textInputSteps {
		t.Run(step.String(), func(t *testing.T) {
			m := NewPromptModel()
			m.currentStep = step
			m.cursor = 0

			// Try to move down with arrow key
			msg := tea.KeyMsg{Type: tea.KeyDown}
			updatedModel, _ := m.Update(msg)
			updated := updatedModel.(PromptModel)

			// Cursor should NOT change on text input steps
			if updated.cursor != 0 {
				t.Errorf("Arrow keys should not work on text input step %v, but cursor changed from 0 to %d", step, updated.cursor)
			}
		})
	}
}

func TestGetChoiceCount(t *testing.T) {
	tests := []struct {
		step          Step
		expectedCount int
	}{
		{StepProvider, 5}, // Anthropic, OpenAI, Google, Meta Llama, Ollama
		{StepAnthropicModel, 4},
		{StepOpenAIModel, 3},
		{StepGoogleModel, 2},
		{StepMetaLlamaModel, 4}, // Added Meta Llama model step
		{StepColorScheme, 3},
		{StepAnthropicAPIKey, 0}, // Text input step
		{StepExtendedThinking, 0}, // Yes/no step
	}

	for _, tt := range tests {
		t.Run(tt.step.String(), func(t *testing.T) {
			m := NewPromptModel()
			m.currentStep = tt.step

			count := m.getChoiceCount()
			if count != tt.expectedCount {
				t.Errorf("Expected choice count %d for step %v, got %d", tt.expectedCount, tt.step, count)
			}
		})
	}
}

func TestCursorResetOnStepTransition(t *testing.T) {
	m := NewPromptModel()
	m.currentStep = StepProvider
	m.cursor = 2 // Select third option
	m.Selections["provider"] = "google"

	// Transition to next step
	updatedModel, _ := m.nextStep()
	updated := updatedModel.(PromptModel)

	// Cursor should be reset to 0
	if updated.cursor != 0 {
		t.Errorf("Expected cursor to be reset to 0 after step transition, got %d", updated.cursor)
	}
}

// Helper method for testing
func (s Step) String() string {
	switch s {
	case StepProvider:
		return "StepProvider"
	case StepAnthropicAPIKey:
		return "StepAnthropicAPIKey"
	case StepAnthropicModel:
		return "StepAnthropicModel"
	case StepExtendedThinking:
		return "StepExtendedThinking"
	case StepOpenAIAPIKey:
		return "StepOpenAIAPIKey"
	case StepOpenAIModel:
		return "StepOpenAIModel"
	case StepGoogleAPIKey:
		return "StepGoogleAPIKey"
	case StepGoogleModel:
		return "StepGoogleModel"
	case StepOllamaURL:
		return "StepOllamaURL"
	case StepOllamaModel:
		return "StepOllamaModel"
	case StepMetaLlamaAPIKey:
		return "StepMetaLlamaAPIKey"
	case StepMetaLlamaModel:
		return "StepMetaLlamaModel"
	case StepAINativeLogin:
		return "StepAINativeLogin"
	case StepAINativeAPIKey:
		return "StepAINativeAPIKey"
	case StepColorScheme:
		return "StepColorScheme"
	case StepPromptCaching:
		return "StepPromptCaching"
	case StepComplete:
		return "StepComplete"
	default:
		return "Unknown"
	}
}
