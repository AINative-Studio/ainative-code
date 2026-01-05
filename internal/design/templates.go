package design

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

// TemplateEngine provides template-based code generation functionality
type TemplateEngine struct {
	templates map[string]*template.Template
	funcMap   template.FuncMap
}

// NewTemplateEngine creates a new template engine with built-in functions
func NewTemplateEngine() *TemplateEngine {
	funcMap := template.FuncMap{
		"kebabCase":  toKebabCase,
		"camelCase":  toCamelCase,
		"pascalCase": toPascalCase,
		"snakeCase":  toSnakeCase,
		"upper":      strings.ToUpper,
		"lower":      strings.ToLower,
		"trim":       strings.TrimSpace,
		"join":       strings.Join,
		"quote":      quote,
		"indent":     indent,
	}

	return &TemplateEngine{
		templates: make(map[string]*template.Template),
		funcMap:   funcMap,
	}
}

// RegisterTemplate registers a new template with the engine
func (te *TemplateEngine) RegisterTemplate(name string, content string) error {
	tmpl, err := template.New(name).Funcs(te.funcMap).Parse(content)
	if err != nil {
		return fmt.Errorf("failed to parse template '%s': %w", name, err)
	}
	te.templates[name] = tmpl
	return nil
}

// Execute executes a registered template with the given data
func (te *TemplateEngine) Execute(name string, data interface{}) (string, error) {
	tmpl, ok := te.templates[name]
	if !ok {
		return "", fmt.Errorf("template '%s' not found", name)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template '%s': %w", name, err)
	}

	return buf.String(), nil
}

// ExecuteString executes an inline template string with the given data
func (te *TemplateEngine) ExecuteString(templateStr string, data interface{}) (string, error) {
	tmpl, err := template.New("inline").Funcs(te.funcMap).Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse inline template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute inline template: %w", err)
	}

	return buf.String(), nil
}

// Helper functions for template transformations

// toKebabCase converts a string to kebab-case
func toKebabCase(s string) string {
	// First handle underscores and dots
	s = strings.ReplaceAll(s, "_", "-")
	s = strings.ReplaceAll(s, ".", "-")

	// Handle camelCase by adding dashes before uppercase letters
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			// Don't add dash if previous char was already a dash or uppercase
			prev := rune(s[i-1])
			if prev != '-' && !(prev >= 'A' && prev <= 'Z') {
				result.WriteRune('-')
			}
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// toCamelCase converts a string to camelCase
func toCamelCase(s string) string {
	s = strings.ReplaceAll(s, "-", " ")
	s = strings.ReplaceAll(s, "_", " ")
	s = strings.ReplaceAll(s, ".", " ")
	words := strings.Fields(s)
	if len(words) == 0 {
		return ""
	}

	result := strings.ToLower(words[0])
	for i := 1; i < len(words); i++ {
		result += strings.Title(strings.ToLower(words[i]))
	}
	return result
}

// toPascalCase converts a string to PascalCase
func toPascalCase(s string) string {
	s = strings.ReplaceAll(s, "-", " ")
	s = strings.ReplaceAll(s, "_", " ")
	s = strings.ReplaceAll(s, ".", " ")
	words := strings.Fields(s)
	var result string
	for _, word := range words {
		result += strings.Title(strings.ToLower(word))
	}
	return result
}

// toSnakeCase converts a string to snake_case
func toSnakeCase(s string) string {
	// First handle dashes and dots
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, ".", "_")

	// Handle camelCase by adding underscores before uppercase letters
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			// Don't add underscore if previous char was already an underscore or uppercase
			prev := rune(s[i-1])
			if prev != '_' && !(prev >= 'A' && prev <= 'Z') {
				result.WriteRune('_')
			}
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// quote wraps a string in double quotes
func quote(s string) string {
	return fmt.Sprintf("\"%s\"", s)
}

// indent indents each line of a string by the specified number of spaces
func indent(spaces int, s string) string {
	prefix := strings.Repeat(" ", spaces)
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = prefix + line
		}
	}
	return strings.Join(lines, "\n")
}

// Built-in template definitions

// TailwindTemplate is the template for Tailwind config generation
const TailwindTemplate = `/** @type {import('tailwindcss').Config} */
module.exports = {
  theme: {
    extend: {
{{- if .Colors }}
      colors: {
{{- range .Colors }}
        {{ quote .Name | kebabCase }}: {{ quote .Value }},
{{- end }}
      },
{{- end }}
{{- if .Spacing }}
      spacing: {
{{- range .Spacing }}
        {{ quote .Name | kebabCase }}: {{ quote .Value }},
{{- end }}
      },
{{- end }}
{{- if .FontFamily }}
      fontFamily: {
{{- range .FontFamily }}
        {{ quote .Name | kebabCase }}: {{ .Value }},
{{- end }}
      },
{{- end }}
{{- if .FontSize }}
      fontSize: {
{{- range .FontSize }}
        {{ quote .Name | kebabCase }}: {{ quote .Value }},
{{- end }}
      },
{{- end }}
{{- if .BorderRadius }}
      borderRadius: {
{{- range .BorderRadius }}
        {{ quote .Name | kebabCase }}: {{ quote .Value }},
{{- end }}
      },
{{- end }}
{{- if .BoxShadow }}
      boxShadow: {
{{- range .BoxShadow }}
        {{ quote .Name | kebabCase }}: {{ quote .Value }},
{{- end }}
      },
{{- end }}
    },
  },
  plugins: [],
}
`

// CSSTemplate is the template for CSS variables generation
const CSSTemplate = `/**
 * Design Tokens - CSS Custom Properties
 * Generated from design system
 */

:root {
{{- range .Tokens }}
  --{{ .Name | kebabCase }}: {{ .Value }};
{{- end }}
}
`

// SCSSTemplate is the template for SCSS variables generation
const SCSSTemplate = `/**
 * Design Tokens - SCSS Variables
 * Generated from design system
 */

{{- range .Tokens }}
${{ .Name | kebabCase }}: {{ .Value }};
{{- end }}
`

// TypeScriptTemplate is the template for TypeScript constants generation
const TypeScriptTemplate = `/**
 * Design Tokens - TypeScript Constants
 * Generated from design system
 */

export const DesignTokens = {
{{- range .Tokens }}
  {{ .Name | camelCase }}: {{ quote .Value }},
{{- end }}
} as const;

export type DesignToken = typeof DesignTokens;
`

// JavaScriptTemplate is the template for JavaScript constants generation
const JavaScriptTemplate = `/**
 * Design Tokens - JavaScript Constants
 * Generated from design system
 */

export const DesignTokens = {
{{- range .Tokens }}
  {{ .Name | camelCase }}: {{ quote .Value }},
{{- end }}
};
`
