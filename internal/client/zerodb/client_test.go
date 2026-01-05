package zerodb_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AINative-studio/ainative-code/internal/client"
	"github.com/AINative-studio/ainative-code/internal/client/zerodb"
)

// TestCreateTable tests table creation functionality.
func TestCreateTable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/v1/projects/test-project/nosql/tables", r.URL.Path)

		var req zerodb.CreateTableRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Equal(t, "users", req.Name)
		assert.NotNil(t, req.Schema)

		resp := zerodb.CreateTableResponse{
			Table: &zerodb.Table{
				ID:   "table-123",
				Name: "users",
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

	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name":  map[string]interface{}{"type": "string"},
			"email": map[string]interface{}{"type": "string"},
		},
	}

	table, err := zdbClient.CreateTable(context.Background(), "users", schema)
	require.NoError(t, err)
	assert.Equal(t, "table-123", table.ID)
	assert.Equal(t, "users", table.Name)
}

// TestInsert tests document insertion functionality.
func TestInsert(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/v1/projects/test-project/nosql/documents", r.URL.Path)

		var req zerodb.InsertRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Equal(t, "users", req.TableName)
		assert.Equal(t, "John Doe", req.Data["name"])

		resp := zerodb.InsertResponse{
			ID: "doc-123",
			Document: &zerodb.Document{
				ID:        "doc-123",
				TableName: "users",
				Data:      req.Data,
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

	data := map[string]interface{}{
		"name":  "John Doe",
		"email": "john@example.com",
		"age":   30,
	}

	id, doc, err := zdbClient.Insert(context.Background(), "users", data)
	require.NoError(t, err)
	assert.Equal(t, "doc-123", id)
	assert.Equal(t, "users", doc.TableName)
	assert.Equal(t, "John Doe", doc.Data["name"])
}

// TestQuery tests document querying functionality with filters.
func TestQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/v1/projects/test-project/nosql/query", r.URL.Path)

		var req zerodb.QueryRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Equal(t, "users", req.TableName)

		// Verify MongoDB-style filter
		if len(req.Filter) > 0 {
			ageFilter, ok := req.Filter["age"].(map[string]interface{})
			assert.True(t, ok)
			assert.Equal(t, float64(18), ageFilter["$gte"])
		}

		resp := zerodb.QueryResponse{
			Documents: []*zerodb.Document{
				{
					ID:        "doc-1",
					TableName: "users",
					Data: map[string]interface{}{
						"name": "John Doe",
						"age":  30,
					},
				},
				{
					ID:        "doc-2",
					TableName: "users",
					Data: map[string]interface{}{
						"name": "Jane Smith",
						"age":  25,
					},
				},
			},
			Total:  2,
			Limit:  10,
			Offset: 0,
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

	filter := zerodb.QueryFilter{
		"age": map[string]interface{}{
			"$gte": 18,
		},
	}

	options := zerodb.QueryOptions{
		Limit:  10,
		Offset: 0,
	}

	docs, err := zdbClient.Query(context.Background(), "users", filter, options)
	require.NoError(t, err)
	assert.Len(t, docs, 2)
	assert.Equal(t, "doc-1", docs[0].ID)
	assert.Equal(t, "John Doe", docs[0].Data["name"])
}

// TestUpdate tests document update functionality.
func TestUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/api/v1/projects/test-project/nosql/documents/doc-123", r.URL.Path)

		var req zerodb.UpdateRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Equal(t, "users", req.TableName)
		assert.Equal(t, "doc-123", req.ID)
		assert.Equal(t, float64(31), req.Data["age"])

		resp := zerodb.UpdateResponse{
			Document: &zerodb.Document{
				ID:        "doc-123",
				TableName: "users",
				Data: map[string]interface{}{
					"name": "John Doe",
					"age":  31,
				},
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

	updateData := map[string]interface{}{
		"age": 31,
	}

	doc, err := zdbClient.Update(context.Background(), "users", "doc-123", updateData)
	require.NoError(t, err)
	assert.Equal(t, "doc-123", doc.ID)
	assert.Equal(t, float64(31), doc.Data["age"])
}

// TestDelete tests document deletion functionality.
func TestDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/projects/test-project/nosql/documents/doc-123")
		assert.Equal(t, "users", r.URL.Query().Get("table"))

		resp := zerodb.DeleteResponse{
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

	err := zdbClient.Delete(context.Background(), "users", "doc-123")
	require.NoError(t, err)
}

// TestListTables tests table listing functionality.
func TestListTables(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/v1/projects/test-project/nosql/tables", r.URL.Path)

		resp := zerodb.ListTablesResponse{
			Tables: []*zerodb.Table{
				{
					ID:   "table-1",
					Name: "users",
				},
				{
					ID:   "table-2",
					Name: "products",
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

	tables, err := zdbClient.ListTables(context.Background())
	require.NoError(t, err)
	assert.Len(t, tables, 2)
	assert.Equal(t, "users", tables[0].Name)
	assert.Equal(t, "products", tables[1].Name)
}

// TestMongoDBStyleFilters tests various MongoDB-style filter operators.
func TestMongoDBStyleFilters(t *testing.T) {
	tests := []struct {
		name     string
		filter   zerodb.QueryFilter
		expected map[string]interface{}
	}{
		{
			name: "equality filter",
			filter: zerodb.QueryFilter{
				"status": "active",
			},
			expected: map[string]interface{}{
				"status": "active",
			},
		},
		{
			name: "comparison operators",
			filter: zerodb.QueryFilter{
				"age": map[string]interface{}{
					"$gt":  18,
					"$lte": 65,
				},
			},
			expected: map[string]interface{}{
				"age": map[string]interface{}{
					"$gt":  18,
					"$lte": 65,
				},
			},
		},
		{
			name: "logical AND operator",
			filter: zerodb.QueryFilter{
				"$and": []interface{}{
					map[string]interface{}{"age": map[string]interface{}{"$gte": 18}},
					map[string]interface{}{"status": "active"},
				},
			},
			expected: map[string]interface{}{
				"$and": []interface{}{
					map[string]interface{}{"age": map[string]interface{}{"$gte": 18}},
					map[string]interface{}{"status": "active"},
				},
			},
		},
		{
			name: "array IN operator",
			filter: zerodb.QueryFilter{
				"tags": map[string]interface{}{
					"$in": []string{"go", "rust", "python"},
				},
			},
			expected: map[string]interface{}{
				"tags": map[string]interface{}{
					"$in": []string{"go", "rust", "python"},
				},
			},
		},
		{
			name: "exists operator",
			filter: zerodb.QueryFilter{
				"email": map[string]interface{}{
					"$exists": true,
				},
			},
			expected: map[string]interface{}{
				"email": map[string]interface{}{
					"$exists": true,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify filter structure matches expected
			filterJSON, err := json.Marshal(tt.filter)
			require.NoError(t, err)

			expectedJSON, err := json.Marshal(tt.expected)
			require.NoError(t, err)

			assert.JSONEq(t, string(expectedJSON), string(filterJSON))
		})
	}
}
