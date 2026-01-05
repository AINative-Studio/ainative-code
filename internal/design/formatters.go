package design

import (
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// Formatter formats design tokens into various output formats.
type Formatter interface {
	Format(tokens []Token) (string, error)
}

// JSONFormatter formats tokens as JSON.
type JSONFormatter struct {
	Pretty bool
}

// NewJSONFormatter creates a new JSON formatter.
func NewJSONFormatter(pretty bool) *JSONFormatter {
	return &JSONFormatter{Pretty: pretty}
}

// Format converts tokens to JSON format.
func (f *JSONFormatter) Format(tokens []Token) (string, error) {
	collection := TokenCollection{
		Tokens: tokens,
		Metadata: map[string]string{
			"version": "1.0.0",
			"format":  "json",
		},
	}

	var data []byte
	var err error

	if f.Pretty {
		data, err = json.MarshalIndent(collection, "", "  ")
	} else {
		data, err = json.Marshal(collection)
	}

	if err != nil {
		return "", fmt.Errorf("failed to marshal tokens to JSON: %w", err)
	}

	return string(data), nil
}

// YAMLFormatter formats tokens as YAML.
type YAMLFormatter struct{}

// NewYAMLFormatter creates a new YAML formatter.
func NewYAMLFormatter() *YAMLFormatter {
	return &YAMLFormatter{}
}

// Format converts tokens to YAML format.
func (f *YAMLFormatter) Format(tokens []Token) (string, error) {
	collection := TokenCollection{
		Tokens: tokens,
		Metadata: map[string]string{
			"version": "1.0.0",
			"format":  "yaml",
		},
	}

	data, err := yaml.Marshal(collection)
	if err != nil {
		return "", fmt.Errorf("failed to marshal tokens to YAML: %w", err)
	}

	return string(data), nil
}

// TailwindFormatter formats tokens as a Tailwind CSS configuration.
type TailwindFormatter struct {
	IncludeComments bool
}

// NewTailwindFormatter creates a new Tailwind formatter.
func NewTailwindFormatter(includeComments bool) *TailwindFormatter {
	return &TailwindFormatter{IncludeComments: includeComments}
}

// Format converts tokens to Tailwind config format.
func (f *TailwindFormatter) Format(tokens []Token) (string, error) {
	var builder strings.Builder

	// Header
	builder.WriteString("module.exports = {\n")
	builder.WriteString("  theme: {\n")
	builder.WriteString("    extend: {\n")

	// Group tokens by type
	colorTokens := filterTokensByType(tokens, TokenTypeColor)
	spacingTokens := filterTokensByType(tokens, TokenTypeSpacing)
	fontFamilyTokens := filterTypographyTokens(tokens, "font-family")
	fontSizeTokens := filterTypographyTokens(tokens, "font-size")
	lineHeightTokens := filterTypographyTokens(tokens, "line-height")
	shadowTokens := filterTokensByType(tokens, TokenTypeShadow)
	borderRadiusTokens := filterTokensByType(tokens, TokenTypeBorderRadius)

	// Colors
	if len(colorTokens) > 0 {
		builder.WriteString("      colors: {\n")
		if f.IncludeComments {
			builder.WriteString("        // Extracted color tokens\n")
		}
		for _, token := range colorTokens {
			name := f.formatTokenName(token.Name)
			builder.WriteString(fmt.Sprintf("        '%s': '%s',\n", name, token.Value))
		}
		builder.WriteString("      },\n")
	}

	// Spacing
	if len(spacingTokens) > 0 {
		builder.WriteString("      spacing: {\n")
		if f.IncludeComments {
			builder.WriteString("        // Extracted spacing tokens\n")
		}
		for _, token := range spacingTokens {
			name := f.formatTokenName(token.Name)
			builder.WriteString(fmt.Sprintf("        '%s': '%s',\n", name, token.Value))
		}
		builder.WriteString("      },\n")
	}

	// Font Family
	if len(fontFamilyTokens) > 0 {
		builder.WriteString("      fontFamily: {\n")
		if f.IncludeComments {
			builder.WriteString("        // Extracted font family tokens\n")
		}
		for _, token := range fontFamilyTokens {
			name := f.formatTokenName(token.Name)
			// Tailwind expects font families as arrays
			families := strings.Split(token.Value, ",")
			formattedFamilies := make([]string, len(families))
			for i, family := range families {
				formattedFamilies[i] = fmt.Sprintf("'%s'", strings.TrimSpace(family))
			}
			builder.WriteString(fmt.Sprintf("        '%s': [%s],\n", name, strings.Join(formattedFamilies, ", ")))
		}
		builder.WriteString("      },\n")
	}

	// Font Size
	if len(fontSizeTokens) > 0 {
		builder.WriteString("      fontSize: {\n")
		if f.IncludeComments {
			builder.WriteString("        // Extracted font size tokens\n")
		}
		for _, token := range fontSizeTokens {
			name := f.formatTokenName(token.Name)
			builder.WriteString(fmt.Sprintf("        '%s': '%s',\n", name, token.Value))
		}
		builder.WriteString("      },\n")
	}

	// Line Height
	if len(lineHeightTokens) > 0 {
		builder.WriteString("      lineHeight: {\n")
		if f.IncludeComments {
			builder.WriteString("        // Extracted line height tokens\n")
		}
		for _, token := range lineHeightTokens {
			name := f.formatTokenName(token.Name)
			builder.WriteString(fmt.Sprintf("        '%s': '%s',\n", name, token.Value))
		}
		builder.WriteString("      },\n")
	}

	// Box Shadow
	if len(shadowTokens) > 0 {
		builder.WriteString("      boxShadow: {\n")
		if f.IncludeComments {
			builder.WriteString("        // Extracted shadow tokens\n")
		}
		for _, token := range shadowTokens {
			name := f.formatTokenName(token.Name)
			builder.WriteString(fmt.Sprintf("        '%s': '%s',\n", name, token.Value))
		}
		builder.WriteString("      },\n")
	}

	// Border Radius
	if len(borderRadiusTokens) > 0 {
		builder.WriteString("      borderRadius: {\n")
		if f.IncludeComments {
			builder.WriteString("        // Extracted border radius tokens\n")
		}
		for _, token := range borderRadiusTokens {
			name := f.formatTokenName(token.Name)
			builder.WriteString(fmt.Sprintf("        '%s': '%s',\n", name, token.Value))
		}
		builder.WriteString("      },\n")
	}

	// Footer
	builder.WriteString("    },\n")
	builder.WriteString("  },\n")
	builder.WriteString("  plugins: [],\n")
	builder.WriteString("}\n")

	return builder.String(), nil
}

// formatTokenName formats a token name for Tailwind config.
func (f *TailwindFormatter) formatTokenName(name string) string {
	// Remove prefixes like 'color-', 'spacing-', etc.
	name = strings.TrimPrefix(name, "color-")
	name = strings.TrimPrefix(name, "spacing-")
	name = strings.TrimPrefix(name, "font-")
	name = strings.TrimPrefix(name, "shadow-")
	name = strings.TrimPrefix(name, "border-")
	name = strings.TrimPrefix(name, "radius-")

	// Convert to kebab-case
	name = strings.ReplaceAll(name, "_", "-")

	return name
}

// filterTokensByType filters tokens by their type.
func filterTokensByType(tokens []Token, tokenType TokenType) []Token {
	result := make([]Token, 0)
	for _, token := range tokens {
		if token.Type == tokenType {
			result = append(result, token)
		}
	}
	return result
}

// filterTypographyTokens filters typography tokens by property name.
func filterTypographyTokens(tokens []Token, property string) []Token {
	result := make([]Token, 0)
	for _, token := range tokens {
		if token.Type == TokenTypeTypography && strings.Contains(strings.ToLower(token.Name), property) {
			result = append(result, token)
		}
	}
	return result
}

// StyleDictionaryFormatter formats tokens in Style Dictionary format.
type StyleDictionaryFormatter struct {
	Pretty bool
}

// NewStyleDictionaryFormatter creates a new Style Dictionary formatter.
func NewStyleDictionaryFormatter(pretty bool) *StyleDictionaryFormatter {
	return &StyleDictionaryFormatter{Pretty: pretty}
}

// Format converts tokens to Style Dictionary format.
func (f *StyleDictionaryFormatter) Format(tokens []Token) (string, error) {
	// Group tokens by category
	grouped := make(map[string]map[string]interface{})

	for _, token := range tokens {
		category := string(token.Type)
		if token.Category != "" {
			category = token.Category
		}

		if grouped[category] == nil {
			grouped[category] = make(map[string]interface{})
		}

		// Create token object in Style Dictionary format
		tokenObj := map[string]interface{}{
			"value": token.Value,
		}

		if token.Description != "" {
			tokenObj["comment"] = token.Description
		}

		if token.Type != "" {
			tokenObj["type"] = string(token.Type)
		}

		// Use token name as key
		grouped[category][token.Name] = tokenObj
	}

	var data []byte
	var err error

	if f.Pretty {
		data, err = json.MarshalIndent(grouped, "", "  ")
	} else {
		data, err = json.Marshal(grouped)
	}

	if err != nil {
		return "", fmt.Errorf("failed to marshal tokens to Style Dictionary format: %w", err)
	}

	return string(data), nil
}

// CSSVariablesFormatter formats tokens as CSS custom properties.
type CSSVariablesFormatter struct {
	Prefix string
}

// NewCSSVariablesFormatter creates a new CSS variables formatter.
func NewCSSVariablesFormatter(prefix string) *CSSVariablesFormatter {
	if prefix == "" {
		prefix = "token"
	}
	return &CSSVariablesFormatter{Prefix: prefix}
}

// Format converts tokens to CSS custom properties.
func (f *CSSVariablesFormatter) Format(tokens []Token) (string, error) {
	var builder strings.Builder

	builder.WriteString(":root {\n")

	// Group tokens by category for better organization
	categories := make(map[string][]Token)
	for _, token := range tokens {
		category := token.Category
		if category == "" {
			category = string(token.Type)
		}
		categories[category] = append(categories[category], token)
	}

	// Output tokens by category
	for category, categoryTokens := range categories {
		if len(categoryTokens) > 0 {
			builder.WriteString(fmt.Sprintf("\n  /* %s */\n", category))
			for _, token := range categoryTokens {
				varName := f.formatVariableName(token.Name)
				builder.WriteString(fmt.Sprintf("  --%s-%s: %s;\n", f.Prefix, varName, token.Value))
			}
		}
	}

	builder.WriteString("}\n")

	return builder.String(), nil
}

// formatVariableName formats a token name for CSS custom properties.
func (f *CSSVariablesFormatter) formatVariableName(name string) string {
	// Convert to kebab-case
	name = strings.ReplaceAll(name, "_", "-")
	name = strings.ToLower(name)
	return name
}

// SCSSVariablesFormatter formats tokens as SCSS variables.
type SCSSVariablesFormatter struct{}

// NewSCSSVariablesFormatter creates a new SCSS variables formatter.
func NewSCSSVariablesFormatter() *SCSSVariablesFormatter {
	return &SCSSVariablesFormatter{}
}

// Format converts tokens to SCSS variables.
func (f *SCSSVariablesFormatter) Format(tokens []Token) (string, error) {
	var builder strings.Builder

	builder.WriteString("// Design Tokens\n")
	builder.WriteString("// Generated from extraction\n\n")

	// Group tokens by category
	categories := make(map[string][]Token)
	for _, token := range tokens {
		category := token.Category
		if category == "" {
			category = string(token.Type)
		}
		categories[category] = append(categories[category], token)
	}

	// Output tokens by category
	for category, categoryTokens := range categories {
		if len(categoryTokens) > 0 {
			builder.WriteString(fmt.Sprintf("// %s\n", category))
			for _, token := range categoryTokens {
				varName := f.formatVariableName(token.Name)
				if token.Description != "" {
					builder.WriteString(fmt.Sprintf("// %s\n", token.Description))
				}
				builder.WriteString(fmt.Sprintf("$%s: %s;\n", varName, token.Value))
			}
			builder.WriteString("\n")
		}
	}

	return builder.String(), nil
}

// formatVariableName formats a token name for SCSS variables.
func (f *SCSSVariablesFormatter) formatVariableName(name string) string {
	// Convert to kebab-case
	name = strings.ReplaceAll(name, "_", "-")
	name = strings.ToLower(name)
	return name
}

// GetFormatter returns the appropriate formatter based on the output format.
func GetFormatter(format OutputFormat, options map[string]interface{}) (Formatter, error) {
	switch format {
	case OutputFormatJSON:
		pretty := true
		if val, ok := options["pretty"].(bool); ok {
			pretty = val
		}
		return NewJSONFormatter(pretty), nil

	case OutputFormatYAML:
		return NewYAMLFormatter(), nil

	case OutputFormatTailwind:
		includeComments := true
		if val, ok := options["includeComments"].(bool); ok {
			includeComments = val
		}
		return NewTailwindFormatter(includeComments), nil

	default:
		return nil, fmt.Errorf("unsupported output format: %s (supported: json, yaml, tailwind)", format)
	}
}
