package setup

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Step represents each step in the wizard
type Step int

const (
	StepProvider Step = iota
	StepAnthropicAPIKey
	StepAnthropicModel
	StepExtendedThinking
	StepOpenAIAPIKey
	StepOpenAIModel
	StepGoogleAPIKey
	StepGoogleModel
	StepOllamaURL
	StepOllamaModel
	StepAINativeLogin
	StepAINativeAPIKey
	StepColorScheme
	StepPromptCaching
	StepComplete
)

// PromptModel represents the state of the interactive prompts
type PromptModel struct {
	currentStep Step
	Selections  map[string]interface{}
	textInput   textinput.Model
	cursor      int
	err         error
}

// NewPromptModel creates a new prompt model
func NewPromptModel() PromptModel {
	ti := textinput.New()
	ti.Placeholder = "Enter your choice..."
	ti.CharLimit = 200
	ti.Width = 50

	return PromptModel{
		currentStep: StepProvider,
		Selections:  make(map[string]interface{}),
		textInput:   ti,
		cursor:      0,
	}
}

// Init initializes the model
func (m PromptModel) Init() tea.Cmd {
	// Focus textinput if we're starting on a text input step
	if m.isTextInputStep() {
		m.textInput.Focus()
		return textinput.Blink
	}
	return nil
}

// Update handles messages
func (m PromptModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "enter":
			return m.handleEnter()

		case "up", "k":
			if !m.isTextInputStep() && m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if !m.isTextInputStep() && m.cursor < m.getChoiceCount()-1 {
				m.cursor++
			}

		case "y", "Y":
			// Handle yes/no prompts
			if m.isYesNoStep() {
				m.Selections[m.getStepKey()] = true
				return m.nextStep()
			}

		case "n", "N":
			// Handle yes/no prompts
			if m.isYesNoStep() {
				m.Selections[m.getStepKey()] = false
				return m.nextStep()
			}

		default:
			if m.isTextInputStep() {
				var cmd tea.Cmd
				m.textInput, cmd = m.textInput.Update(msg)
				return m, cmd
			}
		}

	case error:
		m.err = msg
		return m, nil
	}

	return m, nil
}

// View renders the UI
func (m PromptModel) View() string {
	if m.currentStep == StepComplete {
		return ""
	}

	var s strings.Builder

	// Title style
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		MarginBottom(1)

	// Question style
	questionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("250")).
		MarginBottom(1)

	// Render based on current step
	switch m.currentStep {
	case StepProvider:
		s.WriteString(titleStyle.Render("LLM Provider Selection"))
		s.WriteString("\n\n")
		s.WriteString(questionStyle.Render("Which LLM provider would you like to use?"))
		s.WriteString("\n\n")
		s.WriteString(m.renderChoices([]string{
			"Anthropic (Claude)",
			"OpenAI (GPT)",
			"Google (Gemini)",
			"Ollama (Local)",
		}))
		s.WriteString("\n\n")
		s.WriteString(m.renderHelpText("Use arrow keys to navigate, Enter to select"))

	case StepAnthropicAPIKey:
		s.WriteString(titleStyle.Render("Anthropic Configuration"))
		s.WriteString("\n\n")
		s.WriteString(questionStyle.Render("Enter your Anthropic API key:"))
		s.WriteString("\n")
		s.WriteString(m.renderTextInput())
		s.WriteString("\n\n")
		s.WriteString(m.renderHelpText("Get your API key from: https://console.anthropic.com/"))

	case StepAnthropicModel:
		s.WriteString(titleStyle.Render("Anthropic Model Selection"))
		s.WriteString("\n\n")
		s.WriteString(questionStyle.Render("Which Claude model would you like to use?"))
		s.WriteString("\n\n")
		s.WriteString(m.renderChoices([]string{
			"claude-3-5-sonnet-20241022 (Recommended)",
			"claude-3-opus-20240229",
			"claude-3-sonnet-20240229",
			"claude-3-haiku-20240307",
		}))
		s.WriteString("\n\n")
		s.WriteString(m.renderHelpText("Sonnet 3.5 offers the best balance of performance and cost"))

	case StepExtendedThinking:
		s.WriteString(titleStyle.Render("Extended Thinking"))
		s.WriteString("\n\n")
		s.WriteString(questionStyle.Render("Enable extended thinking mode for complex reasoning?"))
		s.WriteString("\n")
		s.WriteString(questionStyle.Render("(This allows Claude to think longer on complex problems)"))
		s.WriteString("\n\n")
		s.WriteString(m.renderYesNo())
		s.WriteString("\n\n")
		s.WriteString(m.renderHelpText("Extended thinking provides deeper analysis but may be slower"))

	case StepOpenAIAPIKey:
		s.WriteString(titleStyle.Render("OpenAI Configuration"))
		s.WriteString("\n\n")
		s.WriteString(questionStyle.Render("Enter your OpenAI API key:"))
		s.WriteString("\n")
		s.WriteString(m.renderTextInput())
		s.WriteString("\n\n")
		s.WriteString(m.renderHelpText("Get your API key from: https://platform.openai.com/api-keys"))

	case StepOpenAIModel:
		s.WriteString(titleStyle.Render("OpenAI Model Selection"))
		s.WriteString("\n\n")
		s.WriteString(questionStyle.Render("Which OpenAI model would you like to use?"))
		s.WriteString("\n\n")
		s.WriteString(m.renderChoices([]string{
			"gpt-4-turbo-preview (Recommended)",
			"gpt-4",
			"gpt-3.5-turbo",
		}))
		s.WriteString("\n\n")
		s.WriteString(m.renderHelpText("GPT-4 Turbo offers the best performance"))

	case StepGoogleAPIKey:
		s.WriteString(titleStyle.Render("Google Gemini Configuration"))
		s.WriteString("\n\n")
		s.WriteString(questionStyle.Render("Enter your Google API key:"))
		s.WriteString("\n")
		s.WriteString(m.renderTextInput())
		s.WriteString("\n\n")
		s.WriteString(m.renderHelpText("Get your API key from: https://makersuite.google.com/app/apikey"))

	case StepGoogleModel:
		s.WriteString(titleStyle.Render("Google Model Selection"))
		s.WriteString("\n\n")
		s.WriteString(questionStyle.Render("Which Gemini model would you like to use?"))
		s.WriteString("\n\n")
		s.WriteString(m.renderChoices([]string{
			"gemini-pro (Recommended)",
			"gemini-pro-vision",
		}))
		s.WriteString("\n\n")
		s.WriteString(m.renderHelpText("Gemini Pro is optimized for text tasks"))

	case StepOllamaURL:
		s.WriteString(titleStyle.Render("Ollama Configuration"))
		s.WriteString("\n\n")
		s.WriteString(questionStyle.Render("Enter Ollama server URL:"))
		s.WriteString("\n")
		s.WriteString(m.renderTextInput())
		s.WriteString("\n\n")
		s.WriteString(m.renderHelpText("Default: http://localhost:11434"))

	case StepOllamaModel:
		s.WriteString(titleStyle.Render("Ollama Model Selection"))
		s.WriteString("\n\n")
		s.WriteString(questionStyle.Render("Enter the model name (e.g., llama2, codellama):"))
		s.WriteString("\n")
		s.WriteString(m.renderTextInput())
		s.WriteString("\n\n")
		s.WriteString(m.renderHelpText("Make sure the model is already pulled in Ollama"))

	case StepAINativeLogin:
		s.WriteString(titleStyle.Render("AINative Platform"))
		s.WriteString("\n\n")
		s.WriteString(questionStyle.Render("Would you like to connect to the AINative platform?"))
		s.WriteString("\n")
		s.WriteString(questionStyle.Render("(Optional - enables advanced features like ZeroDB, Design tools, etc.)"))
		s.WriteString("\n\n")
		s.WriteString(m.renderYesNo())
		s.WriteString("\n\n")
		s.WriteString(m.renderHelpText("You can configure this later with: ainative-code auth login"))

	case StepAINativeAPIKey:
		s.WriteString(titleStyle.Render("AINative Platform API Key"))
		s.WriteString("\n\n")
		s.WriteString(questionStyle.Render("Enter your AINative API key:"))
		s.WriteString("\n")
		s.WriteString(m.renderTextInput())
		s.WriteString("\n\n")
		s.WriteString(m.renderHelpText("Get your API key from: https://ainative.studio/dashboard"))

	case StepColorScheme:
		s.WriteString(titleStyle.Render("Color Scheme"))
		s.WriteString("\n\n")
		s.WriteString(questionStyle.Render("Choose your preferred color scheme:"))
		s.WriteString("\n\n")
		s.WriteString(m.renderChoices([]string{
			"Auto (Match terminal)",
			"Light",
			"Dark",
		}))

	case StepPromptCaching:
		s.WriteString(titleStyle.Render("Prompt Caching"))
		s.WriteString("\n\n")
		s.WriteString(questionStyle.Render("Enable prompt caching for faster responses?"))
		s.WriteString("\n")
		s.WriteString(questionStyle.Render("(Caches common prompts to reduce latency and costs)"))
		s.WriteString("\n\n")
		s.WriteString(m.renderYesNo())
		s.WriteString("\n\n")
		s.WriteString(m.renderHelpText("Recommended for most users"))
	}

	return s.String()
}

// Helper methods

func (m PromptModel) renderChoices(choices []string) string {
	var s strings.Builder

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Bold(true)

	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("250"))

	for i, choice := range choices {
		cursor := "  "
		if m.cursor == i {
			cursor = "> "
			s.WriteString(selectedStyle.Render(cursor + choice))
		} else {
			s.WriteString(normalStyle.Render(cursor + choice))
		}
		s.WriteString("\n")
	}

	return s.String()
}

func (m PromptModel) renderTextInput() string {
	// Don't focus here - focusing should happen in Init() or when transitioning steps
	return m.textInput.View()
}

func (m PromptModel) renderYesNo() string {
	yesStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true)
	noStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("250"))

	return normalStyle.Render("[") + yesStyle.Render("y") + normalStyle.Render("es / ") +
		noStyle.Render("n") + normalStyle.Render("o]")
}

func (m PromptModel) renderHelpText(text string) string {
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true)

	return helpStyle.Render(text)
}

func (m PromptModel) handleEnter() (tea.Model, tea.Cmd) {
	switch m.currentStep {
	case StepProvider:
		providers := []string{"anthropic", "openai", "google", "ollama"}
		m.Selections["provider"] = providers[m.cursor]

	case StepAnthropicAPIKey:
		m.Selections["anthropic_api_key"] = m.textInput.Value()
		m.textInput.SetValue("")

	case StepAnthropicModel:
		models := []string{
			"claude-3-5-sonnet-20241022",
			"claude-3-opus-20240229",
			"claude-3-sonnet-20240229",
			"claude-3-haiku-20240307",
		}
		m.Selections["anthropic_model"] = models[m.cursor]

	case StepOpenAIAPIKey:
		m.Selections["openai_api_key"] = m.textInput.Value()
		m.textInput.SetValue("")

	case StepOpenAIModel:
		models := []string{
			"gpt-4-turbo-preview",
			"gpt-4",
			"gpt-3.5-turbo",
		}
		m.Selections["openai_model"] = models[m.cursor]

	case StepGoogleAPIKey:
		m.Selections["google_api_key"] = m.textInput.Value()
		m.textInput.SetValue("")

	case StepGoogleModel:
		models := []string{"gemini-pro", "gemini-pro-vision"}
		m.Selections["google_model"] = models[m.cursor]

	case StepOllamaURL:
		url := m.textInput.Value()
		if url == "" {
			url = "http://localhost:11434"
		}
		m.Selections["ollama_url"] = url
		m.textInput.SetValue("")

	case StepOllamaModel:
		m.Selections["ollama_model"] = m.textInput.Value()
		m.textInput.SetValue("")

	case StepAINativeAPIKey:
		m.Selections["ainative_api_key"] = m.textInput.Value()
		m.textInput.SetValue("")

	case StepColorScheme:
		schemes := []string{"auto", "light", "dark"}
		m.Selections["color_scheme"] = schemes[m.cursor]

	default:
		// For yes/no steps, handled by key press
		return m, nil
	}

	return m.nextStep()
}

func (m PromptModel) nextStep() (tea.Model, tea.Cmd) {
	m.cursor = 0

	// Determine next step based on selections
	switch m.currentStep {
	case StepProvider:
		provider := m.Selections["provider"].(string)
		switch provider {
		case "anthropic":
			m.currentStep = StepAnthropicAPIKey
		case "openai":
			m.currentStep = StepOpenAIAPIKey
		case "google":
			m.currentStep = StepGoogleAPIKey
		case "ollama":
			m.currentStep = StepOllamaURL
		}

	case StepAnthropicAPIKey:
		m.currentStep = StepAnthropicModel

	case StepAnthropicModel:
		m.currentStep = StepExtendedThinking

	case StepExtendedThinking:
		m.currentStep = StepAINativeLogin

	case StepOpenAIAPIKey:
		m.currentStep = StepOpenAIModel

	case StepOpenAIModel:
		m.currentStep = StepAINativeLogin

	case StepGoogleAPIKey:
		m.currentStep = StepGoogleModel

	case StepGoogleModel:
		m.currentStep = StepAINativeLogin

	case StepOllamaURL:
		m.currentStep = StepOllamaModel

	case StepOllamaModel:
		m.currentStep = StepAINativeLogin

	case StepAINativeLogin:
		if loginEnabled, ok := m.Selections["ainative_login"].(bool); ok && loginEnabled {
			m.currentStep = StepAINativeAPIKey
		} else {
			m.currentStep = StepColorScheme
		}

	case StepAINativeAPIKey:
		m.currentStep = StepColorScheme

	case StepColorScheme:
		m.currentStep = StepPromptCaching

	case StepPromptCaching:
		m.currentStep = StepComplete
		return m, tea.Quit

	default:
		return m, tea.Quit
	}

	// Focus textinput if transitioning to a text input step
	if m.isTextInputStep() {
		m.textInput.Focus()
		return m, textinput.Blink
	}

	// Blur textinput if transitioning away from a text input step
	m.textInput.Blur()

	return m, nil
}

func (m PromptModel) isTextInputStep() bool {
	return m.currentStep == StepAnthropicAPIKey ||
		m.currentStep == StepOpenAIAPIKey ||
		m.currentStep == StepGoogleAPIKey ||
		m.currentStep == StepOllamaURL ||
		m.currentStep == StepOllamaModel ||
		m.currentStep == StepAINativeAPIKey
}

func (m PromptModel) isYesNoStep() bool {
	return m.currentStep == StepExtendedThinking ||
		m.currentStep == StepAINativeLogin ||
		m.currentStep == StepPromptCaching
}

func (m PromptModel) getStepKey() string {
	switch m.currentStep {
	case StepExtendedThinking:
		return "extended_thinking"
	case StepAINativeLogin:
		return "ainative_login"
	case StepPromptCaching:
		return "prompt_caching"
	default:
		return ""
	}
}

func (m PromptModel) getChoiceCount() int {
	switch m.currentStep {
	case StepProvider:
		return 4 // Anthropic, OpenAI, Google, Ollama
	case StepAnthropicModel:
		return 4 // 4 Claude models
	case StepOpenAIModel:
		return 3 // 3 GPT models
	case StepGoogleModel:
		return 2 // 2 Gemini models
	case StepColorScheme:
		return 3 // Auto, Light, Dark
	default:
		return 0
	}
}

// SummaryModel displays the configuration summary
type SummaryModel struct {
	Config     interface{}
	Selections map[string]interface{}
	Confirmed  bool
	cursor     int
}

// NewSummaryModel creates a new summary model
func NewSummaryModel(cfg interface{}, selections map[string]interface{}) SummaryModel {
	return SummaryModel{
		Config:     cfg,
		Selections: selections,
		Confirmed:  false,
		cursor:     0,
	}
}

// Init initializes the summary model
func (m SummaryModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for summary
func (m SummaryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "y", "Y":
			m.Confirmed = true
			return m, tea.Quit

		case "n", "N":
			m.Confirmed = false
			return m, tea.Quit
		}
	}

	return m, nil
}

// View renders the summary
func (m SummaryModel) View() string {
	var s strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		MarginBottom(1)

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("250")).
		Bold(true)

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("46"))

	s.WriteString(titleStyle.Render("Configuration Summary"))
	s.WriteString("\n\n")

	// Display selections
	provider := m.Selections["provider"].(string)
	s.WriteString(labelStyle.Render("LLM Provider: "))
	s.WriteString(valueStyle.Render(provider))
	s.WriteString("\n")

	switch provider {
	case "anthropic":
		s.WriteString(labelStyle.Render("Model: "))
		s.WriteString(valueStyle.Render(m.Selections["anthropic_model"].(string)))
		s.WriteString("\n")
		if et, ok := m.Selections["extended_thinking"].(bool); ok {
			s.WriteString(labelStyle.Render("Extended Thinking: "))
			s.WriteString(valueStyle.Render(fmt.Sprintf("%v", et)))
			s.WriteString("\n")
		}

	case "openai":
		s.WriteString(labelStyle.Render("Model: "))
		s.WriteString(valueStyle.Render(m.Selections["openai_model"].(string)))
		s.WriteString("\n")

	case "google":
		s.WriteString(labelStyle.Render("Model: "))
		s.WriteString(valueStyle.Render(m.Selections["google_model"].(string)))
		s.WriteString("\n")

	case "ollama":
		s.WriteString(labelStyle.Render("Server: "))
		s.WriteString(valueStyle.Render(m.Selections["ollama_url"].(string)))
		s.WriteString("\n")
		s.WriteString(labelStyle.Render("Model: "))
		s.WriteString(valueStyle.Render(m.Selections["ollama_model"].(string)))
		s.WriteString("\n")
	}

	if loginEnabled, ok := m.Selections["ainative_login"].(bool); ok {
		s.WriteString(labelStyle.Render("AINative Platform: "))
		s.WriteString(valueStyle.Render(fmt.Sprintf("%v", loginEnabled)))
		s.WriteString("\n")
	}

	s.WriteString("\n")
	s.WriteString(labelStyle.Render("Confirm and save this configuration? [y/n]: "))

	return s.String()
}
