package session

import (
	"context"
	"database/sql"
	"fmt"
)

// SearchAllMessages performs full-text search across all conversation messages
func (m *SQLiteManager) SearchAllMessages(ctx context.Context, opts *SearchOptions) (*SearchResultSet, error) {
	if opts == nil {
		return nil, NewSessionError("SearchMessages", ErrEmptySearchQuery, "options are nil")
	}

	// Validate options
	if err := opts.Validate(); err != nil {
		return nil, NewSessionError("SearchMessages", err, "invalid search options")
	}

	// Build and execute the appropriate query based on filters
	var results []SearchResult
	var totalCount int64
	var err error

	// Determine which query to use based on filters
	hasDateFilter := opts.DateFrom != nil && opts.DateTo != nil
	hasProviderFilter := opts.Provider != ""

	if hasDateFilter && hasProviderFilter {
		results, totalCount, err = m.searchWithAllFilters(ctx, opts)
	} else if hasDateFilter {
		results, totalCount, err = m.searchWithDateRange(ctx, opts)
	} else if hasProviderFilter {
		results, totalCount, err = m.searchWithProvider(ctx, opts)
	} else {
		results, totalCount, err = m.searchBasic(ctx, opts)
	}

	if err != nil {
		return nil, NewSessionError("SearchMessages", err, "search query failed")
	}

	return &SearchResultSet{
		Results:    results,
		TotalCount: totalCount,
		Query:      opts.Query,
		Limit:      opts.Limit,
		Offset:     opts.Offset,
	}, nil
}

// searchBasic performs basic full-text search without filters
func (m *SQLiteManager) searchBasic(ctx context.Context, opts *SearchOptions) ([]SearchResult, int64, error) {
	query := `
		SELECT
			m.id, m.session_id, m.role, m.content, m.timestamp, m.parent_id,
			m.tokens_used, m.model, m.finish_reason, m.metadata,
			s.name as session_name, s.status as session_status,
			snippet(messages_fts, 3, '<mark>', '</mark>', '...', 32) as snippet,
			bm25(messages_fts) as relevance_score
		FROM messages_fts fts
		JOIN messages m ON fts.message_id = m.id
		JOIN sessions s ON m.session_id = s.id
		WHERE messages_fts MATCH ?
		ORDER BY relevance_score
		LIMIT ? OFFSET ?
	`

	rows, err := m.db.DB().QueryContext(ctx, query, opts.Query, opts.Limit, opts.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute search query: %w", err)
	}
	defer rows.Close()

	results, err := m.scanSearchResults(rows)
	if err != nil {
		return nil, 0, err
	}

	// Get total count
	totalCount, err := m.searchCount(ctx, opts.Query)
	if err != nil {
		return nil, 0, err
	}

	return results, totalCount, nil
}

// searchWithDateRange performs search with date range filter
func (m *SQLiteManager) searchWithDateRange(ctx context.Context, opts *SearchOptions) ([]SearchResult, int64, error) {
	query := `
		SELECT
			m.id, m.session_id, m.role, m.content, m.timestamp, m.parent_id,
			m.tokens_used, m.model, m.finish_reason, m.metadata,
			s.name as session_name, s.status as session_status,
			snippet(messages_fts, 3, '<mark>', '</mark>', '...', 32) as snippet,
			bm25(messages_fts) as relevance_score
		FROM messages_fts fts
		JOIN messages m ON fts.message_id = m.id
		JOIN sessions s ON m.session_id = s.id
		WHERE messages_fts MATCH ?
			AND m.timestamp >= ?
			AND m.timestamp <= ?
		ORDER BY relevance_score
		LIMIT ? OFFSET ?
	`

	dateFrom := formatTimestamp(*opts.DateFrom)
	dateTo := formatTimestamp(*opts.DateTo)

	rows, err := m.db.DB().QueryContext(ctx, query, opts.Query, dateFrom, dateTo, opts.Limit, opts.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute search query with date range: %w", err)
	}
	defer rows.Close()

	results, err := m.scanSearchResults(rows)
	if err != nil {
		return nil, 0, err
	}

	// Get total count
	totalCount, err := m.searchCountWithDateRange(ctx, opts.Query, dateFrom, dateTo)
	if err != nil {
		return nil, 0, err
	}

	return results, totalCount, nil
}

// searchWithProvider performs search with provider filter
func (m *SQLiteManager) searchWithProvider(ctx context.Context, opts *SearchOptions) ([]SearchResult, int64, error) {
	query := `
		SELECT
			m.id, m.session_id, m.role, m.content, m.timestamp, m.parent_id,
			m.tokens_used, m.model, m.finish_reason, m.metadata,
			s.name as session_name, s.status as session_status,
			snippet(messages_fts, 3, '<mark>', '</mark>', '...', 32) as snippet,
			bm25(messages_fts) as relevance_score
		FROM messages_fts fts
		JOIN messages m ON fts.message_id = m.id
		JOIN sessions s ON m.session_id = s.id
		WHERE messages_fts MATCH ?
			AND m.model LIKE ?
		ORDER BY relevance_score
		LIMIT ? OFFSET ?
	`

	providerPattern := "%" + opts.Provider + "%"

	rows, err := m.db.DB().QueryContext(ctx, query, opts.Query, providerPattern, opts.Limit, opts.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute search query with provider: %w", err)
	}
	defer rows.Close()

	results, err := m.scanSearchResults(rows)
	if err != nil {
		return nil, 0, err
	}

	// Get total count
	totalCount, err := m.searchCountWithProvider(ctx, opts.Query, providerPattern)
	if err != nil {
		return nil, 0, err
	}

	return results, totalCount, nil
}

// searchWithAllFilters performs search with all filters
func (m *SQLiteManager) searchWithAllFilters(ctx context.Context, opts *SearchOptions) ([]SearchResult, int64, error) {
	query := `
		SELECT
			m.id, m.session_id, m.role, m.content, m.timestamp, m.parent_id,
			m.tokens_used, m.model, m.finish_reason, m.metadata,
			s.name as session_name, s.status as session_status,
			snippet(messages_fts, 3, '<mark>', '</mark>', '...', 32) as snippet,
			bm25(messages_fts) as relevance_score
		FROM messages_fts fts
		JOIN messages m ON fts.message_id = m.id
		JOIN sessions s ON m.session_id = s.id
		WHERE messages_fts MATCH ?
			AND m.timestamp >= ?
			AND m.timestamp <= ?
			AND m.model LIKE ?
		ORDER BY relevance_score
		LIMIT ? OFFSET ?
	`

	dateFrom := formatTimestamp(*opts.DateFrom)
	dateTo := formatTimestamp(*opts.DateTo)
	providerPattern := "%" + opts.Provider + "%"

	rows, err := m.db.DB().QueryContext(ctx, query, opts.Query, dateFrom, dateTo, providerPattern, opts.Limit, opts.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute search query with all filters: %w", err)
	}
	defer rows.Close()

	results, err := m.scanSearchResults(rows)
	if err != nil {
		return nil, 0, err
	}

	// Get total count
	totalCount, err := m.searchCountWithAllFilters(ctx, opts.Query, dateFrom, dateTo, providerPattern)
	if err != nil {
		return nil, 0, err
	}

	return results, totalCount, nil
}

// scanSearchResults scans database rows into SearchResult structs
func (m *SQLiteManager) scanSearchResults(rows *sql.Rows) ([]SearchResult, error) {
	var results []SearchResult

	for rows.Next() {
		var (
			id, sessionID, role, content, timestamp string
			parentID, model, finishReason, metadata *string
			tokensUsed                               *int64
			sessionName, sessionStatus               string
			snippet                                  string
			relevanceScore                           float64
		)

		err := rows.Scan(
			&id, &sessionID, &role, &content, &timestamp, &parentID,
			&tokensUsed, &model, &finishReason, &metadata,
			&sessionName, &sessionStatus,
			&snippet, &relevanceScore,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan search result: %w", err)
		}

		// Parse timestamp
		ts, err := parseTimestamp(timestamp)
		if err != nil {
			return nil, fmt.Errorf("failed to parse timestamp: %w", err)
		}

		// Unmarshal metadata
		var metadataMap map[string]any
		if metadata != nil && *metadata != "" {
			metadataMap, err = UnmarshalMetadata(*metadata)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		result := SearchResult{
			Message: Message{
				ID:           id,
				SessionID:    sessionID,
				Role:         MessageRole(role),
				Content:      content,
				Timestamp:    ts,
				ParentID:     parentID,
				TokensUsed:   tokensUsed,
				Model:        model,
				FinishReason: finishReason,
				Metadata:     metadataMap,
			},
			SessionName:    sessionName,
			SessionStatus:  SessionStatus(sessionStatus),
			Snippet:        snippet,
			RelevanceScore: relevanceScore,
		}

		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating search results: %w", err)
	}

	return results, nil
}

// searchCount gets the total count of search results
func (m *SQLiteManager) searchCount(ctx context.Context, query string) (int64, error) {
	countQuery := `
		SELECT COUNT(*)
		FROM messages_fts fts
		JOIN messages m ON fts.message_id = m.id
		WHERE messages_fts MATCH ?
	`

	var count int64
	err := m.db.DB().QueryRowContext(ctx, countQuery, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get search count: %w", err)
	}

	return count, nil
}

// searchCountWithDateRange gets the count with date range filter
func (m *SQLiteManager) searchCountWithDateRange(ctx context.Context, query, dateFrom, dateTo string) (int64, error) {
	countQuery := `
		SELECT COUNT(*)
		FROM messages_fts fts
		JOIN messages m ON fts.message_id = m.id
		WHERE messages_fts MATCH ?
			AND m.timestamp >= ?
			AND m.timestamp <= ?
	`

	var count int64
	err := m.db.DB().QueryRowContext(ctx, countQuery, query, dateFrom, dateTo).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get search count with date range: %w", err)
	}

	return count, nil
}

// searchCountWithProvider gets the count with provider filter
func (m *SQLiteManager) searchCountWithProvider(ctx context.Context, query, providerPattern string) (int64, error) {
	countQuery := `
		SELECT COUNT(*)
		FROM messages_fts fts
		JOIN messages m ON fts.message_id = m.id
		WHERE messages_fts MATCH ?
			AND m.model LIKE ?
	`

	var count int64
	err := m.db.DB().QueryRowContext(ctx, countQuery, query, providerPattern).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get search count with provider: %w", err)
	}

	return count, nil
}

// searchCountWithAllFilters gets the count with all filters
func (m *SQLiteManager) searchCountWithAllFilters(ctx context.Context, query, dateFrom, dateTo, providerPattern string) (int64, error) {
	countQuery := `
		SELECT COUNT(*)
		FROM messages_fts fts
		JOIN messages m ON fts.message_id = m.id
		WHERE messages_fts MATCH ?
			AND m.timestamp >= ?
			AND m.timestamp <= ?
			AND m.model LIKE ?
	`

	var count int64
	err := m.db.DB().QueryRowContext(ctx, countQuery, query, dateFrom, dateTo, providerPattern).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get search count with all filters: %w", err)
	}

	return count, nil
}

// RebuildSearchIndex rebuilds the FTS5 search index
func (m *SQLiteManager) RebuildSearchIndex(ctx context.Context) error {
	query := `INSERT INTO messages_fts(messages_fts) VALUES('rebuild')`

	_, err := m.db.DB().ExecContext(ctx, query)
	if err != nil {
		return NewSessionError("RebuildSearchIndex", err, "failed to rebuild FTS index")
	}

	return nil
}

// OptimizeSearchIndex optimizes the FTS5 search index
func (m *SQLiteManager) OptimizeSearchIndex(ctx context.Context) error {
	query := `INSERT INTO messages_fts(messages_fts) VALUES('optimize')`

	_, err := m.db.DB().ExecContext(ctx, query)
	if err != nil {
		return NewSessionError("OptimizeSearchIndex", err, "failed to optimize FTS index")
	}

	return nil
}
