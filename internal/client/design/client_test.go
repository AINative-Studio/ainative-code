package design

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AINative-studio/ainative-code/internal/client"
	"github.com/AINative-studio/ainative-code/internal/design"
)

// TestUploadTokens tests the UploadTokens method.
func TestUploadTokens(t *testing.T) {
	tests := []struct {
		name               string
		tokens             []*design.Token
		conflictResolution design.ConflictResolutionStrategyUpload
		mockResponse       UploadTokensResponse
		mockStatusCode     int
		expectError        bool
		expectedUploaded   int
	}{
		{
			name: "successful upload with overwrite",
			tokens: []*design.Token{
				{
					Name:     "primary-color",
					Value:    "#007bff",
					Type:     "color",
					Category: "colors",
				},
				{
					Name:     "font-size-base",
					Value:    "16px",
					Type:     "font-size",
					Category: "typography",
				},
			},
			conflictResolution: design.ConflictOverwrite,
			mockResponse: UploadTokensResponse{
				Success:       true,
				UploadedCount: 2,
				SkippedCount:  0,
				UpdatedCount:  0,
			},
			mockStatusCode:   http.StatusOK,
			expectError:      false,
			expectedUploaded: 2,
		},
		{
			name: "upload with merge conflict resolution",
			tokens: []*design.Token{
				{
					Name:     "primary-color",
					Value:    "#0056b3",
					Type:     "color",
					Category: "colors",
				},
			},
			conflictResolution: design.ConflictMerge,
			mockResponse: UploadTokensResponse{
				Success:       true,
				UploadedCount: 0,
				SkippedCount:  0,
				UpdatedCount:  1,
			},
			mockStatusCode:   http.StatusOK,
			expectError:      false,
			expectedUploaded: 0,
		},
		{
			name: "upload with skip conflict resolution",
			tokens: []*design.Token{
				{
					Name:     "primary-color",
					Value:    "#007bff",
					Type:     "color",
					Category: "colors",
				},
			},
			conflictResolution: design.ConflictSkip,
			mockResponse: UploadTokensResponse{
				Success:       true,
				UploadedCount: 0,
				SkippedCount:  1,
				UpdatedCount:  0,
			},
			mockStatusCode:   http.StatusOK,
			expectError:      false,
			expectedUploaded: 0,
		},
		{
			name:               "empty token list",
			tokens:             []*design.Token{},
			conflictResolution: design.ConflictOverwrite,
			mockStatusCode:     http.StatusOK,
			expectError:        true,
		},
		{
			name: "server error",
			tokens: []*design.Token{
				{
					Name:     "primary-color",
					Value:    "#007bff",
					Type:     "color",
					Category: "colors",
				},
			},
			conflictResolution: design.ConflictOverwrite,
			mockStatusCode:     http.StatusInternalServerError,
			expectError:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request method and path
				if r.Method != http.MethodPost {
					t.Errorf("expected POST request, got %s", r.Method)
				}
				if r.URL.Path != "/api/v1/design/tokens/upload" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}

				// Verify request body
				var req UploadTokensRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Errorf("failed to decode request body: %v", err)
				}

				if req.ProjectID != "test-project" {
					t.Errorf("expected project_id 'test-project', got '%s'", req.ProjectID)
				}

				if req.ConflictResolution != tt.conflictResolution {
					t.Errorf("expected conflict_resolution '%s', got '%s'", tt.conflictResolution, req.ConflictResolution)
				}

				// Send mock response
				w.WriteHeader(tt.mockStatusCode)
				if tt.mockStatusCode == http.StatusOK {
					json.NewEncoder(w).Encode(tt.mockResponse)
				} else {
					json.NewEncoder(w).Encode(map[string]string{"error": "server error"})
				}
			}))
			defer server.Close()

			// Create client
			apiClient := client.New(
				client.WithBaseURL(server.URL),
				client.WithTimeout(5*time.Second),
			)

			designClient := New(
				WithAPIClient(apiClient),
				WithProjectID("test-project"),
			)

			// Call UploadTokens
			ctx := context.Background()
			result, err := designClient.UploadTokens(ctx, tt.tokens, tt.conflictResolution, nil)

			// Verify results
			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result.UploadedCount != tt.expectedUploaded {
				t.Errorf("expected %d uploaded, got %d", tt.expectedUploaded, result.UploadedCount)
			}
		})
	}
}

// TestUploadTokensWithProgress tests upload with progress callback.
func TestUploadTokensWithProgress(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(UploadTokensResponse{
			Success:       true,
			UploadedCount: 10,
			SkippedCount:  0,
			UpdatedCount:  0,
		})
	}))
	defer server.Close()

	// Create client
	apiClient := client.New(
		client.WithBaseURL(server.URL),
		client.WithTimeout(5*time.Second),
	)

	designClient := New(
		WithAPIClient(apiClient),
		WithProjectID("test-project"),
	)

	// Create test tokens
	tokens := make([]*design.Token, 10)
	for i := 0; i < 10; i++ {
		tokens[i] = &design.Token{
			Name:     "token-" + string(rune(i)),
			Value:    "#000000",
			Type:     "color",
			Category: "colors",
		}
	}

	// Track progress callbacks
	progressCalls := 0
	progressCallback := func(uploaded, total int) {
		progressCalls++
		if uploaded > total {
			t.Errorf("uploaded (%d) should not exceed total (%d)", uploaded, total)
		}
	}

	// Call UploadTokens with progress callback
	ctx := context.Background()
	_, err := designClient.UploadTokens(ctx, tokens, design.ConflictOverwrite, progressCallback)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if progressCalls == 0 {
		t.Error("progress callback was not called")
	}
}

// TestGetTokens tests the GetTokens method.
func TestGetTokens(t *testing.T) {
	tests := []struct {
		name           string
		types          []string
		category       string
		limit          int
		offset         int
		mockTokens     []*design.Token
		mockTotal      int
		mockStatusCode int
		expectError    bool
	}{
		{
			name:     "get all tokens",
			types:    nil,
			category: "",
			limit:    100,
			offset:   0,
			mockTokens: []*design.Token{
				{
					Name:     "primary-color",
					Value:    "#007bff",
					Type:     "color",
					Category: "colors",
				},
				{
					Name:     "font-size-base",
					Value:    "16px",
					Type:     "font-size",
					Category: "typography",
				},
			},
			mockTotal:      2,
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:     "get tokens by type",
			types:    []string{"color"},
			category: "",
			limit:    100,
			offset:   0,
			mockTokens: []*design.Token{
				{
					Name:     "primary-color",
					Value:    "#007bff",
					Type:     "color",
					Category: "colors",
				},
			},
			mockTotal:      1,
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:     "get tokens by category",
			types:    nil,
			category: "colors",
			limit:    100,
			offset:   0,
			mockTokens: []*design.Token{
				{
					Name:     "primary-color",
					Value:    "#007bff",
					Type:     "color",
					Category: "colors",
				},
			},
			mockTotal:      1,
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "server error",
			types:          nil,
			category:       "",
			limit:          100,
			offset:         0,
			mockStatusCode: http.StatusInternalServerError,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request
				if r.Method != http.MethodPost {
					t.Errorf("expected POST request, got %s", r.Method)
				}

				// Send mock response
				w.WriteHeader(tt.mockStatusCode)
				if tt.mockStatusCode == http.StatusOK {
					json.NewEncoder(w).Encode(TokenQueryResponse{
						Tokens: tt.mockTokens,
						Total:  tt.mockTotal,
					})
				} else {
					json.NewEncoder(w).Encode(map[string]string{"error": "server error"})
				}
			}))
			defer server.Close()

			// Create client
			apiClient := client.New(
				client.WithBaseURL(server.URL),
				client.WithTimeout(5*time.Second),
			)

			designClient := New(
				WithAPIClient(apiClient),
				WithProjectID("test-project"),
			)

			// Call GetTokens
			ctx := context.Background()
			tokens, total, err := designClient.GetTokens(ctx, tt.types, tt.category, tt.limit, tt.offset)

			// Verify results
			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(tokens) != len(tt.mockTokens) {
				t.Errorf("expected %d tokens, got %d", len(tt.mockTokens), len(tokens))
			}

			if total != tt.mockTotal {
				t.Errorf("expected total %d, got %d", tt.mockTotal, total)
			}
		})
	}
}

// TestDeleteToken tests the DeleteToken method.
func TestDeleteToken(t *testing.T) {
	tests := []struct {
		name           string
		tokenName      string
		mockSuccess    bool
		mockStatusCode int
		expectError    bool
	}{
		{
			name:           "successful delete",
			tokenName:      "primary-color",
			mockSuccess:    true,
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "empty token name",
			tokenName:      "",
			mockStatusCode: http.StatusOK,
			expectError:    true,
		},
		{
			name:           "delete failed",
			tokenName:      "primary-color",
			mockSuccess:    false,
			mockStatusCode: http.StatusOK,
			expectError:    true,
		},
		{
			name:           "server error",
			tokenName:      "primary-color",
			mockStatusCode: http.StatusInternalServerError,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Send mock response
				w.WriteHeader(tt.mockStatusCode)
				if tt.mockStatusCode == http.StatusOK {
					json.NewEncoder(w).Encode(DeleteTokenResponse{
						Success: tt.mockSuccess,
						Message: "Token deleted",
					})
				} else {
					json.NewEncoder(w).Encode(map[string]string{"error": "server error"})
				}
			}))
			defer server.Close()

			// Create client
			apiClient := client.New(
				client.WithBaseURL(server.URL),
				client.WithTimeout(5*time.Second),
			)

			designClient := New(
				WithAPIClient(apiClient),
				WithProjectID("test-project"),
			)

			// Call DeleteToken
			ctx := context.Background()
			err := designClient.DeleteToken(ctx, tt.tokenName)

			// Verify results
			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestValidateTokens tests the ValidateTokens method.
func TestValidateTokens(t *testing.T) {
	designClient := New()

	tests := []struct {
		name        string
		tokens      []*design.Token
		expectValid bool
		expectError int
	}{
		{
			name: "valid tokens",
			tokens: []*design.Token{
				{
					Name:     "primary-color",
					Value:    "#007bff",
					Type:     "color",
					Category: "colors",
				},
				{
					Name:     "spacing-base",
					Value:    "16px",
					Type:     "spacing",
					Category: "spacing",
				},
			},
			expectValid: true,
			expectError: 0,
		},
		{
			name: "invalid color format",
			tokens: []*design.Token{
				{
					Name:     "primary-color",
					Value:    "not-a-color",
					Type:     "color",
					Category: "colors",
				},
			},
			expectValid: false,
			expectError: 1,
		},
		{
			name: "missing required fields",
			tokens: []*design.Token{
				{
					Name:  "test-token",
					Value: "",
					Type:  "",
				},
			},
			expectValid: false,
			expectError: 2,
		},
		{
			name: "duplicate token names",
			tokens: []*design.Token{
				{
					Name:     "primary-color",
					Value:    "#007bff",
					Type:     "color",
					Category: "colors",
				},
				{
					Name:     "primary-color",
					Value:    "#0056b3",
					Type:     "color",
					Category: "colors",
				},
			},
			expectValid: false,
			expectError: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := designClient.ValidateTokens(tt.tokens)

			if result.Valid != tt.expectValid {
				t.Errorf("expected valid=%v, got %v", tt.expectValid, result.Valid)
			}

			if len(result.Errors) != tt.expectError {
				t.Errorf("expected %d errors, got %d", tt.expectError, len(result.Errors))
			}
		})
	}
}

// TestClientWithoutProjectID tests that operations fail without project ID.
func TestClientWithoutProjectID(t *testing.T) {
	apiClient := client.New(client.WithBaseURL("http://localhost"))
	designClient := New(WithAPIClient(apiClient))

	ctx := context.Background()

	// Test UploadTokens
	tokens := []*design.Token{
		{Name: "test", Value: "#000", Type: "color", Category: "colors"},
	}
	_, err := designClient.UploadTokens(ctx, tokens, design.ConflictOverwrite, nil)
	if err == nil {
		t.Error("UploadTokens should fail without project ID")
	}

	// Test GetTokens
	_, _, err = designClient.GetTokens(ctx, nil, "", 100, 0)
	if err == nil {
		t.Error("GetTokens should fail without project ID")
	}

	// Test DeleteToken
	err = designClient.DeleteToken(ctx, "test")
	if err == nil {
		t.Error("DeleteToken should fail without project ID")
	}
}

// TestLargeBatchUpload tests uploading a large number of tokens.
func TestLargeBatchUpload(t *testing.T) {
	// Create mock server that counts batches
	batchCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		batchCount++

		var req UploadTokensRequest
		json.NewDecoder(r.Body).Decode(&req)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(UploadTokensResponse{
			Success:       true,
			UploadedCount: len(req.Tokens),
			SkippedCount:  0,
			UpdatedCount:  0,
		})
	}))
	defer server.Close()

	// Create client
	apiClient := client.New(
		client.WithBaseURL(server.URL),
		client.WithTimeout(5*time.Second),
	)

	designClient := New(
		WithAPIClient(apiClient),
		WithProjectID("test-project"),
	)

	// Create 250 tokens (should be split into 3 batches of 100, 100, 50)
	tokens := make([]*design.Token, 250)
	for i := 0; i < 250; i++ {
		tokens[i] = &design.Token{
			Name:     "color-" + string(rune(i)),
			Value:    "#000000",
			Type:     "color",
			Category: "colors",
		}
	}

	// Upload tokens
	ctx := context.Background()
	result, err := designClient.UploadTokens(ctx, tokens, design.ConflictOverwrite, nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if result.UploadedCount != 250 {
		t.Errorf("expected 250 uploaded, got %d", result.UploadedCount)
	}

	if batchCount != 3 {
		t.Errorf("expected 3 batches, got %d", batchCount)
	}
}
