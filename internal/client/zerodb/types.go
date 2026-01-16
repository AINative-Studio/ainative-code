package zerodb

import "time"

// Table represents a ZeroDB NoSQL table.
type Table struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Schema    map[string]interface{} `json:"schema"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// Document represents a document in a ZeroDB table.
type Document struct {
	ID        string                 `json:"id"`
	TableName string                 `json:"table_name"`
	Data      map[string]interface{} `json:"data"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// QueryFilter represents a MongoDB-style query filter.
//
// Examples:
//
//	// Simple equality
//	{"name": "John"}
//
//	// Comparison operators
//	{"age": {"$gt": 18, "$lt": 65}}
//
//	// Logical operators
//	{"$and": [{"age": {"$gte": 18}}, {"status": "active"}]}
//
//	// Array operators
//	{"tags": {"$in": ["go", "rust"]}}
//
//	// Existence check
//	{"email": {"$exists": true}}
type QueryFilter map[string]interface{}

// QueryOptions represents options for querying documents.
type QueryOptions struct {
	Limit  int                    `json:"limit,omitempty"`
	Offset int                    `json:"offset,omitempty"`
	Sort   map[string]int         `json:"sort,omitempty"` // 1 for asc, -1 for desc
	Fields map[string]interface{} `json:"fields,omitempty"`
}

// CreateTableRequest represents a request to create a new table.
type CreateTableRequest struct {
	Name   string                 `json:"name"`
	Schema map[string]interface{} `json:"schema"`
}

// CreateTableResponse represents the response from creating a table.
type CreateTableResponse struct {
	Table *Table `json:"table"`
}

// InsertRequest represents a request to insert a document.
type InsertRequest struct {
	TableName string                 `json:"table_name"`
	Data      map[string]interface{} `json:"data"`
}

// InsertResponse represents the response from inserting a document.
type InsertResponse struct {
	ID       string    `json:"id"`
	Document *Document `json:"document"`
}

// QueryRequest represents a request to query documents.
type QueryRequest struct {
	TableName string       `json:"table_name"`
	Filter    QueryFilter  `json:"filter,omitempty"`
	Options   QueryOptions `json:"options,omitempty"`
}

// QueryResponse represents the response from querying documents.
type QueryResponse struct {
	Documents []*Document `json:"documents"`
	Total     int         `json:"total"`
	Offset    int         `json:"offset"`
	Limit     int         `json:"limit"`
}

// UpdateRequest represents a request to update a document.
type UpdateRequest struct {
	TableName string                 `json:"table_name"`
	ID        string                 `json:"id"`
	Data      map[string]interface{} `json:"data"`
}

// UpdateResponse represents the response from updating a document.
type UpdateResponse struct {
	Document *Document `json:"document"`
}

// DeleteRequest represents a request to delete a document.
type DeleteRequest struct {
	TableName string `json:"table_name"`
	ID        string `json:"id"`
}

// DeleteResponse represents the response from deleting a document.
type DeleteResponse struct {
	Success bool `json:"success"`
}

// ListTablesResponse represents the response from listing tables.
type ListTablesResponse struct {
	Tables []*Table `json:"tables"`
	Total  int      `json:"total"`
}

// Memory represents a stored memory entry for agent memory operations.
type Memory struct {
	ID         string                 `json:"id"`
	AgentID    string                 `json:"agent_id"`
	SessionID  string                 `json:"session_id,omitempty"`
	Content    string                 `json:"content"`
	Role       string                 `json:"role,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Similarity float64                `json:"similarity,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

// MemoryStoreRequest represents a request to store agent memory.
type MemoryStoreRequest struct {
	AgentID   string                 `json:"agent_id"`
	Content   string                 `json:"content"`
	Role      string                 `json:"role,omitempty"`
	SessionID string                 `json:"session_id,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// MemoryStoreResponse represents the response from storing memory.
type MemoryStoreResponse struct {
	Memory *Memory `json:"memory"`
}

// MemoryRetrieveRequest represents a request to retrieve agent memories using semantic search.
type MemoryRetrieveRequest struct {
	AgentID   string `json:"agent_id"`
	Query     string `json:"query"`
	Limit     int    `json:"limit,omitempty"`
	SessionID string `json:"session_id,omitempty"`
}

// MemoryRetrieveResponse represents the response from retrieving memories.
type MemoryRetrieveResponse struct {
	Memories []*Memory `json:"memories"`
	Total    int       `json:"total"`
}

// MemoryClearRequest represents a request to clear agent memories.
type MemoryClearRequest struct {
	AgentID   string `json:"agent_id"`
	SessionID string `json:"session_id,omitempty"`
}

// MemoryClearResponse represents the response from clearing memories.
type MemoryClearResponse struct {
	Deleted int    `json:"deleted"`
	Message string `json:"message"`
}

// MemoryListRequest represents a request to list agent memories.
type MemoryListRequest struct {
	AgentID   string `json:"agent_id"`
	SessionID string `json:"session_id,omitempty"`
	Limit     int    `json:"limit,omitempty"`
	Offset    int    `json:"offset,omitempty"`
}

// MemoryListResponse represents the response from listing memories.
type MemoryListResponse struct {
	Memories []*Memory `json:"memories"`
	Total    int       `json:"total"`
	Limit    int       `json:"limit"`
	Offset   int       `json:"offset"`
}

// EmbedAndStoreRequest represents a request to embed and store documents.
// The actual API expects texts as an array of strings, not document objects.
type EmbedAndStoreRequest struct {
	Texts     []string                 `json:"texts"`
	Namespace string                   `json:"namespace,omitempty"`
	Metadata  []map[string]interface{} `json:"metadata,omitempty"`
}

// EmbedAndStoreResponse represents the response from embed-and-store.
type EmbedAndStoreResponse struct {
	Success              bool    `json:"success"`
	VectorsStored        int     `json:"vectors_stored"`
	EmbeddingsGenerated  int     `json:"embeddings_generated"`
	Model                string  `json:"model"`
	Dimensions           int     `json:"dimensions"`
	TargetColumn         string  `json:"target_column"`
	Namespace            string  `json:"namespace"`
	ProjectID            string  `json:"project_id"`
	ProcessingTimeMs     float64 `json:"processing_time_ms"`
}

// SearchEmbeddingsRequest represents a request to search embeddings.
type SearchEmbeddingsRequest struct {
	Query     string                 `json:"query"`
	TopK      int                    `json:"top_k,omitempty"`
	Namespace string                 `json:"namespace,omitempty"`
	Filter    map[string]interface{} `json:"filter,omitempty"`
}

// SearchResult represents a single search result from the embeddings API.
type SearchResult struct {
	VectorID       string                 `json:"vector_id"`
	Document       string                 `json:"document"`
	VectorMetadata map[string]interface{} `json:"vector_metadata"`
	Namespace      string                 `json:"namespace"`
	ProjectID      string                 `json:"project_id"`
	Source         *string                `json:"source"`
	CreatedAt      string                 `json:"created_at"`
	Score          float64                `json:"score,omitempty"` // May not always be present
}

// SearchEmbeddingsResponse represents the response from search.
type SearchEmbeddingsResponse struct {
	Results []SearchResult `json:"results"`
	Total   int            `json:"total"`
}

// VectorCollection represents a vector collection for storing embeddings.
type VectorCollection struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Dimensions int       `json:"dimensions"`
	Metric     string    `json:"metric,omitempty"`     // similarity metric (cosine, euclidean, dot_product)
	Count      int       `json:"count,omitempty"`      // number of vectors in collection
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Vector represents a vector embedding with metadata.
type Vector struct {
	ID         string                 `json:"id"`
	Collection string                 `json:"collection"`
	Vector     []float64              `json:"vector"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Score      float64                `json:"score,omitempty"` // similarity score for search results
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

// CreateCollectionRequest represents a request to create a vector collection.
type CreateCollectionRequest struct {
	Name       string `json:"name"`
	Dimensions int    `json:"dimensions"`
	Metric     string `json:"metric,omitempty"` // cosine (default), euclidean, dot_product
}

// CreateCollectionResponse represents the response from creating a collection.
type CreateCollectionResponse struct {
	Collection *VectorCollection `json:"collection"`
}

// VectorInsertRequest represents a request to insert a vector.
type VectorInsertRequest struct {
	Collection string                 `json:"collection"`
	Vector     []float64              `json:"vector"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	ID         string                 `json:"id,omitempty"` // optional ID for upsert behavior
}

// VectorInsertResponse represents the response from inserting a vector.
type VectorInsertResponse struct {
	ID     string  `json:"id"`
	Vector *Vector `json:"vector"`
}

// VectorSearchRequest represents a request to search for similar vectors.
type VectorSearchRequest struct {
	Collection  string      `json:"collection"`
	QueryVector []float64   `json:"query_vector"`
	Limit       int         `json:"limit,omitempty"`
	Filter      QueryFilter `json:"filter,omitempty"` // metadata filter
}

// VectorSearchResponse represents the response from searching vectors.
type VectorSearchResponse struct {
	Vectors []*Vector `json:"vectors"`
	Total   int       `json:"total"`
}

// VectorDeleteRequest represents a request to delete a vector.
type VectorDeleteRequest struct {
	Collection string `json:"collection"`
	ID         string `json:"id"`
}

// VectorDeleteResponse represents the response from deleting a vector.
type VectorDeleteResponse struct {
	Success bool `json:"success"`
}

// ListCollectionsResponse represents the response from listing vector collections.
type ListCollectionsResponse struct {
	Collections []*VectorCollection `json:"collections"`
	Total       int                 `json:"total"`
}

// QuantumVector represents a quantum-enhanced vector with entanglement state.
type QuantumVector struct {
	ID             string                 `json:"id"`
	Vector         []float64              `json:"vector"`
	Dimension      int                    `json:"dimension"`
	IsEntangled    bool                   `json:"is_entangled"`
	EntangledWith  []string               `json:"entangled_with,omitempty"`
	QuantumState   string                 `json:"quantum_state,omitempty"`
	CompressionRatio float64              `json:"compression_ratio,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// QuantumEntangleRequest represents a request to entangle two vectors.
type QuantumEntangleRequest struct {
	VectorID1 string `json:"vector_id_1"`
	VectorID2 string `json:"vector_id_2"`
}

// QuantumEntangleResponse represents the response from entangling vectors.
type QuantumEntangleResponse struct {
	Vector1         *QuantumVector `json:"vector_1"`
	Vector2         *QuantumVector `json:"vector_2"`
	EntanglementID  string         `json:"entanglement_id"`
	CorrelationScore float64       `json:"correlation_score"`
	Message         string         `json:"message"`
}

// QuantumMeasureRequest represents a request to measure a quantum vector.
type QuantumMeasureRequest struct {
	VectorID string `json:"vector_id"`
}

// QuantumMeasureResponse represents the response from measuring a vector.
type QuantumMeasureResponse struct {
	Vector          *QuantumVector         `json:"vector"`
	QuantumState    string                 `json:"quantum_state"`
	Entropy         float64                `json:"entropy"`
	Coherence       float64                `json:"coherence"`
	Properties      map[string]interface{} `json:"properties"`
}

// QuantumCompressRequest represents a request to compress a vector using quantum techniques.
type QuantumCompressRequest struct {
	VectorID         string  `json:"vector_id"`
	CompressionRatio float64 `json:"compression_ratio"`
}

// QuantumCompressResponse represents the response from compressing a vector.
type QuantumCompressResponse struct {
	Vector            *QuantumVector `json:"vector"`
	OriginalDimension int            `json:"original_dimension"`
	CompressedDimension int          `json:"compressed_dimension"`
	CompressionRatio  float64        `json:"compression_ratio"`
	InformationLoss   float64        `json:"information_loss"`
	Message           string         `json:"message"`
}

// QuantumDecompressRequest represents a request to decompress a vector.
type QuantumDecompressRequest struct {
	VectorID string `json:"vector_id"`
}

// QuantumDecompressResponse represents the response from decompressing a vector.
type QuantumDecompressResponse struct {
	Vector              *QuantumVector `json:"vector"`
	OriginalDimension   int            `json:"original_dimension"`
	DecompressedDimension int          `json:"decompressed_dimension"`
	RestorationAccuracy float64        `json:"restoration_accuracy"`
	Message             string         `json:"message"`
}

// QuantumSearchRequest represents a request to perform quantum-enhanced vector search.
type QuantumSearchRequest struct {
	QueryVector      []float64              `json:"query_vector"`
	Limit            int                    `json:"limit,omitempty"`
	UseQuantumBoost  bool                   `json:"use_quantum_boost,omitempty"`
	IncludeEntangled bool                   `json:"include_entangled,omitempty"`
	Filters          map[string]interface{} `json:"filters,omitempty"`
}

// QuantumSearchResult represents a single search result with quantum metrics.
type QuantumSearchResult struct {
	Vector            *QuantumVector `json:"vector"`
	Similarity        float64        `json:"similarity"`
	QuantumSimilarity float64        `json:"quantum_similarity,omitempty"`
	Rank              int            `json:"rank"`
}

// QuantumSearchResponse represents the response from quantum vector search.
type QuantumSearchResponse struct {
	Results          []*QuantumSearchResult `json:"results"`
	Total            int                    `json:"total"`
	QuantumBoostUsed bool                   `json:"quantum_boost_used"`
	SearchLatency    float64                `json:"search_latency_ms"`
}
