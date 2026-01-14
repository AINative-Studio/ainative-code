package tui

import (
	"github.com/AINative-studio/ainative-code/internal/rlhf"
	"github.com/AINative-studio/ainative-code/internal/tui/dialogs"
	"github.com/AINative-studio/ainative-code/internal/tui/layout"
	"github.com/AINative-studio/ainative-code/internal/tui/syntax"
	"github.com/AINative-studio/ainative-code/internal/tui/theme"
	"github.com/AINative-studio/ainative-code/internal/tui/toast"
	"github.com/AINative-studio/ainative-code/pkg/lsp"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/viewport"
)

// Model represents the TUI application state
type Model struct {
	viewport         viewport.Model
	textInput        textinput.Model
	messages         []Message
	thinkingState    *ThinkingState
	thinkingConfig   ThinkingConfig
	width            int
	height           int
	ready            bool
	quitting         bool
	streaming        bool
	err              error
	lspClient        *lsp.Client
	lspEnabled       bool
	currentDocument  string
	completionItems  []lsp.CompletionItem
	showCompletion   bool
	completionIndex  int
	hoverInfo        *lsp.Hover
	showHover        bool
	navigationResult []lsp.Location
	showNavigation   bool

	// RLHF auto-collection (TASK-064)
	rlhfCollector       *rlhf.Collector
	rlhfEnabled         bool
	lastInteractionID   string
	showFeedbackPrompt  bool
	feedbackPromptModel *rlhf.FeedbackPromptModel

	// Syntax highlighting (TASK-022)
	syntaxHighlighter *syntax.Highlighter
	syntaxEnabled     bool

	// Dialog system (TASK-133)
	dialogManager *dialogs.DialogManager

	// Layout management (TASK-132)
	layoutManager layout.LayoutManager

	// Theme management (TASK-137)
	themeManager *theme.ThemeManager

	// Toast notification system (TASK-138)
	toastManager *toast.ToastManager
}

// Message represents a chat message
type Message struct {
	Role    string // "user", "assistant", "system"
	Content string
}

// NewModel creates a new TUI model
func NewModel() Model {
	ti := textinput.New()
	ti.Placeholder = "Type your message..."
	ti.Focus()
	ti.CharLimit = 0
	ti.Width = 50

	// Initialize syntax highlighter with AINative branding
	highlighter := syntax.NewHighlighter(syntax.AINativeConfig())

	// Initialize dialog manager
	dialogMgr := dialogs.NewDialogManager()

	// Initialize theme manager with built-in themes
	themeMgr := theme.NewThemeManager()
	theme.RegisterBuiltinThemes(themeMgr)
	themeMgr.SetTheme("AINative") // Set AINative as default
	themeMgr.LoadConfig()         // Try to load saved theme preference

	// Initialize toast manager
	toastMgr := toast.NewToastManager()
	toastMgr.SetMaxToasts(3)              // Max 3 visible toasts
	toastMgr.SetPosition(toast.TopRight)  // Default position

	return Model{
		textInput:         ti,
		messages:          []Message{},
		thinkingState:     NewThinkingState(),
		thinkingConfig:    DefaultThinkingConfig(),
		ready:             false,
		quitting:          false,
		streaming:         false,
		lspClient:         nil,
		lspEnabled:        false,
		completionItems:   []lsp.CompletionItem{},
		showCompletion:    false,
		completionIndex:   0,
		hoverInfo:         nil,
		showHover:         false,
		navigationResult:  []lsp.Location{},
		showNavigation:    false,
		syntaxHighlighter: highlighter,
		syntaxEnabled:     true,
		dialogManager:     dialogMgr,
		themeManager:      themeMgr,
		toastManager:      toastMgr,
	}
}

// NewModelWithLSP creates a new TUI model with LSP enabled
func NewModelWithLSP(workspace string) Model {
	m := NewModel()
	m.lspClient = lsp.NewClient()
	m.lspEnabled = true
	m.currentDocument = workspace
	return m
}

// SetSize updates the model dimensions
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height

	// Initialize layout manager on first call
	if m.layoutManager == nil {
		m.layoutManager = m.createLayoutManager()
	}

	// Update layout manager with new size
	m.layoutManager.SetAvailableSize(width, height)
	_ = m.layoutManager.RecalculateLayout()

	// Apply calculated bounds to components
	viewportBounds := m.layoutManager.GetComponentBounds("viewport")
	inputBounds := m.layoutManager.GetComponentBounds("input")

	if !m.ready {
		// Initialize viewport with calculated dimensions
		m.viewport = viewport.New(viewportBounds.Width, viewportBounds.Height)
		m.viewport.YPosition = viewportBounds.Y
		m.ready = true
	} else {
		// Update existing viewport dimensions
		m.viewport.Width = viewportBounds.Width
		m.viewport.Height = viewportBounds.Height
		m.viewport.YPosition = viewportBounds.Y
	}

	// Update text input width (subtract 2 for prompt "► ")
	if inputBounds.Width > 2 {
		m.textInput.Width = inputBounds.Width - 2
	} else {
		m.textInput.Width = inputBounds.Width
	}

	// Update dialog manager size
	m.dialogManager.SetSize(width, height)

	// Update toast manager size
	m.toastManager.SetSize(width, height)

	// Update toast manager size
	m.toastManager.SetSize(width, height)
}

// createLayoutManager initializes the layout manager with component constraints
func (m *Model) createLayoutManager() layout.LayoutManager {
	// Create vertical box layout
	vbox := layout.NewBoxLayout(layout.Vertical)
	mgr := layout.NewManager(vbox)

	// Register viewport - flexible component that grows to fill space
	_ = mgr.RegisterComponent("viewport", layout.FlexConstraints(10, 1))

	// Register input area - fixed height (3 lines: separator + input + padding)
	_ = mgr.RegisterComponent("input", layout.Constraints{
		MinHeight: 3,
		MaxHeight: 3,
		Grow:      false,
		Shrink:    false,
		Weight:    0,
	})

	// Register status bar - fixed height (1 line)
	_ = mgr.RegisterComponent("statusbar", layout.FixedConstraints(0, 1))

	return mgr
}

// AddMessage adds a new message to the conversation
func (m *Model) AddMessage(role, content string) {
	m.messages = append(m.messages, Message{
		Role:    role,
		Content: content,
	})
}

// ClearMessages removes all messages
func (m *Model) ClearMessages() {
	m.messages = []Message{}
}

// GetUserInput returns and clears the current input
func (m *Model) GetUserInput() string {
	input := m.textInput.Value()
	m.textInput.SetValue("")
	return input
}

// SetError sets an error state
func (m *Model) SetError(err error) {
	m.err = err
}

// ClearError clears the error state
func (m *Model) ClearError() {
	m.err = nil
}

// IsReady returns whether the TUI is ready to display
func (m *Model) IsReady() bool {
	return m.ready
}

// IsQuitting returns whether the TUI is quitting
func (m *Model) IsQuitting() bool {
	return m.quitting
}

// IsStreaming returns whether a response is being streamed
func (m *Model) IsStreaming() bool {
	return m.streaming
}

// SetStreaming sets the streaming state
func (m *Model) SetStreaming(streaming bool) {
	m.streaming = streaming
}

// SetQuitting sets the quitting state
func (m *Model) SetQuitting(quitting bool) {
	m.quitting = quitting
}

// Thinking-related methods

// ToggleThinkingDisplay toggles the display of thinking blocks
func (m *Model) ToggleThinkingDisplay() {
	m.thinkingState.ToggleDisplay()
}

// AddThinking adds a new thinking block
func (m *Model) AddThinking(content string, depth int) {
	m.thinkingState.AddThinkingBlock(content, depth)
}

// AppendThinking appends content to the current thinking block
func (m *Model) AppendThinking(content string) {
	m.thinkingState.AppendToCurrentBlock(content)
}

// CollapseAllThinking collapses all thinking blocks
func (m *Model) CollapseAllThinking() {
	m.thinkingState.CollapseAll()
}

// ExpandAllThinking expands all thinking blocks
func (m *Model) ExpandAllThinking() {
	m.thinkingState.ExpandAll()
}

// ClearThinking removes all thinking blocks
func (m *Model) ClearThinking() {
	m.thinkingState.ClearBlocks()
}

// IsThinkingVisible returns whether thinking blocks are visible
func (m *Model) IsThinkingVisible() bool {
	return m.thinkingState.ShowThinking
}

// GetThinkingState returns the thinking state
func (m *Model) GetThinkingState() *ThinkingState {
	return m.thinkingState
}

// GetThinkingConfig returns the thinking configuration
func (m *Model) GetThinkingConfig() ThinkingConfig {
	return m.thinkingConfig
}

// SetThinkingConfig sets the thinking configuration
func (m *Model) SetThinkingConfig(config ThinkingConfig) {
	m.thinkingConfig = config
}

// LSP-related methods

// GetLSPClient returns the LSP client
func (m *Model) GetLSPClient() *lsp.Client {
	return m.lspClient
}

// IsLSPEnabled returns whether LSP is enabled
func (m *Model) IsLSPEnabled() bool {
	return m.lspEnabled && m.lspClient != nil
}

// GetLSPStatus returns the LSP connection status
func (m *Model) GetLSPStatus() lsp.ConnectionStatus {
	if m.lspClient == nil {
		return lsp.StatusDisconnected
	}
	return m.lspClient.GetStatus()
}

// SetCompletionItems sets the completion items
func (m *Model) SetCompletionItems(items []lsp.CompletionItem) {
	m.completionItems = items
	m.showCompletion = len(items) > 0
	m.completionIndex = 0
}

// ClearCompletion clears the completion popup
func (m *Model) ClearCompletion() {
	m.completionItems = []lsp.CompletionItem{}
	m.showCompletion = false
	m.completionIndex = 0
}

// NextCompletion moves to the next completion item
func (m *Model) NextCompletion() {
	if len(m.completionItems) > 0 {
		m.completionIndex = (m.completionIndex + 1) % len(m.completionItems)
	}
}

// PrevCompletion moves to the previous completion item
func (m *Model) PrevCompletion() {
	if len(m.completionItems) > 0 {
		m.completionIndex = (m.completionIndex - 1 + len(m.completionItems)) % len(m.completionItems)
	}
}

// GetSelectedCompletion returns the currently selected completion item
func (m *Model) GetSelectedCompletion() *lsp.CompletionItem {
	if len(m.completionItems) > 0 && m.completionIndex >= 0 && m.completionIndex < len(m.completionItems) {
		return &m.completionItems[m.completionIndex]
	}
	return nil
}

// SetHoverInfo sets the hover information
func (m *Model) SetHoverInfo(hover *lsp.Hover) {
	m.hoverInfo = hover
	m.showHover = hover != nil
}

// ClearHover clears the hover popup
func (m *Model) ClearHover() {
	m.hoverInfo = nil
	m.showHover = false
}

// SetNavigationResult sets the navigation result
func (m *Model) SetNavigationResult(locations []lsp.Location) {
	m.navigationResult = locations
	m.showNavigation = len(locations) > 0
}

// ClearNavigation clears the navigation result
func (m *Model) ClearNavigation() {
	m.navigationResult = []lsp.Location{}
	m.showNavigation = false
}

// GetShowCompletion returns whether the completion popup is shown
func (m *Model) GetShowCompletion() bool {
	return m.showCompletion
}

// GetShowHover returns whether the hover popup is shown
func (m *Model) GetShowHover() bool {
	return m.showHover
}

// GetHoverInfo returns the hover information
func (m *Model) GetHoverInfo() *lsp.Hover {
	return m.hoverInfo
}

// GetShowNavigation returns whether the navigation popup is shown
func (m *Model) GetShowNavigation() bool {
	return m.showNavigation
}

// GetNavigationResult returns the navigation results
func (m *Model) GetNavigationResult() []lsp.Location {
	return m.navigationResult
}

// SetValue sets the input value (for testing)
func (m *Model) SetValue(value string) {
	m.textInput.SetValue(value)
}

// Syntax highlighting methods

// EnableSyntaxHighlighting enables syntax highlighting
func (m *Model) EnableSyntaxHighlighting() {
	m.syntaxEnabled = true
}

// DisableSyntaxHighlighting disables syntax highlighting
func (m *Model) DisableSyntaxHighlighting() {
	m.syntaxEnabled = false
}

// IsSyntaxHighlightingEnabled returns whether syntax highlighting is enabled
func (m *Model) IsSyntaxHighlightingEnabled() bool {
	return m.syntaxEnabled
}

// GetSyntaxHighlighter returns the syntax highlighter
func (m *Model) GetSyntaxHighlighter() *syntax.Highlighter {
	return m.syntaxHighlighter
}

// RLHF-related methods (TASK-064)

// SetRLHFCollector sets the RLHF collector
func (m *Model) SetRLHFCollector(collector *rlhf.Collector) {
	m.rlhfCollector = collector
	m.rlhfEnabled = collector != nil
}

// CaptureInteraction captures an interaction for RLHF
func (m *Model) CaptureInteraction(prompt, response, modelID string) {
	if m.rlhfCollector == nil || !m.rlhfEnabled {
		return
	}

	interactionID := m.rlhfCollector.CaptureInteraction(prompt, response, modelID)
	m.lastInteractionID = interactionID

	// Check if we should prompt for feedback
	if m.rlhfCollector.ShouldPromptForFeedback() {
		m.showFeedbackPrompt = true
		model := rlhf.NewFeedbackPromptModel(interactionID)
		m.feedbackPromptModel = &model
	}
}

// RecordImplicitFeedback records an implicit feedback signal
func (m *Model) RecordImplicitFeedback(action rlhf.FeedbackAction) {
	if m.rlhfCollector == nil || !m.rlhfEnabled || m.lastInteractionID == "" {
		return
	}

	m.rlhfCollector.RecordImplicitFeedback(m.lastInteractionID, action)
}

// RecordExplicitFeedback records explicit user feedback
func (m *Model) RecordExplicitFeedback(interactionID string, score float64, feedback string) {
	if m.rlhfCollector == nil || !m.rlhfEnabled {
		return
	}

	m.rlhfCollector.RecordExplicitFeedback(interactionID, score, feedback)
}

// GetShowFeedbackPrompt returns whether the feedback prompt should be shown
func (m *Model) GetShowFeedbackPrompt() bool {
	return m.showFeedbackPrompt
}

// DismissFeedbackPrompt dismisses the feedback prompt
func (m *Model) DismissFeedbackPrompt() {
	m.showFeedbackPrompt = false
	m.feedbackPromptModel = nil
}

// GetFeedbackPromptModel returns the feedback prompt model
func (m *Model) GetFeedbackPromptModel() *rlhf.FeedbackPromptModel {
	return m.feedbackPromptModel
}

// Theme-related methods (TASK-137)

// GetThemeManager returns the theme manager
func (m *Model) GetThemeManager() *theme.ThemeManager {
	return m.themeManager
}

// GetCurrentTheme returns the current theme
func (m *Model) GetCurrentTheme() *theme.Theme {
	if m.themeManager == nil {
		return nil
	}
	return m.themeManager.CurrentTheme()
}

// SwitchTheme switches to a different theme by name
func (m *Model) SwitchTheme(name string) error {
	if m.themeManager == nil {
		return nil
	}
	err := m.themeManager.SetTheme(name)
	if err == nil {
		// Save theme preference
		_ = m.themeManager.SaveConfig()
	}
	return err
}

// CycleTheme cycles to the next theme
func (m *Model) CycleTheme() error {
	if m.themeManager == nil {
		return nil
	}
	err := m.themeManager.CycleTheme()
	if err == nil {
		// Save theme preference
		_ = m.themeManager.SaveConfig()
	}
	return err
}

// Toast notification methods (TASK-138)

// GetToastManager returns the toast manager
func (m *Model) GetToastManager() *toast.ToastManager {
	return m.toastManager
}

// ShowToast displays a custom toast notification
func (m *Model) ShowToast(config toast.ToastConfig) tea.Cmd {
	if m.toastManager == nil {
		return nil
	}
	return m.toastManager.ShowToast(config)
}

// ShowInfoToast displays an info toast
func (m *Model) ShowInfoToast(message string) tea.Cmd {
	if m.toastManager == nil {
		return nil
	}
	return m.toastManager.ShowInfo(message)
}

// ShowSuccessToast displays a success toast
func (m *Model) ShowSuccessToast(message string) tea.Cmd {
	if m.toastManager == nil {
		return nil
	}
	return m.toastManager.ShowSuccess(message)
}

// ShowWarningToast displays a warning toast
func (m *Model) ShowWarningToast(message string) tea.Cmd {
	if m.toastManager == nil {
		return nil
	}
	return m.toastManager.ShowWarning(message)
}

// ShowErrorToast displays an error toast
func (m *Model) ShowErrorToast(message string) tea.Cmd {
	if m.toastManager == nil {
		return nil
	}
	return m.toastManager.ShowError(message)
}

// ShowLoadingToast displays a loading toast
func (m *Model) ShowLoadingToast(message string) tea.Cmd {
	if m.toastManager == nil {
		return nil
	}
	return m.toastManager.ShowLoading(message)
}

// DismissToast dismisses a specific toast by ID
func (m *Model) DismissToast(id string) tea.Cmd {
	if m.toastManager == nil {
		return nil
	}
	return m.toastManager.DismissToast(id)
}

// DismissAllToasts dismisses all visible toasts
func (m *Model) DismissAllToasts() tea.Cmd {
	if m.toastManager == nil {
		return nil
	}
	return m.toastManager.DismissAll()
}

