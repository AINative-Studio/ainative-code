package benchmark

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/AINative-studio/ainative-code/internal/provider"
)

const (
	// Target: Streaming latency < 50ms (time to first token)
	StreamingLatencyTargetMs = 50.0
)

// MockStreamingProvider implements a mock streaming provider for benchmarking
type MockStreamingProvider struct {
	latency time.Duration
}

func (m *MockStreamingProvider) Chat(ctx context.Context, messages []provider.Message, opts ...provider.ChatOption) (provider.Response, error) {
	return provider.Response{
		Content: "Test response",
		Usage: provider.Usage{
			PromptTokens:     10,
			CompletionTokens: 20,
			TotalTokens:      30,
		},
		Model: "mock-model",
	}, nil
}

func (m *MockStreamingProvider) Stream(ctx context.Context, messages []provider.Message, opts ...provider.StreamOption) (<-chan provider.Event, error) {
	ch := make(chan provider.Event, 10)

	go func() {
		defer close(ch)

		// Simulate latency to first token
		time.Sleep(m.latency)

		// Send first token
		ch <- provider.Event{
			Type:    provider.EventTypeContentStart,
			Content: "",
			Done:    false,
		}

		// Send content deltas
		content := "This is a test streaming response with multiple tokens."
		for i := 0; i < len(content); i += 5 {
			end := i + 5
			if end > len(content) {
				end = len(content)
			}

			select {
			case <-ctx.Done():
				return
			case ch <- provider.Event{
				Type:    provider.EventTypeContentDelta,
				Content: content[i:end],
				Done:    false,
			}:
			}

			// Small delay between tokens
			time.Sleep(1 * time.Millisecond)
		}

		// Send completion
		ch <- provider.Event{
			Type:    provider.EventTypeContentEnd,
			Content: "",
			Done:    true,
		}
	}()

	return ch, nil
}

func (m *MockStreamingProvider) Name() string {
	return "mock"
}

func (m *MockStreamingProvider) Models() []string {
	return []string{"mock-model"}
}

func (m *MockStreamingProvider) Close() error {
	return nil
}

// BenchmarkStreamingTimeToFirstToken measures latency to receive first token
func BenchmarkStreamingTimeToFirstToken(b *testing.B) {
	ctx := context.Background()

	// Test with different latency values
	latencies := []time.Duration{
		10 * time.Millisecond,
		25 * time.Millisecond,
		50 * time.Millisecond,
	}

	for _, latency := range latencies {
		b.Run(fmt.Sprintf("Latency_%dms", latency.Milliseconds()), func(b *testing.B) {
			provider := &MockStreamingProvider{latency: latency}
			messages := []provider.Message{
				{Role: "user", Content: "Test message"},
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				start := time.Now()

				eventCh, err := provider.Stream(ctx, messages)
				if err != nil {
					b.Fatalf("Stream failed: %v", err)
				}

				// Wait for first token
				firstToken := false
				for event := range eventCh {
					if event.Type == provider.EventTypeContentStart || event.Type == provider.EventTypeContentDelta {
						if !firstToken {
							elapsed := time.Since(start)
							firstToken = true

							if i == 0 {
								elapsedMs := float64(elapsed.Nanoseconds()) / 1_000_000
								b.ReportMetric(elapsedMs, "ms/first-token")

								if elapsedMs > StreamingLatencyTargetMs {
									b.Logf("WARNING: First token latency %.2fms exceeds target of %.2fms", elapsedMs, StreamingLatencyTargetMs)
								} else {
									b.Logf("SUCCESS: First token latency %.2fms meets target of %.2fms", elapsedMs, StreamingLatencyTargetMs)
								}
							}
							break
						}
					}
				}
			}
		})
	}
}

// BenchmarkStreamingResponseLatency measures overall streaming response latency
func BenchmarkStreamingResponseLatency(b *testing.B) {
	ctx := context.Background()
	provider := &MockStreamingProvider{latency: 25 * time.Millisecond}

	messages := []provider.Message{
		{Role: "user", Content: "Test message"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()

		eventCh, err := provider.Stream(ctx, messages)
		if err != nil {
			b.Fatalf("Stream failed: %v", err)
		}

		// Process all events
		for range eventCh {
			// Consume events
		}

		elapsed := time.Since(start)

		if i == 0 {
			b.ReportMetric(float64(elapsed.Nanoseconds())/1_000_000, "ms/total")
		}
	}
}

// BenchmarkStreamingWithVariousMessageSizes tests streaming with different message sizes
func BenchmarkStreamingWithVariousMessageSizes(b *testing.B) {
	ctx := context.Background()
	provider := &MockStreamingProvider{latency: 25 * time.Millisecond}

	messageSizes := []int{10, 100, 1000, 10000}

	for _, size := range messageSizes {
		b.Run(fmt.Sprintf("MessageSize_%d", size), func(b *testing.B) {
			// Create message of specified size
			content := make([]byte, size)
			for i := range content {
				content[i] = 'a'
			}

			messages := []provider.Message{
				{Role: "user", Content: string(content)},
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				start := time.Now()

				eventCh, err := provider.Stream(ctx, messages)
				if err != nil {
					b.Fatalf("Stream failed: %v", err)
				}

				// Wait for first token
				for event := range eventCh {
					if event.Type == provider.EventTypeContentStart || event.Type == provider.EventTypeContentDelta {
						elapsed := time.Since(start)

						if i == 0 {
							b.ReportMetric(float64(elapsed.Nanoseconds())/1_000_000, "ms/first-token")
						}
						break
					}
				}
			}
		})
	}
}

// BenchmarkStreamingChannelOverhead measures the overhead of channel operations
func BenchmarkStreamingChannelOverhead(b *testing.B) {
	ctx := context.Background()

	b.Run("ChannelSend", func(b *testing.B) {
		ch := make(chan provider.Event, 100)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ch <- provider.Event{
				Type:    provider.EventTypeContentDelta,
				Content: "test",
			}
		}
		close(ch)

		// Drain channel
		for range ch {
		}
	})

	b.Run("ChannelReceive", func(b *testing.B) {
		ch := make(chan provider.Event, b.N)

		// Fill channel
		for i := 0; i < b.N; i++ {
			ch <- provider.Event{
				Type:    provider.EventTypeContentDelta,
				Content: "test",
			}
		}
		close(ch)

		b.ResetTimer()
		for range ch {
			// Consume events
		}
	})

	b.Run("ChannelRoundTrip", func(b *testing.B) {
		provider := &MockStreamingProvider{latency: 1 * time.Millisecond}
		messages := []provider.Message{
			{Role: "user", Content: "Test"},
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			eventCh, _ := provider.Stream(ctx, messages)
			for range eventCh {
			}
		}
	})
}

// BenchmarkStreamingConcurrentStreams measures performance with concurrent streams
func BenchmarkStreamingConcurrentStreams(b *testing.B) {
	ctx := context.Background()
	provider := &MockStreamingProvider{latency: 10 * time.Millisecond}

	messages := []provider.Message{
		{Role: "user", Content: "Test message"},
	}

	concurrency := []int{1, 5, 10, 20}

	for _, n := range concurrency {
		b.Run(fmt.Sprintf("Concurrent_%d", n), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Start concurrent streams
				done := make(chan bool, n)

				for j := 0; j < n; j++ {
					go func() {
						eventCh, err := provider.Stream(ctx, messages)
						if err != nil {
							done <- false
							return
						}

						for range eventCh {
						}
						done <- true
					}()
				}

				// Wait for all to complete
				for j := 0; j < n; j++ {
					<-done
				}
			}
		})
	}
}

// BenchmarkStreamingContextCancellation measures context cancellation overhead
func BenchmarkStreamingContextCancellation(b *testing.B) {
	provider := &MockStreamingProvider{latency: 10 * time.Millisecond}

	messages := []provider.Message{
		{Role: "user", Content: "Test message"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)

		eventCh, err := provider.Stream(ctx, messages)
		if err != nil {
			cancel()
			continue
		}

		// Process until context is cancelled
		for range eventCh {
		}

		cancel()
	}
}

// BenchmarkStreamingThinkingBlocks measures streaming with thinking blocks
func BenchmarkStreamingThinkingBlocks(b *testing.B) {
	ctx := context.Background()

	// Mock provider that includes thinking blocks
	provider := &MockStreamingProvider{latency: 20 * time.Millisecond}

	messages := []provider.Message{
		{Role: "user", Content: "Complex question requiring thinking"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()

		eventCh, err := provider.Stream(ctx, messages)
		if err != nil {
			b.Fatalf("Stream failed: %v", err)
		}

		// Process events including thinking blocks
		thinkingCount := 0
		for event := range eventCh {
			if event.Type == provider.EventTypeThinking {
				thinkingCount++
			}
		}

		elapsed := time.Since(start)

		if i == 0 {
			b.ReportMetric(float64(elapsed.Nanoseconds())/1_000_000, "ms/with-thinking")
			b.Logf("Processed %d thinking blocks", thinkingCount)
		}
	}
}
