package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/AINative-studio/ainative-code/tests/integration/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestZeroDBVectorOperations tests vector CRUD operations
func TestZeroDBVectorOperations(t *testing.T) {
	t.Run("should upsert vector successfully", func(t *testing.T) {
		// Given: Mock ZeroDB server
		server := mocks.NewZeroDBServer()
		defer server.Close()

		// And: Vector data
		vectorData := map[string]interface{}{
			"id":     "vec_001",
			"vector": []float64{0.1, 0.2, 0.3, 0.4},
			"metadata": map[string]interface{}{
				"source": "test",
				"type":   "embedding",
			},
		}

		// When: Upserting vector
		body, _ := json.Marshal(vectorData)
		req, _ := http.NewRequest("POST", server.GetURL()+"/api/vectors/upsert", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", "test-key")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)

		// Then: Should succeed
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, server.VectorsCalled)
	})

	t.Run("should search vectors by similarity", func(t *testing.T) {
		// Given: Mock ZeroDB server with vectors
		server := mocks.NewZeroDBServer()
		defer server.Close()

		// And: Search query
		searchData := map[string]interface{}{
			"vector": []float64{0.1, 0.2, 0.3, 0.4},
			"limit":  10,
		}

		// When: Searching vectors
		body, _ := json.Marshal(searchData)
		req, _ := http.NewRequest("POST", server.GetURL()+"/api/vectors/search", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", "test-key")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)

		// Then: Should return results
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, server.VectorsCalled)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Contains(t, result, "results")
	})

	t.Run("should list all vectors", func(t *testing.T) {
		// Given: Mock ZeroDB server
		server := mocks.NewZeroDBServer()
		defer server.Close()

		// When: Listing vectors
		req, _ := http.NewRequest("GET", server.GetURL()+"/api/vectors/list", nil)
		req.Header.Set("X-API-Key", "test-key")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)

		// Then: Should return vector list
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, server.VectorsCalled)
	})

	t.Run("should reject empty vector", func(t *testing.T) {
		// Given: Mock ZeroDB server
		server := mocks.NewZeroDBServer()
		defer server.Close()

		// And: Invalid vector data
		vectorData := map[string]interface{}{
			"id":     "vec_002",
			"vector": []float64{},
		}

		// When: Upserting empty vector
		body, _ := json.Marshal(vectorData)
		req, _ := http.NewRequest("POST", server.GetURL()+"/api/vectors/upsert", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", "test-key")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)

		// Then: Should be rejected
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

// TestZeroDBTableOperations tests NoSQL table operations
func TestZeroDBTableOperations(t *testing.T) {
	t.Run("should create new table", func(t *testing.T) {
		// Given: Mock ZeroDB server
		server := mocks.NewZeroDBServer()
		defer server.Close()

		// And: Table creation request
		tableData := map[string]interface{}{
			"table_name": "sessions",
		}

		// When: Creating table
		body, _ := json.Marshal(tableData)
		req, _ := http.NewRequest("POST", server.GetURL()+"/api/tables/create", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", "test-key")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)

		// Then: Should succeed
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, server.TablesCalled)
	})

	t.Run("should insert record into table", func(t *testing.T) {
		// Given: Mock ZeroDB server with table
		server := mocks.NewZeroDBServer()
		defer server.Close()

		// Create table first
		server.Tables["sessions"] = make([]mocks.TableRecord, 0)

		// And: Record data
		recordData := map[string]interface{}{
			"table_name": "sessions",
			"data": map[string]interface{}{
				"user_id":    "user_123",
				"session_id": "sess_456",
				"created_at": time.Now().Unix(),
			},
		}

		// When: Inserting record
		body, _ := json.Marshal(recordData)
		req, _ := http.NewRequest("POST", server.GetURL()+"/api/tables/insert", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", "test-key")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)

		// Then: Should succeed
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("should query records from table", func(t *testing.T) {
		// Given: Mock ZeroDB server with table and data
		server := mocks.NewZeroDBServer()
		defer server.Close()

		server.Tables["sessions"] = []mocks.TableRecord{
			{
				ID: "rec_1",
				Data: map[string]interface{}{
					"user_id": "user_123",
				},
			},
		}

		// And: Query request
		queryData := map[string]interface{}{
			"table_name": "sessions",
			"limit":      10,
		}

		// When: Querying table
		body, _ := json.Marshal(queryData)
		req, _ := http.NewRequest("POST", server.GetURL()+"/api/tables/query", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", "test-key")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)

		// Then: Should return records
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Contains(t, result, "records")
	})

	t.Run("should update record in table", func(t *testing.T) {
		// Given: Mock ZeroDB server with table and record
		server := mocks.NewZeroDBServer()
		defer server.Close()

		server.Tables["sessions"] = []mocks.TableRecord{
			{
				ID: "rec_1",
				Data: map[string]interface{}{
					"user_id": "user_123",
				},
			},
		}

		// And: Update request
		updateData := map[string]interface{}{
			"table_name": "sessions",
			"record_id":  "rec_1",
			"data": map[string]interface{}{
				"user_id": "user_456",
				"updated": true,
			},
		}

		// When: Updating record
		body, _ := json.Marshal(updateData)
		req, _ := http.NewRequest("POST", server.GetURL()+"/api/tables/update", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", "test-key")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)

		// Then: Should succeed
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("should reject operation on non-existent table", func(t *testing.T) {
		// Given: Mock ZeroDB server
		server := mocks.NewZeroDBServer()
		defer server.Close()

		// And: Query on non-existent table
		queryData := map[string]interface{}{
			"table_name": "non_existent",
		}

		// When: Querying
		body, _ := json.Marshal(queryData)
		req, _ := http.NewRequest("POST", server.GetURL()+"/api/tables/query", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", "test-key")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)

		// Then: Should return 404
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

// TestZeroDBMemoryOperations tests agent memory operations
func TestZeroDBMemoryOperations(t *testing.T) {
	t.Run("should store agent memory", func(t *testing.T) {
		// Given: Mock ZeroDB server
		server := mocks.NewZeroDBServer()
		defer server.Close()

		// And: Memory data
		memoryData := map[string]interface{}{
			"session_id": "sess_123",
			"content":    "User asked about authentication",
			"vector":     []float64{0.5, 0.6, 0.7},
		}

		// When: Storing memory
		body, _ := json.Marshal(memoryData)
		req, _ := http.NewRequest("POST", server.GetURL()+"/api/memory/store", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", "test-key")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)

		// Then: Should succeed
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, server.MemoryCalled)
	})

	t.Run("should search agent memory", func(t *testing.T) {
		// Given: Mock ZeroDB server with memory
		server := mocks.NewZeroDBServer()
		defer server.Close()

		server.Memories = []mocks.MemoryRecord{
			{
				ID:        "mem_1",
				SessionID: "sess_123",
				Content:   "Previous conversation",
				Vector:    []float64{0.5, 0.6, 0.7},
			},
		}

		// And: Search request
		searchData := map[string]interface{}{
			"session_id": "sess_123",
			"vector":     []float64{0.5, 0.6, 0.7},
			"limit":      5,
		}

		// When: Searching memory
		body, _ := json.Marshal(searchData)
		req, _ := http.NewRequest("POST", server.GetURL()+"/api/memory/search", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", "test-key")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)

		// Then: Should return results
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Contains(t, result, "results")
	})
}

// TestZeroDBErrorHandling tests error scenarios
func TestZeroDBErrorHandling(t *testing.T) {
	t.Run("should handle 401 unauthorized", func(t *testing.T) {
		// Given: Mock server with auth failure
		server := mocks.NewZeroDBServer()
		defer server.Close()

		server.ShouldFailAuth = true

		// When: Making request without auth
		req, _ := http.NewRequest("GET", server.GetURL()+"/api/vectors/list", nil)
		req.Header.Set("X-API-Key", "invalid-key")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)

		// Then: Should return 401
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("should handle 429 rate limit", func(t *testing.T) {
		// Given: Mock server with rate limiting
		server := mocks.NewZeroDBServer()
		defer server.Close()

		server.ShouldRateLimit = true

		// When: Making request
		req, _ := http.NewRequest("GET", server.GetURL()+"/api/vectors/list", nil)
		req.Header.Set("X-API-Key", "test-key")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)

		// Then: Should return 429
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
		assert.NotEmpty(t, resp.Header.Get("Retry-After"))
	})

	t.Run("should validate API request format", func(t *testing.T) {
		// Given: Mock ZeroDB server
		server := mocks.NewZeroDBServer()
		defer server.Close()

		// When: Sending invalid JSON
		req, _ := http.NewRequest("POST", server.GetURL()+"/api/vectors/upsert", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", "test-key")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)

		// Then: Should return 400
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}
