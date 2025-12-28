// Package branding contains AINative Code branding constants and utilities.
// © 2024 AINative Studio. All rights reserved.
package branding

import (
	"strings"
	"testing"
)

func TestConstants(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{"ProductName", ProductName, "AINative Code"},
		{"ProductTagline", ProductTagline, "AI-Native Development, Natively"},
		{"CompanyName", CompanyName, "AINative Studio"},
		{"BinaryName", BinaryName, "ainative-code"},
		{"ConfigFileName", ConfigFileName, ".ainative-code.yaml"},
		{"EnvPrefix", EnvPrefix, "AINATIVE_CODE"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != tt.expected {
				t.Errorf("%s = %v, want %v", tt.name, tt.value, tt.expected)
			}
		})
	}
}

func TestBrandColors(t *testing.T) {
	colors := map[string]string{
		"Primary":   ColorPrimary,
		"Secondary": ColorSecondary,
		"Success":   ColorSuccess,
		"Error":     ColorError,
		"Accent":    ColorAccent,
		"Warning":   ColorWarning,
		"Info":      ColorInfo,
	}

	for name, color := range colors {
		t.Run(name, func(t *testing.T) {
			// Check if color is a valid hex code
			if !strings.HasPrefix(color, "#") {
				t.Errorf("Color %s (%s) does not start with #", name, color)
			}
			if len(color) != 7 {
				t.Errorf("Color %s (%s) is not 7 characters long", name, color)
			}
		})
	}
}

func TestServiceEndpoints(t *testing.T) {
	endpoints := map[string]string{
		"Auth":   AuthServiceURL,
		"ZeroDB": ZeroDBServiceURL,
		"Design": DesignServiceURL,
		"Strapi": StrapiServiceURL,
		"RLHF":   RLHFServiceURL,
	}

	for name, url := range endpoints {
		t.Run(name, func(t *testing.T) {
			// Check if URL starts with https://
			if !strings.HasPrefix(url, "https://") {
				t.Errorf("Service endpoint %s (%s) does not start with https://", name, url)
			}
		})
	}
}

func TestGetFullProductName(t *testing.T) {
	expected := "AINative Code - AI-Native Development, Natively"
	result := GetFullProductName()
	if result != expected {
		t.Errorf("GetFullProductName() = %v, want %v", result, expected)
	}
}

func TestGetVersionString(t *testing.T) {
	result := GetVersionString()
	if !strings.HasPrefix(result, "AINative Code v") {
		t.Errorf("GetVersionString() should start with 'AINative Code v', got %v", result)
	}
}

func TestGetCopyrightNotice(t *testing.T) {
	result := GetCopyrightNotice()
	if !strings.Contains(result, "2024") {
		t.Errorf("Copyright notice should contain year 2024, got %v", result)
	}
	if !strings.Contains(result, "AINative Studio") {
		t.Errorf("Copyright notice should contain 'AINative Studio', got %v", result)
	}
}

func TestGetWelcomeMessage(t *testing.T) {
	result := GetWelcomeMessage()

	// Check for key components
	if !strings.Contains(result, "AINative Code") {
		t.Error("Welcome message should contain 'AINative Code'")
	}
	if !strings.Contains(result, "AI-Native Development, Natively") {
		t.Error("Welcome message should contain tagline")
	}
	if !strings.Contains(result, "2024") {
		t.Error("Welcome message should contain copyright year")
	}
	if !strings.Contains(result, Version) {
		t.Error("Welcome message should contain version")
	}
	if !strings.Contains(result, DocsURL) {
		t.Error("Welcome message should contain docs URL")
	}
}

// TestNoCrushReferences ensures there are no references to "Crush" in the branding package
func TestNoCrushReferences(t *testing.T) {
	values := []string{
		ProductName,
		ProductTagline,
		ProductDescription,
		CompanyName,
		Copyright,
		WebsiteURL,
		DocsURL,
		SupportEmail,
		RepositoryURL,
		BinaryName,
		ConfigFileName,
		ConfigDirName,
		DataDirName,
		EnvPrefix,
	}

	for _, value := range values {
		if strings.Contains(strings.ToLower(value), "crush") {
			t.Errorf("Found 'crush' reference in branding: %s", value)
		}
	}
}
