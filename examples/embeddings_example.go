package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"

	"github.com/AINative-studio/ainative-code/internal/embeddings"
)

func main() {
	// Example 1: Basic Embeddings
	basicEmbeddingsExample()

	// Example 2: Batch Embeddings
	batchEmbeddingsExample()

	// Example 3: Similarity Search
	similaritySearchExample()
}

// basicEmbeddingsExample demonstrates basic embedding generation
func basicEmbeddingsExample() {
	fmt.Println("\n=== Example 1: Basic Embeddings ===")

	// IMPORTANT: We use AINative platform for ALL embeddings
	// DO NOT use OpenAI embeddings endpoint
	client, err := embeddings.NewAINativeEmbeddingsClient(embeddings.Config{
		APIKey: os.Getenv("AINATIVE_API_KEY"),
	})
	if err != nil {
		log.Fatalf("Failed to create embeddings client: %v", err)
	}
	defer client.Close()

	// Generate embedding for a single text
	ctx := context.Background()
	result, err := client.Embed(ctx, []string{"Hello, world!"}, "default")
	if err != nil {
		log.Fatalf("Embedding failed: %v", err)
	}

	fmt.Printf("Model: %s\n", result.Model)
	fmt.Printf("Tokens used: %d\n", result.TotalTokens)
	fmt.Printf("Embedding dimension: %d\n", len(result.Embeddings[0]))
	fmt.Printf("First 5 values: %v\n", result.Embeddings[0][:5])
}

// batchEmbeddingsExample demonstrates batch embedding generation
func batchEmbeddingsExample() {
	fmt.Println("\n=== Example 2: Batch Embeddings ===")

	client, err := embeddings.NewAINativeEmbeddingsClient(embeddings.Config{
		APIKey: os.Getenv("AINATIVE_API_KEY"),
	})
	if err != nil {
		log.Fatalf("Failed to create embeddings client: %v", err)
	}
	defer client.Close()

	// Generate embeddings for multiple texts
	texts := []string{
		"Machine learning is fascinating",
		"Artificial intelligence is the future",
		"Deep learning models are powerful",
		"Natural language processing is useful",
	}

	ctx := context.Background()
	result, err := client.Embed(ctx, texts, "default")
	if err != nil {
		log.Fatalf("Batch embedding failed: %v", err)
	}

	fmt.Printf("Generated %d embeddings\n", len(result.Embeddings))
	fmt.Printf("Total tokens: %d\n", result.TotalTokens)

	// Display embedding info for each text
	for i, text := range texts {
		fmt.Printf("\nText %d: \"%s\"\n", i+1, text)
		fmt.Printf("  Embedding dimension: %d\n", len(result.Embeddings[i]))
		fmt.Printf("  First 3 values: %v\n", result.Embeddings[i][:3])
	}
}

// similaritySearchExample demonstrates using embeddings for similarity
func similaritySearchExample() {
	fmt.Println("\n=== Example 3: Similarity Search ===")

	client, err := embeddings.NewAINativeEmbeddingsClient(embeddings.Config{
		APIKey: os.Getenv("AINATIVE_API_KEY"),
	})
	if err != nil {
		log.Fatalf("Failed to create embeddings client: %v", err)
	}
	defer client.Close()

	// Documents to search
	documents := []string{
		"The cat sat on the mat",
		"Dogs are loyal animals",
		"Python is a programming language",
		"JavaScript is used for web development",
		"The kitten played with a ball",
	}

	// Query
	query := "Information about cats"

	ctx := context.Background()

	// Embed all documents
	fmt.Println("Embedding documents...")
	docResult, err := client.Embed(ctx, documents, "default")
	if err != nil {
		log.Fatalf("Document embedding failed: %v", err)
	}

	// Embed query
	fmt.Println("Embedding query...")
	queryResult, err := client.Embed(ctx, []string{query}, "default")
	if err != nil {
		log.Fatalf("Query embedding failed: %v", err)
	}

	queryEmbedding := queryResult.Embeddings[0]

	// Calculate cosine similarity for each document
	fmt.Printf("\nQuery: \"%s\"\n\n", query)
	fmt.Println("Similarity scores:")

	type docScore struct {
		index int
		score float32
	}

	var scores []docScore
	for i, docEmbedding := range docResult.Embeddings {
		similarity := cosineSimilarity(queryEmbedding, docEmbedding)
		scores = append(scores, docScore{index: i, score: similarity})
		fmt.Printf("  Document %d: %.4f - \"%s\"\n", i+1, similarity, documents[i])
	}

	// Find most similar document
	var maxScore float32
	var maxIndex int
	for _, s := range scores {
		if s.score > maxScore {
			maxScore = s.score
			maxIndex = s.index
		}
	}

	fmt.Printf("\nMost similar document (score: %.4f):\n", maxScore)
	fmt.Printf("  \"%s\"\n", documents[maxIndex])
}

// cosineSimilarity calculates the cosine similarity between two vectors
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

// Example: Error Handling for Embeddings
func embeddingsErrorHandlingExample() {
	fmt.Println("\n=== Example 4: Error Handling ===")

	client, err := embeddings.NewAINativeEmbeddingsClient(embeddings.Config{
		APIKey: os.Getenv("AINATIVE_API_KEY"),
	})
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Example 1: Empty texts
	_, err = client.Embed(ctx, []string{}, "default")
	if err != nil {
		fmt.Printf("Empty texts error: %v\n", err)
	}

	// Example 2: Too many texts
	manyTexts := make([]string, 101) // Max is 100
	for i := range manyTexts {
		manyTexts[i] = fmt.Sprintf("Text %d", i)
	}
	_, err = client.Embed(ctx, manyTexts, "default")
	if err != nil {
		fmt.Printf("Batch size error: %v\n", err)
	}

	// Example 3: Check error types
	_, err = client.Embed(ctx, []string{"test"}, "default")
	if err != nil {
		if apiErr, ok := err.(*embeddings.EmbeddingAPIError); ok {
			if apiErr.IsAuthenticationError() {
				fmt.Println("Authentication error: Check your API key")
			} else if apiErr.IsRateLimitError() {
				fmt.Println("Rate limit error: Please retry later")
			} else if apiErr.IsQuotaExceededError() {
				fmt.Println("Quota exceeded: Upgrade your plan")
			}
		}
	}
}
