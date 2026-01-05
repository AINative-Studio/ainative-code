package design

import (
	"testing"
)

func TestValidateColor(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name        string
		value       string
		expectValid bool
	}{
		// Hex colors
		{"valid 3-digit hex", "#fff", true},
		{"valid 6-digit hex", "#ffffff", true},
		{"valid 8-digit hex with alpha", "#ffffffff", true},
		{"uppercase hex", "#FFFFFF", true},
		{"mixed case hex", "#FfFfFf", true},
		{"invalid hex - too short", "#ff", false},
		{"invalid hex - too long", "#fffffffff", false},
		{"invalid hex - no hash", "ffffff", false},

		// RGB colors
		{"valid rgb", "rgb(255, 255, 255)", true},
		{"valid rgb no spaces", "rgb(255,255,255)", true},
		{"invalid rgb - out of range", "rgb(256, 255, 255)", false},

		// RGBA colors
		{"valid rgba", "rgba(255, 255, 255, 0.5)", true},
		{"valid rgba integer alpha", "rgba(255, 255, 255, 1)", true},
		{"invalid rgba - missing alpha", "rgba(255, 255, 255)", false},

		// HSL colors
		{"valid hsl", "hsl(180, 50%, 50%)", true},
		{"invalid hsl - missing percent", "hsl(180, 50, 50)", false},

		// HSLA colors
		{"valid hsla", "hsla(180, 50%, 50%, 0.5)", true},

		// Named colors
		{"named color - white", "white", true},
		{"named color - black", "black", true},
		{"named color - red", "red", true},
		{"named color - transparent", "transparent", true},
		{"invalid named color", "notacolor", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateColor(tt.value)
			isValid := err == nil

			if isValid != tt.expectValid {
				t.Errorf("validateColor(%q) = %v, want valid=%v (error: %v)", tt.value, isValid, tt.expectValid, err)
			}
		})
	}
}

func TestValidateSizing(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name        string
		value       string
		expectValid bool
	}{
		{"valid px", "16px", true},
		{"valid rem", "1.5rem", true},
		{"valid em", "2em", true},
		{"valid percent", "100%", true},
		{"valid vh", "50vh", true},
		{"valid vw", "75vw", true},
		{"valid zero", "0", true},
		{"valid auto", "auto", true},
		{"negative value", "-10px", true},
		{"decimal value", "1.5px", true},
		{"invalid - no unit", "16", false},
		{"invalid - unknown unit", "16xyz", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateSizing(tt.value)
			isValid := err == nil

			if isValid != tt.expectValid {
				t.Errorf("validateSizing(%q) = %v, want valid=%v (error: %v)", tt.value, isValid, tt.expectValid, err)
			}
		})
	}
}

func TestValidateFontWeight(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name        string
		value       string
		expectValid bool
	}{
		{"valid normal", "normal", true},
		{"valid bold", "bold", true},
		{"valid lighter", "lighter", true},
		{"valid bolder", "bolder", true},
		{"valid 100", "100", true},
		{"valid 400", "400", true},
		{"valid 700", "700", true},
		{"valid 900", "900", true},
		{"invalid 50", "50", false},
		{"invalid 1000", "1000", false},
		{"invalid text", "heavy", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateFontWeight(tt.value)
			isValid := err == nil

			if isValid != tt.expectValid {
				t.Errorf("validateFontWeight(%q) = %v, want valid=%v (error: %v)", tt.value, isValid, tt.expectValid, err)
			}
		})
	}
}

func TestValidateLineHeight(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name        string
		value       string
		expectValid bool
	}{
		{"valid normal", "normal", true},
		{"valid unitless", "1.5", true},
		{"valid integer unitless", "2", true},
		{"valid px", "24px", true},
		{"valid rem", "1.5rem", true},
		{"valid percent", "150%", true},
		{"invalid text", "large", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateLineHeight(tt.value)
			isValid := err == nil

			if isValid != tt.expectValid {
				t.Errorf("validateLineHeight(%q) = %v, want valid=%v (error: %v)", tt.value, isValid, tt.expectValid, err)
			}
		})
	}
}

func TestValidateOpacity(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name        string
		value       string
		expectValid bool
	}{
		{"valid 0", "0", true},
		{"valid 1", "1", true},
		{"valid 0.5", "0.5", true},
		{"valid .5", ".5", true},
		{"valid 0.75", "0.75", true},
		{"invalid 1.5", "1.5", false},
		{"invalid negative", "-0.5", false},
		{"invalid text", "half", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateOpacity(tt.value)
			isValid := err == nil

			if isValid != tt.expectValid {
				t.Errorf("validateOpacity(%q) = %v, want valid=%v (error: %v)", tt.value, isValid, tt.expectValid, err)
			}
		})
	}
}

func TestValidateZIndex(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name        string
		value       string
		expectValid bool
	}{
		{"valid auto", "auto", true},
		{"valid positive", "10", true},
		{"valid negative", "-1", true},
		{"valid zero", "0", true},
		{"invalid decimal", "1.5", false},
		{"invalid text", "high", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateZIndex(tt.value)
			isValid := err == nil

			if isValid != tt.expectValid {
				t.Errorf("validateZIndex(%q) = %v, want valid=%v (error: %v)", tt.value, isValid, tt.expectValid, err)
			}
		})
	}
}

func TestValidateDuration(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name        string
		value       string
		expectValid bool
	}{
		{"valid ms", "100ms", true},
		{"valid s", "1s", true},
		{"valid decimal ms", "250.5ms", true},
		{"valid decimal s", "0.5s", true},
		{"invalid no unit", "100", false},
		{"invalid wrong unit", "100px", false},
		{"invalid text", "fast", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateDuration(tt.value)
			isValid := err == nil

			if isValid != tt.expectValid {
				t.Errorf("validateDuration(%q) = %v, want valid=%v (error: %v)", tt.value, isValid, tt.expectValid, err)
			}
		})
	}
}

func TestValidateToken(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name        string
		token       *Token
		expectValid bool
		expectError int
	}{
		{
			name: "valid color token",
			token: &Token{
				Name:     "primary-color",
				Value:    "#007bff",
				Type:     "color",
				Category: "colors",
			},
			expectValid: true,
			expectError: 0,
		},
		{
			name: "valid spacing token",
			token: &Token{
				Name:     "spacing-base",
				Value:    "16px",
				Type:     "spacing",
				Category: "spacing",
			},
			expectValid: true,
			expectError: 0,
		},
		{
			name: "missing name",
			token: &Token{
				Name:     "",
				Value:    "#007bff",
				Type:     "color",
				Category: "colors",
			},
			expectValid: false,
			expectError: 1,
		},
		{
			name: "missing value",
			token: &Token{
				Name:     "primary-color",
				Value:    "",
				Type:     "color",
				Category: "colors",
			},
			expectValid: false,
			expectError: 1,
		},
		{
			name: "missing type",
			token: &Token{
				Name:     "primary-color",
				Value:    "#007bff",
				Type:     "",
				Category: "colors",
			},
			expectValid: false,
			expectError: 1,
		},
		{
			name: "invalid type",
			token: &Token{
				Name:     "primary-color",
				Value:    "#007bff",
				Type:     "invalid-type",
				Category: "colors",
			},
			expectValid: false,
			expectError: 1,
		},
		{
			name: "invalid name format - uppercase",
			token: &Token{
				Name:     "PrimaryColor",
				Value:    "#007bff",
				Type:     "color",
				Category: "colors",
			},
			expectValid: false,
			expectError: 1,
		},
		{
			name: "invalid name format - spaces",
			token: &Token{
				Name:     "primary color",
				Value:    "#007bff",
				Type:     "color",
				Category: "colors",
			},
			expectValid: false,
			expectError: 1,
		},
		{
			name: "valid name with dots",
			token: &Token{
				Name:     "colors.primary.base",
				Value:    "#007bff",
				Type:     "color",
				Category: "colors",
			},
			expectValid: true,
			expectError: 0,
		},
		{
			name: "valid name with dashes",
			token: &Token{
				Name:     "primary-color-base",
				Value:    "#007bff",
				Type:     "color",
				Category: "colors",
			},
			expectValid: true,
			expectError: 0,
		},
		{
			name: "invalid color value",
			token: &Token{
				Name:     "primary-color",
				Value:    "not-a-color",
				Type:     "color",
				Category: "colors",
			},
			expectValid: false,
			expectError: 1,
		},
		{
			name: "multiple errors",
			token: &Token{
				Name:     "",
				Value:    "",
				Type:     "",
				Category: "",
			},
			expectValid: false,
			expectError: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.Validate(tt.token)

			if result.Valid != tt.expectValid {
				t.Errorf("Validate() valid = %v, want %v", result.Valid, tt.expectValid)
			}

			if len(result.Errors) != tt.expectError {
				t.Errorf("Validate() errors = %d, want %d", len(result.Errors), tt.expectError)
				for _, err := range result.Errors {
					t.Logf("  - %s", err.Error())
				}
			}
		})
	}
}

func TestValidateBatch(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name        string
		tokens      []*Token
		expectValid bool
		expectError int
	}{
		{
			name: "valid batch",
			tokens: []*Token{
				{
					Name:     "primary-color",
					Value:    "#007bff",
					Type:     "color",
					Category: "colors",
				},
				{
					Name:     "spacing-base",
					Value:    "16px",
					Type:     "spacing",
					Category: "spacing",
				},
			},
			expectValid: true,
			expectError: 0,
		},
		{
			name: "duplicate names",
			tokens: []*Token{
				{
					Name:     "primary-color",
					Value:    "#007bff",
					Type:     "color",
					Category: "colors",
				},
				{
					Name:     "primary-color",
					Value:    "#0056b3",
					Type:     "color",
					Category: "colors",
				},
			},
			expectValid: false,
			expectError: 1,
		},
		{
			name: "mixed valid and invalid",
			tokens: []*Token{
				{
					Name:     "primary-color",
					Value:    "#007bff",
					Type:     "color",
					Category: "colors",
				},
				{
					Name:     "invalid-token",
					Value:    "",
					Type:     "",
					Category: "",
				},
			},
			expectValid: false,
			expectError: 2,
		},
		{
			name: "empty batch",
			tokens: []*Token{},
			expectValid: true,
			expectError: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateBatch(tt.tokens)

			if result.Valid != tt.expectValid {
				t.Errorf("ValidateBatch() valid = %v, want %v", result.Valid, tt.expectValid)
			}

			if len(result.Errors) != tt.expectError {
				t.Errorf("ValidateBatch() errors = %d, want %d", len(result.Errors), tt.expectError)
				for _, err := range result.Errors {
					t.Logf("  - %s", err.Error())
				}
			}
		})
	}
}

func TestConflictResolution(t *testing.T) {
	tests := []struct {
		resolution ConflictResolutionStrategyUpload
		expected   string
	}{
		{ConflictOverwrite, "overwrite"},
		{ConflictMerge, "merge"},
		{ConflictSkip, "skip"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if string(tt.resolution) != tt.expected {
				t.Errorf("ConflictResolution = %v, want %v", tt.resolution, tt.expected)
			}
		})
	}
}

func TestValidationError(t *testing.T) {
	err := &ValidationError{
		TokenName: "test-token",
		Field:     "value",
		Message:   "invalid value",
	}

	expected := "token 'test-token': value - invalid value"
	if err.Error() != expected {
		t.Errorf("ValidationError.Error() = %q, want %q", err.Error(), expected)
	}
}
