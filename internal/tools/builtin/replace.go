// Package builtin provides built-in tools for the tool execution framework.
package builtin

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/AINative-studio/ainative-code/internal/tools"
)

// SearchReplaceTool implements find and replace functionality with regex support.
type SearchReplaceTool struct {
	sandbox *Sandbox
}

// NewSearchReplaceTool creates a new SearchReplaceTool instance with the specified sandbox.
func NewSearchReplaceTool(sandbox *Sandbox) *SearchReplaceTool {
	return &SearchReplaceTool{
		sandbox: sandbox,
	}
}

// Name returns the unique name of the tool.
func (t *SearchReplaceTool) Name() string {
	return "search_replace"
}

// Description returns a human-readable description of what the tool does.
func (t *SearchReplaceTool) Description() string {
	return "Performs find and replace operations across files with regex support, backup options, and dry-run mode"
}

// Schema returns the JSON schema for the tool's input parameters.
func (t *SearchReplaceTool) Schema() tools.ToolSchema {
	maxPatternLength := 1024
	maxReplacementLength := 4096

	return tools.ToolSchema{
		Type: "object",
		Properties: map[string]tools.PropertyDef{
			"pattern": {
				Type:        "string",
				Description: "Regular expression pattern to search for",
				MaxLength:   &maxPatternLength,
			},
			"replacement": {
				Type:        "string",
				Description: "Replacement string (supports regex capture groups like $1, $2)",
				MaxLength:   &maxReplacementLength,
			},
			"path": {
				Type:        "string",
				Description: "File or directory path to perform replacements (default: current working directory)",
			},
			"file_pattern": {
				Type:        "string",
				Description: "Glob pattern to filter files (e.g., '*.go', '*.md')",
			},
			"recursive": {
				Type:        "boolean",
				Description: "Search recursively through subdirectories (default: false)",
				Default:     false,
			},
			"case_sensitive": {
				Type:        "boolean",
				Description: "Perform case-sensitive search (default: true)",
				Default:     true,
			},
			"dry_run": {
				Type:        "boolean",
				Description: "Preview changes without modifying files (default: false)",
				Default:     false,
			},
			"backup": {
				Type:        "boolean",
				Description: "Create backup files with .bak extension (default: true)",
				Default:     true,
			},
			"max_files": {
				Type:        "integer",
				Description: "Maximum number of files to process (default: 100)",
				Default:     100,
			},
		},
		Required: []string{"pattern", "replacement"},
	}
}

// Execute runs the tool with the provided input and returns the result.
func (t *SearchReplaceTool) Execute(ctx context.Context, input map[string]interface{}) (*tools.Result, error) {
	// Extract and validate pattern
	patternRaw, ok := input["pattern"]
	if !ok {
		return nil, &tools.ErrInvalidInput{
			ToolName: t.Name(),
			Field:    "pattern",
			Reason:   "pattern is required",
		}
	}

	patternStr, ok := patternRaw.(string)
	if !ok {
		return nil, &tools.ErrInvalidInput{
			ToolName: t.Name(),
			Field:    "pattern",
			Reason:   fmt.Sprintf("pattern must be a string, got %T", patternRaw),
		}
	}

	if patternStr == "" {
		return nil, &tools.ErrInvalidInput{
			ToolName: t.Name(),
			Field:    "pattern",
			Reason:   "pattern cannot be empty",
		}
	}

	// Extract replacement string
	replacementRaw, ok := input["replacement"]
	if !ok {
		return nil, &tools.ErrInvalidInput{
			ToolName: t.Name(),
			Field:    "replacement",
			Reason:   "replacement is required",
		}
	}

	replacement, ok := replacementRaw.(string)
	if !ok {
		return nil, &tools.ErrInvalidInput{
			ToolName: t.Name(),
			Field:    "replacement",
			Reason:   fmt.Sprintf("replacement must be a string, got %T", replacementRaw),
		}
	}

	// Extract case sensitivity
	caseSensitive := extractBoolParam(input, "case_sensitive", true)

	// Compile regex pattern
	var pattern *regexp.Regexp
	var err error
	if caseSensitive {
		pattern, err = regexp.Compile(patternStr)
	} else {
		pattern, err = regexp.Compile("(?i)" + patternStr)
	}
	if err != nil {
		return nil, &tools.ErrInvalidInput{
			ToolName: t.Name(),
			Field:    "pattern",
			Reason:   fmt.Sprintf("invalid regex pattern: %v", err),
		}
	}

	// Extract path with default
	searchPath := t.sandbox.WorkingDirectory
	if pathRaw, exists := input["path"]; exists {
		p, ok := pathRaw.(string)
		if !ok {
			return nil, &tools.ErrInvalidInput{
				ToolName: t.Name(),
				Field:    "path",
				Reason:   fmt.Sprintf("path must be a string, got %T", pathRaw),
			}
		}

		// Resolve and validate path
		resolvedPath, err := t.sandbox.resolveAbsolutePath(p)
		if err != nil {
			return nil, &tools.ErrInvalidInput{
				ToolName: t.Name(),
				Field:    "path",
				Reason:   fmt.Sprintf("cannot resolve path: %v", err),
			}
		}

		if err := t.sandbox.ValidatePath(resolvedPath); err != nil {
			if permErr, ok := err.(*tools.ErrPermissionDenied); ok {
				permErr.ToolName = t.Name()
			}
			return nil, err
		}

		searchPath = resolvedPath
	} else if searchPath != "" {
		// Validate default path
		if err := t.sandbox.ValidatePath(searchPath); err != nil {
			return nil, &tools.ErrExecutionFailed{
				ToolName: t.Name(),
				Reason:   fmt.Sprintf("default search path is invalid: %v", err),
			}
		}
	}

	// Extract other parameters
	filePattern := extractStringParam(input, "file_pattern", "*")
	recursive := extractBoolParam(input, "recursive", false)
	dryRun := extractBoolParam(input, "dry_run", false)
	backup := extractBoolParam(input, "backup", true)
	maxFiles := extractIntParam(input, "max_files", 100)

	// Validate maxFiles
	if maxFiles <= 0 || maxFiles > 1000 {
		return nil, &tools.ErrInvalidInput{
			ToolName: t.Name(),
			Field:    "max_files",
			Reason:   "max_files must be between 1 and 1000",
		}
	}

	// Build replace options
	opts := &replaceOptions{
		pattern:     pattern,
		replacement: replacement,
		filePattern: filePattern,
		recursive:   recursive,
		dryRun:      dryRun,
		backup:      backup,
		maxFiles:    maxFiles,
	}

	// Perform the replacement
	result, err := t.replace(ctx, searchPath, opts)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Category returns the category this tool belongs to.
func (t *SearchReplaceTool) Category() tools.Category {
	return tools.CategoryText
}

// RequiresConfirmation returns true if this tool requires user confirmation before execution.
func (t *SearchReplaceTool) RequiresConfirmation() bool {
	return true // Modifying files requires confirmation
}

// replaceOptions contains replacement configuration
type replaceOptions struct {
	pattern     *regexp.Regexp
	replacement string
	filePattern string
	recursive   bool
	dryRun      bool
	backup      bool
	maxFiles    int
}

// replace performs the actual find and replace operation
func (t *SearchReplaceTool) replace(ctx context.Context, searchPath string, opts *replaceOptions) (*tools.Result, error) {
	// Check if path exists
	info, err := os.Stat(searchPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &tools.ErrExecutionFailed{
				ToolName: t.Name(),
				Reason:   fmt.Sprintf("search path does not exist: %s", searchPath),
			}
		}
		return nil, &tools.ErrExecutionFailed{
			ToolName: t.Name(),
			Reason:   fmt.Sprintf("cannot access search path: %s", searchPath),
			Cause:    err,
		}
	}

	var filesToProcess []string

	if info.IsDir() {
		// Find files in directory
		if opts.recursive {
			filesToProcess, err = t.findFilesRecursive(searchPath, opts.filePattern, opts.maxFiles)
		} else {
			filesToProcess, err = t.findFilesNonRecursive(searchPath, opts.filePattern, opts.maxFiles)
		}
		if err != nil {
			return nil, err
		}
	} else {
		// Single file
		filesToProcess = []string{searchPath}
	}

	// Process files
	var filesChanged []string
	var filesWithMatches []string
	totalReplacements := 0

	for _, filePath := range filesToProcess {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return nil, &tools.ErrExecutionFailed{
				ToolName: t.Name(),
				Reason:   "replacement cancelled",
				Cause:    ctx.Err(),
			}
		default:
		}

		replacements, err := t.replaceInFile(filePath, opts)
		if err != nil {
			// Skip files that can't be processed
			continue
		}

		if replacements > 0 {
			filesWithMatches = append(filesWithMatches, filePath)
			if !opts.dryRun {
				filesChanged = append(filesChanged, filePath)
			}
			totalReplacements += replacements
		}
	}

	// Format output
	output := t.formatReplaceResults(filesWithMatches, totalReplacements, opts)

	// Build result
	result := &tools.Result{
		Success: true,
		Output:  output,
		Metadata: map[string]interface{}{
			"pattern":            opts.pattern.String(),
			"replacement":        opts.replacement,
			"search_path":        searchPath,
			"total_replacements": totalReplacements,
			"files_changed":      len(filesChanged),
			"files_matched":      len(filesWithMatches),
			"dry_run":            opts.dryRun,
			"backup_created":     opts.backup && !opts.dryRun,
		},
	}

	return result, nil
}

// replaceInFile performs replacement in a single file
func (t *SearchReplaceTool) replaceInFile(filePath string, opts *replaceOptions) (int, error) {
	// Read file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return 0, err
	}

	originalContent := string(content)

	// Perform replacement
	newContent := opts.pattern.ReplaceAllString(originalContent, opts.replacement)

	// Check if any replacements were made
	if newContent == originalContent {
		return 0, nil
	}

	// Count replacements
	replacements := len(opts.pattern.FindAllString(originalContent, -1))

	// In dry-run mode, don't write file
	if opts.dryRun {
		return replacements, nil
	}

	// Create backup if requested
	if opts.backup {
		backupPath := filePath + ".bak"
		err := os.WriteFile(backupPath, content, 0644)
		if err != nil {
			return 0, fmt.Errorf("failed to create backup: %w", err)
		}
	}

	// Write modified content atomically
	tempPath := filePath + ".tmp"
	err = os.WriteFile(tempPath, []byte(newContent), 0644)
	if err != nil {
		return 0, fmt.Errorf("failed to write temp file: %w", err)
	}

	// Get original file permissions
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		os.Remove(tempPath)
		return 0, fmt.Errorf("failed to stat original file: %w", err)
	}

	// Set permissions on temp file
	err = os.Chmod(tempPath, fileInfo.Mode())
	if err != nil {
		os.Remove(tempPath)
		return 0, fmt.Errorf("failed to set permissions: %w", err)
	}

	// Atomically replace original file
	err = os.Rename(tempPath, filePath)
	if err != nil {
		os.Remove(tempPath)
		return 0, fmt.Errorf("failed to replace file: %w", err)
	}

	return replacements, nil
}

// findFilesRecursive finds files recursively matching the pattern
func (t *SearchReplaceTool) findFilesRecursive(dir, pattern string, maxFiles int) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files with errors
		}

		if info.IsDir() {
			return nil
		}

		// Check file pattern
		matched, err := filepath.Match(pattern, filepath.Base(path))
		if err != nil {
			return nil
		}

		if matched {
			files = append(files, path)
			if len(files) >= maxFiles {
				return filepath.SkipDir
			}
		}

		return nil
	})

	return files, err
}

// findFilesNonRecursive finds files in a directory without recursion
func (t *SearchReplaceTool) findFilesNonRecursive(dir, pattern string, maxFiles int) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if len(files) >= maxFiles {
			break
		}

		if entry.IsDir() {
			continue
		}

		matched, err := filepath.Match(pattern, entry.Name())
		if err != nil {
			continue
		}

		if matched {
			files = append(files, filepath.Join(dir, entry.Name()))
		}
	}

	return files, nil
}

// formatReplaceResults formats replacement results for output
func (t *SearchReplaceTool) formatReplaceResults(files []string, totalReplacements int, opts *replaceOptions) string {
	var sb strings.Builder

	if opts.dryRun {
		sb.WriteString("DRY RUN - No files were modified\n\n")
	}

	if totalReplacements == 0 {
		return "No matches found for pattern: " + opts.pattern.String()
	}

	sb.WriteString(fmt.Sprintf("Total replacements: %d\n", totalReplacements))
	sb.WriteString(fmt.Sprintf("Files affected: %d\n\n", len(files)))

	if opts.dryRun {
		sb.WriteString("The following files would be modified:\n")
	} else {
		sb.WriteString("Modified files:\n")
	}

	for _, file := range files {
		sb.WriteString(fmt.Sprintf("  - %s\n", file))
	}

	if opts.backup && !opts.dryRun {
		sb.WriteString("\nBackup files created with .bak extension\n")
	}

	return sb.String()
}
