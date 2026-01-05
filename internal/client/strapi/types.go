package strapi

import "time"

// BlogPost represents a blog post in Strapi CMS.
type BlogPost struct {
	ID          int                    `json:"id"`
	Attributes  *BlogPostAttributes    `json:"attributes"`
	Meta        map[string]interface{} `json:"meta,omitempty"`
}

// BlogPostAttributes contains the attributes of a blog post.
type BlogPostAttributes struct {
	Title       string                 `json:"title"`
	Content     string                 `json:"content"`
	Slug        string                 `json:"slug,omitempty"`
	Author      string                 `json:"author,omitempty"`
	Status      string                 `json:"status"` // draft, published
	PublishedAt *time.Time             `json:"publishedAt,omitempty"`
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`
	Tags        []string               `json:"tags,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// CreateBlogPostRequest represents a request to create a new blog post.
type CreateBlogPostRequest struct {
	Data *BlogPostData `json:"data"`
}

// BlogPostData contains the data for creating or updating a blog post.
type BlogPostData struct {
	Title    string                 `json:"title"`
	Content  string                 `json:"content"`
	Slug     string                 `json:"slug,omitempty"`
	Author   string                 `json:"author,omitempty"`
	Status   string                 `json:"status,omitempty"` // draft (default), published
	Tags     []string               `json:"tags,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// CreateBlogPostResponse represents the response from creating a blog post.
type CreateBlogPostResponse struct {
	Data *BlogPost              `json:"data"`
	Meta map[string]interface{} `json:"meta,omitempty"`
}

// UpdateBlogPostRequest represents a request to update a blog post.
type UpdateBlogPostRequest struct {
	Data *BlogPostData `json:"data"`
}

// UpdateBlogPostResponse represents the response from updating a blog post.
type UpdateBlogPostResponse struct {
	Data *BlogPost              `json:"data"`
	Meta map[string]interface{} `json:"meta,omitempty"`
}

// ListBlogPostsResponse represents the response from listing blog posts.
type ListBlogPostsResponse struct {
	Data []*BlogPost `json:"data"`
	Meta *ListMeta   `json:"meta"`
}

// ListMeta contains pagination metadata for list responses.
type ListMeta struct {
	Pagination *Pagination `json:"pagination"`
}

// Pagination contains pagination information.
type Pagination struct {
	Page      int `json:"page"`
	PageSize  int `json:"pageSize"`
	PageCount int `json:"pageCount"`
	Total     int `json:"total"`
}

// GetBlogPostResponse represents the response from getting a single blog post.
type GetBlogPostResponse struct {
	Data *BlogPost              `json:"data"`
	Meta map[string]interface{} `json:"meta,omitempty"`
}

// DeleteBlogPostResponse represents the response from deleting a blog post.
type DeleteBlogPostResponse struct {
	Data *BlogPost              `json:"data"`
	Meta map[string]interface{} `json:"meta,omitempty"`
}

// PublishBlogPostRequest represents a request to publish a blog post.
type PublishBlogPostRequest struct {
	Data *PublishBlogPostData `json:"data"`
}

// PublishBlogPostData contains the data for publishing a blog post.
type PublishBlogPostData struct {
	PublishedAt *time.Time `json:"publishedAt,omitempty"`
}

// PublishBlogPostResponse represents the response from publishing a blog post.
type PublishBlogPostResponse struct {
	Data *BlogPost              `json:"data"`
	Meta map[string]interface{} `json:"meta,omitempty"`
}

// ContentType represents a Strapi content type.
type ContentType struct {
	UID         string                 `json:"uid"`
	DisplayName string                 `json:"displayName"`
	Kind        string                 `json:"kind"` // singleType, collectionType
	Info        *ContentTypeInfo       `json:"info,omitempty"`
	Attributes  map[string]interface{} `json:"attributes,omitempty"`
}

// ContentTypeInfo contains information about a content type.
type ContentTypeInfo struct {
	DisplayName string `json:"displayName"`
	Description string `json:"description,omitempty"`
	Singular    string `json:"singularName"`
	Plural      string `json:"pluralName"`
}

// ListContentTypesResponse represents the response from listing content types.
type ListContentTypesResponse struct {
	Data []*ContentType `json:"data"`
}

// ContentEntry represents a generic content entry in Strapi.
type ContentEntry struct {
	ID         int                    `json:"id"`
	Attributes map[string]interface{} `json:"attributes"`
}

// CreateContentRequest represents a request to create content.
type CreateContentRequest struct {
	Data map[string]interface{} `json:"data"`
}

// CreateContentResponse represents the response from creating content.
type CreateContentResponse struct {
	Data *ContentEntry          `json:"data"`
	Meta map[string]interface{} `json:"meta,omitempty"`
}

// ListContentResponse represents the response from listing content entries.
type ListContentResponse struct {
	Data []*ContentEntry `json:"data"`
	Meta *ListMeta       `json:"meta"`
}

// ListOptions contains options for listing content.
type ListOptions struct {
	Page      int                    `json:"page,omitempty"`
	PageSize  int                    `json:"pageSize,omitempty"`
	Sort      []string               `json:"sort,omitempty"`
	Filters   map[string]interface{} `json:"filters,omitempty"`
	Fields    []string               `json:"fields,omitempty"`
	Populate  interface{}            `json:"populate,omitempty"`
}

// ErrorResponse represents a Strapi API error response.
type ErrorResponse struct {
	Error *ErrorDetail `json:"error"`
}

// ErrorDetail contains error details.
type ErrorDetail struct {
	Status  int                    `json:"status"`
	Name    string                 `json:"name"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}
