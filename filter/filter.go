// Package filter provides database-agnostic filtering utilities for query builders.
//
// This package works with any ORM or query builder (Ent, GORM, sqlc, etc.) by accepting
// functions that know how to build predicates and apply WHERE clauses for your specific query type.
package filter

import "strings"

// Config contains the query-building functions needed for filtering.
// The Predicate type parameter represents whatever your ORM uses for WHERE conditions.
type Config[Q any, P any] struct {
	// Where applies a WHERE clause to the query
	Where func(Q, P) Q
}

// Combinators provides functions for combining predicates with AND/OR logic.
// These are used by BuildFilterMap to construct accurate "always true" and "always false" predicates.
//
// Example for Ent:
//
//	combinators := filter.Combinators[predicate.User]{
//	    Or:  user.Or,
//	    And: user.And,
//	}
type Combinators[P any] struct {
	// Or combines predicates with OR logic
	Or func(predicates ...P) P

	// And combines predicates with AND logic
	And func(predicates ...P) P
}

// PredicateBuilder builds predicates for combining and ID filtering.
// Users implement this to adapt their ORM's predicate system.
//
// Example for Ent:
//
//	predicates := filter.PredicateBuilder[predicate.User]{
//	    IDIn: user.IDIn,
//	    Or:   user.Or,
//	    And:  user.And,
//	}
type PredicateBuilder[P any] struct {
	// IDIn creates a predicate that matches any of the given IDs
	IDIn func(ids ...any) P

	// Or combines predicates with OR logic
	Or func(predicates ...P) P

	// And combines predicates with AND logic
	And func(predicates ...P) P
}

// SearchFields maps database field names to functions that create case-insensitive
// LIKE/ILIKE predicates for those fields.
//
// Example for Ent:
//
//	fields := filter.SearchFields[predicate.User]{
//	    user.FieldEmail:     user.EmailContainsFold,
//	    user.FieldFirstName: user.FirstNameContainsFold,
//	    user.FieldLastName:  user.LastNameContainsFold,
//	}
type SearchFields[P any] map[string]func(string) P

// ApplyIDs filters a query to only include records with the given IDs.
// This supports the Refine.js "getMany" pattern where multiple IDs are passed.
//
// Parameters:
//   - query: The query to filter
//   - cfg: Configuration containing the where function
//   - builder: PredicateBuilder for creating ID predicates
//   - ids: List of IDs to match (any type - uuid.UUID, int, string, etc.)
//
// Returns the modified query. If ids is empty, returns query unchanged.
//
// Example:
//
//	query = filter.ApplyIDs(query, cfg, builder, []any{id1, id2, id3})
func ApplyIDs[Q any, P any](
	query Q,
	cfg Config[Q, P],
	builder PredicateBuilder[P],
	ids []any,
) Q {
	if len(ids) == 0 {
		return query
	}

	predicate := builder.IDIn(ids...)
	return cfg.Where(query, predicate)
}

// ApplySearch applies full-text search filtering to a query.
//
// Search behavior:
//   - Tokenizes the search string on whitespace
//   - Each token must match at least one of the searchable fields (OR logic within token)
//   - All tokens must match (AND logic between tokens)
//   - Case-insensitive matching
//
// Example: searching "john doe" across first_name and last_name fields will match
// records where (first_name ILIKE '%john%' OR last_name ILIKE '%john%') AND
// (first_name ILIKE '%doe%' OR last_name ILIKE '%doe%')
//
// Parameters:
//   - query: The query to filter
//   - cfg: Configuration containing the where function
//   - fields: Map of field names to predicate builder functions
//   - builder: PredicateBuilder for combining predicates
//   - search: The search string from the user
//
// Returns the modified query. If search is empty, returns query unchanged.
//
// Example:
//
//	searchFields := filter.SearchFields[predicate.User]{
//	    user.FieldEmail:     user.EmailContainsFold,
//	    user.FieldFirstName: user.FirstNameContainsFold,
//	}
//	query = filter.ApplySearch(query, cfg, searchFields, builder, "john doe")
func ApplySearch[Q any, P any](
	query Q,
	cfg Config[Q, P],
	fields SearchFields[P],
	builder PredicateBuilder[P],
	search string,
) Q {
	if search == "" {
		return query
	}

	// Tokenize search string on whitespace
	tokens := strings.Fields(search)
	if len(tokens) == 0 {
		return query
	}

	// For each token, create OR predicates across all searchable fields
	tokenPredicates := make([]P, 0, len(tokens))
	for _, token := range tokens {
		fieldPredicates := make([]P, 0, len(fields))
		for _, containsFoldFunc := range fields {
			fieldPredicates = append(fieldPredicates, containsFoldFunc(token))
		}
		// Combine field predicates with OR for this token
		tokenPredicates = append(tokenPredicates, builder.Or(fieldPredicates...))
	}

	// Combine all token predicates with AND
	var combinedPredicate P
	if len(tokenPredicates) == 1 {
		combinedPredicate = tokenPredicates[0]
	} else {
		combinedPredicate = builder.And(tokenPredicates...)
	}

	return cfg.Where(query, combinedPredicate)
}

// ApplyWhere applies a single predicate to a query.
// This is a simple wrapper for adding custom WHERE clauses.
//
// Example:
//
//	isActivePredicate := user.IsActive(true)
//	query = filter.ApplyWhere(query, cfg, isActivePredicate)
func ApplyWhere[Q any, P any](
	query Q,
	cfg Config[Q, P],
	predicate P,
) Q {
	return cfg.Where(query, predicate)
}
