package strapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/AINative-studio/ainative-code/internal/client"
	"github.com/AINative-studio/ainative-code/internal/logger"
)

// Client represents a client for Strapi CMS operations.
type Client struct {
	apiClient *client.Client
	baseURL   string
}

// Option is a functional option for configuring the Client.
type Option func(*Client)

// WithAPIClient sets the underlying HTTP API client.
func WithAPIClient(apiClient *client.Client) Option {
	return func(c *Client) {
		c.apiClient = apiClient
	}
}

// WithBaseURL sets the Strapi base URL.
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.baseURL = strings.TrimSuffix(baseURL, "/")
	}
}

// New creates a new Strapi client with the specified options.
func New(opts ...Option) *Client {
	c := &Client{}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// CreateBlogPost creates a new blog post in Strapi.
func (c *Client) CreateBlogPost(ctx context.Context, req *CreateBlogPostRequest) (*BlogPost, error) {
	logger.InfoEvent().
		Str("title", req.Data.Title).
		Str("author", req.Data.Author).
		Str("status", req.Data.Status).
		Msg("Creating blog post in Strapi")

	// Set default status if not provided
	if req.Data.Status == "" {
		req.Data.Status = "draft"
	}

	// Validate required fields
	if req.Data.Title == "" {
		return nil, fmt.Errorf("title is required")
	}
	if req.Data.Content == "" {
		return nil, fmt.Errorf("content is required")
	}

	path := "/api/blog-posts"
	respData, err := c.apiClient.Post(ctx, path, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create blog post: %w", err)
	}

	var resp CreateBlogPostResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	logger.InfoEvent().
		Int("id", resp.Data.ID).
		Str("title", resp.Data.Attributes.Title).
		Msg("Blog post created successfully")

	return resp.Data, nil
}

// ListBlogPosts lists blog posts with optional filters.
func (c *Client) ListBlogPosts(ctx context.Context, opts *ListOptions) ([]*BlogPost, *ListMeta, error) {
	logger.DebugEvent().Msg("Listing blog posts from Strapi")

	if opts == nil {
		opts = &ListOptions{}
	}

	// Set defaults
	if opts.PageSize == 0 {
		opts.PageSize = 25
	}
	if opts.Page == 0 {
		opts.Page = 1
	}

	// Build query parameters
	queryParams := c.buildQueryParams(opts)
	path := "/api/blog-posts?" + queryParams

	respData, err := c.apiClient.Get(ctx, path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list blog posts: %w", err)
	}

	var resp ListBlogPostsResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		return nil, nil, fmt.Errorf("failed to parse response: %w", err)
	}

	logger.DebugEvent().
		Int("count", len(resp.Data)).
		Int("total", resp.Meta.Pagination.Total).
		Msg("Blog posts listed successfully")

	return resp.Data, resp.Meta, nil
}

// GetBlogPost retrieves a single blog post by ID.
func (c *Client) GetBlogPost(ctx context.Context, id int) (*BlogPost, error) {
	logger.DebugEvent().
		Int("id", id).
		Msg("Getting blog post from Strapi")

	path := fmt.Sprintf("/api/blog-posts/%d", id)
	respData, err := c.apiClient.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to get blog post: %w", err)
	}

	var resp GetBlogPostResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	logger.DebugEvent().
		Int("id", id).
		Str("title", resp.Data.Attributes.Title).
		Msg("Blog post retrieved successfully")

	return resp.Data, nil
}

// UpdateBlogPost updates an existing blog post.
func (c *Client) UpdateBlogPost(ctx context.Context, id int, req *UpdateBlogPostRequest) (*BlogPost, error) {
	logger.InfoEvent().
		Int("id", id).
		Str("title", req.Data.Title).
		Msg("Updating blog post in Strapi")

	path := fmt.Sprintf("/api/blog-posts/%d", id)
	respData, err := c.apiClient.Put(ctx, path, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update blog post: %w", err)
	}

	var resp UpdateBlogPostResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	logger.InfoEvent().
		Int("id", id).
		Str("title", resp.Data.Attributes.Title).
		Msg("Blog post updated successfully")

	return resp.Data, nil
}

// PublishBlogPost publishes a blog post by updating its status and publishedAt timestamp.
func (c *Client) PublishBlogPost(ctx context.Context, id int) (*BlogPost, error) {
	logger.InfoEvent().
		Int("id", id).
		Msg("Publishing blog post in Strapi")

	now := time.Now()
	req := &UpdateBlogPostRequest{
		Data: &BlogPostData{
			Status: "published",
		},
	}

	path := fmt.Sprintf("/api/blog-posts/%d", id)
	respData, err := c.apiClient.Put(ctx, path, req)
	if err != nil {
		return nil, fmt.Errorf("failed to publish blog post: %w", err)
	}

	var resp UpdateBlogPostResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Update publishedAt if not already set
	if resp.Data.Attributes.PublishedAt == nil {
		resp.Data.Attributes.PublishedAt = &now
	}

	logger.InfoEvent().
		Int("id", id).
		Str("title", resp.Data.Attributes.Title).
		Msg("Blog post published successfully")

	return resp.Data, nil
}

// DeleteBlogPost deletes a blog post by ID.
func (c *Client) DeleteBlogPost(ctx context.Context, id int) error {
	logger.InfoEvent().
		Int("id", id).
		Msg("Deleting blog post from Strapi")

	path := fmt.Sprintf("/api/blog-posts/%d", id)
	respData, err := c.apiClient.Delete(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to delete blog post: %w", err)
	}

	var resp DeleteBlogPostResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	logger.InfoEvent().
		Int("id", id).
		Msg("Blog post deleted successfully")

	return nil
}

// ListContentTypes lists all available content types in Strapi.
func (c *Client) ListContentTypes(ctx context.Context) ([]*ContentType, error) {
	logger.DebugEvent().Msg("Listing content types from Strapi")

	path := "/api/content-type-builder/content-types"
	respData, err := c.apiClient.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to list content types: %w", err)
	}

	var resp ListContentTypesResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	logger.DebugEvent().
		Int("count", len(resp.Data)).
		Msg("Content types listed successfully")

	return resp.Data, nil
}

// CreateContent creates a new entry in a specified content type.
func (c *Client) CreateContent(ctx context.Context, contentType string, data map[string]interface{}) (*ContentEntry, error) {
	logger.InfoEvent().
		Str("content_type", contentType).
		Msg("Creating content entry in Strapi")

	req := &CreateContentRequest{
		Data: data,
	}

	path := fmt.Sprintf("/api/%s", contentType)
	respData, err := c.apiClient.Post(ctx, path, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create content: %w", err)
	}

	var resp CreateContentResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	logger.InfoEvent().
		Str("content_type", contentType).
		Int("id", resp.Data.ID).
		Msg("Content entry created successfully")

	return resp.Data, nil
}

// ListContent lists entries from a specified content type.
func (c *Client) ListContent(ctx context.Context, contentType string, opts *ListOptions) ([]*ContentEntry, *ListMeta, error) {
	logger.DebugEvent().
		Str("content_type", contentType).
		Msg("Listing content entries from Strapi")

	if opts == nil {
		opts = &ListOptions{}
	}

	// Set defaults
	if opts.PageSize == 0 {
		opts.PageSize = 25
	}
	if opts.Page == 0 {
		opts.Page = 1
	}

	// Build query parameters
	queryParams := c.buildQueryParams(opts)
	path := fmt.Sprintf("/api/%s?%s", contentType, queryParams)

	respData, err := c.apiClient.Get(ctx, path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list content: %w", err)
	}

	var resp ListContentResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		return nil, nil, fmt.Errorf("failed to parse response: %w", err)
	}

	logger.DebugEvent().
		Str("content_type", contentType).
		Int("count", len(resp.Data)).
		Int("total", resp.Meta.Pagination.Total).
		Msg("Content entries listed successfully")

	return resp.Data, resp.Meta, nil
}

// UpdateContent updates an entry in a specified content type.
func (c *Client) UpdateContent(ctx context.Context, contentType string, id int, data map[string]interface{}) (*ContentEntry, error) {
	logger.InfoEvent().
		Str("content_type", contentType).
		Int("id", id).
		Msg("Updating content entry in Strapi")

	req := &CreateContentRequest{
		Data: data,
	}

	path := fmt.Sprintf("/api/%s/%d", contentType, id)
	respData, err := c.apiClient.Put(ctx, path, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update content: %w", err)
	}

	var resp CreateContentResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	logger.InfoEvent().
		Str("content_type", contentType).
		Int("id", id).
		Msg("Content entry updated successfully")

	return resp.Data, nil
}

// DeleteContent deletes an entry from a specified content type.
func (c *Client) DeleteContent(ctx context.Context, contentType string, id int) error {
	logger.InfoEvent().
		Str("content_type", contentType).
		Int("id", id).
		Msg("Deleting content entry from Strapi")

	path := fmt.Sprintf("/api/%s/%d", contentType, id)
	_, err := c.apiClient.Delete(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to delete content: %w", err)
	}

	logger.InfoEvent().
		Str("content_type", contentType).
		Int("id", id).
		Msg("Content entry deleted successfully")

	return nil
}

// buildQueryParams builds URL query parameters from ListOptions.
func (c *Client) buildQueryParams(opts *ListOptions) string {
	params := url.Values{}

	if opts.Page > 0 {
		params.Set("pagination[page]", strconv.Itoa(opts.Page))
	}
	if opts.PageSize > 0 {
		params.Set("pagination[pageSize]", strconv.Itoa(opts.PageSize))
	}

	// Add sort parameters
	for i, sortField := range opts.Sort {
		params.Set(fmt.Sprintf("sort[%d]", i), sortField)
	}

	// Add filter parameters
	if opts.Filters != nil {
		for key, value := range opts.Filters {
			switch v := value.(type) {
			case string:
				params.Set(fmt.Sprintf("filters[%s][$eq]", key), v)
			case []string:
				for i, item := range v {
					params.Set(fmt.Sprintf("filters[%s][$in][%d]", key, i), item)
				}
			case map[string]interface{}:
				// Handle complex filters
				for op, val := range v {
					params.Set(fmt.Sprintf("filters[%s][%s]", key, op), fmt.Sprintf("%v", val))
				}
			default:
				params.Set(fmt.Sprintf("filters[%s][$eq]", key), fmt.Sprintf("%v", v))
			}
		}
	}

	// Add field selection
	for i, field := range opts.Fields {
		params.Set(fmt.Sprintf("fields[%d]", i), field)
	}

	// Add populate
	if opts.Populate != nil {
		switch p := opts.Populate.(type) {
		case string:
			params.Set("populate", p)
		case []string:
			for i, field := range p {
				params.Set(fmt.Sprintf("populate[%d]", i), field)
			}
		case bool:
			if p {
				params.Set("populate", "*")
			}
		}
	}

	return params.Encode()
}
