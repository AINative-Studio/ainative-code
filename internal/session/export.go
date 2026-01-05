package session

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

//go:embed templates/*.tmpl
var templatesFS embed.FS

// ExporterOptions holds configuration for the exporter
type ExporterOptions struct {
	// TemplateDir is the directory containing custom templates
	TemplateDir string

	// IncludeMetadata controls whether to include detailed metadata in exports
	IncludeMetadata bool

	// PrettyPrint controls JSON formatting
	PrettyPrint bool
}

// Exporter handles session export operations
type Exporter struct {
	options *ExporterOptions
}

// NewExporter creates a new Exporter instance
func NewExporter(options *ExporterOptions) *Exporter {
	if options == nil {
		options = &ExporterOptions{
			IncludeMetadata: true,
			PrettyPrint:     true,
		}
	}
	return &Exporter{
		options: options,
	}
}

// ExportData represents the data passed to templates
type ExportData struct {
	Session  *Session
	Messages []*Message
	Metadata ExportMetadata
}

// ExportMetadata contains export-specific metadata
type ExportMetadata struct {
	ExportedAt    time.Time
	ExporterName  string
	ExporterVer   string
	MessageCount  int
	TotalTokens   int64
	FirstMessage  time.Time
	LastMessage   time.Time
	Provider      string
}

// ExportToJSON exports session to JSON format
func (e *Exporter) ExportToJSON(w io.Writer, session *Session, messages []*Message) error {
	if w == nil {
		return fmt.Errorf("writer cannot be nil")
	}
	if session == nil {
		return fmt.Errorf("session cannot be nil")
	}

	// Convert to proper slice type
	messageSlice := make([]Message, 0, len(messages))
	for _, msg := range messages {
		if msg != nil {
			messageSlice = append(messageSlice, *msg)
		}
	}

	export := SessionExport{
		Session:  *session,
		Messages: messageSlice,
	}

	encoder := json.NewEncoder(w)
	if e.options.PrettyPrint {
		encoder.SetIndent("", "  ")
	}

	if err := encoder.Encode(export); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}

// ExportToJSONWithContext exports session to JSON with context support
func (e *Exporter) ExportToJSONWithContext(ctx context.Context, w io.Writer, session *Session, messages []*Message) error {
	// Check context before starting
	if ctx.Err() != nil {
		return fmt.Errorf("context error: %w", ctx.Err())
	}

	return e.ExportToJSON(w, session, messages)
}

// ExportToMarkdown exports session to Markdown format
func (e *Exporter) ExportToMarkdown(w io.Writer, session *Session, messages []*Message) error {
	if w == nil {
		return fmt.Errorf("writer cannot be nil")
	}
	if session == nil {
		return fmt.Errorf("session cannot be nil")
	}

	// Load markdown template
	tmpl, err := e.loadTemplate("markdown.tmpl")
	if err != nil {
		return fmt.Errorf("failed to load markdown template: %w", err)
	}

	// Prepare export data
	data := e.prepareExportData(session, messages)

	// Execute template
	if err := tmpl.Execute(w, data); err != nil {
		return fmt.Errorf("failed to execute markdown template: %w", err)
	}

	return nil
}

// ExportToHTML exports session to HTML format
func (e *Exporter) ExportToHTML(w io.Writer, session *Session, messages []*Message) error {
	if w == nil {
		return fmt.Errorf("writer cannot be nil")
	}
	if session == nil {
		return fmt.Errorf("session cannot be nil")
	}

	// Load HTML template
	tmpl, err := e.loadTemplate("html.tmpl")
	if err != nil {
		return fmt.Errorf("failed to load HTML template: %w", err)
	}

	// Prepare export data
	data := e.prepareExportData(session, messages)

	// Execute template
	if err := tmpl.Execute(w, data); err != nil {
		return fmt.Errorf("failed to execute HTML template: %w", err)
	}

	return nil
}

// ExportWithTemplate exports session using a custom template
func (e *Exporter) ExportWithTemplate(w io.Writer, templatePath string, session *Session, messages []*Message) error {
	if w == nil {
		return fmt.Errorf("writer cannot be nil")
	}
	if session == nil {
		return fmt.Errorf("session cannot be nil")
	}

	var tmpl *template.Template
	var err error

	if templatePath == "" {
		// Use default markdown template
		tmpl, err = e.loadTemplate("markdown.tmpl")
	} else {
		// Load custom template
		tmpl, err = e.loadCustomTemplate(templatePath)
	}

	if err != nil {
		return fmt.Errorf("failed to load template: %w", err)
	}

	// Prepare export data
	data := e.prepareExportData(session, messages)

	// Execute template
	if err := tmpl.Execute(w, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

// loadTemplate loads a built-in template from embedded filesystem
func (e *Exporter) loadTemplate(name string) (*template.Template, error) {
	// Read template from embedded filesystem
	templatePath := filepath.Join("templates", name)
	data, err := templatesFS.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded template %s: %w", name, err)
	}

	// Parse template with helper functions
	tmpl, err := template.New(name).Funcs(e.templateFuncs()).Parse(string(data))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template %s: %w", name, err)
	}

	return tmpl, nil
}

// loadCustomTemplate loads a custom template from filesystem
func (e *Exporter) loadCustomTemplate(path string) (*template.Template, error) {
	// Check if file exists
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("template file not found: %w", err)
	}

	// Read template file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file: %w", err)
	}

	// Parse template with helper functions
	name := filepath.Base(path)
	tmpl, err := template.New(name).Funcs(e.templateFuncs()).Parse(string(data))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	return tmpl, nil
}

// templateFuncs returns template helper functions
func (e *Exporter) templateFuncs() template.FuncMap {
	return template.FuncMap{
		"formatTime": func(t time.Time) string {
			return t.Format("2006-01-02 15:04:05 MST")
		},
		"formatTimeISO": func(t time.Time) string {
			return t.Format(time.RFC3339)
		},
		"formatDate": func(t time.Time) string {
			return t.Format("2006-01-02")
		},
		"add": func(a, b int) int {
			return a + b
		},
		"upper": func(s string) string {
			return strings.ToUpper(s)
		},
		"lower": func(s string) string {
			return strings.ToLower(s)
		},
		"title": func(s string) string {
			return strings.Title(strings.ToLower(s))
		},
		"truncate": func(s string, length int) string {
			if len(s) <= length {
				return s
			}
			return s[:length] + "..."
		},
		"nl2br": func(s string) template.HTML {
			return template.HTML(strings.ReplaceAll(template.HTMLEscapeString(s), "\n", "<br>"))
		},
		"escapeHTML": func(s string) string {
			return template.HTMLEscapeString(s)
		},
		"markdownCode": func(content, language string) string {
			return fmt.Sprintf("```%s\n%s\n```", language, content)
		},
		"hasCodeBlock": func(content string) bool {
			return strings.Contains(content, "```")
		},
		"roleClass": func(role MessageRole) string {
			return fmt.Sprintf("message-%s", strings.ToLower(string(role)))
		},
		"roleLabel": func(role MessageRole) string {
			switch role {
			case RoleUser:
				return "User"
			case RoleAssistant:
				return "Assistant"
			case RoleSystem:
				return "System"
			case RoleTool:
				return "Tool"
			default:
				return string(role)
			}
		},
	}
}

// prepareExportData prepares data for template execution
func (e *Exporter) prepareExportData(session *Session, messages []*Message) ExportData {
	metadata := ExportMetadata{
		ExportedAt:   time.Now().UTC(),
		ExporterName: "AINative Code Session Exporter",
		ExporterVer:  "1.0.0",
		MessageCount: len(messages),
	}

	// Calculate total tokens and message times
	var totalTokens int64
	for i, msg := range messages {
		if msg.TokensUsed != nil {
			totalTokens += *msg.TokensUsed
		}
		if i == 0 {
			metadata.FirstMessage = msg.Timestamp
		}
		if i == len(messages)-1 {
			metadata.LastMessage = msg.Timestamp
		}
	}
	metadata.TotalTokens = totalTokens

	// Extract provider from session settings
	if session.Settings != nil {
		if provider, ok := session.Settings["provider"].(string); ok {
			metadata.Provider = provider
		}
	}

	return ExportData{
		Session:  session,
		Messages: messages,
		Metadata: metadata,
	}
}

// ExportToFile exports session to a file with the specified format
func (e *Exporter) ExportToFile(filePath string, format ExportFormat, session *Session, messages []*Message) error {
	if !format.IsValid() {
		return fmt.Errorf("invalid export format: %s", format)
	}

	// Create output file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	// Export to appropriate format
	switch format {
	case ExportFormatJSON:
		return e.ExportToJSON(file, session, messages)
	case ExportFormatMarkdown:
		return e.ExportToMarkdown(file, session, messages)
	case ExportFormatHTML:
		return e.ExportToHTML(file, session, messages)
	default:
		return fmt.Errorf("unsupported export format: %s", format)
	}
}

// DetectCodeLanguage attempts to detect the programming language from code content
func DetectCodeLanguage(content string) string {
	// Simple heuristics for common languages
	if strings.Contains(content, "package main") || strings.Contains(content, "func ") {
		return "go"
	}
	if strings.Contains(content, "def ") || strings.Contains(content, "import ") {
		return "python"
	}
	if strings.Contains(content, "function ") || strings.Contains(content, "const ") {
		return "javascript"
	}
	if strings.Contains(content, "public class ") || strings.Contains(content, "public static") {
		return "java"
	}
	return ""
}

// FormatCodeBlock formats a code block for markdown
func FormatCodeBlock(content, language string) string {
	if language == "" {
		language = DetectCodeLanguage(content)
	}
	return fmt.Sprintf("```%s\n%s\n```", language, content)
}
