# AWS Bedrock Provider

The AWS Bedrock provider enables access to Claude and other AI models through Amazon Bedrock. This provider implements AWS Signature Version 4 authentication and supports both streaming and non-streaming requests.

## Features

- Full AWS Signature V4 authentication
- Support for multiple Claude models on Bedrock
- Streaming and non-streaming responses
- Comprehensive error handling
- AWS credentials from environment or explicit configuration
- IAM role support via session tokens
- Multiple AWS regions

## Supported Models

The following Claude models are supported on AWS Bedrock:

- `anthropic.claude-3-5-sonnet-20241022-v2:0` - Latest Claude 3.5 Sonnet
- `anthropic.claude-3-opus-20240229-v1:0` - Claude 3 Opus
- `anthropic.claude-3-sonnet-20240229-v1:0` - Claude 3 Sonnet
- `anthropic.claude-3-haiku-20240307-v1:0` - Claude 3 Haiku (fastest)
- `anthropic.claude-v2` - Claude 2
- `anthropic.claude-instant-v1` - Claude Instant

## Installation

```go
import (
    "github.com/AINative-studio/ainative-code/internal/provider"
    "github.com/AINative-studio/ainative-code/internal/provider/bedrock"
)
```

## Configuration

### Basic Configuration

```go
config := bedrock.Config{
    Region:    "us-east-1",
    AccessKey: "YOUR_AWS_ACCESS_KEY",
    SecretKey: "YOUR_AWS_SECRET_KEY",
}

provider, err := bedrock.NewBedrockProvider(config)
if err != nil {
    log.Fatal(err)
}
defer provider.Close()
```

### Using Environment Variables

The provider can automatically load credentials from environment variables:

```bash
export AWS_ACCESS_KEY_ID="YOUR_ACCESS_KEY"
export AWS_SECRET_ACCESS_KEY="YOUR_SECRET_KEY"
export AWS_REGION="us-east-1"
export AWS_SESSION_TOKEN="YOUR_SESSION_TOKEN"  # Optional, for temporary credentials
```

```go
config := bedrock.Config{}
config.MergeWithEnvironment()

provider, err := bedrock.NewBedrockProvider(config)
```

### Using IAM Roles with Session Tokens

For temporary credentials from IAM roles:

```go
config := bedrock.Config{
    Region:       "us-east-1",
    AccessKey:    "TEMPORARY_ACCESS_KEY",
    SecretKey:    "TEMPORARY_SECRET_KEY",
    SessionToken: "SESSION_TOKEN",
}

provider, err := bedrock.NewBedrockProvider(config)
```

### Custom Endpoint

For VPC endpoints or testing:

```go
config := bedrock.Config{
    Region:    "us-east-1",
    AccessKey: "YOUR_ACCESS_KEY",
    SecretKey: "YOUR_SECRET_KEY",
    Endpoint:  "https://vpce-1234567890abcdef0-abc12345.bedrock-runtime.us-east-1.vpce.amazonaws.com",
}

provider, err := bedrock.NewBedrockProvider(config)
```

## Usage Examples

### Basic Chat Request

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/AINative-studio/ainative-code/internal/provider"
    "github.com/AINative-studio/ainative-code/internal/provider/bedrock"
)

func main() {
    // Create provider
    config := bedrock.Config{
        Region:    "us-east-1",
        AccessKey: "YOUR_ACCESS_KEY",
        SecretKey: "YOUR_SECRET_KEY",
    }

    p, err := bedrock.NewBedrockProvider(config)
    if err != nil {
        log.Fatal(err)
    }
    defer p.Close()

    // Create messages
    messages := []provider.Message{
        {Role: "user", Content: "What is the capital of France?"},
    }

    // Send request
    ctx := context.Background()
    resp, err := p.Chat(ctx, messages,
        provider.WithModel("anthropic.claude-3-5-sonnet-20241022-v2:0"),
        provider.WithMaxTokens(1024),
        provider.WithTemperature(0.7),
    )

    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Response:", resp.Content)
    fmt.Printf("Tokens used: %d input, %d output\n",
        resp.Usage.PromptTokens,
        resp.Usage.CompletionTokens)
}
```

### Streaming Response

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/AINative-studio/ainative-code/internal/provider"
    "github.com/AINative-studio/ainative-code/internal/provider/bedrock"
)

func main() {
    // Create provider
    config := bedrock.Config{
        Region:    "us-east-1",
        AccessKey: "YOUR_ACCESS_KEY",
        SecretKey: "YOUR_SECRET_KEY",
    }

    p, err := bedrock.NewBedrockProvider(config)
    if err != nil {
        log.Fatal(err)
    }
    defer p.Close()

    // Create messages
    messages := []provider.Message{
        {Role: "user", Content: "Write a short poem about AI."},
    }

    // Start streaming
    ctx := context.Background()
    eventChan, err := p.Stream(ctx, messages,
        provider.StreamWithModel("anthropic.claude-3-haiku-20240307-v1:0"),
        provider.StreamWithMaxTokens(500),
    )

    if err != nil {
        log.Fatal(err)
    }

    // Process events
    fmt.Print("Response: ")
    for event := range eventChan {
        switch event.Type {
        case provider.EventTypeContentDelta:
            fmt.Print(event.Content)
        case provider.EventTypeError:
            log.Fatal(event.Error)
        case provider.EventTypeContentEnd:
            fmt.Println()
        }
    }
}
```

### With System Prompt

```go
messages := []provider.Message{
    {
        Role:    "system",
        Content: "You are a helpful assistant that explains complex topics simply.",
    },
    {
        Role:    "user",
        Content: "Explain quantum computing.",
    },
}

resp, err := p.Chat(ctx, messages,
    provider.WithModel("anthropic.claude-3-opus-20240229-v1:0"),
    provider.WithMaxTokens(2048),
)
```

### Conversation with Multiple Messages

```go
messages := []provider.Message{
    {Role: "user", Content: "Hello! I need help with Python."},
    {Role: "assistant", Content: "I'd be happy to help with Python! What would you like to know?"},
    {Role: "user", Content: "How do I read a CSV file?"},
}

resp, err := p.Chat(ctx, messages,
    provider.WithModel("anthropic.claude-3-sonnet-20240229-v1:0"),
    provider.WithMaxTokens(1024),
)
```

### Advanced Options

```go
resp, err := p.Chat(ctx, messages,
    provider.WithModel("anthropic.claude-3-5-sonnet-20241022-v2:0"),
    provider.WithMaxTokens(2048),
    provider.WithTemperature(0.8),
    provider.WithTopP(0.95),
    provider.WithStopSequences("END", "STOP"),
    provider.WithSystemPrompt("You are an expert programmer."),
)
```

## Error Handling

The provider includes comprehensive error handling for common AWS Bedrock errors:

```go
resp, err := p.Chat(ctx, messages, provider.WithModel("anthropic.claude-3-5-sonnet-20241022-v2:0"))
if err != nil {
    var authErr *provider.AuthenticationError
    var rateLimitErr *provider.RateLimitError
    var contextErr *provider.ContextLengthError
    var providerErr *provider.ProviderError

    switch {
    case errors.As(err, &authErr):
        // Handle authentication error
        log.Printf("Authentication failed: %v", err)

    case errors.As(err, &rateLimitErr):
        // Handle rate limiting
        log.Printf("Rate limited, retry after %d seconds", rateLimitErr.RetryAfter)

    case errors.As(err, &contextErr):
        // Handle context length exceeded
        log.Printf("Context too long: %d tokens requested, %d max",
            contextErr.RequestTokens, contextErr.MaxTokens)

    case errors.As(err, &providerErr):
        // Handle other provider errors
        log.Printf("Provider error: %v", err)

    default:
        log.Printf("Unknown error: %v", err)
    }
}
```

## AWS Credentials Configuration

### Option 1: Explicit Configuration

```go
config := bedrock.Config{
    Region:    "us-east-1",
    AccessKey: "AKIAIOSFODNN7EXAMPLE",
    SecretKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
}
```

### Option 2: Environment Variables

```bash
export AWS_ACCESS_KEY_ID="AKIAIOSFODNN7EXAMPLE"
export AWS_SECRET_ACCESS_KEY="wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
export AWS_REGION="us-east-1"
```

### Option 3: IAM Role with Temporary Credentials

```bash
export AWS_ACCESS_KEY_ID="ASIAIOSFODNN7EXAMPLE"
export AWS_SECRET_ACCESS_KEY="wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
export AWS_SESSION_TOKEN="AQoDYXdzEJr..."
export AWS_REGION="us-east-1"
```

### Option 4: AWS Credentials File

While the provider doesn't automatically read `~/.aws/credentials`, you can implement this yourself:

```go
import (
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/credentials"
)

// Load AWS config
awsConfig, err := config.LoadDefaultConfig(context.TODO())
if err != nil {
    log.Fatal(err)
}

// Get credentials
creds, err := awsConfig.Credentials.Retrieve(context.TODO())
if err != nil {
    log.Fatal(err)
}

// Create Bedrock provider
bedrockConfig := bedrock.Config{
    Region:       awsConfig.Region,
    AccessKey:    creds.AccessKeyID,
    SecretKey:    creds.SecretAccessKey,
    SessionToken: creds.SessionToken,
}

provider, err := bedrock.NewBedrockProvider(bedrockConfig)
```

## AWS Regions

Supported AWS regions (as of 2024):

- `us-east-1` - US East (N. Virginia)
- `us-west-2` - US West (Oregon)
- `ap-southeast-1` - Asia Pacific (Singapore)
- `ap-northeast-1` - Asia Pacific (Tokyo)
- `eu-central-1` - Europe (Frankfurt)
- `eu-west-1` - Europe (Ireland)
- `eu-west-3` - Europe (Paris)

## Model Selection Guide

### Claude 3.5 Sonnet (anthropic.claude-3-5-sonnet-20241022-v2:0)
- **Best for**: Complex reasoning, coding, analysis
- **Speed**: Fast
- **Context**: 200K tokens
- **Cost**: Medium

### Claude 3 Opus (anthropic.claude-3-opus-20240229-v1:0)
- **Best for**: Highly complex tasks, research, deep analysis
- **Speed**: Slower
- **Context**: 200K tokens
- **Cost**: Highest

### Claude 3 Haiku (anthropic.claude-3-haiku-20240307-v1:0)
- **Best for**: Simple tasks, quick responses, high volume
- **Speed**: Fastest
- **Context**: 200K tokens
- **Cost**: Lowest

### Claude 2 (anthropic.claude-v2)
- **Best for**: Legacy applications
- **Context**: 100K tokens
- **Cost**: Medium

## Embeddings

**IMPORTANT**: For vector embeddings and semantic search, use the **AINative platform APIs**, not AWS Bedrock embeddings. This ensures consistency across the platform and leverages AINative's optimized embedding infrastructure.

```go
// DO NOT use AWS Bedrock for embeddings
// Instead, use AINative platform APIs:
// - /zerodb-vector-upsert for storing embeddings
// - /zerodb-vector-search for semantic search
```

## Performance Optimization

### Connection Pooling

The provider uses HTTP connection pooling by default. For custom configuration:

```go
transport := &http.Transport{
    MaxIdleConns:        100,
    MaxIdleConnsPerHost: 10,
    IdleConnTimeout:     90 * time.Second,
}

httpClient := &http.Client{
    Transport: transport,
    Timeout:   60 * time.Second,
}

config := bedrock.Config{
    Region:     "us-east-1",
    AccessKey:  "YOUR_KEY",
    SecretKey:  "YOUR_SECRET",
    HTTPClient: httpClient,
}
```

### Context Timeouts

Always use context timeouts for production:

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

resp, err := provider.Chat(ctx, messages, provider.WithModel("..."))
```

### Retry Logic

The provider includes automatic retry logic for transient errors:
- 429 (Rate Limit): Retries with exponential backoff
- 500, 502, 503, 504: Retries with exponential backoff
- Maximum 3 retries by default

## Security Best Practices

1. **Never hardcode credentials**
   ```go
   // BAD
   config := bedrock.Config{
       AccessKey: "AKIAIOSFODNN7EXAMPLE",
       SecretKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
   }

   // GOOD
   config := bedrock.Config{
       AccessKey: os.Getenv("AWS_ACCESS_KEY_ID"),
       SecretKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
   }
   ```

2. **Use IAM roles when possible**
   - Prefer IAM roles over long-term access keys
   - Use temporary credentials with session tokens

3. **Apply least privilege**
   - Only grant `bedrock:InvokeModel` permissions
   - Restrict to specific model ARNs if needed

4. **Monitor usage**
   - Track token usage via response.Usage
   - Set up AWS CloudWatch alarms

## Testing

The implementation includes comprehensive tests:

```bash
# Run unit tests
go test ./internal/provider/bedrock/...

# Run with coverage
go test ./internal/provider/bedrock/... -coverprofile=coverage.out

# View coverage
go tool cover -html=coverage.out

# Run integration tests
go test ./tests/integration/bedrock_test.go -v
```

## Architecture

### Package Structure

```
internal/provider/bedrock/
├── client.go       # Main provider implementation
├── auth.go         # AWS Signature V4 authentication
├── config.go       # Configuration management
├── messages.go     # Message format conversion
├── streaming.go    # Streaming event handling
├── errors.go       # Error parsing and handling
└── *_test.go      # Comprehensive unit tests
```

### Authentication Flow

1. Request is created with JSON body
2. AWS Signature V4 signing:
   - Calculate payload hash (SHA256)
   - Create canonical request
   - Create string to sign
   - Calculate signature using HMAC-SHA256
   - Add Authorization header
3. Request is sent to Bedrock endpoint

### Message Flow

1. Provider messages → Bedrock format conversion
2. System prompts extracted and formatted separately
3. JSON serialization
4. AWS signing
5. HTTP request
6. Response parsing
7. Bedrock format → Provider format conversion

## Troubleshooting

### Authentication Errors

```
Error: authentication failed for provider "bedrock": The security token included in the request is invalid
```

**Solution**: Check your AWS credentials and ensure they have the correct permissions.

### Rate Limiting

```
Error: rate limit exceeded for provider "bedrock"
```

**Solution**: Implement exponential backoff or request a quota increase through AWS Support.

### Context Length Errors

```
Error: context length exceeded for provider "bedrock" model "anthropic.claude-3-5-sonnet-20241022-v2:0"
```

**Solution**: Reduce the message size or use a model with a larger context window.

### Region Errors

```
Error: Could not resolve the foundation model
```

**Solution**: Ensure the model is available in your selected region and that you have access enabled.

## License

This implementation is part of the AINative-Code project.

## Support

For issues and questions:
- GitHub Issues: https://github.com/AINative-studio/ainative-code/issues
- Documentation: https://ainative.studio/docs
