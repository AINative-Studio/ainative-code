package design

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCSSParser_Parse_ColorTokens(t *testing.T) {
	tests := []struct {
		name     string
		css      string
		expected []Token
	}{
		{
			name: "CSS custom property with hex color",
			css: `:root {
				--primary-color: #6366F1;
			}`,
			expected: []Token{
				{
					Name:  "primary-color",
					Type:  TokenTypeColor,
					Value: "#6366F1",
					Metadata: map[string]string{
						"source": "css-variable",
						"line":   "2",
					},
				},
			},
		},
		{
			name: "CSS custom property with RGB color",
			css: `:root {
				--secondary-color: rgb(139, 92, 246);
			}`,
			expected: []Token{
				{
					Name:  "secondary-color",
					Type:  TokenTypeColor,
					Value: "rgb(139, 92, 246)",
					Metadata: map[string]string{
						"source": "css-variable",
						"line":   "2",
					},
				},
			},
		},
		{
			name: "CSS custom property with RGBA color",
			css: `:root {
				--overlay-color: rgba(0, 0, 0, 0.5);
			}`,
			expected: []Token{
				{
					Name:  "overlay-color",
					Type:  TokenTypeColor,
					Value: "rgba(0, 0, 0, 0.5)",
					Metadata: map[string]string{
						"source": "css-variable",
						"line":   "2",
					},
				},
			},
		},
		{
			name: "CSS custom property with HSL color",
			css: `:root {
				--accent-color: hsl(250, 100%, 50%);
			}`,
			expected: []Token{
				{
					Name:  "accent-color",
					Type:  TokenTypeColor,
					Value: "hsl(250, 100%, 50%)",
					Metadata: map[string]string{
						"source": "css-variable",
						"line":   "2",
					},
				},
			},
		},
		{
			name: "CSS property with color",
			css: `.button {
				background-color: #10B981;
			}`,
			expected: []Token{
				{
					Name:  "background-color",
					Type:  TokenTypeColor,
					Value: "#10B981",
					Metadata: map[string]string{
						"source": "css-property",
						"line":   "2",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewCSSParser()
			reader := strings.NewReader(tt.css)

			result, err := parser.Parse(reader)

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Len(t, result.Tokens, len(tt.expected))

			if len(result.Tokens) > 0 && len(tt.expected) > 0 {
				assert.Equal(t, tt.expected[0].Name, result.Tokens[0].Name)
				assert.Equal(t, tt.expected[0].Type, result.Tokens[0].Type)
				assert.Equal(t, tt.expected[0].Value, result.Tokens[0].Value)
			}
		})
	}
}

func TestCSSParser_Parse_TypographyTokens(t *testing.T) {
	tests := []struct {
		name     string
		css      string
		expected []Token
	}{
		{
			name: "Font family variable",
			css: `:root {
				--font-family-base: "Inter", sans-serif;
			}`,
			expected: []Token{
				{
					Name:  "font-family-base",
					Type:  TokenTypeTypography,
					Value: `"Inter", sans-serif`,
				},
			},
		},
		{
			name: "Font size variable",
			css: `:root {
				--font-size-lg: 1.125rem;
			}`,
			expected: []Token{
				{
					Name:  "font-size-lg",
					Type:  TokenTypeTypography,
					Value: "1.125rem",
				},
			},
		},
		{
			name: "Line height variable",
			css: `:root {
				--line-height-normal: 1.5;
			}`,
			expected: []Token{
				{
					Name:  "line-height-normal",
					Type:  TokenTypeTypography,
					Value: "1.5",
				},
			},
		},
		{
			name: "Typography property",
			css: `.heading {
				font-size: 2rem;
			}`,
			expected: []Token{
				{
					Name:     "font-size",
					Type:     TokenTypeTypography,
					Value:    "2rem",
					Category: "typography",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewCSSParser()
			reader := strings.NewReader(tt.css)

			result, err := parser.Parse(reader)

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Len(t, result.Tokens, len(tt.expected))

			if len(result.Tokens) > 0 && len(tt.expected) > 0 {
				assert.Equal(t, tt.expected[0].Name, result.Tokens[0].Name)
				assert.Equal(t, tt.expected[0].Type, result.Tokens[0].Type)
				assert.Equal(t, tt.expected[0].Value, result.Tokens[0].Value)
			}
		})
	}
}

func TestCSSParser_Parse_SpacingTokens(t *testing.T) {
	tests := []struct {
		name     string
		css      string
		expected []Token
	}{
		{
			name: "Spacing variable",
			css: `:root {
				--spacing-md: 1rem;
			}`,
			expected: []Token{
				{
					Name:  "spacing-md",
					Type:  TokenTypeSpacing,
					Value: "1rem",
				},
			},
		},
		{
			name: "Margin property",
			css: `.container {
				margin: 2rem;
			}`,
			expected: []Token{
				{
					Name:     "margin",
					Type:     TokenTypeSpacing,
					Value:    "2rem",
					Category: "spacing",
				},
			},
		},
		{
			name: "Padding property",
			css: `.box {
				padding: 1.5rem;
			}`,
			expected: []Token{
				{
					Name:     "padding",
					Type:     TokenTypeSpacing,
					Value:    "1.5rem",
					Category: "spacing",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewCSSParser()
			reader := strings.NewReader(tt.css)

			result, err := parser.Parse(reader)

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Len(t, result.Tokens, len(tt.expected))

			if len(result.Tokens) > 0 && len(tt.expected) > 0 {
				assert.Equal(t, tt.expected[0].Name, result.Tokens[0].Name)
				assert.Equal(t, tt.expected[0].Type, result.Tokens[0].Type)
				assert.Equal(t, tt.expected[0].Value, result.Tokens[0].Value)
			}
		})
	}
}

func TestCSSParser_Parse_ShadowTokens(t *testing.T) {
	tests := []struct {
		name     string
		css      string
		expected []Token
	}{
		{
			name: "Box shadow variable",
			css: `:root {
				--shadow-md: 0 4px 6px rgba(0, 0, 0, 0.1);
			}`,
			expected: []Token{
				{
					Name:  "shadow-md",
					Type:  TokenTypeShadow,
					Value: "0 4px 6px rgba(0, 0, 0, 0.1)",
				},
			},
		},
		{
			name: "Box shadow property",
			css: `.card {
				box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
			}`,
			expected: []Token{
				{
					Name:     "box-shadow",
					Type:     TokenTypeShadow,
					Value:    "0 2px 4px rgba(0, 0, 0, 0.05)",
					Category: "shadow",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewCSSParser()
			reader := strings.NewReader(tt.css)

			result, err := parser.Parse(reader)

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Len(t, result.Tokens, len(tt.expected))

			if len(result.Tokens) > 0 && len(tt.expected) > 0 {
				assert.Equal(t, tt.expected[0].Name, result.Tokens[0].Name)
				assert.Equal(t, tt.expected[0].Type, result.Tokens[0].Type)
				assert.Equal(t, tt.expected[0].Value, result.Tokens[0].Value)
			}
		})
	}
}

func TestCSSParser_Parse_BorderRadiusTokens(t *testing.T) {
	tests := []struct {
		name     string
		css      string
		expected []Token
	}{
		{
			name: "Border radius variable",
			css: `:root {
				--rounded-lg: 0.5rem;
			}`,
			expected: []Token{
				{
					Name:  "rounded-lg",
					Type:  TokenTypeBorderRadius,
					Value: "0.5rem",
				},
			},
		},
		{
			name: "Border radius property",
			css: `.button {
				border-radius: 0.375rem;
			}`,
			expected: []Token{
				{
					Name:     "border-radius",
					Type:     TokenTypeBorderRadius,
					Value:    "0.375rem",
					Category: "border-radius",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewCSSParser()
			reader := strings.NewReader(tt.css)

			result, err := parser.Parse(reader)

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Len(t, result.Tokens, len(tt.expected))

			if len(result.Tokens) > 0 && len(tt.expected) > 0 {
				assert.Equal(t, tt.expected[0].Name, result.Tokens[0].Name)
				assert.Equal(t, tt.expected[0].Type, result.Tokens[0].Type)
				assert.Equal(t, tt.expected[0].Value, result.Tokens[0].Value)
			}
		})
	}
}

func TestCSSParser_Parse_Comments(t *testing.T) {
	css := `
		/* This is a comment */
		:root {
			/* Another comment */
			--primary-color: #6366F1;
			// Single line comment (not standard CSS but should be handled)
			--secondary-color: #8B5CF6;
		}
		/* Multi-line
		   comment
		   block */
	`

	parser := NewCSSParser()
	reader := strings.NewReader(css)

	result, err := parser.Parse(reader)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.GreaterOrEqual(t, len(result.Tokens), 2)

	// Verify we extracted the non-commented variables
	foundPrimary := false
	foundSecondary := false
	for _, token := range result.Tokens {
		if token.Name == "primary-color" {
			foundPrimary = true
		}
		if token.Name == "secondary-color" {
			foundSecondary = true
		}
	}

	assert.True(t, foundPrimary, "Should extract primary-color")
	assert.True(t, foundSecondary, "Should extract secondary-color")
}

func TestCSSParser_Parse_EmptyInput(t *testing.T) {
	parser := NewCSSParser()
	reader := strings.NewReader("")

	result, err := parser.Parse(reader)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Empty(t, result.Tokens)
}

func TestCSSParser_Parse_ComplexCSS(t *testing.T) {
	css := `
		:root {
			/* Colors */
			--primary-color: #6366F1;
			--secondary-color: rgb(139, 92, 246);
			--success-color: #10B981;

			/* Typography */
			--font-family-base: "Inter", sans-serif;
			--font-size-sm: 0.875rem;
			--font-size-base: 1rem;
			--line-height-normal: 1.5;

			/* Spacing */
			--spacing-xs: 0.25rem;
			--spacing-sm: 0.5rem;
			--spacing-md: 1rem;

			/* Shadows */
			--shadow-sm: 0 1px 2px rgba(0, 0, 0, 0.05);
			--shadow-md: 0 4px 6px rgba(0, 0, 0, 0.1);

			/* Border Radius */
			--rounded-sm: 0.25rem;
			--rounded-md: 0.375rem;
		}

		.button {
			background-color: #6366F1;
			color: #FFFFFF;
			padding: 0.5rem 1rem;
			border-radius: 0.375rem;
			font-size: 1rem;
			box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
		}
	`

	parser := NewCSSParser()
	reader := strings.NewReader(css)

	result, err := parser.Parse(reader)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Greater(t, len(result.Tokens), 10, "Should extract multiple tokens")

	// Verify we have tokens of each type
	typeCount := make(map[TokenType]int)
	for _, token := range result.Tokens {
		typeCount[token.Type]++
	}

	assert.Greater(t, typeCount[TokenTypeColor], 0, "Should have color tokens")
	assert.Greater(t, typeCount[TokenTypeTypography], 0, "Should have typography tokens")
	assert.Greater(t, typeCount[TokenTypeSpacing], 0, "Should have spacing tokens")
	assert.Greater(t, typeCount[TokenTypeShadow], 0, "Should have shadow tokens")
	assert.Greater(t, typeCount[TokenTypeBorderRadius], 0, "Should have border-radius tokens")
}

func TestIsColorProperty(t *testing.T) {
	tests := []struct {
		property string
		expected bool
	}{
		{"color", true},
		{"background-color", true},
		{"border-color", true},
		{"outline-color", true},
		{"font-size", false},
		{"margin", false},
	}

	for _, tt := range tests {
		t.Run(tt.property, func(t *testing.T) {
			result := isColorProperty(tt.property)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsTypographyProperty(t *testing.T) {
	tests := []struct {
		property string
		expected bool
	}{
		{"font-family", true},
		{"font-size", true},
		{"font-weight", true},
		{"line-height", true},
		{"letter-spacing", true},
		{"color", false},
		{"margin", false},
	}

	for _, tt := range tests {
		t.Run(tt.property, func(t *testing.T) {
			result := isTypographyProperty(tt.property)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsSpacingProperty(t *testing.T) {
	tests := []struct {
		property string
		expected bool
	}{
		{"margin", true},
		{"margin-top", true},
		{"margin-bottom", true},
		{"padding", true},
		{"padding-left", true},
		{"gap", true},
		{"row-gap", true},
		{"column-gap", true},
		{"color", false},
		{"font-size", false},
	}

	for _, tt := range tests {
		t.Run(tt.property, func(t *testing.T) {
			result := isSpacingProperty(tt.property)
			assert.Equal(t, tt.expected, result)
		})
	}
}
