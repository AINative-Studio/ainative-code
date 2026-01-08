# Session Management Guide

This guide covers how to manage chat sessions, including creating, resuming, exporting, and organizing conversations.

## Table of Contents

1. [Overview](#overview)
2. [Understanding Sessions](#understanding-sessions)
3. [Creating Sessions](#creating-sessions)
4. [Listing Sessions](#listing-sessions)
5. [Viewing Session Details](#viewing-session-details)
6. [Resuming Sessions](#resuming-sessions)
7. [Exporting Sessions](#exporting-sessions)
8. [Importing Sessions](#importing-sessions)
9. [Deleting Sessions](#deleting-sessions)
10. [Session Metadata](#session-metadata)
11. [Best Practices](#best-practices)
12. [Advanced Usage](#advanced-usage)

## Overview

Sessions in AINative Code allow you to:

- Maintain conversation context across multiple interactions
- Organize different projects or topics
- Resume work from where you left off
- Share conversations with team members
- Keep a history of problem-solving approaches
- Archive important conversations

### Session Storage

Sessions are stored in:
- **Local SQLite database**: `~/.config/ainative-code/sessions.db`
- **ZeroDB** (if enabled): For cloud backup and sync

## Understanding Sessions

### What is a Session?

A session is a conversation thread that includes:

- All messages exchanged with the AI
- Conversation context and history
- Metadata (title, created time, last updated)
- Provider and model information
- Session ID for reference

### Session Lifecycle

```
Create → Active → Pause → Resume → Archive/Delete
```

### Session Identifiers

Each session has a unique ID:
```
abc123def456  (12-character alphanumeric)
```

## Creating Sessions

### Manual Session Creation

You can manually create a session with custom settings using the `session create` command:

```bash
# Create a basic session
ainative-code session create --title "Bug Investigation"

# Create with tags
ainative-code session create --title "API Development" --tags "golang,api,rest"

# Create with specific provider and model
ainative-code session create --title "Code Review" --provider anthropic --model claude-3-5-sonnet-20241022

# Create with custom metadata
ainative-code session create --title "Project Planning" --metadata '{"project":"myapp","priority":"high"}'

# Create without activating it
ainative-code session create --title "Draft Session" --no-activate
```

**Available Flags:**
- `--title` (required): Session title
- `--tags`: Comma-separated list of tags
- `--provider`: AI provider name (anthropic, openai, azure, bedrock, gemini, ollama, meta)
- `--model`: Specific model name
- `--metadata`: JSON string with custom metadata
- `--no-activate`: Don't activate the session after creation (default: activates)

**Returns:**
- Session ID
- Session details (title, tags, model, status)
- Activation message (if not using `--no-activate`)

### Automatic Session Creation

A new session is automatically created when you start a chat:

```bash
ainative-code chat
```

### Naming Sessions

Sessions are automatically titled based on the first message, but you can specify a custom title:

```bash
ainative-code chat --title "REST API Development"
```

### Session with Initial Message

```bash
ainative-code chat --title "OAuth Implementation" "How do I implement OAuth 2.0 in Go?"
```

## Listing Sessions

### List Recent Sessions

```bash
# List 10 most recent sessions (default)
ainative-code session list

# Output:
# ID           TITLE                     CREATED              MESSAGES
# abc123def456 REST API Development      2024-01-15 10:30     12
# def789ghi012 OAuth Implementation      2024-01-14 15:45     8
# ...
```

### List All Sessions

```bash
ainative-code session list --all
```

### Limit Number of Sessions

```bash
# Show 20 sessions
ainative-code session list --limit 20

# Show 5 sessions
ainative-code session list -n 5
```

### Filter Sessions

```bash
# Search by title
ainative-code session list --search "OAuth"

# Filter by date
ainative-code session list --since "2024-01-01"
ainative-code session list --until "2024-01-31"

# Filter by provider
ainative-code session list --provider anthropic
```

### Sort Sessions

```bash
# Sort by date (default: newest first)
ainative-code session list --sort date

# Sort by title
ainative-code session list --sort title

# Sort by message count
ainative-code session list --sort messages

# Reverse sort order
ainative-code session list --reverse
```

## Viewing Session Details

### Show Session Content

```bash
# View full session
ainative-code session show abc123def456
```

Output:
```
Session: REST API Development
ID: abc123def456
Created: 2024-01-15 10:30:00
Last Updated: 2024-01-15 14:22:00
Provider: anthropic
Model: claude-3-5-sonnet-20241022
Messages: 12

─────────────────────────────────────────

[User - 10:30:05]
How do I create a REST API in Go?

[Assistant - 10:30:15]
I'll help you create a REST API in Go. Here's a comprehensive guide:
[... full response ...]

[User - 10:35:20]
How do I add authentication?

[Assistant - 10:35:35]
For authentication, you can use JWT tokens...
[... full response ...]

─────────────────────────────────────────
```

### Show Summary Only

```bash
# Show metadata without messages
ainative-code session show abc123def456 --summary
```

### Show Specific Messages

```bash
# Show first 5 messages
ainative-code session show abc123def456 --limit 5

# Show messages 10-20
ainative-code session show abc123def456 --skip 10 --limit 10
```

### Export to Different Formats

```bash
# View as JSON
ainative-code session show abc123def456 --format json

# View as Markdown
ainative-code session show abc123def456 --format markdown

# View as plain text
ainative-code session show abc123def456 --format text
```

## Resuming Sessions

### Resume a Session

```bash
# Resume by session ID
ainative-code chat --session-id abc123def456

# Or use shorthand
ainative-code chat -s abc123def456
```

The AI will have full context of previous messages.

### Resume Last Session

```bash
# Automatically resume the most recent session
ainative-code chat --resume

# Or use shorthand
ainative-code chat -r
```

### Resume from Session List

```bash
# List sessions and select one
ainative-code session list

# Resume the selected session
ainative-code chat -s <session-id>
```

## Exporting Sessions

### Export to JSON

```bash
# Export to default location (session-<id>.json)
ainative-code session export abc123def456

# Export to specific file
ainative-code session export abc123def456 -o my-session.json

# Export to specific directory
ainative-code session export abc123def456 -o ~/exports/rest-api.json
```

### Export Format

Exported JSON structure:

```json
{
  "id": "abc123def456",
  "title": "REST API Development",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T14:22:00Z",
  "provider": "anthropic",
  "model": "claude-3-5-sonnet-20241022",
  "metadata": {
    "tags": ["go", "rest-api"],
    "project": "backend-service"
  },
  "messages": [
    {
      "role": "user",
      "content": "How do I create a REST API in Go?",
      "timestamp": "2024-01-15T10:30:05Z"
    },
    {
      "role": "assistant",
      "content": "I'll help you create a REST API in Go...",
      "timestamp": "2024-01-15T10:30:15Z"
    }
  ]
}
```

### Export Multiple Sessions

```bash
# Export all sessions
ainative-code session export --all -o sessions-backup.json

# Export sessions from specific date range
ainative-code session export --since "2024-01-01" -o january-sessions.json
```

### Export to Markdown

```bash
# Export as Markdown documentation
ainative-code session export abc123def456 --format markdown -o session.md
```

## Importing Sessions

### Import from JSON

```bash
# Import a session
ainative-code session import -i session.json

# Import and assign new ID
ainative-code session import -i session.json --new-id

# Import with custom title
ainative-code session import -i session.json --title "Imported Session"
```

### Import Multiple Sessions

```bash
# Import from backup file
ainative-code session import -i sessions-backup.json

# Import from directory (all .json files)
ainative-code session import --dir ~/session-backups/
```

### Import Validation

The import command validates:
- JSON structure
- Required fields
- Message format
- Metadata consistency

## Deleting Sessions

### Delete a Single Session

```bash
# Delete by ID
ainative-code session delete abc123def456

# With confirmation prompt
ainative-code session delete abc123def456 --confirm
```

### Delete Multiple Sessions

```bash
# Delete specific sessions
ainative-code session delete abc123 def456 ghi789

# Delete sessions matching criteria
ainative-code session delete --before "2024-01-01"
ainative-code session delete --older-than 30d
```

### Delete All Sessions

```bash
# Delete all sessions (with confirmation)
ainative-code session delete --all --confirm

# Force delete without confirmation (dangerous!)
ainative-code session delete --all --force
```

## Session Metadata

### Adding Metadata

```bash
# Add tags to a session
ainative-code session tag abc123def456 go rest-api authentication

# Add project association
ainative-code session meta abc123def456 project "backend-service"

# Add custom metadata
ainative-code session meta abc123def456 priority high
ainative-code session meta abc123def456 status active
```

### Updating Session Title

```bash
ainative-code session rename abc123def456 "OAuth 2.0 Implementation Guide"
```

### Viewing Metadata

```bash
# Show metadata
ainative-code session meta abc123def456

# Output:
# Title: REST API Development
# Tags: go, rest-api, authentication
# Project: backend-service
# Priority: high
# Status: active
```

### Searching by Metadata

```bash
# Find sessions by tag
ainative-code session list --tag go

# Find sessions by project
ainative-code session list --project backend-service

# Combine filters
ainative-code session list --tag go --project backend-service
```

## Best Practices

### 1. Use Descriptive Titles

```bash
# Good
ainative-code chat --title "Implementing JWT Authentication in Go API"

# Less helpful
ainative-code chat --title "Auth stuff"
```

### 2. Organize with Tags

```bash
# Tag sessions for easy filtering
ainative-code session tag abc123 go rest-api jwt authentication
ainative-code session tag def456 go websocket real-time
```

### 3. Regular Exports

Create backups of important sessions:

```bash
# Weekly backup
ainative-code session export --all -o ~/backups/sessions-$(date +%Y%m%d).json
```

### 4. Clean Up Old Sessions

```bash
# Delete sessions older than 90 days
ainative-code session delete --older-than 90d --confirm
```

### 5. Session per Topic

Keep related conversations in one session:

```bash
# One session for entire feature
ainative-code chat --title "User Authentication Feature" -r

# Instead of multiple fragmented sessions
```

### 6. Use Metadata for Project Management

```bash
# Associate with projects
ainative-code session meta abc123 project "ecommerce-api"
ainative-code session meta abc123 sprint "sprint-12"
ainative-code session meta abc123 story "USER-123"
```

### 7. Export Before Major Changes

```bash
# Save a copy before continuing sensitive work
ainative-code session export abc123 -o backup-before-refactor.json
```

### 8. Archive Completed Work

```bash
# Mark as archived (custom metadata)
ainative-code session meta abc123 archived true
ainative-code session meta abc123 completed_date "2024-01-20"
```

## Advanced Usage

### Session Templates

Create template sessions for common workflows:

```bash
# Create template
ainative-code chat --title "API Development Template"
# Add initial prompts and structure
# Export as template
ainative-code session export <id> -o templates/api-dev.json

# Use template
ainative-code session import -i templates/api-dev.json --new-id
```

### Session Sharing

Share sessions with team members:

```bash
# Export session
ainative-code session export abc123 -o shared-session.json

# Share file with team
# They can import:
ainative-code session import -i shared-session.json
```

### Session Branching

Create a new session from an existing one:

```bash
# Export current state
ainative-code session export abc123 -o base-session.json

# Import as new session
ainative-code session import -i base-session.json --new-id --title "Branched Session"
```

### Batch Operations

```bash
# Tag multiple sessions
for id in abc123 def456 ghi789; do
  ainative-code session tag $id project-x
done

# Export multiple sessions
ainative-code session list --project project-x --format json | \
  jq -r '.[] | .id' | \
  xargs -I {} ainative-code session export {} -o {}.json
```

### Session Analytics

```bash
# Count sessions
ainative-code session list --all --format json | jq length

# Sessions per provider
ainative-code session list --all --format json | \
  jq 'group_by(.provider) | map({provider: .[0].provider, count: length})'

# Average messages per session
ainative-code session list --all --format json | \
  jq '[.[] | .message_count] | add / length'
```

### ZeroDB Sync

If ZeroDB is enabled, sessions are automatically synced:

```bash
# Force sync to ZeroDB
ainative-code session sync --push

# Pull sessions from ZeroDB
ainative-code session sync --pull

# Two-way sync
ainative-code session sync --bidirectional
```

### Session Search

Advanced search capabilities:

```bash
# Full-text search in messages
ainative-code session search "JWT authentication"

# Search with filters
ainative-code session search "OAuth" --provider anthropic --since 2024-01-01

# Export search results
ainative-code session search "Go REST API" --export results.json
```

### Session Statistics

```bash
# Show session statistics
ainative-code session stats abc123

# Output:
# Session Statistics for abc123
# ─────────────────────────────
# Total Messages: 24
# User Messages: 12
# Assistant Messages: 12
# Total Tokens: 15,234
# Average Response Time: 3.2s
# Duration: 2h 15m
# First Message: 2024-01-15 10:30:00
# Last Message: 2024-01-15 12:45:00
```

## Troubleshooting

### Session Not Found

```bash
# Verify session exists
ainative-code session list | grep abc123

# Check full ID
ainative-code session list --all --format json | jq '.[] | select(.id | contains("abc"))'
```

### Export Fails

```bash
# Check permissions
ls -la ~/exports/

# Create directory if needed
mkdir -p ~/exports/

# Try with full path
ainative-code session export abc123 -o $(pwd)/session.json
```

### Import Validation Errors

```bash
# Validate JSON
cat session.json | jq .

# Check required fields
cat session.json | jq '{id, title, messages}'
```

### Sync Issues

```bash
# Check ZeroDB connection
ainative-code zerodb ping

# Force re-sync
ainative-code session sync --force --pull
```

## Integration with AINative Platform

### Cloud Backup

Sessions are automatically backed up to ZeroDB when enabled:

```yaml
services:
  zerodb:
    enabled: true
    endpoint: postgresql://zerodb.ainative.studio:5432
```

### Team Collaboration

Share sessions through AINative platform:

```bash
# Share session with team
ainative-code session share abc123 --team engineering

# List shared sessions
ainative-code session list --shared

# Clone shared session
ainative-code session clone shared-abc123
```

### Cross-Device Sync

Sessions sync across devices with AINative authentication:

```bash
# Login
ainative-code auth login

# Sessions automatically sync
# Work on device A
ainative-code chat -s abc123

# Continue on device B
ainative-code chat -s abc123  # Synced automatically
```

## Next Steps

- [Getting Started Guide](getting-started.md) - Basic usage
- [Tools Guide](tools.md) - Available tools and capabilities
- [AINative Integrations](ainative-integrations.md) - Platform features
- [Troubleshooting Guide](troubleshooting.md) - Common issues
