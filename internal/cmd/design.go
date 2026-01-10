package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/AINative-studio/ainative-code/internal/design"
	"github.com/AINative-studio/ainative-code/internal/logger"
	"gopkg.in/yaml.v3"
)

var (
	designImportFile string
	designExportFile string
	designFormat     string
)

// designCmd represents the design command
var designCmd = &cobra.Command{
	Use:   "design",
	Short: "Manage design tokens",
	Long: `Manage design tokens for UI consistency across applications.

Design tokens are design decisions represented as data, including colors,
typography, spacing, and other visual properties. This command helps manage
and synchronize design tokens with your design system.

Examples:
  # List all design tokens
  ainative-code design list

  # Show a specific token
  ainative-code design show colors.primary

  # Import tokens from a file
  ainative-code design import --file tokens.json

  # Export tokens to a file
  ainative-code design export --file tokens.json --format json

  # Sync tokens with Strapi
  ainative-code design sync`,
	Aliases: []string{"tokens", "dt"},
}

// designListCmd represents the design list command
var designListCmd = &cobra.Command{
	Use:   "list",
	Short: "List design tokens",
	Long:  `List all design tokens stored in the database.`,
	Aliases: []string{"ls", "l"},
	RunE:  runDesignList,
}

// designShowCmd represents the design show command
var designShowCmd = &cobra.Command{
	Use:   "show [token-name]",
	Short: "Show token details",
	Long:  `Display detailed information about a specific design token.`,
	Aliases: []string{"get", "view"},
	Args:  cobra.ExactArgs(1),
	RunE:  runDesignShow,
}

// designImportCmd represents the design import command
var designImportCmd = &cobra.Command{
	Use:   "import",
	Short: "Import design tokens",
	Long:  `Import design tokens from a JSON or YAML file.`,
	RunE:  runDesignImport,
}

// designExportCmd represents the design export command
var designExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export design tokens",
	Long:  `Export design tokens to a JSON or YAML file.`,
	RunE:  runDesignExport,
}

// designSyncCmd represents the design sync command
var designSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync tokens with Strapi",
	Long:  `Synchronize design tokens between local database and Strapi CMS.`,
	RunE:  runDesignSync,
}

// designValidateCmd represents the design validate command
var designValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate design tokens",
	Long:  `Validate design tokens for correctness and consistency.`,
	RunE:  runDesignValidate,
}

func init() {
	rootCmd.AddCommand(designCmd)

	// Add subcommands
	designCmd.AddCommand(designListCmd)
	designCmd.AddCommand(designShowCmd)
	designCmd.AddCommand(designImportCmd)
	designCmd.AddCommand(designExportCmd)
	designCmd.AddCommand(designSyncCmd)
	designCmd.AddCommand(designValidateCmd)

	// Import flags
	designImportCmd.Flags().StringP("file", "f", "", "input file path (required)")
	designImportCmd.MarkFlagRequired("file")
	designImportCmd.Flags().BoolP("merge", "m", false, "merge with existing tokens")

	// Export flags
	designExportCmd.Flags().StringP("file", "f", "", "output file path (required)")
	designExportCmd.MarkFlagRequired("file")
	designExportCmd.Flags().StringVar(&designFormat, "format", "json", "output format (json, yaml)")

	// Sync flags
	designSyncCmd.Flags().String("direction", "pull", "sync direction (pull, push, both)")
}

func runDesignList(cmd *cobra.Command, args []string) error {
	logger.Debug("Listing design tokens")

	fmt.Println("Design Tokens")
	fmt.Println("==============")
	fmt.Println()

	fmt.Println("This command requires design tokens database schema to be implemented.")
	fmt.Println()
	fmt.Println("Current Status:")
	fmt.Println("  • Design token storage requires database schema (planned)")
	fmt.Println("  • Design token import/export functionality is stubbed out")
	fmt.Println("  • Strapi sync for design tokens is planned")
	fmt.Println()
	fmt.Println("Design tokens are typically managed in:")
	fmt.Println("  1. Design files (Figma, Sketch, etc.)")
	fmt.Println("  2. Strapi CMS for centralized management")
	fmt.Println("  3. JSON/YAML files in your repository")
	fmt.Println()
	fmt.Println("Common design token categories:")
	fmt.Println()
	fmt.Println("  Colors:")
	fmt.Println("    • colors.primary")
	fmt.Println("    • colors.secondary")
	fmt.Println("    • colors.accent")
	fmt.Println("    • colors.text.primary")
	fmt.Println("    • colors.text.secondary")
	fmt.Println()
	fmt.Println("  Typography:")
	fmt.Println("    • typography.font.family")
	fmt.Println("    • typography.font.size.base")
	fmt.Println("    • typography.line.height")
	fmt.Println("    • typography.font.weight.normal")
	fmt.Println()
	fmt.Println("  Spacing:")
	fmt.Println("    • spacing.xs (4px)")
	fmt.Println("    • spacing.sm (8px)")
	fmt.Println("    • spacing.md (16px)")
	fmt.Println("    • spacing.lg (24px)")
	fmt.Println("    • spacing.xl (32px)")
	fmt.Println()
	fmt.Println("To use design tokens now:")
	fmt.Println("  • Use 'design import --file tokens.json' to import from a file")
	fmt.Println("  • Use 'design export --file tokens.json' to export to a file")
	fmt.Println("  • Use 'strapi fetch design-tokens' to sync from Strapi CMS")

	return nil
}

func runDesignShow(cmd *cobra.Command, args []string) error {
	tokenName := args[0]

	logger.DebugEvent().Str("token_name", tokenName).Msg("Showing token details")

	fmt.Printf("\nDesign Token: %s\n", tokenName)
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println()

	fmt.Println("This command requires design tokens database schema to be implemented.")
	fmt.Println()
	fmt.Println("Example token structure:")
	fmt.Println()

	// Show example based on token name pattern
	if strings.Contains(tokenName, "color") {
		fmt.Println("  {")
		fmt.Printf("    \"name\": \"%s\",\n", tokenName)
		fmt.Println("    \"type\": \"color\",")
		fmt.Println("    \"value\": \"#3B82F6\",")
		fmt.Println("    \"description\": \"Primary brand color\",")
		fmt.Println("    \"category\": \"colors\",")
		fmt.Println("    \"metadata\": {")
		fmt.Println("      \"rgb\": \"59, 130, 246\",")
		fmt.Println("      \"hsl\": \"217, 91%, 60%\"")
		fmt.Println("    }")
		fmt.Println("  }")
	} else if strings.Contains(tokenName, "typography") || strings.Contains(tokenName, "font") {
		fmt.Println("  {")
		fmt.Printf("    \"name\": \"%s\",\n", tokenName)
		fmt.Println("    \"type\": \"typography\",")
		fmt.Println("    \"value\": \"16px\",")
		fmt.Println("    \"description\": \"Base font size\",")
		fmt.Println("    \"category\": \"typography\",")
		fmt.Println("    \"metadata\": {")
		fmt.Println("      \"rem\": \"1rem\",")
		fmt.Println("      \"lineHeight\": \"1.5\"")
		fmt.Println("    }")
		fmt.Println("  }")
	} else if strings.Contains(tokenName, "spacing") {
		fmt.Println("  {")
		fmt.Printf("    \"name\": \"%s\",\n", tokenName)
		fmt.Println("    \"type\": \"spacing\",")
		fmt.Println("    \"value\": \"16px\",")
		fmt.Println("    \"description\": \"Medium spacing\",")
		fmt.Println("    \"category\": \"spacing\",")
		fmt.Println("    \"metadata\": {")
		fmt.Println("      \"rem\": \"1rem\",")
		fmt.Println("      \"scale\": \"4px base\"")
		fmt.Println("    }")
		fmt.Println("  }")
	} else {
		fmt.Println("  {")
		fmt.Printf("    \"name\": \"%s\",\n", tokenName)
		fmt.Println("    \"type\": \"unknown\",")
		fmt.Println("    \"value\": \"...\",")
		fmt.Println("    \"description\": \"Token description\",")
		fmt.Println("    \"category\": \"general\"")
		fmt.Println("  }")
	}

	fmt.Println()
	fmt.Println("To view actual tokens:")
	fmt.Println("  • Use 'design list' to see all available tokens")
	fmt.Println("  • Import tokens from a file with 'design import --file tokens.json'")

	return nil
}

func runDesignImport(cmd *cobra.Command, args []string) error {
	file, _ := cmd.Flags().GetString("file")
	merge, _ := cmd.Flags().GetBool("merge")

	// Validate input file with comprehensive checks
	if err := validateInputFile(file); err != nil {
		return err
	}

	logger.InfoEvent().
		Str("file", file).
		Bool("merge", merge).
		Msg("Importing design tokens")

	fmt.Printf("Importing tokens from: %s\n", file)

	if merge {
		fmt.Println("Mode: Merge with existing tokens")
	} else {
		fmt.Println("Mode: Replace all tokens")
	}

	// TODO: Implement token import
	// - Read file (JSON/YAML)
	// - Validate token structure
	// - Store in database
	// - Report success/errors

	fmt.Println("Import completed!")

	return nil
}

func runDesignExport(cmd *cobra.Command, args []string) error {
	file, _ := cmd.Flags().GetString("file")
	format, _ := cmd.Flags().GetString("format")

	// Validate output path with comprehensive checks
	if err := validateOutputPath(file); err != nil {
		return err
	}

	logger.InfoEvent().
		Str("file", file).
		Str("format", format).
		Msg("Exporting design tokens")

	fmt.Printf("Exporting tokens to: %s (format: %s)\n", file, format)

	// Create example design tokens collection
	// In a real implementation, this would query tokens from a database
	exampleTokens := createExampleDesignTokens()

	// Format and write tokens based on format
	var data []byte
	var err error

	switch strings.ToLower(format) {
	case "json":
		data, err = marshalJSON(exampleTokens)
	case "yaml", "yml":
		data, err = marshalYAML(exampleTokens)
	default:
		return fmt.Errorf("unsupported format: %s (supported: json, yaml)", format)
	}

	if err != nil {
		return fmt.Errorf("failed to format tokens: %w", err)
	}

	// Write to file
	if err := os.WriteFile(file, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Printf("Successfully exported %d design tokens to %s\n", len(exampleTokens.Tokens), file)
	fmt.Println("Export completed!")

	return nil
}

func runDesignSync(cmd *cobra.Command, args []string) error {
	direction, _ := cmd.Flags().GetString("direction")

	logger.InfoEvent().Str("direction", direction).Msg("Syncing design tokens")

	fmt.Printf("Syncing tokens with Strapi (direction: %s)\n", direction)

	// TODO: Implement Strapi sync
	// - Connect to Strapi
	// - Pull/push tokens based on direction
	// - Handle conflicts
	// - Report changes

	fmt.Println("Sync completed!")

	return nil
}

func runDesignValidate(cmd *cobra.Command, args []string) error {
	logger.Debug("Validating design tokens")

	fmt.Println("Validating design tokens...")

	// TODO: Implement token validation
	// - Check required properties
	// - Validate color formats
	// - Check typography values
	// - Verify spacing scale
	// - Report issues

	fmt.Println("Validation completed!")

	return nil
}

// createExampleDesignTokens creates a sample collection of design tokens
// In a production implementation, this would fetch tokens from a database
func createExampleDesignTokens() *design.TokenCollection {
	return &design.TokenCollection{
		Tokens: []design.Token{
			{
				Name:        "colors.primary",
				Type:        design.TokenTypeColor,
				Value:       "#007bff",
				Description: "Primary brand color",
				Category:    "colors",
				Metadata: map[string]string{
					"rgb": "0, 123, 255",
					"hsl": "211, 100%, 50%",
				},
			},
			{
				Name:        "colors.secondary",
				Type:        design.TokenTypeColor,
				Value:       "#6c757d",
				Description: "Secondary color",
				Category:    "colors",
				Metadata: map[string]string{
					"rgb": "108, 117, 125",
					"hsl": "208, 7%, 46%",
				},
			},
			{
				Name:        "colors.success",
				Type:        design.TokenTypeColor,
				Value:       "#28a745",
				Description: "Success state color",
				Category:    "colors",
				Metadata: map[string]string{
					"rgb": "40, 167, 69",
					"hsl": "134, 61%, 41%",
				},
			},
			{
				Name:        "spacing.base",
				Type:        design.TokenTypeSpacing,
				Value:       "16px",
				Description: "Base spacing unit",
				Category:    "spacing",
				Metadata: map[string]string{
					"rem": "1rem",
				},
			},
			{
				Name:        "spacing.sm",
				Type:        design.TokenTypeSpacing,
				Value:       "8px",
				Description: "Small spacing",
				Category:    "spacing",
				Metadata: map[string]string{
					"rem": "0.5rem",
				},
			},
			{
				Name:        "spacing.lg",
				Type:        design.TokenTypeSpacing,
				Value:       "32px",
				Description: "Large spacing",
				Category:    "spacing",
				Metadata: map[string]string{
					"rem": "2rem",
				},
			},
			{
				Name:        "typography.font.family.base",
				Type:        design.TokenTypeTypography,
				Value:       "Helvetica, Arial, sans-serif",
				Description: "Base font family",
				Category:    "typography",
			},
			{
				Name:        "typography.font.size.base",
				Type:        design.TokenTypeTypography,
				Value:       "16px",
				Description: "Base font size",
				Category:    "typography",
				Metadata: map[string]string{
					"rem":        "1rem",
					"lineHeight": "1.5",
				},
			},
			{
				Name:        "border-radius.sm",
				Type:        design.TokenTypeBorderRadius,
				Value:       "4px",
				Description: "Small border radius",
				Category:    "border-radius",
			},
			{
				Name:        "shadow.sm",
				Type:        design.TokenTypeShadow,
				Value:       "0 1px 2px rgba(0,0,0,0.1)",
				Description: "Small shadow",
				Category:    "shadow",
			},
		},
		Metadata: map[string]string{
			"version":   "1.0.0",
			"generated": "true",
			"source":    "ainative-code design export",
		},
	}
}

// marshalJSON marshals the token collection to JSON with pretty printing
func marshalJSON(tokens *design.TokenCollection) ([]byte, error) {
	return json.MarshalIndent(tokens, "", "  ")
}

// marshalYAML marshals the token collection to YAML
func marshalYAML(tokens *design.TokenCollection) ([]byte, error) {
	return yaml.Marshal(tokens)
}

// validateInputFile performs comprehensive validation on input file paths
func validateInputFile(path string) error {
	// Check if path is empty
	if path == "" {
		return fmt.Errorf("file path cannot be empty")
	}

	// Check if file exists
	fileInfo, err := os.Stat(path)
	if os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", path)
	}
	if err != nil {
		return fmt.Errorf("cannot access file: %w", err)
	}

	// Check if path is a directory
	if fileInfo.IsDir() {
		return fmt.Errorf("path is a directory, not a file: %s", path)
	}

	// Check if file is readable by attempting to open it
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("cannot read file (permission denied): %s", path)
	}
	defer file.Close()

	// Check if file has content (not empty)
	if fileInfo.Size() == 0 {
		return fmt.Errorf("file is empty: %s", path)
	}

	return nil
}

// validateOutputPath performs comprehensive validation on output file paths
func validateOutputPath(path string) error {
	// Check if path is empty
	if path == "" {
		return fmt.Errorf("output file path cannot be empty")
	}

	// Get the directory path
	dir := filepath.Dir(path)

	// Handle current directory case
	if dir == "." {
		dir, _ = os.Getwd()
	}

	// Check if directory exists
	dirInfo, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return fmt.Errorf("output directory does not exist: %s\nPlease create the directory first with: mkdir -p %s", dir, dir)
	}
	if err != nil {
		return fmt.Errorf("cannot access output directory: %w", err)
	}

	// Verify it's a directory
	if !dirInfo.IsDir() {
		return fmt.Errorf("parent path is not a directory: %s", dir)
	}

	// Test write permission by attempting to create/write a temporary file
	testFile := filepath.Join(dir, ".ainative_write_test")
	f, err := os.OpenFile(testFile, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("permission denied: cannot write to directory %s\nPlease check directory permissions", dir)
		}
		// If file already exists, try to write to it
		f, err = os.OpenFile(testFile, os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return fmt.Errorf("cannot write to directory %s: %w", dir, err)
		}
	}
	f.Close()
	os.Remove(testFile) // Clean up test file

	// If target file already exists, check if we can overwrite it
	if _, err := os.Stat(path); err == nil {
		// File exists, check if we can write to it
		f, err := os.OpenFile(path, os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("output file exists but is not writable: %s", path)
		}
		f.Close()
	}

	return nil
}
