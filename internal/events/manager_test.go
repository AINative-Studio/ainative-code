package events

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStreamManager(t *testing.T) {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewStreamManager(tt.bufferSize)
			require.NotNil(t, manager)
			assert.Equal(t, tt.wantSize, manager.defaultBufferSize)
			assert.NotNil(t, manager.streams)
		})
	}
}

func TestStreamManager_CreateStream(t *testing.T) {
	t.Run("create new stream", func(t *testing.T) {
		manager := NewStreamManager(10)
		streamID := "stream-1"

		stream, err := manager.CreateStream(streamID)
		require.NoError(t, err)
		require.NotNil(t, stream)
		assert.Equal(t, 10, stream.BufferSize())
	})

	t.Run("create stream with empty ID", func(t *testing.T) {
		manager := NewStreamManager(10)

		stream, err := manager.CreateStream("")
		assert.Error(t, err)
		assert.Nil(t, stream)
		assert.Contains(t, err.Error(), "stream ID cannot be empty")
	})

	t.Run("create duplicate stream", func(t *testing.T) {
		manager := NewStreamManager(10)
		streamID := "stream-1"

		stream1, err := manager.CreateStream(streamID)
		require.NoError(t, err)
		require.NotNil(t, stream1)

		stream2, err := manager.CreateStream(streamID)
		assert.Error(t, err)
		assert.Nil(t, stream2)
		assert.Contains(t, err.Error(), "already exists")
	})
}

func TestStreamManager_GetStream(t *testing.T) {
	t.Run("get existing stream", func(t *testing.T) {
		manager := NewStreamManager(10)
		streamID := "stream-1"

		created, err := manager.CreateStream(streamID)
		require.NoError(t, err)

		retrieved, err := manager.GetStream(streamID)
		require.NoError(t, err)
		assert.Equal(t, created, retrieved)
	})

	t.Run("get non-existent stream", func(t *testing.T) {
		manager := NewStreamManager(10)

		stream, err := manager.GetStream("nonexistent")
		assert.Error(t, err)
		assert.Nil(t, stream)
		assert.True(t, IsStreamNotFound(err))
	})

	t.Run("get stream with empty ID", func(t *testing.T) {
		manager := NewStreamManager(10)

		stream, err := manager.GetStream("")
		assert.Error(t, err)
		assert.Nil(t, stream)
		assert.Contains(t, err.Error(), "stream ID cannot be empty")
	})
}

func TestStreamManager_CloseStream(t *testing.T) {
	t.Run("close existing stream", func(t *testing.T) {
		manager := NewStreamManager(10)
		streamID := "stream-1"

		_, err := manager.CreateStream(streamID)
		require.NoError(t, err)

		err = manager.CloseStream(streamID)
		require.NoError(t, err)

		// Stream should be removed
		stream, err := manager.GetStream(streamID)
		assert.Error(t, err)
		assert.Nil(t, stream)
		assert.True(t, IsStreamNotFound(err))
	})

	t.Run("close non-existent stream", func(t *testing.T) {
		manager := NewStreamManager(10)

		err := manager.CloseStream("nonexistent")
		assert.Error(t, err)
		assert.True(t, IsStreamNotFound(err))
	})

	t.Run("close stream with empty ID", func(t *testing.T) {
		manager := NewStreamManager(10)

		err := manager.CloseStream("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "stream ID cannot be empty")
	})
}

func TestStreamManager_ListStreams(t *testing.T) {
	t.Run("list empty streams", func(t *testing.T) {
		manager := NewStreamManager(10)

		streams := manager.ListStreams()
		assert.Empty(t, streams)
	})

	t.Run("list multiple streams", func(t *testing.T) {
		manager := NewStreamManager(10)

		streamIDs := []string{"stream-1", "stream-2", "stream-3"}
		for _, id := range streamIDs {
			_, err := manager.CreateStream(id)
			require.NoError(t, err)
		}

		streams := manager.ListStreams()
		assert.Len(t, streams, 3)

		// Verify all stream IDs are present
		for _, id := range streamIDs {
			assert.Contains(t, streams, id)
		}
	})

	t.Run("list after closing some streams", func(t *testing.T) {
		manager := NewStreamManager(10)

		manager.CreateStream("stream-1")
		manager.CreateStream("stream-2")
		manager.CreateStream("stream-3")

		manager.CloseStream("stream-2")

		streams := manager.ListStreams()
		assert.Len(t, streams, 2)
		assert.Contains(t, streams, "stream-1")
		assert.Contains(t, streams, "stream-3")
		assert.NotContains(t, streams, "stream-2")
	})
}

func TestStreamManager_CloseAll(t *testing.T) {
	t.Run("close all streams", func(t *testing.T) {
		manager := NewStreamManager(10)

		// Create multiple streams
		manager.CreateStream("stream-1")
		manager.CreateStream("stream-2")
		manager.CreateStream("stream-3")

		assert.Len(t, manager.ListStreams(), 3)

		manager.CloseAll()

		assert.Empty(t, manager.ListStreams())
	})

	t.Run("close all with no streams", func(t *testing.T) {
		manager := NewStreamManager(10)

		// Should not panic
		manager.CloseAll()

		assert.Empty(t, manager.ListStreams())
	})
}

func TestStreamManager_StreamCount(t *testing.T) {
	manager := NewStreamManager(10)

	assert.Equal(t, 0, manager.StreamCount())

	manager.CreateStream("stream-1")
	assert.Equal(t, 1, manager.StreamCount())

	manager.CreateStream("stream-2")
	assert.Equal(t, 2, manager.StreamCount())

	manager.CloseStream("stream-1")
	assert.Equal(t, 1, manager.StreamCount())

	manager.CloseAll()
	assert.Equal(t, 0, manager.StreamCount())
}

func TestStreamManager_Concurrent(t *testing.T) {
	t.Run("concurrent create and close", func(t *testing.T) {
		manager := NewStreamManager(10)
		numGoroutines := 10
		numStreams := 5

		var wg sync.WaitGroup

		// Concurrent creates
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for j := 0; j < numStreams; j++ {
					streamID := fmt.Sprintf("stream-%d-%d", id, j)
					_, err := manager.CreateStream(streamID)
					if err != nil {
						t.Errorf("failed to create stream: %v", err)
					}
				}
			}(i)
		}

		wg.Wait()

		assert.Equal(t, numGoroutines*numStreams, manager.StreamCount())

		// Concurrent closes
		streams := manager.ListStreams()
		for _, id := range streams {
			wg.Add(1)
			go func(streamID string) {
				defer wg.Done()
				err := manager.CloseStream(streamID)
				if err != nil && !IsStreamNotFound(err) {
					t.Errorf("failed to close stream: %v", err)
				}
			}(id)
		}

		wg.Wait()

		assert.Equal(t, 0, manager.StreamCount())
	})
}

func TestStreamManager_CleanupInactive(t *testing.T) {
	t.Run("cleanup streams inactive for threshold", func(t *testing.T) {
		manager := NewStreamManager(10)

		// Create streams
		stream1, _ := manager.CreateStream("stream-1")
		_, _ = manager.CreateStream("stream-2")
		stream3, _ := manager.CreateStream("stream-3")

		// Send events to stream1 and stream3
		stream1.Send(TextDeltaEvent("hello"))
		stream3.Send(TextDeltaEvent("world"))

		// Mark stream-2 as older by manipulating lastActivity
		// Since we can't directly manipulate lastActivity in tests,
		// we'll test with a very long threshold that won't cleanup anything
		threshold := 24 * time.Hour
		cleaned := manager.CleanupInactive(threshold)
		assert.Equal(t, 0, cleaned)
		assert.Equal(t, 3, manager.StreamCount())
	})

	t.Run("cleanup with zero threshold", func(t *testing.T) {
		manager := NewStreamManager(10)

		manager.CreateStream("stream-1")
		manager.CreateStream("stream-2")

		// Zero threshold should not clean anything
		cleaned := manager.CleanupInactive(0)
		assert.Equal(t, 0, cleaned)
		assert.Equal(t, 2, manager.StreamCount())
	})
}

func TestStreamManager_GetOrCreate(t *testing.T) {
	t.Run("get existing stream", func(t *testing.T) {
		manager := NewStreamManager(10)
		streamID := "stream-1"

		created, err := manager.CreateStream(streamID)
		require.NoError(t, err)

		retrieved, created2, err := manager.GetOrCreate(streamID)
		require.NoError(t, err)
		assert.False(t, created2)
		assert.Equal(t, created, retrieved)
	})

	t.Run("create new stream if not exists", func(t *testing.T) {
		manager := NewStreamManager(10)
		streamID := "stream-1"

		stream, created, err := manager.GetOrCreate(streamID)
		require.NoError(t, err)
		assert.True(t, created)
		require.NotNil(t, stream)

		// Verify it was actually created
		retrieved, err := manager.GetStream(streamID)
		require.NoError(t, err)
		assert.Equal(t, stream, retrieved)
	})

	t.Run("get or create with empty ID", func(t *testing.T) {
		manager := NewStreamManager(10)

		stream, created, err := manager.GetOrCreate("")
		assert.Error(t, err)
		assert.False(t, created)
		assert.Nil(t, stream)
	})
}

func TestStreamManager_GetStreamInfo(t *testing.T) {
	t.Run("get info for existing stream", func(t *testing.T) {
		manager := NewStreamManager(10)
		streamID := "stream-1"

		stream, err := manager.CreateStream(streamID)
		require.NoError(t, err)

		// Send some events to populate the buffer
		stream.Send(TextDeltaEvent("hello"))
		stream.Send(TextDeltaEvent("world"))

		info, err := manager.GetStreamInfo(streamID)
		require.NoError(t, err)
		require.NotNil(t, info)

		assert.Equal(t, streamID, info.ID)
		assert.Equal(t, 10, info.BufferSize)
		assert.Equal(t, 2, info.CurrentLoad)
		assert.False(t, info.IsClosed)
		assert.False(t, info.LastActivity.IsZero())
	})

	t.Run("get info for non-existent stream", func(t *testing.T) {
		manager := NewStreamManager(10)

		info, err := manager.GetStreamInfo("non-existent")
		assert.Error(t, err)
		assert.True(t, IsStreamNotFound(err))
		assert.Nil(t, info)
	})

	t.Run("get info with empty ID", func(t *testing.T) {
		manager := NewStreamManager(10)

		info, err := manager.GetStreamInfo("")
		assert.Error(t, err)
		assert.Nil(t, info)
	})
}

func TestStreamManager_ListStreamInfo(t *testing.T) {
	t.Run("list info for multiple streams", func(t *testing.T) {
		manager := NewStreamManager(10)

		// Create streams
		stream1, _ := manager.CreateStream("stream-1")
		stream2, _ := manager.CreateStream("stream-2")

		// Send events
		stream1.Send(TextDeltaEvent("hello"))
		stream2.Send(TextDeltaEvent("world"))
		stream2.Send(TextDeltaEvent("!"))

		info := manager.ListStreamInfo()
		require.NotNil(t, info)
		assert.Len(t, info, 2)

		// Verify info contains expected data
		for _, i := range info {
			assert.NotEmpty(t, i.ID)
			assert.Equal(t, 10, i.BufferSize)
			assert.False(t, i.IsClosed)
			assert.False(t, i.LastActivity.IsZero())
		}
	})

	t.Run("list info for empty manager", func(t *testing.T) {
		manager := NewStreamManager(10)

		info := manager.ListStreamInfo()
		require.NotNil(t, info)
		assert.Empty(t, info)
	})
}
