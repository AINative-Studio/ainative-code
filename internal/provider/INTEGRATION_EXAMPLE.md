# Provider Selector Integration Example

## Quick Start Integration

This example shows how to integrate the provider selector into the AINative CLI.

## Basic CLI Integration

```go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/AINative-studio/ainative-code/internal/backend"
    "github.com/AINative-studio/ainative-code/internal/provider"
)

func main() {
    // 1. Get user configuration (from config file or API)
    user := &provider.User{
        Email:   "user@example.com",
        Credits: 150,
        Tier:    "pro",
    }

    // 2. Create provider selector with user preferences
    selector := provider.NewSelector(
        provider.WithProviders("anthropic", "openai", "google"),
        provider.WithUserPreference("anthropic"),
        provider.WithCreditThreshold(50),
    )

    // 3. Define request requirements (based on CLI flags or prompt analysis)
    req := &provider.SelectionRequest{
        RequiresVision:          false,  // Set to true if images attached
        RequiresFunctionCalling: false,  // Set to true if tools needed
        RequiresStreaming:       true,   // Set based on --stream flag
        Model:                   "auto",
    }

    // 4. Select optimal provider
    selectedProvider, err := selector.Select(context.Background(), user, req)
    if err != nil {
        handleProviderError(err)
        return
    }

    // 5. Display credit warning if needed
    if selectedProvider.LowCreditWarning {
        fmt.Fprintf(os.Stderr, "⚠️  Warning: Low credits remaining (%d). Consider upgrading.\n", user.Credits)
    }

    // 6. Use selected provider with backend client
    client := backend.NewClient("http://localhost:8000")

    chatReq := &backend.ChatRequest{
        Provider: selectedProvider.Name,
        Messages: []backend.Message{
            {Role: "user", Content: "Hello, world!"},
        },
        Stream: req.RequiresStreaming,
    }

    // 7. Send request
    response, err := client.Chat(context.Background(), chatReq)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        return
    }

    // 8. Display response
    fmt.Println(response.Content)

    // 9. Show provider used
    fmt.Fprintf(os.Stderr, "\n[Provider: %s, Tokens: %d]\n",
        selectedProvider.DisplayName,
        response.Usage.TotalTokens)
}

func handleProviderError(err error) {
    switch {
    case errors.Is(err, provider.ErrInsufficientCredits):
        fmt.Fprintln(os.Stderr, "❌ Error: Insufficient credits.")
        fmt.Fprintln(os.Stderr, "   Visit https://ainative.studio/upgrade to add credits.")
        os.Exit(1)

    case errors.Is(err, provider.ErrNoProviderAvailable):
        fmt.Fprintln(os.Stderr, "❌ Error: No provider available that meets your requirements.")
        fmt.Fprintln(os.Stderr, "   Please adjust your request or configuration.")
        os.Exit(1)

    default:
        fmt.Fprintf(os.Stderr, "❌ Error selecting provider: %v\n", err)
        os.Exit(1)
    }
}
```

## CLI Command Example

```go
package cmd

import (
    "context"
    "fmt"

    "github.com/spf13/cobra"
    "github.com/AINative-studio/ainative-code/internal/config"
    "github.com/AINative-studio/ainative-code/internal/provider"
)

var chatCmd = &cobra.Command{
    Use:   "chat",
    Short: "Chat with AI",
    RunE: func(cmd *cobra.Command, args []string) error {
        // Load user config
        cfg, err := config.Load()
        if err != nil {
            return err
        }

        // Get flags
        preferredProvider, _ := cmd.Flags().GetString("provider")
        requiresVision, _ := cmd.Flags().GetBool("vision")
        requiresStreaming, _ := cmd.Flags().GetBool("stream")

        // Create selector
        selector := provider.NewSelector(
            provider.WithProviders("anthropic", "openai", "google"),
            provider.WithUserPreference(preferredProvider),
            provider.WithCreditThreshold(50),
        )

        // Build selection request
        req := &provider.SelectionRequest{
            RequiresVision:    requiresVision,
            RequiresStreaming: requiresStreaming,
            Model:             "auto",
        }

        // Select provider
        selectedProvider, err := selector.Select(
            context.Background(),
            &provider.User{
                Email:   cfg.User.Email,
                Credits: cfg.User.Credits,
                Tier:    cfg.User.Tier,
            },
            req,
        )
        if err != nil {
            return fmt.Errorf("provider selection failed: %w", err)
        }

        // Check credit warning
        if selectedProvider.LowCreditWarning {
            fmt.Printf("⚠️  Low credits: %d remaining\n", cfg.User.Credits)
        }

        // Continue with chat...
        fmt.Printf("Using provider: %s\n", selectedProvider.DisplayName)

        return nil
    },
}

func init() {
    chatCmd.Flags().String("provider", "", "Preferred provider (anthropic, openai, google)")
    chatCmd.Flags().Bool("vision", false, "Enable vision capabilities")
    chatCmd.Flags().Bool("stream", true, "Stream responses")
}
```

## Advanced: Dynamic Provider Selection

```go
package chatservice

import (
    "context"
    "strings"

    "github.com/AINative-studio/ainative-code/internal/provider"
)

// ChatService handles chat interactions with intelligent provider selection
type ChatService struct {
    selector *provider.Selector
    user     *provider.User
}

// NewChatService creates a new chat service
func NewChatService(user *provider.User, preferredProvider string) *ChatService {
    return &ChatService{
        selector: provider.NewSelector(
            provider.WithProviders("anthropic", "openai", "google"),
            provider.WithUserPreference(preferredProvider),
            provider.WithCreditThreshold(50),
        ),
        user: user,
    }
}

// SendMessage sends a message and automatically selects the best provider
func (s *ChatService) SendMessage(ctx context.Context, message string, images []string) (string, error) {
    // Analyze message to determine requirements
    req := s.analyzeRequirements(message, images)

    // Select provider
    selectedProvider, err := s.selector.Select(ctx, s.user, req)
    if err != nil {
        return "", err
    }

    // Log provider selection
    fmt.Printf("Selected: %s (Vision: %v, Functions: %v)\n",
        selectedProvider.Name,
        selectedProvider.SupportsVision,
        selectedProvider.SupportsFunctionCalling,
    )

    // Send to backend (implementation details omitted)
    return s.sendToBackend(ctx, selectedProvider, message, images)
}

// analyzeRequirements analyzes the message to determine capability requirements
func (s *ChatService) analyzeRequirements(message string, images []string) *provider.SelectionRequest {
    req := &provider.SelectionRequest{
        Model: "auto",
    }

    // Check for vision requirements
    req.RequiresVision = len(images) > 0

    // Check for function calling requirements
    // (simple keyword detection - could be more sophisticated)
    keywords := []string{"calculate", "search", "tool", "function", "execute"}
    for _, keyword := range keywords {
        if strings.Contains(strings.ToLower(message), keyword) {
            req.RequiresFunctionCalling = true
            break
        }
    }

    // Default to streaming for better UX
    req.RequiresStreaming = true

    return req
}

func (s *ChatService) sendToBackend(ctx context.Context, prov *provider.ProviderInfo, msg string, imgs []string) (string, error) {
    // Implementation would use the backend client
    return "", nil
}
```

## Testing Integration

```go
package chatservice_test

import (
    "context"
    "testing"

    "github.com/AINative-studio/ainative-code/internal/provider"
)

func TestChatService_ProviderSelection(t *testing.T) {
    tests := []struct {
        name            string
        user            *provider.User
        message         string
        images          []string
        expectedVision  bool
        expectLowCredit bool
    }{
        {
            name: "low credits should warn",
            user: &provider.User{
                Email:   "test@example.com",
                Credits: 10,
                Tier:    "free",
            },
            message:         "Hello",
            images:          nil,
            expectedVision:  false,
            expectLowCredit: true,
        },
        {
            name: "images should require vision",
            user: &provider.User{
                Email:   "test@example.com",
                Credits: 100,
                Tier:    "pro",
            },
            message:         "What's in this image?",
            images:          []string{"image.jpg"},
            expectedVision:  true,
            expectLowCredit: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            service := NewChatService(tt.user, "anthropic")

            // Test would verify provider selection logic
            // (simplified for example)
        })
    }
}
```

## Configuration File Support

Example configuration file (`~/.ainative/config.yaml`):

```yaml
user:
  email: user@example.com
  credits: 150
  tier: pro
  preferred_provider: anthropic

providers:
  available:
    - anthropic
    - openai
    - google

  settings:
    credit_threshold: 50
    fallback_enabled: true

capabilities:
  vision: true
  function_calling: true
  streaming: true
```

Loading configuration:

```go
package config

import (
    "os"
    "path/filepath"

    "gopkg.in/yaml.v3"
    "github.com/AINative-studio/ainative-code/internal/provider"
)

type Config struct {
    User      UserConfig      `yaml:"user"`
    Providers ProviderConfig  `yaml:"providers"`
}

type UserConfig struct {
    Email             string `yaml:"email"`
    Credits           int    `yaml:"credits"`
    Tier              string `yaml:"tier"`
    PreferredProvider string `yaml:"preferred_provider"`
}

type ProviderConfig struct {
    Available []string        `yaml:"available"`
    Settings  ProviderSettings `yaml:"settings"`
}

type ProviderSettings struct {
    CreditThreshold int  `yaml:"credit_threshold"`
    FallbackEnabled bool `yaml:"fallback_enabled"`
}

func Load() (*Config, error) {
    home, err := os.UserHomeDir()
    if err != nil {
        return nil, err
    }

    path := filepath.Join(home, ".ainative", "config.yaml")
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }

    var cfg Config
    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return nil, err
    }

    return &cfg, nil
}

// CreateSelector creates a provider selector from config
func (c *Config) CreateSelector() *provider.Selector {
    return provider.NewSelector(
        provider.WithProviders(c.Providers.Available...),
        provider.WithUserPreference(c.User.PreferredProvider),
        provider.WithCreditThreshold(c.Providers.Settings.CreditThreshold),
        provider.WithFallback(c.Providers.Settings.FallbackEnabled),
    )
}

// GetUser returns a provider.User from config
func (c *Config) GetUser() *provider.User {
    return &provider.User{
        Email:   c.User.Email,
        Credits: c.User.Credits,
        Tier:    c.User.Tier,
    }
}
```

## Usage in CLI

```go
package main

import (
    "context"
    "fmt"

    "github.com/AINative-studio/ainative-code/internal/config"
)

func main() {
    // Load configuration
    cfg, err := config.Load()
    if err != nil {
        fmt.Printf("Error loading config: %v\n", err)
        return
    }

    // Create selector from config
    selector := cfg.CreateSelector()

    // Get user from config
    user := cfg.GetUser()

    // Use selector
    provider, err := selector.Select(context.Background(), user)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }

    fmt.Printf("Selected provider: %s\n", provider.DisplayName)
}
```

## Summary

This integration example demonstrates:

1. **Basic CLI Integration**: Simple provider selection in CLI commands
2. **Advanced Service Layer**: Automatic requirement detection
3. **Configuration Management**: YAML-based configuration with selector factory
4. **Error Handling**: Proper error handling for all failure modes
5. **Testing**: Integration testing patterns

The provider selector seamlessly integrates into the AINative CLI architecture while maintaining clean separation of concerns and testability.
