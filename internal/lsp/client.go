package lsp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
)

// Client represents an LSP client
type Client struct {
	config *LanguageServerConfig
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser

	// Request/response tracking
	responses map[interface{}]chan *JSONRPCResponse
	responseMu sync.RWMutex
	nextID    int64

	// State
	initialized bool
	shutdown    bool
	mu          sync.Mutex

	// Background workers
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewClient creates a new LSP client
func NewClient(config *LanguageServerConfig) (*Client, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	client := &Client{
		config:    config,
		responses: make(map[interface{}]chan *JSONRPCResponse),
		nextID:    1,
		ctx:       ctx,
		cancel:    cancel,
	}

	return client, nil
}

// Start starts the language server process
func (c *Client) Start() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cmd != nil {
		return &JSONRPCError{
			Code:    InternalError,
			Message: "language server already started",
		}
	}

	// Create command
	cmd := exec.CommandContext(c.ctx, c.config.Command, c.config.Args...)

	// Set environment variables
	cmd.Env = os.Environ()
	for k, v := range c.config.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	// Set up pipes
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		stdin.Close()
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		stdin.Close()
		stdout.Close()
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the process
	if err := cmd.Start(); err != nil {
		stdin.Close()
		stdout.Close()
		stderr.Close()
		return fmt.Errorf("failed to start language server: %w", err)
	}

	c.cmd = cmd
	c.stdin = stdin
	c.stdout = stdout
	c.stderr = stderr

	// Start response handler
	c.wg.Add(1)
	go c.handleResponses()

	return nil
}

// Initialize sends the initialize request
func (c *Client) Initialize(ctx context.Context, rootURI *string, initOptions interface{}) (*InitializeResult, error) {
	c.mu.Lock()
	if c.initialized {
		c.mu.Unlock()
		return nil, &JSONRPCError{
			Code:    InternalError,
			Message: "client already initialized",
		}
	}
	c.mu.Unlock()

	processID := os.Getpid()
	params := InitializeParams{
		ProcessID:             &processID,
		RootURI:               rootURI,
		InitializationOptions: initOptions,
		Capabilities:          c.buildClientCapabilities(),
		Trace:                 "off",
	}

	if initOptions == nil && c.config.InitializationOptions != nil {
		params.InitializationOptions = c.config.InitializationOptions
	}

	// Apply initialization timeout
	if c.config.InitTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.config.InitTimeout)
		defer cancel()
	}

	var result InitializeResult
	if err := c.sendRequest(ctx, "initialize", params, &result); err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.initialized = true
	c.mu.Unlock()

	return &result, nil
}

// Initialized sends the initialized notification
func (c *Client) Initialized(ctx context.Context) error {
	return c.sendNotification("initialized", struct{}{})
}

// Shutdown sends the shutdown request
func (c *Client) Shutdown(ctx context.Context) error {
	c.mu.Lock()
	if c.shutdown {
		c.mu.Unlock()
		return nil
	}
	c.mu.Unlock()

	if err := c.sendRequest(ctx, "shutdown", nil, nil); err != nil {
		return err
	}

	c.mu.Lock()
	c.shutdown = true
	c.mu.Unlock()

	return nil
}

// Exit sends the exit notification
func (c *Client) Exit(ctx context.Context) error {
	return c.sendNotification("exit", nil)
}

// Completion sends a textDocument/completion request
func (c *Client) Completion(ctx context.Context, params CompletionParams) (*CompletionList, error) {
	if !c.isInitialized() {
		return nil, &JSONRPCError{
			Code:    ServerNotInitialized,
			Message: "server not initialized",
		}
	}

	var result CompletionList
	if err := c.sendRequest(ctx, "textDocument/completion", params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Hover sends a textDocument/hover request
func (c *Client) Hover(ctx context.Context, params HoverParams) (*Hover, error) {
	if !c.isInitialized() {
		return nil, &JSONRPCError{
			Code:    ServerNotInitialized,
			Message: "server not initialized",
		}
	}

	var result Hover
	if err := c.sendRequest(ctx, "textDocument/hover", params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Definition sends a textDocument/definition request
func (c *Client) Definition(ctx context.Context, params DefinitionParams) (*Location, error) {
	if !c.isInitialized() {
		return nil, &JSONRPCError{
			Code:    ServerNotInitialized,
			Message: "server not initialized",
		}
	}

	var result Location
	if err := c.sendRequest(ctx, "textDocument/definition", params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// References sends a textDocument/references request
func (c *Client) References(ctx context.Context, params ReferenceParams) ([]Location, error) {
	if !c.isInitialized() {
		return nil, &JSONRPCError{
			Code:    ServerNotInitialized,
			Message: "server not initialized",
		}
	}

	var result []Location
	if err := c.sendRequest(ctx, "textDocument/references", params, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// Close closes the client and terminates the language server
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cancel != nil {
		c.cancel()
	}

	if c.stdin != nil {
		c.stdin.Close()
	}

	if c.stdout != nil {
		c.stdout.Close()
	}

	if c.stderr != nil {
		c.stderr.Close()
	}

	if c.cmd != nil && c.cmd.Process != nil {
		c.cmd.Process.Kill()
		c.cmd.Wait()
	}

	c.wg.Wait()

	return nil
}

// sendRequest sends a JSON-RPC request and waits for response
func (c *Client) sendRequest(ctx context.Context, method string, params interface{}, result interface{}) error {
	id := c.getNextID()

	request := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}

	// Create response channel
	respChan := make(chan *JSONRPCResponse, 1)
	c.responseMu.Lock()
	c.responses[id] = respChan
	c.responseMu.Unlock()

	defer func() {
		c.responseMu.Lock()
		delete(c.responses, id)
		c.responseMu.Unlock()
	}()

	// Send request
	if err := c.writeMessage(&request); err != nil {
		return err
	}

	// Apply request timeout if configured
	if c.config.RequestTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.config.RequestTimeout)
		defer cancel()
	}

	// Wait for response
	select {
	case resp := <-respChan:
		if resp.Error != nil {
			return resp.Error
		}

		if result != nil && len(resp.Result) > 0 {
			if err := json.Unmarshal(resp.Result, result); err != nil {
				return fmt.Errorf("failed to unmarshal result: %w", err)
			}
		}

		return nil

	case <-ctx.Done():
		return ctx.Err()
	}
}

// sendNotification sends a JSON-RPC notification (no response expected)
func (c *Client) sendNotification(method string, params interface{}) error {
	notification := JSONRPCNotification{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}

	return c.writeMessage(&notification)
}

// writeMessage writes a message to the language server
func (c *Client) writeMessage(msg interface{}) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(data))

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.stdin == nil {
		return &JSONRPCError{
			Code:    InternalError,
			Message: "stdin not available",
		}
	}

	if _, err := c.stdin.Write([]byte(header)); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	if _, err := c.stdin.Write(data); err != nil {
		return fmt.Errorf("failed to write data: %w", err)
	}

	return nil
}

// handleResponses reads and processes responses from the language server
func (c *Client) handleResponses() {
	defer c.wg.Done()

	reader := bufio.NewReader(c.stdout)

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		// Read Content-Length header
		header, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				// Log error
			}
			return
		}

		if !strings.HasPrefix(header, "Content-Length: ") {
			continue
		}

		var length int
		if _, err := fmt.Sscanf(header, "Content-Length: %d\r\n", &length); err != nil {
			continue
		}

		// Read empty line
		if _, err := reader.ReadString('\n'); err != nil {
			return
		}

		// Read content
		content := make([]byte, length)
		if _, err := io.ReadFull(reader, content); err != nil {
			return
		}

		// Try to parse as response first
		var response JSONRPCResponse
		if err := json.Unmarshal(content, &response); err == nil && response.ID != nil {
			c.responseMu.RLock()
			respChan, ok := c.responses[response.ID]
			c.responseMu.RUnlock()

			if ok {
				respChan <- &response
			}
			continue
		}

		// Could be a notification (no ID)
		// We ignore notifications for now
	}
}

// buildClientCapabilities builds the client capabilities
func (c *Client) buildClientCapabilities() ClientCapabilities {
	return ClientCapabilities{
		Workspace: WorkspaceClientCapabilities{
			ApplyEdit: true,
			WorkspaceEdit: &WorkspaceEditCapabilities{
				DocumentChanges: true,
			},
			DidChangeConfiguration: &DidChangeConfigurationCapabilities{
				DynamicRegistration: true,
			},
		},
		TextDocument: TextDocumentClientCapabilities{
			Synchronization: &TextDocumentSyncClientCapabilities{
				DynamicRegistration: true,
				WillSave:            true,
				WillSaveWaitUntil:   true,
				DidSave:             true,
			},
			Completion: &CompletionCapabilities{
				DynamicRegistration: true,
				CompletionItem: &CompletionItemCapabilities{
					SnippetSupport:          true,
					CommitCharactersSupport: true,
					DocumentationFormat:     []string{MarkupKindMarkdown, MarkupKindPlainText},
				},
			},
			Hover: &HoverCapabilities{
				DynamicRegistration: true,
				ContentFormat:       []string{MarkupKindMarkdown, MarkupKindPlainText},
			},
			Definition: &DefinitionCapabilities{
				DynamicRegistration: true,
				LinkSupport:         true,
			},
			References: &ReferencesCapabilities{
				DynamicRegistration: true,
			},
		},
	}
}

// getNextID returns the next request ID
func (c *Client) getNextID() int64 {
	return atomic.AddInt64(&c.nextID, 1) - 1
}

// isInitialized returns true if the client is initialized
func (c *Client) isInitialized() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.initialized
}

// IsShutdown returns true if the client is shutdown
func (c *Client) IsShutdown() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.shutdown
}
