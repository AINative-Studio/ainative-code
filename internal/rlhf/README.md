# RLHF Auto-Collection Package

This package implements automatic collection of RLHF (Reinforcement Learning from Human Feedback) data for improving AI model performance.

## Architecture

### Components

- **Collector**: Main service that manages interaction capture, queuing, and submission
- **FeedbackPrompt**: TUI component for collecting explicit user feedback
- **Types**: Data structures for interactions and feedback signals

### Key Features

1. **Automatic Capture**: Captures all user-AI interactions automatically
2. **Implicit Feedback**: Tracks user actions (regenerate, copy, edit, continue)
3. **Explicit Feedback**: Periodic prompts for user ratings and comments
4. **Background Processing**: Non-blocking batch submission to API
5. **Privacy Controls**: Opt-out mechanism and review-before-submit option
6. **Retry Logic**: Automatic retry with exponential backoff on failures

## Usage

### Initializing the Collector

```go
import (
    "github.com/AINative-studio/ainative-code/internal/rlhf"
    "github.com/AINative-studio/ainative-code/internal/config"
    rlhfClient "github.com/AINative-studio/ainative-code/internal/client/rlhf"
)

// Load configuration
cfg := config.LoadRLHFConfig()

// Create RLHF API client
client := rlhfClient.New(
    rlhfClient.WithBaseURL(cfg.Endpoint),
)

// Create collector
collector := rlhf.NewCollector(cfg, client)

// Start background worker
if err := collector.Start(); err != nil {
    log.Fatal(err)
}
defer collector.Stop()
```

### Capturing Interactions

```go
// Capture a user-AI interaction
interactionID := collector.CaptureInteraction(
    "What is the capital of France?",  // prompt
    "The capital of France is Paris.", // response
    "claude-3-5-sonnet-20241022",      // model ID
)
```

### Recording Implicit Feedback

```go
// User regenerates response (negative signal)
collector.RecordImplicitFeedback(interactionID, rlhf.ActionRegenerate)

// User copies response (positive signal)
collector.RecordImplicitFeedback(interactionID, rlhf.ActionCopy)

// User edits response (negative signal)
collector.RecordImplicitFeedback(interactionID, rlhf.ActionEdit)

// User continues conversation (positive signal)
collector.RecordImplicitFeedback(interactionID, rlhf.ActionContinue)
```

### Recording Explicit Feedback

```go
// User provides rating and comment
collector.RecordExplicitFeedback(
    interactionID,
    0.95,                    // score (0.0-1.0)
    "Very helpful response!" // comment
)
```

### Checking for Feedback Prompts

```go
// Check if user should be prompted for feedback
if collector.ShouldPromptForFeedback() {
    // Show feedback prompt UI
    model := rlhf.NewFeedbackPromptModel(interactionID)
    // ... run TUI model
}
```

## Configuration

### Required Settings

```yaml
services:
  rlhf:
    enabled: true
    endpoint: https://api.ainative.studio/v1/rlhf
    api_key: your_api_key
    auto_collect: true
```

### Optional Settings

```yaml
services:
  rlhf:
    # Privacy controls
    opt_out: false
    review_before_submit: false

    # Batch settings
    batch_size: 10
    batch_interval: 5m

    # Feedback prompt settings
    prompt_interval: 5

    # Implicit feedback scores
    implicit_feedback:
      enabled: true
      regenerate_score: 0.2
      edit_response_score: 0.3
      copy_response_score: 0.8
      continue_score: 0.7
```

## API Reference

### Collector

#### NewCollector
```go
func NewCollector(cfg *config.RLHFConfig, client *rlhf.Client) *Collector
```
Creates a new RLHF collector instance.

#### Start
```go
func (c *Collector) Start() error
```
Starts the background worker for batch processing.

#### Stop
```go
func (c *Collector) Stop() error
```
Gracefully stops the collector and flushes remaining interactions.

#### CaptureInteraction
```go
func (c *Collector) CaptureInteraction(prompt, response, modelID string) string
```
Captures a user-AI interaction. Returns the interaction ID.

#### RecordImplicitFeedback
```go
func (c *Collector) RecordImplicitFeedback(interactionID string, action FeedbackAction)
```
Records an implicit feedback signal for an interaction.

#### RecordExplicitFeedback
```go
func (c *Collector) RecordExplicitFeedback(interactionID string, score float64, feedback string)
```
Records explicit user feedback for an interaction.

#### ShouldPromptForFeedback
```go
func (c *Collector) ShouldPromptForFeedback() bool
```
Returns true if the user should be prompted for feedback.

### FeedbackAction Types

```go
const (
    ActionRegenerate FeedbackAction = "regenerate"
    ActionEdit       FeedbackAction = "edit"
    ActionCopy       FeedbackAction = "copy"
    ActionContinue   FeedbackAction = "continue"
)
```

## Data Structures

### InteractionData

```go
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
```

### ImplicitSignal

```go
type ImplicitSignal struct {
    Action    string
    Timestamp time.Time
    Score     float64
}
```

## Testing

### Running Tests

```bash
# Run all tests
go test ./internal/rlhf/...

# Run with coverage
go test -cover ./internal/rlhf/...

# Run specific test
go test -run TestCollector_CaptureInteraction ./internal/rlhf/...
```

### Mock Client

Use the provided MockRLHFClient for testing:

```go
mockClient := &MockRLHFClient{
    SubmittedBatches: make([]*rlhf.BatchInteractionFeedback, 0),
    ShouldFail:       false,
}

collector := rlhf.NewCollector(cfg, mockClient)
```

## Privacy & Security

### Data Collected

- User prompts
- AI responses
- Timestamps
- Model IDs
- Session IDs (anonymous UUIDs)
- Implicit signals (user actions)
- Explicit feedback (ratings and comments)

### Data NOT Collected

- Personal identifying information (PII)
- File paths or system information
- API keys or credentials
- IP addresses

### Privacy Features

- **Opt-out**: Set `opt_out: true` to disable all collection
- **Review before submit**: User can review data before API submission
- **Session isolation**: Each session has a unique anonymous ID
- **No disk storage**: Data only in memory until submitted

## Performance

### Benchmarks

Typical performance characteristics:

- **Capture**: < 1ms per interaction
- **Memory**: ~1KB per queued interaction
- **CPU**: < 0.1% additional usage
- **Network**: 1 API call per batch (default: 10 interactions)

### Optimization Tips

1. **Adjust batch size**: Larger batches = fewer API calls
2. **Tune interval**: Balance between latency and freshness
3. **Monitor queue size**: Use `GetQueueSize()` to check backlog
4. **Handle failures**: Retry logic prevents data loss

## Troubleshooting

### Common Issues

#### No interactions captured
- Check `auto_collect: true` in config
- Verify `opt_out: false`
- Ensure collector is started with `Start()`

#### High memory usage
- Reduce `batch_size`
- Decrease `batch_interval` for more frequent submission
- Check for API connection issues

#### Submissions failing
- Verify API endpoint is reachable
- Check API key is valid
- Review error logs

### Logging

Enable debug logging to troubleshoot:

```go
logger.SetLevel("debug")
```

Look for RLHF-related log messages:

```
[INFO] Starting RLHF auto-collector
[DEBUG] Captured interaction: id=abc123
[DEBUG] Recorded implicit feedback: action=copy score=0.8
[INFO] Processing RLHF batch: batch_size=10
[INFO] RLHF batch submitted: success_count=10
```

## Contributing

When contributing to this package:

1. **Add tests**: All new features must have unit tests
2. **Update docs**: Document new configuration options
3. **Privacy first**: Always respect user privacy settings
4. **Performance**: Profile any changes that affect capture path
5. **Backwards compatibility**: Don't break existing configs

## License

Copyright 2026 AINative Studio. All rights reserved.
