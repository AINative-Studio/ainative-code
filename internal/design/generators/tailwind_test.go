package generators

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/AINative-studio/ainative-code/internal/design"
)

func TestNewTailwindGenerator(t *testing.T) {
	t.Run("should create Tailwind generator", func(t *testing.T) {
		// When
		gen := NewTailwindGenerator()

		// Then
		assert.NotNil(t, gen)
		assert.NotNil(t, gen.engine)
	})
}

func TestTailwindGenerator_Generate(t *testing.T) {
	t.Run("should generate Tailwind config with colors", func(t *testing.T) {
		// Given
		gen := NewTailwindGenerator()
		tokens := []design.Token{
			{Name: "primary-color", Type: "color", Value: "#007bff"},
			{Name: "secondary-color", Type: "color", Value: "#6c757d"},
		}

		// When
		output, err := gen.Generate(tokens)

		// Then
		require.NoError(t, err)
		assert.Contains(t, output, "colors:")
		assert.Contains(t, output, "\"primary-color\": \"#007bff\"")
		assert.Contains(t, output, "\"secondary-color\": \"#6c757d\"")
		assert.Contains(t, output, "module.exports")
	})

	t.Run("should generate Tailwind config with spacing", func(t *testing.T) {
		// Given
		gen := NewTailwindGenerator()
		tokens := []design.Token{
			{Name: "spacing-sm", Type: "spacing", Value: "8px"},
			{Name: "spacing-md", Type: "spacing", Value: "16px"},
		}

		// When
		output, err := gen.Generate(tokens)

		// Then
		require.NoError(t, err)
		assert.Contains(t, output, "spacing:")
		assert.Contains(t, output, "\"spacing-sm\": \"8px\"")
	})

	t.Run("should generate Tailwind config with font families", func(t *testing.T) {
		// Given
		gen := NewTailwindGenerator()
		tokens := []design.Token{
			{Name: "font-sans", Type: "font-family", Value: "Helvetica, Arial, sans-serif"},
		}

		// When
		output, err := gen.Generate(tokens)

		// Then
		require.NoError(t, err)
		assert.Contains(t, output, "fontFamily:")
		assert.Contains(t, output, "'Helvetica'")
		assert.Contains(t, output, "'Arial'")
	})

	t.Run("should generate Tailwind config with multiple token types", func(t *testing.T) {
		// Given
		gen := NewTailwindGenerator()
		tokens := []design.Token{
			{Name: "primary", Type: "color", Value: "#007bff"},
			{Name: "spacing-base", Type: "spacing", Value: "16px"},
			{Name: "font-size-base", Type: "font-size", Value: "16px"},
			{Name: "radius-md", Type: "border-radius", Value: "8px"},
			{Name: "shadow-sm", Type: "shadow", Value: "0 1px 2px rgba(0,0,0,0.1)"},
		}

		// When
		output, err := gen.Generate(tokens)

		// Then
		require.NoError(t, err)
		assert.Contains(t, output, "colors:")
		assert.Contains(t, output, "spacing:")
		assert.Contains(t, output, "fontSize:")
		assert.Contains(t, output, "borderRadius:")
		assert.Contains(t, output, "boxShadow:")
	})

	t.Run("should handle empty tokens", func(t *testing.T) {
		// Given
		gen := NewTailwindGenerator()
		tokens := []design.Token{}

		// When
		output, err := gen.Generate(tokens)

		// Then
		require.NoError(t, err)
		assert.Contains(t, output, "module.exports")
		// Should still have valid structure even with no tokens
	})
}

func TestTailwindGenerator_FormatFontFamily(t *testing.T) {
	t.Run("should format single font", func(t *testing.T) {
		// Given
		gen := NewTailwindGenerator()
		value := "Helvetica"

		// When
		result := gen.formatFontFamily(value)

		// Then
		assert.Equal(t, "['Helvetica']", result)
	})

	t.Run("should format multiple fonts", func(t *testing.T) {
		// Given
		gen := NewTailwindGenerator()
		value := "Helvetica, Arial, sans-serif"

		// When
		result := gen.formatFontFamily(value)

		// Then
		assert.Equal(t, "['Helvetica', 'Arial', 'sans-serif']", result)
	})

	t.Run("should remove existing quotes", func(t *testing.T) {
		// Given
		gen := NewTailwindGenerator()
		value := "\"Helvetica Neue\", 'Arial', sans-serif"

		// When
		result := gen.formatFontFamily(value)

		// Then
		assert.Contains(t, result, "'Helvetica Neue'")
		assert.Contains(t, result, "'Arial'")
	})
}

func TestTailwindGenerator_Metadata(t *testing.T) {
	t.Run("should return correct name", func(t *testing.T) {
		// Given
		gen := NewTailwindGenerator()

		// When
		name := gen.Name()

		// Then
		assert.Contains(t, strings.ToLower(name), "tailwind")
	})

	t.Run("should return supported formats", func(t *testing.T) {
		// Given
		gen := NewTailwindGenerator()

		// When
		formats := gen.SupportedFormats()

		// Then
		assert.Contains(t, formats, "tailwind")
	})
}
