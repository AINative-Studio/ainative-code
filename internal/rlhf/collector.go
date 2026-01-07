package rlhf

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/AINative-studio/ainative-code/internal/client/rlhf"
	"github.com/AINative-studio/ainative-code/internal/config"
	"github.com/AINative-studio/ainative-code/internal/logger"
	"github.com/google/uuid"
)

// Collector manages automatic collection and submission of RLHF data
type Collector struct {
	config        *config.RLHFConfig
	client        *rlhf.Client
	queue         []*InteractionData
	mu            sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	sessionID     string
	interactionCount int
}

// InteractionData represents a captured interaction with metadata
type InteractionData struct {
	ID               string
	Prompt           string
	Response         string
	Timestamp        time.Time
	ModelID          string
	SessionID        string
	ImplicitScore    float64
	ExplicitScore    float64
	UserFeedback     string
	Metadata         map[string]interface{}
	ImplicitSignals  []ImplicitSignal
	HasExplicitScore bool
}

// ImplicitSignal represents an implicit feedback action
type ImplicitSignal struct {
	Action    string    // "regenerate", "edit", "copy", "continue"
	Timestamp time.Time
	Score     float64
}

// FeedbackAction represents user actions that indicate quality
type FeedbackAction string

const (
	ActionRegenerate FeedbackAction = "regenerate"
	ActionEdit       FeedbackAction = "edit"
	ActionCopy       FeedbackAction = "copy"
	ActionContinue   FeedbackAction = "continue"
)

// NewCollector creates a new RLHF collector instance
func NewCollector(cfg *config.RLHFConfig, client *rlhf.Client) *Collector {
	ctx, cancel := context.WithCancel(context.Background())

	return &Collector{
		config:    cfg,
		client:    client,
		queue:     make([]*InteractionData, 0),
		ctx:       ctx,
		cancel:    cancel,
		sessionID: uuid.New().String(),
	}
}

// Start begins the background collection and submission worker
func (c *Collector) Start() error {
	if c.config.OptOut {
		logger.Info("RLHF auto-collection is opted out")
		return nil
	}

	if !c.config.AutoCollect {
		logger.Debug("RLHF auto-collection is disabled")
		return nil
	}

	logger.InfoEvent().
		Bool("auto_collect", c.config.AutoCollect).
		Int("batch_size", c.config.BatchSize).
		Dur("batch_interval", c.config.BatchInterval).
		Msg("Starting RLHF auto-collector")

	// Start background worker
	c.wg.Add(1)
	go c.worker()

	return nil
}

// Stop gracefully shuts down the collector
func (c *Collector) Stop() error {
	logger.Info("Stopping RLHF auto-collector")

	// Cancel context to signal worker to stop
	c.cancel()

	// Wait for worker to finish
	c.wg.Wait()

	// Flush remaining interactions
	if err := c.flush(); err != nil {
		logger.ErrorEvent().Err(err).Msg("Failed to flush remaining interactions")
		return err
	}

	logger.Info("RLHF auto-collector stopped")
	return nil
}

// CaptureInteraction captures a new interaction for potential submission
func (c *Collector) CaptureInteraction(prompt, response, modelID string) string {
	if c.config.OptOut || !c.config.AutoCollect {
		return ""
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	interactionID := uuid.New().String()

	interaction := &InteractionData{
		ID:              interactionID,
		Prompt:          prompt,
		Response:        response,
		Timestamp:       time.Now(),
		ModelID:         modelID,
		SessionID:       c.sessionID,
		ImplicitScore:   0.0,
		ImplicitSignals: make([]ImplicitSignal, 0),
		Metadata: map[string]interface{}{
			"session_id": c.sessionID,
			"auto_captured": true,
		},
	}

	c.queue = append(c.queue, interaction)
	c.interactionCount++

	logger.DebugEvent().
		Str("interaction_id", interactionID).
		Int("queue_size", len(c.queue)).
		Msg("Captured interaction")

	return interactionID
}

// RecordImplicitFeedback records an implicit feedback signal
func (c *Collector) RecordImplicitFeedback(interactionID string, action FeedbackAction) {
	if c.config.OptOut || !c.config.AutoCollect {
		return
	}

	if c.config.ImplicitFeedback == nil || !c.config.ImplicitFeedback.Enabled {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Find the interaction
	var interaction *InteractionData
	for _, inter := range c.queue {
		if inter.ID == interactionID {
			interaction = inter
			break
		}
	}

	if interaction == nil {
		logger.WarnEvent().
			Str("interaction_id", interactionID).
			Msg("Interaction not found for implicit feedback")
		return
	}

	// Calculate score based on action
	var score float64
	switch action {
	case ActionRegenerate:
		score = c.config.ImplicitFeedback.RegenerateScore
	case ActionEdit:
		score = c.config.ImplicitFeedback.EditResponseScore
	case ActionCopy:
		score = c.config.ImplicitFeedback.CopyResponseScore
	case ActionContinue:
		score = c.config.ImplicitFeedback.ContinueScore
	default:
		logger.WarnEvent().
			Str("action", string(action)).
			Msg("Unknown implicit feedback action")
		return
	}

	signal := ImplicitSignal{
		Action:    string(action),
		Timestamp: time.Now(),
		Score:     score,
	}

	interaction.ImplicitSignals = append(interaction.ImplicitSignals, signal)

	// Update implicit score (weighted average or latest score)
	interaction.ImplicitScore = c.calculateImplicitScore(interaction.ImplicitSignals)

	logger.DebugEvent().
		Str("interaction_id", interactionID).
		Str("action", string(action)).
		Float64("score", score).
		Float64("total_implicit_score", interaction.ImplicitScore).
		Msg("Recorded implicit feedback")
}

// RecordExplicitFeedback records explicit user feedback
func (c *Collector) RecordExplicitFeedback(interactionID string, score float64, feedback string) {
	if c.config.OptOut || !c.config.AutoCollect {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Find the interaction
	var interaction *InteractionData
	for _, inter := range c.queue {
		if inter.ID == interactionID {
			interaction = inter
			break
		}
	}

	if interaction == nil {
		logger.WarnEvent().
			Str("interaction_id", interactionID).
			Msg("Interaction not found for explicit feedback")
		return
	}

	interaction.ExplicitScore = score
	interaction.UserFeedback = feedback
	interaction.HasExplicitScore = true

	logger.InfoEvent().
		Str("interaction_id", interactionID).
		Float64("score", score).
		Msg("Recorded explicit feedback")
}

// ShouldPromptForFeedback determines if user should be prompted for feedback
func (c *Collector) ShouldPromptForFeedback() bool {
	if c.config.OptOut || !c.config.AutoCollect {
		return false
	}

	if c.config.PromptInterval <= 0 {
		return false
	}

	return c.interactionCount > 0 && c.interactionCount%c.config.PromptInterval == 0
}

// GetQueueSize returns the current queue size
func (c *Collector) GetQueueSize() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.queue)
}

// GetInteractionCount returns the total interaction count
func (c *Collector) GetInteractionCount() int {
	return c.interactionCount
}

// worker runs in the background to periodically submit batches
func (c *Collector) worker() {
	defer c.wg.Done()

	ticker := time.NewTicker(c.config.BatchInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			logger.Debug("RLHF collector worker shutting down")
			return
		case <-ticker.C:
			c.processBatch()
		}
	}
}

// processBatch processes and submits a batch of interactions
func (c *Collector) processBatch() {
	c.mu.Lock()

	if len(c.queue) == 0 {
		c.mu.Unlock()
		return
	}

	// Don't submit if batch size not reached (unless interval is long)
	if len(c.queue) < c.config.BatchSize && c.config.BatchInterval < 5*time.Minute {
		c.mu.Unlock()
		return
	}

	// Get batch to submit (up to batch size)
	batchSize := c.config.BatchSize
	if len(c.queue) < batchSize {
		batchSize = len(c.queue)
	}

	batch := c.queue[:batchSize]
	c.queue = c.queue[batchSize:]
	c.mu.Unlock()

	logger.InfoEvent().
		Int("batch_size", len(batch)).
		Msg("Processing RLHF batch")

	if err := c.submitBatch(batch); err != nil {
		logger.ErrorEvent().
			Err(err).
			Int("batch_size", len(batch)).
			Msg("Failed to submit RLHF batch, re-queuing")

		// Re-queue failed batch
		c.mu.Lock()
		c.queue = append(batch, c.queue...)
		c.mu.Unlock()
	}
}

// submitBatch submits a batch of interactions to the RLHF API
func (c *Collector) submitBatch(batch []*InteractionData) error {
	if c.client == nil {
		return fmt.Errorf("RLHF client not initialized")
	}

	// Convert to RLHF client format
	feedbackBatch := &rlhf.BatchInteractionFeedback{
		Interactions: make([]*rlhf.InteractionFeedback, 0, len(batch)),
	}

	for _, interaction := range batch {
		// Calculate final score (prefer explicit over implicit)
		score := interaction.ImplicitScore
		if interaction.HasExplicitScore {
			score = interaction.ExplicitScore
		}

		// Normalize score to 0.0-1.0 range if needed
		if score < 0.0 {
			score = 0.0
		} else if score > 1.0 {
			score = 1.0
		}

		// Add metadata about implicit signals
		metadata := interaction.Metadata
		if metadata == nil {
			metadata = make(map[string]interface{})
		}
		metadata["implicit_signals_count"] = len(interaction.ImplicitSignals)
		metadata["has_explicit_feedback"] = interaction.HasExplicitScore
		if interaction.UserFeedback != "" {
			metadata["user_feedback"] = interaction.UserFeedback
		}

		feedback := &rlhf.InteractionFeedback{
			Prompt:    interaction.Prompt,
			Response:  interaction.Response,
			Score:     score,
			ModelID:   interaction.ModelID,
			SessionID: interaction.SessionID,
			Timestamp: interaction.Timestamp,
			Metadata:  metadata,
		}

		feedbackBatch.Interactions = append(feedbackBatch.Interactions, feedback)
	}

	// Submit batch
	ctx, cancel := context.WithTimeout(c.ctx, c.config.Timeout)
	defer cancel()

	result, err := c.client.SubmitBatchInteractionFeedback(ctx, feedbackBatch)
	if err != nil {
		return err
	}

	logger.InfoEvent().
		Int("success_count", result.SuccessCount).
		Int("failure_count", result.FailureCount).
		Msg("RLHF batch submitted")

	return nil
}

// flush submits all remaining interactions in the queue
func (c *Collector) flush() error {
	c.mu.Lock()

	if len(c.queue) == 0 {
		c.mu.Unlock()
		return nil
	}

	batch := c.queue
	c.queue = make([]*InteractionData, 0)
	c.mu.Unlock()

	logger.InfoEvent().
		Int("batch_size", len(batch)).
		Msg("Flushing remaining RLHF interactions")

	return c.submitBatch(batch)
}

// calculateImplicitScore calculates the overall implicit score from signals
func (c *Collector) calculateImplicitScore(signals []ImplicitSignal) float64 {
	if len(signals) == 0 {
		return 0.5 // Neutral score if no signals
	}

	// Use weighted average, with more recent signals weighted higher
	totalWeight := 0.0
	weightedSum := 0.0

	for i, signal := range signals {
		// More recent signals have higher weight (exponential decay)
		weight := float64(i+1) / float64(len(signals))
		weightedSum += signal.Score * weight
		totalWeight += weight
	}

	if totalWeight == 0 {
		return 0.5
	}

	score := weightedSum / totalWeight

	// Normalize to 0.0-1.0 range
	if score < 0.0 {
		score = 0.0
	} else if score > 1.0 {
		score = 1.0
	}

	return score
}
