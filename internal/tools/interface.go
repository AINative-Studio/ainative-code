package tools

import (
	"context"
)

// Tool defines the interface that all tools must implement
type Tool interface {
	// Name returns the unique identifier for the tool
	Name() string

	// Description returns a human-readable description of the tool's purpose
	Description() string

	// Schema returns the JSON schema defining the tool's input parameters
	Schema() ToolSchema

	// Execute runs the tool with the given input and returns the result
	Execute(ctx context.Context, input map[string]interface{}) (string, error)
}

// ToolSchema defines the structure for tool input validation
type ToolSchema struct {
	// Type is the JSON schema type (typically "object")
	Type string `json:"type"`

	// Properties defines the input parameters
	Properties map[string]PropertySchema `json:"properties"`

	// Required lists the required parameter names
	Required []string `json:"required,omitempty"`

	// Description provides additional schema documentation
	Description string `json:"description,omitempty"`
}

// PropertySchema defines a single input parameter
type PropertySchema struct {
	// Type is the JSON schema type (string, number, boolean, array, object)
	Type string `json:"type"`

	// Description explains the parameter's purpose
	Description string `json:"description,omitempty"`

	// Enum lists allowed values for the parameter
	Enum []interface{} `json:"enum,omitempty"`

	// Default provides a default value if not specified
	Default interface{} `json:"default,omitempty"`

	// Format specifies the data format (e.g., "uri", "email", "date-time")
	Format string `json:"format,omitempty"`

	// Items defines the schema for array items
	Items *PropertySchema `json:"items,omitempty"`

	// Properties defines nested object properties
	Properties map[string]PropertySchema `json:"properties,omitempty"`

	// Required lists required properties for nested objects
	Required []string `json:"required,omitempty"`

	// Minimum defines the minimum value for numbers
	Minimum *float64 `json:"minimum,omitempty"`

	// Maximum defines the maximum value for numbers
	Maximum *float64 `json:"maximum,omitempty"`

	// MinLength defines the minimum string length
	MinLength *int `json:"minLength,omitempty"`

	// MaxLength defines the maximum string length
	MaxLength *int `json:"maxLength,omitempty"`
}

// ToolResult represents the outcome of a tool execution
type ToolResult struct {
	// Success indicates whether the tool executed successfully
	Success bool `json:"success"`

	// Output contains the tool's result data
	Output string `json:"output"`

	// Error contains error information if the tool failed
	Error string `json:"error,omitempty"`

	// Metadata contains additional execution information
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ToolExecutionOptions provides configuration for tool execution
type ToolExecutionOptions struct {
	// Timeout specifies the maximum execution duration
	Timeout int64 `json:"timeout,omitempty"`

	// RequireConfirmation indicates if user confirmation is needed before execution
	RequireConfirmation bool `json:"require_confirmation,omitempty"`

	// Sandbox enables sandboxed execution for security
	Sandbox bool `json:"sandbox,omitempty"`

	// MaxRetries specifies the number of retry attempts on failure
	MaxRetries int `json:"max_retries,omitempty"`
}
