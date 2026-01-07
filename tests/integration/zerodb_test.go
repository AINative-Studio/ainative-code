// +build integration

package integration

import (
	"context"
	"testing"

	"github.com/AINative-studio/ainative-code/internal/client"
	"github.com/AINative-studio/ainative-code/internal/client/zerodb"
	"github.com/AINative-studio/ainative-code/tests/integration/helpers"
	"github.com/stretchr/testify/suite"
)

// ZeroDBIntegrationTestSuite tests ZeroDB operations.
type ZeroDBIntegrationTestSuite struct {
	suite.Suite
	client  *zerodb.Client
	cleanup func()
}

// SetupTest runs before each test in the suite.
func (s *ZeroDBIntegrationTestSuite) SetupTest() {
	server, cleanup := helpers.MockZeroDBServer(s.T())
	s.cleanup = cleanup

	// Create API client pointing to mock server
	apiClient := client.New(
		client.WithBaseURL(server.URL),
	)

	// Create ZeroDB client
	s.client = zerodb.New(
		zerodb.WithAPIClient(apiClient),
		zerodb.WithProjectID("test-project"),
	)
}

// TearDownTest runs after each test in the suite.
func (s *ZeroDBIntegrationTestSuite) TearDownTest() {
	if s.cleanup != nil {
		s.cleanup()
	}
}

// TestTableCreationAndListing tests creating and listing tables.
func (s *ZeroDBIntegrationTestSuite) TestTableCreationAndListing() {
	// Given: A table schema
	ctx := context.Background()
	schema := map[string]interface{}{
		"name":  "string",
		"email": "string",
		"age":   "integer",
	}

	// When: Creating a table
	table, err := s.client.CreateTable(ctx, "users", schema)
	s.Require().NoError(err, "Failed to create table")
	s.NotNil(table)
	s.Equal("test_table", table.Name)

	// Then: Table should be listable
	tables, err := s.client.ListTables(ctx)
	s.Require().NoError(err, "Failed to list tables")
	s.NotEmpty(tables, "Expected at least one table")
}

// TestDocumentInsertAndQuery tests inserting and querying documents.
func (s *ZeroDBIntegrationTestSuite) TestDocumentInsertAndQuery() {
	// Given: A table with a document
	ctx := context.Background()
	data := map[string]interface{}{
		"name":  "John Doe",
		"email": "john@example.com",
		"age":   30,
	}

	// When: Inserting a document
	docID, doc, err := s.client.Insert(ctx, "users", data)
	s.Require().NoError(err, "Failed to insert document")
	s.NotEmpty(docID)
	s.NotNil(doc)

	// Then: Document should be queryable
	filter := zerodb.QueryFilter{
		"name": "John Doe",
	}
	options := zerodb.QueryOptions{
		Limit:  10,
		Offset: 0,
	}

	docs, err := s.client.Query(ctx, "users", filter, options)
	s.Require().NoError(err, "Failed to query documents")
	s.NotEmpty(docs, "Expected at least one document")
}

// TestDocumentUpdateAndDelete tests updating and deleting documents.
func (s *ZeroDBIntegrationTestSuite) TestDocumentUpdateAndDelete() {
	// Given: An existing document
	ctx := context.Background()
	data := map[string]interface{}{
		"name":  "Jane Doe",
		"email": "jane@example.com",
	}

	docID, _, err := s.client.Insert(ctx, "users", data)
	s.Require().NoError(err)

	// When: Updating the document
	updateData := map[string]interface{}{
		"email": "jane.doe@example.com",
	}
	updatedDoc, err := s.client.Update(ctx, "users", docID, updateData)
	s.Require().NoError(err, "Failed to update document")
	s.NotNil(updatedDoc)

	// When: Deleting the document
	err = s.client.Delete(ctx, "users", docID)
	s.Require().NoError(err, "Failed to delete document")
}

// TestMemoryStoreAndRetrieve tests agent memory storage and retrieval.
func (s *ZeroDBIntegrationTestSuite) TestMemoryStoreAndRetrieve() {
	// Given: Memory content to store
	ctx := context.Background()
	storeReq := &zerodb.MemoryStoreRequest{
		AgentID:   "test-agent",
		SessionID: "test-session",
		Content:   "Test memory content",
		Metadata: map[string]interface{}{
			"type": "conversation",
		},
	}

	// When: Storing memory
	memory, err := s.client.StoreMemory(ctx, storeReq)
	s.Require().NoError(err, "Failed to store memory")
	s.NotNil(memory)
	s.Equal("mem_123", memory.ID)

	// Then: Memory should be retrievable
	retrieveReq := &zerodb.MemoryRetrieveRequest{
		AgentID:   "test-agent",
		SessionID: "test-session",
		Query:     "conversation",
		Limit:     10,
	}

	memories, err := s.client.RetrieveMemory(ctx, retrieveReq)
	s.Require().NoError(err, "Failed to retrieve memories")
	s.NotEmpty(memories, "Expected at least one memory")
}

// TestMemoryListAndClear tests listing and clearing memories.
func (s *ZeroDBIntegrationTestSuite) TestMemoryListAndClear() {
	// Given: Stored memories
	ctx := context.Background()

	// When: Listing memories
	listReq := &zerodb.MemoryListRequest{
		AgentID:   "test-agent",
		SessionID: "test-session",
		Limit:     100,
		Offset:    0,
	}

	memories, total, err := s.client.ListMemory(ctx, listReq)
	s.Require().NoError(err, "Failed to list memories")
	s.NotNil(memories)
	s.GreaterOrEqual(total, 0)

	// When: Clearing memories
	clearReq := &zerodb.MemoryClearRequest{
		AgentID:   "test-agent",
		SessionID: "test-session",
	}

	clearResp, err := s.client.ClearMemory(ctx, clearReq)
	s.Require().NoError(err, "Failed to clear memories")
	s.NotNil(clearResp)
	s.GreaterOrEqual(clearResp.Deleted, 0)
}

// TestVectorCollectionOperations tests vector collection CRUD.
func (s *ZeroDBIntegrationTestSuite) TestVectorCollectionOperations() {
	// Given: A vector collection configuration
	ctx := context.Background()

	// When: Creating a collection
	collection, err := s.client.CreateCollection(ctx, "embeddings", 768, "cosine")
	s.Require().NoError(err, "Failed to create collection")
	s.NotNil(collection)

	// Then: Collection should be listable
	collections, err := s.client.ListCollections(ctx)
	s.Require().NoError(err, "Failed to list collections")
	s.NotEmpty(collections, "Expected at least one collection")
}

// TestVectorInsertAndSearch tests vector insertion and similarity search.
func (s *ZeroDBIntegrationTestSuite) TestVectorInsertAndSearch() {
	// Given: A vector to insert
	ctx := context.Background()
	vector := make([]float64, 768)
	for i := range vector {
		vector[i] = 0.5
	}
	metadata := map[string]interface{}{
		"text": "sample text",
	}

	// When: Inserting a vector
	vectorID, vec, err := s.client.InsertVector(ctx, "embeddings", vector, metadata, "")
	s.Require().NoError(err, "Failed to insert vector")
	s.NotEmpty(vectorID)
	s.NotNil(vec)

	// Then: Vector should be searchable
	queryVector := make([]float64, 768)
	for i := range queryVector {
		queryVector[i] = 0.5
	}

	results, err := s.client.SearchVectors(ctx, "embeddings", queryVector, 10, nil)
	s.Require().NoError(err, "Failed to search vectors")
	s.NotEmpty(results, "Expected at least one result")
}

// TestVectorDeletion tests deleting vectors.
func (s *ZeroDBIntegrationTestSuite) TestVectorDeletion() {
	// Given: An inserted vector
	ctx := context.Background()
	vector := make([]float64, 768)
	vectorID, _, err := s.client.InsertVector(ctx, "embeddings", vector, nil, "")
	s.Require().NoError(err)

	// When: Deleting the vector
	err = s.client.DeleteVector(ctx, "embeddings", vectorID)
	s.Require().NoError(err, "Failed to delete vector")
}

// TestQuantumEntanglement tests quantum vector entanglement.
func (s *ZeroDBIntegrationTestSuite) TestQuantumEntanglement() {
	// Given: Two vectors
	ctx := context.Background()
	vector1ID := "vec1"
	vector2ID := "vec2"

	// When: Entangling vectors
	resp, err := s.client.QuantumEntangle(ctx, vector1ID, vector2ID)
	s.Require().NoError(err, "Failed to entangle vectors")
	s.NotNil(resp)
	s.NotEmpty(resp.EntanglementID)
}

// TestQuantumMeasurement tests quantum state measurement.
func (s *ZeroDBIntegrationTestSuite) TestQuantumMeasurement() {
	// Given: A vector
	ctx := context.Background()
	vectorID := "test-vector"

	// When: Measuring quantum state
	resp, err := s.client.QuantumMeasure(ctx, vectorID)
	s.Require().NoError(err, "Failed to measure vector")
	s.NotNil(resp)
	s.NotEmpty(resp.QuantumState)
}

// TestQuantumCompression tests quantum vector compression.
func (s *ZeroDBIntegrationTestSuite) TestQuantumCompression() {
	// Given: A vector to compress
	ctx := context.Background()
	vectorID := "test-vector"
	compressionRatio := 0.5

	// When: Compressing the vector
	resp, err := s.client.QuantumCompress(ctx, vectorID, compressionRatio)
	s.Require().NoError(err, "Failed to compress vector")
	s.NotNil(resp)
	s.Greater(resp.OriginalDimension, resp.CompressedDimension)
}

// TestQuantumDecompression tests quantum vector decompression.
func (s *ZeroDBIntegrationTestSuite) TestQuantumDecompression() {
	// Given: A compressed vector
	ctx := context.Background()
	vectorID := "compressed-vector"

	// When: Decompressing the vector
	resp, err := s.client.QuantumDecompress(ctx, vectorID)
	s.Require().NoError(err, "Failed to decompress vector")
	s.NotNil(resp)
	s.Greater(resp.DecompressedDimension, 0)
}

// TestQuantumSearch tests quantum-enhanced search.
func (s *ZeroDBIntegrationTestSuite) TestQuantumSearch() {
	// Given: A quantum search request
	ctx := context.Background()
	queryVector := make([]float64, 768)
	for i := range queryVector {
		queryVector[i] = 0.5
	}

	req := &zerodb.QuantumSearchRequest{
		QueryVector:     queryVector,
		Limit:           10,
		UseQuantumBoost: true,
	}

	// When: Performing quantum search
	results, err := s.client.QuantumSearch(ctx, req)
	s.Require().NoError(err, "Failed to perform quantum search")
	s.NotNil(results)
}

// TestErrorHandling tests error conditions and validation.
func (s *ZeroDBIntegrationTestSuite) TestErrorHandling() {
	ctx := context.Background()

	// Test missing required fields
	_, err := s.client.StoreMemory(ctx, &zerodb.MemoryStoreRequest{})
	s.Error(err, "Should error with missing agent_id")

	_, err = s.client.RetrieveMemory(ctx, &zerodb.MemoryRetrieveRequest{})
	s.Error(err, "Should error with missing agent_id")

	_, err = s.client.QuantumEntangle(ctx, "", "")
	s.Error(err, "Should error with empty vector IDs")

	_, err = s.client.QuantumCompress(ctx, "test", 1.5)
	s.Error(err, "Should error with invalid compression ratio")
}

// TestZeroDBIntegrationTestSuite runs the test suite.
func TestZeroDBIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(ZeroDBIntegrationTestSuite))
}
