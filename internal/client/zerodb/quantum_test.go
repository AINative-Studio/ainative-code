package zerodb_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AINative-studio/ainative-code/internal/client"
	"github.com/AINative-studio/ainative-code/internal/client/zerodb"
)

// TestQuantumEntangle tests vector entanglement functionality.
func TestQuantumEntangle(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Setup
	ctx := context.Background()
	zdbClient := setupTestClient(t)

	// Test cases
	tests := []struct {
		name        string
		vectorID1   string
		vectorID2   string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "successful entanglement",
			vectorID1:   "vec_test_1",
			vectorID2:   "vec_test_2",
			expectError: false,
		},
		{
			name:        "empty vector ID 1",
			vectorID1:   "",
			vectorID2:   "vec_test_2",
			expectError: true,
			errorMsg:    "both vector IDs are required",
		},
		{
			name:        "empty vector ID 2",
			vectorID1:   "vec_test_1",
			vectorID2:   "",
			expectError: true,
			errorMsg:    "both vector IDs are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			resp, err := zdbClient.QuantumEntangle(ctx, tt.vectorID1, tt.vectorID2)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, resp)
				assert.NotEmpty(t, resp.EntanglementID)
				assert.NotNil(t, resp.Vector1)
				assert.NotNil(t, resp.Vector2)
				assert.Equal(t, tt.vectorID1, resp.Vector1.ID)
				assert.Equal(t, tt.vectorID2, resp.Vector2.ID)
				assert.True(t, resp.Vector1.IsEntangled)
				assert.True(t, resp.Vector2.IsEntangled)
				assert.GreaterOrEqual(t, resp.CorrelationScore, 0.0)
				assert.LessOrEqual(t, resp.CorrelationScore, 1.0)
			}
		})
	}
}

// TestQuantumMeasure tests quantum state measurement functionality.
func TestQuantumMeasure(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Setup
	ctx := context.Background()
	zdbClient := setupTestClient(t)

	// Test cases
	tests := []struct {
		name        string
		vectorID    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "successful measurement",
			vectorID:    "vec_test_measure",
			expectError: false,
		},
		{
			name:        "empty vector ID",
			vectorID:    "",
			expectError: true,
			errorMsg:    "vector_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			resp, err := zdbClient.QuantumMeasure(ctx, tt.vectorID)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, resp)
				assert.NotNil(t, resp.Vector)
				assert.Equal(t, tt.vectorID, resp.Vector.ID)
				assert.NotEmpty(t, resp.QuantumState)
				assert.GreaterOrEqual(t, resp.Entropy, 0.0)
				assert.GreaterOrEqual(t, resp.Coherence, 0.0)
				assert.LessOrEqual(t, resp.Coherence, 1.0)
				assert.NotNil(t, resp.Properties)
			}
		})
	}
}

// TestQuantumCompress tests quantum vector compression functionality.
func TestQuantumCompress(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Setup
	ctx := context.Background()
	zdbClient := setupTestClient(t)

	// Test cases
	tests := []struct {
		name             string
		vectorID         string
		compressionRatio float64
		expectError      bool
		errorMsg         string
	}{
		{
			name:             "successful compression 50%",
			vectorID:         "vec_test_compress",
			compressionRatio: 0.5,
			expectError:      false,
		},
		{
			name:             "successful compression 30%",
			vectorID:         "vec_test_compress",
			compressionRatio: 0.3,
			expectError:      false,
		},
		{
			name:             "invalid ratio - too low",
			vectorID:         "vec_test_compress",
			compressionRatio: 0.0,
			expectError:      true,
			errorMsg:         "compression_ratio must be between 0 and 1",
		},
		{
			name:             "invalid ratio - too high",
			vectorID:         "vec_test_compress",
			compressionRatio: 1.0,
			expectError:      true,
			errorMsg:         "compression_ratio must be between 0 and 1",
		},
		{
			name:             "invalid ratio - negative",
			vectorID:         "vec_test_compress",
			compressionRatio: -0.5,
			expectError:      true,
			errorMsg:         "compression_ratio must be between 0 and 1",
		},
		{
			name:             "empty vector ID",
			vectorID:         "",
			compressionRatio: 0.5,
			expectError:      true,
			errorMsg:         "vector_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			resp, err := zdbClient.QuantumCompress(ctx, tt.vectorID, tt.compressionRatio)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, resp)
				assert.NotNil(t, resp.Vector)
				assert.Equal(t, tt.vectorID, resp.Vector.ID)
				assert.Greater(t, resp.OriginalDimension, 0)
				assert.Greater(t, resp.CompressedDimension, 0)
				assert.Less(t, resp.CompressedDimension, resp.OriginalDimension)
				assert.Equal(t, tt.compressionRatio, resp.CompressionRatio)
				assert.GreaterOrEqual(t, resp.InformationLoss, 0.0)
				assert.LessOrEqual(t, resp.InformationLoss, 1.0)
			}
		})
	}
}

// TestQuantumDecompress tests quantum vector decompression functionality.
func TestQuantumDecompress(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Setup
	ctx := context.Background()
	zdbClient := setupTestClient(t)

	// Test cases
	tests := []struct {
		name        string
		vectorID    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "successful decompression",
			vectorID:    "vec_test_decompress",
			expectError: false,
		},
		{
			name:        "empty vector ID",
			vectorID:    "",
			expectError: true,
			errorMsg:    "vector_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			resp, err := zdbClient.QuantumDecompress(ctx, tt.vectorID)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, resp)
				assert.NotNil(t, resp.Vector)
				assert.Equal(t, tt.vectorID, resp.Vector.ID)
				assert.Greater(t, resp.OriginalDimension, 0)
				assert.Greater(t, resp.DecompressedDimension, 0)
				assert.GreaterOrEqual(t, resp.RestorationAccuracy, 0.0)
				assert.LessOrEqual(t, resp.RestorationAccuracy, 1.0)
			}
		})
	}
}

// TestQuantumSearch tests quantum-enhanced vector search functionality.
func TestQuantumSearch(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Setup
	ctx := context.Background()
	zdbClient := setupTestClient(t)

	// Test cases
	tests := []struct {
		name        string
		request     *zerodb.QuantumSearchRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful basic search",
			request: &zerodb.QuantumSearchRequest{
				QueryVector: []float64{0.1, 0.2, 0.3, 0.4, 0.5},
				Limit:       10,
			},
			expectError: false,
		},
		{
			name: "search with quantum boost",
			request: &zerodb.QuantumSearchRequest{
				QueryVector:     []float64{0.1, 0.2, 0.3},
				Limit:           5,
				UseQuantumBoost: true,
			},
			expectError: false,
		},
		{
			name: "search with entangled vectors",
			request: &zerodb.QuantumSearchRequest{
				QueryVector:      []float64{0.1, 0.2, 0.3},
				Limit:            5,
				IncludeEntangled: true,
			},
			expectError: false,
		},
		{
			name: "search with all options",
			request: &zerodb.QuantumSearchRequest{
				QueryVector:      []float64{0.1, 0.2, 0.3},
				Limit:            10,
				UseQuantumBoost:  true,
				IncludeEntangled: true,
			},
			expectError: false,
		},
		{
			name: "empty query vector",
			request: &zerodb.QuantumSearchRequest{
				QueryVector: []float64{},
				Limit:       10,
			},
			expectError: true,
			errorMsg:    "query_vector is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			results, err := zdbClient.QuantumSearch(ctx, tt.request)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				assert.Nil(t, results)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, results)

				// Validate results
				for _, result := range results {
					assert.NotNil(t, result.Vector)
					assert.NotEmpty(t, result.Vector.ID)
					assert.GreaterOrEqual(t, result.Similarity, 0.0)
					assert.LessOrEqual(t, result.Similarity, 1.0)
					assert.Greater(t, result.Rank, 0)

					if tt.request.UseQuantumBoost {
						assert.GreaterOrEqual(t, result.QuantumSimilarity, 0.0)
						assert.LessOrEqual(t, result.QuantumSimilarity, 1.0)
					}
				}

				// Validate ranking order
				for i := 1; i < len(results); i++ {
					assert.GreaterOrEqual(t, results[i-1].Similarity, results[i].Similarity,
						"results should be ordered by similarity descending")
				}
			}
		})
	}
}

// TestQuantumCompressDecompressCycle tests the full compress-decompress cycle.
func TestQuantumCompressDecompressCycle(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Setup
	ctx := context.Background()
	zdbClient := setupTestClient(t)
	vectorID := "vec_test_cycle"

	// Compress
	compressResp, err := zdbClient.QuantumCompress(ctx, vectorID, 0.5)
	require.NoError(t, err)
	require.NotNil(t, compressResp)
	originalDim := compressResp.OriginalDimension
	compressedDim := compressResp.CompressedDimension

	// Decompress
	decompressResp, err := zdbClient.QuantumDecompress(ctx, vectorID)
	require.NoError(t, err)
	require.NotNil(t, decompressResp)

	// Assertions
	assert.Equal(t, originalDim, decompressResp.OriginalDimension,
		"original dimension should match")
	assert.Equal(t, compressedDim, decompressResp.OriginalDimension,
		"decompression should reference the compressed dimension as original")
	assert.GreaterOrEqual(t, decompressResp.RestorationAccuracy, 0.5,
		"restoration accuracy should be reasonable for 50% compression")
}

// TestQuantumEntangleMeasure tests entanglement followed by measurement.
func TestQuantumEntangleMeasure(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Setup
	ctx := context.Background()
	zdbClient := setupTestClient(t)
	vectorID1 := "vec_test_entangle_1"
	vectorID2 := "vec_test_entangle_2"

	// Entangle vectors
	entangleResp, err := zdbClient.QuantumEntangle(ctx, vectorID1, vectorID2)
	require.NoError(t, err)
	require.NotNil(t, entangleResp)

	// Measure first vector
	measureResp1, err := zdbClient.QuantumMeasure(ctx, vectorID1)
	require.NoError(t, err)
	require.NotNil(t, measureResp1)
	assert.True(t, measureResp1.Vector.IsEntangled,
		"measured vector should show as entangled")
	assert.Contains(t, measureResp1.Vector.EntangledWith, vectorID2,
		"measured vector should list its entanglement partner")

	// Measure second vector
	measureResp2, err := zdbClient.QuantumMeasure(ctx, vectorID2)
	require.NoError(t, err)
	require.NotNil(t, measureResp2)
	assert.True(t, measureResp2.Vector.IsEntangled,
		"measured vector should show as entangled")
	assert.Contains(t, measureResp2.Vector.EntangledWith, vectorID1,
		"measured vector should list its entanglement partner")
}

// TestQuantumSearchWithEntangledVectors tests search behavior with entangled vectors.
func TestQuantumSearchWithEntangledVectors(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Setup
	ctx := context.Background()
	zdbClient := setupTestClient(t)

	// Search without entangled vectors
	req1 := &zerodb.QuantumSearchRequest{
		QueryVector:      []float64{0.1, 0.2, 0.3},
		Limit:            10,
		IncludeEntangled: false,
	}
	results1, err := zdbClient.QuantumSearch(ctx, req1)
	require.NoError(t, err)
	count1 := len(results1)

	// Search with entangled vectors
	req2 := &zerodb.QuantumSearchRequest{
		QueryVector:      []float64{0.1, 0.2, 0.3},
		Limit:            10,
		IncludeEntangled: true,
	}
	results2, err := zdbClient.QuantumSearch(ctx, req2)
	require.NoError(t, err)
	count2 := len(results2)

	// Entangled search may return more or different results
	assert.GreaterOrEqual(t, count2, count1,
		"search with entangled vectors should return at least as many results")
}

// setupTestClient creates a test ZeroDB client.
func setupTestClient(t *testing.T) *zerodb.Client {
	t.Helper()

	// Create HTTP client
	httpClient := client.New(
		client.WithBaseURL("https://api.ainative.studio"),
		client.WithTimeout(30*time.Second),
	)

	// Create ZeroDB client
	zdbClient := zerodb.New(
		zerodb.WithAPIClient(httpClient),
		zerodb.WithProjectID("test-project"),
	)

	return zdbClient
}
