# Strapi Blog Operations

This guide covers the Strapi blog post management features in AINative Code, including creating, listing, updating, publishing, and deleting blog posts.

## Table of Contents

- [Overview](#overview)
- [Configuration](#configuration)
- [Command Reference](#command-reference)
  - [Create Blog Post](#create-blog-post)
  - [List Blog Posts](#list-blog-posts)
  - [Update Blog Post](#update-blog-post)
  - [Publish Blog Post](#publish-blog-post)
  - [Delete Blog Post](#delete-blog-post)
- [Markdown Support](#markdown-support)
- [Integration with Strapi API](#integration-with-strapi-api)
- [Examples](#examples)
- [Troubleshooting](#troubleshooting)

## Overview

The Strapi blog operations feature provides a command-line interface for managing blog posts in a Strapi CMS instance. Key features include:

- **Full CRUD Operations**: Create, read, update, and delete blog posts
- **Markdown Support**: Native markdown content support with file input
- **Filtering**: Filter posts by status, author, and other attributes
- **Pagination**: Efficient handling of large blog post collections
- **Publishing Workflow**: Draft and publish workflow for content management
- **JSON Output**: Machine-readable JSON output for automation

## Configuration

Before using Strapi blog operations, configure your Strapi CMS connection:

```bash
# Set Strapi URL
ainative-code strapi config --url https://your-strapi-instance.com

# Set API token (if required)
ainative-code strapi config --token your-api-token

# Verify configuration
ainative-code strapi config
```

Configuration is stored in your AINative Code configuration file (`~/.ainative-code.yaml`).

## Command Reference

### Create Blog Post

Create a new blog post in Strapi CMS.

#### Usage

```bash
ainative-code strapi blog create [flags]
```

#### Flags

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `--title` | string | Yes | Blog post title |
| `--content` | string | Yes | Blog post content or `@filename` for file input |
| `--author` | string | No | Blog post author |
| `--slug` | string | No | URL-friendly slug (auto-generated if not provided) |
| `--tags` | []string | No | Comma-separated list of tags |
| `--status` | string | No | Post status: `draft` (default) or `published` |
| `--json` | bool | No | Output as JSON |

#### Examples

**Create a draft post with inline content:**
```bash
ainative-code strapi blog create \
  --title "My First Post" \
  --content "# Hello World\n\nThis is my first blog post!" \
  --author "John Doe"
```

**Create from a markdown file:**
```bash
ainative-code strapi blog create \
  --title "Technical Deep Dive" \
  --content @article.md \
  --author "Jane Smith" \
  --tags go,rust,performance
```

**Create and immediately publish:**
```bash
ainative-code strapi blog create \
  --title "Breaking News" \
  --content @news.md \
  --author "Editor" \
  --status published
```

**Get JSON output for automation:**
```bash
ainative-code strapi blog create \
  --title "Automated Post" \
  --content @content.md \
  --author "Bot" \
  --json
```

#### Output

```
Blog post created successfully!

ID: 42
Title: My First Post
Author: John Doe
Status: draft
Created: 2024-01-15 10:30:00
```

### List Blog Posts

List blog posts with optional filtering and pagination.

#### Usage

```bash
ainative-code strapi blog list [flags]
```

#### Flags

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `--status` | string | No | Filter by status (`draft` or `published`) |
| `--author` | string | No | Filter by author name |
| `--limit` | int | No | Maximum posts to return (default: 25) |
| `--page` | int | No | Page number for pagination (default: 1) |
| `--json` | bool | No | Output as JSON |

#### Examples

**List all posts:**
```bash
ainative-code strapi blog list
```

**List only published posts:**
```bash
ainative-code strapi blog list --status published
```

**List posts by a specific author:**
```bash
ainative-code strapi blog list --author "John Doe"
```

**List with pagination:**
```bash
ainative-code strapi blog list --limit 50 --page 2
```

**Combine filters:**
```bash
ainative-code strapi blog list --status published --author "Jane Smith" --limit 10
```

**Get JSON output:**
```bash
ainative-code strapi blog list --json | jq '.data[] | {id, title: .attributes.title}'
```

#### Output

```
ID  TITLE                                   AUTHOR      STATUS     CREATED
--  -----                                   ------      ------     -------
42  My First Post                           John Doe    draft      2024-01-15
41  Technical Deep Dive                     Jane Smith  published  2024-01-14
40  Getting Started with AINative          Editor      published  2024-01-13

Page 1 of 5 (Total: 123 posts)
```

### Update Blog Post

Update an existing blog post.

#### Usage

```bash
ainative-code strapi blog update [flags]
```

#### Flags

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `--id` | int | Yes | Blog post ID to update |
| `--title` | string | No | New title |
| `--content` | string | No | New content or `@filename` |
| `--author` | string | No | New author |
| `--slug` | string | No | New slug |
| `--tags` | []string | No | New tags (comma-separated) |
| `--json` | bool | No | Output as JSON |

**Note:** At least one field to update must be provided.

#### Examples

**Update title only:**
```bash
ainative-code strapi blog update --id 42 --title "Updated Title"
```

**Update content from file:**
```bash
ainative-code strapi blog update --id 42 --content @revised-content.md
```

**Update multiple fields:**
```bash
ainative-code strapi blog update \
  --id 42 \
  --title "Revised Article" \
  --author "Jane Doe" \
  --tags go,rust,best-practices
```

#### Output

```
Blog post updated successfully!

ID: 42
Title: Updated Title
Author: John Doe
Status: draft
Updated: 2024-01-15 14:20:00
```

### Publish Blog Post

Publish a draft blog post by changing its status to published.

#### Usage

```bash
ainative-code strapi blog publish [flags]
```

#### Flags

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `--id` | int | Yes | Blog post ID to publish |
| `--json` | bool | No | Output as JSON |

#### Examples

**Publish a post:**
```bash
ainative-code strapi blog publish --id 42
```

**Publish with JSON output:**
```bash
ainative-code strapi blog publish --id 42 --json
```

#### Output

```
Blog post published successfully!

ID: 42
Title: My First Post
Status: published
Published: 2024-01-15 15:00:00
```

### Delete Blog Post

Delete a blog post from Strapi CMS.

**WARNING:** This action cannot be undone.

#### Usage

```bash
ainative-code strapi blog delete [flags]
```

#### Flags

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `--id` | int | Yes | Blog post ID to delete |

#### Examples

**Delete a post:**
```bash
ainative-code strapi blog delete --id 42
```

#### Output

```
WARNING: You are about to delete blog post ID 42. This action cannot be undone.
Type 'yes' to confirm: yes
Blog post ID 42 deleted successfully.
```

## Markdown Support

The Strapi blog operations fully support markdown content:

### Inline Markdown

```bash
ainative-code strapi blog create \
  --title "Quick Note" \
  --content "# Heading\n\n**Bold text** and *italic text*\n\n- List item 1\n- List item 2" \
  --author "Author"
```

### File Input

Use the `@filename` syntax to read content from a markdown file:

```bash
ainative-code strapi blog create \
  --title "From File" \
  --content @article.md \
  --author "Author"
```

### Example Markdown File

```markdown
# Introduction to AINative Code

AINative Code is a powerful CLI tool for managing your development workflow.

## Features

- **ZeroDB Integration**: Seamless database operations
- **Strapi CMS**: Content management made easy
- **RLHF Feedback**: Continuous improvement through feedback

## Getting Started

1. Install AINative Code
2. Configure your services
3. Start building!

```bash
ainative-code --version
```

## Code Examples

\`\`\`go
func main() {
    fmt.Println("Hello, AINative!")
}
\`\`\`
```

## Integration with Strapi API

The Strapi blog operations integrate with the Strapi REST API using the following conventions:

### API Endpoints

- **Create**: `POST /api/blog-posts`
- **List**: `GET /api/blog-posts`
- **Get**: `GET /api/blog-posts/{id}`
- **Update**: `PUT /api/blog-posts/{id}`
- **Delete**: `DELETE /api/blog-posts/{id}`

### Authentication

API requests include JWT authentication tokens configured via:

```bash
ainative-code strapi config --token your-api-token
```

### Request Format

Blog posts follow the Strapi data format:

```json
{
  "data": {
    "title": "Blog Post Title",
    "content": "Markdown content here",
    "author": "Author Name",
    "status": "draft",
    "tags": ["tag1", "tag2"]
  }
}
```

### Response Format

Strapi returns blog posts in this format:

```json
{
  "data": {
    "id": 42,
    "attributes": {
      "title": "Blog Post Title",
      "content": "Markdown content here",
      "author": "Author Name",
      "status": "draft",
      "slug": "blog-post-title",
      "publishedAt": null,
      "createdAt": "2024-01-15T10:30:00.000Z",
      "updatedAt": "2024-01-15T10:30:00.000Z"
    }
  }
}
```

## Examples

### Workflow: Draft to Publish

```bash
# 1. Create a draft post
ainative-code strapi blog create \
  --title "New Feature Announcement" \
  --content @announcement.md \
  --author "Product Team" \
  --tags announcements,features

# Output: ID: 42

# 2. Review and update if needed
ainative-code strapi blog update \
  --id 42 \
  --content @announcement-revised.md

# 3. Publish when ready
ainative-code strapi blog publish --id 42
```

### Automated Publishing

```bash
#!/bin/bash
# publish-daily-update.sh

# Create post from generated content
POST_ID=$(ainative-code strapi blog create \
  --title "Daily Update: $(date +%Y-%m-%d)" \
  --content @daily-report.md \
  --author "Automation Bot" \
  --status published \
  --json | jq -r '.id')

echo "Published daily update with ID: $POST_ID"
```

### Bulk Operations

```bash
# List all draft posts and publish them
ainative-code strapi blog list --status draft --json | \
  jq -r '.data[].id' | \
  while read id; do
    ainative-code strapi blog publish --id $id
  done
```

### Content Migration

```bash
# Export all published posts to local files
ainative-code strapi blog list --status published --json | \
  jq -r '.data[] | "\(.id)|\(.attributes.title)|\(.attributes.content)"' | \
  while IFS='|' read id title content; do
    echo "$content" > "export/post-${id}.md"
    echo "Exported: $title"
  done
```

## Troubleshooting

### Common Issues

#### 1. Strapi URL Not Configured

**Error:**
```
Strapi URL not configured. Use 'ainative-code strapi config --url <url>' to set it up
```

**Solution:**
```bash
ainative-code strapi config --url https://your-strapi-instance.com
```

#### 2. Authentication Failed

**Error:**
```
HTTP 401: Unauthorized
```

**Solution:**
```bash
# Configure your API token
ainative-code strapi config --token your-api-token

# Or check your existing token
ainative-code strapi config
```

#### 3. Post Not Found

**Error:**
```
HTTP 404: Not Found
```

**Solution:**
```bash
# Verify the post ID exists
ainative-code strapi blog list

# Use the correct ID
ainative-code strapi blog update --id <correct-id> --title "Updated"
```

#### 4. File Not Found

**Error:**
```
failed to read file article.md: no such file or directory
```

**Solution:**
```bash
# Use absolute path
ainative-code strapi blog create \
  --title "Post" \
  --content @/full/path/to/article.md \
  --author "Author"

# Or use relative path from current directory
ainative-code strapi blog create \
  --title "Post" \
  --content @./relative/path/article.md \
  --author "Author"
```

#### 5. Invalid Markdown Syntax

**Issue:** Markdown not rendering correctly in Strapi

**Solution:**
- Ensure your markdown file is valid
- Check for special characters that need escaping
- Verify Strapi's markdown parser configuration

### Debug Mode

Enable verbose logging for troubleshooting:

```bash
export AINATIVE_CODE_LOG_LEVEL=debug
ainative-code strapi blog create --title "Test" --content "Content" --author "Test"
```

### API Connectivity Test

```bash
# Test Strapi connection
ainative-code strapi test

# Check API status
curl -I https://your-strapi-instance.com/api
```

## API Reference

For detailed API documentation, see:

- [Strapi Client API Documentation](../internal/client/strapi/README.md)
- [CLI Command Reference](./commands.md)

## Related Documentation

- [Strapi CMS Configuration](./strapi-setup.md)
- [Content Management Workflows](./content-workflows.md)
- [API Authentication](./authentication.md)
- [Markdown Guide](./markdown.md)

## Support

For issues or questions:

- GitHub Issues: https://github.com/AINative-studio/ainative-code/issues
- Documentation: https://docs.ainative.studio
- Community: https://community.ainative.studio
