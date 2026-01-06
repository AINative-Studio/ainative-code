package benchmark

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/AINative-studio/ainative-code/internal/design"
	"github.com/AINative-studio/ainative-code/internal/design/generators"
)

// BenchmarkTokenExtraction measures design token extraction performance
func BenchmarkTokenExtraction(b *testing.B) {
	helper := NewTestHelper(b)
	defer helper.Cleanup()

	// Create test CSS file
	cssContent := `
:root {
  --color-primary: #007bff;
  --color-secondary: #6c757d;
  --color-success: #28a745;
  --color-danger: #dc3545;
  --color-warning: #ffc107;
  --color-info: #17a2b8;

  --font-family-base: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto;
  --font-size-base: 1rem;
  --font-size-lg: 1.25rem;
  --font-size-sm: 0.875rem;

  --spacing-1: 0.25rem;
  --spacing-2: 0.5rem;
  --spacing-3: 1rem;
  --spacing-4: 1.5rem;
  --spacing-5: 3rem;
}
`
	cssFile := filepath.Join(helper.TempDir, "tokens.css")
	if err := os.WriteFile(cssFile, []byte(cssContent), 0644); err != nil {
		b.Fatalf("Failed to write CSS file: %v", err)
	}

	extractor := design.NewExtractor(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()

		result, err := extractor.ExtractFromFile(cssFile)
		elapsed := time.Since(start)

		if err != nil {
			b.Fatalf("Extraction failed: %v", err)
		}

		if i == 0 {
			b.ReportMetric(float64(elapsed.Nanoseconds())/1_000_000, "ms/extract")
			b.Logf("Extracted %d tokens", len(result.Tokens))
		}
	}
}

// BenchmarkTokenExtractionLargeFile measures extraction from large CSS file
func BenchmarkTokenExtractionLargeFile(b *testing.B) {
	helper := NewTestHelper(b)
	defer helper.Cleanup()

	// Generate large CSS file with many tokens
	var cssBuilder strings.Builder
	cssBuilder.WriteString(":root {\n")

	// Generate 500 tokens
	for i := 0; i < 500; i++ {
		cssBuilder.WriteString(fmt.Sprintf("  --color-%d: #%06x;\n", i, i*1000))
		cssBuilder.WriteString(fmt.Sprintf("  --spacing-%d: %drem;\n", i, i))
		cssBuilder.WriteString(fmt.Sprintf("  --font-size-%d: %drem;\n", i, i))
	}
	cssBuilder.WriteString("}\n")

	cssFile := filepath.Join(helper.TempDir, "large.css")
	if err := os.WriteFile(cssFile, []byte(cssBuilder.String()), 0644); err != nil {
		b.Fatalf("Failed to write CSS file: %v", err)
	}

	extractor := design.NewExtractor(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()

		result, err := extractor.ExtractFromFile(cssFile)
		elapsed := time.Since(start)

		if err != nil {
			b.Fatalf("Extraction failed: %v", err)
		}

		if i == 0 {
			b.ReportMetric(float64(elapsed.Nanoseconds())/1_000_000, "ms/large-extract")
			b.Logf("Extracted %d tokens from large file", len(result.Tokens))
		}
	}
}

// BenchmarkTokenValidation measures token validation performance
func BenchmarkTokenValidation(b *testing.B) {
	validator := design.NewValidator()

	testToken := &design.Token{
		Name:     "color-primary",
		Value:    "#007bff",
		Type:     design.TokenTypeColor,
		Category: "color",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()

		result := validator.Validate(testToken)
		elapsed := time.Since(start)

		if i == 0 {
			b.ReportMetric(float64(elapsed.Nanoseconds())/1_000, "μs/validate")
			b.Logf("Validation result: valid=%v", result.Valid)
		}
	}
}

// BenchmarkTokenCategorization measures token categorization performance
func BenchmarkTokenCategorization(b *testing.B) {
	categorizer := design.NewTokenCategorizer()

	tokens := []design.Token{
		{Name: "color-primary", Value: "#007bff"},
		{Name: "spacing-md", Value: "1rem"},
		{Name: "font-size-base", Value: "16px"},
		{Name: "shadow-lg", Value: "0 10px 15px rgba(0,0,0,0.1)"},
		{Name: "border-radius", Value: "4px"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := range tokens {
			start := time.Now()

			categorizer.Categorize(&tokens[j])
			elapsed := time.Since(start)

			if i == 0 && j == 0 {
				b.ReportMetric(float64(elapsed.Nanoseconds())/1_000, "μs/categorize")
			}
		}
	}
}

// BenchmarkTokenParsing measures different parser performance
func BenchmarkTokenParsing(b *testing.B) {
	b.Run("CSS", func(b *testing.B) {
		parser := design.NewCSSParser()
		cssContent := `:root {
  --color-primary: #007bff;
  --spacing-md: 1rem;
  --font-size-base: 16px;
}`

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			start := time.Now()

			result, err := parser.Parse(strings.NewReader(cssContent))
			elapsed := time.Since(start)

			if err != nil {
				b.Fatalf("Parse failed: %v", err)
			}

			if i == 0 {
				b.ReportMetric(float64(elapsed.Nanoseconds())/1_000, "μs/css-parse")
				b.Logf("Parsed %d CSS tokens", len(result.Tokens))
			}
		}
	})

	b.Run("SCSS", func(b *testing.B) {
		parser := design.NewSCSSParser()
		scssContent := `$color-primary: #007bff;
$spacing-md: 1rem;
$font-size-base: 16px;`

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			start := time.Now()

			result, err := parser.Parse(strings.NewReader(scssContent))
			elapsed := time.Since(start)

			if err != nil {
				b.Fatalf("Parse failed: %v", err)
			}

			if i == 0 {
				b.ReportMetric(float64(elapsed.Nanoseconds())/1_000, "μs/scss-parse")
				b.Logf("Parsed %d SCSS tokens", len(result.Tokens))
			}
		}
	})

	b.Run("LESS", func(b *testing.B) {
		parser := design.NewLESSParser()
		lessContent := `@color-primary: #007bff;
@spacing-md: 1rem;
@font-size-base: 16px;`

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			start := time.Now()

			result, err := parser.Parse(strings.NewReader(lessContent))
			elapsed := time.Since(start)

			if err != nil {
				b.Fatalf("Parse failed: %v", err)
			}

			if i == 0 {
				b.ReportMetric(float64(elapsed.Nanoseconds())/1_000, "μs/less-parse")
				b.Logf("Parsed %d LESS tokens", len(result.Tokens))
			}
		}
	})
}

// BenchmarkTokenFormatting measures token formatting performance
func BenchmarkTokenFormatting(b *testing.B) {
	tokens := []design.Token{
		{Name: "color-primary", Value: "#007bff", Type: design.TokenTypeColor},
		{Name: "spacing-md", Value: "1rem", Type: design.TokenTypeSpacing},
		{Name: "font-size-base", Value: "16px", Type: design.TokenTypeTypography},
	}

	b.Run("JSON", func(b *testing.B) {
		formatter := design.NewJSONFormatter(true)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			start := time.Now()

			output, err := formatter.Format(tokens)
			elapsed := time.Since(start)

			if err != nil {
				b.Fatalf("Format failed: %v", err)
			}

			if i == 0 {
				b.ReportMetric(float64(elapsed.Nanoseconds())/1_000, "μs/json-format")
				b.Logf("Formatted %d bytes", len(output))
			}
		}
	})

	b.Run("CSS", func(b *testing.B) {
		formatter := design.NewCSSVariablesFormatter("")

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			start := time.Now()

			output, err := formatter.Format(tokens)
			elapsed := time.Since(start)

			if err != nil {
				b.Fatalf("Format failed: %v", err)
			}

			if i == 0 {
				b.ReportMetric(float64(elapsed.Nanoseconds())/1_000, "μs/css-format")
				b.Logf("Formatted %d bytes", len(output))
			}
		}
	})

	b.Run("SCSS", func(b *testing.B) {
		formatter := design.NewSCSSVariablesFormatter()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			start := time.Now()

			output, err := formatter.Format(tokens)
			elapsed := time.Since(start)

			if err != nil {
				b.Fatalf("Format failed: %v", err)
			}

			if i == 0 {
				b.ReportMetric(float64(elapsed.Nanoseconds())/1_000, "μs/scss-format")
				b.Logf("Formatted %d bytes", len(output))
			}
		}
	})
}

// BenchmarkCodeGeneration measures code generation performance
func BenchmarkCodeGeneration(b *testing.B) {
	tokens := []design.Token{
		{Name: "color-primary", Value: "#007bff", Type: design.TokenTypeColor},
		{Name: "color-secondary", Value: "#6c757d", Type: design.TokenTypeColor},
		{Name: "spacing-sm", Value: "0.5rem", Type: design.TokenTypeSpacing},
		{Name: "spacing-md", Value: "1rem", Type: design.TokenTypeSpacing},
		{Name: "font-size-base", Value: "16px", Type: design.TokenTypeTypography},
	}

	b.Run("TypeScript", func(b *testing.B) {
		generator := generators.NewTypeScriptGenerator()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			start := time.Now()

			code, err := generator.Generate(tokens, "ts")
			elapsed := time.Since(start)

			if err != nil {
				b.Fatalf("Generate failed: %v", err)
			}

			if i == 0 {
				b.ReportMetric(float64(elapsed.Nanoseconds())/1_000, "μs/ts-gen")
				b.Logf("Generated %d bytes of TypeScript", len(code))
			}
		}
	})

	b.Run("JavaScript", func(b *testing.B) {
		generator := generators.NewTypeScriptGenerator()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			start := time.Now()

			code, err := generator.Generate(tokens, "js")
			elapsed := time.Since(start)

			if err != nil {
				b.Fatalf("Generate failed: %v", err)
			}

			if i == 0 {
				b.ReportMetric(float64(elapsed.Nanoseconds())/1_000, "μs/js-gen")
				b.Logf("Generated %d bytes of JavaScript", len(code))
			}
		}
	})
}

// BenchmarkTokenResolutionEndToEnd measures complete token resolution pipeline
func BenchmarkTokenResolutionEndToEnd(b *testing.B) {
	helper := NewTestHelper(b)
	defer helper.Cleanup()

	// Create test CSS file
	cssContent := `
:root {
  --color-primary: #007bff;
  --color-secondary: #6c757d;
  --spacing-md: 1rem;
  --font-size-base: 16px;
}
`
	cssFile := filepath.Join(helper.TempDir, "tokens.css")
	if err := os.WriteFile(cssFile, []byte(cssContent), 0644); err != nil {
		b.Fatalf("Failed to write CSS file: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()

		// Extract tokens
		extractor := design.NewExtractor(nil)
		result, err := extractor.ExtractFromFile(cssFile)
		if err != nil {
			b.Fatalf("Extraction failed: %v", err)
		}

		// Generate code
		generator := generators.NewTypeScriptGenerator()
		_, err = generator.Generate(result.Tokens, "ts")
		if err != nil {
			b.Fatalf("Generation failed: %v", err)
		}

		elapsed := time.Since(start)

		if i == 0 {
			b.ReportMetric(float64(elapsed.Nanoseconds())/1_000_000, "ms/end-to-end")
		}
	}
}
