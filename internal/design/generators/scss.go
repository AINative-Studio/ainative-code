package generators

import (
	"fmt"

	"github.com/AINative-studio/ainative-code/internal/design"
)

// SCSSGenerator generates SCSS variables from design tokens
type SCSSGenerator struct {
	engine *design.TemplateEngine
}

// NewSCSSGenerator creates a new SCSS variables generator
func NewSCSSGenerator() *SCSSGenerator {
	engine := design.NewTemplateEngine()
	if err := engine.RegisterTemplate("scss", design.SCSSTemplate); err != nil {
		panic(fmt.Sprintf("failed to register scss template: %v", err))
	}

	return &SCSSGenerator{
		engine: engine,
	}
}

// Generate generates SCSS variables from design tokens
func (g *SCSSGenerator) Generate(tokens []design.Token) (string, error) {
	data := map[string]interface{}{
		"Tokens": tokens,
	}
	return g.engine.Execute("scss", data)
}

// Name returns the generator name
func (g *SCSSGenerator) Name() string {
	return "SCSS Variables Generator"
}

// SupportedFormats returns supported output formats
func (g *SCSSGenerator) SupportedFormats() []string {
	return []string{"scss", "sass"}
}
