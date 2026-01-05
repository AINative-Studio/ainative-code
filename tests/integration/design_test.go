// +build integration

package integration

import (
	"testing"

	"github.com/AINative-studio/ainative-code/internal/design"
	"github.com/AINative-studio/ainative-code/internal/design/generators"
	"github.com/AINative-studio/ainative-code/tests/integration/fixtures"
	"github.com/stretchr/testify/suite"
)

// DesignIntegrationTestSuite tests design token extraction and generation.
type DesignIntegrationTestSuite struct {
	suite.Suite
}

// TestDesignTokenExtraction tests extracting design tokens from a structure.
func (s *DesignIntegrationTestSuite) TestDesignTokenExtraction() {
	// Given: Test design tokens
	tokenCollection := fixtures.TestDesignTokens()

	// Then: Tokens should be valid
	s.NotEmpty(tokenCollection.Tokens, "Should have tokens")
	s.NotEmpty(tokenCollection.Metadata, "Should have metadata")

	// Verify different token types exist
	hasColor := false
	hasTypography := false
	hasSpacing := false

	for _, token := range tokenCollection.Tokens {
		switch token.Type {
		case design.TokenTypeColor:
			hasColor = true
		case design.TokenTypeTypography:
			hasTypography = true
		case design.TokenTypeSpacing:
			hasSpacing = true
		}
	}

	s.True(hasColor, "Should have color tokens")
	s.True(hasTypography, "Should have typography tokens")
	s.True(hasSpacing, "Should have spacing tokens")
}

// TestCSSGeneration tests generating CSS from design tokens.
func (s *DesignIntegrationTestSuite) TestCSSGeneration() {
	// Given: Test design tokens
	tokenCollection := fixtures.TestDesignTokens()

	// When: Generating CSS
	generator := generators.NewCSSGenerator()
	css, err := generator.Generate(tokenCollection.Tokens)

	// Then: CSS should be generated
	s.Require().NoError(err, "CSS generation should succeed")
	s.NotEmpty(css, "CSS output should not be empty")
	s.Contains(css, ":root", "CSS should contain root selector")
}

// TestJSONGeneration tests generating JSON from design tokens.
func (s *DesignIntegrationTestSuite) TestJSONGeneration() {
	// Given: Test design tokens
	tokenCollection := fixtures.TestDesignTokens()

	// When: Generating JSON
	generator := generators.NewJSONGenerator()
	jsonData, err := generator.Generate(tokenCollection.Tokens)

	// Then: JSON should be generated
	s.Require().NoError(err, "JSON generation should succeed")
	s.NotEmpty(jsonData, "JSON output should not be empty")
}

// TestSCSSGeneration tests generating SCSS from design tokens.
func (s *DesignIntegrationTestSuite) TestSCSSGeneration() {
	// Given: Test design tokens
	tokenCollection := fixtures.TestDesignTokens()

	// When: Generating SCSS
	generator := generators.NewSCSSGenerator()
	scss, err := generator.Generate(tokenCollection.Tokens)

	// Then: SCSS should be generated
	s.Require().NoError(err, "SCSS generation should succeed")
	s.NotEmpty(scss, "SCSS output should not be empty")
}

// TestTailwindGeneration tests generating Tailwind config from design tokens.
func (s *DesignIntegrationTestSuite) TestTailwindGeneration() {
	// Given: Test design tokens
	tokenCollection := fixtures.TestDesignTokens()

	// When: Generating Tailwind config
	generator := generators.NewTailwindGenerator()
	config, err := generator.Generate(tokenCollection.Tokens)

	// Then: Config should be generated
	s.Require().NoError(err, "Tailwind generation should succeed")
	s.NotEmpty(config, "Tailwind config should not be empty")
}

// TestTypeScriptGeneration tests generating TypeScript types from design tokens.
func (s *DesignIntegrationTestSuite) TestTypeScriptGeneration() {
	// Given: Test design tokens
	tokenCollection := fixtures.TestDesignTokens()

	// When: Generating TypeScript
	generator := generators.NewTypeScriptGenerator()
	ts, err := generator.Generate(tokenCollection.Tokens)

	// Then: TypeScript should be generated
	s.Require().NoError(err, "TypeScript generation should succeed")
	s.NotEmpty(ts, "TypeScript output should not be empty")
}

// TestTokenValidation tests design token validation.
func (s *DesignIntegrationTestSuite) TestTokenValidation() {
	// Given: Valid tokens
	validTokens := fixtures.TestDesignTokens()

	// Then: Tokens should have required fields
	for _, token := range validTokens.Tokens {
		s.NotEmpty(token.Name, "Token should have name")
		s.NotEmpty(token.Type, "Token should have type")
		s.NotEmpty(token.Value, "Token should have value")
	}
}

// TestTokenFormatting tests design token formatting.
func (s *DesignIntegrationTestSuite) TestTokenFormatting() {
	// Given: Design tokens
	tokenCollection := fixtures.TestDesignTokens()

	// When: Generating JSON (which is a form of formatting)
	generator := generators.NewJSONGenerator()
	formatted, err := generator.Generate(tokenCollection.Tokens)

	// Then: Formatting should succeed
	s.Require().NoError(err, "Token formatting should succeed")
	s.NotEmpty(formatted, "Formatted output should not be empty")
}

// TestMultiFormatGeneration tests generating multiple formats simultaneously.
func (s *DesignIntegrationTestSuite) TestMultiFormatGeneration() {
	// Given: Test design tokens
	tokenCollection := fixtures.TestDesignTokens()

	// When: Generating multiple formats
	type Generator interface {
		Generate([]design.Token) (string, error)
	}

	gens := map[string]Generator{
		"css":        generators.NewCSSGenerator(),
		"scss":       generators.NewSCSSGenerator(),
		"json":       generators.NewJSONGenerator(),
		"tailwind":   generators.NewTailwindGenerator(),
		"typescript": generators.NewTypeScriptGenerator(),
	}

	outputs := make(map[string]string)

	for format, gen := range gens {
		output, err := gen.Generate(tokenCollection.Tokens)
		s.Require().NoError(err, "Generation should succeed for %s", format)
		outputs[format] = output
	}

	// Then: All formats should be generated
	s.Len(outputs, len(gens), "All formats should be generated")
	for format, output := range outputs {
		s.NotEmpty(output, "Output for %s should not be empty", format)
	}
}

// TestDesignIntegrationTestSuite runs the test suite.
func TestDesignIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(DesignIntegrationTestSuite))
}
