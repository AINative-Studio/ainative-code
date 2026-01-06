# Strapi CMS Integration Guide

## Overview

Strapi is a leading open-source headless CMS that AINative Code integrates with for content management. This guide covers blog post management, content operations, and AI-assisted content creation workflows.

## Table of Contents

1. [Quick Start](#quick-start)
2. [Configuration](#configuration)
3. [Blog Operations](#blog-operations)
4. [Content Management](#content-management)
5. [AI-Assisted Content Creation](#ai-assisted-content-creation)
6. [Automation Workflows](#automation-workflows)
7. [Best Practices](#best-practices)
8. [Troubleshooting](#troubleshooting)

## Quick Start

### Prerequisites

```bash
# Login to AINative platform
ainative-code auth login

# Configure Strapi connection
export STRAPI_URL="https://cms.ainative.studio"
export STRAPI_API_KEY="your-api-key"
```

### Your First Blog Post

```bash
# Create a draft blog post
ainative-code strapi blog create \
  --title "Getting Started with AINative Code" \
  --content "$(cat article.md)" \
  --author "John Doe" \
  --tags "tutorial,getting-started"

# List all posts
ainative-code strapi blog list

# Publish the post
ainative-code strapi blog publish --id 42
```

## Configuration

### Setting Up Strapi Connection

**Method 1: Environment Variables**

```bash
export STRAPI_URL="https://cms.ainative.studio"
export STRAPI_API_KEY="your-api-key"
```

**Method 2: Configuration File**

```yaml
# ~/.config/ainative-code/config.yaml
services:
  strapi:
    enabled: true
    endpoint: "https://cms.ainative.studio"
    api_key: "${STRAPI_API_KEY}"
    timeout: 30s
```

**Method 3: CLI Command**

```bash
ainative-code strapi config --url "https://cms.ainative.studio"
ainative-code strapi config --token "your-api-key"
```

### Verify Configuration

```bash
# Test connection
ainative-code strapi test

# View current config
ainative-code strapi config
```

**Output:**

```
Strapi Configuration:
  URL: https://cms.ainative.studio
  API Key: sk-strapi-****...****
  Timeout: 30s
  Status: Connected ✓
```

## Blog Operations

### Creating Blog Posts

**Basic Creation:**

```bash
ainative-code strapi blog create \
  --title "My First Post" \
  --content "# Hello World\n\nThis is my first blog post!" \
  --author "John Doe"
```

**From Markdown File:**

```bash
ainative-code strapi blog create \
  --title "Technical Deep Dive" \
  --content @article.md \
  --author "Jane Smith" \
  --tags "go,rust,performance" \
  --status draft
```

**With Custom Slug:**

```bash
ainative-code strapi blog create \
  --title "Getting Started with Kubernetes" \
  --slug "k8s-getting-started" \
  --content @k8s-guide.md \
  --author "DevOps Team"
```

**Publish Immediately:**

```bash
ainative-code strapi blog create \
  --title "Breaking News" \
  --content @news.md \
  --author "Editor" \
  --status published
```

### Listing Blog Posts

**List All Posts:**

```bash
ainative-code strapi blog list
```

**Filter by Status:**

```bash
# Only published posts
ainative-code strapi blog list --status published

# Only drafts
ainative-code strapi blog list --status draft
```

**Filter by Author:**

```bash
ainative-code strapi blog list --author "John Doe"
```

**Pagination:**

```bash
# Get second page with 50 posts per page
ainative-code strapi blog list --limit 50 --page 2
```

**Combine Filters:**

```bash
ainative-code strapi blog list \
  --status published \
  --author "Jane Smith" \
  --limit 10
```

**JSON Output (for scripting):**

```bash
ainative-code strapi blog list --json | \
  jq '.data[] | {id, title: .attributes.title}'
```

### Updating Blog Posts

**Update Title:**

```bash
ainative-code strapi blog update --id 42 --title "Updated Title"
```

**Update Content from File:**

```bash
ainative-code strapi blog update --id 42 --content @revised-article.md
```

**Update Multiple Fields:**

```bash
ainative-code strapi blog update \
  --id 42 \
  --title "Revised: Getting Started Guide" \
  --author "Jane Doe" \
  --tags "tutorial,beginner,updated"
```

### Publishing Posts

**Publish a Draft:**

```bash
ainative-code strapi blog publish --id 42
```

**Unpublish (back to draft):**

```bash
ainative-code strapi blog update --id 42 --status draft
```

### Deleting Posts

```bash
ainative-code strapi blog delete --id 42
```

**Warning:** This action cannot be undone!

**Confirmation:**

```
WARNING: You are about to delete blog post ID 42. This action cannot be undone.
Type 'yes' to confirm: yes
Blog post ID 42 deleted successfully.
```

## Content Management

### Working with Content Types

**List All Content Types:**

```bash
ainative-code strapi content types
```

**Output:**

```
Content Types:
  - blog-post
  - page
  - author
  - category
  - tag
  - media
```

### Managing Pages

**Create a Page:**

```bash
ainative-code strapi content create page \
  --data '{
    "title": "About Us",
    "slug": "about",
    "content": "We are a team of passionate developers...",
    "meta_description": "Learn about our mission and team",
    "published": true
  }'
```

**List Pages:**

```bash
ainative-code strapi content list page --limit 50
```

**Update Page:**

```bash
ainative-code strapi content update page 123 \
  --data '{
    "content": "Updated content here..."
  }'
```

### Managing Authors

**Create Author:**

```bash
ainative-code strapi content create author \
  --data '{
    "name": "John Doe",
    "email": "john@example.com",
    "bio": "Senior Developer with 10+ years of experience",
    "avatar": "https://example.com/john.jpg"
  }'
```

**List Authors:**

```bash
ainative-code strapi content list author
```

### Managing Categories

**Create Category:**

```bash
ainative-code strapi content create category \
  --data '{
    "name": "Tutorials",
    "slug": "tutorials",
    "description": "Step-by-step guides and tutorials"
  }'
```

**Link Post to Category:**

```bash
ainative-code strapi blog update --id 42 \
  --data '{"category": "tutorials"}'
```

## AI-Assisted Content Creation

### Generate Blog Post with AI

**Interactive Mode:**

```bash
ainative-code chat
```

```
User: Create a blog post about Docker best practices

AI: I'll create a comprehensive blog post about Docker best practices.

[Generates content]

Title: Docker Best Practices for Production Environments
Content: [generated markdown article]

Would you like me to publish this to Strapi?

User: Yes, as a draft

AI: [Creates in Strapi as draft]
✓ Published to Strapi as draft (ID: 456)
You can review at: https://cms.ainative.studio/blog-posts/456
```

### Generate Multiple Posts

**Batch Content Generation:**

```bash
# Create a script
cat > generate-posts.sh << 'EOF'
#!/bin/bash

TOPICS=(
  "Getting Started with Kubernetes"
  "Docker Security Best Practices"
  "CI/CD with GitHub Actions"
  "Microservices Architecture Patterns"
)

for topic in "${TOPICS[@]}"; do
  echo "Generating: $topic"

  # Generate content with AI
  ainative-code chat "Generate a technical blog post about: $topic.
    Format as markdown with code examples." > "content-${topic// /-}.md"

  # Create draft in Strapi
  ainative-code strapi blog create \
    --title "$topic" \
    --content "@content-${topic// /-}.md" \
    --author "AI Assistant" \
    --tags "tutorial,technical" \
    --status draft

  echo "✓ Created draft: $topic"
done
EOF

chmod +x generate-posts.sh
./generate-posts.sh
```

### AI Content Enhancement

**Improve Existing Content:**

```bash
# Get existing post
ainative-code strapi blog list --id 42 --json > post-42.json

# Extract content
CONTENT=$(jq -r '.data.attributes.content' post-42.json)

# Enhance with AI
ainative-code chat "Improve this blog post:
- Add more technical depth
- Include code examples
- Improve SEO
- Add conclusion

Original content:
$CONTENT" > enhanced-content.md

# Update post
ainative-code strapi blog update --id 42 --content @enhanced-content.md
```

## Automation Workflows

### Scheduled Publishing

**Cron Job Example:**

```bash
# Add to crontab
0 9 * * 1 /path/to/weekly-publish.sh
```

**weekly-publish.sh:**

```bash
#!/bin/bash

# Get all draft posts
DRAFTS=$(ainative-code strapi blog list --status draft --json)

# Get post IDs
IDS=$(echo "$DRAFTS" | jq -r '.data[].id')

# Publish first draft
FIRST_ID=$(echo "$IDS" | head -1)

if [ -n "$FIRST_ID" ]; then
  ainative-code strapi blog publish --id "$FIRST_ID"
  echo "Published post ID: $FIRST_ID"
else
  echo "No drafts to publish"
fi
```

### Content Backup

```bash
#!/bin/bash

# Backup all published posts
ainative-code strapi blog list --status published --json > \
  "backups/posts-$(date +%Y%m%d).json"

# Backup pages
ainative-code strapi content list page --json > \
  "backups/pages-$(date +%Y%m%d).json"

echo "✓ Backup completed"
```

### Content Migration

```bash
#!/bin/bash

# Export from old CMS (example: WordPress)
wp post list --format=json > wordpress-posts.json

# Convert and import to Strapi
jq -c '.[]' wordpress-posts.json | while read post; do
  TITLE=$(echo "$post" | jq -r '.title')
  CONTENT=$(echo "$post" | jq -r '.content')
  AUTHOR=$(echo "$post" | jq -r '.author')

  ainative-code strapi blog create \
    --title "$TITLE" \
    --content "$CONTENT" \
    --author "$AUTHOR" \
    --status draft

  echo "✓ Migrated: $TITLE"
done
```

### RSS Feed Generation

```bash
#!/bin/bash

# Get published posts
POSTS=$(ainative-code strapi blog list --status published --json)

# Generate RSS feed
cat > rss.xml << 'EOF'
<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>AINative Blog</title>
    <link>https://ainative.studio/blog</link>
    <description>Latest posts from AINative</description>
EOF

echo "$POSTS" | jq -r '.data[] |
  "<item>
    <title>\(.attributes.title)</title>
    <link>https://ainative.studio/blog/\(.attributes.slug)</link>
    <description>\(.attributes.content | .[0:200])...</description>
    <pubDate>\(.attributes.publishedAt)</pubDate>
  </item>"' >> rss.xml

cat >> rss.xml << 'EOF'
  </channel>
</rss>
EOF

echo "✓ RSS feed generated"
```

## Best Practices

### 1. Content Organization

```bash
# Use consistent naming
--title "Clear, Descriptive Title"
--slug "url-friendly-slug"
--tags "relevant,searchable,tags"

# Structure content
# - Clear H1 (title)
# - H2 for sections
# - H3 for subsections
# - Code blocks with syntax highlighting
# - Images with alt text
```

### 2. SEO Optimization

```json
{
  "title": "Docker Best Practices for Production - Complete Guide 2024",
  "meta_description": "Learn Docker best practices for production deployments. Security, performance, and reliability tips with real-world examples.",
  "slug": "docker-production-best-practices",
  "tags": ["docker", "devops", "production", "tutorial"],
  "featured_image": {
    "alt": "Docker containers in production environment",
    "url": "https://cdn.example.com/docker-production.jpg"
  }
}
```

### 3. Version Control

```bash
# Keep content in Git
git add content/
git commit -m "Add Docker best practices article"
git push

# Deploy to Strapi from CI/CD
ainative-code strapi blog create \
  --title "$(cat article-title.txt)" \
  --content @article.md \
  --status draft
```

### 4. Review Workflow

```bash
# 1. Create as draft
ainative-code strapi blog create \
  --title "..." \
  --content @article.md \
  --status draft

# 2. Review in Strapi admin

# 3. Update if needed
ainative-code strapi blog update --id 42 --content @revised.md

# 4. Publish when ready
ainative-code strapi blog publish --id 42
```

### 5. Error Handling

```bash
#!/bin/bash

create_post() {
  local title="$1"
  local content="$2"

  # Try to create post
  result=$(ainative-code strapi blog create \
    --title "$title" \
    --content "$content" \
    --json 2>&1)

  # Check for errors
  if echo "$result" | grep -q "error"; then
    echo "❌ Failed to create post: $title"
    echo "$result" | jq -r '.error.message'
    return 1
  else
    post_id=$(echo "$result" | jq -r '.data.id')
    echo "✓ Created post: $title (ID: $post_id)"
    return 0
  fi
}

# Usage
create_post "My Title" "@content.md" || exit 1
```

### 6. Bulk Operations

```bash
# Publish all drafts by specific author
ainative-code strapi blog list \
  --status draft \
  --author "AI Assistant" \
  --json | \
  jq -r '.data[].id' | \
  while read id; do
    ainative-code strapi blog publish --id "$id"
    echo "✓ Published post $id"
  done
```

## Troubleshooting

### Authentication Errors

**Problem:** 401 Unauthorized

**Solutions:**

```bash
# Check API token
echo $STRAPI_API_KEY

# Regenerate token in Strapi
# Settings > API Tokens > Create New Token

# Set new token
export STRAPI_API_KEY="new-token-here"

# Test authentication
ainative-code strapi test
```

### Connection Errors

**Problem:** Cannot connect to Strapi

**Solutions:**

```bash
# Verify URL
curl -I https://cms.ainative.studio/api

# Check network
ping cms.ainative.studio

# Verify HTTPS
curl https://cms.ainative.studio/api/blog-posts \
  -H "Authorization: Bearer $STRAPI_API_KEY"

# Check firewall
# Ensure outbound HTTPS (443) is allowed
```

### Post Not Found

**Problem:** 404 Not Found when updating/deleting

**Solutions:**

```bash
# List all posts to find correct ID
ainative-code strapi blog list --json | jq '.data[] | {id, title}'

# Verify post exists
ainative-code strapi blog list --id 42 --json

# Check permissions
# Ensure your API token has write access
```

### File Upload Issues

**Problem:** Cannot upload content from file

**Solutions:**

```bash
# Check file exists
ls -la article.md

# Use absolute path
ainative-code strapi blog create \
  --title "Title" \
  --content @/full/path/to/article.md

# Check file encoding
file article.md  # Should be UTF-8

# Check file size
du -h article.md  # Strapi may have size limits
```

### Markdown Rendering Issues

**Problem:** Markdown not rendering correctly

**Solutions:**

```bash
# Validate markdown syntax
npx markdownlint article.md

# Preview locally
npx marked article.md > preview.html

# Check for special characters
# Escape backticks, quotes, etc.

# Use proper code blocks
# ```language
# code here
# ```
```

### Slug Conflicts

**Problem:** Slug already exists

**Solutions:**

```bash
# Use unique slug
ainative-code strapi blog create \
  --title "Getting Started" \
  --slug "getting-started-2024" \
  --content @article.md

# Let Strapi auto-generate
# Omit --slug flag
ainative-code strapi blog create \
  --title "Getting Started" \
  --content @article.md
```

## Next Steps

- [ZeroDB Integration](zerodb-integration.md)
- [Design Token Integration](design-token-integration.md)
- [RLHF Feedback System](rlhf-integration.md)
- [Authentication Setup](authentication-setup.md)

## Resources

- [Strapi Documentation](https://docs.strapi.io/)
- [Strapi API Reference](https://docs.strapi.io/dev-docs/api/rest)
- [Markdown Guide](https://www.markdownguide.org/)
- [Content Strategy Best Practices](https://contentstrategy.com/)
