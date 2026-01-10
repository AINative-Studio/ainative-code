package session

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSanitizeFTS5Query_Whitespace tests the sanitization function with whitespace
func TestSanitizeFTS5Query_Whitespace(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "spaces only",
			input:    "   ",
			expected: `"   "`,
		},
		{
			name:     "tab only",
			input:    "\t",
			expected: "\"\t\"",
		},
		{
			name:     "newline only",
			input:    "\n",
			expected: "\"\n\"",
		},
		{
			name:     "mixed whitespace",
			input:    "  \t\n  ",
			expected: "\"  \t\n  \"",
		},
		{
			name:     "valid query",
			input:    "test query",
			expected: `"test query"`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := sanitizeFTS5Query(tc.input)
			assert.Equal(t, tc.expected, result)

			// Log what would be sent to FTS5
			t.Logf("Input: %q -> Sanitized: %q", tc.input, result)

			// Check if the sanitized query would be problematic
			if len(tc.input) > 0 && len(result) == 2 {
				t.Errorf("Whitespace-only input %q resulted in empty FTS5 query", tc.input)
			}
		})
	}
}
