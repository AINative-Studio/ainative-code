package design

import (
	"io"
	"strings"
	"testing"
)

func TestExtractorOptions(t *testing.T) {
	t.Run("creates extractor with default options", func(t *testing.T) {
		extractor := NewExtractor(nil)

		if extractor == nil {
			t.Fatal("expected extractor to be created")
		}

		if !extractor.enableValidation {
			t.Error("expected validation to be enabled by default")
		}
	})

	t.Run("creates extractor with custom options", func(t *testing.T) {
		opts := &ExtractorOptions{
			EnableValidation: false,
			IncludeComments:  false,
			InferCategories:  false,
		}

		extractor := NewExtractor(opts)

		if extractor.enableValidation {
			t.Error("expected validation to be disabled")
		}
	})
}

func TestSCSSParser_Parse(t *testing.T) {
	t.Run("extracts color variables from SCSS", func(t *testing.T) {
		scss := `
$primary-color: #3490dc;
$secondary-color: rgb(255, 99, 71);
$accent-color: hsl(120, 100%, 50%);
`
		parser := NewSCSSParser()
		result, err := parser.Parse(strings.NewReader(scss))

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result.Tokens) != 3 {
			t.Errorf("expected 3 tokens, got %d", len(result.Tokens))
		}

		// Check first token
		if result.Tokens[0].Name != "primary-color" {
			t.Errorf("expected token name 'primary-color', got '%s'", result.Tokens[0].Name)
		}

		if result.Tokens[0].Type != TokenTypeColor {
			t.Errorf("expected token type 'color', got '%s'", result.Tokens[0].Type)
		}

		if result.Tokens[0].Value != "#3490dc" {
			t.Errorf("expected token value '#3490dc', got '%s'", result.Tokens[0].Value)
		}
	})

	t.Run("extracts spacing variables from SCSS", func(t *testing.T) {
		scss := `
$spacing-xs: 0.25rem;
$spacing-sm: 0.5rem;
$spacing-md: 1rem;
$margin-large: 2rem;
$padding-small: 0.5rem;
`
		parser := NewSCSSParser()
		result, err := parser.Parse(strings.NewReader(scss))

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result.Tokens) != 5 {
			t.Errorf("expected 5 tokens, got %d", len(result.Tokens))
		}

		for _, token := range result.Tokens {
			if token.Type != TokenTypeSpacing {
				t.Errorf("expected token type 'spacing', got '%s'", token.Type)
			}
		}
	})

	t.Run("extracts typography variables from SCSS", func(t *testing.T) {
		scss := `
$font-family-base: 'Inter', sans-serif;
$font-size-base: 16px;
$line-height-base: 1.5;
`
		parser := NewSCSSParser()
		result, err := parser.Parse(strings.NewReader(scss))

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result.Tokens) != 3 {
			t.Errorf("expected 3 tokens, got %d", len(result.Tokens))
		}

		for _, token := range result.Tokens {
			if token.Type != TokenTypeTypography {
				t.Errorf("expected token type 'typography', got '%s' for token '%s'", token.Type, token.Name)
			}
		}
	})

	t.Run("extracts shadow variables from SCSS", func(t *testing.T) {
		scss := `
$shadow-sm: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
$shadow-md: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
`
		parser := NewSCSSParser()
		result, err := parser.Parse(strings.NewReader(scss))

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result.Tokens) != 2 {
			t.Errorf("expected 2 tokens, got %d", len(result.Tokens))
		}

		for _, token := range result.Tokens {
			if token.Type != TokenTypeShadow {
				t.Errorf("expected token type 'shadow', got '%s'", token.Type)
			}
		}
	})

	t.Run("extracts border-radius variables from SCSS", func(t *testing.T) {
		scss := `
$border-radius-sm: 0.25rem;
$border-radius-md: 0.5rem;
$rounded-full: 9999px;
`
		parser := NewSCSSParser()
		result, err := parser.Parse(strings.NewReader(scss))

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result.Tokens) != 3 {
			t.Errorf("expected 3 tokens, got %d", len(result.Tokens))
		}

		for _, token := range result.Tokens {
			if token.Type != TokenTypeBorderRadius {
				t.Errorf("expected token type 'border-radius', got '%s'", token.Type)
			}
		}
	})

	t.Run("skips comments", func(t *testing.T) {
		scss := `
// This is a comment
$primary-color: #3490dc;
/* Multi-line
   comment */
$secondary-color: #ff6347;
`
		parser := NewSCSSParser()
		result, err := parser.Parse(strings.NewReader(scss))

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result.Tokens) != 2 {
			t.Errorf("expected 2 tokens (comments should be skipped), got %d", len(result.Tokens))
		}
	})

	t.Run("skips variable references", func(t *testing.T) {
		scss := `
$base-color: #3490dc;
$primary-color: $base-color;
`
		parser := NewSCSSParser()
		result, err := parser.Parse(strings.NewReader(scss))

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should only extract the base-color, not the reference
		if len(result.Tokens) != 1 {
			t.Errorf("expected 1 token, got %d", len(result.Tokens))
		}

		if result.Tokens[0].Name != "base-color" {
			t.Errorf("expected token name 'base-color', got '%s'", result.Tokens[0].Name)
		}
	})

	t.Run("extracts from SCSS maps", func(t *testing.T) {
		scss := `
$colors: (
  primary: #3490dc,
  secondary: #ff6347,
  accent: #38c172
);
`
		parser := NewSCSSParser()
		result, err := parser.Parse(strings.NewReader(scss))

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result.Tokens) != 3 {
			t.Errorf("expected 3 tokens from map, got %d", len(result.Tokens))
		}

		// Check that tokens have map name prefix
		expectedNames := []string{"colors-primary", "colors-secondary", "colors-accent"}
		for i, token := range result.Tokens {
			if token.Name != expectedNames[i] {
				t.Errorf("expected token name '%s', got '%s'", expectedNames[i], token.Name)
			}
		}
	})
}

func TestLESSParser_Parse(t *testing.T) {
	t.Run("extracts color variables from LESS", func(t *testing.T) {
		less := `
@primary-color: #3490dc;
@secondary-color: rgb(255, 99, 71);
@accent-color: hsl(120, 100%, 50%);
`
		parser := NewLESSParser()
		result, err := parser.Parse(strings.NewReader(less))

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result.Tokens) != 3 {
			t.Errorf("expected 3 tokens, got %d", len(result.Tokens))
		}

		// Check first token
		if result.Tokens[0].Name != "primary-color" {
			t.Errorf("expected token name 'primary-color', got '%s'", result.Tokens[0].Name)
		}

		if result.Tokens[0].Type != TokenTypeColor {
			t.Errorf("expected token type 'color', got '%s'", result.Tokens[0].Type)
		}
	})

	t.Run("extracts spacing variables from LESS", func(t *testing.T) {
		less := `
@spacing-xs: 0.25rem;
@spacing-sm: 0.5rem;
@margin-large: 2rem;
`
		parser := NewLESSParser()
		result, err := parser.Parse(strings.NewReader(less))

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result.Tokens) != 3 {
			t.Errorf("expected 3 tokens, got %d", len(result.Tokens))
		}

		for _, token := range result.Tokens {
			if token.Type != TokenTypeSpacing {
				t.Errorf("expected token type 'spacing', got '%s'", token.Type)
			}
		}
	})

	t.Run("skips variable references", func(t *testing.T) {
		less := `
@base-color: #3490dc;
@primary-color: @base-color;
`
		parser := NewLESSParser()
		result, err := parser.Parse(strings.NewReader(less))

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should only extract the base-color, not the reference
		if len(result.Tokens) != 1 {
			t.Errorf("expected 1 token, got %d", len(result.Tokens))
		}
	})
}

func TestCSSParser_Parse(t *testing.T) {
	t.Run("extracts CSS custom properties", func(t *testing.T) {
		css := `
:root {
  --primary-color: #3490dc;
  --spacing-base: 1rem;
  --font-family: 'Inter', sans-serif;
}
`
		parser := NewCSSParser()
		result, err := parser.Parse(strings.NewReader(css))

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result.Tokens) != 3 {
			t.Errorf("expected 3 tokens, got %d", len(result.Tokens))
		}
	})

	t.Run("extracts from CSS properties", func(t *testing.T) {
		css := `
.button {
  color: #3490dc;
  background-color: #ffffff;
  font-size: 16px;
  padding: 1rem;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
  border-radius: 0.5rem;
}
`
		parser := NewCSSParser()
		result, err := parser.Parse(strings.NewReader(css))

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should extract color, background-color, font-size, padding, box-shadow, border-radius
		if len(result.Tokens) < 6 {
			t.Errorf("expected at least 6 tokens, got %d", len(result.Tokens))
		}
	})
}

func TestTokenCategorizer(t *testing.T) {
	categorizer := NewTokenCategorizer()

	t.Run("categorizes color tokens", func(t *testing.T) {
		token := &Token{
			Name:  "primary-500",
			Type:  TokenTypeColor,
			Value: "#3490dc",
		}

		categorizer.Categorize(token)

		if token.Category != "colors-primary" {
			t.Errorf("expected category 'colors-primary', got '%s'", token.Category)
		}
	})

	t.Run("categorizes spacing tokens", func(t *testing.T) {
		token := &Token{
			Name:  "md-spacing",
			Type:  TokenTypeSpacing,
			Value: "1rem",
		}

		categorizer.Categorize(token)

		if token.Category != "spacing-md" {
			t.Errorf("expected category 'spacing-md', got '%s'", token.Category)
		}
	})

	t.Run("categorizes typography tokens", func(t *testing.T) {
		token := &Token{
			Name:  "heading-large",
			Type:  TokenTypeTypography,
			Value: "2rem",
		}

		categorizer.Categorize(token)

		if token.Category != "typography-heading" {
			t.Errorf("expected category 'typography-heading', got '%s'", token.Category)
		}
	})

	t.Run("does not override existing category", func(t *testing.T) {
		token := &Token{
			Name:     "custom",
			Type:     TokenTypeColor,
			Value:    "#3490dc",
			Category: "existing-category",
		}

		categorizer.Categorize(token)

		if token.Category != "existing-category" {
			t.Errorf("expected category to remain 'existing-category', got '%s'", token.Category)
		}
	})
}

func TestDeduplicateTokens(t *testing.T) {
	t.Run("removes duplicate tokens", func(t *testing.T) {
		tokens := []Token{
			{Name: "primary", Type: TokenTypeColor, Value: "#3490dc"},
			{Name: "primary", Type: TokenTypeColor, Value: "#ff0000"},
			{Name: "secondary", Type: TokenTypeColor, Value: "#ff6347"},
		}

		result := deduplicateTokens(tokens)

		if len(result) != 2 {
			t.Errorf("expected 2 unique tokens, got %d", len(result))
		}

		// Should keep the first occurrence
		if result[0].Value != "#3490dc" {
			t.Errorf("expected first occurrence to be kept, got value '%s'", result[0].Value)
		}
	})

	t.Run("keeps tokens with same name but different type", func(t *testing.T) {
		tokens := []Token{
			{Name: "base", Type: TokenTypeColor, Value: "#3490dc"},
			{Name: "base", Type: TokenTypeSpacing, Value: "1rem"},
		}

		result := deduplicateTokens(tokens)

		if len(result) != 2 {
			t.Errorf("expected 2 tokens (different types), got %d", len(result))
		}
	})
}

func TestExtractor_Extract(t *testing.T) {
	t.Run("validates tokens when enabled", func(t *testing.T) {
		extractor := NewExtractor(&ExtractorOptions{
			EnableValidation: true,
		})

		scss := `
$INVALID-NAME: #3490dc;
$valid-color: invalid-color-value;
`
		parser := NewSCSSParser()
		result, err := extractor.Extract(strings.NewReader(scss), parser)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should have warnings for invalid tokens (uppercase name and invalid color value)
		if len(result.Warnings) == 0 {
			t.Error("expected validation warnings")
		}
	})

	t.Run("skips validation when disabled", func(t *testing.T) {
		extractor := NewExtractor(&ExtractorOptions{
			EnableValidation: false,
		})

		scss := `
$invalid-name: #3490dc;
$valid-color: invalid-color-value;
`
		parser := NewSCSSParser()
		result, err := extractor.Extract(strings.NewReader(scss), parser)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should not have warnings when validation is disabled
		if len(result.Warnings) > 0 {
			t.Error("expected no validation warnings when validation is disabled")
		}
	})

	t.Run("categorizes tokens", func(t *testing.T) {
		extractor := NewExtractor(&ExtractorOptions{
			InferCategories: true,
		})

		scss := `
$primary-500: #3490dc;
$md-spacing: 1rem;
`
		parser := NewSCSSParser()
		result, err := extractor.Extract(strings.NewReader(scss), parser)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		for _, token := range result.Tokens {
			if token.Category == "" || token.Category == "uncategorized" {
				t.Errorf("expected token '%s' to be categorized", token.Name)
			}
		}
	})

	t.Run("deduplicates tokens", func(t *testing.T) {
		extractor := NewExtractor(nil)

		scss := `
$primary: #3490dc;
$primary: #3490dc;
$secondary: #ff6347;
`
		parser := NewSCSSParser()
		result, err := extractor.Extract(strings.NewReader(scss), parser)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result.Tokens) != 2 {
			t.Errorf("expected 2 unique tokens, got %d", len(result.Tokens))
		}
	})
}

func TestExtractor_ExtractFromFile(t *testing.T) {
	t.Run("detects file type from extension", func(t *testing.T) {
		// Mock the openFile function for testing
		originalOpenFile := openFile
		defer func() { openFile = originalOpenFile }()

		openFile = func(path string) (io.ReadCloser, error) {
			content := "$primary: #3490dc;"
			return io.NopCloser(strings.NewReader(content)), nil
		}

		extractor := NewExtractor(nil)

		tests := []struct {
			filename string
			wantErr  bool
		}{
			{"styles.css", false},
			{"variables.scss", false},
			{"theme.sass", false},
			{"styles.less", false},
			{"unknown.txt", true},
		}

		for _, tt := range tests {
			t.Run(tt.filename, func(t *testing.T) {
				_, err := extractor.ExtractFromFile(tt.filename)
				if (err != nil) != tt.wantErr {
					t.Errorf("ExtractFromFile() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
	})
}

func TestSCSSParser_InferTokenType(t *testing.T) {
	parser := NewSCSSParser()

	tests := []struct {
		name      string
		varName   string
		value     string
		wantType  TokenType
	}{
		{"color by name", "primary-color", "#3490dc", TokenTypeColor},
		{"color by value", "brand", "#ff6347", TokenTypeColor},
		{"typography by name", "font-base", "16px", TokenTypeTypography},
		{"spacing by name", "margin-sm", "0.5rem", TokenTypeSpacing},
		{"shadow by name", "shadow-lg", "0 10px 15px rgba(0,0,0,0.1)", TokenTypeShadow},
		{"border-radius by name", "rounded-md", "0.5rem", TokenTypeBorderRadius},
		{"unknown type", "unknown", "somevalue", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.inferTokenType(tt.varName, tt.value)
			if got != tt.wantType {
				t.Errorf("inferTokenType() = %v, want %v", got, tt.wantType)
			}
		})
	}
}

func TestSCSSParser_ExtractTokensFromProperties(t *testing.T) {
	parser := NewSCSSParser()

	t.Run("extracts color from CSS properties", func(t *testing.T) {
		tokens := parser.extractTokensFromProperty("color", "#3490dc", 1)
		if len(tokens) != 1 {
			t.Errorf("expected 1 token, got %d", len(tokens))
		}
		if tokens[0].Type != TokenTypeColor {
			t.Errorf("expected token type 'color', got '%s'", tokens[0].Type)
		}
	})

	t.Run("extracts typography from CSS properties", func(t *testing.T) {
		tokens := parser.extractTokensFromProperty("font-size", "16px", 1)
		if len(tokens) != 1 {
			t.Errorf("expected 1 token, got %d", len(tokens))
		}
		if tokens[0].Type != TokenTypeTypography {
			t.Errorf("expected token type 'typography', got '%s'", tokens[0].Type)
		}
	})

	t.Run("extracts spacing from CSS properties", func(t *testing.T) {
		tokens := parser.extractTokensFromProperty("padding", "1rem", 1)
		if len(tokens) != 1 {
			t.Errorf("expected 1 token, got %d", len(tokens))
		}
		if tokens[0].Type != TokenTypeSpacing {
			t.Errorf("expected token type 'spacing', got '%s'", tokens[0].Type)
		}
	})

	t.Run("extracts shadow from CSS properties", func(t *testing.T) {
		tokens := parser.extractTokensFromProperty("box-shadow", "0 2px 4px rgba(0,0,0,0.1)", 1)
		if len(tokens) != 1 {
			t.Errorf("expected 1 token, got %d", len(tokens))
		}
		if tokens[0].Type != TokenTypeShadow {
			t.Errorf("expected token type 'shadow', got '%s'", tokens[0].Type)
		}
	})

	t.Run("extracts border-radius from CSS properties", func(t *testing.T) {
		tokens := parser.extractTokensFromProperty("border-radius", "0.5rem", 1)
		if len(tokens) != 1 {
			t.Errorf("expected 1 token, got %d", len(tokens))
		}
		if tokens[0].Type != TokenTypeBorderRadius {
			t.Errorf("expected token type 'border-radius', got '%s'", tokens[0].Type)
		}
	})
}

func TestLESSParser_ExtractTokensFromProperties(t *testing.T) {
	parser := NewLESSParser()

	t.Run("extracts color from CSS properties", func(t *testing.T) {
		tokens := parser.extractTokensFromProperty("background-color", "#ffffff", 1)
		if len(tokens) != 1 {
			t.Errorf("expected 1 token, got %d", len(tokens))
		}
		if tokens[0].Type != TokenTypeColor {
			t.Errorf("expected token type 'color', got '%s'", tokens[0].Type)
		}
	})

	t.Run("extracts typography from CSS properties", func(t *testing.T) {
		tokens := parser.extractTokensFromProperty("line-height", "1.5", 1)
		if len(tokens) != 1 {
			t.Errorf("expected 1 token, got %d", len(tokens))
		}
		if tokens[0].Type != TokenTypeTypography {
			t.Errorf("expected token type 'typography', got '%s'", tokens[0].Type)
		}
	})
}
