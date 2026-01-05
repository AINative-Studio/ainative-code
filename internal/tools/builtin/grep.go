// Package builtin provides built-in tools for the tool execution framework.
package builtin

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/AINative-studio/ainative-code/internal/tools"
)

// GrepTool implements a tool for searching across files with regex support.
type GrepTool struct {
	sandbox *Sandbox
}

// NewGrepTool creates a new GrepTool instance with the specified sandbox.
func NewGrepTool(sandbox *Sandbox) *GrepTool {
	return &GrepTool{
		sandbox: sandbox,
	}
}

// Name returns the unique name of the tool.
func (t *GrepTool) Name() string {
	return "grep"
}

// Description returns a human-readable description of what the tool does.
func (t *GrepTool) Description() string {
	return "Searches for patterns across files using regex with context lines, file filtering, and performance limits"
}

// Schema returns the JSON schema for the tool's input parameters.
func (t *GrepTool) Schema() tools.ToolSchema {
	maxPatternLength := 1024
	maxGlobLength := 256
	defaultMaxMatches := 1000
	defaultMaxFiles := 100

	return tools.ToolSchema{
		Type: "object",
		Properties: map[string]tools.PropertyDef{
			"pattern": {
				Type:        "string",
				Description: "Regular expression pattern to search for",
				MaxLength:   &maxPatternLength,
			},
			"path": {
				Type:        "string",
				Description: "Directory or file path to search (default: current working directory)",
			},
			"file_pattern": {
				Type:        "string",
				Description: "Glob pattern to filter files (e.g., '*.go', '*.md')",
				MaxLength:   &maxGlobLength,
			},
			"recursive": {
				Type:        "boolean",
				Description: "Search recursively through subdirectories (default: true)",
				Default:     true,
			},
			"case_sensitive": {
				Type:        "boolean",
				Description: "Perform case-sensitive search (default: true)",
				Default:     true,
			},
			"context_before": {
				Type:        "integer",
				Description: "Number of lines to show before each match (default: 0)",
				Default:     0,
			},
			"context_after": {
				Type:        "integer",
				Description: "Number of lines to show after each match (default: 0)",
				Default:     0,
			},
			"max_matches": {
				Type:        "integer",
				Description: fmt.Sprintf("Maximum number of matches to return (default: %d)", defaultMaxMatches),
				Default:     defaultMaxMatches,
			},
			"max_files": {
				Type:        "integer",
				Description: fmt.Sprintf("Maximum number of files to search (default: %d)", defaultMaxFiles),
				Default:     defaultMaxFiles,
			},
			"show_line_numbers": {
				Type:        "boolean",
				Description: "Show line numbers in results (default: true)",
				Default:     true,
			},
		},
		Required: []string{"pattern"},
	}
}

// Execute runs the tool with the provided input and returns the result.
func (t *GrepTool) Execute(ctx context.Context, input map[string]interface{}) (*tools.Result, error) {
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

	// Extract case sensitivity
	caseSensitive := true
	if csRaw, exists := input["case_sensitive"]; exists {
		cs, ok := csRaw.(bool)
		if !ok {
			return nil, &tools.ErrInvalidInput{
				ToolName: t.Name(),
				Field:    "case_sensitive",
				Reason:   fmt.Sprintf("case_sensitive must be a boolean, got %T", csRaw),
			}
		}
		caseSensitive = cs
	}

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
	recursive := extractBoolParam(input, "recursive", true)
	contextBefore := extractIntParam(input, "context_before", 0)
	contextAfter := extractIntParam(input, "context_after", 0)
	maxMatches := extractIntParam(input, "max_matches", 1000)
	maxFiles := extractIntParam(input, "max_files", 100)
	showLineNumbers := extractBoolParam(input, "show_line_numbers", true)

	// Validate numeric parameters
	if contextBefore < 0 || contextBefore > 10 {
		return nil, &tools.ErrInvalidInput{
			ToolName: t.Name(),
			Field:    "context_before",
			Reason:   "context_before must be between 0 and 10",
		}
	}
	if contextAfter < 0 || contextAfter > 10 {
		return nil, &tools.ErrInvalidInput{
			ToolName: t.Name(),
			Field:    "context_after",
			Reason:   "context_after must be between 0 and 10",
		}
	}
	if maxMatches <= 0 || maxMatches > 10000 {
		return nil, &tools.ErrInvalidInput{
			ToolName: t.Name(),
			Field:    "max_matches",
			Reason:   "max_matches must be between 1 and 10000",
		}
	}
	if maxFiles <= 0 || maxFiles > 1000 {
		return nil, &tools.ErrInvalidInput{
			ToolName: t.Name(),
			Field:    "max_files",
			Reason:   "max_files must be between 1 and 1000",
		}
	}

	// Build search options
	opts := &grepOptions{
		pattern:         pattern,
		filePattern:     filePattern,
		recursive:       recursive,
		contextBefore:   contextBefore,
		contextAfter:    contextAfter,
		maxMatches:      maxMatches,
		maxFiles:        maxFiles,
		showLineNumbers: showLineNumbers,
	}

	// Perform the search
	results, err := t.search(ctx, searchPath, opts)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// Category returns the category this tool belongs to.
func (t *GrepTool) Category() tools.Category {
	return tools.CategoryText
}

// RequiresConfirmation returns true if this tool requires user confirmation before execution.
func (t *GrepTool) RequiresConfirmation() bool {
	return false // Reading/searching is generally safe
}

// grepOptions contains search configuration
type grepOptions struct {
	pattern         *regexp.Regexp
	filePattern     string
	recursive       bool
	contextBefore   int
	contextAfter    int
	maxMatches      int
	maxFiles        int
	showLineNumbers bool
}

// search performs the actual grep operation
func (t *GrepTool) search(ctx context.Context, searchPath string, opts *grepOptions) (*tools.Result, error) {
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

	var filesToSearch []string

	if info.IsDir() {
		// Search directory
		if opts.recursive {
			filesToSearch, err = t.findFilesRecursive(searchPath, opts.filePattern, opts.maxFiles)
		} else {
			filesToSearch, err = t.findFilesNonRecursive(searchPath, opts.filePattern, opts.maxFiles)
		}
		if err != nil {
			return nil, err
		}
	} else {
		// Single file
		filesToSearch = []string{searchPath}
	}

	// Search files
	var matches []matchResult
	totalMatches := 0
	filesSearched := 0

	for _, filePath := range filesToSearch {
		if filesSearched >= opts.maxFiles {
			break
		}

		// Check context cancellation
		select {
		case <-ctx.Done():
			return nil, &tools.ErrExecutionFailed{
				ToolName: t.Name(),
				Reason:   "search cancelled",
				Cause:    ctx.Err(),
			}
		default:
		}

		fileMatches, err := t.searchFile(filePath, opts)
		if err != nil {
			// Skip files that can't be read
			continue
		}

		if len(fileMatches) > 0 {
			matches = append(matches, fileMatches...)
			totalMatches += len(fileMatches)
			filesSearched++
		}

		if totalMatches >= opts.maxMatches {
			break
		}
	}

	// Format output
	output := t.formatResults(matches, opts)

	// Build result
	result := &tools.Result{
		Success: true,
		Output:  output,
		Metadata: map[string]interface{}{
			"pattern":        opts.pattern.String(),
			"search_path":    searchPath,
			"total_matches":  totalMatches,
			"files_searched": filesSearched,
			"files_with_matches": len(filesToSearch),
			"truncated":      totalMatches >= opts.maxMatches,
		},
	}

	return result, nil
}

// matchResult represents a single match
type matchResult struct {
	filePath   string
	lineNumber int
	line       string
	context    []string
}

// searchFile searches a single file for matches
func (t *GrepTool) searchFile(filePath string, opts *grepOptions) ([]matchResult, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var matches []matchResult
	var lineBuffer []string
	scanner := bufio.NewScanner(file)
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()

		// Keep context buffer
		if opts.contextBefore > 0 {
			lineBuffer = append(lineBuffer, line)
			if len(lineBuffer) > opts.contextBefore {
				lineBuffer = lineBuffer[1:]
			}
		}

		// Check if line matches
		if opts.pattern.MatchString(line) {
			match := matchResult{
				filePath:   filePath,
				lineNumber: lineNumber,
				line:       line,
			}

			// Add before context
			if opts.contextBefore > 0 && len(lineBuffer) > 1 {
				match.context = append(match.context, lineBuffer[:len(lineBuffer)-1]...)
			}

			// Add after context
			if opts.contextAfter > 0 {
				afterLines := []string{}
				for i := 0; i < opts.contextAfter && scanner.Scan(); i++ {
					lineNumber++
					afterLines = append(afterLines, scanner.Text())
				}
				match.context = append(match.context, afterLines...)
			}

			matches = append(matches, match)

			if len(matches) >= opts.maxMatches {
				break
			}
		}
	}

	return matches, scanner.Err()
}

// findFilesRecursive finds files recursively matching the pattern
func (t *GrepTool) findFilesRecursive(dir, pattern string, maxFiles int) ([]string, error) {
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
func (t *GrepTool) findFilesNonRecursive(dir, pattern string, maxFiles int) ([]string, error) {
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

// formatResults formats match results for output
func (t *GrepTool) formatResults(matches []matchResult, opts *grepOptions) string {
	if len(matches) == 0 {
		return "No matches found"
	}

	var sb strings.Builder
	currentFile := ""

	for _, match := range matches {
		if match.filePath != currentFile {
			if currentFile != "" {
				sb.WriteString("\n")
			}
			sb.WriteString(fmt.Sprintf("=== %s ===\n", match.filePath))
			currentFile = match.filePath
		}

		if opts.showLineNumbers {
			sb.WriteString(fmt.Sprintf("%d: %s\n", match.lineNumber, match.line))
		} else {
			sb.WriteString(fmt.Sprintf("%s\n", match.line))
		}

		// Add context if present
		if len(match.context) > 0 {
			for _, contextLine := range match.context {
				sb.WriteString(fmt.Sprintf("    %s\n", contextLine))
			}
		}
	}

	return sb.String()
}

// Helper functions to extract parameters with defaults
func extractStringParam(input map[string]interface{}, key, defaultValue string) string {
	if val, exists := input[key]; exists {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}

func extractBoolParam(input map[string]interface{}, key string, defaultValue bool) bool {
	if val, exists := input[key]; exists {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return defaultValue
}

func extractIntParam(input map[string]interface{}, key string, defaultValue int) int {
	if val, exists := input[key]; exists {
		switch v := val.(type) {
		case int:
			return v
		case float64:
			return int(v)
		case int64:
			return int(v)
		}
	}
	return defaultValue
}
