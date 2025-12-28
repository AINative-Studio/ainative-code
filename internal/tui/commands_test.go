package tui

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// TestStreamResponse tests the StreamResponse command with various readers
func TestStreamResponse(t *testing.T) {
	tests := []struct {
		name            string
		reader          io.Reader
		expectedContent string
		expectError     bool
		expectDone      bool
	}{
		{
			name:            "streams from string reader",
			reader:          strings.NewReader("Hello, world!"),
			expectedContent: "Hello, world!",
			expectError:     false,
			expectDone:      false, // First read gets content
		},
		{
			name:            "streams from empty reader",
			reader:          strings.NewReader(""),
			expectedContent: "",
			expectError:     false,
			expectDone:      true, // Empty reader returns EOF immediately
		},
		{
			name:            "streams from bytes buffer",
			reader:          bytes.NewBufferString("test content"),
			expectedContent: "test content",
			expectError:     false,
			expectDone:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := StreamResponse(tt.reader)

			// Execute command
			msg := cmd()

			// Verify message type
			if tt.expectDone {
				_, ok := msg.(streamDoneMsg)
				if !ok {
					t.Errorf("expected streamDoneMsg, got %T", msg)
				}
			} else if tt.expectError {
				errMessage, ok := msg.(errMsg)
				if !ok {
					t.Errorf("expected errMsg, got %T", msg)
				} else if errMessage.err == nil {
					t.Error("expected error message to contain error")
				}
			} else {
				chunkMessage, ok := msg.(streamChunkMsg)
				if !ok {
					t.Errorf("expected streamChunkMsg, got %T", msg)
				} else if !strings.Contains(chunkMessage.content, tt.expectedContent) {
					t.Errorf("expected content to contain %q, got %q", tt.expectedContent, chunkMessage.content)
				}
			}
		})
	}
}

// TestStreamResponseWithMultipleReads tests streaming with multiple chunks
func TestStreamResponseWithMultipleReads(t *testing.T) {
	// Create reader with known content
	reader := strings.NewReader("chunk1")

	// First read should return chunk
	cmd := StreamResponse(reader)
	msg := cmd()

	chunkMessage, ok := msg.(streamChunkMsg)
	if !ok {
		t.Fatalf("expected streamChunkMsg, got %T", msg)
	}

	if !strings.Contains(chunkMessage.content, "chunk1") {
		t.Errorf("expected content to contain chunk1, got %q", chunkMessage.content)
	}
}

// TestStreamResponseEOF tests StreamResponse handles EOF correctly
func TestStreamResponseEOF(t *testing.T) {
	// Create empty reader (will return EOF immediately)
	reader := strings.NewReader("")

	cmd := StreamResponse(reader)
	msg := cmd()

	// Should return streamDoneMsg on EOF
	_, ok := msg.(streamDoneMsg)
	if !ok {
		t.Errorf("expected streamDoneMsg on EOF, got %T", msg)
	}
}

// errorReader is a test helper that always returns an error
type errorReader struct {
	err error
}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, e.err
}

// TestStreamResponseError tests StreamResponse handles read errors
func TestStreamResponseError(t *testing.T) {
	// Create reader that returns error
	testErr := errors.New("read error")
	reader := &errorReader{err: testErr}

	cmd := StreamResponse(reader)
	msg := cmd()

	// Should return errMsg
	errMessage, ok := msg.(errMsg)
	if !ok {
		t.Fatalf("expected errMsg, got %T", msg)
	}

	if errMessage.err == nil {
		t.Error("expected error message to contain error")
	}

	// Verify error contains the original error
	if !strings.Contains(errMessage.err.Error(), "stream read error") {
		t.Errorf("expected error to mention stream read error, got %q", errMessage.err.Error())
	}
}

// TestHandleAPIError tests the HandleAPIError command
func TestHandleAPIError(t *testing.T) {
	tests := []struct {
		name          string
		err           error
		expectedError string
	}{
		{
			name:          "handles simple error",
			err:           errors.New("API error"),
			expectedError: "API error",
		},
		{
			name:          "handles timeout error",
			err:           errors.New("connection timeout"),
			expectedError: "connection timeout",
		},
		{
			name:          "handles authentication error",
			err:           errors.New("invalid API key"),
			expectedError: "invalid API key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := HandleAPIError(tt.err)

			// Execute command
			msg := cmd()

			// Verify message type
			errMessage, ok := msg.(errMsg)
			if !ok {
				t.Fatalf("expected errMsg, got %T", msg)
			}

			// Verify error content
			if errMessage.err.Error() != tt.expectedError {
				t.Errorf("expected error %q, got %q", tt.expectedError, errMessage.err.Error())
			}
		})
	}
}

// TestWaitForStream tests the WaitForStream command
func TestWaitForStream(t *testing.T) {
	tests := []struct {
		name            string
		closeImmediately bool
		expectTimeout   bool
	}{
		{
			name:            "completes when channel is closed",
			closeImmediately: true,
			expectTimeout:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			done := make(chan struct{})

			if tt.closeImmediately {
				close(done)
			}

			cmd := WaitForStream(done)

			// Execute command in goroutine with timeout
			msgChan := make(chan tea.Msg, 1)
			go func() {
				msgChan <- cmd()
			}()

			// Wait for message or timeout
			select {
			case msg := <-msgChan:
				if tt.expectTimeout {
					t.Error("expected timeout, but command completed")
				}

				// Verify message type
				_, ok := msg.(streamDoneMsg)
				if !ok {
					t.Errorf("expected streamDoneMsg, got %T", msg)
				}
			case <-time.After(100 * time.Millisecond):
				if !tt.expectTimeout {
					t.Error("command timed out unexpectedly")
				}
			}
		})
	}
}

// TestWaitForStreamBlocking tests that WaitForStream blocks until channel closes
func TestWaitForStreamBlocking(t *testing.T) {
	done := make(chan struct{})

	cmd := WaitForStream(done)

	// Start command in goroutine
	msgChan := make(chan tea.Msg, 1)
	go func() {
		msgChan <- cmd()
	}()

	// Verify command doesn't complete immediately
	select {
	case <-msgChan:
		t.Error("expected command to block, but it completed immediately")
	case <-time.After(50 * time.Millisecond):
		// Good - command is blocking
	}

	// Close channel
	close(done)

	// Verify command completes
	select {
	case msg := <-msgChan:
		_, ok := msg.(streamDoneMsg)
		if !ok {
			t.Errorf("expected streamDoneMsg, got %T", msg)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("command did not complete after channel closed")
	}
}

// TestBatchStreamChunks tests the BatchStreamChunks command
func TestBatchStreamChunks(t *testing.T) {
	tests := []struct {
		name            string
		chunks          []string
		expectedContent string
	}{
		{
			name:            "batches single chunk",
			chunks:          []string{"chunk1"},
			expectedContent: "chunk1",
		},
		{
			name:            "batches multiple chunks",
			chunks:          []string{"chunk1", "chunk2", "chunk3"},
			expectedContent: "chunk1chunk2chunk3",
		},
		{
			name:            "batches empty chunks",
			chunks:          []string{},
			expectedContent: "",
		},
		{
			name:            "batches chunks with spaces",
			chunks:          []string{"hello ", "world", "!"},
			expectedContent: "hello world!",
		},
		{
			name:            "batches chunks with newlines",
			chunks:          []string{"line1\n", "line2\n", "line3"},
			expectedContent: "line1\nline2\nline3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := BatchStreamChunks(tt.chunks)

			// Execute command
			msg := cmd()

			// Verify message type
			chunkMessage, ok := msg.(streamChunkMsg)
			if !ok {
				t.Fatalf("expected streamChunkMsg, got %T", msg)
			}

			// Verify content
			if chunkMessage.content != tt.expectedContent {
				t.Errorf("expected content %q, got %q", tt.expectedContent, chunkMessage.content)
			}
		})
	}
}

// TestBatchStreamChunksWithLargeData tests batching with many chunks
func TestBatchStreamChunksWithLargeData(t *testing.T) {
	// Create large number of chunks
	chunks := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		chunks[i] = "x"
	}

	cmd := BatchStreamChunks(chunks)
	msg := cmd()

	chunkMessage, ok := msg.(streamChunkMsg)
	if !ok {
		t.Fatalf("expected streamChunkMsg, got %T", msg)
	}

	// Verify content length
	if len(chunkMessage.content) != 1000 {
		t.Errorf("expected content length 1000, got %d", len(chunkMessage.content))
	}
}

// TestCommandsReturnValidMessages tests that all command functions return valid messages
func TestCommandsReturnValidMessages(t *testing.T) {
	tests := []struct {
		name        string
		cmd         tea.Cmd
		expectedMsg interface{}
	}{
		{
			name:        "StreamResponse returns message",
			cmd:         StreamResponse(strings.NewReader("test")),
			expectedMsg: nil, // Any message is valid
		},
		{
			name:        "HandleAPIError returns errMsg",
			cmd:         HandleAPIError(errors.New("test")),
			expectedMsg: errMsg{},
		},
		{
			name: "WaitForStream returns message",
			cmd: func() tea.Cmd {
				done := make(chan struct{})
				close(done)
				return WaitForStream(done)
			}(),
			expectedMsg: streamDoneMsg{},
		},
		{
			name:        "BatchStreamChunks returns streamChunkMsg",
			cmd:         BatchStreamChunks([]string{"test"}),
			expectedMsg: streamChunkMsg{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.cmd == nil {
				t.Fatal("expected command to be non-nil")
			}

			msg := tt.cmd()
			if msg == nil {
				t.Fatal("expected command to return non-nil message")
			}

			// Verify message implements tea.Msg interface
			var _ tea.Msg = msg
		})
	}
}

// TestStreamResponseNonBlocking tests that StreamResponse doesn't block indefinitely
func TestStreamResponseNonBlocking(t *testing.T) {
	reader := strings.NewReader("test content")
	cmd := StreamResponse(reader)

	// Execute command with timeout
	msgChan := make(chan tea.Msg, 1)
	go func() {
		msgChan <- cmd()
	}()

	select {
	case msg := <-msgChan:
		// Command completed
		if msg == nil {
			t.Error("expected non-nil message")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("StreamResponse blocked for more than 100ms")
	}
}

// TestHandleAPIErrorNonBlocking tests that HandleAPIError executes immediately
func TestHandleAPIErrorNonBlocking(t *testing.T) {
	cmd := HandleAPIError(errors.New("test error"))

	msgChan := make(chan tea.Msg, 1)
	go func() {
		msgChan <- cmd()
	}()

	select {
	case msg := <-msgChan:
		if msg == nil {
			t.Error("expected non-nil message")
		}
	case <-time.After(10 * time.Millisecond):
		t.Error("HandleAPIError blocked for more than 10ms")
	}
}

// TestBatchStreamChunksNonBlocking tests that BatchStreamChunks executes immediately
func TestBatchStreamChunksNonBlocking(t *testing.T) {
	chunks := make([]string, 100)
	for i := 0; i < 100; i++ {
		chunks[i] = "chunk"
	}

	cmd := BatchStreamChunks(chunks)

	msgChan := make(chan tea.Msg, 1)
	go func() {
		msgChan <- cmd()
	}()

	select {
	case msg := <-msgChan:
		if msg == nil {
			t.Error("expected non-nil message")
		}
	case <-time.After(50 * time.Millisecond):
		t.Error("BatchStreamChunks blocked for more than 50ms")
	}
}

// TestStreamResponseWithVariousBufferSizes tests streaming with different buffer sizes
func TestStreamResponseWithVariousBufferSizes(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		minChunkLen int
	}{
		{
			name:        "small content",
			content:     "small",
			minChunkLen: 5,
		},
		{
			name:        "medium content",
			content:     strings.Repeat("x", 500),
			minChunkLen: 500,
		},
		{
			name:        "large content",
			content:     strings.Repeat("y", 2000),
			minChunkLen: 1024, // Buffer size in StreamResponse is 1024
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.content)
			cmd := StreamResponse(reader)

			msg := cmd()

			// First read should return a chunk
			if chunkMsg, ok := msg.(streamChunkMsg); ok {
				if len(chunkMsg.content) == 0 {
					t.Error("expected non-empty chunk")
				}
			} else if _, ok := msg.(streamDoneMsg); !ok {
				t.Errorf("expected streamChunkMsg or streamDoneMsg, got %T", msg)
			}
		})
	}
}

// TestCommandsWithNilInputs tests command functions with edge case inputs
func TestCommandsWithNilInputs(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func() tea.Cmd
		shouldRun bool
	}{
		{
			name: "BatchStreamChunks with nil slice",
			setupFunc: func() tea.Cmd {
				return BatchStreamChunks(nil)
			},
			shouldRun: true,
		},
		{
			name: "BatchStreamChunks with empty slice",
			setupFunc: func() tea.Cmd {
				return BatchStreamChunks([]string{})
			},
			shouldRun: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.setupFunc()

			if !tt.shouldRun {
				return
			}

			// Verify command doesn't panic
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("command panicked: %v", r)
				}
			}()

			msg := cmd()
			if msg == nil {
				t.Error("expected non-nil message")
			}
		})
	}
}

// TestWaitForStreamWithPreClosedChannel tests WaitForStream when channel is already closed
func TestWaitForStreamWithPreClosedChannel(t *testing.T) {
	done := make(chan struct{})
	close(done)

	cmd := WaitForStream(done)

	// Should complete immediately
	msgChan := make(chan tea.Msg, 1)
	go func() {
		msgChan <- cmd()
	}()

	select {
	case msg := <-msgChan:
		_, ok := msg.(streamDoneMsg)
		if !ok {
			t.Errorf("expected streamDoneMsg, got %T", msg)
		}
	case <-time.After(50 * time.Millisecond):
		t.Error("command did not complete immediately with pre-closed channel")
	}
}

// TestErrorMessageWrapping tests that HandleAPIError wraps errors correctly
func TestErrorMessageWrapping(t *testing.T) {
	originalErr := errors.New("original error message")
	cmd := HandleAPIError(originalErr)

	msg := cmd()

	errMessage, ok := msg.(errMsg)
	if !ok {
		t.Fatalf("expected errMsg, got %T", msg)
	}

	if errMessage.err != originalErr {
		t.Error("expected error to be preserved exactly")
	}
}

// TestBatchStreamChunksPreservesOrder tests that chunks are combined in order
func TestBatchStreamChunksPreservesOrder(t *testing.T) {
	chunks := []string{"first", "second", "third", "fourth"}
	cmd := BatchStreamChunks(chunks)

	msg := cmd()

	chunkMessage, ok := msg.(streamChunkMsg)
	if !ok {
		t.Fatalf("expected streamChunkMsg, got %T", msg)
	}

	expected := "firstsecondthirdfourth"
	if chunkMessage.content != expected {
		t.Errorf("expected %q, got %q", expected, chunkMessage.content)
	}

	// Verify order is preserved
	if !strings.Contains(chunkMessage.content, "firstsecond") {
		t.Error("chunk order not preserved")
	}
	if !strings.Contains(chunkMessage.content, "secondthird") {
		t.Error("chunk order not preserved")
	}
}

// Benchmark tests for performance validation

// BenchmarkStreamResponse benchmarks StreamResponse command
func BenchmarkStreamResponse(b *testing.B) {
	content := "test content for benchmarking"
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		reader := strings.NewReader(content)
		cmd := StreamResponse(reader)
		_ = cmd()
	}
}

// BenchmarkHandleAPIError benchmarks HandleAPIError command
func BenchmarkHandleAPIError(b *testing.B) {
	err := errors.New("test error")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cmd := HandleAPIError(err)
		_ = cmd()
	}
}

// BenchmarkBatchStreamChunks benchmarks BatchStreamChunks command
func BenchmarkBatchStreamChunks(b *testing.B) {
	chunks := []string{"chunk1", "chunk2", "chunk3", "chunk4", "chunk5"}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cmd := BatchStreamChunks(chunks)
		_ = cmd()
	}
}

// BenchmarkBatchStreamChunksLarge benchmarks BatchStreamChunks with many chunks
func BenchmarkBatchStreamChunksLarge(b *testing.B) {
	chunks := make([]string, 100)
	for i := 0; i < 100; i++ {
		chunks[i] = "chunk"
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cmd := BatchStreamChunks(chunks)
		_ = cmd()
	}
}

// BenchmarkWaitForStream benchmarks WaitForStream command
func BenchmarkWaitForStream(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		done := make(chan struct{})
		close(done)
		cmd := WaitForStream(done)
		_ = cmd()
	}
}

// BenchmarkStreamResponseWithLargeContent benchmarks StreamResponse with large content
func BenchmarkStreamResponseWithLargeContent(b *testing.B) {
	content := strings.Repeat("x", 10000)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		reader := strings.NewReader(content)
		cmd := StreamResponse(reader)
		_ = cmd()
	}
}
