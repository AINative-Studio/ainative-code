package rlhf

import (
	"context"
	"testing"
	"time"

	"github.com/AINative-studio/ainative-code/internal/client/rlhf"
	"github.com/AINative-studio/ainative-code/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockRLHFClient is a mock implementation of the RLHF client for testing
type MockRLHFClient struct {
	SubmittedBatches []*rlhf.BatchInteractionFeedback
	ShouldFail       bool
}

func (m *MockRLHFClient) SubmitBatchInteractionFeedback(ctx context.Context, batch *rlhf.BatchInteractionFeedback) (*rlhf.BatchInteractionFeedbackResponse, error) {
	if m.ShouldFail {
		return nil, assert.AnError
	}

	m.SubmittedBatches = append(m.SubmittedBatches, batch)

	return &rlhf.BatchInteractionFeedbackResponse{
		Successful:     make([]string, len(batch.Interactions)),
		TotalProcessed: len(batch.Interactions),
		SuccessCount:   len(batch.Interactions),
		FailureCount:   0,
	}, nil
}

func TestNewCollector(t *testing.T) {
	cfg := &config.RLHFConfig{
		AutoCollect:   true,
		BatchSize:     10,
		BatchInterval: 1 * time.Minute,
	}

	collector := NewCollector(cfg, nil)

	assert.NotNil(t, collector)
	assert.Equal(t, cfg, collector.config)
	assert.NotNil(t, collector.sessionID)
	assert.Empty(t, collector.queue)
}

func TestCollector_CaptureInteraction(t *testing.T) {
	tests := []struct {
		name      string
		optOut    bool
		autoCollect bool
		wantCaptured bool
	}{
		{
			name:         "capture when enabled",
			optOut:       false,
			autoCollect:  true,
			wantCaptured: true,
		},
		{
			name:         "skip when opted out",
			optOut:       true,
			autoCollect:  true,
			wantCaptured: false,
		},
		{
			name:         "skip when disabled",
			optOut:       false,
			autoCollect:  false,
			wantCaptured: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.RLHFConfig{
				AutoCollect: tt.autoCollect,
				OptOut:      tt.optOut,
				BatchSize:   10,
			}

			collector := NewCollector(cfg, nil)

			interactionID := collector.CaptureInteraction("test prompt", "test response", "model-123")

			if tt.wantCaptured {
				assert.NotEmpty(t, interactionID)
				assert.Equal(t, 1, collector.GetQueueSize())
				assert.Equal(t, 1, collector.GetInteractionCount())
			} else {
				assert.Empty(t, interactionID)
				assert.Equal(t, 0, collector.GetQueueSize())
			}
		})
	}
}

func TestCollector_RecordImplicitFeedback(t *testing.T) {
	cfg := &config.RLHFConfig{
		AutoCollect: true,
		BatchSize:   10,
		ImplicitFeedback: &config.ImplicitFeedbackConfig{
			Enabled:           true,
			RegenerateScore:   0.2,
			EditResponseScore: 0.3,
			CopyResponseScore: 0.8,
			ContinueScore:     0.7,
		},
	}

	collector := NewCollector(cfg, nil)

	// Capture an interaction
	interactionID := collector.CaptureInteraction("test", "response", "model-1")
	require.NotEmpty(t, interactionID)

	// Record implicit feedback
	collector.RecordImplicitFeedback(interactionID, ActionRegenerate)

	// Verify signal was recorded
	collector.mu.RLock()
	interaction := collector.queue[0]
	collector.mu.RUnlock()

	assert.Equal(t, 1, len(interaction.ImplicitSignals))
	assert.Equal(t, "regenerate", interaction.ImplicitSignals[0].Action)
	assert.Equal(t, 0.2, interaction.ImplicitSignals[0].Score)
	assert.Equal(t, 0.2, interaction.ImplicitScore)
}

func TestCollector_RecordExplicitFeedback(t *testing.T) {
	cfg := &config.RLHFConfig{
		AutoCollect: true,
		BatchSize:   10,
	}

	collector := NewCollector(cfg, nil)

	// Capture an interaction
	interactionID := collector.CaptureInteraction("test", "response", "model-1")
	require.NotEmpty(t, interactionID)

	// Record explicit feedback
	collector.RecordExplicitFeedback(interactionID, 0.95, "Great response!")

	// Verify feedback was recorded
	collector.mu.RLock()
	interaction := collector.queue[0]
	collector.mu.RUnlock()

	assert.True(t, interaction.HasExplicitScore)
	assert.Equal(t, 0.95, interaction.ExplicitScore)
	assert.Equal(t, "Great response!", interaction.UserFeedback)
}

func TestCollector_ShouldPromptForFeedback(t *testing.T) {
	tests := []struct {
		name             string
		promptInterval   int
		interactionCount int
		want             bool
	}{
		{
			name:             "prompt at interval",
			promptInterval:   5,
			interactionCount: 5,
			want:             true,
		},
		{
			name:             "no prompt before interval",
			promptInterval:   5,
			interactionCount: 3,
			want:             false,
		},
		{
			name:             "no prompt when disabled",
			promptInterval:   0,
			interactionCount: 10,
			want:             false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.RLHFConfig{
				AutoCollect:    true,
				PromptInterval: tt.promptInterval,
			}

			collector := NewCollector(cfg, nil)
			collector.interactionCount = tt.interactionCount

			got := collector.ShouldPromptForFeedback()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCollector_CalculateImplicitScore(t *testing.T) {
	cfg := &config.RLHFConfig{
		AutoCollect: true,
	}

	collector := NewCollector(cfg, nil)

	tests := []struct {
		name     string
		signals  []ImplicitSignal
		wantMin  float64
		wantMax  float64
	}{
		{
			name:     "no signals returns neutral",
			signals:  []ImplicitSignal{},
			wantMin:  0.5,
			wantMax:  0.5,
		},
		{
			name: "single positive signal",
			signals: []ImplicitSignal{
				{Action: "copy", Score: 0.8},
			},
			wantMin: 0.8,
			wantMax: 0.8,
		},
		{
			name: "mixed signals weighted average",
			signals: []ImplicitSignal{
				{Action: "regenerate", Score: 0.2},
				{Action: "copy", Score: 0.8},
			},
			wantMin: 0.0,
			wantMax: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := collector.calculateImplicitScore(tt.signals)
			assert.GreaterOrEqual(t, score, tt.wantMin)
			assert.LessOrEqual(t, score, tt.wantMax)
		})
	}
}

func TestCollector_ImplicitFeedbackActions(t *testing.T) {
	cfg := &config.RLHFConfig{
		AutoCollect: true,
		ImplicitFeedback: &config.ImplicitFeedbackConfig{
			Enabled:           true,
			RegenerateScore:   0.2,
			EditResponseScore: 0.3,
			CopyResponseScore: 0.8,
			ContinueScore:     0.7,
		},
	}

	collector := NewCollector(cfg, nil)

	tests := []struct {
		name        string
		action      FeedbackAction
		expectScore float64
	}{
		{
			name:        "regenerate action",
			action:      ActionRegenerate,
			expectScore: 0.2,
		},
		{
			name:        "edit action",
			action:      ActionEdit,
			expectScore: 0.3,
		},
		{
			name:        "copy action",
			action:      ActionCopy,
			expectScore: 0.8,
		},
		{
			name:        "continue action",
			action:      ActionContinue,
			expectScore: 0.7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interactionID := collector.CaptureInteraction("test", "response", "model")
			collector.RecordImplicitFeedback(interactionID, tt.action)

			collector.mu.RLock()
			interaction := collector.queue[len(collector.queue)-1]
			collector.mu.RUnlock()

			require.Len(t, interaction.ImplicitSignals, 1)
			assert.Equal(t, tt.expectScore, interaction.ImplicitSignals[0].Score)
		})
	}
}

func TestCollector_QueueManagement(t *testing.T) {
	cfg := &config.RLHFConfig{
		AutoCollect: true,
		BatchSize:   3,
	}

	collector := NewCollector(cfg, nil)

	// Add multiple interactions
	for i := 0; i < 5; i++ {
		collector.CaptureInteraction("prompt", "response", "model")
	}

	assert.Equal(t, 5, collector.GetQueueSize())
	assert.Equal(t, 5, collector.GetInteractionCount())
}

func TestCollector_PrivacyOptOut(t *testing.T) {
	cfg := &config.RLHFConfig{
		AutoCollect: true,
		OptOut:      true,
	}

	collector := NewCollector(cfg, nil)

	// Try to capture interaction
	interactionID := collector.CaptureInteraction("test", "response", "model")

	// Should not capture when opted out
	assert.Empty(t, interactionID)
	assert.Equal(t, 0, collector.GetQueueSize())
}

func TestCollector_MultipleImplicitSignals(t *testing.T) {
	cfg := &config.RLHFConfig{
		AutoCollect: true,
		ImplicitFeedback: &config.ImplicitFeedbackConfig{
			Enabled:           true,
			RegenerateScore:   0.2,
			EditResponseScore: 0.3,
			CopyResponseScore: 0.8,
			ContinueScore:     0.7,
		},
	}

	collector := NewCollector(cfg, nil)

	// Capture interaction
	interactionID := collector.CaptureInteraction("test", "response", "model")

	// Record multiple signals
	collector.RecordImplicitFeedback(interactionID, ActionRegenerate)
	time.Sleep(10 * time.Millisecond)
	collector.RecordImplicitFeedback(interactionID, ActionCopy)

	collector.mu.RLock()
	interaction := collector.queue[0]
	collector.mu.RUnlock()

	// Should have both signals
	assert.Len(t, interaction.ImplicitSignals, 2)

	// Score should be weighted (more recent signals weighted higher)
	assert.Greater(t, interaction.ImplicitScore, 0.2)
	assert.Less(t, interaction.ImplicitScore, 0.8)
}

func TestCollector_ExplicitOverridesImplicit(t *testing.T) {
	cfg := &config.RLHFConfig{
		AutoCollect: true,
		ImplicitFeedback: &config.ImplicitFeedbackConfig{
			Enabled:         true,
			RegenerateScore: 0.2,
		},
	}

	collector := NewCollector(cfg, nil)

	// Capture interaction
	interactionID := collector.CaptureInteraction("test", "response", "model")

	// Record implicit feedback (low score)
	collector.RecordImplicitFeedback(interactionID, ActionRegenerate)

	// Record explicit feedback (high score)
	collector.RecordExplicitFeedback(interactionID, 0.95, "Actually good!")

	collector.mu.RLock()
	interaction := collector.queue[0]
	collector.mu.RUnlock()

	// Explicit should be preferred
	assert.True(t, interaction.HasExplicitScore)
	assert.Equal(t, 0.95, interaction.ExplicitScore)
	assert.Equal(t, 0.2, interaction.ImplicitScore) // Implicit still recorded
}
