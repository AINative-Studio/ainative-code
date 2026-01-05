package strapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AINative-studio/ainative-code/internal/client"
)

// TestCreateBlogPost tests creating a blog post.
func TestCreateBlogPost(t *testing.T) {
	tests := []struct {
		name           string
		request        *CreateBlogPostRequest
		mockResponse   *CreateBlogPostResponse
		mockStatusCode int
		wantErr        bool
		errContains    string
	}{
		{
			name: "successful creation",
			request: &CreateBlogPostRequest{
				Data: &BlogPostData{
					Title:   "Test Post",
					Content: "# Test Content",
					Author:  "John Doe",
					Status:  "draft",
				},
			},
			mockResponse: &CreateBlogPostResponse{
				Data: &BlogPost{
					ID: 1,
					Attributes: &BlogPostAttributes{
						Title:     "Test Post",
						Content:   "# Test Content",
						Author:    "John Doe",
						Status:    "draft",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
				},
			},
			mockStatusCode: http.StatusOK,
			wantErr:        false,
		},
		{
			name: "missing title",
			request: &CreateBlogPostRequest{
				Data: &BlogPostData{
					Content: "# Test Content",
					Author:  "John Doe",
				},
			},
			wantErr:     true,
			errContains: "title is required",
		},
		{
			name: "missing content",
			request: &CreateBlogPostRequest{
				Data: &BlogPostData{
					Title:  "Test Post",
					Author: "John Doe",
				},
			},
			wantErr:     true,
			errContains: "content is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/blog-posts" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				if r.Method != http.MethodPost {
					t.Errorf("unexpected method: %s", r.Method)
				}

				w.WriteHeader(tt.mockStatusCode)
				if tt.mockResponse != nil {
					json.NewEncoder(w).Encode(tt.mockResponse)
				}
			}))
			defer server.Close()

			// Create client
			apiClient := client.New(
				client.WithBaseURL(server.URL),
			)
			strapiClient := New(
				WithAPIClient(apiClient),
				WithBaseURL(server.URL),
			)

			// Execute test
			ctx := context.Background()
			post, err := strapiClient.CreateBlogPost(ctx, tt.request)

			// Check error
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				} else if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("expected error to contain %q, got %q", tt.errContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Validate response
			if post.ID != tt.mockResponse.Data.ID {
				t.Errorf("expected ID %d, got %d", tt.mockResponse.Data.ID, post.ID)
			}
			if post.Attributes.Title != tt.mockResponse.Data.Attributes.Title {
				t.Errorf("expected title %q, got %q", tt.mockResponse.Data.Attributes.Title, post.Attributes.Title)
			}
		})
	}
}

// TestListBlogPosts tests listing blog posts.
func TestListBlogPosts(t *testing.T) {
	tests := []struct {
		name           string
		options        *ListOptions
		mockResponse   *ListBlogPostsResponse
		mockStatusCode int
		wantErr        bool
	}{
		{
			name: "successful list",
			options: &ListOptions{
				Page:     1,
				PageSize: 10,
			},
			mockResponse: &ListBlogPostsResponse{
				Data: []*BlogPost{
					{
						ID: 1,
						Attributes: &BlogPostAttributes{
							Title:     "Post 1",
							Content:   "Content 1",
							Status:    "published",
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						},
					},
					{
						ID: 2,
						Attributes: &BlogPostAttributes{
							Title:     "Post 2",
							Content:   "Content 2",
							Status:    "draft",
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						},
					},
				},
				Meta: &ListMeta{
					Pagination: &Pagination{
						Page:      1,
						PageSize:  10,
						PageCount: 1,
						Total:     2,
					},
				},
			},
			mockStatusCode: http.StatusOK,
			wantErr:        false,
		},
		{
			name: "with filters",
			options: &ListOptions{
				Page:     1,
				PageSize: 10,
				Filters: map[string]interface{}{
					"status": "published",
					"author": "John Doe",
				},
			},
			mockResponse: &ListBlogPostsResponse{
				Data: []*BlogPost{
					{
						ID: 1,
						Attributes: &BlogPostAttributes{
							Title:     "Published Post",
							Content:   "Content",
							Author:    "John Doe",
							Status:    "published",
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						},
					},
				},
				Meta: &ListMeta{
					Pagination: &Pagination{
						Page:      1,
						PageSize:  10,
						PageCount: 1,
						Total:     1,
					},
				},
			},
			mockStatusCode: http.StatusOK,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("unexpected method: %s", r.Method)
				}

				w.WriteHeader(tt.mockStatusCode)
				if tt.mockResponse != nil {
					json.NewEncoder(w).Encode(tt.mockResponse)
				}
			}))
			defer server.Close()

			// Create client
			apiClient := client.New(
				client.WithBaseURL(server.URL),
			)
			strapiClient := New(
				WithAPIClient(apiClient),
				WithBaseURL(server.URL),
			)

			// Execute test
			ctx := context.Background()
			posts, meta, err := strapiClient.ListBlogPosts(ctx, tt.options)

			// Check error
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Validate response
			if len(posts) != len(tt.mockResponse.Data) {
				t.Errorf("expected %d posts, got %d", len(tt.mockResponse.Data), len(posts))
			}
			if meta.Pagination.Total != tt.mockResponse.Meta.Pagination.Total {
				t.Errorf("expected total %d, got %d", tt.mockResponse.Meta.Pagination.Total, meta.Pagination.Total)
			}
		})
	}
}

// TestUpdateBlogPost tests updating a blog post.
func TestUpdateBlogPost(t *testing.T) {
	tests := []struct {
		name           string
		postID         int
		request        *UpdateBlogPostRequest
		mockResponse   *UpdateBlogPostResponse
		mockStatusCode int
		wantErr        bool
	}{
		{
			name:   "successful update",
			postID: 1,
			request: &UpdateBlogPostRequest{
				Data: &BlogPostData{
					Title:   "Updated Title",
					Content: "Updated Content",
				},
			},
			mockResponse: &UpdateBlogPostResponse{
				Data: &BlogPost{
					ID: 1,
					Attributes: &BlogPostAttributes{
						Title:     "Updated Title",
						Content:   "Updated Content",
						Status:    "draft",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
				},
			},
			mockStatusCode: http.StatusOK,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPut {
					t.Errorf("unexpected method: %s", r.Method)
				}

				w.WriteHeader(tt.mockStatusCode)
				if tt.mockResponse != nil {
					json.NewEncoder(w).Encode(tt.mockResponse)
				}
			}))
			defer server.Close()

			// Create client
			apiClient := client.New(
				client.WithBaseURL(server.URL),
			)
			strapiClient := New(
				WithAPIClient(apiClient),
				WithBaseURL(server.URL),
			)

			// Execute test
			ctx := context.Background()
			post, err := strapiClient.UpdateBlogPost(ctx, tt.postID, tt.request)

			// Check error
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Validate response
			if post.Attributes.Title != tt.mockResponse.Data.Attributes.Title {
				t.Errorf("expected title %q, got %q", tt.mockResponse.Data.Attributes.Title, post.Attributes.Title)
			}
		})
	}
}

// TestPublishBlogPost tests publishing a blog post.
func TestPublishBlogPost(t *testing.T) {
	tests := []struct {
		name           string
		postID         int
		mockResponse   *UpdateBlogPostResponse
		mockStatusCode int
		wantErr        bool
	}{
		{
			name:   "successful publish",
			postID: 1,
			mockResponse: &UpdateBlogPostResponse{
				Data: &BlogPost{
					ID: 1,
					Attributes: &BlogPostAttributes{
						Title:       "Test Post",
						Content:     "Test Content",
						Status:      "published",
						PublishedAt: timePtr(time.Now()),
						CreatedAt:   time.Now(),
						UpdatedAt:   time.Now(),
					},
				},
			},
			mockStatusCode: http.StatusOK,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPut {
					t.Errorf("unexpected method: %s", r.Method)
				}

				w.WriteHeader(tt.mockStatusCode)
				if tt.mockResponse != nil {
					json.NewEncoder(w).Encode(tt.mockResponse)
				}
			}))
			defer server.Close()

			// Create client
			apiClient := client.New(
				client.WithBaseURL(server.URL),
			)
			strapiClient := New(
				WithAPIClient(apiClient),
				WithBaseURL(server.URL),
			)

			// Execute test
			ctx := context.Background()
			post, err := strapiClient.PublishBlogPost(ctx, tt.postID)

			// Check error
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Validate response
			if post.Attributes.Status != "published" {
				t.Errorf("expected status %q, got %q", "published", post.Attributes.Status)
			}
		})
	}
}

// TestDeleteBlogPost tests deleting a blog post.
func TestDeleteBlogPost(t *testing.T) {
	tests := []struct {
		name           string
		postID         int
		mockResponse   *DeleteBlogPostResponse
		mockStatusCode int
		wantErr        bool
	}{
		{
			name:   "successful delete",
			postID: 1,
			mockResponse: &DeleteBlogPostResponse{
				Data: &BlogPost{
					ID: 1,
					Attributes: &BlogPostAttributes{
						Title:     "Deleted Post",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
				},
			},
			mockStatusCode: http.StatusOK,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodDelete {
					t.Errorf("unexpected method: %s", r.Method)
				}

				w.WriteHeader(tt.mockStatusCode)
				if tt.mockResponse != nil {
					json.NewEncoder(w).Encode(tt.mockResponse)
				}
			}))
			defer server.Close()

			// Create client
			apiClient := client.New(
				client.WithBaseURL(server.URL),
			)
			strapiClient := New(
				WithAPIClient(apiClient),
				WithBaseURL(server.URL),
			)

			// Execute test
			ctx := context.Background()
			err := strapiClient.DeleteBlogPost(ctx, tt.postID)

			// Check error
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
		})
	}
}

// TestBuildQueryParams tests query parameter building.
func TestBuildQueryParams(t *testing.T) {
	tests := []struct {
		name     string
		options  *ListOptions
		expected []string // expected params to be present
	}{
		{
			name: "pagination only",
			options: &ListOptions{
				Page:     2,
				PageSize: 50,
			},
			expected: []string{
				"pagination%5Bpage%5D=2",
				"pagination%5BpageSize%5D=50",
			},
		},
		{
			name: "with filters",
			options: &ListOptions{
				Page:     1,
				PageSize: 25,
				Filters: map[string]interface{}{
					"status": "published",
				},
			},
			expected: []string{
				"pagination%5Bpage%5D=1",
				"pagination%5BpageSize%5D=25",
				"filters%5Bstatus%5D",
			},
		},
		{
			name: "with sort",
			options: &ListOptions{
				Page:     1,
				PageSize: 25,
				Sort:     []string{"createdAt:desc", "title:asc"},
			},
			expected: []string{
				"pagination%5Bpage%5D=1",
				"pagination%5BpageSize%5D=25",
				"sort",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{}
			params := client.buildQueryParams(tt.options)

			for _, key := range tt.expected {
				if !containsSubstring(params, key) {
					t.Errorf("expected query param %q to be present in %q", key, params)
				}
			}
		})
	}
}

// Helper functions

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) >= len(substr) && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func timePtr(t time.Time) *time.Time {
	return &t
}
