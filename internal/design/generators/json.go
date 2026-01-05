package generators

import (
	"encoding/json"

	"github.com/AINative-studio/ainative-code/internal/design"
)

// JSONGenerator generates JSON output from design tokens
type JSONGenerator struct {
	pretty bool
}

// NewJSONGenerator creates a new JSON generator
func NewJSONGenerator(pretty bool) *JSONGenerator {
	return &JSONGenerator{
		pretty: pretty,
	}
}

// Generate generates JSON from design tokens
func (g *JSONGenerator) Generate(tokens []design.Token) (string, error) {
	output := map[string]interface{}{
		"tokens": tokens,
		"count":  len(tokens),
	}

	var bytes []byte
	var err error

	if g.pretty {
		bytes, err = json.MarshalIndent(output, "", "  ")
	} else {
		bytes, err = json.Marshal(output)
	}

	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

// Name returns the generator name
func (g *JSONGenerator) Name() string {
	return "JSON Generator"
}

// SupportedFormats returns supported output formats
func (g *JSONGenerator) SupportedFormats() []string {
	return []string{"json"}
}
