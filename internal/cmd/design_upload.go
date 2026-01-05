package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/AINative-studio/ainative-code/internal/client"
	"github.com/AINative-studio/ainative-code/internal/client/design"
	designpkg "github.com/AINative-studio/ainative-code/internal/design"
	"github.com/AINative-studio/ainative-code/internal/logger"
)

var (
	uploadTokensFile       string
	uploadProject          string
	uploadConflictMode     string
	uploadValidateOnly     bool
	uploadShowProgress     bool
)

// designUploadCmd represents the design upload command
var designUploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload design tokens to AINative Design system",
	Long: `Upload design tokens to the AINative Design system.

This command uploads design tokens from a JSON or YAML file to your AINative
Design project. The tokens will be validated before upload to ensure they
conform to the design token specification.

Conflict Resolution Modes:
  - overwrite: Replace existing tokens with new values (default)
  - merge: Merge new tokens with existing, preferring new values
  - skip: Skip conflicting tokens and keep existing values

Examples:
  # Upload tokens with overwrite mode
  ainative-code design upload --tokens tokens.json --project my-project

  # Upload with merge conflict resolution
  ainative-code design upload --tokens tokens.yaml --project my-project --conflict merge

  # Validate tokens without uploading
  ainative-code design upload --tokens tokens.json --validate-only

  # Upload with progress indicator
  ainative-code design upload --tokens tokens.json --project my-project --progress`,
	Aliases: []string{"push"},
	RunE:    runDesignUpload,
}

func init() {
	designCmd.AddCommand(designUploadCmd)

	designUploadCmd.Flags().StringVarP(&uploadTokensFile, "tokens", "t", "", "path to design tokens file (JSON or YAML) (required)")
	designUploadCmd.MarkFlagRequired("tokens")

	designUploadCmd.Flags().StringVarP(&uploadProject, "project", "p", "", "project ID for design tokens (required unless set in config)")

	designUploadCmd.Flags().StringVar(&uploadConflictMode, "conflict", "overwrite", "conflict resolution mode (overwrite, merge, skip)")

	designUploadCmd.Flags().BoolVar(&uploadValidateOnly, "validate-only", false, "only validate tokens without uploading")

	designUploadCmd.Flags().BoolVar(&uploadShowProgress, "progress", false, "show progress indicator for large token sets")
}

func runDesignUpload(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	logger.InfoEvent().
		Str("tokens_file", uploadTokensFile).
		Str("project", uploadProject).
		Str("conflict_mode", uploadConflictMode).
		Bool("validate_only", uploadValidateOnly).
		Msg("Starting design token upload")

	// Read tokens from file
	tokens, err := readTokensFromFileForUpload(uploadTokensFile)
	if err != nil {
		return fmt.Errorf("failed to read tokens file: %w", err)
	}

	fmt.Printf("📦 Loaded %d tokens from %s\n", len(tokens), uploadTokensFile)

	// Validate tokens
	validator := designpkg.NewValidator()
	validationResult := validator.ValidateBatch(tokens)

	if !validationResult.Valid {
		fmt.Println("\n❌ Token validation failed:")
		for _, err := range validationResult.Errors {
			fmt.Printf("  - %s\n", err.Error())
		}
		return fmt.Errorf("token validation failed with %d errors", len(validationResult.Errors))
	}

	fmt.Println("✅ All tokens validated successfully")

	// If validate-only mode, stop here
	if uploadValidateOnly {
		fmt.Println("\n✨ Validation complete (upload skipped)")
		return nil
	}

	// Check project ID
	if uploadProject == "" {
		return fmt.Errorf("project ID is required (use --project flag or set in config)")
	}

	// Parse conflict resolution mode
	conflictResolution, err := parseConflictResolution(uploadConflictMode)
	if err != nil {
		return err
	}

	// Create API client
	// TODO: Get base URL from config
	apiClient := client.New(
		client.WithBaseURL("https://api.ainative.studio"),
		client.WithTimeout(30*time.Second),
	)

	// Create design client
	designClient := design.New(
		design.WithAPIClient(apiClient),
		design.WithProjectID(uploadProject),
	)

	// Upload tokens with optional progress callback
	var progressCallback design.ProgressCallback
	if uploadShowProgress {
		progressCallback = func(uploaded, total int) {
			percentage := float64(uploaded) / float64(total) * 100
			fmt.Printf("\r⬆️  Uploading: %d/%d tokens (%.1f%%)", uploaded, total, percentage)
		}
	}

	fmt.Printf("\n🚀 Uploading tokens to project '%s' (conflict mode: %s)...\n", uploadProject, uploadConflictMode)

	result, err := designClient.UploadTokens(ctx, tokens, conflictResolution, progressCallback)
	if err != nil {
		return fmt.Errorf("failed to upload tokens: %w", err)
	}

	if uploadShowProgress {
		fmt.Println() // New line after progress indicator
	}

	// Print upload summary
	fmt.Println("\n📊 Upload Summary:")
	fmt.Printf("  ✅ Uploaded: %d tokens\n", result.UploadedCount)
	if result.UpdatedCount > 0 {
		fmt.Printf("  🔄 Updated: %d tokens\n", result.UpdatedCount)
	}
	if result.SkippedCount > 0 {
		fmt.Printf("  ⏭️  Skipped: %d tokens\n", result.SkippedCount)
	}

	if result.Message != "" {
		fmt.Printf("\n%s\n", result.Message)
	}

	logger.InfoEvent().
		Int("uploaded", result.UploadedCount).
		Int("updated", result.UpdatedCount).
		Int("skipped", result.SkippedCount).
		Msg("Design token upload completed")

	return nil
}

// readTokensFromFileForUpload reads design tokens from a JSON or YAML file for upload.
func readTokensFromFileForUpload(filePath string) ([]*designpkg.Token, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Detect file format by extension
	var tokens []*designpkg.Token

	if isYAMLFile(filePath) {
		// Parse YAML
		var yamlData struct {
			Tokens []designpkg.Token `yaml:"tokens"`
		}
		if err := yaml.Unmarshal(data, &yamlData); err != nil {
			return nil, fmt.Errorf("failed to parse YAML: %w", err)
		}
		// Convert to pointers
		for i := range yamlData.Tokens {
			tokens = append(tokens, &yamlData.Tokens[i])
		}
	} else {
		// Parse JSON
		var jsonData struct {
			Tokens []designpkg.Token `json:"tokens"`
		}
		if err := json.Unmarshal(data, &jsonData); err != nil {
			return nil, fmt.Errorf("failed to parse JSON: %w", err)
		}
		// Convert to pointers
		for i := range jsonData.Tokens {
			tokens = append(tokens, &jsonData.Tokens[i])
		}
	}

	if len(tokens) == 0 {
		return nil, fmt.Errorf("no tokens found in file")
	}

	return tokens, nil
}

// isYAMLFile checks if the file has a YAML extension.
func isYAMLFile(filePath string) bool {
	return len(filePath) >= 5 && (filePath[len(filePath)-5:] == ".yaml" || filePath[len(filePath)-4:] == ".yml")
}

// parseConflictResolution parses the conflict resolution mode string.
func parseConflictResolution(mode string) (designpkg.ConflictResolutionStrategyUpload, error) {
	switch mode {
	case "overwrite":
		return designpkg.ConflictOverwrite, nil
	case "merge":
		return designpkg.ConflictMerge, nil
	case "skip":
		return designpkg.ConflictSkip, nil
	default:
		return "", fmt.Errorf("invalid conflict resolution mode '%s' (must be: overwrite, merge, or skip)", mode)
	}
}
