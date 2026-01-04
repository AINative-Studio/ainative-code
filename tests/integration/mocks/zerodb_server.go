package mocks

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"
)

// ZeroDBServer represents a mock ZeroDB API server
type ZeroDBServer struct {
	Server           *httptest.Server
	Vectors          map[string]VectorRecord
	Tables           map[string][]TableRecord
	Memories         []MemoryRecord
	mu               sync.RWMutex
	ShouldFailAuth   bool
	ShouldRateLimit  bool
	ShouldTimeout    bool
	ResponseDelay    time.Duration
	VectorsCalled    bool
	TablesCalled     bool
	MemoryCalled     bool
}

// VectorRecord represents a vector embedding with metadata
type VectorRecord struct {
	ID        string                 `json:"id"`
	Vector    []float64              `json:"vector"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt time.Time              `json:"created_at"`
}

// TableRecord represents a NoSQL table record
type TableRecord struct {
	ID        string                 `json:"id"`
	Data      map[string]interface{} `json:"data"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// MemoryRecord represents an agent memory record
type MemoryRecord struct {
	ID        string    `json:"id"`
	SessionID string    `json:"session_id"`
	Content   string    `json:"content"`
	Vector    []float64 `json:"vector"`
	Timestamp time.Time `json:"timestamp"`
}

// NewZeroDBServer creates a new mock ZeroDB server
func NewZeroDBServer() *ZeroDBServer {
	zs := &ZeroDBServer{
		Vectors:  make(map[string]VectorRecord),
		Tables:   make(map[string][]TableRecord),
		Memories: make([]MemoryRecord, 0),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/vectors/upsert", zs.handleVectorUpsert)
	mux.HandleFunc("/api/vectors/search", zs.handleVectorSearch)
	mux.HandleFunc("/api/vectors/list", zs.handleVectorList)
	mux.HandleFunc("/api/tables/create", zs.handleTableCreate)
	mux.HandleFunc("/api/tables/insert", zs.handleTableInsert)
	mux.HandleFunc("/api/tables/query", zs.handleTableQuery)
	mux.HandleFunc("/api/tables/update", zs.handleTableUpdate)
	mux.HandleFunc("/api/memory/store", zs.handleMemoryStore)
	mux.HandleFunc("/api/memory/search", zs.handleMemorySearch)

	zs.Server = httptest.NewServer(zs.authMiddleware(mux))
	return zs
}

// Close shuts down the mock server
func (zs *ZeroDBServer) Close() {
	zs.Server.Close()
}

// authMiddleware validates API authentication
func (zs *ZeroDBServer) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if zs.ResponseDelay > 0 {
			time.Sleep(zs.ResponseDelay)
		}

		if zs.ShouldFailAuth {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "unauthorized",
				"message": "Invalid API key",
			})
			return
		}

		if zs.ShouldRateLimit {
			w.Header().Set("Retry-After", "60")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "rate_limit_exceeded",
				"message": "Too many requests",
			})
			return
		}

		if zs.ShouldTimeout {
			time.Sleep(10 * time.Second)
			return
		}

		// Validate API key header
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "missing_api_key",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}

// handleVectorUpsert handles vector upsert requests
func (zs *ZeroDBServer) handleVectorUpsert(w http.ResponseWriter, r *http.Request) {
	zs.VectorsCalled = true
	zs.mu.Lock()
	defer zs.mu.Unlock()

	var req struct {
		ID       string                 `json:"id"`
		Vector   []float64              `json:"vector"`
		Metadata map[string]interface{} `json:"metadata"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid_request"})
		return
	}

	// Validate vector dimensions
	if len(req.Vector) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "empty_vector"})
		return
	}

	record := VectorRecord{
		ID:        req.ID,
		Vector:    req.Vector,
		Metadata:  req.Metadata,
		CreatedAt: time.Now(),
	}

	zs.Vectors[req.ID] = record

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"id":      req.ID,
	})
}

// handleVectorSearch handles vector similarity search
func (zs *ZeroDBServer) handleVectorSearch(w http.ResponseWriter, r *http.Request) {
	zs.VectorsCalled = true
	zs.mu.RLock()
	defer zs.mu.RUnlock()

	var req struct {
		Vector []float64 `json:"vector"`
		Limit  int       `json:"limit"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid_request"})
		return
	}

	if req.Limit == 0 {
		req.Limit = 10
	}

	// Mock search results (return all vectors with mock similarity scores)
	results := make([]map[string]interface{}, 0)
	for id, vec := range zs.Vectors {
		results = append(results, map[string]interface{}{
			"id":         id,
			"score":      0.95, // Mock similarity score
			"metadata":   vec.Metadata,
			"created_at": vec.CreatedAt,
		})
		if len(results) >= req.Limit {
			break
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"results": results,
		"count":   len(results),
	})
}

// handleVectorList lists all vectors
func (zs *ZeroDBServer) handleVectorList(w http.ResponseWriter, r *http.Request) {
	zs.VectorsCalled = true
	zs.mu.RLock()
	defer zs.mu.RUnlock()

	vectors := make([]map[string]interface{}, 0)
	for id, vec := range zs.Vectors {
		vectors = append(vectors, map[string]interface{}{
			"id":         id,
			"metadata":   vec.Metadata,
			"created_at": vec.CreatedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"vectors": vectors,
		"count":   len(vectors),
	})
}

// handleTableCreate creates a new table
func (zs *ZeroDBServer) handleTableCreate(w http.ResponseWriter, r *http.Request) {
	zs.TablesCalled = true
	zs.mu.Lock()
	defer zs.mu.Unlock()

	var req struct {
		TableName string `json:"table_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid_request"})
		return
	}

	if req.TableName == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "table_name_required"})
		return
	}

	// Check if table already exists
	if _, exists := zs.Tables[req.TableName]; exists {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{"error": "table_exists"})
		return
	}

	zs.Tables[req.TableName] = make([]TableRecord, 0)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"table_name": req.TableName,
	})
}

// handleTableInsert inserts data into a table
func (zs *ZeroDBServer) handleTableInsert(w http.ResponseWriter, r *http.Request) {
	zs.TablesCalled = true
	zs.mu.Lock()
	defer zs.mu.Unlock()

	var req struct {
		TableName string                 `json:"table_name"`
		Data      map[string]interface{} `json:"data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid_request"})
		return
	}

	table, exists := zs.Tables[req.TableName]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "table_not_found"})
		return
	}

	record := TableRecord{
		ID:        fmt.Sprintf("rec_%d", len(table)+1),
		Data:      req.Data,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	zs.Tables[req.TableName] = append(table, record)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"id":      record.ID,
	})
}

// handleTableQuery queries data from a table
func (zs *ZeroDBServer) handleTableQuery(w http.ResponseWriter, r *http.Request) {
	zs.TablesCalled = true
	zs.mu.RLock()
	defer zs.mu.RUnlock()

	var req struct {
		TableName string `json:"table_name"`
		Limit     int    `json:"limit"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid_request"})
		return
	}

	table, exists := zs.Tables[req.TableName]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "table_not_found"})
		return
	}

	if req.Limit == 0 {
		req.Limit = 100
	}

	results := table
	if len(results) > req.Limit {
		results = results[:req.Limit]
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"records": results,
		"count":   len(results),
	})
}

// handleTableUpdate updates a record in a table
func (zs *ZeroDBServer) handleTableUpdate(w http.ResponseWriter, r *http.Request) {
	zs.TablesCalled = true
	zs.mu.Lock()
	defer zs.mu.Unlock()

	var req struct {
		TableName string                 `json:"table_name"`
		RecordID  string                 `json:"record_id"`
		Data      map[string]interface{} `json:"data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid_request"})
		return
	}

	table, exists := zs.Tables[req.TableName]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "table_not_found"})
		return
	}

	// Find and update record
	found := false
	for i, record := range table {
		if record.ID == req.RecordID {
			record.Data = req.Data
			record.UpdatedAt = time.Now()
			zs.Tables[req.TableName][i] = record
			found = true
			break
		}
	}

	if !found {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "record_not_found"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

// handleMemoryStore stores agent memory
func (zs *ZeroDBServer) handleMemoryStore(w http.ResponseWriter, r *http.Request) {
	zs.MemoryCalled = true
	zs.mu.Lock()
	defer zs.mu.Unlock()

	var req struct {
		SessionID string    `json:"session_id"`
		Content   string    `json:"content"`
		Vector    []float64 `json:"vector"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid_request"})
		return
	}

	memory := MemoryRecord{
		ID:        fmt.Sprintf("mem_%d", len(zs.Memories)+1),
		SessionID: req.SessionID,
		Content:   req.Content,
		Vector:    req.Vector,
		Timestamp: time.Now(),
	}

	zs.Memories = append(zs.Memories, memory)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"id":      memory.ID,
	})
}

// handleMemorySearch searches agent memory
func (zs *ZeroDBServer) handleMemorySearch(w http.ResponseWriter, r *http.Request) {
	zs.MemoryCalled = true
	zs.mu.RLock()
	defer zs.mu.RUnlock()

	var req struct {
		SessionID string    `json:"session_id"`
		Vector    []float64 `json:"vector"`
		Limit     int       `json:"limit"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid_request"})
		return
	}

	if req.Limit == 0 {
		req.Limit = 10
	}

	// Filter by session and return mock results
	results := make([]map[string]interface{}, 0)
	for _, mem := range zs.Memories {
		if mem.SessionID == req.SessionID {
			results = append(results, map[string]interface{}{
				"id":        mem.ID,
				"content":   mem.Content,
				"score":     0.92,
				"timestamp": mem.Timestamp,
			})
			if len(results) >= req.Limit {
				break
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"results": results,
		"count":   len(results),
	})
}

// GetURL returns the base URL of the mock server
func (zs *ZeroDBServer) GetURL() string {
	return zs.Server.URL
}

// Reset clears all stored data
func (zs *ZeroDBServer) Reset() {
	zs.mu.Lock()
	defer zs.mu.Unlock()
	zs.Vectors = make(map[string]VectorRecord)
	zs.Tables = make(map[string][]TableRecord)
	zs.Memories = make([]MemoryRecord, 0)
}
