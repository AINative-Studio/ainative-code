# Component Quick Reference Guide

## At-a-Glance Component Status

### WELL-IMPLEMENTED (Ready to Reuse)
- **Model** - Central state container
- **Message System** - Type-safe event messages  
- **Thinking System** - Collapsible content blocks with depth
- **Styling System** - Color palette and style helpers
- **Status Bar** - Responsive multi-section status display
- **Help System** - Categorized keyboard shortcuts display
- **Event Streams** - Thread-safe event buffering
- **Syntax Highlighting** - Code block highlighting with Chroma
- **Animation System** - Spinners, progress, and visual effects

### PARTIALLY-IMPLEMENTED (Needs Abstraction)
- **Completion Popup** - LSP completion display (render OK, logic could be abstracted)
- **Hover System** - LSP type information display
- **Navigation System** - Definition/reference navigation
- **RLHF Collector** - Background interaction capture

### BASIC/MANUAL (Needs Infrastructure)
- **Layout System** - Hard-coded offsets, no abstraction layer
- **Modal System** - Popup overlay mechanism (works but could be abstracted)
- **Responsive Design** - Pattern-based, no unified framework

### COMPLETELY MISSING
- **Dialog System** - Generic confirmation, input, selection dialogs
- **Golden Test Setup** - Snapshot testing for UI regression
- **Multi-step Wizards** - Step-by-step UI flows
- **Toast/Notification System** - Temporary message display
- **Draggable/Resizable Components** - Advanced component features

---

## Key Files to Know

### Core TUI
- `/internal/tui/model.go` - Main Model struct (422 lines)
- `/internal/tui/update.go` - Message handling (250+ lines)
- `/internal/tui/view.go` - Rendering logic
- `/internal/tui/messages.go` - Message type definitions

### Components
- `/internal/tui/animations.go` - Animation state and effects
- `/internal/tui/thinking.go` - Thinking block logic
- `/internal/tui/thinking_view.go` - Thinking block rendering
- `/internal/tui/statusbar.go` - Status bar state and rendering (376 lines)
- `/internal/tui/help.go` - Help overlay system (451 lines)
- `/internal/tui/completion.go` - Completion popup
- `/internal/tui/hover.go` - Hover information display
- `/internal/tui/navigation.go` - Navigation results display

### Styling
- `/internal/tui/styles.go` - Color palette and style definitions
- `/internal/tui/syntax/highlighter.go` - Code syntax highlighting

### Systems
- `/internal/events/types.go` - Event definitions
- `/internal/events/manager.go` - Stream management (258 lines)
- `/internal/rlhf/collector.go` - RLHF auto-collection

### Tests
- `/internal/tui/*_test.go` - Component unit tests
- `/tests/integration/` - Integration tests
- `/tests/benchmark/` - Performance benchmarks

---

## Constructor Functions

```go
// Model constructors
NewModel() Model
NewModelWithLSP(workspace string) Model

// State constructors  
NewAnimationState() *AnimationState
NewThinkingState() *ThinkingState
NewStatusBarState() *StatusBarState
NewHelpState() *HelpState

// External constructors
NewHighlighter(config HighlighterConfig) *Highlighter
NewStreamManager(bufferSize int) *StreamManager
NewCollector(cfg *config.RLHFConfig, client *rlhf.Client) *Collector
```

---

## Message Types Available

```go
// Core messages
errMsg              // Error events
readyMsg            // TUI ready signal
streamChunkMsg      // Streamed content chunks
streamDoneMsg       // Stream completion
userInputMsg        // User text input
readyMsg            // Ready to display

// Thinking messages
thinkingChunkMsg    // Thinking content chunks
thinkingDoneMsg     // Thinking completion
toggleThinkingMsg   // Toggle display
collapseAllThinkingMsg
expandAllThinkingMsg

// RLHF messages
feedbackPromptMsg         // Show feedback dialog
feedbackSubmittedMsg      // Feedback submission
implicitFeedbackMsg       // Implicit feedback actions
```

---

## State Management Patterns

### Pattern 1: Bubble Tea Model/Update/View
```go
// In update handler
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        // Handle input
    case streamChunkMsg:
        // Handle streaming data
    }
    return m, cmd
}
```

### Pattern 2: Event Streaming
```go
// Create and manage event streams
sm := events.NewStreamManager(100)
stream, _ := sm.CreateStream("stream-id")
stream.Send(event)
defer sm.CloseStream("stream-id")
```

### Pattern 3: State Containers
```go
// Isolated state management
type MyComponent struct {
    State SomeState
    // ... other fields
}

func (m *MyComponent) Update(event string) {
    // Modify state
}
```

---

## Layout Quick Tips

### Responsive Breakpoints
- `< 40 chars` - Minimal/compact mode
- `40-80 chars` - Compact mode with hints
- `80-100 chars` - Standard mode without middle section
- `100+ chars` - Full mode with all sections

### Useful lipgloss Functions
```go
// Sizing
lipgloss.Width(str)  // Calculate rendered width
lipgloss.Height(str) // Calculate rendered height

// Positioning
lipgloss.Place(width, height, placeH, placeV, str)

// Styling
lipgloss.NewStyle().
    Foreground(Color("12")).
    Background(Color("240")).
    Bold(true).
    Padding(1, 2).
    Border(lipgloss.RoundedBorder()).
    Render(content)
```

### Common Patterns
```go
// Centering content
centered := lipgloss.Place(width, height, 
    lipgloss.Center, lipgloss.Center, content)

// Side-by-side layout
left + strings.Repeat(" ", spacingWidth) + right

// Padding
padded := lipgloss.NewStyle().Padding(1, 2).Render(content)
```

---

## Color Palette

### Thinking Block Colors (Depth-Based)
- `Color("141")` - Light purple (depth 0)
- `Color("105")` - Medium purple (depth 1)
- `Color("99")` - Deep purple (depth 2)
- `Color("63")` - Dark purple (depth 3)

### UI Colors
- `Color("240")` - Gray (borders, muted)
- `Color("13")` - Magenta (headers)
- `Color("252")` - Light gray (text)
- `Color("12")` - Blue (accents, input)
- `Color("10")` - Green (ready, success)
- `Color("9")` - Red (error)
- `Color("14")` - Cyan (info)

---

## Testing Patterns

### Unit Test Template
```go
func TestComponent(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {name: "case 1", input: "x", expected: "y"},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := Component(tt.input)
            if result != tt.expected {
                t.Errorf("got %v, want %v", result, tt.expected)
            }
        })
    }
}
```

### Assertion Patterns
```go
// Value comparison
if actual != expected {
    t.Errorf("expected %v, got %v", expected, actual)
}

// Nil check
if err != nil {
    t.Fatalf("unexpected error: %v", err)
}

// Length check
if len(items) != expectedCount {
    t.Errorf("expected %d items, got %d", expectedCount, len(items))
}
```

---

## Next Steps for Refactor

### High Priority
1. Extract component interfaces for popups
2. Create layout abstraction layer
3. Add dialog system (confirmation, input, selection)

### Medium Priority  
4. Add golden/snapshot testing
5. Wrap animations in component lifecycle
6. Create modal manager for stacking

### Low Priority
7. Centralize theme system
8. Add toast/notification system
9. Implement draggable/resizable components

