# Component Architecture

This document describes the detailed design of major components in AINative Code.

## Provider Architecture

### Overview

The provider layer abstracts different LLM providers behind a common interface, enabling seamless switching between models and implementing fallback strategies.

### Interface Design

```go
// Provider defines the interface all LLM providers must implement
type Provider interface {
    // Name returns the provider identifier
    Name() string

    // Chat sends a message and returns a complete response
    Chat(ctx context.Context, messages []Message, opts ...Option) (*Response, error)

    // Stream sends a message and returns a channel of streaming events
    Stream(ctx context.Context, messages []Message, opts ...Option) (<-chan StreamEvent, error)

    // ValidateConfig checks if the provider configuration is valid
    ValidateConfig(config *Config) error
}

// Message represents a chat message
type Message struct {
    Role    string      // user, assistant, system, tool_use, tool_result
    Content interface{} // string or structured content
}

// Response represents a complete LLM response
type Response struct {
    Content    string
    ToolCalls  []ToolCall
    Usage      Usage
    StopReason string
}

// StreamEvent represents a streaming chunk
type StreamEvent struct {
    Type    string // text_delta, tool_call, usage, done
    Content string
    Delta   string
}
```

### Implementation Pattern

Each provider implements the interface with provider-specific logic:

```
anthropic/
├── client.go       # Anthropic API client
├── provider.go     # Provider interface implementation
├── messages.go     # Message format conversion
├── streaming.go    # Streaming response handler
└── errors.go       # Error mapping

openai/
├── client.go
├── provider.go
├── messages.go
├── streaming.go
└── errors.go
```

### Message Format Translation

Each provider translates the common message format to its specific API format:

**Common Format → Anthropic**:
```go
func (p *Provider) formatMessages(msgs []Message) []anthropic.Message {
    // Convert common format to Anthropic's format
    // Handle system messages separately
    // Convert tool calls to Anthropic format
}
```

**Common Format → OpenAI**:
```go
func (p *Provider) formatMessages(msgs []Message) []openai.Message {
    // Convert to OpenAI format
    // Include system message in messages array
    // Convert tool calls to OpenAI function calling format
}
```

### Streaming Implementation

Streaming is implemented using Go channels:

```go
func (p *Provider) Stream(ctx context.Context, msgs []Message, opts ...Option) (<-chan StreamEvent, error) {
    events := make(chan StreamEvent)

    go func() {
        defer close(events)

        // Create SSE reader
        reader := sse.NewReader(resp.Body)

        for {
            select {
            case <-ctx.Done():
                return
            default:
                event, err := reader.Next()
                if err != nil {
                    events <- StreamEvent{Type: "error", Content: err.Error()}
                    return
                }

                // Parse and send event
                events <- parseStreamEvent(event)
            }
        }
    }()

    return events, nil
}
```

## Session Management

### Database Schema

```sql
CREATE TABLE sessions (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    metadata TEXT, -- JSON
    provider TEXT,
    model TEXT,
    total_tokens INTEGER DEFAULT 0,
    total_cost REAL DEFAULT 0.0
);

CREATE TABLE messages (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    role TEXT NOT NULL,
    content TEXT NOT NULL,
    timestamp INTEGER NOT NULL,
    token_count INTEGER DEFAULT 0,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
);

CREATE TABLE attachments (
    id TEXT PRIMARY KEY,
    message_id TEXT NOT NULL,
    type TEXT NOT NULL, -- image, file, url
    data BLOB,
    metadata TEXT, -- JSON
    FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE
);

CREATE INDEX idx_sessions_updated ON sessions(updated_at DESC);
CREATE INDEX idx_messages_session ON messages(session_id, timestamp);
```

### Session Manager

```go
type SessionManager struct {
    db     *sql.DB
    cache  *Cache
    events chan Event
}

// Create a new session
func (m *SessionManager) Create(title string, opts ...Option) (*Session, error)

// Load an existing session
func (m *SessionManager) Load(id string) (*Session, error)

// Save session changes
func (m *SessionManager) Save(session *Session) error

// Add message to session
func (m *SessionManager) AddMessage(sessionID string, msg Message) error

// List sessions with filtering
func (m *SessionManager) List(filter Filter, page int, pageSize int) ([]Session, error)

// Export session
func (m *SessionManager) Export(sessionID string, format string) ([]byte, error)
```

### Auto-Save Strategy

Sessions are automatically saved:
- After every message exchange
- On graceful shutdown
- Every 10 messages (configurable)
- On session switch

Crash recovery:
- Checkpoint file written before risky operations
- Recovery on startup checks for incomplete sessions
- User prompted to restore or discard

## Authentication System

### Three-Tier Validation

```go
type TokenValidator struct {
    rsaValidator   *RSAValidator
    apiValidator   *APIValidator
    localValidator *LocalValidator
    cache          *KeyCache
}

func (v *TokenValidator) Validate(token string) (*Claims, error) {
    // Tier 1: Local RSA validation
    if claims, err := v.rsaValidator.Validate(token); err == nil {
        return claims, nil
    }

    // Tier 2: API validation
    if claims, err := v.apiValidator.Validate(token); err == nil {
        // Cache public key for future validations
        v.cache.Store(claims.KeyID, claims.PublicKey)
        return claims, nil
    }

    // Tier 3: Local auth fallback
    return v.localValidator.Validate(token)
}
```

### OAuth Flow Implementation

```go
type OAuthManager struct {
    config       *oauth2.Config
    verifier     string // PKCE code verifier
    state        string // CSRF protection
    callbackSrv  *http.Server
    resultChan   chan *oauth2.Token
}

func (m *OAuthManager) Login(ctx context.Context) (*Token, error) {
    // 1. Generate PKCE verifier and challenge
    m.verifier = generateVerifier()
    challenge := sha256Challenge(m.verifier)

    // 2. Generate state for CSRF protection
    m.state = generateState()

    // 3. Build authorization URL
    authURL := m.config.AuthCodeURL(m.state,
        oauth2.SetAuthURLParam("code_challenge", challenge),
        oauth2.SetAuthURLParam("code_challenge_method", "S256"),
    )

    // 4. Start local callback server
    m.startCallbackServer()

    // 5. Open browser to authorization URL
    browser.Open(authURL)

    // 6. Wait for callback
    select {
    case token := <-m.resultChan:
        return token, nil
    case <-ctx.Done():
        return nil, ctx.Err()
    }
}

func (m *OAuthManager) handleCallback(w http.ResponseWriter, r *http.Request) {
    // Validate state
    if r.URL.Query().Get("state") != m.state {
        http.Error(w, "Invalid state", http.StatusBadRequest)
        return
    }

    // Exchange code for token
    code := r.URL.Query().Get("code")
    token, err := m.config.Exchange(context.Background(), code,
        oauth2.SetAuthURLParam("code_verifier", m.verifier),
    )

    if err != nil {
        http.Error(w, "Token exchange failed", http.StatusInternalServerError)
        return
    }

    // Send token back to login flow
    m.resultChan <- token

    // Show success page
    w.Write([]byte("Authentication successful! You can close this window."))
}
```

### Token Refresh

```go
func (m *AuthManager) refreshToken(ctx context.Context) error {
    // Get current refresh token
    refreshToken, err := m.keyring.Get("refresh_token")
    if err != nil {
        return err
    }

    // Exchange for new access token
    newToken, err := m.oauth.RefreshToken(ctx, refreshToken)
    if err != nil {
        return err
    }

    // Store new tokens
    m.keyring.Set("access_token", newToken.AccessToken)
    m.keyring.Set("refresh_token", newToken.RefreshToken)

    // Update cache
    m.tokenCache.Set(newToken.AccessToken, newToken.ExpiresAt)

    // Schedule next refresh
    m.scheduleRefresh(newToken.ExpiresAt)

    return nil
}
```

## Tool Execution Framework

### Tool Interface

```go
type Tool interface {
    Name() string
    Description() string
    Parameters() []Parameter
    Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error)
}

type Parameter struct {
    Name        string
    Type        string
    Description string
    Required    bool
    Default     interface{}
}

type ToolResult struct {
    Success bool
    Output  string
    Error   string
    Data    interface{}
}
```

### Built-in Tools

#### Bash Tool

```go
type BashTool struct {
    allowList  []string
    blockList  []string
    workDir    string
    timeout    time.Duration
    sandboxed  bool
}

func (t *BashTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
    command := args["command"].(string)

    // Check permissions
    if !t.isAllowed(command) {
        return nil, ErrCommandBlocked
    }

    // Create context with timeout
    ctx, cancel := context.WithTimeout(ctx, t.timeout)
    defer cancel()

    // Execute command
    cmd := exec.CommandContext(ctx, "sh", "-c", command)
    cmd.Dir = t.workDir

    output, err := cmd.CombinedOutput()

    return &ToolResult{
        Success: err == nil,
        Output:  string(output),
        Error:   errorString(err),
    }, nil
}
```

#### File Operations Tool

```go
type FileOpsTool struct {
    basePath   string
    maxSize    int64
    allowedExt []string
}

func (t *FileOpsTool) Read(path string) (*ToolResult, error) {
    // Validate path is within basePath
    absPath, err := filepath.Abs(path)
    if err != nil || !strings.HasPrefix(absPath, t.basePath) {
        return nil, ErrPathNotAllowed
    }

    // Check file size
    stat, err := os.Stat(absPath)
    if err != nil {
        return nil, err
    }
    if stat.Size() > t.maxSize {
        return nil, ErrFileTooLarge
    }

    // Read file
    content, err := os.ReadFile(absPath)
    return &ToolResult{
        Success: err == nil,
        Output:  string(content),
        Error:   errorString(err),
    }, nil
}
```

### Permission System

```go
type PermissionManager struct {
    config   *Config
    prompter Prompter
}

func (m *PermissionManager) CheckPermission(tool string, args map[string]interface{}) (bool, error) {
    // Check global tool permissions
    if !m.config.Tools[tool].Enabled {
        return false, ErrToolDisabled
    }

    // Check if confirmation required
    if m.config.Tools[tool].RequireConfirmation {
        return m.promptUser(tool, args)
    }

    // Check for dangerous operations
    if m.isDangerous(tool, args) {
        return m.promptUser(tool, args)
    }

    return true, nil
}

func (m *PermissionManager) isDangerous(tool string, args map[string]interface{}) bool {
    switch tool {
    case "bash":
        cmd := args["command"].(string)
        dangerousPatterns := []string{"rm -rf", "sudo", "mkfs", "dd if="}
        for _, pattern := range dangerousPatterns {
            if strings.Contains(cmd, pattern) {
                return true
            }
        }
    case "file_write":
        path := args["path"].(string)
        systemPaths := []string{"/etc", "/bin", "/usr/bin", "/System"}
        for _, syspath := range systemPaths {
            if strings.HasPrefix(path, syspath) {
                return true
            }
        }
    }
    return false
}
```

## Event System

### Event Bus Architecture

```go
type EventBus struct {
    subscribers map[string][]Subscriber
    mu          sync.RWMutex
    buffer      chan Event
}

type Event struct {
    Type      string
    Timestamp time.Time
    Data      interface{}
    Source    string
}

type Subscriber func(Event)

func (b *EventBus) Subscribe(eventType string, subscriber Subscriber) {
    b.mu.Lock()
    defer b.mu.Unlock()

    b.subscribers[eventType] = append(b.subscribers[eventType], subscriber)
}

func (b *EventBus) Publish(event Event) {
    b.buffer <- event
}

func (b *EventBus) run() {
    for event := range b.buffer {
        b.mu.RLock()
        subscribers := b.subscribers[event.Type]
        b.mu.RUnlock()

        for _, subscriber := range subscribers {
            go subscriber(event) // Async notification
        }
    }
}
```

### Event Types

```go
const (
    EventMessageSent     = "message.sent"
    EventMessageReceived = "message.received"
    EventStreamStart     = "stream.start"
    EventStreamChunk     = "stream.chunk"
    EventStreamEnd       = "stream.end"
    EventToolExecuted    = "tool.executed"
    EventSessionCreated  = "session.created"
    EventSessionUpdated  = "session.updated"
    EventAuthLogin       = "auth.login"
    EventAuthLogout      = "auth.logout"
    EventProviderSwitch  = "provider.switch"
    EventError           = "error"
)
```

## Configuration System

### Hierarchical Configuration

```go
type ConfigManager struct {
    viper    *viper.Viper
    resolver *KeyResolver
    watchers []Watcher
}

// Load configuration from multiple sources
func (m *ConfigManager) Load() (*Config, error) {
    // 1. Set defaults
    m.viper.SetDefault("llm.default_provider", "anthropic")
    m.viper.SetDefault("ui.theme", "default")

    // 2. Read config file
    m.viper.SetConfigName("config")
    m.viper.SetConfigType("yaml")
    m.viper.AddConfigPath("$HOME/.config/ainative-code")
    m.viper.ReadInConfig()

    // 3. Read environment variables
    m.viper.SetEnvPrefix("AINATIVE_CODE")
    m.viper.AutomaticEnv()

    // 4. Parse into struct
    var config Config
    if err := m.viper.Unmarshal(&config); err != nil {
        return nil, err
    }

    // 5. Resolve dynamic values
    m.resolver.Resolve(&config)

    return &config, nil
}
```

### Dynamic API Key Resolution

```go
type KeyResolver struct {
    cache map[string]string
    ttl   time.Duration
}

func (r *KeyResolver) Resolve(value string) (string, error) {
    // Check cache first
    if cached, ok := r.cache[value]; ok {
        return cached, nil
    }

    // Match pattern: $(command)
    re := regexp.MustCompile(`\$\(([^)]+)\)`)
    matches := re.FindStringSubmatch(value)
    if len(matches) == 0 {
        // Not a command, return as-is
        return value, nil
    }

    // Execute command to get value
    cmd := exec.Command("sh", "-c", matches[1])
    output, err := cmd.Output()
    if err != nil {
        return "", err
    }

    resolved := strings.TrimSpace(string(output))

    // Cache result
    r.cache[value] = resolved

    return resolved, nil
}
```

## TUI Architecture

### Bubble Tea Model

```go
type Model struct {
    // View state
    activeView  View
    chatView    *ChatView
    configView  *ConfigView
    sessionView *SessionView

    // Data
    session  *Session
    messages []Message

    // UI state
    width   int
    height  int
    spinner spinner.Model

    // Services
    provider   Provider
    sessionMgr *SessionManager
    eventBus   *EventBus
}

func (m Model) Init() tea.Cmd {
    return tea.Batch(
        m.spinner.Tick,
        listenForEvents(m.eventBus),
    )
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        return m.handleKeyPress(msg)

    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
        return m, nil

    case StreamEventMsg:
        return m.handleStreamEvent(msg)

    case ErrorMsg:
        return m.handleError(msg)
    }

    return m, nil
}

func (m Model) View() string {
    return m.activeView.Render(m.width, m.height)
}
```

### Chat View Component

```go
type ChatView struct {
    viewport    viewport.Model
    input       textinput.Model
    messages    []RenderedMessage
    streaming   bool
    currentResp strings.Builder
}

func (v *ChatView) Render(width, height int) string {
    // Render messages in viewport
    content := v.renderMessages()
    v.viewport.SetContent(content)

    // Render input box
    inputBox := v.renderInput()

    // Combine with layout
    return lipgloss.JoinVertical(
        lipgloss.Left,
        v.viewport.View(),
        inputBox,
    )
}

func (v *ChatView) renderMessages() string {
    var rendered []string

    for _, msg := range v.messages {
        style := v.styleForRole(msg.Role)
        rendered = append(rendered, style.Render(msg.Content))
    }

    // Add current streaming response
    if v.streaming {
        style := v.styleForRole("assistant")
        rendered = append(rendered,
            style.Render(v.currentResp.String()+" ▋"))
    }

    return strings.Join(rendered, "\n\n")
}
```

## References

- [System Overview](system-overview.md)
- [Database Guide](../database-guide.md)
- [Configuration Guide](../configuration.md)
- [Development Guide](../development/README.md)

---

**Document Version**: 1.0
**Last Updated**: January 2025
