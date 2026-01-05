// Package helpers provides test helper utilities for integration tests.
package helpers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// MockServerConfig holds configuration for the mock HTTP server.
type MockServerConfig struct {
	// Handlers maps request paths to handler functions
	Handlers map[string]http.HandlerFunc
	// DefaultHandler is used when no specific handler is found
	DefaultHandler http.HandlerFunc
}

// SetupMockServer creates a mock HTTP server for testing external API calls.
func SetupMockServer(t *testing.T, config *MockServerConfig) (*httptest.Server, func()) {
	t.Helper()

	if config == nil {
		config = &MockServerConfig{
			Handlers: make(map[string]http.HandlerFunc),
		}
	}

	if config.DefaultHandler == nil {
		config.DefaultHandler = func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "not found",
			})
		}
	}

	mux := http.NewServeMux()

	// Register all configured handlers
	for path, handler := range config.Handlers {
		mux.HandleFunc(path, handler)
	}

	// Register default handler for all other paths
	mux.HandleFunc("/", config.DefaultHandler)

	server := httptest.NewServer(mux)

	cleanup := func() {
		server.Close()
	}

	return server, cleanup
}

// MockAuthServer creates a mock OAuth server for authentication testing.
func MockAuthServer(t *testing.T) (*httptest.Server, func()) {
	t.Helper()

	config := &MockServerConfig{
		Handlers: map[string]http.HandlerFunc{
			"/oauth/token": func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					w.WriteHeader(http.StatusMethodNotAllowed)
					return
				}

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"access_token":  "mock_access_token",
					"refresh_token": "mock_refresh_token",
					"token_type":    "Bearer",
					"expires_in":    3600,
				})
			},
			"/oauth/authorize": func(w http.ResponseWriter, r *http.Request) {
				// Return authorization code
				w.Header().Set("Location", "http://localhost:8080/callback?code=mock_auth_code")
				w.WriteHeader(http.StatusFound)
			},
		},
	}

	return SetupMockServer(t, config)
}

// MockZeroDBServer creates a mock ZeroDB API server for testing.
func MockZeroDBServer(t *testing.T) (*httptest.Server, func()) {
	t.Helper()

	config := &MockServerConfig{
		Handlers: map[string]http.HandlerFunc{
			"/api/v1/projects/test-project/nosql/tables": func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case http.MethodPost:
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(map[string]interface{}{
						"table": map[string]interface{}{
							"id":   "table_123",
							"name": "test_table",
						},
					})
				case http.MethodGet:
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(map[string]interface{}{
						"tables": []map[string]interface{}{
							{"id": "table_123", "name": "test_table"},
						},
					})
				}
			},
			"/api/v1/projects/test-project/nosql/documents": func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"id": "doc_123",
					"document": map[string]interface{}{
						"data": "test_data",
					},
				})
			},
			"/api/v1/projects/test-project/nosql/query": func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"documents": []map[string]interface{}{
						{"id": "doc_123", "data": "test_data"},
					},
					"total": 1,
				})
			},
			"/api/v1/projects/test-project/memory/store": func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"memory": map[string]interface{}{
						"id":      "mem_123",
						"content": "test memory",
					},
				})
			},
			"/api/v1/projects/test-project/memory/retrieve": func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"memories": []map[string]interface{}{
						{"id": "mem_123", "content": "test memory"},
					},
					"total": 1,
				})
			},
			"/api/v1/projects/test-project/memory/list": func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"memories": []map[string]interface{}{},
					"total":    0,
				})
			},
			"/api/v1/projects/test-project/memory/clear": func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"deleted": 0,
				})
			},
			"/api/v1/projects/test-project/vectors/collections": func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case http.MethodPost:
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(map[string]interface{}{
						"collection": map[string]interface{}{
							"id":         "coll_123",
							"name":       "embeddings",
							"dimensions": 768,
						},
					})
				case http.MethodGet:
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(map[string]interface{}{
						"collections": []map[string]interface{}{
							{"id": "coll_123", "name": "embeddings"},
						},
					})
				}
			},
			"/api/v1/projects/test-project/vectors": func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"id": "vec_123",
					"vector": map[string]interface{}{
						"id": "vec_123",
					},
				})
			},
			"/api/v1/projects/test-project/vectors/search": func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"vectors": []map[string]interface{}{
						{"id": "vec_123", "score": 0.95},
					},
					"total": 1,
				})
			},
			"/api/v1/projects/test-project/quantum/entangle": func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"entanglement_id":   "ent_123",
					"correlation_score": 0.98,
				})
			},
			"/api/v1/projects/test-project/quantum/measure": func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"quantum_state": "superposition",
					"entropy":       0.75,
					"coherence":     0.85,
				})
			},
			"/api/v1/projects/test-project/quantum/compress": func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"original_dimension":   768,
					"compressed_dimension": 384,
					"information_loss":     0.05,
				})
			},
			"/api/v1/projects/test-project/quantum/decompress": func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"decompressed_dimension": 768,
					"restoration_accuracy":   0.95,
				})
			},
			"/api/v1/projects/test-project/quantum/search": func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"results": []map[string]interface{}{
						{"id": "vec_123", "score": 0.98},
					},
					"total":              1,
					"quantum_boost_used": true,
					"search_latency":     15.5,
				})
			},
		},
	}

	return SetupMockServer(t, config)
}

// MockStrapiServer creates a mock Strapi API server for testing.
func MockStrapiServer(t *testing.T) (*httptest.Server, func()) {
	t.Helper()

	config := &MockServerConfig{
		Handlers: map[string]http.HandlerFunc{
			"/api/blog-posts": func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case http.MethodPost:
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(map[string]interface{}{
						"data": map[string]interface{}{
							"id": 1,
							"attributes": map[string]interface{}{
								"title":   "Test Post",
								"content": "Test content",
							},
						},
					})
				case http.MethodGet:
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(map[string]interface{}{
						"data": []map[string]interface{}{
							{
								"id": 1,
								"attributes": map[string]interface{}{
									"title":   "Test Post",
									"content": "Test content",
								},
							},
						},
					})
				}
			},
		},
	}

	return SetupMockServer(t, config)
}
