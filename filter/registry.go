package filter

// FieldBuilder represents a configured filter builder for a single field.
// This is used to build up the filter builder map declaratively.
type FieldBuilder[P any] struct {
	Name    string
	Builder FieldFilterBuilder[P]
}

// BuildFilterMap constructs a map of field names to filter builders from a list of field configurations.
// This provides a declarative way to register all filters for a model.
//
// Example usage:
//
//	filterBuilders := BuildFilterMap(
//	    StringField("email", user.EmailEQ, user.EmailNEQ, ...),
//	    NullableStringField("first_name", user.FirstNameEQ, ...),
//	    BoolField("is_active", user.IsActiveEQ, user.IsActiveNEQ, ...),
//	    TimeField("created_at", user.CreatedAtEQ, user.CreatedAtNEQ, ...),
//	)
func BuildFilterMap[P any](fields ...FieldBuilder[P]) map[string]FieldFilterBuilder[P] {
	result := make(map[string]FieldFilterBuilder[P], len(fields))
	for _, field := range fields {
		result[field.Name] = field.Builder
	}
	return result
}

// StringField creates a FieldBuilder for a non-nullable string field.
// This is a convenience function that wraps NewStringFilter.
//
// Example:
//
//	StringField("email", StringPredicates[predicate.User]{
//	    Eq:         user.EmailEQ,
//	    Ne:         user.EmailNEQ,
//	    Gt:         user.EmailGT,
//	    Gte:        user.EmailGTE,
//	    Lt:         user.EmailLT,
//	    Lte:        user.EmailLTE,
//	    In:         user.EmailIn,
//	    Nin:        user.EmailNotIn,
//	    Contains:   user.EmailContainsFold,
//	    StartsWith: user.EmailHasPrefix,
//	    EndsWith:   user.EmailHasSuffix,
//	})
func StringField[P any](
	name string,
	predicates StringPredicates[P],
) FieldBuilder[P] {
	return FieldBuilder[P]{
		Name:    name,
		Builder: NewStringFilter(predicates),
	}
}

// NullableStringField creates a FieldBuilder for a nullable string field.
//
// Example:
//
//	NullableStringField("first_name", StringPredicates[predicate.User]{
//	    Eq:         user.FirstNameEQ,
//	    Ne:         user.FirstNameNEQ,
//	    // ...
//	    IsNil:      user.FirstNameIsNil,
//	    IsNotNil:   user.FirstNameNotNil,
//	})
func NullableStringField[P any](
	name string,
	predicates StringPredicates[P],
) FieldBuilder[P] {
	return FieldBuilder[P]{
		Name:    name,
		Builder: NewNullableStringFilter(predicates),
	}
}

// BoolField creates a FieldBuilder for a boolean field.
//
// Example:
//
//	BoolField("is_active", BoolPredicates[predicate.User]{
//	    Eq:  user.IsActiveEQ,
//	    Ne:  user.IsActiveNEQ,
//	    Or:  user.Or,
//	    And: user.And,
//	})
func BoolField[P any](
	name string,
	predicates BoolPredicates[P],
) FieldBuilder[P] {
	return FieldBuilder[P]{
		Name:    name,
		Builder: NewBoolFilter(predicates),
	}
}

// TimeField creates a FieldBuilder for a non-nullable time/timestamp field.
//
// Example:
//
//	TimeField("created_at", TimePredicates[predicate.User]{
//	    Eq:  user.CreatedAtEQ,
//	    Ne:  user.CreatedAtNEQ,
//	    Gt:  user.CreatedAtGT,
//	    Gte: user.CreatedAtGTE,
//	    Lt:  user.CreatedAtLT,
//	    Lte: user.CreatedAtLTE,
//	    In:  user.CreatedAtIn,
//	    Nin: user.CreatedAtNotIn,
//	})
func TimeField[P any](
	name string,
	predicates TimePredicates[P],
) FieldBuilder[P] {
	return FieldBuilder[P]{
		Name:    name,
		Builder: NewTimeFilter(predicates),
	}
}

// NullableTimeField creates a FieldBuilder for a nullable time field.
func NullableTimeField[P any](
	name string,
	predicates TimePredicates[P],
) FieldBuilder[P] {
	return FieldBuilder[P]{
		Name:    name,
		Builder: NewNullableTimeFilter(predicates),
	}
}
