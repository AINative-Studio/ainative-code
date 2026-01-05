package tui

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// RenderThinkingBlock renders a single thinking block with all styling
func RenderThinkingBlock(block *ThinkingBlock, config ThinkingConfig) string {
	var sb strings.Builder

	// Add depth indicator if enabled
	if config.ShowDepthIndicator && block.Depth > 0 {
		sb.WriteString(block.GetDepthIndicator())
	}

	// Add collapse/expand icon
	if block.Collapsed {
		icon := CollapsedIndicatorStyle.Render(GetCollapsedIcon())
		sb.WriteString(icon)
	} else {
		icon := ExpandedIndicatorStyle.Render(GetExpandedIcon())
		sb.WriteString(icon)
	}

	sb.WriteString(" ")

	// Add thinking label
	label := ThinkingLabelStyle.Render("Thinking")
	sb.WriteString(label)

	// Add timestamp if enabled
	if config.ShowTimestamps {
		timestamp := ThinkingLabelStyle.Render(
			fmt.Sprintf(" [%s]", block.Timestamp.Format("15:04:05")),
		)
		sb.WriteString(timestamp)
	}

	// Add line count indicator
	if !block.Collapsed {
		lineCount := ThinkingLabelStyle.Render(
			fmt.Sprintf(" (%d lines)", block.GetLineCount()),
		)
		sb.WriteString(lineCount)
	}

	sb.WriteString("\n")

	// Render content based on collapsed state
	if block.Collapsed {
		// Show preview when collapsed
		preview := block.GetPreview(config.MaxPreviewLength)
		if preview != "" {
			previewStyle := lipgloss.NewStyle().
				Foreground(ThinkingMutedColor).
				Italic(true)
			sb.WriteString("  ")
			sb.WriteString(previewStyle.Render(preview))
		}
	} else {
		// Show full content when expanded
		content := block.Content
		if config.SyntaxHighlighting {
			content = ApplySyntaxHighlighting(content)
		}

		// Apply styling and indentation
		styledContent := ThinkingContentStyle.Render(content)
		sb.WriteString(styledContent)
	}

	// Apply border based on depth
	borderStyle := ThinkingBorderStyle(block.Depth)
	return borderStyle.Render(sb.String())
}

// RenderAllThinkingBlocks renders all thinking blocks
func RenderAllThinkingBlocks(state *ThinkingState, config ThinkingConfig) string {
	if !state.ShowThinking || len(state.Blocks) == 0 {
		return ""
	}

	var sb strings.Builder

	// Add header separator
	separator := ThinkingLabelStyle.Render(strings.Repeat("─", 50))
	sb.WriteString(separator)
	sb.WriteString("\n")

	// Render each block
	for i, block := range state.GetVisibleBlocks() {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(RenderThinkingBlock(block, config))
	}

	// Add footer separator
	sb.WriteString("\n")
	sb.WriteString(separator)

	return sb.String()
}

// ApplySyntaxHighlighting applies basic syntax highlighting to thinking content
func ApplySyntaxHighlighting(content string) string {
	// Apply highlighting for common code patterns
	result := content

	// Highlight code blocks (```...```)
	codeBlockPattern := regexp.MustCompile("(?s)```([\\w]*)\\n(.*?)```")
	result = codeBlockPattern.ReplaceAllStringFunc(result, func(match string) string {
		parts := codeBlockPattern.FindStringSubmatch(match)
		if len(parts) >= 3 {
			language := parts[1]
			code := parts[2]

			// Apply code block styling
			header := ""
			if language != "" {
				header = InlineCodeStyle.Render(language) + "\n"
			}

			highlightedCode := highlightCode(code)
			return header + CodeBlockStyle.Render(highlightedCode)
		}
		return match
	})

	// Highlight inline code (`...`)
	inlineCodePattern := regexp.MustCompile("`([^`]+)`")
	result = inlineCodePattern.ReplaceAllStringFunc(result, func(match string) string {
		code := strings.Trim(match, "`")
		return InlineCodeStyle.Render(code)
	})

	return result
}

// highlightCode applies syntax highlighting to code snippets
func highlightCode(code string) string {
	result := code

	// Highlight keywords (common programming keywords)
	keywords := []string{
		"function", "const", "let", "var", "if", "else", "for", "while",
		"return", "class", "interface", "type", "struct", "import", "export",
		"package", "func", "def", "async", "await", "try", "catch", "finally",
	}

	keywordStyle := lipgloss.NewStyle().Foreground(CodeKeywordColor).Bold(true)
	for _, keyword := range keywords {
		// Match whole words only
		pattern := regexp.MustCompile(`\b` + keyword + `\b`)
		result = pattern.ReplaceAllStringFunc(result, func(match string) string {
			return keywordStyle.Render(match)
		})
	}

	// Highlight strings ("..." and '...')
	stringPattern := regexp.MustCompile(`["']([^"']*?)["']`)
	stringStyle := lipgloss.NewStyle().Foreground(CodeStringColor)
	result = stringPattern.ReplaceAllStringFunc(result, func(match string) string {
		return stringStyle.Render(match)
	})

	// Highlight numbers
	numberPattern := regexp.MustCompile(`\b\d+\.?\d*\b`)
	numberStyle := lipgloss.NewStyle().Foreground(CodeNumberColor)
	result = numberPattern.ReplaceAllStringFunc(result, func(match string) string {
		return numberStyle.Render(match)
	})

	// Highlight comments (// and # style)
	commentPattern := regexp.MustCompile(`(//.*?$|#.*?$)`)
	commentStyle := lipgloss.NewStyle().Foreground(CodeCommentColor).Italic(true)
	result = commentPattern.ReplaceAllStringFunc(result, func(match string) string {
		return commentStyle.Render(match)
	})

	// Highlight function calls (word followed by parentheses)
	functionPattern := regexp.MustCompile(`\b([a-zA-Z_][a-zA-Z0-9_]*)\s*\(`)
	functionStyle := lipgloss.NewStyle().Foreground(CodeFunctionColor)
	result = functionPattern.ReplaceAllStringFunc(result, func(match string) string {
		parts := functionPattern.FindStringSubmatch(match)
		if len(parts) >= 2 {
			return functionStyle.Render(parts[1]) + "("
		}
		return match
	})

	return result
}

// RenderThinkingToggleHint renders a hint about the thinking toggle
func RenderThinkingToggleHint(showThinking bool) string {
	hintStyle := lipgloss.NewStyle().
		Foreground(ThinkingMutedColor).
		Italic(true)

	if showThinking {
		return hintStyle.Render("Press 't' to hide thinking | 'e' to expand all | 'c' to collapse all")
	}
	return hintStyle.Render("Press 't' to show thinking")
}

// RenderThinkingSummary renders a summary of thinking blocks
func RenderThinkingSummary(state *ThinkingState) string {
	if len(state.Blocks) == 0 {
		return ""
	}

	totalBlocks := len(state.Blocks)
	collapsedCount := 0
	totalLines := 0

	for _, block := range state.Blocks {
		if block.Collapsed {
			collapsedCount++
		}
		totalLines += block.GetLineCount()
	}

	summaryStyle := lipgloss.NewStyle().
		Foreground(ThinkingMutedColor).
		Italic(true)

	summary := fmt.Sprintf(
		"%s %d thinking blocks (%d collapsed, %d lines total)",
		GetThinkingIcon(),
		totalBlocks,
		collapsedCount,
		totalLines,
	)

	return summaryStyle.Render(summary)
}

// RenderThinkingHeader renders a header for the thinking section
func RenderThinkingHeader(state *ThinkingState) string {
	if !state.ShowThinking || len(state.Blocks) == 0 {
		return ""
	}

	var sb strings.Builder

	// Add icon and title
	header := ThinkingHeaderStyle.Render(
		fmt.Sprintf("%s Extended Thinking", GetThinkingIcon()),
	)
	sb.WriteString(header)

	// Add summary
	summary := RenderThinkingSummary(state)
	if summary != "" {
		sb.WriteString(" ")
		sb.WriteString(summary)
	}

	sb.WriteString("\n")

	// Add toggle hint
	hint := RenderThinkingToggleHint(state.ShowThinking)
	sb.WriteString(hint)
	sb.WriteString("\n")

	return sb.String()
}

// FormatThinkingContent formats thinking content for display
func FormatThinkingContent(content string, config ThinkingConfig) string {
	// Apply syntax highlighting if enabled
	if config.SyntaxHighlighting {
		content = ApplySyntaxHighlighting(content)
	}

	// Apply content styling
	return ThinkingContentStyle.Render(content)
}

// WrapThinkingContent wraps thinking content to a specific width
func WrapThinkingContent(content string, width int) string {
	if width <= 0 {
		return content
	}

	var wrapped strings.Builder
	lines := strings.Split(content, "\n")

	for i, line := range lines {
		if i > 0 {
			wrapped.WriteString("\n")
		}

		// Simple word wrapping
		if len(line) <= width {
			wrapped.WriteString(line)
			continue
		}

		words := strings.Fields(line)
		currentLine := ""

		for _, word := range words {
			testLine := currentLine
			if testLine != "" {
				testLine += " "
			}
			testLine += word

			if len(testLine) > width {
				if currentLine != "" {
					wrapped.WriteString(currentLine)
					wrapped.WriteString("\n")
				}
				currentLine = word
			} else {
				currentLine = testLine
			}
		}

		if currentLine != "" {
			wrapped.WriteString(currentLine)
		}
	}

	return wrapped.String()
}
