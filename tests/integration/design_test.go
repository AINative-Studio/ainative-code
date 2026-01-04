package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/AINative-studio/ainative-code/tests/integration/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDesignTokenExtraction_CSS tests CSS parsing
func TestDesignTokenExtraction_CSS(t *testing.T) {
	t.Run("should parse CSS file and extract color tokens", func(t *testing.T) {
		// Given: Mock design server
		server := mocks.NewDesignServer()
		defer server.Close()

		// And: CSS content with colors
		cssContent := `
:root {
  --primary-color: #007bff;
  --secondary-color: #6c757d;
  --success-color: #28a745;
}
`
		parseRequest := map[string]interface{}{
			"content":   cssContent,
			"file_type": "css",
		}

		// When: Parsing CSS
		body, _ := json.Marshal(parseRequest)
		req, _ := http.NewRequest("POST", server.GetURL()+"/api/tokens/parse", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)

		// Then: Should extract tokens
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, server.ParseCalled)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Contains(t, result, "tokens")
		assert.Greater(t, int(result["count"].(float64)), 0)
	})

	t.Run("should parse SCSS file and extract typography tokens", func(t *testing.T) {
		// Given: Mock design server
		server := mocks.NewDesignServer()
		defer server.Close()

		// And: SCSS content with typography
		scssContent := `
$font-family-base: 'Helvetica Neue', sans-serif;
$font-size-base: 16px;
$font-weight-normal: 400;
$font-weight-bold: 700;
`
		parseRequest := map[string]interface{}{
			"content":   scssContent,
			"file_type": "scss",
		}

		// When: Parsing SCSS
		body, _ := json.Marshal(parseRequest)
		req, _ := http.NewRequest("POST", server.GetURL()+"/api/tokens/parse", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)

		// Then: Should extract typography tokens
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("should extract specific token types", func(t *testing.T) {
		// Given: Mock design server
		server := mocks.NewDesignServer()
		defer server.Close()

		// And: Mixed content
		content := `
:root {
  --primary-color: #007bff;
  --font-family: system-ui;
  --spacing-base: 16px;
}
`
		extractRequest := map[string]interface{}{
			"content":     content,
			"token_types": []string{"color"},
		}

		// When: Extracting only color tokens
		body, _ := json.Marshal(extractRequest)
		req, _ := http.NewRequest("POST", server.GetURL()+"/api/tokens/extract", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)

		// Then: Should return only color tokens
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, server.ExtractCalled)
	})

	t.Run("should reject empty content", func(t *testing.T) {
		// Given: Mock design server
		server := mocks.NewDesignServer()
		defer server.Close()

		// And: Empty content
		parseRequest := map[string]interface{}{
			"content":   "",
			"file_type": "css",
		}

		// When: Parsing empty content
		body, _ := json.Marshal(parseRequest)
		req, _ := http.NewRequest("POST", server.GetURL()+"/api/tokens/parse", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)

		// Then: Should reject
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

// TestDesignTokenExtraction_Formats tests token output formats
func TestDesignTokenExtraction_Formats(t *testing.T) {
	t.Run("should export tokens as JSON", func(t *testing.T) {
		// Given: Mock design server with tokens
		server := mocks.NewDesignServer()
		defer server.Close()

		server.AddToken(mocks.DesignToken{
			Name:     "primary-color",
			Value:    "#007bff",
			Type:     "color",
			Category: "colors",
		})

		// When: Exporting as JSON
		req, _ := http.NewRequest("GET", server.GetURL()+"/api/tokens/export?format=json", nil)

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)

		// Then: Should return JSON
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Contains(t, result, "tokens")
	})

	t.Run("should export tokens as CSS", func(t *testing.T) {
		// Given: Mock design server with tokens
		server := mocks.NewDesignServer()
		defer server.Close()

		server.AddToken(mocks.DesignToken{
			Name:  "primary-color",
			Value: "#007bff",
			Type:  "color",
		})

		// When: Exporting as CSS
		req, _ := http.NewRequest("GET", server.GetURL()+"/api/tokens/export?format=css", nil)

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)

		// Then: Should return CSS
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "text/css", resp.Header.Get("Content-Type"))

		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		cssOutput := buf.String()

		assert.Contains(t, cssOutput, ":root")
		assert.Contains(t, cssOutput, "--primary-color")
	})

	t.Run("should export tokens as SCSS", func(t *testing.T) {
		// Given: Mock design server with tokens
		server := mocks.NewDesignServer()
		defer server.Close()

		server.AddToken(mocks.DesignToken{
			Name:  "primary-color",
			Value: "#007bff",
			Type:  "color",
		})

		// When: Exporting as SCSS
		req, _ := http.NewRequest("GET", server.GetURL()+"/api/tokens/export?format=scss", nil)

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)

		// Then: Should return SCSS
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		scssOutput := buf.String()

		assert.Contains(t, scssOutput, "$primary-color")
	})

	t.Run("should reject unsupported format", func(t *testing.T) {
		// Given: Mock design server
		server := mocks.NewDesignServer()
		defer server.Close()

		// When: Requesting unsupported format
		req, _ := http.NewRequest("GET", server.GetURL()+"/api/tokens/export?format=xml", nil)

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)

		// Then: Should reject
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

// TestDesignTokenValidation tests token structure validation
func TestDesignTokenValidation(t *testing.T) {
	t.Run("should validate correct token structure", func(t *testing.T) {
		// Given: Mock design server
		server := mocks.NewDesignServer()
		defer server.Close()

		// And: Valid tokens
		tokens := []map[string]interface{}{
			{
				"name":     "primary-color",
				"value":    "#007bff",
				"type":     "color",
				"category": "colors",
			},
		}

		validateRequest := map[string]interface{}{
			"tokens": tokens,
		}

		// When: Validating tokens
		body, _ := json.Marshal(validateRequest)
		req, _ := http.NewRequest("POST", server.GetURL()+"/api/tokens/validate", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)

		// Then: Should be valid
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, server.ValidateCalled)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.True(t, result["valid"].(bool))
	})

	t.Run("should reject invalid token structure", func(t *testing.T) {
		// Given: Mock design server
		server := mocks.NewDesignServer()
		defer server.Close()

		// And: Invalid tokens (missing required fields)
		tokens := []map[string]interface{}{
			{
				"name": "incomplete-token",
				// Missing value and type
			},
		}

		validateRequest := map[string]interface{}{
			"tokens": tokens,
		}

		// When: Validating tokens
		body, _ := json.Marshal(validateRequest)
		req, _ := http.NewRequest("POST", server.GetURL()+"/api/tokens/validate", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)

		// Then: Should be invalid
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.False(t, result["valid"].(bool))
		assert.NotEmpty(t, result["errors"])
	})
}

// TestDesignTokenErrorHandling tests error scenarios
func TestDesignTokenErrorHandling(t *testing.T) {
	t.Run("should handle 401 unauthorized", func(t *testing.T) {
		// Given: Mock server with auth failure
		server := mocks.NewDesignServer()
		defer server.Close()

		server.ShouldFailAuth = true

		// When: Making request
		req, _ := http.NewRequest("GET", server.GetURL()+"/api/tokens/export", nil)

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)

		// Then: Should return 401
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("should handle 429 rate limit", func(t *testing.T) {
		// Given: Mock server with rate limiting
		server := mocks.NewDesignServer()
		defer server.Close()

		server.ShouldRateLimit = true

		// When: Making request
		req, _ := http.NewRequest("GET", server.GetURL()+"/api/tokens/export", nil)

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)

		// Then: Should return 429
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
	})

	t.Run("should handle invalid JSON", func(t *testing.T) {
		// Given: Mock design server
		server := mocks.NewDesignServer()
		defer server.Close()

		// When: Sending invalid JSON
		req, _ := http.NewRequest("POST", server.GetURL()+"/api/tokens/parse", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)

		// Then: Should return 400
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}
