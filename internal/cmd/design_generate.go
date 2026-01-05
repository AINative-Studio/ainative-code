package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/AINative-studio/ainative-code/internal/design"
	"github.com/AINative-studio/ainative-code/internal/design/generators"
	"github.com/AINative-studio/ainative-code/internal/logger"
)

var (
	generateTokensFile string
	generateFormat     string
	generateOutput     string
	generatePretty     bool
	generateTemplate   string
)

// designGenerateCmd represents the design generate command
var designGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate code from design tokens",
	Long: `Generate code from design tokens in various formats.

This command reads design tokens from a JSON file and generates code in the
specified format. Supported formats include:

  - tailwind: Tailwind CSS configuration file
  - css: CSS custom properties (variables)
  - scss: SCSS variables
  - typescript/ts: TypeScript constants
  - javascript/js: JavaScript constants
  - json: JSON format

Examples:
  # Generate Tailwind config from tokens
  ainative-code design generate --tokens tokens.json --format tailwind --output tailwind.config.js

  # Generate CSS variables
  ainative-code design generate --tokens tokens.json --format css --output design-tokens.css

  # Generate TypeScript constants
  ainative-code design generate --tokens tokens.json --format typescript --output tokens.ts

  # Generate SCSS variables
  ainative-code design generate --tokens tokens.json --format scss --output _tokens.scss

  # Use custom template
  ainative-code design generate --tokens tokens.json --format css --output custom.css --template my-template.tmpl`,
	Aliases: []string{"gen", "g"},
	RunE:    runDesignGenerate,
}

func init() {
	designCmd.AddCommand(designGenerateCmd)

	// Flags
	designGenerateCmd.Flags().StringVarP(&generateTokensFile, "tokens", "t", "", "input tokens file (JSON) (required)")
	designGenerateCmd.Flags().StringVarP(&generateFormat, "format", "f", "json", "output format (tailwind, css, scss, typescript, javascript, json)")
	designGenerateCmd.Flags().StringVarP(&generateOutput, "output", "o", "", "output file path (prints to stdout if not specified)")
	designGenerateCmd.Flags().BoolVarP(&generatePretty, "pretty", "p", true, "pretty-print output (for JSON)")
	designGenerateCmd.Flags().StringVar(&generateTemplate, "template", "", "path to custom template file")

	designGenerateCmd.MarkFlagRequired("tokens")
}

func runDesignGenerate(cmd *cobra.Command, args []string) error {
	logger.InfoEvent().
		Str("tokens_file", generateTokensFile).
		Str("format", generateFormat).
		Str("output", generateOutput).
		Msg("Generating code from design tokens")

	// Validate format
	validFormats := []string{"tailwind", "tw", "css", "scss", "sass", "typescript", "ts", "javascript", "js", "json"}
	if !contains(validFormats, generateFormat) {
		return fmt.Errorf("invalid format '%s'. Supported formats: %s", generateFormat, strings.Join(validFormats, ", "))
	}

	// Read tokens from file
	tokens, err := readTokensFromFile(generateTokensFile)
	if err != nil {
		return fmt.Errorf("failed to read tokens: %w", err)
	}

	logger.InfoEvent().Int("token_count", len(tokens)).Msg("Loaded design tokens")

	// Generate code based on format
	var output string
	var fileExt string

	if generateTemplate != "" {
		// Use custom template
		output, err = generateWithCustomTemplate(tokens, generateTemplate)
		fileExt = ".txt"
	} else {
		// Use built-in generators
		output, fileExt, err = generateWithBuiltInGenerator(tokens, generateFormat)
	}

	if err != nil {
		return fmt.Errorf("failed to generate code: %w", err)
	}

	// Write output
	if generateOutput == "" {
		// Print to stdout
		fmt.Println(output)
	} else {
		// Ensure output has correct extension if not specified
		if filepath.Ext(generateOutput) == "" {
			generateOutput += fileExt
		}

		// Create directory if it doesn't exist
		outputDir := filepath.Dir(generateOutput)
		if outputDir != "." && outputDir != "" {
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				return fmt.Errorf("failed to create output directory: %w", err)
			}
		}

		// Write file
		if err := os.WriteFile(generateOutput, []byte(output), 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}

		logger.InfoEvent().Str("file", generateOutput).Msg("Code generated successfully")
		fmt.Printf("✓ Generated %s code: %s\n", generateFormat, generateOutput)
	}

	return nil
}

// readTokensFromFile reads design tokens from a JSON file
func readTokensFromFile(filename string) ([]design.Token, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file '%s': %w", filename, err)
	}

	// Try to parse as token array directly
	var tokens []design.Token
	if err := json.Unmarshal(data, &tokens); err == nil {
		return tokens, nil
	}

	// Try to parse as object with tokens field
	var obj struct {
		Tokens []design.Token `json:"tokens"`
	}
	if err := json.Unmarshal(data, &obj); err != nil {
		return nil, fmt.Errorf("failed to parse tokens JSON: %w", err)
	}

	if len(obj.Tokens) == 0 {
		return nil, fmt.Errorf("no tokens found in file '%s'", filename)
	}

	return obj.Tokens, nil
}

// generateWithBuiltInGenerator generates code using built-in generators
func generateWithBuiltInGenerator(tokens []design.Token, format string) (string, string, error) {
	switch format {
	case "tailwind", "tw":
		gen := generators.NewTailwindGenerator()
		output, err := gen.Generate(tokens)
		return output, ".js", err

	case "css":
		gen := generators.NewCSSGenerator()
		output, err := gen.Generate(tokens)
		return output, ".css", err

	case "scss", "sass":
		gen := generators.NewSCSSGenerator()
		output, err := gen.Generate(tokens)
		return output, ".scss", err

	case "typescript", "ts":
		gen := generators.NewTypeScriptGenerator()
		output, err := gen.Generate(tokens, "typescript")
		return output, ".ts", err

	case "javascript", "js":
		gen := generators.NewTypeScriptGenerator()
		output, err := gen.Generate(tokens, "javascript")
		return output, ".js", err

	case "json":
		gen := generators.NewJSONGenerator(generatePretty)
		output, err := gen.Generate(tokens)
		return output, ".json", err

	default:
		return "", "", fmt.Errorf("unsupported format: %s", format)
	}
}

// generateWithCustomTemplate generates code using a custom template
func generateWithCustomTemplate(tokens []design.Token, templatePath string) (string, error) {
	// Read template file
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to read template file '%s': %w", templatePath, err)
	}

	// Execute template
	engine := design.NewTemplateEngine()
	data := map[string]interface{}{
		"Tokens": tokens,
	}

	return engine.ExecuteString(string(templateContent), data)
}

// contains checks if a string slice contains a value
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
