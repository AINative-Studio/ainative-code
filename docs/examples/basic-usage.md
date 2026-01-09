# Basic Usage Examples

This guide provides practical examples of using AINative Code for common tasks.

## Getting Started

### First Conversation

```bash
# Start a new chat session
ainative-code chat

# One-shot question
ainative-code chat "What is the difference between interface{} and any in Go?"

# Resume last session
ainative-code chat --resume
```

### Configuration

```bash
# Initialize configuration
ainative-code setup

# Set default provider
ainative-code config set llm.default_provider anthropic

# Set API key
export ANTHROPIC_API_KEY="your-api-key"
ainative-code config set llm.anthropic.api_key "${ANTHROPIC_API_KEY}"

# View current configuration
ainative-code config show

# Edit configuration file
ainative-code config edit
```

## Code Generation

### Generate a REST API Handler

```bash
ainative-code chat "Create a REST API handler in Go for user authentication with:
- POST /api/auth/login endpoint
- Email and password validation
- JWT token generation
- Error handling
- Unit tests"
```

**Example Response**:

```go
package auth

import (
    "encoding/json"
    "net/http"
    "time"

    "github.com/golang-jwt/jwt/v5"
)

type LoginRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

type LoginResponse struct {
    Token     string    `json:"token"`
    ExpiresAt time.Time `json:"expires_at"`
}

func HandleLogin(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var req LoginRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Validate email
    if !isValidEmail(req.Email) {
        http.Error(w, "Invalid email", http.StatusBadRequest)
        return
    }

    // Authenticate user
    user, err := authenticateUser(req.Email, req.Password)
    if err != nil {
        http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        return
    }

    // Generate JWT token
    token, expiresAt, err := generateJWT(user.ID)
    if err != nil {
        http.Error(w, "Failed to generate token", http.StatusInternalServerError)
        return
    }

    // Send response
    resp := LoginResponse{
        Token:     token,
        ExpiresAt: expiresAt,
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp)
}

func generateJWT(userID string) (string, time.Time, error) {
    expiresAt := time.Now().Add(24 * time.Hour)

    claims := jwt.MapClaims{
        "sub": userID,
        "exp": expiresAt.Unix(),
        "iat": time.Now().Unix(),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString([]byte("your-secret-key"))

    return tokenString, expiresAt, err
}
```

### Generate Database Schema

```bash
ainative-code chat "Create a PostgreSQL schema for an e-commerce application with:
- Users table
- Products table
- Orders table
- Order items table
- Proper foreign keys and indexes"
```

### Generate Configuration File

```bash
ainative-code chat "Create a comprehensive config.yaml for a microservice with:
- Server settings (host, port, timeout)
- Database connection (PostgreSQL)
- Redis cache settings
- Logging configuration
- Environment-specific overrides"
```

## Code Review and Improvement

### Review Code for Security

```bash
ainative-code chat "Review this code for security issues:

\`\`\`go
func GetUser(w http.ResponseWriter, r *http.Request) {
    id := r.URL.Query().Get(\"id\")
    query := \"SELECT * FROM users WHERE id = \" + id
    rows, _ := db.Query(query)
    // ... rest of code
}
\`\`\`

What security vulnerabilities exist and how should I fix them?"
```

### Optimize Performance

```bash
ainative-code chat "Optimize this function for performance:

\`\`\`go
func ProcessItems(items []Item) []Result {
    var results []Result
    for _, item := range items {
        result := expensiveOperation(item)
        results = append(results, result)
    }
    return results
}
\`\`\`"
```

### Add Error Handling

```bash
ainative-code chat "Add proper error handling to this code:

\`\`\`go
func SaveUser(user User) {
    data, _ := json.Marshal(user)
    os.WriteFile(\"users.json\", data, 0644)
}
\`\`\`"
```

## Debugging

### Debug an Error

```bash
ainative-code chat "I'm getting this error:

panic: runtime error: invalid memory address or nil pointer dereference

Here's the code:

\`\`\`go
func ProcessRequest(req *Request) {
    user := getUserByID(req.UserID)
    log.Printf(\"Processing for user: %s\", user.Name)
}
\`\`\`

What's wrong and how do I fix it?"
```

### Understand Stack Trace

```bash
ainative-code chat "Explain this stack trace and help me fix it:

goroutine 1 [running]:
main.processData(0xc0001a4000, 0x5, 0x5)
    /app/main.go:45 +0x123
main.main()
    /app/main.go:20 +0x85"
```

## Learning and Documentation

### Learn a Concept

```bash
ainative-code chat "Explain Go interfaces with practical examples"

ainative-code chat "What are the differences between mutex and channels in Go?"

ainative-code chat "How does garbage collection work in Go?"
```

### Generate Documentation

```bash
ainative-code chat "Generate godoc comments for this function:

\`\`\`go
func CalculateDiscount(price float64, quantity int, customerTier string) float64 {
    baseDiscount := 0.0
    if quantity >= 10 {
        baseDiscount = 0.1
    }
    if customerTier == \"premium\" {
        baseDiscount += 0.05
    }
    return price * (1 - baseDiscount)
}
\`\`\`"
```

### Create README

```bash
ainative-code chat "Create a comprehensive README.md for my project. It's a CLI tool for managing Docker containers with features like:
- List containers
- Start/stop containers
- View logs
- Clean up unused containers
Written in Go using Cobra CLI framework"
```

## Testing

### Generate Unit Tests

```bash
ainative-code chat "Generate comprehensive unit tests for this function:

\`\`\`go
func ValidateEmail(email string) bool {
    re := regexp.MustCompile(\`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$\`)
    return re.MatchString(email)
}
\`\`\`"
```

### Generate Table-Driven Tests

```bash
ainative-code chat "Create table-driven tests for a function that converts temperature from Celsius to Fahrenheit"
```

### Generate Mocks

```bash
ainative-code chat "Create a mock implementation for this interface:

\`\`\`go
type UserRepository interface {
    GetByID(id string) (*User, error)
    Create(user *User) error
    Update(user *User) error
    Delete(id string) error
}
\`\`\`"
```

## Refactoring

### Extract Function

```bash
ainative-code chat "Refactor this code by extracting the validation logic into separate functions:

\`\`\`go
func ProcessOrder(order Order) error {
    if order.Total <= 0 {
        return errors.New(\"invalid total\")
    }
    if order.CustomerID == \"\" {
        return errors.New(\"customer ID required\")
    }
    if len(order.Items) == 0 {
        return errors.New(\"no items in order\")
    }

    // ... rest of processing
}
\`\`\`"
```

### Simplify Complex Function

```bash
ainative-code chat "Simplify this function:

\`\`\`go
func GetUserPermissions(userID string, resourceID string) ([]string, error) {
    user, err := db.GetUser(userID)
    if err != nil {
        return nil, err
    }

    resource, err := db.GetResource(resourceID)
    if err != nil {
        return nil, err
    }

    if user.IsAdmin {
        return []string{\"read\", \"write\", \"delete\", \"admin\"}, nil
    }

    if resource.OwnerID == userID {
        return []string{\"read\", \"write\", \"delete\"}, nil
    }

    // ... more conditions
}
\`\`\`"
```

## Session Management

### Create Named Session

```bash
# Create a session for a specific task
ainative-code chat --new --title "OAuth Implementation"
```

### List Sessions

```bash
# View all sessions
ainative-code session list

# Search sessions
ainative-code session list --filter "oauth"

# View recent sessions
ainative-code session list --limit 5
```

### Resume Session

```bash
# Resume specific session
ainative-code chat --session-id abc123

# Resume last session
ainative-code chat --resume
```

### Export Session

```bash
# Export to markdown
ainative-code session export abc123 --format markdown > session.md

# Export to JSON
ainative-code session export abc123 --format json > session.json
```

## Provider Management

### Switch Providers

```bash
# Use specific provider for one conversation
ainative-code chat --provider openai "Explain async/await"

# Set default provider
ainative-code config set llm.default_provider anthropic

# List available providers
ainative-code provider list
```

### Provider-Specific Features

```bash
# Use extended thinking (Claude)
ainative-code chat --extended-thinking "Design a distributed caching system"

# Use vision (Claude, GPT-4, Gemini)
ainative-code chat --image screenshot.png "What's wrong with this UI?"

# Use local model (Ollama)
ainative-code chat --provider ollama --model llama3 "Quick question about Go"
```

## Advanced Usage

### Multi-File Context

```bash
ainative-code chat "Review these files for consistency:

\`\`\`main.go
// paste file content
\`\`\`

\`\`\`handler.go
// paste file content
\`\`\`

Are the error handling patterns consistent?"
```

### Interactive Debugging Session

```bash
# Start debugging session
ainative-code chat --new --title "Debug Memory Leak"

# In chat:
# You: "I have a memory leak in my application"
# Assistant: "Can you share the code and memory profile?"
# You: [paste code and pprof output]
# Assistant: "I see the issue. You're not closing the HTTP response body..."
```

### Code Generation Workflow

```bash
# Step 1: Generate initial code
ainative-code chat "Create a user service with CRUD operations"

# Step 2: Add tests
ainative-code chat --resume "Now generate comprehensive unit tests"

# Step 3: Add documentation
ainative-code chat --resume "Add godoc comments and a usage example"

# Step 4: Review
ainative-code chat --resume "Review the code for any issues"
```

## Integration with Tools

### File Operations

The AI can directly read and write files:

```bash
ainative-code chat "Read the main.go file and add proper error handling"
```

The AI will:
1. Read the file
2. Analyze the code
3. Write the improved version
4. Show you the diff

### Execute Commands

The AI can run commands (with your permission):

```bash
ainative-code chat "Run the tests and analyze any failures"
```

The AI will:
1. Execute `go test ./...`
2. Analyze the output
3. Suggest fixes

### Web Research

```bash
ainative-code chat "Research the latest best practices for Go error handling in 2025"
```

The AI will:
1. Search the web
2. Analyze current best practices
3. Provide recommendations

## Tips for Effective Usage

### 1. Be Specific

**Instead of**: "Help with my code"
**Try**: "Review this authentication function for security issues and suggest improvements"

### 2. Provide Context

**Instead of**: "This doesn't work"
**Try**: "This function should return users sorted by name, but it's returning them unsorted. Here's the code: [code]"

### 3. Use Code Blocks

Always use code blocks with language specification:

````
```go
func MyFunction() {
    // Your code here
}
```
````

### 4. Iterate and Refine

Start with a broad request, then refine:
1. "Create a web server"
2. "Add authentication"
3. "Add rate limiting"
4. "Add comprehensive error handling"

### 5. Ask for Explanations

Don't just ask for code, ask for understanding:
- "Explain why this approach is better"
- "What are the trade-offs?"
- "When should I use this pattern?"

### 6. Request Tests

Always ask for tests with your code:
- "Include unit tests"
- "Generate test cases covering edge cases"
- "Create both positive and negative test cases"

## Next Steps

- Read the [Configuration Guide](../user-guide/configuration.md) for advanced configuration
- See [Provider Guide](../user-guide/providers.md) for provider-specific features
- Check [Tools Guide](../user-guide/tools.md) for advanced tool usage
- Explore [AINative Integrations](../user-guide/ainative-integrations.md) for platform features

---

**Last Updated**: January 2025
