package tools

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockTool is a simple test implementation of the Tool interface
type MockTool struct {
	name                 string
	description          string
	schema               ToolSchema
	category             Category
	requiresConfirmation bool
	executeFunc          func(ctx context.Context, input map[string]interface{}) (*Result, error)
}

func (m *MockTool) Name() string {
	return m.name
}

func (m *MockTool) Description() string {
	return m.description
}

func (m *MockTool) Schema() ToolSchema {
	return m.schema
}

func (m *MockTool) Category() Category {
	return m.category
}

func (m *MockTool) RequiresConfirmation() bool {
	return m.requiresConfirmation
}

func (m *MockTool) Execute(ctx context.Context, input map[string]interface{}) (*Result, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, input)
	}
	return &Result{
		Success: true,
		Output:  "mock output",
	}, nil
}

// createMockTool creates a basic mock tool for testing
func createMockTool(name string) *MockTool {
	maxLen := 100
	return &MockTool{
		name:        name,
		description: "A mock tool for testing",
		schema: ToolSchema{
			Type: "object",
			Properties: map[string]PropertyDef{
				"input": {
					Type:        "string",
					Description: "Test input parameter",
					MaxLength:   &maxLen,
				},
			},
			Required: []string{"input"},
		},
		category:             CategorySystem,
		requiresConfirmation: false,
	}
}

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()

	assert.NotNil(t, registry)
	assert.NotNil(t, registry.tools)
	assert.NotNil(t, registry.validator)
	assert.Empty(t, registry.tools)
}

func TestRegistry_Register(t *testing.T) {
	tests := []struct {
		name        string
		tool        Tool
		expectError bool
		errorType   error
	}{
		{
			name:        "successful registration",
			tool:        createMockTool("test_tool"),
			expectError: false,
		},
		{
			name:        "nil tool",
			tool:        nil,
			expectError: true,
		},
		{
			name: "empty tool name",
			tool: &MockTool{
				name: "",
			},
			expectError: true,
		},
		{
			name:        "duplicate tool name",
			tool:        createMockTool("duplicate"),
			expectError: false, // First registration succeeds
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := NewRegistry()

			err := registry.Register(tt.tool)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Test duplicate registration
				if tt.tool != nil && tt.tool.Name() != "" {
					err = registry.Register(tt.tool)
					assert.Error(t, err)
					var conflictErr *ErrToolConflict
					assert.ErrorAs(t, err, &conflictErr)
					assert.Equal(t, tt.tool.Name(), conflictErr.ToolName)
				}
			}
		})
	}
}

func TestRegistry_Unregister(t *testing.T) {
	tests := []struct {
		name        string
		setupTools  []string
		unregister  string
		expectError bool
	}{
		{
			name:        "successful unregistration",
			setupTools:  []string{"tool1", "tool2"},
			unregister:  "tool1",
			expectError: false,
		},
		{
			name:        "unregister non-existent tool",
			setupTools:  []string{"tool1"},
			unregister:  "nonexistent",
			expectError: true,
		},
		{
			name:        "unregister from empty registry",
			setupTools:  []string{},
			unregister:  "tool1",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := NewRegistry()

			// Setup tools
			for _, name := range tt.setupTools {
				err := registry.Register(createMockTool(name))
				require.NoError(t, err)
			}

			// Unregister
			err := registry.Unregister(tt.unregister)

			if tt.expectError {
				assert.Error(t, err)
				var notFoundErr *ErrToolNotFound
				assert.ErrorAs(t, err, &notFoundErr)
			} else {
				assert.NoError(t, err)

				// Verify tool is gone
				_, err := registry.Get(tt.unregister)
				assert.Error(t, err)
			}
		})
	}
}

func TestRegistry_Get(t *testing.T) {
	tests := []struct {
		name        string
		setupTools  []string
		getTool     string
		expectError bool
	}{
		{
			name:        "get existing tool",
			setupTools:  []string{"tool1", "tool2"},
			getTool:     "tool1",
			expectError: false,
		},
		{
			name:        "get non-existent tool",
			setupTools:  []string{"tool1"},
			getTool:     "nonexistent",
			expectError: true,
		},
		{
			name:        "get from empty registry",
			setupTools:  []string{},
			getTool:     "tool1",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := NewRegistry()

			// Setup tools
			for _, name := range tt.setupTools {
				err := registry.Register(createMockTool(name))
				require.NoError(t, err)
			}

			// Get tool
			tool, err := registry.Get(tt.getTool)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, tool)
				var notFoundErr *ErrToolNotFound
				assert.ErrorAs(t, err, &notFoundErr)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, tool)
				assert.Equal(t, tt.getTool, tool.Name())
			}
		})
	}
}

func TestRegistry_List(t *testing.T) {
	tests := []struct {
		name       string
		setupTools []string
		expected   int
	}{
		{
			name:       "list multiple tools",
			setupTools: []string{"tool1", "tool2", "tool3"},
			expected:   3,
		},
		{
			name:       "list empty registry",
			setupTools: []string{},
			expected:   0,
		},
		{
			name:       "list single tool",
			setupTools: []string{"tool1"},
			expected:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := NewRegistry()

			// Setup tools
			for _, name := range tt.setupTools {
				err := registry.Register(createMockTool(name))
				require.NoError(t, err)
			}

			// List tools
			tools := registry.List()

			assert.Len(t, tools, tt.expected)

			// Verify all tools are present
			toolNames := make(map[string]bool)
			for _, tool := range tools {
				toolNames[tool.Name()] = true
			}

			for _, expectedName := range tt.setupTools {
				assert.True(t, toolNames[expectedName], "Expected tool %s not found", expectedName)
			}
		})
	}
}

func TestRegistry_ListByCategory(t *testing.T) {
	registry := NewRegistry()

	// Create tools with different categories
	fileTool := createMockTool("file_tool")
	fileTool.category = CategoryFilesystem

	netTool := createMockTool("net_tool")
	netTool.category = CategoryNetwork

	sysTool1 := createMockTool("sys_tool1")
	sysTool1.category = CategorySystem

	sysTool2 := createMockTool("sys_tool2")
	sysTool2.category = CategorySystem

	require.NoError(t, registry.Register(fileTool))
	require.NoError(t, registry.Register(netTool))
	require.NoError(t, registry.Register(sysTool1))
	require.NoError(t, registry.Register(sysTool2))

	tests := []struct {
		name     string
		category Category
		expected int
	}{
		{
			name:     "filesystem category",
			category: CategoryFilesystem,
			expected: 1,
		},
		{
			name:     "network category",
			category: CategoryNetwork,
			expected: 1,
		},
		{
			name:     "system category",
			category: CategorySystem,
			expected: 2,
		},
		{
			name:     "database category (empty)",
			category: CategoryDatabase,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tools := registry.ListByCategory(tt.category)
			assert.Len(t, tools, tt.expected)

			// Verify all tools have the correct category
			for _, tool := range tools {
				assert.Equal(t, tt.category, tool.Category())
			}
		})
	}
}

func TestRegistry_Execute(t *testing.T) {
	tests := []struct {
		name          string
		toolName      string
		input         map[string]interface{}
		execCtx       *ExecutionContext
		setupTool     func() Tool
		expectError   bool
		errorType     interface{}
		validateResult func(t *testing.T, result *Result)
	}{
		{
			name:     "successful execution",
			toolName: "success_tool",
			input: map[string]interface{}{
				"input": "test value",
			},
			setupTool: func() Tool {
				return createMockTool("success_tool")
			},
			expectError: false,
			validateResult: func(t *testing.T, result *Result) {
				assert.True(t, result.Success)
				assert.NotEmpty(t, result.Output)
				assert.Contains(t, result.Metadata, "tool_name")
			},
		},
		{
			name:     "tool not found",
			toolName: "nonexistent_tool",
			input: map[string]interface{}{
				"input": "test",
			},
			setupTool:   nil,
			expectError: true,
			errorType:   &ErrToolNotFound{},
		},
		{
			name:     "invalid input - missing required field",
			toolName: "validation_tool",
			input:    map[string]interface{}{}, // Missing "input" field
			setupTool: func() Tool {
				return createMockTool("validation_tool")
			},
			expectError: true,
			errorType:   &ErrInvalidInput{},
		},
		{
			name:     "execution timeout",
			toolName: "timeout_tool",
			input: map[string]interface{}{
				"input": "test",
			},
			execCtx: &ExecutionContext{
				Timeout: 10 * time.Millisecond,
			},
			setupTool: func() Tool {
				tool := createMockTool("timeout_tool")
				tool.executeFunc = func(ctx context.Context, input map[string]interface{}) (*Result, error) {
					// Simulate long-running operation
					time.Sleep(100 * time.Millisecond)
					return &Result{Success: true, Output: "done"}, nil
				}
				return tool
			},
			expectError: true,
			errorType:   &ErrTimeout{},
		},
		{
			name:     "execution failure",
			toolName: "error_tool",
			input: map[string]interface{}{
				"input": "test",
			},
			setupTool: func() Tool {
				tool := createMockTool("error_tool")
				tool.executeFunc = func(ctx context.Context, input map[string]interface{}) (*Result, error) {
					return nil, errors.New("execution failed")
				}
				return tool
			},
			expectError: true,
			errorType:   &ErrExecutionFailed{},
		},
		{
			name:     "dry run mode",
			toolName: "dry_run_tool",
			input: map[string]interface{}{
				"input": "test",
			},
			execCtx: &ExecutionContext{
				DryRun: true,
			},
			setupTool: func() Tool {
				return createMockTool("dry_run_tool")
			},
			expectError: false,
			validateResult: func(t *testing.T, result *Result) {
				assert.True(t, result.Success)
				assert.Contains(t, result.Output, "Dry-run mode")
				assert.Equal(t, true, result.Metadata["dry_run"])
			},
		},
		{
			name:     "output too large",
			toolName: "large_output_tool",
			input: map[string]interface{}{
				"input": "test",
			},
			execCtx: &ExecutionContext{
				MaxOutputSize: 10, // Very small limit
			},
			setupTool: func() Tool {
				tool := createMockTool("large_output_tool")
				tool.executeFunc = func(ctx context.Context, input map[string]interface{}) (*Result, error) {
					return &Result{
						Success: true,
						Output:  "This is a very long output that exceeds the maximum size limit",
					}, nil
				}
				return tool
			},
			expectError: true,
			errorType:   &ErrOutputTooLarge{},
		},
		{
			name:     "context cancellation",
			toolName: "cancel_tool",
			input: map[string]interface{}{
				"input": "test",
			},
			setupTool: func() Tool {
				tool := createMockTool("cancel_tool")
				tool.executeFunc = func(ctx context.Context, input map[string]interface{}) (*Result, error) {
					<-ctx.Done()
					return nil, ctx.Err()
				}
				return tool
			},
			expectError: true,
			errorType:   &ErrExecutionFailed{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := NewRegistry()

			// Setup tool if provided
			if tt.setupTool != nil {
				tool := tt.setupTool()
				err := registry.Register(tool)
				require.NoError(t, err)
			}

			// Create context
			ctx := context.Background()
			if tt.name == "context cancellation" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				// Cancel immediately
				cancel()
			}

			// Execute
			result, err := registry.Execute(ctx, tt.toolName, tt.input, tt.execCtx)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorType != nil {
					assert.ErrorAs(t, err, &tt.errorType)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
		})
	}
}

func TestRegistry_Schemas(t *testing.T) {
	registry := NewRegistry()

	// Register multiple tools
	tool1 := createMockTool("tool1")
	tool2 := createMockTool("tool2")
	tool3 := createMockTool("tool3")

	require.NoError(t, registry.Register(tool1))
	require.NoError(t, registry.Register(tool2))
	require.NoError(t, registry.Register(tool3))

	// Get schemas
	schemas := registry.Schemas()

	assert.Len(t, schemas, 3)
	assert.Contains(t, schemas, "tool1")
	assert.Contains(t, schemas, "tool2")
	assert.Contains(t, schemas, "tool3")

	// Verify schema structure
	for name, schema := range schemas {
		assert.Equal(t, "object", schema.Type)
		assert.NotNil(t, schema.Properties)
		t.Logf("Tool %s has schema: %+v", name, schema)
	}
}

func TestExecutionContext(t *testing.T) {
	t.Run("NewExecutionContext with defaults", func(t *testing.T) {
		ctx := NewExecutionContext()

		assert.NotNil(t, ctx)
		assert.Equal(t, 30*time.Second, ctx.Timeout)
		assert.NotNil(t, ctx.Environment)
		assert.NotNil(t, ctx.AllowedPaths)
		assert.Equal(t, int64(10*1024*1024), ctx.MaxOutputSize)
		assert.False(t, ctx.DryRun)
	})

	t.Run("NewExecutionContext with options", func(t *testing.T) {
		ctx := NewExecutionContext(
			WithTimeout(60*time.Second),
			WithWorkingDirectory("/tmp"),
			WithEnvironment(map[string]string{"KEY": "value"}),
			WithAllowedPaths([]string{"/safe/path"}),
			WithMaxOutputSize(1024),
			WithDryRun(true),
		)

		assert.Equal(t, 60*time.Second, ctx.Timeout)
		assert.Equal(t, "/tmp", ctx.WorkingDirectory)
		assert.Equal(t, "value", ctx.Environment["KEY"])
		assert.Contains(t, ctx.AllowedPaths, "/safe/path")
		assert.Equal(t, int64(1024), ctx.MaxOutputSize)
		assert.True(t, ctx.DryRun)
	})
}

func TestRegistry_ConcurrentAccess(t *testing.T) {
	registry := NewRegistry()

	// Register initial tool
	initialTool := createMockTool("initial")
	require.NoError(t, registry.Register(initialTool))

	// Run concurrent operations
	done := make(chan bool)

	// Concurrent registrations
	for i := 0; i < 10; i++ {
		go func(id int) {
			tool := createMockTool(string(rune('a' + id)))
			_ = registry.Register(tool)
			done <- true
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 10; i++ {
		go func() {
			_ = registry.List()
			_, _ = registry.Get("initial")
			_ = registry.Schemas()
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}

	// Verify registry is still functional
	tools := registry.List()
	assert.GreaterOrEqual(t, len(tools), 1)
}

func TestMockTool_Implementation(t *testing.T) {
	tool := createMockTool("test")

	assert.Equal(t, "test", tool.Name())
	assert.NotEmpty(t, tool.Description())
	assert.Equal(t, CategorySystem, tool.Category())
	assert.False(t, tool.RequiresConfirmation())

	schema := tool.Schema()
	assert.Equal(t, "object", schema.Type)
	assert.Contains(t, schema.Properties, "input")
	assert.Contains(t, schema.Required, "input")

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"input": "test",
	})
	assert.NoError(t, err)
	assert.True(t, result.Success)
	assert.NotEmpty(t, result.Output)
}
