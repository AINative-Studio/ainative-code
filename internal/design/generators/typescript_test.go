package generators

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/AINative-studio/ainative-code/internal/design"
)

func TestNewTypeScriptGenerator(t *testing.T) {
	t.Run("should create TypeScript generator", func(t *testing.T) {
		// When
		gen := NewTypeScriptGenerator()

		// Then
		assert.NotNil(t, gen)
		assert.NotNil(t, gen.engine)
	})
}

func TestTypeScriptGenerator_GenerateTypeScript(t *testing.T) {
	t.Run("should generate TypeScript constants", func(t *testing.T) {
		// Given
		gen := NewTypeScriptGenerator()
		tokens := []design.Token{
			{Name: "primary-color", Type: "color", Value: "#007bff"},
			{Name: "secondary-color", Type: "color", Value: "#6c757d"},
		}

		// When
		output, err := gen.Generate(tokens, "typescript")

		// Then
		require.NoError(t, err)
		assert.Contains(t, output, "export const DesignTokens")
		assert.Contains(t, output, "primaryColor: \"#007bff\"")
		assert.Contains(t, output, "secondaryColor: \"#6c757d\"")
		assert.Contains(t, output, "as const")
		assert.Contains(t, output, "export type DesignToken")
	})

	t.Run("should convert names to camelCase", func(t *testing.T) {
		// Given
		gen := NewTypeScriptGenerator()
		tokens := []design.Token{
			{Name: "font-size-base", Type: "font-size", Value: "16px"},
			{Name: "spacing_large", Type: "spacing", Value: "32px"},
		}

		// When
		output, err := gen.Generate(tokens, "typescript")

		// Then
		require.NoError(t, err)
		assert.Contains(t, output, "fontSizeBase:")
		assert.Contains(t, output, "spacingLarge:")
	})

	t.Run("should handle empty tokens", func(t *testing.T) {
		// Given
		gen := NewTypeScriptGenerator()
		tokens := []design.Token{}

		// When
		output, err := gen.Generate(tokens, "typescript")

		// Then
		require.NoError(t, err)
		assert.Contains(t, output, "export const DesignTokens")
	})
}

func TestTypeScriptGenerator_GenerateJavaScript(t *testing.T) {
	t.Run("should generate JavaScript constants", func(t *testing.T) {
		// Given
		gen := NewTypeScriptGenerator()
		tokens := []design.Token{
			{Name: "primary-color", Type: "color", Value: "#007bff"},
		}

		// When
		output, err := gen.Generate(tokens, "javascript")

		// Then
		require.NoError(t, err)
		assert.Contains(t, output, "export const DesignTokens")
		assert.Contains(t, output, "primaryColor: \"#007bff\"")
		// Should not contain TypeScript-specific syntax
		assert.NotContains(t, output, "as const")
		assert.NotContains(t, output, "export type")
	})
}

func TestTypeScriptGenerator_Metadata(t *testing.T) {
	t.Run("should return correct name", func(t *testing.T) {
		// Given
		gen := NewTypeScriptGenerator()

		// When
		name := gen.Name()

		// Then
		assert.Contains(t, strings.ToLower(name), "typescript")
	})

	t.Run("should return supported formats", func(t *testing.T) {
		// Given
		gen := NewTypeScriptGenerator()

		// When
		formats := gen.SupportedFormats()

		// Then
		assert.Contains(t, formats, "typescript")
		assert.Contains(t, formats, "ts")
		assert.Contains(t, formats, "javascript")
		assert.Contains(t, formats, "js")
	})
}
