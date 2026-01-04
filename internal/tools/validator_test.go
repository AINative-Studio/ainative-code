package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewValidator(t *testing.T) {
	validator := NewValidator()
	assert.NotNil(t, validator)
}

func TestValidator_Validate(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name        string
		schema      ToolSchema
		input       map[string]interface{}
		expectError bool
		errorField  string
	}{
		{
			name: "valid input - all fields present",
			schema: ToolSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"name": {
						Type:        "string",
						Description: "Name field",
					},
					"age": {
						Type:        "number",
						Description: "Age field",
					},
				},
				Required: []string{"name"},
			},
			input: map[string]interface{}{
				"name": "John",
				"age":  30.0,
			},
			expectError: false,
		},
		{
			name: "missing required field",
			schema: ToolSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"name": {
						Type: "string",
					},
				},
				Required: []string{"name"},
			},
			input:       map[string]interface{}{},
			expectError: true,
			errorField:  "name",
		},
		{
			name: "extra fields allowed (permissive)",
			schema: ToolSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"name": {
						Type: "string",
					},
				},
				Required: []string{"name"},
			},
			input: map[string]interface{}{
				"name":  "John",
				"extra": "field",
			},
			expectError: false,
		},
		{
			name: "invalid schema type",
			schema: ToolSchema{
				Type: "array", // Not "object"
				Properties: map[string]PropertySchema{
					"name": {
						Type: "string",
					},
				},
			},
			input: map[string]interface{}{
				"name": "John",
			},
			expectError: true,
		},
		{
			name: "wrong type - string expected, number provided",
			schema: ToolSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"name": {
						Type: "string",
					},
				},
			},
			input: map[string]interface{}{
				"name": 123,
			},
			expectError: true,
			errorField:  "name",
		},
		{
			name: "wrong type - number expected, string provided",
			schema: ToolSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"count": {
						Type: "number",
					},
				},
			},
			input: map[string]interface{}{
				"count": "not a number",
			},
			expectError: true,
			errorField:  "count",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.schema, tt.input)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorField != "" {
					var inputErr *ErrInvalidInput
					if assert.ErrorAs(t, err, &inputErr) {
						assert.Equal(t, tt.errorField, inputErr.Field)
					}
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidator_ValidateType(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name         string
		fieldName    string
		value        interface{}
		expectedType string
		expectError  bool
	}{
		{
			name:         "valid string",
			fieldName:    "name",
			value:        "test",
			expectedType: "string",
			expectError:  false,
		},
		{
			name:         "valid integer",
			fieldName:    "count",
			value:        42,
			expectedType: "integer",
			expectError:  false,
		},
		{
			name:         "valid number - float",
			fieldName:    "price",
			value:        19.99,
			expectedType: "number",
			expectError:  false,
		},
		{
			name:         "valid number - int",
			fieldName:    "quantity",
			value:        10,
			expectedType: "number",
			expectError:  false,
		},
		{
			name:         "valid boolean",
			fieldName:    "enabled",
			value:        true,
			expectedType: "boolean",
			expectError:  false,
		},
		{
			name:         "valid array - interface slice",
			fieldName:    "items",
			value:        []interface{}{"a", "b", "c"},
			expectedType: "array",
			expectError:  false,
		},
		{
			name:         "valid array - string slice",
			fieldName:    "tags",
			value:        []string{"tag1", "tag2"},
			expectedType: "array",
			expectError:  false,
		},
		{
			name:         "valid object",
			fieldName:    "config",
			value:        map[string]interface{}{"key": "value"},
			expectedType: "object",
			expectError:  false,
		},
		{
			name:         "invalid type - string expected, int provided",
			fieldName:    "name",
			value:        123,
			expectedType: "string",
			expectError:  true,
		},
		{
			name:         "invalid type - number expected, string provided",
			fieldName:    "count",
			value:        "abc",
			expectedType: "number",
			expectError:  true,
		},
		{
			name:         "invalid type - boolean expected, string provided",
			fieldName:    "flag",
			value:        "true",
			expectedType: "boolean",
			expectError:  true,
		},
		{
			name:         "invalid type - array expected, string provided",
			fieldName:    "items",
			value:        "not an array",
			expectedType: "array",
			expectError:  true,
		},
		{
			name:         "invalid type - object expected, string provided",
			fieldName:    "config",
			value:        "not an object",
			expectedType: "object",
			expectError:  true,
		},
		{
			name:         "unsupported type",
			fieldName:    "field",
			value:        "value",
			expectedType: "unsupported_type",
			expectError:  true,
		},
		{
			name:         "integer validation - whole number float",
			fieldName:    "count",
			value:        42.0,
			expectedType: "integer",
			expectError:  false,
		},
		{
			name:         "integer validation - fractional float",
			fieldName:    "count",
			value:        42.5,
			expectedType: "integer",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateType(tt.fieldName, tt.value, tt.expectedType)

			if tt.expectError {
				assert.Error(t, err)
				var inputErr *ErrInvalidInput
				if assert.ErrorAs(t, err, &inputErr) {
					assert.Equal(t, tt.fieldName, inputErr.Field)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidator_ValidateEnum(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name        string
		fieldName   string
		value       string
		enumValues  []interface{}
		expectError bool
	}{
		{
			name:        "valid enum value",
			fieldName:   "status",
			value:       "active",
			enumValues:  []interface{}{"active", "inactive", "pending"},
			expectError: false,
		},
		{
			name:        "invalid enum value",
			fieldName:   "status",
			value:       "deleted",
			enumValues:  []interface{}{"active", "inactive", "pending"},
			expectError: true,
		},
		{
			name:        "empty enum list - value accepted",
			fieldName:   "status",
			value:       "any",
			enumValues:  []interface{}{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert []interface{} to []string for the validator
			stringEnums := make([]string, len(tt.enumValues))
			for i, v := range tt.enumValues {
				stringEnums[i] = v.(string)
			}

			err := validator.validateEnum(tt.fieldName, tt.value, stringEnums)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidator_ValidatePattern(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name        string
		fieldName   string
		value       string
		pattern     string
		expectError bool
	}{
		{
			name:        "valid pattern - email",
			fieldName:   "email",
			value:       "test@example.com",
			pattern:     `^[\w\.\-]+@[\w\.\-]+\.\w+$`,
			expectError: false,
		},
		{
			name:        "invalid pattern - email",
			fieldName:   "email",
			value:       "not-an-email",
			pattern:     `^[\w\.\-]+@[\w\.\-]+\.\w+$`,
			expectError: true,
		},
		{
			name:        "valid pattern - alphanumeric",
			fieldName:   "username",
			value:       "user123",
			pattern:     `^[a-zA-Z0-9]+$`,
			expectError: false,
		},
		{
			name:        "invalid pattern - alphanumeric",
			fieldName:   "username",
			value:       "user@123",
			pattern:     `^[a-zA-Z0-9]+$`,
			expectError: true,
		},
		{
			name:        "invalid regex pattern",
			fieldName:   "field",
			value:       "value",
			pattern:     `[invalid(`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validatePattern(tt.fieldName, tt.value, tt.pattern)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidator_StringValidations(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name        string
		schema      ToolSchema
		input       map[string]interface{}
		expectError bool
		description string
	}{
		{
			name: "valid min length",
			schema: ToolSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"password": {
						Type:      "string",
						MinLength: intPtr(8),
					},
				},
			},
			input: map[string]interface{}{
				"password": "mypassword123",
			},
			expectError: false,
			description: "String meets minimum length requirement",
		},
		{
			name: "invalid min length",
			schema: ToolSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"password": {
						Type:      "string",
						MinLength: intPtr(8),
					},
				},
			},
			input: map[string]interface{}{
				"password": "short",
			},
			expectError: true,
			description: "String is shorter than minimum length",
		},
		{
			name: "valid max length",
			schema: ToolSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"username": {
						Type:      "string",
						MaxLength: intPtr(20),
					},
				},
			},
			input: map[string]interface{}{
				"username": "validuser",
			},
			expectError: false,
			description: "String meets maximum length requirement",
		},
		{
			name: "invalid max length",
			schema: ToolSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"username": {
						Type:      "string",
						MaxLength: intPtr(10),
					},
				},
			},
			input: map[string]interface{}{
				"username": "thisusernameistoolong",
			},
			expectError: true,
			description: "String exceeds maximum length",
		},
		{
			name: "valid pattern match",
			schema: ToolSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"code": {
						Type:    "string",
						Pattern: `^[A-Z]{3}\d{3}$`,
					},
				},
			},
			input: map[string]interface{}{
				"code": "ABC123",
			},
			expectError: false,
			description: "String matches pattern",
		},
		{
			name: "invalid pattern match",
			schema: ToolSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"code": {
						Type:    "string",
						Pattern: `^[A-Z]{3}\d{3}$`,
					},
				},
			},
			input: map[string]interface{}{
				"code": "abc123",
			},
			expectError: true,
			description: "String does not match pattern",
		},
		{
			name: "valid enum value",
			schema: ToolSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"status": {
						Type: "string",
						Enum: []interface{}{"active", "inactive", "pending"},
					},
				},
			},
			input: map[string]interface{}{
				"status": "active",
			},
			expectError: false,
			description: "String is in enum list",
		},
		{
			name: "invalid enum value",
			schema: ToolSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"status": {
						Type: "string",
						Enum: []interface{}{"active", "inactive", "pending"},
					},
				},
			},
			input: map[string]interface{}{
				"status": "deleted",
			},
			expectError: true,
			description: "String is not in enum list",
		},
		{
			name: "combined validations - all pass",
			schema: ToolSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"username": {
						Type:      "string",
						MinLength: intPtr(3),
						MaxLength: intPtr(15),
						Pattern:   `^[a-z]+$`,
					},
				},
			},
			input: map[string]interface{}{
				"username": "johndoe",
			},
			expectError: false,
			description: "All string validations pass",
		},
		{
			name: "combined validations - pattern fails",
			schema: ToolSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"username": {
						Type:      "string",
						MinLength: intPtr(3),
						MaxLength: intPtr(15),
						Pattern:   `^[a-z]+$`,
					},
				},
			},
			input: map[string]interface{}{
				"username": "john123",
			},
			expectError: true,
			description: "Pattern validation fails",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.schema, tt.input)

			if tt.expectError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
			}
		})
	}
}

func TestValidator_ComplexSchemas(t *testing.T) {
	validator := NewValidator()

	t.Run("nested required fields", func(t *testing.T) {
		schema := ToolSchema{
			Type: "object",
			Properties: map[string]PropertySchema{
				"user": {
					Type: "object",
					Properties: map[string]PropertySchema{
						"name": {
							Type: "string",
						},
						"email": {
							Type: "string",
						},
					},
					Required: []string{"name"},
				},
			},
			Required: []string{"user"},
		}

		// Valid nested object
		input := map[string]interface{}{
			"user": map[string]interface{}{
				"name":  "John",
				"email": "john@example.com",
			},
		}
		err := validator.Validate(schema, input)
		assert.NoError(t, err)

		// Missing required parent field
		invalidInput := map[string]interface{}{}
		err = validator.Validate(schema, invalidInput)
		assert.Error(t, err)
	})

	t.Run("multiple types", func(t *testing.T) {
		schema := ToolSchema{
			Type: "object",
			Properties: map[string]PropertySchema{
				"name": {
					Type: "string",
				},
				"age": {
					Type: "integer",
				},
				"active": {
					Type: "boolean",
				},
				"tags": {
					Type: "array",
				},
				"metadata": {
					Type: "object",
				},
			},
			Required: []string{"name", "age"},
		}

		input := map[string]interface{}{
			"name":     "John",
			"age":      30,
			"active":   true,
			"tags":     []interface{}{"user", "admin"},
			"metadata": map[string]interface{}{"role": "admin"},
		}

		err := validator.Validate(schema, input)
		assert.NoError(t, err)
	})
}

// Helper function to create int pointers
func intPtr(i int) *int {
	return &i
}
