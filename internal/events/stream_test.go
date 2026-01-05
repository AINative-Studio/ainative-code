package events

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEventStream(t *testing.T) {
	tests := []struct {
		name       string
		bufferSize int
		wantSize   int
	}{
		{
			name:       "default buffer size",
			bufferSize: 0,
			wantSize:   defaultBufferSize,
		},
		{
			name:       "custom buffer size",
			bufferSize: 50,
			wantSize:   50,
		},
		{
			name:       "large buffer size",
			bufferSize: 1000,
			wantSize:   1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stream := NewEventStream(tt.bufferSize)
			require.NotNil(t, stream)
			assert.Equal(t, tt.wantSize, stream.bufferSize)
			assert.Equal(t, tt.wantSize, cap(stream.events))
			assert.False(t, stream.closed)
		})
	}
}

func TestEventStream_Send(t *testing.T) {
	t.Run("send event successfully", func(t *testing.T) {
		stream := NewEventStream(10)
		event := TextDeltaEvent("hello")

		err := stream.Send(event)
		require.NoError(t, err)

		// Receive the event
		select {
		case received := <-stream.Receive():
			assert.Equal(t, event.Type, received.Type)
			assert.Equal(t, event.Data["text"], received.Data["text"])
		case <-time.After(1 * time.Second):
			t.Fatal("timeout waiting for event")
		}
	})

	t.Run("send to closed stream", func(t *testing.T) {
		stream := NewEventStream(10)
		stream.Close()

		event := TextDeltaEvent("hello")
		err := stream.Send(event)
		assert.Error(t, err)
		assert.True(t, IsStreamClosed(err))
	})

	t.Run("send nil event", func(t *testing.T) {
		stream := NewEventStream(10)
		defer stream.Close()

		err := stream.Send(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "event cannot be nil")
	})

	t.Run("send invalid event", func(t *testing.T) {
		stream := NewEventStream(10)
		defer stream.Close()

		// Create invalid event (zero timestamp)
		event := &Event{
			Type: EventTextDelta,
			Data: map[string]interface{}{"text": "hello"},
		}

		err := stream.Send(event)
		assert.Error(t, err)
		assert.True(t, IsInvalidEvent(err))
	})
}

func TestEventStream_SendWithContext(t *testing.T) {
	t.Run("send with valid context", func(t *testing.T) {
		stream := NewEventStream(10)
		defer stream.Close()

		ctx := context.Background()
		event := TextDeltaEvent("hello")

		err := stream.SendWithContext(ctx, event)
		require.NoError(t, err)

		// Receive the event
		select {
		case received := <-stream.Receive():
			assert.Equal(t, event.Type, received.Type)
		case <-time.After(1 * time.Second):
			t.Fatal("timeout waiting for event")
		}
	})

	t.Run("send with cancelled context", func(t *testing.T) {
		stream := NewEventStream(1) // Small buffer to force blocking
		defer stream.Close()

		// Fill the buffer
		stream.Send(TextDeltaEvent("filler"))

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		event := TextDeltaEvent("hello")
		err := stream.SendWithContext(ctx, event)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})

	t.Run("send with timeout context", func(t *testing.T) {
		stream := NewEventStream(1) // Small buffer
		defer stream.Close()

		// Fill the buffer
		stream.Send(TextDeltaEvent("event1"))

		// Try to send with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		event := TextDeltaEvent("event2")
		err := stream.SendWithContext(ctx, event)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "context deadline exceeded")
	})
}

func TestEventStream_Receive(t *testing.T) {
	t.Run("receive events in order", func(t *testing.T) {
		stream := NewEventStream(10)
		defer stream.Close()

		// Send multiple events
		events := []*Event{
			TextDeltaEvent("first"),
			TextDeltaEvent("second"),
			TextDeltaEvent("third"),
		}

		for _, e := range events {
			err := stream.Send(e)
			require.NoError(t, err)
		}

		// Receive events and verify order
		for i, expected := range events {
			select {
			case received := <-stream.Receive():
				assert.Equal(t, expected.Data["text"], received.Data["text"], "event %d mismatch", i)
			case <-time.After(1 * time.Second):
				t.Fatalf("timeout waiting for event %d", i)
			}
		}
	})

	t.Run("receive from closed stream", func(t *testing.T) {
		stream := NewEventStream(10)

		// Send an event
		stream.Send(TextDeltaEvent("hello"))

		// Close the stream
		stream.Close()

		// Should still receive the buffered event
		select {
		case event := <-stream.Receive():
			assert.Equal(t, "hello", event.Data["text"])
		case <-time.After(1 * time.Second):
			t.Fatal("timeout waiting for buffered event")
		}

		// Channel should be closed now
		select {
		case _, ok := <-stream.Receive():
			assert.False(t, ok, "channel should be closed")
		case <-time.After(1 * time.Second):
			t.Fatal("timeout waiting for channel close")
		}
	})
}

func TestEventStream_Close(t *testing.T) {
	t.Run("close stream once", func(t *testing.T) {
		stream := NewEventStream(10)

		err := stream.Close()
		assert.NoError(t, err)
		assert.True(t, stream.IsClosed())

		// Verify channel is closed
		select {
		case _, ok := <-stream.Receive():
			assert.False(t, ok, "channel should be closed")
		case <-time.After(1 * time.Second):
			t.Fatal("timeout waiting for channel close")
		}
	})

	t.Run("close stream multiple times", func(t *testing.T) {
		stream := NewEventStream(10)

		err1 := stream.Close()
		assert.NoError(t, err1)

		err2 := stream.Close()
		assert.Error(t, err2)
		assert.True(t, IsStreamClosed(err2))
	})
}

func TestEventStream_IsClosed(t *testing.T) {
	stream := NewEventStream(10)

	assert.False(t, stream.IsClosed())

	stream.Close()

	assert.True(t, stream.IsClosed())
}

func TestEventStream_Backpressure(t *testing.T) {
	t.Run("block policy - blocks when full", func(t *testing.T) {
		bufferSize := 2
		stream := NewEventStream(bufferSize)
		stream.SetBackpressurePolicy(BackpressureBlock)
		defer stream.Close()

		// Fill the buffer
		for i := 0; i < bufferSize; i++ {
			err := stream.Send(TextDeltaEvent("event"))
			require.NoError(t, err)
		}

		// Next send should block
		done := make(chan bool)
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()
			err := stream.SendWithContext(ctx, TextDeltaEvent("blocked"))
			assert.Error(t, err) // Should timeout
			done <- true
		}()

		select {
		case <-done:
			// Expected behavior - send blocked and timed out
		case <-time.After(500 * time.Millisecond):
			t.Fatal("goroutine did not complete")
		}
	})

	t.Run("drop policy - drops when full", func(t *testing.T) {
		bufferSize := 2
		stream := NewEventStream(bufferSize)
		stream.SetBackpressurePolicy(BackpressureDrop)
		defer stream.Close()

		// Fill the buffer
		for i := 0; i < bufferSize; i++ {
			err := stream.Send(TextDeltaEvent("event"))
			require.NoError(t, err)
		}

		// Next send should return error immediately
		err := stream.Send(TextDeltaEvent("dropped"))
		assert.Error(t, err)
		assert.True(t, IsStreamFull(err))
	})
}

func TestEventStream_Concurrent(t *testing.T) {
	t.Run("concurrent send and receive", func(t *testing.T) {
		stream := NewEventStream(100)
		defer stream.Close()

		numSenders := 5
		numEvents := 20
		totalEvents := numSenders * numEvents

		var wg sync.WaitGroup

		// Start senders
		for i := 0; i < numSenders; i++ {
			wg.Add(1)
			go func(senderID int) {
				defer wg.Done()
				for j := 0; j < numEvents; j++ {
					event := TextDeltaEvent("hello")
					err := stream.Send(event)
					if err != nil && !IsStreamClosed(err) {
						t.Errorf("sender %d: unexpected error: %v", senderID, err)
					}
				}
			}(i)
		}

		// Start receiver
		received := 0
		receiveDone := make(chan bool)
		go func() {
			for range stream.Receive() {
				received++
				if received >= totalEvents {
					break
				}
			}
			receiveDone <- true
		}()

		// Wait for all senders
		wg.Wait()

		// Wait for receiver
		select {
		case <-receiveDone:
			assert.Equal(t, totalEvents, received)
		case <-time.After(5 * time.Second):
			t.Fatalf("timeout: only received %d/%d events", received, totalEvents)
		}
	})
}

func TestEventStream_SetBackpressurePolicy(t *testing.T) {
	stream := NewEventStream(10)
	defer stream.Close()

	// Default should be block
	assert.Equal(t, BackpressureBlock, stream.backpressurePolicy)

	// Change to drop
	stream.SetBackpressurePolicy(BackpressureDrop)
	assert.Equal(t, BackpressureDrop, stream.backpressurePolicy)

	// Change back to block
	stream.SetBackpressurePolicy(BackpressureBlock)
	assert.Equal(t, BackpressureBlock, stream.backpressurePolicy)
}

func TestEventStream_BufferSizeAndLen(t *testing.T) {
	bufferSize := 10
	stream := NewEventStream(bufferSize)
	defer stream.Close()

	// Test BufferSize
	assert.Equal(t, bufferSize, stream.BufferSize())

	// Test Len - should start empty
	assert.Equal(t, 0, stream.Len())

	// Add some events
	stream.Send(TextDeltaEvent("event1"))
	stream.Send(TextDeltaEvent("event2"))
	stream.Send(TextDeltaEvent("event3"))

	// Len should reflect buffered events
	assert.Equal(t, 3, stream.Len())

	// Receive one event
	<-stream.Receive()

	// Len should decrease
	assert.Equal(t, 2, stream.Len())
}
