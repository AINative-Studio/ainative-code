package generators

import (
	"fmt"

	"github.com/AINative-studio/ainative-code/internal/design"
)

// CSSGenerator generates CSS custom properties from design tokens
type CSSGenerator struct {
	engine *design.TemplateEngine
}

// NewCSSGenerator creates a new CSS variables generator
func NewCSSGenerator() *CSSGenerator {
	engine := design.NewTemplateEngine()
	if err := engine.RegisterTemplate("css", design.CSSTemplate); err != nil {
		panic(fmt.Sprintf("failed to register css template: %v", err))
	}

	return &CSSGenerator{
		engine: engine,
	}
}

// Generate generates CSS custom properties from design tokens
func (g *CSSGenerator) Generate(tokens []design.Token) (string, error) {
	data := map[string]interface{}{
		"Tokens": tokens,
	}
	return g.engine.Execute("css", data)
}

// Name returns the generator name
func (g *CSSGenerator) Name() string {
	return "CSS Variables Generator"
}

// SupportedFormats returns supported output formats
func (g *CSSGenerator) SupportedFormats() []string {
	return []string{"css"}
}
