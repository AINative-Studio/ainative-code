# LLM Providers Guide

This guide covers all supported LLM providers, their setup, features, and best practices.

## Table of Contents

1. [Overview](#overview)
2. [Supported Providers](#supported-providers)
3. [Anthropic Claude](#anthropic-claude)
4. [OpenAI](#openai)
5. [Google Gemini](#google-gemini)
6. [AWS Bedrock](#aws-bedrock)
7. [Azure OpenAI](#azure-openai)
8. [Ollama (Local)](#ollama-local)
9. [Provider Comparison](#provider-comparison)
10. [Fallback Configuration](#fallback-configuration)
11. [Best Practices](#best-practices)

## Overview

AINative Code supports multiple LLM providers, allowing you to:

- Choose the best model for your task
- Switch between providers seamlessly
- Configure automatic fallbacks
- Use local models for privacy
- Optimize for cost vs. performance

### Provider Selection

Set your default provider:

```bash
ainative-code config set llm.default_provider anthropic
```

Or use a specific provider per command:

```bash
ainative-code chat --provider openai
```

## Supported Providers

| Provider | Cloud/Local | Best For | Cost | Speed |
|----------|------------|----------|------|-------|
| Anthropic Claude | Cloud | Complex reasoning, code generation | $$$ | Fast |
| OpenAI GPT | Cloud | General purpose, wide model selection | $$-$$$ | Fast |
| Google Gemini | Cloud | Multimodal tasks, integration with Google | $$ | Fast |
| AWS Bedrock | Cloud | Enterprise, AWS integration | $$$ | Fast |
| Azure OpenAI | Cloud | Enterprise, Microsoft integration | $$$ | Fast |
| Ollama | Local | Privacy, offline use, cost-free | Free | Medium |

## Anthropic Claude

Anthropic's Claude models excel at complex reasoning, detailed code generation, and following instructions precisely.

### Features

- Extended thinking for complex problems
- Large context windows (200K+ tokens)
- Strong code generation capabilities
- Excellent instruction following
- Built-in safety features

### Models

| Model | Context | Best For | Cost |
|-------|---------|----------|------|
| claude-3-5-sonnet-20241022 | 200K | Balanced performance/cost | $$ |
| claude-3-opus-20240229 | 200K | Complex tasks | $$$ |
| claude-3-haiku-20240307 | 200K | Fast responses | $ |

### Setup

#### 1. Get API Key

Visit [Anthropic Console](https://console.anthropic.com/) to create an API key.

#### 2. Configure

```yaml
llm:
  default_provider: anthropic
  anthropic:
    api_key: "${ANTHROPIC_API_KEY}"
    model: claude-3-5-sonnet-20241022
    max_tokens: 4096
    temperature: 0.7
    top_p: 0.9
    top_k: 40
    timeout: 300s
    retry_attempts: 3
    api_version: "2023-06-01"
```

#### 3. Set Environment Variable

```bash
export ANTHROPIC_API_KEY="sk-ant-api03-..."
```

### Extended Thinking

Enable Claude's extended thinking for complex reasoning:

```yaml
llm:
  anthropic:
    extended_thinking:
      enabled: true
      auto_expand: false
      max_depth: 5
```

Usage:

```bash
ainative-code chat "Design a distributed microservices architecture"
# Claude will show detailed reasoning process
```

### Usage Examples

**Complex Code Generation:**
```bash
ainative-code chat "Create a complete REST API with authentication, middleware, and database integration in Go"
```

**Code Review:**
```bash
ainative-code chat "Review this code for security vulnerabilities and performance issues: [code]"
```

**Architecture Design:**
```bash
ainative-code chat "Design a scalable event-driven architecture for an e-commerce platform"
```

## OpenAI

OpenAI provides GPT models with excellent general-purpose capabilities and a wide selection of models.

### Features

- Multiple model options (GPT-4, GPT-3.5)
- Function calling
- JSON mode for structured outputs
- Vision capabilities (GPT-4V)
- DALL-E integration

### Models

| Model | Context | Best For | Cost |
|-------|---------|----------|------|
| gpt-4-turbo-preview | 128K | Latest features, long context | $$$ |
| gpt-4 | 8K | High quality, reliable | $$$ |
| gpt-4-32k | 32K | Long documents | $$$$ |
| gpt-3.5-turbo | 16K | Fast, cost-effective | $ |

### Setup

#### 1. Get API Key

Visit [OpenAI Platform](https://platform.openai.com/api-keys) to create an API key.

#### 2. Configure

```yaml
llm:
  default_provider: openai
  openai:
    api_key: "${OPENAI_API_KEY}"
    model: gpt-4-turbo-preview
    organization: ""  # Optional
    max_tokens: 4096
    temperature: 0.7
    top_p: 1.0
    frequency_penalty: 0.0
    presence_penalty: 0.0
    timeout: 300s
    retry_attempts: 3
```

#### 3. Set Environment Variable

```bash
export OPENAI_API_KEY="sk-..."
```

### Organization (Optional)

If you're part of multiple OpenAI organizations:

```yaml
llm:
  openai:
    organization: "org-..."
```

### Usage Examples

**Quick Responses:**
```bash
ainative-code chat --model gpt-3.5-turbo "Explain how to use Go channels"
```

**Complex Tasks:**
```bash
ainative-code chat --model gpt-4 "Implement OAuth 2.0 authorization code flow"
```

## Google Gemini

Google's Gemini models offer strong multimodal capabilities and integration with Google Cloud.

### Features

- Multimodal (text, images, video)
- Large context windows
- Integration with Google Cloud
- Competitive pricing
- Function calling

### Models

| Model | Context | Best For | Cost |
|-------|---------|----------|------|
| gemini-pro | 32K | Text generation | $$ |
| gemini-pro-vision | 16K | Multimodal tasks | $$ |
| gemini-ultra | 32K | Complex reasoning | $$$ |

### Setup

#### 1. Get API Key

Visit [Google AI Studio](https://makersuite.google.com/app/apikey) for an API key.

#### 2. Configure

```yaml
llm:
  default_provider: google
  google:
    api_key: "${GOOGLE_API_KEY}"
    model: gemini-pro
    project_id: ""  # Optional for Vertex AI
    location: us-central1
    max_tokens: 2048
    temperature: 0.7
    top_p: 0.95
    top_k: 40
    timeout: 300s
    retry_attempts: 3
```

#### 3. Set Environment Variable

```bash
export GOOGLE_API_KEY="..."
```

### Vertex AI (Enterprise)

For Vertex AI deployment:

```yaml
llm:
  google:
    project_id: "your-gcp-project"
    location: us-central1
```

Authenticate with:
```bash
gcloud auth application-default login
```

### Usage Examples

```bash
ainative-code chat --provider google "Explain TensorFlow concepts"
```

## AWS Bedrock

Amazon's managed AI service provides access to multiple foundation models including Claude.

### Features

- Multiple foundation models
- AWS integration
- Enterprise security
- Pay-per-use pricing
- VPC support

### Models

| Model ID | Provider | Best For |
|----------|----------|----------|
| anthropic.claude-3-sonnet-20240229-v1:0 | Anthropic | Balanced |
| anthropic.claude-3-opus-20240229-v1:0 | Anthropic | Complex |
| anthropic.claude-3-haiku-20240307-v1:0 | Anthropic | Fast |
| amazon.titan-text-express-v1 | Amazon | General |

### Setup

#### 1. Configure AWS Credentials

```bash
aws configure
```

Or use environment variables:
```bash
export AWS_ACCESS_KEY_ID="..."
export AWS_SECRET_ACCESS_KEY="..."
export AWS_REGION="us-east-1"
```

#### 2. Configure

```yaml
llm:
  default_provider: bedrock
  bedrock:
    region: us-east-1
    model: anthropic.claude-3-sonnet-20240229-v1:0
    profile: default  # AWS profile name
    max_tokens: 4096
    temperature: 0.7
    top_p: 0.9
    timeout: 300s
    retry_attempts: 3
```

### IAM Permissions

Required IAM permissions:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "bedrock:InvokeModel",
        "bedrock:InvokeModelWithResponseStream"
      ],
      "Resource": "arn:aws:bedrock:*::foundation-model/*"
    }
  ]
}
```

### Usage Examples

```bash
ainative-code chat --provider bedrock "Create AWS Lambda function in Go"
```

## Azure OpenAI

Microsoft's Azure deployment of OpenAI models with enterprise features.

### Features

- OpenAI models on Azure infrastructure
- Enterprise compliance
- Virtual network support
- Private endpoints
- Managed identity support

### Setup

#### 1. Create Azure OpenAI Resource

Create a resource in [Azure Portal](https://portal.azure.com).

#### 2. Deploy a Model

Deploy a model (e.g., GPT-4) through the Azure OpenAI Studio.

#### 3. Configure

```yaml
llm:
  default_provider: azure
  azure:
    api_key: "${AZURE_OPENAI_API_KEY}"
    endpoint: https://your-resource.openai.azure.com/
    deployment_name: gpt-4-deployment  # Your deployment name
    api_version: "2024-02-15-preview"
    max_tokens: 4096
    temperature: 0.7
    top_p: 1.0
    timeout: 300s
    retry_attempts: 3
```

#### 4. Set Environment Variable

```bash
export AZURE_OPENAI_API_KEY="..."
```

### Managed Identity

For Azure VMs/services:

```yaml
llm:
  azure:
    endpoint: https://your-resource.openai.azure.com/
    # No api_key needed with managed identity
```

### Usage Examples

```bash
ainative-code chat --provider azure "Create Azure Function in Go"
```

## Ollama (Local)

Run open-source LLMs locally for privacy and offline use.

### Features

- Complete privacy (no data sent to cloud)
- Offline operation
- No API costs
- Multiple open-source models
- GPU acceleration support

### Models

Popular models for coding:

| Model | Size | Best For |
|-------|------|----------|
| codellama | 7B-34B | Code generation |
| deepseek-coder | 6.7B-33B | Code completion |
| phind-codellama | 34B | Code explanation |
| llama2 | 7B-70B | General purpose |
| mistral | 7B | Fast, efficient |
| mixtral | 8x7B | High quality |

### Setup

#### 1. Install Ollama

**macOS:**
```bash
brew install ollama
```

**Linux:**
```bash
curl -fsSL https://ollama.ai/install.sh | sh
```

**Windows:**
Download from [ollama.ai](https://ollama.ai)

#### 2. Start Ollama

```bash
ollama serve
```

#### 3. Pull Models

```bash
# Download CodeLlama
ollama pull codellama

# Download DeepSeek Coder
ollama pull deepseek-coder

# Download Mistral
ollama pull mistral
```

#### 4. Configure

```yaml
llm:
  default_provider: ollama
  ollama:
    base_url: http://localhost:11434
    model: codellama
    max_tokens: 4096
    temperature: 0.7
    top_p: 0.9
    top_k: 40
    timeout: 600s
    retry_attempts: 3
    keep_alive: 5m
```

### Usage Examples

```bash
# Use default Ollama model
ainative-code chat "Explain Go interfaces"

# Use specific model
ainative-code chat --model deepseek-coder "Generate unit tests"
```

### GPU Acceleration

Ollama automatically uses GPU if available:

```bash
# Check GPU usage
nvidia-smi  # NVIDIA GPUs

# For AMD GPUs
rocm-smi
```

### Custom Models

Create custom models with Modelfile:

```dockerfile
# Modelfile
FROM codellama

SYSTEM "You are an expert Go developer. Always provide idiomatic Go code."

PARAMETER temperature 0.7
PARAMETER top_p 0.9
```

Create the model:
```bash
ollama create my-go-assistant -f Modelfile
```

Use it:
```yaml
llm:
  ollama:
    model: my-go-assistant
```

## Provider Comparison

### Performance Comparison

| Feature | Anthropic | OpenAI | Google | Bedrock | Azure | Ollama |
|---------|-----------|--------|--------|---------|-------|--------|
| Code Quality | Excellent | Excellent | Very Good | Excellent | Excellent | Good |
| Speed | Fast | Fast | Fast | Fast | Fast | Medium |
| Context Length | 200K | 128K | 32K | 200K | 128K | 4K-32K |
| Cost | $$ | $$-$$$ | $$ | $$$ | $$$ | Free |
| Privacy | Cloud | Cloud | Cloud | Cloud | Cloud | Local |
| Offline | No | No | No | No | No | Yes |

### Cost Comparison

Approximate costs per 1M tokens (input/output):

| Provider | Model | Input | Output |
|----------|-------|-------|--------|
| Anthropic | Claude 3.5 Sonnet | $3 | $15 |
| Anthropic | Claude 3 Opus | $15 | $75 |
| Anthropic | Claude 3 Haiku | $0.25 | $1.25 |
| OpenAI | GPT-4 Turbo | $10 | $30 |
| OpenAI | GPT-3.5 Turbo | $0.50 | $1.50 |
| Google | Gemini Pro | $0.50 | $1.50 |
| Ollama | Any | Free | Free |

### Use Case Recommendations

**For Code Generation:**
- Primary: Anthropic Claude 3.5 Sonnet
- Alternative: OpenAI GPT-4 Turbo
- Budget: Anthropic Claude 3 Haiku or Ollama CodeLlama

**For Code Review:**
- Primary: Anthropic Claude 3 Opus
- Alternative: OpenAI GPT-4

**For Quick Questions:**
- Primary: Anthropic Claude 3 Haiku
- Alternative: OpenAI GPT-3.5 Turbo or Ollama Mistral

**For Privacy/Offline:**
- Only option: Ollama (DeepSeek-Coder recommended)

**For Enterprise:**
- AWS users: AWS Bedrock
- Azure users: Azure OpenAI
- GCP users: Google Gemini (Vertex AI)

## Fallback Configuration

Configure automatic fallback between providers for reliability.

### Basic Fallback

```yaml
llm:
  default_provider: anthropic

  anthropic:
    api_key: "${ANTHROPIC_API_KEY}"

  openai:
    api_key: "${OPENAI_API_KEY}"

  fallback:
    enabled: true
    providers:
      - anthropic
      - openai
    max_retries: 2
    retry_delay: 5s
```

### Multi-Tier Fallback

```yaml
llm:
  fallback:
    enabled: true
    providers:
      - anthropic      # Try first (best quality)
      - openai         # Then this (reliable)
      - google         # Then this (cost-effective)
      - ollama         # Finally this (free, offline)
    max_retries: 3
    retry_delay: 3s
```

### Fallback Behavior

The system will automatically:

1. Try the primary provider
2. On failure, try the next provider
3. Continue through the list
4. Return error if all providers fail
5. Log fallback events for monitoring

## Best Practices

### 1. Choose the Right Model

Match the model to your task:

```yaml
# For complex tasks
model: claude-3-opus-20240229

# For balanced use
model: claude-3-5-sonnet-20241022

# For quick responses
model: claude-3-haiku-20240307
```

### 2. Optimize Token Usage

```yaml
llm:
  anthropic:
    max_tokens: 2048  # Reduce for simple tasks
    temperature: 0.3  # Lower for more focused responses
```

### 3. Use Environment Variables for Keys

```bash
# .env
ANTHROPIC_API_KEY=sk-ant-...
OPENAI_API_KEY=sk-...
```

Never commit API keys to version control!

### 4. Configure Timeouts

```yaml
llm:
  anthropic:
    timeout: 300s  # 5 minutes
    retry_attempts: 3
```

### 5. Enable Fallbacks

Always configure fallback providers for reliability:

```yaml
llm:
  fallback:
    enabled: true
    providers:
      - anthropic
      - openai
```

### 6. Monitor Usage

Track your API usage:

```bash
# Enable verbose logging
ainative-code --verbose chat

# Check provider usage
ainative-code stats providers
```

### 7. Use Appropriate Temperature

```yaml
# For code generation (more deterministic)
temperature: 0.3

# For creative tasks
temperature: 0.7

# For very creative tasks
temperature: 0.9
```

### 8. Local Development

Use Ollama for local development:

```yaml
# Development config
llm:
  default_provider: ollama
  ollama:
    model: codellama
```

Switch to cloud for production:

```yaml
# Production config
llm:
  default_provider: anthropic
```

### 9. Cost Management

Monitor and control costs:

```bash
# Use cheaper models for simple tasks
ainative-code chat --model claude-3-haiku "Quick question"

# Use expensive models only when needed
ainative-code chat --model claude-3-opus "Complex architecture design"
```

### 10. Security

- Use environment variables for API keys
- Enable config encryption for sensitive deployments
- Rotate API keys regularly
- Use least-privilege API keys when possible

## Troubleshooting

### Authentication Errors

```bash
# Verify API key is set
echo $ANTHROPIC_API_KEY

# Test API key
curl https://api.anthropic.com/v1/messages \
  -H "x-api-key: $ANTHROPIC_API_KEY" \
  -H "content-type: application/json"
```

### Rate Limiting

```yaml
# Add retry logic
llm:
  anthropic:
    retry_attempts: 5
    timeout: 600s
```

### Slow Responses

- Use faster models (e.g., Claude Haiku, GPT-3.5)
- Reduce max_tokens
- Enable streaming
- Check network connectivity

### Provider Failures

- Check provider status pages
- Enable fallback configuration
- Verify API keys are valid
- Check rate limits

## Next Steps

- [Configuration Guide](configuration.md) - Detailed configuration options
- [Getting Started](getting-started.md) - Start using AINative Code
- [Authentication Guide](authentication.md) - Platform authentication
- [Troubleshooting Guide](troubleshooting.md) - Common issues
