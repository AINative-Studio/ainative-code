// Package fixtures provides test data fixtures for integration tests.
package fixtures

import (
	"github.com/AINative-studio/ainative-code/internal/design"
)

// TestFigmaFile returns a sample Figma file URL for testing.
func TestFigmaFile() string {
	return "https://www.figma.com/file/test123/Test-Design-File"
}

// TestDesignTokens returns sample design tokens for testing.
func TestDesignTokens() *design.TokenCollection {
	return &design.TokenCollection{
		Tokens: []design.Token{
			{
				Name:        "primary",
				Type:        design.TokenTypeColor,
				Value:       "#3B82F6",
				Description: "Primary brand color",
				Category:    "colors",
			},
			{
				Name:        "secondary",
				Type:        design.TokenTypeColor,
				Value:       "#10B981",
				Description: "Secondary brand color",
				Category:    "colors",
			},
			{
				Name:     "heading-1-family",
				Type:     design.TokenTypeTypography,
				Value:    "Inter",
				Category: "typography",
			},
			{
				Name:     "heading-1-size",
				Type:     design.TokenTypeTypography,
				Value:    "32px",
				Category: "typography",
			},
			{
				Name:     "xs",
				Type:     design.TokenTypeSpacing,
				Value:    "4px",
				Category: "spacing",
			},
			{
				Name:     "sm",
				Type:     design.TokenTypeSpacing,
				Value:    "8px",
				Category: "spacing",
			},
			{
				Name:     "md",
				Type:     design.TokenTypeSpacing,
				Value:    "16px",
				Category: "spacing",
			},
			{
				Name:     "shadow-sm",
				Type:     design.TokenTypeShadow,
				Value:    "0px 1px 2px 0px rgba(0, 0, 0, 0.05)",
				Category: "effects",
			},
		},
		Metadata: map[string]string{
			"version": "1.0.0",
		},
	}
}

// TestCSSOutput returns expected CSS output for test tokens.
func TestCSSOutput() string {
	return `:root {
  /* Colors */
  --color-primary: #3B82F6;
  --color-secondary: #10B981;

  /* Typography */
  --font-heading-1-family: Inter;
  --font-heading-1-size: 32px;
  --font-heading-1-weight: 700;
  --font-heading-1-line-height: 1.2;

  --font-body-family: Inter;
  --font-body-size: 16px;
  --font-body-weight: 400;
  --font-body-line-height: 1.5;

  /* Spacing */
  --spacing-xs: 4px;
  --spacing-sm: 8px;
  --spacing-md: 16px;

  /* Effects */
  --effect-shadow-sm: 0px 1px 2px 0px rgba(0, 0, 0, 0.05);
}
`
}
