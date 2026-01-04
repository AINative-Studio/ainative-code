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

// TestStrapiCMSOperations_CreatePost tests blog post creation
func TestStrapiCMSOperations_CreatePost(t *testing.T) {
	t.Run("should create new blog post successfully", func(t *testing.T) {
		// Given: Mock Strapi server
		server := mocks.NewStrapiServer()
		defer server.Close()

		// And: Post data
		postData := map[string]interface{}{
			"data": map[string]interface{}{
				"title":   "Test Blog Post",
				"content": "This is a test post",
				"author":  "Test Author",
				"tags":    []string{"test", "blog"},
			},
		}

		// When: Creating post
		body, _ := json.Marshal(postData)
		req, _ := http.NewRequest("POST", server.GetURL()+"/api/posts", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-token")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)

		// Then: Should succeed
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		assert.True(t, server.CreateCalled)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Contains(t, result, "data")
	})

	t.Run("should reject post with invalid data", func(t *testing.T) {
		// Given: Mock Strapi server
		server := mocks.NewStrapiServer()
		defer server.Close()

		// When: Creating post with invalid JSON
		req, _ := http.NewRequest("POST", server.GetURL()+"/api/posts", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-token")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)

		// Then: Should reject
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

// TestStrapiCMSOperations_ListPosts tests listing blog posts
func TestStrapiCMSOperations_ListPosts(t *testing.T) {
	t.Run("should list all blog posts", func(t *testing.T) {
		// Given: Mock Strapi server with posts
		server := mocks.NewStrapiServer()
		defer server.Close()

		server.AddPost(map[string]interface{}{
			"title":   "Post 1",
			"content": "Content 1",
		})
		server.AddPost(map[string]interface{}{
			"title":   "Post 2",
			"content": "Content 2",
		})

		// When: Listing posts
		req, _ := http.NewRequest("GET", server.GetURL()+"/api/posts", nil)
		req.Header.Set("Authorization", "Bearer test-token")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)

		// Then: Should return posts
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, server.ListCalled)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Contains(t, result, "data")
		assert.Contains(t, result, "meta")

		data := result["data"].([]interface{})
		assert.Equal(t, 2, len(data))
	})

	t.Run("should return empty list when no posts", func(t *testing.T) {
		// Given: Mock Strapi server with no posts
		server := mocks.NewStrapiServer()
		defer server.Close()

		// When: Listing posts
		req, _ := http.NewRequest("GET", server.GetURL()+"/api/posts", nil)
		req.Header.Set("Authorization", "Bearer test-token")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)

		// Then: Should return empty list
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		data := result["data"].([]interface{})
		assert.Equal(t, 0, len(data))
	})
}

// TestStrapiCMSOperations_UpdatePost tests updating blog posts
func TestStrapiCMSOperations_UpdatePost(t *testing.T) {
	t.Run("should update existing post", func(t *testing.T) {
		// Given: Mock Strapi server with post
		server := mocks.NewStrapiServer()
		defer server.Close()

		postID := server.AddPost(map[string]interface{}{
			"title":   "Original Title",
			"content": "Original Content",
		})

		// And: Update data
		updateData := map[string]interface{}{
			"data": map[string]interface{}{
				"title":   "Updated Title",
				"content": "Updated Content",
			},
		}

		// When: Updating post
		body, _ := json.Marshal(updateData)
		req, _ := http.NewRequest("PUT", server.GetURL()+"/api/posts/"+postID, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-token")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)

		// Then: Should succeed
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, server.UpdateCalled)
	})

	t.Run("should return 404 for non-existent post", func(t *testing.T) {
		// Given: Mock Strapi server
		server := mocks.NewStrapiServer()
		defer server.Close()

		// And: Update data
		updateData := map[string]interface{}{
			"data": map[string]interface{}{
				"title": "Updated Title",
			},
		}

		// When: Updating non-existent post
		body, _ := json.Marshal(updateData)
		req, _ := http.NewRequest("PUT", server.GetURL()+"/api/posts/999", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-token")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)

		// Then: Should return 404
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

// TestStrapiCMSOperations_PublishUnpublish tests publish/unpublish workflow
func TestStrapiCMSOperations_PublishUnpublish(t *testing.T) {
	t.Run("should publish a post", func(t *testing.T) {
		// Given: Mock Strapi server with unpublished post
		server := mocks.NewStrapiServer()
		defer server.Close()

		postID := server.AddPost(map[string]interface{}{
			"title":   "Draft Post",
			"content": "Draft Content",
		})

		// And: Publish request
		publishData := map[string]interface{}{
			"data": map[string]interface{}{
				"publishedAt": time.Now().Format(time.RFC3339),
			},
		}

		// When: Publishing post
		body, _ := json.Marshal(publishData)
		req, _ := http.NewRequest("PUT", server.GetURL()+"/api/posts/"+postID, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-token")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)

		// Then: Should succeed
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, server.PublishCalled)
	})

	t.Run("should unpublish a post", func(t *testing.T) {
		// Given: Mock Strapi server with published post
		server := mocks.NewStrapiServer()
		defer server.Close()

		postID := server.AddPost(map[string]interface{}{
			"title":   "Published Post",
			"content": "Published Content",
		})

		// And: Unpublish request
		unpublishData := map[string]interface{}{
			"data": map[string]interface{}{
				"publishedAt": nil,
			},
		}

		// When: Unpublishing post
		body, _ := json.Marshal(unpublishData)
		req, _ := http.NewRequest("PUT", server.GetURL()+"/api/posts/"+postID, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-token")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)

		// Then: Should succeed
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

// TestStrapiCMSOperations_DeletePost tests deleting blog posts
func TestStrapiCMSOperations_DeletePost(t *testing.T) {
	t.Run("should delete existing post", func(t *testing.T) {
		// Given: Mock Strapi server with post
		server := mocks.NewStrapiServer()
		defer server.Close()

		postID := server.AddPost(map[string]interface{}{
			"title":   "Post to Delete",
			"content": "Content",
		})

		// When: Deleting post
		req, _ := http.NewRequest("DELETE", server.GetURL()+"/api/posts/"+postID, nil)
		req.Header.Set("Authorization", "Bearer test-token")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)

		// Then: Should succeed
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, server.DeleteCalled)
	})

	t.Run("should return 404 when deleting non-existent post", func(t *testing.T) {
		// Given: Mock Strapi server
		server := mocks.NewStrapiServer()
		defer server.Close()

		// When: Deleting non-existent post
		req, _ := http.NewRequest("DELETE", server.GetURL()+"/api/posts/999", nil)
		req.Header.Set("Authorization", "Bearer test-token")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)

		// Then: Should return 404
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

// TestStrapiCMSErrorHandling tests error scenarios
func TestStrapiCMSErrorHandling(t *testing.T) {
	t.Run("should handle 401 unauthorized", func(t *testing.T) {
		// Given: Mock server with auth failure
		server := mocks.NewStrapiServer()
		defer server.Close()

		server.ShouldFailAuth = true

		// When: Making request without auth
		req, _ := http.NewRequest("GET", server.GetURL()+"/api/posts", nil)

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)

		// Then: Should return 401
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("should handle 429 rate limit", func(t *testing.T) {
		// Given: Mock server with rate limiting
		server := mocks.NewStrapiServer()
		defer server.Close()

		server.ShouldRateLimit = true

		// When: Making request
		req, _ := http.NewRequest("GET", server.GetURL()+"/api/posts", nil)
		req.Header.Set("Authorization", "Bearer test-token")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)

		// Then: Should return 429
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
		assert.NotEmpty(t, resp.Header.Get("Retry-After"))
	})

	t.Run("should require authorization header", func(t *testing.T) {
		// Given: Mock Strapi server
		server := mocks.NewStrapiServer()
		defer server.Close()

		// When: Making request without auth header
		req, _ := http.NewRequest("GET", server.GetURL()+"/api/posts", nil)

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)

		// Then: Should return 401
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}
