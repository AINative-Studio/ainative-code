package design

import (
	"fmt"
	"regexp"
	"strings"
)

// Note: Token type is defined in types.go

// ConflictResolutionStrategy defines how to handle conflicting tokens during upload.
// Renamed from ConflictResolution to avoid conflict with sync_types.go
type ConflictResolutionStrategyUpload string

const (
	// ConflictOverwrite replaces existing tokens with new values
	ConflictOverwrite ConflictResolutionStrategyUpload = "overwrite"

	// ConflictMerge merges new tokens with existing, preferring new values
	ConflictMerge ConflictResolutionStrategyUpload = "merge"

	// ConflictSkip skips conflicting tokens and keeps existing values
	ConflictSkip ConflictResolutionStrategyUpload = "skip"
)

// ValidationError represents a token validation error.
type ValidationError struct {
	TokenName string
	Field     string
	Message   string
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("token '%s': %s - %s", e.TokenName, e.Field, e.Message)
}

// ValidationResult contains the results of token validation.
type ValidationResult struct {
	Valid  bool
	Errors []*ValidationError
}

// Validator provides design token validation functionality.
type Validator struct {
	colorRegex       *regexp.Regexp
	hexColorRegex    *regexp.Regexp
	rgbColorRegex    *regexp.Regexp
	rgbaColorRegex   *regexp.Regexp
	hslColorRegex    *regexp.Regexp
	hslaColorRegex   *regexp.Regexp
	shadowRegex      *regexp.Regexp
	sizingRegex      *regexp.Regexp
	validTokenTypes  map[string]bool
}

// NewValidator creates a new token validator.
func NewValidator() *Validator {
	return &Validator{
		hexColorRegex:  regexp.MustCompile(`^#([A-Fa-f0-9]{3}|[A-Fa-f0-9]{6}|[A-Fa-f0-9]{8})$`),
		rgbColorRegex:  regexp.MustCompile(`^rgb\(\s*\d{1,3}\s*,\s*\d{1,3}\s*,\s*\d{1,3}\s*\)$`),
		rgbaColorRegex: regexp.MustCompile(`^rgba\(\s*\d{1,3}\s*,\s*\d{1,3}\s*,\s*\d{1,3}\s*,\s*[0-9.]+\s*\)$`),
		hslColorRegex:  regexp.MustCompile(`^hsl\(\s*\d{1,3}\s*,\s*\d{1,3}%\s*,\s*\d{1,3}%\s*\)$`),
		hslaColorRegex: regexp.MustCompile(`^hsla\(\s*\d{1,3}\s*,\s*\d{1,3}%\s*,\s*\d{1,3}%\s*,\s*[0-9.]+\s*\)$`),
		shadowRegex:    regexp.MustCompile(`^(-?\d+\.?\d*(px|rem|em)?)\s+(-?\d+\.?\d*(px|rem|em)?)\s+(-?\d+\.?\d*(px|rem|em)?)\s+((-?\d+\.?\d*(px|rem|em)?)\s+)?`),
		sizingRegex:    regexp.MustCompile(`^-?\d+\.?\d*(px|rem|em|%|vh|vw|vmin|vmax)$`),
		validTokenTypes: map[string]bool{
			"color":         true,
			"typography":    true,
			"spacing":       true,
			"shadow":        true,
			"border-radius": true,
			"font-family":   true,
			"font-size":     true,
			"font-weight":   true,
			"line-height":   true,
			"letter-spacing": true,
			"duration":      true,
			"easing":        true,
			"z-index":       true,
			"opacity":       true,
		},
	}
}

// Validate validates a single design token.
func (v *Validator) Validate(token *Token) *ValidationResult {
	result := &ValidationResult{
		Valid:  true,
		Errors: make([]*ValidationError, 0),
	}

	// Check required fields
	if token.Name == "" {
		result.Valid = false
		result.Errors = append(result.Errors, &ValidationError{
			TokenName: token.Name,
			Field:     "name",
			Message:   "token name is required",
		})
	}

	if token.Value == "" {
		result.Valid = false
		result.Errors = append(result.Errors, &ValidationError{
			TokenName: token.Name,
			Field:     "value",
			Message:   "token value is required",
		})
	}

	if token.Type == "" {
		result.Valid = false
		result.Errors = append(result.Errors, &ValidationError{
			TokenName: token.Name,
			Field:     "type",
			Message:   "token type is required",
		})
	}

	// Validate token type
	if token.Type != "" && !v.validTokenTypes[string(token.Type)] {
		result.Valid = false
		result.Errors = append(result.Errors, &ValidationError{
			TokenName: token.Name,
			Field:     "type",
			Message:   fmt.Sprintf("invalid token type '%s'", token.Type),
		})
	}

	// Type-specific validation
	if token.Type != "" && token.Value != "" {
		if err := v.validateTokenValue(string(token.Type), token.Value); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, &ValidationError{
				TokenName: token.Name,
				Field:     "value",
				Message:   err.Error(),
			})
		}
	}

	// Validate token name format (should be kebab-case or dot-notation)
	if token.Name != "" {
		validNameRegex := regexp.MustCompile(`^[a-z0-9]+([-.][a-z0-9]+)*$`)
		if !validNameRegex.MatchString(token.Name) {
			result.Valid = false
			result.Errors = append(result.Errors, &ValidationError{
				TokenName: token.Name,
				Field:     "name",
				Message:   "token name must be lowercase alphanumeric with dashes or dots as separators",
			})
		}
	}

	return result
}

// ValidateBatch validates a batch of design tokens.
func (v *Validator) ValidateBatch(tokens []*Token) *ValidationResult {
	result := &ValidationResult{
		Valid:  true,
		Errors: make([]*ValidationError, 0),
	}

	// Validate each token
	for _, token := range tokens {
		tokenResult := v.Validate(token)
		if !tokenResult.Valid {
			result.Valid = false
			result.Errors = append(result.Errors, tokenResult.Errors...)
		}
	}

	// Check for duplicate names
	nameMap := make(map[string]bool)
	for _, token := range tokens {
		if nameMap[token.Name] {
			result.Valid = false
			result.Errors = append(result.Errors, &ValidationError{
				TokenName: token.Name,
				Field:     "name",
				Message:   "duplicate token name in batch",
			})
		}
		nameMap[token.Name] = true
	}

	return result
}

// validateTokenValue validates the token value based on its type.
func (v *Validator) validateTokenValue(tokenType string, value string) error {
	switch tokenType {
	case "color":
		return v.validateColor(value)
	case "spacing", "font-size", "letter-spacing", "border-radius":
		return v.validateSizing(value)
	case "shadow":
		return v.validateShadow(value)
	case "font-weight":
		return v.validateFontWeight(value)
	case "line-height":
		return v.validateLineHeight(value)
	case "opacity":
		return v.validateOpacity(value)
	case "z-index":
		return v.validateZIndex(value)
	case "duration":
		return v.validateDuration(value)
	case "font-family", "easing", "typography":
		// These types accept any string value
		return nil
	default:
		// Unknown type - allow any value
		return nil
	}
}

// validateColor validates color values (hex, rgb, rgba, hsl, hsla, named colors).
func (v *Validator) validateColor(value string) error {
	value = strings.TrimSpace(value)

	// Check for hex colors
	if v.hexColorRegex.MatchString(value) {
		return nil
	}

	// Check for rgb/rgba colors
	if v.rgbColorRegex.MatchString(value) || v.rgbaColorRegex.MatchString(value) {
		return nil
	}

	// Check for hsl/hsla colors
	if v.hslColorRegex.MatchString(value) || v.hslaColorRegex.MatchString(value) {
		return nil
	}

	// Check for named colors
	namedColors := map[string]bool{
		"transparent": true, "black": true, "white": true, "red": true, "blue": true,
		"green": true, "yellow": true, "orange": true, "purple": true, "pink": true,
		"gray": true, "brown": true, "cyan": true, "magenta": true,
		// Add more as needed
	}
	if namedColors[strings.ToLower(value)] {
		return nil
	}

	return fmt.Errorf("invalid color format: %s (expected hex, rgb, rgba, hsl, hsla, or named color)", value)
}

// validateSizing validates size values (px, rem, em, %, etc.).
func (v *Validator) validateSizing(value string) error {
	value = strings.TrimSpace(value)

	if value == "0" || value == "auto" {
		return nil
	}

	if v.sizingRegex.MatchString(value) {
		return nil
	}

	return fmt.Errorf("invalid sizing format: %s (expected number with unit like px, rem, em, %%)", value)
}

// validateShadow validates shadow values.
func (v *Validator) validateShadow(value string) error {
	value = strings.TrimSpace(value)

	if value == "none" {
		return nil
	}

	// Simplified shadow validation - should contain offset-x, offset-y, blur, and optional spread
	if strings.Contains(value, "px") || strings.Contains(value, "rem") || strings.Contains(value, "em") {
		return nil
	}

	return fmt.Errorf("invalid shadow format: %s", value)
}

// validateFontWeight validates font-weight values.
func (v *Validator) validateFontWeight(value string) error {
	value = strings.TrimSpace(value)

	// Numeric values (100-900)
	numericRegex := regexp.MustCompile(`^[1-9]00$`)
	if numericRegex.MatchString(value) {
		return nil
	}

	// Named values
	namedWeights := map[string]bool{
		"normal": true, "bold": true, "lighter": true, "bolder": true,
		"100": true, "200": true, "300": true, "400": true, "500": true,
		"600": true, "700": true, "800": true, "900": true,
	}
	if namedWeights[value] {
		return nil
	}

	return fmt.Errorf("invalid font-weight: %s (expected 100-900 or normal/bold/lighter/bolder)", value)
}

// validateLineHeight validates line-height values.
func (v *Validator) validateLineHeight(value string) error {
	value = strings.TrimSpace(value)

	if value == "normal" {
		return nil
	}

	// Can be unitless number or size with unit
	unitlessRegex := regexp.MustCompile(`^\d+\.?\d*$`)
	if unitlessRegex.MatchString(value) || v.sizingRegex.MatchString(value) {
		return nil
	}

	return fmt.Errorf("invalid line-height: %s", value)
}

// validateOpacity validates opacity values (0-1).
func (v *Validator) validateOpacity(value string) error {
	value = strings.TrimSpace(value)

	opacityRegex := regexp.MustCompile(`^(0|1|0?\.\d+)$`)
	if opacityRegex.MatchString(value) {
		return nil
	}

	return fmt.Errorf("invalid opacity: %s (expected 0-1)", value)
}

// validateZIndex validates z-index values.
func (v *Validator) validateZIndex(value string) error {
	value = strings.TrimSpace(value)

	if value == "auto" {
		return nil
	}

	zIndexRegex := regexp.MustCompile(`^-?\d+$`)
	if zIndexRegex.MatchString(value) {
		return nil
	}

	return fmt.Errorf("invalid z-index: %s (expected integer or 'auto')", value)
}

// validateDuration validates duration values (for transitions/animations).
func (v *Validator) validateDuration(value string) error {
	value = strings.TrimSpace(value)

	durationRegex := regexp.MustCompile(`^\d+\.?\d*(ms|s)$`)
	if durationRegex.MatchString(value) {
		return nil
	}

	return fmt.Errorf("invalid duration: %s (expected number with ms or s)", value)
}
