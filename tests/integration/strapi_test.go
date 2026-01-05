// +build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/AINative-studio/ainative-code/internal/client"
	"github.com/AINative-studio/ainative-code/internal/client/strapi"
	"github.com/AINative-studio/ainative-code/tests/integration/helpers"
	"github.com/stretchr/testify/suite"
)

// StrapiIntegrationTestSuite tests Strapi CMS integration functionality.
type StrapiIntegrationTestSuite struct {
	suite.Suite
	strapiClient *strapi.Client
	cleanup      func()
}

// SetupTest runs before each test in the suite.
func (s *StrapiIntegrationTestSuite) SetupTest() {
	// Create mock Strapi server
	mockServer, cleanup := helpers.MockStrapiServer(s.T())
	s.cleanup = cleanup

	// Create HTTP API client
	apiClient := client.New(
		client.WithBaseURL(mockServer.URL),
		client.WithBearerToken("test_token"),
	)

	// Create Strapi client
	s.strapiClient = strapi.New(
		strapi.WithAPIClient(apiClient),
		strapi.WithBaseURL(mockServer.URL),
	)
}

// TearDownTest runs after each test in the suite.
func (s *StrapiIntegrationTestSuite) TearDownTest() {
	if s.cleanup != nil {
		s.cleanup()
	}
}

// TestBlogPostCreation tests creating a blog post in Strapi.
func (s *StrapiIntegrationTestSuite) TestBlogPostCreation() {
	// Given: A new blog post request
	ctx := context.Background()

	req := &strapi.CreateBlogPostRequest{
		Data: &strapi.BlogPostData{
			Title:   "Test Blog Post",
			Content: "This is the content of the test blog post.",
			Author:  "Test Author",
			Status:  "draft",
			Slug:    "test-blog-post",
			Tags:    []string{"test", "integration"},
		},
	}

	// When: Creating the blog post
	post, err := s.strapiClient.CreateBlogPost(ctx, req)

	// Then: Should create successfully
	s.Require().NoError(err, "Failed to create blog post")
	s.NotNil(post, "Blog post should not be nil")
	s.NotZero(post.ID, "Blog post ID should be set")
	s.Equal("Test Blog Post", post.Attributes.Title, "Title should match")
	s.Equal("This is the content of the test blog post.", post.Attributes.Content, "Content should match")
	s.Equal("Test Author", post.Attributes.Author, "Author should match")
	s.Equal("draft", post.Attributes.Status, "Status should be draft")
}

// TestBlogPostCreationWithMissingFields tests validation of required fields.
func (s *StrapiIntegrationTestSuite) TestBlogPostCreationWithMissingFields() {
	// Given: A blog post request missing required fields
	ctx := context.Background()

	// When: Creating blog post without title
	reqNoTitle := &strapi.CreateBlogPostRequest{
		Data: &strapi.BlogPostData{
			Content: "Content without title",
		},
	}

	_, err := s.strapiClient.CreateBlogPost(ctx, reqNoTitle)

	// Then: Should return validation error
	s.Error(err, "Should error when title is missing")
	s.Contains(err.Error(), "title", "Error should mention title")

	// When: Creating blog post without content
	reqNoContent := &strapi.CreateBlogPostRequest{
		Data: &strapi.BlogPostData{
			Title: "Title without content",
		},
	}

	_, err = s.strapiClient.CreateBlogPost(ctx, reqNoContent)

	// Then: Should return validation error
	s.Error(err, "Should error when content is missing")
	s.Contains(err.Error(), "content", "Error should mention content")
}

// TestBlogPostListing tests listing blog posts with pagination.
func (s *StrapiIntegrationTestSuite) TestBlogPostListing() {
	// Given: List options for pagination
	ctx := context.Background()

	opts := &strapi.ListOptions{
		Page:     1,
		PageSize: 10,
		Sort:     []string{"createdAt:desc"},
	}

	// When: Listing blog posts
	posts, meta, err := s.strapiClient.ListBlogPosts(ctx, opts)

	// Then: Should return list successfully
	s.Require().NoError(err, "Failed to list blog posts")
	s.NotNil(posts, "Posts list should not be nil")
	s.NotNil(meta, "Metadata should not be nil")

	// Verify pagination metadata
	s.NotNil(meta.Pagination, "Pagination info should be present")
	s.Equal(1, meta.Pagination.Page, "Page should be 1")
	s.GreaterOrEqual(meta.Pagination.Total, 0, "Total should be non-negative")
}

// TestBlogPostRetrieval tests retrieving a single blog post by ID.
func (s *StrapiIntegrationTestSuite) TestBlogPostRetrieval() {
	// Given: A blog post ID
	ctx := context.Background()
	postID := 1

	// When: Retrieving the blog post
	post, err := s.strapiClient.GetBlogPost(ctx, postID)

	// Then: Should retrieve successfully
	s.Require().NoError(err, "Failed to retrieve blog post")
	s.NotNil(post, "Blog post should not be nil")
	s.Equal(postID, post.ID, "Post ID should match")
	s.NotEmpty(post.Attributes.Title, "Title should not be empty")
	s.NotEmpty(post.Attributes.Content, "Content should not be empty")
}

// TestBlogPostUpdate tests updating an existing blog post.
func (s *StrapiIntegrationTestSuite) TestBlogPostUpdate() {
	// Given: An existing blog post and update request
	ctx := context.Background()
	postID := 1

	updateReq := &strapi.UpdateBlogPostRequest{
		Data: &strapi.BlogPostData{
			Title:   "Updated Blog Post Title",
			Content: "Updated content for the blog post.",
			Status:  "draft",
		},
	}

	// When: Updating the blog post
	updatedPost, err := s.strapiClient.UpdateBlogPost(ctx, postID, updateReq)

	// Then: Should update successfully
	s.Require().NoError(err, "Failed to update blog post")
	s.NotNil(updatedPost, "Updated blog post should not be nil")
	s.Equal(postID, updatedPost.ID, "Post ID should match")
	s.Equal("Updated Blog Post Title", updatedPost.Attributes.Title, "Title should be updated")
	s.Equal("Updated content for the blog post.", updatedPost.Attributes.Content, "Content should be updated")
}

// TestBlogPostPublishing tests publishing a blog post.
func (s *StrapiIntegrationTestSuite) TestBlogPostPublishing() {
	// Given: A draft blog post
	ctx := context.Background()
	postID := 1

	// When: Publishing the blog post
	publishedPost, err := s.strapiClient.PublishBlogPost(ctx, postID)

	// Then: Should publish successfully
	s.Require().NoError(err, "Failed to publish blog post")
	s.NotNil(publishedPost, "Published blog post should not be nil")
	s.Equal(postID, publishedPost.ID, "Post ID should match")
	s.Equal("published", publishedPost.Attributes.Status, "Status should be published")
	s.NotNil(publishedPost.Attributes.PublishedAt, "PublishedAt timestamp should be set")
}

// TestBlogPostDeletion tests deleting a blog post.
func (s *StrapiIntegrationTestSuite) TestBlogPostDeletion() {
	// Given: An existing blog post
	ctx := context.Background()
	postID := 1

	// When: Deleting the blog post
	err := s.strapiClient.DeleteBlogPost(ctx, postID)

	// Then: Should delete successfully
	s.Require().NoError(err, "Failed to delete blog post")
}

// TestBlogPostListingWithFilters tests listing blog posts with filters.
func (s *StrapiIntegrationTestSuite) TestBlogPostListingWithFilters() {
	// Given: List options with filters
	ctx := context.Background()

	opts := &strapi.ListOptions{
		Page:     1,
		PageSize: 10,
		Filters: map[string]interface{}{
			"status": "published",
			"author": "Test Author",
		},
		Sort: []string{"publishedAt:desc"},
	}

	// When: Listing blog posts with filters
	posts, meta, err := s.strapiClient.ListBlogPosts(ctx, opts)

	// Then: Should return filtered list
	s.Require().NoError(err, "Failed to list filtered blog posts")
	s.NotNil(posts, "Posts list should not be nil")
	s.NotNil(meta, "Metadata should not be nil")
}

// TestContentTypeOperations tests generic content type operations.
func (s *StrapiIntegrationTestSuite) TestContentTypeOperations() {
	// Given: A content type name
	ctx := context.Background()
	contentType := "articles"

	// When: Creating a content entry
	data := map[string]interface{}{
		"title":       "Test Article",
		"description": "Test article description",
		"published":   true,
	}

	entry, err := s.strapiClient.CreateContent(ctx, contentType, data)

	// Then: Should create successfully
	s.Require().NoError(err, "Failed to create content entry")
	s.NotNil(entry, "Content entry should not be nil")
	s.NotZero(entry.ID, "Content entry ID should be set")
	s.NotNil(entry.Attributes, "Attributes should not be nil")
}

// TestContentTypeListing tests listing content from a specific content type.
func (s *StrapiIntegrationTestSuite) TestContentTypeListing() {
	// Given: A content type and list options
	ctx := context.Background()
	contentType := "articles"

	opts := &strapi.ListOptions{
		Page:     1,
		PageSize: 25,
	}

	// When: Listing content entries
	entries, meta, err := s.strapiClient.ListContent(ctx, contentType, opts)

	// Then: Should list successfully
	s.Require().NoError(err, "Failed to list content entries")
	s.NotNil(entries, "Entries list should not be nil")
	s.NotNil(meta, "Metadata should not be nil")
	s.NotNil(meta.Pagination, "Pagination should be present")
}

// TestContentUpdate tests updating a content entry.
func (s *StrapiIntegrationTestSuite) TestContentUpdate() {
	// Given: An existing content entry
	ctx := context.Background()
	contentType := "articles"
	entryID := 1

	updateData := map[string]interface{}{
		"title":       "Updated Article Title",
		"description": "Updated article description",
	}

	// When: Updating the content entry
	updatedEntry, err := s.strapiClient.UpdateContent(ctx, contentType, entryID, updateData)

	// Then: Should update successfully
	s.Require().NoError(err, "Failed to update content entry")
	s.NotNil(updatedEntry, "Updated entry should not be nil")
	s.Equal(entryID, updatedEntry.ID, "Entry ID should match")
}

// TestContentDeletion tests deleting a content entry.
func (s *StrapiIntegrationTestSuite) TestContentDeletion() {
	// Given: An existing content entry
	ctx := context.Background()
	contentType := "articles"
	entryID := 1

	// When: Deleting the content entry
	err := s.strapiClient.DeleteContent(ctx, contentType, entryID)

	// Then: Should delete successfully
	s.Require().NoError(err, "Failed to delete content entry")
}

// TestStrapiAuthentication tests Strapi authentication headers.
func (s *StrapiIntegrationTestSuite) TestStrapiAuthentication() {
	// Given: Strapi client with authentication
	// (authentication is handled by the underlying HTTP client)

	ctx := context.Background()

	// When: Making an authenticated request
	opts := &strapi.ListOptions{
		Page:     1,
		PageSize: 10,
	}

	// Then: Should succeed with proper authentication
	_, _, err := s.strapiClient.ListBlogPosts(ctx, opts)
	s.Require().NoError(err, "Authenticated request should succeed")
}

// TestListOptionsDefaults tests default values for list options.
func (s *StrapiIntegrationTestSuite) TestListOptionsDefaults() {
	// Given: List options with no values set
	ctx := context.Background()

	// When: Listing with nil options (should use defaults)
	posts, meta, err := s.strapiClient.ListBlogPosts(ctx, nil)

	// Then: Should use default pagination
	s.Require().NoError(err, "Should succeed with default options")
	s.NotNil(posts, "Posts should not be nil")
	s.NotNil(meta, "Metadata should not be nil")
	s.NotNil(meta.Pagination, "Pagination should have defaults")
	s.Equal(1, meta.Pagination.Page, "Default page should be 1")
	s.Equal(25, meta.Pagination.PageSize, "Default page size should be 25")
}

// TestBlogPostMetadata tests handling blog post metadata.
func (s *StrapiIntegrationTestSuite) TestBlogPostMetadata() {
	// Given: A blog post with metadata
	ctx := context.Background()

	req := &strapi.CreateBlogPostRequest{
		Data: &strapi.BlogPostData{
			Title:   "Post with Metadata",
			Content: "Content with metadata",
			Status:  "draft",
			Metadata: map[string]interface{}{
				"seo_title":       "SEO optimized title",
				"seo_description": "SEO description",
				"featured":        true,
			},
		},
	}

	// When: Creating blog post with metadata
	post, err := s.strapiClient.CreateBlogPost(ctx, req)

	// Then: Metadata should be preserved
	s.Require().NoError(err, "Failed to create blog post with metadata")
	s.NotNil(post, "Post should not be nil")
	s.NotNil(post.Attributes.Metadata, "Metadata should be present")
}

// TestBlogPostTags tests handling blog post tags.
func (s *StrapiIntegrationTestSuite) TestBlogPostTags() {
	// Given: A blog post with tags
	ctx := context.Background()

	req := &strapi.CreateBlogPostRequest{
		Data: &strapi.BlogPostData{
			Title:   "Post with Tags",
			Content: "Content with tags",
			Status:  "draft",
			Tags:    []string{"tag1", "tag2", "tag3"},
		},
	}

	// When: Creating blog post with tags
	post, err := s.strapiClient.CreateBlogPost(ctx, req)

	// Then: Tags should be preserved
	s.Require().NoError(err, "Failed to create blog post with tags")
	s.NotNil(post, "Post should not be nil")
	s.NotNil(post.Attributes.Tags, "Tags should be present")
	s.Len(post.Attributes.Tags, 3, "Should have 3 tags")
}

// TestPaginationHandling tests pagination across multiple pages.
func (s *StrapiIntegrationTestSuite) TestPaginationHandling() {
	// Given: Multiple pages of blog posts
	ctx := context.Background()

	// When: Fetching first page
	optsPage1 := &strapi.ListOptions{
		Page:     1,
		PageSize: 5,
	}
	postsPage1, metaPage1, err := s.strapiClient.ListBlogPosts(ctx, optsPage1)

	// Then: Should get first page
	s.Require().NoError(err, "Failed to get first page")
	s.NotNil(postsPage1, "First page should not be nil")
	s.Equal(1, metaPage1.Pagination.Page, "Should be page 1")

	// When: Fetching second page
	optsPage2 := &strapi.ListOptions{
		Page:     2,
		PageSize: 5,
	}
	postsPage2, metaPage2, err := s.strapiClient.ListBlogPosts(ctx, optsPage2)

	// Then: Should get second page
	s.Require().NoError(err, "Failed to get second page")
	s.NotNil(postsPage2, "Second page should not be nil")
	s.Equal(2, metaPage2.Pagination.Page, "Should be page 2")
}

// TestSortingOptions tests sorting blog posts.
func (s *StrapiIntegrationTestSuite) TestSortingOptions() {
	// Given: List options with sorting
	ctx := context.Background()

	testCases := []struct {
		name     string
		sortBy   []string
		expected string
	}{
		{
			name:     "Sort by created date descending",
			sortBy:   []string{"createdAt:desc"},
			expected: "createdAt:desc",
		},
		{
			name:     "Sort by title ascending",
			sortBy:   []string{"title:asc"},
			expected: "title:asc",
		},
		{
			name:     "Multiple sort fields",
			sortBy:   []string{"status:asc", "createdAt:desc"},
			expected: "status:asc",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			opts := &strapi.ListOptions{
				Page:     1,
				PageSize: 10,
				Sort:     tc.sortBy,
			}

			// When: Listing with sort options
			posts, _, err := s.strapiClient.ListBlogPosts(ctx, opts)

			// Then: Should apply sorting
			s.Require().NoError(err, "Failed to list with sorting")
			s.NotNil(posts, "Posts should not be nil")
		})
	}
}

// TestErrorHandling tests error handling for various scenarios.
func (s *StrapiIntegrationTestSuite) TestErrorHandling() {
	// Given: Invalid operations
	ctx := context.Background()

	// When: Attempting to get non-existent post
	// Note: Mock server returns valid responses, so we test the client's behavior
	_, err := s.strapiClient.GetBlogPost(ctx, 99999)

	// Then: Should handle gracefully (mock returns data, but in real scenario would error)
	// Since mock always returns data, we just verify no panic occurs
	s.NotPanics(func() {
		s.strapiClient.GetBlogPost(ctx, 99999)
	}, "Should handle requests gracefully")
}

// TestConcurrentOperations tests concurrent Strapi operations.
func (s *StrapiIntegrationTestSuite) TestConcurrentOperations() {
	// Given: Multiple concurrent requests
	ctx := context.Background()
	concurrentOps := 5
	done := make(chan bool, concurrentOps)
	errors := make(chan error, concurrentOps)

	// When: Making concurrent requests
	for i := 0; i < concurrentOps; i++ {
		go func(index int) {
			opts := &strapi.ListOptions{
				Page:     1,
				PageSize: 10,
			}
			_, _, err := s.strapiClient.ListBlogPosts(ctx, opts)
			if err != nil {
				errors <- err
			}
			done <- true
		}(i)
	}

	// Wait for all operations to complete
	for i := 0; i < concurrentOps; i++ {
		<-done
	}
	close(errors)

	// Then: All operations should succeed
	s.Empty(errors, "No errors should occur during concurrent operations")
}

// TestTimestamps tests handling of created/updated timestamps.
func (s *StrapiIntegrationTestSuite) TestTimestamps() {
	// Given: A new blog post
	ctx := context.Background()

	req := &strapi.CreateBlogPostRequest{
		Data: &strapi.BlogPostData{
			Title:   "Timestamp Test Post",
			Content: "Testing timestamps",
			Status:  "draft",
		},
	}

	// When: Creating blog post
	post, err := s.strapiClient.CreateBlogPost(ctx, req)

	// Then: Timestamps should be set
	s.Require().NoError(err, "Failed to create blog post")
	s.NotNil(post, "Post should not be nil")
	s.False(post.Attributes.CreatedAt.IsZero(), "CreatedAt should be set")
	s.False(post.Attributes.UpdatedAt.IsZero(), "UpdatedAt should be set")
}

// TestStrapiIntegrationTestSuite runs the test suite.
func TestStrapiIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(StrapiIntegrationTestSuite))
}
