# Design Token Synchronization

This document describes the bidirectional design token synchronization feature that allows you to sync design tokens between your local files and the AINative Design system.

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Command Reference](#command-reference)
- [Sync Directions](#sync-directions)
- [Conflict Resolution](#conflict-resolution)
- [Watch Mode](#watch-mode)
- [Examples](#examples)
- [Architecture](#architecture)
- [Troubleshooting](#troubleshooting)

## Overview

The `ainative-code design sync` command provides bidirectional synchronization of design tokens with the following capabilities:

- **Pull**: Download tokens from AINative Design to local file
- **Push**: Upload tokens from local file to AINative Design
- **Bidirectional**: Sync in both directions with intelligent conflict resolution
- **Watch Mode**: Continuous sync with automatic file change detection
- **Conflict Detection**: Automatic detection and resolution of conflicting changes
- **Dry Run**: Preview changes before applying them

## Quick Start

### Basic Pull (Download from AINative Design)

```bash
ainative-code design sync --project my-project --direction pull
```

### Basic Push (Upload to AINative Design)

```bash
ainative-code design sync --project my-project --direction push
```

### Bidirectional Sync with Watch Mode

```bash
ainative-code design sync --project my-project --watch
```

## Command Reference

### Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--project` | `-p` | string | *required* | Project ID for synchronization |
| `--direction` | `-d` | string | `bidirectional` | Sync direction: `pull`, `push`, or `bidirectional` |
| `--local-path` | `-l` | string | `./design-tokens.json` | Local file path for tokens |
| `--conflict` | `-c` | string | `prompt` | Conflict resolution strategy |
| `--watch` | `-w` | bool | `false` | Enable watch mode for continuous sync |
| `--watch-interval` | | duration | `2s` | Debounce interval for watch mode |
| `--dry-run` | | bool | `false` | Perform dry run without making changes |
| `--verbose` | `-v` | bool | `false` | Enable verbose logging |

### Conflict Resolution Strategies

- **`local`**: Prefer local changes over remote
- **`remote`**: Prefer remote changes over local
- **`newest`**: Prefer the newest changes based on timestamp
- **`prompt`**: Prompt the user to resolve conflicts interactively
- **`merge`**: Attempt to merge changes automatically

## Sync Directions

### Pull (Remote → Local)

Downloads tokens from AINative Design and updates your local file.

```bash
ainative-code design sync \
  --project my-project \
  --direction pull \
  --local-path ./tokens/design-tokens.json
```

**Behavior:**
- New remote tokens are added to local file
- Modified remote tokens update local versions
- Tokens deleted remotely are removed from local file

### Push (Local → Remote)

Uploads tokens from your local file to AINative Design.

```bash
ainative-code design sync \
  --project my-project \
  --direction push \
  --local-path ./tokens/design-tokens.json
```

**Behavior:**
- New local tokens are uploaded to remote
- Modified local tokens update remote versions
- Tokens deleted locally are removed from remote

### Bidirectional (Local ↔ Remote)

Syncs in both directions, intelligently merging changes.

```bash
ainative-code design sync \
  --project my-project \
  --direction bidirectional \
  --conflict remote
```

**Behavior:**
- New tokens on either side are synced to the other
- Conflicts are resolved using the specified strategy
- Both local and remote stay in sync

## Conflict Resolution

### Interactive Prompt (Default)

When conflicts are detected, you'll be prompted to choose:

```
=== Conflict Detected ===
Token: color.primary
Type: both_modified

Local version:
  Type:  color
  Value: #007bff

Remote version:
  Type:  color
  Value: #0056b3

Choose resolution:
  1) Use local version
  2) Use remote version
  3) Skip this token

Enter choice (1-3):
```

### Automatic Resolution

Use `--conflict` flag to automatically resolve conflicts:

```bash
# Always use remote version
ainative-code design sync --project my-project --conflict remote

# Always use local version
ainative-code design sync --project my-project --conflict local

# Use newest based on timestamp
ainative-code design sync --project my-project --conflict newest

# Attempt to merge metadata
ainative-code design sync --project my-project --conflict merge
```

### Conflict Summary

After sync, a conflict summary is displayed:

```
=== Conflict Summary ===
Total conflicts: 2

1. color.primary (both_modified)
   Resolution: remote - Preferred remote version as per strategy

2. spacing.large (type_change)
   Resolution: local - Preferred local version as per strategy
```

## Watch Mode

Watch mode enables continuous synchronization by monitoring file changes.

### Basic Watch

```bash
ainative-code design sync \
  --project my-project \
  --watch \
  --local-path ./design-tokens.json
```

### Watch with Custom Interval

```bash
ainative-code design sync \
  --project my-project \
  --watch \
  --watch-interval 5s
```

### Watch Mode Features

- **Debouncing**: Multiple rapid changes trigger only one sync
- **Automatic Retry**: Failed syncs are automatically retried (3 attempts by default)
- **Initial Sync**: Performs sync on startup (configurable)
- **Graceful Shutdown**: Handles Ctrl+C cleanly

### Watch Mode Output

```
Starting watch mode...
Watching: /Users/you/project/design-tokens.json
Press Ctrl+C to stop

[14:30:15] File changed, queuing sync...
[14:30:17] Debounce period elapsed, triggering sync
[14:30:18] Sync completed: 2 updated, 0 conflicts

Received interrupt signal, stopping watcher...
Watch mode stopped
```

## Examples

### Example 1: First-Time Setup

Pull all tokens from AINative Design to start local development:

```bash
ainative-code design sync \
  --project ainative-ui \
  --direction pull \
  --local-path ./src/styles/tokens.json
```

### Example 2: Local Development Workflow

Enable watch mode for continuous sync during development:

```bash
ainative-code design sync \
  --project ainative-ui \
  --watch \
  --conflict local \
  --verbose
```

### Example 3: Preview Changes Before Syncing

Use dry run to see what would change:

```bash
ainative-code design sync \
  --project ainative-ui \
  --direction push \
  --dry-run
```

Output:
```
=== Sync Results ===
Duration: 234ms
Added:    5
Updated:  3
Deleted:  1
Conflicts: 2

Dry run completed - no changes were made
```

### Example 4: Team Sync

Push local changes to share with team:

```bash
ainative-code design sync \
  --project ainative-ui \
  --direction push \
  --conflict prompt
```

### Example 5: Production Deployment

Sync with automatic conflict resolution for CI/CD:

```bash
ainative-code design sync \
  --project ainative-ui \
  --direction pull \
  --conflict remote \
  --local-path ./dist/tokens.json
```

## Architecture

### Components

1. **Sync Engine** (`internal/design/sync.go`)
   - Orchestrates synchronization operations
   - Implements pull, push, and bidirectional sync
   - Manages conflict detection and resolution

2. **Design Client Adapter** (`internal/client/design/sync_adapter.go`)
   - Bridges the HTTP API client to the sync engine
   - Handles pagination for large token sets
   - Manages token transformations

3. **Conflict Resolver** (`internal/design/conflicts.go`)
   - Detects conflicts between local and remote tokens
   - Implements multiple resolution strategies
   - Provides interactive conflict resolution

4. **File Watcher** (`internal/design/watcher.go`)
   - Monitors local file system for changes
   - Implements debouncing to prevent rapid re-syncs
   - Provides retry logic for failed syncs

### Data Flow

```
┌─────────────┐         ┌──────────────┐         ┌─────────────┐
│   Local     │◄────────┤     Sync     ├────────►│  AINative   │
│   File      │  Pull   │    Engine    │  Push   │   Design    │
│             │         │              │         │   System    │
└─────────────┘         └──────────────┘         └─────────────┘
       ▲                       │
       │                       │
       │                       ▼
 ┌─────┴──────┐         ┌──────────────┐
 │   File     │         │   Conflict   │
 │  Watcher   │         │   Resolver   │
 └────────────┘         └──────────────┘
```

### Token Format

Local tokens are stored in JSON format:

```json
{
  "tokens": [
    {
      "name": "color.primary",
      "type": "color",
      "value": "#007bff",
      "description": "Primary brand color",
      "category": "colors",
      "metadata": {
        "source": "brand-guidelines",
        "updated_by": "design-team"
      }
    }
  ],
  "metadata": {
    "synced_at": "2026-01-04T14:30:00Z",
    "project": "ainative-ui"
  }
}
```

## Troubleshooting

### Issue: Sync Fails with Authentication Error

**Solution:** Ensure your API credentials are configured:

```bash
ainative-code config set api.key YOUR_API_KEY
ainative-code config set api.url https://api.ainative.com
```

### Issue: Conflicts Not Being Detected

**Symptom:** Changes are overwritten without warning

**Solution:** Use bidirectional sync or specify explicit conflict strategy:

```bash
ainative-code design sync --project my-project --conflict prompt
```

### Issue: Watch Mode Not Detecting Changes

**Possible causes:**
1. File is in a subdirectory not being watched
2. Changes are happening too quickly (debounce period)

**Solution:**

```bash
# Ensure parent directory is being watched
ainative-code design sync --watch --local-path /absolute/path/to/tokens.json

# Reduce debounce interval
ainative-code design sync --watch --watch-interval 1s
```

### Issue: Too Many Conflicts

**Symptom:** Every sync shows many conflicts

**Solution:** Perform a one-time sync to baseline:

```bash
# Reset to remote state
ainative-code design sync --project my-project --direction pull --conflict remote

# Or reset to local state
ainative-code design sync --project my-project --direction push --conflict local
```

### Issue: Slow Sync Performance

**Symptom:** Sync takes a long time with many tokens

**Solution:** The sync engine uses batching (100 tokens per batch). For very large token sets:

1. Use `--verbose` to see batch progress
2. Consider splitting tokens into multiple projects
3. Use `--direction pull` or `--direction push` instead of bidirectional

### Issue: File Permission Errors

**Symptom:** Cannot write to local token file

**Solution:**

```bash
# Check file permissions
ls -la design-tokens.json

# Fix permissions if needed
chmod 644 design-tokens.json

# Ensure directory exists
mkdir -p $(dirname /path/to/tokens.json)
```

## Best Practices

### 1. Use Version Control

Always commit your token files to version control:

```bash
git add design-tokens.json
git commit -m "Update design tokens"
```

### 2. Establish Sync Direction

For teams, establish clear ownership:

- **Design team owns remote**: Use `pull` to get updates
- **Dev team owns local**: Use `push` to publish changes
- **Collaborative editing**: Use `bidirectional` with `prompt`

### 3. Use Dry Run Before Major Syncs

Preview changes before applying:

```bash
ainative-code design sync --project my-project --dry-run
```

### 4. Watch Mode for Development

Enable watch mode during active development:

```bash
ainative-code design sync --project my-project --watch --conflict local
```

### 5. CI/CD Integration

For automated deployments:

```bash
# In your CI/CD pipeline
ainative-code design sync \
  --project $PROJECT_ID \
  --direction pull \
  --conflict remote \
  --local-path ./build/tokens.json
```

### 6. Backup Before Sync

Create backups before major sync operations:

```bash
cp design-tokens.json design-tokens.backup.json
ainative-code design sync --project my-project
```

## Advanced Usage

### Custom Sync Workflow

Combine with other commands for advanced workflows:

```bash
# Validate tokens before pushing
ainative-code design validate ./tokens.json
if [ $? -eq 0 ]; then
  ainative-code design sync --project my-project --direction push
fi
```

### Programmatic Integration

The sync functionality can be used programmatically:

```go
package main

import (
    "context"
    "github.com/AINative-studio/ainative-code/internal/design"
    designclient "github.com/AINative-studio/ainative-code/internal/client/design"
)

func main() {
    // Create client and adapter
    client := designclient.New(...)
    adapter := designclient.NewSyncAdapter(client, "my-project")

    // Create syncer
    syncer := design.NewSyncer(adapter, design.SyncConfig{
        ProjectID: "my-project",
        Direction: design.SyncDirectionPull,
        LocalPath: "./tokens.json",
        ConflictResolution: design.ConflictResolutionRemote,
    })

    // Perform sync
    result, err := syncer.Sync(context.Background())
    if err != nil {
        panic(err)
    }

    // Process results
    println("Added:", result.Added)
    println("Updated:", result.Updated)
}
```

## Related Documentation

- [Design Token Generation](./design-generate.md)
- [Design Token Upload](./design-upload.md)
- [API Client Configuration](./api-client.md)
- [CLI Commands Reference](./cli-reference.md)

## Support

For issues or questions:

- GitHub Issues: https://github.com/AINative-studio/ainative-code/issues
- Documentation: https://docs.ainative.com
- Community: https://community.ainative.com
