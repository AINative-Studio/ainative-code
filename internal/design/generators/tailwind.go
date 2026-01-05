package generators

import (
	"fmt"
	"strings"

	"github.com/AINative-studio/ainative-code/internal/design"
)

// TailwindGenerator generates Tailwind CSS configuration from design tokens
type TailwindGenerator struct {
	engine *design.TemplateEngine
}

// NewTailwindGenerator creates a new Tailwind config generator
func NewTailwindGenerator() *TailwindGenerator {
	engine := design.NewTemplateEngine()
	if err := engine.RegisterTemplate("tailwind", design.TailwindTemplate); err != nil {
		panic(fmt.Sprintf("failed to register tailwind template: %v", err))
	}

	return &TailwindGenerator{
		engine: engine,
	}
}

// Generate generates Tailwind config from design tokens
func (g *TailwindGenerator) Generate(tokens []design.Token) (string, error) {
	data := g.prepareData(tokens)
	return g.engine.Execute("tailwind", data)
}

// prepareData organizes tokens by Tailwind theme categories
func (g *TailwindGenerator) prepareData(tokens []design.Token) map[string]interface{} {
	data := map[string]interface{}{
		"Colors":       []design.Token{},
		"Spacing":      []design.Token{},
		"FontFamily":   []design.Token{},
		"FontSize":     []design.Token{},
		"BorderRadius": []design.Token{},
		"BoxShadow":    []design.Token{},
	}

	for _, token := range tokens {
		switch token.Type {
		case "color":
			data["Colors"] = append(data["Colors"].([]design.Token), token)
		case "spacing":
			data["Spacing"] = append(data["Spacing"].([]design.Token), token)
		case "font-family":
			// Tailwind expects font families as arrays
			token.Value = g.formatFontFamily(token.Value)
			data["FontFamily"] = append(data["FontFamily"].([]design.Token), token)
		case "font-size":
			data["FontSize"] = append(data["FontSize"].([]design.Token), token)
		case "border-radius":
			data["BorderRadius"] = append(data["BorderRadius"].([]design.Token), token)
		case "shadow":
			data["BoxShadow"] = append(data["BoxShadow"].([]design.Token), token)
		case "typography":
			// Try to categorize typography tokens
			if strings.Contains(strings.ToLower(token.Name), "font-family") {
				token.Value = g.formatFontFamily(token.Value)
				data["FontFamily"] = append(data["FontFamily"].([]design.Token), token)
			} else if strings.Contains(strings.ToLower(token.Name), "font-size") {
				data["FontSize"] = append(data["FontSize"].([]design.Token), token)
			}
		}
	}

	return data
}

// formatFontFamily formats font-family values for Tailwind
func (g *TailwindGenerator) formatFontFamily(value string) string {
	// Split by comma and wrap each font in quotes
	fonts := strings.Split(value, ",")
	var formatted []string
	for _, font := range fonts {
		font = strings.TrimSpace(font)
		// Remove existing quotes if present
		font = strings.Trim(font, "\"'")
		formatted = append(formatted, fmt.Sprintf("'%s'", font))
	}
	return "[" + strings.Join(formatted, ", ") + "]"
}

// Name returns the generator name
func (g *TailwindGenerator) Name() string {
	return "Tailwind CSS Config Generator"
}

// SupportedFormats returns supported output formats
func (g *TailwindGenerator) SupportedFormats() []string {
	return []string{"tailwind", "tw"}
}
