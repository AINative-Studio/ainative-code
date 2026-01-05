# AINative Embeddings Documentation

## Overview

The AINative Embeddings API provides vector embeddings for text, enabling semantic search, similarity matching, and other vector-based operations.

**IMPORTANT**: This is the ONLY way to obtain embeddings in the AINative-Code platform. We do NOT use OpenAI's embeddings endpoint or any other third-party embedding service directly.

## Why AINative Embeddings?

- **Unified API**: Single endpoint for all embedding needs
- **Cost-effective**: Optimized pricing and quota management
- **Consistent**: Same API regardless of underlying model
- **Secure**: Centralized authentication and access control
- **Normalized vectors**: Always returns normalized vectors for cosine similarity

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/AINative-studio/ainative-code/internal/embeddings"
)

func main() {
    // Create embeddings client
    client, err := embeddings.NewAINativeEmbeddingsClient(embeddings.Config{
        APIKey: "your-ainative-api-key",
    })
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Generate embeddings
    ctx := context.Background()
    result, err := client.Embed(ctx, []string{"Hello, world!"}, "default")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Generated %d embeddings\n", len(result.Embeddings))
    fmt.Printf("Dimension: %d\n", len(result.Embeddings[0]))
}
```

## Configuration

### Basic Configuration

```go
config := embeddings.Config{
    APIKey: "ainative-key-...",  // Required
}
```

### Advanced Configuration

```go
config := embeddings.Config{
    APIKey:     "ainative-key-...",              // Required
    Endpoint:   "https://custom.api/embeddings", // Optional: Custom endpoint
    HTTPClient: customHTTPClient,                // Optional: Custom HTTP client
    Logger:     customLogger,                    // Optional: Logger
    MaxRetries: 5,                               // Optional: Max retry attempts (default: 3)
    RetryDelay: 2 * time.Second,                 // Optional: Retry delay (default: 1s)
}
```

## API Reference

### Embed Method

Generate embeddings for one or more texts:

```go
result, err := client.Embed(ctx context.Context, texts []string, model string) (*EmbeddingResult, error)
```

**Parameters**:
- `ctx`: Context for cancellation and timeouts
- `texts`: Array of strings to embed (max 100 per request)
- `model`: Embedding model to use (use "default" for platform default)

**Returns**:
- `EmbeddingResult`: Contains embeddings, model used, and token count
- `error`: Error if request failed

### EmbeddingResult Structure

```go
type EmbeddingResult struct {
    Embeddings  [][]float32  // Vector embeddings (one per input text)
    Model       string       // Model used for embedding
    TotalTokens int          // Total tokens processed
}
```

## Batch Processing

You can embed up to 100 texts in a single request:

```go
texts := []string{
    "First document",
    "Second document",
    "Third document",
    // ... up to 100 total
}

result, err := client.Embed(ctx, texts, "default")
if err != nil {
    log.Fatal(err)
}

// Process each embedding
for i, embedding := range result.Embeddings {
    fmt.Printf("Document %d: %d dimensions\n", i, len(embedding))
}
```

## Use Cases

### 1. Semantic Search

```go
// Embed documents
documents := []string{
    "Machine learning tutorial",
    "Python programming guide",
    "Database optimization tips",
}

docResult, _ := client.Embed(ctx, documents, "default")

// Embed query
queryResult, _ := client.Embed(ctx, []string{"How to learn ML?"}, "default")

// Calculate similarities
for i, docEmb := range docResult.Embeddings {
    similarity := cosineSimilarity(queryResult.Embeddings[0], docEmb)
    fmt.Printf("Document %d similarity: %.4f\n", i, similarity)
}
```

### 2. Document Clustering

```go
// Embed multiple documents
texts := loadDocuments()
result, _ := client.Embed(ctx, texts, "default")

// Use embeddings for clustering (e.g., k-means)
clusters := performClustering(result.Embeddings)
```

### 3. Similarity Detection

```go
// Check if two texts are similar
text1 := "The cat sat on the mat"
text2 := "A feline rested on the rug"

result, _ := client.Embed(ctx, []string{text1, text2}, "default")

similarity := cosineSimilarity(result.Embeddings[0], result.Embeddings[1])
fmt.Printf("Similarity: %.4f\n", similarity)
```

## Error Handling

### Error Types

The embeddings API returns `EmbeddingAPIError` with specific methods:

```go
if err != nil {
    if apiErr, ok := err.(*embeddings.EmbeddingAPIError); ok {
        if apiErr.IsAuthenticationError() {
            // Handle auth error
        } else if apiErr.IsRateLimitError() {
            // Handle rate limit
        } else if apiErr.IsQuotaExceededError() {
            // Handle quota exceeded
        }
    }
}
```

### Error Methods

- `IsAuthenticationError()`: Returns true for 401/403 errors
- `IsRateLimitError()`: Returns true for 429 rate limit errors
- `IsQuotaExceededError()`: Returns true when quota is exhausted

### Example Error Handling

```go
result, err := client.Embed(ctx, texts, "default")
if err != nil {
    if apiErr, ok := err.(*embeddings.EmbeddingAPIError); ok {
        switch {
        case apiErr.IsAuthenticationError():
            log.Println("Invalid API key - check your credentials")
        case apiErr.IsRateLimitError():
            log.Println("Rate limited - retry after delay")
        case apiErr.IsQuotaExceededError():
            log.Println("Quota exceeded - upgrade plan or wait for reset")
        default:
            log.Printf("API error: %s", apiErr.Message)
        }
    } else {
        log.Printf("Request error: %v", err)
    }
    return
}
```

## Retry Logic

The client automatically retries failed requests:

```go
client, _ := embeddings.NewAINativeEmbeddingsClient(embeddings.Config{
    APIKey:     "key",
    MaxRetries: 5,                    // Retry up to 5 times
    RetryDelay: 2 * time.Second,      // 2s delay between retries
})
```

**Retry Behavior**:
- Retries on server errors (500, 502, 503, 504)
- Retries on rate limits (429)
- Does NOT retry on client errors (400, 401, 403, 404)
- Uses exponential backoff with retry delay

## Cosine Similarity

Helper function for calculating similarity between embeddings:

```go
func cosineSimilarity(a, b []float32) float32 {
    if len(a) != len(b) {
        return 0
    }

    var dotProduct, normA, normB float64
    for i := range a {
        dotProduct += float64(a[i]) * float64(b[i])
        normA += float64(a[i]) * float64(a[i])
        normB += float64(b[i]) * float64(b[i])
    }

    if normA == 0 || normB == 0 {
        return 0
    }

    return float32(dotProduct / (math.Sqrt(normA) * math.Sqrt(normB)))
}
```

The AINative API returns normalized vectors, so you can also use dot product directly for similarity.

## Best Practices

1. **Batch requests**: Embed multiple texts at once (up to 100) to reduce API calls
2. **Cache embeddings**: Store embeddings to avoid re-computing for same text
3. **Use appropriate context**: Set reasonable timeouts in context
4. **Handle errors**: Implement retry logic for transient failures
5. **Monitor usage**: Track token consumption to manage costs
6. **Normalize on client side**: Although vectors are normalized, verify before similarity calculations

## Limits and Quotas

- **Maximum batch size**: 100 texts per request
- **Rate limits**: Enforced by the platform (see API error responses)
- **Token limits**: Based on your plan and quota
- **Timeout**: 30 seconds default (configurable via HTTP client)

## Testing

The embeddings client includes comprehensive tests:

- **Unit tests**: `internal/embeddings/ainative_test.go`
- **Coverage**: 91.3%+ code coverage

Run tests:

```bash
go test ./internal/embeddings/...

# With coverage
go test -cover ./internal/embeddings/...
```

## Integration with Vector Databases

Embeddings can be stored in vector databases for efficient similarity search:

```go
// Generate embeddings
result, _ := client.Embed(ctx, documents, "default")

// Store in vector database (pseudo-code)
for i, embedding := range result.Embeddings {
    vectorDB.Insert(documents[i], embedding)
}

// Query
queryEmb, _ := client.Embed(ctx, []string{query}, "default")
results := vectorDB.Search(queryEmb.Embeddings[0], topK=10)
```

## Examples

See `examples/embeddings_example.go` for complete working examples including:

- Basic embeddings
- Batch processing
- Similarity search
- Error handling

## FAQ

### Q: Why not use OpenAI embeddings directly?

A: The AINative platform provides a unified embeddings API that:
- Centralizes quota and cost management
- Provides consistent authentication
- Enables switching between embedding models transparently
- Offers better monitoring and analytics

### Q: What embedding models are available?

A: Use `"default"` as the model parameter. The platform will use the optimal embedding model for your use case. Specific model selection may be added in future updates.

### Q: Are vectors normalized?

A: Yes, all embeddings are normalized by default. You can use dot product for cosine similarity calculations.

### Q: What's the embedding dimension?

A: The dimension depends on the underlying model. Check `len(result.Embeddings[0])` to get the actual dimension.

## Support

For issues or questions about embeddings:

- Review the examples in `examples/embeddings_example.go`
- Check the unit tests for usage patterns
- Consult the AINative platform documentation
- Contact platform support for quota or billing questions
