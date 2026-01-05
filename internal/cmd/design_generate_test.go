package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/AINative-studio/ainative-code/internal/design"
)

func TestReadTokensFromFile(t *testing.T) {
	t.Run("should read tokens from JSON array", func(t *testing.T) {
		// Given
		tmpDir := t.TempDir()
		tokensFile := filepath.Join(tmpDir, "tokens.json")

		tokens := []design.Token{
			{Name: "primary", Type: "color", Value: "#007bff"},
			{Name: "secondary", Type: "color", Value: "#6c757d"},
		}

		data, _ := json.Marshal(tokens)
		os.WriteFile(tokensFile, data, 0644)

		// When
		result, err := readTokensFromFile(tokensFile)

		// Then
		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "primary", result[0].Name)
	})

	t.Run("should read tokens from object with tokens field", func(t *testing.T) {
		// Given
		tmpDir := t.TempDir()
		tokensFile := filepath.Join(tmpDir, "tokens.json")

		obj := struct {
			Tokens []design.Token `json:"tokens"`
		}{
			Tokens: []design.Token{
				{Name: "test", Type: "color", Value: "#000"},
			},
		}

		data, _ := json.Marshal(obj)
		os.WriteFile(tokensFile, data, 0644)

		// When
		result, err := readTokensFromFile(tokensFile)

		// Then
		require.NoError(t, err)
		assert.Len(t, result, 1)
	})

	t.Run("should return error for non-existent file", func(t *testing.T) {
		// When
		_, err := readTokensFromFile("/nonexistent/file.json")

		// Then
		assert.Error(t, err)
	})

	t.Run("should return error for invalid JSON", func(t *testing.T) {
		// Given
		tmpDir := t.TempDir()
		tokensFile := filepath.Join(tmpDir, "invalid.json")
		os.WriteFile(tokensFile, []byte("invalid json"), 0644)

		// When
		_, err := readTokensFromFile(tokensFile)

		// Then
		assert.Error(t, err)
	})

	t.Run("should return error for empty tokens", func(t *testing.T) {
		// Given
		tmpDir := t.TempDir()
		tokensFile := filepath.Join(tmpDir, "empty.json")
		os.WriteFile(tokensFile, []byte(`{"tokens": []}`), 0644)

		// When
		_, err := readTokensFromFile(tokensFile)

		// Then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no tokens found")
	})
}

func TestGenerateWithBuiltInGenerator(t *testing.T) {
	tokens := []design.Token{
		{Name: "primary", Type: "color", Value: "#007bff"},
	}

	tests := []struct {
		name        string
		format      string
		expectedExt string
		shouldError bool
	}{
		{"tailwind format", "tailwind", ".js", false},
		{"tw alias", "tw", ".js", false},
		{"css format", "css", ".css", false},
		{"scss format", "scss", ".scss", false},
		{"sass alias", "sass", ".scss", false},
		{"typescript format", "typescript", ".ts", false},
		{"ts alias", "ts", ".ts", false},
		{"javascript format", "javascript", ".js", false},
		{"js alias", "js", ".js", false},
		{"json format", "json", ".json", false},
		{"unsupported format", "xml", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When
			output, ext, err := generateWithBuiltInGenerator(tokens, tt.format)

			// Then
			if tt.shouldError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, output)
				assert.Equal(t, tt.expectedExt, ext)
			}
		})
	}
}

func TestGenerateWithCustomTemplate(t *testing.T) {
	t.Run("should generate code with custom template", func(t *testing.T) {
		// Given
		tmpDir := t.TempDir()
		templateFile := filepath.Join(tmpDir, "template.tmpl")
		templateContent := `Colors:
{{- range .Tokens }}
{{ .Name }}: {{ .Value }}
{{- end }}`
		os.WriteFile(templateFile, []byte(templateContent), 0644)

		tokens := []design.Token{
			{Name: "primary", Type: "color", Value: "#007bff"},
		}

		// When
		output, err := generateWithCustomTemplate(tokens, templateFile)

		// Then
		require.NoError(t, err)
		assert.Contains(t, output, "Colors:")
		assert.Contains(t, output, "primary: #007bff")
	})

	t.Run("should return error for non-existent template", func(t *testing.T) {
		// When
		_, err := generateWithCustomTemplate([]design.Token{}, "/nonexistent/template.tmpl")

		// Then
		assert.Error(t, err)
	})

	t.Run("should return error for invalid template", func(t *testing.T) {
		// Given
		tmpDir := t.TempDir()
		templateFile := filepath.Join(tmpDir, "invalid.tmpl")
		os.WriteFile(templateFile, []byte("{{ .Invalid syntax"), 0644)

		// When
		_, err := generateWithCustomTemplate([]design.Token{}, templateFile)

		// Then
		assert.Error(t, err)
	})
}

func TestContains(t *testing.T) {
	t.Run("should return true if item exists", func(t *testing.T) {
		// Given
		slice := []string{"a", "b", "c"}

		// When/Then
		assert.True(t, contains(slice, "b"))
	})

	t.Run("should return false if item does not exist", func(t *testing.T) {
		// Given
		slice := []string{"a", "b", "c"}

		// When/Then
		assert.False(t, contains(slice, "d"))
	})

	t.Run("should handle empty slice", func(t *testing.T) {
		// Given
		slice := []string{}

		// When/Then
		assert.False(t, contains(slice, "a"))
	})
}

func TestDesignGenerateCommandIntegration(t *testing.T) {
	t.Run("should generate all formats successfully", func(t *testing.T) {
		// Given
		tmpDir := t.TempDir()
		tokensFile := filepath.Join(tmpDir, "tokens.json")

		tokens := []design.Token{
			{Name: "primary-color", Type: "color", Value: "#007bff"},
			{Name: "spacing-base", Type: "spacing", Value: "16px"},
			{Name: "font-family", Type: "typography", Value: "Helvetica, Arial"},
		}

		data, _ := json.Marshal(map[string]interface{}{
			"tokens": tokens,
		})
		os.WriteFile(tokensFile, data, 0644)

		formats := []string{"tailwind", "css", "scss", "typescript", "javascript", "json"}

		for _, format := range formats {
			// When
			outputFile := filepath.Join(tmpDir, "output")
			generateTokensFile = tokensFile
			generateFormat = format
			generateOutput = outputFile
			generatePretty = true

			err := runDesignGenerate(designGenerateCmd, []string{})

			// Then
			require.NoError(t, err, "Format: %s", format)

			// Verify output file exists
			files, _ := os.ReadDir(tmpDir)
			found := false
			for _, file := range files {
				if !file.IsDir() && file.Name() != "tokens.json" {
					found = true
					break
				}
			}
			assert.True(t, found, "Output file should exist for format: %s", format)
		}
	})
}
