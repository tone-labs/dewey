// Package sort provides database-agnostic sorting utilities for query builders.
//
// This package works with any ORM or query builder (Ent, GORM, sqlc, etc.) by accepting
// functions that know how to apply ordering to your specific query type.
package sort

// Fields maps JSON field names (from API requests) to database field names.
// This allows you to expose clean API field names while using actual DB columns for sorting.
//
// Example:
//
//	fields := sort.Fields{
//	    "email":      "email",           // JSON name -> DB column
//	    "created_at": "created_at",
//	    "full_name":  "CONCAT(first_name, ' ', last_name)", // Computed fields work too
//	}
type Fields map[string]string

// OrderBuilder creates ORDER BY clauses for a specific ORM.
// Users implement this to adapt their ORM's ordering mechanism.
//
// Example for Ent:
//
//	type EntOrderBuilder struct{}
//
//	func (EntOrderBuilder) Asc(field string) any {
//	    return func(s *sql.Selector) {
//	        s.OrderBy(sql.Asc(field))
//	    }
//	}
//
//	func (EntOrderBuilder) Desc(field string) any {
//	    return func(s *sql.Selector) {
//	        s.OrderBy(sql.Desc(field))
//	    }
//	}
type OrderBuilder interface {
	// Asc creates an ascending order option for the given database field
	Asc(dbField string) any

	// Desc creates a descending order option for the given database field
	Desc(dbField string) any
}

// Config contains the query-building functions needed for sorting.
type Config[Q any] struct {
	// Order applies ORDER BY clauses to the query.
	// The variadic parameter accepts whatever order options your ORM uses.
	Order func(Q, ...any) Q
}

// Order represents sort order
type Order string

const (
	Asc  Order = "asc"
	Desc Order = "desc"
)

// Apply applies sorting to a query based on a JSON field name and direction.
//
// Parameters:
//   - query: The query to sort
//   - cfg: Configuration containing the order function for your ORM
//   - fields: Mapping of JSON field names to database field names
//   - builder: OrderBuilder that creates order options for your ORM
//   - sortBy: The JSON field name to sort by (from API request)
//   - sortDir: The sort direction ("asc" or "desc")
//
// Returns the modified query with sorting applied. If sortBy is empty or not found
// in the fields map, returns the query unchanged.
//
// Example:
//
//	fields := sort.Fields{
//	    "email": user.FieldEmail,
//	    "created_at": user.FieldCreatedAt,
//	}
//	query = sort.Apply(query, cfg, fields, builder, "email", "desc")
func Apply[Q any](
	query Q,
	cfg Config[Q],
	fields Fields,
	builder OrderBuilder,
	sortBy string,
	sortDir string,
) Q {
	if sortBy == "" {
		return query
	}

	// Look up the database field that corresponds to the JSON field
	dbField, ok := fields[sortBy]
	if !ok {
		// Unknown field, ignore sorting
		return query
	}

	// Create the order option based on direction
	var orderOpt any
	if sortDir == string(Desc) {
		orderOpt = builder.Desc(dbField)
	} else {
		orderOpt = builder.Asc(dbField)
	}

	// Apply the order option to the query
	return cfg.Order(query, orderOpt)
}

// ApplyMultiple applies multiple sort criteria to a query.
// Sorts are applied in the order given (first sort is primary, second is tie-breaker, etc.).
//
// Parameters:
//   - query: The query to sort
//   - cfg: Configuration containing the order function for your ORM
//   - fields: Mapping of JSON field names to database field names
//   - builder: OrderBuilder that creates order options for your ORM
//   - sorts: List of sort criteria (field name and direction pairs)
//
// Returns the modified query with all sorts applied.
//
// Example:
//
//	sorts := []sort.Criteria{
//	    {Field: "last_name", Order: sort.Asc},
//	    {Field: "first_name", Order: sort.Asc},
//	}
//	query = sort.ApplyMultiple(query, cfg, fields, builder, sorts)
func ApplyMultiple[Q any](
	query Q,
	cfg Config[Q],
	fields Fields,
	builder OrderBuilder,
	sorts []Criteria,
) Q {
	if len(sorts) == 0 {
		return query
	}

	orderOpts := make([]any, 0, len(sorts))

	for _, s := range sorts {
		dbField, ok := fields[s.Field]
		if !ok {
			// Skip unknown fields
			continue
		}

		var orderOpt any
		if s.Order == Desc {
			orderOpt = builder.Desc(dbField)
		} else {
			orderOpt = builder.Asc(dbField)
		}

		orderOpts = append(orderOpts, orderOpt)
	}

	if len(orderOpts) == 0 {
		return query
	}

	return cfg.Order(query, orderOpts...)
}

// Criteria represents a single sort criterion
type Criteria struct {
	Field string `json:"field"` // JSON field name
	Order Order  `json:"order"` // Sort order (asc or desc)
}
