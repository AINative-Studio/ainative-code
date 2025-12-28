package session

// ListOptions contains options for listing sessions
type ListOptions struct {
	Status SessionStatus
	Limit  int64
	Offset int64
	SortBy string
}

// ListOption is a functional option for configuring ListOptions
type ListOption func(*ListOptions)

// WithStatus filters sessions by status
func WithStatus(status SessionStatus) ListOption {
	return func(opts *ListOptions) {
		opts.Status = status
	}
}

// WithLimit sets the maximum number of sessions to return
func WithLimit(limit int64) ListOption {
	return func(opts *ListOptions) {
		opts.Limit = limit
	}
}

// WithOffset sets the offset for pagination
func WithOffset(offset int64) ListOption {
	return func(opts *ListOptions) {
		opts.Offset = offset
	}
}

// WithSortBy sets the field to sort by
func WithSortBy(sortBy string) ListOption {
	return func(opts *ListOptions) {
		opts.SortBy = sortBy
	}
}

// DefaultListOptions returns default list options
func DefaultListOptions() *ListOptions {
	return &ListOptions{
		Status: StatusActive,
		Limit:  50,
		Offset: 0,
		SortBy: "updated_at",
	}
}

// ApplyListOptions applies functional options to ListOptions
func ApplyListOptions(opts ...ListOption) *ListOptions {
	options := DefaultListOptions()
	for _, opt := range opts {
		opt(options)
	}
	return options
}

// SearchOptions contains options for searching
type SearchOptions struct {
	Query  string
	Limit  int64
	Offset int64
}

// SearchOption is a functional option for configuring SearchOptions
type SearchOption func(*SearchOptions)

// WithQuery sets the search query
func WithQuery(query string) SearchOption {
	return func(opts *SearchOptions) {
		opts.Query = query
	}
}

// WithSearchLimit sets the maximum number of results
func WithSearchLimit(limit int64) SearchOption {
	return func(opts *SearchOptions) {
		opts.Limit = limit
	}
}

// WithSearchOffset sets the offset for search pagination
func WithSearchOffset(offset int64) SearchOption {
	return func(opts *SearchOptions) {
		opts.Offset = offset
	}
}

// DefaultSearchOptions returns default search options
func DefaultSearchOptions() *SearchOptions {
	return &SearchOptions{
		Query:  "",
		Limit:  50,
		Offset: 0,
	}
}

// ApplySearchOptions applies functional options to SearchOptions
func ApplySearchOptions(opts ...SearchOption) *SearchOptions {
	options := DefaultSearchOptions()
	for _, opt := range opts {
		opt(options)
	}
	return options
}
