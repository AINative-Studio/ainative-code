// Package tools provides an extensible tool execution framework for LLM assistants.
package tools

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Registry manages tool registration and execution with thread safety.
type Registry struct {
	mu        sync.RWMutex
	tools     map[string]Tool
	validator *Validator
}

// NewRegistry creates a new Registry instance.
func NewRegistry() *Registry {
	return &Registry{
		tools:     make(map[string]Tool),
		validator: NewValidator(),
	}
}

// Register registers a new tool in the registry.
// Returns ErrToolConflict if a tool with the same name is already registered.
func (r *Registry) Register(tool Tool) error {
	if tool == nil {
		return fmt.Errorf("cannot register nil tool")
	}

	name := tool.Name()
	if name == "" {
		return fmt.Errorf("tool name cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tools[name]; exists {
		return &ErrToolConflict{
			ToolName: name,
		}
	}

	r.tools[name] = tool
	return nil
}

// Unregister removes a tool from the registry.
// Returns ErrToolNotFound if the tool is not registered.
func (r *Registry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tools[name]; !exists {
		return &ErrToolNotFound{
			ToolName: name,
		}
	}

	delete(r.tools, name)
	return nil
}

// Get retrieves a tool by name.
// Returns ErrToolNotFound if the tool is not registered.
func (r *Registry) Get(name string) (Tool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tool, exists := r.tools[name]
	if !exists {
		return nil, &ErrToolNotFound{
			ToolName: name,
		}
	}

	return tool, nil
}

// List returns all registered tools.
func (r *Registry) List() []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tools := make([]Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}

	return tools
}

// ListByCategory returns all tools in a specific category.
func (r *Registry) ListByCategory(category Category) []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tools := make([]Tool, 0)
	for _, tool := range r.tools {
		if tool.Category() == category {
			tools = append(tools, tool)
		}
	}

	return tools
}

// Execute executes a tool with the provided input and execution context.
// This method handles validation, timeout enforcement, and error handling.
func (r *Registry) Execute(ctx context.Context, toolName string, input map[string]interface{}, execCtx *ExecutionContext) (*Result, error) {
	// Get the tool
	tool, err := r.Get(toolName)
	if err != nil {
		return nil, err
	}

	// Validate input against tool schema
	schema := tool.Schema()
	if err := r.validator.Validate(schema, input); err != nil {
		// Add tool name to validation error if not already present
		if valErr, ok := err.(*ErrInvalidInput); ok && valErr.ToolName == "" {
			valErr.ToolName = toolName
			return nil, valErr
		}
		return nil, err
	}

	// Use default execution context if not provided
	if execCtx == nil {
		execCtx = NewExecutionContext()
	}

	// Dry-run mode: return success without executing
	if execCtx.DryRun {
		return &Result{
			Success: true,
			Output:  fmt.Sprintf("Dry-run mode: would execute tool %s", toolName),
			Metadata: map[string]interface{}{
				"dry_run":   true,
				"tool_name": toolName,
			},
		}, nil
	}

	// Create context with timeout
	execCtxWithTimeout, cancel := context.WithTimeout(ctx, execCtx.Timeout)
	defer cancel()

	// Execute the tool in a goroutine to handle timeout
	resultChan := make(chan *Result, 1)
	errChan := make(chan error, 1)

	go func() {
		result, err := tool.Execute(execCtxWithTimeout, input)
		if err != nil {
			errChan <- err
			return
		}
		resultChan <- result
	}()

	// Wait for execution or timeout
	select {
	case <-execCtxWithTimeout.Done():
		// Check if it was a timeout or cancellation
		if execCtxWithTimeout.Err() == context.DeadlineExceeded {
			return nil, &ErrTimeout{
				ToolName: toolName,
				Duration: execCtx.Timeout.String(),
			}
		}
		// Context was cancelled
		return nil, &ErrExecutionFailed{
			ToolName: toolName,
			Reason:   "execution cancelled",
			Cause:    execCtxWithTimeout.Err(),
		}

	case err := <-errChan:
		// Execution failed
		if execErr, ok := err.(*ErrExecutionFailed); ok && execErr.ToolName == "" {
			execErr.ToolName = toolName
			return nil, execErr
		}
		return nil, &ErrExecutionFailed{
			ToolName: toolName,
			Reason:   "tool execution failed",
			Cause:    err,
		}

	case result := <-resultChan:
		// Execution succeeded
		// Check output size limit
		if execCtx.MaxOutputSize > 0 && int64(len(result.Output)) > execCtx.MaxOutputSize {
			return nil, &ErrOutputTooLarge{
				ToolName:   toolName,
				OutputSize: int64(len(result.Output)),
				MaxSize:    execCtx.MaxOutputSize,
			}
		}

		// Add execution metadata
		if result.Metadata == nil {
			result.Metadata = make(map[string]interface{})
		}
		result.Metadata["tool_name"] = toolName
		result.Metadata["execution_time"] = time.Now().Format(time.RFC3339)

		return result, nil
	}
}

// Schemas returns a map of tool names to their schemas.
func (r *Registry) Schemas() map[string]ToolSchema {
	r.mu.RLock()
	defer r.mu.RUnlock()

	schemas := make(map[string]ToolSchema, len(r.tools))
	for name, tool := range r.tools {
		schemas[name] = tool.Schema()
	}

	return schemas
}
