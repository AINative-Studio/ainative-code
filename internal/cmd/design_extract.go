package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/AINative-studio/ainative-code/internal/design"
	"github.com/AINative-studio/ainative-code/internal/logger"
)

var (
	extractSourceFile   string
	extractOutputFile   string
	extractFormat       string
	extractPretty       bool
	extractValidate     bool
	extractIncludeComments bool
)

// designExtractCmd represents the design extract command
var designExtractCmd = &cobra.Command{
	Use:   "extract",
	Short: "Extract design tokens from CSS/SCSS/LESS files",
	Long: `Extract design tokens from CSS, SCSS, or LESS files and output them in various formats.

The extract command parses CSS-like files and identifies design tokens including:
- Colors (hex, rgb, rgba, hsl, hsla)
- Typography (font-family, font-size, line-height, font-weight, letter-spacing)
- Spacing (margin, padding, gap)
- Shadows (box-shadow, text-shadow)
- Border radius

Supports multiple output formats:
- JSON: Standard JSON format for easy integration
- YAML: Human-readable YAML format
- Tailwind: Tailwind CSS configuration format

Examples:
  # Extract from CSS file to JSON
  ainative-code design extract --source styles.css --output tokens.json --format json

  # Extract from SCSS file to Tailwind config
  ainative-code design extract --source variables.scss --output tailwind.config.js --format tailwind

  # Extract from LESS file to YAML
  ainative-code design extract --source theme.less --output tokens.yaml --format yaml

  # Extract with validation enabled
  ainative-code design extract --source styles.css --output tokens.json --validate

  # Extract with pretty formatting
  ainative-code design extract --source styles.css --output tokens.json --pretty`,
	Aliases: []string{"ext", "parse"},
	RunE:    runDesignExtract,
}

func init() {
	designCmd.AddCommand(designExtractCmd)

	// Flags
	designExtractCmd.Flags().StringVarP(&extractSourceFile, "source", "s", "", "source CSS/SCSS/LESS file (required)")
	designExtractCmd.Flags().StringVarP(&extractOutputFile, "output", "o", "", "output file path (required)")
	designExtractCmd.Flags().StringVarP(&extractFormat, "format", "f", "json", "output format (json, yaml, tailwind)")
	designExtractCmd.Flags().BoolVar(&extractPretty, "pretty", true, "pretty print output (for json)")
	designExtractCmd.Flags().BoolVar(&extractValidate, "validate", true, "validate extracted tokens")
	designExtractCmd.Flags().BoolVar(&extractIncludeComments, "include-comments", true, "include comments in output (for tailwind)")

	// Mark required flags
	designExtractCmd.MarkFlagRequired("source")
	designExtractCmd.MarkFlagRequired("output")
}

func runDesignExtract(cmd *cobra.Command, args []string) error {
	logger.InfoEvent().
		Str("source", extractSourceFile).
		Str("output", extractOutputFile).
		Str("format", extractFormat).
		Msg("Extracting design tokens")

	// Validate source file exists
	if _, err := os.Stat(extractSourceFile); os.IsNotExist(err) {
		return fmt.Errorf("source file does not exist: %s", extractSourceFile)
	}

	// Validate output format
	var outputFormat design.OutputFormat
	switch strings.ToLower(extractFormat) {
	case "json":
		outputFormat = design.OutputFormatJSON
	case "yaml", "yml":
		outputFormat = design.OutputFormatYAML
	case "tailwind", "tw":
		outputFormat = design.OutputFormatTailwind
	default:
		return fmt.Errorf("unsupported output format: %s (supported: json, yaml, tailwind)", extractFormat)
	}

	// Create extractor
	extractor := design.NewExtractor(&design.ExtractorOptions{
		EnableValidation: extractValidate,
		IncludeComments:  extractIncludeComments,
		InferCategories:  true,
	})

	// Extract tokens from file
	logger.Debug("Parsing source file...")
	result, err := extractor.ExtractFromFile(extractSourceFile)
	if err != nil {
		return fmt.Errorf("failed to extract tokens: %w", err)
	}

	// Display extraction summary
	fmt.Printf("Extracted %d tokens from %s\n", len(result.Tokens), extractSourceFile)

	if len(result.Warnings) > 0 {
		fmt.Printf("\nWarnings (%d):\n", len(result.Warnings))
		for _, warning := range result.Warnings {
			fmt.Printf("  - %s\n", warning)
		}
	}

	if len(result.Errors) > 0 {
		fmt.Printf("\nErrors (%d):\n", len(result.Errors))
		for _, err := range result.Errors {
			fmt.Printf("  - %s\n", err.Error())
		}
		return fmt.Errorf("extraction completed with errors")
	}

	// Display token summary by type
	tokensByType := groupTokensByType(result.Tokens)
	fmt.Println("\nToken Summary:")
	for tokenType, tokens := range tokensByType {
		fmt.Printf("  - %s: %d\n", tokenType, len(tokens))
	}

	// Get formatter
	formatterOptions := map[string]interface{}{
		"pretty":          extractPretty,
		"includeComments": extractIncludeComments,
	}
	formatter, err := design.GetFormatter(outputFormat, formatterOptions)
	if err != nil {
		return fmt.Errorf("failed to create formatter: %w", err)
	}

	// Format tokens
	logger.Debug("Formatting tokens...")
	output, err := formatter.Format(result.Tokens)
	if err != nil {
		return fmt.Errorf("failed to format tokens: %w", err)
	}

	// Ensure output directory exists
	outputDir := filepath.Dir(extractOutputFile)
	if outputDir != "" && outputDir != "." {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	// Write output file
	logger.Debug("Writing output file...")
	if err := os.WriteFile(extractOutputFile, []byte(output), 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	fmt.Printf("\nTokens successfully extracted to: %s\n", extractOutputFile)
	fmt.Printf("Format: %s\n", extractFormat)

	// Display sample tokens if verbose
	if len(result.Tokens) > 0 && len(result.Tokens) <= 5 {
		fmt.Println("\nSample Tokens:")
		for i, token := range result.Tokens {
			if i >= 5 {
				break
			}
			fmt.Printf("  %d. %s (%s): %s\n", i+1, token.Name, token.Type, token.Value)
		}
	} else if len(result.Tokens) > 5 {
		fmt.Println("\nFirst 5 Tokens:")
		for i := 0; i < 5; i++ {
			token := result.Tokens[i]
			fmt.Printf("  %d. %s (%s): %s\n", i+1, token.Name, token.Type, token.Value)
		}
		fmt.Printf("  ... and %d more\n", len(result.Tokens)-5)
	}

	return nil
}

// groupTokensByType groups tokens by their type.
func groupTokensByType(tokens []design.Token) map[string][]design.Token {
	grouped := make(map[string][]design.Token)
	for _, token := range tokens {
		tokenType := string(token.Type)
		grouped[tokenType] = append(grouped[tokenType], token)
	}
	return grouped
}
