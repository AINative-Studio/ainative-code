package mocks

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"
)

// StrapiServer represents a mock Strapi CMS server
type StrapiServer struct {
	Server          *httptest.Server
	Posts           map[string]*BlogPost
	mu              sync.RWMutex
	NextID          int
	ShouldFailAuth  bool
	ShouldRateLimit bool
	ResponseDelay   time.Duration
	CreateCalled    bool
	ListCalled      bool
	UpdateCalled    bool
	DeleteCalled    bool
	PublishCalled   bool
}

// BlogPost represents a Strapi blog post
type BlogPost struct {
	ID          int                    `json:"id"`
	Attributes  map[string]interface{} `json:"attributes"`
	PublishedAt *time.Time             `json:"publishedAt"`
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`
}

// NewStrapiServer creates a new mock Strapi server
func NewStrapiServer() *StrapiServer {
	ss := &StrapiServer{
		Posts:  make(map[string]*BlogPost),
		NextID: 1,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/posts", ss.handlePosts)
	mux.HandleFunc("/api/posts/", ss.handlePostByID)

	ss.Server = httptest.NewServer(ss.authMiddleware(mux))
	return ss
}

// Close shuts down the mock server
func (ss *StrapiServer) Close() {
	ss.Server.Close()
}

// authMiddleware validates API authentication
func (ss *StrapiServer) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ss.ResponseDelay > 0 {
			time.Sleep(ss.ResponseDelay)
		}

		if ss.ShouldFailAuth {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": map[string]interface{}{
					"status":  401,
					"name":    "UnauthorizedError",
					"message": "Missing or invalid credentials",
				},
			})
			return
		}

		if ss.ShouldRateLimit {
			w.Header().Set("Retry-After", "60")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": map[string]interface{}{
					"status":  429,
					"name":    "RateLimitError",
					"message": "Too many requests",
				},
			})
			return
		}

		// Validate authorization header
		auth := r.Header.Get("Authorization")
		if auth == "" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": map[string]interface{}{
					"status":  401,
					"message": "Missing authorization header",
				},
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}

// handlePosts handles POST and GET requests to /api/posts
func (ss *StrapiServer) handlePosts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		ss.handleCreatePost(w, r)
	case http.MethodGet:
		ss.handleListPosts(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "method_not_allowed"})
	}
}

// handlePostByID handles GET, PUT, and DELETE requests to /api/posts/:id
func (ss *StrapiServer) handlePostByID(w http.ResponseWriter, r *http.Request) {
	// Extract ID from path
	id := r.URL.Path[len("/api/posts/"):]

	switch r.Method {
	case http.MethodGet:
		ss.handleGetPost(w, r, id)
	case http.MethodPut:
		ss.handleUpdatePost(w, r, id)
	case http.MethodDelete:
		ss.handleDeletePost(w, r, id)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "method_not_allowed"})
	}
}

// handleCreatePost creates a new blog post
func (ss *StrapiServer) handleCreatePost(w http.ResponseWriter, r *http.Request) {
	ss.CreateCalled = true
	ss.mu.Lock()
	defer ss.mu.Unlock()

	var req struct {
		Data map[string]interface{} `json:"data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]string{
				"message": "Invalid request body",
			},
		})
		return
	}

	now := time.Now()
	post := &BlogPost{
		ID:         ss.NextID,
		Attributes: req.Data,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	id := fmt.Sprintf("%d", ss.NextID)
	ss.Posts[id] = post
	ss.NextID++

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": post,
	})
}

// handleListPosts lists all blog posts
func (ss *StrapiServer) handleListPosts(w http.ResponseWriter, r *http.Request) {
	ss.ListCalled = true
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	posts := make([]*BlogPost, 0, len(ss.Posts))
	for _, post := range ss.Posts {
		posts = append(posts, post)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": posts,
		"meta": map[string]interface{}{
			"pagination": map[string]interface{}{
				"page":      1,
				"pageSize":  25,
				"pageCount": 1,
				"total":     len(posts),
			},
		},
	})
}

// handleGetPost retrieves a specific blog post
func (ss *StrapiServer) handleGetPost(w http.ResponseWriter, r *http.Request, id string) {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	post, exists := ss.Posts[id]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"status":  404,
				"message": "Post not found",
			},
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": post,
	})
}

// handleUpdatePost updates a blog post
func (ss *StrapiServer) handleUpdatePost(w http.ResponseWriter, r *http.Request, id string) {
	ss.UpdateCalled = true
	ss.mu.Lock()
	defer ss.mu.Unlock()

	post, exists := ss.Posts[id]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"status":  404,
				"message": "Post not found",
			},
		})
		return
	}

	var req struct {
		Data map[string]interface{} `json:"data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]string{
				"message": "Invalid request body",
			},
		})
		return
	}

	// Check if publishing/unpublishing
	if publishedAt, ok := req.Data["publishedAt"]; ok {
		if publishedAt != nil {
			ss.PublishCalled = true
			now := time.Now()
			post.PublishedAt = &now
		} else {
			post.PublishedAt = nil
		}
	}

	// Update attributes
	for k, v := range req.Data {
		post.Attributes[k] = v
	}
	post.UpdatedAt = time.Now()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": post,
	})
}

// handleDeletePost deletes a blog post
func (ss *StrapiServer) handleDeletePost(w http.ResponseWriter, r *http.Request, id string) {
	ss.DeleteCalled = true
	ss.mu.Lock()
	defer ss.mu.Unlock()

	post, exists := ss.Posts[id]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"status":  404,
				"message": "Post not found",
			},
		})
		return
	}

	delete(ss.Posts, id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": post,
	})
}

// GetURL returns the base URL of the mock server
func (ss *StrapiServer) GetURL() string {
	return ss.Server.URL
}

// Reset clears all stored posts
func (ss *StrapiServer) Reset() {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	ss.Posts = make(map[string]*BlogPost)
	ss.NextID = 1
}

// AddPost adds a post for testing
func (ss *StrapiServer) AddPost(attributes map[string]interface{}) string {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	now := time.Now()
	post := &BlogPost{
		ID:         ss.NextID,
		Attributes: attributes,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	id := fmt.Sprintf("%d", ss.NextID)
	ss.Posts[id] = post
	ss.NextID++

	return id
}
