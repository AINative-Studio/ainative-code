package zerodb

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/AINative-studio/ainative-code/internal/client"
	"github.com/AINative-studio/ainative-code/internal/logger"
	"github.com/google/uuid"
)

// Client represents a client for ZeroDB NoSQL operations.
type Client struct {
	apiClient *client.Client
	projectID string
}

// Option is a functional option for configuring the Client.
type Option func(*Client)

// WithAPIClient sets the underlying HTTP API client.
func WithAPIClient(apiClient *client.Client) Option {
	return func(c *Client) {
		c.apiClient = apiClient
	}
}

// WithProjectID sets the ZeroDB project ID.
func WithProjectID(projectID string) Option {
	return func(c *Client) {
		c.projectID = projectID
	}
}

// New creates a new ZeroDB client with the specified options.
func New(opts ...Option) *Client {
	c := &Client{}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// CreateTable creates a new NoSQL table with the specified schema.
func (c *Client) CreateTable(ctx context.Context, name string, schema map[string]interface{}) (*Table, error) {
	logger.InfoEvent().
		Str("table", name).
		Msg("Creating ZeroDB table")

	req := &CreateTableRequest{
		Name:   name,
		Schema: schema,
	}

	path := fmt.Sprintf("/api/v1/projects/%s/nosql/tables", c.projectID)
	respData, err := c.apiClient.Post(ctx, path, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	var resp CreateTableResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	logger.InfoEvent().
		Str("table", name).
		Str("id", resp.Table.ID).
		Msg("Table created successfully")

	return resp.Table, nil
}

// Insert inserts a new document into the specified table.
func (c *Client) Insert(ctx context.Context, tableName string, data map[string]interface{}) (string, *Document, error) {
	logger.DebugEvent().
		Str("table", tableName).
		Msg("Inserting document")

	req := &InsertRequest{
		TableName: tableName,
		Data:      data,
	}

	path := fmt.Sprintf("/api/v1/projects/%s/nosql/documents", c.projectID)
	respData, err := c.apiClient.Post(ctx, path, req)
	if err != nil {
		return "", nil, fmt.Errorf("failed to insert document: %w", err)
	}

	var resp InsertResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		return "", nil, fmt.Errorf("failed to parse response: %w", err)
	}

	logger.DebugEvent().
		Str("table", tableName).
		Str("id", resp.ID).
		Msg("Document inserted successfully")

	return resp.ID, resp.Document, nil
}

// Query queries documents from the specified table with optional filter.
func (c *Client) Query(ctx context.Context, tableName string, filter QueryFilter, options QueryOptions) ([]*Document, error) {
	logger.DebugEvent().
		Str("table", tableName).
		Msg("Querying documents")

	req := &QueryRequest{
		TableName: tableName,
		Filter:    filter,
		Options:   options,
	}

	path := fmt.Sprintf("/api/v1/projects/%s/nosql/query", c.projectID)
	respData, err := c.apiClient.Post(ctx, path, req)
	if err != nil {
		return nil, fmt.Errorf("failed to query documents: %w", err)
	}

	var resp QueryResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	logger.DebugEvent().
		Str("table", tableName).
		Int("count", len(resp.Documents)).
		Int("total", resp.Total).
		Msg("Query completed")

	return resp.Documents, nil
}

// Update updates a document in the specified table.
func (c *Client) Update(ctx context.Context, tableName string, id string, data map[string]interface{}) (*Document, error) {
	logger.DebugEvent().
		Str("table", tableName).
		Str("id", id).
		Msg("Updating document")

	req := &UpdateRequest{
		TableName: tableName,
		ID:        id,
		Data:      data,
	}

	path := fmt.Sprintf("/api/v1/projects/%s/nosql/documents/%s", c.projectID, id)
	respData, err := c.apiClient.Put(ctx, path, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update document: %w", err)
	}

	var resp UpdateResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	logger.DebugEvent().
		Str("table", tableName).
		Str("id", id).
		Msg("Document updated successfully")

	return resp.Document, nil
}

// Delete deletes a document from the specified table.
func (c *Client) Delete(ctx context.Context, tableName string, id string) error {
	logger.DebugEvent().
		Str("table", tableName).
		Str("id", id).
		Msg("Deleting document")

	path := fmt.Sprintf("/api/v1/projects/%s/nosql/documents/%s?table=%s", c.projectID, id, tableName)
	respData, err := c.apiClient.Delete(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	var resp DeleteResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("delete operation failed")
	}

	logger.DebugEvent().
		Str("table", tableName).
		Str("id", id).
		Msg("Document deleted successfully")

	return nil
}

// ListTables lists all tables in the project.
func (c *Client) ListTables(ctx context.Context) ([]*Table, error) {
	logger.DebugEvent().Msg("Listing tables")

	path := fmt.Sprintf("/api/v1/projects/%s/nosql/tables", c.projectID)
	respData, err := c.apiClient.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to list tables: %w", err)
	}

	var resp ListTablesResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	logger.DebugEvent().
		Int("count", len(resp.Tables)).
		Msg("Tables listed successfully")

	return resp.Tables, nil
}

// StoreMemory stores agent memory content using the embeddings API.
func (c *Client) StoreMemory(ctx context.Context, req *MemoryStoreRequest) (*Memory, error) {
	logger.InfoEvent().
		Str("agent_id", req.AgentID).
		Str("session_id", req.SessionID).
		Msg("Storing agent memory")

	if req.AgentID == "" {
		return nil, fmt.Errorf("agent_id is required")
	}
	if req.Content == "" {
		return nil, fmt.Errorf("content is required")
	}

	// Generate unique ID for this memory
	memoryID := fmt.Sprintf("memory_%s", uuid.New().String())

	// Build metadata object for this memory
	metadata := make(map[string]interface{})
	metadata["memory_id"] = memoryID
	metadata["agent_id"] = req.AgentID
	if req.SessionID != "" {
		metadata["session_id"] = req.SessionID
	}
	if req.Role != "" {
		metadata["role"] = req.Role
	}
	// Merge any additional metadata
	for k, v := range req.Metadata {
		metadata[k] = v
	}

	// Create embed-and-store request with correct API format
	embedReq := EmbedAndStoreRequest{
		Texts:     []string{req.Content},
		Namespace: "agent_memories",
		Metadata:  []map[string]interface{}{metadata},
	}

	path := fmt.Sprintf("/v1/public/%s/embeddings/embed-and-store", c.projectID)
	respData, err := c.apiClient.Post(ctx, path, embedReq)
	if err != nil {
		return nil, fmt.Errorf("failed to store memory: %w", err)
	}

	var embedResp EmbedAndStoreResponse
	if err := json.Unmarshal(respData, &embedResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !embedResp.Success || embedResp.VectorsStored == 0 {
		return nil, fmt.Errorf("failed to store memory: success=%v, vectors_stored=%d", embedResp.Success, embedResp.VectorsStored)
	}

	// Construct Memory object to return
	now := time.Now()
	memory := &Memory{
		ID:        memoryID,
		AgentID:   req.AgentID,
		SessionID: req.SessionID,
		Content:   req.Content,
		Role:      req.Role,
		Metadata:  req.Metadata,
		CreatedAt: now,
		UpdatedAt: now,
	}

	logger.InfoEvent().
		Str("agent_id", req.AgentID).
		Str("memory_id", memoryID).
		Msg("Memory stored successfully")

	return memory, nil
}

// RetrieveMemory retrieves agent memories using semantic search via embeddings API.
func (c *Client) RetrieveMemory(ctx context.Context, req *MemoryRetrieveRequest) ([]*Memory, error) {
	logger.DebugEvent().
		Str("agent_id", req.AgentID).
		Str("query", req.Query).
		Int("limit", req.Limit).
		Msg("Retrieving agent memories")

	if req.AgentID == "" {
		return nil, fmt.Errorf("agent_id is required")
	}
	if req.Query == "" {
		return nil, fmt.Errorf("query is required")
	}

	if req.Limit == 0 {
		req.Limit = 10
	}

	// Build filter for agent_id and optional session_id
	filter := make(map[string]interface{})
	filter["agent_id"] = req.AgentID
	if req.SessionID != "" {
		filter["session_id"] = req.SessionID
	}

	// Create search request
	searchReq := SearchEmbeddingsRequest{
		Query:     req.Query,
		TopK:      req.Limit,
		Namespace: "agent_memories",
		Filter:    filter,
	}

	path := fmt.Sprintf("/v1/public/%s/embeddings/search", c.projectID)
	respData, err := c.apiClient.Post(ctx, path, searchReq)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve memories: %w", err)
	}

	var searchResp SearchEmbeddingsResponse
	if err := json.Unmarshal(respData, &searchResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Convert search results to Memory objects
	memories := make([]*Memory, 0, len(searchResp.Results))
	for _, result := range searchResp.Results {
		// Parse created_at timestamp
		var createdAt time.Time
		if result.CreatedAt != "" {
			if t, err := time.Parse(time.RFC3339, result.CreatedAt); err == nil {
				createdAt = t
			}
		}

		memory := &Memory{
			ID:         result.VectorID,
			Content:    result.Document,
			Similarity: result.Score,
			CreatedAt:  createdAt,
			UpdatedAt:  createdAt,
		}

		// Extract metadata fields from vector_metadata
		if result.VectorMetadata != nil {
			if agentID, ok := result.VectorMetadata["agent_id"].(string); ok {
				memory.AgentID = agentID
			}
			if sessionID, ok := result.VectorMetadata["session_id"].(string); ok {
				memory.SessionID = sessionID
			}
			if role, ok := result.VectorMetadata["role"].(string); ok {
				memory.Role = role
			}
			// Store full metadata
			memory.Metadata = result.VectorMetadata
		}

		memories = append(memories, memory)
	}

	logger.DebugEvent().
		Str("agent_id", req.AgentID).
		Int("count", len(memories)).
		Int("total", searchResp.Total).
		Msg("Memories retrieved successfully")

	return memories, nil
}

// ClearMemory clears agent memories by searching and deleting them.
// Note: This performs a search followed by individual deletes since the embeddings API
// doesn't provide a bulk delete endpoint.
func (c *Client) ClearMemory(ctx context.Context, req *MemoryClearRequest) (*MemoryClearResponse, error) {
	logger.InfoEvent().
		Str("agent_id", req.AgentID).
		Str("session_id", req.SessionID).
		Msg("Clearing agent memories")

	if req.AgentID == "" {
		return nil, fmt.Errorf("agent_id is required")
	}

	// First, search for all memories matching the criteria
	filter := make(map[string]interface{})
	filter["agent_id"] = req.AgentID
	if req.SessionID != "" {
		filter["session_id"] = req.SessionID
	}

	searchReq := SearchEmbeddingsRequest{
		Query:     " ",
		TopK:      1000, // Get a large batch
		Namespace: "agent_memories",
		Filter:    filter,
	}

	searchPath := fmt.Sprintf("/v1/public/%s/embeddings/search", c.projectID)
	respData, err := c.apiClient.Post(ctx, searchPath, searchReq)
	if err != nil {
		return nil, fmt.Errorf("failed to search memories for deletion: %w", err)
	}

	var searchResp SearchEmbeddingsResponse
	if err := json.Unmarshal(respData, &searchResp); err != nil {
		return nil, fmt.Errorf("failed to parse search response: %w", err)
	}

	// Delete each memory
	deleted := 0
	for _, result := range searchResp.Results {
		// Try to delete using the embeddings namespace pattern
		deletePath := fmt.Sprintf("/v1/public/%s/embeddings/%s?namespace=agent_memories", c.projectID, result.VectorID)
		_, err := c.apiClient.Delete(ctx, deletePath)
		if err != nil {
			// Log error but continue with other deletions
			logger.WarnEvent().
				Str("memory_id", result.VectorID).
				Err(err).
				Msg("Failed to delete memory")
			continue
		}
		deleted++
	}

	logger.InfoEvent().
		Str("agent_id", req.AgentID).
		Int("deleted", deleted).
		Int("total_found", len(searchResp.Results)).
		Msg("Memories cleared")

	return &MemoryClearResponse{
		Deleted: deleted,
		Message: fmt.Sprintf("Deleted %d memories for agent %s", deleted, req.AgentID),
	}, nil
}

// ListMemory lists agent memories using embeddings search with a wildcard query.
// Note: Since embeddings API doesn't have a direct "list" endpoint, we use a broad search query.
func (c *Client) ListMemory(ctx context.Context, req *MemoryListRequest) ([]*Memory, int, error) {
	logger.DebugEvent().
		Str("agent_id", req.AgentID).
		Int("limit", req.Limit).
		Int("offset", req.Offset).
		Msg("Listing agent memories")

	if req.AgentID == "" {
		return nil, 0, fmt.Errorf("agent_id is required")
	}

	if req.Limit == 0 {
		req.Limit = 100
	}

	// Build filter for agent_id and optional session_id
	filter := make(map[string]interface{})
	filter["agent_id"] = req.AgentID
	if req.SessionID != "" {
		filter["session_id"] = req.SessionID
	}

	// Use a space as a broad query - embeddings API requires a query
	// The filter will constrain results to the specific agent
	searchReq := SearchEmbeddingsRequest{
		Query:     " ",
		TopK:      req.Limit,
		Namespace: "agent_memories",
		Filter:    filter,
	}

	path := fmt.Sprintf("/v1/public/%s/embeddings/search", c.projectID)
	respData, err := c.apiClient.Post(ctx, path, searchReq)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list memories: %w", err)
	}

	var searchResp SearchEmbeddingsResponse
	if err := json.Unmarshal(respData, &searchResp); err != nil {
		return nil, 0, fmt.Errorf("failed to parse response: %w", err)
	}

	// Convert search results to Memory objects
	memories := make([]*Memory, 0, len(searchResp.Results))
	for _, result := range searchResp.Results {
		// Parse created_at timestamp
		var createdAt time.Time
		if result.CreatedAt != "" {
			if t, err := time.Parse(time.RFC3339, result.CreatedAt); err == nil {
				createdAt = t
			}
		}

		memory := &Memory{
			ID:         result.VectorID,
			Content:    result.Document,
			Similarity: result.Score,
			CreatedAt:  createdAt,
			UpdatedAt:  createdAt,
		}

		// Extract metadata fields from vector_metadata
		if result.VectorMetadata != nil {
			if agentID, ok := result.VectorMetadata["agent_id"].(string); ok {
				memory.AgentID = agentID
			}
			if sessionID, ok := result.VectorMetadata["session_id"].(string); ok {
				memory.SessionID = sessionID
			}
			if role, ok := result.VectorMetadata["role"].(string); ok {
				memory.Role = role
			}
			// Store full metadata
			memory.Metadata = result.VectorMetadata
		}

		memories = append(memories, memory)
	}

	logger.DebugEvent().
		Str("agent_id", req.AgentID).
		Int("count", len(memories)).
		Int("total", searchResp.Total).
		Msg("Memories listed successfully")

	return memories, searchResp.Total, nil
}

// CreateCollection creates a new vector collection with the specified dimensions.
func (c *Client) CreateCollection(ctx context.Context, name string, dimensions int, metric string) (*VectorCollection, error) {
	logger.InfoEvent().
		Str("collection", name).
		Int("dimensions", dimensions).
		Str("metric", metric).
		Msg("Creating vector collection")

	if metric == "" {
		metric = "cosine"
	}

	req := &CreateCollectionRequest{
		Name:       name,
		Dimensions: dimensions,
		Metric:     metric,
	}

	path := fmt.Sprintf("/api/v1/projects/%s/vectors/collections", c.projectID)
	respData, err := c.apiClient.Post(ctx, path, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create collection: %w", err)
	}

	var resp CreateCollectionResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	logger.InfoEvent().
		Str("collection", name).
		Str("id", resp.Collection.ID).
		Msg("Collection created successfully")

	return resp.Collection, nil
}

// InsertVector inserts a new vector into the specified collection.
func (c *Client) InsertVector(ctx context.Context, collection string, vector []float64, metadata map[string]interface{}, id string) (string, *Vector, error) {
	logger.DebugEvent().
		Str("collection", collection).
		Int("dimensions", len(vector)).
		Msg("Inserting vector")

	req := &VectorInsertRequest{
		Collection: collection,
		Vector:     vector,
		Metadata:   metadata,
		ID:         id,
	}

	path := fmt.Sprintf("/api/v1/projects/%s/vectors", c.projectID)
	respData, err := c.apiClient.Post(ctx, path, req)
	if err != nil {
		return "", nil, fmt.Errorf("failed to insert vector: %w", err)
	}

	var resp VectorInsertResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		return "", nil, fmt.Errorf("failed to parse response: %w", err)
	}

	logger.DebugEvent().
		Str("collection", collection).
		Str("id", resp.ID).
		Msg("Vector inserted successfully")

	return resp.ID, resp.Vector, nil
}

// SearchVectors searches for vectors similar to the query vector.
func (c *Client) SearchVectors(ctx context.Context, collection string, queryVector []float64, limit int, filter QueryFilter) ([]*Vector, error) {
	logger.DebugEvent().
		Str("collection", collection).
		Int("dimensions", len(queryVector)).
		Int("limit", limit).
		Msg("Searching vectors")

	if limit == 0 {
		limit = 10
	}

	req := &VectorSearchRequest{
		Collection:  collection,
		QueryVector: queryVector,
		Limit:       limit,
		Filter:      filter,
	}

	path := fmt.Sprintf("/api/v1/projects/%s/vectors/search", c.projectID)
	respData, err := c.apiClient.Post(ctx, path, req)
	if err != nil {
		return nil, fmt.Errorf("failed to search vectors: %w", err)
	}

	var resp VectorSearchResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	logger.DebugEvent().
		Str("collection", collection).
		Int("count", len(resp.Vectors)).
		Int("total", resp.Total).
		Msg("Vector search completed")

	return resp.Vectors, nil
}

// DeleteVector deletes a vector from the specified collection.
func (c *Client) DeleteVector(ctx context.Context, collection string, id string) error {
	logger.DebugEvent().
		Str("collection", collection).
		Str("id", id).
		Msg("Deleting vector")

	path := fmt.Sprintf("/api/v1/projects/%s/vectors/%s?collection=%s", c.projectID, id, collection)
	respData, err := c.apiClient.Delete(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to delete vector: %w", err)
	}

	var resp VectorDeleteResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("delete operation failed")
	}

	logger.DebugEvent().
		Str("collection", collection).
		Str("id", id).
		Msg("Vector deleted successfully")

	return nil
}

// ListCollections lists all vector collections in the project.
func (c *Client) ListCollections(ctx context.Context) ([]*VectorCollection, error) {
	logger.DebugEvent().Msg("Listing vector collections")

	path := fmt.Sprintf("/api/v1/projects/%s/vectors/collections", c.projectID)
	respData, err := c.apiClient.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to list collections: %w", err)
	}

	var resp ListCollectionsResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	logger.DebugEvent().
		Int("count", len(resp.Collections)).
		Msg("Collections listed successfully")

	return resp.Collections, nil
}

// QuantumEntangle entangles two vectors to create a quantum correlation.
func (c *Client) QuantumEntangle(ctx context.Context, vectorID1, vectorID2 string) (*QuantumEntangleResponse, error) {
	logger.InfoEvent().
		Str("vector_id_1", vectorID1).
		Str("vector_id_2", vectorID2).
		Msg("Entangling quantum vectors")

	if vectorID1 == "" || vectorID2 == "" {
		return nil, fmt.Errorf("both vector IDs are required")
	}

	req := &QuantumEntangleRequest{
		VectorID1: vectorID1,
		VectorID2: vectorID2,
	}

	path := fmt.Sprintf("/api/v1/projects/%s/quantum/entangle", c.projectID)
	respData, err := c.apiClient.Post(ctx, path, req)
	if err != nil {
		return nil, fmt.Errorf("failed to entangle vectors: %w", err)
	}

	var resp QuantumEntangleResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	logger.InfoEvent().
		Str("entanglement_id", resp.EntanglementID).
		Float64("correlation_score", resp.CorrelationScore).
		Msg("Vectors entangled successfully")

	return &resp, nil
}

// QuantumMeasure measures the quantum state of a vector.
func (c *Client) QuantumMeasure(ctx context.Context, vectorID string) (*QuantumMeasureResponse, error) {
	logger.DebugEvent().
		Str("vector_id", vectorID).
		Msg("Measuring quantum vector state")

	if vectorID == "" {
		return nil, fmt.Errorf("vector_id is required")
	}

	req := &QuantumMeasureRequest{
		VectorID: vectorID,
	}

	path := fmt.Sprintf("/api/v1/projects/%s/quantum/measure", c.projectID)
	respData, err := c.apiClient.Post(ctx, path, req)
	if err != nil {
		return nil, fmt.Errorf("failed to measure vector: %w", err)
	}

	var resp QuantumMeasureResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	logger.DebugEvent().
		Str("vector_id", vectorID).
		Str("quantum_state", resp.QuantumState).
		Float64("entropy", resp.Entropy).
		Float64("coherence", resp.Coherence).
		Msg("Vector measured successfully")

	return &resp, nil
}

// QuantumCompress compresses a vector using quantum compression techniques.
func (c *Client) QuantumCompress(ctx context.Context, vectorID string, compressionRatio float64) (*QuantumCompressResponse, error) {
	logger.InfoEvent().
		Str("vector_id", vectorID).
		Float64("compression_ratio", compressionRatio).
		Msg("Compressing vector with quantum techniques")

	if vectorID == "" {
		return nil, fmt.Errorf("vector_id is required")
	}
	if compressionRatio <= 0 || compressionRatio >= 1 {
		return nil, fmt.Errorf("compression_ratio must be between 0 and 1")
	}

	req := &QuantumCompressRequest{
		VectorID:         vectorID,
		CompressionRatio: compressionRatio,
	}

	path := fmt.Sprintf("/api/v1/projects/%s/quantum/compress", c.projectID)
	respData, err := c.apiClient.Post(ctx, path, req)
	if err != nil {
		return nil, fmt.Errorf("failed to compress vector: %w", err)
	}

	var resp QuantumCompressResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	logger.InfoEvent().
		Str("vector_id", vectorID).
		Int("original_dimension", resp.OriginalDimension).
		Int("compressed_dimension", resp.CompressedDimension).
		Float64("information_loss", resp.InformationLoss).
		Msg("Vector compressed successfully")

	return &resp, nil
}

// QuantumDecompress decompresses a previously compressed vector.
func (c *Client) QuantumDecompress(ctx context.Context, vectorID string) (*QuantumDecompressResponse, error) {
	logger.InfoEvent().
		Str("vector_id", vectorID).
		Msg("Decompressing quantum vector")

	if vectorID == "" {
		return nil, fmt.Errorf("vector_id is required")
	}

	req := &QuantumDecompressRequest{
		VectorID: vectorID,
	}

	path := fmt.Sprintf("/api/v1/projects/%s/quantum/decompress", c.projectID)
	respData, err := c.apiClient.Post(ctx, path, req)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress vector: %w", err)
	}

	var resp QuantumDecompressResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	logger.InfoEvent().
		Str("vector_id", vectorID).
		Int("decompressed_dimension", resp.DecompressedDimension).
		Float64("restoration_accuracy", resp.RestorationAccuracy).
		Msg("Vector decompressed successfully")

	return &resp, nil
}

// QuantumSearch performs quantum-enhanced vector similarity search.
func (c *Client) QuantumSearch(ctx context.Context, req *QuantumSearchRequest) ([]*QuantumSearchResult, error) {
	logger.DebugEvent().
		Int("vector_dimension", len(req.QueryVector)).
		Int("limit", req.Limit).
		Bool("use_quantum_boost", req.UseQuantumBoost).
		Msg("Performing quantum vector search")

	if len(req.QueryVector) == 0 {
		return nil, fmt.Errorf("query_vector is required")
	}

	if req.Limit == 0 {
		req.Limit = 10
	}

	path := fmt.Sprintf("/api/v1/projects/%s/quantum/search", c.projectID)
	respData, err := c.apiClient.Post(ctx, path, req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform quantum search: %w", err)
	}

	var resp QuantumSearchResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	logger.DebugEvent().
		Int("results", len(resp.Results)).
		Int("total", resp.Total).
		Bool("quantum_boost_used", resp.QuantumBoostUsed).
		Float64("latency_ms", resp.SearchLatency).
		Msg("Quantum search completed")

	return resp.Results, nil
}
