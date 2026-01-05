package zerodb_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AINative-studio/ainative-code/internal/client"
	"github.com/AINative-studio/ainative-code/internal/client/zerodb"
)

// TestCreateCollection tests vector collection creation functionality.
func TestCreateCollection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/v1/projects/test-project/vectors/collections", r.URL.Path)

		var req zerodb.CreateCollectionRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Equal(t, "embeddings", req.Name)
		assert.Equal(t, 1536, req.Dimensions)
		assert.Equal(t, "cosine", req.Metric)

		resp := zerodb.CreateCollectionResponse{
			Collection: &zerodb.VectorCollection{
				ID:         "col-123",
				Name:       "embeddings",
				Dimensions: 1536,
				Metric:     "cosine",
				Count:      0,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	httpClient := client.New(
		client.WithBaseURL(server.URL),
	)

	zdbClient := zerodb.New(
		zerodb.WithAPIClient(httpClient),
		zerodb.WithProjectID("test-project"),
	)

	collection, err := zdbClient.CreateCollection(context.Background(), "embeddings", 1536, "cosine")
	require.NoError(t, err)
	assert.Equal(t, "col-123", collection.ID)
	assert.Equal(t, "embeddings", collection.Name)
	assert.Equal(t, 1536, collection.Dimensions)
	assert.Equal(t, "cosine", collection.Metric)
	assert.Equal(t, 0, collection.Count)
}

// TestCreateCollectionDefaultMetric tests that default metric is applied.
func TestCreateCollectionDefaultMetric(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req zerodb.CreateCollectionRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		// Verify default metric is set
		assert.Equal(t, "cosine", req.Metric)

		resp := zerodb.CreateCollectionResponse{
			Collection: &zerodb.VectorCollection{
				ID:         "col-123",
				Name:       "test",
				Dimensions: 768,
				Metric:     "cosine",
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	httpClient := client.New(
		client.WithBaseURL(server.URL),
	)

	zdbClient := zerodb.New(
		zerodb.WithAPIClient(httpClient),
		zerodb.WithProjectID("test-project"),
	)

	collection, err := zdbClient.CreateCollection(context.Background(), "test", 768, "")
	require.NoError(t, err)
	assert.Equal(t, "cosine", collection.Metric)
}

// TestInsertVector tests vector insertion functionality.
func TestInsertVector(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/v1/projects/test-project/vectors", r.URL.Path)

		var req zerodb.VectorInsertRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Equal(t, "embeddings", req.Collection)
		assert.Equal(t, 3, len(req.Vector))
		assert.Equal(t, 0.1, req.Vector[0])
		assert.Equal(t, 0.2, req.Vector[1])
		assert.Equal(t, 0.3, req.Vector[2])
		assert.Equal(t, "hello world", req.Metadata["text"])

		resp := zerodb.VectorInsertResponse{
			ID: "vec-123",
			Vector: &zerodb.Vector{
				ID:         "vec-123",
				Collection: "embeddings",
				Vector:     req.Vector,
				Metadata:   req.Metadata,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	httpClient := client.New(
		client.WithBaseURL(server.URL),
	)

	zdbClient := zerodb.New(
		zerodb.WithAPIClient(httpClient),
		zerodb.WithProjectID("test-project"),
	)

	vector := []float64{0.1, 0.2, 0.3}
	metadata := map[string]interface{}{
		"text": "hello world",
	}

	id, vec, err := zdbClient.InsertVector(context.Background(), "embeddings", vector, metadata, "")
	require.NoError(t, err)
	assert.Equal(t, "vec-123", id)
	assert.Equal(t, "embeddings", vec.Collection)
	assert.Equal(t, 3, len(vec.Vector))
	assert.Equal(t, "hello world", vec.Metadata["text"])
}

// TestInsertVectorWithCustomID tests vector insertion with custom ID (upsert).
func TestInsertVectorWithCustomID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req zerodb.VectorInsertRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Equal(t, "custom-123", req.ID)

		resp := zerodb.VectorInsertResponse{
			ID: "custom-123",
			Vector: &zerodb.Vector{
				ID:         "custom-123",
				Collection: req.Collection,
				Vector:     req.Vector,
				Metadata:   req.Metadata,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	httpClient := client.New(
		client.WithBaseURL(server.URL),
	)

	zdbClient := zerodb.New(
		zerodb.WithAPIClient(httpClient),
		zerodb.WithProjectID("test-project"),
	)

	vector := []float64{0.1, 0.2, 0.3}
	id, vec, err := zdbClient.InsertVector(context.Background(), "embeddings", vector, nil, "custom-123")
	require.NoError(t, err)
	assert.Equal(t, "custom-123", id)
	assert.Equal(t, "custom-123", vec.ID)
}

// TestSearchVectors tests vector similarity search functionality.
func TestSearchVectors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/v1/projects/test-project/vectors/search", r.URL.Path)

		var req zerodb.VectorSearchRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Equal(t, "embeddings", req.Collection)
		assert.Equal(t, 3, len(req.QueryVector))
		assert.Equal(t, 5, req.Limit)

		resp := zerodb.VectorSearchResponse{
			Vectors: []*zerodb.Vector{
				{
					ID:         "vec-123",
					Collection: "embeddings",
					Vector:     []float64{0.11, 0.21, 0.31},
					Metadata:   map[string]interface{}{"text": "hello world"},
					Score:      0.95,
					CreatedAt:  time.Now(),
					UpdatedAt:  time.Now(),
				},
				{
					ID:         "vec-456",
					Collection: "embeddings",
					Vector:     []float64{0.12, 0.22, 0.32},
					Metadata:   map[string]interface{}{"text": "hello there"},
					Score:      0.92,
					CreatedAt:  time.Now(),
					UpdatedAt:  time.Now(),
				},
			},
			Total: 2,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	httpClient := client.New(
		client.WithBaseURL(server.URL),
	)

	zdbClient := zerodb.New(
		zerodb.WithAPIClient(httpClient),
		zerodb.WithProjectID("test-project"),
	)

	queryVector := []float64{0.1, 0.2, 0.3}
	vectors, err := zdbClient.SearchVectors(context.Background(), "embeddings", queryVector, 5, nil)
	require.NoError(t, err)
	assert.Equal(t, 2, len(vectors))
	assert.Equal(t, "vec-123", vectors[0].ID)
	assert.Equal(t, 0.95, vectors[0].Score)
	assert.Equal(t, "vec-456", vectors[1].ID)
	assert.Equal(t, 0.92, vectors[1].Score)
}

// TestSearchVectorsWithFilter tests vector search with metadata filtering.
func TestSearchVectorsWithFilter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req zerodb.VectorSearchRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		// Verify filter is passed correctly
		assert.NotNil(t, req.Filter)
		assert.Equal(t, "doc1", req.Filter["source"])

		resp := zerodb.VectorSearchResponse{
			Vectors: []*zerodb.Vector{
				{
					ID:         "vec-123",
					Collection: "embeddings",
					Vector:     []float64{0.1, 0.2, 0.3},
					Metadata:   map[string]interface{}{"source": "doc1"},
					Score:      0.95,
					CreatedAt:  time.Now(),
					UpdatedAt:  time.Now(),
				},
			},
			Total: 1,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	httpClient := client.New(
		client.WithBaseURL(server.URL),
	)

	zdbClient := zerodb.New(
		zerodb.WithAPIClient(httpClient),
		zerodb.WithProjectID("test-project"),
	)

	queryVector := []float64{0.1, 0.2, 0.3}
	filter := zerodb.QueryFilter{"source": "doc1"}
	vectors, err := zdbClient.SearchVectors(context.Background(), "embeddings", queryVector, 10, filter)
	require.NoError(t, err)
	assert.Equal(t, 1, len(vectors))
	assert.Equal(t, "doc1", vectors[0].Metadata["source"])
}

// TestSearchVectorsDefaultLimit tests that default limit is applied.
func TestSearchVectorsDefaultLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req zerodb.VectorSearchRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		// Verify default limit is set
		assert.Equal(t, 10, req.Limit)

		resp := zerodb.VectorSearchResponse{
			Vectors: []*zerodb.Vector{},
			Total:   0,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	httpClient := client.New(
		client.WithBaseURL(server.URL),
	)

	zdbClient := zerodb.New(
		zerodb.WithAPIClient(httpClient),
		zerodb.WithProjectID("test-project"),
	)

	queryVector := []float64{0.1, 0.2, 0.3}
	_, err := zdbClient.SearchVectors(context.Background(), "embeddings", queryVector, 0, nil)
	require.NoError(t, err)
}

// TestDeleteVector tests vector deletion functionality.
func TestDeleteVector(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/api/v1/projects/test-project/vectors/vec-123", r.URL.Path)
		assert.Equal(t, "embeddings", r.URL.Query().Get("collection"))

		resp := zerodb.VectorDeleteResponse{
			Success: true,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	httpClient := client.New(
		client.WithBaseURL(server.URL),
	)

	zdbClient := zerodb.New(
		zerodb.WithAPIClient(httpClient),
		zerodb.WithProjectID("test-project"),
	)

	err := zdbClient.DeleteVector(context.Background(), "embeddings", "vec-123")
	require.NoError(t, err)
}

// TestDeleteVectorFailure tests error handling for failed deletions.
func TestDeleteVectorFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := zerodb.VectorDeleteResponse{
			Success: false,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	httpClient := client.New(
		client.WithBaseURL(server.URL),
	)

	zdbClient := zerodb.New(
		zerodb.WithAPIClient(httpClient),
		zerodb.WithProjectID("test-project"),
	)

	err := zdbClient.DeleteVector(context.Background(), "embeddings", "vec-123")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "delete operation failed")
}

// TestListCollections tests listing vector collections functionality.
func TestListCollections(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/v1/projects/test-project/vectors/collections", r.URL.Path)

		resp := zerodb.ListCollectionsResponse{
			Collections: []*zerodb.VectorCollection{
				{
					ID:         "col-123",
					Name:       "embeddings",
					Dimensions: 1536,
					Metric:     "cosine",
					Count:      100,
					CreatedAt:  time.Now(),
					UpdatedAt:  time.Now(),
				},
				{
					ID:         "col-456",
					Name:       "images",
					Dimensions: 512,
					Metric:     "euclidean",
					Count:      50,
					CreatedAt:  time.Now(),
					UpdatedAt:  time.Now(),
				},
			},
			Total: 2,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	httpClient := client.New(
		client.WithBaseURL(server.URL),
	)

	zdbClient := zerodb.New(
		zerodb.WithAPIClient(httpClient),
		zerodb.WithProjectID("test-project"),
	)

	collections, err := zdbClient.ListCollections(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 2, len(collections))
	assert.Equal(t, "embeddings", collections[0].Name)
	assert.Equal(t, 1536, collections[0].Dimensions)
	assert.Equal(t, 100, collections[0].Count)
	assert.Equal(t, "images", collections[1].Name)
	assert.Equal(t, 512, collections[1].Dimensions)
	assert.Equal(t, 50, collections[1].Count)
}

// TestListCollectionsEmpty tests listing when no collections exist.
func TestListCollectionsEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := zerodb.ListCollectionsResponse{
			Collections: []*zerodb.VectorCollection{},
			Total:       0,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	httpClient := client.New(
		client.WithBaseURL(server.URL),
	)

	zdbClient := zerodb.New(
		zerodb.WithAPIClient(httpClient),
		zerodb.WithProjectID("test-project"),
	)

	collections, err := zdbClient.ListCollections(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 0, len(collections))
}

// TestVectorOperationsErrorHandling tests error handling for API failures.
func TestVectorOperationsErrorHandling(t *testing.T) {
	// Test server returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
	}))
	defer server.Close()

	httpClient := client.New(
		client.WithBaseURL(server.URL),
	)

	zdbClient := zerodb.New(
		zerodb.WithAPIClient(httpClient),
		zerodb.WithProjectID("test-project"),
	)

	ctx := context.Background()

	// Test CreateCollection error
	_, err := zdbClient.CreateCollection(ctx, "test", 768, "cosine")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create collection")

	// Test InsertVector error
	vector := []float64{0.1, 0.2, 0.3}
	_, _, err = zdbClient.InsertVector(ctx, "test", vector, nil, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to insert vector")

	// Test SearchVectors error
	_, err = zdbClient.SearchVectors(ctx, "test", vector, 10, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to search vectors")

	// Test DeleteVector error
	err = zdbClient.DeleteVector(ctx, "test", "vec-123")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete vector")

	// Test ListCollections error
	_, err = zdbClient.ListCollections(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list collections")
}
