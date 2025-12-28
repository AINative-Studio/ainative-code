// Package tools provides an extensible tool execution framework for LLM assistants.
package tools

import (
	"fmt"
	"regexp"
	"strings"
)

// Validator validates tool input against JSON schemas.
type Validator struct{}

// NewValidator creates a new Validator instance.
func NewValidator() *Validator {
	return &Validator{}
}

// Validate validates the input against the provided schema.
func (v *Validator) Validate(schema ToolSchema, input map[string]interface{}) error {
	// Validate that schema type is "object"
	if schema.Type != "object" {
		return fmt.Errorf("schema type must be 'object', got '%s'", schema.Type)
	}

	// Check required fields
	for _, requiredField := range schema.Required {
		if _, exists := input[requiredField]; !exists {
			return &ErrInvalidInput{
				Field:  requiredField,
				Reason: "required field is missing",
			}
		}
	}

	// Validate each input field against schema properties
	for fieldName, fieldValue := range input {
		propertyDef, exists := schema.Properties[fieldName]
		if !exists {
			// Field not in schema - this could be strict or permissive
			// For now, we'll allow extra fields (permissive mode)
			continue
		}

		if err := v.validateProperty(fieldName, fieldValue, propertyDef); err != nil {
			return err
		}
	}

	return nil
}

// validateProperty validates a single property value against its definition.
func (v *Validator) validateProperty(fieldName string, value interface{}, propDef PropertyDef) error {
	// Type validation
	if err := v.validateType(fieldName, value, propDef.Type); err != nil {
		return err
	}

	// String-specific validations
	if propDef.Type == "string" {
		strValue, ok := value.(string)
		if !ok {
			return &ErrInvalidInput{
				Field:  fieldName,
				Reason: fmt.Sprintf("expected string, got %T", value),
			}
		}

		// Enum validation
		if len(propDef.Enum) > 0 {
			if err := v.validateEnum(fieldName, strValue, propDef.Enum); err != nil {
				return err
			}
		}

		// MinLength validation
		if propDef.MinLength != nil {
			if len(strValue) < *propDef.MinLength {
				return &ErrInvalidInput{
					Field:  fieldName,
					Reason: fmt.Sprintf("string length %d is less than minimum %d", len(strValue), *propDef.MinLength),
				}
			}
		}

		// MaxLength validation
		if propDef.MaxLength != nil {
			if len(strValue) > *propDef.MaxLength {
				return &ErrInvalidInput{
					Field:  fieldName,
					Reason: fmt.Sprintf("string length %d exceeds maximum %d", len(strValue), *propDef.MaxLength),
				}
			}
		}

		// Pattern validation
		if propDef.Pattern != "" {
			if err := v.validatePattern(fieldName, strValue, propDef.Pattern); err != nil {
				return err
			}
		}
	}

	return nil
}

// validateType validates that the value matches the expected type.
func (v *Validator) validateType(fieldName string, value interface{}, expectedType string) error {
	switch expectedType {
	case "string":
		if _, ok := value.(string); !ok {
			return &ErrInvalidInput{
				Field:  fieldName,
				Reason: fmt.Sprintf("expected type string, got %T", value),
			}
		}
	case "number":
		switch value.(type) {
		case float64, float32, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			// Valid numeric type
		default:
			return &ErrInvalidInput{
				Field:  fieldName,
				Reason: fmt.Sprintf("expected type number, got %T", value),
			}
		}
	case "integer":
		switch value.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			// Valid integer type
		case float64:
			// JSON unmarshaling often produces float64 for all numbers
			// Check if it's a whole number
			if floatVal := value.(float64); floatVal != float64(int64(floatVal)) {
				return &ErrInvalidInput{
					Field:  fieldName,
					Reason: fmt.Sprintf("expected integer, got float with fractional part: %v", floatVal),
				}
			}
		default:
			return &ErrInvalidInput{
				Field:  fieldName,
				Reason: fmt.Sprintf("expected type integer, got %T", value),
			}
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return &ErrInvalidInput{
				Field:  fieldName,
				Reason: fmt.Sprintf("expected type boolean, got %T", value),
			}
		}
	case "array":
		// Check for slice or array types
		switch value.(type) {
		case []interface{}, []string, []int, []float64, []bool:
			// Valid array types
		default:
			return &ErrInvalidInput{
				Field:  fieldName,
				Reason: fmt.Sprintf("expected type array, got %T", value),
			}
		}
	case "object":
		if _, ok := value.(map[string]interface{}); !ok {
			return &ErrInvalidInput{
				Field:  fieldName,
				Reason: fmt.Sprintf("expected type object, got %T", value),
			}
		}
	default:
		return &ErrInvalidInput{
			Field:  fieldName,
			Reason: fmt.Sprintf("unsupported type in schema: %s", expectedType),
		}
	}

	return nil
}

// validateEnum validates that the string value is one of the allowed enum values.
func (v *Validator) validateEnum(fieldName, value string, enumValues []string) error {
	for _, allowedValue := range enumValues {
		if value == allowedValue {
			return nil
		}
	}

	return &ErrInvalidInput{
		Field:  fieldName,
		Reason: fmt.Sprintf("value '%s' is not in allowed enum values: [%s]", value, strings.Join(enumValues, ", ")),
	}
}

// validatePattern validates that the string value matches the regex pattern.
func (v *Validator) validatePattern(fieldName, value, pattern string) error {
	matched, err := regexp.MatchString(pattern, value)
	if err != nil {
		return &ErrInvalidInput{
			Field:  fieldName,
			Reason: fmt.Sprintf("invalid regex pattern '%s': %v", pattern, err),
		}
	}

	if !matched {
		return &ErrInvalidInput{
			Field:  fieldName,
			Reason: fmt.Sprintf("value '%s' does not match pattern '%s'", value, pattern),
		}
	}

	return nil
}
