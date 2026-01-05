package design

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTemplateEngine(t *testing.T) {
	t.Run("should create template engine with built-in functions", func(t *testing.T) {
		// When
		engine := NewTemplateEngine()

		// Then
		assert.NotNil(t, engine)
		assert.NotNil(t, engine.templates)
		assert.NotNil(t, engine.funcMap)
	})
}

func TestTemplateEngine_RegisterTemplate(t *testing.T) {
	t.Run("should register valid template", func(t *testing.T) {
		// Given
		engine := NewTemplateEngine()
		templateContent := "Hello {{ .Name }}"

		// When
		err := engine.RegisterTemplate("test", templateContent)

		// Then
		assert.NoError(t, err)
	})

	t.Run("should return error for invalid template syntax", func(t *testing.T) {
		// Given
		engine := NewTemplateEngine()
		templateContent := "Hello {{ .Name "

		// When
		err := engine.RegisterTemplate("test", templateContent)

		// Then
		assert.Error(t, err)
	})
}

func TestTemplateEngine_Execute(t *testing.T) {
	t.Run("should execute registered template", func(t *testing.T) {
		// Given
		engine := NewTemplateEngine()
		engine.RegisterTemplate("greeting", "Hello {{ .Name }}!")
		data := map[string]string{"Name": "World"}

		// When
		result, err := engine.Execute("greeting", data)

		// Then
		require.NoError(t, err)
		assert.Equal(t, "Hello World!", result)
	})

	t.Run("should return error for non-existent template", func(t *testing.T) {
		// Given
		engine := NewTemplateEngine()

		// When
		_, err := engine.Execute("nonexistent", nil)

		// Then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestTemplateEngine_ExecuteString(t *testing.T) {
	t.Run("should execute inline template", func(t *testing.T) {
		// Given
		engine := NewTemplateEngine()
		templateStr := "Color: {{ .Color }}"
		data := map[string]string{"Color": "#ff0000"}

		// When
		result, err := engine.ExecuteString(templateStr, data)

		// Then
		require.NoError(t, err)
		assert.Equal(t, "Color: #ff0000", result)
	})
}

func TestTemplateHelperFunctions(t *testing.T) {
	t.Run("kebabCase should convert to kebab-case", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"primaryColor", "primary-color"},
			{"primary_color", "primary-color"},
			{"primary.color", "primary-color"},
			{"PRIMARY_COLOR", "primary-color"},
		}

		for _, tt := range tests {
			result := toKebabCase(tt.input)
			assert.Equal(t, tt.expected, result, "Input: %s", tt.input)
		}
	})

	t.Run("camelCase should convert to camelCase", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"primary-color", "primaryColor"},
			{"primary_color", "primaryColor"},
			{"primary.color", "primaryColor"},
			{"PRIMARY-COLOR", "primaryColor"},
		}

		for _, tt := range tests {
			result := toCamelCase(tt.input)
			assert.Equal(t, tt.expected, result, "Input: %s", tt.input)
		}
	})

	t.Run("pascalCase should convert to PascalCase", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"primary-color", "PrimaryColor"},
			{"primary_color", "PrimaryColor"},
			{"primary.color", "PrimaryColor"},
		}

		for _, tt := range tests {
			result := toPascalCase(tt.input)
			assert.Equal(t, tt.expected, result, "Input: %s", tt.input)
		}
	})

	t.Run("snakeCase should convert to snake_case", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"primaryColor", "primary_color"},
			{"primary-color", "primary_color"},
			{"primary.color", "primary_color"},
			{"PRIMARY-COLOR", "primary_color"},
		}

		for _, tt := range tests {
			result := toSnakeCase(tt.input)
			assert.Equal(t, tt.expected, result, "Input: %s", tt.input)
		}
	})

	t.Run("quote should wrap in double quotes", func(t *testing.T) {
		result := quote("test")
		assert.Equal(t, "\"test\"", result)
	})

	t.Run("indent should indent lines", func(t *testing.T) {
		input := "line1\nline2\nline3"
		result := indent(2, input)
		expected := "  line1\n  line2\n  line3"
		assert.Equal(t, expected, result)
	})
}

func TestTemplateEngine_WithHelperFunctions(t *testing.T) {
	t.Run("should use kebabCase in template", func(t *testing.T) {
		// Given
		engine := NewTemplateEngine()
		templateStr := "{{ .Name | kebabCase }}"
		data := map[string]string{"Name": "primaryColor"}

		// When
		result, err := engine.ExecuteString(templateStr, data)

		// Then
		require.NoError(t, err)
		assert.Equal(t, "primary-color", result)
	})

	t.Run("should use camelCase in template", func(t *testing.T) {
		// Given
		engine := NewTemplateEngine()
		templateStr := "{{ .Name | camelCase }}"
		data := map[string]string{"Name": "primary-color"}

		// When
		result, err := engine.ExecuteString(templateStr, data)

		// Then
		require.NoError(t, err)
		assert.Equal(t, "primaryColor", result)
	})

	t.Run("should chain multiple functions", func(t *testing.T) {
		// Given
		engine := NewTemplateEngine()
		templateStr := "{{ .Name | kebabCase | upper }}"
		data := map[string]string{"Name": "primaryColor"}

		// When
		result, err := engine.ExecuteString(templateStr, data)

		// Then
		require.NoError(t, err)
		assert.Equal(t, "PRIMARY-COLOR", result)
	})
}
