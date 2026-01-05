package design

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Extractor extracts design tokens from CSS/SCSS/LESS files.
type Extractor struct {
	validator        *Validator
	cssParser        *CSSParser
	scssParser       *SCSSParser
	lessParser       *LESSParser
	categorizer      *TokenCategorizer
	enableValidation bool
}

// ExtractorOptions configures the token extractor.
type ExtractorOptions struct {
	EnableValidation bool
	IncludeComments  bool
	InferCategories  bool
}

// NewExtractor creates a new design token extractor.
func NewExtractor(opts *ExtractorOptions) *Extractor {
	if opts == nil {
		opts = &ExtractorOptions{
			EnableValidation: true,
			IncludeComments:  true,
			InferCategories:  true,
		}
	}

	return &Extractor{
		validator:        NewValidator(),
		cssParser:        NewCSSParser(),
		scssParser:       NewSCSSParser(),
		lessParser:       NewLESSParser(),
		categorizer:      NewTokenCategorizer(),
		enableValidation: opts.EnableValidation,
	}
}

// ExtractFromFile extracts design tokens from a file.
func (e *Extractor) ExtractFromFile(filePath string) (*ExtractionResult, error) {
	ext := strings.ToLower(filepath.Ext(filePath))

	var parser Parser
	switch ext {
	case ".css":
		parser = e.cssParser
	case ".scss", ".sass":
		parser = e.scssParser
	case ".less":
		parser = e.lessParser
	default:
		return nil, fmt.Errorf("unsupported file type: %s (supported: .css, .scss, .sass, .less)", ext)
	}

	file, err := openFile(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return e.Extract(file, parser)
}

// Extract extracts design tokens using the specified parser.
func (e *Extractor) Extract(reader io.Reader, parser Parser) (*ExtractionResult, error) {
	result, err := parser.Parse(reader)
	if err != nil {
		return nil, err
	}

	// Categorize tokens
	for i := range result.Tokens {
		e.categorizer.Categorize(&result.Tokens[i])
	}

	// Validate tokens if enabled
	if e.enableValidation {
		for _, token := range result.Tokens {
			validationResult := e.validator.Validate(&token)
			if !validationResult.Valid {
				for _, validationErr := range validationResult.Errors {
					result.Warnings = append(result.Warnings, validationErr.Error())
				}
			}
		}
	}

	// Deduplicate tokens
	result.Tokens = deduplicateTokens(result.Tokens)

	return result, nil
}

// SCSSParser parses SCSS files to extract design tokens.
type SCSSParser struct {
	colorRegex     *regexp.Regexp
	variableRegex  *regexp.Regexp
	propertyRegex  *regexp.Regexp
	mixinRegex     *regexp.Regexp
	functionRegex  *regexp.Regexp
	mapRegex       *regexp.Regexp
}

// NewSCSSParser creates a new SCSS parser.
func NewSCSSParser() *SCSSParser {
	return &SCSSParser{
		colorRegex:    regexp.MustCompile(`(?i)(#[0-9a-f]{3,8}|rgba?\([^)]+\)|hsla?\([^)]+\))`),
		variableRegex: regexp.MustCompile(`\$([a-zA-Z0-9-_]+)\s*:\s*([^;]+);`),
		propertyRegex: regexp.MustCompile(`([a-zA-Z-]+)\s*:\s*([^;]+);`),
		mixinRegex:    regexp.MustCompile(`@mixin\s+([a-zA-Z0-9-_]+)`),
		functionRegex: regexp.MustCompile(`@function\s+([a-zA-Z0-9-_]+)`),
		mapRegex:      regexp.MustCompile(`\$([a-zA-Z0-9-_]+)\s*:\s*\(`),
	}
}

// Parse parses an SCSS file and extracts design tokens.
func (p *SCSSParser) Parse(reader io.Reader) (*ExtractionResult, error) {
	result := &ExtractionResult{
		Tokens:   make([]Token, 0),
		Warnings: make([]string, 0),
		Errors:   make([]error, 0),
	}

	scanner := bufio.NewScanner(reader)
	lineNumber := 0
	inComment := false
	inMap := false
	currentMapName := ""
	mapTokens := make([]Token, 0)

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

		// Check for SCSS map start
		if matches := p.mapRegex.FindStringSubmatch(line); matches != nil {
			inMap = true
			currentMapName = matches[1]
			mapTokens = make([]Token, 0)
			continue
		}

		// Check for map end
		if inMap && strings.Contains(line, ");") {
			// Add all map tokens with prefix
			for _, token := range mapTokens {
				token.Name = currentMapName + "-" + token.Name
				result.Tokens = append(result.Tokens, token)
			}
			inMap = false
			currentMapName = ""
			mapTokens = make([]Token, 0)
			continue
		}

		// Extract from map entries
		if inMap {
			if matches := p.variableRegex.FindStringSubmatch(line); matches != nil {
				token := p.extractTokenFromVariable(matches[1], matches[2], lineNumber)
				if token != nil {
					mapTokens = append(mapTokens, *token)
				}
			}
			// Also check for quoted keys in maps
			keyValueRegex := regexp.MustCompile(`['"]?([a-zA-Z0-9-_]+)['"]?\s*:\s*([^,]+)`)
			if matches := keyValueRegex.FindStringSubmatch(line); matches != nil {
				token := p.extractTokenFromVariable(matches[1], matches[2], lineNumber)
				if token != nil {
					mapTokens = append(mapTokens, *token)
				}
			}
			continue
		}

		// Extract SCSS variables
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
		return nil, fmt.Errorf("error reading SCSS file: %w", err)
	}

	return result, nil
}

// extractTokenFromVariable extracts a token from an SCSS variable.
func (p *SCSSParser) extractTokenFromVariable(name, value string, line int) *Token {
	value = strings.TrimSpace(value)
	value = strings.TrimSuffix(value, ",") // Remove trailing comma from maps

	// Skip variables that reference other variables (we'll resolve them later)
	if strings.HasPrefix(value, "$") {
		return nil
	}

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
			"source": "scss-variable",
			"line":   fmt.Sprintf("%d", line),
		},
	}
}

// extractTokensFromProperty extracts tokens from a CSS/SCSS property.
func (p *SCSSParser) extractTokensFromProperty(property, value string, line int) []Token {
	tokens := make([]Token, 0)
	property = strings.TrimSpace(property)
	value = strings.TrimSpace(value)

	// Skip variable references
	if strings.HasPrefix(value, "$") {
		return tokens
	}

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
func (p *SCSSParser) extractColorToken(property, value string, line int) *Token {
	if !p.colorRegex.MatchString(value) {
		return nil
	}

	return &Token{
		Name:  property,
		Type:  TokenTypeColor,
		Value: value,
		Metadata: map[string]string{
			"source": "scss-property",
			"line":   fmt.Sprintf("%d", line),
		},
	}
}

// extractTypographyToken creates a typography token.
func (p *SCSSParser) extractTypographyToken(property, value string, line int) *Token {
	return &Token{
		Name:     property,
		Type:     TokenTypeTypography,
		Value:    value,
		Category: "typography",
		Metadata: map[string]string{
			"source": "scss-property",
			"line":   fmt.Sprintf("%d", line),
		},
	}
}

// extractSpacingToken creates a spacing token.
func (p *SCSSParser) extractSpacingToken(property, value string, line int) *Token {
	return &Token{
		Name:     property,
		Type:     TokenTypeSpacing,
		Value:    value,
		Category: "spacing",
		Metadata: map[string]string{
			"source": "scss-property",
			"line":   fmt.Sprintf("%d", line),
		},
	}
}

// extractShadowToken creates a shadow token.
func (p *SCSSParser) extractShadowToken(property, value string, line int) *Token {
	return &Token{
		Name:     property,
		Type:     TokenTypeShadow,
		Value:    value,
		Category: "shadow",
		Metadata: map[string]string{
			"source": "scss-property",
			"line":   fmt.Sprintf("%d", line),
		},
	}
}

// extractBorderRadiusToken creates a border-radius token.
func (p *SCSSParser) extractBorderRadiusToken(property, value string, line int) *Token {
	return &Token{
		Name:     property,
		Type:     TokenTypeBorderRadius,
		Value:    value,
		Category: "border-radius",
		Metadata: map[string]string{
			"source": "scss-property",
			"line":   fmt.Sprintf("%d", line),
		},
	}
}

// inferTokenType infers the token type from the variable name and value.
func (p *SCSSParser) inferTokenType(name, value string) TokenType {
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

// LESSParser parses LESS files to extract design tokens.
type LESSParser struct {
	colorRegex    *regexp.Regexp
	variableRegex *regexp.Regexp
	propertyRegex *regexp.Regexp
}

// NewLESSParser creates a new LESS parser.
func NewLESSParser() *LESSParser {
	return &LESSParser{
		colorRegex:    regexp.MustCompile(`(?i)(#[0-9a-f]{3,8}|rgba?\([^)]+\)|hsla?\([^)]+\))`),
		variableRegex: regexp.MustCompile(`@([a-zA-Z0-9-_]+)\s*:\s*([^;]+);`),
		propertyRegex: regexp.MustCompile(`([a-zA-Z-]+)\s*:\s*([^;]+);`),
	}
}

// Parse parses a LESS file and extracts design tokens.
func (p *LESSParser) Parse(reader io.Reader) (*ExtractionResult, error) {
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

		// Extract LESS variables
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
		return nil, fmt.Errorf("error reading LESS file: %w", err)
	}

	return result, nil
}

// extractTokenFromVariable extracts a token from a LESS variable.
func (p *LESSParser) extractTokenFromVariable(name, value string, line int) *Token {
	value = strings.TrimSpace(value)

	// Skip variables that reference other variables
	if strings.HasPrefix(value, "@") {
		return nil
	}

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
			"source": "less-variable",
			"line":   fmt.Sprintf("%d", line),
		},
	}
}

// extractTokensFromProperty extracts tokens from a CSS/LESS property.
func (p *LESSParser) extractTokensFromProperty(property, value string, line int) []Token {
	tokens := make([]Token, 0)
	property = strings.TrimSpace(property)
	value = strings.TrimSpace(value)

	// Skip variable references
	if strings.HasPrefix(value, "@") {
		return tokens
	}

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
func (p *LESSParser) extractColorToken(property, value string, line int) *Token {
	if !p.colorRegex.MatchString(value) {
		return nil
	}

	return &Token{
		Name:  property,
		Type:  TokenTypeColor,
		Value: value,
		Metadata: map[string]string{
			"source": "less-property",
			"line":   fmt.Sprintf("%d", line),
		},
	}
}

// extractTypographyToken creates a typography token.
func (p *LESSParser) extractTypographyToken(property, value string, line int) *Token {
	return &Token{
		Name:     property,
		Type:     TokenTypeTypography,
		Value:    value,
		Category: "typography",
		Metadata: map[string]string{
			"source": "less-property",
			"line":   fmt.Sprintf("%d", line),
		},
	}
}

// extractSpacingToken creates a spacing token.
func (p *LESSParser) extractSpacingToken(property, value string, line int) *Token {
	return &Token{
		Name:     property,
		Type:     TokenTypeSpacing,
		Value:    value,
		Category: "spacing",
		Metadata: map[string]string{
			"source": "less-property",
			"line":   fmt.Sprintf("%d", line),
		},
	}
}

// extractShadowToken creates a shadow token.
func (p *LESSParser) extractShadowToken(property, value string, line int) *Token {
	return &Token{
		Name:     property,
		Type:     TokenTypeShadow,
		Value:    value,
		Category: "shadow",
		Metadata: map[string]string{
			"source": "less-property",
			"line":   fmt.Sprintf("%d", line),
		},
	}
}

// extractBorderRadiusToken creates a border-radius token.
func (p *LESSParser) extractBorderRadiusToken(property, value string, line int) *Token {
	return &Token{
		Name:     property,
		Type:     TokenTypeBorderRadius,
		Value:    value,
		Category: "border-radius",
		Metadata: map[string]string{
			"source": "less-property",
			"line":   fmt.Sprintf("%d", line),
		},
	}
}

// inferTokenType infers the token type from the variable name and value.
func (p *LESSParser) inferTokenType(name, value string) TokenType {
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

// TokenCategorizer categorizes tokens based on their properties.
type TokenCategorizer struct {
	colorPrefixes  []string
	spacingPrefixes []string
	typographyPrefixes []string
}

// NewTokenCategorizer creates a new token categorizer.
func NewTokenCategorizer() *TokenCategorizer {
	return &TokenCategorizer{
		colorPrefixes:  []string{"primary", "secondary", "accent", "neutral", "success", "warning", "error", "info"},
		spacingPrefixes: []string{"xs", "sm", "md", "lg", "xl", "2xl", "3xl"},
		typographyPrefixes: []string{"heading", "body", "caption", "label"},
	}
}

// Categorize assigns a category to a token based on its name and type.
func (c *TokenCategorizer) Categorize(token *Token) {
	if token.Category != "" {
		return // Already categorized
	}

	nameLower := strings.ToLower(token.Name)

	switch token.Type {
	case TokenTypeColor:
		token.Category = c.categorizeColor(nameLower)
	case TokenTypeTypography:
		token.Category = c.categorizeTypography(nameLower)
	case TokenTypeSpacing:
		token.Category = c.categorizeSpacing(nameLower)
	case TokenTypeShadow:
		token.Category = "shadows"
	case TokenTypeBorderRadius:
		token.Category = "borders"
	}

	if token.Category == "" {
		token.Category = "uncategorized"
	}
}

func (c *TokenCategorizer) categorizeColor(name string) string {
	for _, prefix := range c.colorPrefixes {
		if strings.HasPrefix(name, prefix) || strings.Contains(name, "-"+prefix) {
			return "colors-" + prefix
		}
	}
	return "colors"
}

func (c *TokenCategorizer) categorizeTypography(name string) string {
	for _, prefix := range c.typographyPrefixes {
		if strings.HasPrefix(name, prefix) || strings.Contains(name, "-"+prefix) {
			return "typography-" + prefix
		}
	}
	return "typography"
}

func (c *TokenCategorizer) categorizeSpacing(name string) string {
	for _, prefix := range c.spacingPrefixes {
		if strings.HasPrefix(name, prefix) || strings.Contains(name, "-"+prefix) {
			return "spacing-" + prefix
		}
	}
	return "spacing"
}

// deduplicateTokens removes duplicate tokens, keeping the first occurrence.
func deduplicateTokens(tokens []Token) []Token {
	seen := make(map[string]bool)
	result := make([]Token, 0)

	for _, token := range tokens {
		key := token.Name + "|" + string(token.Type)
		if !seen[key] {
			seen[key] = true
			result = append(result, token)
		}
	}

	return result
}

// openFile is a helper function to open a file (can be stubbed for testing).
var openFile = func(path string) (io.ReadCloser, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	file, err := os.Open(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return file, nil
}
