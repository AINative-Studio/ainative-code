# Tools Usage Guide

This guide covers the available tools in AINative Code, including built-in tools, MCP (Model Context Protocol) servers, and custom tool integration.

## Table of Contents

1. [Overview](#overview)
2. [Built-in Tools](#built-in-tools)
3. [MCP Server Integration](#mcp-server-integration)
4. [Tool Permissions](#tool-permissions)
5. [Custom Tools](#custom-tools)
6. [Tool Configuration](#tool-configuration)
7. [Best Practices](#best-practices)

## Overview

Tools extend the AI's capabilities beyond text generation, allowing it to:

- Execute bash commands
- Read and write files
- Browse the web
- Analyze code
- Interact with databases
- Call external APIs

### Tool Types

1. **Built-in Tools**: Pre-configured tools available out of the box
2. **MCP Tools**: Tools from Model Context Protocol servers
3. **Custom Tools**: User-defined tools for specific workflows

## Built-in Tools

### Bash Tool

Execute shell commands and scripts.

**Capabilities:**
- Run terminal commands
- Execute scripts
- Capture output
- Handle long-running processes

**Example Usage:**

```
User: Run git status in my project
AI: [Executes: git status]
Here's the current git status:
On branch main
Your branch is up to date with 'origin/main'.
```

**Configuration:**

```yaml
tools:
  terminal:
    enabled: true
    allowed_commands:
      - git
      - npm
      - go
      - python
      - docker
      - kubectl
    blocked_commands:
      - rm -rf /
      - dd
      - mkfs
    timeout: 300s
    working_dir: /workspace
```

**Security:**
- Commands run in restricted environment
- Blocked commands list prevents dangerous operations
- Timeout prevents hanging processes
- Working directory restriction

### File Operations Tool

Read, write, and manage files.

**Capabilities:**
- Read file contents
- Write/update files
- List directories
- Create/delete files
- File search

**Example Usage:**

```
User: Read the main.go file
AI: [Reads file]
Here's the content of main.go:
[... file contents ...]

User: Create a new file called config.yaml with default settings
AI: [Creates file]
I've created config.yaml with the following content:
[... generated config ...]
```

**Configuration:**

```yaml
tools:
  filesystem:
    enabled: true
    allowed_paths:
      - /home/user/projects
      - /workspace
    blocked_paths:
      - /etc
      - /sys
      - /proc
    max_file_size: 10485760  # 10 MB
    allowed_extensions:
      - .go
      - .js
      - .ts
      - .py
      - .java
      - .rs
      - .md
      - .json
      - .yaml
```

**Safety:**
- Path restrictions prevent access to system files
- File size limits prevent memory issues
- Extension filtering controls file types
- Automatic backups for write operations

### Web Fetch Tool

Retrieve and analyze web content.

**Capabilities:**
- Fetch web pages
- Parse HTML/JSON
- Extract information
- Follow redirects

**Example Usage:**

```
User: Fetch the latest Go release notes
AI: [Fetches https://go.dev/doc/devel/release]
Here are the latest Go release notes:
[... extracted information ...]
```

**Configuration:**

```yaml
tools:
  browser:
    enabled: true
    headless: true
    timeout: 30s
    user_agent: "AINative-Code/0.1.0"
    allowed_domains:
      - golang.org
      - github.com
      - stackoverflow.com
```

**Features:**
- JavaScript rendering (when headless browser enabled)
- Cookie handling
- Custom headers
- Rate limiting

### Code Analysis Tool

Analyze code structure and patterns.

**Capabilities:**
- Parse code syntax
- Detect patterns
- Calculate metrics
- Find dependencies
- Generate documentation

**Example Usage:**

```
User: Analyze this Go package for code smells
AI: [Analyzes code]
I've analyzed the package. Here are the findings:
1. High cyclomatic complexity in main()
2. Duplicate error handling patterns
3. Missing test coverage in auth package
```

**Configuration:**

```yaml
tools:
  code_analysis:
    enabled: true
    languages:
      - go
      - javascript
      - typescript
      - python
    max_file_size: 5242880  # 5 MB
    include_tests: true
```

## MCP Server Integration

Model Context Protocol (MCP) allows external servers to provide tools to AINative Code.

### Understanding MCP

MCP servers can provide:
- Custom tools specific to your workflow
- Integration with internal systems
- Domain-specific capabilities
- Team-shared tools

### MCP Server Commands

```bash
# List configured MCP servers
ainative-code mcp list-servers

# Add a new MCP server
ainative-code mcp add-server \
  --name mytools \
  --url http://localhost:3000 \
  --timeout 30s

# Remove an MCP server
ainative-code mcp remove-server mytools

# List available tools from all MCP servers
ainative-code mcp list-tools

# Test a specific tool
ainative-code mcp test-tool mytools.deploy \
  --args '{"environment": "staging"}'

# Discover tools from all servers
ainative-code mcp discover
```

### Configuring MCP Servers

```yaml
# config.yaml
mcp:
  servers:
    - name: mytools
      url: http://localhost:3000
      timeout: 30s
      headers:
        Authorization: "Bearer ${MCP_TOKEN}"
      enabled: true

    - name: github-tools
      url: https://mcp.github.com
      timeout: 60s
      enabled: true

    - name: slack-tools
      url: http://localhost:3001
      timeout: 30s
      enabled: true
```

### Using MCP Tools

Once configured, MCP tools are automatically available to the AI:

```
User: Deploy the app to staging using mytools
AI: [Calls mytools.deploy with parameters]
Deployment initiated to staging environment.
[... deployment output ...]

User: Create a GitHub issue for this bug
AI: [Calls github-tools.create_issue]
Issue created: https://github.com/user/repo/issues/123
```

### Creating an MCP Server

Example MCP server in Go:

```go
package main

import (
    "encoding/json"
    "net/http"
    "github.com/AINative-studio/ainative-code/pkg/mcp"
)

func main() {
    server := mcp.NewServer()

    // Register a tool
    server.RegisterTool(mcp.Tool{
        Name:        "deploy",
        Description: "Deploy application to environment",
        Parameters: map[string]interface{}{
            "environment": "string",
            "version":     "string",
        },
        Handler: func(args map[string]interface{}) (interface{}, error) {
            env := args["environment"].(string)
            ver := args["version"].(string)

            // Deployment logic
            result := deploy(env, ver)

            return result, nil
        },
    })

    http.ListenAndServe(":3000", server)
}
```

### MCP Tool Discovery

Tools are automatically discovered and shown in:

```bash
ainative-code mcp list-tools
```

Output:
```
TOOL NAME              SERVER      DESCRIPTION
─────────────────────  ──────────  ────────────────────────────
mytools.deploy         mytools     Deploy application to environment
mytools.rollback       mytools     Rollback to previous version
github-tools.create_issue  github  Create a GitHub issue
github-tools.close_pr  github      Close a pull request
slack-tools.send       slack       Send message to Slack channel
```

## Tool Permissions

### Permission Levels

Tools operate under different permission levels:

1. **Read-only**: Can read data but not modify
2. **Read-write**: Can read and modify data
3. **Execute**: Can execute commands
4. **Admin**: Full system access

### Configuring Permissions

```yaml
tools:
  filesystem:
    enabled: true
    permissions: read-write
    allowed_paths:
      - /workspace

  terminal:
    enabled: true
    permissions: execute
    allowed_commands:
      - git
      - npm
```

### Permission Prompts

For sensitive operations, the AI will ask for confirmation:

```
AI: I need to execute 'rm -rf dist/'. This will delete the dist directory.
    Do you want to proceed? [y/N]
```

### Automatic Approval

For trusted workflows, configure automatic approval:

```yaml
tools:
  auto_approve:
    enabled: true
    patterns:
      - "git status"
      - "git diff"
      - "npm install"
```

## Custom Tools

### Creating Custom Tools

Define custom tools in your configuration:

```yaml
tools:
  custom:
    - name: run_tests
      description: "Run all unit tests"
      command: "go test ./..."
      working_dir: "/workspace"

    - name: build_docker
      description: "Build Docker image"
      command: "docker build -t myapp:latest ."

    - name: deploy_staging
      description: "Deploy to staging"
      command: "./scripts/deploy.sh staging"
      require_confirmation: true
```

### Tool Scripts

Create reusable tool scripts:

```bash
# ~/.config/ainative-code/tools/deploy.sh
#!/bin/bash
set -e

ENVIRONMENT=$1
VERSION=$2

echo "Deploying version $VERSION to $ENVIRONMENT"
kubectl set image deployment/myapp myapp=myapp:$VERSION -n $ENVIRONMENT
kubectl rollout status deployment/myapp -n $ENVIRONMENT
echo "Deployment complete"
```

Register the script:

```yaml
tools:
  custom:
    - name: deploy
      description: "Deploy application"
      script: "~/.config/ainative-code/tools/deploy.sh"
      parameters:
        - environment
        - version
```

Usage:

```
User: Deploy version 1.2.3 to staging
AI: [Executes deploy.sh staging 1.2.3]
Deploying version 1.2.3 to staging
[... deployment output ...]
```

### Tool Composition

Combine multiple tools into workflows:

```yaml
tools:
  workflows:
    - name: full_deployment
      description: "Complete deployment workflow"
      steps:
        - tool: run_tests
        - tool: build_docker
        - tool: deploy_staging
          on_success: notify_slack
```

## Tool Configuration

### Global Tool Settings

```yaml
tools:
  # Global settings
  enabled: true
  require_confirmation: false
  timeout: 300s
  max_concurrent: 5

  # Logging
  log_tool_calls: true
  log_tool_output: true

  # Error handling
  retry_on_failure: true
  max_retries: 3
```

### Tool-Specific Settings

```yaml
tools:
  terminal:
    timeout: 600s  # 10 minutes
    max_output_size: 1048576  # 1 MB

  filesystem:
    max_file_size: 10485760  # 10 MB
    create_backups: true

  browser:
    timeout: 30s
    max_page_size: 5242880  # 5 MB
```

### Environment Variables

Tools can access environment variables:

```yaml
tools:
  custom:
    - name: deploy
      command: "./deploy.sh"
      env:
        AWS_REGION: us-east-1
        ENVIRONMENT: "${DEPLOY_ENV}"
        API_KEY: "${DEPLOY_API_KEY}"
```

## Best Practices

### 1. Principle of Least Privilege

Only enable tools you need:

```yaml
tools:
  filesystem:
    enabled: true  # Needed
  terminal:
    enabled: true  # Needed
  browser:
    enabled: false  # Not needed, disabled
```

### 2. Restrict Access

Limit tool access to necessary paths:

```yaml
tools:
  filesystem:
    allowed_paths:
      - /workspace/myproject  # Specific project only
    blocked_paths:
      - /workspace/myproject/secrets  # Block sensitive dirs
```

### 3. Use Timeouts

Prevent hanging operations:

```yaml
tools:
  terminal:
    timeout: 300s  # 5 minute timeout
```

### 4. Enable Logging

Track tool usage for debugging:

```yaml
tools:
  log_tool_calls: true
  log_file: ~/.config/ainative-code/tools.log
```

### 5. Require Confirmation for Destructive Operations

```yaml
tools:
  custom:
    - name: drop_database
      command: "dropdb myapp"
      require_confirmation: true  # Always confirm
      confirm_message: "This will delete the database. Are you sure?"
```

### 6. Validate Tool Inputs

```yaml
tools:
  custom:
    - name: deploy
      parameters:
        environment:
          type: string
          allowed:
            - staging
            - production
          required: true
        version:
          type: string
          pattern: "^v?[0-9]+\\.[0-9]+\\.[0-9]+$"
          required: true
```

### 7. Use MCP for Complex Tools

For complex tools, create an MCP server instead of shell scripts:

```go
// Provides better error handling, validation, and maintainability
server.RegisterTool(mcp.Tool{
    Name: "deploy",
    Validate: func(args map[string]interface{}) error {
        // Validation logic
    },
    Handler: func(args map[string]interface{}) (interface{}, error) {
        // Deployment logic with proper error handling
    },
})
```

### 8. Version Control Tool Configurations

Store tool configs in version control:

```bash
# .ainative-tools.yaml in project root
tools:
  custom:
    - name: run_tests
      command: "make test"
    - name: build
      command: "make build"
```

Reference in main config:

```yaml
tools:
  import:
    - .ainative-tools.yaml
```

### 9. Create Tool Documentation

Document custom tools:

```yaml
tools:
  custom:
    - name: deploy
      description: |
        Deploy application to specified environment.

        Usage:
          Deploy to staging: deploy staging v1.2.3
          Deploy to production: deploy production v1.2.3

        Requirements:
          - kubectl configured
          - Access to cluster
          - Valid version tag
      command: "./scripts/deploy.sh"
```

### 10. Test Tools Independently

Test tools before using with AI:

```bash
# Test MCP tool
ainative-code mcp test-tool mytools.deploy --args '{"environment": "staging"}'

# Test custom tool
ainative-code tools test run_tests
```

## Security Considerations

### 1. Sandbox Tools

Run tools in isolated environments:

```yaml
tools:
  sandbox:
    enabled: true
    type: docker  # or: vm, namespace
    image: ainative/sandbox:latest
```

### 2. Audit Logging

Log all tool executions:

```yaml
tools:
  audit:
    enabled: true
    log_file: /var/log/ainative-code/audit.log
    include_output: true
    include_environment: false  # Don't log env vars (may contain secrets)
```

### 3. Rate Limiting

Prevent tool abuse:

```yaml
tools:
  rate_limit:
    enabled: true
    max_calls_per_minute: 60
    max_concurrent: 5
```

### 4. Secret Management

Never hardcode secrets:

```yaml
# Bad
tools:
  custom:
    - name: deploy
      env:
        API_KEY: "hardcoded-secret"  # Don't do this!

# Good
tools:
  custom:
    - name: deploy
      env:
        API_KEY: "${DEPLOY_API_KEY}"  # Use environment variable
```

## Troubleshooting

### Tool Not Found

```bash
# List available tools
ainative-code tools list

# Check tool configuration
ainative-code config get tools
```

### Permission Denied

```bash
# Check file permissions
ls -la ~/.config/ainative-code/tools/

# Make script executable
chmod +x ~/.config/ainative-code/tools/deploy.sh
```

### MCP Server Not Responding

```bash
# Check server health
ainative-code mcp list-servers

# Test server connection
curl http://localhost:3000/health

# Restart MCP server
# (depends on how you're running it)
```

### Tool Timeout

```bash
# Increase timeout for slow operations
ainative-code config set tools.terminal.timeout 600s
```

## Next Steps

- [MCP Integration Guide](ainative-integrations.md) - AINative platform tools
- [Configuration Guide](configuration.md) - Detailed tool configuration
- [Security Best Practices](../security.md) - Secure tool usage
