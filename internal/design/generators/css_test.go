package generators

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/AINative-studio/ainative-code/internal/design"
)

func TestNewCSSGenerator(t *testing.T) {
	t.Run("should create CSS generator", func(t *testing.T) {
		// When
		gen := NewCSSGenerator()

		// Then
		assert.NotNil(t, gen)
		assert.NotNil(t, gen.engine)
	})
}

func TestCSSGenerator_Generate(t *testing.T) {
	t.Run("should generate CSS custom properties", func(t *testing.T) {
		// Given
		gen := NewCSSGenerator()
		tokens := []design.Token{
			{Name: "primary-color", Type: "color", Value: "#007bff"},
			{Name: "secondary-color", Type: "color", Value: "#6c757d"},
			{Name: "spacing-base", Type: "spacing", Value: "16px"},
		}

		// When
		output, err := gen.Generate(tokens)

		// Then
		require.NoError(t, err)
		assert.Contains(t, output, ":root {")
		assert.Contains(t, output, "--primary-color: #007bff;")
		assert.Contains(t, output, "--secondary-color: #6c757d;")
		assert.Contains(t, output, "--spacing-base: 16px;")
		assert.Contains(t, output, "}")
	})

	t.Run("should handle empty tokens", func(t *testing.T) {
		// Given
		gen := NewCSSGenerator()
		tokens := []design.Token{}

		// When
		output, err := gen.Generate(tokens)

		// Then
		require.NoError(t, err)
		assert.Contains(t, output, ":root {")
		assert.Contains(t, output, "}")
	})

	t.Run("should include comment header", func(t *testing.T) {
		// Given
		gen := NewCSSGenerator()
		tokens := []design.Token{
			{Name: "test", Type: "color", Value: "#000"},
		}

		// When
		output, err := gen.Generate(tokens)

		// Then
		require.NoError(t, err)
		assert.Contains(t, output, "Design Tokens")
		assert.Contains(t, output, "CSS Custom Properties")
	})

	t.Run("should handle special characters in names", func(t *testing.T) {
		// Given
		gen := NewCSSGenerator()
		tokens := []design.Token{
			{Name: "color.primary.500", Type: "color", Value: "#007bff"},
		}

		// When
		output, err := gen.Generate(tokens)

		// Then
		require.NoError(t, err)
		// Should convert dots to dashes
		assert.Contains(t, output, "--color-primary-500")
	})
}

func TestCSSGenerator_Metadata(t *testing.T) {
	t.Run("should return correct name", func(t *testing.T) {
		// Given
		gen := NewCSSGenerator()

		// When
		name := gen.Name()

		// Then
		assert.Contains(t, strings.ToLower(name), "css")
	})

	t.Run("should return supported formats", func(t *testing.T) {
		// Given
		gen := NewCSSGenerator()

		// When
		formats := gen.SupportedFormats()

		// Then
		assert.Contains(t, formats, "css")
	})
}
