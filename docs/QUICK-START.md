# AINative Code - Quick Start Guide

## Prerequisites

- Go 1.25.5 or later
- CGO enabled (required for SQLite3)
- Git

## Initial Setup

### 1. Clone and Setup

```bash
cd /Users/aideveloper/AINative-Code
go mod download
```

### 2. Verify Installation

```bash
./verify-deps.sh
```

Expected output:
```
=== All required dependencies are installed! ===
```

### 3. Install SQLC (if not already installed)

```bash
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

SQLC will be installed at `~/go/bin/sqlc`. Add it to your PATH:

```bash
export PATH="$HOME/go/bin:$PATH"
```

## Core Dependencies

### TUI Framework (Bubble Tea)
```go
import tea "github.com/charmbracelet/bubbletea"
```

### CLI Framework (Cobra)
```go
import "github.com/spf13/cobra"
```

### Configuration (Viper)
```go
import "github.com/spf13/viper"
```

### JWT Authentication
```go
import "github.com/golang-jwt/jwt/v5"
```

### Database (SQLite)
```go
import (
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
)
```

### HTTP Client (Resty)
```go
import "github.com/go-resty/resty/v2"
```

### Logging (Zerolog)
```go
import "github.com/rs/zerolog/log"
```

## Project Structure

```
/Users/aideveloper/AINative-Code/
├── cmd/                    # Application entry points
├── internal/              # Private application code
│   ├── api/              # API integration (Claude, etc.)
│   ├── auth/             # Authentication logic
│   ├── config/           # Configuration management
│   ├── database/         # Database models and queries
│   ├── logger/           # Logging setup
│   ├── tui/              # Terminal UI components
│   └── ui/               # UI logic
├── pkg/                   # Public libraries
├── configs/              # Configuration files
├── scripts/              # Build and deployment scripts
├── tests/                # Test files
├── docs/                 # Documentation
├── go.mod                # Go module definition
├── go.sum                # Dependency checksums
├── sqlc.yaml            # SQLC configuration
└── Makefile             # Build automation
```

## Common Commands

### Build
```bash
make build
```

### Run Tests
```bash
make test
```

### Generate SQLC Code
```bash
sqlc generate
# or
~/go/bin/sqlc generate
```

### Run Linter
```bash
make lint
```

### Format Code
```bash
make fmt
```

## Development Workflow

1. **Create a new feature branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Write tests first** (TDD approach)
   ```bash
   # Create test file in tests/ or *_test.go
   go test ./...
   ```

3. **Implement the feature**

4. **Generate database code** (if needed)
   ```bash
   sqlc generate
   ```

5. **Run tests and linter**
   ```bash
   make test
   make lint
   ```

6. **Commit your changes**
   ```bash
   git add .
   git commit -m "feat: description of your feature"
   ```

## SQLC Workflow

### 1. Define Schema
Create SQL schema files in `internal/database/schema/`:

```sql
-- schema.sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### 2. Write Queries
Create query files in `internal/database/queries/`:

```sql
-- users.sql
-- name: GetUser :one
SELECT * FROM users WHERE id = ? LIMIT 1;

-- name: ListUsers :many
SELECT * FROM users ORDER BY created_at DESC;

-- name: CreateUser :one
INSERT INTO users (email) VALUES (?) RETURNING *;
```

### 3. Generate Code
```bash
sqlc generate
```

This creates type-safe Go code in `internal/database/`.

### 4. Use Generated Code
```go
import "github.com/AINative-studio/ainative-code/internal/database"

queries := database.New(db)
user, err := queries.GetUser(ctx, userID)
```

## Configuration Management with Viper

### 1. Create config file
```yaml
# configs/config.yaml
claude:
  api_key: ${CLAUDE_API_KEY}
  model: claude-3-opus-20240229

database:
  path: ./data/ainative.db

logging:
  level: info
  file: ./logs/ainative.log
```

### 2. Load configuration
```go
viper.SetConfigName("config")
viper.SetConfigType("yaml")
viper.AddConfigPath("./configs")
viper.AutomaticEnv()

if err := viper.ReadInConfig(); err != nil {
    log.Fatal().Err(err).Msg("Failed to read config")
}
```

## HTTP Requests with Resty

```go
client := resty.New()

resp, err := client.R().
    SetHeader("Content-Type", "application/json").
    SetAuthToken("your-api-key").
    SetBody(request).
    Post("https://api.anthropic.com/v1/messages")
```

## Logging with Zerolog

```go
import (
    "github.com/rs/zerolog/log"
    "gopkg.in/natefinch/lumberjack.v2"
)

// Setup with rotation
log.Logger = log.Output(&lumberjack.Logger{
    Filename:   "./logs/ainative.log",
    MaxSize:    10, // MB
    MaxBackups: 3,
    MaxAge:     28, // days
})

// Use it
log.Info().Msg("Application started")
log.Error().Err(err).Msg("Failed to connect")
```

## TUI with Bubble Tea

```go
import tea "github.com/charmbracelet/bubbletea"

type model struct {
    // your state
}

func (m model) Init() tea.Cmd {
    return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // handle messages
    return m, nil
}

func (m model) View() string {
    // render UI
    return "Hello, World!"
}

func main() {
    p := tea.NewProgram(model{})
    if err := p.Start(); err != nil {
        log.Fatal(err)
    }
}
```

## CLI with Cobra

```go
import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
    Use:   "ainative",
    Short: "AINative Code - AI-powered coding assistant",
}

var chatCmd = &cobra.Command{
    Use:   "chat",
    Short: "Start interactive chat",
    Run: func(cmd *cobra.Command, args []string) {
        // implementation
    },
}

func init() {
    rootCmd.AddCommand(chatCmd)
}
```

## Troubleshooting

### CGO errors with SQLite
Ensure you have a C compiler installed:
```bash
# macOS
xcode-select --install

# Linux
sudo apt-get install build-essential
```

### SQLC command not found
Add Go bin to PATH:
```bash
export PATH="$HOME/go/bin:$PATH"
```

Or use full path:
```bash
~/go/bin/sqlc generate
```

### Module errors
```bash
go mod tidy
go mod verify
```

## Resources

- **Bubble Tea**: https://github.com/charmbracelet/bubbletea
- **Cobra**: https://github.com/spf13/cobra
- **Viper**: https://github.com/spf13/viper
- **SQLC**: https://docs.sqlc.dev/
- **Zerolog**: https://github.com/rs/zerolog
- **Resty**: https://github.com/go-resty/resty

## Getting Help

- Check DEPENDENCIES.md for detailed dependency information
- Review TASK-004-SUMMARY.md for installation details
- Run `./verify-deps.sh` to check your setup
- Consult the PRD.md for project requirements
- Review backlog.md for planned features

## Next Steps

1. Implement database schema (see TASK-005)
2. Set up authentication (see TASK-006)
3. Build CLI structure (see TASK-007)
4. Create TUI components (see TASK-008)
5. Integrate Claude API (see TASK-010)
