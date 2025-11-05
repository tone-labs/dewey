package filter

// Operator represents a filter comparison operator
type Operator string

const (
	// Equality operators
	OpEq Operator = "eq" // Equal
	OpNe Operator = "ne" // Not equal

	// Comparison operators
	OpGt  Operator = "gt"  // Greater than
	OpGte Operator = "gte" // Greater than or equal to
	OpLt  Operator = "lt"  // Less than
	OpLte Operator = "lte" // Less than or equal to

	// Array operators
	OpIn  Operator = "in"  // In array
	OpNin Operator = "nin" // Not in array

	// String operators
	OpContains   Operator = "contains"   // Contains substring (case-insensitive)
	OpStartsWith Operator = "startswith" // Starts with (case-insensitive)
	OpEndsWith   Operator = "endswith"   // Ends with (case-insensitive)

	// Null operators
	OpNull  Operator = "null"  // Is null
	OpNnull Operator = "nnull" // Is not null
)

// Filter represents a single field filter with an operator and value
type Filter struct {
	Field    string   // The field name to filter on
	Operator Operator // The comparison operator
	Value    any      // The value to compare against
}

// FilterGroup represents a group of filters combined with AND or OR logic
type FilterGroup struct {
	Filters []Filter // The filters to apply
	Logic   string   // "and" or "or"
}

// FieldFilterBuilder builds predicates for a specific field with various operators.
// Implementations should handle type conversion and validation for their field type.
type FieldFilterBuilder[P any] interface {
	// Equality
	Eq(value any) P
	Ne(value any) P

	// Comparison (for numeric/date fields)
	Gt(value any) P
	Gte(value any) P
	Lt(value any) P
	Lte(value any) P

	// Array membership
	In(values []any) P
	Nin(values []any) P

	// String operations (for string fields)
	Contains(value string) P
	StartsWith(value string) P
	EndsWith(value string) P

	// Null checks
	IsNull() P
	IsNotNull() P
}

// ApplyStructuredFilters applies a group of structured filters to a query.
//
// Parameters:
//   - query: The query to filter
//   - cfg: Configuration containing the where function
//   - group: The filter group to apply (filters + AND/OR logic)
//   - fieldBuilders: Map of field names to FieldFilterBuilder implementations
//   - predicates: PredicateBuilder for combining predicates
//
// Returns the modified query with all filters applied.
//
// Example:
//
//	fieldBuilders := map[string]filter.FieldFilterBuilder[predicate.User]{
//	    "email": NewUserEmailFilterBuilder(),
//	    "age":   NewUserAgeFilterBuilder(),
//	}
//
//	group := filter.FilterGroup{
//	    Filters: []filter.Filter{
//	        {Field: "email", Operator: filter.OpContains, Value: "john"},
//	        {Field: "age", Operator: filter.OpGte, Value: 18},
//	    },
//	    Logic: "and",
//	}
//
//	query = filter.ApplyStructuredFilters(query, cfg, group, fieldBuilders, predicates)
func ApplyStructuredFilters[Q any, P any](
	query Q,
	cfg Config[Q, P],
	group FilterGroup,
	fieldBuilders map[string]FieldFilterBuilder[P],
	predicates PredicateBuilder[P],
) Q {
	if len(group.Filters) == 0 {
		return query
	}

	// Build predicates for each filter
	filterPredicates := make([]P, 0, len(group.Filters))
	for _, f := range group.Filters {
		builder, ok := fieldBuilders[f.Field]
		if !ok {
			// Skip unknown fields
			continue
		}

		predicate := buildPredicate(builder, f)
		filterPredicates = append(filterPredicates, predicate)
	}

	if len(filterPredicates) == 0 {
		return query
	}

	// Combine predicates based on logic
	var combinedPredicate P
	if len(filterPredicates) == 1 {
		combinedPredicate = filterPredicates[0]
	} else {
		if group.Logic == "or" {
			combinedPredicate = predicates.Or(filterPredicates...)
		} else {
			// Default to AND
			combinedPredicate = predicates.And(filterPredicates...)
		}
	}

	return cfg.Where(query, combinedPredicate)
}

// buildPredicate builds a single predicate from a filter using the field builder
func buildPredicate[P any](builder FieldFilterBuilder[P], f Filter) P {
	switch f.Operator {
	case OpEq:
		return builder.Eq(f.Value)
	case OpNe:
		return builder.Ne(f.Value)
	case OpGt:
		return builder.Gt(f.Value)
	case OpGte:
		return builder.Gte(f.Value)
	case OpLt:
		return builder.Lt(f.Value)
	case OpLte:
		return builder.Lte(f.Value)
	case OpIn:
		// Convert value to []any if it isn't already
		values, ok := f.Value.([]any)
		if !ok {
			// Try to handle slice of concrete type
			// For now, wrap single value in slice
			values = []any{f.Value}
		}
		return builder.In(values)
	case OpNin:
		values, ok := f.Value.([]any)
		if !ok {
			values = []any{f.Value}
		}
		return builder.Nin(values)
	case OpContains:
		return builder.Contains(toString(f.Value))
	case OpStartsWith:
		return builder.StartsWith(toString(f.Value))
	case OpEndsWith:
		return builder.EndsWith(toString(f.Value))
	case OpNull:
		return builder.IsNull()
	case OpNnull:
		return builder.IsNotNull()
	default:
		// For unknown operators, default to Eq
		return builder.Eq(f.Value)
	}
}

// toString converts a value to string, handling nil
func toString(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
