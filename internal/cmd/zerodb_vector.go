package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/AINative-studio/ainative-code/internal/client/zerodb"
)

var (
	// Vector create-collection flags
	vectorCollectionName       string
	vectorCollectionDimensions int
	vectorCollectionMetric     string

	// Vector insert flags
	vectorInsertCollection string
	vectorInsertVector     string
	vectorInsertMetadata   string
	vectorInsertID         string

	// Vector search flags
	vectorSearchCollection string
	vectorSearchVector     string
	vectorSearchLimit      int
	vectorSearchFilter     string

	// Vector delete flags
	vectorDeleteCollection string
	vectorDeleteVectorID   string

	// Vector output flags
	vectorOutputJSON bool
)

// zerodbVectorCmd represents the zerodb vector command
var zerodbVectorCmd = &cobra.Command{
	Use:   "vector",
	Short: "Manage ZeroDB vector operations",
	Long: `Manage ZeroDB vector database operations for storing and searching embeddings.

ZeroDB Vector provides high-performance vector similarity search with support
for various distance metrics and metadata filtering.

Examples:
  # Create a vector collection
  ainative-code zerodb vector create-collection --name embeddings --dimensions 1536

  # Insert a vector
  ainative-code zerodb vector insert --collection embeddings --vector '[0.1,0.2,0.3]' --metadata '{"text":"hello"}'

  # Search for similar vectors
  ainative-code zerodb vector search --collection embeddings --query-vector '[0.1,0.2,0.3]' --limit 5

  # Delete a vector
  ainative-code zerodb vector delete --collection embeddings --id vec_123

  # List all collections
  ainative-code zerodb vector list-collections`,
	Aliases: []string{"vec", "vectors"},
}

// zerodbVectorCreateCollectionCmd represents the create-collection command
var zerodbVectorCreateCollectionCmd = &cobra.Command{
	Use:   "create-collection",
	Short: "Create a new vector collection",
	Long: `Create a new vector collection with the specified dimensions.

The collection stores vector embeddings with support for similarity search.
You can specify the distance metric for similarity calculations.

Supported metrics:
  - cosine: Cosine similarity (default, range 0-1, higher is more similar)
  - euclidean: Euclidean distance (L2 norm, lower is more similar)
  - dot_product: Dot product similarity (higher is more similar)

Examples:
  # Create collection with default cosine metric
  ainative-code zerodb vector create-collection --name embeddings --dimensions 1536

  # Create collection with euclidean metric
  ainative-code zerodb vector create-collection --name image_vectors --dimensions 512 --metric euclidean

  # Create collection with dot product metric
  ainative-code zerodb vector create-collection --name text_embeddings --dimensions 768 --metric dot_product`,
	RunE: runVectorCreateCollection,
}

// zerodbVectorInsertCmd represents the insert command
var zerodbVectorInsertCmd = &cobra.Command{
	Use:   "insert",
	Short: "Insert a vector into a collection",
	Long: `Insert a new vector embedding into the specified collection.

The vector must match the dimensionality of the collection.
You can optionally provide metadata as a JSON object and a custom ID.

Examples:
  # Insert a vector
  ainative-code zerodb vector insert --collection embeddings --vector '[0.1,0.2,0.3,0.4]'

  # Insert with metadata
  ainative-code zerodb vector insert --collection embeddings --vector '[0.1,0.2,0.3]' --metadata '{"text":"hello world","source":"doc1"}'

  # Insert with custom ID (upsert behavior)
  ainative-code zerodb vector insert --collection embeddings --vector '[0.1,0.2,0.3]' --id custom_vec_123`,
	RunE: runVectorInsert,
}

// zerodbVectorSearchCmd represents the search command
var zerodbVectorSearchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search for similar vectors",
	Long: `Search for vectors similar to the query vector using similarity search.

Returns vectors ranked by similarity score according to the collection's
distance metric. You can optionally filter results by metadata.

Examples:
  # Basic search
  ainative-code zerodb vector search --collection embeddings --query-vector '[0.1,0.2,0.3]'

  # Search with limit
  ainative-code zerodb vector search --collection embeddings --query-vector '[0.1,0.2,0.3]' --limit 10

  # Search with metadata filter
  ainative-code zerodb vector search --collection embeddings --query-vector '[0.1,0.2,0.3]' --filter '{"source":"doc1"}'

  # JSON output
  ainative-code zerodb vector search --collection embeddings --query-vector '[0.1,0.2,0.3]' --json`,
	RunE: runVectorSearch,
}

// zerodbVectorDeleteCmd represents the delete command
var zerodbVectorDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a vector from a collection",
	Long: `Delete a vector from the specified collection by its ID.

This operation is permanent and cannot be undone.

Examples:
  # Delete a vector
  ainative-code zerodb vector delete --collection embeddings --id vec_123`,
	RunE: runVectorDelete,
}

// zerodbVectorListCollectionsCmd represents the list-collections command
var zerodbVectorListCollectionsCmd = &cobra.Command{
	Use:   "list-collections",
	Short: "List all vector collections",
	Long: `List all vector collections in the current project.

Displays collection name, ID, dimensions, metric, and vector count.

Examples:
  # List all collections
  ainative-code zerodb vector list-collections

  # List with JSON output
  ainative-code zerodb vector list-collections --json`,
	RunE: runVectorListCollections,
}

func init() {
	zerodbCmd.AddCommand(zerodbVectorCmd)

	// Add vector subcommands
	zerodbVectorCmd.AddCommand(zerodbVectorCreateCollectionCmd)
	zerodbVectorCmd.AddCommand(zerodbVectorInsertCmd)
	zerodbVectorCmd.AddCommand(zerodbVectorSearchCmd)
	zerodbVectorCmd.AddCommand(zerodbVectorDeleteCmd)
	zerodbVectorCmd.AddCommand(zerodbVectorListCollectionsCmd)

	// Create-collection flags
	zerodbVectorCreateCollectionCmd.Flags().StringVar(&vectorCollectionName, "name", "", "collection name (required)")
	zerodbVectorCreateCollectionCmd.Flags().IntVar(&vectorCollectionDimensions, "dimensions", 0, "vector dimensions (required)")
	zerodbVectorCreateCollectionCmd.Flags().StringVar(&vectorCollectionMetric, "metric", "cosine", "distance metric (cosine, euclidean, dot_product)")
	zerodbVectorCreateCollectionCmd.MarkFlagRequired("name")
	zerodbVectorCreateCollectionCmd.MarkFlagRequired("dimensions")

	// Insert flags
	zerodbVectorInsertCmd.Flags().StringVar(&vectorInsertCollection, "collection", "", "collection name (required)")
	zerodbVectorInsertCmd.Flags().StringVar(&vectorInsertVector, "vector", "", "vector as JSON array (required)")
	zerodbVectorInsertCmd.Flags().StringVar(&vectorInsertMetadata, "metadata", "", "metadata as JSON object")
	zerodbVectorInsertCmd.Flags().StringVar(&vectorInsertID, "id", "", "custom vector ID (optional, for upsert)")
	zerodbVectorInsertCmd.MarkFlagRequired("collection")
	zerodbVectorInsertCmd.MarkFlagRequired("vector")

	// Search flags
	zerodbVectorSearchCmd.Flags().StringVar(&vectorSearchCollection, "collection", "", "collection name (required)")
	zerodbVectorSearchCmd.Flags().StringVar(&vectorSearchVector, "query-vector", "", "query vector as JSON array (required)")
	zerodbVectorSearchCmd.Flags().IntVar(&vectorSearchLimit, "limit", 10, "maximum number of results")
	zerodbVectorSearchCmd.Flags().StringVar(&vectorSearchFilter, "filter", "", "metadata filter as JSON object")
	zerodbVectorSearchCmd.MarkFlagRequired("collection")
	zerodbVectorSearchCmd.MarkFlagRequired("query-vector")

	// Delete flags
	zerodbVectorDeleteCmd.Flags().StringVar(&vectorDeleteCollection, "collection", "", "collection name (required)")
	zerodbVectorDeleteCmd.Flags().StringVar(&vectorDeleteVectorID, "id", "", "vector ID (required)")
	zerodbVectorDeleteCmd.MarkFlagRequired("collection")
	zerodbVectorDeleteCmd.MarkFlagRequired("id")

	// Global output flag for all vector commands
	zerodbVectorCmd.PersistentFlags().BoolVar(&vectorOutputJSON, "json", false, "output in JSON format")
}

func runVectorCreateCollection(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Validate dimensions
	if vectorCollectionDimensions <= 0 {
		return fmt.Errorf("dimensions must be greater than 0")
	}

	// Validate metric
	validMetrics := map[string]bool{
		"cosine":      true,
		"euclidean":   true,
		"dot_product": true,
	}
	if !validMetrics[vectorCollectionMetric] {
		return fmt.Errorf("invalid metric: %s (must be cosine, euclidean, or dot_product)", vectorCollectionMetric)
	}

	// Create ZeroDB client
	zdbClient, err := createZeroDBClient()
	if err != nil {
		return err
	}

	// Create collection
	collection, err := zdbClient.CreateCollection(ctx, vectorCollectionName, vectorCollectionDimensions, vectorCollectionMetric)
	if err != nil {
		return fmt.Errorf("failed to create collection: %w", err)
	}

	// Output result
	if vectorOutputJSON {
		return zerodbOutputJSON(collection)
	}

	fmt.Printf("Collection created successfully!\n")
	fmt.Printf("  ID:         %s\n", collection.ID)
	fmt.Printf("  Name:       %s\n", collection.Name)
	fmt.Printf("  Dimensions: %d\n", collection.Dimensions)
	fmt.Printf("  Metric:     %s\n", collection.Metric)
	fmt.Printf("  Created At: %s\n", collection.CreatedAt.Format(time.RFC3339))

	return nil
}

func runVectorInsert(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Parse vector JSON
	var vector []float64
	if err := json.Unmarshal([]byte(vectorInsertVector), &vector); err != nil {
		return fmt.Errorf("invalid vector JSON: %w", err)
	}

	if len(vector) == 0 {
		return fmt.Errorf("vector cannot be empty")
	}

	// Parse metadata JSON if provided
	var metadata map[string]interface{}
	if vectorInsertMetadata != "" {
		if err := json.Unmarshal([]byte(vectorInsertMetadata), &metadata); err != nil {
			return fmt.Errorf("invalid metadata JSON: %w", err)
		}
	}

	// Create ZeroDB client
	zdbClient, err := createZeroDBClient()
	if err != nil {
		return err
	}

	// Insert vector
	id, vec, err := zdbClient.InsertVector(ctx, vectorInsertCollection, vector, metadata, vectorInsertID)
	if err != nil {
		return fmt.Errorf("failed to insert vector: %w", err)
	}

	// Output result
	if vectorOutputJSON {
		return zerodbOutputJSON(vec)
	}

	fmt.Printf("Vector inserted successfully!\n")
	fmt.Printf("  ID:         %s\n", id)
	fmt.Printf("  Collection: %s\n", vectorInsertCollection)
	fmt.Printf("  Dimensions: %d\n", len(vector))
	fmt.Printf("  Created At: %s\n", vec.CreatedAt.Format(time.RFC3339))

	return nil
}

func runVectorSearch(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Parse query vector JSON
	var queryVector []float64
	if err := json.Unmarshal([]byte(vectorSearchVector), &queryVector); err != nil {
		return fmt.Errorf("invalid query vector JSON: %w", err)
	}

	if len(queryVector) == 0 {
		return fmt.Errorf("query vector cannot be empty")
	}

	// Parse filter JSON if provided
	var filter zerodb.QueryFilter
	if vectorSearchFilter != "" {
		if err := json.Unmarshal([]byte(vectorSearchFilter), &filter); err != nil {
			return fmt.Errorf("invalid filter JSON: %w", err)
		}
	}

	// Create ZeroDB client
	zdbClient, err := createZeroDBClient()
	if err != nil {
		return err
	}

	// Search vectors
	vectors, err := zdbClient.SearchVectors(ctx, vectorSearchCollection, queryVector, vectorSearchLimit, filter)
	if err != nil {
		return fmt.Errorf("failed to search vectors: %w", err)
	}

	// Output result
	if vectorOutputJSON {
		return zerodbOutputJSON(vectors)
	}

	if len(vectors) == 0 {
		fmt.Println("No similar vectors found.")
		return nil
	}

	fmt.Printf("Found %d similar vector(s):\n\n", len(vectors))
	for i, vec := range vectors {
		fmt.Printf("Result %d (score: %.4f):\n", i+1, vec.Score)
		fmt.Printf("  ID:         %s\n", vec.ID)
		fmt.Printf("  Dimensions: %d\n", len(vec.Vector))
		fmt.Printf("  Created At: %s\n", vec.CreatedAt.Format(time.RFC3339))

		if len(vec.Metadata) > 0 {
			fmt.Printf("  Metadata:\n")
			metaJSON, _ := json.MarshalIndent(vec.Metadata, "    ", "  ")
			fmt.Printf("    %s\n", string(metaJSON))
		}

		// Show first few dimensions of vector
		vecPreview := vec.Vector
		if len(vecPreview) > 5 {
			vecPreview = vecPreview[:5]
			fmt.Printf("  Vector:     [%s...] (%d dimensions)\n", formatFloats(vecPreview), len(vec.Vector))
		} else {
			fmt.Printf("  Vector:     %s\n", formatFloats(vecPreview))
		}
		fmt.Println()
	}

	return nil
}

func runVectorDelete(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Create ZeroDB client
	zdbClient, err := createZeroDBClient()
	if err != nil {
		return err
	}

	// Delete vector
	if err := zdbClient.DeleteVector(ctx, vectorDeleteCollection, vectorDeleteVectorID); err != nil {
		return fmt.Errorf("failed to delete vector: %w", err)
	}

	// Output result
	if vectorOutputJSON {
		return zerodbOutputJSON(map[string]interface{}{
			"success":    true,
			"id":         vectorDeleteVectorID,
			"collection": vectorDeleteCollection,
		})
	}

	fmt.Printf("Vector deleted successfully!\n")
	fmt.Printf("  ID:         %s\n", vectorDeleteVectorID)
	fmt.Printf("  Collection: %s\n", vectorDeleteCollection)

	return nil
}

func runVectorListCollections(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Create ZeroDB client
	zdbClient, err := createZeroDBClient()
	if err != nil {
		return err
	}

	// List collections
	collections, err := zdbClient.ListCollections(ctx)
	if err != nil {
		return fmt.Errorf("failed to list collections: %w", err)
	}

	// Output result
	if vectorOutputJSON {
		return zerodbOutputJSON(collections)
	}

	if len(collections) == 0 {
		fmt.Println("No vector collections found.")
		return nil
	}

	// Create table writer for formatted output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "NAME\tDIMENSIONS\tMETRIC\tCOUNT\tCREATED AT")
	fmt.Fprintln(w, "----\t----------\t------\t-----\t----------")

	for _, collection := range collections {
		fmt.Fprintf(w, "%s\t%d\t%s\t%d\t%s\n",
			collection.Name,
			collection.Dimensions,
			collection.Metric,
			collection.Count,
			collection.CreatedAt.Format("2006-01-02 15:04:05"),
		)
	}

	w.Flush()

	fmt.Printf("\nTotal: %d collection(s)\n", len(collections))

	return nil
}

// formatFloats formats a slice of floats for display
func formatFloats(floats []float64) string {
	var parts []string
	for _, f := range floats {
		parts = append(parts, strconv.FormatFloat(f, 'f', 4, 64))
	}
	return strings.Join(parts, ", ")
}
