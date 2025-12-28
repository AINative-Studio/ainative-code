// Package tools provides an extensible tool execution framework for LLM assistants.
package tools

import (
	"context"
	"time"
)

// Category represents the category of a tool.
type Category string

const (
	CategoryFilesystem Category = "filesystem"
	CategoryNetwork    Category = "network"
	CategorySystem     Category = "system"
	CategoryDatabase   Category = "database"
	CategoryText       Category = "text"
)

// Tool defines the interface that all tools must implement.
type Tool interface {
	// Name returns the unique name of the tool.
	Name() string

	// Description returns a human-readable description of what the tool does.
	Description() string

	// Schema returns the JSON schema for the tool's input parameters.
	Schema() ToolSchema

	// Execute runs the tool with the provided input and returns the result.
	Execute(ctx context.Context, input map[string]interface{}) (*Result, error)

	// Category returns the category this tool belongs to.
	Category() Category

	// RequiresConfirmation returns true if this tool requires user confirmation before execution.
	RequiresConfirmation() bool
}

// ToolSchema defines the JSON schema for tool input validation.
type ToolSchema struct {
	Type       string                 `json:"type"`
	Properties map[string]PropertyDef `json:"properties"`
	Required   []string               `json:"required,omitempty"`
}

// PropertyDef defines a property in the tool schema.
type PropertyDef struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"`
	Default     interface{} `json:"default,omitempty"`
	MinLength   *int     `json:"minLength,omitempty"`
	MaxLength   *int     `json:"maxLength,omitempty"`
	Pattern     string   `json:"pattern,omitempty"`
}

// Result represents the result of a tool execution.
type Result struct {
	Success  bool                   `json:"success"`
	Output   string                 `json:"output"`
	Error    error                  `json:"error,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ExecutionContext contains settings for tool execution.
type ExecutionContext struct {
	Timeout          time.Duration
	WorkingDirectory string
	Environment      map[string]string
	AllowedPaths     []string // For sandboxing file operations
	MaxOutputSize    int64    // Maximum output size in bytes
	DryRun           bool     // If true, don't actually execute, just validate
}

// ExecutionOption is a function that configures an ExecutionContext.
type ExecutionOption func(*ExecutionContext)

// NewExecutionContext creates a new ExecutionContext with default values.
func NewExecutionContext(opts ...ExecutionOption) *ExecutionContext {
	ctx := &ExecutionContext{
		Timeout:       30 * time.Second,
		Environment:   make(map[string]string),
		AllowedPaths:  []string{},
		MaxOutputSize: 10 * 1024 * 1024, // 10MB default
		DryRun:        false,
	}

	for _, opt := range opts {
		opt(ctx)
	}

	return ctx
}

// WithTimeout sets the execution timeout.
func WithTimeout(timeout time.Duration) ExecutionOption {
	return func(ctx *ExecutionContext) {
		ctx.Timeout = timeout
	}
}

// WithWorkingDirectory sets the working directory for execution.
func WithWorkingDirectory(dir string) ExecutionOption {
	return func(ctx *ExecutionContext) {
		ctx.WorkingDirectory = dir
	}
}

// WithEnvironment sets environment variables for execution.
func WithEnvironment(env map[string]string) ExecutionOption {
	return func(ctx *ExecutionContext) {
		ctx.Environment = env
	}
}

// WithAllowedPaths sets the allowed paths for file operations (sandboxing).
func WithAllowedPaths(paths []string) ExecutionOption {
	return func(ctx *ExecutionContext) {
		ctx.AllowedPaths = paths
	}
}

// WithMaxOutputSize sets the maximum output size in bytes.
func WithMaxOutputSize(size int64) ExecutionOption {
	return func(ctx *ExecutionContext) {
		ctx.MaxOutputSize = size
	}
}

// WithDryRun enables or disables dry-run mode.
func WithDryRun(dryRun bool) ExecutionOption {
	return func(ctx *ExecutionContext) {
		ctx.DryRun = dryRun
	}
}
