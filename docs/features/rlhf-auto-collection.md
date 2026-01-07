# RLHF Auto-Collection

## Overview

The RLHF (Reinforcement Learning from Human Feedback) Auto-Collection feature automatically captures user interactions with AI responses to improve model performance over time. This feature collects both explicit user feedback (ratings and comments) and implicit feedback signals (user actions) in the background.

## Features

### Automatic Interaction Capture
- Captures all user prompts and AI responses during chat sessions
- Records metadata including model ID, session ID, and timestamps
- Stores interactions temporarily in memory before batch submission
- Minimal performance impact on user experience

### Implicit Feedback Signals
The system automatically tracks user actions that indicate response quality:

| Action | Signal Type | Default Score | Meaning |
|--------|-------------|---------------|---------|
| Regenerate | Negative | 0.2 | User requests a new response (poor quality) |
| Edit Response | Negative | 0.3 | User modifies the response (needs improvement) |
| Copy Response | Positive | 0.8 | User copies response (useful content) |
| Continue Conversation | Positive | 0.7 | User continues the conversation (satisfactory) |

### Explicit Feedback Prompts
- Non-intrusive prompts appear after N interactions (configurable)
- Users can rate responses on a 1-5 scale
- Optional free-text feedback comments
- Can be easily dismissed without interrupting workflow

### Background Submission
- Batches interactions for efficient API submission
- Configurable batch size and submission interval
- Automatic retry logic with exponential backoff
- Network failure handling with local queue persistence
- Graceful shutdown ensures no data loss

### Privacy Controls
- **Opt-in by default**: Auto-collection disabled unless explicitly enabled
- **Opt-out mechanism**: Users can disable collection at any time
- **Review before submit**: Optional setting to review data before submission
- **Clear documentation**: Transparent about what data is collected
- **Session isolation**: Each session has a unique ID for privacy

## Configuration

### Basic Configuration

Add to your configuration file (`~/.config/ainative-code/config.yaml`):

```yaml
services:
  rlhf:
    enabled: true
    endpoint: https://api.ainative.studio/v1/rlhf
    api_key: your_api_key_here

    # Auto-collection settings
    auto_collect: true              # Enable automatic collection
    opt_out: false                  # Set to true to completely disable
    review_before_submit: false     # Prompt before submitting batches
    batch_size: 10                  # Submit after 10 interactions
    batch_interval: 5m              # Or submit every 5 minutes
    prompt_interval: 5              # Prompt for feedback every 5 interactions

    # Implicit feedback configuration
    implicit_feedback:
      enabled: true
      regenerate_score: 0.2         # Negative signal
      edit_response_score: 0.3      # Negative signal
      copy_response_score: 0.8      # Positive signal
      continue_score: 0.7           # Positive signal
```

### Environment Variables

Override configuration with environment variables:

```bash
export AINATIVE_RLHF_AUTO_COLLECT=true
export AINATIVE_RLHF_OPT_OUT=false
export AINATIVE_RLHF_BATCH_SIZE=20
export AINATIVE_RLHF_BATCH_INTERVAL=10m
export AINATIVE_RLHF_PROMPT_INTERVAL=10
```

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `auto_collect` | boolean | `false` | Enable automatic interaction collection |
| `opt_out` | boolean | `false` | Completely disable RLHF features |
| `review_before_submit` | boolean | `false` | Review data before API submission |
| `batch_size` | integer | `10` | Number of interactions per batch |
| `batch_interval` | duration | `5m` | Maximum time between submissions |
| `prompt_interval` | integer | `5` | Prompt for feedback after N interactions |

## Usage

### Enabling Auto-Collection

1. **Edit configuration file**:
   ```bash
   vi ~/.config/ainative-code/config.yaml
   ```

2. **Set auto_collect to true**:
   ```yaml
   services:
     rlhf:
       auto_collect: true
   ```

3. **Restart the application** or reload configuration

### Providing Explicit Feedback

When prompted for feedback:

1. **Rating**: Enter a number between 1-5
   - 1: Very poor response
   - 2: Poor response
   - 3: Adequate response
   - 4: Good response
   - 5: Excellent response

2. **Comment** (optional): Provide specific feedback about the response

3. **Submit**: Press Enter to submit

4. **Skip**: Press Esc to dismiss the prompt

### Opting Out

To completely disable RLHF data collection:

**Option 1: Configuration file**
```yaml
services:
  rlhf:
    opt_out: true
```

**Option 2: Environment variable**
```bash
export AINATIVE_RLHF_OPT_OUT=true
```

**Option 3: Disable auto-collection**
```yaml
services:
  rlhf:
    auto_collect: false
```

## Architecture

### Components

```
┌─────────────────────────────────────────────────────────┐
│                     TUI Application                      │
├─────────────────────────────────────────────────────────┤
│  1. User sends prompt                                    │
│  2. AI generates response                                │
│  3. Collector captures interaction                       │
│  4. Track implicit feedback (copy, regenerate, etc.)    │
│  5. Periodically prompt for explicit feedback            │
└───────────────────┬─────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────────────┐
│                  RLHF Collector                          │
├─────────────────────────────────────────────────────────┤
│  - In-memory queue of interactions                       │
│  - Implicit feedback scoring                             │
│  - Explicit feedback storage                             │
│  - Background worker for batch processing                │
└───────────────────┬─────────────────────────────────────┘
                    │
                    │ Batch submission every N interactions
                    │ or M minutes (whichever comes first)
                    ▼
┌─────────────────────────────────────────────────────────┐
│                    RLHF API                              │
├─────────────────────────────────────────────────────────┤
│  - Receives batch interaction feedback                   │
│  - Stores for model training                             │
│  - Provides analytics                                    │
└─────────────────────────────────────────────────────────┘
```

### Data Flow

1. **Capture**: User interacts with AI, interaction captured
2. **Score**: Implicit signals automatically scored
3. **Queue**: Interaction added to in-memory queue
4. **Batch**: When batch size reached or interval elapsed
5. **Submit**: Background worker submits to API
6. **Retry**: Failed submissions re-queued with backoff
7. **Flush**: On shutdown, remaining interactions submitted

### Scoring Algorithm

The collector uses a weighted average for implicit feedback:

```
implicit_score = Σ(signal_score × weight) / Σ(weight)

where weight = (signal_index + 1) / total_signals
```

This gives more weight to recent signals, assuming user sentiment may change over time.

**Final Score Selection**:
- If explicit feedback exists: Use explicit score
- Otherwise: Use implicit score
- Default (no signals): 0.5 (neutral)

## Privacy & Data Handling

### What Data is Collected

The following data is captured when auto-collection is enabled:

- **User Prompt**: The question or input provided by the user
- **AI Response**: The response generated by the AI model
- **Timestamp**: When the interaction occurred
- **Model ID**: Which AI model was used
- **Session ID**: Unique identifier for the chat session
- **Implicit Signals**: User actions (regenerate, copy, edit, continue)
- **Explicit Feedback**: User ratings and comments (if provided)

### What Data is NOT Collected

- Personal identifying information (PII)
- File paths or system information
- API keys or credentials
- User's IP address or location
- Any data when `opt_out: true`

### Data Retention

- **In-memory**: Interactions stored temporarily until submission
- **On shutdown**: Queue flushed to API before exit
- **Failed submissions**: Re-queued with exponential backoff
- **No persistent storage**: Data not written to disk

### Security

- All API communication uses HTTPS/TLS encryption
- API keys stored securely in configuration
- Session IDs are UUIDs (anonymous)
- No sensitive data logged to console or files

## Performance Impact

The auto-collection system is designed for minimal performance impact:

- **Capture overhead**: < 1ms per interaction
- **Memory usage**: ~1KB per interaction in queue
- **Background processing**: Separate goroutine, non-blocking
- **Network**: Batched submissions reduce API calls
- **CPU**: Negligible (scoring calculations are simple)

**Typical Resource Usage**:
- Memory: ~10KB for 10 queued interactions
- Network: 1 API call per batch (default: every 10 interactions)
- CPU: < 0.1% additional usage

## Troubleshooting

### Auto-collection not working

**Check configuration**:
```bash
ainative-code config show | grep rlhf
```

**Verify settings**:
- `auto_collect: true`
- `opt_out: false`
- `enabled: true`

**Check logs**:
```bash
tail -f ~/.local/share/ainative-code/logs/app.log | grep RLHF
```

### Feedback prompts not appearing

**Check prompt interval**:
- Ensure `prompt_interval > 0`
- Prompts appear after N interactions
- Try setting to a lower value (e.g., `prompt_interval: 2`)

**Verify counter**:
- Restart application to reset counter
- Each session tracks interactions independently

### Submissions failing

**Check API connectivity**:
```bash
curl -I https://api.ainative.studio/v1/rlhf
```

**Verify API key**:
- Check `api_key` in configuration
- Ensure key has RLHF permissions

**Review error logs**:
```bash
grep "Failed to submit" ~/.local/share/ainative-code/logs/app.log
```

### High memory usage

**Reduce batch size**:
```yaml
services:
  rlhf:
    batch_size: 5  # Reduce from default 10
```

**Decrease interval**:
```yaml
services:
  rlhf:
    batch_interval: 2m  # Submit more frequently
```

## Best Practices

### For Users

1. **Enable auto-collection** to help improve the AI
2. **Provide explicit feedback** when prompted (be honest!)
3. **Use implicit signals naturally** (copy good responses, regenerate bad ones)
4. **Review privacy settings** to ensure you're comfortable
5. **Opt out if needed** - it's completely optional

### For Developers

1. **Keep batch sizes reasonable** (10-50 interactions)
2. **Set appropriate intervals** (5-10 minutes)
3. **Monitor queue size** in production
4. **Handle API failures gracefully** (auto-retry implemented)
5. **Respect user privacy** (honor opt-out settings)
6. **Log submission statistics** for monitoring

### For Administrators

1. **Set organization defaults** in shared config
2. **Monitor API usage** and costs
3. **Review analytics** to track improvement
4. **Communicate privacy policy** to users
5. **Provide opt-out instructions** clearly

## Analytics & Reporting

Use the RLHF analytics commands to view collected data:

```bash
# View statistics
ainative-code rlhf stats

# Export data for analysis
ainative-code rlhf export --output feedback.jsonl

# View recent feedback
ainative-code rlhf list --limit 20

# Get analytics for date range
ainative-code rlhf analytics \
  --from 2026-01-01 \
  --to 2026-01-31 \
  --format json
```

## FAQ

**Q: Is my data sent to Anthropic or OpenAI?**
A: No, data is sent only to the AINative RLHF API endpoint configured in your settings.

**Q: Can I review data before it's submitted?**
A: Yes, set `review_before_submit: true` to prompt before each batch submission.

**Q: How often should I provide explicit feedback?**
A: We recommend every 5-10 interactions, but it's entirely up to you.

**Q: What happens if I lose internet connection?**
A: Interactions are queued and submitted when connection is restored.

**Q: Does this slow down the AI responses?**
A: No, collection happens asynchronously and doesn't impact response time.

**Q: Can I delete my submitted feedback?**
A: Yes, use `ainative-code rlhf delete <feedback-id>` to remove entries.

**Q: Is this feature required?**
A: No, it's completely optional and disabled by default.

## Related Documentation

- [RLHF Commands](../commands/rlhf.md)
- [Configuration Guide](../configuration.md)
- [Privacy Policy](../privacy.md)
- [API Documentation](../api/rlhf.md)

## Support

If you encounter issues or have questions:

1. Check the [Troubleshooting](#troubleshooting) section
2. Review [GitHub Issues](https://github.com/AINative-studio/ainative-code/issues)
3. Contact support at support@ainative.studio
4. Join our [Discord community](https://discord.gg/ainative)
