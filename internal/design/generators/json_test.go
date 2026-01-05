package generators

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/AINative-studio/ainative-code/internal/design"
)

func TestNewJSONGenerator(t *testing.T) {
	t.Run("should create JSON generator with pretty printing", func(t *testing.T) {
		// When
		gen := NewJSONGenerator(true)

		// Then
		assert.NotNil(t, gen)
		assert.True(t, gen.pretty)
	})

	t.Run("should create JSON generator without pretty printing", func(t *testing.T) {
		// When
		gen := NewJSONGenerator(false)

		// Then
		assert.NotNil(t, gen)
		assert.False(t, gen.pretty)
	})
}

func TestJSONGenerator_Generate(t *testing.T) {
	t.Run("should generate valid JSON", func(t *testing.T) {
		// Given
		gen := NewJSONGenerator(true)
		tokens := []design.Token{
			{Name: "primary-color", Type: "color", Value: "#007bff"},
			{Name: "spacing-base", Type: "spacing", Value: "16px"},
		}

		// When
		output, err := gen.Generate(tokens)

		// Then
		require.NoError(t, err)

		// Should be valid JSON
		var result map[string]interface{}
		err = json.Unmarshal([]byte(output), &result)
		require.NoError(t, err)

		// Should contain tokens
		assert.Contains(t, result, "tokens")
		assert.Contains(t, result, "count")
		assert.Equal(t, float64(2), result["count"])
	})

	t.Run("should generate pretty-printed JSON", func(t *testing.T) {
		// Given
		gen := NewJSONGenerator(true)
		tokens := []design.Token{
			{Name: "test", Type: "color", Value: "#000"},
		}

		// When
		output, err := gen.Generate(tokens)

		// Then
		require.NoError(t, err)
		// Pretty-printed JSON should have indentation
		assert.Contains(t, output, "\n")
		assert.Contains(t, output, "  ")
	})

	t.Run("should generate minified JSON", func(t *testing.T) {
		// Given
		gen := NewJSONGenerator(false)
		tokens := []design.Token{
			{Name: "test", Type: "color", Value: "#000"},
		}

		// When
		output, err := gen.Generate(tokens)

		// Then
		require.NoError(t, err)
		// Minified JSON should not have extra whitespace
		assert.NotContains(t, output, "\n  ")
	})

	t.Run("should handle empty tokens", func(t *testing.T) {
		// Given
		gen := NewJSONGenerator(true)
		tokens := []design.Token{}

		// When
		output, err := gen.Generate(tokens)

		// Then
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal([]byte(output), &result)
		require.NoError(t, err)

		assert.Equal(t, float64(0), result["count"])
	})

	t.Run("should preserve token structure", func(t *testing.T) {
		// Given
		gen := NewJSONGenerator(true)
		tokens := []design.Token{
			{
				Name:        "primary",
				Type:        "color",
				Value:       "#007bff",
				Category:    "colors",
				Description: "Primary brand color",
			},
		}

		// When
		output, err := gen.Generate(tokens)

		// Then
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal([]byte(output), &result)
		require.NoError(t, err)

		tokensArray := result["tokens"].([]interface{})
		token := tokensArray[0].(map[string]interface{})

		assert.Equal(t, "primary", token["name"])
		assert.Equal(t, "color", token["type"])
		assert.Equal(t, "#007bff", token["value"])
		assert.Equal(t, "colors", token["category"])
		assert.Equal(t, "Primary brand color", token["description"])
	})
}

func TestJSONGenerator_Metadata(t *testing.T) {
	t.Run("should return correct name", func(t *testing.T) {
		// Given
		gen := NewJSONGenerator(true)

		// When
		name := gen.Name()

		// Then
		assert.Contains(t, strings.ToLower(name), "json")
	})

	t.Run("should return supported formats", func(t *testing.T) {
		// Given
		gen := NewJSONGenerator(true)

		// When
		formats := gen.SupportedFormats()

		// Then
		assert.Contains(t, formats, "json")
	})
}
