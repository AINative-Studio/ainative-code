package session_test

import (
	"testing"

	"github.com/AINative-studio/ainative-code/internal/session"
	"github.com/stretchr/testify/assert"
)

// TestSearchOptions_WhitespaceQuery tests validation of whitespace-only queries
// This is a regression test for GitHub Issue #111: Empty session search query causes crash
func TestSearchOptions_WhitespaceQuery(t *testing.T) {
	testCases := []struct {
		name          string
		query         string
		shouldBeValid bool
	}{
		{
			name:          "empty string",
			query:         "",
			shouldBeValid: false,
		},
		{
			name:          "spaces only",
			query:         "   ",
			shouldBeValid: false,
		},
		{
			name:          "tab only",
			query:         "\t",
			shouldBeValid: false,
		},
		{
			name:          "newline only",
			query:         "\n",
			shouldBeValid: false,
		},
		{
			name:          "mixed whitespace",
			query:         "  \t\n  ",
			shouldBeValid: false,
		},
		{
			name:          "valid query with spaces",
			query:         "  test  ",
			shouldBeValid: true,
		},
		{
			name:          "valid query",
			query:         "test",
			shouldBeValid: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			opts := &session.SearchOptions{
				Query:  tc.query,
				Limit:  50,
				Offset: 0,
			}

			err := opts.Validate()

			if tc.shouldBeValid {
				assert.NoError(t, err, "query %q should be valid", tc.query)
			} else {
				assert.ErrorIs(t, err, session.ErrEmptySearchQuery, "query %q should be invalid", tc.query)
			}
		})
	}
}
