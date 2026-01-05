package generators

import (
	"fmt"

	"github.com/AINative-studio/ainative-code/internal/design"
)

// TypeScriptGenerator generates TypeScript constants from design tokens
type TypeScriptGenerator struct {
	engine *design.TemplateEngine
}

// NewTypeScriptGenerator creates a new TypeScript constants generator
func NewTypeScriptGenerator() *TypeScriptGenerator {
	engine := design.NewTemplateEngine()
	if err := engine.RegisterTemplate("typescript", design.TypeScriptTemplate); err != nil {
		panic(fmt.Sprintf("failed to register typescript template: %v", err))
	}
	if err := engine.RegisterTemplate("javascript", design.JavaScriptTemplate); err != nil {
		panic(fmt.Sprintf("failed to register javascript template: %v", err))
	}

	return &TypeScriptGenerator{
		engine: engine,
	}
}

// Generate generates TypeScript/JavaScript constants from design tokens
func (g *TypeScriptGenerator) Generate(tokens []design.Token, format string) (string, error) {
	data := map[string]interface{}{
		"Tokens": tokens,
	}

	templateName := "typescript"
	if format == "js" || format == "javascript" {
		templateName = "javascript"
	}

	return g.engine.Execute(templateName, data)
}

// Name returns the generator name
func (g *TypeScriptGenerator) Name() string {
	return "TypeScript/JavaScript Constants Generator"
}

// SupportedFormats returns supported output formats
func (g *TypeScriptGenerator) SupportedFormats() []string {
	return []string{"typescript", "ts", "javascript", "js"}
}
