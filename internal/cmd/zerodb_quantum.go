package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/AINative-studio/ainative-code/internal/client/zerodb"
)

var (
	// Quantum entangle flags
	entangleVectorID1 string
	entangleVectorID2 string

	// Quantum measure flags
	measureVectorID string

	// Quantum compress flags
	compressVectorID    string
	compressionRatio    float64

	// Quantum decompress flags
	decompressVectorID string

	// Quantum search flags
	quantumSearchQueryVector     string
	quantumSearchLimit           int
	quantumSearchUseQuantumBoost bool
	quantumSearchIncludeEntangled bool

	// Quantum output flags
	quantumOutputJSON bool
)

// zerodbQuantumCmd represents the zerodb quantum command
var zerodbQuantumCmd = &cobra.Command{
	Use:   "quantum",
	Short: "Manage ZeroDB quantum-enhanced features",
	Long: `Manage ZeroDB quantum-enhanced vector operations including entanglement,
measurement, compression, and quantum-boosted search.

Quantum features leverage advanced techniques for enhanced vector similarity,
compression, and correlation analysis. These features are designed to improve
performance and capabilities for AI and machine learning workloads.

Examples:
  # Entangle two vectors
  ainative-code zerodb quantum entangle --vector-id-1 vec_123 --vector-id-2 vec_456

  # Measure quantum state
  ainative-code zerodb quantum measure --vector-id vec_123

  # Compress vector
  ainative-code zerodb quantum compress --vector-id vec_123 --compression-ratio 0.5

  # Decompress vector
  ainative-code zerodb quantum decompress --vector-id vec_123

  # Quantum-enhanced search
  ainative-code zerodb quantum search --query-vector '[0.1,0.2,0.3]' --limit 5`,
	Aliases: []string{"q"},
}

// zerodbQuantumEntangleCmd represents the zerodb quantum entangle command
var zerodbQuantumEntangleCmd = &cobra.Command{
	Use:   "entangle",
	Short: "Entangle two vectors to create quantum correlation",
	Long: `Entangle two vectors to create a quantum correlation between them.

Quantum entanglement creates a special correlation between two vectors, allowing
them to influence each other's search results and enabling advanced similarity
analysis. Entangled vectors are tracked together and can be used for relationship
discovery and enhanced search capabilities.

Use Cases:
- Creating semantic relationships between related concepts
- Linking similar documents or embeddings
- Building knowledge graphs with vector correlations
- Enhancing search results with related content

Examples:
  # Entangle two related concept vectors
  ainative-code zerodb quantum entangle --vector-id-1 vec_product_a --vector-id-2 vec_product_b

  # View entanglement details in JSON
  ainative-code zerodb quantum entangle --vector-id-1 vec_123 --vector-id-2 vec_456 --json`,
	RunE: runQuantumEntangle,
}

// zerodbQuantumMeasureCmd represents the zerodb quantum measure command
var zerodbQuantumMeasureCmd = &cobra.Command{
	Use:   "measure",
	Short: "Measure the quantum state of a vector",
	Long: `Measure the quantum state of a vector to analyze its properties.

Quantum measurement reveals the internal state and properties of a vector,
including entropy, coherence, and other quantum characteristics. This information
can be used to understand vector quality, optimize storage, and improve search
performance.

Measured Properties:
- Quantum State: The current quantum configuration
- Entropy: Measure of information content and randomness
- Coherence: Measure of quantum correlation strength
- Additional quantum properties

Examples:
  # Measure a vector's quantum state
  ainative-code zerodb quantum measure --vector-id vec_123

  # Measure with JSON output for programmatic analysis
  ainative-code zerodb quantum measure --vector-id vec_456 --json`,
	RunE: runQuantumMeasure,
}

// zerodbQuantumCompressCmd represents the zerodb quantum compress command
var zerodbQuantumCompressCmd = &cobra.Command{
	Use:   "compress",
	Short: "Compress a vector using quantum compression techniques",
	Long: `Compress a vector using quantum compression techniques to reduce dimensionality.

Quantum compression uses advanced algorithms to reduce vector dimensions while
preserving semantic meaning and similarity relationships. This can significantly
reduce storage costs and improve search performance with minimal information loss.

Compression Ratio:
- Value between 0 and 1 (e.g., 0.5 for 50% compression)
- Lower values = more compression = more information loss
- Higher values = less compression = better quality
- Recommended range: 0.4 - 0.7 for most use cases

Benefits:
- Reduced storage costs (proportional to compression ratio)
- Faster search performance (fewer dimensions to compare)
- Lower memory usage for large vector collections
- Maintained semantic relationships in most cases

Examples:
  # Compress to 50% of original size
  ainative-code zerodb quantum compress --vector-id vec_123 --compression-ratio 0.5

  # Aggressive compression (30% of original size)
  ainative-code zerodb quantum compress --vector-id vec_456 --compression-ratio 0.3

  # Conservative compression (70% of original size)
  ainative-code zerodb quantum compress --vector-id vec_789 --compression-ratio 0.7`,
	RunE: runQuantumCompress,
}

// zerodbQuantumDecompressCmd represents the zerodb quantum decompress command
var zerodbQuantumDecompressCmd = &cobra.Command{
	Use:   "decompress",
	Short: "Decompress a previously compressed vector",
	Long: `Decompress a previously compressed vector to restore original dimensions.

Quantum decompression attempts to restore a compressed vector to its original
dimensionality. The restoration accuracy depends on the original compression
ratio and the information preserved during compression.

Note: Decompression is a best-effort restoration and may not perfectly recreate
the original vector, especially for aggressive compression ratios. The restoration
accuracy metric indicates how well the original vector was recovered.

Examples:
  # Decompress a compressed vector
  ainative-code zerodb quantum decompress --vector-id vec_123

  # Decompress with JSON output
  ainative-code zerodb quantum decompress --vector-id vec_456 --json`,
	RunE: runQuantumDecompress,
}

// zerodbQuantumSearchCmd represents the zerodb quantum search command
var zerodbQuantumSearchCmd = &cobra.Command{
	Use:   "search",
	Short: "Perform quantum-enhanced vector similarity search",
	Long: `Perform quantum-enhanced vector similarity search with advanced capabilities.

Quantum search uses advanced algorithms to find similar vectors with improved
accuracy and additional features like entanglement awareness. When quantum boost
is enabled, the search algorithm considers quantum properties to enhance result
quality and ranking.

Features:
- Quantum-boosted similarity scoring
- Entanglement-aware search results
- Enhanced ranking algorithms
- Support for standard metadata filtering

Query Vector Format:
- Provide as JSON array: '[0.1, 0.2, 0.3, ...]'
- Must match the collection's dimension

Examples:
  # Basic quantum search
  ainative-code zerodb quantum search --query-vector '[0.1,0.2,0.3,0.4,0.5]' --limit 10

  # Search with quantum boost enabled
  ainative-code zerodb quantum search --query-vector '[0.1,0.2,0.3]' --limit 5 --use-quantum-boost

  # Include entangled vectors in results
  ainative-code zerodb quantum search --query-vector '[0.1,0.2,0.3]' --include-entangled

  # JSON output for programmatic use
  ainative-code zerodb quantum search --query-vector '[0.1,0.2,0.3]' --limit 5 --json`,
	RunE: runQuantumSearch,
}

func init() {
	zerodbCmd.AddCommand(zerodbQuantumCmd)

	// Add quantum subcommands
	zerodbQuantumCmd.AddCommand(zerodbQuantumEntangleCmd)
	zerodbQuantumCmd.AddCommand(zerodbQuantumMeasureCmd)
	zerodbQuantumCmd.AddCommand(zerodbQuantumCompressCmd)
	zerodbQuantumCmd.AddCommand(zerodbQuantumDecompressCmd)
	zerodbQuantumCmd.AddCommand(zerodbQuantumSearchCmd)

	// Quantum entangle flags
	zerodbQuantumEntangleCmd.Flags().StringVar(&entangleVectorID1, "vector-id-1", "", "first vector ID (required)")
	zerodbQuantumEntangleCmd.Flags().StringVar(&entangleVectorID2, "vector-id-2", "", "second vector ID (required)")
	zerodbQuantumEntangleCmd.MarkFlagRequired("vector-id-1")
	zerodbQuantumEntangleCmd.MarkFlagRequired("vector-id-2")

	// Quantum measure flags
	zerodbQuantumMeasureCmd.Flags().StringVar(&measureVectorID, "vector-id", "", "vector ID to measure (required)")
	zerodbQuantumMeasureCmd.MarkFlagRequired("vector-id")

	// Quantum compress flags
	zerodbQuantumCompressCmd.Flags().StringVar(&compressVectorID, "vector-id", "", "vector ID to compress (required)")
	zerodbQuantumCompressCmd.Flags().Float64Var(&compressionRatio, "compression-ratio", 0, "compression ratio between 0 and 1 (required)")
	zerodbQuantumCompressCmd.MarkFlagRequired("vector-id")
	zerodbQuantumCompressCmd.MarkFlagRequired("compression-ratio")

	// Quantum decompress flags
	zerodbQuantumDecompressCmd.Flags().StringVar(&decompressVectorID, "vector-id", "", "vector ID to decompress (required)")
	zerodbQuantumDecompressCmd.MarkFlagRequired("vector-id")

	// Quantum search flags
	zerodbQuantumSearchCmd.Flags().StringVar(&quantumSearchQueryVector, "query-vector", "", "query vector as JSON array (required)")
	zerodbQuantumSearchCmd.Flags().IntVar(&quantumSearchLimit, "limit", 10, "maximum number of results")
	zerodbQuantumSearchCmd.Flags().BoolVar(&quantumSearchUseQuantumBoost, "use-quantum-boost", false, "enable quantum boost for enhanced results")
	zerodbQuantumSearchCmd.Flags().BoolVar(&quantumSearchIncludeEntangled, "include-entangled", false, "include entangled vectors in results")
	zerodbQuantumSearchCmd.MarkFlagRequired("query-vector")

	// Global output flag for all quantum commands
	zerodbQuantumCmd.PersistentFlags().BoolVar(&quantumOutputJSON, "json", false, "output in JSON format")
}

func runQuantumEntangle(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Create ZeroDB client
	zdbClient, err := createZeroDBClient()
	if err != nil {
		return err
	}

	// Entangle vectors
	resp, err := zdbClient.QuantumEntangle(ctx, entangleVectorID1, entangleVectorID2)
	if err != nil {
		return fmt.Errorf("failed to entangle vectors: %w", err)
	}

	// Output result
	if quantumOutputJSON {
		return zerodbOutputJSON(resp)
	}

	fmt.Printf("Vectors entangled successfully!\n\n")
	fmt.Printf("Entanglement ID:     %s\n", resp.EntanglementID)
	fmt.Printf("Correlation Score:   %.4f\n\n", resp.CorrelationScore)

	fmt.Printf("Vector 1:\n")
	fmt.Printf("  ID:              %s\n", resp.Vector1.ID)
	fmt.Printf("  Dimension:       %d\n", resp.Vector1.Dimension)
	fmt.Printf("  Entangled:       %v\n", resp.Vector1.IsEntangled)
	if resp.Vector1.QuantumState != "" {
		fmt.Printf("  Quantum State:   %s\n", resp.Vector1.QuantumState)
	}

	fmt.Printf("\nVector 2:\n")
	fmt.Printf("  ID:              %s\n", resp.Vector2.ID)
	fmt.Printf("  Dimension:       %d\n", resp.Vector2.Dimension)
	fmt.Printf("  Entangled:       %v\n", resp.Vector2.IsEntangled)
	if resp.Vector2.QuantumState != "" {
		fmt.Printf("  Quantum State:   %s\n", resp.Vector2.QuantumState)
	}

	if resp.Message != "" {
		fmt.Printf("\n%s\n", resp.Message)
	}

	return nil
}

func runQuantumMeasure(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Create ZeroDB client
	zdbClient, err := createZeroDBClient()
	if err != nil {
		return err
	}

	// Measure vector
	resp, err := zdbClient.QuantumMeasure(ctx, measureVectorID)
	if err != nil {
		return fmt.Errorf("failed to measure vector: %w", err)
	}

	// Output result
	if quantumOutputJSON {
		return zerodbOutputJSON(resp)
	}

	fmt.Printf("Quantum Measurement Results:\n\n")
	fmt.Printf("Vector ID:         %s\n", resp.Vector.ID)
	fmt.Printf("Dimension:         %d\n", resp.Vector.Dimension)
	fmt.Printf("Quantum State:     %s\n", resp.QuantumState)
	fmt.Printf("Entropy:           %.6f\n", resp.Entropy)
	fmt.Printf("Coherence:         %.6f\n", resp.Coherence)

	if resp.Vector.IsEntangled {
		fmt.Printf("Entangled:         Yes\n")
		if len(resp.Vector.EntangledWith) > 0 {
			fmt.Printf("Entangled With:    %s\n", strings.Join(resp.Vector.EntangledWith, ", "))
		}
	} else {
		fmt.Printf("Entangled:         No\n")
	}

	if resp.Vector.CompressionRatio > 0 {
		fmt.Printf("Compression Ratio: %.2f\n", resp.Vector.CompressionRatio)
	}

	if len(resp.Properties) > 0 {
		fmt.Printf("\nAdditional Properties:\n")
		for key, value := range resp.Properties {
			fmt.Printf("  %s: %v\n", key, value)
		}
	}

	return nil
}

func runQuantumCompress(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Validate compression ratio
	if compressionRatio <= 0 || compressionRatio >= 1 {
		return fmt.Errorf("compression ratio must be between 0 and 1 (exclusive)")
	}

	// Create ZeroDB client
	zdbClient, err := createZeroDBClient()
	if err != nil {
		return err
	}

	// Compress vector
	resp, err := zdbClient.QuantumCompress(ctx, compressVectorID, compressionRatio)
	if err != nil {
		return fmt.Errorf("failed to compress vector: %w", err)
	}

	// Output result
	if quantumOutputJSON {
		return zerodbOutputJSON(resp)
	}

	fmt.Printf("Vector compressed successfully!\n\n")
	fmt.Printf("Vector ID:             %s\n", resp.Vector.ID)
	fmt.Printf("Original Dimension:    %d\n", resp.OriginalDimension)
	fmt.Printf("Compressed Dimension:  %d\n", resp.CompressedDimension)
	fmt.Printf("Compression Ratio:     %.2f\n", resp.CompressionRatio)
	fmt.Printf("Information Loss:      %.4f%%\n", resp.InformationLoss*100)

	// Calculate storage savings
	savings := float64(resp.OriginalDimension-resp.CompressedDimension) / float64(resp.OriginalDimension) * 100
	fmt.Printf("Storage Savings:       %.1f%%\n", savings)

	if resp.Message != "" {
		fmt.Printf("\n%s\n", resp.Message)
	}

	return nil
}

func runQuantumDecompress(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Create ZeroDB client
	zdbClient, err := createZeroDBClient()
	if err != nil {
		return err
	}

	// Decompress vector
	resp, err := zdbClient.QuantumDecompress(ctx, decompressVectorID)
	if err != nil {
		return fmt.Errorf("failed to decompress vector: %w", err)
	}

	// Output result
	if quantumOutputJSON {
		return zerodbOutputJSON(resp)
	}

	fmt.Printf("Vector decompressed successfully!\n\n")
	fmt.Printf("Vector ID:               %s\n", resp.Vector.ID)
	fmt.Printf("Original Dimension:      %d\n", resp.OriginalDimension)
	fmt.Printf("Decompressed Dimension:  %d\n", resp.DecompressedDimension)
	fmt.Printf("Restoration Accuracy:    %.2f%%\n", resp.RestorationAccuracy*100)

	if resp.Message != "" {
		fmt.Printf("\n%s\n", resp.Message)
	}

	// Provide guidance based on accuracy
	if resp.RestorationAccuracy < 0.9 {
		fmt.Printf("\nWarning: Restoration accuracy is below 90%%. The decompressed vector may not\n")
		fmt.Printf("perfectly match the original. Consider using a higher compression ratio for\n")
		fmt.Printf("better restoration quality in the future.\n")
	}

	return nil
}

func runQuantumSearch(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Parse query vector
	var queryVector []float64
	if err := json.Unmarshal([]byte(quantumSearchQueryVector), &queryVector); err != nil {
		return fmt.Errorf("invalid query vector JSON: %w (expected array like [0.1,0.2,0.3])", err)
	}

	if len(queryVector) == 0 {
		return fmt.Errorf("query vector cannot be empty")
	}

	// Create ZeroDB client
	zdbClient, err := createZeroDBClient()
	if err != nil {
		return err
	}

	// Perform quantum search
	req := &zerodb.QuantumSearchRequest{
		QueryVector:      queryVector,
		Limit:            quantumSearchLimit,
		UseQuantumBoost:  quantumSearchUseQuantumBoost,
		IncludeEntangled: quantumSearchIncludeEntangled,
	}

	results, err := zdbClient.QuantumSearch(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to perform quantum search: %w", err)
	}

	// Output result
	if quantumOutputJSON {
		return zerodbOutputJSON(results)
	}

	if len(results) == 0 {
		fmt.Println("No similar vectors found.")
		return nil
	}

	fmt.Printf("Found %d similar vector(s):\n", len(results))
	if quantumSearchUseQuantumBoost {
		fmt.Printf("(Quantum boost: ENABLED)\n")
	}
	fmt.Println()

	// Create table writer for formatted output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

	if quantumSearchUseQuantumBoost {
		fmt.Fprintln(w, "RANK\tVECTOR ID\tSIMILARITY\tQUANTUM SIM\tENTANGLED")
		fmt.Fprintln(w, "----\t---------\t----------\t-----------\t---------")
	} else {
		fmt.Fprintln(w, "RANK\tVECTOR ID\tSIMILARITY\tENTANGLED")
		fmt.Fprintln(w, "----\t---------\t----------\t---------")
	}

	for _, result := range results {
		entangled := "No"
		if result.Vector.IsEntangled {
			entangled = "Yes"
			if len(result.Vector.EntangledWith) > 0 {
				entangled = fmt.Sprintf("Yes (%d)", len(result.Vector.EntangledWith))
			}
		}

		if quantumSearchUseQuantumBoost {
			fmt.Fprintf(w, "%d\t%s\t%.4f\t%.4f\t%s\n",
				result.Rank,
				result.Vector.ID,
				result.Similarity,
				result.QuantumSimilarity,
				entangled,
			)
		} else {
			fmt.Fprintf(w, "%d\t%s\t%.4f\t%s\n",
				result.Rank,
				result.Vector.ID,
				result.Similarity,
				entangled,
			)
		}
	}

	w.Flush()

	// Show detailed information for each result
	fmt.Println("\nDetailed Results:")
	for i, result := range results {
		fmt.Printf("\n%d. Vector %s (Similarity: %.4f)\n", i+1, result.Vector.ID, result.Similarity)
		fmt.Printf("   Dimension:     %d\n", result.Vector.Dimension)

		if quantumSearchUseQuantumBoost {
			fmt.Printf("   Quantum Sim:   %.4f\n", result.QuantumSimilarity)
		}

		if result.Vector.IsEntangled {
			fmt.Printf("   Entangled:     Yes\n")
			if len(result.Vector.EntangledWith) > 0 {
				fmt.Printf("   Entangled With: %s\n", strings.Join(result.Vector.EntangledWith, ", "))
			}
		}

		if result.Vector.QuantumState != "" {
			fmt.Printf("   Quantum State: %s\n", result.Vector.QuantumState)
		}

		if result.Vector.CompressionRatio > 0 {
			fmt.Printf("   Compressed:    %.0f%%\n", result.Vector.CompressionRatio*100)
		}

		if len(result.Vector.Metadata) > 0 {
			fmt.Printf("   Metadata:      ")
			metaJSON, _ := json.Marshal(result.Vector.Metadata)
			fmt.Printf("%s\n", string(metaJSON))
		}
	}

	return nil
}

// parseFloatArray parses a comma-separated string of floats into a slice.
func parseFloatArray(s string) ([]float64, error) {
	parts := strings.Split(s, ",")
	result := make([]float64, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		f, err := strconv.ParseFloat(part, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid float value '%s': %w", part, err)
		}
		result = append(result, f)
	}

	return result, nil
}
