package session

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test data helpers
func createTestSessionExportData() (*Session, []*Message) {
	now := time.Now().UTC()
	model := "claude-3-sonnet"
	temp := 0.7
	maxTokens := int64(4096)

	session := &Session{
		ID:          "test-session-123",
		Name:        "Test Conversation",
		CreatedAt:   now.Add(-1 * time.Hour),
		UpdatedAt:   now,
		Status:      StatusActive,
		Model:       &model,
		Temperature: &temp,
		MaxTokens:   &maxTokens,
		Settings: map[string]any{
			"provider": "anthropic",
			"debug":    true,
		},
	}

	tokens1 := int64(150)
	tokens2 := int64(450)
	finishReason := "stop"

	messages := []*Message{
		{
			ID:        "msg-1",
			SessionID: session.ID,
			Role:      RoleUser,
			Content:   "Hello! Can you help me with Go programming?",
			Timestamp: now.Add(-45 * time.Minute),
		},
		{
			ID:           "msg-2",
			SessionID:    session.ID,
			Role:         RoleAssistant,
			Content:      "Of course! I'd be happy to help you with Go programming.\n\nHere's a simple example:\n\n```go\npackage main\n\nimport \"fmt\"\n\nfunc main() {\n    fmt.Println(\"Hello, World!\")\n}\n```\n\nWhat specific topic would you like to explore?",
			Timestamp:    now.Add(-44 * time.Minute),
			TokensUsed:   &tokens1,
			Model:        &model,
			FinishReason: &finishReason,
			Metadata: map[string]any{
				"temperature": 0.7,
				"provider":    "anthropic",
			},
		},
		{
			ID:        "msg-3",
			SessionID: session.ID,
			Role:      RoleUser,
			Content:   "Can you explain goroutines?",
			Timestamp: now.Add(-40 * time.Minute),
		},
		{
			ID:           "msg-4",
			SessionID:    session.ID,
			Role:         RoleAssistant,
			Content:      "Goroutines are lightweight threads managed by the Go runtime. They allow concurrent execution of functions.\n\n**Key Features:**\n- Very lightweight (only a few KB of stack space)\n- Multiplexed onto OS threads\n- Managed by Go scheduler\n\n**Example:**\n```go\ngo func() {\n    fmt.Println(\"Running concurrently!\")\n}()\n```",
			Timestamp:    now.Add(-39 * time.Minute),
			TokensUsed:   &tokens2,
			Model:        &model,
			FinishReason: &finishReason,
		},
	}

	return session, messages
}

// TestExportToJSON tests JSON export functionality
func TestExportToJSON(t *testing.T) {
	session, messages := createTestSessionExportData()

	t.Run("ValidJSONExport", func(t *testing.T) {
		var buf bytes.Buffer
		exporter := NewExporter(nil)

		err := exporter.ExportToJSON(&buf, session, messages)
		require.NoError(t, err)

		// Verify JSON is valid
		var exported SessionExport
		err = json.Unmarshal(buf.Bytes(), &exported)
		require.NoError(t, err)

		// Verify content
		assert.Equal(t, session.ID, exported.Session.ID)
		assert.Equal(t, session.Name, exported.Session.Name)
		assert.Equal(t, len(messages), len(exported.Messages))
		assert.Equal(t, messages[0].Content, exported.Messages[0].Content)
		assert.Equal(t, messages[1].Content, exported.Messages[1].Content)

		// Verify metadata is preserved
		assert.NotNil(t, exported.Session.Settings)
		assert.Equal(t, "anthropic", exported.Session.Settings["provider"])
		assert.NotNil(t, exported.Messages[1].Metadata)
	})

	t.Run("JSONContainsAllMetadata", func(t *testing.T) {
		var buf bytes.Buffer
		exporter := NewExporter(nil)

		err := exporter.ExportToJSON(&buf, session, messages)
		require.NoError(t, err)

		jsonStr := buf.String()
		assert.Contains(t, jsonStr, "test-session-123")
		assert.Contains(t, jsonStr, "claude-3-sonnet")
		assert.Contains(t, jsonStr, "anthropic")
		assert.Contains(t, jsonStr, "tokens_used")
	})

	t.Run("EmptyMessages", func(t *testing.T) {
		var buf bytes.Buffer
		exporter := NewExporter(nil)

		err := exporter.ExportToJSON(&buf, session, []*Message{})
		require.NoError(t, err)

		var exported SessionExport
		err = json.Unmarshal(buf.Bytes(), &exported)
		require.NoError(t, err)
		assert.Equal(t, 0, len(exported.Messages))
	})
}

// TestExportToMarkdown tests Markdown export functionality
func TestExportToMarkdown(t *testing.T) {
	session, messages := createTestSessionExportData()

	t.Run("ValidMarkdownExport", func(t *testing.T) {
		var buf bytes.Buffer
		exporter := NewExporter(nil)

		err := exporter.ExportToMarkdown(&buf, session, messages)
		require.NoError(t, err)

		content := buf.String()

		// Verify structure
		assert.Contains(t, content, "# Test Conversation")
		assert.Contains(t, content, "test-session-123")
		assert.Contains(t, content, "claude-3-sonnet")
		assert.Contains(t, content, "active")
	})

	t.Run("MarkdownCodeBlocksPreserved", func(t *testing.T) {
		var buf bytes.Buffer
		exporter := NewExporter(nil)

		err := exporter.ExportToMarkdown(&buf, session, messages)
		require.NoError(t, err)

		content := buf.String()

		// Verify code blocks are preserved
		assert.Contains(t, content, "```go")
		assert.Contains(t, content, "package main")
		assert.Contains(t, content, "fmt.Println")
		assert.Contains(t, content, "```")
	})

	t.Run("MarkdownRoleFormatting", func(t *testing.T) {
		var buf bytes.Buffer
		exporter := NewExporter(nil)

		err := exporter.ExportToMarkdown(&buf, session, messages)
		require.NoError(t, err)

		content := buf.String()

		// Verify role-based formatting
		assert.Contains(t, content, "## User")
		assert.Contains(t, content, "## Assistant")
		assert.Contains(t, content, "Hello! Can you help me with Go programming?")
		assert.Contains(t, content, "Can you explain goroutines?")
	})

	t.Run("MarkdownMetadataIncluded", func(t *testing.T) {
		var buf bytes.Buffer
		exporter := NewExporter(nil)

		err := exporter.ExportToMarkdown(&buf, session, messages)
		require.NoError(t, err)

		content := buf.String()

		// Verify metadata is included
		assert.Contains(t, content, "Temperature:")
		assert.Contains(t, content, "Max Tokens:")
		assert.Contains(t, content, "Tokens used:")
	})

	t.Run("MarkdownWithCustomTemplate", func(t *testing.T) {
		// Create temporary custom template
		tmpDir := t.TempDir()
		templatePath := filepath.Join(tmpDir, "custom.tmpl")

		customTemplate := `# Custom Export: {{.Session.Name}}
ID: {{.Session.ID}}
Total Messages: {{len .Messages}}

{{range .Messages}}
[{{.Role}}]: {{.Content}}
{{end}}`

		err := os.WriteFile(templatePath, []byte(customTemplate), 0644)
		require.NoError(t, err)

		var buf bytes.Buffer
		exporter := NewExporter(nil)

		err = exporter.ExportWithTemplate(&buf, templatePath, session, messages)
		require.NoError(t, err)

		content := buf.String()
		assert.Contains(t, content, "# Custom Export: Test Conversation")
		assert.Contains(t, content, "Total Messages: 4")
	})
}

// TestExportToHTML tests HTML export functionality
func TestExportToHTML(t *testing.T) {
	session, messages := createTestSessionExportData()

	t.Run("ValidHTMLExport", func(t *testing.T) {
		var buf bytes.Buffer
		exporter := NewExporter(nil)

		err := exporter.ExportToHTML(&buf, session, messages)
		require.NoError(t, err)

		content := buf.String()

		// Verify HTML structure
		assert.Contains(t, content, "<!DOCTYPE html>")
		assert.Contains(t, content, "<html")
		assert.Contains(t, content, "<head>")
		assert.Contains(t, content, "<body>")
		assert.Contains(t, content, "</html>")
	})

	t.Run("HTMLMetadataIncluded", func(t *testing.T) {
		var buf bytes.Buffer
		exporter := NewExporter(nil)

		err := exporter.ExportToHTML(&buf, session, messages)
		require.NoError(t, err)

		content := buf.String()

		// Verify metadata is included
		assert.Contains(t, content, "test-session-123")
		assert.Contains(t, content, "claude-3-sonnet")
		assert.Contains(t, content, "Test Conversation")
	})

	t.Run("HTMLSyntaxHighlighting", func(t *testing.T) {
		var buf bytes.Buffer
		exporter := NewExporter(nil)

		err := exporter.ExportToHTML(&buf, session, messages)
		require.NoError(t, err)

		content := buf.String()

		// Verify code content is present (HTML templates preserve code in message content)
		assert.Contains(t, content, "package main")
		assert.Contains(t, content, "fmt.Println")
		assert.Contains(t, content, "```go")
	})

	t.Run("HTMLStyling", func(t *testing.T) {
		var buf bytes.Buffer
		exporter := NewExporter(nil)

		err := exporter.ExportToHTML(&buf, session, messages)
		require.NoError(t, err)

		content := buf.String()

		// Verify CSS is included
		assert.Contains(t, content, "<style>")
		assert.Contains(t, content, "</style>")
		// Verify common CSS elements
		assert.Contains(t, content, "background")
		assert.Contains(t, content, "color")
	})

	t.Run("HTMLRoleDistinction", func(t *testing.T) {
		var buf bytes.Buffer
		exporter := NewExporter(nil)

		err := exporter.ExportToHTML(&buf, session, messages)
		require.NoError(t, err)

		content := buf.String()

		// Verify role-based CSS classes
		assert.Contains(t, content, "user")
		assert.Contains(t, content, "assistant")
	})
}

// TestExportWithTemplates tests template-based export
func TestExportWithTemplates(t *testing.T) {
	session, messages := createTestSessionExportData()

	t.Run("UseBuiltInMarkdownTemplate", func(t *testing.T) {
		var buf bytes.Buffer
		exporter := NewExporter(nil)

		err := exporter.ExportWithTemplate(&buf, "", session, messages)
		require.NoError(t, err)

		content := buf.String()
		assert.Contains(t, content, "Test Conversation")
	})

	t.Run("UseBuiltInHTMLTemplate", func(t *testing.T) {
		var buf bytes.Buffer
		exporter := NewExporter(nil)

		err := exporter.ExportToHTML(&buf, session, messages)
		require.NoError(t, err)

		content := buf.String()
		assert.Contains(t, content, "<!DOCTYPE html>")
	})

	t.Run("InvalidTemplatePath", func(t *testing.T) {
		var buf bytes.Buffer
		exporter := NewExporter(nil)

		err := exporter.ExportWithTemplate(&buf, "/nonexistent/template.tmpl", session, messages)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "template")
	})

	t.Run("CustomTemplateWithHelperFunctions", func(t *testing.T) {
		tmpDir := t.TempDir()
		templatePath := filepath.Join(tmpDir, "helpers.tmpl")

		customTemplate := `Session: {{.Session.Name}}
Created: {{formatTime .Session.CreatedAt}}
Messages: {{len .Messages}}
Has Model: {{if .Session.Model}}Yes{{else}}No{{end}}`

		err := os.WriteFile(templatePath, []byte(customTemplate), 0644)
		require.NoError(t, err)

		var buf bytes.Buffer
		exporter := NewExporter(nil)

		err = exporter.ExportWithTemplate(&buf, templatePath, session, messages)
		require.NoError(t, err)

		content := buf.String()
		assert.Contains(t, content, "Has Model: Yes")
		assert.Contains(t, content, "Messages: 4")
	})
}

// TestExportErrors tests error handling
func TestExportErrors(t *testing.T) {
	session, messages := createTestSessionExportData()

	t.Run("NilSession", func(t *testing.T) {
		var buf bytes.Buffer
		exporter := NewExporter(nil)

		err := exporter.ExportToJSON(&buf, nil, messages)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "session")
	})

	t.Run("NilMessages", func(t *testing.T) {
		var buf bytes.Buffer
		exporter := NewExporter(nil)

		// Nil messages should work (empty export)
		err := exporter.ExportToJSON(&buf, session, nil)
		require.NoError(t, err)

		var exported SessionExport
		err = json.Unmarshal(buf.Bytes(), &exported)
		require.NoError(t, err)
		assert.Equal(t, 0, len(exported.Messages))
	})

	t.Run("NilWriter", func(t *testing.T) {
		exporter := NewExporter(nil)

		err := exporter.ExportToJSON(nil, session, messages)
		assert.Error(t, err)
	})
}

// TestExportFormatValidation tests format validation
func TestExportFormatValidation(t *testing.T) {
	t.Run("ValidFormats", func(t *testing.T) {
		formats := []ExportFormat{
			ExportFormatJSON,
			ExportFormatMarkdown,
			ExportFormatHTML,
		}

		for _, format := range formats {
			assert.True(t, format.IsValid(), "Format %s should be valid", format)
		}
	})

	t.Run("InvalidFormat", func(t *testing.T) {
		format := ExportFormat("invalid")
		assert.False(t, format.IsValid())
	})
}

// TestExportMetadataPreservation tests that all metadata is preserved
func TestExportMetadataPreservation(t *testing.T) {
	session, messages := createTestSessionExportData()

	t.Run("JSONPreservesAllMetadata", func(t *testing.T) {
		var buf bytes.Buffer
		exporter := NewExporter(nil)

		err := exporter.ExportToJSON(&buf, session, messages)
		require.NoError(t, err)

		var exported SessionExport
		err = json.Unmarshal(buf.Bytes(), &exported)
		require.NoError(t, err)

		// Verify session metadata
		assert.Equal(t, *session.Model, *exported.Session.Model)
		assert.Equal(t, *session.Temperature, *exported.Session.Temperature)
		assert.Equal(t, *session.MaxTokens, *exported.Session.MaxTokens)
		assert.Equal(t, session.Settings["provider"], exported.Session.Settings["provider"])

		// Verify message metadata
		assert.NotNil(t, exported.Messages[1].TokensUsed)
		assert.Equal(t, *messages[1].TokensUsed, *exported.Messages[1].TokensUsed)
		assert.NotNil(t, exported.Messages[1].Metadata)
		assert.Equal(t, messages[1].Metadata["provider"], exported.Messages[1].Metadata["provider"])
	})
}

// TestExportCodeBlockFormatting tests code block preservation
func TestExportCodeBlockFormatting(t *testing.T) {
	session, messages := createTestSessionExportData()

	t.Run("MarkdownPreservesCodeBlocks", func(t *testing.T) {
		var buf bytes.Buffer
		exporter := NewExporter(nil)

		err := exporter.ExportToMarkdown(&buf, session, messages)
		require.NoError(t, err)

		content := buf.String()

		// Count code block markers
		goBlocks := strings.Count(content, "```go")
		closingBlocks := strings.Count(content, "```")

		assert.Equal(t, 2, goBlocks, "Should have 2 Go code blocks")
		// Each opening ``` should have a closing ``` (so double the count)
		assert.GreaterOrEqual(t, closingBlocks, goBlocks*2)
	})

	t.Run("HTMLPreservesCodeBlocks", func(t *testing.T) {
		var buf bytes.Buffer
		exporter := NewExporter(nil)

		err := exporter.ExportToHTML(&buf, session, messages)
		require.NoError(t, err)

		content := buf.String()

		// Verify code content is present in HTML (preserved in message content div)
		assert.Contains(t, content, "package main")
		assert.Contains(t, content, "fmt.Println")
		assert.Contains(t, content, "```go", "Code blocks should be preserved")
	})
}

// TestExportTemplateCustomization tests template customization
func TestExportTemplateCustomization(t *testing.T) {
	session, messages := createTestSessionExportData()

	t.Run("CustomMarkdownTemplate", func(t *testing.T) {
		tmpDir := t.TempDir()
		templatePath := filepath.Join(tmpDir, "custom_markdown.tmpl")

		customTemplate := `# {{.Session.Name}}

**Exported at:** {{formatTime .Session.UpdatedAt}}

## Conversation

{{range $i, $msg := .Messages}}
### Message {{add $i 1}} - {{$msg.Role}}

{{$msg.Content}}

{{if $msg.TokensUsed}}*Tokens: {{$msg.TokensUsed}}*{{end}}

---
{{end}}`

		err := os.WriteFile(templatePath, []byte(customTemplate), 0644)
		require.NoError(t, err)

		var buf bytes.Buffer
		exporter := NewExporter(nil)

		err = exporter.ExportWithTemplate(&buf, templatePath, session, messages)
		require.NoError(t, err)

		content := buf.String()
		assert.Contains(t, content, "# Test Conversation")
		assert.Contains(t, content, "### Message 1 - user")
		assert.Contains(t, content, "### Message 2 - assistant")
	})

	t.Run("TemplateDirectory", func(t *testing.T) {
		tmpDir := t.TempDir()
		exporter := NewExporter(&ExporterOptions{
			TemplateDir: tmpDir,
		})

		// Create a template in the directory
		templatePath := filepath.Join(tmpDir, "test.tmpl")
		template := `Simple: {{.Session.Name}}`
		err := os.WriteFile(templatePath, []byte(template), 0644)
		require.NoError(t, err)

		var buf bytes.Buffer
		err = exporter.ExportWithTemplate(&buf, templatePath, session, messages)
		require.NoError(t, err)

		content := buf.String()
		assert.Contains(t, content, "Simple: Test Conversation")
	})
}

// TestExportIntegration tests full export workflow
func TestExportIntegration(t *testing.T) {
	session, messages := createTestSessionExportData()

	t.Run("ExportToMultipleFormats", func(t *testing.T) {
		tmpDir := t.TempDir()
		exporter := NewExporter(nil)

		formats := map[ExportFormat]string{
			ExportFormatJSON:     "export.json",
			ExportFormatMarkdown: "export.md",
			ExportFormatHTML:     "export.html",
		}

		for format, filename := range formats {
			filePath := filepath.Join(tmpDir, filename)
			file, err := os.Create(filePath)
			require.NoError(t, err)

			switch format {
			case ExportFormatJSON:
				err = exporter.ExportToJSON(file, session, messages)
			case ExportFormatMarkdown:
				err = exporter.ExportToMarkdown(file, session, messages)
			case ExportFormatHTML:
				err = exporter.ExportToHTML(file, session, messages)
			}

			require.NoError(t, err)
			file.Close()

			// Verify file exists and has content
			info, err := os.Stat(filePath)
			require.NoError(t, err)
			assert.Greater(t, info.Size(), int64(100), "File %s should have content", filename)
		}
	})
}

// TestExportWithContext tests context handling
func TestExportWithContext(t *testing.T) {
	session, messages := createTestSessionExportData()

	t.Run("ExportWithContext", func(t *testing.T) {
		ctx := context.Background()
		var buf bytes.Buffer
		exporter := NewExporter(nil)

		err := exporter.ExportToJSONWithContext(ctx, &buf, session, messages)
		require.NoError(t, err)

		var exported SessionExport
		err = json.Unmarshal(buf.Bytes(), &exported)
		require.NoError(t, err)
		assert.Equal(t, session.ID, exported.Session.ID)
	})

	t.Run("CancelledContext", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		var buf bytes.Buffer
		exporter := NewExporter(nil)

		err := exporter.ExportToJSONWithContext(ctx, &buf, session, messages)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context")
	})
}
