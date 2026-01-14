package tui

import (
	"errors"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// TestInit tests the Init function
func TestInit(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "returns non-nil command",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel()
			cmd := m.Init()

			if cmd == nil {
				t.Error("expected Init to return non-nil command")
			}
		})
	}
}

// TestInitReturnsValidCmd tests that Init returns a valid tea.Cmd
func TestInitReturnsValidCmd(t *testing.T) {
	m := NewModel()
	cmd := m.Init()

	// Verify cmd is not nil
	if cmd == nil {
		t.Fatal("expected Init to return non-nil command")
	}

	// Verify cmd can be called without panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Init command panicked: %v", r)
		}
	}()

	// Execute the command (it's a batch, so it may return nil or a message)
	msg := cmd()
	_ = msg // Batch commands may return nil
}

// TestInitWithDifferentModelStates tests Init with various model states
func TestInitWithDifferentModelStates(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(*Model)
	}{
		{
			name: "init with default model state",
			setupFunc: func(m *Model) {
				// Use default state
			},
		},
		{
			name: "init with ready state set",
			setupFunc: func(m *Model) {
				m.ready = true
			},
		},
		{
			name: "init with messages already present",
			setupFunc: func(m *Model) {
				m.AddMessage("user", "existing message")
			},
		},
		{
			name: "init with size already set",
			setupFunc: func(m *Model) {
				m.SetSize(80, 24)
			},
		},
		{
			name: "init with streaming state",
			setupFunc: func(m *Model) {
				m.streaming = true
			},
		},
		{
			name: "init with error state",
			setupFunc: func(m *Model) {
				m.SetError(errors.New("test error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel()
			tt.setupFunc(&m)

			cmd := m.Init()

			if cmd == nil {
				t.Error("expected Init to return non-nil command regardless of model state")
			}

			// Verify command can be executed
			_ = cmd()
		})
	}
}

// TestInitCommandExecution tests execution of the Init command
func TestInitCommandExecution(t *testing.T) {
	m := NewModel()
	cmd := m.Init()

	if cmd == nil {
		t.Fatal("expected Init to return non-nil command")
	}

	// Execute the command
	// Note: tea.Batch returns a function that executes all batched commands
	// and returns the last non-nil message
	msg := cmd()

	// tea.Batch may return nil if all commands complete synchronously
	// or may return a message from one of the batched commands
	// We don't make assumptions about the return value, just verify no panic
	_ = msg
}

// TestInitMultipleCalls tests that Init can be called multiple times
func TestInitMultipleCalls(t *testing.T) {
	m := NewModel()

	// Call Init multiple times
	cmd1 := m.Init()
	cmd2 := m.Init()
	cmd3 := m.Init()

	// Verify all return non-nil commands
	if cmd1 == nil {
		t.Error("expected first Init call to return non-nil command")
	}
	if cmd2 == nil {
		t.Error("expected second Init call to return non-nil command")
	}
	if cmd3 == nil {
		t.Error("expected third Init call to return non-nil command")
	}

	// Execute all commands to verify they don't panic
	_ = cmd1()
	_ = cmd2()
	_ = cmd3()
}

// TestInitDoesNotModifyModel tests that Init doesn't modify the model state
func TestInitDoesNotModifyModel(t *testing.T) {
	m := NewModel()

	// Capture initial state
	initialReady := m.ready
	initialQuitting := m.quitting
	initialStreaming := m.streaming
	initialMessagesLen := len(m.messages)
	initialWidth := m.width
	initialHeight := m.height

	// Call Init
	cmd := m.Init()
	_ = cmd

	// Verify state hasn't changed
	if m.ready != initialReady {
		t.Errorf("expected ready to remain %v, got %v", initialReady, m.ready)
	}
	if m.quitting != initialQuitting {
		t.Errorf("expected quitting to remain %v, got %v", initialQuitting, m.quitting)
	}
	if m.streaming != initialStreaming {
		t.Errorf("expected streaming to remain %v, got %v", initialStreaming, m.streaming)
	}
	if len(m.messages) != initialMessagesLen {
		t.Errorf("expected messages length to remain %d, got %d", initialMessagesLen, len(m.messages))
	}
	if m.width != initialWidth {
		t.Errorf("expected width to remain %d, got %d", initialWidth, m.width)
	}
	if m.height != initialHeight {
		t.Errorf("expected height to remain %d, got %d", initialWidth, m.height)
	}
}

// TestInitIntegrationWithUpdate tests Init followed by Update
func TestInitIntegrationWithUpdate(t *testing.T) {
	m := NewModel()

	// Call Init
	initCmd := m.Init()
	if initCmd == nil {
		t.Fatal("expected Init to return non-nil command")
	}

	// Execute init command
	initMsg := initCmd()

	// If init returns a message, process it through Update
	if initMsg != nil {
		updatedModel, updateCmd := m.Update(initMsg)
		_ = updatedModel
		_ = updateCmd
	}

	// Verify model is still in a valid state
	if m.quitting {
		t.Error("expected model not to be quitting after Init")
	}
}

// TestInitWithReadyMsg tests that SendReady is part of Init batch
func TestInitWithReadyMsg(t *testing.T) {
	// Create a model
	m := NewModel()

	// Verify initial state
	if m.ready {
		t.Error("expected model to not be ready initially")
	}

	// Call Init
	cmd := m.Init()
	if cmd == nil {
		t.Fatal("expected Init to return non-nil command")
	}

	// Execute the batched command
	// tea.Batch executes all commands and returns the last non-nil message
	msg := cmd()

	// Process the message if one is returned
	if msg != nil {
		// Check if it's a readyMsg
		if _, ok := msg.(readyMsg); ok {
			// readyMsg should update the model to ready state
			updatedModel, _ := m.Update(msg)
			m = updatedModel.(Model)

			if !m.ready {
				t.Error("expected model to be ready after processing readyMsg")
			}
		}
	}
}

// TestInitCommandType tests that Init returns a tea.Cmd function
func TestInitCommandType(t *testing.T) {
	m := NewModel()
	cmd := m.Init()

	// Verify cmd is a function
	if cmd == nil {
		t.Fatal("expected Init to return non-nil command")
	}

	// tea.Cmd is a function type: func() tea.Msg
	// We verify it can be called
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("calling Init command caused panic: %v", r)
		}
	}()

	result := cmd()
	_ = result // May be nil or a message
}

// TestInitConsistency tests that Init produces consistent results
func TestInitConsistency(t *testing.T) {
	// Create multiple models and call Init on each
	models := make([]Model, 5)
	cmds := make([]tea.Cmd, 5)

	for i := 0; i < 5; i++ {
		models[i] = NewModel()
		cmds[i] = models[i].Init()
	}

	// Verify all commands are non-nil
	for i, cmd := range cmds {
		if cmd == nil {
			t.Errorf("expected command %d to be non-nil", i)
		}
	}

	// Execute all commands and verify they don't panic
	for i, cmd := range cmds {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("command %d panicked: %v", i, r)
				}
			}()
			_ = cmd()
		}()
	}
}

// TestInitWithCustomModel tests Init with a manually constructed model
func TestInitWithCustomModel(t *testing.T) {
	// Create a model with custom initial state
	m := Model{
		ready:     false,
		quitting:  false,
		streaming: false,
		width:     100,
		height:    50,
		messages:  []Message{},
		err:       nil,
	}

	// Initialize components that NewModel would initialize
	m.viewport = NewModel().viewport
	m.textInput = NewModel().textInput

	// Call Init
	cmd := m.Init()

	if cmd == nil {
		t.Error("expected Init to return non-nil command even with custom model")
	}

	// Execute command
	_ = cmd()
}

// TestInitReturnsImmediately tests that Init doesn't block
func TestInitReturnsImmediately(t *testing.T) {
	m := NewModel()

	// Init should return immediately without blocking
	done := make(chan bool, 1)

	go func() {
		cmd := m.Init()
		_ = cmd
		done <- true
	}()

	// Wait for completion with timeout
	select {
	case <-done:
		// Success - Init returned immediately
	case <-time.After(100 * time.Millisecond):
		t.Error("Init did not return within 100ms - it may be blocking")
	}
}

// TestInitBatchedCommands tests that Init returns batched commands
func TestInitBatchedCommands(t *testing.T) {
	m := NewModel()
	cmd := m.Init()

	if cmd == nil {
		t.Fatal("expected Init to return non-nil command")
	}

	// Execute the batched command
	// tea.Batch creates a function that executes multiple commands
	// We can't directly inspect the batch, but we can verify it executes
	msg := cmd()

	// The batched command should execute without panic
	// It may return nil or a message
	_ = msg
}

// TestInitWithZeroSizeModel tests Init with a model that has zero dimensions
func TestInitWithZeroSizeModel(t *testing.T) {
	m := NewModel()
	// Default model has zero width and height

	if m.width != 0 || m.height != 0 {
		t.Skip("test assumes NewModel creates zero-size model")
	}

	cmd := m.Init()

	if cmd == nil {
		t.Error("expected Init to handle zero-size model")
	}

	// Execute command
	_ = cmd()
}

// TestInitAfterSetSize tests Init after setting model size
func TestInitAfterSetSize(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
	}{
		{
			name:   "standard terminal size",
			width:  80,
			height: 24,
		},
		{
			name:   "large terminal size",
			width:  200,
			height: 60,
		},
		{
			name:   "small terminal size",
			width:  40,
			height: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel()
			m.SetSize(tt.width, tt.height)

			cmd := m.Init()

			if cmd == nil {
				t.Error("expected Init to return non-nil command after SetSize")
			}

			// Execute command
			_ = cmd()
		})
	}
}

// TestInitEdgeCases tests Init with edge case scenarios
func TestInitEdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func() Model
	}{
		{
			name: "init with all boolean flags true",
			setupFunc: func() Model {
				m := NewModel()
				m.ready = true
				m.quitting = true
				m.streaming = true
				return m
			},
		},
		{
			name: "init with negative dimensions",
			setupFunc: func() Model {
				m := NewModel()
				m.width = -10
				m.height = -20
				return m
			},
		},
		{
			name: "init with very large dimensions",
			setupFunc: func() Model {
				m := NewModel()
				m.width = 10000
				m.height = 10000
				return m
			},
		},
		{
			name: "init with many messages",
			setupFunc: func() Model {
				m := NewModel()
				for i := 0; i < 100; i++ {
					m.AddMessage("user", "test message")
				}
				return m
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.setupFunc()
			cmd := m.Init()

			if cmd == nil {
				t.Error("expected Init to handle edge case and return non-nil command")
			}

			// Verify command can be executed without panic
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Init command panicked on edge case: %v", r)
				}
			}()

			_ = cmd()
		})
	}
}

// Benchmark tests for performance validation

// BenchmarkInit benchmarks the Init function
func BenchmarkInit(b *testing.B) {
	m := NewModel()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = m.Init()
	}
}

// BenchmarkInitAndExecute benchmarks Init and command execution
func BenchmarkInitAndExecute(b *testing.B) {
	m := NewModel()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cmd := m.Init()
		_ = cmd()
	}
}

// BenchmarkInitWithSize benchmarks Init with a sized model
func BenchmarkInitWithSize(b *testing.B) {
	m := NewModel()
	m.SetSize(80, 24)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = m.Init()
	}
}

// BenchmarkInitWithMessages benchmarks Init with messages present
func BenchmarkInitWithMessages(b *testing.B) {
	m := NewModel()
	m.AddMessage("user", "Hello")
	m.AddMessage("assistant", "Hi there!")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = m.Init()
	}
}

// BenchmarkMultipleInitCalls benchmarks multiple Init calls
func BenchmarkMultipleInitCalls(b *testing.B) {
	m := NewModel()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = m.Init()
		_ = m.Init()
		_ = m.Init()
	}
}
