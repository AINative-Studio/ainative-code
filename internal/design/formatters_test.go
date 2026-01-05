package design

import (
	"encoding/json"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestJSONFormatter_Format(t *testing.T) {
	tokens := []Token{
		{
			Name:  "primary-color",
			Type:  TokenTypeColor,
			Value: "#3490dc",
			Metadata: map[string]string{
				"source": "scss-variable",
			},
		},
		{
			Name:  "spacing-base",
			Type:  TokenTypeSpacing,
			Value: "1rem",
		},
	}

	t.Run("formats tokens as pretty JSON", func(t *testing.T) {
		formatter := NewJSONFormatter(true)
		output, err := formatter.Format(tokens)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Check that output is valid JSON
		var collection TokenCollection
		if err := json.Unmarshal([]byte(output), &collection); err != nil {
			t.Fatalf("output is not valid JSON: %v", err)
		}

		if len(collection.Tokens) != 2 {
			t.Errorf("expected 2 tokens in output, got %d", len(collection.Tokens))
		}

		// Check that it's pretty printed (has indentation)
		if !strings.Contains(output, "  ") {
			t.Error("expected output to be pretty printed with indentation")
		}
	})

	t.Run("formats tokens as compact JSON", func(t *testing.T) {
		formatter := NewJSONFormatter(false)
		output, err := formatter.Format(tokens)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Check that output is valid JSON
		var collection TokenCollection
		if err := json.Unmarshal([]byte(output), &collection); err != nil {
			t.Fatalf("output is not valid JSON: %v", err)
		}

		// Check that it's compact (no extra whitespace)
		if strings.Contains(output, "\n  ") {
			t.Error("expected output to be compact without indentation")
		}
	})

	t.Run("includes metadata in output", func(t *testing.T) {
		formatter := NewJSONFormatter(true)
		output, err := formatter.Format(tokens)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var collection TokenCollection
		json.Unmarshal([]byte(output), &collection)

		if collection.Metadata == nil {
			t.Error("expected metadata to be included")
		}

		if collection.Metadata["version"] == "" {
			t.Error("expected version in metadata")
		}
	})
}

func TestYAMLFormatter_Format(t *testing.T) {
	tokens := []Token{
		{
			Name:  "primary-color",
			Type:  TokenTypeColor,
			Value: "#3490dc",
		},
		{
			Name:  "spacing-base",
			Type:  TokenTypeSpacing,
			Value: "1rem",
		},
	}

	t.Run("formats tokens as YAML", func(t *testing.T) {
		formatter := NewYAMLFormatter()
		output, err := formatter.Format(tokens)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Check that output is valid YAML
		var collection TokenCollection
		if err := yaml.Unmarshal([]byte(output), &collection); err != nil {
			t.Fatalf("output is not valid YAML: %v", err)
		}

		if len(collection.Tokens) != 2 {
			t.Errorf("expected 2 tokens in output, got %d", len(collection.Tokens))
		}
	})

	t.Run("includes metadata in YAML output", func(t *testing.T) {
		formatter := NewYAMLFormatter()
		output, err := formatter.Format(tokens)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var collection TokenCollection
		yaml.Unmarshal([]byte(output), &collection)

		if collection.Metadata == nil {
			t.Error("expected metadata to be included")
		}
	})
}

func TestTailwindFormatter_Format(t *testing.T) {
	tokens := []Token{
		{Name: "primary", Type: TokenTypeColor, Value: "#3490dc"},
		{Name: "secondary", Type: TokenTypeColor, Value: "#ff6347"},
		{Name: "spacing-sm", Type: TokenTypeSpacing, Value: "0.5rem"},
		{Name: "spacing-md", Type: TokenTypeSpacing, Value: "1rem"},
		{Name: "font-sans", Type: TokenTypeTypography, Value: "Inter, sans-serif"},
		{Name: "font-size-base", Type: TokenTypeTypography, Value: "16px"},
		{Name: "shadow-sm", Type: TokenTypeShadow, Value: "0 1px 2px rgba(0,0,0,0.05)"},
		{Name: "border-radius-md", Type: TokenTypeBorderRadius, Value: "0.5rem"},
	}

	t.Run("formats tokens as Tailwind config", func(t *testing.T) {
		formatter := NewTailwindFormatter(false)
		output, err := formatter.Format(tokens)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Check for Tailwind config structure
		if !strings.Contains(output, "module.exports") {
			t.Error("expected output to contain 'module.exports'")
		}

		if !strings.Contains(output, "theme:") {
			t.Error("expected output to contain 'theme:'")
		}

		if !strings.Contains(output, "extend:") {
			t.Error("expected output to contain 'extend:'")
		}
	})

	t.Run("includes colors section", func(t *testing.T) {
		formatter := NewTailwindFormatter(false)
		output, err := formatter.Format(tokens)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.Contains(output, "colors:") {
			t.Error("expected output to contain 'colors:' section")
		}

		if !strings.Contains(output, "'primary': '#3490dc'") {
			t.Error("expected output to contain primary color")
		}
	})

	t.Run("includes spacing section", func(t *testing.T) {
		formatter := NewTailwindFormatter(false)
		output, err := formatter.Format(tokens)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.Contains(output, "spacing:") {
			t.Error("expected output to contain 'spacing:' section")
		}

		if !strings.Contains(output, "'sm': '0.5rem'") {
			t.Error("expected output to contain spacing token")
		}
	})

	t.Run("includes boxShadow section", func(t *testing.T) {
		formatter := NewTailwindFormatter(false)
		output, err := formatter.Format(tokens)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.Contains(output, "boxShadow:") {
			t.Error("expected output to contain 'boxShadow:' section")
		}
	})

	t.Run("includes borderRadius section", func(t *testing.T) {
		formatter := NewTailwindFormatter(false)
		output, err := formatter.Format(tokens)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.Contains(output, "borderRadius:") {
			t.Error("expected output to contain 'borderRadius:' section")
		}
	})

	t.Run("includes comments when enabled", func(t *testing.T) {
		formatter := NewTailwindFormatter(true)
		output, err := formatter.Format(tokens)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.Contains(output, "// Extracted") {
			t.Error("expected output to contain comments")
		}
	})

	t.Run("excludes comments when disabled", func(t *testing.T) {
		formatter := NewTailwindFormatter(false)
		output, err := formatter.Format(tokens)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if strings.Contains(output, "// Extracted") {
			t.Error("expected output to not contain comments")
		}
	})

	t.Run("strips prefixes from token names", func(t *testing.T) {
		formatter := NewTailwindFormatter(false)
		output, err := formatter.Format(tokens)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should strip 'spacing-' prefix
		if !strings.Contains(output, "'sm':") {
			t.Error("expected 'spacing-' prefix to be stripped")
		}

		// Should strip 'border-radius-' prefix
		if !strings.Contains(output, "'md':") {
			t.Error("expected 'border-radius-' prefix to be stripped")
		}
	})
}

func TestStyleDictionaryFormatter_Format(t *testing.T) {
	tokens := []Token{
		{
			Name:     "primary-500",
			Type:     TokenTypeColor,
			Value:    "#3490dc",
			Category: "colors",
		},
		{
			Name:     "spacing-md",
			Type:     TokenTypeSpacing,
			Value:    "1rem",
			Category: "spacing",
		},
	}

	t.Run("formats tokens in Style Dictionary format", func(t *testing.T) {
		formatter := NewStyleDictionaryFormatter(true)
		output, err := formatter.Format(tokens)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Check that output is valid JSON
		var result map[string]map[string]interface{}
		if err := json.Unmarshal([]byte(output), &result); err != nil {
			t.Fatalf("output is not valid JSON: %v", err)
		}

		// Check for grouped categories
		if result["colors"] == nil {
			t.Error("expected 'colors' category in output")
		}

		if result["spacing"] == nil {
			t.Error("expected 'spacing' category in output")
		}
	})

	t.Run("includes token type in output", func(t *testing.T) {
		formatter := NewStyleDictionaryFormatter(true)
		output, err := formatter.Format(tokens)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var result map[string]map[string]interface{}
		json.Unmarshal([]byte(output), &result)

		colorToken := result["colors"]["primary-500"].(map[string]interface{})
		if colorToken["type"] != "color" {
			t.Errorf("expected type 'color', got '%v'", colorToken["type"])
		}
	})
}

func TestCSSVariablesFormatter_Format(t *testing.T) {
	tokens := []Token{
		{
			Name:     "primary-color",
			Type:     TokenTypeColor,
			Value:    "#3490dc",
			Category: "colors",
		},
		{
			Name:     "spacing-base",
			Type:     TokenTypeSpacing,
			Value:    "1rem",
			Category: "spacing",
		},
	}

	t.Run("formats tokens as CSS custom properties", func(t *testing.T) {
		formatter := NewCSSVariablesFormatter("token")
		output, err := formatter.Format(tokens)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.Contains(output, ":root {") {
			t.Error("expected output to contain ':root {' selector")
		}

		if !strings.Contains(output, "--token-primary-color: #3490dc;") {
			t.Error("expected output to contain CSS variable for primary-color")
		}

		if !strings.Contains(output, "--token-spacing-base: 1rem;") {
			t.Error("expected output to contain CSS variable for spacing-base")
		}
	})

	t.Run("uses custom prefix", func(t *testing.T) {
		formatter := NewCSSVariablesFormatter("custom")
		output, err := formatter.Format(tokens)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.Contains(output, "--custom-") {
			t.Error("expected output to use 'custom' prefix")
		}
	})

	t.Run("groups tokens by category with comments", func(t *testing.T) {
		formatter := NewCSSVariablesFormatter("token")
		output, err := formatter.Format(tokens)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.Contains(output, "/* colors */") {
			t.Error("expected output to contain category comments")
		}

		if !strings.Contains(output, "/* spacing */") {
			t.Error("expected output to contain category comments")
		}
	})
}

func TestSCSSVariablesFormatter_Format(t *testing.T) {
	tokens := []Token{
		{
			Name:        "primary-color",
			Type:        TokenTypeColor,
			Value:       "#3490dc",
			Category:    "colors",
			Description: "Primary brand color",
		},
		{
			Name:     "spacing-base",
			Type:     TokenTypeSpacing,
			Value:    "1rem",
			Category: "spacing",
		},
	}

	t.Run("formats tokens as SCSS variables", func(t *testing.T) {
		formatter := NewSCSSVariablesFormatter()
		output, err := formatter.Format(tokens)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.Contains(output, "$primary-color: #3490dc;") {
			t.Error("expected output to contain SCSS variable for primary-color")
		}

		if !strings.Contains(output, "$spacing-base: 1rem;") {
			t.Error("expected output to contain SCSS variable for spacing-base")
		}
	})

	t.Run("includes category comments", func(t *testing.T) {
		formatter := NewSCSSVariablesFormatter()
		output, err := formatter.Format(tokens)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.Contains(output, "// colors") {
			t.Error("expected output to contain category comments")
		}

		if !strings.Contains(output, "// spacing") {
			t.Error("expected output to contain category comments")
		}
	})

	t.Run("includes token descriptions", func(t *testing.T) {
		formatter := NewSCSSVariablesFormatter()
		output, err := formatter.Format(tokens)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.Contains(output, "// Primary brand color") {
			t.Error("expected output to contain token description")
		}
	})
}

func TestGetFormatter(t *testing.T) {
	t.Run("returns JSON formatter", func(t *testing.T) {
		formatter, err := GetFormatter(OutputFormatJSON, map[string]interface{}{
			"pretty": true,
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if _, ok := formatter.(*JSONFormatter); !ok {
			t.Error("expected JSONFormatter to be returned")
		}
	})

	t.Run("returns YAML formatter", func(t *testing.T) {
		formatter, err := GetFormatter(OutputFormatYAML, map[string]interface{}{})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if _, ok := formatter.(*YAMLFormatter); !ok {
			t.Error("expected YAMLFormatter to be returned")
		}
	})

	t.Run("returns Tailwind formatter", func(t *testing.T) {
		formatter, err := GetFormatter(OutputFormatTailwind, map[string]interface{}{
			"includeComments": true,
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if _, ok := formatter.(*TailwindFormatter); !ok {
			t.Error("expected TailwindFormatter to be returned")
		}
	})

	t.Run("returns error for unsupported format", func(t *testing.T) {
		_, err := GetFormatter("unsupported", map[string]interface{}{})

		if err == nil {
			t.Error("expected error for unsupported format")
		}
	})
}

func TestFilterTokensByType(t *testing.T) {
	tokens := []Token{
		{Name: "primary", Type: TokenTypeColor, Value: "#3490dc"},
		{Name: "secondary", Type: TokenTypeColor, Value: "#ff6347"},
		{Name: "spacing-sm", Type: TokenTypeSpacing, Value: "0.5rem"},
		{Name: "shadow-sm", Type: TokenTypeShadow, Value: "0 1px 2px rgba(0,0,0,0.05)"},
	}

	t.Run("filters color tokens", func(t *testing.T) {
		result := filterTokensByType(tokens, TokenTypeColor)

		if len(result) != 2 {
			t.Errorf("expected 2 color tokens, got %d", len(result))
		}

		for _, token := range result {
			if token.Type != TokenTypeColor {
				t.Errorf("expected only color tokens, got type '%s'", token.Type)
			}
		}
	})

	t.Run("filters spacing tokens", func(t *testing.T) {
		result := filterTokensByType(tokens, TokenTypeSpacing)

		if len(result) != 1 {
			t.Errorf("expected 1 spacing token, got %d", len(result))
		}
	})

	t.Run("returns empty array when no tokens match", func(t *testing.T) {
		result := filterTokensByType(tokens, TokenTypeBorderRadius)

		if len(result) != 0 {
			t.Errorf("expected 0 border-radius tokens, got %d", len(result))
		}
	})
}

func TestFilterTypographyTokens(t *testing.T) {
	tokens := []Token{
		{Name: "font-family-base", Type: TokenTypeTypography, Value: "Inter"},
		{Name: "font-size-base", Type: TokenTypeTypography, Value: "16px"},
		{Name: "line-height-base", Type: TokenTypeTypography, Value: "1.5"},
		{Name: "spacing-base", Type: TokenTypeSpacing, Value: "1rem"},
	}

	t.Run("filters font-family tokens", func(t *testing.T) {
		result := filterTypographyTokens(tokens, "font-family")

		if len(result) != 1 {
			t.Errorf("expected 1 font-family token, got %d", len(result))
		}

		if result[0].Name != "font-family-base" {
			t.Errorf("expected token name 'font-family-base', got '%s'", result[0].Name)
		}
	})

	t.Run("filters font-size tokens", func(t *testing.T) {
		result := filterTypographyTokens(tokens, "font-size")

		if len(result) != 1 {
			t.Errorf("expected 1 font-size token, got %d", len(result))
		}
	})

	t.Run("does not return non-typography tokens", func(t *testing.T) {
		result := filterTypographyTokens(tokens, "spacing")

		if len(result) != 0 {
			t.Errorf("expected 0 tokens (spacing is not typography), got %d", len(result))
		}
	})
}
