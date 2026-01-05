package generators

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/AINative-studio/ainative-code/internal/design"
)

func TestNewSCSSGenerator(t *testing.T) {
	t.Run("should create SCSS generator", func(t *testing.T) {
		// When
		gen := NewSCSSGenerator()

		// Then
		assert.NotNil(t, gen)
		assert.NotNil(t, gen.engine)
	})
}

func TestSCSSGenerator_Generate(t *testing.T) {
	t.Run("should generate SCSS variables", func(t *testing.T) {
		// Given
		gen := NewSCSSGenerator()
		tokens := []design.Token{
			{Name: "primary-color", Type: "color", Value: "#007bff"},
			{Name: "secondary-color", Type: "color", Value: "#6c757d"},
			{Name: "spacing-base", Type: "spacing", Value: "16px"},
		}

		// When
		output, err := gen.Generate(tokens)

		// Then
		require.NoError(t, err)
		assert.Contains(t, output, "$primary-color: #007bff;")
		assert.Contains(t, output, "$secondary-color: #6c757d;")
		assert.Contains(t, output, "$spacing-base: 16px;")
	})

	t.Run("should handle empty tokens", func(t *testing.T) {
		// Given
		gen := NewSCSSGenerator()
		tokens := []design.Token{}

		// When
		output, err := gen.Generate(tokens)

		// Then
		require.NoError(t, err)
		// Should still generate valid output
		assert.NotEmpty(t, output)
	})

	t.Run("should include comment header", func(t *testing.T) {
		// Given
		gen := NewSCSSGenerator()
		tokens := []design.Token{
			{Name: "test", Type: "color", Value: "#000"},
		}

		// When
		output, err := gen.Generate(tokens)

		// Then
		require.NoError(t, err)
		assert.Contains(t, output, "Design Tokens")
		assert.Contains(t, output, "SCSS Variables")
	})

	t.Run("should handle complex values", func(t *testing.T) {
		// Given
		gen := NewSCSSGenerator()
		tokens := []design.Token{
			{Name: "font-stack", Type: "typography", Value: "Helvetica, Arial, sans-serif"},
			{Name: "box-shadow", Type: "shadow", Value: "0 2px 4px rgba(0,0,0,0.1)"},
		}

		// When
		output, err := gen.Generate(tokens)

		// Then
		require.NoError(t, err)
		assert.Contains(t, output, "$font-stack: Helvetica, Arial, sans-serif;")
		assert.Contains(t, output, "$box-shadow: 0 2px 4px rgba(0,0,0,0.1);")
	})
}

func TestSCSSGenerator_Metadata(t *testing.T) {
	t.Run("should return correct name", func(t *testing.T) {
		// Given
		gen := NewSCSSGenerator()

		// When
		name := gen.Name()

		// Then
		assert.Contains(t, strings.ToLower(name), "scss")
	})

	t.Run("should return supported formats", func(t *testing.T) {
		// Given
		gen := NewSCSSGenerator()

		// When
		formats := gen.SupportedFormats()

		// Then
		assert.Contains(t, formats, "scss")
		assert.Contains(t, formats, "sass")
	})
}
