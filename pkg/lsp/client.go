package lsp

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"
)

// Client represents an LSP client
type Client struct {
	config       Config
	workspace    string
	connected    bool
	status       ConnectionStatus
	mu           sync.RWMutex
	cache        *cache
	pendingReqs  map[string]context.CancelFunc
	reqMu        sync.Mutex
	lastReqTime  time.Time
	debounceTimer *time.Timer
}

// NewClient creates a new LSP client with default configuration
func NewClient() *Client {
	return NewClientWithConfig(DefaultConfig())
}

// NewClientWithConfig creates a new LSP client with custom configuration
func NewClientWithConfig(config Config) *Client {
	client := &Client{
		config:      config,
		connected:   false,
		status:      StatusDisconnected,
		pendingReqs: make(map[string]context.CancelFunc),
	}

	if config.EnableCache {
		client.cache = newCache(config.CacheSize)
	}

	return client
}

// Initialize initializes the LSP client with a workspace
func (c *Client) Initialize(ctx context.Context, workspace string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if workspace exists
	if _, err := os.Stat(workspace); os.IsNotExist(err) {
		return fmt.Errorf("workspace path does not exist: %s", workspace)
	}

	// Check context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	c.workspace = workspace
	c.connected = true
	c.status = StatusConnected

	return nil
}

// IsConnected returns whether the client is connected
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// GetStatus returns the current connection status
func (c *Client) GetStatus() ConnectionStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.status
}

// GetCompletion requests completion items at a given position
func (c *Client) GetCompletion(ctx context.Context, params CompletionParams) ([]CompletionItem, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("LSP client not connected")
	}

	// Apply timeout
	ctx, cancel := context.WithTimeout(ctx, c.config.RequestTimeout)
	defer cancel()

	// Check cache
	if c.cache != nil {
		cacheKey := fmt.Sprintf("completion:%s:%d:%d", params.TextDocument.URI, params.Position.Line, params.Position.Character)
		if cached, ok := c.cache.get(cacheKey); ok {
			if items, ok := cached.([]CompletionItem); ok {
				return items, nil
			}
		}
	}

	// Simulate LSP request (in real implementation, this would call gopls or other LSP server)
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(50 * time.Millisecond):
		// Simulate response
		items := c.generateMockCompletions(params)

		// Cache result
		if c.cache != nil {
			cacheKey := fmt.Sprintf("completion:%s:%d:%d", params.TextDocument.URI, params.Position.Line, params.Position.Character)
			c.cache.set(cacheKey, items)
		}

		return items, nil
	}
}

// GetHover requests hover information at a given position
func (c *Client) GetHover(ctx context.Context, params HoverParams) (*Hover, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("LSP client not connected")
	}

	// Apply timeout
	ctx, cancel := context.WithTimeout(ctx, c.config.RequestTimeout)
	defer cancel()

	// Check cache
	if c.cache != nil {
		cacheKey := fmt.Sprintf("hover:%s:%d:%d", params.TextDocument.URI, params.Position.Line, params.Position.Character)
		if cached, ok := c.cache.get(cacheKey); ok {
			if hover, ok := cached.(*Hover); ok {
				return hover, nil
			}
		}
	}

	// Simulate LSP request
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(30 * time.Millisecond):
		hover := c.generateMockHover(params)

		// Cache result
		if c.cache != nil && hover != nil {
			cacheKey := fmt.Sprintf("hover:%s:%d:%d", params.TextDocument.URI, params.Position.Line, params.Position.Character)
			c.cache.set(cacheKey, hover)
		}

		return hover, nil
	}
}

// GetDefinition requests definition locations for a symbol at a given position
func (c *Client) GetDefinition(ctx context.Context, params DefinitionParams) ([]Location, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("LSP client not connected")
	}

	// Apply timeout
	ctx, cancel := context.WithTimeout(ctx, c.config.RequestTimeout)
	defer cancel()

	// Simulate LSP request
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(40 * time.Millisecond):
		return c.generateMockDefinitions(params), nil
	}
}

// GetReferences requests reference locations for a symbol at a given position
func (c *Client) GetReferences(ctx context.Context, params ReferencesParams) ([]Location, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("LSP client not connected")
	}

	// Apply timeout
	ctx, cancel := context.WithTimeout(ctx, c.config.RequestTimeout)
	defer cancel()

	// Simulate LSP request
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(60 * time.Millisecond):
		return c.generateMockReferences(params), nil
	}
}

// Shutdown shuts down the LSP client
func (c *Client) Shutdown(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return nil // Already shut down
	}

	// Cancel all pending requests
	c.reqMu.Lock()
	for _, cancel := range c.pendingReqs {
		cancel()
	}
	c.pendingReqs = make(map[string]context.CancelFunc)
	c.reqMu.Unlock()

	// Clear cache
	if c.cache != nil {
		c.cache.clear()
	}

	c.connected = false
	c.status = StatusDisconnected

	return nil
}

// Mock data generators for testing

func (c *Client) generateMockCompletions(params CompletionParams) []CompletionItem {
	// Check if position is in a valid range (simple validation)
	if params.Position.Line < 0 || params.Position.Character < 0 {
		return []CompletionItem{}
	}

	return []CompletionItem{
		{
			Label:  "Model",
			Kind:   CompletionItemKindStruct,
			Detail: "type Model struct",
			Documentation: "Model represents the TUI application state",
		},
		{
			Label:  "NewModel",
			Kind:   CompletionItemKindFunction,
			Detail: "func NewModel() Model",
			Documentation: "NewModel creates a new TUI model",
		},
		{
			Label:  "Message",
			Kind:   CompletionItemKindStruct,
			Detail: "type Message struct",
			Documentation: "Message represents a chat message",
		},
	}
}

func (c *Client) generateMockHover(params HoverParams) *Hover {
	// Return nil for invalid positions
	if params.Position.Line <= 0 && params.Position.Character <= 0 {
		return nil
	}

	return &Hover{
		Contents: MarkupContent{
			Kind: "markdown",
			Value: "```go\ntype Model struct\n```\n\n" +
				"Model represents the TUI application state\n\n" +
				"**Fields:**\n" +
				"- viewport: viewport.Model\n" +
				"- textInput: textinput.Model\n" +
				"- messages: []Message",
		},
		Range: &Range{
			Start: params.Position,
			End:   Position{Line: params.Position.Line, Character: params.Position.Character + 5},
		},
	}
}

func (c *Client) generateMockDefinitions(params DefinitionParams) []Location {
	// Return empty for invalid positions
	if params.Position.Line <= 0 && params.Position.Character <= 0 {
		return []Location{}
	}

	return []Location{
		{
			URI: params.TextDocument.URI,
			Range: Range{
				Start: Position{Line: 9, Character: 5},
				End:   Position{Line: 9, Character: 10},
			},
		},
	}
}

func (c *Client) generateMockReferences(params ReferencesParams) []Location {
	// Return empty for invalid positions
	if params.Position.Line <= 0 && params.Position.Character <= 0 {
		return []Location{}
	}

	return []Location{
		{
			URI: params.TextDocument.URI,
			Range: Range{
				Start: Position{Line: 9, Character: 5},
				End:   Position{Line: 9, Character: 10},
			},
		},
		{
			URI: params.TextDocument.URI,
			Range: Range{
				Start: Position{Line: 37, Character: 9},
				End:   Position{Line: 37, Character: 14},
			},
		},
	}
}

// cache implements a simple LRU cache
type cache struct {
	data     map[string]interface{}
	keys     []string
	maxSize  int
	mu       sync.RWMutex
}

func newCache(maxSize int) *cache {
	return &cache{
		data:    make(map[string]interface{}),
		keys:    make([]string, 0, maxSize),
		maxSize: maxSize,
	}
}

func (c *cache) get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.data[key]
	return val, ok
}

func (c *cache) set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// If key exists, update it
	if _, exists := c.data[key]; exists {
		c.data[key] = value
		return
	}

	// If cache is full, remove oldest entry
	if len(c.keys) >= c.maxSize {
		oldestKey := c.keys[0]
		delete(c.data, oldestKey)
		c.keys = c.keys[1:]
	}

	// Add new entry
	c.data[key] = value
	c.keys = append(c.keys, key)
}

func (c *cache) clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = make(map[string]interface{})
	c.keys = make([]string, 0, c.maxSize)
}
