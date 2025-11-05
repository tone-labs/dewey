// Package pagination provides database-agnostic pagination utilities for query builders.
//
// This package works with any ORM or query builder (Ent, GORM, sqlc, etc.) by accepting
// functions that know how to apply limits and offsets to your specific query type.
package pagination

// Config contains the query-building functions needed for pagination.
// Users provide these functions to adapt their specific ORM to this library.
//
// Example for Ent:
//
//	cfg := pagination.Config[*ent.UserQuery]{
//	    Limit: func(q *ent.UserQuery, n int) *ent.UserQuery {
//	        return q.Limit(n)
//	    },
//	    Offset: func(q *ent.UserQuery, n int) *ent.UserQuery {
//	        return q.Offset(n)
//	    },
//	}
type Config[Q any] struct {
	// Limit applies a LIMIT clause to the query
	Limit func(Q, int) Q

	// Offset applies an OFFSET clause to the query
	Offset func(Q, int) Q
}

// Apply applies limit and offset pagination to a query.
//
// Parameters:
//   - query: The query to paginate
//   - cfg: Configuration containing limit/offset functions for your ORM
//   - limit: Maximum number of records to return (0 or negative means no limit)
//   - offset: Number of records to skip (0 or negative means no offset)
//
// Returns the modified query with pagination applied.
//
// Example:
//
//	query := client.User.Query()
//	query = pagination.Apply(query, cfg, 25, 0)  // First page, 25 items
//	query = pagination.Apply(query, cfg, 25, 25) // Second page, 25 items
func Apply[Q any](query Q, cfg Config[Q], limit, offset int) Q {
	if offset > 0 {
		query = cfg.Offset(query, offset)
	}

	if limit > 0 {
		query = cfg.Limit(query, limit)
	}

	return query
}

// Page represents a page of results with metadata.
type Page[T any] struct {
	// Data contains the records for this page
	Data []T `json:"data"`

	// Total is the total number of records across all pages
	Total int `json:"total"`

	// Limit is the maximum number of records per page
	Limit int `json:"limit"`

	// Offset is the number of records skipped
	Offset int `json:"offset"`
}

// NewPage creates a new Page with the given data and pagination metadata.
func NewPage[T any](data []T, total, limit, offset int) Page[T] {
	return Page[T]{
		Data:   data,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}
}

// HasNextPage returns true if there are more pages after this one.
func (p Page[T]) HasNextPage() bool {
	return p.Offset+p.Limit < p.Total
}

// HasPrevPage returns true if there are pages before this one.
func (p Page[T]) HasPrevPage() bool {
	return p.Offset > 0
}

// PageNumber returns the current page number (1-indexed).
func (p Page[T]) PageNumber() int {
	if p.Limit <= 0 {
		return 1
	}
	return (p.Offset / p.Limit) + 1
}

// TotalPages returns the total number of pages.
func (p Page[T]) TotalPages() int {
	if p.Limit <= 0 {
		return 1
	}
	pages := p.Total / p.Limit
	if p.Total%p.Limit > 0 {
		pages++
	}
	return pages
}
