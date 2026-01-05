package design

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// Parser is an interface for parsing different CSS-like formats.
type Parser interface {
	Parse(reader io.Reader) (*ExtractionResult, error)
}

// CSSParser parses CSS files to extract design tokens.
type CSSParser struct {
	// colorRegex matches color values in various formats
	colorRegex *regexp.Regexp
	// variableRegex matches CSS custom properties (--variable-name: value;)
	variableRegex *regexp.Regexp
	// propertyRegex matches CSS properties
	propertyRegex *regexp.Regexp
}

// NewCSSParser creates a new CSS parser.
func NewCSSParser() *CSSParser {
	return &CSSParser{
		colorRegex:    regexp.MustCompile(`(?i)(#[0-9a-f]{3,8}|rgba?\([^)]+\)|hsla?\([^)]+\))`),
		variableRegex: regexp.MustCompile(`--([a-zA-Z0-9-_]+)\s*:\s*([^;]+);`),
		propertyRegex: regexp.MustCompile(`([a-zA-Z-]+)\s*:\s*([^;]+);`),
	}
}

// Parse parses a CSS file and extracts design tokens.
func (p *CSSParser) Parse(reader io.Reader) (*ExtractionResult, error) {
	result := &ExtractionResult{
		Tokens:   make([]Token, 0),
		Warnings: make([]string, 0),
		Errors:   make([]error, 0),
	}

	scanner := bufio.NewScanner(reader)
	lineNumber := 0
	inComment := false

	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines
		if line == "" {
			continue
		}

		// Handle multi-line comments
		if strings.Contains(line, "/*") {
			inComment = true
		}
		if inComment {
			if strings.Contains(line, "*/") {
				inComment = false
			}
			continue
		}

		// Skip single-line comments
		if strings.HasPrefix(line, "//") {
			continue
		}

		// Extract CSS custom properties (variables)
		if matches := p.variableRegex.FindStringSubmatch(line); matches != nil {
			token := p.extractTokenFromVariable(matches[1], matches[2], lineNumber)
			if token != nil {
				result.Tokens = append(result.Tokens, *token)
			}
			continue
		}

		// Extract from regular CSS properties
		if matches := p.propertyRegex.FindStringSubmatch(line); matches != nil {
			tokens := p.extractTokensFromProperty(matches[1], matches[2], lineNumber)
			result.Tokens = append(result.Tokens, tokens...)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading CSS file: %w", err)
	}

	return result, nil
}

// extractTokenFromVariable extracts a token from a CSS custom property.
func (p *CSSParser) extractTokenFromVariable(name, value string, line int) *Token {
	value = strings.TrimSpace(value)

	// Determine token type based on value
	tokenType := p.inferTokenType(name, value)
	if tokenType == "" {
		return nil
	}

	return &Token{
		Name:  name,
		Type:  tokenType,
		Value: value,
		Metadata: map[string]string{
			"source": "css-variable",
			"line":   fmt.Sprintf("%d", line),
		},
	}
}

// extractTokensFromProperty extracts tokens from a CSS property.
func (p *CSSParser) extractTokensFromProperty(property, value string, line int) []Token {
	tokens := make([]Token, 0)
	property = strings.TrimSpace(property)
	value = strings.TrimSpace(value)

	switch {
	case isColorProperty(property):
		if token := p.extractColorToken(property, value, line); token != nil {
			tokens = append(tokens, *token)
		}
	case isTypographyProperty(property):
		if token := p.extractTypographyToken(property, value, line); token != nil {
			tokens = append(tokens, *token)
		}
	case isSpacingProperty(property):
		if token := p.extractSpacingToken(property, value, line); token != nil {
			tokens = append(tokens, *token)
		}
	case isShadowProperty(property):
		if token := p.extractShadowToken(property, value, line); token != nil {
			tokens = append(tokens, *token)
		}
	case isBorderRadiusProperty(property):
		if token := p.extractBorderRadiusToken(property, value, line); token != nil {
			tokens = append(tokens, *token)
		}
	}

	return tokens
}

// extractColorToken creates a color token.
func (p *CSSParser) extractColorToken(property, value string, line int) *Token {
	if !p.colorRegex.MatchString(value) {
		return nil
	}

	return &Token{
		Name:  property,
		Type:  TokenTypeColor,
		Value: value,
		Metadata: map[string]string{
			"source": "css-property",
			"line":   fmt.Sprintf("%d", line),
		},
	}
}

// extractTypographyToken creates a typography token.
func (p *CSSParser) extractTypographyToken(property, value string, line int) *Token {
	return &Token{
		Name:     property,
		Type:     TokenTypeTypography,
		Value:    value,
		Category: "typography",
		Metadata: map[string]string{
			"source": "css-property",
			"line":   fmt.Sprintf("%d", line),
		},
	}
}

// extractSpacingToken creates a spacing token.
func (p *CSSParser) extractSpacingToken(property, value string, line int) *Token {
	return &Token{
		Name:     property,
		Type:     TokenTypeSpacing,
		Value:    value,
		Category: "spacing",
		Metadata: map[string]string{
			"source": "css-property",
			"line":   fmt.Sprintf("%d", line),
		},
	}
}

// extractShadowToken creates a shadow token.
func (p *CSSParser) extractShadowToken(property, value string, line int) *Token {
	return &Token{
		Name:     property,
		Type:     TokenTypeShadow,
		Value:    value,
		Category: "shadow",
		Metadata: map[string]string{
			"source": "css-property",
			"line":   fmt.Sprintf("%d", line),
		},
	}
}

// extractBorderRadiusToken creates a border-radius token.
func (p *CSSParser) extractBorderRadiusToken(property, value string, line int) *Token {
	return &Token{
		Name:     property,
		Type:     TokenTypeBorderRadius,
		Value:    value,
		Category: "border-radius",
		Metadata: map[string]string{
			"source": "css-property",
			"line":   fmt.Sprintf("%d", line),
		},
	}
}

// inferTokenType infers the token type from the variable name and value.
func (p *CSSParser) inferTokenType(name, value string) TokenType {
	nameLower := strings.ToLower(name)

	// Check by name patterns
	if strings.Contains(nameLower, "color") || strings.Contains(nameLower, "bg") ||
		strings.Contains(nameLower, "text") || strings.Contains(nameLower, "border-color") {
		if p.colorRegex.MatchString(value) {
			return TokenTypeColor
		}
	}

	if strings.Contains(nameLower, "font") || strings.Contains(nameLower, "text") ||
		strings.Contains(nameLower, "line-height") {
		return TokenTypeTypography
	}

	if strings.Contains(nameLower, "spacing") || strings.Contains(nameLower, "margin") ||
		strings.Contains(nameLower, "padding") || strings.Contains(nameLower, "gap") {
		return TokenTypeSpacing
	}

	if strings.Contains(nameLower, "shadow") {
		return TokenTypeShadow
	}

	if strings.Contains(nameLower, "radius") || strings.Contains(nameLower, "rounded") {
		return TokenTypeBorderRadius
	}

	// Check by value patterns
	if p.colorRegex.MatchString(value) {
		return TokenTypeColor
	}

	return ""
}

// Helper functions to identify property types

func isColorProperty(property string) bool {
	colorProps := []string{"color", "background-color", "border-color", "outline-color"}
	for _, prop := range colorProps {
		if property == prop {
			return true
		}
	}
	return false
}

func isTypographyProperty(property string) bool {
	typographyProps := []string{"font-family", "font-size", "font-weight", "line-height", "letter-spacing"}
	for _, prop := range typographyProps {
		if property == prop {
			return true
		}
	}
	return false
}

func isSpacingProperty(property string) bool {
	return strings.HasPrefix(property, "margin") || strings.HasPrefix(property, "padding") ||
		property == "gap" || property == "row-gap" || property == "column-gap"
}

func isShadowProperty(property string) bool {
	return property == "box-shadow" || property == "text-shadow"
}

func isBorderRadiusProperty(property string) bool {
	return strings.Contains(property, "border-radius")
}
