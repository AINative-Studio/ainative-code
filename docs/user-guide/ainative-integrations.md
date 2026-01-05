# AINative Platform Integrations

This guide covers the AINative platform integrations including ZeroDB, Design Token Sync, Strapi CMS, and RLHF feedback systems.

## Table of Contents

1. [Overview](#overview)
2. [ZeroDB Integration](#zerodb-integration)
3. [Design Token Sync](#design-token-sync)
4. [Strapi CMS Integration](#strapi-cms-integration)
5. [RLHF Feedback](#rlhf-feedback)
6. [Analytics Integration](#analytics-integration)
7. [Authentication](#authentication)

## Overview

AINative Code integrates with the AINative platform ecosystem to provide:

- **ZeroDB**: Vector and NoSQL database for AI-native applications
- **Design Tokens**: Extract and sync design systems
- **Strapi CMS**: Content management integration
- **RLHF**: Reinforcement Learning from Human Feedback
- **Analytics**: Usage tracking and insights

### Prerequisites

Most platform integrations require authentication:

```bash
# Login to AINative platform
ainative-code auth login
```

## ZeroDB Integration

ZeroDB provides vector search, NoSQL tables, and PostgreSQL capabilities optimized for AI workflows.

### Features

- Vector embeddings storage and semantic search
- NoSQL tables for flexible data storage
- Hybrid quantum-classical search
- Agent memory persistence
- PostgreSQL compatibility

### Configuration

```yaml
services:
  zerodb:
    enabled: true
    endpoint: postgresql://zerodb.ainative.studio:5432
    database: your_database
    ssl: true
    ssl_mode: require
    max_connections: 25
    timeout: 30s
```

### Vector Operations

**Upsert Vectors:**

```bash
# Add vector embedding
ainative-code zerodb vector upsert \
  --id "doc-123" \
  --vector "[0.1, 0.2, 0.3, ...]" \
  --metadata '{"title": "API Documentation", "type": "doc"}' \
  --namespace "docs"
```

**Search Vectors:**

```bash
# Semantic search
ainative-code zerodb vector search \
  --query "How do I authenticate?" \
  --top-k 5 \
  --namespace "docs"

# Output:
# ID: doc-123, Score: 0.95
# ID: doc-456, Score: 0.87
# ID: doc-789, Score: 0.82
```

**List Vectors:**

```bash
# List all vectors
ainative-code zerodb vector list --limit 100

# Filter by namespace
ainative-code zerodb vector list --namespace "docs"
```

**Vector Statistics:**

```bash
# Get vector stats
ainative-code zerodb vector stats

# Output:
# Total Vectors: 1,234
# Namespaces: 5
# Average Dimension: 1536
# Storage Used: 45.2 MB
```

### NoSQL Table Operations

**Create Table:**

```bash
# Create a table
ainative-code zerodb table create users \
  --schema '{
    "id": "string",
    "name": "string",
    "email": "string",
    "metadata": "json"
  }'
```

**Insert Data:**

```bash
# Insert row
ainative-code zerodb table insert users \
  --data '{
    "id": "user-123",
    "name": "John Doe",
    "email": "john@example.com",
    "metadata": {"role": "admin"}
  }'
```

**Query Data:**

```bash
# Query table
ainative-code zerodb table query users \
  --filter 'metadata.role == "admin"' \
  --limit 10

# SQL-style query
ainative-code zerodb table query users \
  --sql "SELECT * FROM users WHERE email LIKE '%@example.com'"
```

**Update Data:**

```bash
# Update rows
ainative-code zerodb table update users \
  --filter 'id == "user-123"' \
  --data '{"metadata": {"role": "user"}}'
```

**List Tables:**

```bash
# List all tables
ainative-code zerodb table list

# Output:
# NAME        ROWS    COLUMNS  CREATED
# users       1,234   4        2024-01-15
# sessions    567     6        2024-01-10
# documents   2,345   8        2024-01-05
```

### Quantum Search

Hybrid quantum-classical vector search for enhanced accuracy:

```bash
# Enable quantum search
ainative-code zerodb quantum-search \
  --query "machine learning optimization" \
  --top-k 10 \
  --quantum-depth 3

# Quantum search provides better semantic understanding
# and can find results traditional vector search might miss
```

### Memory Persistence

Store conversation memory for context:

```bash
# Store memory
ainative-code zerodb memory store \
  --session-id "sess-123" \
  --content "User prefers functional programming" \
  --type "preference"

# Search memory
ainative-code zerodb memory search \
  --session-id "sess-123" \
  --query "programming style"

# Get context window
ainative-code zerodb memory context \
  --session-id "sess-123" \
  --max-items 10
```

### PostgreSQL Access

Direct PostgreSQL access for advanced use cases:

```bash
# Get connection details
ainative-code zerodb postgres connection

# Output:
# Host: zerodb.ainative.studio
# Port: 5432
# Database: your_database
# SSL: required

# Check status
ainative-code zerodb postgres status

# View usage statistics
ainative-code zerodb postgres usage

# View query logs
ainative-code zerodb postgres logs --limit 100
```

### AI Integration

Use ZeroDB in conversations:

```
User: Store this API documentation in ZeroDB for later reference
[paste documentation]

AI: I'll store this in ZeroDB with semantic indexing.
[Extracts vectors and stores with metadata]
Stored in ZeroDB with ID: doc-456

User: Find documentation about authentication

AI: [Searches ZeroDB]
I found these relevant documents:
1. API Authentication (score: 0.95)
2. OAuth 2.0 Setup (score: 0.87)
[... shows results ...]
```

## Design Token Sync

Extract and sync design tokens from design files and tools.

### Features

- Extract design tokens from Figma, Sketch, Adobe XD
- Generate code (CSS, SCSS, JavaScript, Tailwind)
- Sync with design systems
- Version control integration

### Configuration

```yaml
services:
  design:
    enabled: true
    endpoint: https://design.ainative.studio
    timeout: 30s
```

### Extract Design Tokens

**From Figma:**

```bash
# Extract from Figma file
ainative-code design extract \
  --source figma \
  --file-id "ABC123" \
  --token "${FIGMA_TOKEN}" \
  --output tokens.json
```

**From Local Design File:**

```bash
# Extract from Sketch file
ainative-code design extract \
  --source sketch \
  --file design.sketch \
  --output tokens.json
```

### Generate Code

**CSS Variables:**

```bash
# Generate CSS
ainative-code design generate \
  --input tokens.json \
  --format css \
  --output styles/tokens.css

# Generated:
# :root {
#   --color-primary: #6366F1;
#   --color-secondary: #8B5CF6;
#   --spacing-unit: 8px;
# }
```

**Tailwind Config:**

```bash
# Generate Tailwind configuration
ainative-code design generate \
  --input tokens.json \
  --format tailwind \
  --output tailwind.config.js
```

**TypeScript:**

```bash
# Generate TypeScript types
ainative-code design generate \
  --input tokens.json \
  --format typescript \
  --output src/tokens.ts
```

### Sync Design Tokens

**Upload to AINative:**

```bash
# Upload tokens to design system
ainative-code design upload \
  --input tokens.json \
  --project "my-app" \
  --version "1.2.0"
```

**Sync Bidirectionally:**

```bash
# Sync local and remote
ainative-code design sync \
  --project "my-app" \
  --input tokens.json \
  --output tokens.json
```

**Watch for Changes:**

```bash
# Auto-sync on changes
ainative-code design watch \
  --source figma \
  --file-id "ABC123" \
  --output tokens.json \
  --auto-generate css
```

### AI-Assisted Design

Use AI to work with design tokens:

```
User: Extract colors from my Figma design and generate CSS variables

AI: I'll extract the design tokens and generate CSS.
[Extracts tokens]
Found: 12 colors, 8 spacing values, 6 typography styles

Generated tokens.css:
[shows CSS]

User: Now create a dark mode variant

AI: [Generates dark mode tokens]
Created tokens-dark.css with inverted colors and adjusted contrasts
```

## Strapi CMS Integration

Integrate with Strapi headless CMS for content management.

### Features

- Fetch and create content
- Blog post management
- Content type operations
- Media library access

### Configuration

```yaml
services:
  strapi:
    enabled: true
    endpoint: https://cms.ainative.studio
    timeout: 30s
```

### Blog Management

**List Blog Posts:**

```bash
# List all posts
ainative-code strapi blog list

# Filter published posts
ainative-code strapi blog list --status published

# Search posts
ainative-code strapi blog list --search "kubernetes"
```

**Create Blog Post:**

```bash
# Create new post
ainative-code strapi blog create \
  --title "Getting Started with AINative Code" \
  --content "$(cat article.md)" \
  --author "John Doe" \
  --tags "tutorial,getting-started" \
  --status draft
```

**Update Blog Post:**

```bash
# Update existing post
ainative-code strapi blog update 123 \
  --status published \
  --content "$(cat updated-article.md)"
```

**Delete Blog Post:**

```bash
# Delete post
ainative-code strapi blog delete 123
```

### Content Management

**List Content Types:**

```bash
# List all content types
ainative-code strapi content types

# Output:
# blog-post
# page
# author
# category
```

**Fetch Content:**

```bash
# Get content by type
ainative-code strapi content get blog-post 123

# List all content of a type
ainative-code strapi content list page --limit 50
```

**Create Content:**

```bash
# Create content
ainative-code strapi content create page \
  --data '{
    "title": "About Us",
    "slug": "about",
    "content": "..."
  }'
```

### AI-Assisted Content

Create content with AI assistance:

```
User: Create a blog post about Docker best practices

AI: I'll create a comprehensive blog post.
[Generates content]

Title: Docker Best Practices for Production
Content: [generated article]

Shall I publish this to Strapi?

User: Yes, publish as draft

AI: [Creates in Strapi]
Published to Strapi as draft (ID: 456)
You can review at: https://cms.ainative.studio/blog-posts/456
```

## RLHF Feedback

Provide feedback to improve AI responses through Reinforcement Learning from Human Feedback.

### Features

- Rate AI responses
- Provide corrections
- Submit preferences
- Track feedback analytics

### Configuration

```yaml
services:
  rlhf:
    enabled: true
    endpoint: https://rlhf.ainative.studio
    model_id: "claude-3-5-sonnet"
```

### Submit Feedback

**Rate Response:**

```bash
# Rate a response (1-5 stars)
ainative-code rlhf rate \
  --interaction-id "int-123" \
  --rating 5 \
  --comment "Excellent code example with clear explanation"
```

**Submit Correction:**

```bash
# Correct an inaccurate response
ainative-code rlhf correct \
  --interaction-id "int-123" \
  --correction "The actual syntax is: go test -race ./..."
```

**Indicate Preference:**

```bash
# Choose between responses
ainative-code rlhf prefer \
  --interaction-id "int-123" \
  --choice "response-a" \
  --reason "More concise and clearer"
```

### Interactive Feedback

During chat sessions, provide feedback:

```
AI: Here's how to implement OAuth:
[provides implementation]

User: /feedback rate:5 This is perfect!

AI: Thank you for the feedback! This helps improve my responses.

User: /feedback correction The token should expire in 3600 seconds, not 7200

AI: Thank you for the correction. I've logged this for model improvement.
```

### Feedback Analytics

**View Your Feedback:**

```bash
# List your feedback submissions
ainative-code rlhf list

# Filter by rating
ainative-code rlhf list --rating 1-2  # Show low ratings

# Filter by date
ainative-code rlhf list --since 2024-01-01
```

**Feedback Statistics:**

```bash
# View feedback stats
ainative-code rlhf stats

# Output:
# Total Feedback: 145
# Average Rating: 4.2/5
# Corrections: 12
# Preferences: 23
# Most Common Tags: code-quality, explanation, examples
```

### Privacy

RLHF feedback is used to improve models while respecting privacy:

- Personally identifiable information is removed
- Code snippets are anonymized
- You can delete your feedback anytime

```bash
# Delete specific feedback
ainative-code rlhf delete int-123

# Delete all your feedback
ainative-code rlhf delete --all --confirm
```

## Analytics Integration

Track usage and gain insights into your AI-assisted development.

### Features

- Usage statistics
- Command analytics
- Session analytics
- Cost tracking

### View Analytics

**Usage Statistics:**

```bash
# View overall usage
ainative-code analytics usage

# Output:
# Total Sessions: 234
# Total Messages: 1,456
# Providers Used: anthropic, openai
# Most Active Day: 2024-01-15 (45 messages)
```

**Provider Usage:**

```bash
# Provider breakdown
ainative-code analytics providers

# Output:
# PROVIDER    MESSAGES  TOKENS     COST
# anthropic   1,200     890K       $42.50
# openai      256       180K       $8.20
# Total                 1,070K     $50.70
```

**Command Analytics:**

```bash
# Most used commands
ainative-code analytics commands

# Output:
# COMMAND         COUNT   AVG DURATION
# chat            234     45s
# session list    67      0.5s
# zerodb search   23      2.1s
```

**Session Analytics:**

```bash
# Session statistics
ainative-code analytics sessions

# Average session length
# Average messages per session
# Most common topics
```

### Cost Tracking

Monitor AI usage costs:

```bash
# Current month costs
ainative-code analytics cost --month current

# Date range
ainative-code analytics cost --from 2024-01-01 --to 2024-01-31

# Export cost report
ainative-code analytics cost --export costs-january.csv
```

## Authentication

All platform integrations require authentication.

### Login

```bash
# OAuth login
ainative-code auth login

# Opens browser for authentication
# Stores tokens securely
```

### Check Status

```bash
# View auth status
ainative-code auth whoami

# Output:
# Email: user@example.com
# Token Status: Valid
# Expires: 2024-01-20 14:30:00
```

### Token Management

```bash
# Refresh token
ainative-code auth token refresh

# View token status
ainative-code auth token status
```

### Logout

```bash
# Logout (removes tokens)
ainative-code auth logout
```

## Best Practices

### 1. Enable Features You Need

```yaml
services:
  zerodb:
    enabled: true   # Using this
  strapi:
    enabled: false  # Not using, disable
```

### 2. Use ZeroDB for Persistence

Store important information in ZeroDB:

```
User: Save this architecture design for later

AI: [Stores in ZeroDB with semantic indexing]
Saved with ID: design-123
```

### 3. Provide Regular RLHF Feedback

Help improve the AI:
- Rate helpful responses
- Correct inaccuracies
- Indicate preferences

### 4. Sync Design Tokens Regularly

Keep design and code in sync:

```bash
# Set up auto-sync
ainative-code design watch --auto-generate css
```

### 5. Monitor Costs

Track usage to manage costs:

```bash
# Weekly cost check
ainative-code analytics cost --week current
```

## Next Steps

- [Authentication Guide](authentication.md) - Detailed auth setup
- [Configuration Guide](configuration.md) - Configure integrations
- [Tools Guide](tools.md) - Additional tools and capabilities
