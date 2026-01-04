package mocks

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"
)

// DesignServer represents a mock design token API server
type DesignServer struct {
	Server          *httptest.Server
	Tokens          map[string]DesignToken
	mu              sync.RWMutex
	ShouldFailAuth  bool
	ShouldRateLimit bool
	ResponseDelay   time.Duration
	ParseCalled     bool
	ExtractCalled   bool
	ValidateCalled  bool
}

// DesignToken represents a design token
type DesignToken struct {
	Name        string                 `json:"name"`
	Value       string                 `json:"value"`
	Type        string                 `json:"type"` // color, typography, spacing, etc.
	Category    string                 `json:"category"`
	Description string                 `json:"description,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
}

// ParseRequest represents a CSS/SCSS parsing request
type ParseRequest struct {
	Content  string `json:"content"`
	FileType string `json:"file_type"` // css, scss
}

// ParseResponse represents the parsing result
type ParseResponse struct {
	Tokens []DesignToken `json:"tokens"`
	Count  int           `json:"count"`
	Errors []string      `json:"errors,omitempty"`
}

// NewDesignServer creates a new mock design token server
func NewDesignServer() *DesignServer {
	ds := &DesignServer{
		Tokens: make(map[string]DesignToken),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/tokens/parse", ds.handleParse)
	mux.HandleFunc("/api/tokens/extract", ds.handleExtract)
	mux.HandleFunc("/api/tokens/validate", ds.handleValidate)
	mux.HandleFunc("/api/tokens/export", ds.handleExport)

	ds.Server = httptest.NewServer(ds.authMiddleware(mux))
	return ds
}

// Close shuts down the mock server
func (ds *DesignServer) Close() {
	ds.Server.Close()
}

// authMiddleware validates API authentication
func (ds *DesignServer) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ds.ResponseDelay > 0 {
			time.Sleep(ds.ResponseDelay)
		}

		if ds.ShouldFailAuth {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error":   "unauthorized",
				"message": "Invalid API key",
			})
			return
		}

		if ds.ShouldRateLimit {
			w.Header().Set("Retry-After", "60")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]string{
				"error":   "rate_limit_exceeded",
				"message": "Too many requests",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}

// handleParse parses CSS/SCSS files to extract design tokens
func (ds *DesignServer) handleParse(w http.ResponseWriter, r *http.Request) {
	ds.ParseCalled = true

	var req ParseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error":   "invalid_request",
			"message": "Invalid JSON body",
		})
		return
	}

	if req.Content == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error":   "validation_error",
			"message": "content is required",
		})
		return
	}

	// Mock token extraction from CSS/SCSS
	tokens := ds.extractTokensFromContent(req.Content, req.FileType)

	response := ParseResponse{
		Tokens: tokens,
		Count:  len(tokens),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleExtract extracts specific token types
func (ds *DesignServer) handleExtract(w http.ResponseWriter, r *http.Request) {
	ds.ExtractCalled = true

	var req struct {
		Content   string   `json:"content"`
		TokenType []string `json:"token_types"` // color, typography, spacing
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid_request"})
		return
	}

	tokens := ds.extractTokensFromContent(req.Content, "css")

	// Filter by requested types
	if len(req.TokenType) > 0 {
		filtered := make([]DesignToken, 0)
		for _, token := range tokens {
			for _, wantedType := range req.TokenType {
				if token.Type == wantedType {
					filtered = append(filtered, token)
					break
				}
			}
		}
		tokens = filtered
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"tokens": tokens,
		"count":  len(tokens),
	})
}

// handleValidate validates token structure
func (ds *DesignServer) handleValidate(w http.ResponseWriter, r *http.Request) {
	ds.ValidateCalled = true

	var req struct {
		Tokens []DesignToken `json:"tokens"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid_request"})
		return
	}

	errors := make([]string, 0)
	for _, token := range req.Tokens {
		if token.Name == "" {
			errors = append(errors, "Token name is required")
		}
		if token.Value == "" {
			errors = append(errors, "Token value is required")
		}
		if token.Type == "" {
			errors = append(errors, "Token type is required")
		}
	}

	valid := len(errors) == 0

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"valid":  valid,
		"errors": errors,
	})
}

// handleExport exports tokens in various formats
func (ds *DesignServer) handleExport(w http.ResponseWriter, r *http.Request) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json"
	}

	tokens := make([]DesignToken, 0, len(ds.Tokens))
	for _, token := range ds.Tokens {
		tokens = append(tokens, token)
	}

	switch format {
	case "json":
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"tokens": tokens,
			"count":  len(tokens),
		})
	case "css":
		w.Header().Set("Content-Type", "text/css")
		w.Write([]byte(":root {\n"))
		for _, token := range tokens {
			w.Write([]byte("  --" + token.Name + ": " + token.Value + ";\n"))
		}
		w.Write([]byte("}\n"))
	case "scss":
		w.Header().Set("Content-Type", "text/plain")
		for _, token := range tokens {
			w.Write([]byte("$" + token.Name + ": " + token.Value + ";\n"))
		}
	default:
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "unsupported_format",
		})
	}
}

// extractTokensFromContent extracts design tokens from CSS/SCSS content
func (ds *DesignServer) extractTokensFromContent(content, fileType string) []DesignToken {
	tokens := make([]DesignToken, 0)
	now := time.Now()

	// Mock extraction - in reality this would parse CSS/SCSS
	// For testing, we'll create some predictable tokens based on content

	// Extract color tokens (look for color-related patterns)
	if contains(content, "color") || contains(content, "#") {
		tokens = append(tokens, DesignToken{
			Name:      "primary-color",
			Value:     "#007bff",
			Type:      "color",
			Category:  "colors",
			CreatedAt: now,
		})
	}

	// Extract typography tokens
	if contains(content, "font") {
		tokens = append(tokens, DesignToken{
			Name:      "font-family-base",
			Value:     "system-ui, sans-serif",
			Type:      "typography",
			Category:  "fonts",
			CreatedAt: now,
		})
	}

	// Extract spacing tokens
	if contains(content, "spacing") || contains(content, "margin") || contains(content, "padding") {
		tokens = append(tokens, DesignToken{
			Name:      "spacing-base",
			Value:     "16px",
			Type:      "spacing",
			Category:  "spacing",
			CreatedAt: now,
		})
	}

	return tokens
}

// contains is a simple string contains check
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
		 len(s) > len(substr)+1 && containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// GetURL returns the base URL of the mock server
func (ds *DesignServer) GetURL() string {
	return ds.Server.URL
}

// Reset clears all stored tokens
func (ds *DesignServer) Reset() {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.Tokens = make(map[string]DesignToken)
}

// AddToken adds a token for testing
func (ds *DesignServer) AddToken(token DesignToken) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.Tokens[token.Name] = token
}
