package setup

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"gopkg.in/yaml.v3"

	"github.com/AINative-studio/ainative-code/internal/config"
)

// WizardConfig holds configuration for the setup wizard
type WizardConfig struct {
	ConfigPath      string
	SkipValidation  bool
	InteractiveMode bool
	Force           bool
}

// WizardResult represents the output of the wizard
type WizardResult struct {
	Config         *config.Config
	ConfigPath     string
	MarkerCreated  bool
	SkippedSetup   bool
	ValidationPass bool
}

// Wizard orchestrates the first-time setup process
type Wizard struct {
	ctx            context.Context
	config         WizardConfig
	result         *WizardResult
	userSelections map[string]interface{}
}

// NewWizard creates a new setup wizard instance
func NewWizard(ctx context.Context, cfg WizardConfig) *Wizard {
	return &Wizard{
		ctx:            ctx,
		config:         cfg,
		userSelections: make(map[string]interface{}),
		result: &WizardResult{
			Config: &config.Config{},
		},
	}
}

// Run executes the setup wizard flow
func (w *Wizard) Run() (*WizardResult, error) {
	// Check if already initialized (skip if force flag is set)
	if !w.config.Force && w.checkAlreadyInitialized() {
		return w.result, nil
	}

	// Welcome screen
	w.showWelcome()

	// Run interactive prompts
	if w.config.InteractiveMode {
		if err := w.runInteractiveSetup(); err != nil {
			return nil, fmt.Errorf("interactive setup failed: %w", err)
		}
	}

	// Build configuration from selections
	if err := w.buildConfiguration(); err != nil {
		return nil, fmt.Errorf("failed to build configuration: %w", err)
	}

	// Validate configuration
	if !w.config.SkipValidation {
		if err := w.validateConfiguration(); err != nil {
			return nil, fmt.Errorf("configuration validation failed: %w", err)
		}
		w.result.ValidationPass = true
	}

	// Show summary
	if w.config.InteractiveMode {
		confirmed, err := w.showSummary()
		if err != nil {
			return nil, err
		}
		if !confirmed {
			fmt.Println("\nSetup cancelled.")
			return nil, fmt.Errorf("setup cancelled by user")
		}
	}

	// Write configuration file
	if err := w.writeConfiguration(); err != nil {
		return nil, fmt.Errorf("failed to write configuration: %w", err)
	}

	// Create initialization marker
	if err := w.createMarker(); err != nil {
		return nil, fmt.Errorf("failed to create initialization marker: %w", err)
	}

	// Show success message
	w.showSuccess()

	return w.result, nil
}

// checkAlreadyInitialized checks if the CLI has already been initialized
func (w *Wizard) checkAlreadyInitialized() bool {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	markerPath := filepath.Join(homeDir, ".ainative-code-initialized")
	if _, err := os.Stat(markerPath); err == nil {
		// Already initialized
		configPath := filepath.Join(homeDir, ".ainative-code.yaml")
		if _, err := os.Stat(configPath); err == nil {
			w.result.ConfigPath = configPath
			w.result.SkippedSetup = true
			return true
		}
	}

	return false
}

// showWelcome displays the welcome message
func (w *Wizard) showWelcome() {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")). // Blue
		MarginTop(1).
		MarginBottom(1)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("250")). // Light gray
		MarginBottom(1)

	fmt.Println(titleStyle.Render("Welcome to AINative Code!"))
	fmt.Println(descStyle.Render("Let's set up your AI-powered development environment."))
	fmt.Println(descStyle.Render("This wizard will guide you through the configuration process."))
	fmt.Println()
}

// runInteractiveSetup runs the interactive prompt flow
func (w *Wizard) runInteractiveSetup() error {
	model := NewPromptModel()

	// Run bubble tea program
	p := tea.NewProgram(model, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("failed to run interactive setup: %w", err)
	}

	// Extract selections from final model
	if m, ok := finalModel.(PromptModel); ok {
		w.userSelections = m.Selections
	}

	return nil
}

// buildConfiguration constructs the config object from user selections
func (w *Wizard) buildConfiguration() error {
	cfg := &config.Config{
		App: config.AppConfig{
			Name:        "ainative-code",
			Version:     "0.1.0",
			Environment: "development",
			Debug:       false,
		},
		LLM: config.LLMConfig{},
		Platform: config.PlatformConfig{
			Authentication: config.AuthConfig{
				Method:  "none", // Default to none, will be set to api_key if user configures AINative
				Timeout: 10000000000, // 10s in nanoseconds
			},
		},
		Performance: config.PerformanceConfig{
			Cache: config.CacheConfig{
				Enabled: false,
				Type:    "memory",
				TTL:     3600000000000, // 1h
				MaxSize: 100,
			},
			Concurrency: config.ConcurrencyConfig{
				MaxWorkers:    10,
				MaxQueueSize:  100,
				WorkerTimeout: 300000000000, // 5m
			},
		},
		Logging: config.LoggingConfig{
			Level:      "info",
			Format:     "json",
			Output:     "stdout",
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     7,
			Compress:   true,
		},
		Security: config.SecurityConfig{
			EncryptConfig: false,
			TLSEnabled:    false,
		},
	}

	// Set default provider
	if provider, ok := w.userSelections["provider"].(string); ok {
		cfg.LLM.DefaultProvider = provider
	} else {
		cfg.LLM.DefaultProvider = "anthropic"
	}

	// Configure Anthropic
	if cfg.LLM.DefaultProvider == "anthropic" {
		apiKey := ""
		if key, ok := w.userSelections["anthropic_api_key"].(string); ok {
			apiKey = key
		}

		model := "claude-3-5-sonnet-20241022"
		if m, ok := w.userSelections["anthropic_model"].(string); ok && m != "" {
			model = m
		}

		extendedThinking := false
		if et, ok := w.userSelections["extended_thinking"].(bool); ok {
			extendedThinking = et
		}

		cfg.LLM.Anthropic = &config.AnthropicConfig{
			APIKey:        apiKey,
			Model:         model,
			MaxTokens:     4096,
			Temperature:   0.7,
			TopP:          1.0,
			TopK:          0,
			Timeout:       30000000000, // 30s
			RetryAttempts: 3,
			APIVersion:    "2023-06-01",
		}

		if extendedThinking {
			cfg.LLM.Anthropic.ExtendedThinking = &config.ExtendedThinkingConfig{
				Enabled:    true,
				AutoExpand: false,
				MaxDepth:   5,
			}
		}
	}

	// Configure OpenAI
	if cfg.LLM.DefaultProvider == "openai" {
		apiKey := ""
		if key, ok := w.userSelections["openai_api_key"].(string); ok {
			apiKey = key
		}

		model := "gpt-4-turbo-preview"
		if m, ok := w.userSelections["openai_model"].(string); ok && m != "" {
			model = m
		}

		cfg.LLM.OpenAI = &config.OpenAIConfig{
			APIKey:           apiKey,
			Model:            model,
			MaxTokens:        4096,
			Temperature:      0.7,
			TopP:             1.0,
			FrequencyPenalty: 0.0,
			PresencePenalty:  0.0,
			Timeout:          30000000000, // 30s
			RetryAttempts:    3,
		}
	}

	// Configure Google Gemini
	if cfg.LLM.DefaultProvider == "google" {
		apiKey := ""
		if key, ok := w.userSelections["google_api_key"].(string); ok {
			apiKey = key
		}

		model := "gemini-pro"
		if m, ok := w.userSelections["google_model"].(string); ok && m != "" {
			model = m
		}

		cfg.LLM.Google = &config.GoogleConfig{
			APIKey:        apiKey,
			Model:         model,
			MaxTokens:     4096,
			Temperature:   0.7,
			TopP:          1.0,
			TopK:          40,
			Timeout:       30000000000, // 30s
			RetryAttempts: 3,
		}
	}

	// Configure Ollama
	if cfg.LLM.DefaultProvider == "ollama" {
		baseURL := "http://localhost:11434"
		if url, ok := w.userSelections["ollama_url"].(string); ok && url != "" {
			baseURL = url
		}

		model := "llama2"
		if m, ok := w.userSelections["ollama_model"].(string); ok && m != "" {
			model = m
		}

		cfg.LLM.Ollama = &config.OllamaConfig{
			BaseURL:       baseURL,
			Model:         model,
			MaxTokens:     4096,
			Temperature:   0.7,
			TopP:          1.0,
			TopK:          40,
			Timeout:       120000000000, // 120s
			RetryAttempts: 1,
			KeepAlive:     "5m",
		}
	}

	// Configure Meta Llama
	if cfg.LLM.DefaultProvider == "meta_llama" {
		apiKey := ""
		if key, ok := w.userSelections["meta_llama_api_key"].(string); ok {
			apiKey = key
		}

		model := "Llama-4-Maverick-17B-128E-Instruct-FP8"
		if m, ok := w.userSelections["meta_llama_model"].(string); ok && m != "" {
			model = m
		}

		cfg.LLM.MetaLlama = &config.MetaLlamaConfig{
			APIKey:           apiKey,
			Model:            model,
			MaxTokens:        4096,
			Temperature:      0.7,
			TopP:             0.9,
			Timeout:          60000000000, // 60s
			RetryAttempts:    3,
			BaseURL:          "https://api.llama.com/compat/v1",
			PresencePenalty:  0.0,
			FrequencyPenalty: 0.0,
		}
	}

	// AINative platform login (optional)
	if loginEnabled, ok := w.userSelections["ainative_login"].(bool); ok && loginEnabled {
		if apiKey, ok := w.userSelections["ainative_api_key"].(string); ok {
			cfg.Platform.Authentication.Method = "api_key"
			cfg.Platform.Authentication.APIKey = apiKey
		}
	}

	// Configure Strapi (optional)
	if strapiEnabled, ok := w.userSelections["strapi_enabled"].(bool); ok && strapiEnabled {
		strapiURL := ""
		if url, ok := w.userSelections["strapi_url"].(string); ok {
			strapiURL = url
		}

		strapiAPIKey := ""
		if key, ok := w.userSelections["strapi_api_key"].(string); ok {
			strapiAPIKey = key
		}

		cfg.Services.Strapi = &config.StrapiConfig{
			Enabled:       true,
			Endpoint:      strapiURL,
			APIKey:        strapiAPIKey,
			Timeout:       30000000000, // 30s
			RetryAttempts: 3,
		}
	}

	// Configure ZeroDB (optional)
	if zeroDBEnabled, ok := w.userSelections["zerodb_enabled"].(bool); ok && zeroDBEnabled {
		projectID := ""
		if pid, ok := w.userSelections["zerodb_project_id"].(string); ok {
			projectID = pid
		}

		endpoint := ""
		if ep, ok := w.userSelections["zerodb_endpoint"].(string); ok && ep != "" {
			endpoint = ep
		}

		cfg.Services.ZeroDB = &config.ZeroDBConfig{
			Enabled:         true,
			ProjectID:       projectID,
			Endpoint:        endpoint,
			Database:        "default",
			SSL:             true,
			SSLMode:         "require",
			MaxConnections:  10,
			IdleConnections: 5,
			ConnMaxLifetime: 3600000000000, // 1h
			Timeout:         30000000000,   // 30s
			RetryAttempts:   3,
			RetryDelay:      1000000000, // 1s
		}
	}

	w.result.Config = cfg
	return nil
}

// validateConfiguration validates the built configuration
func (w *Wizard) validateConfiguration() error {
	// First validate using config validator
	configValidator := config.NewValidator(w.result.Config)
	if err := configValidator.Validate(); err != nil {
		return err
	}

	// Then validate provider-specific credentials if not skipped
	if !w.config.SkipValidation {
		validator := NewValidator()
		provider := w.userSelections["provider"].(string)
		return validator.ValidateProviderConfig(w.ctx, provider, w.userSelections)
	}

	return nil
}

// showSummary displays configuration summary and asks for confirmation
func (w *Wizard) showSummary() (bool, error) {
	model := NewSummaryModel(w.result.Config, w.userSelections)
	p := tea.NewProgram(model)

	finalModel, err := p.Run()
	if err != nil {
		return false, fmt.Errorf("failed to show summary: %w", err)
	}

	if m, ok := finalModel.(SummaryModel); ok {
		return m.Confirmed, nil
	}

	return false, nil
}

// writeConfiguration writes the configuration to file
func (w *Wizard) writeConfiguration() error {
	// Determine config path
	configPath := w.config.ConfigPath
	if configPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		configPath = filepath.Join(homeDir, ".ainative-code.yaml")
	}

	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal configuration to YAML
	data, err := yaml.Marshal(w.result.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	w.result.ConfigPath = configPath
	return nil
}

// createMarker creates the initialization marker file
func (w *Wizard) createMarker() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	markerPath := filepath.Join(homeDir, ".ainative-code-initialized")
	markerContent := fmt.Sprintf("initialized_at: %s\n", strings.ReplaceAll(strings.ReplaceAll(os.Getenv(""), "\n", ""), "\r", ""))

	if err := os.WriteFile(markerPath, []byte(markerContent), 0644); err != nil {
		return fmt.Errorf("failed to create marker file: %w", err)
	}

	w.result.MarkerCreated = true
	return nil
}

// showSuccess displays the success message with next steps
func (w *Wizard) showSuccess() {
	successStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("46")). // Green
		MarginTop(1).
		MarginBottom(1)

	infoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("250"))

	pathStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")). // Blue
		Bold(true)

	fmt.Println(successStyle.Render("Setup Complete!"))
	fmt.Println()
	fmt.Println(infoStyle.Render("Your configuration has been saved to:"))
	fmt.Println(pathStyle.Render("  " + w.result.ConfigPath))
	fmt.Println()
	fmt.Println(infoStyle.Render("Next steps:"))
	fmt.Println(infoStyle.Render("  1. Start a chat session: ainative-code chat"))
	fmt.Println(infoStyle.Render("  2. View configuration: ainative-code config show"))
	fmt.Println(infoStyle.Render("  3. Check version: ainative-code version"))
	fmt.Println()
	fmt.Println(infoStyle.Render("For help, run: ainative-code --help"))
	fmt.Println()
}

// CheckFirstRun checks if this is a first run and whether setup is needed
func CheckFirstRun() bool {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	markerPath := filepath.Join(homeDir, ".ainative-code-initialized")
	if _, err := os.Stat(markerPath); err != nil {
		// Marker doesn't exist, this is first run
		return true
	}

	return false
}

// SetSelections manually sets user selections (for non-interactive mode)
func (w *Wizard) SetSelections(selections map[string]interface{}) {
	w.userSelections = selections
}
