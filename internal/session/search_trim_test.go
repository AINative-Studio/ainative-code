package session_test

import (
	"testing"

	"github.com/AINative-studio/ainative-code/internal/session"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSearchOptions_QueryTrimming tests that validation trims whitespace from queries
func TestSearchOptions_QueryTrimming(t *testing.T) {
	testCases := []struct {
		name          string
		inputQuery    string
		expectedQuery string
	}{
		{
			name:          "leading spaces",
			inputQuery:    "   test query",
			expectedQuery: "test query",
		},
		{
			name:          "trailing spaces",
			inputQuery:    "test query   ",
			expectedQuery: "test query",
		},
		{
			name:          "leading and trailing spaces",
			inputQuery:    "   test query   ",
			expectedQuery: "test query",
		},
		{
			name:          "tabs and spaces",
			inputQuery:    "\t  test query  \t",
			expectedQuery: "test query",
		},
		{
			name:          "newlines and spaces",
			inputQuery:    "\n  test query  \n",
			expectedQuery: "test query",
		},
		{
			name:          "no trimming needed",
			inputQuery:    "test query",
			expectedQuery: "test query",
		},
		{
			name:          "internal spaces preserved",
			inputQuery:    "  test   query   with   spaces  ",
			expectedQuery: "test   query   with   spaces",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			opts := &session.SearchOptions{
				Query:  tc.inputQuery,
				Limit:  50,
				Offset: 0,
			}

			err := opts.Validate()
			require.NoError(t, err, "validation should succeed for non-empty queries")

			// Verify the query was trimmed
			assert.Equal(t, tc.expectedQuery, opts.Query, "query should be trimmed after validation")
		})
	}
}
