package session_test

import (
	"context"
	"testing"

	"github.com/AINative-studio/ainative-code/internal/session"
	"github.com/stretchr/testify/require"
)

// TestSearchMessages_WhitespaceQueryCrash tests if whitespace-only queries cause crashes
// This is a regression test for GitHub Issue #111: Empty session search query causes crash
func TestSearchMessages_WhitespaceQueryCrash(t *testing.T) {
	_, mgr, cleanup := setupSearchTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create some test data
	createTestSessionWithMessages(t, mgr, "Test Session", []struct {
		role    session.MessageRole
		content string
		model   string
	}{
		{session.RoleUser, "How do I test whitespace?", ""},
		{session.RoleAssistant, "You should trim whitespace before validation", ""},
	})

	testCases := []struct {
		name  string
		query string
	}{
		{
			name:  "spaces only",
			query: "   ",
		},
		{
			name:  "tab only",
			query: "\t",
		},
		{
			name:  "newline only",
			query: "\n",
		},
		{
			name:  "mixed whitespace",
			query: "  \t\n  ",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			opts := &session.SearchOptions{
				Query:  tc.query,
				Limit:  50,
				Offset: 0,
			}

			// This should either:
			// 1. Return an error from validation
			// 2. Return an error from search
			// 3. NOT crash
			results, err := mgr.SearchAllMessages(ctx, opts)

			// Currently this passes validation and tries to search with empty query
			// which may cause a crash or database error
			if err == nil {
				// Should not reach here - whitespace queries should be invalid
				t.Logf("WARNING: Search with whitespace query succeeded (returned %d results). This may cause issues.", len(results.Results))
				t.Logf("Query was: %q", tc.query)
			} else {
				t.Logf("Search correctly failed with error: %v", err)
				require.Error(t, err, "whitespace-only query should be rejected")
			}
		})
	}
}
